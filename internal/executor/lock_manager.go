// Package executor - Lock Manager component
// Manages locks for concurrency control
package executor

import (
	"fmt"
	"sync"
	"time"
)

// LockManager manages locks for concurrency control
// Architecture: Part of Execution Engine Layer
type LockManager struct {
	// Lock tables
	tableLocks map[string]*LockTable // table name -> lock table
	pageLocks  map[PageLockKey]*Lock
	rowLocks   map[RowLockKey]*Lock

	// Deadlock detection
	waitForGraph *WaitForGraph

	// Configuration
	deadlockTimeout time.Duration
	lockTimeout     time.Duration

	mutex sync.RWMutex
}

// PageLockKey identifies a page lock
type PageLockKey struct {
	TableName string
	PageID    uint64
}

// RowLockKey identifies a row lock
type RowLockKey struct {
	TableName string
	PageID    uint64
	SlotID    uint16
}

// LockTable represents locks for a table
type LockTable struct {
	TableName string
	Locks     []*Lock
	mutex     sync.RWMutex
}

// Lock represents a lock on a resource
type Lock struct {
	ID            string
	TransactionID uint64
	LockMode      LockMode
	LockType      LockType
	ResourceID    interface{} // Can be table name, PageLockKey, or RowLockKey
	AcquiredAt    time.Time
	IsWaiting     bool
}

// LockMode defines the lock mode
type LockMode int

const (
	SharedLock                LockMode = iota // S - Read lock
	ExclusiveLock                             // X - Write lock
	IntentSharedLock                          // IS - Intent to acquire S locks
	IntentExclusiveLock                       // IX - Intent to acquire X locks
	SharedIntentExclusiveLock                 // SIX - Shared with intent to acquire X locks
)

// LockType defines the granularity of the lock
type LockType int

const (
	TableLock LockType = iota
	PageLock
	RowLock
)

// WaitForGraph tracks transaction dependencies for deadlock detection
type WaitForGraph struct {
	edges map[uint64][]uint64 // transaction -> transactions it's waiting for
	mutex sync.RWMutex
}

// NewLockManager creates a new lock manager
func NewLockManager() *LockManager {
	return &LockManager{
		tableLocks:      make(map[string]*LockTable),
		pageLocks:       make(map[PageLockKey]*Lock),
		rowLocks:        make(map[RowLockKey]*Lock),
		waitForGraph:    NewWaitForGraph(),
		deadlockTimeout: 30 * time.Second,
		lockTimeout:     10 * time.Second,
	}
}

// NewWaitForGraph creates a new wait-for graph
func NewWaitForGraph() *WaitForGraph {
	return &WaitForGraph{
		edges: make(map[uint64][]uint64),
	}
}

// AcquireTableLock acquires a lock on a table
func (lm *LockManager) AcquireTableLock(txnID uint64, tableName string, mode LockMode) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Get or create lock table
	lockTable, exists := lm.tableLocks[tableName]
	if !exists {
		lockTable = &LockTable{
			TableName: tableName,
			Locks:     make([]*Lock, 0),
		}
		lm.tableLocks[tableName] = lockTable
	}

	// Check for lock compatibility
	if !lm.isCompatible(lockTable.Locks, txnID, mode) {
		// TODO: Implement lock waiting and deadlock detection
		return fmt.Errorf("lock conflict on table %s", tableName)
	}

	// Grant lock
	lock := &Lock{
		ID:            fmt.Sprintf("lock_%d_%s", txnID, tableName),
		TransactionID: txnID,
		LockMode:      mode,
		LockType:      TableLock,
		ResourceID:    tableName,
		AcquiredAt:    time.Now(),
		IsWaiting:     false,
	}

	lockTable.Locks = append(lockTable.Locks, lock)
	return nil
}

