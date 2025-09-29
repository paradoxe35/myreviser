# Makefile for MyReviser
.PHONY: build run clean package-all test install-deps

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

# Default target
all: build

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Fyne command tools..."
	go install fyne.io/fyne/v2/cmd/fyne@latest

# Build for current platform
build:
	@echo "Building MyReviser for current platform..."
	go build -ldflags "$(LDFLAGS)" -o myreviser .

# Run in development mode
run:
	@echo "Running MyReviser in development mode..."
	go run -ldflags "$(LDFLAGS)" .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/ dist/ myreviser myreviser.exe *.AppImage *.dmg *.deb
	rm -f bundled.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Bundle resources
bundle:
	@echo "Bundling resources..."
	fyne bundle -o bundled.go assets/icon.png

# Package for all platforms
package-all: clean bundle
	@echo "Packaging for all platforms..."
	@mkdir -p build/windows build/darwin build/linux

	# Windows
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		go build -ldflags "$(LDFLAGS) -H=windowsgui" -o build/windows/myreviser.exe
	fyne package -os windows -icon assets/icon.ico

	# macOS
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		go build -ldflags "$(LDFLAGS)" -o build/darwin/myreviser
	fyne package -os darwin -icon assets/icon.png

	# Linux
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		go build -ldflags "$(LDFLAGS)" -o build/linux/myreviser
	fyne package -os linux -icon assets/icon.png

# Package for Windows
package-windows: bundle
	@echo "Packaging for Windows..."
	fyne package -os windows -icon assets/icon.ico

# Package for macOS
package-darwin: bundle
	@echo "Packaging for macOS..."
	fyne package -os darwin -icon assets/icon.png

# Package for Linux
package-linux: bundle
	@echo "Packaging for Linux..."
	fyne package -os linux -icon assets/icon.png

# Install locally
install: build
	@echo "Installing MyReviser..."
	fyne install -icon assets/icon.png

# Development with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@command -v air >/dev/null 2>&1 || { \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	}
	air

# Check code quality
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Show help
help:
	@echo "MyReviser Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build for current platform"
	@echo "  make run            - Run in development mode"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make bundle         - Bundle resources"
	@echo "  make package-all    - Package for all platforms"
	@echo "  make package-windows - Package for Windows"
	@echo "  make package-darwin  - Package for macOS"
	@echo "  make package-linux   - Package for Linux"
	@echo "  make install        - Install locally"
	@echo "  make dev            - Run with hot reload"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make update-deps    - Update dependencies"
	@echo "  make help           - Show this help message"