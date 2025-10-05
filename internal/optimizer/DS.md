# Query Optimizer Module Data Structures

## Overview
This document describes the data structures used in the SQL query optimizer for representing logical and physical plans, maintaining statistics, managing costs, and caching optimization results.

## Core Plan Data Structures

### 1. Logical Plan Nodes

#### LogicalPlan Interface
```go
type LogicalPlan interface {
    // Plan identification and properties
    ID() PlanNodeID
    Type() LogicalNodeType
    Schema() *Schema
    Children() []LogicalPlan
    
    // Cost estimation support
    EstimateCardinality(stats *Statistics) int64
    GetRequiredProperties() *RequiredProperties
    
    // Plan transformation
    Clone() LogicalPlan
    ReplaceChild(index int, child LogicalPlan) LogicalPlan
    Accept(visitor LogicalPlanVisitor) error
    
    // String representation
    String() string
}

// Base implementation
type LogicalPlanBase struct {
    id          PlanNodeID
    nodeType    LogicalNodeType
    schema      *Schema
    children    []LogicalPlan
    properties  *RequiredProperties
    cardinality int64
}
```

#### Specific Logical Operators
```go
// Table scan operation
type LogicalTableScan struct {
    LogicalPlanBase
    TableName   string
    Schema      *TableSchema
    Filter      Expression        // WHERE clause predicates
    Projection  []ColumnRef       // Selected columns
    Alias       string
}

// Join operation
type LogicalJoin struct {
    LogicalPlanBase
    JoinType     JoinType         // INNER, LEFT, RIGHT, FULL, SEMI, ANTI
    LeftChild    LogicalPlan
    RightChild   LogicalPlan
    JoinCondition Expression      // ON clause
    Filter       Expression       // Additional WHERE predicates
}

// Aggregation operation
type LogicalAggregation struct {
    LogicalPlanBase
    Child        LogicalPlan
    GroupBy      []Expression     // GROUP BY expressions
    Aggregates   []AggregateFunc  // Aggregate functions
    Having       Expression       // HAVING clause
}

// Projection operation
type LogicalProjection struct {
    LogicalPlanBase
    Child        LogicalPlan
    Projections  []NamedExpression
}

// Selection operation
type LogicalSelection struct {
    LogicalPlanBase
    Child        LogicalPlan
    Condition    Expression
}

// Sort operation
type LogicalSort struct {
    LogicalPlanBase
    Child        LogicalPlan
    SortKeys     []SortKey
    Limit        *int64
    Offset       *int64
}

// Union operation
type LogicalUnion struct {
    LogicalPlanBase
    Children     []LogicalPlan
    UnionType    UnionType        // UNION, UNION ALL
}
```

### 2. Physical Plan Nodes

#### PhysicalPlan Interface
```go
type PhysicalPlan interface {
    // Plan identification and properties
    ID() PlanNodeID
    Type() PhysicalNodeType
    Schema() *Schema
    Children() []PhysicalPlan
    
    // Cost and properties
    Cost() *Cost
    Properties() *PhysicalProperties
    
    // Execution support
    Open(ctx *ExecutionContext) error
    Next() (*Tuple, error)
    Close() error
    
    // Plan transformation
    Clone() PhysicalPlan
    ReplaceChild(index int, child PhysicalPlan) PhysicalPlan
    
    // String representation
    String() string
}

// Base implementation
type PhysicalPlanBase struct {
    id         PlanNodeID
    nodeType   PhysicalNodeType
    schema     *Schema
    children   []PhysicalPlan
    cost       *Cost
    properties *PhysicalProperties
}
```

#### Physical Scan Operators
```go
// Sequential table scan
type PhysicalTableScan struct {
    PhysicalPlanBase
    TableName    string
    Schema       *TableSchema
    Filter       Expression
    Projection   []ColumnRef
    EstimatedRows int64
}

// Index scan operation
type PhysicalIndexScan struct {
    PhysicalPlanBase
    TableName    string
    IndexName    string
    Schema       *TableSchema
    ScanCondition Expression      // Index key conditions
    Filter       Expression       // Additional filter conditions
    ScanDirection ScanDirection   // FORWARD, BACKWARD
    EstimatedRows int64
}
```

