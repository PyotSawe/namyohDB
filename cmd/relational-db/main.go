package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"relational-db/internal/config"
	"relational-db/internal/storage"
)

// Global storage engine (in real implementation, use dependency injection)
var globalStorage *storage.Engine

func main() {
	fmt.Println("Relational Database - Starting...")
	
	// Load configuration
	cfg := config.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	
	fmt.Println(cfg.String())
	
	// Initialize database components
	if err := initializeDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Start the database server
	fmt.Printf("Database server ready to accept connections on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	
	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down database server...")
	
	// Implement graceful shutdown
	if globalStorage != nil {
		if err := globalStorage.Close(); err != nil {
			fmt.Printf("Error closing storage engine: %v\n", err)
		} else {
			fmt.Println("Storage engine closed successfully")
		}
	}
	
	fmt.Println("Database server stopped")
}

// testStorageEngine performs basic tests on the storage engine
func testStorageEngine(engine *storage.Engine) error {
	fmt.Println("Testing storage engine...")
	
	// Test page allocation
	pageID, err := engine.AllocatePage()
	if err != nil {
		return fmt.Errorf("failed to allocate page: %w", err)
	}
	
	// Test page writing
	testData := make([]byte, 4096) // Assuming 4KB page size
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	testPage := &storage.Page{
		ID:   pageID,
		Data: testData,
	}
	
	if err := engine.WritePage(testPage); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}
	
	// Test page reading
	readPage, err := engine.ReadPage(pageID)
	if err != nil {
		return fmt.Errorf("failed to read page: %w", err)
	}
	
	// Verify data
	for i := range testData {
		if readPage.Data[i] != testData[i] {
			return fmt.Errorf("data mismatch at byte %d: expected %d, got %d", i, testData[i], readPage.Data[i])
		}
	}
	
	// Test sync
	if err := engine.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}
	
	fmt.Printf("Storage engine test passed (Page ID: %d)\n", pageID)
	return nil
}

func initializeDatabase(cfg *config.Config) error {
	// Initialize storage engine
	storageEngine, err := storage.NewEngine(&cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage engine: %w", err)
	}
	
	// Test the storage engine
	if err := testStorageEngine(storageEngine); err != nil {
		return fmt.Errorf("storage engine test failed: %w", err)
	}
	
	// TODO: Initialize query processor
	// TODO: Initialize transaction manager with cfg.Database
	// TODO: Initialize connection manager with cfg.Server
	
	fmt.Println("Database components initialized successfully")
	fmt.Println(storageEngine.Stats().String())
	
	// Store the engine globally for now (in a real implementation, use dependency injection)
	globalStorage = storageEngine
	
	// Demonstrate the new SQLite3-style modular architecture
	demonstrateArchitecture()
	demonateSQLProcessing(cfg, storageEngine)
	
	return nil
}
