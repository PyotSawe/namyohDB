package compiler

import "fmt"

// ErrorCode represents specific compilation error codes
type ErrorCode int

const (
	// Name resolution errors (1000-1999)
	ErrTableNotFound ErrorCode = 1001 + iota
	ErrColumnNotFound
	ErrAmbiguousColumn
	ErrInvalidAlias
	ErrNameResolution

	// Type checking errors (2000-2999)
	ErrTypeMismatch ErrorCode = 2001 + iota
	ErrInvalidOperand
	ErrInvalidFunctionArg
	ErrCannotCoerce
	ErrTypeChecking

	// Constraint validation errors (3000-3999)
	ErrNotNullViolation ErrorCode = 3001 + iota
	ErrPrimaryKeyDuplicate
	ErrForeignKeyInvalid
	ErrCheckViolation
	ErrUniqueViolation
	ErrConstraintValidation

	// Semantic analysis errors (4000-4999)
	ErrInvalidAggregate ErrorCode = 4001 + iota
	ErrInvalidGroupBy
	ErrInvalidSubquery
	ErrCircularReference
	ErrSemanticAnalysis
)

// ErrorCategory classifies compilation errors
type ErrorCategory int

const (
	ErrorCategoryNameResolution ErrorCategory = iota
	ErrorCategoryTypeChecking
	ErrorCategoryConstraintValidation
	ErrorCategorySemanticAnalysis
)

// String returns the string representation of ErrorCategory
func (ec ErrorCategory) String() string {
	switch ec {
	case ErrorCategoryNameResolution:
		return "NameResolution"
	case ErrorCategoryTypeChecking:
		return "TypeChecking"
	case ErrorCategoryConstraintValidation:
		return "ConstraintValidation"
	case ErrorCategorySemanticAnalysis:
		return "SemanticAnalysis"
	default:
		return "Unknown"
	}
}

// CompilationError represents a compilation error with context
type CompilationError struct {
	Code      ErrorCode
	Category  ErrorCategory
	Message   string
	Hint      string
	Line      int
	Column    int
	Position  int
	QueryPart string
}

// Error implements the error interface
func (ce *CompilationError) Error() string {
	if ce.Hint != "" {
		return fmt.Sprintf("[%s] %s\nHint: %s", ce.Category, ce.Message, ce.Hint)
	}
	return fmt.Sprintf("[%s] %s", ce.Category, ce.Message)
}

// WithHint adds a hint to the error
func (ce *CompilationError) WithHint(hint string) *CompilationError {
	ce.Hint = hint
	return ce
}

// NewCompilationError creates a new compilation error
func NewCompilationError(code ErrorCode, category ErrorCategory, message string) CompilationError {
	return CompilationError{
		Code:     code,
		Category: category,
		Message:  message,
	}
}