#### Physical Join Operators
```go
// Nested loop join
type PhysicalNestedLoopJoin struct {
    PhysicalPlanBase
    LeftChild     PhysicalPlan
    RightChild    PhysicalPlan
    JoinType      JoinType
    JoinCondition Expression
    Filter        Expression
}

// Hash join
type PhysicalHashJoin struct {
    PhysicalPlanBase
    LeftChild     PhysicalPlan     // Build side
    RightChild    PhysicalPlan     // Probe side
    JoinType      JoinType
    BuildKeys     []Expression     // Hash keys for build side
    ProbeKeys     []Expression     // Hash keys for probe side
    JoinCondition Expression
    Filter        Expression
    HashTableSize int64            // Estimated hash table size
}

// Sort-merge join
type PhysicalSortMergeJoin struct {
    PhysicalPlanBase
    LeftChild     PhysicalPlan
    RightChild    PhysicalPlan
    JoinType      JoinType
    LeftSortKeys  []SortKey
    RightSortKeys []SortKey
    JoinCondition Expression
    Filter        Expression
}
```

#### Physical Aggregation and Sort Operators
```go
// Hash-based aggregation
type PhysicalHashAggregation struct {
    PhysicalPlanBase
    Child         PhysicalPlan
    GroupBy       []Expression
    Aggregates    []AggregateFunc
    Having        Expression
    HashTableSize int64
}

// Sort-based aggregation
type PhysicalSortAggregation struct {
    PhysicalPlanBase
    Child         PhysicalPlan
    GroupBy       []Expression
    Aggregates    []AggregateFunc
    Having        Expression
    RequiresSorting bool
}

// Sort operator
type PhysicalSort struct {
    PhysicalPlanBase
    Child         PhysicalPlan
    SortKeys      []SortKey
    Limit         *int64
    Offset        *int64
    Algorithm     SortAlgorithm    // QUICKSORT, MERGESORT, EXTERNAL
    MemoryLimit   int64
}
```

## Cost and Statistics Data Structures

### 3. Cost Model Data Structures

#### Cost Structure
```go
type Cost struct {
    IOCost     float64    // I/O operations cost
    CPUCost    float64    // CPU processing cost
    NetworkCost float64   // Network transfer cost (for distributed)
    MemoryCost float64    // Memory usage cost
    TotalCost  float64    // Weighted total cost
}

func (c *Cost) Add(other *Cost) *Cost {
    return &Cost{
        IOCost:      c.IOCost + other.IOCost,
        CPUCost:     c.CPUCost + other.CPUCost,
        NetworkCost: c.NetworkCost + other.NetworkCost,
        MemoryCost:  c.MemoryCost + other.MemoryCost,
        TotalCost:   c.TotalCost + other.TotalCost,
    }
}

func (c *Cost) Multiply(factor float64) *Cost {
    return &Cost{
        IOCost:      c.IOCost * factor,
        CPUCost:     c.CPUCost * factor,
        NetworkCost: c.NetworkCost * factor,
        MemoryCost:  c.MemoryCost * factor,
        TotalCost:   c.TotalCost * factor,
    }
}

// Cost factors for different operations
type CostFactors struct {
    SequentialIOCost    float64    // Cost per page read (sequential)
    RandomIOCost        float64    // Cost per page read (random)
    CPUTupleCost        float64    // Cost per tuple processed
    CPUOperatorCost     float64    // Cost per operator evaluation
    MemoryPageCost      float64    // Cost per memory page
    NetworkTransferCost float64    // Cost per byte transferred
}
```

### 4. Statistics Data Structures

#### Table Statistics
```go
type TableStatistics struct {
    TableName     string
    RowCount      int64
    PageCount     int64
    AvgTupleSize  int32
    LastUpdated   time.Time
    
    // Column statistics
    ColumnStats   map[string]*ColumnStatistics
    
    // Index statistics
    IndexStats    map[string]*IndexStatistics
}

type ColumnStatistics struct {
    ColumnName    string
    DataType      DataType
    NullCount     int64
    DistinctCount int64          // Number of distinct values (NDV)
    MinValue      interface{}
    MaxValue      interface{}
    AvgWidth      int32          // Average column width in bytes
    
    // Histogram for value distribution
    Histogram     *Histogram
    
    // Most common values
    MCV           *MostCommonValues
    
    LastUpdated   time.Time
}

type IndexStatistics struct {
    IndexName     string
    TableName     string
    Columns       []string
    IndexType     IndexType
    KeyCount      int64          // Number of index entries
    LeafPages     int64          // Number of leaf pages
    Height        int32          // B+ tree height
    Selectivity   float64        // Overall index selectivity
    Clustered     bool           // Whether index is clustered
    LastUpdated   time.Time
}
```

