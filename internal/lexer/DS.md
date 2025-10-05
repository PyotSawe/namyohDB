# Lexer Module Data Structures

## Overview
This document describes the data structures used in the SQL lexer module, including their design rationale, memory layout, and usage patterns.

## Core Data Structures

### 1. Lexer Structure

```go
type Lexer struct {
    input        string    // Input SQL string to tokenize
    position     int       // Current position in input (points to current char)
    readPosition int       // Current reading position (after current char)
    ch           byte      // Current character under examination
    line         int       // Current line number (1-indexed)
    column       int       // Current column number (1-indexed)
}
```

#### Design Rationale
- **String Input**: Immutable string storage for thread safety
- **Dual Positions**: `position` and `readPosition` enable single-character lookahead
- **Byte Processing**: Using `byte` instead of `rune` for ASCII optimization (can be extended for Unicode)
- **Position Tracking**: Line/column information for precise error reporting

#### Memory Layout
```
┌─────────────────────────────────────────────────────────┐
│                    Lexer (56 bytes)                     │
├─────────────────────────────────────────────────────────┤
│ input        │ 16 bytes │ string header (ptr + len)    │
│ position     │  8 bytes │ int                          │
│ readPosition │  8 bytes │ int                          │
│ ch           │  1 byte  │ byte                         │
│ line         │  8 bytes │ int                          │
│ column       │  8 bytes │ int                          │
│ padding      │  7 bytes │ struct alignment             │
└─────────────────────────────────────────────────────────┘
```

### 2. Token Structure

```go
type Token struct {
    Type     TokenType  // Classification of the token
    Literal  string     // The actual string value
    Position int        // Character position in input
    Line     int        // Line number where token appears
    Column   int        // Column number where token starts
}
```

#### Design Rationale
- **Type Safety**: Strongly typed token classification
- **Literal Preservation**: Exact string representation for accurate reconstruction
- **Source Mapping**: Position information for error reporting and IDE features
- **Immutable Design**: Tokens are immutable once created

#### Memory Layout
```
┌─────────────────────────────────────────────────────────┐
│                    Token (48 bytes)                     │
├─────────────────────────────────────────────────────────┤
│ Type         │  4 bytes │ TokenType (int32)            │
│ Literal      │ 16 bytes │ string header (ptr + len)    │
│ Position     │  8 bytes │ int                          │
│ Line         │  8 bytes │ int                          │
│ Column       │  8 bytes │ int                          │
│ padding      │  4 bytes │ struct alignment             │
└─────────────────────────────────────────────────────────┘
```

### 3. TokenType Enumeration

```go
type TokenType int

const (
    // Special tokens
    ILLEGAL TokenType = iota
    EOF
    
    // Identifiers and literals
    IDENT    // table_name, column_name
    INT      // 123456
    FLOAT    // 123.456
    STRING   // 'hello', "world"
    
    // Keywords
    SELECT
    FROM
    WHERE
    INSERT
    UPDATE
    DELETE
    CREATE
    DROP
    ALTER
    
    // Operators
    ASSIGN    // =
    EQ        // ==
    NOT_EQ    // !=, <>
    LT        // <
    GT        // >
    LTE       // <=
    GTE       // >=
    
    // Logic operators
    AND
    OR
    NOT
    
    // Punctuation
    COMMA     // ,
    SEMICOLON // ;
    LPAREN    // (
    RPAREN    // )
    
    // Additional tokens...
)
```

#### Design Rationale
- **Integer Constants**: Efficient comparison and switch statements
- **Logical Grouping**: Related tokens grouped together for maintainability
- **Extensibility**: Easy to add new token types
- **Performance**: Integer comparison faster than string comparison

### 4. Keywords Map

