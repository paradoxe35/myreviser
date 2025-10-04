# MyReviser - Project Implementation Plan

> **🚀 Latest Update (Oct 2025)**: Successfully migrated system input layer to **Rust FFI** for better cross-platform compatibility and reliability. All system interactions (global hotkeys, clipboard, key simulation) are now handled via Rust FFI (rdev, arboard, enigo).

## Project Overview

MyReviser is a cross-platform text revision tool built with **Go + Fyne UI + Rust FFI** that integrates with AI providers to automatically enhance text quality, fix grammar, and improve clarity while preserving the original intent and language.

**Note**: For the latest Fyne documentation and API references, use Context7 with library ID: `/fyne-io/docs.fyne.io`

## Core Features

- **Real-time text revision** via keyboard shortcuts
- **Multiple AI provider support** (OpenAI, Claude Anthropic, Google Gemini)
- **Cross-platform compatibility** (Windows, macOS, Linux)
- **System tray integration** via Fyne's native support
- **Single instance application**
- **Configurable settings via GUI** (single main window, no nested dialogs or About window)

## Technology Stack

### Core Technologies

- **Language**: Go 1.24+ with Rust FFI
- **GUI Framework**: Fyne (v2.6.3+) - Cross-platform Go UI framework
- **System Tray**: Native Fyne systray support via `desktop.App` interface
- **System Input (via Rust FFI)**:
  - **Global Hotkeys**: Rust `rdev 0.5` via FFI
  - **Clipboard**: Rust `arboard 3.6` via FFI
  - **Key Simulation**: Rust `enigo 0.2` via FFI
  - **FFI Binding**: `cbindgen` for automatic C header generation
  - **Build**: Static Rust libraries (`.a`) linked into Go binary via CGO
- **HTTP Client**: Go standard `net/http` with context support
- **Configuration**: JSON/TOML with `encoding/json` or `BurntSushi/toml`
- **Single Instance**: File lock or named mutex approach
- **Logging**: `log/slog` (Go 1.21+) with file rotation via `gopkg.in/natefinch/lumberjack.v2`

### Rust FFI Architecture

**Why Rust FFI?**
We use battle-tested Rust libraries via FFI for:

- Better cross-platform support
- More reliable system integration
- Higher performance
- Static linking for simpler deployment

**FFI Implementation:**

```
┌─────────────────────────────────────────────────────────┐
│                    Go Application                        │
│  ┌────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ Fyne UI    │  │ Revision     │  │  Config/AI     │  │
│  │ & Systray  │  │ Processor    │  │  Providers     │  │
│  └─────┬──────┘  └──────┬───────┘  └────────────────┘  │
│        │                 │                               │
│  ┌─────▼─────────────────▼──────────────────────────┐  │
│  │         CGO Wrappers (internal/input)            │  │
│  │  • ffi_hotkeys.go  • ffi_clipboard.go           │  │
│  │  • ffi_simulator.go                              │  │
│  └──────────────────┬───────────────────────────────┘  │
└────────────────────│────────────────────────────────────┘
                     │ C ABI (bindings.h)
┌────────────────────▼────────────────────────────────────┐
│              Rust FFI Library (rust-ffi/)               │
│  ┌──────────────────────────────────────────────────┐  │
│  │  FFI Layer (src/ffi/)                            │  │
│  │  • ffi_hotkey.rs   • ffi_clipboard.rs           │  │
│  │  • ffi_simulator.rs • ffi_types.rs              │  │
│  └────────────┬─────────────────────────────────────┘  │
│               │                                          │
│  ┌────────────▼─────────────────────────────────────┐  │
│  │  Core Implementation (src/core/)                 │  │
│  │  • clipboard.rs (arboard)                        │  │
│  │  • simulator.rs (enigo)                          │  │
│  │  • Hotkey manager (rdev)                         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**Build Output:**

- Rust: `lib/libmyreviser_ffi.a` (static library, ~9MB)
- Go: `bin/myreviser` (with embedded Rust, ~24MB with -release flag)

**Memory Safety:**

- All FFI resources use opaque pointers (`*mut c_void`)
- Explicit cleanup with `Close()` methods
- String memory properly managed across boundaries
- No memory leaks (verified via comprehensive analysis)

See [MEMORY_LEAK_ANALYSIS.md](MEMORY_LEAK_ANALYSIS.md) for detailed FFI safety analysis.

### System Requirements

#### Build Requirements

- **Go**: 1.24 or later
- **Rust**: 1.70+ (stable)
- **CGO**: Required for FFI integration

#### Platform-Specific Dependencies

**Linux:**

```bash
# Required X11 development libraries (for Fyne + FFI)
sudo apt-get update && sudo apt-get install -y \
    libgl1-mesa-dev xorg-dev \
    libx11-dev libxkbfile-dev libxtst-dev \
    libxdo-dev libxi-dev \
    libpng-dev libjpeg-dev \
    libxinerama-dev libxcb-xkb-dev \
    libxcursor-dev libxrandr-dev libxrender-dev \
    libxfixes-dev libxxf86vm-dev \
    libxkbcommon-dev libxkbcommon-x11-dev
```

**macOS:**

```bash
# Install Xcode Command Line Tools
xcode-select --install
```

**Windows:**

- MinGW-w64 or MSVC build tools
- WebView2 Runtime (usually pre-installed on Windows 11)

### Key Dependencies

**Go Dependencies (go.mod):**

```go
module github.com/paradoxe35/myreviser

go 1.24.0

require (
    fyne.io/fyne/v2 v2.6.3
    github.com/allan-simon/go-singleinstance v0.0.0-20210120080615-d0997106ab37
)
// Note: All system input operations use Rust FFI (no external Go dependencies)
```

**Rust Dependencies (rust-ffi/Cargo.toml):**

```toml
[dependencies]
rdev = "0.5"          # Global hotkeys and input monitoring
arboard = "3.6"       # Clipboard management
enigo = "0.2"         # Keyboard/mouse simulation
tokio = { version = "1.41", features = ["rt", "sync", "time"] }
anyhow = "1.0"
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["fmt"] }
once_cell = "1.20"
parking_lot = "0.12"
libc = "0.2"

