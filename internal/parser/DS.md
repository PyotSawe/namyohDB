# Parser Module Data Structures

## Overview
This document describes the data structures used in the SQL parser module, including AST node representations, parser state management, and error handling structures.

## Core Data Structures

### 1. Parser State Structure

```go
type Parser struct {
    lexer     *lexer.Lexer      // Token source
    current   lexer.Token       // Current token being examined
    previous  lexer.Token       // Previously consumed token
    errors    []ParseError      // Accumulated parse errors
    panicMode bool              // Error recovery state
    context   []string          // Current parsing context stack
}
```

#### Design Rationale
- **Single Token Lookahead**: Efficient LL(1) parsing with minimal buffering
- **Error Accumulation**: Collect multiple errors for comprehensive reporting
- **Context Tracking**: Maintain parsing context for better error messages
- **Panic Mode**: Enable error recovery and continuation

#### Memory Layout
```
┌─────────────────────────────────────────────────────────┐
│                    Parser (72 bytes)                    │
├─────────────────────────────────────────────────────────┤
│ lexer       │ 8 bytes  │ pointer to Lexer             │
│ current     │ 48 bytes │ Token struct                  │
│ previous    │ 48 bytes │ Token struct                  │
│ errors      │ 24 bytes │ slice header                  │
│ panicMode   │ 1 byte   │ boolean                       │
│ context     │ 24 bytes │ slice header                  │
│ padding     │ 7 bytes  │ struct alignment              │
└─────────────────────────────────────────────────────────┘
```

### 2. Abstract Syntax Tree Node Hierarchy

#### Base Node Interface

```go
type Node interface {
    String() string               // Human-readable representation
    Accept(visitor Visitor) error // Visitor pattern support
    Position() token.Position     // Source code position
    Type() NodeType              // Node type classification
    Children() []Node            // Child nodes for traversal
}

type BaseNode struct {
    Pos   token.Position    // Source position information
    Kind  NodeType         // Node type identifier
}
```

#### Node Type Classification

```go
type NodeType int

const (
    // Statement nodes
    SELECT_STMT NodeType = iota
    INSERT_STMT
    UPDATE_STMT
    DELETE_STMT
    CREATE_TABLE_STMT
    CREATE_INDEX_STMT
    DROP_TABLE_STMT
    DROP_INDEX_STMT
    ALTER_TABLE_STMT
    
    // Expression nodes
    LITERAL_EXPR
    IDENTIFIER_EXPR
    BINARY_EXPR
    UNARY_EXPR
    FUNCTION_CALL_EXPR
    CASE_EXPR
    SUBQUERY_EXPR
    
    // Clause nodes
    WHERE_CLAUSE
    ORDER_CLAUSE
    GROUP_CLAUSE
    HAVING_CLAUSE
    LIMIT_CLAUSE
)
```

### 3. Statement Node Structures

#### SELECT Statement

```go
type SelectStatement struct {
    BaseNode
    Distinct   bool                // DISTINCT keyword present
    SelectList []SelectItem        // Column expressions
    From       TableExpression     // Table sources
    Where      Expression          // Filter conditions (nullable)
    GroupBy    *GroupByClause      // Grouping specification (nullable)
    Having     Expression          // Group filter conditions (nullable)
    OrderBy    *OrderByClause      // Sort specification (nullable)
    Limit      *LimitClause        // Result limitation (nullable)
}

type SelectItem struct {
    Expression Expression    // Column expression
    Alias      *Identifier   // Optional column alias
}
```

#### INSERT Statement

```go
type InsertStatement struct {
    BaseNode
    Table   *Identifier      // Target table
    Columns []Identifier     // Column list (optional)
    Values  [][]Expression   // Values to insert (multiple rows)
    Select  *SelectStatement // INSERT ... SELECT form (nullable)
}
```

#### UPDATE Statement

```go
type UpdateStatement struct {
    BaseNode
    Table       *Identifier         // Target table
    Assignments []Assignment        // SET clause assignments
    Where       Expression          // Filter conditions (nullable)
}

type Assignment struct {
    Column Identifier    // Column to update
    Value  Expression    // New value expression
}
```

#### DELETE Statement

```go
type DeleteStatement struct {
    BaseNode
    Table *Identifier    // Target table
    Where Expression     // Filter conditions (nullable)
}
```

