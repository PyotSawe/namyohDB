// Package executor implements the query execution engine for NamyohDB.
// It executes optimized query plans using the Volcano/Iterator model.
package executor

import (
	"context"
	"fmt"
	"time"

	"relational-db/internal/optimizer"
	"relational-db/internal/storage"
)

// Executor executes optimized query plans
type Executor struct {
	storage    storage.StorageEngine
	bufferPool *storage.BufferPool
	statistics *ExecutionStatistics
	config     *ExecutorConfig
}

// ExecutorConfig contains configuration for the executor
type ExecutorConfig struct {
	// Memory limits
	MaxMemoryBytes int64
	WorkMemBytes   int64 // Memory per operator

	// Parallelism
	MaxParallelism int
	EnableParallel bool

	// Timeouts
	QueryTimeout    time.Duration
	OperatorTimeout time.Duration

	// Optimization flags
	EnablePipelining bool
	EnableBatching   bool
	BatchSize        int
}

// DefaultExecutorConfig returns default executor configuration
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		MaxMemoryBytes:   1024 * 1024 * 1024, // 1GB
		WorkMemBytes:     64 * 1024 * 1024,   // 64MB per operator
		MaxParallelism:   4,
		EnableParallel:   true,
		QueryTimeout:     30 * time.Second,
		OperatorTimeout:  10 * time.Second,
		EnablePipelining: true,
		EnableBatching:   true,
		BatchSize:        1000,
	}
}

// NewExecutor creates a new query executor
func NewExecutor(storageEngine storage.StorageEngine, bufferPool *storage.BufferPool) *Executor {
	return &Executor{
		storage:    storageEngine,
		bufferPool: bufferPool,
		statistics: NewExecutionStatistics(),
		config:     DefaultExecutorConfig(),
	}
}

// NewExecutorWithConfig creates an executor with custom configuration
func NewExecutorWithConfig(
	storageEngine storage.StorageEngine,
	bufferPool *storage.BufferPool,
	config *ExecutorConfig,
) *Executor {
	return &Executor{
		storage:    storageEngine,
		bufferPool: bufferPool,
		statistics: NewExecutionStatistics(),
		config:     config,
	}
}

// Execute executes a query plan and returns results
func (e *Executor) Execute(ctx context.Context, plan *optimizer.QueryPlan) (*ResultSet, error) {
	// Create execution context
	execCtx := NewExecutionContext(ctx, e.config)
	execCtx.SetStorage(e.storage)
	execCtx.SetBufferPool(e.bufferPool)

	// Build operator tree from physical plan
	rootOperator, err := e.buildOperatorTree(plan.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to build operator tree: %w", err)
	}

	// Execute query
	startTime := time.Now()
	defer func() {
		e.statistics.RecordQuery(time.Since(startTime))
	}()

	// Open operator tree (initialize resources)
	if err := rootOperator.Open(execCtx); err != nil {
		return nil, fmt.Errorf("failed to open operator: %w", err)
	}
	defer rootOperator.Close()

	// Pull tuples from root operator
	resultSet := NewResultSet()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get next tuple
		tuple, err := rootOperator.Next()
		if err != nil {
			return nil, fmt.Errorf("execution error: %w", err)
		}

		if tuple == nil {
			break // No more tuples (EOF)
		}

		resultSet.AddTuple(tuple)

		// Check result size limits
		if resultSet.RowCount() > e.config.BatchSize*100 {
			// Too many results, consider pagination
			break
		}
	}

	return resultSet, nil
}

