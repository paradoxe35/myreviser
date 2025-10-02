//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void SetActivationPolicyRegular(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
}

void SetActivationPolicyAccessory(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
*/
import "C"

import "github.com/paradoxe35/myreviser/internal/logger"

// showInDock makes the app appear in Dock (when window is visible)
func showInDock() {
	logger.Info("Setting macOS activation policy to Regular (show in Dock)")
	C.SetActivationPolicyRegular()
}

// hideFromDock makes the app disappear from Dock (when no windows visible)
func hideFromDock() {
	logger.Info("Setting macOS activation policy to Accessory (hide from Dock)")
	C.SetActivationPolicyAccessory()
}
