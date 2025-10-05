# Dispatcher Module Architecture

## Overview
The Dispatcher module serves as the central routing hub for SQL statement execution, analyzing parsed SQL statements and coordinating their execution through the appropriate database subsystems. It acts as the orchestrator between the SQL parsing layer and the query execution engine.

## Architecture Design

### Core Components

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Dispatcher Module                               │
├─────────────────────────────────────────────────────────────────────────┤
│  Input: Abstract Syntax Tree (AST)                                     │
│  Output: Query Execution Results / Status                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌────────────────┐    ┌──────────────────────────┐ │
│  │   Statement │───▶│   Query Type   │───▶│    Execution Router      │ │
│  │   Analyzer  │    │ Classification │    │                          │ │
│  │             │    │                │    │                          │ │
│  └─────────────┘    └────────────────┘    └──────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                   Execution Coordinators                         │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │ │
│  │  │    DQL      │ │    DML      │ │           DDL               │ │ │
│  │  │Coordinator  │ │Coordinator  │ │       Coordinator           │ │ │
│  │  │ (SELECT)    │ │(INSERT,     │ │    (CREATE, DROP,           │ │ │
│  │  │             │ │ UPDATE,     │ │     ALTER)                  │ │ │
│  │  │             │ │ DELETE)     │ │                             │ │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                    Support Systems                                  │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │ │
│  │  │ Session     │ │ Transaction │ │      Performance            │ │ │
│  │  │ Management  │ │ Management  │ │      Monitoring             │ │ │
│  │  │             │ │             │ │                             │ │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. **Command Pattern Implementation**
- **Decision**: Use command pattern for SQL statement execution
- **Rationale**: 
  - Decouples statement analysis from execution
  - Enables request queuing and batching
  - Supports undo/redo operations
  - Facilitates transaction management
- **Benefits**: Clean separation of concerns, extensible execution model

#### 2. **Strategy Pattern for Query Types**
- **Decision**: Different execution strategies for DQL, DML, DDL
- **Rationale**:
  - Each query type has different execution requirements
  - Enables specialized optimization for each type
  - Supports different concurrency models per type
  - Allows independent evolution of execution paths
- **Implementation**: Strategy interface with concrete implementations

#### 3. **Asynchronous Execution Model**
- **Decision**: Support both synchronous and asynchronous execution
- **Rationale**:
  - Long-running queries benefit from async execution
  - Enables progress reporting and cancellation
  - Supports concurrent query execution
  - Better resource utilization
- **Trade-offs**: Added complexity vs. improved performance/UX

#### 4. **Centralized Session Management**
- **Decision**: Dispatcher manages database sessions and connections
- **Rationale**:
  - Single point of control for resource allocation
  - Consistent session state management
  - Connection pooling and reuse
  - Security and access control enforcement
- **Benefits**: Resource efficiency, security, consistency

### Dispatcher Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                   Public API Layer                      │
│            (Execute, ExecuteAsync, Prepare)             │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                 Statement Analysis Layer                │
│        (Type Classification, Security Validation)      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                Routing & Coordination Layer             │
│           (Execution Strategy Selection)                │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                 Execution Engine Layer                  │
│      (Query Optimizer, Transaction Manager, etc.)      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                   Storage Layer                         │
│              (Buffer Pool, File I/O)                    │
└─────────────────────────────────────────────────────────┘
```

## Statement Classification System

### 1. **Query Type Taxonomy**

```go
type QueryType int

