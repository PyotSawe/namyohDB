// Package executor - Transaction Executor component
// Manages transaction execution and coordination
package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"relational-db/internal/optimizer"
)

// TransactionExecutor executes transactions with ACID guarantees
// Architecture: Part of Execution Engine Layer, coordinates with Lock Manager
type TransactionExecutor struct {
	// Core components
	queryExecutor *Executor
	lockManager   *LockManager

	// Transaction management
	activeTransactions map[uint64]*Transaction
	nextTxnID          uint64

	// Configuration
	isolationLevel IsolationLevel
	timeout        time.Duration

	mutex sync.RWMutex
}

// Transaction represents an active transaction
type Transaction struct {
	ID             uint64
	State          TransactionState
	IsolationLevel IsolationLevel
	StartTime      time.Time
	EndTime        time.Time

	// Transaction context
	Context context.Context
	Cancel  context.CancelFunc

	// Operations
	Operations []*TransactionOperation

	// Locks
	AcquiredLocks []*Lock

	// Savepoints
	Savepoints map[string]*Savepoint

	// Statistics
	RowsRead     uint64
	RowsModified uint64

	mutex sync.RWMutex
}

// TransactionState defines transaction states
type TransactionState int

const (
	TxnActive TransactionState = iota
	TxnPreparing
	TxnCommitting
	TxnCommitted
	TxnAborting
	TxnAborted
)

// IsolationLevel defines transaction isolation levels
type IsolationLevel int

const (
	ReadUncommitted IsolationLevel = iota
	ReadCommitted
	RepeatableRead
	Serializable
)

// TransactionOperation represents an operation within a transaction
type TransactionOperation struct {
	Type      OperationType
	TableName string
	Plan      *optimizer.QueryPlan
	Timestamp time.Time
	Result    *ResultSet
}

// OperationType defines types of operations
type OperationType int

const (
	SelectOp OperationType = iota
	InsertOp
	UpdateOp
	DeleteOp
)

// Savepoint represents a transaction savepoint
type Savepoint struct {
	Name      string
	TxnID     uint64
	Timestamp time.Time
	Position  int // Position in operation list
}

// NewTransactionExecutor creates a new transaction executor
func NewTransactionExecutor(queryExecutor *Executor, lockManager *LockManager) *TransactionExecutor {
	return &TransactionExecutor{
		queryExecutor:      queryExecutor,
		lockManager:        lockManager,
		activeTransactions: make(map[uint64]*Transaction),
		nextTxnID:          1,
		isolationLevel:     ReadCommitted,
		timeout:            30 * time.Second,
	}
}

// BeginTransaction starts a new transaction
func (te *TransactionExecutor) BeginTransaction(isolationLevel IsolationLevel) (*Transaction, error) {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	txnID := te.nextTxnID
	te.nextTxnID++

	ctx, cancel := context.WithTimeout(context.Background(), te.timeout)

	txn := &Transaction{
		ID:             txnID,
		State:          TxnActive,
		IsolationLevel: isolationLevel,
		StartTime:      time.Now(),
		Context:        ctx,
		Cancel:         cancel,
		Operations:     make([]*TransactionOperation, 0),
		AcquiredLocks:  make([]*Lock, 0),
		Savepoints:     make(map[string]*Savepoint),
	}

	te.activeTransactions[txnID] = txn
	return txn, nil
}

// CommitTransaction commits a transaction
func (te *TransactionExecutor) CommitTransaction(txnID uint64) error {
	te.mutex.Lock()
	txn, exists := te.activeTransactions[txnID]
	if !exists {
		te.mutex.Unlock()
		return fmt.Errorf("transaction %d not found", txnID)
	}
	te.mutex.Unlock()

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	if txn.State != TxnActive {
		return fmt.Errorf("transaction %d is not active (state: %d)", txnID, txn.State)
	}

	// Change state to committing
	txn.State = TxnCommitting

	// TODO: Write WAL records for all operations

	// TODO: Flush WAL to disk

	// Change state to committed
	txn.State = TxnCommitted
	txn.EndTime = time.Now()

	// Release all locks
	if err := te.lockManager.ReleaseAllLocks(txnID); err != nil {
		return fmt.Errorf("failed to release locks: %w", err)
	}

	// Remove from active transactions
	te.mutex.Lock()
	delete(te.activeTransactions, txnID)
	te.mutex.Unlock()

	// Cancel context
	txn.Cancel()

	return nil
}

// RollbackTransaction aborts a transaction
func (te *TransactionExecutor) RollbackTransaction(txnID uint64) error {
	te.mutex.Lock()
	txn, exists := te.activeTransactions[txnID]
	if !exists {
		te.mutex.Unlock()
		return fmt.Errorf("transaction %d not found", txnID)
	}
	te.mutex.Unlock()

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	if txn.State != TxnActive {
		return fmt.Errorf("transaction %d is not active", txnID)
	}

	// Change state to aborting
	txn.State = TxnAborting

	// TODO: Undo all operations using WAL

	// Change state to aborted
	txn.State = TxnAborted
	txn.EndTime = time.Now()

	// Release all locks
	if err := te.lockManager.ReleaseAllLocks(txnID); err != nil {
		return fmt.Errorf("failed to release locks: %w", err)
	}

	// Remove from active transactions
	te.mutex.Lock()
	delete(te.activeTransactions, txnID)
	te.mutex.Unlock()

	// Cancel context
	txn.Cancel()

	return nil
}

