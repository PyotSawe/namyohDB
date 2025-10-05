# Parser Implementation - Start Here

## 🎯 Your Policy-Compliant Implementation Plan

**POLICY REMINDER**: 
- ✅ Implement **Algorithms** from `internal/parser/ALGO.md`
- ✅ Use **Data Structures** from `internal/parser/DS.md`  
- ✅ Solve **Problems** from `internal/parser/PROBLEMS.md`

---

## 📊 Current Status

### Test Results: 10/15 Passing (66%)

```bash
# Last test run showed:
✅ TestParseSimpleSelect
✅ TestParseSelectStar
✅ TestParseSelectMultipleColumns
✅ TestParseSelectWithWhere
✅ TestParseSelectWithComplexWhere
✅ TestParseSelectWithOrderBy
✅ TestParseSelectWithLimit
✅ TestParseInsertSimple
✅ TestParseInsertWithoutColumns
✅ TestParseCreateTable
❌ TestParseCreateTableWithConstraints (PANIC)
⏸️ TestParseUpdate (not implemented)
⏸️ TestParseDelete (not implemented)
⏸️ TestParseDropTable (not implemented)
⏸️ TestParserErrors (not implemented)
```

---

## 🚀 Task 1: Fix CREATE TABLE Panic (START HERE)

### Policy Compliance
- **Algorithm**: `internal/parser/ALGO.md` Section 5 - "Column Constraint Parsing"
- **Data Structure**: `internal/parser/DS.md` Section 3.5 - "ColumnDefinition Structure"
- **Problem Solved**: `internal/parser/PROBLEMS.md` Section 4 - "Constraint Parsing Problem"

### What To Do

1. **Open File**: `internal/parser/parser.go`

2. **Find Function**: Search for `parseColumnDefinition()` (around line 800-900)

3. **The Problem**: The function panics when parsing:
   ```sql
   CREATE TABLE products (
       name TEXT NOT NULL UNIQUE,
       price REAL DEFAULT 0.0,          -- <- Crashes here
       quantity INTEGER CHECK (quantity >= 0)  -- <- Or here
   )
   ```

4. **Root Cause**: 
   - `parseExpression()` returns `nil` for DEFAULT value
   - Code tries to dereference nil pointer
   - No nil check before using the expression

5. **The Fix** (Based on ALGO.md Section 5.3.4 and 5.3.5):

```go
// Find this in parser.go and ADD nil checks

func (p *Parser) parseColumnDefinition() *ColumnDefinition {
    col := &ColumnDefinition{}
    
    // ... existing code for name and data type ...
    
    // FIND the constraint parsing loop, it looks like this:
    for p.currentToken.Type != lexer.COMMA && 
        p.currentToken.Type != lexer.RPAREN {
        
        switch p.currentToken.Type {
        
        case lexer.DEFAULT:
            p.nextToken() // Skip DEFAULT
            
            // THIS IS THE CRITICAL FIX:
            expr := p.parseExpression()
            if expr == nil {
                p.addError("Expected expression after DEFAULT")
                return nil  // or continue to next constraint
            }
            col.DefaultValue = expr
            
        case lexer.CHECK:
            p.nextToken() // Skip CHECK
            if p.currentToken.Type != lexer.LPAREN {
                p.addError("Expected ( after CHECK")
                return nil
            }
            p.nextToken() // Skip (
            
            // THIS IS THE CRITICAL FIX:
            expr := p.parseExpression()
            if expr == nil {
                p.addError("Expected expression in CHECK constraint")
                return nil
            }
            col.CheckConstraint = expr
            
            if p.currentToken.Type != lexer.RPAREN {
                p.addError("Expected ) after CHECK expression")
                return nil
            }
            p.nextToken() // Skip )
        
        // ... other constraint cases ...
        }
    }
    
    return col
}
```

### Test Your Fix
```bash
cd /home/yathur/2025SRU/GtrLab/DBLab/namyohDB
go test ./tests/unit/parser_test.go -v -run TestParseCreateTableWithConstraints
```

**Expected**: Test should PASS ✅ (no more panic)

---

## 🚀 Task 2: Implement UPDATE Statement

### Policy Compliance
- **Algorithm**: `internal/parser/ALGO.md` Section 4 - "UPDATE Statement Parsing"
- **Data Structure**: `internal/parser/DS.md` Section 3.3 - "UpdateStatement Structure"
- **Problem Solved**: `internal/parser/PROBLEMS.md` Section 3.2 - "DML Statement Parsing"

### What To Do

1. **Open File**: `internal/parser/parser.go`