#### Histogram Data Structures
```go
type Histogram interface {
    EstimateSelectivity(predicate *Predicate) float64
    EstimateCardinality(totalRows int64, predicate *Predicate) int64
    GetBucketCount() int
    Merge(other Histogram) Histogram
}

// Equal-width histogram
type EqualWidthHistogram struct {
    Buckets     []*HistogramBucket
    BucketCount int
    MinValue    interface{}
    MaxValue    interface{}
    TotalCount  int64
}

// Equal-depth histogram (better for skewed data)
type EqualDepthHistogram struct {
    Buckets     []*HistogramBucket
    BucketCount int
    TotalCount  int64
    BucketSize  int64          // Target rows per bucket
}

type HistogramBucket struct {
    LowerBound  interface{}    // Bucket lower bound
    UpperBound  interface{}    // Bucket upper bound
    Count       int64          // Rows in this bucket
    DistinctCount int64        // Distinct values in bucket
}

// Most common values (for handling skewed distributions)
type MostCommonValues struct {
    Values      []interface{}  // The actual values
    Frequencies []int64        // Frequency of each value
    TotalCount  int64          // Total rows represented
}
```

### 5. Plan Properties Data Structures

#### Physical Properties
```go
type PhysicalProperties struct {
    // Ordering properties
    SortOrder     []SortKey      // Current sort order
    
    // Distribution properties (for distributed systems)
    Distribution  *Distribution
    
    // Partitioning properties
    Partitioning  *Partitioning
    
    // Memory requirements
    MemoryReq     int64         // Memory requirement in bytes
    
    // Parallelism properties
    Parallelism   ParallelismInfo
}

type SortKey struct {
    Expression  Expression
    Direction   SortDirection   // ASC, DESC
    NullsFirst  bool
}

type Distribution struct {
    Type        DistributionType  // BROADCAST, HASH, RANGE, REPLICATED
    Keys        []Expression      // Distribution keys
    Partitions  int              // Number of partitions
}

type Partitioning struct {
    Type        PartitionType     // HASH, RANGE, LIST
    Keys        []Expression      // Partitioning keys
    Partitions  []PartitionInfo
}

type ParallelismInfo struct {
    MaxWorkers  int              // Maximum parallel workers
    Current     int              // Current parallelism level
    Scalable    bool             // Can scale parallelism
}
```

#### Required Properties
```go
type RequiredProperties struct {
    // Required sort order (for ORDER BY, merge joins, etc.)
    RequiredSort    []SortKey
    
    // Required distribution (for distributed joins)
    RequiredDist    *Distribution
    
    // Required partitioning
    RequiredPart    *Partitioning
    
    // Limits and constraints
    RowLimit        *int64        // LIMIT clause
    MemoryLimit     *int64        // Memory constraint
}
```

## Plan Enumeration Data Structures

### 6. Plan Space Exploration

#### Plan Alternatives
```go
type PlanAlternative struct {
    Plan        PhysicalPlan
    Cost        *Cost
    Properties  *PhysicalProperties
    Rank        int              // Ranking among alternatives
}

type PlanAlternatives struct {
    Alternatives []*PlanAlternative
    BestPlan     *PlanAlternative
    mutex        sync.RWMutex
}

func (pa *PlanAlternatives) Add(alt *PlanAlternative) {
    pa.mutex.Lock()
    defer pa.mutex.Unlock()
    
    pa.Alternatives = append(pa.Alternatives, alt)
    
    // Update best plan if this alternative is better
    if pa.BestPlan == nil || alt.Cost.TotalCost < pa.BestPlan.Cost.TotalCost {
        pa.BestPlan = alt
    }
}

func (pa *PlanAlternatives) GetBest() *PlanAlternative {
    pa.mutex.RLock()
    defer pa.mutex.RUnlock()
    return pa.BestPlan
}
```

