package executor

import (
	"context"
	"testing"
	"time"
)

// TestNewExecutor tests creating an executor
func TestNewExecutor(t *testing.T) {
	executor := NewExecutor(nil, nil)

	if executor == nil {
		t.Fatal("Expected executor to be created")
	}

	if executor.config == nil {
		t.Error("Expected config to be set")
	}

	if executor.statistics == nil {
		t.Error("Expected statistics to be initialized")
	}
}

// TestExecutorConfig tests executor configuration
func TestExecutorConfig(t *testing.T) {
	config := DefaultExecutorConfig()

	if config.MaxMemoryBytes <= 0 {
		t.Error("Expected positive max memory")
	}

	if config.WorkMemBytes <= 0 {
		t.Error("Expected positive work memory")
	}

	if config.MaxParallelism <= 0 {
		t.Error("Expected positive parallelism")
	}

	if !config.EnablePipelining {
		t.Error("Expected pipelining to be enabled by default")
	}
}

// TestExecutionContext tests execution context
func TestExecutionContext(t *testing.T) {
	ctx := context.Background()
	config := DefaultExecutorConfig()
	execCtx := NewExecutionContext(ctx, config)

	if execCtx == nil {
		t.Fatal("Expected execution context to be created")
	}

	if execCtx.Context() != ctx {
		t.Error("Expected context to match")
	}

	if execCtx.IsTimedOut() {
		t.Error("Should not be timed out immediately")
	}
}

// TestExecutionContextMemory tests memory management
func TestExecutionContextMemory(t *testing.T) {
	ctx := context.Background()
	config := DefaultExecutorConfig()
	config.MaxMemoryBytes = 1024
	execCtx := NewExecutionContext(ctx, config)

	// Test allocation
	err := execCtx.AllocateMemory(512)
	if err != nil {
		t.Errorf("Failed to allocate memory: %v", err)
	}

	if execCtx.GetMemoryUsed() != 512 {
		t.Errorf("Expected 512 bytes used, got %d", execCtx.GetMemoryUsed())
	}

	// Test over-allocation
	err = execCtx.AllocateMemory(1024)
	if err == nil {
		t.Error("Expected error when exceeding memory limit")
	}

	// Test release
	execCtx.ReleaseMemory(512)
	if execCtx.GetMemoryUsed() != 0 {
		t.Errorf("Expected 0 bytes used after release, got %d", execCtx.GetMemoryUsed())
	}
}

