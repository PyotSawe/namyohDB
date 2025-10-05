# Query Compiler Data Structures

## Overview
This document describes the data structures used in the Query Compiler module. The compiler transforms AST nodes into validated, enriched representations with type information, resolved references, and validation metadata.

---

## Core Data Structures

### 1. CompiledQuery

The main output of the compilation process, containing all validated and enriched information.

```go
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
    Errors []CompilationError
    
    // Metadata for optimizer
    Metadata map[string]interface{}
    
    // Compilation timestamp
    CompiledAt time.Time
}
```

**Usage**:
```go
compiled := compiler.Compile(ast)
if !compiled.Validated {
    for _, err := range compiled.Errors {
        fmt.Println(err)
    }
}
```

**Memory Size**: ~200-500 bytes + size of AST and metadata

---

### 2. QueryType Enumeration

Classification of SQL statement types for routing to appropriate handlers.

```go
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
    
    // DCL (Data Control Language)
    QueryTypeGrant
    QueryTypeRevoke
)

func (qt QueryType) String() string
func (qt QueryType) IsDML() bool
func (qt QueryType) IsDDL() bool
func (qt QueryType) IsTCL() bool
```

**Usage**: Route queries to appropriate execution handlers based on type.

---

### 3. ResolvedReferences

Stores all name resolution results linking symbolic names to actual database objects.

```go
type ResolvedReferences struct {
    // Table references: alias/name → TableMetadata
    Tables map[string]*TableMetadata
    
    // Column references: qualified name → ColumnMetadata
    Columns map[string]*ColumnMetadata
    
    // Alias mappings: alias → real table name
    Aliases map[string]string
    
    // Subquery references: subquery ID → CompiledQuery
    Subqueries map[string]*CompiledQuery
    
    // Function references: function name → FunctionMetadata
    Functions map[string]*FunctionMetadata
}

// Constructor
func NewResolvedReferences() *ResolvedReferences

// Methods
func (rr *ResolvedReferences) AddTable(alias string, table *TableMetadata)
func (rr *ResolvedReferences) GetTable(name string) (*TableMetadata, bool)
func (rr *ResolvedReferences) AddColumn(qualifiedName string, col *ColumnMetadata)
func (rr *ResolvedReferences) ResolveColumn(name string, contextTable string) (*ColumnMetadata, error)
```

**Example**:
```go
// For query: SELECT u.name FROM users AS u
refs := NewResolvedReferences()
refs.AddTable("u", usersTableMetadata)
refs.AddColumn("u.name", nameColumnMetadata)
```

**Memory Size**: O(t + c) where t = tables, c = columns

---

### 4. TableMetadata

Complete information about a table from the catalog.

```go
type TableMetadata struct {
    // Basic information
    Name       string
    Schema     string  // Database schema name
    
    // Columns
    Columns    []*ColumnMetadata
    ColumnMap  map[string]*ColumnMetadata  // Fast column lookup
    
    // Constraints
    PrimaryKey *PrimaryKeyConstraint
    ForeignKeys []*ForeignKeyConstraint
    UniqueKeys []*UniqueConstraint
    CheckConstraints []*CheckConstraint
    
    // Statistics (for optimizer)
    RowCount     int64
    TotalSize    int64
    CreatedAt    time.Time
    LastModified time.Time
    
    // Storage information
    TableID      uint64
    StoragePath  string
}

func (tm *TableMetadata) GetColumn(name string) (*ColumnMetadata, error)
func (tm *TableMetadata) HasColumn(name string) bool
func (tm *TableMetadata) GetPrimaryKeyColumns() []*ColumnMetadata
```

**Memory Size**: ~100 bytes + (columns × 50 bytes) + (constraints × 30 bytes)

---

### 5. ColumnMetadata

Complete information about a table column.

```go
type ColumnMetadata struct {
    // Basic information
    Name         string
    TableName    string
    Position     int  // Column position in table (0-indexed)
    
    // Type information
    DataType     DataType
    TypeModifier string    // e.g., "VARCHAR(100)"
    Nullable     bool
    
    // Constraints
    IsPrimaryKey bool
    IsUnique     bool
    HasDefault   bool
    DefaultValue interface{}
    
    // Statistics (for optimizer)
    Cardinality  int64      // Number of distinct values
    MinValue     interface{}
    MaxValue     interface{}
    NullCount    int64
    
    // Storage information
    ColumnID     uint32
    StorageSize  int32
}

func (cm *ColumnMetadata) QualifiedName() string  // Returns "table.column"
func (cm *ColumnMetadata) IsNumeric() bool
func (cm *ColumnMetadata) CanBeNull() bool
```

