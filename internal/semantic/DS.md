# Semantic Analyzer Data Structures

## Core Data Structures

### 1. SemanticInfo
**Purpose**: Stores complete semantic analysis results

```go
type SemanticInfo struct {
    // Source query
    CompiledQuery *compiler.CompiledQuery
    
    // Validation results
    SchemaValid    bool
    SemanticsValid bool
    AccessValid    bool
    
    // Aggregate metadata
    HasAggregates  bool
    AggregateInfo  *AggregateMetadata
    
    // Subquery metadata
    HasSubqueries  bool
    SubqueryInfo   []*SubqueryMetadata
    
    // GROUP BY metadata
    HasGroupBy     bool
    GroupByInfo    *GroupByMetadata
    
    // Semantic errors
    Errors         []SemanticError
    Warnings       []SemanticWarning
    
    // Analysis timestamp
    AnalyzedAt     time.Time
}
```

**Memory**: ~500 bytes base + variable size for errors/metadata
**Lifecycle**: Created once per query, passed to optimizer

---

### 2. AggregateMetadata
**Purpose**: Tracks aggregate function usage and validation

```go
type AggregateMetadata struct {
    // Aggregate functions found in query
    Functions      []*AggregateFunctionInfo
    
    // Positions where aggregates appear
    InSelect       bool
    InHaving       bool
    InOrderBy      bool
    
    // GROUP BY related
    RequiresGroupBy bool
    GroupByColumns  []string
    
    // Validation state
    Validated       bool
    ValidationError error
}

type AggregateFunctionInfo struct {
    Name           string           // COUNT, SUM, AVG, MAX, MIN
    Arguments      []Expression     // Function arguments
    ReturnType     compiler.DataType
    IsDistinct     bool             // COUNT(DISTINCT ...)
    Position       ExpressionLocation
}

type ExpressionLocation int

const (
    LocationUnknown ExpressionLocation = iota
    LocationSelect
    LocationWhere
    LocationGroupBy
    LocationHaving
    LocationOrderBy
)
```

**Memory**: 
- AggregateMetadata: ~200 bytes
- AggregateFunctionInfo: ~100 bytes each
- Total for query with 3 aggregates: ~500 bytes

**Usage**:
```
Query: SELECT COUNT(*), AVG(age) FROM users GROUP BY dept

AggregateMetadata {
    Functions: [
        {Name: "COUNT", Arguments: [Wildcard], ReturnType: Integer},
        {Name: "AVG", Arguments: [Column("age")], ReturnType: Real}
    ],
    InSelect: true,
    RequiresGroupBy: true,
    GroupByColumns: ["dept"]
}
```

---

### 3. GroupByMetadata
**Purpose**: Stores GROUP BY validation information

```go
type GroupByMetadata struct {
    // GROUP BY expressions
    Expressions    []Expression
    
    // Columns referenced in GROUP BY
    Columns        []string
    
    // Non-grouped columns in SELECT
    NonGroupedCols []string
    
    // Aggregate functions used
    Aggregates     []string
    
    // HAVING clause info
    HasHaving      bool
    HavingExpr     Expression
    
    // Validation results
    Valid          bool
    Errors         []string
}
```

**Memory**: ~300 bytes + size of expressions

**Example**:
```sql
SELECT dept, COUNT(*), AVG(salary)
FROM employees
GROUP BY dept
HAVING AVG(salary) > 50000

GroupByMetadata {
    Expressions: [Identifier("dept")],
    Columns: ["dept"],
    Aggregates: ["COUNT", "AVG"],
    HasHaving: true,
    Valid: true
}
```

---

### 4. SubqueryMetadata
**Purpose**: Tracks subquery usage and validation

```go
type SubqueryMetadata struct {
    // Subquery statement
    Query          *parser.SelectStatement
    
    // Subquery type
    Type           SubqueryType
    
    // Location in parent query
    Location       ExpressionLocation
    
    // Correlation info
    IsCorrelated   bool
    CorrelatedRefs []CorrelatedReference
    
    // Result info
    ColumnCount    int
    ExpectedRows   RowCountExpectation
    
    // Validation
    Valid          bool
    Error          error
}

type SubqueryType int

const (
    SubqueryScalar        SubqueryType = iota  // (SELECT x FROM t)
    SubqueryInPredicate                        // col IN (SELECT x FROM t)
    SubqueryExists                             // EXISTS (SELECT ...)
    SubqueryDerivedTable                       // FROM (SELECT ...) AS t
)

type RowCountExpectation int

const (
    RowCountAny        RowCountExpectation = iota  // Any number of rows
    RowCountSingle                                  // Must be exactly 1 row
    RowCountZeroOrOne                               // 0 or 1 rows
)

type CorrelatedReference struct {
    Column        string
    Table         string
    OuterScope    int    // Nesting level
}
```

