package unit

import (
	"testing"

	"relational-db/internal/lexer"
	"relational-db/internal/parser"
)

// TestParseSimpleSelect tests basic SELECT statement parsing
func TestParseSimpleSelect(t *testing.T) {
	sql := "SELECT name FROM users"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	// Validate SELECT clause
	if selectStmt.SelectClause == nil {
		t.Fatal("Expected SelectClause, got nil")
	}

	if len(selectStmt.SelectClause.Columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(selectStmt.SelectClause.Columns))
	}

	// Validate FROM clause
	if selectStmt.FromClause == nil {
		t.Fatal("Expected FromClause, got nil")
	}

	if len(selectStmt.FromClause.Tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(selectStmt.FromClause.Tables))
	}
}

// TestParseSelectStar tests SELECT * parsing
func TestParseSelectStar(t *testing.T) {
	sql := "SELECT * FROM users"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	// Should have at least one column (the * wildcard)
	if selectStmt.SelectClause == nil || len(selectStmt.SelectClause.Columns) == 0 {
		t.Error("Expected columns in SELECT clause")
	}
}

// TestParseSelectMultipleColumns tests SELECT with multiple columns
func TestParseSelectMultipleColumns(t *testing.T) {
	sql := "SELECT name, age, email FROM users"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	if len(selectStmt.SelectClause.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(selectStmt.SelectClause.Columns))
	}
}

// TestParseSelectWithWhere tests SELECT with WHERE clause
func TestParseSelectWithWhere(t *testing.T) {
	sql := "SELECT name FROM users WHERE id = 42"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	if selectStmt.WhereClause == nil {
		t.Fatal("Expected WhereClause, got nil")
	}

	if selectStmt.WhereClause.Condition == nil {
		t.Fatal("Expected WHERE condition, got nil")
	}
}

// TestParseSelectWithComplexWhere tests SELECT with complex WHERE conditions
func TestParseSelectWithComplexWhere(t *testing.T) {
	testCases := []struct {
		name string
		sql  string
	}{
		{
			name: "AND condition",
			sql:  "SELECT name FROM users WHERE age > 18 AND status = 'active'",
		},
		{
			name: "OR condition",
			sql:  "SELECT name FROM users WHERE age < 13 OR age > 65",
		},
		{
			name: "Multiple conditions",
			sql:  "SELECT name FROM users WHERE age > 18 AND status = 'active' AND verified = true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.NewLexer(tc.sql)
			p := parser.NewParser(l)

			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
			}

			selectStmt, ok := stmt.(*parser.SelectStatement)
			if !ok {
				t.Fatalf("Expected *SelectStatement, got %T", stmt)
			}

			if selectStmt.WhereClause == nil {
				t.Fatal("Expected WhereClause, got nil")
			}
		})
	}
}

// TestParseSelectWithOrderBy tests SELECT with ORDER BY
func TestParseSelectWithOrderBy(t *testing.T) {
	sql := "SELECT name, age FROM users ORDER BY age DESC"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	if selectStmt.OrderBy == nil {
		t.Fatal("Expected OrderBy clause, got nil")
	}

	if len(selectStmt.OrderBy.Orders) != 1 {
		t.Errorf("Expected 1 order expression, got %d", len(selectStmt.OrderBy.Orders))
	}
}

// TestParseSelectWithLimit tests SELECT with LIMIT
func TestParseSelectWithLimit(t *testing.T) {
	sql := "SELECT name FROM users LIMIT 10"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	if selectStmt.Limit == nil {
		t.Fatal("Expected Limit clause, got nil")
	}
}

// TestParseInsertSimple tests basic INSERT statement
func TestParseInsertSimple(t *testing.T) {
	sql := "INSERT INTO users (name, age) VALUES ('Alice', 25)"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	insertStmt, ok := stmt.(*parser.InsertStatement)
	if !ok {
		t.Fatalf("Expected *InsertStatement, got %T", stmt)
	}

	if insertStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}

	if len(insertStmt.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(insertStmt.Columns))
	}

	if len(insertStmt.Values) == 0 {
		t.Fatal("Expected values, got none")
	}
}

// TestParseInsertWithoutColumns tests INSERT without column list
func TestParseInsertWithoutColumns(t *testing.T) {
	sql := "INSERT INTO users VALUES (1, 'Bob', 30)"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	insertStmt, ok := stmt.(*parser.InsertStatement)
	if !ok {
		t.Fatalf("Expected *InsertStatement, got %T", stmt)
	}

	if insertStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}
}

