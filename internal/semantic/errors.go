package semantic

import "fmt"

// ErrorCode represents a semantic error code
type ErrorCode int

const (
	// Generic errors (5000-5099)
	ErrGeneric ErrorCode = 5000

	// GROUP BY errors (5100-5199)
	ErrGroupByRequired      ErrorCode = 5101
	ErrColumnNotInGroupBy   ErrorCode = 5102
	ErrAggregateInGroupBy   ErrorCode = 5103
	ErrHavingWithoutGroupBy ErrorCode = 5104

	// Aggregate errors (5200-5299)
	ErrNestedAggregate        ErrorCode = 5201
	ErrAggregateInWhere       ErrorCode = 5202
	ErrInvalidAggregateArg    ErrorCode = 5203
	ErrWrongAggregateArgCount ErrorCode = 5204

	// Subquery errors (5300-5399)
	ErrScalarSubqueryMultiCol  ErrorCode = 5301
	ErrInSubqueryMultiCol      ErrorCode = 5302
	ErrDerivedTableNoAlias     ErrorCode = 5303
	ErrCorrelatedRefNotVisible ErrorCode = 5304

	// Schema errors (5400-5499)
	ErrTableAlreadyExists    ErrorCode = 5401
	ErrDuplicateColumn       ErrorCode = 5402
	ErrForeignKeyRefNotFound ErrorCode = 5403
	ErrCircularDependency    ErrorCode = 5404
	ErrTableNotFound         ErrorCode = 5405
)

// ErrorCategory represents the category of semantic error
type ErrorCategory int

const (
	CategoryGeneral ErrorCategory = iota
	CategoryGroupBy
	CategoryAggregate
	CategorySubquery
	CategorySchema
	CategoryConstraint
)

func (ec ErrorCategory) String() string {
	switch ec {
	case CategoryGroupBy:
		return "GROUP BY"
	case CategoryAggregate:
		return "Aggregate"
	case CategorySubquery:
		return "Subquery"
	case CategorySchema:
		return "Schema"
	case CategoryConstraint:
		return "Constraint"
	default:
		return "General"
	}
}

// SemanticError represents a semantic validation error
type SemanticError struct {
	Code       ErrorCode
	Category   ErrorCategory
	Message    string
	Hint       string
	Location   ErrorLocation
	Expression string // The problematic expression
	Suggestion string // How to fix it
}

// Error implements the error interface
func (se *SemanticError) Error() string {
	if se.Hint != "" {
		return fmt.Sprintf("%s: %s (Hint: %s)", se.Category, se.Message, se.Hint)
	}
	return fmt.Sprintf("%s: %s", se.Category, se.Message)
}

// IsCritical returns true if this is a critical error that should stop analysis
func (se *SemanticError) IsCritical() bool {
	// Schema errors are typically critical
	return se.Category == CategorySchema
}

// WithHint adds a hint to the error
func (se *SemanticError) WithHint(hint string) *SemanticError {
	se.Hint = hint
	return se
}

// WithSuggestion adds a suggestion to the error
func (se *SemanticError) WithSuggestion(suggestion string) *SemanticError {
	se.Suggestion = suggestion
	return se
}

// ErrorLocation represents the location of an error
type ErrorLocation struct {
	Line   int
	Column int
	Offset int
}

// SemanticWarning represents a non-critical semantic issue
type SemanticWarning struct {
	Message string
	Hint    string
}

// NewSemanticError creates a new semantic error
func NewSemanticError(code ErrorCode, category ErrorCategory, message string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Category: category,
		Message:  message,
	}
}

// NewGroupByError creates a GROUP BY related error
func NewGroupByError(code ErrorCode, message string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Category: CategoryGroupBy,
		Message:  message,
	}
}

// NewAggregateError creates an aggregate function related error
func NewAggregateError(code ErrorCode, message string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Category: CategoryAggregate,
		Message:  message,
	}
}

// NewSubqueryError creates a subquery related error
func NewSubqueryError(code ErrorCode, message string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Category: CategorySubquery,
		Message:  message,
	}
}

// NewSchemaError creates a schema related error
func NewSchemaError(code ErrorCode, message string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Category: CategorySchema,
		Message:  message,
	}
}
