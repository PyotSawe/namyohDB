# Relational Database System Architecture

## Project Overview

This is a production-ready relational database management system (RDBMS) implemented in Go, following SQLite3's modular architecture. The system provides a complete SQL database engine with ACID compliance, concurrent access, and persistence.

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Client Applications                            │
└─────────────────────────────────────┬───────────────────────────────────┘
                                      │
┌─────────────────────────────────────┴───────────────────────────────────┐
│                         SQLite3-Compatible API                          │
│                         (pkg/sqlite3)                                   │
└─────────────────────────────────────┬───────────────────────────────────┘
                                      │
┌─────────────────────────────────────┴───────────────────────────────────┐
│                            Core Engine                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌──────────┐ │
│  │ SQL         │───▶│   SQL       │───▶│   Query     │───▶│  Query   │ │
│  │ Dispatcher  │    │   Parser    │    │  Optimizer  │    │ Executor │ │
│  │             │    │   (AST)     │    │             │    │          │ │
│  └─────────────┘    └─────────────┘    └─────────────┘    └──────────┘ │
│         │                   │                   │               │       │
│         ▼                   ▼                   ▼               ▼       │
│  ┌─────────────────────────────────────────────────────────────────────┤
│  │                    SQL Lexer                                        │
│  └─────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│  │ Transaction │    │   Locking   │    │ Concurrency │                │
│  │ Manager     │    │ & Isolation │    │   Control   │                │
│  │ (WAL)       │    │             │    │             │                │
│  └─────────────┘    └─────────────┘    └─────────────┘                │
│                                                                         │
└─────────────────────────────────────┬───────────────────────────────────┘
                                      │
┌─────────────────────────────────────┴───────────────────────────────────┐
│                       Storage Layer                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌──────────┐ │
│  │   B-tree    │    │   Buffer    │    │   File I/O  │    │  Page    │ │
│  │  Indexing   │    │    Pool     │    │    Layer    │    │ Manager  │ │
│  │             │    │   (LRU)     │    │(Journaling) │    │          │ │
│  └─────────────┘    └─────────────┘    └─────────────┘    └──────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Modular Design Principles

#### 1. **Layered Architecture**
- **API Layer**: SQLite3-compatible public interface
- **SQL Processing Layer**: Lexer, Parser, Optimizer, Executor
- **Transaction Layer**: ACID compliance, concurrency control
- **Storage Layer**: B-trees, buffer management, persistence

#### 2. **Separation of Concerns**
- Each module has a single, well-defined responsibility
- Clear interfaces between modules
- Minimal dependencies and coupling

#### 3. **SQLite3 Compatibility**
- Same public API surface as SQLite3
- Compatible SQL dialect and behavior
- Drop-in replacement capability

## Module Architecture Details

### 1. SQL Processing Pipeline

```
SQL Input → Lexer → Parser → AST → Optimizer → Execution Plan → Executor → Results
```

#### SQL Lexer (`internal/lexer`)
- **Purpose**: Tokenize SQL statements into a stream of tokens
- **Key Components**:
  - Character scanner with lookahead
  - Token classification (keywords, identifiers, literals, operators)
  - Position tracking for error reporting
  - Comment handling
- **Architecture**: Single-pass lexical analyzer with state machine
- **Performance**: O(n) linear processing, minimal memory allocations

#### SQL Parser (`internal/parser`)
- **Purpose**: Build Abstract Syntax Tree (AST) from token stream
- **Key Components**:
  - Recursive descent parser
  - AST node types for all SQL constructs
  - Error recovery and reporting
  - Precedence handling for expressions
- **Architecture**: Top-down parser with predictive parsing
- **Output**: Strongly-typed AST representing SQL statements

#### Query Optimizer (`internal/optimizer`)
- **Purpose**: Transform AST into efficient execution plan
- **Key Components**:
  - Cost-based optimization
  - Index selection
  - Join ordering
  - Predicate pushdown
- **Architecture**: Rule-based and cost-based optimization phases
- **Performance**: Polynomial time complexity with pruning

#### Query Executor (`internal/executor`)
- **Purpose**: Execute optimized query plans
- **Key Components**:
  - Operator implementations (scan, join, sort, aggregate)
  - Result set management
  - Memory management
  - Statistics collection