// ReleaseTableLock releases a lock on a table
func (lm *LockManager) ReleaseTableLock(txnID uint64, tableName string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lockTable, exists := lm.tableLocks[tableName]
	if !exists {
		return fmt.Errorf("no locks found for table %s", tableName)
	}

	// Remove lock
	newLocks := make([]*Lock, 0)
	found := false
	for _, lock := range lockTable.Locks {
		if lock.TransactionID != txnID {
			newLocks = append(newLocks, lock)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("lock not found for transaction %d on table %s", txnID, tableName)
	}

	lockTable.Locks = newLocks

	// Clean up empty lock table
	if len(lockTable.Locks) == 0 {
		delete(lm.tableLocks, tableName)
	}

	return nil
}

// AcquirePageLock acquires a lock on a page
func (lm *LockManager) AcquirePageLock(txnID uint64, tableName string, pageID uint64, mode LockMode) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	key := PageLockKey{
		TableName: tableName,
		PageID:    pageID,
	}

	// Check if page is already locked
	if existingLock, exists := lm.pageLocks[key]; exists {
		if existingLock.TransactionID != txnID {
			// Different transaction holds the lock
			if !lm.isLockCompatible(existingLock.LockMode, mode) {
				return fmt.Errorf("lock conflict on page %d", pageID)
			}
		}
	}

	// Grant lock
	lock := &Lock{
		ID:            fmt.Sprintf("lock_%d_%s_%d", txnID, tableName, pageID),
		TransactionID: txnID,
		LockMode:      mode,
		LockType:      PageLock,
		ResourceID:    key,
		AcquiredAt:    time.Now(),
		IsWaiting:     false,
	}

	lm.pageLocks[key] = lock
	return nil
}

// ReleasePageLock releases a lock on a page
func (lm *LockManager) ReleasePageLock(txnID uint64, tableName string, pageID uint64) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	key := PageLockKey{
		TableName: tableName,
		PageID:    pageID,
	}

	lock, exists := lm.pageLocks[key]
	if !exists {
		return fmt.Errorf("lock not found for page %d", pageID)
	}

	if lock.TransactionID != txnID {
		return fmt.Errorf("transaction %d does not hold lock on page %d", txnID, pageID)
	}

	delete(lm.pageLocks, key)
	return nil
}

// AcquireRowLock acquires a lock on a row
func (lm *LockManager) AcquireRowLock(txnID uint64, tableName string, pageID uint64, slotID uint16, mode LockMode) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	key := RowLockKey{
		TableName: tableName,
		PageID:    pageID,
		SlotID:    slotID,
	}

	// Check if row is already locked
	if existingLock, exists := lm.rowLocks[key]; exists {
		if existingLock.TransactionID != txnID {
			// Different transaction holds the lock
			if !lm.isLockCompatible(existingLock.LockMode, mode) {
				return fmt.Errorf("lock conflict on row (%d, %d)", pageID, slotID)
			}
		}
	}

	// Grant lock
	lock := &Lock{
		ID:            fmt.Sprintf("lock_%d_%s_%d_%d", txnID, tableName, pageID, slotID),
		TransactionID: txnID,
		LockMode:      mode,
		LockType:      RowLock,
		ResourceID:    key,
		AcquiredAt:    time.Now(),
		IsWaiting:     false,
	}

	lm.rowLocks[key] = lock
	return nil
}

// ReleaseRowLock releases a lock on a row
func (lm *LockManager) ReleaseRowLock(txnID uint64, tableName string, pageID uint64, slotID uint16) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	key := RowLockKey{
		TableName: tableName,
		PageID:    pageID,
		SlotID:    slotID,
	}

	lock, exists := lm.rowLocks[key]
	if !exists {
		return fmt.Errorf("lock not found for row (%d, %d)", pageID, slotID)
	}

	if lock.TransactionID != txnID {
		return fmt.Errorf("transaction %d does not hold lock on row (%d, %d)", txnID, pageID, slotID)
	}

	delete(lm.rowLocks, key)
	return nil
}

