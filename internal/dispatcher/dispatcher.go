package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"relational-db/internal/config"
	"relational-db/internal/lexer"
	"relational-db/internal/parser"
	"relational-db/internal/storage"
)

// QueryType represents different types of SQL queries
type QueryType int

const (
	QueryTypeSelect QueryType = iota
	QueryTypeInsert
	QueryTypeUpdate
	QueryTypeDelete
	QueryTypeCreateTable
	QueryTypeDropTable
	QueryTypeCreateIndex
	QueryTypeDropIndex
)

func (qt QueryType) String() string {
	switch qt {
	case QueryTypeSelect:
		return "SELECT"
	case QueryTypeInsert:
		return "INSERT"
	case QueryTypeUpdate:
		return "UPDATE"
	case QueryTypeDelete:
		return "DELETE"
	case QueryTypeCreateTable:
		return "CREATE_TABLE"
	case QueryTypeDropTable:
		return "DROP_TABLE"
	case QueryTypeCreateIndex:
		return "CREATE_INDEX"
	case QueryTypeDropIndex:
		return "DROP_INDEX"
	default:
		return "UNKNOWN"
	}
}

// QueryContext holds context information for query execution
type QueryContext struct {
	ConnectionID string
	UserID       string
	DatabaseName string
	StartTime    time.Time
	Timeout      time.Duration
}

// QueryResult represents the result of query execution
type QueryResult struct {
	Columns      []string
	Rows         [][]interface{}
	RowsAffected int64
	LastInsertID int64
	ExecutionTime time.Duration
	Error         error
}

// QueryPlan represents an execution plan for a query
type QueryPlan struct {
	QueryType     QueryType
	AST           parser.Statement
	EstimatedCost float64
	Operations    []Operation
}

// Operation represents a single operation in the execution plan
type Operation struct {
	Type        OperationType
	TableName   string
	IndexName   string
	Conditions  []Condition
	Projections []string
	Cost        float64
}

type OperationType int

const (
	OpTableScan OperationType = iota
	OpIndexScan
	OpNestedLoopJoin
	OpHashJoin
	OpSort
	OpFilter
	OpProjection
	OpInsert
	OpUpdate
	OpDelete
)

