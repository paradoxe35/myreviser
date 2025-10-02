# Platform Build Guide - Rust FFI Static Libraries

## Overview

This guide explains how to build the Rust FFI static library for different platforms. The static library (`libmyreviser_ffi.a`) embeds all Rust code into the Go binary, eliminating runtime dependencies on Rust libraries.

## Important Note: Platform-Specific Builds

⚠️ **The static library MUST be built on the target platform** (or with proper cross-compilation setup).

- **Linux `.a`** file can only be used on Linux
- **macOS `.a`** file can only be used on macOS
- **Windows `.a`** file can only be used on Windows

You cannot use a Linux-built `.a` file on macOS or Windows.

## Current Build Status

✅ **Tested and Working:**
- Linux (x86_64) - Built and tested successfully
- Binary size: ~34 MB
- Dependencies: System libraries only (libX11, libXtst, etc.)

⏳ **Requires Platform-Specific Build:**
- macOS (x86_64 and arm64)
- Windows (x86_64)

## Building for Each Platform

### Linux (Current Platform)

#### Option 1: Standard Build (Recommended)
```bash
cd rust-ffi
cargo build --release

# Copy to lib directory
cp target/release/libmyreviser_ffi.a ../lib/
```

**Output:** `lib/libmyreviser_ffi.a` (~9.2 MB)

#### Option 2: Musl Build (Fully Static - Requires musl-gcc)
```bash
# Install musl tools
sudo apt-get install musl-tools

# Add musl target
rustup target add x86_64-unknown-linux-musl

# Build with musl
cd rust-ffi
RUSTFLAGS="-C target-feature=+crt-static" \
    cargo build --release --target x86_64-unknown-linux-musl

# Copy to lib directory
cp target/x86_64-unknown-linux-musl/release/libmyreviser_ffi.a ../lib/
```

**Note:** Musl build requires additional configuration for X11 dependencies, which is complex. The standard build is recommended.

#### System Requirements (Linux)
```bash
# Ubuntu/Debian
sudo apt-get install -y \
    build-essential \
    libx11-dev \
    libxtst-dev \
    libxdo-dev \
    libxi-dev \
    libxcb-shape0-dev \
    libxcb-xfixes0-dev

# Fedora/RHEL
sudo dnf install -y \
    gcc \
    libX11-devel \
    libXtst-devel \
    xdotool-devel \
    libXi-devel
```

### macOS

**You must build on macOS** - cross-compilation from Linux is very difficult.

#### Standard Build
```bash
cd rust-ffi
cargo build --release

# Copy to lib directory
cp target/release/libmyreviser_ffi.a ../lib/
```

#### System Requirements (macOS)
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

#### macOS-Specific Notes
- macOS cannot do fully static linking due to system frameworks
- The binary will always link against system frameworks:
  - CoreFoundation
  - Security
  - AppKit
  - ApplicationServices
- This is normal and expected on macOS

### Windows

**You must build on Windows** - or use WSL/MinGW cross-compilation (complex).

#### Option 1: Native Windows Build (MSVC)
```powershell
cd rust-ffi
cargo build --release

# Copy to lib directory
copy target\release\myreviser_ffi.lib ..\lib\
```

#### Option 2: MinGW Build (For Static Linking)
```bash
# Install MinGW target
rustup target add x86_64-pc-windows-gnu

# Build
cd rust-ffi
RUSTFLAGS="-C target-feature=+crt-static" \
    cargo build --release --target x86_64-pc-windows-gnu

# Copy to lib directory
cp target/x86_64-pc-windows-gnu/release/libmyreviser_ffi.a ../lib/
```

#### System Requirements (Windows)
- Visual Studio Build Tools OR MinGW-w64
- Rust installed via rustup

## Build Verification

After building, verify the library:

```bash
# Check file size
ls -lh lib/libmyreviser_ffi.a

# Expected: ~9-10 MB

# Check symbols (Linux/macOS)
nm lib/libmyreviser_ffi.a | grep myreviser_clipboard_new
# Should show: T myreviser_clipboard_new

# On Windows
dumpbin /SYMBOLS lib\myreviser_ffi.lib | findstr myreviser_clipboard_new
```

## Building the Go Application

Once you have the platform-specific static library:

```bash
# From project root
export CGO_ENABLED=1
go build -o bin/myreviser .
```

### Platform-Specific Go Build Commands

#### Linux
```bash
CGO_ENABLED=1 go build \
    -ldflags "-s -w" \
    -o bin/myreviser \
    .
```

#### macOS
```bash
CGO_ENABLED=1 go build \
    -ldflags "-s -w" \
    -o bin/myreviser \
    .
```

