# APIMock Parser & AST Documentation

## Overview

This document describes the parser implementation and AST (Abstract Syntax Tree) for the APIMock language.

## Quick Start

### Validate Files

```bash
./apimock validate examples/
```

### Parse and Display AST

```bash
./apimock parse examples/json.apimock
```

### Parse to JSON

```bash
./apimock parse examples/json.apimock --json
```

## Architecture

### Components

1. **Grammar (EBNF)** - [`language.ebnf`](language.ebnf)
   - Formal grammar definition
   - Bottom-up organization
   - Pure EBNF (no ABNF)

2. **AST Structures** - [`ast.go`](ast.go)
   - `APIMockFile` - Root structure
   - `RequestSection` - Request definition
   - `ResponseSection` - Response definition
   - `PathSegment` - Path component with parameter support

3. **Parser** - [`parser.go`](parser.go)
   - Converts `.apimock` files to AST
   - Returns structured data
   - Full error reporting

4. **Validator** - [`validator.go`](validator.go)
   - Grammar validation
   - File format checking
   - Detailed error messages

## AST Structure

### APIMockFile

```go
type APIMockFile struct {
    Request   *RequestSection   // Optional
    Responses []ResponseSection // At least one required
}
```

### RequestSection

```go
type RequestSection struct {
    Method       string            // HTTP method (optional)
    Path         string            // Full path
    PathSegments []PathSegment     // Parsed segments
    QueryParams  map[string]string // Query parameters
    Headers      map[string]string // HTTP headers
    BodySchema   string            // Request body
}
```

### PathSegment

```go
type PathSegment struct {
    Value       string // Segment value
    IsParameter bool   // true for {id}
    Name        string // Parameter name (if IsParameter)
}
```

### ResponseSection

```go
type ResponseSection struct {
    StatusCode  int               // HTTP status code
    Description string            // Optional description
    Headers     map[string]string // Response headers
    Body        string            // Response body
}
```

## Parser Usage

### Basic Parsing

```go
parser, err := NewParser("examples/json.apimock")
if err != nil {
    log.Fatal(err)
}

ast, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

// Access parsed data
fmt.Println("Method:", ast.Request.Method)
fmt.Println("Path:", ast.Request.Path)
fmt.Println("Responses:", len(ast.Responses))
```

### Working with Path Segments

```go
for _, seg := range ast.Request.PathSegments {
    if seg.IsParameter {
        fmt.Printf("Parameter: {%s}\n", seg.Name)
    } else {
        fmt.Printf("Literal: %s\n", seg.Value)
    }
}
```

### Accessing Query Parameters

```go
for key, value := range ast.Request.QueryParams {
    fmt.Printf("%s = %s\n", key, value)
}
```

### Iterating Responses

```go
for i, resp := range ast.Responses {
    fmt.Printf("Response #%d: %d %s\n", 
        i+1, resp.StatusCode, resp.Description)
    
    // Access headers
    contentType := resp.Headers["ContentType"]
    
    // Access body
    body := resp.Body
}
```

## Validation

### Validate Single File

```go
validator, err := NewValidator("file.apimock")
if err != nil {
    log.Fatal(err)
}

if !validator.Validate() {
    // File has errors
}
```

### Validate Directory

```go
valid, invalid, err := ValidateDirectory("examples/")
fmt.Printf("Valid: %d, Invalid: %d\n", valid, invalid)
```

## Testing

### Run All Tests

```bash
go test -v
```

### Run Specific Test

```bash
go test -v -run TestParseJson
```

### Test Coverage

```bash
go test -cover
```

## CLI Commands

### Help

```bash
./apimock help
```

### Validate

```bash
./apimock validate <directory>
```

Example:
```bash
./apimock validate ./examples
```

### Parse

```bash
./apimock parse <file> [--json]
```

Examples:
```bash
# Display AST
./apimock parse examples/json.apimock

# Output JSON
./apimock parse examples/json.apimock --json
```

## Example Output

### Parse Output

```
╔════════════════════════════════════════════╗
║   APIMock Parser - AST Display            ║
╚════════════════════════════════════════════╝

📄 Parsing: json.apimock

✅ Successfully parsed!

╔════════════════════════════════════════════╗
║  Abstract Syntax Tree (AST)                ║
╚════════════════════════════════════════════╝

📤 REQUEST SECTION
──────────────────────────────────────────────────
  Method: POST
  Path: /user
  Path Segments:
    [0] user (literal)
  Headers:
    Accept: application/json
  Body Schema: 197 bytes

📥 RESPONSE SECTIONS (2)
──────────────────────────────────────────────────
  Response #1:
    Status: 201 User created
    Headers:
      ContentType: application/json
    Body: 33 bytes
```

### JSON Output

```json
{
  "Request": {
    "Method": "POST",
    "Path": "/user",
    "PathSegments": [
      {
        "Value": "user",
        "IsParameter": false,
        "Name": ""
      }
    ],
    "QueryParams": {},
    "Headers": {
      "Accept": "application/json"
    },
    "BodySchema": "{\n  \"title\": \"New User\",\n  ..."
  },
  "Responses": [
    {
      "StatusCode": 201,
      "Description": "User created",
      "Headers": {
        "ContentType": "application/json"
      },
      "Body": "{\n  \"name\": \"Luiz\",\n  \"age\": 34\n}"
    }
  ]
}
```

## Grammar Highlights

### Bottom-Up Organization

The EBNF grammar is organized from basic tokens to complex structures:

1. **Basic Tokens** - Characters, whitespace, EOL
2. **Primitive Elements** - Identifiers, status codes
3. **HTTP Elements** - Methods, paths, query params
4. **Properties** - Key-value pairs
5. **Request/Response Components**
6. **Root Rule** - Complete file structure

### Key Features

- Pure EBNF (no ABNF mixing)
- Uses `{ }` for repetition (0+)
- Uses `[ ]` for optional (0 or 1)
- Clear separation of concerns
- Supports path parameters `{id}`
- Supports query strings
- Flexible body content

## Development

### Build

```bash
go build -o apimock .
```

### Run Validator Script

```bash
./validate.sh
```

### Add New Test

1. Create test in `parser_test.go`
2. Add example file in `examples/`
3. Run tests: `go test -v`

## Next Steps

See [README.md](README.md) for integration roadmap and next features.
