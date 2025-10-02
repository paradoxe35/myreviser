//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void SetActivationPolicy(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
*/
import "C"

import "github.com/paradoxe35/myreviser/internal/logger"

func setActivationPolicy() {
	logger.Info("Setting macOS activation policy to Accessory (hide from Dock when no windows)")
	C.SetActivationPolicy()
}
