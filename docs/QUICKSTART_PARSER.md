# Quick Start: Parser Implementation

## Immediate Tasks (Start Here)

### Task 1: Fix CREATE TABLE Panic (30 minutes) ⚠️ CRITICAL

**Location**: `internal/parser/parser.go`  
**Function**: Find `parseColumnDefinition()` around line 800-900

**Problem**: When parsing constraints like `DEFAULT 0.0` or `CHECK (quantity >= 0)`, the parser encounters a NULL pointer.

**Fix Steps**:

1. Find the `parseColumnDefinition()` function
2. Look for constraint parsing logic
3. Add NULL checks before dereferencing pointers
4. Handle these constraint keywords properly:
   - `DEFAULT <expression>`
   - `CHECK (<expression>)`
   - `UNIQUE`
   - `PRIMARY KEY`
   - `NOT NULL`
   - `FOREIGN KEY REFERENCES table(column)`

**Example Fix Pattern**:
```go
func (p *Parser) parseColumnDefinition() *ColumnDefinition {
    col := &ColumnDefinition{}
    
    // Parse column name
    col.Name = p.currentToken.Literal
    p.nextToken()
    
    // Parse data type
    col.DataType = p.parseDataType()
    
    // Parse constraints
    for p.currentToken.Type != lexer.COMMA && 
        p.currentToken.Type != lexer.RPAREN && 
        p.currentToken.Type != lexer.EOF {
        
        switch p.currentToken.Type {
        case lexer.PRIMARY:
            p.nextToken() // Skip PRIMARY
            if p.expectPeek(lexer.KEY) {
                col.IsPrimaryKey = true
            }
            
        case lexer.NOT:
            p.nextToken() // Skip NOT
            if p.expectPeek(lexer.NULL) {
                col.NotNull = true
            }
            
        case lexer.UNIQUE:
            col.IsUnique = true
            p.nextToken()
            
        case lexer.DEFAULT:
            p.nextToken() // Skip DEFAULT
            col.DefaultValue = p.parseExpression()  // <- Make sure this doesn't return nil
            if col.DefaultValue == nil {
                p.addError("Expected expression after DEFAULT")
            }
            
        case lexer.CHECK:
            p.nextToken() // Skip CHECK
            if !p.expectPeek(lexer.LPAREN) {
                break
            }
            p.nextToken() // Skip (
            col.CheckConstraint = p.parseExpression()  // <- Make sure this doesn't return nil
            if col.CheckConstraint == nil {
                p.addError("Expected expression in CHECK constraint")
            }
            if !p.expectPeek(lexer.RPAREN) {
                p.addError("Expected ) after CHECK constraint")
            }
            p.nextToken()
            
        default:
            return col
        }
    }
    
    return col
}
```

