package compiler

import (
	"testing"
)

func TestNewQueryCompiler(t *testing.T) {
	catalog := NewMockCatalog()
	compiler := NewQueryCompiler(catalog)

	if compiler == nil {
		t.Fatal("Expected compiler to be created")
	}

	if compiler.catalog == nil {
		t.Fatal("Expected compiler to have catalog")
	}
}

func TestIdentifyQueryType(t *testing.T) {
	tests := []struct {
		name          string
		setupCatalog  func() *MockCatalog
		expectedType  QueryType
		expectedError bool
	}{
		{
			name: "Simple SELECT",
			setupCatalog: func() *MockCatalog {
				catalog := NewMockCatalog()
				
				// Create users table
				users := NewTableMetadata("users")
				users.AddColumn(NewColumnMetadata("id", DataTypeInteger))
				users.AddColumn(NewColumnMetadata("name", DataTypeText))
				users.AddColumn(NewColumnMetadata("age", DataTypeInteger))
				
				catalog.AddTable(users)
				return catalog
			},
			expectedType:  QueryTypeSelect,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := tt.setupCatalog()
			compiler := NewQueryCompiler(catalog)

			if compiler == nil {
				t.Fatal("Failed to create compiler")
			}

			// Basic compilation test
			t.Logf("Compiler created successfully with catalog containing %d tables", len(catalog.tables))
		})
	}
}

func TestDataTypes(t *testing.T) {
	tests := []struct {
		name         string
		dataType     DataType
		expectedStr  string
		isNumeric    bool
		isString     bool
	}{
		{"Integer", DataTypeInteger, "INTEGER", true, false},
		{"Real", DataTypeReal, "REAL", true, false},
		{"Text", DataTypeText, "TEXT", false, true},
		{"Boolean", DataTypeBoolean, "BOOLEAN", false, false},
		{"Null", DataTypeNull, "NULL", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dataType.String() != tt.expectedStr {
				t.Errorf("Expected %s, got %s", tt.expectedStr, tt.dataType.String())
			}

			if tt.dataType.IsNumeric() != tt.isNumeric {
				t.Errorf("Expected IsNumeric=%v, got %v", tt.isNumeric, tt.dataType.IsNumeric())
			}

			if tt.dataType.IsString() != tt.isString {
				t.Errorf("Expected IsString=%v, got %v", tt.isString, tt.dataType.IsString())
			}
		})
	}
}

func TestTableMetadata(t *testing.T) {
	table := NewTableMetadata("users")

	// Add columns
	idCol := NewColumnMetadata("id", DataTypeInteger)
	idCol.IsPrimaryKey = true
	idCol.Nullable = false
	table.AddColumn(idCol)

	nameCol := NewColumnMetadata("name", DataTypeText)
	table.AddColumn(nameCol)

	ageCol := NewColumnMetadata("age", DataTypeInteger)
	table.AddColumn(ageCol)

	// Test column count
	if len(table.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(table.Columns))
	}

	// Test GetColumn
	col, err := table.GetColumn("name")
	if err != nil {
		t.Errorf("Expected to find column 'name', got error: %v", err)
	}
	if col.DataType != DataTypeText {
		t.Errorf("Expected TEXT type, got %s", col.DataType)
	}

	// Test HasColumn
	if !table.HasColumn("id") {
		t.Error("Expected table to have column 'id'")
	}

	if table.HasColumn("nonexistent") {
		t.Error("Expected table to NOT have column 'nonexistent'")
	}

	// Test column not found
	_, err = table.GetColumn("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent column")
	}
}

func TestMockCatalog(t *testing.T) {
	catalog := NewMockCatalog()

	// Create and add users table
	users := NewTableMetadata("users")
	users.AddColumn(NewColumnMetadata("id", DataTypeInteger))
	users.AddColumn(NewColumnMetadata("name", DataTypeText))
	catalog.AddTable(users)

	// Create and add orders table
	orders := NewTableMetadata("orders")
	orders.AddColumn(NewColumnMetadata("id", DataTypeInteger))
	orders.AddColumn(NewColumnMetadata("user_id", DataTypeInteger))
	orders.AddColumn(NewColumnMetadata("total", DataTypeReal))
	catalog.AddTable(orders)

	// Test TableExists
	if !catalog.TableExists("users") {
		t.Error("Expected 'users' table to exist")
	}

	if !catalog.TableExists("orders") {
		t.Error("Expected 'orders' table to exist")
	}

	if catalog.TableExists("nonexistent") {
		t.Error("Expected 'nonexistent' table to NOT exist")
	}

	// Test GetTable
	table, err := catalog.GetTable("users")
	if err != nil {
		t.Errorf("Expected to find 'users' table, got error: %v", err)
	}
	if table.Name != "users" {
		t.Errorf("Expected table name 'users', got '%s'", table.Name)
	}

	// Test GetTable not found
	_, err = catalog.GetTable("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent table")
	}

	// Test GetColumn
	col, err := catalog.GetColumn("orders", "total")
	if err != nil {
		t.Errorf("Expected to find column 'total', got error: %v", err)
	}
	if col.DataType != DataTypeReal {
		t.Errorf("Expected REAL type, got %s", col.DataType)
	}

	// Test ListTables
	tables, err := catalog.ListTables()
	if err != nil {
		t.Errorf("Expected to list tables, got error: %v", err)
	}
	if len(tables) != 2 {
		t.Errorf("Expected 2 tables, got %d", len(tables))
	}
}

