# Parser Module Architecture

## Overview
The Parser module is responsible for syntactic analysis of SQL statements, converting a stream of tokens from the lexer into an Abstract Syntax Tree (AST) that represents the logical structure of SQL queries.

## Architecture Design

### Core Components

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Parser Module                                  │
├─────────────────────────────────────────────────────────────────────────┤
│  Input: Token Stream                                                    │
│  Output: Abstract Syntax Tree (AST)                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌────────────────┐    ┌──────────────────────────┐ │
│  │   Token     │───▶│   Recursive    │───▶│    AST Node Factory      │ │
│  │   Stream    │    │   Descent      │    │                          │ │
│  │   Reader    │    │   Parser       │    │                          │ │
│  └─────────────┘    └────────────────┘    └──────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                  Grammar Production Rules                         │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │ │
│  │  │   Query     │ │ Expression  │ │        Statement            │ │ │
│  │  │  Parsing    │ │  Parsing    │ │        Parsing              │ │ │
│  │  │ (SELECT,    │ │(Arithmetic, │ │    (DDL, DML)               │ │ │
│  │  │  INSERT)    │ │ Logical,    │ │                             │ │ │
│  │  │             │ │Comparison)  │ │                             │ │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                    Error Handling                                   │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │ │
│  │  │   Syntax    │ │   Error     │ │     Error Recovery          │ │ │
│  │  │   Error     │ │  Messages   │ │     & Continuation          │ │ │
│  │  │ Detection   │ │             │ │                             │ │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. **Recursive Descent Parser Design**
- **Decision**: Use recursive descent parsing with predictive lookahead
- **Rationale**: 
  - Natural mapping from grammar rules to code functions
  - Easy to understand and maintain
  - Efficient for LL(1) and LL(k) grammars
  - Excellent error reporting capabilities
- **Trade-offs**: Less compact than table-driven parsers, but more maintainable

#### 2. **Abstract Syntax Tree (AST) Representation**
- **Decision**: Build strongly-typed AST nodes during parsing
- **Rationale**:
  - Type safety for subsequent compilation phases
  - Clear semantic representation of SQL constructs
  - Enables sophisticated optimization and analysis
  - Facilitates code generation and transformation
- **Benefits**: Eliminates runtime type checking in later phases

#### 3. **Expression Precedence Handling**
- **Decision**: Implement operator precedence parsing for expressions
- **Rationale**:
  - Correct handling of mathematical and logical operator precedence
  - Efficient parsing of complex expressions
  - Matches SQL standard precedence rules
- **Implementation**: Precedence climbing algorithm

#### 4. **Error Recovery Strategy**
- **Decision**: Panic-mode recovery with synchronization points
- **Rationale**:
  - Continue parsing after syntax errors
  - Provide multiple error messages in single pass
  - Maintain parser state consistency
- **Synchronization Points**: Statement boundaries, major keywords

### Parser Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    High-Level API                       │
│              (ParseQuery, ParseStatement)               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                Statement Parsers                        │
│   (SELECT, INSERT, UPDATE, DELETE, CREATE, DROP)       │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                Expression Parsers                       │
│    (Arithmetic, Logical, Comparison, Function)         │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                 Token Management                        │
│           (Advance, Expect, Peek, Match)               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                   Lexer Interface                       │
│                (Token Stream Input)                     │
└─────────────────────────────────────────────────────────┘
```

## AST Node Architecture

### 1. **Base Node Structure**

```go
type Node interface {
    String() string           // Human-readable representation
    Accept(Visitor) error     // Visitor pattern support
    Position() token.Position // Source location
    Type() NodeType          // Node classification
}