// ReleaseAllLocks releases all locks held by a transaction
func (lm *LockManager) ReleaseAllLocks(txnID uint64) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Release table locks
	for tableName, lockTable := range lm.tableLocks {
		newLocks := make([]*Lock, 0)
		for _, lock := range lockTable.Locks {
			if lock.TransactionID != txnID {
				newLocks = append(newLocks, lock)
			}
		}
		lockTable.Locks = newLocks

		if len(lockTable.Locks) == 0 {
			delete(lm.tableLocks, tableName)
		}
	}

	// Release page locks
	for key, lock := range lm.pageLocks {
		if lock.TransactionID == txnID {
			delete(lm.pageLocks, key)
		}
	}

	// Release row locks
	for key, lock := range lm.rowLocks {
		if lock.TransactionID == txnID {
			delete(lm.rowLocks, key)
		}
	}

	// Remove from wait-for graph
	lm.waitForGraph.RemoveTransaction(txnID)

	return nil
}

// isCompatible checks if a new lock is compatible with existing locks
func (lm *LockManager) isCompatible(existingLocks []*Lock, txnID uint64, newMode LockMode) bool {
	for _, lock := range existingLocks {
		if lock.TransactionID == txnID {
			continue // Same transaction
		}

		if !lm.isLockCompatible(lock.LockMode, newMode) {
			return false
		}
	}

	return true
}

// isLockCompatible checks if two lock modes are compatible
func (lm *LockManager) isLockCompatible(mode1, mode2 LockMode) bool {
	// Lock compatibility matrix
	// S is compatible with S and IS
	// X is not compatible with anything
	// IS is compatible with IS and S
	// IX is compatible with IS and IX
	// SIX is compatible with IS

	switch mode1 {
	case SharedLock:
		return mode2 == SharedLock || mode2 == IntentSharedLock
	case ExclusiveLock:
		return false
	case IntentSharedLock:
		return mode2 != ExclusiveLock
	case IntentExclusiveLock:
		return mode2 == IntentSharedLock || mode2 == IntentExclusiveLock
	case SharedIntentExclusiveLock:
		return mode2 == IntentSharedLock
	default:
		return false
	}
}

// DetectDeadlock checks for deadlocks in the wait-for graph
func (lm *LockManager) DetectDeadlock() (bool, []uint64) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	return lm.waitForGraph.DetectCycle()
}

// WaitForGraph methods

// AddEdge adds an edge to the wait-for graph
func (wfg *WaitForGraph) AddEdge(from, to uint64) {
	wfg.mutex.Lock()
	defer wfg.mutex.Unlock()

	if wfg.edges[from] == nil {
		wfg.edges[from] = make([]uint64, 0)
	}

	wfg.edges[from] = append(wfg.edges[from], to)
}

// RemoveTransaction removes a transaction from the graph
func (wfg *WaitForGraph) RemoveTransaction(txnID uint64) {
	wfg.mutex.Lock()
	defer wfg.mutex.Unlock()

	delete(wfg.edges, txnID)

	// Remove as destination
	for from, destinations := range wfg.edges {
		newDests := make([]uint64, 0)
		for _, dest := range destinations {
			if dest != txnID {
				newDests = append(newDests, dest)
			}
		}
		wfg.edges[from] = newDests
	}
}

// DetectCycle detects cycles in the wait-for graph (deadlock)
func (wfg *WaitForGraph) DetectCycle() (bool, []uint64) {
	wfg.mutex.RLock()
	defer wfg.mutex.RUnlock()

	visited := make(map[uint64]bool)
	recStack := make(map[uint64]bool)
	path := make([]uint64, 0)

	for txnID := range wfg.edges {
		if !visited[txnID] {
			if wfg.detectCycleUtil(txnID, visited, recStack, &path) {
				return true, path
			}
		}
	}

	return false, nil
}

// detectCycleUtil is a utility function for cycle detection (DFS)
func (wfg *WaitForGraph) detectCycleUtil(txnID uint64, visited, recStack map[uint64]bool, path *[]uint64) bool {
	visited[txnID] = true
	recStack[txnID] = true
	*path = append(*path, txnID)

	for _, neighbor := range wfg.edges[txnID] {
		if !visited[neighbor] {
			if wfg.detectCycleUtil(neighbor, visited, recStack, path) {
				return true
			}
		} else if recStack[neighbor] {
			*path = append(*path, neighbor)
			return true
		}
	}

	recStack[txnID] = false
	*path = (*path)[:len(*path)-1]
	return false
}
