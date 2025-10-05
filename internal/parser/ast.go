package parser

import (
	"fmt"
	"strings"
	
	"relational-db/internal/lexer"
)

// AST Node types represent different SQL statement components

// Node is the base interface for all AST nodes
type Node interface {
	String() string
	NodeType() string
}

// Statement represents a complete SQL statement
type Statement interface {
	Node
	StatementNode()
}

// Expression represents SQL expressions (WHERE conditions, column references, etc.)
type Expression interface {
	Node
	ExpressionNode()
}

// SelectStatement represents a SELECT query
type SelectStatement struct {
	SelectClause *SelectClause
	FromClause   *FromClause
	WhereClause  *WhereClause
	GroupBy      *GroupByClause
	Having       *HavingClause
	OrderBy      *OrderByClause
	Limit        *LimitClause
}

func (s *SelectStatement) StatementNode() {}
func (s *SelectStatement) NodeType() string { return "SelectStatement" }
func (s *SelectStatement) String() string {
	var parts []string
	
	if s.SelectClause != nil {
		parts = append(parts, s.SelectClause.String())
	}
	if s.FromClause != nil {
		parts = append(parts, s.FromClause.String())
	}
	if s.WhereClause != nil {
		parts = append(parts, s.WhereClause.String())
	}
	if s.GroupBy != nil {
		parts = append(parts, s.GroupBy.String())
	}
	if s.Having != nil {
		parts = append(parts, s.Having.String())
	}
	if s.OrderBy != nil {
		parts = append(parts, s.OrderBy.String())
	}
	if s.Limit != nil {
		parts = append(parts, s.Limit.String())
	}
	
	return strings.Join(parts, " ")
}

// InsertStatement represents an INSERT statement
type InsertStatement struct {
	TableName *Identifier
	Columns   []*Identifier
	Values    [][]Expression
}

func (i *InsertStatement) StatementNode() {}
func (i *InsertStatement) NodeType() string { return "InsertStatement" }
func (i *InsertStatement) String() string {
	var result strings.Builder
	result.WriteString("INSERT INTO ")
	result.WriteString(i.TableName.String())
	
	if len(i.Columns) > 0 {
		result.WriteString(" (")
		for idx, col := range i.Columns {
			if idx > 0 {
				result.WriteString(", ")
			}
			result.WriteString(col.String())
		}
		result.WriteString(")")
	}
	
	result.WriteString(" VALUES ")
	for idx, valueSet := range i.Values {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString("(")
		for vidx, val := range valueSet {
			if vidx > 0 {
				result.WriteString(", ")
			}
			result.WriteString(val.String())
		}
		result.WriteString(")")
	}
	
	return result.String()
}

// UpdateStatement represents an UPDATE statement
type UpdateStatement struct {
	TableName *Identifier
	SetClauses []*SetClause
	WhereClause *WhereClause
}

func (u *UpdateStatement) StatementNode() {}
func (u *UpdateStatement) NodeType() string { return "UpdateStatement" }
func (u *UpdateStatement) String() string {
	var result strings.Builder
	result.WriteString("UPDATE ")
	result.WriteString(u.TableName.String())
	result.WriteString(" SET ")
	
	for idx, setClause := range u.SetClauses {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(setClause.String())
	}
	
	if u.WhereClause != nil {
		result.WriteString(" ")
		result.WriteString(u.WhereClause.String())
	}
	
	return result.String()
}

// DeleteStatement represents a DELETE statement
type DeleteStatement struct {
	TableName   *Identifier
	WhereClause *WhereClause
}

func (d *DeleteStatement) StatementNode() {}
func (d *DeleteStatement) NodeType() string { return "DeleteStatement" }
func (d *DeleteStatement) String() string {
	var result strings.Builder
	result.WriteString("DELETE FROM ")
	result.WriteString(d.TableName.String())
	
	if d.WhereClause != nil {
		result.WriteString(" ")
		result.WriteString(d.WhereClause.String())
	}
	
	return result.String()
}

// CreateTableStatement represents a CREATE TABLE statement
type CreateTableStatement struct {
	TableName *Identifier
	Columns   []*ColumnDefinition
	Constraints []*TableConstraint
}