const (
    // Data Query Language (Read Operations)
    SELECT_QUERY QueryType = iota
    
    // Data Manipulation Language (Write Operations)
    INSERT_QUERY
    UPDATE_QUERY
    DELETE_QUERY
    
    // Data Definition Language (Schema Operations)
    CREATE_TABLE_QUERY
    CREATE_INDEX_QUERY
    DROP_TABLE_QUERY
    DROP_INDEX_QUERY
    ALTER_TABLE_QUERY
    
    // Transaction Control Language
    BEGIN_TRANSACTION_QUERY
    COMMIT_QUERY
    ROLLBACK_QUERY
    
    // Data Control Language
    GRANT_QUERY
    REVOKE_QUERY
    
    // Database Administration
    ANALYZE_QUERY
    VACUUM_QUERY
    PRAGMA_QUERY
)
```

### 2. **Query Classification Logic**

```go
func (d *Dispatcher) classifyQuery(stmt parser.Statement) (QueryType, QueryMetadata) {
    switch s := stmt.(type) {
    case *parser.SelectStatement:
        return SELECT_QUERY, analyzeSelectQuery(s)
    case *parser.InsertStatement:
        return INSERT_QUERY, analyzeInsertQuery(s)
    case *parser.UpdateStatement:
        return UPDATE_QUERY, analyzeUpdateQuery(s)
    case *parser.DeleteStatement:
        return DELETE_QUERY, analyzeDeleteQuery(s)
    case *parser.CreateTableStatement:
        return CREATE_TABLE_QUERY, analyzeCreateTableQuery(s)
    // ... other statement types
    default:
        return UNKNOWN_QUERY, QueryMetadata{}
    }
}
```

### 3. **Query Complexity Analysis**

```go
type QueryMetadata struct {
    Type           QueryType
    Complexity     ComplexityLevel
    EstimatedCost  float64
    TablesAccessed []string
    IndexesUsed    []string
    RequiredLocks  []LockType
    IsReadOnly     bool
    IsTransactional bool
}

type ComplexityLevel int

const (
    SIMPLE_QUERY ComplexityLevel = iota    // Single table, simple predicates
    MODERATE_QUERY                         // Multiple tables, joins
    COMPLEX_QUERY                          // Subqueries, aggregates, window functions
    VERY_COMPLEX_QUERY                     // CTEs, recursive queries, complex analytics
)
```

## Execution Coordination Architecture

### 1. **Execution Coordinator Interface**

```go
type ExecutionCoordinator interface {
    Execute(ctx context.Context, stmt parser.Statement, session *Session) (Result, error)
    ExecuteAsync(ctx context.Context, stmt parser.Statement, session *Session) (<-chan Result, error)
    Prepare(stmt parser.Statement, session *Session) (PreparedStatement, error)
    EstimateCost(stmt parser.Statement, session *Session) (CostEstimate, error)
}
```

### 2. **DQL Coordinator (SELECT Queries)**

```go
type DQLCoordinator struct {
    optimizer    *optimizer.QueryOptimizer
    executor     *executor.QueryExecutor
    resultCache  *cache.ResultCache
    statistics   *stats.StatisticsCollector
}

func (c *DQLCoordinator) Execute(ctx context.Context, stmt *parser.SelectStatement, session *Session) (Result, error) {
    // 1. Semantic analysis and validation
    if err := c.validateQuery(stmt, session); err != nil {
        return nil, err
    }
    
    // 2. Query optimization
    optimizedPlan, err := c.optimizer.Optimize(stmt, session.GetSchema())
    if err != nil {
        return nil, err
    }
    
    // 3. Check result cache
    if cached := c.resultCache.Get(optimizedPlan.Hash()); cached != nil {
        return cached, nil
    }
    
    // 4. Execute query plan
    result, err := c.executor.Execute(ctx, optimizedPlan, session)
    if err != nil {
        return nil, err
    }
    
    // 5. Cache results if appropriate
    if optimizedPlan.IsCacheable() {
        c.resultCache.Put(optimizedPlan.Hash(), result)
    }
    
    // 6. Collect statistics
    c.statistics.RecordExecution(optimizedPlan, result.ExecutionTime())
    
    return result, nil
}
```

### 3. **DML Coordinator (INSERT, UPDATE, DELETE)**

```go
type DMLCoordinator struct {
    transactionMgr *transaction.Manager
    lockManager    *locking.LockManager
    executor       *executor.QueryExecutor
    indexManager   *btree.IndexManager
}

