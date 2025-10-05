// Package compiler implements the Query Compiler for NamyohDB.package compiler

// It transforms parsed AST nodes into validated, compiled query representations
// with name resolution, type checking, and constraint validation.
package compiler

import (
	"fmt"
	"strings"
	"time"

	"relational-db/internal/lexer"
	"relational-db/internal/parser"
)

// QueryType represents the type of SQL query
type QueryType int

const (
	QueryTypeUnknown QueryType = iota

	// DML (Data Manipulation Language)
	QueryTypeSelect
	QueryTypeInsert
	QueryTypeUpdate
	QueryTypeDelete

	// DDL (Data Definition Language)
	QueryTypeCreateTable
	QueryTypeDropTable
	QueryTypeCreateIndex
	QueryTypeDropIndex
	QueryTypeAlterTable

	// TCL (Transaction Control Language)
	QueryTypeBegin
	QueryTypeCommit
	QueryTypeRollback
	QueryTypeSavepoint
)

// String returns the string representation of QueryType
func (qt QueryType) String() string {
	switch qt {
	case QueryTypeSelect:
		return "SELECT"
	case QueryTypeInsert:
		return "INSERT"
	case QueryTypeUpdate:
		return "UPDATE"
	case QueryTypeDelete:
		return "DELETE"
	case QueryTypeCreateTable:
		return "CREATE_TABLE"
	case QueryTypeDropTable:
		return "DROP_TABLE"
	case QueryTypeCreateIndex:
		return "CREATE_INDEX"
	case QueryTypeDropIndex:
		return "DROP_INDEX"
	case QueryTypeAlterTable:
		return "ALTER_TABLE"
	case QueryTypeBegin:
		return "BEGIN"
	case QueryTypeCommit:
		return "COMMIT"
	case QueryTypeRollback:
		return "ROLLBACK"
	case QueryTypeSavepoint:
		return "SAVEPOINT"
	default:
		return "UNKNOWN"
	}
}

// IsDML returns true if this is a DML query
func (qt QueryType) IsDML() bool {
	return qt >= QueryTypeSelect && qt <= QueryTypeDelete
}

// IsDDL returns true if this is a DDL query
func (qt QueryType) IsDDL() bool {
	return qt >= QueryTypeCreateTable && qt <= QueryTypeAlterTable
}

// IsTCL returns true if this is a TCL query
func (qt QueryType) IsTCL() bool {
	return qt >= QueryTypeBegin && qt <= QueryTypeSavepoint
}

// CompiledQuery represents a fully validated and compiled SQL query
type CompiledQuery struct {
	// Original AST from parser
	Statement parser.Statement

	// Query classification
	QueryType QueryType

	// Name resolution results
	ResolvedRefs *ResolvedReferences

	// Type information for all expressions
	TypeInfo *TypeInformation

	// Validation status
	Validated bool
	Errors    []CompilationError

	// Metadata for optimizer
	Metadata map[string]interface{}

	// Compilation timestamp
	CompiledAt time.Time
}

// HasErrors returns true if compilation encountered errors
func (cq *CompiledQuery) HasErrors() bool {
	return len(cq.Errors) > 0
}

// AddError adds a compilation error
func (cq *CompiledQuery) AddError(err CompilationError) {
	cq.Errors = append(cq.Errors, err)
	cq.Validated = false
}

// QueryCompiler compiles parsed AST into validated query representations
type QueryCompiler struct {
	catalog CatalogManager
}

// NewQueryCompiler creates a new query compiler
func NewQueryCompiler(catalog CatalogManager) *QueryCompiler {
	return &QueryCompiler{
		catalog: catalog,
	}
}

