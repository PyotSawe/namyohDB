# Semantic Analyzer Algorithms

## Overview
The Semantic Analyzer performs deep semantic validation beyond basic syntax and type checking. It ensures that queries are semantically correct according to SQL semantics, business rules, and database constraints.

## Core Algorithms

### 1. Semantic Analysis Pipeline

```
Algorithm: AnalyzeQuery(compiledQuery)
Input: CompiledQuery from Query Compiler
Output: SemanticInfo with validation results

1. Initialize SemanticInfo structure
2. context ← Extract context from compiled query
3. 
4. // Phase 1: Schema Validation
5. errors ← ValidateSchema(compiledQuery, context)
6. IF errors NOT empty THEN
7.     RETURN SemanticInfo with errors
8. END IF
9. 
10. // Phase 2: Semantic Rules Validation
11. errors ← ValidateSemanticRules(compiledQuery, context)
12. IF errors NOT empty THEN
13.     RETURN SemanticInfo with errors
14. END IF
15. 
16. // Phase 3: Access Control (if enabled)
17. IF access_control_enabled THEN
18.     errors ← ValidateAccessControl(compiledQuery, context)
19.     IF errors NOT empty THEN
20.         RETURN SemanticInfo with errors
21.     END IF
22. END IF
23. 
24. // Phase 4: Aggregate Semantics
25. IF IsAggregateQuery(compiledQuery) THEN
26.     errors ← ValidateAggregates(compiledQuery, context)
27.     IF errors NOT empty THEN
28.         RETURN SemanticInfo with errors
29.     END IF
30. END IF
31. 
32. // Phase 5: Subquery Semantics
33. IF HasSubqueries(compiledQuery) THEN
34.     errors ← ValidateSubqueries(compiledQuery, context)
35.     IF errors NOT empty THEN
36.         RETURN SemanticInfo with errors
37.     END IF
38. END IF
39. 
40. RETURN SemanticInfo with success
```

**Complexity**: O(n + m + s) where:
- n = number of expressions in query
- m = number of columns in referenced tables
- s = number of subqueries

**Space**: O(n + m) for tracking validation state

---

### 2. GROUP BY Validation

```
Algorithm: ValidateGroupBy(selectStmt, context)
Input: SELECT statement with GROUP BY clause
Output: Validation errors or success

1. groupByExprs ← Extract GROUP BY expressions
2. selectExprs ← Extract SELECT list expressions
3. 
4. // All non-aggregate expressions in SELECT must be in GROUP BY
5. FOR EACH expr IN selectExprs DO
6.     IF IsAggregateExpression(expr) THEN
7.         CONTINUE
8.     END IF
9.     
10.     IF NOT ContainsExpression(groupByExprs, expr) THEN
11.         IF NOT IsConstant(expr) THEN
12.             ERROR: "Column must appear in GROUP BY or be in aggregate"
13.         END IF
14.     END IF
15. END FOR
16. 
17. // Validate HAVING clause
18. IF HasHavingClause(selectStmt) THEN
19.     havingExpr ← selectStmt.Having.Condition
20.     errors ← ValidateHaving(havingExpr, groupByExprs, context)
21.     IF errors NOT empty THEN
22.         RETURN errors
23.     END IF
24. END IF
25. 
26. // Validate ORDER BY with GROUP BY
27. IF HasOrderBy(selectStmt) THEN
28.     FOR EACH orderExpr IN selectStmt.OrderBy DO
29.         IF NOT IsAggregateExpression(orderExpr.Expression) THEN
30.             IF NOT ContainsExpression(groupByExprs, orderExpr.Expression) THEN
31.                 ERROR: "ORDER BY must use aggregate or GROUP BY column"
32.             END IF
33.         END IF
34.     END FOR
35. END IF
36. 
37. RETURN success
```

**Complexity**: O(s × g) where:
- s = number of SELECT expressions
- g = number of GROUP BY expressions

**Key Rule**: Every non-aggregate expression in SELECT/HAVING/ORDER BY must appear in GROUP BY

---

### 3. Aggregate Function Validation

