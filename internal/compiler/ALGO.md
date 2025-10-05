# Query Compiler Algorithms

## Overview
The Query Compiler transforms parsed Abstract Syntax Trees (AST) into validated, compiled query representations. It performs name resolution, type checking, and constraint validation following Derby's compilation pipeline.

## Core Algorithms

### 1. Query Compilation Pipeline

```
Input: parser.Statement (AST)
Output: CompiledQuery (validated representation)

ALGORITHM: CompileQuery(ast)
  1. Identify query type (SELECT, INSERT, UPDATE, DELETE, DDL)
  2. Resolve names (tables, columns, aliases)
  3. Check and infer types for all expressions
  4. Validate constraints and semantics
  5. Build compiled representation
  6. Return CompiledQuery or error
```

**Time Complexity**: O(n) where n is the number of AST nodes
**Space Complexity**: O(n) for storing compiled metadata

---

### 2. Name Resolution Algorithm

```
ALGORITHM: ResolveNames(ast, catalog)
  INPUT: AST node, CatalogManager
  OUTPUT: Map of resolved references, or error
  
  1. Initialize resolvedRefs = empty map
  2. Extract table references from FROM clause
  3. For each table reference:
     a. Lookup table in catalog
     b. If not found, return "table not found" error
     c. Store TableMetadata in resolvedRefs
  4. For each column reference in SELECT, WHERE, etc:
     a. Determine which table(s) it could belong to
     b. If ambiguous, return "ambiguous column" error
     c. If not found in any table, return "column not found" error
     d. Store resolved column metadata
  5. Resolve aliases and qualified names (table.column)
  6. Return resolvedRefs
```

**Key Cases**:
- **Unqualified columns**: `SELECT name FROM users` → resolve `name` to `users.name`
- **Qualified columns**: `SELECT u.name FROM users u` → verify `u` alias exists
- **Ambiguous references**: `SELECT id FROM users, orders` → ERROR if both have `id`
- **Wildcards**: `SELECT * FROM users` → expand to all columns

**Time Complexity**: O(t × c) where t = tables, c = columns per table

---

### 3. Type Inference and Checking Algorithm

```
ALGORITHM: CheckTypes(ast, resolvedRefs)
  INPUT: AST node, resolved name references
  OUTPUT: Map of type information, or error
  
  1. Initialize typeInfo = empty map
  2. For each expression in AST:
     a. Infer type using InferExpressionType()
     b. Store in typeInfo map
  3. For each operation (binary, unary, function call):
     a. Check operand types are compatible
     b. If not compatible, return "type mismatch" error
     c. Determine result type
  4. For assignments (INSERT, UPDATE):
     a. Check value type matches column type
     b. Handle implicit conversions (INTEGER → REAL)
     c. Reject incompatible assignments (TEXT → INTEGER)
  5. Return typeInfo
```

**Type Inference Rules**:
```
InferExpressionType(expr):
  CASE expr.Type OF:
    LiteralExpression:
      - NUMBER → INTEGER or REAL (based on format)
      - STRING → TEXT
      - TRUE/FALSE → BOOLEAN
      - NULL → NULL type
    
    ColumnExpression:
      - Lookup column type from resolvedRefs
    
    BinaryExpression:
      - Arithmetic (+, -, *, /): 
        * INTEGER op INTEGER → INTEGER
        * INTEGER op REAL → REAL
        * REAL op REAL → REAL
      - Comparison (=, <, >, <=, >=, !=):
        * Compatible types → BOOLEAN
      - Logical (AND, OR):
        * BOOLEAN op BOOLEAN → BOOLEAN
    
    UnaryExpression:
      - NOT: BOOLEAN → BOOLEAN
      - Minus: NUMBER → NUMBER (preserve type)
    
    FunctionCall:
      - COUNT(*) → INTEGER
      - SUM(NUMBER) → NUMBER (preserve type)
      - AVG(NUMBER) → REAL
      - MAX/MIN(T) → T (preserve operand type)
```

**Time Complexity**: O(e) where e = number of expressions in query

---

### 4. Constraint Validation Algorithm

```
ALGORITHM: ValidateConstraints(ast, resolvedRefs, catalog)
  INPUT: AST, resolved references, catalog manager
  OUTPUT: Success or constraint violation error
  
  1. For CREATE TABLE statements:
     a. Check PRIMARY KEY uniqueness (only one per table)
     b. Validate FOREIGN KEY references:
        - Referenced table exists
        - Referenced columns exist
        - Column types match
     c. Check CHECK constraint expressions are BOOLEAN
     d. Verify NOT NULL columns don't have NULL defaults
  
  2. For INSERT statements:
     a. Check all NOT NULL columns have values
     b. Verify PRIMARY KEY uniqueness (requires catalog lookup)
     c. Validate FOREIGN KEY constraints
     d. Evaluate CHECK constraints
  
  3. For UPDATE statements:
     a. Check PRIMARY KEY immutability (if PK is updated)
     b. Validate new values against constraints
     c. Check FOREIGN KEY integrity
  
  4. For DELETE statements:
     a. Check FOREIGN KEY cascading rules
     b. Verify no dependent rows exist (if not CASCADE)
  
  5. Return success or detailed constraint violation
```