**Example**:
```go
col := &ColumnMetadata{
    Name: "age",
    TableName: "users",
    DataType: DataTypeInteger,
    Nullable: false,
    IsPrimaryKey: false,
}
```

**Memory Size**: ~80 bytes

---

### 6. TypeInformation

Stores inferred type information for all expressions in the query.

```go
type TypeInformation struct {
    // Expression ID → inferred type
    ExpressionTypes map[string]DataType
    
    // Expression ID → type coercion needed
    Coercions map[string]TypeCoercion
    
    // Expression ID → null possibility
    Nullability map[string]bool
}

type TypeCoercion struct {
    FromType DataType
    ToType   DataType
    Reason   string  // Why coercion is needed
}

func NewTypeInformation() *TypeInformation
func (ti *TypeInformation) SetType(exprID string, dataType DataType)
func (ti *TypeInformation) GetType(exprID string) (DataType, bool)
func (ti *TypeInformation) AddCoercion(exprID string, from, to DataType, reason string)
```

**Example**:
```go
typeInfo := NewTypeInformation()
typeInfo.SetType("expr_1", DataTypeInteger)
typeInfo.AddCoercion("expr_2", DataTypeInteger, DataTypeReal, "arithmetic with REAL")
```

**Memory Size**: O(e) where e = number of expressions

---

### 7. DataType Enumeration

Fundamental data types supported by the database.

```go
type DataType int

const (
    DataTypeUnknown DataType = iota
    
    // Numeric types
    DataTypeInteger    // INT, INTEGER
    DataTypeReal       // REAL, FLOAT, DOUBLE
    DataTypeNumeric    // NUMERIC, DECIMAL
    
    // String types
    DataTypeText       // TEXT, VARCHAR, CHAR
    DataTypeBlob       // BLOB, BINARY
    
    // Other types
    DataTypeBoolean    // BOOLEAN, BOOL
    DataTypeDate       // DATE
    DataTypeTime       // TIME
    DataTypeTimestamp  // TIMESTAMP, DATETIME
    DataTypeNull       // NULL type
)

func (dt DataType) String() string
func (dt DataType) IsNumeric() bool
func (dt DataType) IsString() bool
func (dt DataType) IsComparable(other DataType) bool
func (dt DataType) CanCoerceTo(other DataType) bool
```

**Type Compatibility Matrix**:
```
           INT  REAL  TEXT  BOOL  DATE  NULL
INT        ✓    ✓     ✗     ✗     ✗     ✓
REAL       ✓    ✓     ✗     ✗     ✗     ✓
TEXT       ✗    ✗     ✓     ✗     ✗     ✓
BOOL       ✗    ✗     ✗     ✓     ✗     ✓
DATE       ✗    ✗     ✗     ✗     ✓     ✓
NULL       ✓    ✓     ✓     ✓     ✓     ✓
```

---

### 8. CompilationError

Detailed error information from compilation failures.

```go
type CompilationError struct {
    // Error classification
    Code     ErrorCode
    Category ErrorCategory
    
    // Error details
    Message  string
    Hint     string    // Suggestion for fixing
    
    // Location in source SQL
    Line     int
    Column   int
    Position int
    
    // Context
    QueryPart string   // The part of query that caused error
    
    // Stack trace for debugging
    StackTrace string
}

type ErrorCode int
type ErrorCategory int

const (
    // Error categories
    ErrorCategoryNameResolution ErrorCategory = iota
    ErrorCategoryTypeChecking
    ErrorCategoryConstraintValidation
    ErrorCategorySemanticAnalysis
)

const (
    // Error codes
    ErrTableNotFound ErrorCode = 1001
    ErrColumnNotFound ErrorCode = 1002
    ErrAmbiguousColumn ErrorCode = 1003
    ErrTypeMismatch ErrorCode = 2001
    ErrInvalidOperand ErrorCode = 2002
    ErrNotNullViolation ErrorCode = 3001
    ErrPrimaryKeyDuplicate ErrorCode = 3002
    ErrInvalidAggregate ErrorCode = 4001
    // ... more error codes
)

func (ce *CompilationError) Error() string
func (ce *CompilationError) WithHint(hint string) *CompilationError
```

