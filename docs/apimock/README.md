# APIMock Examples

This directory contains example `.apimock` files demonstrating various features and use cases.

## Files

### Basic Examples

- **`simple.apimock`** - Minimal example with responses only (no request section)
- **`get-json.apimock`** - Simple GET request with JSON response

### Content Types

- **`json.apimock`** - POST request with JSON schema
- **`xml.apimock`** - XML schema and response
- **`yaml.apimock`** - YAML content
- **`raw.apimock`** - Plain text content
- **`octet-stream.apimock`** - Binary data

### Advanced Features

- **`query-path-params.apimock`** - Path parameters (`{id}`) and query strings
- **`form.apimock`** - URL-encoded form data
- **`multipart-form-data.apimock`** - Multipart form submissions

## Usage

### Validate Examples

```bash
# From project root
cd apimock
./validate.sh

# Or using the validator command
validator examples/apimock/
```

### Parse an Example

```bash
# View AST
apimock parse examples/apimock/json.apimock

# Output as JSON
apimock parse examples/apimock/json.apimock --json
```

## File Format

All `.apimock` files follow the EBNF grammar defined in [`docs/apimock/language.ebnf`](../../docs/apimock/language.ebnf).

Basic structure:
```apimock
METHOD /path
Header: value

{request_body}

-- STATUS: Description
ResponseHeader: value

{response_body}
```

See the [documentation](../../docs/apimock/README.md) for complete syntax details.
