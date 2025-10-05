# NamyohDB: Implementation Roadmap

## Strategic Implementation Order

This roadmap prioritizes modules that:
1. Enable end-to-end functionality quickly
2. Allow incremental testing at each step
3. Build on existing working components
4. Follow natural dependencies between layers
5. Provide immediate value for debugging and validation

---

## Phase 1: Complete the Parsing Layer (IMMEDIATE PRIORITY)

### Module 1.1: Complete SQL Parser (`internal/parser`) 🎯 **START HERE**

**Why First:**
- Lexer already works (✅ IMPLEMENTED)
- AST structures already defined (✅ IMPLEMENTED)
- Completes the SQL compilation pipeline up to semantic analysis
- Enables you to parse complete SQL statements for testing
- No storage dependencies - pure transformation logic

**What to Implement:**
```go
// internal/parser/parser.go

// Priority 1: SELECT statement parsing
func (p *Parser) parseSelectStatement() (*SelectStatement, error)
func (p *Parser) parseSelectColumns() ([]Expression, error)
func (p *Parser) parseFromClause() (*TableReference, error)
func (p *Parser) parseWhereClause() (Expression, error)

// Priority 2: Simple DML (for testing)
func (p *Parser) parseInsertStatement() (*InsertStatement, error)
func (p *Parser) parseDeleteStatement() (*DeleteStatement, error)

// Priority 3: DDL (for schema management)
func (p *Parser) parseCreateTableStatement() (*CreateTableStatement, error)
func (p *Parser) parseDropTableStatement() (*DropTableStatement, error)
```

**Testing Strategy:**
```go
// tests/unit/parser_test.go
func TestParseSelect(t *testing.T) {
    input := "SELECT name, age FROM users WHERE id = 42"
    parser := NewParser(input)
    stmt, err := parser.Parse()
    // Validate AST structure
}

func TestParseInsert(t *testing.T) {
    input := "INSERT INTO users (name, age) VALUES ('Alice', 25)"
    // Validate AST
}

func TestParseCreateTable(t *testing.T) {
    input := "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)"
    // Validate schema definition
}
```

**Benefits:**
- ✅ Can parse complete SQL statements
- ✅ Enables end-to-end testing of SQL → AST
- ✅ Foundation for optimizer and executor
- ✅ Immediate validation with comprehensive tests

**Estimated Effort:** 3-5 days
**Complexity:** Medium (recursive descent parsing patterns)

---

## Phase 2: Schema Management (HIGH PRIORITY)

### Module 2.1: Catalog Manager (`internal/storage/catalog.go`) 🎯 **SECOND**

**Why Second:**
- Parser can now generate CREATE TABLE ASTs
- Storage engine already works (✅ IMPLEMENTED)
- Enables persistent schema storage
- Required before any query execution
- Natural bridge between parsing and storage

**What to Implement:**
```go
// internal/storage/catalog.go

type CatalogManager interface {
    CreateTable(schema *TableSchema) error
    DropTable(tableName string) error
    GetTable(tableName string) (*TableMetadata, error)
    ListTables() ([]string, error)
}

type TableMetadata struct {
    ID          TableID
    Name        string
    Columns     []ColumnDefinition
    PrimaryKey  []string
    RootPageID  PageID
    RecordCount int64
}

// Store catalog in special system pages (like SQLite's sqlite_master)
const CATALOG_ROOT_PAGE PageID = 1
```

**Testing Strategy:**
```go
// tests/unit/catalog_test.go
func TestCreateTable(t *testing.T) {
    catalog := NewCatalogManager(storageEngine)
    
    schema := &TableSchema{
        Name: "users",
        Columns: []ColumnDefinition{
            {Name: "id", Type: INTEGER, PrimaryKey: true},
            {Name: "name", Type: TEXT, Nullable: false},
        },
    }
    
    err := catalog.CreateTable(schema)
    assert.NoError(t, err)
    
    // Verify table metadata persisted
    meta, err := catalog.GetTable("users")
    assert.Equal(t, "users", meta.Name)
}
```

**Benefits:**
- ✅ Persistent schema storage
- ✅ Foundation for table operations
- ✅ Enables CREATE/DROP TABLE execution
- ✅ Can now test schema persistence across restarts

**Estimated Effort:** 2-3 days
**Complexity:** Medium (encoding/decoding schema to pages)

---

### Module 2.2: Record Manager (`internal/storage/record.go`) 🎯 **THIRD**

**Why Third:**
- Catalog defines table schemas
- Need to serialize/deserialize records according to schema
- Required before implementing INSERT/SELECT
- Uses existing page structure