```
Algorithm: ValidateAggregateFunction(funcCall, context)
Input: Function call node, validation context
Output: Validation errors or success

1. funcName ← funcCall.Name
2. args ← funcCall.Arguments
3. 
4. // Check if it's an aggregate function
5. IF funcName NOT IN {COUNT, SUM, AVG, MIN, MAX, GROUP_CONCAT} THEN
6.     RETURN success  // Not an aggregate
7. END IF
8. 
9. // Validate argument count
10. expectedArgs ← GetExpectedArgCount(funcName)
11. IF LENGTH(args) < expectedArgs.min OR LENGTH(args) > expectedArgs.max THEN
12.     ERROR: "Wrong number of arguments for {funcName}"
13. END IF
14. 
15. // Validate argument types
16. SWITCH funcName:
17.     CASE COUNT:
18.         // COUNT(*) or COUNT(expression)
19.         IF LENGTH(args) == 1 THEN
20.             IF args[0] IS NOT Wildcard AND args[0] IS NOT Expression THEN
21.                 ERROR: "Invalid argument for COUNT"
22.             END IF
23.         END IF
24.         
25.     CASE SUM, AVG:
26.         // Must be numeric
27.         IF NOT IsNumericType(args[0]) THEN
28.             ERROR: "{funcName} requires numeric argument"
29.         END IF
30.         
31.     CASE MIN, MAX:
32.         // Must be comparable
33.         IF NOT IsComparableType(args[0]) THEN
34.             ERROR: "{funcName} requires comparable argument"
35.         END IF
36. END SWITCH
37. 
38. // Check for nested aggregates (not allowed)
39. FOR EACH arg IN args DO
40.     IF ContainsAggregate(arg) THEN
41.         ERROR: "Nested aggregate functions not allowed"
42.     END IF
43. END FOR
44. 
45. // Check context validity
46. IF context.InWhereClause THEN
47.     ERROR: "Aggregate functions not allowed in WHERE clause"
48. END IF
49. 
50. IF context.InGroupBy THEN
51.     ERROR: "Aggregate functions not allowed in GROUP BY clause"
52. END IF
53. 
54. RETURN success
```

**Complexity**: O(d) where d = depth of expression tree (to check for nested aggregates)

---

### 4. Subquery Validation

```
Algorithm: ValidateSubquery(subquery, context)
Input: Subquery node, parent context
Output: Validation errors or success

1. // Analyze subquery independently
2. subqueryResult ← AnalyzeQuery(subquery)
3. IF subqueryResult.HasErrors THEN
4.     RETURN subqueryResult.Errors
5. END IF
6. 
7. // Validate subquery type based on context
8. SWITCH context.SubqueryType:
9.     CASE SCALAR:
10.         // Must return single row, single column
11.         IF subquery.SelectColumns.Length != 1 THEN
12.             ERROR: "Scalar subquery must return single column"
13.         END IF
14.         // Runtime check for single row
15.         
16.     CASE IN_PREDICATE:
17.         // Must return single column
18.         IF subquery.SelectColumns.Length != 1 THEN
19.             ERROR: "IN subquery must return single column"
20.         END IF
21.         
22.     CASE EXISTS_PREDICATE:
23.         // Any number of columns OK
24.         
25.     CASE FROM_CLAUSE:
26.         // Derived table - any result OK
27.         IF NOT HasAlias(subquery) THEN
28.             ERROR: "Derived table must have alias"
29.         END IF
30. END SWITCH
31. 
32. // Validate correlated references
33. IF IsCorrelated(subquery) THEN
34.     correlatedRefs ← ExtractCorrelatedReferences(subquery)
35.     FOR EACH ref IN correlatedRefs DO
36.         IF NOT IsVisibleInContext(ref, context) THEN
37.             ERROR: "Correlated reference {ref} not visible"
38.         END IF
39.     END FOR
40. END IF
41. 
42. RETURN success
```

**Complexity**: O(n × c) where:
- n = size of subquery AST
- c = number of correlated references

---

### 5. Schema Dependency Validation

```
Algorithm: ValidateSchemaConstraints(stmt, catalog)
Input: DDL statement, catalog manager
Output: Validation errors or success

1. SWITCH stmt.Type:
2.     CASE CREATE_TABLE:
3.         // Check table doesn't already exist
4.         IF catalog.TableExists(stmt.TableName) THEN
5.             IF NOT stmt.IfNotExists THEN
6.                 ERROR: "Table {stmt.TableName} already exists"
7.             END IF
8.         END IF
9.         
10.         // Validate column definitions
11.         columnNames ← SET()
12.         FOR EACH col IN stmt.Columns DO
13.             IF col.Name IN columnNames THEN
14.                 ERROR: "Duplicate column name {col.Name}"
15.             END IF
16.             columnNames.ADD(col.Name)
17.             
18.             // Validate constraints
19.             errors ← ValidateColumnConstraints(col)
20.             IF errors NOT empty THEN
21.                 RETURN errors
22.             END IF
23.         END FOR
24.         
25.         // Validate foreign key references
26.         FOR EACH fk IN stmt.ForeignKeys DO
27.             IF NOT catalog.TableExists(fk.ReferencedTable) THEN
28.                 ERROR: "Referenced table {fk.ReferencedTable} not found"
29.             END IF
30.             
31.             refTable ← catalog.GetTable(fk.ReferencedTable)
32.             FOR EACH col IN fk.ReferencedColumns DO
33.                 IF NOT refTable.HasColumn(col) THEN
34.                     ERROR: "Referenced column {col} not found"
35.                 END IF
36.             END FOR
37.         END FOR
38.         
39.     CASE DROP_TABLE:
40.         // Check table exists
41.         IF NOT catalog.TableExists(stmt.TableName) THEN
42.             IF NOT stmt.IfExists THEN
43.                 ERROR: "Table {stmt.TableName} doesn't exist"
44.             END IF
45.         END IF
46.         
47.         // Check for dependent objects
48.         dependencies ← catalog.GetDependencies(stmt.TableName)
49.         IF dependencies NOT empty AND NOT stmt.Cascade THEN
50.             ERROR: "Cannot drop table with dependencies"
51.         END IF
52. END SWITCH
53. 
54. RETURN success
```

