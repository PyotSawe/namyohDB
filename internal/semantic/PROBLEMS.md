# Problems Solved by Semantic Analyzer

## Problem 1: GROUP BY Semantic Validation

### The Problem
SQL's GROUP BY has complex semantic rules that must be enforced:
- Every non-aggregate expression in SELECT must appear in GROUP BY
- HAVING clause can only reference aggregates or GROUP BY columns
- ORDER BY follows similar rules
- These rules prevent ambiguous query results

**Example of Invalid Query**:
```sql
SELECT dept, name, COUNT(*)
FROM employees
GROUP BY dept
-- ERROR: 'name' not in GROUP BY and not in aggregate
```

**Why It's Invalid**:
- Multiple employees can have the same dept
- Which 'name' should be returned? First? Last? Random?
- SQL requires explicit specification via GROUP BY or aggregate

### Solution Approach

**Algorithm**:
1. Extract all GROUP BY expressions
2. Walk SELECT list expressions
3. For each non-aggregate expression:
   - Check if it appears in GROUP BY
   - Check if it's a constant
   - If neither, report error
4. Validate HAVING and ORDER BY similarly

**Implementation**:
```go
func (v *GroupByValidator) ValidateSelect(stmt *SelectStatement) error {
    groupByExprs := stmt.GroupBy
    
    for _, selectExpr := range stmt.SelectClause.Columns {
        if v.isAggregate(selectExpr) {
            continue
        }
        
        if !v.containsExpression(groupByExprs, selectExpr) {
            if !v.isConstant(selectExpr) {
                return fmt.Errorf(
                    "column %v must appear in GROUP BY or be in aggregate",
                    selectExpr)
            }
        }
    }
    
    return nil
}
```

**Complexity**: O(s × g) where s = SELECT expressions, g = GROUP BY expressions

### Edge Cases Handled

1. **Constants are OK**:
   ```sql
   SELECT dept, 'Manager', COUNT(*)
   FROM employees
   GROUP BY dept
   -- 'Manager' is constant, allowed
   ```

2. **Complex expressions**:
   ```sql
   SELECT dept, UPPER(name), COUNT(*)
   FROM employees
   GROUP BY dept, UPPER(name)
   -- UPPER(name) must match exactly in GROUP BY
   ```

3. **Multiple aggregates**:
   ```sql
   SELECT dept, COUNT(*), AVG(salary), MAX(salary)
   FROM employees
   GROUP BY dept
   -- Multiple aggregates OK
   ```

### Benefits
- Prevents ambiguous query results
- Ensures predictable behavior
- Matches SQL standard semantics
- Clear error messages guide users

---

## Problem 2: Aggregate Function Placement Validation

### The Problem
Aggregate functions have strict placement rules in SQL:
- Cannot appear in WHERE clause (filtered before grouping)
- Cannot appear in GROUP BY clause (circular dependency)
- Can appear in SELECT, HAVING, ORDER BY
- Cannot be nested: `COUNT(SUM(x))` is invalid

**Example of Invalid Queries**:
```sql
-- Error 1: Aggregate in WHERE
SELECT * FROM employees WHERE COUNT(*) > 5

-- Error 2: Nested aggregate
SELECT dept, COUNT(SUM(salary)) FROM employees GROUP BY dept

-- Error 3: Aggregate in GROUP BY
SELECT dept, COUNT(*) FROM employees GROUP BY dept, COUNT(*)
```

### Solution Approach

**Context Tracking**:
```go
type ValidationContext struct {
    InWhereClause   bool
    InGroupByClause bool
    InHavingClause  bool
    AggregateDepth  int
}
```

**Validation Logic**:
```go
func (v *AggregateValidator) ValidateFunction(fn *FunctionCall, ctx *Context) error {
    if !isAggregate(fn.Name) {
        return nil
    }
    
    // Check placement
    if ctx.InWhereClause {
        return fmt.Errorf("aggregate not allowed in WHERE clause")
    }
    
    if ctx.InGroupByClause {
        return fmt.Errorf("aggregate not allowed in GROUP BY clause")
    }
    
    // Check nesting
    ctx.AggregateDepth++
    if ctx.AggregateDepth > 1 {
        return fmt.Errorf("nested aggregates not allowed")
    }
    
    // Validate arguments
    for _, arg := range fn.Arguments {
        if err := v.ValidateExpression(arg, ctx); err != nil {
            return err
        }
    }
    
    ctx.AggregateDepth--
    return nil
}
```

### Why WHERE Cannot Have Aggregates

**Execution Order**:
```
FROM → WHERE → GROUP BY → HAVING → SELECT → ORDER BY

WHERE happens before GROUP BY
↓
Groups don't exist yet
↓
Cannot compute aggregates yet
```