#### Dynamic Programming Table
```go
type DPTable struct {
    // Maps relation sets to optimal plans
    table       map[RelationSet]*PlanAlternatives
    mutex       sync.RWMutex
    
    // Statistics for debugging
    lookups     int64
    hits        int64
    misses      int64
}

type RelationSet struct {
    Relations   []string         // Set of relation names
    hash        uint64           // Cached hash value
}

func (rs *RelationSet) Hash() uint64 {
    if rs.hash == 0 {
        h := fnv.New64a()
        for _, rel := range rs.Relations {
            h.Write([]byte(rel))
        }
        rs.hash = h.Sum64()
    }
    return rs.hash
}

func (dt *DPTable) Get(relations RelationSet) (*PlanAlternatives, bool) {
    dt.mutex.RLock()
    defer dt.mutex.RUnlock()
    
    dt.lookups++
    alts, exists := dt.table[relations]
    if exists {
        dt.hits++
    } else {
        dt.misses++
    }
    return alts, exists
}

func (dt *DPTable) Put(relations RelationSet, alternatives *PlanAlternatives) {
    dt.mutex.Lock()
    defer dt.mutex.Unlock()
    
    if dt.table == nil {
        dt.table = make(map[RelationSet]*PlanAlternatives)
    }
    dt.table[relations] = alternatives
}
```

## Expression and Schema Data Structures

### 7. Expression Trees

#### Expression Interface
```go
type Expression interface {
    // Type information
    Type() DataType
    Nullable() bool
    
    // Expression evaluation
    Evaluate(tuple *Tuple) (interface{}, error)
    
    // Expression analysis
    GetColumns() []ColumnRef
    IsConstant() bool
    IsDeterministic() bool
    
    // Expression transformation
    Clone() Expression
    Accept(visitor ExpressionVisitor) Expression
    
    // String representation
    String() string
}

// Column reference
type ColumnRef struct {
    TableName  string
    ColumnName string
    Alias      string
    DataType   DataType
    Nullable   bool
}

// Binary operation
type BinaryExpression struct {
    Left      Expression
    Right     Expression
    Operator  BinaryOperator
    DataType  DataType
}

// Function call
type FunctionCall struct {
    FunctionName string
    Arguments    []Expression
    ReturnType   DataType
    Aggregate    bool
}

// Constant value
type ConstantExpression struct {
    Value     interface{}
    DataType  DataType
}

// Predicate expressions
type ComparisonExpression struct {
    Left      Expression
    Right     Expression
    Operator  ComparisonOperator  // =, <>, <, <=, >, >=
}

type LogicalExpression struct {
    Operands  []Expression
    Operator  LogicalOperator     // AND, OR, NOT
}
```

### 8. Schema Definitions

#### Schema Structure
```go
type Schema struct {
    Columns     []*ColumnDef
    columnIndex map[string]int    // Column name to index mapping
    mutex       sync.RWMutex
}

type ColumnDef struct {
    Name        string
    DataType    DataType
    Nullable    bool
    DefaultValue interface{}
    Comment     string
}

func (s *Schema) GetColumn(name string) (*ColumnDef, bool) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    if idx, exists := s.columnIndex[name]; exists {
        return s.Columns[idx], true
    }
    return nil, false
}

func (s *Schema) GetColumnIndex(name string) (int, bool) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    idx, exists := s.columnIndex[name]
    return idx, exists
}

func (s *Schema) AddColumn(col *ColumnDef) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.columnIndex == nil {
        s.columnIndex = make(map[string]int)
    }
    
    s.columnIndex[col.Name] = len(s.Columns)
    s.Columns = append(s.Columns, col)
}
```

## Plan Caching Data Structures

### 9. Plan Cache

