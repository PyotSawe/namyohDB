# SQL Compiler Layer - Implementation Roadmap

## 🎯 Current Status: Parser COMPLETE (14/15 tests passing - 93%)

### Parser Implementation Status ✅
- ✅ SELECT statements (all variants)
- ✅ INSERT statements  
- ✅ UPDATE statements
- ✅ DELETE statements
- ✅ CREATE TABLE statements
- ✅ DROP TABLE statements
- ✅ Expression parsing (with operator precedence)
- ⚠️  FROM clause currently optional (acceptable for now)

---

## 📋 SQL Compiler Layer Components (Derby-Style Architecture)

Based on your architecture (ARCH.md, Line 39), the SQL Compiler Layer consists of:

```
┌────────────────────────────────────────────────────────────┐
│              SQL Compiler Layer (Derby-Style)              │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  SQL Input ──▶ ┌──────────┐                               │
│                │  Lexer   │ ✅ COMPLETE                    │
│                └────┬─────┘                               │
│                     ▼                                      │
│                ┌──────────┐                               │
│                │  Parser  │ ✅ COMPLETE (14/15 tests)     │
│                └────┬─────┘                               │
│                     ▼                                      │
│         ┌──────────────────────┐                          │
│         │  Query Compiler      │ 🎯 NEXT (Week 1-2)      │
│         │  • AST Validation    │                          │
│         │  • Type Checking     │                          │
│         │  • Name Resolution   │                          │
│         └────┬─────────────────┘                          │
│              ▼                                             │
│         ┌──────────────────────┐                          │
│         │  Semantic Analyzer   │ 🎯 Week 2-3             │
│         │  • Schema Validation │                          │
│         │  • Constraint Check  │                          │
│         │  • Access Control    │                          │
│         └────┬─────────────────┘                          │
│              ▼                                             │
│         ┌──────────────────────┐                          │
│         │  Query Optimizer     │ 🎯 Week 3-4             │
│         │  • Cost-Based Plans  │                          │
│         │  • Index Selection   │                          │
│         │  • Join Reordering   │                          │
│         └──────────────────────┘                          │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 🚀 Implementation Order

### Phase 1: Query Compiler (Week 1-2) 🎯 START HERE

**Purpose**: Transform parsed AST into validated, compiled query representation

**Location**: `internal/compiler/`

**What to Implement**:

```go
// internal/compiler/compiler.go

type QueryCompiler struct {
    catalog CatalogManager  // For schema lookups
}

type CompiledQuery struct {
    Statement    parser.Statement
    QueryType    QueryType
    ResolvedRefs map[string]*TableRef
    TypeInfo     map[string]DataType
    Validated    bool
}

// Main compilation interface
func (qc *QueryCompiler) Compile(ast parser.Statement) (*CompiledQuery, error) {
    // Step 1: Identify query type
    queryType := identifyQueryType(ast)
    
    // Step 2: Resolve names (table names, column names)
    refs, err := qc.resolveNames(ast)
    if err != nil {
        return nil, err
    }
    
    // Step 3: Infer and check types
    types, err := qc.checkTypes(ast, refs)
    if err != nil {
        return nil, err
    }
    
    // Step 4: Validate constraints
    if err := qc.validateConstraints(ast, refs); err != nil {
        return nil, err
    }
    
    return &CompiledQuery{
        Statement:    ast,
        QueryType:    queryType,
        ResolvedRefs: refs,
        TypeInfo:     types,
        Validated:    true,
    }, nil
}

// Name resolution
func (qc *QueryCompiler) resolveNames(ast parser.Statement) (map[string]*TableRef, error) {
    // Resolve all table references
    // Resolve all column references
    // Check for ambiguous references
    // Verify all references exist in catalog
}

// Type checking
func (qc *QueryCompiler) checkTypes(ast parser.Statement, refs map[string]*TableRef) (map[string]DataType, error) {
    // Infer types for all expressions
    // Check type compatibility in operations
    // Validate function arguments
    // Check assignment compatibility (INSERT, UPDATE)
}

// Constraint validation
func (qc *QueryCompiler) validateConstraints(ast parser.Statement, refs map[string]*TableRef) error {
    // Check PRIMARY KEY uniqueness
    // Check FOREIGN KEY references
    // Check NOT NULL constraints
    // Check CHECK constraints
}
```

**Files to Create**:
1. `internal/compiler/compiler.go` - Main compiler
2. `internal/compiler/ALGO.md` - Compilation algorithms
3. `internal/compiler/DS.md` - Data structures
4. `internal/compiler/PROBLEMS.md` - Problems solved
5. `tests/unit/compiler_test.go` - Unit tests

**Dependencies**: Requires Catalog Manager (stores schema information)

---

### Phase 2: Semantic Analyzer (Week 2-3)

**Purpose**: Deep semantic validation beyond basic syntax

**Location**: `internal/semantic/`

**What to Implement**:

```go
// internal/semantic/analyzer.go

