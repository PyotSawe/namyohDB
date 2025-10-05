package executor

import (
	"relational-db/internal/parser"
)

// SeqScanOperator performs sequential table scan
type SeqScanOperator struct {
	tableName  string
	filter     parser.Expression
	schema     *TupleSchema
	closed     bool
	tuplesRead int64
}

// NewSeqScanOperator creates a new sequential scan operator
func NewSeqScanOperator(tableName string, filter parser.Expression) *SeqScanOperator {
	return &SeqScanOperator{
		tableName: tableName,
		filter:    filter,
		closed:    true,
	}
}

// Open initializes the operator
func (op *SeqScanOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil // Already open
	}

	// TODO: Initialize scan state
	// - Get table schema from catalog
	// - Initialize page iterator
	// - Set up tuple reader

	op.closed = false
	return nil
}

// Next returns the next tuple
func (op *SeqScanOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Implement sequential scan logic
	// 1. Read next tuple from storage
	// 2. Apply filter if present
	// 3. Return tuple or nil if EOF

	return nil, nil // EOF for now
}

// Close releases resources
func (op *SeqScanOperator) Close() error {
	if op.closed {
		return nil
	}

	// TODO: Cleanup scan state
	op.closed = true
	return nil
}

// OperatorType returns the operator type
func (op *SeqScanOperator) OperatorType() string {
	return "SeqScan"
}

// EstimatedCost returns estimated cost
func (op *SeqScanOperator) EstimatedCost() float64 {
	return float64(op.tuplesRead) * 1.0
}

// IndexScanOperator performs index-based table scan
type IndexScanOperator struct {
	tableName  string
	indexName  string
	filter     parser.Expression
	schema     *TupleSchema
	closed     bool
	tuplesRead int64
}

// NewIndexScanOperator creates a new index scan operator
func NewIndexScanOperator(tableName, indexName string, filter parser.Expression) *IndexScanOperator {
	return &IndexScanOperator{
		tableName: tableName,
		indexName: indexName,
		filter:    filter,
		closed:    true,
	}
}

// Open initializes the operator
func (op *IndexScanOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	// TODO: Initialize index scan
	// - Open B-tree index
	// - Position iterator at start key
	// - Set up tuple fetcher

	op.closed = false
	return nil
}

// Next returns the next tuple
func (op *IndexScanOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Implement index scan logic
	// 1. Get next key from index
	// 2. Fetch tuple using RID
	// 3. Apply filter if present
	// 4. Return tuple or nil if EOF

	return nil, nil // EOF for now
}

// Close releases resources
func (op *IndexScanOperator) Close() error {
	if op.closed {
		return nil
	}

	// TODO: Cleanup index scan state
	op.closed = true
	return nil
}

// OperatorType returns the operator type
func (op *IndexScanOperator) OperatorType() string {
	return "IndexScan"
}

// EstimatedCost returns estimated cost
func (op *IndexScanOperator) EstimatedCost() float64 {
	return float64(op.tuplesRead) * 4.0 // Random I/O more expensive
}

// FilterOperator filters tuples based on predicate
type FilterOperator struct {
	child      PhysicalOperator
	predicate  parser.Expression
	evaluator  *ExpressionEvaluator
	closed     bool
	tuplesRead int64
	tuplesOut  int64
}

// NewFilterOperator creates a new filter operator
func NewFilterOperator(child PhysicalOperator, predicate parser.Expression) *FilterOperator {
	return &FilterOperator{
		child:     child,
		predicate: predicate,
		evaluator: NewExpressionEvaluator(),
		closed:    true,
	}
}

// Open initializes the operator
func (op *FilterOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	op.closed = false
	return nil
}

// Next returns the next tuple that passes the filter
func (op *FilterOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	for {
		tuple, err := op.child.Next()
		if err != nil {
			return nil, err
		}

		if tuple == nil {
			return nil, nil // EOF
		}

		op.tuplesRead++

		// Evaluate predicate
		result, err := op.evaluator.Evaluate(op.predicate, tuple)
		if err != nil {
			return nil, err
		}

		// Check if predicate is true
		if boolResult, ok := result.(bool); ok && boolResult {
			op.tuplesOut++
			return tuple, nil
		}

		// Tuple filtered out, continue to next
	}
}

// Close releases resources
func (op *FilterOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *FilterOperator) OperatorType() string {
	return "Filter"
}

// EstimatedCost returns estimated cost
func (op *FilterOperator) EstimatedCost() float64 {
	return op.child.EstimatedCost() + float64(op.tuplesRead)*0.01
}

// ProjectOperator projects specific columns
type ProjectOperator struct {
	child          PhysicalOperator
	projectionList []parser.Expression
	evaluator      *ExpressionEvaluator
	outputSchema   *TupleSchema
	closed         bool
}

// NewProjectOperator creates a new project operator
func NewProjectOperator(child PhysicalOperator, projectionList []parser.Expression) *ProjectOperator {
	return &ProjectOperator{
		child:          child,
		projectionList: projectionList,
		evaluator:      NewExpressionEvaluator(),
		closed:         true,
	}
}

// Open initializes the operator
func (op *ProjectOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	// TODO: Build output schema from projection list
	op.closed = false
	return nil
}

// Next returns the next projected tuple
func (op *ProjectOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	tuple, err := op.child.Next()
	if err != nil {
		return nil, err
	}

	if tuple == nil {
		return nil, nil // EOF
	}

	// TODO: Evaluate projection expressions and build output tuple
	// For now, just return input tuple
	return tuple, nil
}

// Close releases resources
func (op *ProjectOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *ProjectOperator) OperatorType() string {
	return "Project"
}

// EstimatedCost returns estimated cost
func (op *ProjectOperator) EstimatedCost() float64 {
	return op.child.EstimatedCost()
}

// LimitOperator limits the number of output tuples
type LimitOperator struct {
	child   PhysicalOperator
	limit   int64
	offset  int64
	count   int64
	skipped int64
	closed  bool
}

// NewLimitOperator creates a new limit operator
func NewLimitOperator(child PhysicalOperator, limit, offset int64) *LimitOperator {
	return &LimitOperator{
		child:  child,
		limit:  limit,
		offset: offset,
		closed: true,
	}
}

// Open initializes the operator
func (op *LimitOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	op.count = 0
	op.skipped = 0
	op.closed = false
	return nil
}

// Next returns the next tuple within limit
func (op *LimitOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// Skip offset tuples
	for op.skipped < op.offset {
		tuple, err := op.child.Next()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			return nil, nil // EOF before offset reached
		}
		op.skipped++
	}

	// Check limit
	if op.limit > 0 && op.count >= op.limit {
		return nil, nil // Limit reached
	}

	tuple, err := op.child.Next()
	if err != nil {
		return nil, err
	}

	if tuple != nil {
		op.count++
	}

	return tuple, nil
}

// Close releases resources
func (op *LimitOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *LimitOperator) OperatorType() string {
	return "Limit"
}

// EstimatedCost returns estimated cost
func (op *LimitOperator) EstimatedCost() float64 {
	// Limit can terminate early, so cost is reduced
	limitCost := float64(op.offset + op.limit)
	childCost := op.child.EstimatedCost()
	if limitCost < childCost {
		return limitCost
	}
	return childCost
}