#### Cache Entry Structure
```go
type PlanCacheEntry struct {
    QueryHash       uint64
    PhysicalPlan    PhysicalPlan
    Statistics      map[string]*TableStatistics  // Stats when plan was created
    CreationTime    time.Time
    LastAccessed    time.Time
    AccessCount     int64
    MemorySize      int64                        // Memory used by cached plan
    ReferencedTables []string                    // Tables used in query
}

type PlanCache struct {
    // Main cache storage
    cache       map[uint64]*PlanCacheEntry
    
    // LRU tracking
    lruList     *list.List
    lruMap      map[uint64]*list.Element
    
    // Cache configuration
    maxSize     int64              // Maximum cache size in bytes
    maxEntries  int                // Maximum number of entries
    currentSize int64              // Current cache size
    
    // Statistics
    hits        int64
    misses      int64
    evictions   int64
    
    mutex       sync.RWMutex
}

func (pc *PlanCache) Get(queryHash uint64) (*PlanCacheEntry, bool) {
    pc.mutex.Lock()
    defer pc.mutex.Unlock()
    
    entry, exists := pc.cache[queryHash]
    if !exists {
        pc.misses++
        return nil, false
    }
    
    // Update LRU
    if elem, exists := pc.lruMap[queryHash]; exists {
        pc.lruList.MoveToFront(elem)
    }
    
    // Update access statistics
    entry.LastAccessed = time.Now()
    entry.AccessCount++
    
    pc.hits++
    return entry, true
}

func (pc *PlanCache) Put(queryHash uint64, plan PhysicalPlan, stats map[string]*TableStatistics) {
    pc.mutex.Lock()
    defer pc.mutex.Unlock()
    
    // Create new entry
    entry := &PlanCacheEntry{
        QueryHash:       queryHash,
        PhysicalPlan:    plan,
        Statistics:      stats,
        CreationTime:    time.Now(),
        LastAccessed:    time.Now(),
        AccessCount:     1,
        MemorySize:      estimatePlanSize(plan),
        ReferencedTables: extractReferencedTables(plan),
    }
    
    // Check if eviction is needed
    for pc.currentSize + entry.MemorySize > pc.maxSize || len(pc.cache) >= pc.maxEntries {
        pc.evictLRU()
    }
    
    // Add to cache
    pc.cache[queryHash] = entry
    pc.currentSize += entry.MemorySize
    
    // Add to LRU
    elem := pc.lruList.PushFront(queryHash)
    pc.lruMap[queryHash] = elem
}

func (pc *PlanCache) evictLRU() {
    if pc.lruList.Len() == 0 {
        return
    }
    
    // Get least recently used entry
    elem := pc.lruList.Back()
    queryHash := elem.Value.(uint64)
    
    // Remove from cache
    entry := pc.cache[queryHash]
    delete(pc.cache, queryHash)
    pc.currentSize -= entry.MemorySize
    
    // Remove from LRU tracking
    pc.lruList.Remove(elem)
    delete(pc.lruMap, queryHash)
    
    pc.evictions++
}
```

## Join Enumeration Data Structures

### 10. Join Graph

#### Join Graph Structure
```go
type JoinGraph struct {
    // Relations in the query
    Relations   map[string]*RelationNode
    
    // Edges represent join conditions
    Edges       []*JoinEdge
    
    // Adjacency list for efficient traversal
    AdjList     map[string][]string
    
    // Join predicates organized by relation pairs
    Predicates  map[RelationPair][]Expression
}

type RelationNode struct {
    Name        string
    Alias       string
    Schema      *TableSchema
    Filter      Expression         // Local predicates
    Cardinality int64
}

type JoinEdge struct {
    Left        string             // Left relation name
    Right       string             // Right relation name
    JoinType    JoinType
    Condition   Expression         // Join condition
    Selectivity float64            // Estimated join selectivity
}

type RelationPair struct {
    Left  string
    Right string
}

func (jg *JoinGraph) AddRelation(rel *RelationNode) {
    jg.Relations[rel.Name] = rel
    jg.AdjList[rel.Name] = make([]string, 0)
}

func (jg *JoinGraph) AddJoinEdge(edge *JoinEdge) {
    jg.Edges = append(jg.Edges, edge)
    
    // Update adjacency list
    jg.AdjList[edge.Left] = append(jg.AdjList[edge.Left], edge.Right)
    jg.AdjList[edge.Right] = append(jg.AdjList[edge.Right], edge.Left)
    
    // Store join predicate
    pair := RelationPair{Left: edge.Left, Right: edge.Right}
    if jg.Predicates[pair] == nil {
        jg.Predicates[pair] = make([]Expression, 0)
    }
    jg.Predicates[pair] = append(jg.Predicates[pair], edge.Condition)
}

func (jg *JoinGraph) GetJoinableRelations(rel string) []string {
    return jg.AdjList[rel]
}

func (jg *JoinGraph) GetJoinCondition(left, right string) []Expression {
    pair1 := RelationPair{Left: left, Right: right}
    pair2 := RelationPair{Left: right, Right: left}
    
    if predicates, exists := jg.Predicates[pair1]; exists {
        return predicates
    }
    if predicates, exists := jg.Predicates[pair2]; exists {
        return predicates
    }
    return nil
}
```

