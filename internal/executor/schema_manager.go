// Package executor - Schema Manager component
// Manages database schemas and metadata
package executor

import (
	"fmt"
	"sync"
)

// SchemaManager manages database schemas and metadata
// Architecture: Part of Execution Engine Layer
type SchemaManager struct {
	// Schema registry
	schemas map[string]*TableSchema // table name -> schema

	// Schema versioning
	versions map[string]int // table name -> version

	// Schema constraints
	constraints map[string][]*Constraint // table name -> constraints

	mutex sync.RWMutex
}

// TableSchema represents a table's schema definition
type TableSchema struct {
	TableName   string
	Columns     []ColumnInfo
	PrimaryKey  []string // Column names
	ForeignKeys []*ForeignKey
	Indexes     []*IndexInfo
	Version     int
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name       string
	Columns    []string // Source columns
	RefTable   string   // Referenced table
	RefColumns []string // Referenced columns
	OnDelete   ReferentialAction
	OnUpdate   ReferentialAction
}

// ReferentialAction defines foreign key actions
type ReferentialAction int

const (
	NoAction ReferentialAction = iota
	Cascade
	SetNull
	SetDefault
	Restrict
)

// IndexInfo represents index metadata
type IndexInfo struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IndexType IndexType
}

// IndexType defines the type of index
type IndexType int

const (
	BTreeIndex IndexType = iota
	HashIndex
	FullTextIndex
)

// Constraint represents a table constraint
type Constraint struct {
	Name            string
	Type            ConstraintType
	Columns         []string
	CheckExpression string // For CHECK constraints
}

// ConstraintType defines constraint types
type ConstraintType int

const (
	PrimaryKeyConstraint ConstraintType = iota
	UniqueConstraint
	ForeignKeyConstraint
	CheckConstraint
	NotNullConstraint
)

// NewSchemaManager creates a new schema manager
func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		schemas:     make(map[string]*TableSchema),
		versions:    make(map[string]int),
		constraints: make(map[string][]*Constraint),
	}
}

// RegisterSchema registers a new table schema
func (sm *SchemaManager) RegisterSchema(schema *TableSchema) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.TableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Check if schema already exists
	if _, exists := sm.schemas[schema.TableName]; exists {
		return fmt.Errorf("schema for table %s already exists", schema.TableName)
	}

	// Validate schema
	if err := sm.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Register schema
	sm.schemas[schema.TableName] = schema
	sm.versions[schema.TableName] = 1

	return nil
}

// GetSchema retrieves a table schema
func (sm *SchemaManager) GetSchema(tableName string) (*TableSchema, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	schema, exists := sm.schemas[tableName]
	if !exists {
		return nil, fmt.Errorf("schema for table %s not found", tableName)
	}

	return schema, nil
}

// UpdateSchema updates an existing schema
func (sm *SchemaManager) UpdateSchema(schema *TableSchema) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Check if schema exists
	if _, exists := sm.schemas[schema.TableName]; !exists {
		return fmt.Errorf("schema for table %s does not exist", schema.TableName)
	}

	// Validate new schema
	if err := sm.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Update schema and increment version
	sm.schemas[schema.TableName] = schema
	sm.versions[schema.TableName]++
	schema.Version = sm.versions[schema.TableName]

	return nil
}

// DropSchema removes a table schema
func (sm *SchemaManager) DropSchema(tableName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.schemas[tableName]; !exists {
		return fmt.Errorf("schema for table %s does not exist", tableName)
	}

	delete(sm.schemas, tableName)
	delete(sm.versions, tableName)
	delete(sm.constraints, tableName)

	return nil
}

// ListSchemas returns all registered table names
func (sm *SchemaManager) ListSchemas() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	tables := make([]string, 0, len(sm.schemas))
	for tableName := range sm.schemas {
		tables = append(tables, tableName)
	}

	return tables
}

// GetSchemaVersion returns the schema version
func (sm *SchemaManager) GetSchemaVersion(tableName string) (int, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	version, exists := sm.versions[tableName]
	if !exists {
		return 0, fmt.Errorf("schema for table %s not found", tableName)
	}

	return version, nil
}

// AddConstraint adds a constraint to a table
func (sm *SchemaManager) AddConstraint(tableName string, constraint *Constraint) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.schemas[tableName]; !exists {
		return fmt.Errorf("schema for table %s does not exist", tableName)
	}

	if constraint == nil {
		return fmt.Errorf("constraint cannot be nil")
	}

	sm.constraints[tableName] = append(sm.constraints[tableName], constraint)
	return nil
}

// GetConstraints returns all constraints for a table
func (sm *SchemaManager) GetConstraints(tableName string) ([]*Constraint, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if _, exists := sm.schemas[tableName]; !exists {
		return nil, fmt.Errorf("schema for table %s does not exist", tableName)
	}

	return sm.constraints[tableName], nil
}

// validateSchema validates a table schema
func (sm *SchemaManager) validateSchema(schema *TableSchema) error {
	if len(schema.Columns) == 0 {
		return fmt.Errorf("schema must have at least one column")
	}

	// Check for duplicate column names
	columnNames := make(map[string]bool)
	for _, col := range schema.Columns {
		if col.Name == "" {
			return fmt.Errorf("column name cannot be empty")
		}

		if columnNames[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}

		columnNames[col.Name] = true
	}

	// Validate primary key columns exist
	for _, pkCol := range schema.PrimaryKey {
		if !columnNames[pkCol] {
			return fmt.Errorf("primary key column %s does not exist", pkCol)
		}
	}

	// Validate foreign keys
	for _, fk := range schema.ForeignKeys {
		for _, col := range fk.Columns {
			if !columnNames[col] {
				return fmt.Errorf("foreign key column %s does not exist", col)
			}
		}
	}

	// Validate indexes
	for _, idx := range schema.Indexes {
		for _, col := range idx.Columns {
			if !columnNames[col] {
				return fmt.Errorf("index column %s does not exist", col)
			}
		}
	}

	return nil
}

// ConvertToTupleSchema converts TableSchema to TupleSchema
func (sm *SchemaManager) ConvertToTupleSchema(tableSchema *TableSchema) *TupleSchema {
	return &TupleSchema{
		Columns: tableSchema.Columns,
	}
}

// GetColumnInfo retrieves information about a specific column
func (sm *SchemaManager) GetColumnInfo(tableName, columnName string) (*ColumnInfo, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	schema, exists := sm.schemas[tableName]
	if !exists {
		return nil, fmt.Errorf("schema for table %s not found", tableName)
	}

	for _, col := range schema.Columns {
		if col.Name == columnName {
			return &col, nil
		}
	}

	return nil, fmt.Errorf("column %s not found in table %s", columnName, tableName)
}
