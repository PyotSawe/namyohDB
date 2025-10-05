# Lexer Module Architecture

## Overview
The Lexer module is responsible for lexical analysis (tokenization) of SQL statements, converting raw SQL text into a stream of tokens that can be processed by the parser.

## Architecture Design

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Lexer Module                             │
├─────────────────────────────────────────────────────────────┤
│  Input: SQL String                                         │
│  Output: Token Stream                                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌────────────────┐    ┌──────────────┐ │
│  │   Input     │───▶│   Character    │───▶│   Token      │ │
│  │   Buffer    │    │   Scanner      │    │  Generator   │ │
│  └─────────────┘    └────────────────┘    └──────────────┘ │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │               Token Classification                      │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │ │
│  │  │  Keywords   │ │ Identifiers │ │    Literals     │   │ │
│  │  │   (SELECT,  │ │ (table,     │ │ (strings,       │   │ │
│  │  │    FROM)    │ │  column     │ │  numbers)       │   │ │
│  │  │             │ │  names)     │ │                 │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────┘   │ │
│  │                                                         │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │ │
│  │  │  Operators  │ │ Punctuation │ │   Comments      │   │ │
│  │  │   (=, <,    │ │   (;, (, )   │ │ (-- , /* */)    │   │ │
│  │  │    >, AND)  │ │    comma)    │ │                 │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────┘   │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. **Single-Pass Lexical Analysis**
- **Decision**: Process input character by character in a single pass
- **Rationale**: Efficient memory usage and real-time processing capability
- **Trade-offs**: More complex state management vs. memory efficiency

#### 2. **Lookahead Buffering**
- **Decision**: Implement single-character lookahead with `peekChar()`
- **Rationale**: Handle multi-character operators (<=, >=, <>, !=) and comments
- **Implementation**: Minimal buffering to avoid memory overhead

#### 3. **Context-Free Tokenization**
- **Decision**: Tokens are identified independently without context
- **Rationale**: Simplifies lexer design and maintains separation of concerns
- **Example**: `SELECT` is always a keyword token, regardless of context

#### 4. **Position Tracking**
- **Decision**: Track line and column numbers for each token
- **Rationale**: Enable precise error reporting and debugging
- **Implementation**: Automatic position updates during character reading

### Data Flow Architecture

```
SQL Input String
      │
      ▼
┌─────────────┐
│   Lexer     │
│ Constructor │ ──┐
└─────────────┘   │
      │           │
      ▼           ▼
┌─────────────┐ ┌─────────────┐
│ readChar()  │ │ Position    │
│             │ │ Tracking    │
└─────────────┘ └─────────────┘
      │
      ▼
┌─────────────────┐
│  NextToken()    │ ──── State Machine ────┐
│                 │                        │
└─────────────────┘                        │
      │                                    ▼
      ▼                           ┌─────────────────┐
┌─────────────────┐               │   Token Type    │
│  Token Stream   │               │ Classification  │
│                 │               └─────────────────┘
└─────────────────┘
```

### Error Handling Strategy

#### 1. **Error Recovery**
- **Approach**: Continue processing after encountering invalid characters
- **Implementation**: Generate `ILLEGAL` tokens for invalid input
- **Benefit**: Allows for partial parsing and better error reporting

#### 2. **Error Context**
- **Position Information**: Line and column numbers for each error
- **Error Messages**: Descriptive messages with context
- **Error Propagation**: Errors bubble up to parser level

### Performance Considerations

#### 1. **Memory Management**
- **Token Storage**: Minimal memory footprint per token
- **String Handling**: Efficient string building with `strings.Builder`
- **Garbage Collection**: Minimize temporary object allocation

#### 2. **Processing Speed**
- **Linear Time Complexity**: O(n) where n is input length
- **Minimal Backtracking**: Single-pass design eliminates backtracking
- **Efficient Character Processing**: Direct rune handling for Unicode support

### Module Interface

```go
type Lexer struct {
    input    string     // Input SQL string
    position int        // Current position in input
    current  rune       // Current character
    line     int        // Current line number
    column   int        // Current column number
}

type Token struct {
    Type     TokenType  // Token classification
    Value    string     // Token value
    Position int        // Character position
    Line     int        // Line number
    Column   int        // Column number
}
```

### Integration Points

#### 1. **Parser Integration**
- **Input**: Token stream from lexer
- **Protocol**: Iterator pattern with `NextToken()`
- **Error Handling**: Invalid tokens propagate as parse errors

#### 2. **Configuration Integration**
- **SQL Dialect**: Keyword set can be configured
- **Case Sensitivity**: Configurable case handling for identifiers
- **Extensions**: Custom token types for database-specific features

### Threading and Concurrency

#### 1. **Thread Safety**
- **Design**: Lexer instances are NOT thread-safe
- **Usage Pattern**: Each parsing operation uses dedicated lexer instance
- **Concurrency**: Multiple lexers can run concurrently on different inputs

#### 2. **Memory Safety**
- **No Shared State**: Each lexer maintains isolated state
- **Resource Cleanup**: Automatic cleanup when lexer goes out of scope

### Extensibility

#### 1. **Token Type Extensions**
- **New Keywords**: Add to `Keywords` map
- **Custom Operators**: Extend operator recognition
- **Dialect Support**: Conditional keyword recognition

#### 2. **Feature Extensions**
- **Unicode Support**: Full Unicode identifier support
- **Escape Sequences**: Extended escape sequence handling
- **Nested Comments**: Support for nested comment blocks

### Testing Strategy

#### 1. **Unit Testing**
- **Token Recognition**: Test each token type individually
- **Edge Cases**: Empty input, malformed tokens, Unicode
- **Error Conditions**: Invalid characters, unterminated strings

#### 2. **Integration Testing**
- **Parser Integration**: End-to-end SQL processing
- **Performance Testing**: Large input handling
- **Memory Testing**: Long-running tokenization

### Compliance and Standards

#### 1. **SQL Standards**
- **ANSI SQL**: Core keyword compatibility
- **Extensions**: Database-specific extensions support
- **Reserved Words**: Proper handling of reserved keywords

#### 2. **Unicode Support**
- **UTF-8 Encoding**: Full UTF-8 input support
- **Identifier Names**: Unicode identifiers support
- **String Literals**: Unicode string content support