// Compile compiles an AST statement into a validated CompiledQuery
func (qc *QueryCompiler) Compile(ast parser.Statement) (*CompiledQuery, error) {
	compiled := &CompiledQuery{
		Statement:    ast,
		ResolvedRefs: NewResolvedReferences(),
		TypeInfo:     NewTypeInformation(),
		Metadata:     make(map[string]interface{}),
		CompiledAt:   time.Now(),
		Validated:    false,
	}

	// Step 1: Identify query type
	compiled.QueryType = qc.identifyQueryType(ast)

	// Step 2: Resolve names (tables, columns, aliases)
	if err := qc.resolveNames(ast, compiled.ResolvedRefs); err != nil {
		compiled.AddError(CompilationError{
			Code:     ErrNameResolution,
			Category: ErrorCategoryNameResolution,
			Message:  err.Error(),
		})
		return compiled, err
	}

	// Step 3: Check and infer types
	if err := qc.checkTypes(ast, compiled.ResolvedRefs, compiled.TypeInfo); err != nil {
		compiled.AddError(CompilationError{
			Code:     ErrTypeChecking,
			Category: ErrorCategoryTypeChecking,
			Message:  err.Error(),
		})
		return compiled, err
	}

	// Step 4: Validate constraints
	if err := qc.validateConstraints(ast, compiled.ResolvedRefs); err != nil {
		compiled.AddError(CompilationError{
			Code:     ErrConstraintValidation,
			Category: ErrorCategoryConstraintValidation,
			Message:  err.Error(),
		})
		return compiled, err
	}

	// Step 5: Validate aggregates (if SELECT)
	if selectStmt, ok := ast.(*parser.SelectStatement); ok {
		if err := qc.validateAggregates(selectStmt, compiled.ResolvedRefs); err != nil {
			compiled.AddError(CompilationError{
				Code:     ErrInvalidAggregate,
				Category: ErrorCategorySemanticAnalysis,
				Message:  err.Error(),
			})
			return compiled, err
		}
	}

	// Mark as validated if no errors
	compiled.Validated = !compiled.HasErrors()
	return compiled, nil
}

// identifyQueryType determines the type of SQL statement
func (qc *QueryCompiler) identifyQueryType(ast parser.Statement) QueryType {
	switch ast.(type) {
	case *parser.SelectStatement:
		return QueryTypeSelect
	case *parser.InsertStatement:
		return QueryTypeInsert
	case *parser.UpdateStatement:
		return QueryTypeUpdate
	case *parser.DeleteStatement:
		return QueryTypeDelete
	case *parser.CreateTableStatement:
		return QueryTypeCreateTable
	case *parser.DropTableStatement:
		return QueryTypeDropTable
	default:
		return QueryTypeUnknown
	}
}

// resolveNames resolves all table and column references
func (qc *QueryCompiler) resolveNames(ast parser.Statement, refs *ResolvedReferences) error {
	resolver := NewNameResolver(qc.catalog, refs)

	switch stmt := ast.(type) {
	case *parser.SelectStatement:
		return resolver.ResolveSelect(stmt)
	case *parser.InsertStatement:
		return resolver.ResolveInsert(stmt)
	case *parser.UpdateStatement:
		return resolver.ResolveUpdate(stmt)
	case *parser.DeleteStatement:
		return resolver.ResolveDelete(stmt)
	case *parser.CreateTableStatement:
		return resolver.ResolveCreateTable(stmt)
	case *parser.DropTableStatement:
		return resolver.ResolveDropTable(stmt)
	default:
		return fmt.Errorf("unsupported statement type for name resolution")
	}
}

// checkTypes performs type checking and inference
func (qc *QueryCompiler) checkTypes(ast parser.Statement, refs *ResolvedReferences, typeInfo *TypeInformation) error {
	checker := NewTypeChecker(refs, typeInfo)

	switch stmt := ast.(type) {
	case *parser.SelectStatement:
		return checker.CheckSelect(stmt)
	case *parser.InsertStatement:
		return checker.CheckInsert(stmt)
	case *parser.UpdateStatement:
		return checker.CheckUpdate(stmt)
	case *parser.DeleteStatement:
		return checker.CheckDelete(stmt)
	case *parser.CreateTableStatement:
		return checker.CheckCreateTable(stmt)
	default:
		return nil // No type checking needed for other statements
	}
}

// validateConstraints validates SQL constraints
func (qc *QueryCompiler) validateConstraints(ast parser.Statement, refs *ResolvedReferences) error {
	validator := NewConstraintValidator(qc.catalog, refs)

	switch stmt := ast.(type) {
	case *parser.CreateTableStatement:
		return validator.ValidateCreateTable(stmt)
	case *parser.InsertStatement:
		return validator.ValidateInsert(stmt)
	case *parser.UpdateStatement:
		return validator.ValidateUpdate(stmt)
	default:
		return nil // No constraint validation needed
	}
}

// validateAggregates validates aggregate function usage
func (qc *QueryCompiler) validateAggregates(stmt *parser.SelectStatement, refs *ResolvedReferences) error {
	validator := NewAggregateValidator(refs)
	return validator.Validate(stmt)
}