- **Architecture**: Iterator-based execution model
- **Concurrency**: Thread-safe execution with proper locking

### 2. Transaction Management

#### Transaction Manager (`internal/transaction`)
- **Purpose**: Provide ACID guarantees and transaction isolation
- **Key Components**:
  - Write-Ahead Logging (WAL)
  - Transaction state management
  - Rollback and recovery
  - Checkpoint coordination
- **Architecture**: WAL-based with automatic checkpointing
- **Durability**: fsync() coordination for crash recovery

#### Locking and Concurrency (`internal/locking`)
- **Purpose**: Manage concurrent access with appropriate isolation levels
- **Key Components**:
  - Multiple granularity locking (table, page, row)
  - Deadlock detection and resolution
  - Lock escalation
  - Reader-writer locks
- **Architecture**: Hierarchical locking with deadlock prevention
- **Performance**: Lock-free reads when possible

### 3. Storage Engine

#### File I/O Layer (`internal/fileio`)
- **Purpose**: Manage database file operations with reliability
- **Key Components**:
  - Page-based file management
  - Journaling for crash recovery
  - Memory-mapped I/O where appropriate
  - File locking for concurrent access
- **Architecture**: Page-oriented with journal-based recovery
- **Reliability**: Atomic writes with rollback capability

#### B-tree Indexing (`internal/btree`)
- **Purpose**: Provide efficient indexed access to data
- **Key Components**:
  - B+ tree implementation
  - Index maintenance during modifications
  - Range queries and point lookups
  - Index statistics for optimization
- **Architecture**: Multi-level B+ trees with leaf-level data
- **Performance**: O(log n) access time, high fan-out

#### Buffer Pool (`internal/bufferpool`)
- **Purpose**: Manage in-memory caching of database pages
- **Key Components**:
  - LRU eviction policy
  - Page pinning and dirty tracking
  - Write-back caching
  - Statistics and monitoring
- **Architecture**: Hash table + LRU list for O(1) access
- **Concurrency**: Fine-grained locking for high concurrency

### 4. Public API

#### SQLite3-Compatible API (`pkg/sqlite3`)
- **Purpose**: Provide familiar SQLite3 interface for applications
- **Key Components**:
  - Database connection management
  - Statement preparation and execution
  - Result set iteration
  - Transaction control
- **Architecture**: C-compatible API with Go implementation
- **Compatibility**: Drop-in replacement for SQLite3

## Data Flow Architecture

### 1. Query Processing Flow

```
1. SQL Statement Input
   ↓
2. Lexical Analysis (Tokenization)
   ↓
3. Syntax Analysis (AST Generation)
   ↓
4. Semantic Analysis (Validation)
   ↓
5. Query Optimization (Plan Generation)
   ↓
6. Plan Execution
   ↓
7. Result Set Generation
```

### 2. Transaction Processing Flow

```
1. BEGIN Transaction
   ↓
2. Acquire Necessary Locks
   ↓
3. Execute Operations (with WAL logging)
   ↓
4. Validation and Conflict Detection
   ↓
5. COMMIT (flush WAL) or ROLLBACK
   ↓
6. Release Locks
   ↓
7. Background Checkpoint (if needed)
```

### 3. Storage Access Flow

```
1. Page Request (by page number)
   ↓
2. Buffer Pool Lookup
   ↓
3a. Cache Hit → Return Page
   ↓
3b. Cache Miss → Load from Disk
   ↓
4. Page Pinning (prevent eviction)
   ↓
5. Page Access/Modification
   ↓
6. Mark Dirty (if modified)
   ↓
7. Unpin Page
   ↓
8. Background Write (if dirty + unpinned)
```

## Scalability Architecture

### 1. Memory Management
- **Buffer Pool Sizing**: Configurable memory limits
- **Memory-Mapped Files**: For large read-only workloads
- **Streaming Results**: Avoid loading entire result sets in memory
- **Connection Pooling**: Reuse connections to reduce overhead

