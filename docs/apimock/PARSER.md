# APIMock Parser Library

A Go library for parsing `.apimock` files into structured AST (Abstract Syntax Tree).

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
    // Create a parser
    parser, err := apimock.NewParser("example.apimock")
    if err != nil {
        log.Fatal(err)
    }
    
    // Parse the file
    ast, err := parser.Parse()
    if err != nil {
        log.Fatal(err)
    }
    
    // Access the data
    if ast.Request != nil {
        fmt.Printf("Method: %s\n", ast.Request.Method)
        fmt.Printf("Path: %s\n", ast.Request.Path)
    }
    
    for i, resp := range ast.Responses {
        fmt.Printf("Response %d: %d %s\n", i+1, resp.StatusCode, resp.Description)
    }
}
```

### Working with Path Parameters

```go
for _, seg := range ast.Request.PathSegments {
    if seg.IsParameter {
        fmt.Printf("Parameter: {%s}\n", seg.Name)
    } else {
        fmt.Printf("Literal: %s\n", seg.Value)
    }
}
```

### Accessing Headers and Body

```go
// Request headers
contentType := ast.Request.Headers["Content-Type"]

// Request body schema
schema := ast.Request.BodySchema

// Response
for _, resp := range ast.Responses {
    fmt.Printf("Status: %d\n", resp.StatusCode)
    fmt.Printf("Body: %s\n", resp.Body)
    
    for key, value := range resp.Headers {
        fmt.Printf("Header %s: %s\n", key, value)
    }
}
```

## Data Structures

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
    Method       string
    Path         string
    PathSegments []PathSegment
    QueryParams  map[string]string
    Headers      map[string]string
    BodySchema   string
}
```

### PathSegment

```go
type PathSegment struct {
    Value       string // "users" or "{id}"
    IsParameter bool   // true for {id}
    Name        string // "id" (if IsParameter)
}
```

### ResponseSection

```go
type ResponseSection struct {
    StatusCode  int
    Description string
    Headers     map[string]string
    Body        string
}
```

## Grammar

The parser is based on a formal EBNF grammar defined in [`language.ebnf`](language.ebnf).

## Examples

See the [examples directory](../../examples/apimock/) for sample `.apimock` files.

## License

Part of the Anansi Proxy project.
