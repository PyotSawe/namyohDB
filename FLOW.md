# NamyohDB: Processing Flow Architecture

## Overview

This document describes the complete processing flow through NamyohDB's layered architecture, showing how SQL queries are processed from client input to result delivery. The flow follows SQLite3's embedded approach with Derby's modular processing pipeline.

## Complete Query Processing Flow

### High-Level Flow Overview

```
Client Application
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. CLIENT REQUEST (SQL String + Connection)                 │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. SQL INTERFACE LAYER                                      │
│    • Connection validation                                  │
│    • Request routing                                        │
│    • Session management                                     │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. SQL COMPILER LAYER                                       │
│    • Lexical analysis (tokenization)                       │
│    • Syntax analysis (AST generation)                      │
│    • Semantic analysis (validation)                        │
│    • Query optimization                                     │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. EXECUTION ENGINE LAYER                                   │
│    • Execution plan processing                              │
│    • Transaction management                                 │
│    • Lock acquisition                                       │
│    • Cursor management                                      │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. STORAGE MANAGER LAYER                                    │
│    • Table/Index access                                     │
│    • B-tree operations                                      │
│    • Buffer pool management                                 │
│    • Page operations                                        │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. I/O & RECOVERY LAYER                                     │
│    • File system operations                                 │
│    • WAL logging                                            │
│    • Physical I/O                                           │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. RESULT DELIVERY (Back up the stack)                     │
└─────────────────────────────────────────────────────────────┘
```

## Detailed Layer-by-Layer Processing

### Layer 1: Client Layer
**Input**: SQL statement from application
**Components**: Client applications, command-line interface
**Processing Activity**: None (external to database)

---

### Layer 2: SQL Interface Layer

#### 2.1 Connection Management (`cmd/relational-db`, `pkg/database`)
```
┌─── Client Request ───┐
│                      │
▼                      │
┌─────────────────────────────────────────┐
│ Connection Manager                      │
│ ┌─────────────────────────────────────┐ │
│ │ • Validate connection               │ │
│ │ • Check authentication              │ │  
│ │ • Allocate session context          │ │
│ │ • Initialize transaction state      │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Request Router                          │
│ ┌─────────────────────────────────────┐ │
│ │ • Parse request type                │ │
│ │ • Route to appropriate handler      │ │
│ │ • Set up execution context          │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼ (SQL String + Context)
```

**Status**: ✅ **IMPLEMENTED**
- Connection handling in `cmd/relational-db/main.go`
- Basic API structure in `pkg/database/`

---

### Layer 3: SQL Compiler Layer (Derby-Style)

#### 3.1 SQL Lexer (`internal/lexer`) ✅ IMPLEMENTED
```
SQL Input: "SELECT name FROM users WHERE id = 42"
│
▼
┌─────────────────────────────────────────┐
│ Lexical Analyzer                       │
│ ┌─────────────────────────────────────┐ │
│ │ Character Stream Processing:        │ │
│ │ • Scan character by character       │ │
│ │ • Identify token boundaries         │ │
│ │ • Classify token types              │ │
│ │ • Track position information        │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼ Token Stream
[SELECT][name][FROM][users][WHERE][id][=][42]
```

**Token Output**:
```
Token{Type: SELECT, Value: "SELECT", Line: 1, Column: 1}
Token{Type: IDENTIFIER, Value: "name", Line: 1, Column: 8}
Token{Type: FROM, Value: "FROM", Line: 1, Column: 13}
Token{Type: IDENTIFIER, Value: "users", Line: 1, Column: 18}
Token{Type: WHERE, Value: "WHERE", Line: 1, Column: 24}
Token{Type: IDENTIFIER, Value: "id", Line: 1, Column: 30}
Token{Type: EQUALS, Value: "=", Line: 1, Column: 33}
Token{Type: NUMBER, Value: "42", Line: 1, Column: 35}
```

