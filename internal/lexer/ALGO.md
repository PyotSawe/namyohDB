# Lexer Module Algorithms

## Overview
This document describes the core algorithms used in the SQL lexer for tokenization, including character processing, token recognition, and error handling strategies.

## Core Algorithms

### 1. Token Recognition Algorithm

#### Algorithm: `NextToken()`
```
Input: Current lexer state (position, current character)
Output: Next token in the input stream

Algorithm NextToken():
  1. Skip whitespace characters
  2. Check for end of input → return EOF token
  3. Save current position for token creation
  4. Determine token type based on current character:
     a. Letter/underscore → Identifier or Keyword
     b. Digit → Numeric literal
     c. Quote → String literal
     d. Operator character → Operator token
     e. Special character → Punctuation token
     f. Unknown → Illegal token
  5. Read complete token value
  6. Create and return token with position information
```

**Time Complexity**: O(k) where k is the length of the token
**Space Complexity**: O(k) for token value storage

### 2. Character Reading Algorithm

#### Algorithm: `readChar()`
```
Input: Current lexer state
Output: Updates lexer position and current character

Algorithm readChar():
  1. If position >= length of input:
     a. Set current character to null terminator
     b. Return
  2. Set current character to input[position]
  3. Increment position
  4. Update line and column tracking:
     a. If previous character was newline:
        - Increment line number
        - Reset column to 1
     b. Else:
        - Increment column number
```

**Time Complexity**: O(1)
**Space Complexity**: O(1)

### 3. Lookahead Algorithm

#### Algorithm: `peekChar()`
```
Input: Current lexer state
Output: Next character without advancing position

Algorithm peekChar():
  1. If (position + 1) >= length of input:
     a. Return null terminator
  2. Return input[position + 1]
```

**Time Complexity**: O(1)
**Space Complexity**: O(1)

### 4. Identifier Recognition Algorithm

#### Algorithm: `readIdentifier()`
```
Input: Current character (first character of identifier)
Output: Complete identifier string

Algorithm readIdentifier():
  1. Initialize result with current character
  2. Advance to next character
  3. While current character is letter, digit, or underscore:
     a. Append current character to result
     b. Advance to next character
  4. Return result
```

**Time Complexity**: O(n) where n is identifier length
**Space Complexity**: O(n) for identifier storage

### 5. Numeric Literal Recognition Algorithm

#### Algorithm: `readNumber()`
```
Input: Current character (first digit)
Output: Complete numeric literal string

Algorithm readNumber():
  1. Initialize result with current character
  2. Advance to next character
  3. While current character is digit:
     a. Append current character to result
     b. Advance to next character
  4. Check for decimal point:
     a. If current character is '.':
        - Append '.' to result
        - Advance to next character
        - While current character is digit:
          * Append current character to result
          * Advance to next character
  5. Check for scientific notation (e/E):
     a. If current character is 'e' or 'E':
        - Append to result and advance
        - Check for sign (+ or -)
        - Read remaining digits
  6. Return result
```

**Time Complexity**: O(n) where n is number length
**Space Complexity**: O(n) for number storage

### 6. String Literal Recognition Algorithm

#### Algorithm: `readString(quote_char)`
```
Input: Quote character (' or ")
Output: Complete string literal including quotes

Algorithm readString(quote_char):
  1. Initialize result with quote character
  2. Advance to next character
  3. While current character != quote_char AND not EOF:
     a. If current character is escape character ('\'):
        - Append escape character to result
        - Advance to next character
        - If not EOF, append escaped character to result
        - Advance to next character
     b. Else:
        - Append current character to result
        - Advance to next character
  4. If current character is closing quote:
     a. Append closing quote to result
     b. Advance to next character
  5. Return result
```

**Time Complexity**: O(n) where n is string length
**Space Complexity**: O(n) for string storage

### 7. Comment Recognition Algorithm

#### Algorithm: `readComment()`
```
Input: Current character (start of comment)
Output: Complete comment string or skip comment

Algorithm readComment():
  1. If current character is '-' and next is '-':
     a. Skip single-line comment:
        - While current character != newline AND not EOF:
          * Advance to next character
     b. Return null (comment skipped)
  
  2. If current character is '/' and next is '*':
     a. Skip multi-line comment:
        - Advance past '/*'
        - While not EOF:
          * If current is '*' and next is '/':
            - Advance past '*/'
            - Break
          * Advance to next character
     b. Return null (comment skipped)
  
  3. Return null (not a comment)
```

**Time Complexity**: O(n) where n is comment length
**Space Complexity**: O(1) - comments are skipped

### 8. Keyword Recognition Algorithm

#### Algorithm: `lookupIdent(ident)`
```
Input: Identifier string
Output: Token type (KEYWORD or IDENT)

Algorithm lookupIdent(ident):
  1. Convert identifier to uppercase for case-insensitive lookup
  2. Check if identifier exists in keywords map:
     a. If found → return corresponding keyword token type
     b. If not found → return IDENT token type
```

