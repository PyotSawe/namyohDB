package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"relational-db/internal/config"
	"relational-db/internal/dispatcher"
	"relational-db/internal/lexer"
	"relational-db/internal/parser"
	"relational-db/internal/storage"
)

// demonstrateSQLProcessing showcases the new modular SQL processing architecture
func demonstrateSQLProcessing(cfg *config.Config, storageEngine storage.StorageEngine) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 SQLite3-Style Modular Architecture Demonstration")
	fmt.Println(strings.Repeat("=", 60))

	// Create SQL query dispatcher
	sqlDispatcher := dispatcher.NewDispatcher(cfg, storageEngine)

	// Test SQL queries to demonstrate different components
	testQueries := []string{
		"SELECT * FROM users WHERE age > 18",
		"INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')",
		"UPDATE users SET email = 'newemail@example.com' WHERE id = 1",
		"DELETE FROM users WHERE age < 18",
		"CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT NOT NULL, price REAL)",
		"DROP TABLE temp_table",
	}

	fmt.Println("\n📊 Component Demonstrations:")
	
	for i, sql := range testQueries {
		fmt.Printf("\n--- Test Query %d ---\n", i+1)
		fmt.Printf("SQL: %s\n", sql)
		
		// 1. Demonstrate Lexical Analysis
		demonstrateLexer(sql)
		
		// 2. Demonstrate Parser and AST
		demonstrateParser(sql)
		
		// 3. Demonstrate Query Dispatcher
		demonstrateDispatcher(sqlDispatcher, sql)
		
		fmt.Println()
	}
	
	// Show dispatcher statistics
	fmt.Println("\n📈 Query Processing Statistics:")
	stats := sqlDispatcher.GetStats()
	fmt.Println(stats.String())
}

func demonstrateLexer(sql string) {
	fmt.Println("\n🔍 Lexical Analysis:")
	
	tokens, err := lexer.TokenizeSQL(sql)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	
	fmt.Printf("   📝 Tokens (%d): ", len(tokens)-1) // -1 to exclude EOF
	for i, token := range tokens {
		if token.Type == lexer.EOF {
			break
		}
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%s", token.Type.String())
	}
	fmt.Println()
}

func demonstrateParser(sql string) {
	fmt.Println("\n🌳 Syntax Analysis (AST):")
	
	stmt, err := parser.ParseSQL(sql)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	
	fmt.Printf("   🎯 Statement Type: %s\n", stmt.NodeType())
	fmt.Printf("   📋 AST: %s\n", stmt.String())
}