type SemanticAnalyzer struct {
    catalog CatalogManager
}

type AnalysisResult struct {
    Errors   []SemanticError
    Warnings []SemanticWarning
    Metadata map[string]interface{}
}

func (sa *SemanticAnalyzer) Analyze(query *CompiledQuery) (*AnalysisResult, error) {
    result := &AnalysisResult{}
    
    // Schema validation
    if err := sa.validateSchema(query); err != nil {
        result.Errors = append(result.Errors, err)
    }
    
    // Access control (permissions)
    if err := sa.checkPermissions(query); err != nil {
        result.Errors = append(result.Errors, err)
    }
    
    // Semantic checks
    if err := sa.checkSemantics(query); err != nil {
        result.Errors = append(result.Errors, err)
    }
    
    // Collect metadata for optimizer
    result.Metadata = sa.collectMetadata(query)
    
    return result, nil
}

func (sa *SemanticAnalyzer) validateSchema(query *CompiledQuery) error {
    // Verify all referenced tables exist
    // Verify all referenced columns exist
    // Check column types match operations
    // Validate aggregate function usage
}

func (sa *SemanticAnalyzer) checkSemantics(query *CompiledQuery) error {
    // GROUP BY validation (all non-aggregates must be in GROUP BY)
    // HAVING clause can only reference aggregates or GROUP BY columns
    // Subquery correlation validation
    // Window function validation
}
```

**Files to Create**:
1. `internal/semantic/analyzer.go`
2. `internal/semantic/ALGO.md`
3. `internal/semantic/DS.md`
4. `internal/semantic/PROBLEMS.md`
5. `tests/unit/semantic_test.go`

---

### Phase 3: Query Optimizer (Week 3-4)

**Purpose**: Generate optimal execution plans

**Location**: `internal/optimizer/`

**What to Implement**:

```go
// internal/optimizer/optimizer.go

type QueryOptimizer struct {
    catalog   CatalogManager
    statistics StatisticsManager
}

type ExecutionPlan struct {
    Root      PlanNode
    Cost      float64
    Operators []Operator
}

type PlanNode interface {
    Type() NodeType
    Children() []PlanNode
    EstimatedCost() float64
    EstimatedRows() int64
}

func (qo *QueryOptimizer) Optimize(query *CompiledQuery, analysis *AnalysisResult) (*ExecutionPlan, error) {
    // Step 1: Generate logical plan
    logicalPlan := qo.generateLogicalPlan(query)
    
    // Step 2: Apply transformation rules
    transformed := qo.applyRules(logicalPlan)
    
    // Step 3: Generate physical plans
    physicalPlans := qo.generatePhysicalPlans(transformed)
    
    // Step 4: Cost estimation
    bestPlan := qo.selectBestPlan(physicalPlans)
    
    return bestPlan, nil
}

// Optimization rules
func (qo *QueryOptimizer) applyRules(plan PlanNode) PlanNode {
    // Predicate pushdown (push WHERE close to data source)
    // Projection pushdown (select only needed columns early)
    // Join reordering (optimal join order)
    // Constant folding
    // Expression simplification
}

// Physical plan generation
func (qo *QueryOptimizer) generatePhysicalPlans(logical PlanNode) []ExecutionPlan {
    // For each logical operation, generate physical alternatives:
    // - Table Scan vs Index Scan
    // - Nested Loop Join vs Hash Join vs Merge Join
    // - In-memory Sort vs External Sort
}