**Constraint Types**:
- **NOT NULL**: Column cannot contain NULL values
- **PRIMARY KEY**: Unique identifier, implicitly NOT NULL
- **UNIQUE**: Column values must be unique across rows
- **FOREIGN KEY**: References another table's PRIMARY KEY
- **CHECK**: Boolean expression must evaluate to TRUE
- **DEFAULT**: Provides default value if not specified

**Time Complexity**: O(c × r) where c = constraints, r = rows (for INSERT/UPDATE)

---

### 5. Query Type Identification

```
ALGORITHM: IdentifyQueryType(ast)
  INPUT: AST statement
  OUTPUT: QueryType enum
  
  MATCH ast.Type:
    SelectStatement    → SELECT_QUERY
    InsertStatement    → INSERT_COMMAND
    UpdateStatement    → UPDATE_COMMAND
    DeleteStatement    → DELETE_COMMAND
    CreateTableStmt    → DDL_CREATE
    DropTableStmt      → DDL_DROP
    CreateIndexStmt    → DDL_CREATE_INDEX
    BeginTransaction   → TRANSACTION_BEGIN
    CommitTransaction  → TRANSACTION_COMMIT
    RollbackTransaction→ TRANSACTION_ROLLBACK
  
  Return QueryType
```

---

## Advanced Algorithms

### 6. Aggregate Function Validation

```
ALGORITHM: ValidateAggregates(selectStmt)
  INPUT: SELECT statement AST
  OUTPUT: Success or aggregation error
  
  1. Collect all aggregate functions (COUNT, SUM, AVG, MAX, MIN)
  2. Collect all non-aggregate column references
  3. If has GROUP BY clause:
     a. All non-aggregate columns must appear in GROUP BY
     b. GROUP BY expressions can't contain aggregates
  4. If has aggregates but no GROUP BY:
     a. ALL columns in SELECT must be aggregates
     b. No plain column references allowed
  5. If has HAVING clause:
     a. HAVING can only reference aggregates or GROUP BY columns
     b. Validate HAVING expression is BOOLEAN
  
  Return success or error
```

**Examples**:
```sql
-- ✅ Valid: All non-aggregates in GROUP BY
SELECT dept, COUNT(*) FROM emp GROUP BY dept

-- ❌ Invalid: salary not in GROUP BY
SELECT dept, salary, COUNT(*) FROM emp GROUP BY dept

-- ✅ Valid: All SELECT items are aggregates
SELECT COUNT(*), AVG(salary) FROM emp

-- ❌ Invalid: Mixed aggregates and non-aggregates without GROUP BY
SELECT name, COUNT(*) FROM emp
```

---

### 7. Subquery Validation

```
ALGORITHM: ValidateSubquery(subquery, parentContext)
  INPUT: Subquery AST, parent query context
  OUTPUT: Success or subquery error
  
  1. Recursively compile subquery (full compilation)
  2. Check subquery cardinality:
     a. Scalar subquery (SELECT x) → must return 1 row, 1 column
     b. IN subquery (WHERE x IN (SELECT...)) → any rows, 1 column
     c. EXISTS subquery → any rows, any columns
  3. Validate correlated references:
     a. If subquery references parent columns
     b. Verify parent columns are in scope
     c. Check correlation doesn't create circular dependency
  4. Type check subquery result:
     a. For comparisons: subquery type matches comparison operand
     b. For IN: subquery column type matches test expression
  
  Return compiled subquery or error
```

---

### 8. Type Coercion and Compatibility

```
ALGORITHM: CheckTypeCompatibility(type1, type2, operation)
  INPUT: Two data types, operation context
  OUTPUT: Result type or incompatibility error
  
  # Compatibility matrix
  COMPATIBLE_TYPES = {
    (INTEGER, INTEGER): INTEGER,
    (INTEGER, REAL): REAL,      # Coerce INTEGER → REAL
    (REAL, INTEGER): REAL,      # Coerce INTEGER → REAL
    (REAL, REAL): REAL,
    (TEXT, TEXT): TEXT,
    (BOOLEAN, BOOLEAN): BOOLEAN
  }
  
  IF operation in {+, -, *, /}:  # Arithmetic
    IF (type1, type2) in COMPATIBLE_TYPES:
      Return COMPATIBLE_TYPES[(type1, type2)]
    ELSE:
      Return "incompatible types for arithmetic" error
  
  IF operation in {=, !=, <, >, <=, >=}:  # Comparison
    IF type1 and type2 are comparable:
      Return BOOLEAN
    ELSE:
      Return "incompatible types for comparison" error
  
  IF operation is ASSIGNMENT:  # INSERT, UPDATE
    IF type1 == type2:
      Return OK
    IF can_coerce(type2, type1):  # INTEGER → REAL allowed
      Return OK with coercion
    ELSE:
      Return "type mismatch" error
```