func (c *DMLCoordinator) Execute(ctx context.Context, stmt parser.Statement, session *Session) (Result, error) {
    // 1. Begin transaction if not already in one
    tx, err := c.ensureTransaction(session)
    if err != nil {
        return nil, err
    }
    
    // 2. Acquire necessary locks
    locks, err := c.acquireLocks(stmt, tx)
    if err != nil {
        return nil, err
    }
    defer c.releaseLocks(locks)
    
    // 3. Validate constraints and permissions
    if err := c.validateDMLOperation(stmt, session); err != nil {
        return nil, err
    }
    
    // 4. Execute the statement
    result, err := c.executor.ExecuteDML(ctx, stmt, tx)
    if err != nil {
        return nil, err
    }
    
    // 5. Update indexes
    if err := c.updateIndexes(stmt, result, tx); err != nil {
        return nil, err
    }
    
    // 6. Log changes for replication
    c.logChanges(stmt, result, tx)
    
    return result, nil
}
```

### 4. **DDL Coordinator (CREATE, DROP, ALTER)**

```go
type DDLCoordinator struct {
    schemaManager  *schema.Manager
    lockManager    *locking.LockManager
    metadataStore  *metadata.Store
}

func (c *DDLCoordinator) Execute(ctx context.Context, stmt parser.Statement, session *Session) (Result, error) {
    // 1. Acquire exclusive schema lock
    schemaLock, err := c.lockManager.AcquireSchemaLock(ctx)
    if err != nil {
        return nil, err
    }
    defer schemaLock.Release()
    
    // 2. Validate DDL operation
    if err := c.validateDDLOperation(stmt, session); err != nil {
        return nil, err
    }
    
    // 3. Execute schema change
    result, err := c.executeSchemaChange(stmt, session)
    if err != nil {
        return nil, err
    }
    
    // 4. Update system catalog
    if err := c.updateSystemCatalog(stmt, session); err != nil {
        return nil, err
    }
    
    // 5. Invalidate affected cached plans
    c.invalidateCachedPlans(stmt)
    
    return result, nil
}
```

## Session Management Architecture

### 1. **Session State Management**

```go
type Session struct {
    ID              string
    UserID          string
    DatabaseName    string
    ConnectionTime  time.Time
    LastActivity    time.Time
    TransactionID   *string
    IsolationLevel  IsolationLevel
    ReadOnly        bool
    Variables       map[string]interface{}
    PreparedStmts   map[string]*PreparedStatement
    TempTables      map[string]*TempTable
    mutex           sync.RWMutex
}
```

### 2. **Session Pool Management**

```go
type SessionPool struct {
    sessions    map[string]*Session
    maxSessions int
    idleTimeout time.Duration
    mutex       sync.RWMutex
    cleaner     *time.Ticker
}

func (p *SessionPool) AcquireSession(userID, database string) (*Session, error) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    if len(p.sessions) >= p.maxSessions {
        return nil, ErrTooManySessions
    }
    
    session := &Session{
        ID:             generateSessionID(),
        UserID:         userID,
        DatabaseName:   database,
        ConnectionTime: time.Now(),
        LastActivity:   time.Now(),
        Variables:      make(map[string]interface{}),
        PreparedStmts:  make(map[string]*PreparedStatement),
        TempTables:     make(map[string]*TempTable),
    }
    
    p.sessions[session.ID] = session
    return session, nil
}
```

## Performance Monitoring Architecture

### 1. **Query Execution Metrics**

```go
type ExecutionMetrics struct {
    QueryType        QueryType
    ExecutionTime    time.Duration
    RowsAffected     int64
    RowsExamined     int64
    IndexesUsed      []string
    CacheHitRatio    float64
    MemoryUsed       int64
    IOOperations     int64
    CPUTime          time.Duration
    WaitTime         time.Duration
}
```

### 2. **Performance Monitoring System**

```go
type PerformanceMonitor struct {
    metrics         chan ExecutionMetrics
    aggregatedStats map[QueryType]*AggregatedStats
    slowQueryLog    *SlowQueryLog
    alertManager    *AlertManager
    mutex           sync.RWMutex
}

func (m *PerformanceMonitor) RecordExecution(metrics ExecutionMetrics) {
    // Async processing to avoid blocking query execution
    select {
    case m.metrics <- metrics:
    default:
        // Channel full, log warning but don't block
        log.Warn("Performance metrics channel full, dropping metric")
    }
}
```

### 3. **Slow Query Detection**

```go
type SlowQueryLog struct {
    threshold     time.Duration
    logFile       *os.File
    buffer        *bufio.Writer
    maxLogSize    int64
    rotationSize  int64
}

