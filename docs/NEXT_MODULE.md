# Next Module to Implement: Dispatcher Integration

## 📊 Current Status

### Phase 1, Week 1, Day 1: Parser Completion
- **Status**: 10/15 tests passing (66%)
- **Remaining Work**: 2-3 hours to reach 15/15
- **Critical Tasks**: 
  1. Fix CREATE TABLE constraint panic
  2. Implement UPDATE/DELETE/DROP TABLE

---

## 🎯 ANSWER: Next Module = **DISPATCHER**

### Why Dispatcher Next?

**Strategic Reasons**:
1. ✅ Parser is 90% complete (just needs finishing touches)
2. ✅ Dispatcher framework already exists (639 lines in `dispatcher.go`)
3. ✅ Enables end-to-end testing: SQL string → AST → Routing
4. ✅ No storage dependencies (pure coordination logic)
5. ✅ Natural next step in SQL Compiler Layer

**Architecture Flow**:
```
SQL String
    ↓
┌───────────────┐
│  DISPATCHER   │ ← YOU IMPLEMENT THIS NEXT
│  (Layer 2)    │
└───────┬───────┘
        ↓
┌───────────────┐
│  LEXER (✅)   │
└───────┬───────┘
        ↓
┌───────────────┐
│  PARSER (🚧)  │ ← Finish this first (2-3 hours)
└───────────────┘
```

---

## 📋 Implementation Timeline

### Week 1, Day 1-2: Complete Parser (2-3 hours) ⚠️ DO THIS FIRST

See `START_HERE.md` for detailed tasks:
1. Fix CREATE TABLE panic (30 min)
2. Implement UPDATE (45 min)
3. Implement DELETE (30 min)
4. Implement DROP TABLE (20 min)

**Target**: 15/15 parser tests passing

---

### Week 1, Day 3-5: Implement Dispatcher (6-8 hours)

#### Day 3: Core Dispatcher (4 hours)

**What Dispatcher Does**:
- Takes SQL string as input
- Calls lexer → parser → validator
- Routes parsed AST to appropriate handler
- Tracks query statistics

**Files to Create/Modify**:

1. **`internal/dispatcher/validator.go`** (NEW - 2 hours)
```go
// Semantic validation after parsing
type Validator struct {}

func (v *Validator) ValidateStatement(stmt parser.Statement) error {
    // Check:
    // - Required clauses present
    // - Column counts match
    // - Duplicate column names
    // (Schema validation comes later with catalog)
}
```

2. **`internal/dispatcher/dispatcher.go`** (MODIFY - 2 hours)
```go
// Complete the DispatchQuery method
func (d *Dispatcher) DispatchQuery(ctx context.Context, sql string, queryCtx *QueryContext) (*QueryResult, error) {
    // Step 1: Lexer (✅ already works)
    // Step 2: Parser (✅ will work after you finish it)
    // Step 3: Validator (NEW - you implement this)
    // Step 4: Route to handler (placeholder for now)
    // Step 5: Track statistics (✅ already structured)
}

// Add this helper
func (d *Dispatcher) determineQueryType(stmt parser.Statement) QueryType {
    switch stmt.(type) {
    case *parser.SelectStatement:
        return QueryTypeSelect
    // ... etc
    }
}
```

#### Day 4-5: Integration Testing (4 hours)

**File**: `tests/integration/dispatcher_test.go` (NEW)

```go
func TestDispatcherBasicFlow(t *testing.T) {
    // Test SQL → Dispatcher → Parsing → Validation
    // For each query type: SELECT, INSERT, UPDATE, DELETE, DDL
}

func TestDispatcherValidation(t *testing.T) {
    // Test invalid SQL gets caught
}

func TestDispatcherStatistics(t *testing.T) {
    // Test query counting works
}
```

---

## 🎯 Success Criteria

### End of Week 1 (5 days total):

**Parser (Days 1-2)** ✅:
- [x] 15/15 tests passing
- [x] All statement types implemented
- [x] >90% code coverage

**Dispatcher (Days 3-5)** ✅:
- [ ] Query routing complete
- [ ] Validation layer implemented
- [ ] Integration tests passing
- [ ] SQL → AST → Routing works end-to-end

---

## 📊 Module Status Reference

| Module | Status | Tests | Next Action |
|--------|--------|-------|-------------|
| Lexer | ✅ Complete | 100% | None |
| Parser | 🚧 66% | 10/15 | Finish (2-3h) |
| Dispatcher | 🎯 Next | 0% | Implement (6-8h) |
| Catalog | ⏸️ Future | 0% | Week 2 |
| Executor | ⏸️ Future | 0% | Week 3 |

---

## 🚀 Quick Start

### Step 1: Finish Parser (TODAY)
```bash
# Open START_HERE.md and follow Tasks 1-4
# Estimated: 2-3 hours
# Result: 15/15 parser tests passing
```

### Step 2: Implement Dispatcher (DAYS 3-5)
```bash
# Day 3: Create validator.go and complete dispatcher.go
# Day 4-5: Write integration tests
# Result: End-to-end SQL parsing and routing
```

### Step 3: Week 2 - Catalog Manager
```bash
# Persistent schema storage
# Required before query execution
```

---

## 📚 Policy Compliance for Dispatcher

**Remember**: Each module must implement:
- ✅ Algorithms from `internal/dispatcher/ALGO.md`
- ✅ Data structures from `internal/dispatcher/DS.md`
- ✅ Solutions to problems in `internal/dispatcher/PROBLEMS.md`

Check these files exist:
```bash
ls -la internal/dispatcher/ALGO.md
ls -la internal/dispatcher/DS.md
ls -la internal/dispatcher/PROBLEMS.md
```

If they don't exist, create them following the pattern from parser module!

---

## 🎯 Summary

**NEXT MODULE TO IMPLEMENT: DISPATCHER**

**Timeline**:
- Days 1-2 (2-3h): Finish parser → 15/15 tests ✅
- Days 3-5 (6-8h): Implement dispatcher → End-to-end flow ✅

**Why This Order?**:
1. Parser almost done (10/15 tests passing)
2. Dispatcher depends on parser
3. Both are pure coordination logic (no storage needed)
4. Enables complete SQL → AST → Routing pipeline
5. Sets foundation for Week 2 (Catalog + Execution)

**Start Now**: Open `START_HERE.md` and begin Task 1! 🚀
