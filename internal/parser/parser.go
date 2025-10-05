package parser

import (
	"fmt"
	"strconv"

	"relational-db/internal/lexer"
)

// Parser represents the SQL parser
type Parser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
	peekToken    lexer.Token
	errors       []string
}

// NewParser creates a new parser instance
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// Read two tokens, so currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances both currentToken and peekToken
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// Errors returns any parsing errors
func (p *Parser) Errors() []string {
	return p.errors
}

// addError adds an error message
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("Parse error at line %d, column %d: %s",
		p.currentToken.Line, p.currentToken.Column, msg))
}

// expectToken checks if current token matches expected type and advances
func (p *Parser) expectToken(expectedType lexer.TokenType) bool {
	if p.currentToken.Type != expectedType {
		p.addError(fmt.Sprintf("expected %s, got %s", expectedType.String(), p.currentToken.Type.String()))
		return false
	}
	p.nextToken()
	return true
}

// currentTokenIs checks if current token matches the given type
func (p *Parser) currentTokenIs(tokenType lexer.TokenType) bool {
	return p.currentToken.Type == tokenType
}

// peekTokenIs checks if peek token matches the given type
func (p *Parser) peekTokenIs(tokenType lexer.TokenType) bool {
	return p.peekToken.Type == tokenType
}

// ParseStatement parses a complete SQL statement
func (p *Parser) ParseStatement() Statement {
	switch p.currentToken.Type {
	case lexer.SELECT:
		return p.parseSelectStatement()
	case lexer.INSERT:
		return p.parseInsertStatement()
	case lexer.UPDATE:
		return p.parseUpdateStatement()
	case lexer.DELETE:
		return p.parseDeleteStatement()
	case lexer.CREATE:
		return p.parseCreateStatement()
	case lexer.DROP:
		return p.parseDropStatement()
	default:
		p.addError(fmt.Sprintf("unexpected token %s", p.currentToken.Type.String()))
		return nil
	}
}

// parseSelectStatement parses SELECT statements
func (p *Parser) parseSelectStatement() *SelectStatement {
	stmt := &SelectStatement{}

	// Parse SELECT clause
	stmt.SelectClause = p.parseSelectClause()
	if stmt.SelectClause == nil {
		return nil
	}

	// Parse FROM clause (optional)
	if p.currentTokenIs(lexer.FROM) {
		stmt.FromClause = p.parseFromClause()
	}

	// Parse WHERE clause (optional)
	if p.currentTokenIs(lexer.WHERE) {
		stmt.WhereClause = p.parseWhereClause()
	}

	// Parse GROUP BY clause (optional)
	if p.currentTokenIs(lexer.GROUP) {
		stmt.GroupBy = p.parseGroupByClause()
	}

	// Parse HAVING clause (optional)
	if p.currentTokenIs(lexer.HAVING) {
		stmt.Having = p.parseHavingClause()
	}

	// Parse ORDER BY clause (optional)
	if p.currentTokenIs(lexer.ORDER) {
		stmt.OrderBy = p.parseOrderByClause()
	}

	// Parse LIMIT clause (optional)
	if p.currentTokenIs(lexer.LIMIT) {
		stmt.Limit = p.parseLimitClause()
	}

	return stmt
}