[build-dependencies]
cbindgen = "0.28"     # Automatic C header generation

[lib]
crate-type = ["staticlib"]  # Static library for CGO linking
```

## Build Process

### Build with Rust FFI + Fyne

**Quick Build (Current Platform):**

```bash
# 1. Build Rust FFI library
cd rust-ffi
cargo build --release
cd ..
mkdir -p lib
cp rust-ffi/target/release/libmyreviser_ffi.a lib/

# 2. Build Go application with Fyne
export CGO_ENABLED=1
fyne build -release -o bin/myreviser

# Or using Makefile
make build  # Builds both Rust and Go
```

**Platform-Specific Builds:**

```bash
# Linux (must build on Linux)
make build-rust-linux   # Builds Rust for Linux
CGO_ENABLED=1 fyne build -release -o bin/myreviser-linux

# macOS (must build on macOS)
make build-rust-darwin  # Builds Rust for macOS
CGO_ENABLED=1 fyne build -release -o bin/myreviser-darwin

# Windows (must build on Windows)
make build-rust-windows # Builds Rust for Windows
CGO_ENABLED=1 fyne build -release -o bin/myreviser.exe
```

**Important Notes:**

- Each platform requires its own Rust static library build
- CGO must be enabled for FFI integration
- Use `fyne build -release` for optimized binaries (27% smaller)
- Binary size: ~24MB (with -release), ~33MB (debug)

## Architecture Design

### File System Structure

```
$HOME/.myreviser/
├── config.json                 # User configuration file
├── logs/                       # Application logs directory
│   ├── myreviser.log          # Current log file
│   └── myreviser.{date}.log   # Rotated log files
└── .lock                       # Single instance lock file
```

### Project Structure

```
myreviser-go/
├── main.go                     # Application entry point, Fyne app setup
├── app.go                      # Core application structure
├── ui/
│   ├── window.go              # Main window implementation
│   ├── settings.go            # Settings window/panel
│   ├── systray.go             # System tray implementation
│   ├── hotkey_capture.go      # Custom hotkey capture widget
│   └── notifications.go       # Notification manager
├── internal/
│   ├── config/
│   │   ├── config.go          # Configuration structures
│   │   ├── storage.go         # Settings persistence
│   │   └── encryption.go      # API key encryption
│   ├── ai/
│   │   ├── provider.go        # AI provider interface
│   │   ├── factory.go         # Provider factory
│   │   ├── openai.go          # OpenAI implementation
│   │   ├── anthropic.go       # Claude implementation
│   │   └── gemini.go          # Google Gemini implementation
│   ├── revision/
│   │   └── processor.go       # Text processing logic
│   ├── input/                  # ✨ FFI Wrappers (Go → Rust)
│   │   ├── ffi_hotkeys.go     # FFI hotkey manager wrapper
│   │   ├── ffi_clipboard.go   # FFI clipboard wrapper
│   │   ├── ffi_simulator.go   # FFI key simulator wrapper
│   │   └── ffi_test.go        # FFI integration tests
│   ├── instance/
│   │   └── manager.go         # Single instance management
│   └── logger/
│       ├── logger.go          # Logging setup with rotation
│       └── viewer.go          # Log file viewer
├── rust-ffi/                   # ✨ Rust FFI Library
│   ├── Cargo.toml             # Rust dependencies
│   ├── build.rs               # cbindgen build script
│   ├── bindings.h             # Generated C headers (by cbindgen)
│   └── src/
│       ├── lib.rs             # Library entry point
│       ├── core/              # Core Rust implementations
│       │   ├── mod.rs
│       │   ├── clipboard.rs   # arboard clipboard wrapper
│       │   ├── simulator.rs   # enigo key simulator
│       │   └── mod.rs
│       └── ffi/               # FFI layer (Rust → C ABI)
│           ├── mod.rs
│           ├── ffi_types.rs   # FFI type definitions & helpers
│           ├── ffi_clipboard.rs  # Clipboard FFI functions
│           ├── ffi_simulator.rs  # Simulator FFI functions
│           └── ffi_hotkey.rs     # Hotkey FFI functions
├── lib/                        # ✨ Compiled Rust libraries
│   └── libmyreviser_ffi.a     # Static library (platform-specific)
├── bin/                        # Built binaries
│   └── myreviser              # Final executable
├── assets/
│   ├── icon.ico              # Windows tray icon
│   └── icon.png              # Linux/macOS tray icon
├── go.mod                     # Go dependencies
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation with Rust + Go
├── MEMORY_LEAK_ANALYSIS.md    # FFI memory safety documentation
└── .github/
    └── workflows/
        └── build.yml          # CI/CD with Rust + Fyne builds
```

### Core Components

#### 1. Single Instance Manager

```go
// internal/instance/manager.go
type InstanceManager struct {
    lockFile *os.File
    mutex    *sync.Mutex
}

func NewInstanceManager() (*InstanceManager, error) {
    lockPath := filepath.Join(os.Getenv("HOME"), ".myreviser", ".lock")
    // Attempt to create exclusive lock
    file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
    if err != nil {
        return nil, fmt.Errorf("another instance is running")
    }
    return &InstanceManager{lockFile: file}, nil
}
```

#### 2. Revision Queue

```go
// internal/revision/queue.go
type RevisionQueue struct {
    mu        sync.Mutex
    current   *RevisionTask
    processor *Processor
    done      chan struct{}
}

type RevisionTask struct {
    Text      string
    Timestamp time.Time
    Mode      string // "select_all" or "selection"
}
```

#### 3. AI Provider Interface

```go
// internal/ai/provider.go
type Provider interface {
    ReviseText(ctx context.Context, text, prompt string) (string, error)
    ValidateConfig() error
    GetName() string
}