2. **Find Function**: Search for `parseUpdateStatement()` (might be incomplete)

3. **Implement** (Based on ALGO.md Section 4):

```go
func (p *Parser) parseUpdateStatement() *UpdateStatement {
    stmt := &UpdateStatement{}
    
    // UPDATE keyword already consumed
    
    // Parse table name
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Expect SET
    if !p.expectPeek(lexer.SET) {
        p.addError("Expected SET in UPDATE")
        return nil
    }
    p.nextToken()
    
    // Parse SET clauses
    stmt.SetClauses = []*SetClause{}
    for {
        sc := &SetClause{}
        
        if p.currentToken.Type != lexer.IDENT {
            p.addError("Expected column name")
            return nil
        }
        sc.Column = &Identifier{Name: p.currentToken.Literal}
        
        if !p.expectPeek(lexer.ASSIGN) {
            return nil
        }
        p.nextToken()
        
        sc.Value = p.parseExpression()
        if sc.Value == nil {
            p.addError("Expected value expression")
            return nil
        }
        
        stmt.SetClauses = append(stmt.SetClauses, sc)
        
        if p.peekToken.Type != lexer.COMMA {
            break
        }
        p.nextToken() // comma
        p.nextToken() // next column
    }
    
    // Optional WHERE
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

4. **Check AST**: Ensure `SetClause` exists in `internal/parser/ast.go`:

```go
type SetClause struct {
    Column *Identifier
    Value  Expression
}

func (s *SetClause) NodeType() string { return "SetClause" }
func (s *SetClause) String() string {
    return fmt.Sprintf("%s = %s", s.Column.String(), s.Value.String())
}
```

### Test Your Implementation
```bash
go test ./tests/unit/parser_test.go -v -run TestParseUpdate
```

**Expected**: Test should PASS ✅

---

## 🚀 Task 3: Implement DELETE Statement

### Policy Compliance
- **Algorithm**: `internal/parser/ALGO.md` Section 3 - "DELETE Statement Parsing"
- **Data Structure**: `internal/parser/DS.md` Section 3.4 - "DeleteStatement Structure"
- **Problem Solved**: `internal/parser/PROBLEMS.md` Section 3.2 - "DML Statement Parsing"

### What To Do

1. **Open File**: `internal/parser/parser.go`

2. **Find Function**: Search for `parseDeleteStatement()` (might be incomplete)

3. **Implement** (Based on ALGO.md Section 3):

```go
func (p *Parser) parseDeleteStatement() *DeleteStatement {
    stmt := &DeleteStatement{}
    
    // DELETE keyword already consumed
    
    // Expect FROM
    if !p.expectPeek(lexer.FROM) {
        p.addError("Expected FROM after DELETE")
        return nil
    }
    
    // Parse table name
    if !p.expectPeek(lexer.IDENT) {
        p.addError("Expected table name")
        return nil
    }
    stmt.TableName = &Identifier{Name: p.currentToken.Literal}
    
    // Optional WHERE
    if p.peekToken.Type == lexer.WHERE {
        p.nextToken()
        stmt.WhereClause = p.parseWhereClause()
    }
    
    return stmt
}
```

### Test Your Implementation
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDelete
```

**Expected**: Test should PASS ✅

---

## 🚀 Task 4: Implement DROP TABLE Statement

### Policy Compliance
- **Algorithm**: `internal/parser/ALGO.md` Section 6 - "DROP TABLE Statement Parsing"
- **Data Structure**: `internal/parser/DS.md` Section 3.7 - "DropTableStatement Structure"
- **Problem Solved**: `internal/parser/PROBLEMS.md` Section 3.1 - "DDL Statement Parsing"

### What To Do

1. **Open File**: `internal/parser/parser.go`

2. **Find Function**: Search for `parseDropTableStatement()` (might be incomplete)

3. **Implement** (Based on ALGO.md Section 6):

