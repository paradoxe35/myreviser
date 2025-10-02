//go:build !darwin

package main

// showInDock is a no-op on non-macOS platforms
func showInDock() {
	// No-op on Linux/Windows
}

// hideFromDock is a no-op on non-macOS platforms
func hideFromDock() {
	// No-op on Linux/Windows
}