func (c *CreateTableStatement) StatementNode() {}
func (c *CreateTableStatement) NodeType() string { return "CreateTableStatement" }
func (c *CreateTableStatement) String() string {
	var result strings.Builder
	result.WriteString("CREATE TABLE ")
	result.WriteString(c.TableName.String())
	result.WriteString(" (")
	
	for idx, col := range c.Columns {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(col.String())
	}
	
	for _, constraint := range c.Constraints {
		result.WriteString(", ")
		result.WriteString(constraint.String())
	}
	
	result.WriteString(")")
	return result.String()
}

// DropTableStatement represents a DROP TABLE statement
type DropTableStatement struct {
	TableName *Identifier
	IfExists  bool
}

func (d *DropTableStatement) StatementNode() {}
func (d *DropTableStatement) NodeType() string { return "DropTableStatement" }
func (d *DropTableStatement) String() string {
	var result strings.Builder
	result.WriteString("DROP TABLE ")
	if d.IfExists {
		result.WriteString("IF EXISTS ")
	}
	result.WriteString(d.TableName.String())
	return result.String()
}

// SelectClause represents the SELECT part of a query
type SelectClause struct {
	Distinct bool
	Columns  []Expression
}

func (s *SelectClause) NodeType() string { return "SelectClause" }
func (s *SelectClause) String() string {
	var result strings.Builder
	result.WriteString("SELECT ")
	if s.Distinct {
		result.WriteString("DISTINCT ")
	}
	
	for idx, col := range s.Columns {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(col.String())
	}
	
	return result.String()
}

// FromClause represents the FROM part of a query
type FromClause struct {
	Tables []Expression
	Joins  []*JoinClause
}

func (f *FromClause) NodeType() string { return "FromClause" }
func (f *FromClause) String() string {
	var result strings.Builder
	result.WriteString("FROM ")
	
	for idx, table := range f.Tables {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(table.String())
	}
	
	for _, join := range f.Joins {
		result.WriteString(" ")
		result.WriteString(join.String())
	}
	
	return result.String()
}

// WhereClause represents the WHERE part of a query
type WhereClause struct {
	Condition Expression
}

func (w *WhereClause) NodeType() string { return "WhereClause" }
func (w *WhereClause) String() string {
	return "WHERE " + w.Condition.String()
}

// GroupByClause represents the GROUP BY part of a query
type GroupByClause struct {
	Columns []Expression
}

func (g *GroupByClause) NodeType() string { return "GroupByClause" }
func (g *GroupByClause) String() string {
	var result strings.Builder
	result.WriteString("GROUP BY ")
	
	for idx, col := range g.Columns {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(col.String())
	}
	
	return result.String()
}

// HavingClause represents the HAVING part of a query
type HavingClause struct {
	Condition Expression
}

func (h *HavingClause) NodeType() string { return "HavingClause" }
func (h *HavingClause) String() string {
	return "HAVING " + h.Condition.String()
}

// OrderByClause represents the ORDER BY part of a query
type OrderByClause struct {
	Orders []*OrderExpression
}

func (o *OrderByClause) NodeType() string { return "OrderByClause" }
func (o *OrderByClause) String() string {
	var result strings.Builder
	result.WriteString("ORDER BY ")
	
	for idx, order := range o.Orders {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(order.String())
	}
	
	return result.String()
}

// LimitClause represents the LIMIT part of a query
type LimitClause struct {
	Count  Expression
	Offset Expression
}

func (l *LimitClause) NodeType() string { return "LimitClause" }
func (l *LimitClause) String() string {
	result := "LIMIT " + l.Count.String()
	if l.Offset != nil {
		result += " OFFSET " + l.Offset.String()
	}
	return result
}

// JoinClause represents JOIN operations
type JoinClause struct {
	JoinType  JoinType
	Table     Expression
	Condition Expression
}

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullJoin
)

func (j JoinType) String() string {
	switch j {
	case InnerJoin:
		return "INNER JOIN"
	case LeftJoin:
		return "LEFT JOIN"
	case RightJoin:
		return "RIGHT JOIN"
	case FullJoin:
		return "FULL JOIN"
	default:
		return "JOIN"
	}
}

func (j *JoinClause) NodeType() string { return "JoinClause" }
func (j *JoinClause) String() string {
	result := j.JoinType.String() + " " + j.Table.String()
	if j.Condition != nil {
		result += " ON " + j.Condition.String()
	}
	return result
}

// SetClause represents SET operations in UPDATE statements
type SetClause struct {
	Column *Identifier
	Value  Expression
}

