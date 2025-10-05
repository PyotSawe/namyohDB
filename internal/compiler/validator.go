package compiler

import (
	"relational-db/internal/parser"
)

// ConstraintValidator validates SQL constraints
type ConstraintValidator struct {
	catalog CatalogManager
	refs    *ResolvedReferences
}

// NewConstraintValidator creates a new constraint validator
func NewConstraintValidator(catalog CatalogManager, refs *ResolvedReferences) *ConstraintValidator {
	return &ConstraintValidator{
		catalog: catalog,
		refs:    refs,
	}
}

// ValidateCreateTable validates constraints in CREATE TABLE
func (cv *ConstraintValidator) ValidateCreateTable(stmt *parser.CreateTableStatement) error {
	// TODO: Implement CREATE TABLE constraint validation
	// - Check only one PRIMARY KEY
	// - Validate FOREIGN KEY references
	// - Check DEFAULT value types
	return nil
}

// ValidateInsert validates constraints in INSERT
func (cv *ConstraintValidator) ValidateInsert(stmt *parser.InsertStatement) error {
	// TODO: Implement INSERT constraint validation
	// - Check NOT NULL constraints
	// - Validate PRIMARY KEY uniqueness
	// - Check FOREIGN KEY references
	return nil
}

// ValidateUpdate validates constraints in UPDATE
func (cv *ConstraintValidator) ValidateUpdate(stmt *parser.UpdateStatement) error {
	// TODO: Implement UPDATE constraint validation
	// - Check NOT NULL constraints
	// - Validate PRIMARY KEY immutability
	// - Check FOREIGN KEY references
	return nil
}

// AggregateValidator validates aggregate function usage
type AggregateValidator struct {
	refs *ResolvedReferences
}

// NewAggregateValidator creates a new aggregate validator
func NewAggregateValidator(refs *ResolvedReferences) *AggregateValidator {
	return &AggregateValidator{
		refs: refs,
	}
}

// Validate validates aggregate usage in SELECT statement
func (av *AggregateValidator) Validate(stmt *parser.SelectStatement) error {
	// TODO: Implement aggregate validation
	// - Check GROUP BY rules
	// - Validate HAVING clause
	// - Ensure non-aggregates are in GROUP BY
	return nil
}