// Cost-based selection
func (qo *QueryOptimizer) selectBestPlan(plans []ExecutionPlan) *ExecutionPlan {
    // Use statistics to estimate:
    // - Number of rows
    // - I/O cost
    // - CPU cost
    // - Memory cost
    // Select plan with minimum total cost
}
```

**Key Optimization Techniques** (From Derby):
1. **Predicate Pushdown**: Move WHERE clauses closer to data source
2. **Join Ordering**: Dynamic programming for optimal join sequence
3. **Index Selection**: Choose best indexes based on selectivity
4. **Subquery Optimization**: Decorrelate correlated subqueries
5. **Cost Estimation**: Statistics-based cost models

**Files to Create**:
1. `internal/optimizer/optimizer.go` - Main optimizer
2. `internal/optimizer/rules.go` - Transformation rules
3. `internal/optimizer/cost.go` - Cost estimation
4. `internal/optimizer/ALGO.md` - Optimization algorithms
5. `internal/optimizer/DS.md` - Plan data structures
6. `internal/optimizer/PROBLEMS.md` - Optimization problems
7. `tests/unit/optimizer_test.go` - Unit tests

---

## 📊 Timeline Summary

| Phase | Module | Duration | Dependencies | Status |
|-------|--------|----------|--------------|--------|
| 0 | Lexer | ✅ Done | None | Complete |
| 0 | Parser | ✅ Done | Lexer | 93% (14/15) |
| 1 | Query Compiler | Week 1-2 | Parser, Catalog | 🎯 Next |
| 2 | Semantic Analyzer | Week 2-3 | Compiler, Catalog | Pending |
| 3 | Query Optimizer | Week 3-4 | Analyzer, Statistics | Pending |
| 4 | Execution Engine | Week 5-7 | Optimizer, Storage | Future |

---

## 🎯 Immediate Next Steps

### Step 1: Complete Missing Dependencies

**Before implementing Query Compiler, we need**:

1. **Catalog Manager** (`internal/storage/catalog.go`)
   - Store table schemas
   - Retrieve table metadata
   - Column definitions lookup

```go
type CatalogManager interface {
    CreateTable(schema *TableSchema) error
    GetTable(name string) (*TableMetadata, error)
    GetColumn(table, column string) (*ColumnMetadata, error)
    ListTables() ([]string, error)
}
```

2. **Schema Definitions** (`internal/schema/`)
   - TableSchema, ColumnSchema types
   - Data type definitions
   - Constraint representations

### Step 2: Implement Query Compiler (This Week)

Follow the pattern:
1. Create `internal/compiler/ALGO.md` (algorithms)
2. Create `internal/compiler/DS.md` (data structures)
3. Create `internal/compiler/PROBLEMS.md` (problems solved)
4. Implement `internal/compiler/compiler.go`
5. Write comprehensive tests

### Step 3: Move to Semantic Analyzer (Next Week)

After compiler is working and tested.

### Step 4: Implement Query Optimizer (Week 3-4)

Cost-based optimization with statistics.

---

## 🏗️ After SQL Compiler Layer: Execution Engine

Once SQL Compiler Layer is complete, you'll move to:

```
┌──────────────────────────────────────────────────────────┐
│           Execution Engine Layer (SQLite3-Style)         │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  ┌───────────────┐   ┌───────────────┐                 │
│  │ Query         │   │ Result Set    │                 │
│  │ Executor      │──▶│ Builder       │                 │
│  └───────────────┘   └───────────────┘                 │
│         ↓                                                │
│  ┌───────────────┐   ┌───────────────┐                 │
│  │ Catalog       │   │ Transaction   │                 │
│  │ Manager       │   │ Manager       │                 │
│  └───────────────┘   └───────────────┘                 │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

This will implement actual query execution using the optimized plans.

---

## 📚 Architecture References

- **ARCH.md** Line 39: SQL Compiler Layer definition
- **ARCH.md** Lines 175-250: Query processing pipeline
- **FLOW.md**: Data flow through compilation layers
- **DATA.md**: AST transformations

---

## ✅ Success Criteria

### SQL Compiler Layer Complete When:
- ✅ Parser: 15/15 tests passing
- ✅ Query Compiler: Can compile all statement types
- ✅ Semantic Analyzer: Validates all semantic rules
- ✅ Query Optimizer: Generates optimized execution plans
- ✅ Integration: End-to-end compilation working
- ✅ Tests: >90% coverage for all modules

### Ready for Execution Engine When:
- ✅ Can transform SQL → AST → Compiled → Analyzed → Optimized
- ✅ Have execution plans ready for executor
- ✅ All schema validation working
- ✅ Cost estimates available for all operations

---

## 🚀 Start Implementation

**TODAY**: Begin with Catalog Manager (required dependency)
**WEEK 1**: Implement Query Compiler
**WEEK 2**: Implement Semantic Analyzer  
**WEEK 3-4**: Implement Query Optimizer
**WEEK 5+**: Move to Execution Engine Layer

**First Command**:
```bash
mkdir -p internal/compiler internal/semantic internal/storage
touch internal/compiler/{compiler.go,ALGO.md,DS.md,PROBLEMS.md}
touch internal/storage/catalog.go
```

Let's build the SQL Compiler Layer following Derby's proven architecture! 🎯