**Time Complexity**: O(1) average case (hash map lookup)
**Space Complexity**: O(1)

### 9. Operator Recognition Algorithm

#### Algorithm: `readOperator(char)`
```
Input: First character of potential operator
Output: Complete operator token

Algorithm readOperator(char):
  1. Switch on current character:
     a. '=' → Check for '=' to form '=='
     b. '<' → Check for '=', '>', or '<' to form '<=', '<>', '<<'
     c. '>' → Check for '=', or '>' to form '>=', '>>'
     d. '!' → Check for '=' to form '!='
     e. '|' → Check for '|' to form '||'
     f. '&' → Check for '&' to form '&&'
     g. Single character operators → Return immediately
  
  2. If multi-character operator detected:
     a. Consume additional characters
     b. Return compound operator
  
  3. Return single-character operator
```

**Time Complexity**: O(1) - maximum 2 character lookahead
**Space Complexity**: O(1)

## Advanced Algorithms

### 10. Error Recovery Algorithm

#### Algorithm: `recoverFromError()`
```
Input: Current error state
Output: Next valid token position

Algorithm recoverFromError():
  1. Create ILLEGAL token with current character
  2. Advance to next character
  3. Continue lexical analysis
  4. Maintain error count for reporting
```

**Time Complexity**: O(1)
**Space Complexity**: O(1)

### 11. Position Tracking Algorithm

#### Algorithm: `updatePosition(char)`
```
Input: Character just processed
Output: Updated line and column information

Algorithm updatePosition(char):
  1. If char is newline ('\n'):
     a. Increment line number
     b. Reset column to 0
  2. Else if char is tab ('\t'):
     a. Increment column by tab width (usually 4 or 8)
  3. Else:
     a. Increment column by 1
```

**Time Complexity**: O(1)
**Space Complexity**: O(1)

### 12. Unicode Handling Algorithm

#### Algorithm: `readUnicodeChar()`
```
Input: Current position in UTF-8 encoded string
Output: Unicode rune and updated position

Algorithm readUnicodeChar():
  1. Check first byte to determine UTF-8 sequence length:
     a. 0xxxxxxx → 1 byte (ASCII)
     b. 110xxxxx → 2 bytes
     c. 1110xxxx → 3 bytes
     d. 11110xxx → 4 bytes
  
  2. Read appropriate number of bytes
  3. Decode UTF-8 sequence to Unicode rune
  4. Update position by number of bytes consumed
  5. Return rune
```

**Time Complexity**: O(1) - fixed maximum 4 bytes
**Space Complexity**: O(1)

## Performance Optimizations

### 1. Character Classification Optimization

```go
// Fast character type checking using lookup table
var charTypes = [256]int{
    // Initialize with character type classifications
}

func isLetter(ch byte) bool {
    return charTypes[ch] == LETTER_TYPE
}
```

### 2. Keyword Lookup Optimization

```go
// Pre-computed hash map for O(1) keyword lookup
var keywords = map[string]TokenType{
    "SELECT": SELECT,
    "FROM":   FROM,
    "WHERE":  WHERE,
    // ... other keywords
}
```

### 3. String Builder Optimization

```go
// Use strings.Builder for efficient string construction
var builder strings.Builder
builder.Grow(estimatedSize) // Pre-allocate capacity
```

## Algorithm Complexity Analysis

| Algorithm | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|--------|
| NextToken() | O(k) | O(k) | k = token length |
| readChar() | O(1) | O(1) | Constant time character access |
| readIdentifier() | O(n) | O(n) | n = identifier length |
| readNumber() | O(n) | O(n) | n = number length |
| readString() | O(n) | O(n) | n = string length |
| lookupIdent() | O(1) avg | O(1) | Hash map lookup |
| readOperator() | O(1) | O(1) | Maximum 2-char lookahead |

## Edge Cases and Error Handling

### 1. Unterminated String Literals
```
Input: "SELECT * FROM table WHERE name = 'unterminated
Action: Generate string token up to EOF, report error
```

### 2. Invalid Characters
```
Input: SELECT @ FROM table
Action: Generate ILLEGAL token for '@', continue parsing
```

### 3. Numeric Overflow
```
Input: 999999999999999999999999999999
Action: Return complete number string, let parser handle validation
```

### 4. Unicode in Identifiers
```
Input: SELECT résumé FROM employees
Action: Support Unicode identifiers following SQL standards
```

## Testing Algorithms

### 1. Property-Based Testing
- **Property**: Every character in input appears in exactly one token
- **Property**: Token positions are monotonically increasing
- **Property**: Concatenating token values reconstructs original input (minus whitespace/comments)

### 2. Fuzzing Strategy
- Random character sequences
- Malformed SQL statements
- Boundary conditions (very long identifiers, numbers)
- Unicode edge cases

### 3. Performance Benchmarks
- Large SQL file processing
- Token-per-second measurements
- Memory allocation profiling
- Garbage collection impact analysis