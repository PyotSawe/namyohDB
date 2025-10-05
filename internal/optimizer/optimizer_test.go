package optimizer
package optimizer

import (
	"testing"
	"time"

	"relational-db/internal/compiler"
)

// TestNewOptimizer tests creating an optimizer
func TestNewOptimizer(t *testing.T) {
	catalog := compiler.NewMockCatalog()
	stats := NewMockStatisticsManager()

	opt := NewOptimizer(catalog, stats)

	if opt == nil {
		t.Fatal("Expected optimizer to be created")
	}

	if opt.catalog != catalog {
		t.Error("Expected catalog to be set")
	}

	if opt.statistics != stats {
		t.Error("Expected statistics to be set")
	}

	if opt.costModel == nil {
		t.Error("Expected cost model to be initialized")
	}
}

// TestOptimizerConfig tests optimizer configuration
func TestOptimizerConfig(t *testing.T) {
	config := DefaultOptimizerConfig()

	if !config.EnablePredicatePushdown {
		t.Error("Expected predicate pushdown to be enabled by default")
	}

	if !config.EnableJoinReordering {
		t.Error("Expected join reordering to be enabled by default")
	}

	if config.SeqPageCost != 1.0 {
		t.Errorf("Expected SeqPageCost=1.0, got %.2f", config.SeqPageCost)
	}

	if config.RandomPageCost != 4.0 {
		t.Errorf("Expected RandomPageCost=4.0, got %.2f", config.RandomPageCost)
	}

	if config.CPUTupleCost != 0.01 {
		t.Errorf("Expected CPUTupleCost=0.01, got %.4f", config.CPUTupleCost)
	}
}