### 4. Expression Node Structures

#### Binary Expression

```go
type BinaryExpression struct {
    BaseNode
    Left     Expression      // Left operand
    Operator BinaryOperator  // Operation type
    Right    Expression      // Right operand
}

type BinaryOperator int

const (
    // Arithmetic operators
    ADD BinaryOperator = iota  // +
    SUB                        // -
    MUL                        // *
    DIV                        // /
    MOD                        // %
    
    // Comparison operators
    EQ                         // =
    NE                         // <> or !=
    LT                         // <
    LE                         // <=
    GT                         // >
    GE                         // >=
    
    // Logical operators
    AND                        // AND
    OR                         // OR
    
    // String operators
    CONCAT                     // ||
    LIKE                       // LIKE
)
```

#### Unary Expression

```go
type UnaryExpression struct {
    BaseNode
    Operator UnaryOperator   // Operation type
    Operand  Expression      // Target expression
}

type UnaryOperator int

const (
    NOT UnaryOperator = iota  // NOT
    NEGATE                    // - (unary minus)
    PLUS                      // + (unary plus)
)
```

#### Literal Expressions

```go
type LiteralExpression struct {
    BaseNode
    Value interface{}    // The literal value
    Raw   string        // Original string representation
}

// Specific literal types
type StringLiteral struct {
    LiteralExpression
    Value string        // String content (without quotes)
}

type NumericLiteral struct {
    LiteralExpression
    Value   interface{} // int64, float64, or big numbers
    IsFloat bool        // Distinguishes integer from float
}

type BooleanLiteral struct {
    LiteralExpression
    Value bool          // true or false
}

type NullLiteral struct {
    LiteralExpression
    // No additional fields - represents SQL NULL
}
```

#### Function Call Expression

```go
type FunctionCallExpression struct {
    BaseNode
    Name      Identifier      // Function name
    Arguments []Expression    // Function arguments
    Distinct  bool           // DISTINCT keyword in aggregates
}
```

#### Case Expression

```go
type CaseExpression struct {
    BaseNode
    Expression Expression      // CASE expression (nullable for simple CASE)
    WhenClauses []WhenClause   // WHEN ... THEN clauses
    ElseClause  Expression     // ELSE clause (nullable)
}

type WhenClause struct {
    Condition Expression       // WHEN condition
    Result    Expression       // THEN result
}
```

### 5. Table Expression Structures

#### Table Expression Hierarchy

```go
type TableExpression interface {
    Node
    tableExpression() // Marker method
}

type TableName struct {
    BaseNode
    Name  Identifier      // Table name
    Alias *Identifier     // Table alias (nullable)
}

type SubqueryTable struct {
    BaseNode
    Query *SelectStatement // Subquery
    Alias Identifier       // Required alias for subqueries
}

type JoinExpression struct {
    BaseNode
    Left      TableExpression // Left table
    JoinType  JoinType        // Type of join
    Right     TableExpression // Right table
    Condition JoinCondition   // ON or USING condition
}
```

#### Join Types and Conditions

```go
type JoinType int

const (
    INNER_JOIN JoinType = iota
    LEFT_JOIN
    RIGHT_JOIN
    FULL_OUTER_JOIN
    CROSS_JOIN
)

type JoinCondition interface {
    joinCondition() // Marker method
}

type OnCondition struct {
    Expression Expression    // Boolean expression for ON clause
}

type UsingCondition struct {
    Columns []Identifier     // Column list for USING clause
}
```

### 6. Clause Structures

#### WHERE Clause

```go
type WhereClause struct {
    BaseNode
    Condition Expression     // Boolean filter expression
}
```

#### ORDER BY Clause

```go
type OrderByClause struct {
    BaseNode
    Items []OrderByItem     // Sort specifications
}

type OrderByItem struct {
    Expression Expression    // Expression to sort by
    Direction  SortDirection // ASC or DESC
}

type SortDirection int

const (
    ASC SortDirection = iota  // Ascending order
    DESC                      // Descending order
)
```

#### GROUP BY Clause

```go
type GroupByClause struct {
    BaseNode
    Expressions []Expression // Grouping expressions
}
```

#### LIMIT Clause

```go
type LimitClause struct {
    BaseNode
    Count  Expression      // Number of rows to return
    Offset Expression      // Number of rows to skip (nullable)
}
```

