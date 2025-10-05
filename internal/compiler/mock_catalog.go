package compiler

import (
	"fmt"
	"strings"
)

// MockCatalog is a simple in-memory catalog for testing
type MockCatalog struct {
	tables map[string]*TableMetadata
}

// NewMockCatalog creates a new mock catalog
func NewMockCatalog() *MockCatalog {
	return &MockCatalog{
		tables: make(map[string]*TableMetadata),
	}
}

// AddTable adds a table to the mock catalog
func (mc *MockCatalog) AddTable(table *TableMetadata) {
	mc.tables[strings.ToLower(table.Name)] = table
}

// GetTable retrieves table metadata by name
func (mc *MockCatalog) GetTable(name string) (*TableMetadata, error) {
	table, found := mc.tables[strings.ToLower(name)]
	if !found {
		return nil, fmt.Errorf("table not found: %s", name)
	}
	return table, nil
}

// GetColumn retrieves column metadata
func (mc *MockCatalog) GetColumn(table, column string) (*ColumnMetadata, error) {
	t, err := mc.GetTable(table)
	if err != nil {
		return nil, err
	}
	return t.GetColumn(column)
}

// TableExists checks if a table exists
func (mc *MockCatalog) TableExists(name string) bool {
	_, found := mc.tables[strings.ToLower(name)]
	return found
}

// ListTables returns all table names
func (mc *MockCatalog) ListTables() ([]string, error) {
	names := make([]string, 0, len(mc.tables))
	for name := range mc.tables {
		names = append(names, name)
	}
	return names, nil
}
