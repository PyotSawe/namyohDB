package compiler

import (
	"fmt"

	"relational-db/internal/parser"
)

// NameResolver resolves table and column names to schema metadata
type NameResolver struct {
	catalog CatalogManager
	refs    *ResolvedReferences
	scope   *Scope
}

// Scope represents a naming scope (for nested subqueries)
type Scope struct {
	Tables  map[string]*TableMetadata
	Aliases map[string]string
	Parent  *Scope
}

// NewNameResolver creates a new name resolver
func NewNameResolver(catalog CatalogManager, refs *ResolvedReferences) *NameResolver {
	return &NameResolver{
		catalog: catalog,
		refs:    refs,
		scope: &Scope{
			Tables:  make(map[string]*TableMetadata),
			Aliases: make(map[string]string),
		},
	}
}

// ResolveSelect resolves names in a SELECT statement
func (nr *NameResolver) ResolveSelect(stmt *parser.SelectStatement) error {
	// Step 1: Resolve FROM clause (tables)
	if stmt.FromClause != nil {
		if err := nr.resolveFromClause(stmt.FromClause); err != nil {
			return err
		}
	}

	// Step 2: Resolve SELECT columns
	if stmt.SelectClause != nil {
		for _, col := range stmt.SelectClause.Columns {
			if err := nr.resolveExpression(col); err != nil {
				return err
			}
		}
	}

	// Step 3: Resolve WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		if err := nr.resolveExpression(stmt.WhereClause.Condition); err != nil {
			return err
		}
	}

	// Step 4: Resolve GROUP BY
	if stmt.GroupBy != nil {
		for _, expr := range stmt.GroupBy.Columns {
			if err := nr.resolveExpression(expr); err != nil {
				return err
			}
		}
	}

	// Step 5: Resolve HAVING
	if stmt.Having != nil && stmt.Having.Condition != nil {
		if err := nr.resolveExpression(stmt.Having.Condition); err != nil {
			return err
		}
	}

	// Step 6: Resolve ORDER BY
	if stmt.OrderBy != nil {
		for _, order := range stmt.OrderBy.Orders {
			if err := nr.resolveExpression(order.Expression); err != nil {
				return err
			}
		}
	}

	return nil
}

// ResolveInsert resolves names in an INSERT statement
func (nr *NameResolver) ResolveInsert(stmt *parser.InsertStatement) error {
	// Resolve table name
	tableName := stmt.TableName.Value
	table, err := nr.catalog.GetTable(tableName)
	if err != nil {
		return fmt.Errorf("table not found: %s", tableName)
	}

	nr.refs.AddTable(tableName, table)
	nr.scope.Tables[tableName] = table

	// Resolve column names if specified
	for _, col := range stmt.Columns {
		colName := col.Value
		if !table.HasColumn(colName) {
			return fmt.Errorf("column %s not found in table %s", colName, tableName)
		}
	}

	// Resolve value expressions
	for _, valueList := range stmt.Values {
		for _, expr := range valueList {
			if err := nr.resolveExpression(expr); err != nil {
				return err
			}
		}
	}

	return nil
}

// ResolveUpdate resolves names in an UPDATE statement
func (nr *NameResolver) ResolveUpdate(stmt *parser.UpdateStatement) error {
	// Resolve table name
	tableName := stmt.TableName.Value
	table, err := nr.catalog.GetTable(tableName)
	if err != nil {
		return fmt.Errorf("table not found: %s", tableName)
	}

	nr.refs.AddTable(tableName, table)
	nr.scope.Tables[tableName] = table

	// Resolve SET clause columns and values
	for _, setClause := range stmt.SetClauses {
		colName := setClause.Column.Value
		if !table.HasColumn(colName) {
			return fmt.Errorf("column %s not found in table %s", colName, tableName)
		}

		if err := nr.resolveExpression(setClause.Value); err != nil {
			return err
		}
	}

	// Resolve WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		if err := nr.resolveExpression(stmt.WhereClause.Condition); err != nil {
			return err
		}
	}

	return nil
}

// ResolveDelete resolves names in a DELETE statement
func (nr *NameResolver) ResolveDelete(stmt *parser.DeleteStatement) error {
	// Resolve table name
	tableName := stmt.TableName.Value
	table, err := nr.catalog.GetTable(tableName)
	if err != nil {
		return fmt.Errorf("table not found: %s", tableName)
	}

	nr.refs.AddTable(tableName, table)
	nr.scope.Tables[tableName] = table

	// Resolve WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		if err := nr.resolveExpression(stmt.WhereClause.Condition); err != nil {
			return err
		}
	}

	return nil
}

// ResolveCreateTable resolves names in a CREATE TABLE statement
func (nr *NameResolver) ResolveCreateTable(stmt *parser.CreateTableStatement) error {
	// Check if table already exists
	tableName := stmt.TableName.Value
	if nr.catalog.TableExists(tableName) {
		return fmt.Errorf("table %s already exists", tableName)
	}

	// For CREATE TABLE, validate column definitions and constraints
	// Actual table creation happens in execution layer
	for _, colDef := range stmt.Columns {
		// Validate foreign key references if present
		for _, constraint := range colDef.Constraints {
			if constraint.Type == parser.ForeignKey && constraint.References != nil {
				refTableName := constraint.References.Table.Value
				if !nr.catalog.TableExists(refTableName) {
					return fmt.Errorf("referenced table %s does not exist", refTableName)
				}
			}
		}
	}

	return nil
}