func (ot OperationType) String() string {
	switch ot {
	case OpTableScan:
		return "TABLE_SCAN"
	case OpIndexScan:
		return "INDEX_SCAN"
	case OpNestedLoopJoin:
		return "NESTED_LOOP_JOIN"
	case OpHashJoin:
		return "HASH_JOIN"
	case OpSort:
		return "SORT"
	case OpFilter:
		return "FILTER"
	case OpProjection:
		return "PROJECTION"
	case OpInsert:
		return "INSERT"
	case OpUpdate:
		return "UPDATE"
	case OpDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// Condition represents a filter condition
type Condition struct {
	Column   string
	Operator ComparisonOperator
	Value    interface{}
}

type ComparisonOperator int

const (
	OpEqual ComparisonOperator = iota
	OpNotEqual
	OpLessThan
	OpGreaterThan
	OpLessEqual
	OpGreaterEqual
	OpLike
	OpIn
	OpBetween
)

// Dispatcher is the main query dispatcher
type Dispatcher struct {
	mu              sync.RWMutex
	config          *config.Config
	storageEngine   storage.StorageEngine
	
	// Query execution statistics
	queriesExecuted int64
	totalExecutionTime time.Duration
	queryTypeStats  map[QueryType]int64
}

// NewDispatcher creates a new query dispatcher
func NewDispatcher(cfg *config.Config, storageEngine storage.StorageEngine) *Dispatcher {
	return &Dispatcher{
		config:         cfg,
		storageEngine:  storageEngine,
		queryTypeStats: make(map[QueryType]int64),
	}
}

// DispatchQuery processes and routes a SQL query to appropriate subsystems
func (d *Dispatcher) DispatchQuery(ctx context.Context, sql string, queryCtx *QueryContext) (*QueryResult, error) {
	startTime := time.Now()
	
	// Step 1: Lexical Analysis
	_, err := lexer.TokenizeSQL(sql)
	if err != nil {
		return &QueryResult{Error: fmt.Errorf("lexical analysis failed: %w", err)}, nil
	}
	
	// Step 2: Parse SQL into AST
	stmt, err := parser.ParseSQL(sql)
	if err != nil {
		return &QueryResult{Error: fmt.Errorf("parsing failed: %w", err)}, nil
	}
	
	// Step 3: Determine query type
	queryType := d.determineQueryType(stmt)
	
	// Step 4: Create query plan
	plan, err := d.createQueryPlan(ctx, stmt, queryType)
	if err != nil {
		return &QueryResult{Error: fmt.Errorf("query planning failed: %w", err)}, nil
	}
	
	// Step 5: Execute query based on type
	result, err := d.executeQuery(ctx, plan, queryCtx)
	if err != nil {
		return &QueryResult{Error: fmt.Errorf("query execution failed: %w", err)}, nil
	}
	
	// Update statistics
	d.updateStats(queryType, time.Since(startTime))
	
	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// determineQueryType determines the type of SQL query
func (d *Dispatcher) determineQueryType(stmt parser.Statement) QueryType {
	switch stmt.(type) {
	case *parser.SelectStatement:
		return QueryTypeSelect
	case *parser.InsertStatement:
		return QueryTypeInsert
	case *parser.UpdateStatement:
		return QueryTypeUpdate
	case *parser.DeleteStatement:
		return QueryTypeDelete
	case *parser.CreateTableStatement:
		return QueryTypeCreateTable
	case *parser.DropTableStatement:
		return QueryTypeDropTable
	default:
		return QueryType(-1) // Unknown
	}
}

// createQueryPlan creates an execution plan for the query
func (d *Dispatcher) createQueryPlan(ctx context.Context, stmt parser.Statement, queryType QueryType) (*QueryPlan, error) {
	
	switch queryType {
	case QueryTypeSelect:
		return d.planSelectQuery(ctx, stmt.(*parser.SelectStatement))
	case QueryTypeInsert:
		return d.planInsertQuery(ctx, stmt.(*parser.InsertStatement))
	case QueryTypeUpdate:
		return d.planUpdateQuery(ctx, stmt.(*parser.UpdateStatement))
	case QueryTypeDelete:
		return d.planDeleteQuery(ctx, stmt.(*parser.DeleteStatement))
	case QueryTypeCreateTable:
		return d.planCreateTableQuery(ctx, stmt.(*parser.CreateTableStatement))
	case QueryTypeDropTable:
		return d.planDropTableQuery(ctx, stmt.(*parser.DropTableStatement))
	default:
		return nil, fmt.Errorf("unsupported query type: %v", queryType)
	}
}

// planSelectQuery creates an execution plan for SELECT queries
func (d *Dispatcher) planSelectQuery(ctx context.Context, stmt *parser.SelectStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeSelect,
		AST:       stmt,
	}
	
	// For now, create a simple table scan plan
	// TODO: Implement proper query optimization
	if stmt.FromClause != nil && len(stmt.FromClause.Tables) > 0 {
		// Extract table name from first table
		tableName := d.extractTableName(stmt.FromClause.Tables[0])
		
		op := Operation{
			Type:      OpTableScan,
			TableName: tableName,
			Cost:      100.0, // Estimated cost
		}
		
		// Add filter operation if WHERE clause exists
		if stmt.WhereClause != nil {
			filterOp := Operation{
				Type: OpFilter,
				Cost: 10.0,
			}
			plan.Operations = append(plan.Operations, filterOp)
		}
		
		// Add projection operation
		if stmt.SelectClause != nil {
			projOp := Operation{
				Type: OpProjection,
				Cost: 5.0,
			}
			plan.Operations = append(plan.Operations, projOp)
		}
		
		plan.Operations = append([]Operation{op}, plan.Operations...)
		plan.EstimatedCost = d.calculatePlanCost(plan.Operations)
	}
	
	return plan, nil
}

// planInsertQuery creates an execution plan for INSERT queries
func (d *Dispatcher) planInsertQuery(ctx context.Context, stmt *parser.InsertStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeInsert,
		AST:       stmt,
	}
	
	op := Operation{
		Type:      OpInsert,
		TableName: stmt.TableName.Value,
		Cost:      20.0,
	}
	
	plan.Operations = append(plan.Operations, op)
	plan.EstimatedCost = 20.0
	
	return plan, nil
}

// planUpdateQuery creates an execution plan for UPDATE queries
func (d *Dispatcher) planUpdateQuery(ctx context.Context, stmt *parser.UpdateStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeUpdate,
		AST:       stmt,
	}
	
	// Table scan to find rows to update
	scanOp := Operation{
		Type:      OpTableScan,
		TableName: stmt.TableName.Value,
		Cost:      100.0,
	}
	
	// Filter operation if WHERE clause exists
	if stmt.WhereClause != nil {
		filterOp := Operation{
			Type: OpFilter,
			Cost: 10.0,
		}
		plan.Operations = append(plan.Operations, scanOp, filterOp)
	} else {
		plan.Operations = append(plan.Operations, scanOp)
	}
	
	// Update operation
	updateOp := Operation{
		Type:      OpUpdate,
		TableName: stmt.TableName.Value,
		Cost:      30.0,
	}
	plan.Operations = append(plan.Operations, updateOp)
	
	plan.EstimatedCost = d.calculatePlanCost(plan.Operations)
	
	return plan, nil
}