#### 3.2 SQL Parser (`internal/parser`) 🚧 PARTIAL
```
Token Stream Input
│
▼
┌─────────────────────────────────────────┐
│ Syntax Analyzer (Recursive Descent)    │
│ ┌─────────────────────────────────────┐ │
│ │ Parse Functions:                    │ │
│ │ • parseSelectStatement()            │ │
│ │ • parseFromClause()                 │ │
│ │ • parseWhereClause()                │ │
│ │ • parseExpression()                 │ │
│ │ • parseBinaryExpression()           │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼ Abstract Syntax Tree (AST)
```

**AST Output**:
```go
SelectStatement{
    Columns: []Expression{
        IdentifierExpression{Name: "name"}
    },
    From: TableReference{Name: "users"},
    Where: BinaryExpression{
        Left:     IdentifierExpression{Name: "id"},
        Operator: EQUALS,
        Right:    LiteralExpression{Value: 42, Type: INTEGER}
    }
}
```

#### 3.3 Semantic Analyzer 🚧 PLANNED
```
AST Input
│
▼
┌─────────────────────────────────────────┐
│ Semantic Analysis Engine                │
│ ┌─────────────────────────────────────┐ │
│ │ Validation Steps:                   │ │
│ │ • Check table existence             │ │
│ │ • Validate column references        │ │
│ │ • Type checking and coercion        │ │
│ │ • Permission validation             │ │
│ │ • Constraint checking               │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼ Validated AST + Schema Information
```

#### 3.4 Query Optimizer 🚧 PLANNED
```
Validated AST Input
│
▼
┌─────────────────────────────────────────┐
│ Rule-Based Optimization                 │
│ ┌─────────────────────────────────────┐ │
│ │ • Predicate pushdown                │ │
│ │ • Constant folding                  │ │
│ │ • Expression simplification         │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Cost-Based Optimization                 │
│ ┌─────────────────────────────────────┐ │
│ │ • Generate alternative plans        │ │
│ │ • Estimate execution costs          │ │
│ │ • Select optimal plan               │ │
│ │ • Index selection                   │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼ Optimized Execution Plan
```

---

### Layer 4: Execution Engine Layer (SQLite3-Style)

#### 4.1 Query Executor 🚧 PLANNED
```
Execution Plan Input
│
▼
┌─────────────────────────────────────────┐
│ Execution Coordinator                   │
│ ┌─────────────────────────────────────┐ │
│ │ Plan Processing:                    │ │
│ │ • Initialize execution context      │ │
│ │ • Set up operator pipeline          │ │
│ │ • Allocate working memory           │ │
│ │ • Create result cursors             │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Operator Execution Pipeline             │
│ ┌─────────────────────────────────────┐ │
│ │ • TableScanOperator                 │ │
│ │ • FilterOperator (WHERE clause)     │ │
│ │ • ProjectionOperator (SELECT cols)  │ │
│ │ • SortOperator (ORDER BY)           │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### 4.2 Transaction Manager 🚧 PLANNED
```
Execution Request
│
▼
┌─────────────────────────────────────────┐
│ Transaction Control                     │
│ ┌─────────────────────────────────────┐ │
│ │ Transaction Lifecycle:              │ │
│ │ • Begin transaction (if needed)     │ │
│ │ • Acquire transaction ID            │ │
│ │ • Set isolation level               │ │
│ │ • Initialize undo log               │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Concurrency Control                     │
│ ┌─────────────────────────────────────┐ │
│ │ • Acquire necessary locks           │ │
│ │ • Check for conflicts               │ │
│ │ • Deadlock detection                │ │
│ │ • MVCC version management           │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

### Layer 5: Storage Manager Layer (SQLite3-Inspired)

