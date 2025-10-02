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

// Open System Preferences to Accessibility settings
void OpenAccessibilityPreferences(void) {
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility";
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
		logger.Warn("  - Monitor global hotkeys (Ctrl+Alt+Space, Ctrl+Win)")
		logger.Warn("  - Simulate keyboard input (Ctrl+C, Ctrl+V)")
		logger.Warn("  - Access clipboard for text revision")
		logger.Warn("Please grant accessibility permissions in System Preferences")
	} else {
		logger.Info("Accessibility permissions granted")
	}

	return trusted
}

// openAccessibilityPreferences opens System Preferences to Accessibility settings
func openAccessibilityPreferences() {
	logger.Info("Opening System Preferences > Privacy & Security > Accessibility")
	C.OpenAccessibilityPreferences()
}