type ProviderFactory struct {
    providers map[string]Provider
}
```

#### 4. FFI Hotkey Manager

```go
// internal/input/ffi_hotkeys.go
type FFIHotkeyManager struct {
    mu          sync.RWMutex
    handle      C.HotkeyManagerHandle  // Opaque Rust handle
    handlers    map[string]func()
    active      bool
    lastTrigger map[string]time.Time   // Debounce protection
}

func (h *FFIHotkeyManager) RegisterHotkey(binding, action string, handler func()) error {
    // Register via Rust FFI (supports modifier-only bindings)
    result := C.myreviser_hotkey_register(h.handle, cBinding, cAction, callback)
    // ...
}
```

#### 5. Fyne Application Structure

```go
// main.go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/driver/desktop"
    "github.com/paradoxe35/myreviser/internal/config"
    "github.com/paradoxe35/myreviser/ui"
)

func main() {
    // Create Fyne application
    myApp := app.NewWithID("me.pngwasi.myreviser")
    myApp.SetIcon(resourceIconPng) // Generated from assets

    // Check for desktop features
    if desk, ok := myApp.(desktop.App); ok {
        // Setup system tray
        ui.SetupSystemTray(desk)
    }

    // Create main window
    mainWindow := ui.NewMainWindow(myApp)

    // Setup close intercept for system tray
    mainWindow.SetCloseIntercept(func() {
        mainWindow.Hide()
    })

    // Start the application
    mainWindow.ShowAndRun()
}
```

#### 6. Fyne System Tray Implementation

```go
// ui/systray.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/driver/desktop"
)

func SetupSystemTray(desk desktop.App, mainWindow fyne.Window) {
    // Load icon from assets
    iconResource := fyne.NewStaticResource("icon", iconData)

    // Set system tray icon
    desk.SetSystemTrayIcon(iconResource)

    // Create tray menu
    menu := fyne.NewMenu("MyReviser",
        fyne.NewMenuItem("Show", func() {
            mainWindow.Show()
        }),
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Settings", func() {
            // Open settings panel
            ShowSettings(mainWindow)
        }),
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Quit", func() {
            mainWindow.Close()
            fyne.CurrentApp().Quit()
        }),
    )

    desk.SetSystemTrayMenu(menu)
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)

- [x] Project setup with Fyne initialization
- [x] Update go.mod with Fyne dependency
- [ ] Implement single instance mechanism using file locks
- [ ] Create basic Fyne application structure
- [ ] Set up structured logging with slog and rotation
- [ ] Implement configuration management with JSON/TOML
- [ ] Create settings storage system in ~/.myreviser/

### Phase 2: AI Provider Integration (Week 1-2)

- [ ] Define AI provider trait
- [ ] Implement OpenAI provider
- [ ] Implement Claude Anthropic provider
- [ ] Implement Google Gemini provider
- [ ] Add retry logic and error handling
- [ ] Implement rate limiting

### Phase 3: Input Handling (Week 2)

- [ ] Implement global hotkey registration
- [ ] Create key simulation system
- [ ] Implement clipboard management
- [ ] Add clipboard content preservation
- [ ] Create text selection verification
- [ ] Handle OS-specific key combinations

### Phase 4: Revision Processing (Week 2-3)

- [ ] Implement revision queue
- [ ] Create text processor
- [ ] Add character limit validation
- [ ] Implement revision workflow:
  - Capture text
  - Validate length
  - Send to AI
  - Replace text
  - Restore clipboard

### Phase 5: GUI Development (Week 3)

- [ ] Design Fyne UI with native Go widgets
- [ ] Create settings interface in main window using Fyne containers
- [ ] Implement provider selection with `widget.Select`
- [ ] Add secure API key input fields with `widget.PasswordEntry`
- [ ] Create hotkey customization UI with validation
- [ ] Add system prompt text editor with `widget.MultiLineEntry`
- [ ] Implement character limit setting with `widget.Entry` validation
- [ ] Setup data binding between UI and config
- [ ] Add real-time status updates with Fyne data binding

### Phase 6: System Integration (Week 3-4)

- [ ] Implement Fyne system tray integration using `desktop.App` interface
- [ ] Configure tray icon from assets (icon.ico/icon.png)
- [ ] Add tray menu using Fyne API:
  ```go
  desk.SetSystemTrayIcon(iconResource)
  desk.SetSystemTrayMenu(menu)
  ```
- [ ] Create auto-start functionality for each OS
- [ ] Handle OS-specific permissions (accessibility, input monitoring)
- [ ] Implement minimize to tray behavior with `SetCloseIntercept`
- [ ] Load appropriate icon based on OS:
  - Windows: icon.ico
  - Linux/macOS: icon.png

### Phase 7: Testing & Polish (Week 4)

- [ ] Unit tests for core components
- [ ] Integration tests
- [ ] Cross-platform testing
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] Documentation

## Detailed Workflows

### Revision Workflow - Select All

1. **Hotkey Detection**: Listen for CTRL+ALT+SPACE (or OS variant)
2. **State Check**: Ensure no revision in progress
3. **Text Capture**:
   - Store current clipboard content
   - Simulate SELECT_ALL (Ctrl+A / Cmd+A)
   - Simulate COPY (Ctrl+C / Cmd+C)
   - Retrieve text from clipboard
4. **Validation**:
   - Check text length against limit
   - Verify text is not empty
5. **AI Processing**:
   - Send to selected AI provider
   - Include system prompt
   - Handle response/errors
6. **Text Replacement**:
   - Copy revised text to clipboard
   - Simulate SELECT_ALL again
   - Simulate PASTE (Ctrl+V / Cmd+V)
7. **Cleanup**:
   - Restore original clipboard content
   - Clear revision state

### Revision Workflow - Selection

1. **Hotkey Detection**: Listen for CTRL+WIN (or OS variant)
2. **State Check**: Ensure no revision in progress
3. **Text Capture**:
   - Store current clipboard content
   - Simulate COPY (assumes text pre-selected)
   - Retrieve text from clipboard
4. **Validation**:
   - Verify selected text matches clipboard
   - Check text length against limit
