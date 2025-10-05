package semantic

import (
	"fmt"

	"relational-db/internal/compiler"
	"relational-db/internal/parser"
)

// GroupByValidationRule validates GROUP BY semantics
type GroupByValidationRule struct{}

// Name returns the rule name
func (r *GroupByValidationRule) Name() string {
	return "GROUP BY Validation"
}

// Validate validates GROUP BY semantics
func (r *GroupByValidationRule) Validate(compiled *compiler.CompiledQuery, ctx *ValidationContext) error {
	// Only validate SELECT statements with GROUP BY
	selectStmt, ok := compiled.Statement.(*parser.SelectStatement)
	if !ok || selectStmt.GroupBy == nil || len(selectStmt.GroupBy.Columns) == 0 {
		return nil
	}

	ctx.HasGroupBy = true
	ctx.GroupByInfo = &GroupByMetadata{
		Expressions: selectStmt.GroupBy.Columns,
		Columns:     make([]string, 0),
		Aggregates:  make([]string, 0),
		Valid:       true,
		Errors:      make([]string, 0),
	}

	// TODO: Implement full GROUP BY validation
	// For now, just mark as present
	return nil
}

// AggregateValidationRule validates aggregate function usage
type AggregateValidationRule struct{}

// Name returns the rule name
func (r *AggregateValidationRule) Name() string {
	return "Aggregate Function Validation"
}

// Validate validates aggregate function usage
func (r *AggregateValidationRule) Validate(compiled *compiler.CompiledQuery, ctx *ValidationContext) error {
	// Check for aggregate functions in the query
	selectStmt, ok := compiled.Statement.(*parser.SelectStatement)
	if !ok {
		return nil
	}

	// Initialize aggregate metadata
	ctx.AggregateInfo = &AggregateMetadata{
		Functions:      make([]*AggregateFunctionInfo, 0),
		GroupByColumns: make([]string, 0),
	}

	// Walk SELECT clause for aggregates
	for _, col := range selectStmt.SelectClause.Columns {
		if r.hasAggregate(col) {
			ctx.HasAggregates = true
			ctx.AggregateInfo.InSelect = true
		}
	}

	// Walk WHERE clause - aggregates not allowed
	if selectStmt.WhereClause != nil {
		ctx.InWhereClause = true
		if r.hasAggregate(selectStmt.WhereClause.Condition) {
			return NewAggregateError(
				ErrAggregateInWhere,
				"Aggregate functions are not allowed in WHERE clause",
			).WithHint("Use HAVING clause instead")
		}
		ctx.InWhereClause = false
	}

	// Walk HAVING clause for aggregates
	if selectStmt.Having != nil {
		if r.hasAggregate(selectStmt.Having.Condition) {
			ctx.HasAggregates = true
			ctx.AggregateInfo.InHaving = true
		}
	}

	// Walk ORDER BY clause for aggregates
	if selectStmt.OrderBy != nil {
		for _, order := range selectStmt.OrderBy.Orders {
			if r.hasAggregate(order.Expression) {
				ctx.HasAggregates = true
				ctx.AggregateInfo.InOrderBy = true
			}
		}
	}

	return nil
}

// hasAggregate checks if an expression contains an aggregate function
func (r *AggregateValidationRule) hasAggregate(expr parser.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *parser.FunctionCall:
		// Check if it's an aggregate function
		name := e.Name.Value
		if name == "COUNT" || name == "SUM" || name == "AVG" || name == "MAX" || name == "MIN" {
			return true
		}
		// Check arguments
		for _, arg := range e.Arguments {
			if r.hasAggregate(arg) {
				return true
			}
		}

	case *parser.BinaryExpression:
		return r.hasAggregate(e.Left) || r.hasAggregate(e.Right)

	case *parser.UnaryExpression:
		return r.hasAggregate(e.Operand)
	}

	return false
}

// SubqueryValidationRule validates subquery semantics
type SubqueryValidationRule struct{}

// Name returns the rule name
func (r *SubqueryValidationRule) Name() string {
	return "Subquery Validation"
}

// Validate validates subquery semantics
func (r *SubqueryValidationRule) Validate(compiled *compiler.CompiledQuery, ctx *ValidationContext) error {
	// TODO: Implement subquery validation
	// For now, just initialize the metadata
	ctx.SubqueryInfo = make([]*SubqueryMetadata, 0)
	return nil
}

// SchemaValidationRule validates schema-related operations
type SchemaValidationRule struct {
	catalog compiler.CatalogManager
}

// Name returns the rule name
func (r *SchemaValidationRule) Name() string {
	return "Schema Validation"
}

// Validate validates schema-related operations
func (r *SchemaValidationRule) Validate(compiled *compiler.CompiledQuery, ctx *ValidationContext) error {
	switch stmt := compiled.Statement.(type) {
	case *parser.CreateTableStatement:
		return r.validateCreateTable(stmt)

	case *parser.DropTableStatement:
		return r.validateDropTable(stmt)

	default:
		return nil
	}
}

// validateCreateTable validates CREATE TABLE statement
func (r *SchemaValidationRule) validateCreateTable(stmt *parser.CreateTableStatement) error {
	tableName := stmt.TableName.Value

	// Check if table already exists
	if r.catalog != nil && r.catalog.TableExists(tableName) {
		return NewSchemaError(
			ErrTableAlreadyExists,
			fmt.Sprintf("Table '%s' already exists", tableName),
		).WithHint("Use IF NOT EXISTS clause to avoid this error")
	}

	// Check for duplicate column names
	columnNames := make(map[string]bool)
	for _, col := range stmt.Columns {
		colName := col.Name.Value
		if columnNames[colName] {
			return NewSchemaError(
				ErrDuplicateColumn,
				fmt.Sprintf("Duplicate column name '%s' in table '%s'", colName, tableName),
			)
		}
		columnNames[colName] = true
	}

	// Validate foreign key references
	for _, col := range stmt.Columns {
		for _, constraint := range col.Constraints {
			if constraint.References != nil {
				refTable := constraint.References.Table.Value
				if r.catalog != nil && !r.catalog.TableExists(refTable) {
					return NewSchemaError(
						ErrForeignKeyRefNotFound,
						fmt.Sprintf("Foreign key references non-existent table '%s'", refTable),
					)
				}
			}
		}
	}

	return nil
}

// validateDropTable validates DROP TABLE statement
func (r *SchemaValidationRule) validateDropTable(stmt *parser.DropTableStatement) error {
	tableName := stmt.TableName.Value

	// Check if table exists
	if r.catalog != nil && !r.catalog.TableExists(tableName) {
		if !stmt.IfExists {
			return NewSchemaError(
				ErrTableNotFound,
				fmt.Sprintf("Table '%s' does not exist", tableName),
			).WithHint("Use IF EXISTS clause to avoid this error")
		}
	}

	return nil
}