```go
var keywords = map[string]TokenType{
    "SELECT":   SELECT,
    "FROM":     FROM,
    "WHERE":    WHERE,
    "INSERT":   INSERT,
    "UPDATE":   UPDATE,
    "DELETE":   DELETE,
    "CREATE":   CREATE,
    "DROP":     DROP,
    "ALTER":    ALTER,
    "TABLE":    TABLE,
    "INDEX":    INDEX,
    "INTO":     INTO,
    "VALUES":   VALUES,
    "SET":      SET,
    "AND":      AND,
    "OR":       OR,
    "NOT":      NOT,
    "NULL":     NULL,
    "TRUE":     TRUE,
    "FALSE":    FALSE,
    "ASC":      ASC,
    "DESC":     DESC,
    "ORDER":    ORDER,
    "BY":       BY,
    "GROUP":    GROUP,
    "HAVING":   HAVING,
    "LIMIT":    LIMIT,
    "OFFSET":   OFFSET,
    "JOIN":     JOIN,
    "INNER":    INNER,
    "LEFT":     LEFT,
    "RIGHT":    RIGHT,
    "FULL":     FULL,
    "OUTER":    OUTER,
    "ON":       ON,
    "AS":       AS,
    "DISTINCT": DISTINCT,
    "COUNT":    COUNT,
    "SUM":      SUM,
    "AVG":      AVG,
    "MIN":      MIN,
    "MAX":      MAX,
    "CASE":     CASE,
    "WHEN":     WHEN,
    "THEN":     THEN,
    "ELSE":     ELSE,
    "END":      END,
    "IN":       IN,
    "EXISTS":   EXISTS,
    "BETWEEN":  BETWEEN,
    "LIKE":     LIKE,
    "IS":       IS,
}
```

#### Design Rationale
- **Hash Map**: O(1) average case lookup performance
- **String Keys**: Case-sensitive keyword matching (normalized to uppercase)
- **Complete Coverage**: All SQL keywords for comprehensive parsing
- **Static Initialization**: Keywords defined at compile time

#### Memory Characteristics
- **Space Complexity**: O(n) where n is number of keywords
- **Lookup Time**: O(1) average case
- **Memory Usage**: ~2KB for keyword map (64-bit pointers)

### 5. Character Classification Arrays

```go
// Fast character type lookup tables
var isAlpha = [256]bool{
    'A': true, 'B': true, /* ... */, 'Z': true,
    'a': true, 'b': true, /* ... */, 'z': true,
    '_': true,
}

var isDigit = [256]bool{
    '0': true, '1': true, /* ... */, '9': true,
}

var isAlphaNumeric = [256]bool{
    // Combined alpha and numeric characters
}

var isWhitespace = [256]bool{
    ' ': true, '\t': true, '\n': true, '\r': true,
}
```

#### Design Rationale
- **O(1) Lookups**: Array access faster than function calls
- **ASCII Optimization**: 256-byte arrays cover entire ASCII range
- **Memory vs Speed Trade-off**: Small memory cost for significant speed improvement
- **Branch Reduction**: Reduces conditional branches in tight loops

#### Memory Layout
```
┌─────────────────────────────────┐
│ Character Classification Tables │
├─────────────────────────────────┤
│ isAlpha        │ 256 bytes      │
│ isDigit        │ 256 bytes      │ 
│ isAlphaNumeric │ 256 bytes      │
│ isWhitespace   │ 256 bytes      │
│ Total          │ 1KB            │
└─────────────────────────────────┘
```

### 6. Token Buffer (Optional Optimization)

```go
type TokenBuffer struct {
    tokens   []Token    // Pre-allocated token slice
    capacity int        // Maximum number of tokens
    size     int        // Current number of tokens
    index    int        // Current reading position
}
```

#### Design Rationale
- **Memory Pool**: Reuse token storage to reduce GC pressure
- **Batch Processing**: Process multiple tokens at once
- **Streaming**: Support for streaming token processing
- **Performance**: Reduce memory allocations in hot paths

### 7. Error Information Structure

