// Package executor - Cursor Manager component
// Manages cursors for result set navigation
package executor

import (
	"fmt"
	"sync"
)

// CursorManager manages database cursors
// Architecture: Part of Execution Engine Layer
type CursorManager struct {
	cursors      map[string]*Cursor
	nextCursorID uint64
	mutex        sync.RWMutex
}

// Cursor represents a result set cursor
type Cursor struct {
	ID             string
	Name           string
	ResultSet      *ResultSet
	Iterator       *ResultSetIterator
	IsOpen         bool
	IsScrollable   bool
	IsHoldable     bool // Can survive transaction commit
	FetchDirection FetchDirection
	Position       int
	mutex          sync.RWMutex
}

// FetchDirection defines cursor fetch direction
type FetchDirection int

const (
	FetchNext FetchDirection = iota
	FetchPrior
	FetchFirst
	FetchLast
	FetchAbsolute
	FetchRelative
)

// NewCursorManager creates a new cursor manager
func NewCursorManager() *CursorManager {
	return &CursorManager{
		cursors:      make(map[string]*Cursor),
		nextCursorID: 1,
	}
}

// OpenCursor creates and opens a new cursor
func (cm *CursorManager) OpenCursor(name string, resultSet *ResultSet, scrollable, holdable bool) (*Cursor, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if cursor with this name already exists
	if _, exists := cm.cursors[name]; exists {
		return nil, fmt.Errorf("cursor %s already exists", name)
	}

	cursorID := fmt.Sprintf("cursor_%d", cm.nextCursorID)
	cm.nextCursorID++

	cursor := &Cursor{
		ID:             cursorID,
		Name:           name,
		ResultSet:      resultSet,
		Iterator:       NewResultSetIterator(resultSet),
		IsOpen:         true,
		IsScrollable:   scrollable,
		IsHoldable:     holdable,
		FetchDirection: FetchNext,
		Position:       0,
	}

	cm.cursors[name] = cursor
	return cursor, nil
}

// GetCursor retrieves a cursor by name
func (cm *CursorManager) GetCursor(name string) (*Cursor, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cursor, exists := cm.cursors[name]
	if !exists {
		return nil, fmt.Errorf("cursor %s not found", name)
	}

	return cursor, nil
}

// CloseCursor closes and removes a cursor
func (cm *CursorManager) CloseCursor(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cursor, exists := cm.cursors[name]
	if !exists {
		return fmt.Errorf("cursor %s not found", name)
	}

	cursor.mutex.Lock()
	cursor.IsOpen = false
	cursor.mutex.Unlock()

	delete(cm.cursors, name)
	return nil
}

// CloseAllCursors closes all non-holdable cursors (e.g., on transaction commit)
func (cm *CursorManager) CloseAllCursors(includeHoldable bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for name, cursor := range cm.cursors {
		if !includeHoldable && cursor.IsHoldable {
			continue
		}

		cursor.mutex.Lock()
		cursor.IsOpen = false
		cursor.mutex.Unlock()

		delete(cm.cursors, name)
	}

	return nil
}

// ListCursors returns all cursor names
func (cm *CursorManager) ListCursors() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	names := make([]string, 0, len(cm.cursors))
	for name := range cm.cursors {
		names = append(names, name)
	}

	return names
}

// Cursor methods

// Fetch fetches the next tuple based on fetch direction
func (c *Cursor) Fetch(direction FetchDirection, count int) ([]*Tuple, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.IsOpen {
		return nil, fmt.Errorf("cursor %s is closed", c.Name)
	}

	switch direction {
	case FetchNext:
		return c.fetchNext(count)
	case FetchPrior:
		if !c.IsScrollable {
			return nil, fmt.Errorf("cursor %s is not scrollable", c.Name)
		}
		return c.fetchPrior(count)
	case FetchFirst:
		if !c.IsScrollable {
			return nil, fmt.Errorf("cursor %s is not scrollable", c.Name)
		}
		return c.fetchFirst()
	case FetchLast:
		if !c.IsScrollable {
			return nil, fmt.Errorf("cursor %s is not scrollable", c.Name)
		}
		return c.fetchLast()
	default:
		return nil, fmt.Errorf("unsupported fetch direction: %d", direction)
	}
}

// fetchNext fetches the next N tuples
func (c *Cursor) fetchNext(count int) ([]*Tuple, error) {
	tuples := make([]*Tuple, 0, count)

	for i := 0; i < count && c.Iterator.HasNext(); i++ {
		tuple, err := c.Iterator.Next()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			break
		}
		tuples = append(tuples, tuple)
		c.Position++
	}

	return tuples, nil
}

// fetchPrior fetches the previous N tuples
func (c *Cursor) fetchPrior(count int) ([]*Tuple, error) {
	if c.Position == 0 {
		return nil, nil // Already at beginning
	}

	newPos := c.Position - count
	if newPos < 0 {
		newPos = 0
	}

	if err := c.Iterator.Seek(newPos); err != nil {
		return nil, err
	}

	tuples := make([]*Tuple, 0, count)
	for i := 0; i < count && c.Iterator.HasNext(); i++ {
		tuple, err := c.Iterator.Next()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			break
		}
		tuples = append(tuples, tuple)
	}

	c.Position = newPos + len(tuples)
	return tuples, nil
}

// fetchFirst fetches the first tuple
func (c *Cursor) fetchFirst() ([]*Tuple, error) {
	c.Iterator.Reset()
	c.Position = 0

	tuple, err := c.Iterator.Next()
	if err != nil {
		return nil, err
	}

	if tuple == nil {
		return nil, nil
	}

	c.Position = 1
	return []*Tuple{tuple}, nil
}

// fetchLast fetches the last tuple
func (c *Cursor) fetchLast() ([]*Tuple, error) {
	totalRows := len(c.ResultSet.Tuples)
	if totalRows == 0 {
		return nil, nil
	}

	if err := c.Iterator.Seek(totalRows - 1); err != nil {
		return nil, err
	}

	tuple, err := c.Iterator.Next()
	if err != nil {
		return nil, err
	}

	c.Position = totalRows
	return []*Tuple{tuple}, nil
}

// Reset resets the cursor to the beginning
func (c *Cursor) Reset() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.IsOpen {
		return fmt.Errorf("cursor %s is closed", c.Name)
	}

	c.Iterator.Reset()
	c.Position = 0
	return nil
}

// GetPosition returns the current cursor position
func (c *Cursor) GetPosition() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Position
}

// GetRowCount returns the total number of rows
func (c *Cursor) GetRowCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.ResultSet.Tuples)
}

// IsEOF returns true if cursor is at end of result set
func (c *Cursor) IsEOF() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return !c.Iterator.HasNext()
}