**Memory**: ~400 bytes per subquery

**Example**:
```sql
SELECT name, (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id)
FROM users

SubqueryMetadata {
    Type: SubqueryScalar,
    Location: LocationSelect,
    IsCorrelated: true,
    CorrelatedRefs: [{Column: "id", Table: "users", OuterScope: 1}],
    ColumnCount: 1,
    ExpectedRows: RowCountSingle
}
```

---

### 5. SemanticError
**Purpose**: Detailed semantic error reporting

```go
type SemanticError struct {
    Code           ErrorCode
    Category       ErrorCategory
    Message        string
    Hint           string
    Location       ErrorLocation
    Expression     string  // The problematic expression
    Suggestion     string  // How to fix it
}

type ErrorCode int

const (
    // GROUP BY errors (5000-5099)
    ErrGroupByRequired          ErrorCode = 5001
    ErrColumnNotInGroupBy       ErrorCode = 5002
    ErrAggregateInGroupBy       ErrorCode = 5003
    ErrHavingWithoutGroupBy     ErrorCode = 5004
    
    // Aggregate errors (5100-5199)
    ErrNestedAggregate          ErrorCode = 5101
    ErrAggregateInWhere         ErrorCode = 5102
    ErrInvalidAggregateArg      ErrorCode = 5103
    ErrWrongAggregateArgCount   ErrorCode = 5104
    
    // Subquery errors (5200-5299)
    ErrScalarSubqueryMultiCol   ErrorCode = 5201
    ErrInSubqueryMultiCol       ErrorCode = 5202
    ErrDerivedTableNoAlias      ErrorCode = 5203
    ErrCorrelatedRefNotVisible  ErrorCode = 5204
    
    // Schema errors (5300-5399)
    ErrTableAlreadyExists       ErrorCode = 5301
    ErrDuplicateColumn          ErrorCode = 5302
    ErrForeignKeyRefNotFound    ErrorCode = 5303
    ErrCircularDependency       ErrorCode = 5304
)

type ErrorCategory int

const (
    CategoryGroupBy    ErrorCategory = iota
    CategoryAggregate
    CategorySubquery
    CategorySchema
    CategoryConstraint
)

type ErrorLocation struct {
    Line   int
    Column int
    Offset int
}
```

**Memory**: ~300 bytes per error

**Example**:
```go
SemanticError {
    Code: ErrColumnNotInGroupBy,
    Category: CategoryGroupBy,
    Message: "Column 'salary' must appear in GROUP BY or be in aggregate function",
    Hint: "Add 'salary' to GROUP BY clause or wrap in aggregate like AVG(salary)",
    Expression: "salary",
    Suggestion: "GROUP BY dept, salary"
}
```

---

### 6. ValidationContext
**Purpose**: Tracks validation state during analysis

```go
type ValidationContext struct {
    // Current scope
    Scope           *ValidationScope
    
    // Reference tracking
    ResolvedRefs    *compiler.ResolvedReferences
    TypeInfo        *compiler.TypeInformation
    
    // State flags
    InWhereClause   bool
    InGroupByClause bool
    InHavingClause  bool
    InOrderByClause bool
    InSubquery      bool
    
    // Aggregate tracking
    AggregateDepth  int
    HasAggregates   bool
    
    // Subquery tracking
    SubqueryDepth   int
    OuterScopes     []*ValidationScope
    
    // Schema access
    Catalog         compiler.CatalogManager
}

type ValidationScope struct {
    // Tables visible in this scope
    Tables          map[string]*compiler.TableMetadata
    
    // Columns visible in this scope
    Columns         map[string]*compiler.ColumnMetadata
    
    // Aliases in this scope
    Aliases         map[string]string
    
    // Parent scope (for subqueries)
    Parent          *ValidationScope
}
```

**Memory**: ~1KB per scope level

**Usage Pattern**:
```go
// Enter subquery
ctx.SubqueryDepth++
ctx.OuterScopes = append(ctx.OuterScopes, ctx.Scope)
ctx.Scope = NewValidationScope(ctx.Scope)

// Exit subquery
ctx.Scope = ctx.Scope.Parent
ctx.OuterScopes = ctx.OuterScopes[:len(ctx.OuterScopes)-1]
ctx.SubqueryDepth--
```

---

### 7. SchemaValidator
**Purpose**: Validates DDL statements and schema operations

