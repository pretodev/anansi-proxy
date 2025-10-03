#!/bin/bash

# Script to install ApiMock syntax highlighting extension for VS Code
# This script packages and installs the extension locally

set -e

echo "📦 Installing ApiMock Syntax Highlighting for VS Code..."

# Check if we're in the correct directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: package.json not found. Please run this script from tools/vscode directory."
    exit 1
fi

# Check if vsce is installed
if ! command -v vsce &> /dev/null; then
    echo "📥 Installing vsce (VS Code Extension Manager)..."
    npm install -g @vscode/vsce
fi

# Package the extension
echo "📦 Packaging extension..."
vsce package

# Find the generated .vsix file
VSIX_FILE=$(ls -t *.vsix 2>/dev/null | head -1)

if [ -z "$VSIX_FILE" ]; then
    echo "❌ Error: No .vsix file found after packaging."
    exit 1
fi

echo "✅ Package created: $VSIX_FILE"

# Install the extension
echo "🔧 Installing extension..."
code --install-extension "$VSIX_FILE" --force

echo "✨ Done! ApiMock syntax highlighting has been installed."
echo "📝 Open any .apimock file to see the syntax highlighting in action."