**Test**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseCreateTableWithConstraints
```

---

### Task 2: Implement UPDATE Statement (45 minutes)

**Location**: `internal/parser/parser.go`  
**Function**: Find `parseUpdateStatement()` (probably incomplete)

**Expected SQL**: `UPDATE users SET age = 26, status = 'active' WHERE name = 'Alice'`

**Implementation**:
```go
func (p *Parser) parseUpdateStatement() *UpdateStatement {
    stmt := &UpdateStatement{}
    
    // Expect UPDATE keyword (already consumed by ParseStatement)
    
    // Parse table name
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Expect SET keyword
    if !p.expectPeek(lexer.SET) {
        p.addError("Expected SET after table name in UPDATE")
        return nil
    }
    p.nextToken()
    
    // Parse assignment list: column = value, column = value
    stmt.SetClauses = []*SetClause{}
    
    for {
        setClause := &SetClause{}
        
        // Parse column name
        if p.currentToken.Type != lexer.IDENT {
            p.addError("Expected column name in SET clause")
            return nil
        }
        setClause.Column = &Identifier{Name: p.currentToken.Literal}
        
        // Expect =
        if !p.expectPeek(lexer.ASSIGN) {
            p.addError("Expected = after column name in SET clause")
            return nil
        }
        p.nextToken()
        
        // Parse value expression
        setClause.Value = p.parseExpression()
        if setClause.Value == nil {
            p.addError("Expected expression after = in SET clause")
            return nil
        }
        
        stmt.SetClauses = append(stmt.SetClauses, setClause)
        
        // Check for more assignments
        if p.peekToken.Type != lexer.COMMA {
            break
        }
        p.nextToken() // Skip comma
        p.nextToken() // Move to next column
    }
    
    // Parse optional WHERE clause
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

**Note**: You may need to add `SetClause` to `ast.go` if it doesn't exist:
```go
type SetClause struct {
    Column *Identifier
    Value  Expression
}

func (s *SetClause) NodeType() string { return "SetClause" }
func (s *SetClause) String() string {
    return s.Column.String() + " = " + s.Value.String()
}
```

**Test**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseUpdate
```

---

### Task 3: Implement DELETE Statement (30 minutes)

**Location**: `internal/parser/parser.go`  
**Function**: Find `parseDeleteStatement()` (probably incomplete)

**Expected SQL**: `DELETE FROM users WHERE id = 42`

**Implementation**:
```go
func (p *Parser) parseDeleteStatement() *DeleteStatement {
    stmt := &DeleteStatement{}
    
    // Expect DELETE keyword (already consumed by ParseStatement)
    
    // Expect FROM keyword
    if !p.expectPeek(lexer.FROM) {
        p.addError("Expected FROM after DELETE")
        return nil
    }
    
    // Parse table name
    if !p.expectPeek(lexer.IDENT) {
        p.addError("Expected table name after FROM")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Parse optional WHERE clause
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

**Test**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDelete
```

---

### Task 4: Implement DROP TABLE Statement (20 minutes)

**Location**: `internal/parser/parser.go`  
**Function**: Find `parseDropTableStatement()` (probably incomplete)

**Expected SQL**: `DROP TABLE users` or `DROP TABLE IF EXISTS users`

**Implementation**:
```go
func (p *Parser) parseDropTableStatement() *DropTableStatement {
    stmt := &DropTableStatement{}
    
    // Expect DROP keyword (already consumed by ParseStatement)
    
    // Expect TABLE keyword
    if !p.expectPeek(lexer.TABLE) {
        p.addError("Expected TABLE after DROP")
        return nil
    }
    p.nextToken()
    
    // Check for optional IF EXISTS
    if p.currentToken.Type == lexer.IF {
        if p.expectPeek(lexer.EXISTS) {
            stmt.IfExists = true
            p.nextToken()
        }
    }
    
    // Parse table name
    if p.currentToken.Type != lexer.IDENT {
        p.addError("Expected table name")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    return stmt
}
```

**Test**:
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDropTable
```

---

## Running All Tests

After completing the above tasks, run all parser tests:

```bash
cd /home/yathur/2025SRU/GtrLab/DBLab/namyohDB
go test ./tests/unit/parser_test.go -v
```

**Expected Result**: 15/15 tests passing ✅

---

## Common Issues & Solutions

### Issue 1: "undefined: SetClause"
**Solution**: Add `SetClause` struct to `internal/parser/ast.go` (see Task 2 above)

### Issue 2: "parseExpression returns nil"
**Solution**: Make sure `parseExpression()` handles all basic cases:
- Literals (numbers, strings, booleans, NULL)
- Identifiers (column names)
- Binary operators (=, !=, <, >, <=, >=, +, -, *, /, AND, OR)

### Issue 3: "expectPeek fails"
**Solution**: Check that the lexer is producing the correct tokens. Run:
```bash
go test ./tests/unit/storage_test.go -v  # Should still pass (23 tests)
```

### Issue 4: Parser hangs or infinite loop
**Solution**: Always call `p.nextToken()` to advance the parser. Common pattern:
```go
if p.expectPeek(lexer.SOME_TOKEN) {
    // Token is now in currentToken
    p.nextToken()  // Move to next token
}
```

---

## Debugging Tips

### Print Current Token
```go
fmt.Printf("DEBUG: current=%v peek=%v\n", p.currentToken, p.peekToken)
```

### Check Lexer Output
```go
l := lexer.NewLexer("UPDATE users SET age = 26")
for {
    tok := l.NextToken()
    fmt.Printf("Token: Type=%v Literal=%q\n", tok.Type, tok.Literal)
    if tok.Type == lexer.EOF {
        break
    }
}
```

### Run Single Test
```bash
go test ./tests/unit/parser_test.go -v -run TestName
```

### Get Stack Trace for Panic
```bash
go test ./tests/unit/parser_test.go -v 2>&1 | tee test_output.txt
```

---

## Timeline

| Task | Duration | Status |
|------|----------|--------|
| Fix CREATE TABLE panic | 30 min | 🎯 START HERE |
| Implement UPDATE | 45 min | ⏸️ Next |
| Implement DELETE | 30 min | ⏸️ Next |
| Implement DROP TABLE | 20 min | ⏸️ Next |
| Run all tests | 5 min | ⏸️ Final check |
| **TOTAL** | **~2 hours** | **Phase 1, Day 1** |

---

## Success Criteria

✅ All 15 parser tests passing  
✅ No panics or crashes  
✅ CREATE TABLE handles all constraints  
✅ UPDATE, DELETE, DROP TABLE fully functional  
✅ Ready for Day 2: Expression parsing enhancement

---

## What's Next (Day 2-3)

After completing these tasks, you'll move on to:

1. **Expression Parsing** (Day 2-3):
   - Operator precedence
   - Function calls (COUNT, SUM, AVG)
   - Nested expressions
   - Parentheses

2. **SELECT Enhancements** (Day 3-4):
   - Column aliases (AS)
   - Table aliases
   - DISTINCT
   - JOINs

3. **Integration** (Day 5-7):
   - Dispatcher integration
   - Validator creation
   - End-to-end testing

---

**Ready to start?** Begin with **Task 1: Fix CREATE TABLE Panic**

Good luck! 🚀