// ExecuteInTransaction executes a query plan within a transaction
func (te *TransactionExecutor) ExecuteInTransaction(txnID uint64, plan *optimizer.QueryPlan, opType OperationType) (*ResultSet, error) {
	te.mutex.RLock()
	txn, exists := te.activeTransactions[txnID]
	te.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transaction %d not found", txnID)
	}

	txn.mutex.Lock()
	if txn.State != TxnActive {
		txn.mutex.Unlock()
		return nil, fmt.Errorf("transaction %d is not active", txnID)
	}
	txn.mutex.Unlock()

	// Acquire necessary locks based on operation type
	if err := te.acquireLocks(txn, plan, opType); err != nil {
		return nil, fmt.Errorf("failed to acquire locks: %w", err)
	}

	// Execute the query
	result, err := te.queryExecutor.Execute(txn.Context, plan)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Record operation
	op := &TransactionOperation{
		Type:      opType,
		TableName: "", // TODO: Extract from plan
		Plan:      plan,
		Timestamp: time.Now(),
		Result:    result,
	}

	txn.mutex.Lock()
	txn.Operations = append(txn.Operations, op)

	// Update statistics
	if result != nil {
		txn.RowsRead += uint64(result.RowCount())
	}
	txn.mutex.Unlock()

	return result, nil
}

// CreateSavepoint creates a savepoint within a transaction
func (te *TransactionExecutor) CreateSavepoint(txnID uint64, name string) error {
	te.mutex.RLock()
	txn, exists := te.activeTransactions[txnID]
	te.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("transaction %d not found", txnID)
	}

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	if txn.State != TxnActive {
		return fmt.Errorf("transaction %d is not active", txnID)
	}

	savepoint := &Savepoint{
		Name:      name,
		TxnID:     txnID,
		Timestamp: time.Now(),
		Position:  len(txn.Operations),
	}

	txn.Savepoints[name] = savepoint
	return nil
}

// RollbackToSavepoint rolls back to a savepoint
func (te *TransactionExecutor) RollbackToSavepoint(txnID uint64, name string) error {
	te.mutex.RLock()
	txn, exists := te.activeTransactions[txnID]
	te.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("transaction %d not found", txnID)
	}

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	savepoint, exists := txn.Savepoints[name]
	if !exists {
		return fmt.Errorf("savepoint %s not found in transaction %d", name, txnID)
	}

	// TODO: Undo operations after savepoint position
	txn.Operations = txn.Operations[:savepoint.Position]

	return nil
}

// GetTransaction retrieves transaction information
func (te *TransactionExecutor) GetTransaction(txnID uint64) (*Transaction, error) {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	txn, exists := te.activeTransactions[txnID]
	if !exists {
		return nil, fmt.Errorf("transaction %d not found", txnID)
	}

	return txn, nil
}

// ListActiveTransactions returns all active transaction IDs
func (te *TransactionExecutor) ListActiveTransactions() []uint64 {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	txnIDs := make([]uint64, 0, len(te.activeTransactions))
	for txnID := range te.activeTransactions {
		txnIDs = append(txnIDs, txnID)
	}

	return txnIDs
}

// acquireLocks acquires necessary locks for an operation
func (te *TransactionExecutor) acquireLocks(txn *Transaction, plan *optimizer.QueryPlan, opType OperationType) error {
	// TODO: Extract table names from plan
	// TODO: Determine lock mode based on operation type and isolation level

	// For now, just acquire table-level locks
	switch opType {
	case SelectOp:
		// Shared lock for reads
		switch txn.IsolationLevel {
		case ReadUncommitted:
			// No locks needed
			return nil
		case ReadCommitted, RepeatableRead, Serializable:
			// TODO: Acquire shared locks on accessed tables
			return nil
		}
	case InsertOp, UpdateOp, DeleteOp:
		// Exclusive lock for writes
		// TODO: Acquire exclusive locks on modified tables
		return nil
	}

	return nil
}

// SetIsolationLevel sets the default isolation level
func (te *TransactionExecutor) SetIsolationLevel(level IsolationLevel) {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	te.isolationLevel = level
}

// GetIsolationLevel returns the default isolation level
func (te *TransactionExecutor) GetIsolationLevel() IsolationLevel {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	return te.isolationLevel
}

// String methods for enums

func (ts TransactionState) String() string {
	switch ts {
	case TxnActive:
		return "ACTIVE"
	case TxnPreparing:
		return "PREPARING"
	case TxnCommitting:
		return "COMMITTING"
	case TxnCommitted:
		return "COMMITTED"
	case TxnAborting:
		return "ABORTING"
	case TxnAborted:
		return "ABORTED"
	default:
		return "UNKNOWN"
	}
}

func (il IsolationLevel) String() string {
	switch il {
	case ReadUncommitted:
		return "READ_UNCOMMITTED"
	case ReadCommitted:
		return "READ_COMMITTED"
	case RepeatableRead:
		return "REPEATABLE_READ"
	case Serializable:
		return "SERIALIZABLE"
	default:
		return "UNKNOWN"
	}
}
