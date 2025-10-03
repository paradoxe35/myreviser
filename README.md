# MyReviser

[![Build and Release](https://github.com/paradoxe35/myreviser/actions/workflows/build.yml/badge.svg)](https://github.com/paradoxe35/myreviser/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://golang.org/dl/)

AI-powered text revision tool with global hotkeys for instant text enhancement across all applications.

## Features

- Multiple AI providers: OpenAI, Claude (Anthropic), and Google Gemini
- Global hotkeys and system tray controls that work across Windows, macOS, and Linux
- Secure API key storage with encrypted configuration
- Lightweight native Go application with Rust-powered input handling
- Configurable prompts, providers, and preferences per user profile

## Installation

- Download the [latest release](https://github.com/paradoxe35/myreviser/releases/latest) for your operating system.
- Extract or install the package that matches your platform and architecture, then run `MyReviser`.
- On the first launch, grant the application the accessibility or input permissions prompted by your OS so hotkeys work system-wide.

_For detailed build artifacts or manual installation notes, see the release description for the version you choose._

## Quick Start

1. Launch MyReviser and open the tray menu to configure your preferred AI provider and API key.
2. Set the global hotkeys you want to use for "Revise Selection" and "Select All & Revise".
3. Trigger a hotkey in any application—the selected text is automatically sent to the provider and replaced with the improved version.

## Configuration

Configuration files are created automatically after the first run.

| Platform | Path                                                  |
| -------- | ----------------------------------------------------- |
| Linux    | `~/.config/myreviser/config.json`                     |
| macOS    | `~/Library/Application Support/MyReviser/config.json` |
| Windows  | `%USERPROFILE%\.myreviser\config.json`                |

Custom base URLs are supported for OpenAI-compatible endpoints.

## Supported AI Providers

| Provider | Model Examples                                            | Base URL                                    |
| -------- | --------------------------------------------------------- | ------------------------------------------- |
| OpenAI   | `gpt-4o`, `gpt-4o-mini`                                   | `https://api.openai.com/v1`                 |
| Claude   | `claude-3-5-haiku-20241022`, `claude-3-5-sonnet-20241022` | `https://api.anthropic.com`                 |
| Gemini   | `gemini-2.5-flash`, `gemini-2.5-flash-lite`               | `https://generativelanguage.googleapis.com` |

OpenAI-compatible providers (e.g., OpenRouter, Together AI) can be configured by supplying their base URL and API key.

## Rust FFI Core

MyReviser routes global hotkeys, clipboard access, and key simulation through the Rust code in `rust-ffi/` (rdev, arboard, enigo) and exposes those capabilities to Go via CGO. This hybrid approach keeps the UI responsive while providing reliable, cross-platform system integration. For implementation notes, see `rust-ffi/README.md`.

## Development

```bash
# Build for the current platform
make build

# Build release packages for all supported platforms
make package-all

# Run the app locally
make run

# Run tests
make test
```

## Contributing

We welcome issues and pull requests. Before contributing, make sure you have:

- Go 1.24 or newer with `CGO_ENABLED=1`
- A Rust toolchain (stable) and `cargo` (the build script pulls `cbindgen` as a `[build-dependency]`, so no manual install is required)
- `make` and the platform dependencies described in `rust-ffi/README.md`

Check the `rust-ffi/` directory when working on system integrations.

## Troubleshooting

- Ensure only one instance of MyReviser is running (check the system tray).
- Review logs at `~/.myreviser/logs/` if revisions fail to send.
- Confirm the application has accessibility/input permissions when hotkeys do not trigger.

## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with [Fyne](https://fyne.io/), a cross-platform GUI framework for Go.

---

**Note:** You need valid API keys from your chosen AI provider. Sign up at the [OpenAI Platform](https://platform.openai.com/), [Anthropic Console](https://console.anthropic.com/), or [Google AI Studio](https://makersuite.google.com/).