// TestExecutionStatistics tests execution statistics
func TestExecutionStatistics(t *testing.T) {
	stats := NewExecutionStatistics()

	if stats.QueriesExecuted != 0 {
		t.Error("Expected 0 queries executed initially")
	}

	stats.RecordQuery(100 * time.Millisecond)
	if stats.QueriesExecuted != 1 {
		t.Error("Expected 1 query recorded")
	}

	if stats.TotalExecutionTime != 100*time.Millisecond {
		t.Errorf("Expected 100ms total time, got %v", stats.TotalExecutionTime)
	}

	stats.RecordTuples(1000)
	if stats.TuplesProduced != 1000 {
		t.Errorf("Expected 1000 tuples, got %d", stats.TuplesProduced)
	}

	str := stats.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

// TestTupleSchema tests tuple schema
func TestTupleSchema(t *testing.T) {
	columns := []ColumnInfo{
		{Name: "id", Type: TypeInt},
		{Name: "name", Type: TypeString},
		{Name: "age", Type: TypeInt},
	}

	schema := NewTupleSchema(columns)

	if schema.ColumnCount() != 3 {
		t.Errorf("Expected 3 columns, got %d", schema.ColumnCount())
	}

	idx := schema.GetColumnIndex("name")
	if idx != 1 {
		t.Errorf("Expected index 1 for 'name', got %d", idx)
	}

	col, ok := schema.GetColumn("age")
	if !ok {
		t.Error("Expected to find 'age' column")
	}
	if col.Type != TypeInt {
		t.Errorf("Expected INT type, got %v", col.Type)
	}
}

// TestTuple tests tuple operations
func TestTuple(t *testing.T) {
	columns := []ColumnInfo{
		{Name: "id", Type: TypeInt},
		{Name: "name", Type: TypeString},
	}

	schema := NewTupleSchema(columns)
	values := []interface{}{1, "Alice"}
	tuple := NewTuple(schema, values)

	// Test GetColumn
	id, err := tuple.GetColumn("id")
	if err != nil {
		t.Errorf("Failed to get id column: %v", err)
	}
	if id != 1 {
		t.Errorf("Expected id=1, got %v", id)
	}

	name, err := tuple.GetColumn("name")
	if err != nil {
		t.Errorf("Failed to get name column: %v", err)
	}
	if name != "Alice" {
		t.Errorf("Expected name='Alice', got %v", name)
	}

	// Test GetColumnByIndex
	val, err := tuple.GetColumnByIndex(0)
	if err != nil {
		t.Errorf("Failed to get column by index: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	// Test SetColumn
	err = tuple.SetColumn("id", 2)
	if err != nil {
		t.Errorf("Failed to set column: %v", err)
	}

	id, _ = tuple.GetColumn("id")
	if id != 2 {
		t.Errorf("Expected id=2 after update, got %v", id)
	}

	// Test Clone
	cloned := tuple.Clone()
	if cloned.Values[0] != tuple.Values[0] {
		t.Error("Cloned tuple should have same values")
	}

	// Modify clone shouldn't affect original
	cloned.Values[0] = 99
	if tuple.Values[0] == 99 {
		t.Error("Modifying clone should not affect original")
	}
}

// TestResultSet tests result set operations
func TestResultSet(t *testing.T) {
	rs := NewResultSet()

	if rs.RowCount() != 0 {
		t.Error("Expected 0 rows initially")
	}

	columns := []ColumnInfo{{Name: "id", Type: TypeInt}}
	schema := NewTupleSchema(columns)

	tuple1 := NewTuple(schema, []interface{}{1})
	tuple2 := NewTuple(schema, []interface{}{2})

	rs.AddTuple(tuple1)
	rs.AddTuple(tuple2)

	if rs.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", rs.RowCount())
	}

	tuple, err := rs.GetTuple(0)
	if err != nil {
		t.Errorf("Failed to get tuple: %v", err)
	}
	if tuple.Values[0] != 1 {
		t.Errorf("Expected id=1, got %v", tuple.Values[0])
	}
}

// TestColumnType tests column type string conversion
func TestColumnType(t *testing.T) {
	tests := []struct {
		colType  ColumnType
		expected string
	}{
		{TypeInt, "INT"},
		{TypeBigInt, "BIGINT"},
		{TypeFloat, "FLOAT"},
		{TypeDouble, "DOUBLE"},
		{TypeString, "STRING"},
		{TypeBoolean, "BOOLEAN"},
		{TypeDate, "DATE"},
		{TypeTimestamp, "TIMESTAMP"},
		{TypeNull, "NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.colType.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestFilterOperator tests filter operator
func TestFilterOperator(t *testing.T) {
	// Create a simple scan operator (stub)
	scan := NewSeqScanOperator("users", nil)

	// Create filter operator
	filter := NewFilterOperator(scan, nil)

	if filter.OperatorType() != "Filter" {
		t.Errorf("Expected operator type 'Filter', got %s", filter.OperatorType())
	}

	// Test that it starts closed
	_, err := filter.Next()
	if err != ErrOperatorClosed {
		t.Error("Expected ErrOperatorClosed for unopened operator")
	}
}

// TestProjectOperator tests project operator
func TestProjectOperator(t *testing.T) {
	scan := NewSeqScanOperator("users", nil)
	project := NewProjectOperator(scan, nil)

	if project.OperatorType() != "Project" {
		t.Errorf("Expected operator type 'Project', got %s", project.OperatorType())
	}
}

// TestLimitOperator tests limit operator
func TestLimitOperator(t *testing.T) {
	scan := NewSeqScanOperator("users", nil)
	limit := NewLimitOperator(scan, 10, 5)

	if limit.OperatorType() != "Limit" {
		t.Errorf("Expected operator type 'Limit', got %s", limit.OperatorType())
	}

	if limit.limit != 10 {
		t.Errorf("Expected limit=10, got %d", limit.limit)
	}

	if limit.offset != 5 {
		t.Errorf("Expected offset=5, got %d", limit.offset)
	}
}

// TestJoinOperators tests join operator creation
func TestJoinOperators(t *testing.T) {
	left := NewSeqScanOperator("users", nil)
	right := NewSeqScanOperator("orders", nil)

	// Test nested loop join
	nlj := NewNestedLoopJoinOperator(left, right, nil, 0)
	if nlj.OperatorType() != "NestedLoopJoin" {
		t.Errorf("Expected 'NestedLoopJoin', got %s", nlj.OperatorType())
	}

	// Test hash join
	hj := NewHashJoinOperator(left, right, nil, 0)
	if hj.OperatorType() != "HashJoin" {
		t.Errorf("Expected 'HashJoin', got %s", hj.OperatorType())
	}

	// Test merge join
	mj := NewMergeJoinOperator(left, right, nil, nil)
	if mj.OperatorType() != "MergeJoin" {
		t.Errorf("Expected 'MergeJoin', got %s", mj.OperatorType())
	}
}

// TestAggregateOperators tests aggregate operator creation
func TestAggregateOperators(t *testing.T) {
	scan := NewSeqScanOperator("sales", nil)

	// Test hash aggregate
	hashAgg := NewHashAggregateOperator(scan, nil, nil)
	if hashAgg.OperatorType() != "HashAggregate" {
		t.Errorf("Expected 'HashAggregate', got %s", hashAgg.OperatorType())
	}

	// Test sort aggregate
	sortAgg := NewSortAggregateOperator(scan, nil, nil)
	if sortAgg.OperatorType() != "SortAggregate" {
		t.Errorf("Expected 'SortAggregate', got %s", sortAgg.OperatorType())
	}
}

// TestSortOperator tests sort operator
func TestSortOperator(t *testing.T) {
	scan := NewSeqScanOperator("users", nil)
	sort := NewSortOperator(scan, nil)

	if sort.OperatorType() != "Sort" {
		t.Errorf("Expected operator type 'Sort', got %s", sort.OperatorType())
	}
}

// TestAggregateState tests aggregate state
func TestAggregateState(t *testing.T) {
	state := NewAggregateState()

	state.Update(10)
	state.Update(20)
	state.Update(30)

	if state.Count != 3 {
		t.Errorf("Expected count=3, got %d", state.Count)
	}

	count := state.Finalize("COUNT")
	if count != int64(3) {
		t.Errorf("Expected COUNT=3, got %v", count)
	}
}
