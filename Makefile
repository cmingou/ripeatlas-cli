# Makefile for ripeatlas

# Binary name
BINARY=ripeatlas

# Build variables
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)
	@echo "✅ Build complete: $(BINARY)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY)
	@echo "✅ Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run with example
.PHONY: run
run: build
	@echo "Running example..."
	./$(BINARY) traceroute --asns 5384,7713 --target 1.1.1.1

# Install to $GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY)..."
	$(GO) install
	@echo "✅ Install complete"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build    - Build the binary (default)"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make run      - Build and run example"
	@echo "  make install  - Install to \$$GOPATH/bin"
	@echo "  make help     - Show this help message"