// planDeleteQuery creates an execution plan for DELETE queries
func (d *Dispatcher) planDeleteQuery(ctx context.Context, stmt *parser.DeleteStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeDelete,
		AST:       stmt,
	}
	
	// Table scan to find rows to delete
	scanOp := Operation{
		Type:      OpTableScan,
		TableName: stmt.TableName.Value,
		Cost:      100.0,
	}
	
	// Filter operation if WHERE clause exists
	if stmt.WhereClause != nil {
		filterOp := Operation{
			Type: OpFilter,
			Cost: 10.0,
		}
		plan.Operations = append(plan.Operations, scanOp, filterOp)
	} else {
		plan.Operations = append(plan.Operations, scanOp)
	}
	
	// Delete operation
	deleteOp := Operation{
		Type:      OpDelete,
		TableName: stmt.TableName.Value,
		Cost:      25.0,
	}
	plan.Operations = append(plan.Operations, deleteOp)
	
	plan.EstimatedCost = d.calculatePlanCost(plan.Operations)
	
	return plan, nil
}

// planCreateTableQuery creates an execution plan for CREATE TABLE queries
func (d *Dispatcher) planCreateTableQuery(ctx context.Context, stmt *parser.CreateTableStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeCreateTable,
		AST:       stmt,
	}
	
	// Simple operation for table creation
	op := Operation{
		Type:      OpInsert, // Reuse insert type for DDL operations
		TableName: stmt.TableName.Value,
		Cost:      50.0,
	}
	
	plan.Operations = append(plan.Operations, op)
	plan.EstimatedCost = 50.0
	
	return plan, nil
}

// planDropTableQuery creates an execution plan for DROP TABLE queries
func (d *Dispatcher) planDropTableQuery(ctx context.Context, stmt *parser.DropTableStatement) (*QueryPlan, error) {
	plan := &QueryPlan{
		QueryType: QueryTypeDropTable,
		AST:       stmt,
	}
	
	// Simple operation for table dropping
	op := Operation{
		Type:      OpDelete, // Reuse delete type for DDL operations
		TableName: stmt.TableName.Value,
		Cost:      40.0,
	}
	
	plan.Operations = append(plan.Operations, op)
	plan.EstimatedCost = 40.0
	
	return plan, nil
}

// executeQuery executes the query plan
func (d *Dispatcher) executeQuery(ctx context.Context, plan *QueryPlan, queryCtx *QueryContext) (*QueryResult, error) {
	switch plan.QueryType {
	case QueryTypeSelect:
		return d.executeSelectQuery(ctx, plan)
	case QueryTypeInsert:
		return d.executeInsertQuery(ctx, plan)
	case QueryTypeUpdate:
		return d.executeUpdateQuery(ctx, plan)
	case QueryTypeDelete:
		return d.executeDeleteQuery(ctx, plan)
	case QueryTypeCreateTable:
		return d.executeCreateTableQuery(ctx, plan)
	case QueryTypeDropTable:
		return d.executeDropTableQuery(ctx, plan)
	default:
		return nil, fmt.Errorf("unsupported query type for execution: %v", plan.QueryType)
	}
}

// executeSelectQuery executes SELECT queries
func (d *Dispatcher) executeSelectQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual SELECT execution with storage engine
	// For now, return a placeholder result
	return &QueryResult{
		Columns:      []string{"id", "name"},
		Rows:         [][]interface{}{{1, "test"}, {2, "example"}},
		RowsAffected: 0,
		LastInsertID: 0,
	}, nil
}

// executeInsertQuery executes INSERT queries
func (d *Dispatcher) executeInsertQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual INSERT execution with storage engine
	return &QueryResult{
		Columns:      []string{},
		Rows:         [][]interface{}{},
		RowsAffected: 1,
		LastInsertID: 1,
	}, nil
}

// executeUpdateQuery executes UPDATE queries
func (d *Dispatcher) executeUpdateQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual UPDATE execution with storage engine
	return &QueryResult{
		Columns:      []string{},
		Rows:         [][]interface{}{},
		RowsAffected: 1,
		LastInsertID: 0,
	}, nil
}

