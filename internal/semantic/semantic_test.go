package semantic

import (
	"testing"

	"relational-db/internal/compiler"
)

// TestNewSemanticAnalyzer tests creating a semantic analyzer
func TestNewSemanticAnalyzer(t *testing.T) {
	catalog := compiler.NewMockCatalog()
	analyzer := NewSemanticAnalyzer(catalog)

	if analyzer == nil {
		t.Fatal("Expected analyzer to be created")
	}

	if analyzer.catalog != catalog {
		t.Error("Expected catalog to be set")
	}

	// Should have default rules registered
	if len(analyzer.rules) == 0 {
		t.Error("Expected default rules to be registered")
	}
}

// TestSemanticInfo tests SemanticInfo structure
func TestSemanticInfo(t *testing.T) {
	info := &SemanticInfo{
		SchemaValid:    true,
		SemanticsValid: true,
		AccessValid:    true,
		Errors:         make([]SemanticError, 0),
	}

	if !info.IsValid() {
		t.Error("Expected info to be valid")
	}

	if info.HasErrors() {
		t.Error("Expected no errors")
	}

	// Add an error
	info.Errors = append(info.Errors, SemanticError{
		Code:     ErrGeneric,
		Category: CategoryGeneral,
		Message:  "Test error",
	})

	if !info.HasErrors() {
		t.Error("Expected errors to be present")
	}
}

// TestErrorCodes tests error code constants
func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		category ErrorCategory
		name     string
	}{
		{ErrGroupByRequired, CategoryGroupBy, "GROUP BY Required"},
		{ErrAggregateInWhere, CategoryAggregate, "Aggregate in WHERE"},
		{ErrScalarSubqueryMultiCol, CategorySubquery, "Scalar subquery multi-column"},
		{ErrTableAlreadyExists, CategorySchema, "Table already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSemanticError(tt.code, tt.category, tt.name)
			if err.Code != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code)
			}
			if err.Category != tt.category {
				t.Errorf("Expected category %v, got %v", tt.category, err.Category)
			}
		})
	}
}

// TestValidationContext tests ValidationContext
func TestValidationContext(t *testing.T) {
	catalog := compiler.NewMockCatalog()
	compiled := &compiler.CompiledQuery{
		ResolvedRefs: compiler.NewResolvedReferences(),
		TypeInfo:     compiler.NewTypeInformation(),
	}

	ctx := NewValidationContext(compiled, catalog)

	if ctx == nil {
		t.Fatal("Expected context to be created")
	}

	if ctx.Catalog != catalog {
		t.Error("Expected catalog to be set")
	}

	if ctx.SubqueryDepth != 0 {
		t.Error("Expected subquery depth to be 0")
	}

	// Test entering/exiting subquery
	ctx.EnterSubquery()
	if ctx.SubqueryDepth != 1 {
		t.Error("Expected subquery depth to be 1")
	}
	if !ctx.InSubquery {
		t.Error("Expected InSubquery to be true")
	}

	ctx.ExitSubquery()
	if ctx.SubqueryDepth != 0 {
		t.Error("Expected subquery depth to be 0 after exit")
	}
}

// TestValidationScope tests ValidationScope
func TestValidationScope(t *testing.T) {
	scope := NewValidationScope(nil)

	if scope == nil {
		t.Fatal("Expected scope to be created")
	}

	// Add a table
	table := compiler.NewTableMetadata("users")
	scope.AddTable("users", table)

	if !scope.HasTable("users") {
		t.Error("Expected table 'users' to exist")
	}

	if scope.HasTable("nonexistent") {
		t.Error("Expected table 'nonexistent' to not exist")
	}

	// Add an alias
	scope.AddAlias("u", "users")

	if scope.Aliases["u"] != "users" {
		t.Error("Expected alias 'u' to map to 'users'")
	}
}

// TestErrorCategories tests error category string representation
func TestErrorCategories(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{CategoryGeneral, "General"},
		{CategoryGroupBy, "GROUP BY"},
		{CategoryAggregate, "Aggregate"},
		{CategorySubquery, "Subquery"},
		{CategorySchema, "Schema"},
		{CategoryConstraint, "Constraint"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.category.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestExpressionLocation tests expression location string representation
func TestExpressionLocation(t *testing.T) {
	tests := []struct {
		location ExpressionLocation
		expected string
	}{
		{LocationUnknown, "UNKNOWN"},
		{LocationSelect, "SELECT"},
		{LocationWhere, "WHERE"},
		{LocationGroupBy, "GROUP BY"},
		{LocationHaving, "HAVING"},
		{LocationOrderBy, "ORDER BY"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.location.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestSubqueryType tests subquery type string representation
func TestSubqueryType(t *testing.T) {
	tests := []struct {
		stype    SubqueryType
		expected string
	}{
		{SubqueryScalar, "SCALAR"},
		{SubqueryInPredicate, "IN_PREDICATE"},
		{SubqueryExists, "EXISTS"},
		{SubqueryDerivedTable, "DERIVED_TABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.stype.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}