// parseSelectClause parses the SELECT part
func (p *Parser) parseSelectClause() *SelectClause {
	if !p.expectToken(lexer.SELECT) {
		return nil
	}

	clause := &SelectClause{}

	// Check for DISTINCT
	if p.currentTokenIs(lexer.DISTINCT) {
		clause.Distinct = true
		p.nextToken()
	}

	// Parse column list
	for {
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		clause.Columns = append(clause.Columns, expr)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	return clause
}

// parseFromClause parses the FROM part
func (p *Parser) parseFromClause() *FromClause {
	if !p.expectToken(lexer.FROM) {
		return nil
	}

	clause := &FromClause{}

	// Parse table list
	for {
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		clause.Tables = append(clause.Tables, expr)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	// Parse JOINs
	for p.currentTokenIs(lexer.JOIN) || p.currentTokenIs(lexer.INNER) ||
		p.currentTokenIs(lexer.LEFT) || p.currentTokenIs(lexer.RIGHT) {

		join := p.parseJoinClause()
		if join == nil {
			return nil
		}
		clause.Joins = append(clause.Joins, join)
	}

	return clause
}

// parseWhereClause parses the WHERE part
func (p *Parser) parseWhereClause() *WhereClause {
	if !p.expectToken(lexer.WHERE) {
		return nil
	}

	condition := p.parseExpression()
	if condition == nil {
		return nil
	}

	return &WhereClause{Condition: condition}
}

// parseGroupByClause parses the GROUP BY part
func (p *Parser) parseGroupByClause() *GroupByClause {
	if !p.expectToken(lexer.GROUP) {
		return nil
	}
	if !p.expectToken(lexer.BY) {
		return nil
	}

	clause := &GroupByClause{}

	for {
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		clause.Columns = append(clause.Columns, expr)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	return clause
}

// parseHavingClause parses the HAVING part
func (p *Parser) parseHavingClause() *HavingClause {
	if !p.expectToken(lexer.HAVING) {
		return nil
	}

	condition := p.parseExpression()
	if condition == nil {
		return nil
	}

	return &HavingClause{Condition: condition}
}

// parseOrderByClause parses the ORDER BY part
func (p *Parser) parseOrderByClause() *OrderByClause {
	if !p.expectToken(lexer.ORDER) {
		return nil
	}
	if !p.expectToken(lexer.BY) {
		return nil
	}

	clause := &OrderByClause{}

	for {
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}

		direction := Ascending
		if p.currentTokenIs(lexer.ASC) {
			p.nextToken()
		} else if p.currentTokenIs(lexer.DESC) {
			direction = Descending
			p.nextToken()
		}

		order := &OrderExpression{
			Expression: expr,
			Direction:  direction,
		}
		clause.Orders = append(clause.Orders, order)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	return clause
}

// parseLimitClause parses the LIMIT part
func (p *Parser) parseLimitClause() *LimitClause {
	if !p.expectToken(lexer.LIMIT) {
		return nil
	}

	count := p.parseExpression()
	if count == nil {
		return nil
	}

	clause := &LimitClause{Count: count}

	// Parse optional OFFSET
	if p.currentTokenIs(lexer.OFFSET) {
		p.nextToken()
		offset := p.parseExpression()
		if offset == nil {
			return nil
		}
		clause.Offset = offset
	}

	return clause
}

// parseJoinClause parses JOIN operations
func (p *Parser) parseJoinClause() *JoinClause {
	joinType := InnerJoin

	if p.currentTokenIs(lexer.LEFT) {
		joinType = LeftJoin
		p.nextToken()
		if p.currentTokenIs(lexer.OUTER) {
			p.nextToken()
		}
	} else if p.currentTokenIs(lexer.RIGHT) {
		joinType = RightJoin
		p.nextToken()
		if p.currentTokenIs(lexer.OUTER) {
			p.nextToken()
		}
	} else if p.currentTokenIs(lexer.INNER) {
		p.nextToken()
	}

	if !p.expectToken(lexer.JOIN) {
		return nil
	}

	table := p.parseExpression()
	if table == nil {
		return nil
	}

	var condition Expression
	if p.currentTokenIs(lexer.ON) {
		p.nextToken()
		condition = p.parseExpression()
		if condition == nil {
			return nil
		}
	}

	return &JoinClause{
		JoinType:  joinType,
		Table:     table,
		Condition: condition,
	}
}

// parseInsertStatement parses INSERT statements
func (p *Parser) parseInsertStatement() *InsertStatement {
	if !p.expectToken(lexer.INSERT) {
		return nil
	}
	if !p.expectToken(lexer.INTO) {
		return nil
	}

	tableName := p.parseIdentifier()
	if tableName == nil {
		return nil
	}

	stmt := &InsertStatement{TableName: tableName}

	// Parse optional column list
	if p.currentTokenIs(lexer.LPAREN) {
		p.nextToken()
		for {
			col := p.parseIdentifier()
			if col == nil {
				return nil
			}
			stmt.Columns = append(stmt.Columns, col)

			if !p.currentTokenIs(lexer.COMMA) {
				break
			}
			p.nextToken() // consume comma
		}
		if !p.expectToken(lexer.RPAREN) {
			return nil
		}
	}

	// Parse VALUES clause
	if !p.expectToken(lexer.VALUES) {
		return nil
	}

	for {
		if !p.expectToken(lexer.LPAREN) {
			return nil
		}

		var values []Expression
		for {
			expr := p.parseExpression()
			if expr == nil {
				return nil
			}
			values = append(values, expr)

			if !p.currentTokenIs(lexer.COMMA) {
				break
			}
			p.nextToken() // consume comma
		}

		if !p.expectToken(lexer.RPAREN) {
			return nil
		}

		stmt.Values = append(stmt.Values, values)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	return stmt
}

// parseUpdateStatement parses UPDATE statements
func (p *Parser) parseUpdateStatement() *UpdateStatement {
	if !p.expectToken(lexer.UPDATE) {
		return nil
	}

	tableName := p.parseIdentifier()
	if tableName == nil {
		return nil
	}

	if !p.expectToken(lexer.SET) {
		return nil
	}

	stmt := &UpdateStatement{TableName: tableName}

	// Parse SET clauses
	for {
		col := p.parseIdentifier()
		if col == nil {
			return nil
		}

		if !p.expectToken(lexer.EQUALS) {
			return nil
		}

		value := p.parseExpression()
		if value == nil {
			return nil
		}

		setClause := &SetClause{
			Column: col,
			Value:  value,
		}
		stmt.SetClauses = append(stmt.SetClauses, setClause)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	// Parse optional WHERE clause
	if p.currentTokenIs(lexer.WHERE) {
		stmt.WhereClause = p.parseWhereClause()
	}

	return stmt
}

// parseDeleteStatement parses DELETE statements
func (p *Parser) parseDeleteStatement() *DeleteStatement {
	if !p.expectToken(lexer.DELETE) {
		return nil
	}
	if !p.expectToken(lexer.FROM) {
		return nil
	}

	tableName := p.parseIdentifier()
	if tableName == nil {
		return nil
	}

	stmt := &DeleteStatement{TableName: tableName}

	// Parse optional WHERE clause
	if p.currentTokenIs(lexer.WHERE) {
		stmt.WhereClause = p.parseWhereClause()
	}

	return stmt
}

// parseCreateStatement parses CREATE statements
func (p *Parser) parseCreateStatement() Statement {
	if !p.expectToken(lexer.CREATE) {
		return nil
	}

	if p.currentTokenIs(lexer.TABLE) {
		return p.parseCreateTableStatement()
	}

	p.addError("only CREATE TABLE is supported")
	return nil
}

// parseCreateTableStatement parses CREATE TABLE statements
func (p *Parser) parseCreateTableStatement() *CreateTableStatement {
	if !p.expectToken(lexer.TABLE) {
		return nil
	}

	tableName := p.parseIdentifier()
	if tableName == nil {
		return nil
	}

	if !p.expectToken(lexer.LPAREN) {
		return nil
	}

	stmt := &CreateTableStatement{TableName: tableName}

	for {
		// Parse column definition or table constraint
		if p.currentTokenIs(lexer.CONSTRAINT) || p.currentTokenIs(lexer.PRIMARY) ||
			p.currentTokenIs(lexer.FOREIGN) || p.currentTokenIs(lexer.UNIQUE) {
			// Table constraint
			constraint := p.parseTableConstraint()
			if constraint == nil {
				return nil
			}
			stmt.Constraints = append(stmt.Constraints, constraint)
		} else {
			// Column definition
			colDef := p.parseColumnDefinition()
			if colDef == nil {
				return nil
			}
			stmt.Columns = append(stmt.Columns, colDef)
		}

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	if !p.expectToken(lexer.RPAREN) {
		return nil
	}

	return stmt
}

// parseDropStatement parses DROP statements
func (p *Parser) parseDropStatement() Statement {
	if !p.expectToken(lexer.DROP) {
		return nil
	}

	if p.currentTokenIs(lexer.TABLE) {
		return p.parseDropTableStatement()
	}

	p.addError("only DROP TABLE is supported")
	return nil
}

// parseDropTableStatement parses DROP TABLE statements
func (p *Parser) parseDropTableStatement() *DropTableStatement {
	if !p.expectToken(lexer.TABLE) {
		return nil
	}

	stmt := &DropTableStatement{}

	// Check for IF EXISTS
	if p.currentTokenIs(lexer.IF) {
		p.nextToken()
		if p.currentTokenIs(lexer.EXISTS) {
			stmt.IfExists = true
			p.nextToken()
		} else {
			p.addError("expected EXISTS after IF")
			return nil
		}
	}

	tableName := p.parseIdentifier()
	if tableName == nil {
		return nil
	}
	stmt.TableName = tableName

	return stmt
}

// parseColumnDefinition parses column definitions
func (p *Parser) parseColumnDefinition() *ColumnDefinition {
	name := p.parseIdentifier()
	if name == nil {
		return nil
	}

	dataType := p.parseDataType()
	if dataType == nil {
		return nil
	}

	colDef := &ColumnDefinition{
		Name:     name,
		DataType: dataType,
	}

	// Parse column constraints
	for {
		constraint := p.parseColumnConstraint()
		if constraint == nil {
			break
		}
		colDef.Constraints = append(colDef.Constraints, constraint)
	}

	return colDef
}

// parseDataType parses SQL data types
func (p *Parser) parseDataType() *DataType {
	if !p.currentTokenIs(lexer.INTEGER) && !p.currentTokenIs(lexer.TEXT) &&
		!p.currentTokenIs(lexer.REAL) && !p.currentTokenIs(lexer.BLOB) &&
		!p.currentTokenIs(lexer.BOOLEAN) {
		p.addError("expected data type")
		return nil
	}

	dataType := &DataType{Name: p.currentToken.Value}
	p.nextToken()

	// Parse optional length/precision
	if p.currentTokenIs(lexer.LPAREN) {
		p.nextToken()

		if !p.currentTokenIs(lexer.NUMBER) {
			p.addError("expected number after (")
			return nil
		}

		length, err := strconv.Atoi(p.currentToken.Value)
		if err != nil {
			p.addError("invalid number")
			return nil
		}

		dataType.Length = length
		p.nextToken()

		// Check for precision (e.g., DECIMAL(10,2))
		if p.currentTokenIs(lexer.COMMA) {
			p.nextToken()
			if !p.currentTokenIs(lexer.NUMBER) {
				p.addError("expected number after comma")
				return nil
			}

			scale, err := strconv.Atoi(p.currentToken.Value)
			if err != nil {
				p.addError("invalid number")
				return nil
			}

			dataType.Precision = dataType.Length
			dataType.Scale = scale
			dataType.Length = 0
			p.nextToken()
		}

		if !p.expectToken(lexer.RPAREN) {
			return nil
		}
	}

	return dataType
}

// parseColumnConstraint parses column constraints
func (p *Parser) parseColumnConstraint() *ColumnConstraint {
	switch p.currentToken.Type {
	case lexer.NOT:
		p.nextToken()
		if p.expectToken(lexer.NULL) {
			return &ColumnConstraint{Type: NotNull}
		}
		return nil
	case lexer.PRIMARY:
		p.nextToken()
		if p.expectToken(lexer.KEY) {
			return &ColumnConstraint{Type: PrimaryKey}
		}
		return nil
	case lexer.UNIQUE:
		p.nextToken()
		return &ColumnConstraint{Type: UniqueKey}
	case lexer.DEFAULT:
		p.nextToken()
		value := p.parseExpression()
		if value == nil {
			p.addError("expected expression after DEFAULT")
			return nil
		}
		return &ColumnConstraint{Type: Default, DefaultValue: value}
	default:
		return nil
	}
}

// parseTableConstraint parses table constraints
func (p *Parser) parseTableConstraint() *TableConstraint {
	constraint := &TableConstraint{}

	// Parse optional constraint name
	if p.currentTokenIs(lexer.CONSTRAINT) {
		p.nextToken()
		constraint.Name = p.parseIdentifier()
		if constraint.Name == nil {
			return nil
		}
	}

	switch p.currentToken.Type {
	case lexer.PRIMARY:
		p.nextToken()
		if !p.expectToken(lexer.KEY) {
			return nil
		}
		constraint.Type = PrimaryKey
	case lexer.FOREIGN:
		p.nextToken()
		if !p.expectToken(lexer.KEY) {
			return nil
		}
		constraint.Type = ForeignKey
	case lexer.UNIQUE:
		p.nextToken()
		constraint.Type = UniqueKey
	default:
		p.addError("expected constraint type")
		return nil
	}

	// Parse column list
	if !p.expectToken(lexer.LPAREN) {
		return nil
	}

	for {
		col := p.parseIdentifier()
		if col == nil {
			return nil
		}
		constraint.Columns = append(constraint.Columns, col)

		if !p.currentTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
	}

	if !p.expectToken(lexer.RPAREN) {
		return nil
	}

	// Parse REFERENCES clause for foreign keys
	if constraint.Type == ForeignKey && p.currentTokenIs(lexer.REFERENCES) {
		p.nextToken()
		refTable := p.parseIdentifier()
		if refTable == nil {
			return nil
		}

		ref := &ForeignKeyReference{Table: refTable}

		if p.currentTokenIs(lexer.LPAREN) {
			p.nextToken()
			for {
				col := p.parseIdentifier()
				if col == nil {
					return nil
				}
				ref.Columns = append(ref.Columns, col)

				if !p.currentTokenIs(lexer.COMMA) {
					break
				}
				p.nextToken() // consume comma
			}
			if !p.expectToken(lexer.RPAREN) {
				return nil
			}
		}

		constraint.References = ref
	}

	return constraint
}

// parseExpression parses expressions with operator precedence
func (p *Parser) parseExpression() Expression {
	return p.parseLogicalOr()
}

// parseLogicalOr parses OR expressions (lowest precedence)
func (p *Parser) parseLogicalOr() Expression {
	left := p.parseLogicalAnd()
	if left == nil {
		return nil
	}

	for p.currentTokenIs(lexer.OR) {
		p.nextToken()
		right := p.parseLogicalAnd()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: Or, Right: right}
	}

	return left
}

// parseLogicalAnd parses AND expressions
func (p *Parser) parseLogicalAnd() Expression {
	left := p.parseComparison()
	if left == nil {
		return nil
	}

	for p.currentTokenIs(lexer.AND) {
		p.nextToken()
		right := p.parseComparison()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: And, Right: right}
	}

	return left
}

// parseComparison parses comparison expressions
func (p *Parser) parseComparison() Expression {
	left := p.parseAddition()
	if left == nil {
		return nil
	}

	for {
		var op BinaryOperator
		switch p.currentToken.Type {
		case lexer.EQUALS:
			op = Equal
		case lexer.NOT_EQUALS:
			op = NotEqual
		case lexer.LESS_THAN:
			op = LessThan
		case lexer.GREATER_THAN:
			op = GreaterThan
		case lexer.LESS_EQUAL:
			op = LessEqual
		case lexer.GREATER_EQUAL:
			op = GreaterEqual
		case lexer.LIKE:
			op = Like
		case lexer.IN:
			op = In
		case lexer.BETWEEN:
			op = Between
		default:
			return left
		}

		p.nextToken()
		right := p.parseAddition()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: op, Right: right}
	}
}

// parseAddition parses addition and subtraction
func (p *Parser) parseAddition() Expression {
	left := p.parseMultiplication()
	if left == nil {
		return nil
	}

	for p.currentTokenIs(lexer.PLUS) || p.currentTokenIs(lexer.MINUS) {
		var op BinaryOperator
		if p.currentTokenIs(lexer.PLUS) {
			op = Plus
		} else {
			op = Minus
		}
		p.nextToken()

		right := p.parseMultiplication()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: op, Right: right}
	}

	return left
}

// parseMultiplication parses multiplication, division, and modulo
func (p *Parser) parseMultiplication() Expression {
	left := p.parseUnary()
	if left == nil {
		return nil
	}

	for p.currentTokenIs(lexer.MULTIPLY) || p.currentTokenIs(lexer.DIVIDE) || p.currentTokenIs(lexer.MODULO) {
		var op BinaryOperator
		switch p.currentToken.Type {
		case lexer.MULTIPLY:
			op = Multiply
		case lexer.DIVIDE:
			op = Divide
		case lexer.MODULO:
			op = Modulo
		}
		p.nextToken()

		right := p.parseUnary()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: op, Right: right}
	}

	return left
}

// parseUnary parses unary expressions
func (p *Parser) parseUnary() Expression {
	switch p.currentToken.Type {
	case lexer.NOT:
		p.nextToken()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpression{Operator: Not, Operand: operand}
	case lexer.MINUS:
		p.nextToken()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpression{Operator: UnaryMinus, Operand: operand}
	case lexer.PLUS:
		p.nextToken()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpression{Operator: UnaryPlus, Operand: operand}
	default:
		return p.parsePrimary()
	}
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() Expression {
	switch p.currentToken.Type {
	case lexer.IDENTIFIER:
		return p.parseIdentifierExpression()
	case lexer.NUMBER:
		return p.parseNumberLiteral()
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.MULTIPLY:
		// Handle * (wildcard)
		p.nextToken()
		return &Wildcard{}
	case lexer.LPAREN:
		// Handle grouped expressions
		p.nextToken()
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		if !p.expectToken(lexer.RPAREN) {
			return nil
		}
		return expr
	default:
		p.addError(fmt.Sprintf("unexpected token %s", p.currentToken.Type.String()))
		return nil
	}
}

// parseIdentifierExpression parses identifiers, function calls, and column references
func (p *Parser) parseIdentifierExpression() Expression {
	name := p.parseIdentifier()
	if name == nil {
		return nil
	}

	// Check if it's a function call
	if p.currentTokenIs(lexer.LPAREN) {
		p.nextToken()

		funcCall := &FunctionCall{Name: name}

		// Check for DISTINCT
		if p.currentTokenIs(lexer.DISTINCT) {
			funcCall.Distinct = true
			p.nextToken()
		}

		// Parse arguments
		if !p.currentTokenIs(lexer.RPAREN) {
			for {
				arg := p.parseExpression()
				if arg == nil {
					return nil
				}
				funcCall.Arguments = append(funcCall.Arguments, arg)

				if !p.currentTokenIs(lexer.COMMA) {
					break
				}
				p.nextToken() // consume comma
			}
		}

		if !p.expectToken(lexer.RPAREN) {
			return nil
		}

		return funcCall
	}

	// Check if it's a qualified column reference (table.column)
	if p.currentTokenIs(lexer.DOT) {
		p.nextToken()
		if p.currentTokenIs(lexer.MULTIPLY) {
			// table.*
			p.nextToken()
			return &Wildcard{Table: name}
		}
		column := p.parseIdentifier()
		if column == nil {
			return nil
		}
		return &ColumnReference{Table: name, Column: column}
	}

	// Simple identifier
	return name
}

// parseIdentifier parses identifiers with optional aliases
func (p *Parser) parseIdentifier() *Identifier {
	if !p.currentTokenIs(lexer.IDENTIFIER) {
		p.addError("expected identifier")
		return nil
	}

	identifier := &Identifier{Value: p.currentToken.Value}
	p.nextToken()

	// Parse optional alias
	if p.currentTokenIs(lexer.AS) {
		p.nextToken()
		if !p.currentTokenIs(lexer.IDENTIFIER) {
			p.addError("expected identifier after AS")
			return nil
		}
		identifier.Alias = &Identifier{Value: p.currentToken.Value}
		p.nextToken()
	}

	return identifier
}

// parseNumberLiteral parses numeric literals
func (p *Parser) parseNumberLiteral() *Literal {
	value := p.currentToken.Value
	p.nextToken()
	return &Literal{Value: value, Type: lexer.NUMBER}
}

// parseStringLiteral parses string literals
func (p *Parser) parseStringLiteral() *Literal {
	value := p.currentToken.Value
	p.nextToken()
	return &Literal{Value: value, Type: lexer.STRING}
}

// ParseSQL is a convenience function to parse a complete SQL statement
func ParseSQL(sql string) (Statement, error) {
	lexer := lexer.NewLexer(sql)
	parser := NewParser(lexer)

	stmt := parser.ParseStatement()

	if len(parser.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %v", parser.Errors())
	}

	return stmt, nil
}