// ResolveDropTable resolves names in a DROP TABLE statement
func (nr *NameResolver) ResolveDropTable(stmt *parser.DropTableStatement) error {
	// Check if table exists (unless IF EXISTS is used)
	tableName := stmt.TableName.Value
	if !stmt.IfExists && !nr.catalog.TableExists(tableName) {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	return nil
}

// resolveFromClause resolves table references in FROM clause
func (nr *NameResolver) resolveFromClause(from *parser.FromClause) error {
	// Resolve tables in FROM clause
	for _, tableExpr := range from.Tables {
		if err := nr.resolveTableExpression(tableExpr); err != nil {
			return err
		}
	}

	// TODO: Handle JOINs when fully implemented in parser
	for _, join := range from.Joins {
		if join.Table != nil {
			if err := nr.resolveTableExpression(join.Table); err != nil {
				return err
			}
		}
		if join.Condition != nil {
			if err := nr.resolveExpression(join.Condition); err != nil {
				return err
			}
		}
	}

	return nil
}

// resolveTableExpression resolves a table reference expression
func (nr *NameResolver) resolveTableExpression(expr parser.Expression) error {
	switch e := expr.(type) {
	case *parser.Identifier:
		tableName := e.Value
		table, err := nr.catalog.GetTable(tableName)
		if err != nil {
			return fmt.Errorf("table not found: %s", tableName)
		}

		// Use alias if provided, otherwise use table name
		refName := tableName
		if e.Alias != nil {
			refName = e.Alias.Value
			nr.refs.AddAlias(refName, tableName)
			nr.scope.Aliases[refName] = tableName
		}

		nr.refs.AddTable(refName, table)
		nr.scope.Tables[refName] = table

		return nil

	default:
		// Other table expressions (subqueries, etc.) - TODO
		return nil
	}
}

// resolveExpression resolves column references in an expression
func (nr *NameResolver) resolveExpression(expr parser.Expression) error {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *parser.Identifier:
		// Simple column reference
		return nr.resolveColumnReference(e.Value, "")

	case *parser.ColumnReference:
		// Qualified column reference (table.column)
		tableName := ""
		if e.Table != nil {
			tableName = e.Table.Value
		}
		return nr.resolveColumnReference(e.Column.Value, tableName)

	case *parser.BinaryExpression:
		if err := nr.resolveExpression(e.Left); err != nil {
			return err
		}
		return nr.resolveExpression(e.Right)

	case *parser.UnaryExpression:
		return nr.resolveExpression(e.Operand)

	case *parser.FunctionCall:
		for _, arg := range e.Arguments {
			if err := nr.resolveExpression(arg); err != nil {
				return err
			}
		}

	case *parser.Literal:
		// Literals don't need resolution
		return nil

	case *parser.Wildcard:
		// Wildcard: validate table if qualified
		if e.Table != nil {
			tableName := e.Table.Value
			if _, found := nr.scope.Tables[tableName]; !found {
				// Try resolving alias
				if realName, ok := nr.scope.Aliases[tableName]; ok {
					if _, found := nr.scope.Tables[realName]; !found {
						return fmt.Errorf("table not found: %s", tableName)
					}
				} else {
					return fmt.Errorf("table not found: %s", tableName)
				}
			}
		}
		return nil

	default:
		// Other expression types don't need resolution
		return nil
	}

	return nil
}

// resolveColumnReference resolves a column name to schema metadata
func (nr *NameResolver) resolveColumnReference(columnName, tableName string) error {
	// If table is specified (qualified reference)
	if tableName != "" {
		// Resolve table (could be alias)
		table, found := nr.scope.Tables[tableName]
		if !found {
			// Try resolving alias
			if realName, ok := nr.scope.Aliases[tableName]; ok {
				table, found = nr.scope.Tables[realName]
			}
		}

		if !found {
			return fmt.Errorf("table not found: %s", tableName)
		}

		// Resolve column in table
		col, err := table.GetColumn(columnName)
		if err != nil {
			return err
		}

		qualifiedName := tableName + "." + columnName
		nr.refs.AddColumn(qualifiedName, col)
		return nil
	}

	// Unqualified column name - search all tables in scope
	var foundColumn *ColumnMetadata
	var foundInTable string
	matchCount := 0

	for tblName, table := range nr.scope.Tables {
		if col, err := table.GetColumn(columnName); err == nil {
			foundColumn = col
			foundInTable = tblName
			matchCount++
		}
	}

	if matchCount == 0 {
		return fmt.Errorf("column not found: %s", columnName)
	}

	if matchCount > 1 {
		return fmt.Errorf("ambiguous column reference: %s", columnName)
	}

	// Add fully qualified name to refs
	qualifiedName := foundInTable + "." + columnName
	nr.refs.AddColumn(qualifiedName, foundColumn)

	return nil
}