**What to Implement:**
```go
// internal/storage/record.go

type RecordManager interface {
    InsertRecord(tableID TableID, record *Record) (RecordID, error)
    GetRecord(tableID TableID, recordID RecordID) (*Record, error)
    UpdateRecord(tableID TableID, recordID RecordID, record *Record) error
    DeleteRecord(tableID TableID, recordID RecordID) error
    ScanRecords(tableID TableID) (RecordIterator, error)
}

type Record struct {
    ID     RecordID
    Fields map[string]interface{}
}

type RecordID struct {
    PageID PageID
    SlotID uint16
}

// Record encoding (SQLite3-style variable length)
func encodeRecord(record *Record, schema *TableSchema) ([]byte, error)
func decodeRecord(data []byte, schema *TableSchema) (*Record, error)
```

**Testing Strategy:**
```go
// tests/unit/record_test.go
func TestInsertAndRetrieveRecord(t *testing.T) {
    recordMgr := NewRecordManager(storageEngine, catalog)
    
    record := &Record{
        Fields: map[string]interface{}{
            "id":   1,
            "name": "Alice",
            "age":  25,
        },
    }
    
    recordID, err := recordMgr.InsertRecord(tableID, record)
    assert.NoError(t, err)
    
    retrieved, err := recordMgr.GetRecord(tableID, recordID)
    assert.Equal(t, record.Fields, retrieved.Fields)
}
```

**Benefits:**
- ✅ Can store and retrieve actual data
- ✅ Foundation for all DML operations
- ✅ Enables basic INSERT/SELECT testing
- ✅ Can verify data persistence

**Estimated Effort:** 3-4 days
**Complexity:** Medium-High (record encoding/decoding logic)

---

## Phase 3: Basic Query Execution (CRITICAL PATH)

### Module 3.1: Simple Executor (`internal/executor/executor.go`) 🎯 **FOURTH**

**Why Fourth:**
- Parser generates ASTs (✅ COMPLETE)
- Catalog provides schema metadata (✅ COMPLETE from Phase 2)
- Record manager handles storage (✅ COMPLETE from Phase 2)
- Now ready for end-to-end query execution
- Start simple: no optimization, just basic execution

**What to Implement:**
```go
// internal/executor/executor.go

type Executor interface {
    ExecuteStatement(stmt Statement) (ResultSet, error)
}

// Start with simplest operations
type SimpleExecutor struct {
    catalog    CatalogManager
    recordMgr  RecordManager
    storage    StorageEngine
}

// Priority 1: SELECT (read-only, no optimization)
func (e *SimpleExecutor) executeSelect(stmt *SelectStatement) (ResultSet, error) {
    // 1. Get table metadata from catalog
    // 2. Scan all records (no indexes yet)
    // 3. Apply WHERE filter
    // 4. Project selected columns
    // 5. Return results
}

// Priority 2: INSERT (write path)
func (e *SimpleExecutor) executeInsert(stmt *InsertStatement) error {
    // 1. Validate schema
    // 2. Create record from values
    // 3. Insert via record manager
}

// Priority 3: DELETE (write path)
func (e *SimpleExecutor) executeDelete(stmt *DeleteStatement) error {
    // 1. Scan and find matching records
    // 2. Delete via record manager
}
```

**Testing Strategy:**
```go
// tests/integration/query_execution_test.go
func TestEndToEndQuery(t *testing.T) {
    // Setup
    executor := NewSimpleExecutor(catalog, recordMgr, storage)
    
    // Create table
    createSQL := "CREATE TABLE users (id INTEGER, name TEXT, age INTEGER)"
    _, err := executor.ExecuteStatement(parseSQL(createSQL))
    assert.NoError(t, err)
    
    // Insert data
    insertSQL := "INSERT INTO users VALUES (1, 'Alice', 25)"
    _, err = executor.ExecuteStatement(parseSQL(insertSQL))
    assert.NoError(t, err)
    
    // Query data
    selectSQL := "SELECT name, age FROM users WHERE id = 1"
    results, err := executor.ExecuteStatement(parseSQL(selectSQL))
    assert.NoError(t, err)
    
    // Verify results
    assert.Equal(t, 1, results.RowCount())
    assert.Equal(t, "Alice", results.Rows[0]["name"])
    assert.Equal(t, 25, results.Rows[0]["age"])
}
```

**Benefits:**
- ✅ **END-TO-END FUNCTIONALITY!** SQL → Results
- ✅ Can run real SQL queries
- ✅ Foundation for all future optimizations
- ✅ Immediate practical value
- ✅ Can benchmark query performance

**Estimated Effort:** 4-5 days
**Complexity:** Medium-High (integrating all components)

---

