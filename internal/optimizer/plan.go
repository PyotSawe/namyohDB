package optimizer

import (
	"fmt"
	"strings"
)

// PlanType represents the type of logical plan node
type PlanType int

const (
	PlanTypeUnknown PlanType = iota
	PlanTypeSelect
	PlanTypeInsert
	PlanTypeUpdate
	PlanTypeDelete
	PlanTypeScan
	PlanTypeFilter
	PlanTypeJoin
	PlanTypeAggregate
	PlanTypeSort
	PlanTypeProject
	PlanTypeLimit
)

func (pt PlanType) String() string {
	switch pt {
	case PlanTypeSelect:
		return "SELECT"
	case PlanTypeInsert:
		return "INSERT"
	case PlanTypeUpdate:
		return "UPDATE"
	case PlanTypeDelete:
		return "DELETE"
	case PlanTypeScan:
		return "SCAN"
	case PlanTypeFilter:
		return "FILTER"
	case PlanTypeJoin:
		return "JOIN"
	case PlanTypeAggregate:
		return "AGGREGATE"
	case PlanTypeSort:
		return "SORT"
	case PlanTypeProject:
		return "PROJECT"
	case PlanTypeLimit:
		return "LIMIT"
	default:
		return "UNKNOWN"
	}
}

// LogicalPlan represents a logical query plan node
type LogicalPlan struct {
	Type PlanType

	// Child plans
	Children []*LogicalPlan

	// Plan-specific data
	TableName  string      // For scan nodes
	FilterExpr interface{} // For filter nodes
	JoinType   JoinType    // For join nodes
	JoinCond   interface{} // For join nodes

	// Estimated properties
	Cardinality int64
	Selectivity float64
}

// String returns a string representation of the logical plan
func (lp *LogicalPlan) String() string {
	return lp.toString(0)
}

func (lp *LogicalPlan) toString(indent int) string {
	prefix := strings.Repeat("  ", indent)
	result := fmt.Sprintf("%s%s", prefix, lp.Type)

	if lp.TableName != "" {
		result += fmt.Sprintf("(%s)", lp.TableName)
	}

	if len(lp.Children) > 0 {
		for _, child := range lp.Children {
			result += "\n" + child.toString(indent+1)
		}
	}

	return result
}

// JoinType represents the type of join operation
type JoinType int

const (
	JoinTypeInner JoinType = iota
	JoinTypeLeft
	JoinTypeRight
	JoinTypeFull
	JoinTypeCross
)

func (jt JoinType) String() string {
	switch jt {
	case JoinTypeInner:
		return "INNER"
	case JoinTypeLeft:
		return "LEFT"
	case JoinTypeRight:
		return "RIGHT"
	case JoinTypeFull:
		return "FULL"
	case JoinTypeCross:
		return "CROSS"
	default:
		return "UNKNOWN"
	}
}

// PhysicalPlanType represents the type of physical plan node
type PhysicalPlanType int

const (
	PhysicalPlanTypeUnknown PhysicalPlanType = iota
	PhysicalPlanTypeSeqScan
	PhysicalPlanTypeIndexScan
	PhysicalPlanTypeFilter
	PhysicalPlanTypeNestedLoopJoin
	PhysicalPlanTypeHashJoin
	PhysicalPlanTypeMergeJoin
	PhysicalPlanTypeHashAggregate
	PhysicalPlanTypeSortAggregate
	PhysicalPlanTypeSort
	PhysicalPlanTypeProject
	PhysicalPlanTypeLimit
)

func (ppt PhysicalPlanType) String() string {
	switch ppt {
	case PhysicalPlanTypeSeqScan:
		return "SeqScan"
	case PhysicalPlanTypeIndexScan:
		return "IndexScan"
	case PhysicalPlanTypeFilter:
		return "Filter"
	case PhysicalPlanTypeNestedLoopJoin:
		return "NestedLoopJoin"
	case PhysicalPlanTypeHashJoin:
		return "HashJoin"
	case PhysicalPlanTypeMergeJoin:
		return "MergeJoin"
	case PhysicalPlanTypeHashAggregate:
		return "HashAggregate"
	case PhysicalPlanTypeSortAggregate:
		return "SortAggregate"
	case PhysicalPlanTypeSort:
		return "Sort"
	case PhysicalPlanTypeProject:
		return "Project"
	case PhysicalPlanTypeLimit:
		return "Limit"
	default:
		return "Unknown"
	}
}

// PhysicalPlan represents a physical execution plan node
type PhysicalPlan struct {
	Type PhysicalPlanType

	// Child plans
	Children []*PhysicalPlan

	// Plan-specific data
	TableName  string      // For scan nodes
	IndexName  string      // For index scan nodes
	FilterExpr interface{} // For filter nodes
	JoinType   JoinType    // For join nodes
	JoinCond   interface{} // For join nodes

	// Cost estimates
	Cost        float64 // Total estimated cost
	StartupCost float64 // Cost before returning first row
	Cardinality int64   // Estimated number of rows

	// Runtime statistics (filled during execution)
	ActualRows  int64
	ActualTime  float64
	ActualLoops int
}

// String returns a string representation of the physical plan
func (pp *PhysicalPlan) String() string {
	return pp.toString(0)
}

func (pp *PhysicalPlan) toString(indent int) string {
	prefix := strings.Repeat("  ", indent)
	result := fmt.Sprintf("%s%s (cost=%.2f rows=%d)",
		prefix, pp.Type, pp.Cost, pp.Cardinality)

	if pp.TableName != "" {
		result += fmt.Sprintf(" on %s", pp.TableName)
	}

	if pp.IndexName != "" {
		result += fmt.Sprintf(" using %s", pp.IndexName)
	}

	if len(pp.Children) > 0 {
		for _, child := range pp.Children {
			result += "\n" + child.toString(indent+1)
		}
	}

	return result
}

// ScanMethod represents the method for scanning a table
type ScanMethod int

const (
	ScanMethodSequential ScanMethod = iota
	ScanMethodIndex
	ScanMethodBitmap
)

func (sm ScanMethod) String() string {
	switch sm {
	case ScanMethodSequential:
		return "Sequential"
	case ScanMethodIndex:
		return "Index"
	case ScanMethodBitmap:
		return "Bitmap"
	default:
		return "Unknown"
	}
}

// JoinAlgorithm represents the algorithm for join execution
type JoinAlgorithm int

const (
	JoinAlgorithmNestedLoop JoinAlgorithm = iota
	JoinAlgorithmHash
	JoinAlgorithmMerge
)

func (ja JoinAlgorithm) String() string {
	switch ja {
	case JoinAlgorithmNestedLoop:
		return "NestedLoop"
	case JoinAlgorithmHash:
		return "Hash"
	case JoinAlgorithmMerge:
		return "Merge"
	default:
		return "Unknown"
	}
}

// AggregateMethod represents the method for aggregation
type AggregateMethod int

const (
	AggregateMethodHash AggregateMethod = iota
	AggregateMethodSort
)

func (am AggregateMethod) String() string {
	switch am {
	case AggregateMethodHash:
		return "Hash"
	case AggregateMethodSort:
		return "Sort"
	default:
		return "Unknown"
	}
}
