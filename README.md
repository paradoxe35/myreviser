# MyReviser

[![Build and Release](https://github.com/paradoxe35/myreviser-go/actions/workflows/build.yml/badge.svg)](https://github.com/paradoxe35/myreviser-go/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://golang.org/dl/)

AI-powered text revision tool with global hotkeys for instant text enhancement across all applications.

## Features

- **Multiple AI Providers**: OpenAI, Claude (Anthropic), and Google Gemini support
- **Global Hotkeys**: System-wide shortcuts work in any application
- **Cross-Platform**: Native builds for Windows, macOS, and Linux
- **System Tray**: Runs quietly in background with easy access
- **Secure Storage**: Encrypted API key storage
- **Multi-Language**: Works with any language supported by AI models
- **Lightweight**: Native Go application with minimal resource usage

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/paradoxe35/myreviser-go/releases).

#### Windows

**Option 1: Installer (Recommended)**

```bash
# Download and run
myreviser-windows-amd64-installer.exe
```

**Option 2: Portable**

```bash
# Extract and run
unzip myreviser-windows-amd64-portable.zip
myreviser.exe
```

#### macOS

**Option 1: DMG (Recommended)**

```bash
# Download and install
open myreviser-darwin-amd64.dmg  # Intel
open myreviser-darwin-arm64.dmg  # Apple Silicon
```

**Option 2: ZIP**

```bash
# Extract and run
unzip myreviser-darwin-amd64.zip
open MyReviser.app
```

**Note**: On first launch, you may need to allow the app in **System Preferences → Privacy & Security → Accessibility**.

#### Linux

**Option 1: AppImage (Universal)**

```bash
# Download and run
chmod +x myreviser-linux-amd64.AppImage
./myreviser-linux-amd64.AppImage
```

**Option 2: Debian/Ubuntu Package**

```bash
# Install package
sudo dpkg -i myreviser-linux-amd64.deb
sudo apt-get install -f  # Install dependencies if needed

# Run
myreviser
```

**Option 3: Portable Archive**

```bash
# Extract and run
tar -xzf myreviser-linux-amd64.tar.gz
chmod +x myreviser
./myreviser
```

### Build from Source

**Requirements:**

- Go 1.24 or later
- CGO enabled
- Platform-specific dependencies (see below)

**Clone and Build:**

```bash
git clone https://github.com/paradoxe35/myreviser-go.git
cd myreviser-go
make build
```

**Platform-Specific Dependencies:**

<details>
<summary><b>Linux</b></summary>

```bash
sudo apt-get update && sudo apt-get install -y \
    libgl1-mesa-dev xorg-dev \
    libx11-dev libxtst-dev libxkbfile-dev \
    libxinerama-dev libxcb-xkb-dev libxkbcommon-x11-dev \
    libxcursor-dev libxrandr-dev libxi-dev
```

</details>

<details>
<summary><b>macOS</b></summary>

```bash
# Install Xcode Command Line Tools
xcode-select --install
```

</details>

<details>
<summary><b>Windows</b></summary>

- Install MinGW-w64 or Visual Studio build tools
- CGO must be enabled
</details>

## Quick Start

1. **Launch MyReviser** - The app will start minimized in your system tray
2. **Configure Settings** - Right-click tray icon and select Settings
3. **Select AI Provider** - Choose OpenAI, Claude, or Gemini
4. **Enter API Key** - Your key is stored encrypted locally
5. **Customize Hotkeys** - Click "Capture" to set your preferred shortcuts
6. **Start Using** - Press your hotkey in any application to revise text

## Default Hotkeys

| Action                  | Linux/Windows             | macOS               |
| ----------------------- | ------------------------- | ------------------- |
| **Select All & Revise** | `Ctrl+Alt+Space`          | `Ctrl+Option+Space` |
| **Revise Selection**    | `Ctrl+Super` / `Ctrl+Win` | `Ctrl+Cmd`          |

_All hotkeys are customizable in Settings_

## Configuration

### Configuration File Location

- **Linux**: `~/.myreviser/config.json`
- **macOS**: `~/.myreviser/config.json`
- **Windows**: `%USERPROFILE%\.myreviser\config.json`

### Supported AI Providers

| Provider   | Model Examples                                            | Base URL                                    |
| ---------- | --------------------------------------------------------- | ------------------------------------------- |
| **OpenAI** | `gpt-4o`, `gpt-4o-mini`                                   | `https://api.openai.com/v1`                 |
| **Claude** | `claude-3-5-haiku-20241022`, `claude-3-5-sonnet-20241022` | `https://api.anthropic.com`                 |
| **Gemini** | `gemini-2.0-flash-exp`, `gemini-1.5-pro`                  | `https://generativelanguage.googleapis.com` |

_Custom base URLs supported for OpenAI-compatible APIs_

## Usage

### Revise All Text

1. Focus on any text field or document
2. Press your "Select All & Revise" hotkey
3. Text is automatically selected, revised, and replaced

### Revise Selected Text

1. Select text in any application
2. Press your "Revise Selection" hotkey
3. Selected text is automatically revised and replaced

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make package-all

# Run in development mode
make run

# Run tests
make test
```

### Project Structure

```
myreviser-go/
├── main.go              # Application entry point
├── internal/
│   ├── ai/             # AI provider implementations
│   ├── config/         # Configuration management
│   ├── input/          # Hotkey and clipboard handling
│   ├── instance/       # Single instance management
│   ├── logger/         # Logging system
│   └── revision/       # Text processing logic
├── ui/                 # Fyne UI components
└── assets/             # Icons and resources
```

## Security & Privacy

- **API Keys**: Stored encrypted using AES-256-GCM
- **No Data Collection**: Text is only sent to your configured AI provider
- **Local Processing**: No data stored or transmitted except to AI APIs
- **HTTPS Only**: All API communications use secure connections

## Troubleshooting

<details>
<summary><b>Application won't start</b></summary>

- Check if another instance is running in system tray
- Delete lock file: `~/.myscript/myscript.lock`
- Check logs: `~/.myreviser/logs/`
</details>

<details>
<summary><b>Hotkeys not working</b></summary>

**Linux:**

- Ensure X11 libraries are installed
- Try running from terminal to see errors

**macOS:**

- Grant Accessibility permissions: System Preferences → Security & Privacy → Accessibility

**Windows:**

- Try running as administrator
- Check for conflicts with other global hotkey apps
</details>

<details>
<summary><b>API Connection Errors</b></summary>

- Verify API key is correct
- Check internet connection
- Ensure you have API credits/quota
- Review logs: Right-click tray icon → View Logs
</details>

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Paradoxe Ng**

- Email: contact@pngwasi.me
- GitHub: [@paradoxe35](https://github.com/paradoxe35)

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI framework for Go
- Global hotkeys via [golang.design/x/hotkey](https://golang.design/x/hotkey)
- Clipboard operations via [golang.design/x/clipboard](https://golang.design/x/clipboard)
- Key simulation via [robotgo](https://github.com/go-vgo/robotgo)

## System Requirements

**Minimum:**

- OS: Windows 10+, macOS 10.13+, or Linux with X11/Wayland
- RAM: 256MB
- Disk: 50MB free space
- Network: Internet connection for AI API calls

**Recommended:**

- RAM: 512MB+
- Modern CPU for faster processing

---

**Note**: This application requires API keys from your chosen AI provider. Sign up at:

- [OpenAI Platform](https://platform.openai.com/)
- [Anthropic Console](https://console.anthropic.com/)
- [Google AI Studio](https://makersuite.google.com/)
