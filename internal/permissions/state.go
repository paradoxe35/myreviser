package permissions

// Type identifies a macOS permission the application depends on.
type Type string

const (
	// Accessibility allows the app to control input events and access the clipboard.
	Accessibility Type = "accessibility"
	// InputMonitoring allows listening for global hotkeys.
	InputMonitoring Type = "input_monitoring"
)

// DisplayName returns a human readable label for the permission.
func (t Type) DisplayName() string {
	switch t {
	case Accessibility:
		return "Accessibility"
	case InputMonitoring:
		return "Input Monitoring"
	default:
		return string(t)
	}
}

// State captures the current macOS privacy permission status we care about.
type State struct {
	AccessibilityGranted   bool
	InputMonitoringGranted bool
}

// AllGranted reports whether the application has all required permissions.
func (s State) AllGranted() bool {
	return s.AccessibilityGranted && s.InputMonitoringGranted
}

// Granted reports whether the given permission is granted.
func (s State) Granted(t Type) bool {
	switch t {
	case Accessibility:
		return s.AccessibilityGranted
	case InputMonitoring:
		return s.InputMonitoringGranted
	default:
		return false
	}
}

// Missing returns the list of permissions that are not currently granted.
func (s State) Missing() []Type {
	missing := make([]Type, 0, 2)
	if !s.AccessibilityGranted {
		missing = append(missing, Accessibility)
	}
	if !s.InputMonitoringGranted {
		missing = append(missing, InputMonitoring)
	}
	return missing
}