// TestParseCreateTable tests CREATE TABLE statement
func TestParseCreateTable(t *testing.T) {
	sql := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		age INTEGER
	)`
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	createStmt, ok := stmt.(*parser.CreateTableStatement)
	if !ok {
		t.Fatalf("Expected *CreateTableStatement, got %T", stmt)
	}

	if createStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}

	if len(createStmt.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(createStmt.Columns))
	}
}

// TestParseCreateTableWithConstraints tests CREATE TABLE with various constraints
func TestParseCreateTableWithConstraints(t *testing.T) {
	sql := `CREATE TABLE products (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		price REAL DEFAULT 0.0,
		quantity INTEGER
	)`
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	createStmt, ok := stmt.(*parser.CreateTableStatement)
	if !ok {
		t.Fatalf("Expected *CreateTableStatement, got %T", stmt)
	}

	if len(createStmt.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(createStmt.Columns))
	}
}

// TestParseDelete tests DELETE statement
func TestParseDelete(t *testing.T) {
	sql := "DELETE FROM users WHERE id = 42"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	deleteStmt, ok := stmt.(*parser.DeleteStatement)
	if !ok {
		t.Fatalf("Expected *DeleteStatement, got %T", stmt)
	}

	if deleteStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}

	if deleteStmt.WhereClause == nil {
		t.Fatal("Expected WHERE clause, got nil")
	}
}

// TestParseUpdate tests UPDATE statement
func TestParseUpdate(t *testing.T) {
	sql := "UPDATE users SET age = 26 WHERE name = 'Alice'"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	updateStmt, ok := stmt.(*parser.UpdateStatement)
	if !ok {
		t.Fatalf("Expected *UpdateStatement, got %T", stmt)
	}

	if updateStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}

	if len(updateStmt.SetClauses) == 0 {
		t.Fatal("Expected assignments, got none")
	}
}

// TestParseDropTable tests DROP TABLE statement
func TestParseDropTable(t *testing.T) {
	sql := "DROP TABLE users"
	l := lexer.NewLexer(sql)
	p := parser.NewParser(l)

	stmt := p.ParseStatement()

	if stmt == nil {
		t.Fatalf("Expected statement, got nil. Errors: %v", p.Errors())
	}

	dropStmt, ok := stmt.(*parser.DropTableStatement)
	if !ok {
		t.Fatalf("Expected *DropTableStatement, got %T", stmt)
	}

	if dropStmt.TableName == nil {
		t.Fatal("Expected table name, got nil")
	}
}

// TestParserErrors tests error handling for invalid SQL
func TestParserErrors(t *testing.T) {
	testCases := []struct {
		name string
		sql  string
	}{
		{
			name: "Missing FROM",
			sql:  "SELECT name",
		},
		{
			name: "Missing table name",
			sql:  "SELECT name FROM",
		},
		{
			name: "Invalid WHERE",
			sql:  "SELECT name FROM users WHERE",
		},
		{
			name: "Missing VALUES",
			sql:  "INSERT INTO users (name)",
		},
		{
			name: "Invalid CREATE TABLE",
			sql:  "CREATE TABLE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.NewLexer(tc.sql)
			p := parser.NewParser(l)

			stmt := p.ParseStatement()

			// Should either return nil or have errors
			if stmt == nil || len(p.Errors()) > 0 {
				// Expected: parser caught the error
				t.Logf("Parser correctly detected error: %v", p.Errors())
			} else {
				t.Error("Parser should have detected an error but didn't")
			}
		})
	}
}

// TestParseSQLConvenienceFunction tests the ParseSQL helper function
func TestParseSQLConvenienceFunction(t *testing.T) {
	sql := "SELECT name FROM users WHERE id = 42"

	stmt, err := parser.ParseSQL(sql)

	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if stmt == nil {
		t.Fatal("Expected statement, got nil")
	}

	selectStmt, ok := stmt.(*parser.SelectStatement)
	if !ok {
		t.Fatalf("Expected *SelectStatement, got %T", stmt)
	}

	if selectStmt.FromClause == nil {
		t.Fatal("Expected FROM clause, got nil")
	}
}
