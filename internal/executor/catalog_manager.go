// Package executor - Catalog Manager component
// Manages database catalog metadata (tables, indexes, statistics)
package executor

import (
	"fmt"
	"sync"
	"time"
)

// CatalogManager manages the system catalog
// Architecture: Part of Execution Engine Layer, works with Schema Manager
type CatalogManager struct {
	// Table catalog
	tables map[string]*TableCatalogEntry

	// Index catalog
	indexes map[string]*IndexCatalogEntry

	// Statistics
	statistics map[string]*TableStatistics

	// Schema manager dependency
	schemaManager *SchemaManager

	mutex sync.RWMutex
}

// TableCatalogEntry represents a table in the catalog
type TableCatalogEntry struct {
	TableName  string
	TableID    uint64
	SchemaName string
	Owner      string
	CreatedAt  time.Time
	ModifiedAt time.Time
	RowCount   uint64
	PageCount  uint64
	DataSize   uint64 // bytes
	IndexCount int
}

// IndexCatalogEntry represents an index in the catalog
type IndexCatalogEntry struct {
	IndexName string
	IndexID   uint64
	TableName string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
	IndexType IndexType
	CreatedAt time.Time
	PageCount uint64
	KeyCount  uint64
}

// TableStatistics contains statistics for query optimization
type TableStatistics struct {
	TableName    string
	RowCount     uint64
	PageCount    uint64
	AvgRowSize   uint64
	ColumnStats  map[string]*ColumnStatistics
	LastAnalyzed time.Time
}

// ColumnStatistics contains per-column statistics
type ColumnStatistics struct {
	ColumnName    string
	DistinctCount uint64
	NullCount     uint64
	MinValue      interface{}
	MaxValue      interface{}
	AvgSize       uint64
	Histogram     *Histogram
}

// Histogram for distribution estimation
type Histogram struct {
	Buckets     []*HistogramBucket
	BucketCount int
}

// HistogramBucket represents a histogram bucket
type HistogramBucket struct {
	LowerBound interface{}
	UpperBound interface{}
	Count      uint64
	Frequency  float64
}

// NewCatalogManager creates a new catalog manager
func NewCatalogManager(schemaManager *SchemaManager) *CatalogManager {
	return &CatalogManager{
		tables:        make(map[string]*TableCatalogEntry),
		indexes:       make(map[string]*IndexCatalogEntry),
		statistics:    make(map[string]*TableStatistics),
		schemaManager: schemaManager,
	}
}

// CreateTable registers a new table in the catalog
func (cm *CatalogManager) CreateTable(entry *TableCatalogEntry) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if entry == nil {
		return fmt.Errorf("table entry cannot be nil")
	}

	if entry.TableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Check if table already exists
	if _, exists := cm.tables[entry.TableName]; exists {
		return fmt.Errorf("table %s already exists", entry.TableName)
	}

	// Verify schema exists
	if cm.schemaManager != nil {
		if _, err := cm.schemaManager.GetSchema(entry.TableName); err != nil {
			return fmt.Errorf("schema not found: %w", err)
		}
	}

	entry.CreatedAt = time.Now()
	entry.ModifiedAt = time.Now()
	cm.tables[entry.TableName] = entry

	// Initialize statistics
	cm.statistics[entry.TableName] = &TableStatistics{
		TableName:    entry.TableName,
		ColumnStats:  make(map[string]*ColumnStatistics),
		LastAnalyzed: time.Now(),
	}

	return nil
}

// GetTable retrieves table catalog entry
func (cm *CatalogManager) GetTable(tableName string) (*TableCatalogEntry, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.tables[tableName]
	if !exists {
		return nil, fmt.Errorf("table %s not found in catalog", tableName)
	}

	return entry, nil
}

// DropTable removes a table from the catalog
func (cm *CatalogManager) DropTable(tableName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.tables[tableName]; !exists {
		return fmt.Errorf("table %s not found", tableName)
	}

	// Remove all indexes associated with the table
	for indexName, indexEntry := range cm.indexes {
		if indexEntry.TableName == tableName {
			delete(cm.indexes, indexName)
		}
	}

	delete(cm.tables, tableName)
	delete(cm.statistics, tableName)

	return nil
}

