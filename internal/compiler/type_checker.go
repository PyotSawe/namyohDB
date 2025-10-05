package compiler

import (
	"fmt"
	"relational-db/internal/lexer"
	"relational-db/internal/parser"
	"strings"
)

// TypeChecker performs type checking and inference
type TypeChecker struct {
	refs     *ResolvedReferences
	typeInfo *TypeInformation
}

// NewTypeChecker creates a new type checker
func NewTypeChecker(refs *ResolvedReferences, typeInfo *TypeInformation) *TypeChecker {
	return &TypeChecker{
		refs:     refs,
		typeInfo: typeInfo,
	}
}

// CheckSelect checks types in a SELECT statement
func (tc *TypeChecker) CheckSelect(stmt *parser.SelectStatement) error {
	// Type check all SELECT columns
	if stmt.SelectClause != nil {
		for _, col := range stmt.SelectClause.Columns {
			if _, err := tc.inferExpressionType(col); err != nil {
				return err
			}
		}
	}

	// Type check WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		whereType, err := tc.inferExpressionType(stmt.WhereClause.Condition)
		if err != nil {
			return err
		}

		// WHERE must be BOOLEAN
		if whereType != DataTypeBoolean && whereType != DataTypeUnknown {
			return fmt.Errorf("WHERE clause must be boolean, got %s", whereType)
		}
	}

	// Type check HAVING clause
	if stmt.Having != nil && stmt.Having.Condition != nil {
		havingType, err := tc.inferExpressionType(stmt.Having.Condition)
		if err != nil {
			return err
		}

		// HAVING must be BOOLEAN
		if havingType != DataTypeBoolean && havingType != DataTypeUnknown {
			return fmt.Errorf("HAVING clause must be boolean, got %s", havingType)
		}
	}

	return nil
}

// CheckInsert checks types in an INSERT statement
func (tc *TypeChecker) CheckInsert(stmt *parser.InsertStatement) error {
	// TODO: Implement INSERT type checking
	// Check that value types match column types
	return nil
}

// CheckUpdate checks types in an UPDATE statement
func (tc *TypeChecker) CheckUpdate(stmt *parser.UpdateStatement) error {
	// Type check SET clause values
	for _, setClause := range stmt.SetClauses {
		if _, err := tc.inferExpressionType(setClause.Value); err != nil {
			return err
		}
	}

	// Type check WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		whereType, err := tc.inferExpressionType(stmt.WhereClause.Condition)
		if err != nil {
			return err
		}

		// WHERE must be BOOLEAN
		if whereType != DataTypeBoolean && whereType != DataTypeUnknown {
			return fmt.Errorf("WHERE clause must be boolean, got %s", whereType)
		}
	}

	return nil
}

// CheckDelete checks types in a DELETE statement
func (tc *TypeChecker) CheckDelete(stmt *parser.DeleteStatement) error {
	// Type check WHERE clause
	if stmt.WhereClause != nil && stmt.WhereClause.Condition != nil {
		whereType, err := tc.inferExpressionType(stmt.WhereClause.Condition)
		if err != nil {
			return err
		}

		// WHERE must be BOOLEAN
		if whereType != DataTypeBoolean && whereType != DataTypeUnknown {
			return fmt.Errorf("WHERE clause must be boolean, got %s", whereType)
		}
	}

	return nil
}

// CheckCreateTable checks types in a CREATE TABLE statement
func (tc *TypeChecker) CheckCreateTable(stmt *parser.CreateTableStatement) error {
	// Type check DEFAULT values in column definitions
	for _, colDef := range stmt.Columns {
		for _, constraint := range colDef.Constraints {
			if constraint.Type == parser.Default && constraint.DefaultValue != nil {
				// Infer type of default value
				defaultType, err := tc.inferExpressionType(constraint.DefaultValue)
				if err != nil {
					return fmt.Errorf("invalid DEFAULT value: %v", err)
				}

				// TODO: Check DEFAULT value type matches column type
				_ = defaultType
			}
		}
	}

	return nil
}

