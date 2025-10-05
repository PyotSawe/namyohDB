package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the database server
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
}

// ServerConfig holds network server configuration
type ServerConfig struct {
	Host string
	Port int
	MaxConnections int
}

// DatabaseConfig holds database-specific settings
type DatabaseConfig struct {
	Name string
	MaxTransactions int
	QueryTimeout int // seconds
}

// StorageConfig holds storage engine configuration
type StorageConfig struct {
	DataDirectory string
	PageSize     int
	BufferSize   int // number of pages in buffer pool
	MaxFileSize  int64 // maximum file size in bytes
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "localhost",
			Port:           5432,
			MaxConnections: 100,
		},
		Database: DatabaseConfig{
			Name:            "relationaldb",
			MaxTransactions: 1000,
			QueryTimeout:    30,
		},
		Storage: StorageConfig{
			DataDirectory: "./data",
			PageSize:      4096, // 4KB pages
			BufferSize:    1000, // 1000 pages in buffer pool (~4MB)
			MaxFileSize:   1024 * 1024 * 1024, // 1GB max file size
		},
	}
}

// LoadFromEnv loads configuration from environment variables, falling back to defaults
func LoadFromEnv() *Config {
	cfg := Default()
	
	// Server configuration
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}
	if maxConnStr := os.Getenv("DB_MAX_CONNECTIONS"); maxConnStr != "" {
		if maxConn, err := strconv.Atoi(maxConnStr); err == nil {
			cfg.Server.MaxConnections = maxConn
		}
	}
	
	// Database configuration
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}
	if timeoutStr := os.Getenv("DB_QUERY_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Database.QueryTimeout = timeout
		}
	}
	
	// Storage configuration
	if dataDir := os.Getenv("DB_DATA_DIRECTORY"); dataDir != "" {
		cfg.Storage.DataDirectory = dataDir
	}
	if pageSizeStr := os.Getenv("DB_PAGE_SIZE"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			cfg.Storage.PageSize = pageSize
		}
	}
	if bufferSizeStr := os.Getenv("DB_BUFFER_SIZE"); bufferSizeStr != "" {
		if bufferSize, err := strconv.Atoi(bufferSizeStr); err == nil {
			cfg.Storage.BufferSize = bufferSize
		}
	}
	
	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Server.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive: %d", c.Server.MaxConnections)
	}
	
	if c.Storage.PageSize <= 0 || c.Storage.PageSize%512 != 0 {
		return fmt.Errorf("page size must be positive and multiple of 512: %d", c.Storage.PageSize)
	}
	
	if c.Storage.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive: %d", c.Storage.BufferSize)
	}
	
	return nil
}

// String returns a formatted string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(`Database Configuration:
  Server:
    Host: %s
    Port: %d
    Max Connections: %d
  Database:
    Name: %s
    Max Transactions: %d
    Query Timeout: %d seconds
  Storage:
    Data Directory: %s
    Page Size: %d bytes
    Buffer Size: %d pages
    Max File Size: %d bytes`,
		c.Server.Host, c.Server.Port, c.Server.MaxConnections,
		c.Database.Name, c.Database.MaxTransactions, c.Database.QueryTimeout,
		c.Storage.DataDirectory, c.Storage.PageSize, c.Storage.BufferSize, c.Storage.MaxFileSize)
}