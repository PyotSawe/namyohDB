package executor

import (
	"relational-db/internal/parser"
)

// HashAggregateOperator implements hash-based aggregation
type HashAggregateOperator struct {
	child        PhysicalOperator
	groupByKeys  []parser.Expression
	aggFunctions []parser.Expression
	hashTable    map[interface{}]*AggregateState
	evaluator    *ExpressionEvaluator
	iterator     []interface{} // Keys for iteration
	iterPos      int
	closed       bool
}

// NewHashAggregateOperator creates a new hash aggregate operator
func NewHashAggregateOperator(
	child PhysicalOperator,
	groupByKeys []parser.Expression,
	aggFunctions []parser.Expression,
) *HashAggregateOperator {
	return &HashAggregateOperator{
		child:        child,
		groupByKeys:  groupByKeys,
		aggFunctions: aggFunctions,
		evaluator:    NewExpressionEvaluator(),
		closed:       true,
	}
}

// Open initializes the operator
func (op *HashAggregateOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	// Initialize hash table
	op.hashTable = make(map[interface{}]*AggregateState)

	// TODO: Build hash table by consuming all input
	// For each input tuple:
	//   - Compute group key
	//   - Update aggregate state

	// Initialize iterator
	op.iterator = make([]interface{}, 0, len(op.hashTable))
	for key := range op.hashTable {
		op.iterator = append(op.iterator, key)
	}
	op.iterPos = 0

	op.closed = false
	return nil
}

// Next returns the next aggregated tuple
func (op *HashAggregateOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	if op.iterPos >= len(op.iterator) {
		return nil, nil // EOF
	}

	// TODO: Return next group's aggregated result
	op.iterPos++
	return nil, nil
}

// Close releases resources
func (op *HashAggregateOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.hashTable = nil
	op.iterator = nil
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *HashAggregateOperator) OperatorType() string {
	return "HashAggregate"
}

// EstimatedCost returns estimated cost
func (op *HashAggregateOperator) EstimatedCost() float64 {
	return op.child.EstimatedCost()
}

// SortAggregateOperator implements sort-based aggregation
type SortAggregateOperator struct {
	child        PhysicalOperator
	groupByKeys  []parser.Expression
	aggFunctions []parser.Expression
	sortedInput  []*Tuple
	evaluator    *ExpressionEvaluator
	currentPos   int
	currentGroup interface{}
	currentAgg   *AggregateState
	closed       bool
}

// NewSortAggregateOperator creates a new sort aggregate operator
func NewSortAggregateOperator(
	child PhysicalOperator,
	groupByKeys []parser.Expression,
	aggFunctions []parser.Expression,
) *SortAggregateOperator {
	return &SortAggregateOperator{
		child:        child,
		groupByKeys:  groupByKeys,
		aggFunctions: aggFunctions,
		evaluator:    NewExpressionEvaluator(),
		closed:       true,
	}
}

// Open initializes the operator
func (op *SortAggregateOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	// TODO: Sort input by group keys
	op.sortedInput = make([]*Tuple, 0)
	op.currentPos = 0

	op.closed = false
	return nil
}

// Next returns the next aggregated tuple
func (op *SortAggregateOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Process sorted groups
	return nil, nil // EOF for now
}

// Close releases resources
func (op *SortAggregateOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.sortedInput = nil
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *SortAggregateOperator) OperatorType() string {
	return "SortAggregate"
}

// EstimatedCost returns estimated cost
func (op *SortAggregateOperator) EstimatedCost() float64 {
	// Cost includes sorting
	return op.child.EstimatedCost() * 1.5
}

// SortOperator implements external merge sort
type SortOperator struct {
	child      PhysicalOperator
	sortKeys   []parser.Expression
	sortedData []*Tuple
	evaluator  *ExpressionEvaluator
	currentPos int
	closed     bool
}

// NewSortOperator creates a new sort operator
func NewSortOperator(child PhysicalOperator, sortKeys []parser.Expression) *SortOperator {
	return &SortOperator{
		child:     child,
		sortKeys:  sortKeys,
		evaluator: NewExpressionEvaluator(),
		closed:    true,
	}
}

// Open initializes the operator
func (op *SortOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.child.Open(ctx); err != nil {
		return err
	}

	// TODO: Sort all input tuples
	// For now, just collect input
	op.sortedData = make([]*Tuple, 0)
	op.currentPos = 0

	op.closed = false
	return nil
}

// Next returns the next sorted tuple
func (op *SortOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	if op.currentPos >= len(op.sortedData) {
		return nil, nil // EOF
	}

	tuple := op.sortedData[op.currentPos]
	op.currentPos++
	return tuple, nil
}

// Close releases resources
func (op *SortOperator) Close() error {
	if op.closed {
		return nil
	}

	err := op.child.Close()
	op.sortedData = nil
	op.closed = true
	return err
}

// OperatorType returns the operator type
func (op *SortOperator) OperatorType() string {
	return "Sort"
}

// EstimatedCost returns estimated cost
func (op *SortOperator) EstimatedCost() float64 {
	// O(n log n) sorting cost
	return op.child.EstimatedCost() * 1.5
}

// AggregateState holds the state for aggregate computations
type AggregateState struct {
	Count  int64
	Sum    float64
	Min    interface{}
	Max    interface{}
	Values []interface{} // For functions like AVG that need multiple values
}

// NewAggregateState creates a new aggregate state
func NewAggregateState() *AggregateState {
	return &AggregateState{
		Values: make([]interface{}, 0),
	}
}

// Update updates the aggregate state with a new value
func (as *AggregateState) Update(value interface{}) {
	as.Count++
	as.Values = append(as.Values, value)

	// TODO: Update Min, Max, Sum based on value type
}

// Finalize computes final aggregate value
func (as *AggregateState) Finalize(aggType string) interface{} {
	switch aggType {
	case "COUNT":
		return as.Count
	case "SUM":
		return as.Sum
	case "AVG":
		if as.Count > 0 {
			return as.Sum / float64(as.Count)
		}
		return nil
	case "MIN":
		return as.Min
	case "MAX":
		return as.Max
	default:
		return nil
	}
}
