.PHONY: build clean test run install help

# Default target
all: build

# Build the binary
build:
	@echo "Building anansi-proxy..."
	@mkdir -p bin
	@go build -o bin/anansi-proxy ./cmd
	@echo "✅ Binary built in ./bin/"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the proxy with example file
run: build
	@echo "Running anansi-proxy with example..."
	@./bin/anansi-proxy -f docs/apimock/examples/simple.apimock -p 8977

# Run in interactive mode
run-interactive: build
	@echo "Running anansi-proxy in interactive mode..."
	@./bin/anansi-proxy -f docs/apimock/examples/simple.apimock -it -p 8977

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✅ Clean complete"

# Install binary to $GOPATH/bin
install: build
	@echo "Installing binary to $(GOPATH)/bin..."
	@cp bin/anansi-proxy $(GOPATH)/bin/
	@echo "✅ Binary installed"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build           - Build the binary"
	@echo "  make test            - Run all tests"
	@echo "  make run             - Run the proxy with example file"
	@echo "  make run-interactive - Run the proxy in interactive mode"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install         - Install binary to \$$GOPATH/bin"
	@echo "  make help            - Show this help message"