## Phase 4: Transaction Support (ROBUSTNESS)

### Module 4.1: Transaction Manager (`internal/transaction/transaction.go`) 🎯 **FIFTH**

**Why Fifth:**
- Basic queries work (✅ from Phase 3)
- Now add ACID guarantees
- Required for data integrity
- Enables concurrent access patterns

**What to Implement:**
```go
// internal/transaction/transaction.go

type TransactionManager interface {
    Begin() (Transaction, error)
    Commit(tx Transaction) error
    Rollback(tx Transaction) error
}

type Transaction struct {
    ID      TransactionID
    State   TransactionState
    Locks   []Lock
    Changes []Change  // For rollback
}

// Start simple: single-writer model (like SQLite)
// Later: add MVCC for concurrent readers
```

**Testing Strategy:**
```go
func TestTransactionCommit(t *testing.T) {
    tx, _ := txMgr.Begin()
    executor.ExecuteInTransaction(tx, insertSQL)
    txMgr.Commit(tx)
    // Verify data persisted
}

func TestTransactionRollback(t *testing.T) {
    tx, _ := txMgr.Begin()
    executor.ExecuteInTransaction(tx, insertSQL)
    txMgr.Rollback(tx)
    // Verify no data persisted
}
```

**Benefits:**
- ✅ ACID compliance
- ✅ Data integrity guarantees
- ✅ Foundation for concurrent access
- ✅ Production-ready reliability

**Estimated Effort:** 5-7 days
**Complexity:** High (transaction semantics, undo logs)

---

### Module 4.2: Write-Ahead Log (`internal/wal/wal.go`) 🎯 **SIXTH**

**Why Sixth:**
- Transactions need durability
- WAL enables fast commits
- Required for crash recovery
- SQLite3's proven approach

**What to Implement:**
```go
// internal/wal/wal.go

type WALManager interface {
    WriteRecord(record *WALRecord) (LSN, error)
    Checkpoint() error
    Recover() error
}

// Simplified WAL (like SQLite)
// - Append-only log file
// - Checkpoint copies to main database
// - Recovery replays uncommitted transactions
```

**Benefits:**
- ✅ Crash recovery
- ✅ Durability guarantees
- ✅ Fast commit performance
- ✅ Production reliability

**Estimated Effort:** 4-5 days
**Complexity:** Medium-High (log management, recovery)

---

## Phase 5: Performance Optimization (PERFORMANCE)

### Module 5.1: B-Tree Index Manager (`internal/btree/btree.go`) 🎯 **SEVENTH**

**Why Seventh:**
- Queries work but are slow (full table scans)
- Indexes provide O(log n) access
- Required for production performance
- Natural next optimization

**What to Implement:**
```go
// internal/btree/btree.go

type BTreeManager interface {
    CreateIndex(tableName, columnName string) error
    InsertKey(indexID IndexID, key []byte, value RecordID) error
    SearchKey(indexID IndexID, key []byte) (RecordID, error)
    RangeSearch(indexID IndexID, start, end []byte) ([]RecordID, error)
}

// B+ tree implementation
// - Internal nodes: keys + child pointers
// - Leaf nodes: keys + record IDs
// - Linked leaves for range scans
```

**Testing Strategy:**
```go
func TestIndexedQuery(t *testing.T) {
    // Create table and insert 10000 records
    
    // Query without index (baseline)
    start := time.Now()
    results := executor.Execute("SELECT * FROM users WHERE id = 5000")
    noIndexTime := time.Since(start)
    
    // Create index
    executor.Execute("CREATE INDEX idx_id ON users(id)")
    
    // Query with index
    start = time.Now()
    results = executor.Execute("SELECT * FROM users WHERE id = 5000")
    indexTime := time.Since(start)
    
    // Verify speedup
    assert.Less(t, indexTime, noIndexTime/10)
}
```

**Benefits:**
- ✅ Fast lookups: O(log n) vs O(n)
- ✅ Range query support
- ✅ Production performance
- ✅ Enables larger datasets

**Estimated Effort:** 7-10 days
**Complexity:** High (B-tree algorithms, page splits)

---

### Module 5.2: Query Optimizer (`internal/optimizer/optimizer.go`) 🎯 **EIGHTH**

**Why Eighth:**
- Indexes available (✅ from Phase 5.1)
- Can now choose optimal access paths
- Significant performance gains
- Natural progression

**What to Implement:**
```go
// internal/optimizer/optimizer.go

type Optimizer interface {
    Optimize(stmt Statement) (*ExecutionPlan, error)
}

// Start with rule-based optimization
type RuleBasedOptimizer struct {
    catalog CatalogManager
    stats   *Statistics
}

// Optimization rules:
// 1. Index selection for WHERE clauses
// 2. Predicate pushdown
// 3. Constant folding
// 4. Expression simplification
```

