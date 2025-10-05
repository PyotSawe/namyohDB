# Parser Module Algorithms

## Overview
This document describes the core algorithms used in the SQL parser for syntactic analysis, including parsing strategies, AST construction, and error recovery mechanisms.

## Core Parsing Algorithms

### 1. Recursive Descent Parsing Algorithm

#### Algorithm: `parseStatement()`
```
Input: Token stream from lexer
Output: AST node representing SQL statement

Algorithm parseStatement():
  1. Get current token type
  2. Switch on token type:
     a. SELECT → parseSelectStatement()
     b. INSERT → parseInsertStatement()
     c. UPDATE → parseUpdateStatement()
     d. DELETE → parseDeleteStatement()
     e. CREATE → parseCreateStatement()
     f. DROP   → parseDropStatement()
     g. ALTER  → parseAlterStatement()
  3. If unrecognized token:
     a. Generate syntax error
     b. Attempt error recovery
  4. Return AST node or error
```

**Time Complexity**: O(n) where n is the number of tokens
**Space Complexity**: O(d) where d is the maximum nesting depth

#### Algorithm: `parseSelectStatement()`
```
Input: Token stream starting with SELECT
Output: SelectStatement AST node

Algorithm parseSelectStatement():
  1. Expect SELECT token, advance
  2. Parse select list:
     a. Check for DISTINCT keyword (optional)
     b. Parse select items (comma-separated)
     c. Handle asterisk (*) or column expressions
  3. Expect FROM token, advance
  4. Parse table expression:
     a. Parse table name or subquery
     b. Handle table aliases (AS keyword)
     c. Parse JOIN clauses if present
  5. Parse optional clauses:
     a. WHERE clause → parseWhereClause()
     b. GROUP BY clause → parseGroupByClause()
     c. HAVING clause → parseHavingClause()
     d. ORDER BY clause → parseOrderByClause()
     e. LIMIT clause → parseLimitClause()
  6. Create SelectStatement AST node
  7. Return node with all parsed components
```

**Time Complexity**: O(n) where n is tokens in statement
**Space Complexity**: O(d) for nested expressions

### 2. Expression Parsing with Precedence

#### Algorithm: `parseExpression()`
```
Input: Token stream at start of expression
Output: Expression AST node

Algorithm parseExpression():
  1. Start with lowest precedence (OR)
  2. Call parseLogicalOr()
  3. Return resulting expression node

Algorithm parseLogicalOr():
  1. Parse left operand with parseLogicalAnd()
  2. While current token is OR:
     a. Store operator
     b. Advance token
     c. Parse right operand with parseLogicalAnd()
     d. Create BinaryExpression node
     e. Use as new left operand
  3. Return final expression
```

**Precedence Levels** (highest to lowest):
1. Primary expressions (literals, identifiers, parentheses)
2. Unary operators (NOT, -)
3. Multiplicative (*, /, %)
4. Additive (+, -)
5. Comparison (<, >, <=, >=)
6. Equality (=, <>, !=)
7. Logical AND
8. Logical OR

#### Algorithm: `parsePrecedenceLevel(level)`
```
Input: Precedence level, token stream
Output: Expression AST node

Algorithm parsePrecedenceLevel(level):
  1. If at highest precedence level:
     a. Return parsePrimary()
  2. Parse left operand at next higher level
  3. While current operator has current precedence level:
     a. Store operator and position
     b. Advance token
     c. Parse right operand at next higher level
     d. Create BinaryExpression(left, operator, right)
     e. Use as new left operand
  4. Return final expression
```

**Time Complexity**: O(n) for expression length n
**Space Complexity**: O(d) for nesting depth d

### 3. Error Recovery Algorithm