func (l *SlowQueryLog) LogSlowQuery(stmt parser.Statement, metrics ExecutionMetrics) {
    if metrics.ExecutionTime < l.threshold {
        return
    }
    
    entry := SlowQueryEntry{
        Timestamp:     time.Now(),
        Query:         stmt.String(),
        ExecutionTime: metrics.ExecutionTime,
        RowsExamined:  metrics.RowsExamined,
        RowsAffected:  metrics.RowsAffected,
    }
    
    l.writeEntry(entry)
}
```

## Security and Access Control

### 1. **Permission Validation**

```go
type AccessController struct {
    userPermissions map[string][]Permission
    rolePermissions map[string][]Permission
    objectACLs      map[string]ACL
}

func (ac *AccessController) ValidateAccess(session *Session, stmt parser.Statement) error {
    requiredPerms := ac.getRequiredPermissions(stmt)
    userPerms := ac.getUserPermissions(session.UserID)
    
    for _, required := range requiredPerms {
        if !ac.hasPermission(userPerms, required) {
            return ErrAccessDenied
        }
    }
    
    return nil
}
```

### 2. **SQL Injection Prevention**

```go
type SecurityValidator struct {
    suspiciousPatterns []regexp.Regexp
    allowedFunctions   map[string]bool
    maxQueryLength     int
}

func (sv *SecurityValidator) ValidateQuery(stmt parser.Statement, session *Session) error {
    queryText := stmt.String()
    
    // Check query length
    if len(queryText) > sv.maxQueryLength {
        return ErrQueryTooLong
    }
    
    // Check for suspicious patterns
    for _, pattern := range sv.suspiciousPatterns {
        if pattern.MatchString(queryText) {
            return ErrSuspiciousQuery
        }
    }
    
    // Validate function calls
    if err := sv.validateFunctions(stmt); err != nil {
        return err
    }
    
    return nil
}
```

## Integration Points

### 1. **Parser Integration**

```go
type ParserIntegration struct {
    parser *parser.Parser
    cache  *cache.ParseCache
}

func (pi *ParserIntegration) ParseStatement(sql string) (parser.Statement, error) {
    // Check parse cache first
    if cached := pi.cache.Get(sql); cached != nil {
        return cached, nil
    }
    
    // Parse SQL statement
    stmt, err := pi.parser.Parse(sql)
    if err != nil {
        return nil, err
    }
    
    // Cache parsed statement
    pi.cache.Put(sql, stmt)
    return stmt, nil
}
```

### 2. **Query Optimizer Integration**

```go
type OptimizerIntegration struct {
    optimizer *optimizer.QueryOptimizer
    planCache *cache.PlanCache
}

func (oi *OptimizerIntegration) OptimizeQuery(stmt parser.Statement, schema *schema.Schema) (*optimizer.QueryPlan, error) {
    planKey := generatePlanKey(stmt, schema)
    
    // Check plan cache
    if cached := oi.planCache.Get(planKey); cached != nil {
        return cached, nil
    }
    
    // Optimize query
    plan, err := oi.optimizer.Optimize(stmt, schema)
    if err != nil {
        return nil, err
    }
    
    // Cache optimized plan
    oi.planCache.Put(planKey, plan)
    return plan, nil
}
```

### 3. **Transaction Manager Integration**

```go
type TransactionIntegration struct {
    txManager *transaction.Manager
    sessions  map[string]*Session
    mutex     sync.RWMutex
}

func (ti *TransactionIntegration) GetTransaction(session *Session) (*transaction.Transaction, error) {
    ti.mutex.RLock()
    defer ti.mutex.RUnlock()
    
    if session.TransactionID != nil {
        return ti.txManager.GetTransaction(*session.TransactionID)
    }
    
    return nil, nil // No active transaction
}
```

## Threading and Concurrency

### 1. **Thread Safety Design**

```go
type Dispatcher struct {
    coordinators    map[QueryType]ExecutionCoordinator
    sessionPool     *SessionPool
    perfMonitor     *PerformanceMonitor
    accessController *AccessController
    
    // Thread-safe collections
    activeSessions  sync.Map // session_id -> *Session
    activeQueries   sync.Map // query_id -> *QueryExecution
    
    // Configuration (read-only after initialization)
    config *Config
    
    // Metrics and monitoring
    metrics *Metrics
}
```

### 2. **Concurrent Query Execution**

```go
func (d *Dispatcher) ExecuteConcurrent(queries []QueryRequest) []QueryResult {
    results := make([]QueryResult, len(queries))
    var wg sync.WaitGroup
    
    // Execute queries concurrently with controlled parallelism
    semaphore := make(chan struct{}, d.config.MaxConcurrentQueries)
    
    for i, query := range queries {
        wg.Add(1)
        go func(index int, q QueryRequest) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Execute query
            result, err := d.Execute(q.Context, q.SQL, q.Session)
            results[index] = QueryResult{
                Data:  result,
                Error: err,
            }
        }(i, query)
    }
    
    wg.Wait()
    return results
}
```

### 3. **Query Cancellation Support**

```go
type QueryExecution struct {
    ID        string
    Context   context.Context
    Cancel    context.CancelFunc
    Session   *Session
    StartTime time.Time
    Status    ExecutionStatus
}

