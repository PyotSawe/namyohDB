package executor

import (
	"relational-db/internal/optimizer"
	"relational-db/internal/parser"
)

// NestedLoopJoinOperator implements nested loop join
type NestedLoopJoinOperator struct {
	leftChild   PhysicalOperator
	rightChild  PhysicalOperator
	joinCond    parser.Expression
	joinType    optimizer.JoinType
	evaluator   *ExpressionEvaluator
	currentLeft *Tuple
	closed      bool
}

// NewNestedLoopJoinOperator creates a new nested loop join operator
func NewNestedLoopJoinOperator(
	left, right PhysicalOperator,
	condition parser.Expression,
	joinType optimizer.JoinType,
) *NestedLoopJoinOperator {
	return &NestedLoopJoinOperator{
		leftChild:  left,
		rightChild: right,
		joinCond:   condition,
		joinType:   joinType,
		evaluator:  NewExpressionEvaluator(),
		closed:     true,
	}
}

// Open initializes the operator
func (op *NestedLoopJoinOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.leftChild.Open(ctx); err != nil {
		return err
	}

	if err := op.rightChild.Open(ctx); err != nil {
		return err
	}

	op.closed = false
	return nil
}

// Next returns the next joined tuple
func (op *NestedLoopJoinOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Implement nested loop join logic
	return nil, nil // EOF for now
}

// Close releases resources
func (op *NestedLoopJoinOperator) Close() error {
	if op.closed {
		return nil
	}

	err1 := op.leftChild.Close()
	err2 := op.rightChild.Close()
	op.closed = true

	if err1 != nil {
		return err1
	}
	return err2
}

// OperatorType returns the operator type
func (op *NestedLoopJoinOperator) OperatorType() string {
	return "NestedLoopJoin"
}

// EstimatedCost returns estimated cost
func (op *NestedLoopJoinOperator) EstimatedCost() float64 {
	return op.leftChild.EstimatedCost() * op.rightChild.EstimatedCost()
}

// HashJoinOperator implements hash join
type HashJoinOperator struct {
	buildChild PhysicalOperator
	probeChild PhysicalOperator
	joinCond   parser.Expression
	joinType   optimizer.JoinType
	hashTable  map[interface{}][]*Tuple
	evaluator  *ExpressionEvaluator
	closed     bool
}

// NewHashJoinOperator creates a new hash join operator
func NewHashJoinOperator(
	left, right PhysicalOperator,
	condition parser.Expression,
	joinType optimizer.JoinType,
) *HashJoinOperator {
	return &HashJoinOperator{
		buildChild: right, // Smaller relation for build phase
		probeChild: left,
		joinCond:   condition,
		joinType:   joinType,
		evaluator:  NewExpressionEvaluator(),
		closed:     true,
	}
}

// Open initializes the operator
func (op *HashJoinOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.buildChild.Open(ctx); err != nil {
		return err
	}

	if err := op.probeChild.Open(ctx); err != nil {
		return err
	}

	// TODO: Build hash table from build side
	op.hashTable = make(map[interface{}][]*Tuple)

	op.closed = false
	return nil
}

// Next returns the next joined tuple
func (op *HashJoinOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Implement hash join logic
	return nil, nil // EOF for now
}

// Close releases resources
func (op *HashJoinOperator) Close() error {
	if op.closed {
		return nil
	}

	err1 := op.buildChild.Close()
	err2 := op.probeChild.Close()
	op.hashTable = nil
	op.closed = true

	if err1 != nil {
		return err1
	}
	return err2
}

// OperatorType returns the operator type
func (op *HashJoinOperator) OperatorType() string {
	return "HashJoin"
}

// EstimatedCost returns estimated cost
func (op *HashJoinOperator) EstimatedCost() float64 {
	return op.buildChild.EstimatedCost() + op.probeChild.EstimatedCost()
}

// MergeJoinOperator implements sort-merge join
type MergeJoinOperator struct {
	leftChild  PhysicalOperator
	rightChild PhysicalOperator
	joinCond   parser.Expression
	joinKeys   []parser.Expression
	evaluator  *ExpressionEvaluator
	closed     bool
}

// NewMergeJoinOperator creates a new merge join operator
func NewMergeJoinOperator(
	left, right PhysicalOperator,
	condition parser.Expression,
	joinKeys []parser.Expression,
) *MergeJoinOperator {
	return &MergeJoinOperator{
		leftChild:  left,
		rightChild: right,
		joinCond:   condition,
		joinKeys:   joinKeys,
		evaluator:  NewExpressionEvaluator(),
		closed:     true,
	}
}

// Open initializes the operator
func (op *MergeJoinOperator) Open(ctx *ExecutionContext) error {
	if !op.closed {
		return nil
	}

	if err := op.leftChild.Open(ctx); err != nil {
		return err
	}

	if err := op.rightChild.Open(ctx); err != nil {
		return err
	}

	op.closed = false
	return nil
}

// Next returns the next joined tuple
func (op *MergeJoinOperator) Next() (*Tuple, error) {
	if op.closed {
		return nil, ErrOperatorClosed
	}

	// TODO: Implement merge join logic
	return nil, nil // EOF for now
}

// Close releases resources
func (op *MergeJoinOperator) Close() error {
	if op.closed {
		return nil
	}

	err1 := op.leftChild.Close()
	err2 := op.rightChild.Close()
	op.closed = true

	if err1 != nil {
		return err1
	}
	return err2
}

// OperatorType returns the operator type
func (op *MergeJoinOperator) OperatorType() string {
	return "MergeJoin"
}

// EstimatedCost returns estimated cost
func (op *MergeJoinOperator) EstimatedCost() float64 {
	return op.leftChild.EstimatedCost() + op.rightChild.EstimatedCost()
}