#### Algorithm: `synchronize()`
```
Input: Parser in error state
Output: Parser synchronized to safe state

Algorithm synchronize():
  1. Mark parser as in panic mode
  2. Skip tokens until synchronization point:
     a. Statement boundaries: semicolon (;)
     b. Major keywords: SELECT, INSERT, UPDATE, DELETE
     c. Block delimiters: closing parentheses
     d. End of input (EOF)
  3. When sync point found:
     a. Clear panic mode
     b. Position at recovery token
     c. Continue normal parsing
  4. If EOF reached:
     a. End parsing session
     b. Return all accumulated errors
```

**Recovery Points**:
- Statement separators (`;`)
- Statement keywords (`SELECT`, `FROM`, `WHERE`)
- Block boundaries (`(`, `)`, `{`, `}`)

#### Algorithm: `recoverFromError(expectedType)`
```
Input: Expected token type, current error context
Output: Recovery action taken

Algorithm recoverFromError(expectedType):
  1. Create error record:
     a. Current token position
     b. Expected token type
     c. Actual token found
     d. Current parsing context
  2. Add error to error list
  3. Determine recovery strategy:
     a. If in expression: skip to end of expression
     b. If in statement: skip to next statement
     c. If in clause: skip to next clause
  4. Apply recovery strategy
  5. Resume parsing at safe point
```

### 4. AST Node Construction Algorithms

#### Algorithm: `createSelectStatement(components)`
```
Input: Parsed statement components
Output: SelectStatement AST node

Algorithm createSelectStatement(components):
  1. Create SelectStatement node
  2. Set base properties:
     a. Position from first token
     b. Node type as SELECT_STMT
  3. Assign components:
     a. selectList from parsed select items
     b. fromClause from table expression
     c. whereClause from WHERE parsing (nullable)
     d. orderByClause from ORDER BY parsing (nullable)
     e. limitClause from LIMIT parsing (nullable)
  4. Validate node consistency:
     a. Check required fields present
     b. Verify expression types match context
  5. Return constructed node
```

#### Algorithm: `createBinaryExpression(left, operator, right)`
```
Input: Left expression, operator token, right expression
Output: BinaryExpression AST node

Algorithm createBinaryExpression(left, operator, right):
  1. Determine expression type from operator:
     a. Arithmetic: +, -, *, /, %
     b. Comparison: =, <>, <, >, <=, >=
     c. Logical: AND, OR
  2. Create appropriate BinaryExpression subtype
  3. Set node properties:
     a. Left operand
     b. Operator type and position
     c. Right operand
     d. Result type (if determinable)
  4. Validate operand compatibility
  5. Return expression node
```

### 5. Token Management Algorithms

#### Algorithm: `advance()`
```
Input: Current parser state
Output: Updated parser state with next token

Algorithm advance():
  1. Store current token as previous
  2. Get next token from lexer
  3. Update current token
  4. Check for lexical errors:
     a. If ILLEGAL token found
     b. Create parse error
     c. Continue with next token
  5. Return previous token
```

#### Algorithm: `expect(tokenType)`
```
Input: Expected token type
Output: Token if matches, error if not

Algorithm expect(tokenType):
  1. Check if current token matches expected type
  2. If match:
     a. Store token
     b. Advance to next token
     c. Return stored token
  3. If no match:
     a. Create error with expectation
     b. Attempt error recovery
     c. Return error
```

#### Algorithm: `match(tokenTypes...)`
```
Input: List of acceptable token types
Output: Boolean indicating match

Algorithm match(tokenTypes...):
  1. Check current token against each type in list
  2. If any match found:
     a. Advance to next token
     b. Return true
  3. If no match:
     a. Return false (no advance)
```

**Time Complexity**: O(1) per operation
**Space Complexity**: O(1)

## Advanced Parsing Algorithms

### 6. Function Call Parsing

