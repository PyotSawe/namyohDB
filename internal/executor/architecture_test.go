package executor

import (
	"testing"
)

// TestResultBuilder tests the Result Set Builder component
func TestResultBuilder(t *testing.T) {
	schema := &TupleSchema{
		Columns: []ColumnInfo{
			{Name: "id", Type: TypeInt},
			{Name: "name", Type: TypeString},
		},
	}

	builder := NewResultBuilder(schema)

	// Test initial state
	if builder.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", builder.RowCount())
	}

	// Add tuple
	tuple := &Tuple{
		Values: []interface{}{1, "Alice"},
		Schema: schema,
	}

	if err := builder.AddTuple(tuple); err != nil {
		t.Errorf("failed to add tuple: %v", err)
	}

	if builder.RowCount() != 1 {
		t.Errorf("expected 1 row, got %d", builder.RowCount())
	}

	// Build result set
	result := builder.Build()
	if result.RowCount() != 1 {
		t.Errorf("expected 1 row in result, got %d", result.RowCount())
	}

	// Test reset
	builder.Reset()
	if builder.RowCount() != 0 {
		t.Errorf("expected 0 rows after reset, got %d", builder.RowCount())
	}
}

// TestResultSetIterator tests the iterator pattern
func TestResultSetIterator(t *testing.T) {
	schema := &TupleSchema{
		Columns: []ColumnInfo{
			{Name: "id", Type: TypeInt},
		},
	}

	resultSet := &ResultSet{
		Schema: schema,
		Tuples: []*Tuple{
			{Values: []interface{}{1}, Schema: schema},
			{Values: []interface{}{2}, Schema: schema},
			{Values: []interface{}{3}, Schema: schema},
		},
	}

	iterator := NewResultSetIterator(resultSet)

	// Test iteration
	count := 0
	for iterator.HasNext() {
		tuple, err := iterator.Next()
		if err != nil {
			t.Errorf("iteration error: %v", err)
		}
		if tuple == nil {
			break
		}
		count++
	}

	if count != 3 {
		t.Errorf("expected 3 tuples, iterated %d", count)
	}

	// Test reset
	iterator.Reset()
	if iterator.Position() != 0 {
		t.Errorf("expected position 0 after reset, got %d", iterator.Position())
	}
}

// TestSchemaManager tests the Schema Manager component
func TestSchemaManager(t *testing.T) {
	sm := NewSchemaManager()

	// Create a table schema
	tableSchema := &TableSchema{
		TableName: "users",
		Columns: []ColumnInfo{
			{Name: "id", Type: TypeInt, Nullable: false},
			{Name: "name", Type: TypeString, Nullable: false},
			{Name: "email", Type: TypeString, Nullable: true},
		},
		PrimaryKey: []string{"id"},
	}

	// Register schema
	if err := sm.RegisterSchema(tableSchema); err != nil {
		t.Errorf("failed to register schema: %v", err)
	}

	// Get schema
	retrieved, err := sm.GetSchema("users")
	if err != nil {
		t.Errorf("failed to get schema: %v", err)
	}

	if retrieved.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", retrieved.TableName)
	}

	// List schemas
	tables := sm.ListSchemas()
	if len(tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(tables))
	}

	// Drop schema
	if err := sm.DropSchema("users"); err != nil {
		t.Errorf("failed to drop schema: %v", err)
	}

	// Verify dropped
	if _, err := sm.GetSchema("users"); err == nil {
		t.Error("expected error when getting dropped schema")
	}
}