// executeDeleteQuery executes DELETE queries
func (d *Dispatcher) executeDeleteQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual DELETE execution with storage engine
	return &QueryResult{
		Columns:      []string{},
		Rows:         [][]interface{}{},
		RowsAffected: 1,
		LastInsertID: 0,
	}, nil
}

// executeCreateTableQuery executes CREATE TABLE queries
func (d *Dispatcher) executeCreateTableQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual CREATE TABLE execution with storage engine
	return &QueryResult{
		Columns:      []string{},
		Rows:         [][]interface{}{},
		RowsAffected: 0,
		LastInsertID: 0,
	}, nil
}

// executeDropTableQuery executes DROP TABLE queries
func (d *Dispatcher) executeDropTableQuery(ctx context.Context, plan *QueryPlan) (*QueryResult, error) {
	// TODO: Implement actual DROP TABLE execution with storage engine
	return &QueryResult{
		Columns:      []string{},
		Rows:         [][]interface{}{},
		RowsAffected: 0,
		LastInsertID: 0,
	}, nil
}

// Helper functions

// extractTableName extracts table name from expression
func (d *Dispatcher) extractTableName(expr parser.Expression) string {
	switch e := expr.(type) {
	case *parser.Identifier:
		return e.Value
	case *parser.ColumnReference:
		if e.Table != nil {
			return e.Table.Value
		}
	}
	return "unknown_table"
}

// calculatePlanCost calculates the total cost of an execution plan
func (d *Dispatcher) calculatePlanCost(operations []Operation) float64 {
	var totalCost float64
	for _, op := range operations {
		totalCost += op.Cost
	}
	return totalCost
}

// updateStats updates query execution statistics
func (d *Dispatcher) updateStats(queryType QueryType, executionTime time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.queriesExecuted++
	d.totalExecutionTime += executionTime
	d.queryTypeStats[queryType]++
}

// GetStats returns dispatcher statistics
func (d *Dispatcher) GetStats() DispatcherStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	stats := DispatcherStats{
		QueriesExecuted:    d.queriesExecuted,
		TotalExecutionTime: d.totalExecutionTime,
		QueryTypeStats:     make(map[QueryType]int64),
	}
	
	for queryType, count := range d.queryTypeStats {
		stats.QueryTypeStats[queryType] = count
	}
	
	if d.queriesExecuted > 0 {
		stats.AverageExecutionTime = d.totalExecutionTime / time.Duration(d.queriesExecuted)
	}
	
	return stats
}

// DispatcherStats holds statistics about query execution
type DispatcherStats struct {
	QueriesExecuted      int64
	TotalExecutionTime   time.Duration
	AverageExecutionTime time.Duration
	QueryTypeStats       map[QueryType]int64
}

// String returns a string representation of dispatcher statistics
func (ds DispatcherStats) String() string {
	return fmt.Sprintf(`Query Dispatcher Statistics:
  Total Queries: %d
  Total Execution Time: %v
  Average Execution Time: %v
  Query Type Breakdown:
    SELECT: %d
    INSERT: %d
    UPDATE: %d
    DELETE: %d
    CREATE TABLE: %d
    DROP TABLE: %d`,
		ds.QueriesExecuted,
		ds.TotalExecutionTime,
		ds.AverageExecutionTime,
		ds.QueryTypeStats[QueryTypeSelect],
		ds.QueryTypeStats[QueryTypeInsert],
		ds.QueryTypeStats[QueryTypeUpdate],
		ds.QueryTypeStats[QueryTypeDelete],
		ds.QueryTypeStats[QueryTypeCreateTable],
		ds.QueryTypeStats[QueryTypeDropTable])
}

// ValidateQuery performs basic validation on the query
func (d *Dispatcher) ValidateQuery(sql string) error {
	if sql == "" {
		return fmt.Errorf("empty query")
	}
	
	// Basic tokenization check
	tokens, err := lexer.TokenizeSQL(sql)
	if err != nil {
		return fmt.Errorf("invalid SQL syntax: %w", err)
	}
	
	if len(tokens) == 0 {
		return fmt.Errorf("empty query")
	}
	
	return nil
}

// ExplainQuery returns the execution plan for a query without executing it
func (d *Dispatcher) ExplainQuery(ctx context.Context, sql string) (*QueryPlan, error) {
	stmt, err := parser.ParseSQL(sql)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	
	queryType := d.determineQueryType(stmt)
	plan, err := d.createQueryPlan(ctx, stmt, queryType)
	if err != nil {
		return nil, fmt.Errorf("query planning failed: %w", err)
	}
	
	return plan, nil
}