#### Algorithm: `parseFunctionCall()`
```
Input: Function identifier token
Output: FunctionCallExpression AST node

Algorithm parseFunctionCall():
  1. Store function name from current token
  2. Advance past function name
  3. Expect opening parenthesis '('
  4. Parse argument list:
     a. If ')' immediately: empty argument list
     b. Otherwise: parse comma-separated expressions
     c. Handle special cases (DISTINCT, *)
  5. Expect closing parenthesis ')'
  6. Create FunctionCallExpression node:
     a. Function name
     b. Argument list
     c. Position information
  7. Return function call node
```

### 7. Subquery Parsing

#### Algorithm: `parseSubquery()`
```
Input: Opening parenthesis of subquery
Output: SubqueryExpression AST node

Algorithm parseSubquery():
  1. Expect opening parenthesis '('
  2. Create new parser context for subquery
  3. Parse complete SELECT statement
  4. Expect closing parenthesis ')'
  5. Create SubqueryExpression node:
     a. Embedded SELECT statement
     b. Subquery type (scalar, table, exists)
  6. Return subquery node
```

### 8. Table Expression Parsing

#### Algorithm: `parseTableExpression()`
```
Input: Token stream after FROM keyword
Output: TableExpression AST node

Algorithm parseTableExpression():
  1. Parse primary table reference:
     a. Table name with optional alias
     b. Subquery with alias
     c. Table function call
  2. Check for JOIN keywords:
     a. INNER JOIN, LEFT JOIN, RIGHT JOIN
     b. CROSS JOIN, FULL OUTER JOIN
  3. For each JOIN found:
     a. Parse JOIN type and target table
     b. Parse ON condition or USING clause
     c. Create JoinExpression node
  4. Build hierarchical join structure
  5. Return complete table expression
```

## Optimization Algorithms

### 9. Left-Recursion Elimination

#### Algorithm: `eliminateLeftRecursion(rule)`
```
Input: Grammar rule with left recursion
Output: Equivalent right-recursive rule

Algorithm eliminateLeftRecursion(rule):
  1. Identify left-recursive productions:
     A → A α | β
  2. Transform to right-recursive form:
     A → β A'
     A' → α A' | ε
  3. Update parsing functions accordingly
  4. Maintain semantic equivalence
```

### 10. Lookahead Optimization

#### Algorithm: `optimizeLookahead()`
```
Input: Parser with k-token lookahead
Output: Optimized parser with minimal lookahead

Algorithm optimizeLookahead():
  1. Analyze grammar for lookahead requirements
  2. Identify FIRST and FOLLOW sets
  3. Minimize lookahead where possible:
     a. Use LL(1) when sufficient
     b. Apply LL(k) only when necessary
  4. Implement predictive parsing tables
  5. Optimize token buffering
```

## Error Handling Algorithms

### 11. Panic Mode Recovery

#### Algorithm: `panicModeRecovery()`
```
Input: Parser in error state
Output: Parser synchronized to safe continuation point

Algorithm panicModeRecovery():
  1. Enter panic mode:
     a. Set panic flag
     b. Record error position
  2. Skip tokens until synchronization:
     a. Find statement boundary
     b. Or major keyword
     c. Or block delimiter
  3. Synchronize parser state:
     a. Clear panic flag
     b. Reset parsing context
     c. Resume normal operation
  4. Continue parsing from sync point
```

### 12. Error Message Generation

#### Algorithm: `generateErrorMessage(error)`
```
Input: Parse error context
Output: Human-readable error message

Algorithm generateErrorMessage(error):
  1. Identify error type:
     a. Unexpected token
     b. Missing expected token
     c. Invalid expression structure
  2. Generate contextual message:
     a. "Expected X but found Y"
     b. "Missing closing parenthesis"
     c. "Invalid expression in WHERE clause"
  3. Add position information:
     a. Line and column numbers
     b. Surrounding context
  4. Suggest possible fixes:
     a. Add missing tokens
     b. Check syntax rules
     c. Verify operator precedence
  5. Return formatted error message
```

## Memory Management Algorithms

### 13. AST Node Pooling