// TestCatalogManager tests the Catalog Manager component
func TestCatalogManager(t *testing.T) {
	sm := NewSchemaManager()
	cm := NewCatalogManager(sm)

	// Register schema first
	tableSchema := &TableSchema{
		TableName: "products",
		Columns: []ColumnInfo{
			{Name: "id", Type: TypeInt},
			{Name: "name", Type: TypeString},
		},
	}
	sm.RegisterSchema(tableSchema)

	// Create table entry
	tableEntry := &TableCatalogEntry{
		TableName:  "products",
		TableID:    1,
		SchemaName: "public",
		Owner:      "admin",
		RowCount:   0,
		PageCount:  0,
	}

	if err := cm.CreateTable(tableEntry); err != nil {
		t.Errorf("failed to create table: %v", err)
	}

	// Get table
	retrieved, err := cm.GetTable("products")
	if err != nil {
		t.Errorf("failed to get table: %v", err)
	}

	if retrieved.TableName != "products" {
		t.Errorf("expected table name 'products', got '%s'", retrieved.TableName)
	}

	// Update row count
	if err := cm.UpdateRowCount("products", 100); err != nil {
		t.Errorf("failed to update row count: %v", err)
	}

	// Verify update
	retrieved, _ = cm.GetTable("products")
	if retrieved.RowCount != 100 {
		t.Errorf("expected row count 100, got %d", retrieved.RowCount)
	}

	// List tables
	tables := cm.ListTables()
	if len(tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(tables))
	}
}

// TestCursorManager tests the Cursor Manager component
func TestCursorManager(t *testing.T) {
	cm := NewCursorManager()

	schema := &TupleSchema{
		Columns: []ColumnInfo{
			{Name: "id", Type: TypeInt},
		},
	}

	resultSet := &ResultSet{
		Schema: schema,
		Tuples: []*Tuple{
			{Values: []interface{}{1}, Schema: schema},
			{Values: []interface{}{2}, Schema: schema},
			{Values: []interface{}{3}, Schema: schema},
		},
	}

	// Open cursor
	cursor, err := cm.OpenCursor("test_cursor", resultSet, true, false)
	if err != nil {
		t.Errorf("failed to open cursor: %v", err)
	}

	if !cursor.IsOpen {
		t.Error("cursor should be open")
	}

	// Fetch next
	tuples, err := cursor.Fetch(FetchNext, 2)
	if err != nil {
		t.Errorf("fetch failed: %v", err)
	}

	if len(tuples) != 2 {
		t.Errorf("expected 2 tuples, got %d", len(tuples))
	}

	// Get cursor
	retrieved, err := cm.GetCursor("test_cursor")
	if err != nil {
		t.Errorf("failed to get cursor: %v", err)
	}

	if retrieved.Name != "test_cursor" {
		t.Errorf("expected cursor name 'test_cursor', got '%s'", retrieved.Name)
	}

	// Close cursor
	if err := cm.CloseCursor("test_cursor"); err != nil {
		t.Errorf("failed to close cursor: %v", err)
	}

	// Verify closed
	if _, err := cm.GetCursor("test_cursor"); err == nil {
		t.Error("expected error when getting closed cursor")
	}
}

// TestLockManager tests the Lock Manager component
func TestLockManager(t *testing.T) {
	lm := NewLockManager()

	txnID := uint64(1)
	tableName := "test_table"

	// Acquire shared lock
	if err := lm.AcquireTableLock(txnID, tableName, SharedLock); err != nil {
		t.Errorf("failed to acquire shared lock: %v", err)
	}

	// Acquire another shared lock (should succeed - compatible)
	txnID2 := uint64(2)
	if err := lm.AcquireTableLock(txnID2, tableName, SharedLock); err != nil {
		t.Errorf("failed to acquire second shared lock: %v", err)
	}

	// Release first lock
	if err := lm.ReleaseTableLock(txnID, tableName); err != nil {
		t.Errorf("failed to release lock: %v", err)
	}

	// Release second lock
	if err := lm.ReleaseTableLock(txnID2, tableName); err != nil {
		t.Errorf("failed to release second lock: %v", err)
	}

	// Test page lock
	pageID := uint64(1)
	if err := lm.AcquirePageLock(txnID, tableName, pageID, ExclusiveLock); err != nil {
		t.Errorf("failed to acquire page lock: %v", err)
	}

	if err := lm.ReleasePageLock(txnID, tableName, pageID); err != nil {
		t.Errorf("failed to release page lock: %v", err)
	}

	// Test row lock
	slotID := uint16(5)
	if err := lm.AcquireRowLock(txnID, tableName, pageID, slotID, ExclusiveLock); err != nil {
		t.Errorf("failed to acquire row lock: %v", err)
	}

	if err := lm.ReleaseRowLock(txnID, tableName, pageID, slotID); err != nil {
		t.Errorf("failed to release row lock: %v", err)
	}
}

