# Parser Implementation Guide

## 🎯 Policy Compliance

**POLICY**: When implementing a module, ensure it implements:
1. **Algorithms** defined in `internal/parser/ALGO.md`
2. **Data Structures** defined in `internal/parser/DS.md`
3. **Solutions** to problems defined in `internal/parser/PROBLEMS.md`

This guide provides concrete implementation steps that align with these documents.

---

## 📋 Implementation Roadmap

### Phase 1: Fix CREATE TABLE Constraint Parsing (CRITICAL)

**Problem Being Solved**: From `PROBLEMS.md` Section 4
- "Parse column constraints (PRIMARY KEY, NOT NULL, UNIQUE, DEFAULT, CHECK, FOREIGN KEY)"
- "Handle both column-level and table-level constraints"

**Algorithm**: From `ALGO.md` Section 5
- `parseColumnDefinition()` - Algorithm for parsing column specifications with constraints

**Data Structures**: From `DS.md` Section 3.5
- `ColumnDefinition` struct with constraint fields
- `ColumnConstraint` enum for constraint types

#### Implementation Steps

**File**: `internal/parser/parser.go`
**Function**: Locate `parseColumnDefinition()` (around line 800-900)

```go
// From DS.md Section 3.5: ColumnDefinition structure
type ColumnDefinition struct {
    Name            *Identifier
    DataType        *DataType
    IsPrimaryKey    bool
    NotNull         bool
    IsUnique        bool
    DefaultValue    Expression      // Can be nil
    CheckConstraint Expression      // Can be nil
    ForeignKey      *ForeignKeyRef  // Can be nil
}

// Algorithm from ALGO.md Section 5: parseColumnDefinition()
func (p *Parser) parseColumnDefinition() *ColumnDefinition {
    col := &ColumnDefinition{}
    
    // Step 1: Parse column name (required)
    if p.currentToken.Type != lexer.IDENT {
        p.addError("Expected column name")
        return nil
    }
    col.Name = &Identifier{Name: p.currentToken.Literal}
    p.nextToken()
    
    // Step 2: Parse data type (required)
    col.DataType = p.parseDataType()
    if col.DataType == nil {
        p.addError("Expected data type after column name")
        return nil
    }
    
    // Step 3: Parse optional constraints (ALGO.md Section 5.2)
    // Loop until we hit comma, closing paren, or EOF
    for p.currentToken.Type != lexer.COMMA && 
        p.currentToken.Type != lexer.RPAREN && 
        p.currentToken.Type != lexer.EOF {
        
        constraint := p.parseColumnConstraint()
        if constraint == nil {
            break // No more constraints
        }
        
        // Apply constraint to column definition
        switch constraint.Type {
        case CONSTRAINT_PRIMARY_KEY:
            col.IsPrimaryKey = true
        case CONSTRAINT_NOT_NULL:
            col.NotNull = true
        case CONSTRAINT_UNIQUE:
            col.IsUnique = true
        case CONSTRAINT_DEFAULT:
            col.DefaultValue = constraint.Expression
        case CONSTRAINT_CHECK:
            col.CheckConstraint = constraint.Expression
        case CONSTRAINT_FOREIGN_KEY:
            col.ForeignKey = constraint.ForeignKeyRef
        }
    }
    
    return col
}

// Helper function (from ALGO.md Section 5.3)
func (p *Parser) parseColumnConstraint() *ColumnConstraint {
    switch p.currentToken.Type {
    case lexer.PRIMARY:
        return p.parsePrimaryKeyConstraint()
    case lexer.NOT:
        return p.parseNotNullConstraint()
    case lexer.UNIQUE:
        return p.parseUniqueConstraint()
    case lexer.DEFAULT:
        return p.parseDefaultConstraint()
    case lexer.CHECK:
        return p.parseCheckConstraint()
    case lexer.REFERENCES:
        return p.parseForeignKeyConstraint()
    default:
        return nil // No constraint at current position
    }
}

// From ALGO.md Section 5.3.4: DEFAULT constraint parsing
func (p *Parser) parseDefaultConstraint() *ColumnConstraint {
    constraint := &ColumnConstraint{Type: CONSTRAINT_DEFAULT}
    
    // Expect DEFAULT keyword
    if p.currentToken.Type != lexer.DEFAULT {
        return nil
    }
    p.nextToken()
    
    // Parse default value expression
    // CRITICAL: Ensure parseExpression() doesn't return nil
    expr := p.parseExpression()
    if expr == nil {
        p.addError("Expected expression after DEFAULT keyword")
        return nil
    }
    
    constraint.Expression = expr
    return constraint
}

// From ALGO.md Section 5.3.5: CHECK constraint parsing
func (p *Parser) parseCheckConstraint() *ColumnConstraint {
    constraint := &ColumnConstraint{Type: CONSTRAINT_CHECK}
    
    // Expect CHECK keyword
    if p.currentToken.Type != lexer.CHECK {
        return nil
    }
    p.nextToken()
    
    // Expect opening parenthesis
    if p.currentToken.Type != lexer.LPAREN {
        p.addError("Expected ( after CHECK")
        return nil
    }
    p.nextToken()
    
    // Parse check expression
    // CRITICAL: Ensure parseExpression() doesn't return nil
    expr := p.parseExpression()
    if expr == nil {
        p.addError("Expected expression in CHECK constraint")
        return nil
    }
    
    constraint.Expression = expr
    
    // Expect closing parenthesis
    if p.currentToken.Type != lexer.RPAREN {
        p.addError("Expected ) after CHECK expression")
        return nil
    }
    p.nextToken()
    
    return constraint
}
```