// inferExpressionType infers the type of an expression
func (tc *TypeChecker) inferExpressionType(expr parser.Expression) (DataType, error) {
	if expr == nil {
		return DataTypeUnknown, nil
	}

	switch e := expr.(type) {
	case *parser.Literal:
		return tc.inferLiteralType(e)

	case *parser.Identifier:
		return tc.inferIdentifierType(e)

	case *parser.ColumnReference:
		return tc.inferColumnReferenceType(e)

	case *parser.BinaryExpression:
		return tc.inferBinaryType(e)

	case *parser.UnaryExpression:
		return tc.inferUnaryType(e)

	case *parser.FunctionCall:
		return tc.inferFunctionType(e)

	case *parser.Wildcard:
		// Wildcard doesn't have a specific type
		return DataTypeUnknown, nil

	default:
		return DataTypeUnknown, nil
	}
}

// inferLiteralType infers the type of a literal
func (tc *TypeChecker) inferLiteralType(lit *parser.Literal) (DataType, error) {
	switch lit.Type {
	case lexer.NUMBER:
		// Check if it's an integer or real number based on value
		switch lit.Value.(type) {
		case int, int64, int32:
			return DataTypeInteger, nil
		case float64, float32:
			return DataTypeReal, nil
		case string:
			// Parse string to determine if integer or float
			// For now, assume integer (more sophisticated parsing needed)
			return DataTypeInteger, nil
		default:
			return DataTypeNumeric, nil
		}

	case lexer.STRING:
		return DataTypeText, nil

	case lexer.NULL:
		return DataTypeNull, nil

	default:
		// Check if it's a boolean value by examining the value
		if v, ok := lit.Value.(bool); ok {
			if v || !v { // Just to use v
				return DataTypeBoolean, nil
			}
		}
		// Check string representation for TRUE/FALSE
		if v, ok := lit.Value.(string); ok {
			upper := strings.ToUpper(v)
			if upper == "TRUE" || upper == "FALSE" {
				return DataTypeBoolean, nil
			}
		}
		return DataTypeUnknown, fmt.Errorf("unknown literal type: %v", lit.Type)
	}
}

// inferIdentifierType infers the type of an identifier (column reference)
func (tc *TypeChecker) inferIdentifierType(ident *parser.Identifier) (DataType, error) {
	// Try to find the column in resolved references
	// Search all qualified names for matching column name
	columnName := ident.Value

	for qualifiedName, col := range tc.refs.Columns {
		// Check if this qualified name ends with our column name
		if qualifiedName == columnName || endsWithColumnName(qualifiedName, columnName) {
			return col.DataType, nil
		}
	}

	// Column not found - return unknown (might be resolved later)
	return DataTypeUnknown, nil
}

// inferColumnReferenceType infers the type of a qualified column reference
func (tc *TypeChecker) inferColumnReferenceType(colRef *parser.ColumnReference) (DataType, error) {
	tableName := ""
	if colRef.Table != nil {
		tableName = colRef.Table.Value
	}
	columnName := colRef.Column.Value

	// Build qualified name
	qualifiedName := columnName
	if tableName != "" {
		qualifiedName = tableName + "." + columnName
	}

	// Look up in resolved columns
	if col, found := tc.refs.Columns[qualifiedName]; found {
		return col.DataType, nil
	}

	// Try without table prefix
	for qn, col := range tc.refs.Columns {
		if endsWithColumnName(qn, columnName) {
			return col.DataType, nil
		}
	}

	return DataTypeUnknown, fmt.Errorf("column not found: %s", qualifiedName)
}

// endsWithColumnName checks if qualified name ends with column name
func endsWithColumnName(qualifiedName, columnName string) bool {
	if len(qualifiedName) < len(columnName) {
		return false
	}
	if qualifiedName == columnName {
		return true
	}
	// Check if it's "table.columnName"
	if len(qualifiedName) > len(columnName) && qualifiedName[len(qualifiedName)-len(columnName)-1] == '.' {
		return qualifiedName[len(qualifiedName)-len(columnName):] == columnName
	}
	return false
}

