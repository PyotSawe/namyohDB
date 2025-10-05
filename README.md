# Relational Database

A high-performance relational database management system (RDBMS) implemented in Go from scratch.

## 🚀 Overview

This project is a complete implementation of a relational database featuring:

✅ **Storage Engine** - Page-based storage with buffer pool management and LRU eviction
✅ **Configuration System** - Environment-based configuration with validation
✅ **Public API** - Clean database interface with connection and transaction management
✅ **Testing Suite** - Comprehensive unit and integration tests
✅ **Build Tools** - Cross-platform build scripts and Makefiles
🔄 **Query Processor** - SQL parser and query execution (in progress)
🔄 **Transaction Management** - ACID compliance with concurrency control (in progress)
🔄 **Network Protocol** - TCP server with database wire protocol (in progress)

## 🏗️ Architecture

```
relational-db/
├── cmd/
│   └── relational-db/      # Main application entry point
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── storage/           # Storage engine implementation
│   └── query/             # Query processing and execution
├── pkg/                   # Public library code
│   └── database/          # Database client interface
├── tests/                 # Test suites
│   ├── unit/              # Unit tests
│   └── integration/       # Integration tests
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
├── Makefile              # Cross-platform build commands
└── go.mod                # Go module definition
```

## 📋 Prerequisites

- Go 1.21 or later
- Git (for version control)
- Make (optional, for cross-platform builds)

## ⚡ Quick Start

### Option 1: Using Go directly
```bash
# Clone and build
git clone <repository-url>
cd relational-db
go mod download
go build -o bin/relational-db.exe ./cmd/relational-db

# Run the database
./bin/relational-db.exe
```

### Option 2: Using build scripts
```bash
# Windows
scripts\build.bat
scripts\build.bat test    # Build and run tests

# Unix/Linux/Mac (with Make)
make build
make test
make run
```

## 💻 Usage

### Starting the Database Server

```bash
# Default configuration
./bin/relational-db.exe

# With custom configuration
set DB_PORT=5433
set DB_DATA_DIRECTORY=./custom_data
./bin/relational-db.exe
```

### Configuration Options

The database can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | Server host |
| DB_PORT | 5432 | Server port |
| DB_MAX_CONNECTIONS | 100 | Maximum concurrent connections |
| DB_NAME | relationaldb | Database name |
| DB_DATA_DIRECTORY | ./data | Data storage directory |
| DB_PAGE_SIZE | 4096 | Storage page size (bytes) |
| DB_BUFFER_SIZE | 1000 | Buffer pool size (pages) |

### Using the Database API

```go
package main

import (
    "context"
    "relational-db/internal/config"
    "relational-db/internal/storage"
    "relational-db/pkg/database"
)

func main() {
    // Create configuration
    cfg := config.Default()
    
    // Create storage engine
    storageEngine, _ := storage.NewEngine(&cfg.Storage)
    defer storageEngine.Close()
    
    // Create database instance
    db, _ := database.NewDatabase(cfg, storageEngine)
    defer db.Close()
    
    // Create connection
    ctx := context.Background()
    conn, _ := db.Connect(ctx)
    defer conn.Close()
    
    // Use connection...
}
```

## 🔧 Development

### Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test ./tests/unit/...

# Integration tests only
go test ./tests/integration/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Using Make
make test
make coverage
```

### Development Workflow

```bash
# Quick development run
go run ./cmd/relational-db

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Full development cycle
make full-test  # Clean, build, test, lint, vet
```

### Cross-Platform Building

```bash
# Build for multiple platforms
make build-all

# Individual platforms
make build-linux
make build-windows
make build-darwin
```

## 🎯 Features Status

### ✅ Implemented

- **Storage Engine**
  - ✅ Page-based storage with configurable page size
  - ✅ LRU buffer pool with configurable size
  - ✅ File manager with free page tracking
  - ✅ Thread-safe operations with proper locking
  - ✅ Storage statistics and monitoring

- **Configuration System**
  - ✅ Environment-based configuration
  - ✅ Validation with sensible defaults
  - ✅ Server, database, and storage settings

- **Database API**
  - ✅ Clean public interface
  - ✅ Connection management with pooling
  - ✅ Transaction interface (skeleton)
  - ✅ Result set handling
  - ✅ Database statistics and health monitoring

- **Testing & Quality**
  - ✅ Comprehensive unit tests
  - ✅ Integration tests with concurrency
  - ✅ Persistence testing
  - ✅ Configuration validation tests

- **Development Tools**
  - ✅ Cross-platform build scripts
  - ✅ Makefile with comprehensive targets
  - ✅ Code formatting and linting setup
  - ✅ Coverage reporting

### 🔄 In Progress

- **Query Processing**
  - 🔄 SQL parser
  - 🔄 Query optimizer
  - 🔄 Query executor
  - 🔄 Table management

- **Transaction Management**
  - 🔄 ACID compliance implementation
  - 🔄 Lock management
  - 🔄 Isolation levels
  - 🔄 Deadlock detection

- **Network Protocol**
  - 🔄 TCP server
  - 🔄 Wire protocol implementation
  - 🔄 Client-server communication

### 📋 Planned

- **Advanced Features**
  - 📋 B+ tree indexes
  - 📋 Write-ahead logging (WAL)
  - 📋 Backup and recovery
  - 📋 Replication
  - 📋 Query optimization
  - 📋 Performance monitoring dashboard

## 📊 Performance & Statistics

The database provides detailed statistics and monitoring:

```
Storage Statistics:
  Pages: 100 total, 5 free
  Buffer: 45/1000 pages (95.2% hit ratio)
  I/O: 1,250 reads, 892 writes

Database Statistics:
  Connections: 3 active, 127 total
  Queries: 5,432 executed
  Transactions: 2 active, 891 total
  Uptime: 2h 15m 30s
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add comprehensive tests
5. Run the full test suite (`make full-test`)
6. Submit a pull request

### Development Guidelines

- Follow Go best practices and idioms
- Write tests for all new functionality
- Update documentation for API changes
- Ensure thread safety for concurrent operations
- Profile performance for critical paths

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 References

- [Database System Concepts](https://www.db-book.com/)
- [Architecture of a Database System](https://dsf.berkeley.edu/papers/fntdb07-architecture.pdf)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
