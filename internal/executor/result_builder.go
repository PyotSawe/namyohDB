// Package executor - Result Set Builder component
// Implements the Result Set Builder from the Execution Engine Layer architecture
package executor

import (
	"fmt"
	"sync"
)

// ResultBuilder builds result sets from operator output
// Architecture: Query Executor → Result Set Builder
type ResultBuilder struct {
	schema    *TupleSchema
	tuples    []*Tuple
	capacity  int
	buildTime int64 // nanoseconds
	mutex     sync.RWMutex
}

// NewResultBuilder creates a new result builder
func NewResultBuilder(schema *TupleSchema) *ResultBuilder {
	return &ResultBuilder{
		schema:   schema,
		tuples:   make([]*Tuple, 0, 1000), // Initial capacity
		capacity: 1000,
	}
}

// NewResultBuilderWithCapacity creates a result builder with specified capacity
func NewResultBuilderWithCapacity(schema *TupleSchema, capacity int) *ResultBuilder {
	return &ResultBuilder{
		schema:   schema,
		tuples:   make([]*Tuple, 0, capacity),
		capacity: capacity,
	}
}

// AddTuple adds a tuple to the result set
func (rb *ResultBuilder) AddTuple(tuple *Tuple) error {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// Validate schema compatibility
	if tuple.Schema == nil {
		return fmt.Errorf("tuple has no schema")
	}

	if !rb.schemasCompatible(tuple.Schema, rb.schema) {
		return fmt.Errorf("tuple schema mismatch: expected %d columns, got %d",
			len(rb.schema.Columns), len(tuple.Schema.Columns))
	}

	rb.tuples = append(rb.tuples, tuple)
	return nil
}

// AddTuples adds multiple tuples in batch
func (rb *ResultBuilder) AddTuples(tuples []*Tuple) error {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	for _, tuple := range tuples {
		if tuple.Schema == nil {
			return fmt.Errorf("tuple has no schema")
		}

		if !rb.schemasCompatible(tuple.Schema, rb.schema) {
			return fmt.Errorf("tuple schema mismatch")
		}
	}

	rb.tuples = append(rb.tuples, tuples...)
	return nil
}

// Build finalizes and returns the result set
func (rb *ResultBuilder) Build() *ResultSet {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	return &ResultSet{
		Schema: rb.schema,
		Tuples: rb.tuples,
	}
}

// Reset clears the builder for reuse
func (rb *ResultBuilder) Reset() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.tuples = rb.tuples[:0] // Keep capacity
}

// RowCount returns current number of tuples
func (rb *ResultBuilder) RowCount() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	return len(rb.tuples)
}

// Capacity returns the builder capacity
func (rb *ResultBuilder) Capacity() int {
	return rb.capacity
}

// Schema returns the result schema
func (rb *ResultBuilder) Schema() *TupleSchema {
	return rb.schema
}

// schemasCompatible checks if two schemas are compatible
func (rb *ResultBuilder) schemasCompatible(s1, s2 *TupleSchema) bool {
	if s1 == nil || s2 == nil {
		return false
	}

	if len(s1.Columns) != len(s2.Columns) {
		return false
	}

	for i := range s1.Columns {
		if s1.Columns[i].Type != s2.Columns[i].Type {
			return false
		}
	}

	return true
}

// ResultSetIterator provides iterator-style access to results
type ResultSetIterator struct {
	resultSet *ResultSet
	position  int
	mutex     sync.RWMutex
}

// NewResultSetIterator creates an iterator for a result set
func NewResultSetIterator(rs *ResultSet) *ResultSetIterator {
	return &ResultSetIterator{
		resultSet: rs,
		position:  0,
	}
}

// HasNext returns true if more tuples are available
func (it *ResultSetIterator) HasNext() bool {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	return it.position < len(it.resultSet.Tuples)
}

// Next returns the next tuple
func (it *ResultSetIterator) Next() (*Tuple, error) {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	if it.position >= len(it.resultSet.Tuples) {
		return nil, nil // EOF
	}

	tuple := it.resultSet.Tuples[it.position]
	it.position++
	return tuple, nil
}

// Reset resets the iterator to the beginning
func (it *ResultSetIterator) Reset() {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	it.position = 0
}

// Position returns the current position
func (it *ResultSetIterator) Position() int {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	return it.position
}

// Seek sets the iterator position
func (it *ResultSetIterator) Seek(position int) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	if position < 0 || position > len(it.resultSet.Tuples) {
		return fmt.Errorf("invalid seek position: %d", position)
	}

	it.position = position
	return nil
}