```go
func (p *Parser) parseDropTableStatement() *DropTableStatement {
    stmt := &DropTableStatement{}
    
    // DROP keyword already consumed
    
    // Expect TABLE
    if !p.expectPeek(lexer.TABLE) {
        p.addError("Expected TABLE after DROP")
        return nil
    }
    p.nextToken()
    
    // Optional IF EXISTS
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

### Test Your Implementation
```bash
go test ./tests/unit/parser_test.go -v -run TestParseDropTable
```

**Expected**: Test should PASS ✅

---

## ✅ Final Verification

After completing all 4 tasks, run the full test suite:

```bash
cd /home/yathur/2025SRU/GtrLab/DBLab/namyohDB
go test ./tests/unit/parser_test.go -v
```

**Expected Result**: 15/15 tests passing (100%) ✅

---

## 📊 Progress Checklist

Track your progress as you implement:

- [ ] **Task 1**: CREATE TABLE constraint fix (CRITICAL)
  - [ ] Added nil check for DEFAULT expression
  - [ ] Added nil check for CHECK expression
  - [ ] Test passes: `TestParseCreateTableWithConstraints`
  
- [ ] **Task 2**: UPDATE statement implementation
  - [ ] Implemented `parseUpdateStatement()`
  - [ ] Added/verified `SetClause` in ast.go
  - [ ] Test passes: `TestParseUpdate`
  
- [ ] **Task 3**: DELETE statement implementation
  - [ ] Implemented `parseDeleteStatement()`
  - [ ] Test passes: `TestParseDelete`
  
- [ ] **Task 4**: DROP TABLE implementation
  - [ ] Implemented `parseDropTableStatement()`
  - [ ] Test passes: `TestParseDropTable`

- [ ] **Final Verification**: All tests passing
  - [ ] 15/15 tests pass
  - [ ] No panics or crashes
  - [ ] Ready for Phase 1 Day 2 (Expression parsing)

---

## 🔍 Policy Compliance Verification

After implementation, verify you've followed the policy:

### ✅ Algorithm Compliance
- [ ] Used algorithms from `internal/parser/ALGO.md`
- [ ] Section 3: DELETE statement ✓
- [ ] Section 4: UPDATE statement ✓
- [ ] Section 5: Constraint parsing ✓
- [ ] Section 6: DROP TABLE statement ✓

### ✅ Data Structure Compliance
- [ ] Used structures from `internal/parser/DS.md`
- [ ] Section 3.3: UpdateStatement ✓
- [ ] Section 3.4: DeleteStatement ✓
- [ ] Section 3.5: ColumnDefinition ✓
- [ ] Section 3.7: DropTableStatement ✓

### ✅ Problem Solution Compliance
- [ ] Solved problems from `internal/parser/PROBLEMS.md`
- [ ] Section 3.1: DDL statement parsing ✓
- [ ] Section 3.2: DML statement parsing ✓
- [ ] Section 4: Constraint parsing ✓

---

## 🚨 Common Issues

### Issue: Can't find `parseColumnDefinition()`
**Solution**: Search for "ColumnDefinition" in parser.go, or look for CREATE TABLE parsing

### Issue: `expectPeek` not defined
**Solution**: This is a parser helper method. If missing, add:
```go
func (p *Parser) expectPeek(t lexer.TokenType) bool {
    if p.peekToken.Type == t {
        p.nextToken()
        return true
    }
    p.addError(fmt.Sprintf("Expected %v, got %v", t, p.peekToken.Type))
    return false
}
```

### Issue: `SetClause` undefined
**Solution**: Add it to `internal/parser/ast.go` (see Task 2 above)

---

## 📚 Reference Documents

**Module Policy Documents** (LOCAL to parser module):
- `internal/parser/ALGO.md` - Algorithms you MUST use
- `internal/parser/DS.md` - Data structures you MUST use
- `internal/parser/PROBLEMS.md` - Problems you MUST solve

**Implementation Guides**:
- `docs/PARSER_IMPLEMENTATION_GUIDE.md` - Detailed policy-compliant guide
- `docs/QUICKSTART_PARSER.md` - Quick implementation reference
- `docs/PARSER_STATUS.md` - Current status and progress

**Architecture Documents**:
- `ARCH.md` - System architecture
- `FLOW.md` - Processing flow
- `DATA.md` - Data transformations

---

## 🎯 Next Steps After Completion

Once you have 15/15 tests passing, you'll move to:

**Phase 1, Days 2-3**: Expression Parsing Enhancement
- Implement operator precedence (ALGO.md Section 2)
- Add function call support (ALGO.md Section 2.4)
- Create expression tests

**Phase 1, Days 4-5**: SELECT Enhancement
- Add JOIN support (ALGO.md Section 1.2)
- Add aliases (ALGO.md Section 1.2)
- Add subqueries (ALGO.md Section 1.2)

**Phase 1, Days 6-7**: Dispatcher Integration
- Connect parser to dispatcher (Week 2 plan)
- Add validation layer (Week 2 plan)
- Integration testing

---

**START NOW**: Open `internal/parser/parser.go` and begin with Task 1 (CREATE TABLE fix) 🚀

**Estimated Time**: ~2 hours total for all 4 tasks

**Your Goal**: 15/15 tests passing ✅