**Test After Implementation**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseCreateTableWithConstraints
```

---

### Phase 2: Implement UPDATE Statement

**Problem Being Solved**: From `PROBLEMS.md` Section 3
- "Parse UPDATE statements with SET clauses"
- "Handle multiple column assignments"

**Algorithm**: From `ALGO.md` Section 4
- `parseUpdateStatement()` - Complete algorithm provided

**Data Structures**: From `DS.md` Section 3.3
- `UpdateStatement` struct
- `SetClause` struct for assignments

#### Implementation Steps

**File**: `internal/parser/parser.go`
**Function**: Locate or create `parseUpdateStatement()`

```go
// From DS.md Section 3.3: UpdateStatement structure
type UpdateStatement struct {
    TableName   *Identifier
    SetClauses  []*SetClause
    WhereClause *WhereClause
}

type SetClause struct {
    Column *Identifier
    Value  Expression
}

// Algorithm from ALGO.md Section 4: parseUpdateStatement()
func (p *Parser) parseUpdateStatement() *UpdateStatement {
    stmt := &UpdateStatement{}
    
    // Step 1: UPDATE keyword already consumed by ParseStatement()
    
    // Step 2: Parse table name
    if !p.expectPeek(lexer.IDENT) {
        p.addError("Expected table name after UPDATE")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Step 3: Expect SET keyword
    if !p.expectPeek(lexer.SET) {
        p.addError("Expected SET keyword in UPDATE statement")
        return nil
    }
    p.nextToken()
    
    // Step 4: Parse SET clauses (comma-separated assignments)
    stmt.SetClauses = []*SetClause{}
    
    for {
        setClause := &SetClause{}
        
        // Parse column name
        if p.currentToken.Type != lexer.IDENT {
            p.addError("Expected column name in SET clause")
            return nil
        }
        setClause.Column = &Identifier{Name: p.currentToken.Literal}
        
        // Expect equals sign
        if !p.expectPeek(lexer.ASSIGN) {
            p.addError("Expected = after column name")
            return nil
        }
        p.nextToken()
        
        // Parse value expression
        value := p.parseExpression()
        if value == nil {
            p.addError("Expected expression after = in SET clause")
            return nil
        }
        setClause.Value = value
        
        stmt.SetClauses = append(stmt.SetClauses, setClause)
        
        // Check for more assignments (comma-separated)
        if p.peekToken.Type != lexer.COMMA {
            break
        }
        p.nextToken() // consume comma
        p.nextToken() // move to next column
    }
    
    // Step 5: Parse optional WHERE clause
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

**Add to AST** (`internal/parser/ast.go` if not already present):
```go
func (s *SetClause) NodeType() string { return "SetClause" }
func (s *SetClause) String() string {
    return fmt.Sprintf("%s = %s", s.Column.String(), s.Value.String())
}
```

**Test After Implementation**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseUpdate
```

---

### Phase 3: Implement DELETE Statement

**Problem Being Solved**: From `PROBLEMS.md` Section 3
- "Parse DELETE statements with optional WHERE clause"

**Algorithm**: From `ALGO.md` Section 3
- `parseDeleteStatement()` - Complete algorithm provided

**Data Structures**: From `DS.md` Section 3.4
- `DeleteStatement` struct

#### Implementation Steps

**File**: `internal/parser/parser.go`
**Function**: Locate or create `parseDeleteStatement()`

```go
// From DS.md Section 3.4: DeleteStatement structure
type DeleteStatement struct {
    TableName   *Identifier
    WhereClause *WhereClause
}

// Algorithm from ALGO.md Section 3: parseDeleteStatement()
func (p *Parser) parseDeleteStatement() *DeleteStatement {
    stmt := &DeleteStatement{}
    
    // Step 1: DELETE keyword already consumed by ParseStatement()
    
    // Step 2: Expect FROM keyword
    if !p.expectPeek(lexer.FROM) {
        p.addError("Expected FROM after DELETE")
        return nil
    }
    
    // Step 3: Parse table name
    if !p.expectPeek(lexer.IDENT) {
        p.addError("Expected table name after FROM")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Step 4: Parse optional WHERE clause
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

**Test After Implementation**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDelete
```

---

### Phase 4: Implement DROP TABLE Statement

**Problem Being Solved**: From `PROBLEMS.md` Section 3
- "Parse DROP TABLE statements with optional IF EXISTS"

**Algorithm**: From `ALGO.md` Section 6
- `parseDropTableStatement()` - Complete algorithm provided

**Data Structures**: From `DS.md` Section 3.7
- `DropTableStatement` struct

#### Implementation Steps

**File**: `internal/parser/parser.go`
**Function**: Locate or create `parseDropTableStatement()`

```go
// From DS.md Section 3.7: DropTableStatement structure
type DropTableStatement struct {
    TableName *Identifier
    IfExists  bool
}

// Algorithm from ALGO.md Section 6: parseDropTableStatement()
func (p *Parser) parseDropTableStatement() *DropTableStatement {
    stmt := &DropTableStatement{}
    
    // Step 1: DROP keyword already consumed by ParseStatement()
    
    // Step 2: Expect TABLE keyword
    if !p.expectPeek(lexer.TABLE) {
        p.addError("Expected TABLE after DROP")
        return nil
    }
    p.nextToken()
    
    // Step 3: Check for optional IF EXISTS
    if p.currentToken.Type == lexer.IF {
        if p.expectPeek(lexer.EXISTS) {
            stmt.IfExists = true
            p.nextToken()
        }
    }
    
    // Step 4: Parse table name
    if p.currentToken.Type != lexer.IDENT {
        p.addError("Expected table name in DROP TABLE")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    return stmt
}
```

**Test After Implementation**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDropTable
```

---

## 🧪 Verification Checklist

After completing each phase, verify against the policy documents:

### ✅ Algorithm Compliance (ALGO.md)
- [ ] Implementation follows recursive descent pattern (Section 1)
- [ ] Expression parsing uses precedence climbing (Section 2)
- [ ] Error recovery mechanism implemented (Section 8)
- [ ] Time complexity is O(n) for linear parsing (documented in ALGO.md)

### ✅ Data Structure Compliance (DS.md)
- [ ] Parser state structure matches DS.md Section 1
- [ ] AST node hierarchy follows DS.md Section 2
- [ ] Statement structures match DS.md Section 3
- [ ] Expression structures match DS.md Section 4

### ✅ Problem Solution Compliance (PROBLEMS.md)
- [ ] Syntactic analysis problem solved (Section 1)
- [ ] Operator precedence handled correctly (Section 2)
- [ ] DDL statements parsed correctly (Section 3)
- [ ] DML statements parsed correctly (Section 3)
- [ ] Constraints parsed correctly (Section 4)

---

## 🚀 Testing Strategy

### Unit Tests
```bash
# Run all parser tests
go test ./tests/unit/parser_test.go -v

# Run specific test
go test ./tests/unit/parser_test.go -v -run TestName

# Check coverage
go test ./tests/unit/parser_test.go -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Expected Test Results
```
✅ TestParseSimpleSelect              - ALGO.md Section 1.2
✅ TestParseSelectStar                - ALGO.md Section 1.2
✅ TestParseSelectMultipleColumns     - ALGO.md Section 1.2
✅ TestParseSelectWithWhere           - ALGO.md Section 1.2 + 2.1
✅ TestParseSelectWithComplexWhere    - ALGO.md Section 2.2
✅ TestParseSelectWithOrderBy         - ALGO.md Section 1.2
✅ TestParseSelectWithLimit           - ALGO.md Section 1.2
✅ TestParseInsertSimple              - ALGO.md Section 1.3
✅ TestParseInsertWithoutColumns      - ALGO.md Section 1.3
✅ TestParseCreateTable               - ALGO.md Section 1.5
✅ TestParseCreateTableWithConstraints - ALGO.md Section 5 (FIX FIRST)
✅ TestParseUpdate                    - ALGO.md Section 4 (IMPLEMENT)
✅ TestParseDelete                    - ALGO.md Section 3 (IMPLEMENT)
✅ TestParseDropTable                 - ALGO.md Section 6 (IMPLEMENT)
✅ TestParserErrors                   - ALGO.md Section 8

TARGET: 15/15 passing (100%)
```

---

## 📊 Progress Tracking

| Phase | Algorithm | Data Structure | Problem | Status |
|-------|-----------|----------------|---------|--------|
| CREATE TABLE Fix | ALGO.md §5 | DS.md §3.5 | PROBLEMS.md §4 | 🎯 START |
| UPDATE | ALGO.md §4 | DS.md §3.3 | PROBLEMS.md §3 | ⏸️ |
| DELETE | ALGO.md §3 | DS.md §3.4 | PROBLEMS.md §3 | ⏸️ |
| DROP TABLE | ALGO.md §6 | DS.md §3.7 | PROBLEMS.md §3 | ⏸️ |

---

## 🔍 Common Issues & Solutions

### Issue 1: parseExpression() returns nil
**Root Cause**: Expression parsing incomplete (PROBLEMS.md Section 2)
**Solution**: Ensure `parsePrimaryExpression()` handles:
- Literals (numbers, strings, booleans, NULL)
- Identifiers (column names)
- Parenthesized expressions

### Issue 2: Infinite loop in constraint parsing
**Root Cause**: Not advancing token position (ALGO.md Section 5.2)
**Solution**: Always call `p.nextToken()` after consuming a token

### Issue 3: Panic on NULL pointer
**Root Cause**: Not checking for nil before dereferencing (DS.md Section 1)
**Solution**: Add nil checks:
```go
expr := p.parseExpression()
if expr == nil {
    p.addError("Expected expression")
    return nil
}
```

---

## 📚 Quick Reference

**Module Documentation**:
- `internal/parser/ALGO.md` - All parsing algorithms
- `internal/parser/DS.md` - All data structures
- `internal/parser/PROBLEMS.md` - Problems being solved

**Implementation Files**:
- `internal/parser/parser.go` - Main parser implementation
- `internal/parser/ast.go` - AST node definitions
- `tests/unit/parser_test.go` - Parser unit tests

**Architecture Documents**:
- `ARCH.md` - System architecture (6 layers)
- `FLOW.md` - Processing flow through layers
- `DATA.md` - Data transformations
- `docs/PHASE1_PARSER_IMPLEMENTATION.md` - Full Phase 1 plan

---

**Ready to implement?** Start with Phase 1: Fix CREATE TABLE Constraint Parsing 🎯

This implementation ensures compliance with your policy:
- ✅ Algorithms from `ALGO.md`
- ✅ Data structures from `DS.md`
- ✅ Solutions to problems in `PROBLEMS.md`
