# Phase 1: Parser Implementation - Respecting Architecture

## Architecture-Aligned Implementation Plan

Based on our layered architecture (ARCH.md), Phase 1 focuses on completing the **SQL Compiler Layer** components while respecting module boundaries and dependencies.

## Architecture Review

```
┌─────────────────────────────────────────────────────────────┐
│                SQL Compiler Layer (FOCUS)                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │    SQL      │───▶│    SQL      │───▶│   Query     │     │
│  │ Dispatcher  │    │   Parser    │    │  Compiler   │     │
│  │ [PARTIAL]   │    │ [PARTIAL]   │    │ [PLANNED]   │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│         │                   │                   │           │
│         └───────────────────┼───────────────────┘           │
│                             ▼                               │
│  ┌─────────────────────────────────────────────┐           │
│  │          SQL Lexer [✅ IMPLEMENTED]         │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## Module Dependencies (Current State)

### Implemented Modules (✅)
- `internal/lexer` - Tokenization (40+ token types)
- `internal/storage` - Storage engine with buffer pool
- `internal/config` - Configuration management

### Partially Implemented Modules (🚧)
- `internal/parser` - AST structures defined, parsing logic incomplete
- `internal/dispatcher` - Framework exists, needs integration

### Planned Modules for Phase 1 (🎯)
- Complete `internal/parser` parsing logic
- Integrate with `internal/dispatcher`
- Add validation layer (semantic analysis preparation)

## Phase 1 Implementation Steps (Architecture-Aligned)

### Step 1: Complete Parser Module (`internal/parser/`) 🎯 **WEEK 1**

The parser module already has:
- ✅ AST structures (ast.go)
- ✅ Basic parser framework (parser.go)
- 🚧 Incomplete parsing functions

**What to Complete:**

#### 1.1 Expression Parsing (Priority: CRITICAL)
```go
// File: internal/parser/parser.go

// Complete these existing functions:
func (p *Parser) parseExpression() Expression {
    // Currently incomplete - needs full implementation
    // Priority: Binary operators, comparisons, logic
}

func (p *Parser) parsePrimaryExpression() Expression {
    // Handle: identifiers, literals, parentheses, function calls
}

func (p *Parser) parseBinaryExpression(left Expression, precedence int) Expression {
    // Operator precedence climbing algorithm
    // Handle: +, -, *, /, =, !=, <, >, <=, >=, AND, OR
}
```

**Implementation Order:**
1. Literal expressions (numbers, strings, NULL, TRUE, FALSE)
2. Identifier expressions (column names, table.column)
3. Binary expressions (arithmetic, comparison, logical)
4. Unary expressions (NOT, -)
5. Function calls (COUNT, SUM, AVG, etc.)
6. Parenthesized expressions

#### 1.2 SELECT Statement Parsing (Priority: HIGH)
```go
// File: internal/parser/parser.go

// Already has structure, needs completion:
func (p *Parser) parseSelectStatement() *SelectStatement {
    // ✅ Basic framework exists
    // 🚧 Complete: subqueries, DISTINCT, aliases
}

func (p *Parser) parseSelectClause() *SelectClause {
    // ✅ Exists but needs enhancement
    // Add: column aliases (AS), wildcard (*), table.* 
}

func (p *Parser) parseFromClause() *FromClause {
    // ✅ Basic exists
    // Complete: table aliases, subqueries in FROM
}

func (p *Parser) parseWhereClause() *WhereClause {
    // ✅ Exists
    // Enhance: complex boolean expressions, IN, LIKE, BETWEEN
}

func (p *Parser) parseJoinClause() *JoinClause {
    // Complete: INNER JOIN, LEFT JOIN, RIGHT JOIN, FULL JOIN
    // Parse: ON conditions
}
```

#### 1.3 INSERT Statement Parsing (Priority: HIGH)
```go
// File: internal/parser/parser.go

// Already has structure, needs completion:
func (p *Parser) parseInsertStatement() *InsertStatement {
    // ✅ Basic framework exists
    // Complete: column list, VALUES, multiple rows
}

