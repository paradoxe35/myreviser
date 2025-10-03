//go:build darwin

package permissions

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

// Return true when the app already has Accessibility permission.
bool IsAccessibilityTrusted(void) {
    return AXIsProcessTrusted();
}

// Ask macOS to display the Accessibility permission prompt.
bool RequestAccessibilityPermissions(void) {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

// Dummy callback used to probe Input Monitoring permissions.
static CGEventRef DummyEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    return event;
}

// Return true when the app already has Input Monitoring permission.
bool HasInputMonitoringPermission(void) {
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        CGEventMaskBit(kCGEventKeyDown),
        DummyEventCallback,
        NULL
    );

    if (eventTap) {
        CFRelease(eventTap);
        return true;
    }

    return false;
}

// Open System Settings to the Accessibility section.
void OpenAccessibilityPreferences(void) {
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility";
    [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:urlString]];
}

// Open System Settings to the Input Monitoring section.
void OpenInputMonitoringPreferences(void) {
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent";
    [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:urlString]];
}
*/
import "C"

// IsSupported reports whether the current build targets macOS specific permissions.
func IsSupported() bool {
	return true
}

// CurrentState reads the latest macOS permission state.
func CurrentState() State {
	return State{
		AccessibilityGranted:   bool(C.IsAccessibilityTrusted()),
		InputMonitoringGranted: bool(C.HasInputMonitoringPermission()),
	}
}

// RequestPermission attempts to trigger the system prompt for the given permission.
// For Input Monitoring there is no programmatic prompt, so we fall back to opening preferences.
func RequestPermission(t Type) {
	switch t {
	case Accessibility:
		C.RequestAccessibilityPermissions()
	case InputMonitoring:
		OpenPreference(InputMonitoring)
	}
}

// OpenPreference opens the relevant System Settings pane for the given permission.
func OpenPreference(t Type) {
	switch t {
	case Accessibility:
		C.OpenAccessibilityPreferences()
	case InputMonitoring:
		C.OpenInputMonitoringPreferences()
	}
}
