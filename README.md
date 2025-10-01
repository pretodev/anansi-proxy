# Anansi Proxy

A lightweight HTTP mock server that serves predefined responses from configuration files. Anansi Proxy allows you to quickly mock REST APIs, test HTTP clients, and simulate various server responses without setting up complex backend infrastructure. Perfect for development, testing, and API prototyping.

## Features

- 📄 Parse HTTP response definitions from simple text format (`.hresp` files)
- 🎯 Interactive terminal UI to dynamically select which response to serve
- 🚀 Lightweight HTTP server that serves the selected response to all requests
- ⚙️ Support for custom status codes, content types, and response headers
- 🔄 Real-time response switching without server restart
- 📁 Multiple example response files included

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
anansi-proxy -f <response-file> [-p <port>] [-it]
```

### Command Line Options

- `-f, --file`: Path to the HTTP response file (required)
- `-p, --port`: Port number for the HTTP server (default: 8977)
- `-it`: Enable interactive mode with terminal UI for response selection

### Usage Examples

#### Interactive Mode (Default)
```bash
# Run with interactive UI to select responses
anansi-proxy -f examples/simple.hresp -p 8080 -it
```

#### Non-Interactive Mode
```bash
# Run with the first response automatically selected
anansi-proxy -f examples/simple.hresp -p 8080
```

#### Quick Start
```bash
# Install the tool
go install github.com/pretodev/anansi-proxy@latest

# Run with provided example
anansi-proxy -f examples/simple.hresp -it
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

## Examples

The `examples/` directory contains various `.hresp` files demonstrating different response types:

- `simple.hresp` - Basic JSON and text responses with different status codes
- `json.hresp` - JSON response examples
- `xml.hresp` - XML response format
- `yaml.hresp` - YAML response format
- `form.hresp` - Form data responses
- `multipart-form-data.hresp` - Multipart form responses
- `octet-stream.hresp` - Binary data responses
- `raw.hresp` - Raw text responses
- `get-json.hresp` - GET request JSON responses

### Example Response File

See `examples/simple.hresp` for a basic example:

## Requirements

- Go 1.22 or later

## License

MIT License