// Add helper:
func (p *Parser) parseValuesList() [][]Expression {
    // Parse: VALUES (val1, val2), (val3, val4), ...
}
```

#### 1.4 CREATE TABLE Statement Parsing (Priority: HIGH)
```go
// File: internal/parser/parser.go

// Already exists, needs completion:
func (p *Parser) parseCreateTableStatement() *CreateTableStatement {
    // ✅ Basic framework exists
    // Complete: all column definitions, constraints
}

func (p *Parser) parseColumnDefinition() *ColumnDefinition {
    // Parse: column_name data_type [constraints]
    // Constraints: PRIMARY KEY, NOT NULL, UNIQUE, DEFAULT, CHECK
}

func (p *Parser) parseTableConstraint() *TableConstraint {
    // Parse: PRIMARY KEY (col1, col2), FOREIGN KEY, CHECK
}
```

#### 1.5 DELETE Statement Parsing (Priority: MEDIUM)
```go
// Already exists:
func (p *Parser) parseDeleteStatement() *DeleteStatement {
    // ✅ Basic exists
    // Enhance: ensure WHERE clause parsing works correctly
}
```

#### 1.6 UPDATE Statement Parsing (Priority: MEDIUM)
```go
// Already exists:
func (p *Parser) parseUpdateStatement() *UpdateStatement {
    // ✅ Basic exists
    // Complete: SET clause with multiple assignments
}
```

#### 1.7 DROP TABLE Statement Parsing (Priority: LOW)
```go
// Already exists:
func (p *Parser) parseDropTableStatement() *DropTableStatement {
    // ✅ Basic exists
    // Add: IF EXISTS clause
}
```

### Step 2: Enhance Dispatcher Module (`internal/dispatcher/`) 🎯 **WEEK 2**

The dispatcher is the bridge between parsing and execution.

#### 2.1 Query Dispatch Logic
```go
// File: internal/dispatcher/dispatcher.go

// Already has QueryType enum ✅
// Add dispatch logic:

func (d *Dispatcher) DispatchQuery(sql string) (*QueryResult, error) {
    // 1. Tokenize with lexer
    tokens, err := d.lexer.Tokenize(sql)
    if err != nil {
        return nil, fmt.Errorf("lexer error: %w", err)
    }
    
    // 2. Parse to AST
    stmt, err := d.parser.Parse(tokens)
    if err != nil {
        return nil, fmt.Errorf("parser error: %w", err)
    }
    
    // 3. Classify query type
    queryType := d.classifyQuery(stmt)
    
    // 4. Route to appropriate handler
    switch queryType {
    case QueryTypeSelect:
        return d.handleSelect(stmt.(*parser.SelectStatement))
    case QueryTypeInsert:
        return d.handleInsert(stmt.(*parser.InsertStatement))
    case QueryTypeCreateTable:
        return d.handleCreateTable(stmt.(*parser.CreateTableStatement))
    // ... other cases
    }
}

func (d *Dispatcher) classifyQuery(stmt parser.Statement) QueryType {
    // Determine query type from AST node type
}
```

#### 2.2 Validation Layer (Semantic Analysis Preparation)
```go
// File: internal/dispatcher/validator.go (NEW)

type Validator struct {
    catalog CatalogManager  // For schema lookups
}

func (v *Validator) ValidateStatement(stmt parser.Statement) error {
    // Basic validation before execution:
    // - Table names exist
    // - Column names exist
    // - Data types match
    // - Constraints are valid
    
    switch s := stmt.(type) {
    case *parser.SelectStatement:
        return v.validateSelect(s)
    case *parser.InsertStatement:
        return v.validateInsert(s)
    case *parser.CreateTableStatement:
        return v.validateCreateTable(s)
    // ... other cases
    }
}