**Complexity**: O(c + f × r) where:
- c = number of columns
- f = number of foreign keys
- r = number of referenced columns per FK

---

### 6. Expression Semantics Validation

```
Algorithm: ValidateExpressionSemantics(expr, context)
Input: Expression node, validation context
Output: Validation errors or success

1. SWITCH expr.Type:
2.     CASE COLUMN_REFERENCE:
3.         // Must be resolved
4.         IF NOT IsResolved(expr, context.ResolvedRefs) THEN
5.             ERROR: "Unresolved column reference"
6.         END IF
7.         
8.     CASE BINARY_EXPRESSION:
9.         // Validate operands
10.         leftErrors ← ValidateExpressionSemantics(expr.Left, context)
11.         rightErrors ← ValidateExpressionSemantics(expr.Right, context)
12.         IF leftErrors OR rightErrors THEN
13.             RETURN leftErrors + rightErrors
14.         END IF
15.         
16.         // Validate operator semantics
17.         errors ← ValidateOperator(expr.Operator, expr.Left, expr.Right)
18.         IF errors NOT empty THEN
19.             RETURN errors
20.         END IF
21.         
22.     CASE FUNCTION_CALL:
23.         // Validate function exists
24.         IF NOT IsFunctionDefined(expr.Name) THEN
25.             ERROR: "Undefined function {expr.Name}"
26.         END IF
27.         
28.         // Validate arguments
29.         IF IsAggregateFunction(expr.Name) THEN
30.             errors ← ValidateAggregateFunction(expr, context)
31.         ELSE
32.             errors ← ValidateScalarFunction(expr, context)
33.         END IF
34.         
35.         IF errors NOT empty THEN
36.             RETURN errors
37.         END IF
38.         
39.     CASE SUBQUERY:
40.         errors ← ValidateSubquery(expr, context)
41.         IF errors NOT empty THEN
42.             RETURN errors
43.         END IF
44. END SWITCH
45. 
46. RETURN success
```

**Complexity**: O(d) where d = depth of expression tree

---

## Performance Characteristics

### Time Complexity Summary
- **Overall Pipeline**: O(n + m + s + a + d)
  - n = query size
  - m = schema size
  - s = subqueries
  - a = aggregates
  - d = expression depth

### Space Complexity
- **Context Storage**: O(m) for resolved references
- **Error Collection**: O(e) where e = number of errors
- **Validation State**: O(n) for expression tracking

### Optimization Strategies

1. **Early Exit**: Stop on first critical error
2. **Lazy Validation**: Only validate what's needed
3. **Cached Results**: Cache function signatures, schema info
4. **Parallel Validation**: Independent validations can run concurrently

---

## Error Recovery Strategies

### Graceful Degradation
```
1. Continue validation even after non-critical errors
2. Collect all errors for better user feedback
3. Mark query as invalid but provide full error report
```

### Error Prioritization
```
1. Critical errors (schema violations) - fail fast
2. Semantic errors (GROUP BY violations) - collect all
3. Warnings (deprecated features) - report but allow
```

---

## Validation Rules Summary

### 1. GROUP BY Rules
- Non-aggregate SELECT expressions must be in GROUP BY
- HAVING can only reference aggregates or GROUP BY columns
- ORDER BY follows same rules as SELECT

### 2. Aggregate Rules
- No nested aggregates allowed
- Cannot appear in WHERE clause
- Cannot appear in GROUP BY clause
- Type-specific requirements (SUM/AVG need numeric)

### 3. Subquery Rules
- Scalar subquery returns 1 row × 1 column
- IN subquery returns N rows × 1 column
- Derived tables must have aliases
- Correlated references must be visible

### 4. Schema Rules
- No duplicate column names in CREATE TABLE
- Foreign keys must reference existing tables/columns
- Cannot drop tables with dependencies (without CASCADE)
- CHECK constraints must be boolean expressions

---

## Integration with Query Compiler

```
Compiler Output → Semantic Analyzer → Optimizer Input

CompiledQuery:
  - Statement (AST)
  - ResolvedRefs (name resolution)
  - TypeInfo (type checking)
  
↓ Semantic Analysis ↓

SemanticInfo:
  - SchemaValid (boolean)
  - SemanticsValid (boolean)
  - AggregateInfo (metadata)
  - SubqueryInfo (metadata)
  - Errors (list)
  
↓ To Optimizer ↓

Validated query ready for optimization
```

---

## Testing Strategy

### Unit Tests
- Test each validation rule independently
- Test error detection and reporting
- Test edge cases (empty GROUP BY, nested aggregates)

### Integration Tests
- Test full semantic analysis pipeline
- Test interaction between validation phases
- Test error recovery and collection

### Performance Tests
- Benchmark large schema validation
- Benchmark complex aggregate queries
- Benchmark deeply nested subqueries
