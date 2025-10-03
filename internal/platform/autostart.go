package platform

// AutoStart provides cross-platform auto-start functionality
type AutoStart interface {
	Enable() error
	Disable() error
	IsEnabled() bool
}

// GetAutoStart returns the platform-specific auto-start implementation
func GetAutoStart() AutoStart {
	return &autoStart{}
}