**Coercion Rules**:
- **INTEGER → REAL**: Always allowed (widening)
- **REAL → INTEGER**: Requires explicit CAST (narrowing)
- **TEXT → NUMBER**: Requires explicit CAST
- **NULL → Any Type**: Always allowed

---

## Error Handling Strategy

### Error Categories

1. **Name Resolution Errors**:
   - `ERR_TABLE_NOT_FOUND`: Referenced table doesn't exist
   - `ERR_COLUMN_NOT_FOUND`: Referenced column doesn't exist
   - `ERR_AMBIGUOUS_COLUMN`: Column name matches multiple tables
   - `ERR_INVALID_ALIAS`: Alias reference doesn't exist

2. **Type Errors**:
   - `ERR_TYPE_MISMATCH`: Incompatible types in operation
   - `ERR_INVALID_OPERAND`: Wrong type for operator
   - `ERR_INVALID_FUNCTION_ARG`: Function argument type mismatch
   - `ERR_CANNOT_COERCE`: Type coercion not possible

3. **Constraint Errors**:
   - `ERR_NOT_NULL_VIOLATION`: NULL value in NOT NULL column
   - `ERR_PRIMARY_KEY_DUPLICATE`: Duplicate PRIMARY KEY constraint
   - `ERR_FOREIGN_KEY_INVALID`: Invalid FOREIGN KEY reference
   - `ERR_CHECK_VIOLATION`: CHECK constraint expression failed

4. **Semantic Errors**:
   - `ERR_INVALID_AGGREGATE`: Invalid aggregate function usage
   - `ERR_INVALID_GROUP_BY`: GROUP BY validation failed
   - `ERR_INVALID_SUBQUERY`: Subquery cardinality mismatch
   - `ERR_CIRCULAR_REFERENCE`: Circular dependency detected

---

## Performance Optimizations

### 1. Caching Strategy
- **Schema Cache**: Cache table/column metadata to avoid repeated catalog lookups
- **Type Cache**: Cache inferred types for common expressions
- **Validation Cache**: Cache constraint validation results

### 2. Early Exit
- **Fail Fast**: Return error on first validation failure (configurable)
- **Lazy Evaluation**: Only validate necessary parts for simple queries

### 3. Parallel Validation
- **Independent Checks**: Type checking and constraint validation can run in parallel
- **Goroutine Pool**: Use worker pool for validating multiple subqueries

---

## Integration with Other Layers

### Input from Parser
```go
// Parser provides AST
ast := parser.ParseSQL(sql)
```

### Output to Semantic Analyzer
```go
// Compiler provides validated query
compiled := compiler.Compile(ast)
```

### Interaction with Catalog
```go
// Compiler queries catalog for metadata
table := catalog.GetTable("users")
column := catalog.GetColumn("users", "name")
```

---

## Testing Strategy

### Unit Test Coverage
1. **Name Resolution**: Test all resolution cases (qualified, unqualified, aliases)
2. **Type Checking**: Test all type combinations and coercions
3. **Constraint Validation**: Test each constraint type independently
4. **Error Handling**: Test all error conditions with specific error messages

### Integration Tests
1. **End-to-End Compilation**: Full compilation of complex queries
2. **Catalog Integration**: Test with actual catalog manager
3. **Performance Tests**: Measure compilation time for large schemas

---

## Algorithm Complexity Summary

| Algorithm | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|-------|
| CompileQuery | O(n) | O(n) | n = AST nodes |
| ResolveNames | O(t × c) | O(t + c) | t = tables, c = columns |
| CheckTypes | O(e) | O(e) | e = expressions |
| ValidateConstraints | O(c × r) | O(c) | c = constraints, r = rows |
| ValidateAggregates | O(s) | O(s) | s = SELECT items |
| ValidateSubquery | O(n²) | O(d) | d = nesting depth |

---

## Future Enhancements

1. **Incremental Compilation**: Cache partial results for repeated queries
2. **Query Fingerprinting**: Identify structurally similar queries
3. **Advanced Type Inference**: ML-based type prediction
4. **Constraint Dependency Graph**: Optimize constraint validation order
5. **Parallel Compilation**: Compile independent parts concurrently

---

*This document describes the core algorithms used in the Query Compiler module of NamyohDB.*