**Example**:
```go
err := &CompilationError{
    Code: ErrColumnNotFound,
    Category: ErrorCategoryNameResolution,
    Message: "Column 'namee' does not exist",
    Hint: "Did you mean 'name'?",
    Line: 1,
    Column: 8,
    QueryPart: "SELECT namee FROM users",
}
```

**Memory Size**: ~200 bytes per error

---

### 9. Constraint Data Structures

#### PrimaryKeyConstraint
```go
type PrimaryKeyConstraint struct {
    Name       string
    Columns    []string  // Column names in PK
    TableName  string
}

func (pk *PrimaryKeyConstraint) IsSingleColumn() bool
func (pk *PrimaryKeyConstraint) Contains(columnName string) bool
```

#### ForeignKeyConstraint
```go
type ForeignKeyConstraint struct {
    Name              string
    SourceTable       string
    SourceColumns     []string
    ReferencedTable   string
    ReferencedColumns []string
    OnDelete          ReferentialAction  // CASCADE, SET NULL, RESTRICT
    OnUpdate          ReferentialAction
}

type ReferentialAction int

const (
    ActionNoAction ReferentialAction = iota
    ActionRestrict
    ActionCascade
    ActionSetNull
    ActionSetDefault
)
```

#### UniqueConstraint
```go
type UniqueConstraint struct {
    Name       string
    Columns    []string
    TableName  string
}
```

#### CheckConstraint
```go
type CheckConstraint struct {
    Name       string
    Expression string  // CHECK expression as string
    TableName  string
}
```

---

### 10. FunctionMetadata

Information about SQL functions (aggregate and scalar).

```go
type FunctionMetadata struct {
    Name         string
    FunctionType FunctionType
    ReturnType   DataType
    ArgumentTypes []DataType
    IsAggregate  bool
    IsBuiltin    bool
}

type FunctionType int

const (
    FunctionTypeScalar FunctionType = iota  // Returns single value
    FunctionTypeAggregate                    // Returns aggregate (COUNT, SUM, etc.)
    FunctionTypeWindow                       // Window function
    FunctionTypeTable                        // Table-valued function
)

// Built-in functions
var BuiltinFunctions = map[string]*FunctionMetadata{
    "COUNT":  {Name: "COUNT", IsAggregate: true, ReturnType: DataTypeInteger},
    "SUM":    {Name: "SUM", IsAggregate: true, ReturnType: DataTypeNumeric},
    "AVG":    {Name: "AVG", IsAggregate: true, ReturnType: DataTypeReal},
    "MAX":    {Name: "MAX", IsAggregate: true, ReturnType: DataTypeUnknown},
    "MIN":    {Name: "MIN", IsAggregate: true, ReturnType: DataTypeUnknown},
    "UPPER":  {Name: "UPPER", IsAggregate: false, ReturnType: DataTypeText},
    "LOWER":  {Name: "LOWER", IsAggregate: false, ReturnType: DataTypeText},
    "LENGTH": {Name: "LENGTH", IsAggregate: false, ReturnType: DataTypeInteger},
}

func (fm *FunctionMetadata) ValidateArguments(args []DataType) error
```

---

## Helper Data Structures

### 11. NameResolver

Internal structure for resolving names during compilation.

```go
type NameResolver struct {
    catalog       CatalogManager
    currentQuery  *CompiledQuery
    resolvedRefs  *ResolvedReferences
    scopeStack    []Scope  // For nested subqueries
}

type Scope struct {
    Tables  map[string]*TableMetadata
    Aliases map[string]string
    Parent  *Scope  // Parent scope for nested queries
}

func NewNameResolver(catalog CatalogManager) *NameResolver
func (nr *NameResolver) PushScope()
func (nr *NameResolver) PopScope()
func (nr *NameResolver) ResolveTable(name string) (*TableMetadata, error)
func (nr *NameResolver) ResolveColumn(name, tableContext string) (*ColumnMetadata, error)
```

---

### 12. TypeChecker

Internal structure for type checking and inference.