### 7. Error Handling Structures

#### Parse Error

```go
type ParseError struct {
    Token    lexer.Token        // Token where error occurred
    Message  string             // Error description
    Expected []lexer.TokenType  // Expected token types
    Context  []string           // Parsing context stack
    Position token.Position     // Precise error location
}

func (e ParseError) Error() string {
    return fmt.Sprintf("Parse error at line %d, column %d: %s", 
                      e.Position.Line, e.Position.Column, e.Message)
}
```

#### Error Recovery Information

```go
type RecoveryInfo struct {
    ErrorCount      int              // Number of errors encountered
    RecoveryPoints  []token.Position // Where recovery occurred
    SkippedTokens   []lexer.Token    // Tokens skipped during recovery
    SyncTokens      []lexer.TokenType // Tokens that trigger synchronization
}
```

### 8. Visitor Pattern Support

#### Visitor Interface

```go
type Visitor interface {
    // Statement visitors
    VisitSelectStatement(stmt *SelectStatement) error
    VisitInsertStatement(stmt *InsertStatement) error
    VisitUpdateStatement(stmt *UpdateStatement) error
    VisitDeleteStatement(stmt *DeleteStatement) error
    
    // Expression visitors
    VisitBinaryExpression(expr *BinaryExpression) error
    VisitUnaryExpression(expr *UnaryExpression) error
    VisitLiteralExpression(expr *LiteralExpression) error
    VisitIdentifierExpression(expr *IdentifierExpression) error
    VisitFunctionCallExpression(expr *FunctionCallExpression) error
    
    // Additional visitors for other node types...
}
```

#### Visitor Implementation Example

```go
type ASTDumper struct {
    output strings.Builder
    indent int
}

func (v *ASTDumper) VisitSelectStatement(stmt *SelectStatement) error {
    v.writeLine("SELECT Statement")
    v.indent++
    
    // Visit select list
    for _, item := range stmt.SelectList {
        item.Expression.Accept(v)
    }
    
    // Visit other clauses
    if stmt.From != nil {
        stmt.From.Accept(v)
    }
    if stmt.Where != nil {
        stmt.Where.Accept(v)
    }
    
    v.indent--
    return nil
}
```

## Advanced Data Structures

### 9. Symbol Table Integration

```go
type SymbolInfo struct {
    Name     string          // Symbol name
    Type     DataType        // Symbol data type
    Scope    ScopeLevel      // Visibility scope
    Position token.Position  // Definition location
}

type SymbolTable struct {
    symbols map[string]*SymbolInfo  // Symbol definitions
    parent  *SymbolTable            // Parent scope (nullable)
    level   ScopeLevel              // Current scope level
}
```

### 10. Type System Support

```go
type DataType interface {
    String() string
    Compatible(other DataType) bool
    Size() int64
}

type BasicType struct {
    Kind     TypeKind    // INTEGER, VARCHAR, etc.
    Size     int64       // Type size in bytes
    Nullable bool        // Can contain NULL values
}

type TypeKind int

const (
    INTEGER_TYPE TypeKind = iota
    VARCHAR_TYPE
    DECIMAL_TYPE
    BOOLEAN_TYPE
    DATE_TYPE
    TIME_TYPE
    TIMESTAMP_TYPE
)
```

### 11. Optimization Hints

```go
type OptimizationHint struct {
    HintType HintKind       // Type of hint
    Target   string         // Target object (table, index, etc.)
    Value    interface{}    // Hint-specific value
}

type HintKind int

const (
    INDEX_HINT HintKind = iota
    JOIN_ORDER_HINT
    PARALLEL_HINT
    CACHE_HINT
)
```

## Memory Management Strategies

### 1. Node Pooling

```go
type NodePool struct {
    selectStmts   sync.Pool    // Pool for SELECT statements
    insertStmts   sync.Pool    // Pool for INSERT statements
    expressions   sync.Pool    // Pool for expressions
    identifiers   sync.Pool    // Pool for identifiers
}

func (p *NodePool) GetSelectStatement() *SelectStatement {
    if stmt := p.selectStmts.Get(); stmt != nil {
        return stmt.(*SelectStatement)
    }
    return &SelectStatement{}
}

func (p *NodePool) PutSelectStatement(stmt *SelectStatement) {
    // Reset statement state
    *stmt = SelectStatement{}
    p.selectStmts.Put(stmt)
}
```