#### Windows (PowerShell)
```powershell
$env:CGO_ENABLED=1
go build -ldflags "-s -w" -o bin\myreviser.exe .
```

## Makefile Targets

The updated Makefile provides convenient targets:

```bash
# Build Rust + Go for current platform
make build

# Build Rust library only
make build-rust

# Build Go app only (requires Rust library exists)
make build-go

# Clean everything
make clean

# Verify static linking
make verify-static
```

## CGO Linker Flags (Updated)

The following linker flags are configured for each platform:

### Linux
```go
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo linux LDFLAGS: -lpthread -ldl -lm -lxdo -lX11 -lXtst -lXi
```

### macOS
```go
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security
#cgo darwin LDFLAGS: -framework AppKit -framework ApplicationServices
```

### Windows
```go
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -static
```

## CI/CD Recommendations

For automated builds across platforms, use platform-specific runners:

### GitHub Actions Example
```yaml
name: Build Multi-Platform

on: [push]

jobs:
  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libx11-dev libxtst-dev libxdo-dev
      - name: Build Rust FFI
        run: cd rust-ffi && cargo build --release
      - name: Build Go app
        run: |
          export CGO_ENABLED=1
          go build -o bin/myreviser-linux .
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-linux
          path: bin/myreviser-linux

  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
      - name: Build Rust FFI
        run: cd rust-ffi && cargo build --release
      - name: Build Go app
        run: |
          export CGO_ENABLED=1
          go build -o bin/myreviser-macos .
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-macos
          path: bin/myreviser-macos

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
      - name: Build Rust FFI
        run: cd rust-ffi && cargo build --release
      - name: Build Go app
        run: |
          $env:CGO_ENABLED=1
          go build -o bin\myreviser-windows.exe .
      - uses: actions/upload-artifact@v3
        with:
          name: myreviser-windows
          path: bin/myreviser-windows.exe
```

## Library Distribution Strategy

### Option 1: Build on Each Platform (Recommended)
- Most reliable
- No cross-compilation complexity
- Use CI/CD with platform-specific runners

### Option 2: Pre-built Libraries
Store pre-built libraries for each platform:

```
lib/
├── linux/
│   └── libmyreviser_ffi.a      # Built on Linux
├── darwin/
│   └── libmyreviser_ffi.a      # Built on macOS
└── windows/
    └── libmyreviser_ffi.a      # Built on Windows
```

Update CGO flags to use platform-specific paths:
```go
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/linux/libmyreviser_ffi.a
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/darwin/libmyreviser_ffi.a
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/windows/libmyreviser_ffi.a
```

### Option 3: Download Pre-built from Releases
Store platform-specific builds in GitHub Releases, download during build:

```bash
# Download script
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
wget https://github.com/user/repo/releases/download/v1.0.0/libmyreviser_ffi-${PLATFORM}.a \
    -O lib/libmyreviser_ffi.a
```

## Troubleshooting

### Build Errors

**Error: "cannot find -lxdo"**
```bash
# Linux
sudo apt-get install libxdo-dev

# macOS - not needed
# Windows - not needed
```

**Error: "undefined reference to xdo_*"**
- Missing `-lxdo` in CGO LDFLAGS
- Check updated ffi_*.go files have correct linker flags

**Error: "failed to run custom build command for x11"**
- Missing X11 development libraries
- Run: `sudo apt-get install libx11-dev libxtst-dev`

### Runtime Errors

**Error: "libxdo.so.3: cannot open shared object file"**
```bash
# Linux runtime dependency
sudo apt-get install xdotool
```

**Error on macOS: "library not loaded"**
- Ensure you built on macOS, not cross-compiled
- Check that binary links against system frameworks (normal)

## Binary Size Comparison

| Platform | Rust .a Size | Final Binary | Notes |
|----------|-------------|--------------|-------|
| Linux | ~9.2 MB | ~34 MB | System libs (X11, GL) linked dynamically |
| macOS | ~10 MB | ~30 MB | System frameworks linked dynamically |
| Windows | ~9 MB | ~32 MB | MinGW static or MSVC |

## Summary

✅ **Current Status:**
- Linux build: **Working** (tested successfully)
- Static library: **9.2 MB**
- Final binary: **34 MB**
- All Rust code embedded in binary

⚠️ **For macOS/Windows:**
- Must build on respective platform
- Or set up CI/CD with platform-specific runners
- Cross-compilation is complex and not recommended

**Recommended Workflow:**
1. Use CI/CD (GitHub Actions) with platform-specific runners
2. Build Rust FFI library on each platform
3. Build Go application on each platform
4. Distribute platform-specific binaries

---

**Last Updated:** 2025-10-02
**Status:** Linux build tested and working ✅