5. **AI Processing**: (same as above)
6. **Text Replacement**:
   - Copy revised text to clipboard
   - Simulate PASTE (replaces selection)
7. **Cleanup**: (same as above)

## Security & Permissions

### Required Permissions

- **Windows**:
  - Accessibility API access
  - Run as background service
  - Global hotkey registration
- **macOS**:
  - Accessibility permissions (System Preferences)
  - Input monitoring
  - Automation permissions
- **Linux**:
  - X11 access for global hotkeys
  - Input device access (may require sudo)

### Security Considerations

- Secure API key storage (OS keychain integration)
- Memory cleanup for sensitive data
- HTTPS only for API communications
- Input validation and sanitization
- Rate limiting to prevent API abuse

## Configuration Schema

```go
// internal/config/config.go
type Config struct {
    AIProvider  AIProviderConfig  `json:"ai_provider"`
    Hotkeys     HotkeyConfig      `json:"hotkeys"`
    Revision    RevisionConfig    `json:"revision"`
    Appearance  AppearanceConfig  `json:"appearance"`
}

type AIProviderConfig struct {
    Provider string `json:"provider"` // "openai" | "claude" | "gemini"
    APIKey   string `json:"api_key"`  // Encrypted in storage
    BaseURL  string `json:"base_url,omitempty"`
}

type HotkeyConfig struct {
    SelectAll string `json:"select_all"` // e.g., "ctrl+alt+space"
    Selection string `json:"selection"`  // e.g., "ctrl+super"
}

type RevisionConfig struct {
    CharacterLimit int    `json:"character_limit"`
    SystemPrompt   string `json:"system_prompt"`
    TimeoutSeconds int    `json:"timeout_seconds"`
}

type AppearanceConfig struct {
    Theme           string `json:"theme"` // "auto" | "light" | "dark"
    StartMinimized  bool   `json:"start_minimized"`
}

// Helper functions
func ConfigPath() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".myreviser", "config.json")
}

func LogDir() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".myreviser", "logs")
}
```

## Default Settings

```json
{
  "ai_provider": {
    "provider": "openai",
    "base_url": "https://api.openai.com/v1"
  },
  "hotkeys": {
    "select_all": "ctrl+alt+space",
    "selection": "ctrl+super"
  },
  "revision": {
    "character_limit": 1000,
    "system_prompt": "You are a multilingual text enhancer: fix errors, improve clarity and quality while preserving tone, context, and intent in the original language. Return only the enhanced version without additional text.",
    "timeout_seconds": 30
  },
  "appearance": {
    "theme": "auto",
    "start_minimized": true
  }
}
```

### Platform-Specific Hotkey Defaults

```go
func GetPlatformHotkeys() HotkeyConfig {
    switch runtime.GOOS {
    case "darwin":
        return HotkeyConfig{
            SelectAll: "ctrl+option+space",
            Selection: "ctrl+cmd",
        }
    case "windows":
        return HotkeyConfig{
            SelectAll: "ctrl+alt+space",
            Selection: "ctrl+win",
        }
    default: // Linux
        return HotkeyConfig{
            SelectAll: "ctrl+alt+space",
            Selection: "ctrl+super",
        }
    }
}
```

## Error Handling Strategy

### User-Facing Errors

- Toast notifications for revision failures
- Clear error messages in settings GUI
- Fallback to original text on failure
- Retry mechanism with exponential backoff

### Internal Errors

- Comprehensive logging with `tracing`
- Error recovery strategies
- Graceful degradation
- Automatic error reporting (opt-in)

## Performance Considerations

- **Lazy loading**: Load AI providers on-demand
- **Async operations**: Non-blocking API calls
- **Resource pooling**: Reuse HTTP connections
- **Memory management**: Clear large strings after use
- **CPU optimization**: Minimal overhead during idle

## Build & Distribution

### Build Commands

#### Development

```bash
# Install Fyne command tools
go install go install fyne.io/tools/cmd/fyne@latest

# Run in development mode
go run .

# Run with live reload (using air or similar)
air
```

#### Production Build

```bash
# Build for current platform
fyne package -os [darwin|linux|windows]

# Build with metadata
fyne bundle -o bundled.go assets/
fyne package -icon assets/icon.png

# Platform-specific builds
fyne package -os windows -icon assets/icon.ico
fyne package -os darwin -icon assets/icon.png
fyne package -os linux -icon assets/icon.png

# Cross-compilation (requires proper toolchain)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    fyne package -os windows
```

### Makefile Build Automation

```makefile
# Makefile
.PHONY: build run clean package-all

VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

build:
	go build -ldflags "$(LDFLAGS)" -o myreviser .

run:
	go run -ldflags "$(LDFLAGS)" .

clean:
	rm -rf build/ dist/ myreviser

package-all:
	# Windows
	fyne package -os windows -icon assets/icon.ico
	# macOS
	fyne package -os darwin -icon assets/icon.png
	# Linux
	fyne package -os linux -icon assets/icon.png
```

### Distribution Formats

- **Windows**: NSIS installer (.exe), portable ZIP
- **macOS**: DMG installer, .app bundle
- **Linux**: AppImage, DEB package, RPM package, tar.gz

## Fyne UI Implementation

### Data Models for UI Binding

```go
// internal/models/models.go
// These structs will be used with Fyne data binding

package models

// Config represents the application configuration
// @description Main configuration structure for MyReviser
type Config struct {
    AIProvider  AIProviderConfig  `json:"ai_provider"`
    Hotkeys     HotkeyConfig      `json:"hotkeys"`
    Revision    RevisionConfig    `json:"revision"`
    Appearance  AppearanceConfig  `json:"appearance"`
}

// RevisionResult represents the result of a text revision
type RevisionResult struct {
    Original    string `json:"original"`
    Revised     string `json:"revised"`
    Provider    string `json:"provider"`
    ProcessTime int64  `json:"process_time_ms"`
    Error       string `json:"error,omitempty"`
}

// StatusUpdate represents a status update event
type StatusUpdate struct {
    Status    string `json:"status"`    // "ready" | "processing" | "error"
    Message   string `json:"message"`
    Progress  int    `json:"progress"`   // 0-100
    Timestamp int64  `json:"timestamp"`
}
```