// CompileSQL is a convenience function that parses and compiles SQL in one step
func (qc *QueryCompiler) CompileSQL(sql string) (*CompiledQuery, error) {
	// Parse SQL to AST
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)
	ast := p.ParseStatement()

	// Check for parse errors
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse error: %v", p.Errors()[0])
	}

	if ast == nil {
		return nil, fmt.Errorf("failed to parse SQL statement")
	}

	// Compile AST
	return qc.Compile(ast)
}

// CatalogManager interface for schema metadata access
type CatalogManager interface {
	// GetTable retrieves table metadata by name
	GetTable(name string) (*TableMetadata, error)

	// GetColumn retrieves column metadata
	GetColumn(table, column string) (*ColumnMetadata, error)

	// TableExists checks if a table exists
	TableExists(name string) bool

	// ListTables returns all table names
	ListTables() ([]string, error)
}

// ResolvedReferences stores all name resolution results
type ResolvedReferences struct {
	// Table references: alias/name → TableMetadata
	Tables map[string]*TableMetadata

	// Column references: qualified name → ColumnMetadata
	Columns map[string]*ColumnMetadata

	// Alias mappings: alias → real table name
	Aliases map[string]string
}

// NewResolvedReferences creates a new ResolvedReferences
func NewResolvedReferences() *ResolvedReferences {
	return &ResolvedReferences{
		Tables:  make(map[string]*TableMetadata),
		Columns: make(map[string]*ColumnMetadata),
		Aliases: make(map[string]string),
	}
}

// AddTable adds a table reference
func (rr *ResolvedReferences) AddTable(name string, table *TableMetadata) {
	rr.Tables[name] = table
}

// GetTable retrieves a table by name or alias
func (rr *ResolvedReferences) GetTable(name string) (*TableMetadata, bool) {
	// Check if it's an alias first
	if realName, ok := rr.Aliases[name]; ok {
		table, found := rr.Tables[realName]
		return table, found
	}

	// Direct lookup
	table, found := rr.Tables[name]
	return table, found
}

// AddAlias adds an alias mapping
func (rr *ResolvedReferences) AddAlias(alias, realName string) {
	rr.Aliases[alias] = realName
}

// AddColumn adds a column reference
func (rr *ResolvedReferences) AddColumn(qualifiedName string, col *ColumnMetadata) {
	rr.Columns[qualifiedName] = col
}

// GetColumn retrieves a column by qualified name
func (rr *ResolvedReferences) GetColumn(qualifiedName string) (*ColumnMetadata, bool) {
	col, found := rr.Columns[qualifiedName]
	return col, found
}

// TypeInformation stores inferred type information
type TypeInformation struct {
	// Expression ID → inferred type
	ExpressionTypes map[string]DataType

	// Expression ID → type coercion needed
	Coercions map[string]TypeCoercion

	// Expression ID → null possibility
	Nullability map[string]bool
}

// NewTypeInformation creates a new TypeInformation
func NewTypeInformation() *TypeInformation {
	return &TypeInformation{
		ExpressionTypes: make(map[string]DataType),
		Coercions:       make(map[string]TypeCoercion),
		Nullability:     make(map[string]bool),
	}
}

// SetType sets the inferred type for an expression
func (ti *TypeInformation) SetType(exprID string, dataType DataType) {
	ti.ExpressionTypes[exprID] = dataType
}

// GetType retrieves the inferred type for an expression
func (ti *TypeInformation) GetType(exprID string) (DataType, bool) {
	dt, found := ti.ExpressionTypes[exprID]
	return dt, found
}

// AddCoercion records a type coercion
func (ti *TypeInformation) AddCoercion(exprID string, from, to DataType, reason string) {
	ti.Coercions[exprID] = TypeCoercion{
		FromType: from,
		ToType:   to,
		Reason:   reason,
	}
}

// TypeCoercion represents a type conversion
type TypeCoercion struct {
	FromType DataType
	ToType   DataType
	Reason   string
}

// DataType represents SQL data types
type DataType int

const (
	DataTypeUnknown DataType = iota

	// Numeric types
	DataTypeInteger // INT, INTEGER
	DataTypeReal    // REAL, FLOAT, DOUBLE
	DataTypeNumeric // NUMERIC, DECIMAL

	// String types
	DataTypeText // TEXT, VARCHAR, CHAR
	DataTypeBlob // BLOB, BINARY

	// Other types
	DataTypeBoolean   // BOOLEAN, BOOL
	DataTypeDate      // DATE
	DataTypeTime      // TIME
	DataTypeTimestamp // TIMESTAMP, DATETIME
	DataTypeNull      // NULL type
)

