package unit

import (
	"os"
	"path/filepath"
	"testing"
	
	"relational-db/internal/config"
	"relational-db/internal/storage"
)

func TestStorageEngine(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "relational_db_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test configuration
	cfg := &config.StorageConfig{
		DataDirectory: tempDir,
		PageSize:      4096,
		BufferSize:    10,
		MaxFileSize:   1024 * 1024,
	}
	
	// Create storage engine
	engine, err := storage.NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage engine: %v", err)
	}
	defer engine.Close()
	
	t.Run("AllocatePage", func(t *testing.T) {
		pageID, err := engine.AllocatePage()
		if err != nil {
			t.Errorf("Failed to allocate page: %v", err)
		}
		if pageID == 0 {
			t.Error("Allocated page ID should not be 0")
		}
	})
	
	t.Run("WriteAndReadPage", func(t *testing.T) {
		// Allocate a page
		pageID, err := engine.AllocatePage()
		if err != nil {
			t.Fatalf("Failed to allocate page: %v", err)
		}
		
		// Create test data
		testData := make([]byte, cfg.PageSize)
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		
		// Write page
		page := &storage.Page{
			ID:   pageID,
			Data: testData,
		}
		
		if err := engine.WritePage(page); err != nil {
			t.Errorf("Failed to write page: %v", err)
		}
		
		// Read page back
		readPage, err := engine.ReadPage(pageID)
		if err != nil {
			t.Errorf("Failed to read page: %v", err)
		}
		
		// Verify data
		if readPage.ID != pageID {
			t.Errorf("Page ID mismatch: expected %d, got %d", pageID, readPage.ID)
		}
		
		if len(readPage.Data) != len(testData) {
			t.Errorf("Data length mismatch: expected %d, got %d", len(testData), len(readPage.Data))
		}
		
		for i, expected := range testData {
			if readPage.Data[i] != expected {
				t.Errorf("Data mismatch at byte %d: expected %d, got %d", i, expected, readPage.Data[i])
			}
		}
	})
	
	t.Run("DeallocatePage", func(t *testing.T) {
		// Allocate a page
		pageID, err := engine.AllocatePage()
		if err != nil {
			t.Fatalf("Failed to allocate page: %v", err)
		}
		
		// Deallocate the page
		if err := engine.DeallocatePage(pageID); err != nil {
			t.Errorf("Failed to deallocate page: %v", err)
		}
	})
	
	t.Run("Stats", func(t *testing.T) {
		stats := engine.Stats()
		
		if stats.BufferSize <= 0 {
			t.Error("Buffer size should be positive")
		}
		
		if stats.TotalPages < 0 {
			t.Error("Total pages should be non-negative")
		}
	})
	
	t.Run("Sync", func(t *testing.T) {
		if err := engine.Sync(); err != nil {
			t.Errorf("Failed to sync: %v", err)
		}
	})
}

func TestFileManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	pageSize := 4096
	
	// Create file manager
	fm, err := storage.NewFileManager(tempDir, pageSize)
	if err != nil {
		t.Fatalf("Failed to create file manager: %v", err)
	}
	defer fm.Close()
	
	t.Run("AllocateAndReadPage", func(t *testing.T) {
		// Allocate a page
		pageID, err := fm.AllocatePage()
		if err != nil {
			t.Fatalf("Failed to allocate page: %v", err)
		}
		
		// Create test data
		testData := make([]byte, pageSize)
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		
		// Write page
		page := &storage.Page{
			ID:   pageID,
			Data: testData,
		}
		
		if err := fm.WritePage(page); err != nil {
			t.Errorf("Failed to write page: %v", err)
		}
		
		// Read page back
		readPage, err := fm.ReadPage(pageID)
		if err != nil {
			t.Errorf("Failed to read page: %v", err)
		}
		
		// Verify data
		for i, expected := range testData {
			if readPage.Data[i] != expected {
				t.Errorf("Data mismatch at byte %d: expected %d, got %d", i, expected, readPage.Data[i])
			}
		}
	})
	
	t.Run("InvalidPageID", func(t *testing.T) {
		// Try to read page with ID 0 (invalid)
		_, err := fm.ReadPage(0)
		if err != storage.ErrInvalidPageID {
			t.Errorf("Expected ErrInvalidPageID, got %v", err)
		}
		
		// Try to write page with ID 0 (invalid)
		page := &storage.Page{
			ID:   0,
			Data: make([]byte, pageSize),
		}
		err = fm.WritePage(page)
		if err != storage.ErrInvalidPageID {
			t.Errorf("Expected ErrInvalidPageID, got %v", err)
		}
	})
	
	t.Run("FileCreation", func(t *testing.T) {
		// Check if data files were created
		dataFile := filepath.Join(tempDir, "data.db")
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			t.Error("Data file was not created")
		}
		
		freePagesFile := filepath.Join(tempDir, "free_pages.db")
		if _, err := os.Stat(freePagesFile); os.IsNotExist(err) {
			t.Error("Free pages file was not created")
		}
	})
}

