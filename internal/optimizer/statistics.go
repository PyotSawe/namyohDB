package optimizer

import (
	"fmt"
	"time"
)

// StatisticsManager manages table and column statistics for cost estimation
type StatisticsManager interface {
	// GetTableStatistics returns statistics for a table
	GetTableStatistics(tableName string) (*TableStatistics, error)

	// GetColumnStatistics returns statistics for a column
	GetColumnStatistics(tableName, columnName string) (*ColumnStatistics, error)

	// UpdateStatistics updates statistics for a table
	UpdateStatistics(tableName string) error

	// GetIndexStatistics returns statistics for an index
	GetIndexStatistics(tableName, indexName string) (*IndexStatistics, error)
}

// TableStatistics contains statistics about a table
type TableStatistics struct {
	TableName string
	RowCount  int64
	PageCount int64
	TotalSize int64 // In bytes

	// Timestamps
	CreatedAt    time.Time
	LastAnalyzed time.Time
	LastModified time.Time
}

// ColumnStatistics contains statistics about a column
type ColumnStatistics struct {
	TableName  string
	ColumnName string

	// Cardinality statistics
	DistinctValues int64   // Number of distinct values (NDV)
	NullFraction   float64 // Fraction of NULL values (0.0 to 1.0)

	// Value distribution
	MinValue interface{} // Minimum value
	MaxValue interface{} // Maximum value
	AvgWidth int         // Average column width in bytes

	// Histogram (optional)
	Histogram *Histogram

	// Most common values (optional)
	MostCommonValues []interface{}
	MCVFrequencies   []float64 // Frequencies of MCVs

	// Timestamps
	LastAnalyzed time.Time
}

// IndexStatistics contains statistics about an index
type IndexStatistics struct {
	TableName string
	IndexName string

	// Index structure
	IndexType string // "BTREE", "HASH", etc.
	Columns   []string
	IsUnique  bool
	IsPrimary bool

	// Size statistics
	LeafPages  int64
	TotalPages int64
	Height     int // B-tree height
	AvgKeySize int // Average key size in bytes

	// Selectivity
	Density float64 // Average selectivity (1.0 / distinct_keys)

	// Timestamps
	LastAnalyzed time.Time
}

// Histogram represents the distribution of values in a column
type Histogram struct {
	Type        HistogramType
	BucketCount int
	Buckets     []*HistogramBucket
}

// HistogramType represents the type of histogram
type HistogramType int

const (
	HistogramTypeEquiWidth HistogramType = iota // Equal width buckets
	HistogramTypeEquiDepth                      // Equal depth (height) buckets
)

func (ht HistogramType) String() string {
	switch ht {
	case HistogramTypeEquiWidth:
		return "EquiWidth"
	case HistogramTypeEquiDepth:
		return "EquiDepth"
	default:
		return "Unknown"
	}
}

// HistogramBucket represents a bucket in a histogram
type HistogramBucket struct {
	LowerBound     interface{} // Lower bound of bucket (inclusive)
	UpperBound     interface{} // Upper bound of bucket (exclusive)
	Frequency      float64     // Fraction of values in this bucket
	DistinctValues int64       // Number of distinct values in bucket
}

// MockStatisticsManager is a simple in-memory statistics manager for testing
type MockStatisticsManager struct {
	tables  map[string]*TableStatistics
	columns map[string]map[string]*ColumnStatistics
	indexes map[string]map[string]*IndexStatistics
}

// NewMockStatisticsManager creates a new mock statistics manager
func NewMockStatisticsManager() *MockStatisticsManager {
	return &MockStatisticsManager{
		tables:  make(map[string]*TableStatistics),
		columns: make(map[string]map[string]*ColumnStatistics),
		indexes: make(map[string]map[string]*IndexStatistics),
	}
}

// GetTableStatistics returns statistics for a table
func (msm *MockStatisticsManager) GetTableStatistics(tableName string) (*TableStatistics, error) {
	stats, ok := msm.tables[tableName]
	if !ok {
		return nil, fmt.Errorf("no statistics for table %s", tableName)
	}
	return stats, nil
}

// GetColumnStatistics returns statistics for a column
func (msm *MockStatisticsManager) GetColumnStatistics(tableName, columnName string) (*ColumnStatistics, error) {
	tableColumns, ok := msm.columns[tableName]
	if !ok {
		return nil, fmt.Errorf("no statistics for table %s", tableName)
	}

	stats, ok := tableColumns[columnName]
	if !ok {
		return nil, fmt.Errorf("no statistics for column %s.%s", tableName, columnName)
	}

	return stats, nil
}

// UpdateStatistics updates statistics for a table
func (msm *MockStatisticsManager) UpdateStatistics(tableName string) error {
	// Mock implementation - do nothing
	return nil
}

// GetIndexStatistics returns statistics for an index
func (msm *MockStatisticsManager) GetIndexStatistics(tableName, indexName string) (*IndexStatistics, error) {
	tableIndexes, ok := msm.indexes[tableName]
	if !ok {
		return nil, fmt.Errorf("no indexes for table %s", tableName)
	}

	stats, ok := tableIndexes[indexName]
	if !ok {
		return nil, fmt.Errorf("no index %s on table %s", indexName, tableName)
	}

	return stats, nil
}

// AddTableStatistics adds table statistics (for testing)
func (msm *MockStatisticsManager) AddTableStatistics(stats *TableStatistics) {
	msm.tables[stats.TableName] = stats
}

// AddColumnStatistics adds column statistics (for testing)
func (msm *MockStatisticsManager) AddColumnStatistics(stats *ColumnStatistics) {
	if msm.columns[stats.TableName] == nil {
		msm.columns[stats.TableName] = make(map[string]*ColumnStatistics)
	}
	msm.columns[stats.TableName][stats.ColumnName] = stats
}

// AddIndexStatistics adds index statistics (for testing)
func (msm *MockStatisticsManager) AddIndexStatistics(stats *IndexStatistics) {
	if msm.indexes[stats.TableName] == nil {
		msm.indexes[stats.TableName] = make(map[string]*IndexStatistics)
	}
	msm.indexes[stats.TableName][stats.IndexName] = stats
}

// EstimateJoinCardinality estimates the cardinality of a join
func EstimateJoinCardinality(leftCard, rightCard, leftDistinct, rightDistinct int64) int64 {
	if leftDistinct == 0 || rightDistinct == 0 {
		// Cross join
		return leftCard * rightCard
	}

	// Assumption: uniform distribution
	// Join selectivity = 1 / max(left_distinct, right_distinct)
	maxDistinct := leftDistinct
	if rightDistinct > maxDistinct {
		maxDistinct = rightDistinct
	}

	selectivity := 1.0 / float64(maxDistinct)
	result := float64(leftCard*rightCard) * selectivity

	// At least 1 row
	if result < 1.0 {
		return 1
	}

	return int64(result)
}

// EstimateGroupByCardinality estimates the cardinality after GROUP BY
func EstimateGroupByCardinality(inputCard int64, groupByDistinct []int64) int64 {
	if len(groupByDistinct) == 0 {
		// No GROUP BY - single aggregate result
		return 1
	}

	// Estimate: product of distinct values in GROUP BY columns
	// But bounded by input cardinality
	distinctProduct := int64(1)
	for _, distinct := range groupByDistinct {
		distinctProduct *= distinct
		if distinctProduct > inputCard {
			// Can't have more groups than input rows
			return inputCard
		}
	}

	return distinctProduct
}