type BaseNode struct {
    Pos  token.Position
    Kind NodeType
}
```

### 2. **Statement Node Hierarchy**

```
Statement (interface)
├── SelectStatement
├── InsertStatement  
├── UpdateStatement
├── DeleteStatement
├── CreateTableStatement
├── CreateIndexStatement
├── DropTableStatement
├── DropIndexStatement
└── AlterTableStatement
```

### 3. **Expression Node Hierarchy**

```
Expression (interface)
├── LiteralExpression
│   ├── StringLiteral
│   ├── NumericLiteral
│   ├── BooleanLiteral
│   └── NullLiteral
├── IdentifierExpression
├── BinaryExpression
│   ├── ArithmeticExpression (+, -, *, /, %)
│   ├── ComparisonExpression (=, <>, <, >, <=, >=)
│   └── LogicalExpression (AND, OR)
├── UnaryExpression
│   ├── NotExpression
│   └── NegationExpression
├── FunctionCallExpression
├── CaseExpression
└── SubqueryExpression
```

## Grammar Productions

### 1. **SQL Statement Grammar**

```
Query           → Statement EOF
Statement       → SelectStmt | InsertStmt | UpdateStmt | DeleteStmt | DDLStmt
SelectStmt      → SELECT SelectList FROM TableExpr WhereClause? OrderClause? LimitClause?
SelectList      → ASTERISK | SelectItem (, SelectItem)*
SelectItem      → Expression (AS Identifier)?
TableExpr       → TableName (AS Identifier)? | JoinExpr
WhereClause     → WHERE Expression
OrderClause     → ORDER BY OrderItem (, OrderItem)*
OrderItem       → Expression (ASC | DESC)?
LimitClause     → LIMIT NumericLiteral (OFFSET NumericLiteral)?
```

### 2. **Expression Grammar**

```
Expression      → LogicalOr
LogicalOr       → LogicalAnd (OR LogicalAnd)*
LogicalAnd      → Equality (AND Equality)*
Equality        → Comparison ((= | <> | !=) Comparison)*
Comparison      → Term ((< | <= | > | >=) Term)*
Term            → Factor ((+ | -) Factor)*
Factor          → Unary ((* | / | %) Unary)*
Unary           → (NOT | -)? Primary
Primary         → Literal | Identifier | FunctionCall | (Expression)
```

## Parser State Management

### 1. **Parser State Structure**

```go
type Parser struct {
    lexer    *lexer.Lexer     // Token source
    current  lexer.Token      // Current token
    previous lexer.Token      // Previous token
    errors   []ParseError     // Accumulated errors
    panic    bool             // Error recovery mode
}
```

### 2. **Token Management**

```go
// Core token operations
func (p *Parser) advance() lexer.Token
func (p *Parser) peek() lexer.Token
func (p *Parser) match(types ...lexer.TokenType) bool
func (p *Parser) expect(tokenType lexer.TokenType) (lexer.Token, error)
func (p *Parser) synchronize() // Error recovery
```

## Error Handling Architecture

### 1. **Error Types**

```go
type ParseError struct {
    Token    lexer.Token      // Error location
    Message  string           // Error description
    Expected []lexer.TokenType // Expected tokens
    Context  string           // Parsing context
}

type ErrorType int
const (
    SyntaxError ErrorType = iota
    UnexpectedToken
    MissingToken
    InvalidExpression
    UnsupportedFeature
)
```

### 2. **Error Recovery Strategy**

#### Panic Mode Recovery
```
1. When error detected:
   - Record error with context
   - Enter panic mode
   - Skip tokens until synchronization point

2. Synchronization points:
   - Statement boundaries (semicolon)
   - Major keywords (SELECT, FROM, WHERE)
   - Block delimiters

3. Resume normal parsing:
   - Exit panic mode
   - Continue with next statement
```

#### Error Message Generation
```go
func (p *Parser) errorAt(token lexer.Token, message string) {
    err := ParseError{
        Token:   token,
        Message: message,
        Context: p.currentContext(),
    }
    p.errors = append(p.errors, err)
}
```

## Performance Considerations

### 1. **Memory Management**
- **AST Node Pooling**: Reuse node objects to reduce GC pressure
- **String Interning**: Share common identifiers and keywords
- **Lazy Evaluation**: Defer expensive operations where possible

### 2. **Parsing Speed**
- **Predictive Parsing**: LL(1) grammar eliminates backtracking
- **Token Buffering**: Minimal lookahead reduces lexer calls
- **Direct AST Construction**: No intermediate representations

### 3. **Space Efficiency**
```go
// Compact node representation
type SelectStatement struct {
    BaseNode
    SelectList []Expression   // Only what's needed
    From       TableExpression // Embedded structs
    Where      Expression     // Nullable fields as pointers
    OrderBy    *OrderClause   // Optional clauses
    Limit      *LimitClause
}
```

## Integration Points

### 1. **Lexer Integration**
- **Token Stream**: Consumes tokens from lexer sequentially
- **Position Tracking**: Maintains source location through AST
- **Error Coordination**: Lexical errors propagate to parse errors

### 2. **Semantic Analysis Integration**
- **Symbol Resolution**: AST provides structure for symbol tables
- **Type Checking**: Strong typing enables static type analysis
- **Constraint Validation**: AST supports constraint checking

### 3. **Query Optimizer Integration**
- **Plan Generation**: AST provides logical query structure
- **Transformation**: AST enables query rewriting
- **Cost Estimation**: AST supports cost model calculations

## Threading and Concurrency

### 1. **Thread Safety**
- **Parser Instances**: Not thread-safe (single-threaded use)
- **AST Nodes**: Immutable after construction (thread-safe)
- **Shared Resources**: Read-only grammar and precedence tables

### 2. **Concurrent Parsing**
- **Multiple Parsers**: Independent parser instances for concurrent queries
- **Shared Grammar**: Static grammar rules shared across parsers
- **Memory Isolation**: Each parser maintains isolated state

## Extensibility Architecture

### 1. **Grammar Extensions**
- **New Statements**: Add new statement node types
- **New Expressions**: Extend expression hierarchy
- **Custom Functions**: Add function call parsing
- **Dialect Support**: Conditional grammar rules

### 2. **Visitor Pattern Support**
```go
type Visitor interface {
    VisitSelectStatement(stmt *SelectStatement) error
    VisitInsertStatement(stmt *InsertStatement) error
    VisitExpression(expr Expression) error
    // ... other visit methods
}

