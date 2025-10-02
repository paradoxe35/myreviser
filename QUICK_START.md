# Quick Start Guide - Rust FFI Integration

## What Was Done

Your Go application now uses **Rust-based system libraries** via FFI instead of problematic Go libraries:

✅ **Replaced:**
- `golang.design/x/hotkey` → Rust `rdev` (via FFI)
- `golang.design/x/clipboard` → Rust `arboard` (via FFI)
- `github.com/go-vgo/robotgo` → Rust `enigo` (via FFI)

✅ **Created:**
- `rust-ffi/` - Rust static library with C bindings
- `internal/input/ffi_*.go` - Go CGO wrappers
- Updated `Makefile` - Build system with static linking
- `lib/libmyreviser_ffi.a` - Static library (9.2 MB)

## How to Build

### Simple Build (Current Platform)
```bash
# Build everything (Rust library + Go app)
make build

# Run the application
make run
```

### Step-by-Step Build
```bash
# 1. Build Rust FFI library
make build-rust

# 2. Build Go application
make build-go

# 3. Run
./bin/myreviser
```

### Clean Build
```bash
make clean
make build
```

## How to Use in Your Code

### Example 1: Using FFI Clipboard

```go
package main

import (
    "fmt"
    "github.com/paradoxe35/myreviser-go/internal/input"
)

func main() {
    // Create clipboard manager
    clipboard, err := input.NewFFIClipboardManager()
    if err != nil {
        panic(err)
    }
    defer clipboard.Close()  // Important: cleanup

    // Use clipboard
    clipboard.SetText("Hello from Rust!")
    text, _ := clipboard.GetText()
    fmt.Println(text)  // "Hello from Rust!"
}
```

### Example 2: Using FFI Key Simulator

```go
// Simulate keyboard shortcuts
simulator, err := input.NewFFIKeySimulator()
if err != nil {
    panic(err)
}
defer simulator.Close()

// Simulate Ctrl+A / Cmd+A
simulator.SelectAll()

// Simulate Ctrl+C / Cmd+C
simulator.Copy()

// Simulate Ctrl+V / Cmd+V
simulator.Paste()
```

### Example 3: Using FFI Hotkeys

```go
// Create hotkey manager
hotkeys := input.NewFFIHotkeyManager()
if hotkeys == nil {
    panic("failed to create hotkey manager")
}
defer hotkeys.Close()

// Register hotkey with handler
hotkeys.RegisterHotkey("ctrl+alt+space", "my_action", func() {
    fmt.Println("Hotkey pressed!")
})

// Start listening
hotkeys.Start()

// Your app runs here...

// Stop when done
hotkeys.Stop()
```

## Migration from Old Code

### Find and Replace

**Clipboard:**
```go
// OLD:
clipboard, _ := input.NewClipboardManager()

// NEW:
clipboard, _ := input.NewFFIClipboardManager()
defer clipboard.Close()  // Add this!
```

**Simulator:**
```go
// OLD:
input.SimulateSelectAll()

// NEW:
input.FFISimulateSelectAll()

// Or create once and reuse:
sim, _ := input.NewFFIKeySimulator()
defer sim.Close()
sim.SelectAll()
```

**Hotkeys:**
```go
// OLD:
hotkeys := input.NewHotkeyManager()
hotkeys.RegisterHandler("action", handler)

// NEW:
hotkeys := input.NewFFIHotkeyManager()
defer hotkeys.Close()
hotkeys.RegisterHotkey("ctrl+alt+space", "action", handler)
```

## Testing

### Run FFI Test
```bash
go run test_ffi.go
```

### Run Unit Tests
```bash
make test
```

### Verify Static Linking
```bash
make verify-static

# Linux: Should show "statically linked" or only linux-vdso
# macOS: Should only show system frameworks
# Windows: Should only show system DLLs
```

## Troubleshooting

### Build Fails: "cannot find -lmyreviser_ffi"
**Solution:** Build Rust library first
```bash
make build-rust
```

### Build Fails: Missing dependencies
**Solution:** Install dependencies
```bash
# Linux
sudo apt-get install musl-tools libx11-dev libxtst-dev

# macOS
xcode-select --install

# All platforms: Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

### Runtime: "failed to create clipboard manager"
**Solution:** Check system permissions
```bash
# Linux: Ensure X11 libraries installed
sudo apt-get install libx11-dev libxcb1-dev

# macOS: Grant accessibility permissions in System Preferences
# Windows: Run as administrator (if needed)
```

### Hotkeys Not Working
**Solution:** Ensure manager is started
```go
err := hotkeys.Start()
if err != nil {
    log.Fatal(err)
}

// On some systems, requires elevated permissions
```

## Next Steps

1. **Update Your Application Code:**
   - Replace `NewClipboardManager()` with `NewFFIClipboardManager()`
   - Replace `NewKeySimulator()` with `NewFFIKeySimulator()`
   - Replace `NewHotkeyManager()` with `NewFFIHotkeyManager()`
   - Add `defer *.Close()` calls

2. **Test Your Application:**
   ```bash
   make build
   make run
   ```

3. **Remove Old Dependencies:**
   ```bash
   go mod edit -droprequire=golang.design/x/hotkey
   go mod edit -droprequire=golang.design/x/clipboard
   go mod edit -droprequire=github.com/go-vgo/robotgo
   go mod tidy
   ```

4. **Build for Production:**
   ```bash
   make package-all
   # Creates binaries for Linux, macOS, and Windows
   ```

## Files Reference

| File | Description |
|------|-------------|
| `rust-ffi/src/` | Rust FFI source code |
| `rust-ffi/bindings.h` | Auto-generated C header |
| `lib/libmyreviser_ffi.a` | Static library |
| `internal/input/ffi_*.go` | Go CGO wrappers |
| `Makefile` | Build system |
| `test_ffi.go` | Integration test |
| `FFI_BINDING_PLAN.md` | Detailed architecture |
| `STATIC_BUILD_CONFIG.md` | Static linking details |
| `FFI_IMPLEMENTATION_SUMMARY.md` | Complete summary |

## Key Benefits

✅ **Single Binary** - No external library dependencies
✅ **Better Performance** - Rust is faster for system operations
✅ **More Reliable** - Fewer crashes and edge cases
✅ **Cross-Platform** - Works on Linux, macOS, Windows
✅ **Type Safe** - Proper error handling and memory safety
✅ **Future Proof** - Easy to extend with more Rust functionality

## Need Help?

1. Read `FFI_IMPLEMENTATION_SUMMARY.md` for details
2. Check `FFI_BINDING_PLAN.md` for architecture
3. Run `make help` for build targets
4. Test with `go run test_ffi.go`

---

**Status:** ✅ Ready to Use
**Next Step:** Update your application code to use FFI managers