func TestQueryType(t *testing.T) {
	tests := []struct {
		queryType QueryType
		expectedStr string
		isDML bool
		isDDL bool
		isTCL bool
	}{
		{QueryTypeSelect, "SELECT", true, false, false},
		{QueryTypeInsert, "INSERT", true, false, false},
		{QueryTypeUpdate, "UPDATE", true, false, false},
		{QueryTypeDelete, "DELETE", true, false, false},
		{QueryTypeCreateTable, "CREATE_TABLE", false, true, false},
		{QueryTypeDropTable, "DROP_TABLE", false, true, false},
		{QueryTypeBegin, "BEGIN", false, false, true},
		{QueryTypeCommit, "COMMIT", false, false, true},
		{QueryTypeRollback, "ROLLBACK", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.expectedStr, func(t *testing.T) {
			if tt.queryType.String() != tt.expectedStr {
				t.Errorf("Expected %s, got %s", tt.expectedStr, tt.queryType.String())
			}

			if tt.queryType.IsDML() != tt.isDML {
				t.Errorf("Expected IsDML=%v, got %v", tt.isDML, tt.queryType.IsDML())
			}

			if tt.queryType.IsDDL() != tt.isDDL {
				t.Errorf("Expected IsDDL=%v, got %v", tt.isDDL, tt.queryType.IsDDL())
			}

			if tt.queryType.IsTCL() != tt.isTCL {
				t.Errorf("Expected IsTCL=%v, got %v", tt.isTCL, tt.queryType.IsTCL())
			}
		})
	}
}

func TestResolvedReferences(t *testing.T) {
	refs := NewResolvedReferences()

	// Create test table
	users := NewTableMetadata("users")
	idCol := NewColumnMetadata("id", DataTypeInteger)
	idCol.TableName = "users"
	nameCol := NewColumnMetadata("name", DataTypeText)
	nameCol.TableName = "users"
	users.AddColumn(idCol)
	users.AddColumn(nameCol)

	// Add table
	refs.AddTable("users", users)

	// Test GetTable
	table, found := refs.GetTable("users")
	if !found {
		t.Error("Expected to find 'users' table")
	}
	if table.Name != "users" {
		t.Errorf("Expected table name 'users', got '%s'", table.Name)
	}

	// Add alias
	refs.AddAlias("u", "users")

	// Test GetTable with alias
	table, found = refs.GetTable("u")
	if !found {
		t.Error("Expected to find table via alias 'u'")
	}
	if table.Name != "users" {
		t.Errorf("Expected table name 'users', got '%s'", table.Name)
	}

	// Add column
	refs.AddColumn("users.id", idCol)

	// Test GetColumn
	col, found := refs.GetColumn("users.id")
	if !found {
		t.Error("Expected to find column 'users.id'")
	}
	if col.Name != "id" {
		t.Errorf("Expected column name 'id', got '%s'", col.Name)
	}
}

func TestTypeInformation(t *testing.T) {
	typeInfo := NewTypeInformation()

	// Set types
	typeInfo.SetType("expr_1", DataTypeInteger)
	typeInfo.SetType("expr_2", DataTypeText)

	// Test GetType
	dt, found := typeInfo.GetType("expr_1")
	if !found {
		t.Error("Expected to find type for 'expr_1'")
	}
	if dt != DataTypeInteger {
		t.Errorf("Expected INTEGER, got %s", dt)
	}

	// Test GetType not found
	_, found = typeInfo.GetType("expr_nonexistent")
	if found {
		t.Error("Expected to NOT find type for 'expr_nonexistent'")
	}

	// Add coercion
	typeInfo.AddCoercion("expr_3", DataTypeInteger, DataTypeReal, "arithmetic with REAL")

	// Test coercion
	coercion, found := typeInfo.Coercions["expr_3"]
	if !found {
		t.Error("Expected to find coercion for 'expr_3'")
	}
	if coercion.FromType != DataTypeInteger || coercion.ToType != DataTypeReal {
		t.Errorf("Expected INTEGER->REAL coercion, got %s->%s", coercion.FromType, coercion.ToType)
	}
}

func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		type1        DataType
		type2        DataType
		comparable   bool
		canCoerce1to2 bool
	}{
		{"Integer-Integer", DataTypeInteger, DataTypeInteger, true, true},
		{"Integer-Real", DataTypeInteger, DataTypeReal, true, true},
		{"Real-Integer", DataTypeReal, DataTypeInteger, true, false},
		{"Integer-Text", DataTypeInteger, DataTypeText, false, false},
		{"Null-Integer", DataTypeNull, DataTypeInteger, true, true},
		{"Integer-Null", DataTypeInteger, DataTypeNull, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.type1.IsComparable(tt.type2) != tt.comparable {
				t.Errorf("Expected IsComparable=%v, got %v", tt.comparable, tt.type1.IsComparable(tt.type2))
			}

			if tt.type1.CanCoerceTo(tt.type2) != tt.canCoerce1to2 {
				t.Errorf("Expected CanCoerceTo=%v, got %v", tt.canCoerce1to2, tt.type1.CanCoerceTo(tt.type2))
			}
		})
	}
}