func (v *Validator) validateSelect(stmt *parser.SelectStatement) error {
    // Validate:
    // - Table exists
    // - Columns exist in table
    // - Expressions are valid
}
```

### Step 3: Testing Infrastructure 🎯 **ONGOING**

Create comprehensive tests for each component.

#### 3.1 Parser Unit Tests
```go
// File: tests/unit/parser_test.go (NEW)

package unit

import (
    "testing"
    "relational-db/internal/lexer"
    "relational-db/internal/parser"
)

func TestParseSimpleSelect(t *testing.T) {
    sql := "SELECT name FROM users"
    l := lexer.NewLexer(sql)
    p := parser.NewParser(l)
    
    stmt := p.ParseStatement()
    
    if stmt == nil {
        t.Fatalf("Expected statement, got nil")
    }
    
    selectStmt, ok := stmt.(*parser.SelectStatement)
    if !ok {
        t.Fatalf("Expected SelectStatement, got %T", stmt)
    }
    
    // Validate AST structure
    if len(selectStmt.SelectClause.Columns) != 1 {
        t.Errorf("Expected 1 column, got %d", len(selectStmt.SelectClause.Columns))
    }
    
    // More validations...
}

func TestParseSelectWithWhere(t *testing.T) {
    sql := "SELECT name, age FROM users WHERE id = 42"
    // Test WHERE clause parsing
}

func TestParseSelectWithJoin(t *testing.T) {
    sql := "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id"
    // Test JOIN parsing
}

func TestParseInsert(t *testing.T) {
    sql := "INSERT INTO users (name, age) VALUES ('Alice', 25)"
    // Test INSERT parsing
}

func TestParseCreateTable(t *testing.T) {
    sql := `CREATE TABLE users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        age INTEGER,
        email TEXT UNIQUE
    )`
    // Test CREATE TABLE parsing
}

func TestParserErrors(t *testing.T) {
    testCases := []struct{
        sql string
        expectedError string
    }{
        {"SELECT FROM users", "expected column list"},
        {"SELECT name users", "expected FROM keyword"},
        {"INSERT users VALUES (1)", "expected INTO keyword"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.sql, func(t *testing.T) {
            // Test error handling
        })
    }
}
```

#### 3.2 Dispatcher Integration Tests
```go
// File: tests/integration/dispatcher_test.go (NEW)

func TestDispatcherSelectQuery(t *testing.T) {
    // Setup storage and catalog
    dispatcher := createTestDispatcher(t)
    
    // Create table first
    createSQL := "CREATE TABLE users (id INTEGER, name TEXT)"
    _, err := dispatcher.DispatchQuery(createSQL)
    assert.NoError(t, err)
    
    // Insert data
    insertSQL := "INSERT INTO users VALUES (1, 'Alice')"
    _, err = dispatcher.DispatchQuery(insertSQL)
    assert.NoError(t, err)
    
    // Query data
    selectSQL := "SELECT name FROM users WHERE id = 1"
    result, err := dispatcher.DispatchQuery(selectSQL)
    assert.NoError(t, err)
    assert.Equal(t, 1, len(result.Rows))
}
```

## Implementation Timeline

### Week 1: Parser Completion
```
Day 1-2: Expression Parsing
  - Literal expressions
  - Identifier expressions  
  - Binary operators (arithmetic, comparison)
  - Test coverage: 80%+

Day 3-4: SELECT Statement
  - Complete SELECT clause parsing
  - FROM clause with aliases
  - WHERE clause with complex conditions
  - JOIN clauses
  - Test coverage: 80%+

Day 5-6: INSERT & CREATE TABLE
  - INSERT statement with multiple rows
  - CREATE TABLE with all constraints
  - Test coverage: 80%+

Day 7: Review & Bug Fixes
  - Fix parser edge cases
  - Improve error messages
  - Documentation updates
```

### Week 2: Dispatcher Integration
```
Day 1-2: Dispatcher Core
  - Query dispatch logic
  - Query type classification
  - Handler routing

Day 3-4: Validation Layer
  - Basic semantic validation
  - Schema existence checks
  - Type validation