// Enable tree traversal and transformation
func (stmt *SelectStatement) Accept(v Visitor) error {
    return v.VisitSelectStatement(stmt)
}
```

### 3. **Plugin Architecture**
- **Custom Parsers**: Register parsing functions for extensions
- **AST Transformers**: Post-processing AST modifications
- **Validation Rules**: Custom semantic validation

## Testing Architecture

### 1. **Unit Testing Strategy**
- **Grammar Rules**: Test each production rule independently
- **Error Cases**: Verify error detection and recovery
- **Edge Cases**: Boundary conditions and malformed input

### 2. **Integration Testing**
- **End-to-End**: Complete SQL statement parsing
- **Performance**: Large query parsing benchmarks
- **Memory**: AST memory usage analysis

### 3. **Test Data Organization**
```
tests/
├── unit/
│   ├── expressions/
│   ├── statements/
│   └── error_handling/
├── integration/
│   ├── complex_queries/
│   └── performance/
└── fixtures/
    ├── valid_sql/
    └── invalid_sql/
```

## Debugging and Diagnostics

### 1. **Debug Information**
- **Parse Trees**: Visual representation of parsing process
- **Token Traces**: Debug token consumption
- **AST Dumps**: Human-readable AST output
- **Error Context**: Rich error reporting with suggestions

### 2. **Profiling Support**
- **Timing Metrics**: Parse time measurements
- **Memory Profiling**: AST memory usage tracking
- **Bottleneck Analysis**: Identify slow parsing paths

### 3. **Development Tools**
```go
// Debug utilities
func (p *Parser) DumpAST(node Node) string
func (p *Parser) TraceTokens(enabled bool)
func (p *Parser) GetParseMetrics() ParseMetrics
```

## Standards Compliance

### 1. **SQL Standards**
- **ANSI SQL**: Core SQL standard compliance
- **SQL-92/99/2003**: Extended features support
- **Common Extensions**: Popular database vendor extensions

### 2. **Error Message Standards**
- **Informative Messages**: Clear, actionable error descriptions
- **Position Information**: Precise error location reporting
- **Suggestion System**: Helpful recovery suggestions

### 3. **API Consistency**
- **Go Conventions**: Follows Go language standards
- **Error Handling**: Idiomatic Go error handling
- **Documentation**: Comprehensive API documentation

## Future Enhancements

### 1. **Advanced SQL Features**
- **Common Table Expressions (CTEs)**: WITH clause support
- **Window Functions**: OVER clause parsing
- **Recursive Queries**: Recursive CTE support
- **JSON Operations**: JSON path expressions

### 2. **Performance Optimizations**
- **Parallel Parsing**: Multi-threaded parsing for large queries
- **Incremental Parsing**: Parse only changed portions
- **Compiled Grammar**: Pre-compiled parsing tables

### 3. **Developer Experience**
- **LSP Integration**: Language Server Protocol support
- **Syntax Highlighting**: Rich syntax highlighting information
- **Code Completion**: Intelligent completion suggestions
- **Refactoring Support**: AST-based query refactoring