package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a SQL token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	WHITESPACE
	COMMENT

	// Literals
	IDENTIFIER
	NUMBER
	STRING

	// Keywords
	SELECT
	FROM
	WHERE
	INSERT
	INTO
	VALUES
	UPDATE
	SET
	DELETE
	CREATE
	TABLE
	DROP
	ALTER
	INDEX
	PRIMARY
	KEY
	FOREIGN
	REFERENCES
	NOT
	NULL
	UNIQUE
	DEFAULT
	AUTO_INCREMENT
	CONSTRAINT

	// Data types
	INTEGER
	TEXT
	REAL
	BLOB
	BOOLEAN

	// Operators
	EQUALS      // =
	NOT_EQUALS  // !=, <>
	LESS_THAN   // <
	GREATER_THAN // >
	LESS_EQUAL  // <=
	GREATER_EQUAL // >=
	LIKE        // LIKE
	IN          // IN
	BETWEEN     // BETWEEN
	IS          // IS

	// Logical operators
	AND
	OR

	// Punctuation
	SEMICOLON // ;
	COMMA     // ,
	DOT       // .
	LPAREN    // (
	RPAREN    // )
	LBRACKET  // [
	RBRACKET  // ]

	// Arithmetic operators
	PLUS     // +
	MINUS    // -
	MULTIPLY // *
	DIVIDE   // /
	MODULO   // %

	// Aggregate functions
	COUNT
	SUM
	AVG
	MIN
	MAX

	// Other keywords
	ORDER
	BY
	ASC
	DESC
	GROUP
	HAVING
	LIMIT
	OFFSET
	JOIN
	INNER
	LEFT
	RIGHT
	OUTER
	ON
	AS
	DISTINCT
	ALL
	IF
	EXISTS
)

// Token represents a single token in the SQL statement
type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WHITESPACE:
		return "WHITESPACE"
	case COMMENT:
		return "COMMENT"
	case IDENTIFIER:
		return "IDENTIFIER"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case SELECT:
		return "SELECT"
	case FROM:
		return "FROM"
	case WHERE:
		return "WHERE"
	case INSERT:
		return "INSERT"
	case INTO:
		return "INTO"
	case VALUES:
		return "VALUES"
	case UPDATE:
		return "UPDATE"
	case SET:
		return "SET"
	case DELETE:
		return "DELETE"
	case CREATE:
		return "CREATE"
	case TABLE:
		return "TABLE"
	case DROP:
		return "DROP"
	case EQUALS:
		return "EQUALS"
	case SEMICOLON:
		return "SEMICOLON"
	case COMMA:
		return "COMMA"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	default:
		return fmt.Sprintf("TokenType(%d)", int(t))
	}
}

// Keywords maps SQL keywords to their token types
var Keywords = map[string]TokenType{
	"SELECT":         SELECT,
	"FROM":           FROM,
	"WHERE":          WHERE,
	"INSERT":         INSERT,
	"INTO":           INTO,
	"VALUES":         VALUES,
	"UPDATE":         UPDATE,
	"SET":            SET,
	"DELETE":         DELETE,
	"CREATE":         CREATE,
	"TABLE":          TABLE,
	"DROP":           DROP,
	"ALTER":          ALTER,
	"INDEX":          INDEX,
	"PRIMARY":        PRIMARY,
	"KEY":            KEY,
	"FOREIGN":        FOREIGN,
	"REFERENCES":     REFERENCES,
	"NOT":            NOT,
	"NULL":           NULL,
	"UNIQUE":         UNIQUE,
	"DEFAULT":        DEFAULT,
	"AUTO_INCREMENT": AUTO_INCREMENT,
	"CONSTRAINT":     CONSTRAINT,
	"INTEGER":        INTEGER,
	"TEXT":           TEXT,
	"REAL":           REAL,
	"BLOB":           BLOB,
	"BOOLEAN":        BOOLEAN,
	"AND":            AND,
	"OR":             OR,
	"LIKE":           LIKE,
	"IN":             IN,
	"BETWEEN":        BETWEEN,
	"IS":             IS,
	"COUNT":          COUNT,
	"SUM":            SUM,
	"AVG":            AVG,
	"MIN":            MIN,
	"MAX":            MAX,
	"ORDER":          ORDER,
	"BY":             BY,
	"ASC":            ASC,
	"DESC":           DESC,
	"GROUP":          GROUP,
	"HAVING":         HAVING,
	"LIMIT":          LIMIT,
	"OFFSET":         OFFSET,
	"JOIN":           JOIN,
	"INNER":          INNER,
	"LEFT":           LEFT,
	"RIGHT":          RIGHT,
	"OUTER":          OUTER,
	"ON":             ON,
	"AS":             AS,
	"DISTINCT":       DISTINCT,
	"ALL":            ALL,
	"IF":             IF,
	"EXISTS":         EXISTS,
}