// ListTables returns all table names
func (cm *CatalogManager) ListTables() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	tables := make([]string, 0, len(cm.tables))
	for tableName := range cm.tables {
		tables = append(tables, tableName)
	}

	return tables
}

// CreateIndex registers a new index in the catalog
func (cm *CatalogManager) CreateIndex(entry *IndexCatalogEntry) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if entry == nil {
		return fmt.Errorf("index entry cannot be nil")
	}

	if entry.IndexName == "" {
		return fmt.Errorf("index name cannot be empty")
	}

	// Check if index already exists
	if _, exists := cm.indexes[entry.IndexName]; exists {
		return fmt.Errorf("index %s already exists", entry.IndexName)
	}

	// Verify table exists
	if _, exists := cm.tables[entry.TableName]; !exists {
		return fmt.Errorf("table %s not found", entry.TableName)
	}

	entry.CreatedAt = time.Now()
	cm.indexes[entry.IndexName] = entry

	// Update table index count
	if tableEntry, exists := cm.tables[entry.TableName]; exists {
		tableEntry.IndexCount++
	}

	return nil
}

// GetIndex retrieves index catalog entry
func (cm *CatalogManager) GetIndex(indexName string) (*IndexCatalogEntry, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.indexes[indexName]
	if !exists {
		return nil, fmt.Errorf("index %s not found in catalog", indexName)
	}

	return entry, nil
}

// DropIndex removes an index from the catalog
func (cm *CatalogManager) DropIndex(indexName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	entry, exists := cm.indexes[indexName]
	if !exists {
		return fmt.Errorf("index %s not found", indexName)
	}

	// Update table index count
	if tableEntry, exists := cm.tables[entry.TableName]; exists {
		tableEntry.IndexCount--
	}

	delete(cm.indexes, indexName)
	return nil
}

// ListIndexes returns all indexes for a table
func (cm *CatalogManager) ListIndexes(tableName string) []*IndexCatalogEntry {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	indexes := make([]*IndexCatalogEntry, 0)
	for _, entry := range cm.indexes {
		if entry.TableName == tableName {
			indexes = append(indexes, entry)
		}
	}

	return indexes
}

// UpdateTableStatistics updates statistics for a table
func (cm *CatalogManager) UpdateTableStatistics(stats *TableStatistics) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if stats == nil {
		return fmt.Errorf("statistics cannot be nil")
	}

	// Verify table exists
	if _, exists := cm.tables[stats.TableName]; !exists {
		return fmt.Errorf("table %s not found", stats.TableName)
	}

	stats.LastAnalyzed = time.Now()
	cm.statistics[stats.TableName] = stats

	// Update catalog entry row count
	if tableEntry, exists := cm.tables[stats.TableName]; exists {
		tableEntry.RowCount = stats.RowCount
		tableEntry.PageCount = stats.PageCount
		tableEntry.ModifiedAt = time.Now()
	}

	return nil
}

// GetTableStatistics retrieves statistics for a table
func (cm *CatalogManager) GetTableStatistics(tableName string) (*TableStatistics, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	stats, exists := cm.statistics[tableName]
	if !exists {
		return nil, fmt.Errorf("statistics for table %s not found", tableName)
	}

	return stats, nil
}

// UpdateRowCount updates the row count for a table
func (cm *CatalogManager) UpdateRowCount(tableName string, rowCount uint64) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	entry, exists := cm.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s not found", tableName)
	}

	entry.RowCount = rowCount
	entry.ModifiedAt = time.Now()

	// Update statistics
	if stats, exists := cm.statistics[tableName]; exists {
		stats.RowCount = rowCount
	}

	return nil
}

// GetCatalogInfo returns overall catalog information
func (cm *CatalogManager) GetCatalogInfo() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return map[string]interface{}{
		"total_tables":  len(cm.tables),
		"total_indexes": len(cm.indexes),
		"total_rows":    cm.getTotalRows(),
		"total_pages":   cm.getTotalPages(),
	}
}

// getTotalRows returns total row count across all tables
func (cm *CatalogManager) getTotalRows() uint64 {
	var total uint64
	for _, entry := range cm.tables {
		total += entry.RowCount
	}
	return total
}

// getTotalPages returns total page count across all tables
func (cm *CatalogManager) getTotalPages() uint64 {
	var total uint64
	for _, entry := range cm.tables {
		total += entry.PageCount
	}
	return total
}