func (s *SetClause) NodeType() string { return "SetClause" }
func (s *SetClause) String() string {
	return s.Column.String() + " = " + s.Value.String()
}

// OrderExpression represents ordering in ORDER BY
type OrderExpression struct {
	Expression Expression
	Direction  OrderDirection
}

type OrderDirection int

const (
	Ascending OrderDirection = iota
	Descending
)

func (o OrderDirection) String() string {
	switch o {
	case Ascending:
		return "ASC"
	case Descending:
		return "DESC"
	default:
		return "ASC"
	}
}

func (o *OrderExpression) NodeType() string { return "OrderExpression" }
func (o *OrderExpression) String() string {
	return o.Expression.String() + " " + o.Direction.String()
}

// ColumnDefinition represents column definitions in CREATE TABLE
type ColumnDefinition struct {
	Name         *Identifier
	DataType     *DataType
	Constraints  []*ColumnConstraint
}

func (c *ColumnDefinition) NodeType() string { return "ColumnDefinition" }
func (c *ColumnDefinition) String() string {
	result := c.Name.String() + " " + c.DataType.String()
	for _, constraint := range c.Constraints {
		result += " " + constraint.String()
	}
	return result
}

// DataType represents SQL data types
type DataType struct {
	Name   string
	Length int
	Precision int
	Scale  int
}

func (d *DataType) NodeType() string { return "DataType" }
func (d *DataType) String() string {
	result := d.Name
	if d.Length > 0 {
		result += fmt.Sprintf("(%d)", d.Length)
	} else if d.Precision > 0 {
		if d.Scale > 0 {
			result += fmt.Sprintf("(%d,%d)", d.Precision, d.Scale)
		} else {
			result += fmt.Sprintf("(%d)", d.Precision)
		}
	}
	return result
}

// ColumnConstraint represents column-level constraints
type ColumnConstraint struct {
	Type        ConstraintType
	Name        *Identifier
	References  *ForeignKeyReference
	DefaultValue Expression
}

// TableConstraint represents table-level constraints
type TableConstraint struct {
	Type       ConstraintType
	Name       *Identifier
	Columns    []*Identifier
	References *ForeignKeyReference
}

func (t *TableConstraint) NodeType() string { return "TableConstraint" }
func (t *TableConstraint) String() string {
	var result strings.Builder
	
	if t.Name != nil {
		result.WriteString("CONSTRAINT ")
		result.WriteString(t.Name.String())
		result.WriteString(" ")
	}
	
	switch t.Type {
	case PrimaryKey:
		result.WriteString("PRIMARY KEY (")
		for idx, col := range t.Columns {
			if idx > 0 {
				result.WriteString(", ")
			}
			result.WriteString(col.String())
		}
		result.WriteString(")")
	case ForeignKey:
		result.WriteString("FOREIGN KEY (")
		for idx, col := range t.Columns {
			if idx > 0 {
				result.WriteString(", ")
			}
			result.WriteString(col.String())
		}
		result.WriteString(") ")
		result.WriteString(t.References.String())
	case UniqueKey:
		result.WriteString("UNIQUE (")
		for idx, col := range t.Columns {
			if idx > 0 {
				result.WriteString(", ")
			}
			result.WriteString(col.String())
		}
		result.WriteString(")")
	}
	
	return result.String()
}

type ConstraintType int

const (
	NotNull ConstraintType = iota
	PrimaryKey
	UniqueKey
	ForeignKey
	Check
	Default
)

func (c *ColumnConstraint) NodeType() string { return "ColumnConstraint" }
func (c *ColumnConstraint) String() string {
	switch c.Type {
	case NotNull:
		return "NOT NULL"
	case PrimaryKey:
		return "PRIMARY KEY"
	case UniqueKey:
		return "UNIQUE"
	case ForeignKey:
		return "REFERENCES " + c.References.String()
	case Default:
		return "DEFAULT " + c.DefaultValue.String()
	default:
		return ""
	}
}

// ForeignKeyReference represents foreign key references
type ForeignKeyReference struct {
	Table   *Identifier
	Columns []*Identifier
}

func (f *ForeignKeyReference) NodeType() string { return "ForeignKeyReference" }
func (f *ForeignKeyReference) String() string {
	result := "REFERENCES " + f.Table.String()
	if len(f.Columns) > 0 {
		result += " ("
		for idx, col := range f.Columns {
			if idx > 0 {
				result += ", "
			}
			result += col.String()
		}
		result += ")"
	}
	return result
}

