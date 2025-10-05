// Package executor - Operator interface definitions
package executor

import (
	"relational-db/internal/parser"
)

// PhysicalOperator is the interface all physical operators must implement
// following the Volcano/Iterator model
type PhysicalOperator interface {
	// Open initializes the operator and allocates resources
	Open(ctx *ExecutionContext) error

	// Next returns the next tuple or nil if EOF
	Next() (*Tuple, error)

	// Close releases resources and performs cleanup
	Close() error

	// OperatorType returns the type of this operator
	OperatorType() string

	// EstimatedCost returns the estimated execution cost
	EstimatedCost() float64
}

// Tuple represents a row of data
type Tuple struct {
	Values []interface{}
	Schema *TupleSchema
}

// NewTuple creates a new tuple
func NewTuple(schema *TupleSchema, values []interface{}) *Tuple {
	return &Tuple{
		Schema: schema,
		Values: values,
	}
}

// GetColumn returns value by column name
func (t *Tuple) GetColumn(name string) (interface{}, error) {
	idx := t.Schema.GetColumnIndex(name)
	if idx < 0 || idx >= len(t.Values) {
		return nil, ErrColumnNotFound
	}
	return t.Values[idx], nil
}

// GetColumnByIndex returns value by index
func (t *Tuple) GetColumnByIndex(idx int) (interface{}, error) {
	if idx < 0 || idx >= len(t.Values) {
		return nil, ErrColumnNotFound
	}
	return t.Values[idx], nil
}

// SetColumn sets value by column name
func (t *Tuple) SetColumn(name string, value interface{}) error {
	idx := t.Schema.GetColumnIndex(name)
	if idx < 0 || idx >= len(t.Values) {
		return ErrColumnNotFound
	}
	t.Values[idx] = value
	return nil
}

// Clone creates a copy of the tuple
func (t *Tuple) Clone() *Tuple {
	values := make([]interface{}, len(t.Values))
	copy(values, t.Values)
	return &Tuple{
		Schema: t.Schema,
		Values: values,
	}
}

// TupleSchema defines the schema of a tuple
type TupleSchema struct {
	Columns []ColumnInfo
	nameMap map[string]int // column name to index mapping
}

// ColumnInfo contains metadata about a column
type ColumnInfo struct {
	Name      string
	Type      ColumnType
	Nullable  bool
	TableName string // Optional table qualifier
}

// ColumnType represents the data type of a column
type ColumnType int

const (
	TypeInt ColumnType = iota
	TypeBigInt
	TypeFloat
	TypeDouble
	TypeString
	TypeBoolean
	TypeDate
	TypeTimestamp
	TypeNull
)

// String returns string representation of column type
func (ct ColumnType) String() string {
	switch ct {
	case TypeInt:
		return "INT"
	case TypeBigInt:
		return "BIGINT"
	case TypeFloat:
		return "FLOAT"
	case TypeDouble:
		return "DOUBLE"
	case TypeString:
		return "STRING"
	case TypeBoolean:
		return "BOOLEAN"
	case TypeDate:
		return "DATE"
	case TypeTimestamp:
		return "TIMESTAMP"
	case TypeNull:
		return "NULL"
	default:
		return "UNKNOWN"
	}
}

// NewTupleSchema creates a new tuple schema
func NewTupleSchema(columns []ColumnInfo) *TupleSchema {
	nameMap := make(map[string]int, len(columns))
	for i, col := range columns {
		nameMap[col.Name] = i
	}
	return &TupleSchema{
		Columns: columns,
		nameMap: nameMap,
	}
}

// GetColumnIndex returns the index of a column by name
func (ts *TupleSchema) GetColumnIndex(name string) int {
	if idx, ok := ts.nameMap[name]; ok {
		return idx
	}
	return -1
}

// GetColumn returns column info by name
func (ts *TupleSchema) GetColumn(name string) (*ColumnInfo, bool) {
	idx := ts.GetColumnIndex(name)
	if idx < 0 {
		return nil, false
	}
	return &ts.Columns[idx], true
}

// ColumnCount returns the number of columns
func (ts *TupleSchema) ColumnCount() int {
	return len(ts.Columns)
}