### Fyne UI Components

```go
// ui/window.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/data/binding"
)

type MainWindow struct {
    window fyne.Window
    app    fyne.App
    config binding.Untyped
}

func NewMainWindow(app fyne.App) *MainWindow {
    w := app.NewWindow("MyReviser")
    w.Resize(fyne.NewSize(600, 400))

    // Create UI components
    providerSelect := widget.NewSelect(
        []string{"OpenAI", "Claude", "Gemini"},
        func(value string) {
            // Handle provider change
        },
    )

    apiKeyEntry := widget.NewPasswordEntry()
    apiKeyEntry.SetPlaceHolder("Enter API Key")

    hotkeyEntry := widget.NewEntry()
    hotkeyEntry.SetPlaceHolder("Ctrl+Alt+Space")

    promptEntry := widget.NewMultiLineEntry()
    promptEntry.SetPlaceHolder("System prompt...")

    // Layout
    content := container.NewVBox(
        widget.NewLabel("AI Provider:"),
        providerSelect,
        widget.NewLabel("API Key:"),
        apiKeyEntry,
        widget.NewLabel("Hotkey:"),
        hotkeyEntry,
        widget.NewLabel("System Prompt:"),
        promptEntry,
        widget.NewButton("Save", func() {
            // Save configuration
        }),
    )

    w.SetContent(container.NewScroll(content))

    return &MainWindow{window: w, app: app}
}
```

### FFI Clipboard Operations

```go
// internal/input/ffi_clipboard.go
package input

type FFIClipboardManager struct {
    handle C.ClipboardHandle  // Opaque Rust handle
}

func NewFFIClipboardManager() (*FFIClipboardManager, error) {
    handle := C.myreviser_clipboard_new()
    if handle == nil {
        return nil, fmt.Errorf("failed to create clipboard manager")
    }
    return &FFIClipboardManager{handle: handle}, nil
}

func (c *FFIClipboardManager) GetText() (string, error) {
    cText := C.myreviser_clipboard_get_text(c.handle)
    if cText == nil {
        return "", fmt.Errorf("failed to get clipboard text")
    }
    defer C.myreviser_free_string(cText)
    return C.GoString(cText), nil
}

func (c *FFIClipboardManager) SetText(text string) error {
    cText := C.CString(text)
    defer C.free(unsafe.Pointer(cText))

    result := C.myreviser_clipboard_set_text(c.handle, cText)
    if result != 0 {
        return fmt.Errorf("failed to set clipboard text")
    }
    return nil
}

func (c *FFIClipboardManager) CaptureSelectedText() (string, error) {
    // Save clipboard, simulate Ctrl+C, get text
    // Caller must restore clipboard after paste
    if err := c.SaveCurrent(); err != nil {
        return "", err
    }

    sim, _ := NewFFIKeySimulator()
    defer sim.Close()

    sim.Copy()
    time.Sleep(100 * time.Millisecond)

    return c.GetText()
}
```

### FFI Hotkey Implementation

```go
// internal/input/ffi_hotkeys.go
package input

type FFIHotkeyManager struct {
    mu          sync.RWMutex
    handle      C.HotkeyManagerHandle
    handlers    map[string]func()
    active      bool
    lastTrigger map[string]time.Time  // Debounce protection
}

func NewFFIHotkeyManager() *FFIHotkeyManager {
    handle := C.myreviser_hotkey_manager_new()
    if handle == nil {
        return nil
    }

    return &FFIHotkeyManager{
        handle:      handle,
        handlers:    make(map[string]func()),
        lastTrigger: make(map[string]time.Time),
    }
}

func (h *FFIHotkeyManager) RegisterHotkey(binding, action string, handler func()) error {
    h.mu.Lock()
    h.handlers[action] = handler
    h.mu.Unlock()

    cBinding := C.CString(binding)
    cAction := C.CString(action)
    defer C.free(unsafe.Pointer(cBinding))
    defer C.free(unsafe.Pointer(cAction))

    // Register with Rust FFI (supports modifier-only bindings like "ctrl+win")
    result := C.myreviser_hotkey_register(h.handle, cBinding, cAction,
                                         C.HotkeyCallback(C.hotkeyCallbackGateway))
    if result != 0 {
        return fmt.Errorf("failed to register hotkey: %s", binding)
    }
    return nil
}

// Callback from Rust when hotkey is triggered
//export hotkeyCallbackGateway
func hotkeyCallbackGateway(action *C.char) {
    defer C.myreviser_free_string(action)
    actionStr := C.GoString(action)

    // Debounce and execute handler in goroutine
    if handler, exists := manager.handlers[actionStr]; exists {
        go handler()
    }
}
```

## GitHub Actions CI/CD

### Automated Release Workflow

The project includes a comprehensive GitHub Actions workflow for automated multi-platform Fyne builds and releases.

**Workflow Location**: `.github/workflows/build.yml` (in the project directory)

#### Key Implementation Details:

- Pure Go implementation with native UI widgets
- Uses Fyne CLI (`go install fyne.io/tools/cmd/fyne@latest`) for building and packaging
- Requires OpenGL and X11 dependencies for Linux
- Uses `fyne bundle` for resource embedding
- Uses `fyne package` for platform-specific packaging
- Supports all major packaging formats (portable, installers, AppImage, DMG, etc.)

#### Workflow Features:

- **Multi-platform builds**: Linux (amd64/arm64), Windows (amd64/arm64), macOS (amd64/arm64)
- **Automated packaging**:
  - Linux: Portable tar.gz, AppImage, .deb packages
  - Windows: Portable ZIP, NSIS installer
  - macOS: .app bundle, ZIP archive, DMG installer