#### 5.1 Table Manager 🚧 PLANNED
```
Storage Request (Table: "users", Operation: SCAN)
│
▼
┌─────────────────────────────────────────┐
│ Table Access Manager                    │
│ ┌─────────────────────────────────────┐ │
│ │ • Resolve table name to table ID    │ │
│ │ • Load table metadata               │ │
│ │ • Determine access method           │ │
│ │ • Initialize scan cursor            │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Index Selection                         │
│ ┌─────────────────────────────────────┐ │
│ │ • Evaluate available indexes        │ │
│ │ • Choose optimal access path        │ │
│ │ • Set up index scan or table scan   │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### 5.2 B-Tree Manager 🚧 PLANNED
```
Access Request (Method: INDEX_SCAN, Key: id=42)
│
▼
┌─────────────────────────────────────────┐
│ B-Tree Navigation                       │
│ ┌─────────────────────────────────────┐ │
│ │ • Start at root page                │ │
│ │ • Navigate to leaf page             │ │
│ │ • Binary search within page         │ │
│ │ • Position cursor at key            │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Record Retrieval                        │
│ ┌─────────────────────────────────────┐ │
│ │ • Read record from leaf page        │ │
│ │ • Deserialize record data           │ │
│ │ • Apply filters                     │ │
│ │ • Continue scan if needed           │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### 5.3 Buffer Pool Manager ✅ IMPLEMENTED
```
Page Request (PageID: 1024)
│
▼
┌─────────────────────────────────────────┐
│ Buffer Pool Lookup                      │
│ ┌─────────────────────────────────────┐ │
│ │ Current Implementation:             │ │
│ │ • Hash table lookup O(1)            │ │
│ │ • Check if page in memory           │ │
│ │ • Update LRU position               │ │
│ │ • Return cached page if found       │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
├─ Cache Hit ──────────────────────────┐
│                                      │
▼                                      │
Return Page                            │
│                                      │
▼                                      │
┌─────────────────────────────────────────┐
│ Cache Miss - Load from Storage          │
│ ┌─────────────────────────────────────┐ │
│ │ • Allocate buffer frame             │ │
│ │ • Request page from file manager    │ │
│ │ • Load page into buffer             │ │
│ │ • Add to hash table                 │ │
│ │ • Update LRU chain                  │ │
│ │ • Evict LRU page if needed          │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

### Layer 6: I/O & Recovery Layer

#### 6.1 File Manager ✅ IMPLEMENTED
```
Page I/O Request (PageID: 1024, Operation: READ)
│
▼
┌─────────────────────────────────────────┐
│ File Operations Manager                 │
│ ┌─────────────────────────────────────┐ │
│ │ Current Implementation:             │ │
│ │ • Calculate file offset             │ │
│ │ • Validate page boundaries          │ │
│ │ • Perform atomic read operation     │ │
│ │ • Handle I/O errors gracefully      │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Operating System I/O                    │
│ ┌─────────────────────────────────────┐ │
│ │ • System call (read/write)          │ │
│ │ • File system operations            │ │
│ │ • Disk I/O                          │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### 6.2 WAL Manager 🚧 PLANNED
```
Write Operation
│
▼
┌─────────────────────────────────────────┐
│ Write-Ahead Logging                     │
│ ┌─────────────────────────────────────┐ │
│ │ • Generate WAL record               │ │
│ │ • Assign LSN (Log Sequence Number)  │ │
│ │ • Append to WAL file                │ │
│ │ • Force WAL to disk (durability)    │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Checkpoint Processing                   │
│ ┌─────────────────────────────────────┐ │
│ │ • Flush dirty pages to main DB      │ │
│ │ • Truncate WAL file                 │ │
│ │ • Update database header            │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## Transaction Processing Flow

### ACID Transaction Lifecycle

#### Begin Transaction
```
1. Client Issues BEGIN
   │
   ▼
2. Transaction Manager
   • Assign Transaction ID
   • Initialize transaction context
   • Set isolation level
   • Create undo log space
   │
   ▼
3. Lock Manager
   • Prepare for lock acquisition
   • Initialize deadlock detection
```

#### Execute Operations (SELECT Example)
```
1. Parse & Optimize (Layers 2-3)
   │
   ▼
