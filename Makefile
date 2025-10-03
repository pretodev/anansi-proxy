.PHONY: build clean test install help

# Default target
all: build

# Build all binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	@go build -o bin/apimock ./cmd/apimock
	@go build -o bin/validator ./cmd/validator
	@echo "✅ Binaries built in ./bin/"

# Run tests
test:
	@echo "Running apimock package tests..."
	@go test -v ./pkg/apimock/...

# Validate example files
validate:
	@echo "Validating example files..."
	@./bin/validator apimock/examples/apimock/

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f apimock/apimock apimock/validator
	@echo "✅ Clean complete"

# Install binaries to $GOPATH/bin
install: build
	@echo "Installing binaries to $(GOPATH)/bin..."
	@cp bin/apimock $(GOPATH)/bin/
	@cp bin/validator $(GOPATH)/bin/
	@echo "✅ Binaries installed"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build     - Build all binaries"
	@echo "  make test      - Run tests"
	@echo "  make validate  - Validate example files"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make install   - Install binaries to \$$GOPATH/bin"
	@echo "  make help      - Show this help message"
