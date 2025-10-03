# ApiMock Syntax Highlighting for VS Code

This extension provides syntax highlighting for `.apimock` files used by Anansi Proxy.

## Features

- Syntax highlighting for HTTP methods (GET, POST, PUT, DELETE, etc.)
- Status code highlighting in response sections
- Path parameter highlighting (e.g., `{userId}`)
- Property/header highlighting
- Embedded JSON, XML, and YAML support
- Comment support with `#`

## Installation

### Option 1: Automated Script (Recommended)

1. Navigate to this directory:
   ```bash
   cd tools/vscode
   ```

2. Run the install script:
   ```bash
   ./install.sh
   ```

This script will:
- Install `vsce` if needed
- Package the extension
- Install it in VS Code

### Option 2: Manual Installation

1. Install `vsce` (if not already installed):
   ```bash
   npm install -g @vscode/vsce
   ```

2. Navigate to this directory:
   ```bash
   cd tools/vscode
   ```

3. Package the extension:
   ```bash
   vsce package
   ```

4. Install the generated `.vsix` file:
   ```bash
   code --install-extension apimock-syntax-0.1.0.vsix
   ```

### Option 3: Development Mode

Link the extension directly (for development):

1. Create a symlink in your VS Code extensions folder:
   ```bash
   # macOS/Linux
   ln -s $(pwd) ~/.vscode/extensions/apimock-syntax
   
   # Then reload VS Code
   ```

## Usage

Once installed, any file with the `.apimock` extension will automatically use this syntax highlighting.

### Example

```apimock
POST /api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}

-- 201: User created successfully
Content-Type: application/json

{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}

-- 400: Invalid request
Content-Type: application/json

{
  "error": "Invalid email format"
}
```

## License

MIT
