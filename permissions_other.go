//go:build !darwin

package main

// checkAccessibilityPermissions is a no-op on non-macOS platforms
func checkAccessibilityPermissions() bool {
	// Always return true on Linux/Windows (no permission prompt needed)
	return true
}

// openAccessibilityPreferences is a no-op on non-macOS platforms
func openAccessibilityPreferences() {
	// No-op on Linux/Windows
}