### 2. String Interning

```go
type StringInterner struct {
    strings map[string]string    // Interned strings
    mutex   sync.RWMutex        // Thread-safe access
}

func (si *StringInterner) Intern(s string) string {
    si.mutex.RLock()
    if interned, exists := si.strings[s]; exists {
        si.mutex.RUnlock()
        return interned
    }
    si.mutex.RUnlock()
    
    si.mutex.Lock()
    si.strings[s] = s
    si.mutex.Unlock()
    return s
}
```

### 3. Memory Usage Analysis

#### Node Size Analysis

| Node Type | Size (bytes) | Common Count | Memory Impact |
|-----------|--------------|--------------|---------------|
| SelectStatement | 120 | 1 per query | Low |
| BinaryExpression | 48 | High | Medium |
| Identifier | 32 | Very High | High |
| LiteralExpression | 40 | High | Medium |
| Function Call | 56 | Medium | Low |

#### Memory Optimization Strategies

1. **Flyweight Pattern**: Share identical literal values
2. **Compact Representations**: Use smaller integer types where possible
3. **Lazy Initialization**: Create optional fields only when needed
4. **Memory Pooling**: Reuse node instances across parsing sessions

## Performance Characteristics

### 1. Space Complexity

| Structure | Space Complexity | Notes |
|-----------|------------------|-------|
| AST Tree | O(n) | n = number of tokens |
| Symbol Table | O(s) | s = number of symbols |
| Error List | O(e) | e = number of errors |
| Parser State | O(1) | Fixed size |
| Node Pool | O(p) | p = pool capacity |

### 2. Access Time Complexity

| Operation | Time Complexity | Notes |
|-----------|-----------------|-------|
| Node Creation | O(1) | With pooling |
| Tree Traversal | O(n) | Visit each node once |
| Symbol Lookup | O(1) | Hash table access |
| Error Recording | O(1) | Append to slice |
| Node Pool Access | O(1) | Sync.Pool operations |

### 3. Memory Access Patterns

- **Sequential Access**: During parsing and tree construction
- **Random Access**: Symbol table lookups and cross-references
- **Write-Once**: Most AST nodes are immutable after creation
- **Read-Heavy**: Multiple passes over AST for analysis and optimization

## Thread Safety Considerations

### 1. Immutable Structures
- **AST Nodes**: Immutable after construction (thread-safe for reading)
- **Tokens**: Immutable token values from lexer
- **Error Records**: Immutable error information

### 2. Mutable Structures
- **Parser State**: Single-threaded use only
- **Symbol Tables**: Require synchronization for concurrent access
- **Node Pools**: Thread-safe with sync.Pool

### 3. Concurrent Usage Patterns

```go
// Safe: Multiple readers of same AST
func analyzeAST(ast *SelectStatement) {
    // Read-only operations are thread-safe
    ast.Accept(analyzer)
}

// Unsafe: Concurrent modification of parser
func parseQueries(queries []string) {
    // Each goroutine needs its own parser instance
    for _, query := range queries {
        parser := NewParser(query)
        ast, err := parser.Parse()
        // Process ast...
    }
}
```

## Testing Data Structures

### 1. Test Case Representation

```go
type ParseTestCase struct {
    Name     string        // Test case name
    Input    string        // SQL input to parse
    Expected Node          // Expected AST structure
    Error    bool          // Whether error is expected
}
```

### 2. AST Comparison Utilities

```go
type ASTComparator struct {
    ignorePositions bool    // Ignore source positions in comparison
    ignoreWhitespace bool   // Ignore whitespace differences
}

func (c *ASTComparator) Equal(a, b Node) bool {
    // Deep comparison of AST structures
    return c.compareNodes(a, b)
}
```

## Future Enhancements

### 1. Streaming AST Construction
- Build AST incrementally during parsing
- Support for very large SQL statements
- Memory-bounded parsing with spilling

### 2. Persistent AST Structures
- Immutable data structures with structural sharing
- Efficient AST transformation with minimal copying
- Version control for AST modifications

### 3. Typed AST Variants
- Strongly typed AST with compile-time guarantees
- Generic node types with type parameters
- Enhanced type safety for transformations