func (d *Dispatcher) CancelQuery(queryID string) error {
    if exec, exists := d.activeQueries.Load(queryID); exists {
        execution := exec.(*QueryExecution)
        execution.Cancel()
        execution.Status = StatusCancelled
        return nil
    }
    return ErrQueryNotFound
}
```

## Configuration and Extensibility

### 1. **Configuration Management**

```go
type Config struct {
    MaxConcurrentQueries    int           `yaml:"max_concurrent_queries"`
    QueryTimeout           time.Duration `yaml:"query_timeout"`
    SlowQueryThreshold     time.Duration `yaml:"slow_query_threshold"`
    MaxSessionsPerUser     int           `yaml:"max_sessions_per_user"`
    CacheSize              int           `yaml:"cache_size"`
    EnableQueryLogging     bool          `yaml:"enable_query_logging"`
    SecurityLevel          string        `yaml:"security_level"`
}
```

### 2. **Plugin Architecture**

```go
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}) error
    PreExecute(ctx context.Context, stmt parser.Statement, session *Session) error
    PostExecute(ctx context.Context, stmt parser.Statement, result Result, session *Session) error
    Shutdown() error
}

type PluginManager struct {
    plugins []Plugin
    enabled map[string]bool
}

func (pm *PluginManager) RegisterPlugin(plugin Plugin) {
    pm.plugins = append(pm.plugins, plugin)
    pm.enabled[plugin.Name()] = true
}
```

## Error Handling and Recovery

### 1. **Error Classification**

```go
type DispatcherError struct {
    Code      ErrorCode
    Message   string
    Query     string
    Session   *Session
    Cause     error
    Timestamp time.Time
}

type ErrorCode int

const (
    ErrInvalidQuery ErrorCode = iota
    ErrAccessDenied
    ErrResourceExhausted
    ErrQueryTimeout
    ErrTransactionConflict
    ErrSystemError
)
```

### 2. **Circuit Breaker Pattern**

```go
type CircuitBreaker struct {
    maxFailures   int
    resetTimeout  time.Duration
    state         CircuitState
    failures      int
    lastFailTime  time.Time
    mutex         sync.Mutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if cb.state == StateOpen {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = StateHalfOpen
        } else {
            return ErrCircuitBreakerOpen
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
        return err
    }
    
    cb.failures = 0
    cb.state = StateClosed
    return nil
}
```

## Future Architecture Considerations

### 1. **Distributed Execution**
- **Query Distribution**: Distribute queries across multiple nodes
- **Result Aggregation**: Combine results from distributed execution
- **Fault Tolerance**: Handle node failures gracefully
- **Load Balancing**: Balance query load across available nodes

### 2. **Real-time Analytics**
- **Stream Processing**: Support for continuous queries
- **Event-driven Execution**: React to data changes in real-time
- **Materialized Views**: Automatically maintained aggregate views
- **Complex Event Processing**: Pattern matching over data streams

### 3. **Machine Learning Integration**
- **Query Performance Prediction**: ML models for execution time prediction
- **Automatic Query Optimization**: AI-driven query optimization
- **Anomaly Detection**: Detect unusual query patterns
- **Resource Usage Prediction**: Predict resource requirements

### 4. **Cloud Integration**
- **Serverless Execution**: Function-as-a-Service query execution
- **Auto-scaling**: Automatic resource scaling based on load
- **Multi-tenant Isolation**: Secure multi-tenant query execution
- **Cost Optimization**: Minimize cloud resource costs