## Memory Management Data Structures

### 11. Memory Pool for Optimization

#### Memory Pool Structure
```go
type OptimizerMemoryPool struct {
    // Memory pools for different object types
    planPool        *sync.Pool      // For plan nodes
    exprPool        *sync.Pool      // For expressions
    costPool        *sync.Pool      // For cost objects
    statisticsPool  *sync.Pool      // For statistics objects
    
    // Memory tracking
    totalAllocated  int64
    peakAllocated   int64
    currentUsage    int64
    
    // Pool configuration
    maxPoolSize     int64
    enableTracking  bool
    
    mutex           sync.RWMutex
}

func (mp *OptimizerMemoryPool) GetPlan() *PhysicalPlanBase {
    obj := mp.planPool.Get()
    if obj == nil {
        return &PhysicalPlanBase{}
    }
    return obj.(*PhysicalPlanBase)
}

func (mp *OptimizerMemoryPool) ReturnPlan(plan *PhysicalPlanBase) {
    // Reset plan for reuse
    plan.Reset()
    mp.planPool.Put(plan)
}

func (mp *OptimizerMemoryPool) GetCost() *Cost {
    obj := mp.costPool.Get()
    if obj == nil {
        return &Cost{}
    }
    return obj.(*Cost)
}

func (mp *OptimizerMemoryPool) ReturnCost(cost *Cost) {
    cost.Reset()
    mp.costPool.Put(cost)
}

func (mp *OptimizerMemoryPool) TrackAllocation(size int64) {
    if !mp.enableTracking {
        return
    }
    
    mp.mutex.Lock()
    defer mp.mutex.Unlock()
    
    mp.currentUsage += size
    mp.totalAllocated += size
    
    if mp.currentUsage > mp.peakAllocated {
        mp.peakAllocated = mp.currentUsage
    }
}

func (mp *OptimizerMemoryPool) TrackDeallocation(size int64) {
    if !mp.enableTracking {
        return
    }
    
    mp.mutex.Lock()
    defer mp.mutex.Unlock()
    
    mp.currentUsage -= size
}
```

## Data Structure Size Analysis

### Memory Usage Patterns

| Data Structure | Base Size (bytes) | Variable Components | Growth Pattern |
|---------------|-------------------|-------------------|----------------|
| LogicalPlanBase | 64 | + children + expressions | O(plan_depth) |
| PhysicalPlanBase | 80 | + children + cost + properties | O(plan_depth) |
| TableStatistics | 96 | + columns * 200 + indexes * 150 | O(schema_size) |
| ColumnStatistics | 200 | + histogram buckets * 50 | O(histogram_size) |
| PlanCacheEntry | 128 | + plan size + statistics | O(plan_complexity) |
| JoinGraph | 48 | + relations * 100 + edges * 80 | O(relations²) |
| DPTable | 32 | + entries * (relations * 8 + plan_size) | O(2ⁿ) |
| Cost | 40 | Fixed size | O(1) |
| Schema | 32 | + columns * 80 | O(column_count) |
| Expression | 24-200 | Varies by expression type | O(expression_depth) |

### Optimization Memory Requirements

```go
type OptimizerMemoryEstimator struct{}

func (ome *OptimizerMemoryEstimator) EstimateOptimizationMemory(query *Query) int64 {
    baseMemory := int64(1024 * 1024) // 1MB base
    
    // Plan enumeration memory (exponential in number of relations)
    relationCount := int64(len(query.Relations))
    if relationCount <= 10 {
        baseMemory += (1 << relationCount) * 500 // DP table
    } else {
        baseMemory += relationCount * relationCount * 1000 // Heuristic
    }
    
    // Statistics memory
    baseMemory += relationCount * 50000 // Table stats
    
    // Expression memory
    exprCount := ome.countExpressions(query)
    baseMemory += exprCount * 200
    
    // Plan cache overhead
    baseMemory += 10 * 1024 * 1024 // 10MB for plan cache
    
    return baseMemory
}

func (ome *OptimizerMemoryEstimator) countExpressions(query *Query) int64 {
    // Count expressions in WHERE, SELECT, JOIN conditions, etc.
    count := int64(0)
    
    // Implement expression counting logic
    // This is a simplified version
    count += int64(len(query.SelectList)) * 2
    count += int64(len(query.WhereConditions)) * 3
    count += int64(len(query.JoinConditions)) * 2
    count += int64(len(query.GroupBy)) * 1
    count += int64(len(query.OrderBy)) * 1
    
    return count
}
```