```go
type LexError struct {
    Message  string  // Human-readable error message
    Position int     // Character position of error
    Line     int     // Line number of error
    Column   int     // Column number of error
    Context  string  // Surrounding context for error
}
```

#### Design Rationale
- **Rich Error Context**: Comprehensive error information
- **Source Mapping**: Precise error location
- **User Experience**: Clear error messages for debugging
- **Debugging Support**: Context information for error analysis

## Advanced Data Structures

### 8. State Machine Representation

```go
type LexState int

const (
    StateStart LexState = iota
    StateIdentifier
    StateNumber
    StateString
    StateOperator
    StateComment
    StateError
)

type StateTransition struct {
    From      LexState
    To        LexState
    Condition func(byte) bool
    Action    func(*Lexer) Token
}
```

#### Design Rationale
- **Explicit State Management**: Clear state transitions
- **Extensibility**: Easy to add new states
- **Debugging**: State information for troubleshooting
- **Correctness**: Prevents invalid state transitions

### 9. Unicode Support Structures (Future Extension)

```go
type RuneLexer struct {
    input        []rune    // Unicode runes instead of bytes
    position     int       // Current position in runes
    readPosition int       // Reading position
    ch           rune      // Current rune
    line         int       // Line number
    column       int       // Column number
}
```

#### Design Considerations
- **Full Unicode Support**: Handle all Unicode characters
- **Performance Trade-off**: Larger memory usage vs Unicode support
- **Compatibility**: Can coexist with byte-based lexer
- **Standards Compliance**: Support for Unicode identifiers in SQL

## Data Structure Interactions

### 1. Token Stream Generation Flow

```
Input String
     ↓
   Lexer
     ↓
Position Tracking → Character Reading → Token Classification
     ↓                     ↓                      ↓
Line/Column Info    Current Character        Keyword Lookup
     ↓                     ↓                      ↓
     └─────────── Token Creation ←─────────────────┘
                         ↓
                   Token Stream
```

### 2. Memory Access Patterns

- **Sequential Access**: Input string processed sequentially
- **Random Access**: Keyword map lookups
- **Cache-Friendly**: Character classification arrays fit in CPU cache
- **Minimal Allocation**: Reuse structures where possible

## Performance Characteristics

### 1. Space Complexity

| Component | Space Complexity | Notes |
|-----------|------------------|-------|
| Lexer | O(1) | Fixed size structure |
| Token | O(k) | k = token literal length |
| Keywords Map | O(n) | n = number of keywords |
| Char Tables | O(1) | Fixed 256-byte arrays |
| Token Stream | O(m) | m = number of tokens |

### 2. Time Complexity

| Operation | Time Complexity | Notes |
|-----------|-----------------|-------|
| Character Read | O(1) | Array access |
| Keyword Lookup | O(1) avg | Hash map |
| Token Creation | O(k) | k = token length |
| Full Tokenization | O(n) | n = input length |

### 3. Cache Performance

- **Spatial Locality**: Sequential string processing
- **Temporal Locality**: Keyword map reuse
- **Cache Lines**: Character tables fit in L1 cache
- **Branch Prediction**: Predictable branching patterns

## Memory Management

### 1. Allocation Patterns

```go
// Stack allocated lexer
lexer := NewLexer(input)

// Minimal heap allocations
token := Token{
    Type:     tokenType,
    Literal:  input[start:position], // String slice, no copy
    Position: start,
    Line:     line,
    Column:   column,
}
```

### 2. Garbage Collection Impact

- **String Slicing**: Creates views into original string
- **Minimal Objects**: Few heap-allocated objects
- **Short-Lived Tokens**: Tokens typically have short lifespans
- **GC Pressure**: Low pressure due to minimal allocations

### 3. Memory Safety

- **Bounds Checking**: Go's automatic bounds checking
- **String Immutability**: Input string cannot be modified
- **No Manual Memory**: Automatic memory management
- **Race Condition Free**: No shared mutable state