Day 5-6: Integration Testing
  - End-to-end parser → dispatcher tests
  - Error handling tests
  - Performance benchmarks

Day 7: Documentation & Cleanup
  - Update ARCH.md with implementation status
  - Code cleanup and refactoring
  - Prepare for Phase 2
```

## Testing Strategy

### Unit Tests (Parser Module)
- **Coverage Target**: 90%+
- **Focus**: Each parsing function independently
- **Test Data**: Valid SQL, invalid SQL, edge cases

### Integration Tests (Parser + Dispatcher)
- **Coverage Target**: 80%+
- **Focus**: Complete query processing pipeline
- **Test Data**: Real-world SQL queries

### Performance Tests
- **Metrics**: Parsing speed, memory usage
- **Baseline**: Should parse 1000 simple queries in < 100ms

## Success Criteria

### Week 1 Completion
- ✅ Parser can handle all basic SQL statements
- ✅ All AST nodes properly constructed
- ✅ 90%+ test coverage for parser
- ✅ Zero parsing errors for valid SQL
- ✅ Clear error messages for invalid SQL

### Week 2 Completion  
- ✅ Dispatcher routes queries correctly
- ✅ Basic validation works
- ✅ Integration tests pass
- ✅ Ready for Phase 2 (Catalog Manager)

## Module Interaction Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Input (SQL String)                 │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│              Dispatcher (internal/dispatcher)                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ 1. Receive SQL string                                  │ │
│  │ 2. Create Lexer instance                               │ │
│  │ 3. Create Parser instance                              │ │
│  │ 4. Parse to AST                                        │ │
│  │ 5. Validate AST (basic)                                │ │
│  │ 6. Route to handler                                    │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────┬──────────────────────────────┬─────────────────┘
             │                              │
             ▼                              ▼
┌──────────────────────────┐  ┌──────────────────────────────┐
│   Lexer (internal/lexer) │  │  Parser (internal/parser)    │
│  ┌────────────────────┐  │  │  ┌────────────────────────┐  │
│  │ Tokenize SQL       │  │  │  │ Build AST from tokens  │  │
│  │ Return token stream│  │  │  │ Validate syntax        │  │
│  └────────────────────┘  │  │  │ Return Statement       │  │
│  Status: ✅ COMPLETE    │  │  └────────────────────────┘  │
└──────────────────────────┘  │  Status: 🚧 IN PROGRESS    │
                              └──────────────────────────────┘
```

## Files to Create/Modify

### Create New Files:
1. `tests/unit/parser_test.go` - Comprehensive parser tests
2. `tests/unit/expression_test.go` - Expression parsing tests
3. `tests/integration/dispatcher_test.go` - Dispatcher integration tests
4. `internal/dispatcher/validator.go` - Semantic validation layer
5. `docs/PARSER_GRAMMAR.md` - SQL grammar documentation

### Modify Existing Files:
1. `internal/parser/parser.go` - Complete all parsing functions
2. `internal/parser/ast.go` - Add any missing AST node types
3. `internal/dispatcher/dispatcher.go` - Add dispatch logic
4. `ARCH.md` - Update implementation status
5. `WARP.md` - Update with Phase 1 progress

## Next Steps After Phase 1

Once Phase 1 is complete, we move to Phase 2:
- **Catalog Manager** (`internal/storage/catalog.go`) - Schema persistence
- **Record Manager** (`internal/storage/record.go`) - Data serialization
- These will use the parser output (AST) to store schemas

This ensures we follow the architecture's natural flow:
```
SQL → Lexer → Parser → Dispatcher → [Phase 2: Storage] → Results
```

## Key Architectural Principles to Maintain

1. **Layer Separation**: Parser only creates AST, doesn't execute
2. **Module Independence**: Parser doesn't know about storage
3. **Interface-Driven**: All modules interact through clean interfaces
4. **Testability**: Each module can be tested independently
5. **Error Handling**: Errors propagate up through layers cleanly

This plan respects the architecture and provides a clear path to complete Phase 1! 🚀