**Correct Usage**:
```sql
-- Use HAVING instead
SELECT dept, COUNT(*) as cnt
FROM employees
GROUP BY dept
HAVING COUNT(*) > 5
```

### Type-Specific Validation

**SUM/AVG require numeric**:
```go
if fn.Name == "SUM" || fn.Name == "AVG" {
    argType := inferType(fn.Arguments[0])
    if !argType.IsNumeric() {
        return fmt.Errorf("%s requires numeric argument", fn.Name)
    }
}
```

**COUNT accepts anything**:
```go
if fn.Name == "COUNT" {
    // COUNT(*) or COUNT(expr) both valid
    return nil
}
```

### Benefits
- Enforces SQL execution semantics
- Prevents meaningless queries
- Clear error messages explain why
- Suggests correct alternative (HAVING vs WHERE)

---

## Problem 3: Subquery Semantic Validation

### The Problem
Subqueries have different requirements based on context:
- **Scalar subquery**: Must return 1 row × 1 column
- **IN predicate**: Must return N rows × 1 column
- **EXISTS predicate**: Can return any shape
- **Derived table**: Can return any shape but needs alias

**Example Issues**:
```sql
-- Error: Scalar subquery returns multiple columns
SELECT (SELECT id, name FROM users WHERE id = 1)

-- Error: IN subquery returns multiple columns
SELECT * FROM orders WHERE user_id IN (SELECT id, name FROM users)

-- Error: Derived table without alias
SELECT * FROM (SELECT * FROM users)
```

### Solution Approach

**Subquery Type Detection**:
```go
type SubqueryType int

const (
    SubqueryScalar        // (SELECT x FROM t)
    SubqueryInPredicate   // col IN (SELECT x FROM t)
    SubqueryExists        // EXISTS (SELECT ...)
    SubqueryDerivedTable  // FROM (SELECT ...) AS t
)

func detectSubqueryType(parent Expression) SubqueryType {
    switch parent.(type) {
    case *BinaryExpression:
        if parent.Operator == IN {
            return SubqueryInPredicate
        }
        return SubqueryScalar
    case *FunctionCall:
        if parent.Name == "EXISTS" {
            return SubqueryExists
        }
    case *TableExpression:
        return SubqueryDerivedTable
    }
    return SubqueryScalar
}
```

**Validation Logic**:
```go
func (v *SubqueryValidator) Validate(subquery *SelectStatement, stype SubqueryType) error {
    columnCount := len(subquery.SelectClause.Columns)
    
    switch stype {
    case SubqueryScalar:
        if columnCount != 1 {
            return fmt.Errorf("scalar subquery must return single column, got %d", columnCount)
        }
        // Row count checked at runtime
        
    case SubqueryInPredicate:
        if columnCount != 1 {
            return fmt.Errorf("IN subquery must return single column, got %d", columnCount)
        }
        
    case SubqueryExists:
        // Any shape OK
        
    case SubqueryDerivedTable:
        if subquery.Alias == "" {
            return fmt.Errorf("derived table must have alias")
        }
    }
    
    return nil
}
```

### Correlated Subquery Validation

**Problem**: Correlated references must be visible in outer scope

```sql
SELECT name,
       (SELECT COUNT(*) 
        FROM orders 
        WHERE orders.user_id = users.id)  -- Correlated
FROM users
```

**Solution**:
```go
type CorrelatedReference struct {
    Column     string
    Table      string
    OuterScope int
}

func (v *SubqueryValidator) ValidateCorrelation(ref *ColumnReference, ctx *Context) error {
    // Check if reference exists in outer scopes
    for i := len(ctx.OuterScopes) - 1; i >= 0; i-- {
        scope := ctx.OuterScopes[i]
        if scope.HasColumn(ref.Table, ref.Column) {
            return nil
        }
    }
    
    return fmt.Errorf("correlated reference %s.%s not visible", ref.Table, ref.Column)
}
```

### Benefits
- Prevents runtime errors (scalar subquery returning multiple rows)
- Clear compile-time errors for shape mismatches
- Validates correlated references
- Ensures derived tables are usable (have aliases)

---

## Problem 4: Schema Constraint Validation

### The Problem
DDL statements must enforce schema integrity:
- No duplicate column names in CREATE TABLE
- Foreign key references must exist
- Cannot create duplicate tables (without IF NOT EXISTS)
- Cannot drop tables with dependencies (without CASCADE)
- CHECK constraints must be valid boolean expressions

**Example Issues**:
```sql
-- Error: Duplicate column
CREATE TABLE users (
    id INTEGER,
    name TEXT,
    id INTEGER  -- Duplicate!
)

-- Error: Foreign key references non-existent table
CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER REFERENCES users(id)  -- users doesn't exist
)

-- Error: Circular dependency
CREATE TABLE a (id INT REFERENCES b(id))
CREATE TABLE b (id INT REFERENCES a(id))
```