// TestTransactionExecutor tests the Transaction Executor component
func TestTransactionExecutor(t *testing.T) {
	// Create dependencies (mocked for now)
	executor := &Executor{
		statistics: NewExecutionStatistics(),
		config:     DefaultExecutorConfig(),
	}
	lockManager := NewLockManager()

	te := NewTransactionExecutor(executor, lockManager)

	// Begin transaction
	txn, err := te.BeginTransaction(ReadCommitted)
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}

	if txn.State != TxnActive {
		t.Errorf("expected transaction state ACTIVE, got %s", txn.State)
	}

	if txn.IsolationLevel != ReadCommitted {
		t.Errorf("expected isolation level READ_COMMITTED, got %s", txn.IsolationLevel)
	}

	// Get transaction
	retrieved, err := te.GetTransaction(txn.ID)
	if err != nil {
		t.Errorf("failed to get transaction: %v", err)
	}

	if retrieved.ID != txn.ID {
		t.Errorf("expected transaction ID %d, got %d", txn.ID, retrieved.ID)
	}

	// Create savepoint
	if err := te.CreateSavepoint(txn.ID, "sp1"); err != nil {
		t.Errorf("failed to create savepoint: %v", err)
	}

	// List active transactions
	activeTxns := te.ListActiveTransactions()
	if len(activeTxns) != 1 {
		t.Errorf("expected 1 active transaction, got %d", len(activeTxns))
	}

	// Commit transaction
	if err := te.CommitTransaction(txn.ID); err != nil {
		t.Errorf("failed to commit transaction: %v", err)
	}

	// Verify committed
	if txn.State != TxnCommitted {
		t.Errorf("expected transaction state COMMITTED, got %s", txn.State)
	}

	// Verify removed from active list
	activeTxns = te.ListActiveTransactions()
	if len(activeTxns) != 0 {
		t.Errorf("expected 0 active transactions after commit, got %d", len(activeTxns))
	}
}

// TestTransactionRollback tests transaction rollback
func TestTransactionRollback(t *testing.T) {
	executor := &Executor{
		statistics: NewExecutionStatistics(),
		config:     DefaultExecutorConfig(),
	}
	lockManager := NewLockManager()

	te := NewTransactionExecutor(executor, lockManager)

	// Begin transaction
	txn, err := te.BeginTransaction(RepeatableRead)
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}

	// Rollback transaction
	if err := te.RollbackTransaction(txn.ID); err != nil {
		t.Errorf("failed to rollback transaction: %v", err)
	}

	// Verify aborted
	if txn.State != TxnAborted {
		t.Errorf("expected transaction state ABORTED, got %s", txn.State)
	}
}

// TestIsolationLevels tests isolation level configuration
func TestIsolationLevels(t *testing.T) {
	executor := &Executor{
		statistics: NewExecutionStatistics(),
		config:     DefaultExecutorConfig(),
	}
	lockManager := NewLockManager()

	te := NewTransactionExecutor(executor, lockManager)

	// Test default level
	if te.GetIsolationLevel() != ReadCommitted {
		t.Errorf("expected default READ_COMMITTED, got %s", te.GetIsolationLevel())
	}

	// Set new level
	te.SetIsolationLevel(Serializable)

	if te.GetIsolationLevel() != Serializable {
		t.Errorf("expected SERIALIZABLE, got %s", te.GetIsolationLevel())
	}

	// Begin transaction with specific level
	txn, _ := te.BeginTransaction(ReadUncommitted)

	if txn.IsolationLevel != ReadUncommitted {
		t.Errorf("expected READ_UNCOMMITTED, got %s", txn.IsolationLevel)
	}

	te.CommitTransaction(txn.ID)
}