// Expression types

// Identifier represents table names, column names, etc.
type Identifier struct {
	Value string
	Alias *Identifier
}

func (i *Identifier) ExpressionNode() {}
func (i *Identifier) NodeType() string { return "Identifier" }
func (i *Identifier) String() string {
	result := i.Value
	if i.Alias != nil {
		result += " AS " + i.Alias.Value
	}
	return result
}

// Literal represents literal values (strings, numbers, etc.)
type Literal struct {
	Value interface{}
	Type  lexer.TokenType
}

func (l *Literal) ExpressionNode() {}
func (l *Literal) NodeType() string { return "Literal" }
func (l *Literal) String() string {
	switch l.Type {
	case lexer.STRING:
		return fmt.Sprintf("'%v'", l.Value)
	case lexer.NUMBER:
		return fmt.Sprintf("%v", l.Value)
	default:
		return fmt.Sprintf("%v", l.Value)
	}
}

// BinaryExpression represents binary operations (=, <, AND, OR, etc.)
type BinaryExpression struct {
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (b *BinaryExpression) ExpressionNode() {}
func (b *BinaryExpression) NodeType() string { return "BinaryExpression" }
func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator.String(), b.Right.String())
}

type BinaryOperator int

const (
	Equal BinaryOperator = iota
	NotEqual
	LessThan
	GreaterThan
	LessEqual
	GreaterEqual
	Like
	In
	Between
	And
	Or
	Plus
	Minus
	Multiply
	Divide
	Modulo
)

func (b BinaryOperator) String() string {
	switch b {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case LessThan:
		return "<"
	case GreaterThan:
		return ">"
	case LessEqual:
		return "<="
	case GreaterEqual:
		return ">="
	case Like:
		return "LIKE"
	case In:
		return "IN"
	case Between:
		return "BETWEEN"
	case And:
		return "AND"
	case Or:
		return "OR"
	case Plus:
		return "+"
	case Minus:
		return "-"
	case Multiply:
		return "*"
	case Divide:
		return "/"
	case Modulo:
		return "%"
	default:
		return "UNKNOWN"
	}
}

// UnaryExpression represents unary operations (NOT, -, etc.)
type UnaryExpression struct {
	Operator UnaryOperator
	Operand  Expression
}

func (u *UnaryExpression) ExpressionNode() {}
func (u *UnaryExpression) NodeType() string { return "UnaryExpression" }
func (u *UnaryExpression) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator.String(), u.Operand.String())
}

type UnaryOperator int

const (
	Not UnaryOperator = iota
	UnaryMinus
	UnaryPlus
)

func (u UnaryOperator) String() string {
	switch u {
	case Not:
		return "NOT "
	case UnaryMinus:
		return "-"
	case UnaryPlus:
		return "+"
	default:
		return "UNKNOWN"
	}
}

// FunctionCall represents function calls
type FunctionCall struct {
	Name      *Identifier
	Arguments []Expression
	Distinct  bool
}

func (f *FunctionCall) ExpressionNode() {}
func (f *FunctionCall) NodeType() string { return "FunctionCall" }
func (f *FunctionCall) String() string {
	var result strings.Builder
	result.WriteString(f.Name.String())
	result.WriteString("(")
	if f.Distinct {
		result.WriteString("DISTINCT ")
	}
	for idx, arg := range f.Arguments {
		if idx > 0 {
			result.WriteString(", ")
		}
		result.WriteString(arg.String())
	}
	result.WriteString(")")
	return result.String()
}

// ColumnReference represents qualified column references (table.column)
type ColumnReference struct {
	Table  *Identifier
	Column *Identifier
}

func (c *ColumnReference) ExpressionNode() {}
func (c *ColumnReference) NodeType() string { return "ColumnReference" }
func (c *ColumnReference) String() string {
	if c.Table != nil {
		return c.Table.String() + "." + c.Column.String()
	}
	return c.Column.String()
}

// Wildcard represents * in SELECT clauses
type Wildcard struct {
	Table *Identifier
}

func (w *Wildcard) ExpressionNode() {}
func (w *Wildcard) NodeType() string { return "Wildcard" }
func (w *Wildcard) String() string {
	if w.Table != nil {
		return w.Table.String() + ".*"
	}
	return "*"
}