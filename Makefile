# Makefile for MyReviser with Rust FFI Static Linking
.PHONY: build run clean test install-deps help
.PHONY: build-rust build-rust-linux build-rust-darwin build-rust-windows
.PHONY: build-go package-all lint fmt update-deps

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

# Detect current OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    CURRENT_OS := linux
    RUST_TARGET := x86_64-unknown-linux-musl
    LIB_EXT := a
    BIN_EXT :=
    CC := musl-gcc
    EXTRA_LDFLAGS := -linkmode external -extldflags '-static'
endif
ifeq ($(UNAME_S),Darwin)
    CURRENT_OS := darwin
    RUST_TARGET := x86_64-apple-darwin
    LIB_EXT := a
    BIN_EXT :=
    CC := clang
    EXTRA_LDFLAGS :=
endif
ifeq ($(OS),Windows_NT)
    CURRENT_OS := windows
    RUST_TARGET := x86_64-pc-windows-gnu
    LIB_EXT := a
    BIN_EXT := .exe
    CC := x86_64-w64-mingw32-gcc
    EXTRA_LDFLAGS := -H windowsgui -linkmode external -extldflags '-static'
endif

# Directories
RUST_FFI_DIR := rust-ffi
LIB_DIR := lib
BIN_DIR := bin

# ============================================================================
# Default target
# ============================================================================
all: build

# ============================================================================
# Help
# ============================================================================
help:
	@echo "MyReviser Makefile - Rust FFI Static Build"
	@echo ""
	@echo "Available targets:"
	@echo "  make build                - Build Rust FFI + Go app for current platform"
	@echo "  make build-rust           - Build Rust FFI static library for current platform"
	@echo "  make build-rust-linux     - Build Rust FFI for Linux (musl static)"
	@echo "  make build-rust-darwin    - Build Rust FFI for macOS"
	@echo "  make build-rust-windows   - Build Rust FFI for Windows (MinGW)"
	@echo "  make build-go             - Build Go application (requires Rust library)"
	@echo "  make run                  - Build and run in development mode"
	@echo "  make clean                - Clean all build artifacts"
	@echo "  make test                 - Run all tests (Rust + Go)"
	@echo "  make test-rust            - Run Rust tests only"
	@echo "  make test-go              - Run Go tests only"
	@echo "  make package-all          - Package for all platforms"
	@echo "  make install-deps         - Install build dependencies"
	@echo "  make lint                 - Run linters"
	@echo "  make fmt                  - Format code"
	@echo "  make verify-static        - Verify binary is statically linked"
	@echo ""
	@echo "Current platform: $(CURRENT_OS)"
	@echo "Rust target: $(RUST_TARGET)"

# ============================================================================
# Install Dependencies
# ============================================================================
install-deps:
	@echo "Installing dependencies for $(CURRENT_OS)..."
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Rust toolchain..."
	@command -v rustc >/dev/null 2>&1 || curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
	@echo "Installing Rust target: $(RUST_TARGET)..."
	rustup target add $(RUST_TARGET)
ifeq ($(CURRENT_OS),linux)
	@echo "Installing musl-tools for static builds..."
	@command -v musl-gcc >/dev/null 2>&1 || { \
		echo "Please install musl-tools:"; \
		echo "  sudo apt-get install musl-tools libx11-dev libxtst-dev"; \
	}
endif
ifeq ($(CURRENT_OS),darwin)
	@echo "Checking for Xcode Command Line Tools..."
	@xcode-select -p >/dev/null 2>&1 || { \
		echo "Please install Xcode Command Line Tools:"; \
		echo "  xcode-select --install"; \
	}
endif
	@echo "Dependencies check complete!"

# ============================================================================
# Build Rust FFI Static Library
# ============================================================================