### Solution Approach

**Duplicate Column Detection**:
```go
func (v *SchemaValidator) ValidateCreateTable(stmt *CreateTableStatement) error {
    seenColumns := make(map[string]bool)
    
    for _, col := range stmt.Columns {
        if seenColumns[col.Name] {
            return fmt.Errorf("duplicate column name: %s", col.Name)
        }
        seenColumns[col.Name] = true
    }
    
    return nil
}
```

**Foreign Key Validation**:
```go
func (v *SchemaValidator) ValidateForeignKeys(stmt *CreateTableStatement) error {
    for _, constraint := range stmt.Constraints {
        if constraint.Type != ConstraintForeignKey {
            continue
        }
        
        fk := constraint.ForeignKey
        
        // Check referenced table exists
        if !v.catalog.TableExists(fk.ReferencedTable) {
            return fmt.Errorf(
                "foreign key references non-existent table: %s",
                fk.ReferencedTable)
        }
        
        // Check referenced columns exist
        refTable, _ := v.catalog.GetTable(fk.ReferencedTable)
        for _, col := range fk.ReferencedColumns {
            if !refTable.HasColumn(col) {
                return fmt.Errorf(
                    "foreign key references non-existent column: %s.%s",
                    fk.ReferencedTable, col)
            }
        }
    }
    
    return nil
}
```

**Circular Dependency Detection**:
```go
func (v *SchemaValidator) CheckCircularDependencies(tableName string, visited map[string]bool) error {
    if visited[tableName] {
        return fmt.Errorf("circular dependency detected involving table: %s", tableName)
    }
    
    visited[tableName] = true
    
    // Get all foreign keys from this table
    table, _ := v.catalog.GetTable(tableName)
    for _, fk := range table.ForeignKeys {
        if err := v.CheckCircularDependencies(fk.ReferencedTable, visited); err != nil {
            return err
        }
    }
    
    delete(visited, tableName)
    return nil
}
```

### CHECK Constraint Validation

**Problem**: CHECK constraints must be boolean expressions

```sql
-- Valid
CREATE TABLE users (
    age INTEGER CHECK (age >= 18)
)

-- Invalid: not boolean
CREATE TABLE users (
    age INTEGER CHECK (age + 10)  -- Returns integer, not boolean
)
```

**Solution**:
```go
func (v *SchemaValidator) ValidateCheckConstraint(constraint *CheckConstraint) error {
    exprType, err := v.typeChecker.InferType(constraint.Expression)
    if err != nil {
        return err
    }
    
    if exprType != DataTypeBoolean {
        return fmt.Errorf(
            "CHECK constraint must be boolean expression, got %s",
            exprType)
    }
    
    return nil
}
```

### Benefits
- Prevents schema corruption
- Ensures referential integrity setup is valid
- Catches errors early (compile-time vs runtime)
- Provides clear error messages with table/column names

---

## Problem 5: Expression Context Validation

### The Problem
SQL expressions have context-dependent validity rules:
- Window functions only in SELECT/ORDER BY
- Column aliases from SELECT not visible in WHERE
- Correlation names must be unique within scope
- Expression depth limits (prevent stack overflow)

**Example Issues**:
```sql
-- Error: Alias not visible in WHERE
SELECT salary * 1.1 AS new_salary
FROM employees
WHERE new_salary > 50000  -- new_salary not visible here!

-- Correct version
SELECT salary * 1.1 AS new_salary
FROM employees
WHERE salary * 1.1 > 50000
```

### Solution Approach

**Scope Management**:
```go
type ValidationScope struct {
    Tables    map[string]*TableMetadata
    Columns   map[string]*ColumnMetadata
    Aliases   map[string]Expression  // SELECT clause aliases
    Parent    *ValidationScope
}

func (s *ValidationScope) CanSeeAlias(name string, inClause ClauseType) bool {
    // Aliases only visible in ORDER BY and later
    if inClause == ClauseWhere || inClause == ClauseGroupBy {
        return false
    }
    
    _, exists := s.Aliases[name]
    return exists
}
```

**Expression Depth Limiting**:
```go
const MaxExpressionDepth = 100

func (v *ExpressionValidator) ValidateExpression(expr Expression, depth int) error {
    if depth > MaxExpressionDepth {
        return fmt.Errorf("expression too deeply nested (max: %d)", MaxExpressionDepth)
    }
    
    switch e := expr.(type) {
    case *BinaryExpression:
        if err := v.ValidateExpression(e.Left, depth+1); err != nil {
            return err
        }
        if err := v.ValidateExpression(e.Right, depth+1); err != nil {
            return err
        }
    case *UnaryExpression:
        if err := v.ValidateExpression(e.Operand, depth+1); err != nil {
            return err
        }
    case *FunctionCall:
        for _, arg := range e.Arguments {
            if err := v.ValidateExpression(arg, depth+1); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### SQL Clause Visibility Rules

```
Clause Execution Order:
FROM → WHERE → GROUP BY → HAVING → SELECT → ORDER BY