// Lexer represents the lexical analyzer
type Lexer struct {
	input    string
	position int
	current  rune
	line     int
	column   int
}

// NewLexer creates a new lexer instance
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.position >= len(l.input) {
		l.current = 0 // ASCII NUL represents EOF
	} else {
		l.current = rune(l.input[l.position])
	}
	l.position++
	if l.current == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position])
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var token Token

	l.skipWhitespace()

	token.Position = l.position - 1
	token.Line = l.line
	token.Column = l.column

	switch l.current {
	case '=':
		token = Token{Type: EQUALS, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '!':
		if l.peekChar() == '=' {
			token = Token{Type: NOT_EQUALS, Value: "!=", Position: token.Position, Line: token.Line, Column: token.Column}
			l.readChar()
		} else {
			token = Token{Type: ILLEGAL, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	case '<':
		if l.peekChar() == '=' {
			token = Token{Type: LESS_EQUAL, Value: "<=", Position: token.Position, Line: token.Line, Column: token.Column}
			l.readChar()
		} else if l.peekChar() == '>' {
			token = Token{Type: NOT_EQUALS, Value: "<>", Position: token.Position, Line: token.Line, Column: token.Column}
			l.readChar()
		} else {
			token = Token{Type: LESS_THAN, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	case '>':
		if l.peekChar() == '=' {
			token = Token{Type: GREATER_EQUAL, Value: ">=", Position: token.Position, Line: token.Line, Column: token.Column}
			l.readChar()
		} else {
			token = Token{Type: GREATER_THAN, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	case ';':
		token = Token{Type: SEMICOLON, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case ',':
		token = Token{Type: COMMA, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '.':
		token = Token{Type: DOT, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '(':
		token = Token{Type: LPAREN, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case ')':
		token = Token{Type: RPAREN, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '[':
		token = Token{Type: LBRACKET, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case ']':
		token = Token{Type: RBRACKET, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '+':
		token = Token{Type: PLUS, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '-':
		if l.peekChar() == '-' {
			// Single-line comment
			token = l.readComment()
		} else {
			token = Token{Type: MINUS, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	case '*':
		token = Token{Type: MULTIPLY, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '/':
		if l.peekChar() == '*' {
			// Multi-line comment
			token = l.readMultiLineComment()
		} else {
			token = Token{Type: DIVIDE, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	case '%':
		token = Token{Type: MODULO, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
	case '\'', '"':
		token = l.readString()
	case '`':
		token = l.readIdentifier()
	case 0:
		token = Token{Type: EOF, Value: "", Position: token.Position, Line: token.Line, Column: token.Column}
	default:
		if isLetter(l.current) {
			token = l.readIdentifierOrKeyword()
		} else if isDigit(l.current) {
			token = l.readNumber()
		} else {
			token = Token{Type: ILLEGAL, Value: string(l.current), Position: token.Position, Line: token.Line, Column: token.Column}
		}
	}

	l.readChar()
	return token
}

// readIdentifierOrKeyword reads an identifier or keyword
func (l *Lexer) readIdentifierOrKeyword() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	var identifier strings.Builder
	for isLetter(l.current) || isDigit(l.current) || l.current == '_' {
		identifier.WriteRune(l.current)
		l.readChar()
	}
	l.position-- // Back up one character
	l.column--

	value := identifier.String()
	tokenType := lookupIdentifier(value)

	return Token{
		Type:     tokenType,
		Value:    value,
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// readIdentifier reads a quoted identifier
func (l *Lexer) readIdentifier() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	quote := l.current
	l.readChar() // Skip opening quote

	var identifier strings.Builder
	for l.current != quote && l.current != 0 {
		if l.current == '\\' {
			l.readChar()
			if l.current != 0 {
				identifier.WriteRune(l.current)
			}
		} else {
			identifier.WriteRune(l.current)
		}
		l.readChar()
	}

	return Token{
		Type:     IDENTIFIER,
		Value:    identifier.String(),
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// readString reads a string literal
func (l *Lexer) readString() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	quote := l.current
	l.readChar() // Skip opening quote

	var str strings.Builder
	for l.current != quote && l.current != 0 {
		if l.current == '\\' {
			l.readChar()
			switch l.current {
			case 'n':
				str.WriteRune('\n')
			case 't':
				str.WriteRune('\t')
			case 'r':
				str.WriteRune('\r')
			case '\\':
				str.WriteRune('\\')
			case '\'':
				str.WriteRune('\'')
			case '"':
				str.WriteRune('"')
			default:
				str.WriteRune(l.current)
			}
		} else {
			str.WriteRune(l.current)
		}
		l.readChar()
	}

	return Token{
		Type:     STRING,
		Value:    str.String(),
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// readNumber reads a numeric literal
func (l *Lexer) readNumber() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	var number strings.Builder
	for isDigit(l.current) {
		number.WriteRune(l.current)
		l.readChar()
	}

	// Check for decimal point
	if l.current == '.' && isDigit(l.peekChar()) {
		number.WriteRune(l.current)
		l.readChar()
		for isDigit(l.current) {
			number.WriteRune(l.current)
			l.readChar()
		}
	}

	// Check for scientific notation
	if l.current == 'e' || l.current == 'E' {
		number.WriteRune(l.current)
		l.readChar()
		if l.current == '+' || l.current == '-' {
			number.WriteRune(l.current)
			l.readChar()
		}
		for isDigit(l.current) {
			number.WriteRune(l.current)
			l.readChar()
		}
	}

	l.position-- // Back up one character
	l.column--

	return Token{
		Type:     NUMBER,
		Value:    number.String(),
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// readComment reads a single-line comment
func (l *Lexer) readComment() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	var comment strings.Builder
	for l.current != '\n' && l.current != 0 {
		comment.WriteRune(l.current)
		l.readChar()
	}

	return Token{
		Type:     COMMENT,
		Value:    comment.String(),
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// readMultiLineComment reads a multi-line comment
func (l *Lexer) readMultiLineComment() Token {
	position := l.position - 1
	line := l.line
	column := l.column

	var comment strings.Builder
	l.readChar() // Skip '/'
	l.readChar() // Skip '*'

	for {
		if l.current == 0 {
			break
		}
		if l.current == '*' && l.peekChar() == '/' {
			l.readChar() // Skip '*'
			break
		}
		comment.WriteRune(l.current)
		l.readChar()
	}

	return Token{
		Type:     COMMENT,
		Value:    comment.String(),
		Position: position,
		Line:     line,
		Column:   column,
	}
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.current == ' ' || l.current == '\t' || l.current == '\n' || l.current == '\r' {
		l.readChar()
	}
}

// isLetter checks if a character is a letter
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit checks if a character is a digit
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// lookupIdentifier checks if an identifier is a keyword
func lookupIdentifier(identifier string) TokenType {
	if tokenType, exists := Keywords[strings.ToUpper(identifier)]; exists {
		return tokenType
	}
	return IDENTIFIER
}

// TokenizeSQL tokenizes a complete SQL statement and returns all tokens
func TokenizeSQL(sql string) ([]Token, error) {
	lexer := NewLexer(sql)
	var tokens []Token

	for {
		token := lexer.NextToken()
		if token.Type == EOF {
			tokens = append(tokens, token)
			break
		}
		if token.Type == ILLEGAL {
			return nil, fmt.Errorf("illegal token '%s' at position %d (line %d, column %d)",
				token.Value, token.Position, token.Line, token.Column)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}