#### Algorithm: `nodePooling()`
```
Input: Request for new AST node
Output: Reused or new node instance

Algorithm nodePooling():
  1. Check node pool for available instance
  2. If pool has instance:
     a. Remove from pool
     b. Reset node state
     c. Return pooled instance
  3. If pool empty:
     a. Allocate new node
     b. Initialize node
     c. Return new instance
  4. When node no longer needed:
     a. Clean node state
     b. Return to pool for reuse
```

### 14. String Interning

#### Algorithm: `internString(str)`
```
Input: String literal from parsing
Output: Interned string reference

Algorithm internString(str):
  1. Check intern table for existing string
  2. If found:
     a. Return reference to existing string
  3. If not found:
     a. Add string to intern table
     b. Return reference to new entry
  4. Benefit: Multiple references to same string
  5. Memory savings for repeated identifiers/keywords
```

## Performance Analysis

### Algorithm Complexity Summary

| Algorithm | Time Complexity | Space Complexity | Notes |
|-----------|-----------------|------------------|--------|
| parseStatement() | O(n) | O(d) | n=tokens, d=depth |
| parseExpression() | O(n) | O(d) | Linear in expression size |
| synchronize() | O(k) | O(1) | k=tokens to sync point |
| advance() | O(1) | O(1) | Constant time operation |
| expect() | O(1) | O(1) | With error handling |
| createAST() | O(1) | O(n) | n=node size |
| panicRecovery() | O(k) | O(1) | k=recovery distance |

### Memory Usage Patterns

1. **AST Node Storage**: O(n) where n is number of nodes
2. **Token Buffering**: O(k) where k is lookahead distance
3. **Error List**: O(e) where e is number of errors
4. **Parser Stack**: O(d) where d is recursion depth

## Parsing Strategies Comparison

### 1. Recursive Descent vs Table-Driven

| Aspect | Recursive Descent | Table-Driven |
|--------|------------------|--------------|
| Code Size | Larger | Smaller |
| Readability | High | Low |
| Error Handling | Excellent | Limited |
| Performance | Fast | Very Fast |
| Maintenance | Easy | Difficult |

### 2. Precedence Parsing Methods

#### Operator Precedence
```
Algorithm operatorPrecedence():
  1. Use precedence table
  2. Shift-reduce decisions
  3. Handle precedence conflicts
  4. Efficient for expression parsing
```

#### Precedence Climbing  
```
Algorithm precedenceClimbing(minPrec):
  1. Parse left operand
  2. While operator precedence >= minPrec:
     a. Parse right operand recursively
     b. Combine with binary operation
  3. Return combined expression
```

## Testing Algorithms

### 15. Grammar Validation

#### Algorithm: `validateGrammar()`
```
Input: Grammar production rules
Output: Grammar correctness report

Algorithm validateGrammar():
  1. Check for left recursion
  2. Verify FIRST/FOLLOW set consistency
  3. Detect ambiguous productions
  4. Ensure grammar completeness
  5. Report potential issues
```

### 16. Parse Tree Comparison

#### Algorithm: `compareParseTree(expected, actual)`
```
Input: Two AST trees for comparison
Output: Structural equivalence result

Algorithm compareParseTree(expected, actual):
  1. Compare root node types
  2. Recursively compare children:
     a. Same number of children
     b. Same child node types
     c. Same attribute values
  3. Report differences found
  4. Return equivalence boolean
```

## Future Algorithm Enhancements

### 1. Incremental Parsing
- Parse only changed portions of SQL
- Maintain parse state between edits
- Efficient re-parsing for IDE integration

### 2. Parallel Parsing
- Parse independent clauses concurrently
- Merge results into single AST
- Handle dependencies between clauses

### 3. Error-Tolerant Parsing
- Continue parsing despite multiple errors
- Infer missing constructs intelligently
- Provide comprehensive error reporting

### 4. Semantic-Aware Parsing
- Integrate symbol table during parsing
- Context-sensitive parsing decisions
- Early semantic validation