```go
type TypeChecker struct {
    typeInfo     *TypeInformation
    resolvedRefs *ResolvedReferences
}

func NewTypeChecker(refs *ResolvedReferences) *TypeChecker
func (tc *TypeChecker) InferExpressionType(expr parser.Expression) (DataType, error)
func (tc *TypeChecker) CheckBinaryOperation(left, right DataType, op string) (DataType, error)
func (tc *TypeChecker) CheckAssignment(target, source DataType) error
func (tc *TypeChecker) CheckFunctionCall(funcName string, args []DataType) (DataType, error)
```

---

### 13. ConstraintValidator

Internal structure for constraint validation.

```go
type ConstraintValidator struct {
    catalog      CatalogManager
    resolvedRefs *ResolvedReferences
    typeInfo     *TypeInformation
}

func NewConstraintValidator(catalog CatalogManager, refs *ResolvedReferences, types *TypeInformation) *ConstraintValidator
func (cv *ConstraintValidator) ValidateNotNull(tableName, columnName string, value interface{}) error
func (cv *ConstraintValidator) ValidatePrimaryKey(constraint *PrimaryKeyConstraint) error
func (cv *ConstraintValidator) ValidateForeignKey(constraint *ForeignKeyConstraint) error
func (cv *ConstraintValidator) ValidateCheck(constraint *CheckConstraint) error
```

---

## Data Structure Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                      CompiledQuery                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Statement (AST)                                        │ │
│  │ QueryType                                              │ │
│  │ ResolvedReferences ──────────┐                        │ │
│  │ TypeInformation ─────────────┼───┐                    │ │
│  │ Errors                       │   │                    │ │
│  └──────────────────────────────┼───┼────────────────────┘ │
└─────────────────────────────────┼───┼──────────────────────┘
                                  │   │
                  ┌───────────────┘   └───────────────┐
                  ▼                                    ▼
    ┌─────────────────────────┐         ┌─────────────────────────┐
    │  ResolvedReferences     │         │   TypeInformation       │
    ├─────────────────────────┤         ├─────────────────────────┤
    │ Tables ──────────┐      │         │ ExpressionTypes         │
    │ Columns          │      │         │ Coercions               │
    │ Aliases          │      │         │ Nullability             │
    └──────────────────┼──────┘         └─────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────┐
        │    TableMetadata         │
        ├──────────────────────────┤
        │ Columns ─────────┐       │
        │ PrimaryKey       │       │
        │ ForeignKeys      │       │
        │ Statistics       │       │
        └──────────────────┼───────┘
                           ▼
            ┌──────────────────────────┐
            │   ColumnMetadata         │
            ├──────────────────────────┤
            │ DataType                 │
            │ Constraints              │
            │ Statistics               │
            └──────────────────────────┘
```

---

## Memory Usage Analysis

### Typical Query Compilation Memory Profile

```
Small Query (SELECT * FROM users WHERE id = 1):
- CompiledQuery:        ~300 bytes
- ResolvedReferences:   ~200 bytes (1 table, 5 columns)
- TypeInformation:      ~150 bytes (3 expressions)
- TableMetadata:        ~500 bytes (5 columns)
- Total:                ~1,150 bytes

Medium Query (JOIN with WHERE and GROUP BY):
- CompiledQuery:        ~500 bytes
- ResolvedReferences:   ~800 bytes (3 tables, 20 columns)
- TypeInformation:      ~600 bytes (15 expressions)
- TableMetadata:        ~1,500 bytes (3 tables × ~500 bytes)
- Total:                ~3,400 bytes

Large Query (Multiple JOINs, subqueries):
- CompiledQuery:        ~1,000 bytes
- ResolvedReferences:   ~3,000 bytes (10 tables, 80 columns)
- TypeInformation:      ~2,000 bytes (50 expressions)
- TableMetadata:        ~5,000 bytes (10 tables)
- Total:                ~11,000 bytes (~11 KB)
```

**Conclusion**: Even complex queries have reasonable memory footprint (<100 KB).

---

## Thread Safety Considerations

### Immutable Structures
- `TableMetadata`: Immutable after creation (read-only from catalog)
- `ColumnMetadata`: Immutable after creation
- `DataType`: Enum (inherently thread-safe)

### Mutable Structures (Require Synchronization)
- `CompiledQuery`: Built during compilation, immutable after
- `ResolvedReferences`: Modified during resolution phase only
- `TypeInformation`: Modified during type checking phase only

**Thread Safety Strategy**: Each compilation creates new instances, no shared mutable state.

---

*This document describes all major data structures in the Query Compiler module of NamyohDB.*
