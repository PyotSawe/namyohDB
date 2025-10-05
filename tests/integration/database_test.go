package integration

import (
	"os"
	"testing"
	"time"
	
	"relational-db/internal/config"
	"relational-db/internal/storage"
)

func TestDatabaseIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           5433, // Different from default to avoid conflicts
			MaxConnections: 10,
		},
		Database: config.DatabaseConfig{
			Name:            "testdb",
			MaxTransactions: 100,
			QueryTimeout:    10,
		},
		Storage: config.StorageConfig{
			DataDirectory: tempDir,
			PageSize:      4096,
			BufferSize:    50,
			MaxFileSize:   10 * 1024 * 1024, // 10MB
		},
	}
	
	t.Run("FullSystemTest", func(t *testing.T) {
		// Test storage engine initialization
		engine, err := storage.NewEngine(&cfg.Storage)
		if err != nil {
			t.Fatalf("Failed to create storage engine: %v", err)
		}
		defer engine.Close()
		
		// Test multiple concurrent page operations
		const numPages = 100
		pageIDs := make([]storage.PageID, numPages)
		
		// Allocate pages
		for i := 0; i < numPages; i++ {
			pageID, err := engine.AllocatePage()
			if err != nil {
				t.Fatalf("Failed to allocate page %d: %v", i, err)
			}
			pageIDs[i] = pageID
		}
		
		// Write data to pages
		for i, pageID := range pageIDs {
			testData := make([]byte, cfg.Storage.PageSize)
			for j := range testData {
				testData[j] = byte((i + j) % 256)
			}
			
			page := &storage.Page{
				ID:   pageID,
				Data: testData,
			}
			
			if err := engine.WritePage(page); err != nil {
				t.Fatalf("Failed to write page %d: %v", i, err)
			}
		}
		
		// Read and verify all pages
		for i, pageID := range pageIDs {
			page, err := engine.ReadPage(pageID)
			if err != nil {
				t.Fatalf("Failed to read page %d: %v", i, err)
			}
			
			// Verify data
			for j, expected := range page.Data {
				if expected != byte((i+j)%256) {
					t.Errorf("Page %d byte %d: expected %d, got %d", i, j, byte((i+j)%256), expected)
				}
			}
		}
		
		// Test persistence by syncing
		if err := engine.Sync(); err != nil {
			t.Fatalf("Failed to sync: %v", err)
		}
		
		// Get stats
		stats := engine.Stats()
		t.Logf("Storage statistics: %s", stats.String())
		
		if stats.TotalPages < uint64(numPages) {
			t.Errorf("Expected at least %d pages, got %d", numPages, stats.TotalPages)
		}
		
		// Test page deallocation
		for i := 0; i < numPages/2; i++ {
			if err := engine.DeallocatePage(pageIDs[i]); err != nil {
				t.Fatalf("Failed to deallocate page %d: %v", i, err)
			}
		}
		
		// Verify free pages increased
		newStats := engine.Stats()
		if newStats.FreePages == 0 {
			t.Error("Expected some free pages after deallocation")
		}
	})
	
	t.Run("ConcurrencyTest", func(t *testing.T) {
		engine, err := storage.NewEngine(&cfg.Storage)
		if err != nil {
			t.Fatalf("Failed to create storage engine: %v", err)
		}
		defer engine.Close()
		
		// Test concurrent page operations
		const numWorkers = 10
		const pagesPerWorker = 10
		
		done := make(chan bool, numWorkers)
		errors := make(chan error, numWorkers)
		
		// Start concurrent workers
		for w := 0; w < numWorkers; w++ {
			go func(workerID int) {
				defer func() { done <- true }()
				
				// Each worker allocates and writes pages
				for i := 0; i < pagesPerWorker; i++ {
					pageID, err := engine.AllocatePage()
					if err != nil {
						errors <- err
						return
					}
					
					testData := make([]byte, cfg.Storage.PageSize)
					for j := range testData {
						testData[j] = byte((workerID*1000 + i + j) % 256)
					}
					
					page := &storage.Page{
						ID:   pageID,
						Data: testData,
					}
					
					if err := engine.WritePage(page); err != nil {
						errors <- err
						return
					}
					
					// Read back to verify
					readPage, err := engine.ReadPage(pageID)
					if err != nil {
						errors <- err
						return
					}
					
					// Quick verification
					if readPage.ID != pageID {
						errors <- err
						return
					}
				}
			}(w)
		}
		
		// Wait for all workers to complete
		completedWorkers := 0
		timeout := time.After(30 * time.Second)
		
		for completedWorkers < numWorkers {
			select {
			case <-done:
				completedWorkers++
			case err := <-errors:
				t.Fatalf("Worker error: %v", err)
			case <-timeout:
				t.Fatal("Test timed out")
			}
		}
		
		// Verify final state
		stats := engine.Stats()
		expectedMinPages := uint64(numWorkers * pagesPerWorker)
		if stats.TotalPages < expectedMinPages {
			t.Errorf("Expected at least %d pages, got %d", expectedMinPages, stats.TotalPages)
		}
		
		t.Logf("Concurrency test completed: %s", stats.String())
	})
	
	t.Run("PersistenceTest", func(t *testing.T) {
		const testPageCount = 5
		var pageIDs []storage.PageID
		testDataMap := make(map[storage.PageID][]byte)
		
		// Phase 1: Create and populate pages
		{
			engine, err := storage.NewEngine(&cfg.Storage)
			if err != nil {
				t.Fatalf("Failed to create storage engine: %v", err)
			}
			
			for i := 0; i < testPageCount; i++ {
				pageID, err := engine.AllocatePage()
				if err != nil {
					t.Fatalf("Failed to allocate page: %v", err)
				}
				pageIDs = append(pageIDs, pageID)
				
				testData := make([]byte, cfg.Storage.PageSize)
				for j := range testData {
					testData[j] = byte((i*100 + j) % 256)
				}
				testDataMap[pageID] = make([]byte, len(testData))
				copy(testDataMap[pageID], testData)
				
				page := &storage.Page{
					ID:   pageID,
					Data: testData,
				}
				
				if err := engine.WritePage(page); err != nil {
					t.Fatalf("Failed to write page: %v", err)
				}
			}
			
			// Sync and close
			if err := engine.Sync(); err != nil {
				t.Fatalf("Failed to sync: %v", err)
			}
			
			if err := engine.Close(); err != nil {
				t.Fatalf("Failed to close engine: %v", err)
			}
		}
		
		// Phase 2: Reopen and verify persistence
		{
			engine, err := storage.NewEngine(&cfg.Storage)
			if err != nil {
				t.Fatalf("Failed to reopen storage engine: %v", err)
			}
			defer engine.Close()
			
			// Read and verify all pages
			for _, pageID := range pageIDs {
				page, err := engine.ReadPage(pageID)
				if err != nil {
					t.Fatalf("Failed to read persisted page %d: %v", pageID, err)
				}
				
				expectedData := testDataMap[pageID]
				for j, expected := range expectedData {
					if page.Data[j] != expected {
						t.Errorf("Persistence failed for page %d byte %d: expected %d, got %d", 
							pageID, j, expected, page.Data[j])
					}
				}
			}
			
			t.Logf("Persistence test passed for %d pages", len(pageIDs))
		}
	})
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		cfg := config.Default()
		if err := cfg.Validate(); err != nil {
			t.Errorf("Default configuration should be valid: %v", err)
		}
	})
	
	t.Run("InvalidPort", func(t *testing.T) {
		cfg := config.Default()
		cfg.Server.Port = 0
		if err := cfg.Validate(); err == nil {
			t.Error("Configuration with port 0 should be invalid")
		}
	})
	
	t.Run("InvalidPageSize", func(t *testing.T) {
		cfg := config.Default()
		cfg.Storage.PageSize = 100 // Not multiple of 512
		if err := cfg.Validate(); err == nil {
			t.Error("Configuration with invalid page size should be invalid")
		}
	})
	
	t.Run("EnvironmentOverrides", func(t *testing.T) {
		// Set environment variable
		os.Setenv("DB_PORT", "9999")
		defer os.Unsetenv("DB_PORT")
		
		cfg := config.LoadFromEnv()
		if cfg.Server.Port != 9999 {
			t.Errorf("Environment variable not applied: expected port 9999, got %d", cfg.Server.Port)
		}
	})
}