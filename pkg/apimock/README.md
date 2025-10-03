# APIMock Package

The `apimock` package provides a parser and lexer for `.apimock` files, which define HTTP API mock specifications including requests and responses.

## Features

- **Lexical Analysis**: Tokenizes `.apimock` files into a stream of tokens
- **Parsing**: Builds an Abstract Syntax Tree (AST) from tokenized input
- **Validation**: Validates AST for semantic correctness
- **Error Handling**: Provides detailed error messages with file context and line numbers
- **Type Safety**: Strongly-typed Go structures for API definitions
- **Performance Cache**: Built-in caching for improved parsing performance (10-15x faster)

## Installation

```bash
go get github.com/pretodev/anansi-proxy/pkg/apimock
```

## Usage

### Basic Parsing

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/pretodev/anansi-proxy/pkg/apimock"
)

func main() {
    // Create a parser for an .apimock file
    parser, err := apimock.NewParser("example.apimock")
    if err != nil {
        log.Fatalf("Failed to create parser: %v", err)
    }
    
    // Parse the file
    ast, err := parser.Parse()
    if err != nil {
        log.Fatalf("Failed to parse file: %v", err)
    }
    
    // Validate the AST
    if err := ast.Validate(); err != nil {
        log.Fatalf("Validation failed: %v", err)
    }
    
    // Access the parsed data
    if ast.Request != nil {
        fmt.Printf("Method: %s\n", ast.Request.Method)
        fmt.Printf("Path: %s\n", ast.Request.Path)
    }
    
    for _, resp := range ast.Responses {
        fmt.Printf("Response %d: %s\n", resp.StatusCode, resp.Description)
    }
}
```

### Working with Path Parameters

```go
if ast.Request != nil {
    // Check if path has parameters
    if ast.Request.HasPathParameters() {
        params := ast.Request.GetPathParameters()
        fmt.Printf("Path parameters: %v\n", params)
    }
    
    // Iterate through path segments
    for _, seg := range ast.Request.PathSegments {
        if seg.IsParameter {
            fmt.Printf("Parameter: %s\n", seg.Name)
        } else {
            fmt.Printf("Static segment: %s\n", seg.Value)
        }
    }
}
```

### Using Cache for Better Performance

For production use or when parsing the same files multiple times, use `CachedParser`:

```go
// Create a cached parser (10-15x faster for repeated parses)
parser := apimock.NewCachedParser(apimock.DefaultCacheConfig())

// First parse - reads from disk
file, err := parser.ParseFile("example.apimock")

// Second parse - uses cache (much faster!)
file, err = parser.ParseFile("example.apimock")

// Check cache statistics
fmt.Printf("Hit rate: %.2f%%\n", parser.HitRate())
```

**See [CACHE.md](CACHE.md) for complete caching documentation.**

### Error Handling

The package provides detailed error messages with context:

```go
ast, err := parser.Parse()
if err != nil {
    // Check for parse errors with line numbers
    if parseErr, ok := err.(*apimock.ParseError); ok {
        fmt.Printf("Parse error at %s:%d: %s\n", 
            parseErr.Filename, parseErr.Line, parseErr.Message)
    }
}

// Validate AST
if err := ast.Validate(); err != nil {
    // Check for validation errors
    if valErr, ok := err.(*apimock.ValidationError); ok {
        fmt.Printf("Validation error in %s: %s\n", 
            valErr.Field, valErr.Message)
    }
}
```

## APIMock File Format

An `.apimock` file consists of:

1. **Optional Request Section**: Defines the HTTP request
2. **Response Sections**: One or more HTTP response definitions

### Example

```
POST /api/users/{id}
  ?active=true
Content-Type: application/json
Authorization: Bearer token123

{
  "name": "John Doe",
  "email": "john@example.com"
}

-- 200: Success
Content-Type: application/json

{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com"
}

-- 400: Bad Request
Content-Type: application/json

{
  "error": "Invalid input"
}
```

## API Reference

### Types

#### APIMockFile
Represents a complete parsed `.apimock` file.
- `Request *RequestSection`: Optional request section
- `Responses []ResponseSection`: One or more response sections
- `Validate() error`: Validates the file structure

#### RequestSection
Represents an HTTP request definition.
- `Method string`: HTTP method (GET, POST, etc.)
- `Path string`: Request path
- `PathSegments []PathSegment`: Parsed path segments
- `QueryParams map[string]string`: Query parameters
- `Headers map[string]string`: HTTP headers
- `BodySchema string`: Request body content
- `GetPathParameters() []string`: Returns all path parameter names
- `HasPathParameters() bool`: Checks if path has parameters
- `Validate() error`: Validates the request section

#### ResponseSection
Represents an HTTP response definition.
- `StatusCode int`: HTTP status code
- `Description string`: Response description
- `Headers map[string]string`: Response headers
- `Body string`: Response body content
- `Validate() error`: Validates the response section

#### PathSegment
Represents a segment in a URL path.
- `Value string`: The segment value
- `IsParameter bool`: True if it's a parameter placeholder
- `Name string`: Parameter name (if IsParameter is true)
- `String() string`: Returns string representation

### Constants

#### HTTP Methods
- `MethodGet`, `MethodPost`, `MethodPut`, `MethodDelete`, `MethodPatch`
- `MethodHead`, `MethodOptions`, `MethodTrace`, `MethodConnect`

#### HTTP Status Codes
- `MinHTTPStatusCode = 100`
- `MaxHTTPStatusCode = 599`

### Helper Functions

- `IsValidHTTPMethod(method string) bool`: Validates HTTP method
- `IsValidHTTPStatusCode(code int) bool`: Validates HTTP status code

## Testing

The package includes comprehensive test coverage (86.5%):

```bash
go test ./pkg/apimock/... -v -cover
```

## Error Types

### ParseError
Errors that occur during parsing, with filename and line number context.

### ValidationError
Errors found during semantic validation of the AST.

## License

See the main project LICENSE file.
