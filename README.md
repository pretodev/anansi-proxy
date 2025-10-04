# Anansi Proxy

A lightweight HTTP mock server that serves predefined responses from configuration files. Anansi Proxy allows you to quickly mock REST APIs, test HTTP clients, and simulate various server responses without setting up complex backend infrastructure. Perfect for development, testing, and API prototyping.

## Features

- 📄 Parse HTTP response definitions from `.apimock` files
- 🎯 Interactive terminal UI to dynamically select which response to serve
- 🚀 Lightweight HTTP server that serves the selected response to all requests
- ⚙️ Support for custom status codes, content types, and response headers
- 🔄 Real-time response switching without server restart
- 📁 Multiple example response files included
- 🌳 Recursive directory scanning for `.apimock` files
- 🔀 Serve multiple endpoints from multiple files simultaneously
- 📂 Automatic exclusion of common directories (.git, node_modules, etc.)

## Installation

### Using go install (Recommended)

```bash
go install github.com/pretodev/anansi-proxy@latest
```

### From Source

```bash
git clone https://github.com/pretodev/anansi-proxy.git
cd anansi-proxy
go build -o anansi-proxy ./cmd/main.go
```

## Usage

```bash
anansi-proxy [options] <file_or_directory>...
```

### Command Line Options

- `<file_or_directory>...`: One or more paths to `.apimock` files or directories (required)
- `-p, --port`: Port number for the HTTP server (default: 8977)
- `-it`: Enable interactive mode with terminal UI for response selection

### Usage Examples

#### Single File
```bash
# Serve a single .apimock file
anansi-proxy ./docs/apimock/examples/simple.apimock
```

#### Multiple Files
```bash
# Serve multiple .apimock files
anansi-proxy ./docs/apimock/examples/simple.apimock ./docs/apimock/examples/json.apimock
```

#### Directory (Recursive)
```bash
# Serve all .apimock files from a directory (searches recursively)
anansi-proxy ./docs/apimock/examples
```

#### With Custom Port
```bash
# Specify a custom port
anansi-proxy -p 8080 ./docs/apimock/examples
```

#### Interactive Mode
```bash
# Run with interactive UI to select responses
anansi-proxy -it ./docs/apimock/examples
```

#### Quick Start
```bash
# Install the tool
go install github.com/pretodev/anansi-proxy@latest

# Run with provided examples
anansi-proxy ./docs/apimock/examples
```

## Response File Format

Create a `.hresp` file with the following format:

```apimock
POST /api/users
Content-Type: application/json

{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "email": {"type": "string"}
  }
}

-- 201: User created
Content-Type: application/json

{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}

-- 400: Validation error
Content-Type: application/json

{
  "error": "Invalid email format"
}
```

### Format Rules

- Each response starts with `### <title>`
- Optional headers can be specified as `Key: Value`
- Supported headers: `Status-Code`, `Content-Type`
- Response body follows after headers (or title if no headers)
- Multiple responses are separated by `###`

## Interactive UI

Once started in interactive mode (`-it`), use the terminal UI to:

- Navigate responses with arrow keys or `j`/`k` (up/down)
- Switch between endpoints with arrow keys or `h`/`l` (left/right) when serving multiple files
- Select which response the server should return for each endpoint
- Quit with `q` or `Ctrl+C`

The HTTP server will serve the currently selected response for each endpoint to all incoming requests.

## Examples

The `docs/apimock/examples/` directory contains various `.apimock` files demonstrating different response types:

- `simple.apimock` - Basic JSON and text responses with different status codes
- `json.apimock` - JSON response examples with request schema
- `xml.apimock` - XML response format
- `yaml.apimock` - YAML response format
- `form.apimock` - Form data responses
- `multipart-form-data.apimock` - Multipart form responses
- `octet-stream.apimock` - Binary data responses
- `raw.apimock` - Raw text responses
- `get-json.apimock` - GET request JSON responses
- `query-path-params.apimock` - Query and path parameter examples

### Directory Scanning

When you provide a directory path, Anansi Proxy will:
- Recursively search for all `.apimock` files
- Automatically exclude common directories like `.git`, `node_modules`, `.idea`, `.vscode`, `vendor`, `build`, `dist`, etc.
- Serve all found endpoints on the specified port

### Example Response File

See `docs/apimock/examples/simple.apimock` for a basic example:

## Requirements

- Go 1.22 or later

## License

MIT License