# Build for current platform
build-rust:
	@echo "Building Rust FFI static library for $(CURRENT_OS)..."
	@mkdir -p $(LIB_DIR)
	cd $(RUST_FFI_DIR) && \
		RUSTFLAGS="-C target-feature=+crt-static" \
		cargo build --release --target $(RUST_TARGET)
	@echo "Copying static library to $(LIB_DIR)..."
	cp $(RUST_FFI_DIR)/target/$(RUST_TARGET)/release/libmyreviser_ffi.$(LIB_EXT) $(LIB_DIR)/
	@echo "Copying C header bindings..."
	test -f $(RUST_FFI_DIR)/bindings.h && cp $(RUST_FFI_DIR)/bindings.h $(RUST_FFI_DIR)/ || true
	@echo "Rust FFI library built successfully!"

# Build for Linux (musl for static binary)
build-rust-linux:
	@echo "Building Rust FFI static library for Linux (musl)..."
	@mkdir -p $(LIB_DIR)
	cd $(RUST_FFI_DIR) && \
		rustup target add x86_64-unknown-linux-musl && \
		RUSTFLAGS="-C target-feature=+crt-static" \
		cargo build --release --target x86_64-unknown-linux-musl
	cp $(RUST_FFI_DIR)/target/x86_64-unknown-linux-musl/release/libmyreviser_ffi.a $(LIB_DIR)/
	@echo "Linux Rust FFI library built!"

# Build for macOS
build-rust-darwin:
	@echo "Building Rust FFI static library for macOS..."
	@mkdir -p $(LIB_DIR)
	cd $(RUST_FFI_DIR) && \
		cargo build --release --target x86_64-apple-darwin
	cp $(RUST_FFI_DIR)/target/x86_64-apple-darwin/release/libmyreviser_ffi.a $(LIB_DIR)/
	@echo "macOS Rust FFI library built!"

# Build for Windows (MinGW)
build-rust-windows:
	@echo "Building Rust FFI static library for Windows (MinGW)..."
	@mkdir -p $(LIB_DIR)
	cd $(RUST_FFI_DIR) && \
		rustup target add x86_64-pc-windows-gnu && \
		RUSTFLAGS="-C target-feature=+crt-static" \
		cargo build --release --target x86_64-pc-windows-gnu
	cp $(RUST_FFI_DIR)/target/x86_64-pc-windows-gnu/release/libmyreviser_ffi.a $(LIB_DIR)/
	@echo "Windows Rust FFI library built!"

# ============================================================================
# Build Go Application
# ============================================================================
build-go:
	@echo "Building Go application with Fyne for $(CURRENT_OS)..."
	@mkdir -p $(BIN_DIR)
	@test -f $(LIB_DIR)/libmyreviser_ffi.a || { \
		echo "Error: Rust FFI library not found. Run 'make build-rust' first."; \
		exit 1; \
	}
	@command -v fyne >/dev/null 2>&1 || { \
		echo "Installing Fyne CLI..."; \
		go install fyne.io/fyne/v2/cmd/fyne@latest; \
	}
	CGO_ENABLED=1 \
	CC=$(CC) \
	fyne build -release -o $(BIN_DIR)/myreviser$(BIN_EXT)
	@echo "Go application built successfully!"
	@echo "Binary: $(BIN_DIR)/myreviser$(BIN_EXT)"

# Build everything (Rust + Go)
build: build-rust build-go
	@echo ""
	@echo "✓ Build complete!"
	@echo "  Binary: $(BIN_DIR)/myreviser$(BIN_EXT)"
	@echo ""
	@echo "Run with: ./$(BIN_DIR)/myreviser$(BIN_EXT)"

# ============================================================================
# Run
# ============================================================================
run: build
	@echo "Running MyReviser..."
	./$(BIN_DIR)/myreviser$(BIN_EXT)

# ============================================================================
# Testing
# ============================================================================
test: test-rust test-go

test-rust:
	@echo "Running Rust tests..."
	cd $(RUST_FFI_DIR) && cargo test --release

test-go:
	@echo "Running Go tests..."
	CGO_ENABLED=1 go test -v ./...

