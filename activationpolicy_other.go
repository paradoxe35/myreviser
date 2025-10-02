//go:build !darwin

package main

// setActivationPolicy is a no-op on non-macOS platforms
func setActivationPolicy() {
	// No-op on Linux/Windows
}