// ResultSet holds query execution results
type ResultSet struct {
	Schema *TupleSchema
	Tuples []*Tuple
}

// NewResultSet creates a new empty result set
func NewResultSet() *ResultSet {
	return &ResultSet{
		Tuples: make([]*Tuple, 0),
	}
}

// NewResultSetWithSchema creates a result set with schema
func NewResultSetWithSchema(schema *TupleSchema) *ResultSet {
	return &ResultSet{
		Schema: schema,
		Tuples: make([]*Tuple, 0),
	}
}

// AddTuple adds a tuple to the result set
func (rs *ResultSet) AddTuple(tuple *Tuple) {
	if rs.Schema == nil && tuple.Schema != nil {
		rs.Schema = tuple.Schema
	}
	rs.Tuples = append(rs.Tuples, tuple)
}

// RowCount returns the number of rows
func (rs *ResultSet) RowCount() int {
	return len(rs.Tuples)
}

// GetTuple returns a tuple by index
func (rs *ResultSet) GetTuple(idx int) (*Tuple, error) {
	if idx < 0 || idx >= len(rs.Tuples) {
		return nil, ErrInvalidIndex
	}
	return rs.Tuples[idx], nil
}

// ExpressionEvaluator evaluates expressions against tuples
type ExpressionEvaluator struct{}

// NewExpressionEvaluator creates a new expression evaluator
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{}
}

// Evaluate evaluates an expression against a tuple
func (ee *ExpressionEvaluator) Evaluate(expr parser.Expression, tuple *Tuple) (interface{}, error) {
	if expr == nil {
		return nil, nil
	}

	switch e := expr.(type) {
	case *parser.Literal:
		return ee.evaluateLiteral(e)

	case *parser.Identifier:
		return tuple.GetColumn(e.Value)

	case *parser.BinaryExpression:
		return ee.evaluateBinary(e, tuple)

	case *parser.UnaryExpression:
		return ee.evaluateUnary(e, tuple)

	case *parser.FunctionCall:
		return ee.evaluateFunction(e, tuple)

	default:
		return nil, ErrUnsupportedExpression
	}
}

// evaluateLiteral evaluates a literal expression
func (ee *ExpressionEvaluator) evaluateLiteral(lit *parser.Literal) (interface{}, error) {
	return lit.Value, nil
}

// evaluateBinary evaluates a binary expression
func (ee *ExpressionEvaluator) evaluateBinary(expr *parser.BinaryExpression, tuple *Tuple) (interface{}, error) {
	left, err := ee.Evaluate(expr.Left, tuple)
	if err != nil {
		return nil, err
	}

	right, err := ee.Evaluate(expr.Right, tuple)
	if err != nil {
		return nil, err
	}

	// Apply operator
	return ee.applyBinaryOperator(expr.Operator, left, right)
}

// evaluateUnary evaluates a unary expression
func (ee *ExpressionEvaluator) evaluateUnary(expr *parser.UnaryExpression, tuple *Tuple) (interface{}, error) {
	operand, err := ee.Evaluate(expr.Operand, tuple)
	if err != nil {
		return nil, err
	}

	// Apply unary operator
	return ee.applyUnaryOperator(expr.Operator, operand)
}

// evaluateFunction evaluates a function call
func (ee *ExpressionEvaluator) evaluateFunction(expr *parser.FunctionCall, tuple *Tuple) (interface{}, error) {
	// TODO: Implement function evaluation
	return nil, ErrNotImplemented
}

// applyBinaryOperator applies a binary operator
func (ee *ExpressionEvaluator) applyBinaryOperator(op parser.BinaryOperator, left, right interface{}) (interface{}, error) {
	// TODO: Implement binary operators (=, <, >, +, -, *, /, etc.)
	return nil, ErrNotImplemented
}

// applyUnaryOperator applies a unary operator
func (ee *ExpressionEvaluator) applyUnaryOperator(op parser.UnaryOperator, operand interface{}) (interface{}, error) {
	// TODO: Implement unary operators (NOT, -, etc.)
	return nil, ErrNotImplemented
}
