//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

// Check if accessibility permissions are granted
bool CheckAccessibilityPermissions(void) {
    // AXIsProcessTrusted returns true if the app has accessibility permissions
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean trusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    return trusted;
}

// Dummy callback for permission check
static CGEventRef DummyEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    return event;
}

// Check if Input Monitoring permissions are granted
// Returns true if the app can listen to keyboard events
bool CheckInputMonitoringPermissions(void) {
    // This checks if the app is in the Input Monitoring list
    // There's no direct API, but we can check if CGEventTapCreate works
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

// Open System Preferences to Accessibility settings
void OpenAccessibilityPreferences(void) {
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility";
    [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:urlString]];
}

// Open System Preferences to Input Monitoring settings
void OpenInputMonitoringPreferences(void) {
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent";
    [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:urlString]];
}
*/
import "C"

import (
	"github.com/paradoxe35/myreviser/internal/logger"
)

// checkAccessibilityPermissions checks if the app has accessibility permissions
// and prompts the user to grant them if not
func checkAccessibilityPermissions() bool {
	trusted := bool(C.CheckAccessibilityPermissions())

	if !trusted {
		logger.Warn("Accessibility permissions not granted. The app needs accessibility permissions to:")
		logger.Warn("  - Simulate keyboard input (Ctrl+C, Ctrl+V)")
		logger.Warn("  - Access clipboard for text revision")
		logger.Warn("Please grant accessibility permissions in System Preferences")
		logger.Warn("Go to: System Settings > Privacy & Security > Accessibility")
	} else {
		logger.Info("Accessibility permissions granted")
	}

	return trusted
}

// checkInputMonitoringPermissions checks if the app has input monitoring permissions
// Required for global hotkey listening on macOS
func checkInputMonitoringPermissions() bool {
	hasPermission := bool(C.CheckInputMonitoringPermissions())

	if !hasPermission {
		logger.Warn("Input Monitoring permissions not granted. The app needs Input Monitoring to:")
		logger.Warn("  - Listen for global hotkeys (Ctrl+Alt+Space, Ctrl+Win)")
		logger.Warn("  - Detect keyboard shortcuts anywhere on your system")
		logger.Warn("Please grant Input Monitoring permissions in System Settings")
		logger.Warn("Go to: System Settings > Privacy & Security > Input Monitoring")
		logger.Warn("Then add MyReviser to the list and restart the app")
	} else {
		logger.Info("Input Monitoring permissions granted")
	}

	return hasPermission
}

// openAccessibilityPreferences opens System Preferences to Accessibility settings
func openAccessibilityPreferences() {
	logger.Info("Opening System Settings > Privacy & Security > Accessibility")
	C.OpenAccessibilityPreferences()
}

// openInputMonitoringPreferences opens System Preferences to Input Monitoring settings
func openInputMonitoringPreferences() {
	logger.Info("Opening System Settings > Privacy & Security > Input Monitoring")
	C.OpenInputMonitoringPreferences()
}
