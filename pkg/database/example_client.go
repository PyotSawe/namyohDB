package database

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"relational-db/internal/config"
	"relational-db/internal/storage"
)

// ExampleClient demonstrates how to use the database API
func ExampleClient() {
	fmt.Println("Database API Example")
	fmt.Println("====================")
	
	// Create configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           5432,
			MaxConnections: 10,
		},
		Database: config.DatabaseConfig{
			Name:            "example_db",
			MaxTransactions: 100,
			QueryTimeout:    30,
		},
		Storage: config.StorageConfig{
			DataDirectory: "./example_data",
			PageSize:      4096,
			BufferSize:    100,
			MaxFileSize:   10 * 1024 * 1024,
		},
	}
	
	// Create storage engine
	storageEngine, err := storage.NewEngine(&cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to create storage engine: %v", err)
	}
	defer storageEngine.Close()
	
	// Create database instance
	db, err := NewDatabase(cfg, storageEngine)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	fmt.Printf("Database created successfully\n")
	
	// Demonstrate connection management
	demonstrateConnections(db)
	
	// Demonstrate transactions
	demonstrateTransactions(db)
	
	// Show database statistics
	showDatabaseStats(db)
	
	// Show health status
	showHealthStatus(db)
	
	fmt.Println("Example completed successfully!")
}

func demonstrateConnections(db Database) {
	fmt.Println("\n--- Connection Management ---")
	
	ctx := context.Background()
	
	// Create multiple connections
	connections := make([]Connection, 3)
	for i := 0; i < 3; i++ {
		conn, err := db.Connect(ctx)
		if err != nil {
			log.Printf("Failed to create connection %d: %v", i, err)
			continue
		}
		connections[i] = conn
		
		// Test ping
		if err := conn.Ping(); err != nil {
			log.Printf("Failed to ping connection %d: %v", i, err)
		} else {
			fmt.Printf("Connection %d: OK\n", i)
		}
	}
	
	// Test query execution (will fail since not implemented yet)
	if len(connections) > 0 && connections[0] != nil {
		_, err := connections[0].Execute("SELECT 1")
		fmt.Printf("Query execution test: %v (expected - not implemented yet)\n", err)
	}
	
	// Close connections
	for i, conn := range connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Printf("Failed to close connection %d: %v", i, err)
			} else {
				fmt.Printf("Connection %d closed\n", i)
			}
		}
	}
}

func demonstrateTransactions(db Database) {
	fmt.Println("\n--- Transaction Management ---")
	
	ctx := context.Background()
	
	// Create connection
	conn, err := db.Connect(ctx)
	if err != nil {
		log.Printf("Failed to create connection: %v", err)
		return
	}
	defer conn.Close()
	
	// Begin transaction
	tx, err := conn.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	
	fmt.Printf("Transaction started, status: %s\n", tx.Status().String())
	
	// Test transactional query (will fail since not implemented)
	_, err = tx.Execute("INSERT INTO test_table (id, name) VALUES (1, 'test')")
	fmt.Printf("Transactional query test: %v (expected - not implemented yet)\n", err)
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	} else {
		fmt.Printf("Transaction committed, final status: %s\n", tx.Status().String())
	}
	
	// Demonstrate rollback
	tx2, err := conn.Begin()
	if err != nil {
		log.Printf("Failed to begin second transaction: %v", err)
		return
	}
	
	fmt.Printf("Second transaction started, status: %s\n", tx2.Status().String())
	
	// Rollback transaction
	if err := tx2.Rollback(); err != nil {
		log.Printf("Failed to rollback transaction: %v", err)
	} else {
		fmt.Printf("Transaction rolled back, final status: %s\n", tx2.Status().String())
	}
}

func showDatabaseStats(db Database) {
	fmt.Println("\n--- Database Statistics ---")
	
	stats := db.Stats()
	
	fmt.Printf("Connections:\n")
	fmt.Printf("  Active: %d\n", stats.ConnectionsActive)
	fmt.Printf("  Total: %d\n", stats.ConnectionsTotal)
	
	fmt.Printf("Queries:\n")
	fmt.Printf("  Executed: %d\n", stats.QueriesExecuted)
	
	fmt.Printf("Transactions:\n")
	fmt.Printf("  Active: %d\n", stats.TransactionsActive)
	fmt.Printf("  Total: %d\n", stats.TransactionsTotal)
	
	fmt.Printf("Storage:\n")
	fmt.Printf("  %s\n", stats.StorageStats.String())
	
	fmt.Printf("Uptime: %v\n", stats.Uptime)
}

func showHealthStatus(db Database) {
	fmt.Println("\n--- Health Status ---")
	
	health := db.Health()
	
	fmt.Printf("Status: %s\n", health.Status)
	fmt.Printf("Uptime: %v\n", health.Uptime)
	fmt.Printf("Last Check: %s\n", health.LastCheck.Format(time.RFC3339))
	
	fmt.Printf("Details:\n")
	for key, value := range health.Details {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// CreateExampleSchema demonstrates how to define table schemas
func CreateExampleSchema() TableSchema {
	return TableSchema{
		Name: "users",
		Columns: []ColumnSchema{
			{
				Name:       "id",
				Type:       TypeBigInt,
				PrimaryKey: true,
				AutoIncrement: true,
				Nullable:   false,
			},
			{
				Name:     "username",
				Type:     TypeVarChar,
				Size:     50,
				Nullable: false,
			},
			{
				Name:     "email",
				Type:     TypeVarChar,
				Size:     100,
				Nullable: false,
			},
			{
				Name:     "created_at",
				Type:     TypeTimestamp,
				Nullable: false,
				Default:  "CURRENT_TIMESTAMP",
			},
			{
				Name:     "is_active",
				Type:     TypeBool,
				Nullable: false,
				Default:  true,
			},
		},
		Indexes: []IndexSchema{
			{
				Name:    "idx_users_username",
				Columns: []string{"username"},
				Unique:  true,
			},
			{
				Name:    "idx_users_email",
				Columns: []string{"email"},
				Unique:  true,
			},
		},
	}
}

// ExampleSchemaOperations demonstrates schema operations (once implemented)
func ExampleSchemaOperations(db Database) {
	fmt.Println("\n--- Schema Operations Example ---")
	
	// Create example schema
	schema := CreateExampleSchema()
	fmt.Printf("Created schema for table: %s\n", schema.Name)
	fmt.Printf("Columns: %d\n", len(schema.Columns))
	fmt.Printf("Indexes: %d\n", len(schema.Indexes))
	
	// Print column details
	for _, col := range schema.Columns {
		fmt.Printf("  Column: %s (%s)", col.Name, col.Type.String())
		if col.PrimaryKey {
			fmt.Printf(" [PRIMARY KEY]")
		}
		if col.AutoIncrement {
			fmt.Printf(" [AUTO INCREMENT]")
		}
		if !col.Nullable {
			fmt.Printf(" [NOT NULL]")
		}
		if col.Default != nil {
			fmt.Printf(" [DEFAULT: %v]", col.Default)
		}
		fmt.Println()
	}
	
	// Print index details
	for _, idx := range schema.Indexes {
		fmt.Printf("  Index: %s on (%s)", idx.Name, idx.Columns)
		if idx.Unique {
			fmt.Printf(" [UNIQUE]")
		}
		fmt.Println()
	}
	
	// Attempt to create table (will fail until implemented)
	err := db.CreateTable(schema.Name, schema)
	fmt.Printf("Table creation test: %v (expected - not implemented yet)\n", err)
}