func TestBufferPool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "buffer_pool_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	pageSize := 4096
	bufferSize := 3
	
	// Create file manager
	fm, err := storage.NewFileManager(tempDir, pageSize)
	if err != nil {
		t.Fatalf("Failed to create file manager: %v", err)
	}
	defer fm.Close()
	
	// Create buffer pool
	bp := storage.NewBufferPool(bufferSize, fm)
	
	t.Run("BasicOperations", func(t *testing.T) {
		// Allocate a page through file manager
		pageID, err := fm.AllocatePage()
		if err != nil {
			t.Fatalf("Failed to allocate page: %v", err)
		}
		
		// Create test data
		testData := make([]byte, pageSize)
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		
		// Put page in buffer
		page := &storage.Page{
			ID:   pageID,
			Data: testData,
		}
		
		if err := bp.PutPage(page); err != nil {
			t.Errorf("Failed to put page in buffer: %v", err)
		}
		
		// Get page from buffer
		retrievedPage, err := bp.GetPage(pageID)
		if err != nil {
			t.Errorf("Failed to get page from buffer: %v", err)
		}
		
		// Verify data
		for i, expected := range testData {
			if retrievedPage.Data[i] != expected {
				t.Errorf("Data mismatch at byte %d: expected %d, got %d", i, expected, retrievedPage.Data[i])
			}
		}
	})
	
	t.Run("LRUEviction", func(t *testing.T) {
		// Fill buffer pool beyond capacity to test eviction
		pageIDs := make([]storage.PageID, bufferSize+2)
		
		for i := 0; i < bufferSize+2; i++ {
			pageID, err := fm.AllocatePage()
			if err != nil {
				t.Fatalf("Failed to allocate page: %v", err)
			}
			pageIDs[i] = pageID
			
			testData := make([]byte, pageSize)
			for j := range testData {
				testData[j] = byte(i)
			}
			
			page := &storage.Page{
				ID:   pageID,
				Data: testData,
			}
			
			if err := bp.PutPage(page); err != nil {
				t.Errorf("Failed to put page %d in buffer: %v", i, err)
			}
		}
		
		// Buffer should now be at capacity
		hits, misses, used, capacity := bp.Stats()
		if used > capacity {
			t.Errorf("Buffer pool exceeded capacity: used=%d, capacity=%d", used, capacity)
		}
		
		if hits == 0 && misses == 0 {
			t.Error("No buffer statistics recorded")
		}
	})
	
	t.Run("FlushOperations", func(t *testing.T) {
		// Allocate and add a page
		pageID, err := fm.AllocatePage()
		if err != nil {
			t.Fatalf("Failed to allocate page: %v", err)
		}
		
		testData := make([]byte, pageSize)
		for i := range testData {
			testData[i] = byte(42)
		}
		
		page := &storage.Page{
			ID:   pageID,
			Data: testData,
		}
		
		if err := bp.PutPage(page); err != nil {
			t.Errorf("Failed to put page in buffer: %v", err)
		}
		
		// Flush the page
		if err := bp.FlushPage(pageID); err != nil {
			t.Errorf("Failed to flush page: %v", err)
		}
		
		// Flush all pages
		if err := bp.FlushAll(); err != nil {
			t.Errorf("Failed to flush all pages: %v", err)
		}
	})
}