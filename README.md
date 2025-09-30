# Anansi Proxy

A simple HTTP proxy server that serves predefined responses from a configuration file. Perfect for mocking APIs and testing HTTP clients.

## Features

- Parse HTTP response definitions from a simple text format
- Interactive terminal UI to select which response to serve
- HTTP server that serves the selected response
- Support for custom status codes, content types, and response bodies

## Installation

### Using go install (Recommended)

```bash
go install github.com/pretodev/anansi-proxy@latest
```

### From Source

```bash
git clone https://github.com/pretodev/anansi-proxy.git
cd anansi-proxy
go build -o anansi-proxy .
```

## Usage

```bash
anansi-proxy -f <response-file> [-p <port>]
```

### Options

- `-f, -file`: Path to the HTTP response file (required)
- `-p, -port`: Port number for the HTTP server (default: 8977)

### Example

```bash
# Install the tool
go install github.com/pretodev/anansi-proxy@latest

# Run with a response file
anansi-proxy -f responses.hresp -p 8080
```

## Response File Format

Create a `.hresp` file with the following format:

```
### Success: Request completed

{
  "message": "success"
}

### Success: Text message

Success

### Error: Not found resource
Status-Code: 404
Content-Type: application/json

{
  "message": "Resource not found"
}
```

### Format Rules

- Each response starts with `### <title>`
- Optional headers can be specified as `Key: Value`
- Supported headers: `Status-Code`, `Content-Type`
- Response body follows after headers (or title if no headers)
- Multiple responses are separated by `###`

## Interactive UI

Once started, use the interactive terminal UI to:

- Navigate responses with arrow keys or `j`/`k`
- Select which response the server should return
- Quit with `q` or `Ctrl+C`

The HTTP server will serve the currently selected response to all incoming requests.

## Example Response File

See `resources/api-rest.hresp` for a complete example.

## Requirements

- Go 1.22 or later

## License

MIT License