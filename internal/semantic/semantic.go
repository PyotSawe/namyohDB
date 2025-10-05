// Package semantic implements the Semantic Analyzer for NamyohDB.package semantic

// It performs deep semantic validation beyond basic syntax and type checking.
package semantic

import (
	"fmt"
	"time"

	"relational-db/internal/compiler"
	"relational-db/internal/parser"
)

// SemanticAnalyzer performs semantic validation on compiled queries
type SemanticAnalyzer struct {
	catalog compiler.CatalogManager
	rules   []SemanticRule
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer(catalog compiler.CatalogManager) *SemanticAnalyzer {
	sa := &SemanticAnalyzer{
		catalog: catalog,
		rules:   make([]SemanticRule, 0),
	}

	// Register default validation rules
	sa.RegisterDefaultRules()

	return sa
}

// RegisterDefaultRules registers the standard semantic validation rules
func (sa *SemanticAnalyzer) RegisterDefaultRules() {
	// Add GROUP BY validation rule
	sa.AddRule(&GroupByValidationRule{})

	// Add aggregate validation rule
	sa.AddRule(&AggregateValidationRule{})

	// Add subquery validation rule
	sa.AddRule(&SubqueryValidationRule{})

	// Add schema validation rule
	sa.AddRule(&SchemaValidationRule{catalog: sa.catalog})
}

// AddRule adds a custom validation rule
func (sa *SemanticAnalyzer) AddRule(rule SemanticRule) {
	sa.rules = append(sa.rules, rule)
}

// Analyze performs semantic analysis on a compiled query
func (sa *SemanticAnalyzer) Analyze(compiled *compiler.CompiledQuery) (*SemanticInfo, error) {
	info := &SemanticInfo{
		CompiledQuery:  compiled,
		SchemaValid:    true,
		SemanticsValid: true,
		AccessValid:    true,
		Errors:         make([]SemanticError, 0),
		Warnings:       make([]SemanticWarning, 0),
		AnalyzedAt:     time.Now(),
	}

	// Create validation context
	ctx := NewValidationContext(compiled, sa.catalog)

	// Run all validation rules
	for _, rule := range sa.rules {
		if err := rule.Validate(compiled, ctx); err != nil {
			if semErr, ok := err.(*SemanticError); ok {
				info.Errors = append(info.Errors, *semErr)
				info.SemanticsValid = false
			} else {
				// Convert to SemanticError
				info.Errors = append(info.Errors, SemanticError{
					Code:     ErrGeneric,
					Category: CategoryGeneral,
					Message:  err.Error(),
				})
				info.SemanticsValid = false
			}
		}

		// Check for warnings
		if len(ctx.Warnings) > 0 {
			info.Warnings = append(info.Warnings, ctx.Warnings...)
			ctx.Warnings = ctx.Warnings[:0] // Clear for next rule
		}

		// Stop on critical errors
		if ctx.HasCriticalError() {
			break
		}
	}

	// Extract metadata from context
	info.HasAggregates = ctx.HasAggregates
	info.HasSubqueries = ctx.HasSubqueries
	info.HasGroupBy = ctx.HasGroupBy

	if ctx.AggregateInfo != nil {
		info.AggregateInfo = ctx.AggregateInfo
	}

	if ctx.SubqueryInfo != nil {
		info.SubqueryInfo = ctx.SubqueryInfo
	}

	if ctx.GroupByInfo != nil {
		info.GroupByInfo = ctx.GroupByInfo
	}

	return info, nil
}

// SemanticInfo stores the results of semantic analysis
type SemanticInfo struct {
	// Source query
	CompiledQuery *compiler.CompiledQuery

	// Validation results
	SchemaValid    bool
	SemanticsValid bool
	AccessValid    bool

	// Aggregate metadata
	HasAggregates bool
	AggregateInfo *AggregateMetadata

	// Subquery metadata
	HasSubqueries bool
	SubqueryInfo  []*SubqueryMetadata

	// GROUP BY metadata
	HasGroupBy  bool
	GroupByInfo *GroupByMetadata

	// Errors and warnings
	Errors   []SemanticError
	Warnings []SemanticWarning

	// Analysis timestamp
	AnalyzedAt time.Time
}

// HasErrors returns true if semantic analysis found errors
func (si *SemanticInfo) HasErrors() bool {
	return len(si.Errors) > 0
}

// IsValid returns true if the query passed all semantic checks
func (si *SemanticInfo) IsValid() bool {
	return si.SchemaValid && si.SemanticsValid && si.AccessValid && !si.HasErrors()
}

// SemanticRule is an interface for pluggable validation rules
type SemanticRule interface {
	Name() string
	Validate(compiled *compiler.CompiledQuery, ctx *ValidationContext) error
}

// ValidationContext tracks state during semantic analysis
type ValidationContext struct {
	// Compiled query info
	CompiledQuery *compiler.CompiledQuery
	ResolvedRefs  *compiler.ResolvedReferences
	TypeInfo      *compiler.TypeInformation

	// Catalog access
	Catalog compiler.CatalogManager

	// Current scope
	Scope *ValidationScope

	// State flags
	InWhereClause   bool
	InGroupByClause bool
	InHavingClause  bool
	InOrderByClause bool
	InSubquery      bool

	// Aggregate tracking
	AggregateDepth int
	HasAggregates  bool
	AggregateInfo  *AggregateMetadata

	// Subquery tracking
	SubqueryDepth int
	HasSubqueries bool
	SubqueryInfo  []*SubqueryMetadata
	OuterScopes   []*ValidationScope

	// GROUP BY tracking
	HasGroupBy  bool
	GroupByInfo *GroupByMetadata

	// Error collection
	Errors        []SemanticError
	Warnings      []SemanticWarning
	CriticalError bool
}

// NewValidationContext creates a new validation context
func NewValidationContext(compiled *compiler.CompiledQuery, catalog compiler.CatalogManager) *ValidationContext {
	return &ValidationContext{
		CompiledQuery: compiled,
		ResolvedRefs:  compiled.ResolvedRefs,
		TypeInfo:      compiled.TypeInfo,
		Catalog:       catalog,
		Scope:         NewValidationScope(nil),
		OuterScopes:   make([]*ValidationScope, 0),
		Errors:        make([]SemanticError, 0),
		Warnings:      make([]SemanticWarning, 0),
	}
}

// AddError adds a semantic error
func (ctx *ValidationContext) AddError(err SemanticError) {
	ctx.Errors = append(ctx.Errors, err)
	if err.IsCritical() {
		ctx.CriticalError = true
	}
}

// AddWarning adds a semantic warning
func (ctx *ValidationContext) AddWarning(warning SemanticWarning) {
	ctx.Warnings = append(ctx.Warnings, warning)
}

// HasCriticalError returns true if a critical error was encountered
func (ctx *ValidationContext) HasCriticalError() bool {
	return ctx.CriticalError
}

// EnterSubquery prepares context for analyzing a subquery
func (ctx *ValidationContext) EnterSubquery() {
	ctx.SubqueryDepth++
	ctx.OuterScopes = append(ctx.OuterScopes, ctx.Scope)
	ctx.Scope = NewValidationScope(ctx.Scope)
	ctx.InSubquery = true
}

// ExitSubquery restores context after analyzing a subquery
func (ctx *ValidationContext) ExitSubquery() {
	if len(ctx.OuterScopes) > 0 {
		ctx.Scope = ctx.OuterScopes[len(ctx.OuterScopes)-1]
		ctx.OuterScopes = ctx.OuterScopes[:len(ctx.OuterScopes)-1]
	}
	ctx.SubqueryDepth--
	ctx.InSubquery = ctx.SubqueryDepth > 0
}

// ValidationScope represents a visibility scope
type ValidationScope struct {
	// Tables visible in this scope
	Tables map[string]*compiler.TableMetadata

	// Columns visible in this scope
	Columns map[string]*compiler.ColumnMetadata

	// Aliases in this scope
	Aliases map[string]string

	// Parent scope (for nested scopes)
	Parent *ValidationScope
}

// NewValidationScope creates a new validation scope
func NewValidationScope(parent *ValidationScope) *ValidationScope {
	return &ValidationScope{
		Tables:  make(map[string]*compiler.TableMetadata),
		Columns: make(map[string]*compiler.ColumnMetadata),
		Aliases: make(map[string]string),
		Parent:  parent,
	}
}

// HasTable checks if a table is visible in this scope
func (vs *ValidationScope) HasTable(name string) bool {
	if _, ok := vs.Tables[name]; ok {
		return true
	}
	if vs.Parent != nil {
		return vs.Parent.HasTable(name)
	}
	return false
}

// HasColumn checks if a column is visible in this scope
func (vs *ValidationScope) HasColumn(table, column string) bool {
	key := fmt.Sprintf("%s.%s", table, column)
	if _, ok := vs.Columns[key]; ok {
		return true
	}
	if vs.Parent != nil {
		return vs.Parent.HasColumn(table, column)
	}
	return false
}

// AddTable adds a table to the scope
func (vs *ValidationScope) AddTable(name string, table *compiler.TableMetadata) {
	vs.Tables[name] = table
}

// AddColumn adds a column to the scope
func (vs *ValidationScope) AddColumn(table, column string, meta *compiler.ColumnMetadata) {
	key := fmt.Sprintf("%s.%s", table, column)
	vs.Columns[key] = meta
}

// AddAlias adds an alias mapping
func (vs *ValidationScope) AddAlias(alias, realName string) {
	vs.Aliases[alias] = realName
}

// AggregateMetadata stores information about aggregate functions
type AggregateMetadata struct {
	// Aggregate functions found in query
	Functions []*AggregateFunctionInfo

	// Positions where aggregates appear
	InSelect  bool
	InHaving  bool
	InOrderBy bool

	// GROUP BY related
	RequiresGroupBy bool
	GroupByColumns  []string

	// Validation state
	Validated       bool
	ValidationError error
}

// AggregateFunctionInfo describes an aggregate function usage
type AggregateFunctionInfo struct {
	Name       string              // COUNT, SUM, AVG, MAX, MIN
	Arguments  []parser.Expression // Function arguments
	ReturnType compiler.DataType   // Return type
	IsDistinct bool                // COUNT(DISTINCT ...)
	Position   ExpressionLocation  // Where it appears
}

// ExpressionLocation indicates where an expression appears
type ExpressionLocation int

const (
	LocationUnknown ExpressionLocation = iota
	LocationSelect
	LocationWhere
	LocationGroupBy
	LocationHaving
	LocationOrderBy
)

func (el ExpressionLocation) String() string {
	switch el {
	case LocationSelect:
		return "SELECT"
	case LocationWhere:
		return "WHERE"
	case LocationGroupBy:
		return "GROUP BY"
	case LocationHaving:
		return "HAVING"
	case LocationOrderBy:
		return "ORDER BY"
	default:
		return "UNKNOWN"
	}
}

// GroupByMetadata stores GROUP BY validation information
type GroupByMetadata struct {
	// GROUP BY expressions
	Expressions []parser.Expression

	// Columns referenced in GROUP BY
	Columns []string

	// Non-grouped columns in SELECT
	NonGroupedCols []string

	// Aggregate functions used
	Aggregates []string

	// HAVING clause info
	HasHaving  bool
	HavingExpr parser.Expression

	// Validation results
	Valid  bool
	Errors []string
}

// SubqueryMetadata describes a subquery
type SubqueryMetadata struct {
	// Subquery statement
	Query *parser.SelectStatement

	// Subquery type
	Type SubqueryType

	// Location in parent query
	Location ExpressionLocation

	// Correlation info
	IsCorrelated   bool
	CorrelatedRefs []CorrelatedReference

	// Result info
	ColumnCount  int
	ExpectedRows RowCountExpectation

	// Validation
	Valid bool
	Error error
}

// SubqueryType indicates the type of subquery
type SubqueryType int

const (
	SubqueryScalar       SubqueryType = iota // (SELECT x FROM t)
	SubqueryInPredicate                      // col IN (SELECT x FROM t)
	SubqueryExists                           // EXISTS (SELECT ...)
	SubqueryDerivedTable                     // FROM (SELECT ...) AS t
)

func (st SubqueryType) String() string {
	switch st {
	case SubqueryScalar:
		return "SCALAR"
	case SubqueryInPredicate:
		return "IN_PREDICATE"
	case SubqueryExists:
		return "EXISTS"
	case SubqueryDerivedTable:
		return "DERIVED_TABLE"
	default:
		return "UNKNOWN"
	}
}

// RowCountExpectation indicates expected row count
type RowCountExpectation int

const (
	RowCountAny       RowCountExpectation = iota // Any number of rows
	RowCountSingle                               // Must be exactly 1 row
	RowCountZeroOrOne                            // 0 or 1 rows
)

// CorrelatedReference describes a correlated column reference
type CorrelatedReference struct {
	Column     string
	Table      string
	OuterScope int // Nesting level
}
