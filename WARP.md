# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

This is a custom relational database management system (RDBMS) implemented in Go from scratch. The project aims to build a complete database system with:

- Storage engine for persistent data management
- SQL query processor and executor  
- Transaction management with ACID properties
- Index management for efficient data retrieval
- Connection management for client sessions

## Architecture

The codebase follows Go's standard project layout with clear separation of concerns:

### Core Components

- **Storage Engine** (`internal/storage/`): Handles persistent data storage, page management, B+ tree indexes, and transaction logging
- **Query Processing** (`internal/query/`): SQL parser, query optimizer, and query executor
- **Database Interface** (`pkg/database/`): Public API for client connections and database operations
- **Main Application** (`cmd/relational-db/`): Entry point that initializes and coordinates all components

### Directory Structure

- `cmd/relational-db/` - Main application entry point with server initialization
- `internal/` - Private application code (not importable by external packages)
  - `storage/` - Storage engine implementation
  - `query/` - Query processing and execution logic
- `pkg/database/` - Public library interface for database clients
- `tests/` - Integration and unit tests
- `docs/` - Project documentation

### Key Design Patterns

The project is structured as a layered architecture where:
1. The storage layer handles data persistence and indexing
2. The query layer processes SQL and coordinates with storage
3. The database package provides the public interface
4. The main application orchestrates server lifecycle

## Development Commands

### Prerequisites
- Go 1.21 or later
- Git for version control

### Building and Running

Build the application:
```bash
go build -o bin/relational-db ./cmd/relational-db
```

Run from source (development):
```bash
go run ./cmd/relational-db
```

Run the built binary:
```bash
./bin/relational-db
```

### Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests for a specific package:
```bash
go test ./internal/storage
go test ./internal/query
go test ./pkg/database
```

Run a specific test:
```bash
go test -run TestFunctionName ./path/to/package
```

### Development Utilities

Download dependencies:
```bash
go mod download
```

Update dependencies:
```bash
go mod tidy
```

Format code:
```bash
go fmt ./...
```

Lint code (requires golangci-lint):
```bash
golangci-lint run
```

Generate test coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Project Status

This is an early-stage project with basic structure in place. The main components are:
- Basic project scaffolding and directory structure ✅
- Main application entry point with initialization framework ✅
- Storage engine implementation (planned)
- Query processing system (planned)
- Transaction management (planned)
- Client interface and networking (planned)

When working on this codebase, focus on implementing the core database components in the `internal/` packages while maintaining clean interfaces in the `pkg/database/` public API.

Phase 1 (Weeks 1-2): Parser Completion 🎯 START HERE
    └─ Complete recursive descent parser for SELECT, INSERT, CREATE TABLE

Phase 2 (Weeks 3-4): Schema Management
    ├─ Catalog Manager (persist table schemas)
    └─ Record Manager (serialize/deserialize records)

Phase 3 (Weeks 5-6): Query Execution ⭐ MILESTONE
    └─ Simple Executor (END-TO-END SQL QUERIES!)

Phase 4 (Weeks 7-9): Transaction Support
    ├─ Transaction Manager (ACID)
    └─ WAL Manager (durability & crash recovery)

Phase 5 (Weeks 10-12): Performance Optimization
    ├─ B-Tree Indexes (fast lookups)
    └─ Query Optimizer (intelligent plans)