### 2. Concurrency Design
- **Reader-Writer Locks**: Multiple readers, single writer per resource
- **Lock-Free Structures**: Where possible, use atomic operations
- **Work Stealing**: For parallel query execution
- **Connection Per Thread**: Isolated execution contexts

### 3. Performance Optimizations
- **Index-Only Scans**: Avoid table access when possible
- **Vectorized Execution**: Process multiple rows per operation
- **Bloom Filters**: Reduce I/O for join operations
- **Compression**: Reduce storage and I/O overhead

## Configuration Architecture

### 1. Configuration System (`internal/config`)
- **Environment Variables**: Runtime configuration
- **Configuration Files**: Structured settings (YAML/JSON)
- **Command Line**: Override settings for testing
- **Runtime Changes**: Dynamic reconfiguration where safe

### 2. Key Configuration Areas
- **Memory Limits**: Buffer pool size, query memory limits
- **I/O Settings**: Page size, sync behavior, cache sizes
- **Concurrency**: Max connections, lock timeouts
- **Logging**: Log levels, file rotation, performance metrics

## Error Handling Architecture

### 1. Error Categories
- **Syntax Errors**: SQL parsing and validation errors
- **Runtime Errors**: Constraint violations, type mismatches
- **System Errors**: I/O failures, memory exhaustion
- **Concurrency Errors**: Deadlocks, lock timeouts

### 2. Error Recovery Strategy
- **Graceful Degradation**: Continue operation when possible
- **Transaction Rollback**: Automatic rollback on errors
- **Connection Recovery**: Reset connection state on errors
- **Logging and Monitoring**: Comprehensive error tracking

## Testing Architecture

### 1. Testing Levels
- **Unit Tests**: Individual module testing
- **Integration Tests**: Cross-module interaction testing
- **System Tests**: Full end-to-end testing
- **Performance Tests**: Benchmarking and profiling
- **Chaos Testing**: Fault injection and recovery testing

### 2. Test Data Management
- **Test Fixtures**: Reusable test databases
- **Data Generation**: Synthetic data for testing
- **Test Isolation**: Each test runs in isolation
- **Cleanup**: Automatic test environment cleanup

## Deployment Architecture

### 1. Build System
- **Cross-Platform**: Windows, Linux, macOS support
- **Static Linking**: Single binary deployment
- **Optimization**: Release builds with full optimization
- **Debug Symbols**: Separate debug information

### 2. Distribution
- **Binary Releases**: Pre-built binaries for common platforms
- **Package Managers**: Integration with Go modules
- **Container Images**: Docker images for containerized deployment
- **Source Distribution**: Complete source code packages

## Monitoring and Observability

### 1. Metrics Collection
- **Performance Metrics**: Query execution times, throughput
- **Resource Metrics**: Memory usage, I/O statistics
- **Error Metrics**: Error rates, types of failures
- **Business Metrics**: Database size, query patterns

### 2. Logging Strategy
- **Structured Logging**: JSON-formatted logs
- **Log Levels**: Configurable verbosity levels
- **Log Rotation**: Automatic log file management
- **Performance Impact**: Minimal overhead logging

## Security Architecture

### 1. Access Control
- **Connection Security**: Authentication and authorization
- **SQL Injection Prevention**: Parameterized queries
- **Privilege Management**: Least privilege principle
- **Audit Logging**: Security event logging

### 2. Data Protection
- **Encryption at Rest**: Optional database encryption
- **Secure Communication**: TLS for network connections
- **Memory Protection**: Secure memory handling
- **Backup Security**: Encrypted backup support

## Future Architecture Considerations

### 1. Horizontal Scaling
- **Replication**: Master-slave replication
- **Sharding**: Horizontal partitioning
- **Distributed Queries**: Cross-shard query execution
- **Consensus**: Distributed transaction coordination

### 2. Cloud Integration
- **Object Storage**: S3/GCS backend support
- **Kubernetes**: Cloud-native deployment
- **Serverless**: Function-as-a-Service integration
- **Multi-Cloud**: Cloud-agnostic deployment

### 3. Advanced Features
- **Columnar Storage**: OLAP workload optimization
- **Vector Search**: Similarity search capabilities
- **Graph Queries**: Graph database functionality
- **Stream Processing**: Real-time data processing