func demonstrateDispatcher(sqlDispatcher *dispatcher.Dispatcher, sql string) {
	fmt.Println("\n⚡ Query Dispatch & Execution:")
	
	// Validate query
	if err := sqlDispatcher.ValidateQuery(sql); err != nil {
		fmt.Printf("   ❌ Validation Error: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Validation: Passed\n")
	
	// Get execution plan
	ctx := context.Background()
	plan, err := sqlDispatcher.ExplainQuery(ctx, sql)
	if err != nil {
		fmt.Printf("   ❌ Planning Error: %v\n", err)
		return
	}
	
	fmt.Printf("   🎯 Query Type: %s\n", plan.QueryType.String())
	fmt.Printf("   💰 Estimated Cost: %.1f\n", plan.EstimatedCost)
	fmt.Printf("   📋 Operations: ")
	for i, op := range plan.Operations {
		if i > 0 {
			fmt.Print(" → ")
		}
		fmt.Printf("%s", op.Type.String())
	}
	fmt.Println()
	
	// Execute query
	queryCtx := &dispatcher.QueryContext{
		ConnectionID: "demo-connection",
		UserID:       "demo-user",
		DatabaseName: "demo-db",
		StartTime:    time.Now(),
		Timeout:      30 * time.Second,
	}
	
	result, err := sqlDispatcher.DispatchQuery(ctx, sql, queryCtx)
	if err != nil {
		fmt.Printf("   ❌ Execution Error: %v\n", err)
		return
	}
	
	if result.Error != nil {
		fmt.Printf("   ❌ Query Error: %v\n", result.Error)
		return
	}
	
	fmt.Printf("   ⏱️  Execution Time: %v\n", result.ExecutionTime)
	fmt.Printf("   📊 Rows Affected: %d\n", result.RowsAffected)
	if len(result.Rows) > 0 {
		fmt.Printf("   📄 Sample Results: %d rows returned\n", len(result.Rows))
		for i, row := range result.Rows {
			if i >= 2 { // Show only first 2 rows
				fmt.Printf("   ... (%d more rows)\n", len(result.Rows)-2)
				break
			}
			fmt.Printf("      Row %d: %v\n", i+1, row)
		}
	}
}

// demonstrateArchitecture shows the complete architecture flow
func demonstrateArchitecture() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🏗️  Database Architecture Overview")
	fmt.Println(strings.Repeat("=", 60))
	
	architectureDiagram := `
┌─────────────────────────────────────────────────────────┐
│                 Client Application                      │
│         (Mobile, Desktop, or Embedded App)              │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│            SQLite3 API Interface                       │
│       (pkg/sqlite3 - Database Client API)              │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│              SQL Query Dispatcher                      │
│    (internal/dispatcher - Routes to Subsystems)        │
└─────┬──────────────────────────────────────────┬──────┘
      │                                          │
      ▼                                          ▼
┌──────────────────┐                    ┌──────────────────┐
│   SQL Parser     │                    │ Database Engine  │
│ (internal/lexer  │                    │ (Query Execution │
│  internal/parser)│                    │     Engine)      │
└─────┬────────────┘                    └─────┬────────────┘
      │                                       │
      ▼                                       ▼
┌──────────────────┐              ┌─────────────────────────┐
│ Query Optimizer  │              │  Query Execution       │
│ (Execution Plan  │              │  Engine (Scan, Join,   │
│  Generation)     │              │  Select, etc.)         │
└──────────────────┘              └──────┬──────────────────┘
                                         │
                                         ▼
┌────────────────────────────┬──────────────────────────────┐
│   Transaction Manager      │   Locking & Concurrency     │
│ (ACID, Rollback, WAL)      │   Control (Row/Table Locks)  │
└────────────┬───────────────┴──────────────────────────────┘
             │
             ▼
┌─────────────────────────────┬──────────────────────────────┐
│  Database File I/O Layer    │     Buffer Management        │
│ (internal/fileio - Pages,   │  (internal/buffer - Page     │
│  Journals, WAL)             │   Caching in Memory)         │
└─────────────┬───────────────┴──────────────────────────────┘
              │
              ▼
┌──────────────────────────────┬─────────────────────────────┐
│   File-based Storage System  │    B+ Tree Indexing        │
│ (internal/storage - Single   │  (internal/btree - Index   │
│  File DB, WAL)               │   Storage & Management)     │
└──────────────────────────────┴─────────────────────────────┘
`
	
	fmt.Print(architectureDiagram)
	
	fmt.Println("\n📊 Implementation Status:")
	fmt.Println("✅ Lexical Analyzer - Complete")
	fmt.Println("✅ SQL Parser & AST - Complete") 
	fmt.Println("✅ Query Dispatcher - Complete")
	fmt.Println("✅ Storage Engine - Complete")
	fmt.Println("✅ Buffer Pool - Complete")
	fmt.Println("✅ File Manager - Complete")
	fmt.Println("🔄 Query Optimizer - Basic implementation")
	fmt.Println("🔄 Execution Engine - Placeholder implementation")
	fmt.Println("📋 Transaction Manager - Planned")
	fmt.Println("📋 Locking System - Planned")
	fmt.Println("📋 WAL System - Planned")
	fmt.Println("📋 B+ Tree Indexes - Planned")
	fmt.Println("📋 Network Protocol - Planned")
	
	fmt.Println("\n🏆 Architecture Benefits:")
	fmt.Println("• Modular design with clear separation of concerns")
	fmt.Println("• SQLite3-inspired professional architecture")
	fmt.Println("• Extensible components for future enhancements")
	fmt.Println("• Production-ready foundation")
	fmt.Println("• Comprehensive error handling and validation")
	fmt.Println("• Performance monitoring and statistics")
}