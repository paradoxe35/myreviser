# Static Build Configuration for Rust FFI

## Overview
This document details the configuration for static linking of the Rust FFI library with the Go application, ensuring a single self-contained binary without external library dependencies.

## Rust Static Library Configuration

### Cargo.toml (rust-ffi/Cargo.toml)
```toml
[package]
name = "myreviser-ffi"
version = "0.1.0"
edition = "2021"

[lib]
name = "myreviser_ffi"
crate-type = ["staticlib"]  # Static library only for embedding in Go

[dependencies]
rdev = "0.5"
arboard = "3.6"
enigo = "0.2"
tokio = { version = "1.41", features = ["rt", "sync", "time"] }
anyhow = "1.0"
tracing = "0.1"
once_cell = "1.20"
parking_lot = "0.12"
libc = "0.2"

[build-dependencies]
cbindgen = "0.28"

[profile.release]
opt-level = "z"        # Optimize for size
lto = true             # Link-time optimization
codegen-units = 1      # Single codegen unit for better optimization
strip = true           # Strip debug symbols
panic = "abort"        # Reduce binary size (no unwinding)
```

## Platform-Specific Build Commands

### Linux (Full Static)
```bash
# Install musl target for truly static binaries
rustup target add x86_64-unknown-linux-musl

# Build Rust static library with musl
cd rust-ffi
RUSTFLAGS="-C target-feature=+crt-static" \
    cargo build --release --target x86_64-unknown-linux-musl

# Copy static library
cp target/x86_64-unknown-linux-musl/release/libmyreviser_ffi.a ../lib/
```

### macOS (Partial Static)
```bash
# macOS cannot do full static linking due to system frameworks
# But Rust code will be statically linked
cd rust-ffi
cargo build --release

# Copy static library
cp target/release/libmyreviser_ffi.a ../lib/
```

### Windows (Full Static with MinGW)
```bash
# Install Windows GNU target
rustup target add x86_64-pc-windows-gnu

# Build Rust static library
cd rust-ffi
RUSTFLAGS="-C target-feature=+crt-static" \
    cargo build --release --target x86_64-pc-windows-gnu

# Copy static library
cp target/x86_64-pc-windows-gnu/release/libmyreviser_ffi.a ../lib/
```

## CGO Configuration for Static Linking

### internal/input/ffi_hotkeys.go
```go
// +build linux darwin windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux static linking
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo linux LDFLAGS: -lpthread -ldl -lm

// macOS linking (partial static, frameworks required)
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security
#cgo darwin LDFLAGS: -framework AppKit -framework ApplicationServices

// Windows static linking
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -lntdll
#cgo windows LDFLAGS: -static -static-libgcc

#include <stdlib.h>
#include "bindings.h"

extern void hotkeyCallbackWrapper(char* action);
*/
import "C"
```

### internal/input/ffi_clipboard.go
```go
// +build linux darwin windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a -lpthread -ldl -lm

// macOS
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security -framework AppKit

// Windows
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -static

#include <stdlib.h>
#include "bindings.h"
*/
import "C"
```

### internal/input/ffi_simulator.go
```go
// +build linux darwin windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a -lpthread -ldl -lm

// macOS
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework AppKit -framework ApplicationServices

// Windows
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -static

#include <stdlib.h>
#include "bindings.h"
*/
import "C"
```

## Go Build Commands

### Linux (Fully Static)
```bash
# Use musl for truly static binary
CGO_ENABLED=1 \
CC=musl-gcc \
go build \
    -ldflags "-s -w -linkmode external -extldflags '-static'" \
    -o bin/myreviser-linux-amd64 \
    .
```

### macOS (Best Effort Static)
```bash
# macOS binaries will link against system frameworks (required)
CGO_ENABLED=1 \
go build \
    -ldflags "-s -w" \
    -o bin/myreviser-darwin-amd64 \
    .
```

### Windows (Fully Static with MinGW)
```bash
# Using MinGW-w64 for static linking
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
GOOS=windows \
GOARCH=amd64 \
go build \
    -ldflags "-s -w -H windowsgui -linkmode external -extldflags '-static'" \
    -o bin/myreviser-windows-amd64.exe \
    .
```

## Verification

### Check for Dynamic Dependencies

#### Linux
```bash
ldd bin/myreviser-linux-amd64
# Should show "not a dynamic executable" or only linux-vdso
```

#### macOS
```bash
otool -L bin/myreviser-darwin-amd64
# Will show system frameworks (unavoidable on macOS)
```

#### Windows
```bash
objdump -p bin/myreviser-windows-amd64.exe | grep DLL
# Should only show system DLLs (KERNEL32.dll, etc.)
```

### Size Comparison
Static binaries will be larger but self-contained:
- Dynamic build: ~15-20 MB (requires .so/.dll/.dylib)
- Static build: ~25-35 MB (includes all Rust code)

## Distribution Benefits

1. **Single Binary**: No need to ship separate .so/.dll/.dylib files
2. **Version Control**: No library version conflicts
3. **Simplified Deployment**: Just copy the binary
4. **Reduced Support**: Fewer "missing library" issues

## Build Dependencies

### Linux (Ubuntu/Debian)
```bash
# Install musl for static builds
sudo apt-get install musl-tools

# Install Rust musl target
rustup target add x86_64-unknown-linux-musl

# Install development libraries (for building, not runtime)
sudo apt-get install libx11-dev libxtst-dev libxcb-shape0-dev libxcb-xfixes0-dev
```

### macOS
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

### Windows (Cross-compile from Linux)
```bash
# Install MinGW cross-compiler
sudo apt-get install gcc-mingw-w64-x86-64

# Install Rust Windows target
rustup target add x86_64-pc-windows-gnu

# Configure cargo for MinGW
mkdir -p ~/.cargo
cat >> ~/.cargo/config.toml <<EOF
[target.x86_64-pc-windows-gnu]
linker = "x86_64-w64-mingw32-gcc"
ar = "x86_64-w64-mingw32-ar"
EOF
```

## Troubleshooting

### Issue: "cannot find -lmyreviser_ffi"
**Solution**: Ensure the static library was built and copied to `lib/` directory

### Issue: Undefined symbols during linking
**Solution**: Check that all system libraries are included in LDFLAGS

### Issue: Large binary size
**Solution**: This is normal for static builds. Use `strip` command to reduce further:
```bash
strip -s bin/myreviser-linux-amd64
```

### Issue: macOS "library not loaded"
**Solution**: macOS requires frameworks, cannot be fully static. Use `install_name_tool` if needed.

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build Static Binaries

on: [push, pull_request]

jobs:
  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
          target: x86_64-unknown-linux-musl
      - run: sudo apt-get install musl-tools libx11-dev
      - run: make build-rust-linux-static
      - run: make build-go-linux-static
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-linux-amd64
          path: bin/myreviser-linux-amd64

  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      - run: make build-rust-darwin
      - run: make build-go-darwin
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-darwin-amd64
          path: bin/myreviser-darwin-amd64

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable-gnu
          target: x86_64-pc-windows-gnu
      - run: make build-rust-windows-static
      - run: make build-go-windows-static
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-windows-amd64.exe
          path: bin/myreviser-windows-amd64.exe
```
