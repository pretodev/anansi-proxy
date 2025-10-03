#!/bin/bash

# APIMock EBNF Grammar Validator
# Validates all .apimock files against the EBNF grammar definition

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLES_DIR="${SCRIPT_DIR}/examples"

echo "Building validator..."
cd "$SCRIPT_DIR"
go build -o validator validator.go

echo ""
echo "Running validation..."
echo ""

./validator "$EXAMPLES_DIR"

# Cleanup
rm -f validator

echo ""
echo "✨ Validation complete!"