// inferBinaryType infers the type of a binary expression
func (tc *TypeChecker) inferBinaryType(bin *parser.BinaryExpression) (DataType, error) {
	leftType, err := tc.inferExpressionType(bin.Left)
	if err != nil {
		return DataTypeUnknown, err
	}

	rightType, err := tc.inferExpressionType(bin.Right)
	if err != nil {
		return DataTypeUnknown, err
	}

	// Check operator type
	switch bin.Operator {
	case parser.Plus, parser.Minus, parser.Multiply, parser.Divide, parser.Modulo:
		// Arithmetic operators
		return tc.inferArithmeticType(leftType, rightType)

	case parser.Equal, parser.NotEqual, parser.LessThan, parser.GreaterThan, parser.LessEqual, parser.GreaterEqual:
		// Comparison operators
		if leftType.IsComparable(rightType) {
			return DataTypeBoolean, nil
		}
		return DataTypeUnknown, fmt.Errorf("incomparable types: %s and %s", leftType, rightType)

	case parser.And, parser.Or:
		// Logical operators
		if leftType == DataTypeBoolean && rightType == DataTypeBoolean {
			return DataTypeBoolean, nil
		}
		if leftType == DataTypeUnknown || rightType == DataTypeUnknown {
			return DataTypeBoolean, nil // Assume boolean
		}
		return DataTypeUnknown, fmt.Errorf("logical operators require boolean operands")

	case parser.Like:
		// String comparison
		return DataTypeBoolean, nil

	case parser.In:
		// IN operator
		return DataTypeBoolean, nil

	default:
		return DataTypeUnknown, fmt.Errorf("unknown operator: %v", bin.Operator)
	}
}

// inferArithmeticType infers the result type of arithmetic operations
func (tc *TypeChecker) inferArithmeticType(left, right DataType) (DataType, error) {
	// INTEGER op INTEGER = INTEGER
	if left == DataTypeInteger && right == DataTypeInteger {
		return DataTypeInteger, nil
	}

	// Any numeric op REAL = REAL
	if (left.IsNumeric() || left == DataTypeUnknown) && right == DataTypeReal {
		return DataTypeReal, nil
	}
	if left == DataTypeReal && (right.IsNumeric() || right == DataTypeUnknown) {
		return DataTypeReal, nil
	}

	// REAL op REAL = REAL
	if left == DataTypeReal && right == DataTypeReal {
		return DataTypeReal, nil
	}

	// Unknown types - assume numeric for now
	if left == DataTypeUnknown || right == DataTypeUnknown {
		return DataTypeUnknown, nil
	}

	return DataTypeUnknown, fmt.Errorf("arithmetic operation on non-numeric types: %s and %s", left, right)
}

// inferUnaryType infers the type of a unary expression
func (tc *TypeChecker) inferUnaryType(unary *parser.UnaryExpression) (DataType, error) {
	operandType, err := tc.inferExpressionType(unary.Operand)
	if err != nil {
		return DataTypeUnknown, err
	}

	switch unary.Operator {
	case parser.UnaryMinus, parser.UnaryPlus:
		// Numeric negation
		if operandType.IsNumeric() || operandType == DataTypeUnknown {
			return operandType, nil
		}
		return DataTypeUnknown, fmt.Errorf("unary minus/plus requires numeric operand")

	case parser.Not:
		// Logical NOT
		if operandType == DataTypeBoolean || operandType == DataTypeUnknown {
			return DataTypeBoolean, nil
		}
		return DataTypeUnknown, fmt.Errorf("NOT requires boolean operand")

	default:
		return DataTypeUnknown, fmt.Errorf("unknown unary operator: %v", unary.Operator)
	}
}

// inferFunctionType infers the return type of a function call
func (tc *TypeChecker) inferFunctionType(fn *parser.FunctionCall) (DataType, error) {
	funcName := fn.Name.Value

	// Built-in aggregate functions
	switch funcName {
	case "COUNT":
		return DataTypeInteger, nil
	case "SUM":
		// Return type depends on argument type
		if len(fn.Arguments) > 0 {
			argType, err := tc.inferExpressionType(fn.Arguments[0])
			if err != nil {
				return DataTypeUnknown, err
			}
			if argType.IsNumeric() {
				return argType, nil
			}
		}
		return DataTypeNumeric, nil
	case "AVG":
		return DataTypeReal, nil
	case "MAX", "MIN":
		// Return type depends on argument type
		if len(fn.Arguments) > 0 {
			return tc.inferExpressionType(fn.Arguments[0])
		}
		return DataTypeUnknown, nil
	case "UPPER", "LOWER", "TRIM", "LTRIM", "RTRIM":
		return DataTypeText, nil
	case "LENGTH", "CHAR_LENGTH":
		return DataTypeInteger, nil
	default:
		return DataTypeUnknown, fmt.Errorf("unknown function: %s", funcName)
	}
}
