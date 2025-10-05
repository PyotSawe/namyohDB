package executor

import (
	"context"
	"time"

	"relational-db/internal/storage"
)

// ExecutionContext holds the context for query execution
type ExecutionContext struct {
	ctx         context.Context
	config      *ExecutorConfig
	storage     storage.StorageEngine
	bufferPool  *storage.BufferPool
	startTime   time.Time
	memoryUsed  int64
	memoryLimit int64
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, config *ExecutorConfig) *ExecutionContext {
	return &ExecutionContext{
		ctx:         ctx,
		config:      config,
		startTime:   time.Now(),
		memoryLimit: config.MaxMemoryBytes,
	}
}

// SetStorage sets the storage engine
func (ec *ExecutionContext) SetStorage(engine storage.StorageEngine) {
	ec.storage = engine
}

// SetBufferPool sets the buffer pool
func (ec *ExecutionContext) SetBufferPool(pool *storage.BufferPool) {
	ec.bufferPool = pool
}

// GetStorage returns the storage engine
func (ec *ExecutionContext) GetStorage() storage.StorageEngine {
	return ec.storage
}

// GetBufferPool returns the buffer pool
func (ec *ExecutionContext) GetBufferPool() *storage.BufferPool {
	return ec.bufferPool
}

// Context returns the underlying context
func (ec *ExecutionContext) Context() context.Context {
	return ec.ctx
}

// IsTimedOut checks if execution has timed out
func (ec *ExecutionContext) IsTimedOut() bool {
	if ec.config.QueryTimeout == 0 {
		return false
	}
	return time.Since(ec.startTime) > ec.config.QueryTimeout
}

// AllocateMemory tries to allocate memory for an operator
func (ec *ExecutionContext) AllocateMemory(bytes int64) error {
	if ec.memoryUsed+bytes > ec.memoryLimit {
		return ErrInsufficientMemory
	}
	ec.memoryUsed += bytes
	return nil
}

// ReleaseMemory releases allocated memory
func (ec *ExecutionContext) ReleaseMemory(bytes int64) {
	ec.memoryUsed -= bytes
	if ec.memoryUsed < 0 {
		ec.memoryUsed = 0
	}
}

// GetMemoryUsed returns current memory usage
func (ec *ExecutionContext) GetMemoryUsed() int64 {
	return ec.memoryUsed
}

// GetMemoryLimit returns memory limit
func (ec *ExecutionContext) GetMemoryLimit() int64 {
	return ec.memoryLimit
}

// GetWorkMemory returns memory available per operator
func (ec *ExecutionContext) GetWorkMemory() int64 {
	return ec.config.WorkMemBytes
}

// Elapsed returns time since execution started
func (ec *ExecutionContext) Elapsed() time.Duration {
	return time.Since(ec.startTime)
}