- **Dependency management**: Automatically installs platform-specific dependencies
- **Resource bundling**: Uses `fyne bundle` for embedding assets
- **Version injection**: Injects version and build time into binaries
- **Release automation**: Creates GitHub releases with checksums for all artifacts
- **Triggers**: On push to main/develop, pull requests, tags, and manual dispatch

### Build Dependencies

#### Fyne + Rust FFI Requirements

Both Fyne GUI and Rust FFI require CGO:

**Linux:**

```bash
# Ubuntu/Debian - Fyne + X11 + Rust FFI dependencies
sudo apt-get install -y \
    libgl1-mesa-dev xorg-dev \
    libx11-dev libxkbfile-dev libxtst-dev \
    libxdo-dev libxi-dev \
    libpng-dev libjpeg-dev \
    libxinerama-dev libxcb-xkb-dev \
    libxcursor-dev libxrandr-dev libxrender-dev \
    libxfixes-dev libxxf86vm-dev \
    libxkbcommon-dev libxkbcommon-x11-dev
```

**Windows:**

- MinGW-w64 or MSVC build tools (for CGO)
- Rust toolchain with `x86_64-pc-windows-gnu` target
- No additional permissions required

**macOS:**

- Xcode Command Line Tools
- Rust toolchain
- Requires Accessibility permissions in System Preferences for global hotkeys

### Release Process

1. **Version Tagging**:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Automatic Build**: GitHub Actions will:

   - Build for all platforms using Fyne
   - Create platform-specific packages
   - Generate AppImage and Debian packages for Linux
   - Create GitHub Release with all artifacts

3. **Manual Trigger**: Can also trigger workflow manually from GitHub Actions tab

### Local Development with Fyne

#### Quick Start

```bash
# Clone repository
git clone https://github.com/paradoxe35/myreviser.git
cd myreviser-go/fyne

# Install dependencies
go mod download

# Install Fyne tool
go install go install fyne.io/tools/cmd/fyne@latest

# Run application
go run .

# Or with hot reload (install air first)
go install github.com/air-verse/air@latest
air
```

#### Testing System Tray

```bash
# Run with system tray support
go run . -systray
```

### Documentation Resources

**Important**: Always refer to the latest Fyne documentation using Context7:

- Library ID: `/fyne-io/docs.fyne.io`
- Topics to explore:
  - System tray implementation
  - Data binding
  - Custom widgets
  - Preferences API
  - Clipboard handling
  - Keyboard shortcuts
  - Resource bundling

### Logs Viewer Feature

The application includes a "View Logs" menu item in the system tray that allows users to quickly access application logs:

```go
// internal/logger/viewer.go
package logger

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
)

// OpenLogFile opens the latest log file in the system's default text editor
// If unable to open the file, it opens the logs directory instead
func OpenLogFile() error {
    logFile := GetCurrentLogFile()
    logDir := GetLogDirectory()

    // Platform-specific command to open file in default editor
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        // Windows: use 'start' command
        cmd = exec.Command("cmd", "/c", "start", "", logFile)
    case "darwin":
        // macOS: use 'open' command
        cmd = exec.Command("open", logFile)
    case "linux":
        // Linux: use 'xdg-open' command
        cmd = exec.Command("xdg-open", logFile)
    default:
        // Fallback: try to open directory
        return OpenLogDirectory()
    }

    // Try to open the log file
    if err := cmd.Start(); err != nil {
        // If failed, open the logs directory instead
        return OpenLogDirectory()
    }

    return nil
}

// OpenLogDirectory opens the logs directory in the system's file manager
func OpenLogDirectory() error {
    logDir := GetLogDirectory()

    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("explorer", logDir)
    case "darwin":
        cmd = exec.Command("open", logDir)
    case "linux":
        cmd = exec.Command("xdg-open", logDir)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }

    return cmd.Start()
}

// GetCurrentLogFile returns the path to the current log file
func GetCurrentLogFile() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".myreviser", "logs", "myreviser.log")
}

// GetLogDirectory returns the path to the logs directory
func GetLogDirectory() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".myreviser", "logs")
}
```

This feature provides:

- Quick access to logs for debugging
- Platform-specific implementation for Windows, macOS, and Linux
- Fallback to opening logs directory if file opening fails
- User-friendly way to check application behavior and errors

## Testing Strategy

### Unit Tests

- AI provider implementations
- Text processing logic
- Configuration management
- Hotkey parsing

### Integration Tests

- End-to-end revision workflow
- Multi-provider switching
- Clipboard operations
- Settings persistence

### Manual Testing Checklist

- [ ] All hotkey combinations
- [ ] Each AI provider
- [ ] Various text lengths
- [ ] Special characters/Unicode
- [ ] Multiple languages
- [ ] Concurrent revision attempts
- [ ] Settings changes during operation

## Future Enhancements

### Version 1.1

- Custom AI models support
- Batch text processing
- Revision history
- Undo/redo functionality

### Version 1.2

- Browser extension integration
- Mobile companion app
- Cloud sync for settings
- Custom revision profiles

### Version 2.0

- Local LLM support
- Voice-to-text revision
- Real-time collaboration
- Advanced prompt templates

## Development Timeline

- **Week 1**: Core infrastructure, Fyne setup, and AI providers
- **Week 2**: Input handling with gohook and revision processing
- **Week 3**: GUI development with Fyne widgets and system tray integration
- **Week 4**: Testing, packaging, and CI/CD setup

## Technical Implementation Details

### FFI Clipboard Operations Example

```go
// Using Rust FFI via arboard
import "C"

func captureAndReplace() {
    clipMgr, _ := input.NewFFIClipboardManager()
    defer clipMgr.Close()

    // Save current clipboard
    clipMgr.SaveCurrent()

    // Simulate Ctrl+C to copy selection
    sim, _ := input.NewFFIKeySimulator()
    defer sim.Close()
    sim.Copy()
    time.Sleep(100 * time.Millisecond)

    // Get selected text
    selectedText, _ := clipMgr.GetText()

    // Process text with AI
    revisedText := processWithAI(selectedText)

    // Write revised text and paste
    clipMgr.SetText(revisedText)
    sim.Paste()
    time.Sleep(300 * time.Millisecond)

    // Restore original clipboard
    clipMgr.Restore()
}
```