Visibility:
- WHERE can see: FROM tables/columns
- GROUP BY can see: FROM tables/columns
- HAVING can see: FROM tables/columns, GROUP BY expressions
- SELECT can see: FROM tables/columns, GROUP BY expressions
- ORDER BY can see: FROM tables/columns, SELECT aliases
```

**Implementation**:
```go
type ClauseType int

const (
    ClauseFrom ClauseType = iota
    ClauseWhere
    ClauseGroupBy
    ClauseHaving
    ClauseSelect
    ClauseOrderBy
)

func (v *ScopeValidator) ValidateReference(ref *ColumnReference, clause ClauseType) error {
    // Check if it's a table column
    if v.scope.HasColumn(ref.Table, ref.Column) {
        return nil
    }
    
    // Check if it's a SELECT alias
    if clause == ClauseOrderBy {
        if _, exists := v.scope.Aliases[ref.Column]; exists {
            return nil
        }
    }
    
    return fmt.Errorf("column %s not visible in %s clause", ref.Column, clause)
}
```

### Benefits
- Enforces SQL's logical execution model
- Prevents confusing errors (aliases in WHERE)
- Protects against stack overflow attacks
- Matches standard SQL behavior

---

## Design Challenges and Solutions

### Challenge 1: Error Collection vs Fail-Fast

**Problem**: Should we stop at first error or collect all errors?

**Solution**: Collect all non-critical errors
```go
type ErrorCollector struct {
    errors   []SemanticError
    warnings []SemanticWarning
    
    criticalErrorFound bool
}

func (ec *ErrorCollector) AddError(err SemanticError) {
    ec.errors = append(ec.errors, err)
    
    // Critical errors stop analysis
    if err.IsCritical() {
        ec.criticalErrorFound = true
    }
}

func (ec *ErrorCollector) ShouldContinue() bool {
    return !ec.criticalErrorFound
}
```

**Benefits**:
- Users see all errors at once
- Better developer experience
- Faster iteration (fix multiple issues together)

### Challenge 2: Performance with Large Schemas

**Problem**: Validating queries against schemas with thousands of tables is slow

**Solution**: Lazy validation + caching
```go
type CachedCatalog struct {
    underlying CatalogManager
    cache      map[string]*TableMetadata
    mu         sync.RWMutex
}

func (cc *CachedCatalog) GetTable(name string) (*TableMetadata, error) {
    cc.mu.RLock()
    if table, ok := cc.cache[name]; ok {
        cc.mu.RUnlock()
        return table, nil
    }
    cc.mu.RUnlock()
    
    // Cache miss - fetch and cache
    table, err := cc.underlying.GetTable(name)
    if err != nil {
        return nil, err
    }
    
    cc.mu.Lock()
    cc.cache[name] = table
    cc.mu.Unlock()
    
    return table, nil
}
```

### Challenge 3: Extensibility for Custom Rules

**Problem**: Different databases have different semantic rules

**Solution**: Plugin architecture
```go
type SemanticRule interface {
    Name() string
    Validate(query *CompiledQuery, ctx *ValidationContext) error
}

type SemanticAnalyzer struct {
    rules []SemanticRule
}

func (sa *SemanticAnalyzer) AddRule(rule SemanticRule) {
    sa.rules = append(sa.rules, rule)
}

func (sa *SemanticAnalyzer) Analyze(query *CompiledQuery) (*SemanticInfo, error) {
    ctx := NewValidationContext(query)
    
    for _, rule := range sa.rules {
        if err := rule.Validate(query, ctx); err != nil {
            ctx.AddError(err)
        }
    }
    
    return ctx.BuildSemanticInfo(), nil
}
```

---

## Summary

The Semantic Analyzer solves five major categories of problems:

1. **GROUP BY Validation**: Ensures non-ambiguous grouping semantics
2. **Aggregate Placement**: Enforces SQL execution order constraints
3. **Subquery Validation**: Ensures shape and correlation correctness
4. **Schema Constraints**: Maintains schema integrity
5. **Expression Context**: Enforces visibility and scoping rules

These solutions enable:
- ✅ Compile-time error detection (catch bugs early)
- ✅ Clear, actionable error messages
- ✅ SQL standard compliance
- ✅ Predictable query behavior
- ✅ Performance optimization opportunities

The analyzer builds on the Query Compiler's name resolution and type checking to provide comprehensive semantic validation before query optimization and execution.
