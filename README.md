# MyReviser

AI-powered text revision tool for improving grammar and clarity across multiple languages.

## Author

**Paradoxe Ng**
Email: contact@pngwasi.me
Repository: [https://github.com/paradoxe35/myreviser-go](https://github.com/paradoxe35/myreviser-go)

## Features

- **Multi-AI Provider Support**: Works with OpenAI, Claude, and Gemini APIs
- **Global Hotkeys**: Quick text revision with customizable keyboard shortcuts
- **Cross-Platform**: Runs on Windows, macOS, and Linux
- **System Tray Integration**: Runs in background with easy access
- **Secure Configuration**: Encrypted API key storage
- **Smart Text Processing**: Preserves formatting and context
- **Multi-Language Support**: Works with any language supported by the AI models

## Installation

### From Source

1. Ensure you have Go 1.21+ installed
2. Clone the repository:
```bash
git clone https://github.com/paradoxe35/myreviser-go.git
cd myreviser-go/fyne
```

3. Install dependencies:
```bash
make install-deps
```

4. Build the application:
```bash
make build
```

### Pre-built Binaries

Download the latest release from the [GitHub Releases](https://github.com/paradoxe35/myreviser-go/releases) page.

## Configuration

On first launch, MyReviser will create a configuration file at:
- **Linux**: `~/.myreviser/config.json`
- **macOS**: `~/Library/Application Support/MyReviser/config.json`
- **Windows**: `%APPDATA%\Local\MyReviser\config.json`

### API Configuration

1. Open the application settings
2. Select your preferred AI provider (OpenAI, Claude, or Gemini)
3. Enter your API key
4. Optionally customize the model and base URL

### Hotkeys

Default hotkeys (customizable in settings):
- **Select All & Revise**: `Ctrl+Alt+Space` (Linux/Windows), `Ctrl+Option+Space` (macOS)
- **Selection & Revise**: `Ctrl+Super` (Linux), `Ctrl+Win` (Windows), `Ctrl+Cmd` (macOS)

## Usage

1. **Start the application** - It will minimize to system tray by default
2. **Select text** in any application
3. **Press the hotkey** to revise the selected text
4. The revised text will automatically replace the selection

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make package-all

# Run in development mode
make run

# Run with hot reload
make dev
```

### Testing

```bash
make test
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint
```

## System Requirements

- **Operating System**: Windows 10+, macOS 10.14+, or Linux with X11/Wayland
- **Memory**: 256MB RAM minimum
- **Display**: Any resolution
- **Network**: Internet connection required for AI API calls

## Security

- API keys are stored encrypted using system keyring when available
- No text data is stored locally
- All API communications use HTTPS

## Troubleshooting

### Application won't start
- Check if another instance is already running in the system tray
- Delete the lock file if the application crashed:
  - Linux: `~/.myreviser/.lock`
  - macOS: `~/Library/Application Support/MyReviser/.lock`
  - Windows: `%APPDATA%\Local\MyReviser\.lock`

### Hotkeys not working
- Ensure the application has accessibility permissions (macOS)
- Try running as administrator (Windows)
- Check for conflicts with other applications

### API errors
- Verify your API key is correct
- Check your internet connection
- Ensure you have available API credits

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI framework for Go
- Uses [robotgo](https://github.com/go-vgo/robotgo) for hotkey management
- Clipboard operations via [golang.design/x/clipboard](https://golang.design/x/clipboard)