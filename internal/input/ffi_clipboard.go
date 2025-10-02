//go:build linux || darwin || windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux static linking
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a -lpthread -ldl -lm -lxdo -lX11 -lXtst

// macOS linking (partial static, frameworks required)
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security -framework AppKit -framework Carbon

// Windows static linking
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -lntdll -static

#include <stdlib.h>
#include "bindings.h"
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

// FFIClipboardManager wraps the Rust FFI clipboard manager
type FFIClipboardManager struct {
	handle C.ClipboardHandle
}

// NewFFIClipboardManager creates a new FFI-based clipboard manager
func NewFFIClipboardManager() (*FFIClipboardManager, error) {
	handle := C.myreviser_clipboard_new()
	if handle == nil {
		return nil, fmt.Errorf("failed to create clipboard manager: %s", getLastError())
	}

	return &FFIClipboardManager{handle: handle}, nil
}

// GetText gets text from clipboard
func (c *FFIClipboardManager) GetText() (string, error) {
	if c.handle == nil {
		return "", fmt.Errorf("clipboard manager not initialized")
	}

	cStr := C.myreviser_clipboard_get_text(c.handle)
	if cStr == nil {
		return "", fmt.Errorf("failed to get clipboard text: %s", getLastError())
	}
	defer C.myreviser_free_string(cStr)

	return C.GoString(cStr), nil
}

// SetText sets text to clipboard
func (c *FFIClipboardManager) SetText(text string) error {
	if c.handle == nil {
		return fmt.Errorf("clipboard manager not initialized")
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	result := C.myreviser_clipboard_set_text(c.handle, cText)
	if result != 0 {
		return fmt.Errorf("failed to set clipboard text: %s", getLastError())
	}

	return nil
}

// SaveCurrent saves the current clipboard content
func (c *FFIClipboardManager) SaveCurrent() error {
	if c.handle == nil {
		return fmt.Errorf("clipboard manager not initialized")
	}

	result := C.myreviser_clipboard_save(c.handle)
	if result != 0 {
		return fmt.Errorf("failed to save clipboard: %s", getLastError())
	}

	return nil
}

// Restore restores the saved clipboard content
func (c *FFIClipboardManager) Restore() error {
	if c.handle == nil {
		return fmt.Errorf("clipboard manager not initialized")
	}

	result := C.myreviser_clipboard_restore(c.handle)
	if result != 0 {
		return fmt.Errorf("failed to restore clipboard: %s", getLastError())
	}

	return nil
}

// Close frees the clipboard manager resources
func (c *FFIClipboardManager) Close() {
	if c.handle != nil {
		C.myreviser_clipboard_free(c.handle)
		c.handle = nil
	}
}

// CaptureSelectedText captures the currently selected text
// Note: Saves clipboard but does NOT restore - caller must restore after paste
func (c *FFIClipboardManager) CaptureSelectedText() (string, error) {
	// Save current clipboard for later restore
	if err := c.SaveCurrent(); err != nil {
		return "", fmt.Errorf("failed to save clipboard: %w", err)
	}

	// Simulate Ctrl+C to copy selected text
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return "", fmt.Errorf("failed to create simulator: %w", err)
	}
	defer sim.Close()

	if err := sim.Copy(); err != nil {
		return "", fmt.Errorf("failed to simulate copy: %w", err)
	}

	// Wait for clipboard to update after simulated Ctrl+C
	// This prevents race condition where we read before the OS updates clipboard
	time.Sleep(200 * time.Millisecond)

	// Get the captured text (now in clipboard)
	text, err := c.GetText()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard text: %w", err)
	}

	// Return the captured text
	// Original clipboard is saved and will be restored by caller after paste
	return text, nil
}

// ReplaceSelectedText replaces the selected text with new text
// Note: Does NOT save/restore clipboard - caller is responsible for that
func (c *FFIClipboardManager) ReplaceSelectedText(newText string) error {
	// Set new text to clipboard
	if err := c.SetText(newText); err != nil {
		return fmt.Errorf("failed to set clipboard text: %w", err)
	}

	// Simulate Ctrl+V to paste
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return fmt.Errorf("failed to create simulator: %w", err)
	}
	defer sim.Close()

	if err := sim.Paste(); err != nil {
		return fmt.Errorf("failed to simulate paste: %w", err)
	}

	// Wait for paste operation to complete
	// Increased delay for more reliable paste completion
	time.Sleep(300 * time.Millisecond)

	return nil
}

// getLastError retrieves the last error message from Rust
func getLastError() string {
	cErr := C.myreviser_get_last_error()
	if cErr == nil {
		return "unknown error"
	}
	defer C.myreviser_free_string((*C.char)(unsafe.Pointer(cErr)))

	return C.GoString(cErr)
}