// String returns the string representation of DataType
func (dt DataType) String() string {
	switch dt {
	case DataTypeInteger:
		return "INTEGER"
	case DataTypeReal:
		return "REAL"
	case DataTypeNumeric:
		return "NUMERIC"
	case DataTypeText:
		return "TEXT"
	case DataTypeBlob:
		return "BLOB"
	case DataTypeBoolean:
		return "BOOLEAN"
	case DataTypeDate:
		return "DATE"
	case DataTypeTime:
		return "TIME"
	case DataTypeTimestamp:
		return "TIMESTAMP"
	case DataTypeNull:
		return "NULL"
	default:
		return "UNKNOWN"
	}
}

// IsNumeric returns true if this is a numeric type
func (dt DataType) IsNumeric() bool {
	return dt == DataTypeInteger || dt == DataTypeReal || dt == DataTypeNumeric
}

// IsString returns true if this is a string type
func (dt DataType) IsString() bool {
	return dt == DataTypeText || dt == DataTypeBlob
}

// IsComparable returns true if two types can be compared
func (dt DataType) IsComparable(other DataType) bool {
	// NULL is comparable with everything
	if dt == DataTypeNull || other == DataTypeNull {
		return true
	}

	// Same type is always comparable
	if dt == other {
		return true
	}

	// Numeric types are comparable with each other
	if dt.IsNumeric() && other.IsNumeric() {
		return true
	}

	return false
}

// CanCoerceTo returns true if this type can be coerced to another
func (dt DataType) CanCoerceTo(other DataType) bool {
	// Same type - no coercion needed
	if dt == other {
		return true
	}

	// NULL can coerce to anything
	if dt == DataTypeNull {
		return true
	}

	// Integer can coerce to Real (widening)
	if dt == DataTypeInteger && other == DataTypeReal {
		return true
	}

	// Integer can coerce to Numeric
	if dt == DataTypeInteger && other == DataTypeNumeric {
		return true
	}

	return false
}

// TableMetadata contains table schema information
type TableMetadata struct {
	Name      string
	Schema    string
	Columns   []*ColumnMetadata
	ColumnMap map[string]*ColumnMetadata
	RowCount  int64
	TotalSize int64
	TableID   uint64
	CreatedAt time.Time
}

// NewTableMetadata creates a new TableMetadata
func NewTableMetadata(name string) *TableMetadata {
	return &TableMetadata{
		Name:      name,
		Columns:   make([]*ColumnMetadata, 0),
		ColumnMap: make(map[string]*ColumnMetadata),
		CreatedAt: time.Now(),
	}
}

// AddColumn adds a column to the table
func (tm *TableMetadata) AddColumn(col *ColumnMetadata) {
	tm.Columns = append(tm.Columns, col)
	tm.ColumnMap[col.Name] = col
	col.Position = len(tm.Columns) - 1
}

// GetColumn retrieves a column by name
func (tm *TableMetadata) GetColumn(name string) (*ColumnMetadata, error) {
	col, found := tm.ColumnMap[strings.ToLower(name)]
	if !found {
		return nil, fmt.Errorf("column %s not found in table %s", name, tm.Name)
	}
	return col, nil
}

// HasColumn checks if a column exists
func (tm *TableMetadata) HasColumn(name string) bool {
	_, found := tm.ColumnMap[strings.ToLower(name)]
	return found
}

// ColumnMetadata contains column schema information
type ColumnMetadata struct {
	Name         string
	TableName    string
	Position     int
	DataType     DataType
	Nullable     bool
	IsPrimaryKey bool
	IsUnique     bool
	HasDefault   bool
	DefaultValue interface{}
	ColumnID     uint32
}

// NewColumnMetadata creates a new ColumnMetadata
func NewColumnMetadata(name string, dataType DataType) *ColumnMetadata {
	return &ColumnMetadata{
		Name:     name,
		DataType: dataType,
		Nullable: true, // Default to nullable
	}
}

// QualifiedName returns the fully qualified column name
func (cm *ColumnMetadata) QualifiedName() string {
	if cm.TableName != "" {
		return cm.TableName + "." + cm.Name
	}
	return cm.Name
}

// IsNumeric returns true if this is a numeric column
func (cm *ColumnMetadata) IsNumeric() bool {
	return cm.DataType.IsNumeric()
}

// CanBeNull returns true if this column can contain NULL
func (cm *ColumnMetadata) CanBeNull() bool {
	return cm.Nullable && !cm.IsPrimaryKey
}