// buildOperatorTree builds operator tree from physical plan
func (e *Executor) buildOperatorTree(plan *optimizer.PhysicalPlan) (PhysicalOperator, error) {
	if plan == nil {
		return nil, fmt.Errorf("cannot build operator from nil plan")
	}

	// Build children first (bottom-up)
	children := make([]PhysicalOperator, 0, len(plan.Children))
	for _, childPlan := range plan.Children {
		childOp, err := e.buildOperatorTree(childPlan)
		if err != nil {
			return nil, err
		}
		children = append(children, childOp)
	}

	// Create operator based on plan type
	switch plan.Type {
	case optimizer.PhysicalPlanTypeSeqScan:
		return NewSeqScanOperator(plan.TableName, nil), nil

	case optimizer.PhysicalPlanTypeIndexScan:
		return NewIndexScanOperator(plan.TableName, plan.IndexName, nil), nil

	case optimizer.PhysicalPlanTypeFilter:
		if len(children) != 1 {
			return nil, fmt.Errorf("filter operator requires exactly 1 child")
		}
		return NewFilterOperator(children[0], nil), nil

	case optimizer.PhysicalPlanTypeNestedLoopJoin:
		if len(children) != 2 {
			return nil, fmt.Errorf("nested loop join requires exactly 2 children")
		}
		return NewNestedLoopJoinOperator(children[0], children[1], nil, plan.JoinType), nil

	case optimizer.PhysicalPlanTypeHashJoin:
		if len(children) != 2 {
			return nil, fmt.Errorf("hash join requires exactly 2 children")
		}
		return NewHashJoinOperator(children[0], children[1], nil, plan.JoinType), nil

	case optimizer.PhysicalPlanTypeMergeJoin:
		if len(children) != 2 {
			return nil, fmt.Errorf("merge join requires exactly 2 children")
		}
		return NewMergeJoinOperator(children[0], children[1], nil, nil), nil

	case optimizer.PhysicalPlanTypeHashAggregate:
		if len(children) != 1 {
			return nil, fmt.Errorf("hash aggregate requires exactly 1 child")
		}
		return NewHashAggregateOperator(children[0], nil, nil), nil

	case optimizer.PhysicalPlanTypeSortAggregate:
		if len(children) != 1 {
			return nil, fmt.Errorf("sort aggregate requires exactly 1 child")
		}
		return NewSortAggregateOperator(children[0], nil, nil), nil

	case optimizer.PhysicalPlanTypeSort:
		if len(children) != 1 {
			return nil, fmt.Errorf("sort requires exactly 1 child")
		}
		return NewSortOperator(children[0], nil), nil

	case optimizer.PhysicalPlanTypeProject:
		if len(children) != 1 {
			return nil, fmt.Errorf("project requires exactly 1 child")
		}
		return NewProjectOperator(children[0], nil), nil

	case optimizer.PhysicalPlanTypeLimit:
		if len(children) != 1 {
			return nil, fmt.Errorf("limit requires exactly 1 child")
		}
		// TODO: Extract limit and offset from plan
		return NewLimitOperator(children[0], 0, 0), nil

	default:
		return nil, fmt.Errorf("unsupported physical plan type: %v", plan.Type)
	}
}

// ExecutionStatistics tracks execution metrics
type ExecutionStatistics struct {
	QueriesExecuted    int64
	TotalExecutionTime time.Duration
	TuplesProduced     int64
	OperatorsCreated   int64
}

// NewExecutionStatistics creates new execution statistics
func NewExecutionStatistics() *ExecutionStatistics {
	return &ExecutionStatistics{}
}

// RecordQuery records query execution
func (s *ExecutionStatistics) RecordQuery(duration time.Duration) {
	s.QueriesExecuted++
	s.TotalExecutionTime += duration
}

// RecordTuples records produced tuples
func (s *ExecutionStatistics) RecordTuples(count int64) {
	s.TuplesProduced += count
}

// RecordOperator records operator creation
func (s *ExecutionStatistics) RecordOperator() {
	s.OperatorsCreated++
}

// String returns string representation
func (s *ExecutionStatistics) String() string {
	avgTime := time.Duration(0)
	if s.QueriesExecuted > 0 {
		avgTime = s.TotalExecutionTime / time.Duration(s.QueriesExecuted)
	}

	return fmt.Sprintf("ExecutionStats{Queries: %d, TotalTime: %v, AvgTime: %v, Tuples: %d, Operators: %d}",
		s.QueriesExecuted, s.TotalExecutionTime, avgTime, s.TuplesProduced, s.OperatorsCreated)
}