2. Acquire Shared Locks
   • Table-level shared lock
   • Page-level shared locks (as accessed)
   │
   ▼
3. Execute Query Plan
   • Scan table/index pages
   • Apply filters and projections
   • Buffer results
   │
   ▼
4. Return Results
   • Stream results to client
   • Maintain locks until transaction end
```

#### Commit Transaction
```
1. Client Issues COMMIT
   │
   ▼
2. Transaction Manager
   • Validate transaction state
   • Check for conflicts
   │
   ▼
3. WAL Manager
   • Write commit record to WAL
   • Force WAL to disk (durability)
   │
   ▼
4. Lock Manager
   • Release all transaction locks
   • Notify waiting transactions
   │
   ▼
5. Cleanup
   • Free transaction resources
   • Clear undo log space
```

## Error Handling Flow

### Error Propagation Through Layers

```
Error Occurrence (Any Layer)
│
▼
┌─────────────────────────────────────────┐
│ Error Detection & Classification        │
│ ┌─────────────────────────────────────┐ │
│ │ • Syntax Error (Layer 3)            │ │
│ │ • Semantic Error (Layer 3)          │ │
│ │ • Execution Error (Layer 4)         │ │
│ │ • Storage Error (Layer 5)           │ │
│ │ • I/O Error (Layer 6)               │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Error Handling & Recovery               │
│ ┌─────────────────────────────────────┐ │
│ │ • Rollback active operations        │ │
│ │ • Release acquired locks            │ │
│ │ • Clean up allocated resources      │ │
│ │ • Log error for diagnostics         │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────┐
│ Error Response Generation               │
│ ┌─────────────────────────────────────┐ │
│ │ • Format error message              │ │
│ │ • Include context information       │ │
│ │ • Return to client                  │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## Performance Optimization Flow

### Query Performance Pipeline

```
1. Query Analysis
   • Identify bottlenecks
   • Analyze access patterns
   • Check index usage
   │
   ▼
2. Optimizer Decisions
   • Choose optimal access paths
   • Select best join algorithms
   • Determine memory allocation
   │
   ▼
3. Runtime Optimizations
   • Buffer pool hit ratio maximization
   • Concurrent execution where possible
   • Memory-efficient operations
   │
   ▼
4. Statistics Collection
   • Track execution metrics
   • Update optimizer statistics
   • Monitor resource usage
```

## Module Interaction Summary

### Current Implementation Status

| Layer | Module | Status | Key Interactions |
|-------|--------|---------|-----------------|
| 2 | Connection Mgmt | ✅ IMPLEMENTED | → SQL Compiler |
| 3 | SQL Lexer | ✅ IMPLEMENTED | → Parser |
| 3 | SQL Parser | 🚧 PARTIAL | → Semantic Analysis |
| 4 | Query Executor | 🚧 PLANNED | ↔ Storage Manager |
| 5 | Buffer Pool | ✅ IMPLEMENTED | ↔ File Manager |
| 5 | Page Manager | 🚧 PARTIAL | ↔ Buffer Pool |
| 6 | File Manager | ✅ IMPLEMENTED | → OS I/O |

### Next Implementation Priorities

1. **Complete SQL Parser** (Layer 3)
   - Implement remaining parse functions
   - Add error recovery mechanisms
   - Complete AST node implementations

2. **Basic Query Executor** (Layer 4)
   - Implement simple table scan operations
   - Add basic filtering capabilities
   - Create result set management

3. **Schema Management** (Layer 4-5)
   - Implement CREATE/DROP TABLE
   - Add column definitions and constraints
   - Integrate with storage layer

4. **Transaction Framework** (Layer 4)
   - Basic BEGIN/COMMIT/ROLLBACK
   - Simple locking mechanisms
   - WAL integration

This flow architecture provides a clear roadmap for implementing the remaining components while maintaining the proven architectural patterns from SQLite3 and Derby.