### Fyne Data Binding Example

```go
// Using Fyne's data binding for reactive UI
package ui

import (
    "fyne.io/fyne/v2/data/binding"
    "fyne.io/fyne/v2/widget"
)

func CreateSettingsForm(config *Config) fyne.CanvasObject {
    // Create data bindings
    apiKeyBinding := binding.NewString()
    apiKeyBinding.Set(config.APIKey)

    providerBinding := binding.NewString()
    providerBinding.Set(config.Provider)

    // Create widgets bound to data
    apiKeyEntry := widget.NewEntryWithData(apiKeyBinding)

    providerSelect := widget.NewSelect(
        []string{"openai", "claude", "gemini"},
        func(value string) {
            providerBinding.Set(value)
        },
    )

    // Listen for changes
    apiKeyBinding.AddListener(binding.NewDataListener(
        func() {
            value, _ := apiKeyBinding.Get()
            config.APIKey = value
            config.Save()
        },
    ))

    return container.NewVBox(
        apiKeyEntry,
        providerSelect,
    )
}
```

### Fyne Shortcuts and Hotkeys

```go
// internal/input/shortcuts.go
package input

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/driver/desktop"
)

func RegisterShortcuts(window fyne.Window) {
    // Create custom shortcuts
    ctrlAltSpace := &desktop.CustomShortcut{
        KeyName:  fyne.KeySpace,
        Modifier: fyne.KeyModifierControl | fyne.KeyModifierAlt,
    }

    // Add shortcut to canvas
    window.Canvas().AddShortcut(ctrlAltSpace, func(shortcut fyne.Shortcut) {
        // Handle revision for all text
        performRevision("select_all")
    })

    // Custom shortcut for selection
    ctrlWin := &desktop.CustomShortcut{
        KeyName:  fyne.KeyUnknown, // Will use platform-specific
        Modifier: fyne.KeyModifierControl | fyne.KeyModifierSuper,
    }

    window.Canvas().AddShortcut(ctrlWin, func(shortcut fyne.Shortcut) {
        // Handle revision for selected text
        performRevision("selection")
    })
}
```

### Fyne Resource Bundling

```go
// Generate resource bundle from assets
// Run: fyne bundle -o bundled.go assets/icon.png

package main

import "fyne.io/fyne/v2"

// This will be generated by fyne bundle
var resourceIconPng = &fyne.StaticResource{
    StaticName: "icon.png",
    StaticContent: []byte{...},
}

// Use in application
func main() {
    app := app.New()
    app.SetIcon(resourceIconPng)
    // ...
}
```

### Fyne Preferences System

```go
// Using Fyne's built-in preferences
package config

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
)

type ConfigManager struct {
    app fyne.App
}

func NewConfigManager(app fyne.App) *ConfigManager {
    return &ConfigManager{app: app}
}

func (c *ConfigManager) SaveAPIKey(key string) {
    c.app.Preferences().SetString("api_key", key)
}

func (c *ConfigManager) GetAPIKey() string {
    return c.app.Preferences().String("api_key")
}

func (c *ConfigManager) SaveProvider(provider string) {
    c.app.Preferences().SetString("provider", provider)
}

func (c *ConfigManager) GetProvider() string {
    return c.app.Preferences().StringWithFallback("provider", "openai")
}
```

### Fyne Build Metadata

```go
// FyneApp.toml - Fyne application metadata
[Details]
Icon = "assets/icon.png"
Name = "MyReviser"
ID = "me.pngwasi.myreviser"
Version = "1.0.0"
Build = 1
```

### Platform-Specific Build Scripts

```bash
#!/bin/bash
# build.sh - Multi-platform build script

VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME"

# Bundle resources
fyne bundle -o bundled.go assets/

# Windows build
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "$LDFLAGS -H=windowsgui" -o build/windows/myreviser.exe

# macOS build
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
  go build -ldflags "$LDFLAGS" -o build/darwin/myreviser
fyne package -os darwin -icon assets/icon.png

# Linux build
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
  go build -ldflags "$LDFLAGS" -o build/linux/myreviser
fyne package -os linux -icon assets/icon.png
```

## Success Metrics

- **Performance**:

  - Revision latency < 3 seconds
  - Memory usage < 80MB active, < 40MB idle (Fyne is lightweight)
  - CPU usage < 5% active, < 1% idle
  - Startup time < 1 second (native Go binary)

- **Reliability**:

  - 99.9% revision success rate
  - Graceful error handling
  - Automatic recovery from failures
  - Cross-platform feature parity

- **User Experience**:
  - Instant hotkey response (< 100ms)
  - Smooth UI animations
  - Clear status indicators
  - Intuitive configuration

## Phase 8: UI/UX Improvements & Fixes (Completed)

### Improvements Implemented:

- [x] Move single instance lock file to `~/.myreviser/` directory (not `.myreviser`)
- [x] Fix window max width issue - prevent infinite expansion when displaying errors
- [x] Add proper spacing/padding between UI elements for professional look
- [x] Implement "Start Minimized" setting respect in UI
- [x] Show Base URL field only for OpenAI provider (hide for Claude/Gemini)
- [x] Replace Character Limit text entry with number input field (default: 1000)
- [x] Add consistent spacing between all form fields and sections
- [x] Implement custom hotkey capture widget with real-time key detection using Fyne's native events
- [x] **Implemented Rust FFI** for system-wide global hotkeys (rdev), clipboard (arboard), and key simulation (enigo)
- [x] Implement multi-provider configuration storage (map-based per-provider settings)
- [x] Add automatic provider reinitialization on config changes
- [x] Update Makefile: Add CGO_ENABLED=1, output to `./bin/` folder
- [x] Update `.gitignore` to ignore `bin/` directory
- [x] Update GitHub Actions workflow to use Fyne packaging commands
- [x] Implement provider-specific field visibility (API Key, Model, Base URL)
- [x] Fix status bar overflow with text wrapping and scrollable container
- [x] Optimize binary size - Added `-s -w` ldflags for stripping

