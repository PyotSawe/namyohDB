package optimizer

import (
	"math"
)

// CostModel provides cost estimation for query plans
type CostModel struct {
	config *OptimizerConfig
}

// NewCostModel creates a new cost model
func NewCostModel(config *OptimizerConfig) *CostModel {
	return &CostModel{
		config: config,
	}
}

// EstimateCost estimates the total cost of a physical plan
func (cm *CostModel) EstimateCost(plan *PhysicalPlan) float64 {
	switch plan.Type {
	case PhysicalPlanTypeSeqScan:
		return cm.estimateSeqScanCost(plan)

	case PhysicalPlanTypeIndexScan:
		return cm.estimateIndexScanCost(plan)

	case PhysicalPlanTypeFilter:
		return cm.estimateFilterCost(plan)

	case PhysicalPlanTypeNestedLoopJoin:
		return cm.estimateNestedLoopJoinCost(plan)

	case PhysicalPlanTypeHashJoin:
		return cm.estimateHashJoinCost(plan)

	case PhysicalPlanTypeMergeJoin:
		return cm.estimateMergeJoinCost(plan)

	case PhysicalPlanTypeHashAggregate:
		return cm.estimateHashAggregateCost(plan)

	case PhysicalPlanTypeSortAggregate:
		return cm.estimateSortAggregateCost(plan)

	case PhysicalPlanTypeSort:
		return cm.estimateSortCost(plan)

	case PhysicalPlanTypeProject:
		return cm.estimateProjectCost(plan)

	case PhysicalPlanTypeLimit:
		return cm.estimateLimitCost(plan)

	default:
		// Unknown plan type - return high cost
		return 1000000.0
	}
}

// estimateSeqScanCost estimates cost of sequential scan
func (cm *CostModel) estimateSeqScanCost(plan *PhysicalPlan) float64 {
	// Assume table has some pages (would come from statistics)
	// TODO: Get actual page count from statistics
	pageCount := float64(plan.Cardinality) / 100.0 // Assume 100 rows per page
	if pageCount < 1.0 {
		pageCount = 1.0
	}

	// Sequential scan cost = pages * seq_page_cost
	ioCost := pageCount * cm.config.SeqPageCost

	// CPU cost = rows * cpu_tuple_cost
	cpuCost := float64(plan.Cardinality) * cm.config.CPUTupleCost

	return ioCost + cpuCost
}

// estimateIndexScanCost estimates cost of index scan
func (cm *CostModel) estimateIndexScanCost(plan *PhysicalPlan) float64 {
	// Index scan involves:
	// 1. Reading index pages (sequential)
	// 2. Reading data pages (random)

	// Assume selectivity determines how many rows we read
	selectivity := 0.1 // Default selectivity (would come from statistics)

	// Index pages to read (logarithmic in table size)
	indexPages := math.Log2(float64(plan.Cardinality))
	indexCost := indexPages * cm.config.SeqPageCost

	// Data pages to read (random access)
	dataPages := float64(plan.Cardinality) * selectivity / 100.0
	dataCost := dataPages * cm.config.RandomPageCost

	// CPU cost for index and tuple processing
	cpuCost := float64(plan.Cardinality) * selectivity * cm.config.CPUTupleCost

	return indexCost + dataCost + cpuCost
}

// estimateFilterCost estimates cost of filter operation
func (cm *CostModel) estimateFilterCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	// Cost of child + cost of filtering
	childCost := cm.EstimateCost(plan.Children[0])

	// CPU cost to evaluate filter for each input tuple
	filterCost := float64(plan.Children[0].Cardinality) * cm.config.CPUTupleCost

	return childCost + filterCost
}

// estimateNestedLoopJoinCost estimates cost of nested loop join
func (cm *CostModel) estimateNestedLoopJoinCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) < 2 {
		return 0
	}

	leftChild := plan.Children[0]
	rightChild := plan.Children[1]

	// Cost of both children
	leftCost := cm.EstimateCost(leftChild)
	rightCost := cm.EstimateCost(rightChild)

	// Nested loop: for each left tuple, scan right side
	// Cost = left_cost + (left_rows * right_cost)
	loopCost := leftCost + float64(leftChild.Cardinality)*rightCost

	// CPU cost for join condition evaluation
	joinCpuCost := float64(leftChild.Cardinality*rightChild.Cardinality) * cm.config.CPUTupleCost

	return loopCost + joinCpuCost
}