```go
type SchemaValidator struct {
    catalog         compiler.CatalogManager
    checkExistence  bool
    allowDuplicates bool
}

type SchemaValidationResult struct {
    Valid           bool
    TableName       string
    ColumnCount     int
    ConstraintCount int
    Errors          []SchemaError
    Warnings        []SchemaWarning
}

type SchemaError struct {
    Type           SchemaErrorType
    Message        string
    Object         string  // Table/column name
    Constraint     string  // Constraint name if applicable
}

type SchemaErrorType int

const (
    SchemaErrTableExists      SchemaErrorType = iota
    SchemaErrTableNotFound
    SchemaErrDuplicateColumn
    SchemaErrInvalidForeignKey
    SchemaErrCircularReference
    SchemaErrInvalidConstraint
)
```

**Memory**: ~500 bytes + errors

---

## Memory Layout Examples

### Example 1: Simple Aggregate Query
```sql
SELECT dept, COUNT(*) FROM employees GROUP BY dept
```

```
SemanticInfo (500 bytes)
├── AggregateMetadata (200 bytes)
│   └── AggregateFunctionInfo (100 bytes) × 1
├── GroupByMetadata (300 bytes)
└── Errors (0 bytes)

Total: ~1 KB
```

### Example 2: Complex Subquery
```sql
SELECT name, 
       (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id) as order_count
FROM users
WHERE age > 18
```

```
SemanticInfo (500 bytes)
├── AggregateMetadata (200 bytes)
│   └── AggregateFunctionInfo (100 bytes) × 1
├── SubqueryMetadata (400 bytes)
│   └── CorrelatedReference (50 bytes) × 1
└── ValidationContext (1000 bytes)

Total: ~2 KB
```

### Example 3: Error Case
```sql
SELECT name, salary FROM employees GROUP BY name
-- Error: 'salary' not in GROUP BY
```

```
SemanticInfo (500 bytes)
├── GroupByMetadata (300 bytes)
├── Errors (300 bytes) × 1
│   └── SemanticError {
│       Code: 5002,
│       Message: "Column 'salary' must appear in GROUP BY...",
│       Hint: "Add 'salary' to GROUP BY or use aggregate"
│   }
└── SemanticsValid: false

Total: ~1.1 KB
```

---

## Relationships

### Data Flow
```
CompiledQuery → SemanticAnalyzer → SemanticInfo
     ↓                                    ↓
ResolvedRefs                        AggregateMetadata
TypeInfo                            SubqueryMetadata
                                    GroupByMetadata
                                    Errors/Warnings
```

### Dependency Graph
```
SemanticInfo
├── depends on → CompiledQuery
│   ├── ResolvedReferences
│   └── TypeInformation
├── produces → AggregateMetadata
├── produces → SubqueryMetadata
├── produces → GroupByMetadata
└── produces → SemanticError[]
```

---

## Performance Characteristics

### Memory Usage by Query Type

| Query Type | Base | Aggregates | Subqueries | Errors | Total |
|------------|------|------------|------------|--------|-------|
| Simple SELECT | 500B | 0 | 0 | 0 | 500B |
| GROUP BY | 500B | 300B | 0 | 0 | 800B |
| With Aggregate | 500B | 500B | 0 | 0 | 1KB |
| With Subquery | 500B | 0 | 400B | 0 | 900B |
| Complex (all) | 500B | 500B | 800B | 300B | 2.1KB |

### Comparison with Compiler

| Structure | Compiler | Semantic | Total |
|-----------|----------|----------|-------|
| Simple query | 1KB | 0.5KB | 1.5KB |
| Moderate complexity | 5KB | 1KB | 6KB |
| High complexity | 11KB | 2KB | 13KB |

**Note**: Semantic analysis adds 10-20% memory overhead on top of compilation

---

## Design Principles

### 1. Separation of Concerns
- Semantic validation is separate from compilation
- Each validator has single responsibility
- Clear interfaces between components

### 2. Composability
- Validators can be combined
- Context passes through validation chain
- Errors accumulate for better reporting

### 3. Extensibility
- Easy to add new validation rules
- Plugin architecture for custom semantics
- Configurable strictness levels

### 4. Performance
- Lazy evaluation where possible
- Early exit on critical errors
- Minimal memory allocation

---

## Implementation Notes

### Thread Safety
- SemanticInfo is immutable after creation
- ValidationContext is not thread-safe (per-query)
- Catalog access must be synchronized

### Error Handling
- Collect all errors, don't fail fast
- Provide actionable error messages
- Include fix suggestions when possible

### Testing Strategy
- Unit test each validator independently
- Integration test validation pipeline
- Performance test with large schemas