### Technical Details:

#### 1. Lock File Location Fix

- Current: Creates lock in system temp directory
- Required: Create lock in `$HOME/.myreviser/myreviser.lock`
- Update single instance manager initialization

#### 2. Window Size Constraints

- Add maximum width constraint to prevent overflow
- Implement proper text wrapping for error messages
- Use scrollable containers where needed

#### 3. UI Spacing & Layout

```go
// Standard spacing constants
const (
    PaddingSmall  = 5
    PaddingMedium = 10
    PaddingLarge  = 20
)
```

#### 4. Hotkey Capture Widget ✅

**Implemented in `ui/hotkey_capture.go`**

Features:

- Custom Fyne widget using native keyboard events (no external dependencies)
- Click "Capture" button to start listening
- Displays pressed keys as they are detected in real-time
- Automatic detection of modifiers (Ctrl, Alt, Shift, Super/Cmd/Win)
- Validates that combinations include at least one modifier
- ESC to cancel capture, Enter to save
- Clear button to reset hotkey
- Professional UX similar to IDE shortcut configuration
- Platform-specific modifier names (Option on macOS, Alt on others)
- Sorted modifier display (ctrl+alt+shift+key format)

Key mapping includes:

- All modifiers (Ctrl, Alt, Shift, Super/Win/Cmd)
- Letters (a-z)
- Numbers (0-9)
- Function keys (F1-F12)
- Special keys (Space, Enter, Tab, Backspace, Delete)
- Arrow keys (Up, Down, Left, Right)

**Note:** Uses Fyne's built-in keyboard event system for UI capture. System-wide hotkey listening is handled via Rust FFI (rdev) in `internal/input/ffi_hotkeys.go`.

#### 5. Build Optimization

Current binary size: ~42 MB

Optimization strategies:

- Use `-ldflags="-s -w"` for stripping debug info
- Investigate UPX compression (may reduce to ~15-20 MB)
- Remove unused dependencies
- Use `go mod tidy` to clean dependencies
- Consider building with `-buildmode=pie` for smaller size

Expected size after optimization: ~20-25 MB

#### 6. Makefile Updates

```makefile
# Add CGO_ENABLED=1 to all build targets
# Change output directory to ./bin/
build:
    CGO_ENABLED=1 go build -ldflags "-s -w $(LDFLAGS)" -o bin/myreviser .
```

#### 7. Provider-Specific Field Management

```go
// Show/hide fields based on selected provider
providerSelect.OnChanged = func(value string) {
    if value == "openai" {
        baseURLContainer.Show()
    } else {
        baseURLContainer.Hide()
    }
    // Clear and set default values
    updateProviderDefaults(value)
}
```

## Future Enhancements

### Version 1.1

- Custom AI models support
- Batch text processing
- Revision history with undo/redo
- Export/import configuration
- Multiple revision profiles

### Version 1.2

- Browser extension integration via native messaging
- Cloud sync for settings
- Custom prompt templates library
- Statistics dashboard

### Version 2.0

- Local LLM support (Ollama integration)
- Voice-to-text revision
- Plugin system for custom processors
- Team collaboration features

## Implementation Status (Current)

### ✅ Completed Features

#### Rust FFI Integration (October 2025)

- [x] Rust FFI library with cbindgen bindings
- [x] Static library compilation for all platforms
- [x] FFI wrappers in Go (CGO)
- [x] Global hotkeys via Rust rdev
- [x] Clipboard operations via Rust arboard
- [x] Key simulation via Rust enigo
- [x] Modifier-only hotkey support (e.g., `ctrl+win`)
- [x] Memory leak analysis and fixes
- [x] Callback CString lifetime fix
- [x] Comprehensive FFI safety documentation

#### Build System

- [x] Makefile with Rust + Go build targets
- [x] Fyne CLI integration with -release flag
- [x] GitHub Actions CI/CD for multi-platform builds
- [x] Binary size optimization (24MB with -release)
- [x] Platform-specific builds (Linux AMD64, macOS AMD64/ARM64, Windows AMD64)

#### Core Application

- [x] Fyne UI with system tray integration
- [x] Multi-provider AI support (OpenAI, Claude, Gemini)
- [x] Custom hotkey capture widget
- [x] Configuration management with encryption
- [x] Single instance lock
- [x] Logging with rotation
- [x] Theme selection (Light/Dark/Auto)
- [x] Start minimized option

### 📊 Performance Metrics

- **Binary Size**: 24MB (with -release), 33MB (debug)
- **Rust Static Library**: 9.2MB
- **Build Time**: ~8-10 seconds (Rust), ~3-5 seconds (Go)
- **Memory Usage**: <80MB active (tested)
- **FFI Overhead**: Minimal (<1ms per call)

### 🔐 Security & Safety

- **Memory Safety**: All FFI boundaries verified leak-free
- **API Key Encryption**: XOR-based encryption with user-specific key
- **String Management**: Proper allocation/deallocation across FFI
- **Resource Cleanup**: Explicit Close() methods on all FFI handles
- **Thread Safety**: Mutex-protected global state

### 📚 Documentation

- **[PLAN.md](PLAN.md)**: Complete project plan with FFI architecture
- **[MEMORY_LEAK_ANALYSIS.md](MEMORY_LEAK_ANALYSIS.md)**: Detailed FFI safety analysis
- **[README.md](README.md)**: Project overview and quick start

### 🚀 Next Steps

1. **Testing**: Comprehensive cross-platform testing
2. **Performance**: Optimize Tokio runtime usage in FFI
3. **Features**: Add revision history and undo/redo
4. **Distribution**: Create installers for all platforms
5. **Documentation**: User guide and API documentation

---

**Last Updated**: October 2, 2025
**Current Version**: v1.0.0 (FFI Implementation Complete)
**Status**: Production Ready ✅