// estimateHashJoinCost estimates cost of hash join
func (cm *CostModel) estimateHashJoinCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) < 2 {
		return 0
	}

	leftChild := plan.Children[0]
	rightChild := plan.Children[1]

	// Cost of both children
	leftCost := cm.EstimateCost(leftChild)
	rightCost := cm.EstimateCost(rightChild)

	// Build hash table on left (smaller) side
	buildCost := float64(leftChild.Cardinality) * cm.config.CPUTupleCost

	// Probe hash table with right side
	probeCost := float64(rightChild.Cardinality) * cm.config.CPUTupleCost

	return leftCost + rightCost + buildCost + probeCost
}

// estimateMergeJoinCost estimates cost of merge join
func (cm *CostModel) estimateMergeJoinCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) < 2 {
		return 0
	}

	leftChild := plan.Children[0]
	rightChild := plan.Children[1]

	// Cost of both children (assuming they're already sorted)
	leftCost := cm.EstimateCost(leftChild)
	rightCost := cm.EstimateCost(rightChild)

	// If not sorted, add sort cost
	// TODO: Check if children are already sorted
	leftSortCost := cm.estimateSortCostForRows(leftChild.Cardinality)
	rightSortCost := cm.estimateSortCostForRows(rightChild.Cardinality)

	// Merge cost: linear scan of both sides
	mergeCost := float64(leftChild.Cardinality+rightChild.Cardinality) * cm.config.CPUTupleCost

	return leftCost + rightCost + leftSortCost + rightSortCost + mergeCost
}

// estimateHashAggregateCost estimates cost of hash aggregation
func (cm *CostModel) estimateHashAggregateCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	childCost := cm.EstimateCost(plan.Children[0])

	// Build hash table for grouping
	hashCost := float64(plan.Children[0].Cardinality) * cm.config.CPUTupleCost * 2.0

	return childCost + hashCost
}

// estimateSortAggregateCost estimates cost of sort-based aggregation
func (cm *CostModel) estimateSortAggregateCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	childCost := cm.EstimateCost(plan.Children[0])

	// Sort cost
	sortCost := cm.estimateSortCostForRows(plan.Children[0].Cardinality)

	// Aggregation cost (linear scan)
	aggCost := float64(plan.Children[0].Cardinality) * cm.config.CPUTupleCost

	return childCost + sortCost + aggCost
}

// estimateSortCost estimates cost of sort operation
func (cm *CostModel) estimateSortCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	childCost := cm.EstimateCost(plan.Children[0])
	sortCost := cm.estimateSortCostForRows(plan.Children[0].Cardinality)

	return childCost + sortCost
}

// estimateSortCostForRows estimates cost to sort N rows
func (cm *CostModel) estimateSortCostForRows(rows int64) float64 {
	if rows <= 1 {
		return 0
	}

	// Sort cost is O(n log n)
	n := float64(rows)
	comparisons := n * math.Log2(n)

	// CPU cost for comparisons
	return comparisons * cm.config.CPUTupleCost
}

// estimateProjectCost estimates cost of projection
func (cm *CostModel) estimateProjectCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	childCost := cm.EstimateCost(plan.Children[0])

	// CPU cost for projection (usually small)
	projectCost := float64(plan.Children[0].Cardinality) * cm.config.CPUTupleCost * 0.1

	return childCost + projectCost
}

// estimateLimitCost estimates cost of limit operation
func (cm *CostModel) estimateLimitCost(plan *PhysicalPlan) float64 {
	if len(plan.Children) == 0 {
		return 0
	}

	// Limit stops early, so cost is proportional to limit value
	// For now, use child cost (would be refined with actual limit value)
	childCost := cm.EstimateCost(plan.Children[0])

	// Limit is typically cheap
	return childCost * 0.1
}

// EstimateSelectivity estimates the selectivity of a predicate
func (cm *CostModel) EstimateSelectivity(predicate interface{}) float64 {
	// TODO: Implement selectivity estimation based on predicate type
	// For now, return default selectivity

	// Default selectivity factors
	const (
		defaultEquality = 0.005 // 0.5% for equality predicates
		defaultRange    = 0.33  // 33% for range predicates
		defaultLike     = 0.10  // 10% for LIKE predicates
	)

	return defaultEquality
}