## Concurrency and Thread Safety

### 12. Thread-Safe Data Structures

#### Concurrent Plan Cache
```go
type ConcurrentPlanCache struct {
    // Sharded cache to reduce lock contention
    shards      []*PlanCacheShard
    shardCount  int
    shardMask   uint64
}

type PlanCacheShard struct {
    cache       map[uint64]*PlanCacheEntry
    lruList     *list.List
    lruMap      map[uint64]*list.Element
    currentSize int64
    mutex       sync.RWMutex
}

func (cpc *ConcurrentPlanCache) getShard(queryHash uint64) *PlanCacheShard {
    shardIndex := queryHash & cpc.shardMask
    return cpc.shards[shardIndex]
}

func (cpc *ConcurrentPlanCache) Get(queryHash uint64) (*PlanCacheEntry, bool) {
    shard := cpc.getShard(queryHash)
    return shard.Get(queryHash)
}

func (cpc *ConcurrentPlanCache) Put(queryHash uint64, plan PhysicalPlan, stats map[string]*TableStatistics) {
    shard := cpc.getShard(queryHash)
    shard.Put(queryHash, plan, stats)
}
```

#### Lock-Free Statistics Updates
```go
type AtomicColumnStatistics struct {
    ColumnName    string
    DataType      DataType
    NullCount     *int64          // atomic
    DistinctCount *int64          // atomic
    LastUpdated   *int64          // atomic timestamp
    
    // Complex fields protected by RW mutex
    histogram     *Histogram
    mcv           *MostCommonValues
    mutex         sync.RWMutex
}

func (acs *AtomicColumnStatistics) UpdateNullCount(delta int64) {
    atomic.AddInt64(acs.NullCount, delta)
    atomic.StoreInt64(acs.LastUpdated, time.Now().UnixNano())
}

func (acs *AtomicColumnStatistics) UpdateDistinctCount(newValue int64) {
    atomic.StoreInt64(acs.DistinctCount, newValue)
    atomic.StoreInt64(acs.LastUpdated, time.Now().UnixNano())
}

func (acs *AtomicColumnStatistics) GetHistogram() *Histogram {
    acs.mutex.RLock()
    defer acs.mutex.RUnlock()
    return acs.histogram
}

func (acs *AtomicColumnStatistics) SetHistogram(h *Histogram) {
    acs.mutex.Lock()
    defer acs.mutex.Unlock()
    acs.histogram = h
    atomic.StoreInt64(acs.LastUpdated, time.Now().UnixNano())
}
```

## Data Structure Validation

### 13. Validation and Debugging Support

#### Plan Validation Structures
```go
type PlanValidator struct {
    errors   []ValidationError
    warnings []ValidationWarning
}

type ValidationError struct {
    ErrorType   ValidationErrorType
    Message     string
    NodeID      PlanNodeID
    Location    string
}

type ValidationWarning struct {
    WarningType ValidationWarningType  
    Message     string
    NodeID      PlanNodeID
    Severity    WarningSeverity
}

func (pv *PlanValidator) ValidatePlan(plan PhysicalPlan) ValidationResult {
    pv.errors = nil
    pv.warnings = nil
    
    // Validate plan structure
    pv.validateStructure(plan)
    
    // Validate data flow
    pv.validateDataFlow(plan)
    
    // Validate cost estimates
    pv.validateCosts(plan)
    
    // Validate properties
    pv.validateProperties(plan)
    
    return ValidationResult{
        IsValid:  len(pv.errors) == 0,
        Errors:   pv.errors,
        Warnings: pv.warnings,
    }
}
```

This comprehensive data structures documentation provides the foundation for implementing a sophisticated query optimizer with efficient memory management, thread safety, and robust plan representation capabilities.