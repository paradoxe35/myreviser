//go:build !darwin

package permissions

// IsSupported reports whether the current platform requires runtime permission handling.
func IsSupported() bool {
	return false
}

// CurrentState returns a state where all permissions are granted on non-macOS platforms.
func CurrentState() State {
	return State{AccessibilityGranted: true, InputMonitoringGranted: true}
}

// OpenPreference is a no-op on non-macOS platforms.
func OpenPreference(t Type) {}