**Benefits:**
- ✅ Automatic index selection
- ✅ Faster query execution
- ✅ Better resource usage
- ✅ Handles complex queries efficiently

**Estimated Effort:** 5-7 days
**Complexity:** Medium-High (cost estimation, plan selection)

---

## Summary Timeline & Dependencies

```
Week 1-2: Parser Completion
    ├── Complete SELECT parsing
    ├── Add INSERT/DELETE parsing  
    └── Implement CREATE TABLE parsing
    
Week 3-4: Schema Management
    ├── Catalog Manager (persistent schemas)
    └── Record Manager (data serialization)
    
Week 5-6: Basic Query Execution ⭐ MILESTONE
    └── Simple Executor (end-to-end queries)
    
Week 7-9: Transaction Support
    ├── Transaction Manager (ACID)
    └── WAL Manager (durability)
    
Week 10-12: Performance Optimization
    ├── B-Tree Indexes (fast lookups)
    └── Query Optimizer (intelligent execution)
```

## Key Milestones

### 🎯 Milestone 1: "Parse Complete" (Week 2)
- Can parse all basic SQL statements
- Full AST generation tested
- Ready for semantic analysis

### 🎯 Milestone 2: "Storage Complete" (Week 4)
- Schemas persist across restarts
- Records stored and retrieved
- Basic data operations work

### 🎯 Milestone 3: "Query Execution" (Week 6) ⭐ **MAJOR MILESTONE**
- **End-to-end SQL query execution**
- Can run: CREATE TABLE, INSERT, SELECT, DELETE
- Real database functionality!

### 🎯 Milestone 4: "ACID Compliant" (Week 9)
- Transactions work correctly
- Crash recovery functional
- Production-ready reliability

### 🎯 Milestone 5: "Performance Ready" (Week 12)
- Indexes accelerate queries
- Optimizer selects best plans
- Production-ready performance

---

## Why This Order?

### 1. **Fastest Path to Value**
- By Week 6, you have a working database that can execute SQL queries
- Each phase builds working functionality that can be tested independently
- Immediate feedback on design decisions

### 2. **Incremental Testing**
- Parser tests don't need storage
- Storage tests don't need executor
- Executor integrates everything
- Each layer validates the previous layer

### 3. **Natural Dependencies**
```
Parser → Catalog → Record Manager → Executor
                                       ↓
                                  Transaction → WAL
                                       ↓
                                   Indexes → Optimizer
```

### 4. **Risk Mitigation**
- Start with pure logic (parser) - low risk
- Add storage gradually - medium risk
- Tackle complex areas (transactions, B-trees) last when foundation is solid

### 5. **Learning Progression**
- Start with familiar concepts (parsing)
- Build to data structures (records, pages)
- Progress to algorithms (B-trees, optimization)
- Master complex systems (transactions, concurrency)

---

## Testing Strategy per Phase

### Phase 1: Parser
```bash
# Unit tests for each statement type
go test ./internal/parser -v

# Test coverage
go test ./internal/parser -cover
# Target: 90%+ coverage
```

### Phase 2: Schema Management
```bash
# Unit tests for catalog and records
go test ./internal/storage -v

# Integration tests
go test ./tests/integration/schema_test.go -v
# Verify persistence across restarts
```

### Phase 3: Query Execution
```bash
# Integration tests for end-to-end queries
go test ./tests/integration/query_test.go -v

# Performance benchmarks
go test ./tests/integration -bench=. -benchtime=10s
```

### Phase 4: Transactions
```bash
# Transaction correctness tests
go test ./internal/transaction -v

# Concurrency tests
go test ./tests/integration/concurrent_test.go -v

# Crash recovery tests
go test ./tests/integration/recovery_test.go -v
```

### Phase 5: Optimization
```bash
# Performance comparison tests
go test ./tests/performance -v

# Benchmark suite
go test ./tests/performance -bench=. -benchmem
```

---

## Next Immediate Steps

### TODAY: Start with Parser

1. **Create parser test file**:
```bash
touch tests/unit/parser_test.go
```

2. **Implement SELECT parser**:
```bash
# Edit internal/parser/parser.go
# Start with simplest SELECT: "SELECT * FROM table"
```

3. **Test immediately**:
```bash
go test ./internal/parser -v -run TestParseSimpleSelect
```

4. **Iterate**:
- Add WHERE clause parsing
- Add column list parsing
- Add ORDER BY, LIMIT

This roadmap gives you a clear path from current state to production-ready database, with testable milestones at every step! 🚀