package executor

import (
	"errors"
	"fmt"
)

// Common executor errors
var (
	ErrColumnNotFound        = errors.New("column not found")
	ErrInvalidIndex          = errors.New("invalid index")
	ErrUnsupportedExpression = errors.New("unsupported expression type")
	ErrNotImplemented        = errors.New("not implemented")
	ErrOperatorClosed        = errors.New("operator is closed")
	ErrInvalidOperator       = errors.New("invalid operator")
	ErrExecutionTimeout      = errors.New("execution timeout exceeded")
	ErrInsufficientMemory    = errors.New("insufficient memory for operation")
	ErrTypeMismatch          = errors.New("type mismatch")
	ErrNullValue             = errors.New("unexpected null value")
	ErrDivisionByZero        = errors.New("division by zero")
)

// ExecutionError represents an execution error with context
type ExecutionError struct {
	Op      string // Operator that caused the error
	Message string
	Cause   error
}

// NewExecutionError creates a new execution error
func NewExecutionError(op, message string, cause error) *ExecutionError {
	return &ExecutionError{
		Op:      op,
		Message: message,
		Cause:   cause,
	}
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying error
func (e *ExecutionError) Unwrap() error {
	return e.Cause
}