// TestPlanTypes tests plan type enums
func TestPlanTypes(t *testing.T) {
	tests := []struct {
		planType PlanType
		expected string
	}{
		{PlanTypeSelect, "SELECT"},
		{PlanTypeInsert, "INSERT"},
		{PlanTypeUpdate, "UPDATE"},
		{PlanTypeDelete, "DELETE"},
		{PlanTypeScan, "SCAN"},
		{PlanTypeFilter, "FILTER"},
		{PlanTypeJoin, "JOIN"},
		{PlanTypeAggregate, "AGGREGATE"},
		{PlanTypeSort, "SORT"},
		{PlanTypeProject, "PROJECT"},
		{PlanTypeLimit, "LIMIT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.planType.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestJoinTypes tests join type enums
func TestJoinTypes(t *testing.T) {
	tests := []struct {
		joinType JoinType
		expected string
	}{
		{JoinTypeInner, "INNER"},
		{JoinTypeLeft, "LEFT"},
		{JoinTypeRight, "RIGHT"},
		{JoinTypeFull, "FULL"},
		{JoinTypeCross, "CROSS"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.joinType.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestPhysicalPlanTypes tests physical plan type enums
func TestPhysicalPlanTypes(t *testing.T) {
	tests := []struct {
		planType PhysicalPlanType
		expected string
	}{
		{PhysicalPlanTypeSeqScan, "SeqScan"},
		{PhysicalPlanTypeIndexScan, "IndexScan"},
		{PhysicalPlanTypeFilter, "Filter"},
		{PhysicalPlanTypeNestedLoopJoin, "NestedLoopJoin"},
		{PhysicalPlanTypeHashJoin, "HashJoin"},
		{PhysicalPlanTypeMergeJoin, "MergeJoin"},
		{PhysicalPlanTypeHashAggregate, "HashAggregate"},
		{PhysicalPlanTypeSortAggregate, "SortAggregate"},
		{PhysicalPlanTypeSort, "Sort"},
		{PhysicalPlanTypeProject, "Project"},
		{PhysicalPlanTypeLimit, "Limit"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.planType.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestLogicalPlan tests logical plan structure
func TestLogicalPlan(t *testing.T) {
	plan := &LogicalPlan{
		Type:      PlanTypeScan,
		TableName: "users",
		Cardinality: 1000,
	}

	if plan.Type != PlanTypeScan {
		t.Errorf("Expected PlanTypeScan, got %v", plan.Type)
	}

	if plan.TableName != "users" {
		t.Errorf("Expected table name 'users', got %s", plan.TableName)
	}

	// Test string representation
	str := plan.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

// TestPhysicalPlan tests physical plan structure
func TestPhysicalPlan(t *testing.T) {
	plan := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		TableName:   "users",
		Cost:        100.0,
		Cardinality: 1000,
	}

	if plan.Type != PhysicalPlanTypeSeqScan {
		t.Errorf("Expected PhysicalPlanTypeSeqScan, got %v", plan.Type)
	}

	if plan.Cost != 100.0 {
		t.Errorf("Expected cost 100.0, got %.2f", plan.Cost)
	}

	// Test string representation
	str := plan.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

// TestCostModel tests cost model
func TestCostModel(t *testing.T) {
	config := DefaultOptimizerConfig()
	costModel := NewCostModel(config)

	if costModel == nil {
		t.Fatal("Expected cost model to be created")
	}

	// Test seq scan cost estimation
	plan := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		TableName:   "users",
		Cardinality: 1000,
	}

	cost := costModel.EstimateCost(plan)
	if cost <= 0 {
		t.Error("Expected positive cost")
	}
}

// TestSeqScanCost tests sequential scan cost estimation
func TestSeqScanCost(t *testing.T) {
	config := DefaultOptimizerConfig()
	costModel := NewCostModel(config)

	plan := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		Cardinality: 1000,
	}

	cost := costModel.estimateSeqScanCost(plan)

	// Cost should be positive
	if cost <= 0 {
		t.Error("Expected positive cost for seq scan")
	}

	// Larger table should have higher cost
	bigPlan := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		Cardinality: 10000,
	}

	bigCost := costModel.estimateSeqScanCost(bigPlan)
	if bigCost <= cost {
		t.Error("Expected higher cost for larger table")
	}
}

// TestJoinCostComparison tests join cost comparison
func TestJoinCostComparison(t *testing.T) {
	config := DefaultOptimizerConfig()
	costModel := NewCostModel(config)

	leftChild := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		Cardinality: 100,
	}

	rightChild := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		Cardinality: 100,
	}

	// Nested loop join
	nlPlan := &PhysicalPlan{
		Type:     PhysicalPlanTypeNestedLoopJoin,
		Children: []*PhysicalPlan{leftChild, rightChild},
	}

	nlCost := costModel.estimateNestedLoopJoinCost(nlPlan)

	// Hash join
	hashPlan := &PhysicalPlan{
		Type:     PhysicalPlanTypeHashJoin,
		Children: []*PhysicalPlan{leftChild, rightChild},
	}

	hashCost := costModel.estimateHashJoinCost(hashPlan)

	// For small tables, both should have reasonable costs
	if nlCost <= 0 || hashCost <= 0 {
		t.Error("Expected positive costs for joins")
	}

	// For small tables, hash join is usually better
	// (This is a heuristic test, not always true)
	t.Logf("Nested Loop Cost: %.2f, Hash Join Cost: %.2f", nlCost, hashCost)
}

// TestStatisticsManager tests statistics manager
func TestStatisticsManager(t *testing.T) {
	stats := NewMockStatisticsManager()

	// Add table statistics
	tableStats := &TableStatistics{
		TableName:    "users",
		RowCount:     1000,
		PageCount:    10,
		TotalSize:    80000,
		CreatedAt:    time.Now(),
		LastAnalyzed: time.Now(),
	}

	stats.AddTableStatistics(tableStats)

	// Retrieve statistics
	retrieved, err := stats.GetTableStatistics("users")
	if err != nil {
		t.Fatalf("Failed to get table statistics: %v", err)
	}

	if retrieved.RowCount != 1000 {
		t.Errorf("Expected RowCount=1000, got %d", retrieved.RowCount)
	}

	// Test non-existent table
	_, err = stats.GetTableStatistics("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent table")
	}
}

// TestColumnStatistics tests column statistics
func TestColumnStatistics(t *testing.T) {
	stats := NewMockStatisticsManager()

	colStats := &ColumnStatistics{
		TableName:      "users",
		ColumnName:     "age",
		DistinctValues: 50,
		NullFraction:   0.05,
		MinValue:       18,
		MaxValue:       80,
		AvgWidth:       4,
		LastAnalyzed:   time.Now(),
	}

	stats.AddColumnStatistics(colStats)

	retrieved, err := stats.GetColumnStatistics("users", "age")
	if err != nil {
		t.Fatalf("Failed to get column statistics: %v", err)
	}

	if retrieved.DistinctValues != 50 {
		t.Errorf("Expected DistinctValues=50, got %d", retrieved.DistinctValues)
	}

	if retrieved.NullFraction != 0.05 {
		t.Errorf("Expected NullFraction=0.05, got %.2f", retrieved.NullFraction)
	}
}

// TestEstimateJoinCardinality tests join cardinality estimation
func TestEstimateJoinCardinality(t *testing.T) {
	tests := []struct {
		name          string
		leftCard      int64
		rightCard     int64
		leftDistinct  int64
		rightDistinct int64
		expectRange   [2]int64 // Min and max expected
	}{
		{
			name:          "Small join",
			leftCard:      100,
			rightCard:     100,
			leftDistinct:  10,
			rightDistinct: 10,
			expectRange:   [2]int64{500, 2000},
		},
		{
			name:          "Cross join",
			leftCard:      10,
			rightCard:     10,
			leftDistinct:  0,
			rightDistinct: 0,
			expectRange:   [2]int64{100, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateJoinCardinality(
				tt.leftCard, tt.rightCard,
				tt.leftDistinct, tt.rightDistinct,
			)

			if result < tt.expectRange[0] || result > tt.expectRange[1] {
				t.Errorf("Expected result in range [%d, %d], got %d",
					tt.expectRange[0], tt.expectRange[1], result)
			}
		})
	}
}

// TestEstimateGroupByCardinality tests GROUP BY cardinality estimation
func TestEstimateGroupByCardinality(t *testing.T) {
	tests := []struct {
		name            string
		inputCard       int64
		groupByDistinct []int64
		expected        int64
	}{
		{
			name:            "No GROUP BY",
			inputCard:       1000,
			groupByDistinct: []int64{},
			expected:        1,
		},
		{
			name:            "Single column GROUP BY",
			inputCard:       1000,
			groupByDistinct: []int64{10},
			expected:        10,
		},
		{
			name:            "Multi-column GROUP BY",
			inputCard:       1000,
			groupByDistinct: []int64{10, 5},
			expected:        50,
		},
		{
			name:            "GROUP BY exceeds input",
			inputCard:       100,
			groupByDistinct: []int64{10, 20},
			expected:        100, // Bounded by input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateGroupByCardinality(tt.inputCard, tt.groupByDistinct)

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestQueryPlan tests query plan structure
func TestQueryPlan(t *testing.T) {
	physicalPlan := &PhysicalPlan{
		Type:        PhysicalPlanTypeSeqScan,
		Cost:        100.0,
		Cardinality: 1000,
	}

	queryPlan := &QueryPlan{
		Root:             physicalPlan,
		EstimatedCost:    100.0,
		EstimatedRows:    1000,
		OptimizationTime: time.Millisecond * 10,
		Statistics:       make(map[string]interface{}),
	}

	if queryPlan.Root != physicalPlan {
		t.Error("Expected root to be set")
	}

	if queryPlan.EstimatedCost != 100.0 {
		t.Errorf("Expected cost 100.0, got %.2f", queryPlan.EstimatedCost)
	}

	// Test string representation
	str := queryPlan.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}