# ============================================================================
# Packaging for All Platforms
# ============================================================================
package-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)

	@echo ""
	@echo "=== Building for Linux (static) ==="
	$(MAKE) build-rust-linux
	CGO_ENABLED=1 \
		CC=musl-gcc \
		GOOS=linux \
		GOARCH=amd64 \
		go build \
			-ldflags "$(LDFLAGS) -linkmode external -extldflags '-static'" \
			-o $(BIN_DIR)/myreviser-linux-amd64 \
			.

	@echo ""
	@echo "=== Building for macOS ==="
	$(MAKE) build-rust-darwin
	CGO_ENABLED=1 \
		GOOS=darwin \
		GOARCH=amd64 \
		go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BIN_DIR)/myreviser-darwin-amd64 \
			.

	@echo ""
	@echo "=== Building for Windows (static) ==="
	$(MAKE) build-rust-windows
	CGO_ENABLED=1 \
		CC=x86_64-w64-mingw32-gcc \
		GOOS=windows \
		GOARCH=amd64 \
		go build \
			-ldflags "$(LDFLAGS) -H windowsgui -linkmode external -extldflags '-static'" \
			-o $(BIN_DIR)/myreviser-windows-amd64.exe \
			.

	@echo ""
	@echo "✓ All platforms built successfully!"
	@ls -lh $(BIN_DIR)/

# ============================================================================
# Verification
# ============================================================================
verify-static:
	@echo "Verifying static linking..."
ifeq ($(CURRENT_OS),linux)
	@echo "Linux binary dependencies:"
	ldd $(BIN_DIR)/myreviser$(BIN_EXT) || echo "✓ Statically linked (no dynamic dependencies)"
endif
ifeq ($(CURRENT_OS),darwin)
	@echo "macOS binary dependencies:"
	otool -L $(BIN_DIR)/myreviser$(BIN_EXT)
	@echo "Note: macOS system frameworks are required (cannot be fully static)"
endif
ifeq ($(CURRENT_OS),windows)
	@echo "Windows binary dependencies:"
	objdump -p $(BIN_DIR)/myreviser$(BIN_EXT) | grep "DLL Name:" || echo "No DLL dependencies found"
endif

# ============================================================================
# Code Quality
# ============================================================================
lint:
	@echo "Running Rust lints..."
	cd $(RUST_FFI_DIR) && cargo clippy -- -D warnings
	@echo "Running Go lints..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	golangci-lint run

fmt:
	@echo "Formatting Rust code..."
	cd $(RUST_FFI_DIR) && cargo fmt
	@echo "Formatting Go code..."
	go fmt ./...
	gofmt -s -w .

# ============================================================================
# Clean
# ============================================================================
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/*
	rm -rf $(LIB_DIR)/*
	rm -rf $(RUST_FFI_DIR)/target
	rm -f $(RUST_FFI_DIR)/bindings.h
	rm -rf build/ dist/
	rm -f bundled.go
	rm -f *.AppImage *.dmg *.deb
	@echo "Clean complete!"

clean-rust:
	@echo "Cleaning Rust artifacts only..."
	rm -rf $(RUST_FFI_DIR)/target
	rm -rf $(LIB_DIR)/*

clean-go:
	@echo "Cleaning Go artifacts only..."
	rm -rf $(BIN_DIR)/*
	go clean -cache -testcache

# ============================================================================
# Development
# ============================================================================
dev:
	@echo "Starting development mode with hot reload..."
	@command -v air >/dev/null 2>&1 || { \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	}
	air

# Update dependencies
update-deps:
	@echo "Updating Rust dependencies..."
	cd $(RUST_FFI_DIR) && cargo update
	@echo "Updating Go dependencies..."
	go get -u ./...
	go mod tidy

# Install locally (after building)
install: build
	@echo "Installing MyReviser locally..."
	install -m 755 $(BIN_DIR)/myreviser$(BIN_EXT) /usr/local/bin/myreviser
	@echo "Installed to /usr/local/bin/myreviser"

# ============================================================================
# CI/CD Helpers
# ============================================================================
ci-setup:
	@echo "Setting up CI environment..."
	$(MAKE) install-deps

ci-build:
	@echo "CI build..."
	$(MAKE) build

ci-test:
	@echo "CI test..."
	$(MAKE) test

ci-package:
	@echo "CI package..."
	$(MAKE) package-all
