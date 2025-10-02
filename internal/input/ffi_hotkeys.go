//go:build linux || darwin || windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a -lpthread -ldl -lm

// macOS
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security -framework AppKit -framework ApplicationServices

// Windows
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libmyreviser_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -static

#include <stdlib.h>
#include "bindings.h"

// Callback wrapper - Go can't pass Go functions to C directly
// This is called from Rust
extern void hotkeyCallbackGateway(char* action);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// FFIHotkeyManager wraps the Rust FFI hotkey manager
type FFIHotkeyManager struct {
	mu       sync.RWMutex
	handle   C.HotkeyManagerHandle
	handlers map[string]func()
	active   bool
}

// Global instance for callback routing
var globalFFIHotkeyManager *FFIHotkeyManager
var globalFFIMu sync.Mutex

// NewFFIHotkeyManager creates a new FFI-based hotkey manager
func NewFFIHotkeyManager() *FFIHotkeyManager {
	handle := C.myreviser_hotkey_manager_new()
	if handle == nil {
		logger.Error("Failed to create FFI hotkey manager")
		return nil
	}

	manager := &FFIHotkeyManager{
		handle:   handle,
		handlers: make(map[string]func()),
	}

	// Set as global for callback routing
	globalFFIMu.Lock()
	globalFFIHotkeyManager = manager
	globalFFIMu.Unlock()

	return manager
}

// SetBindings sets the hotkey bindings
func (h *FFIHotkeyManager) SetBindings(selectAll, selection string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	logger.Info("FFI: Setting hotkey bindings",
		"select_all", selectAll,
		"selection", selection)

	// Note: We'll register these when handlers are set
}

// RegisterHandler registers a handler for a specific action
func (h *FFIHotkeyManager) RegisterHandler(action string, handler func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[action] = handler
	logger.Info("FFI: Handler registered", "action", action)
}

// RegisterHotkey registers a hotkey with an action and handler
func (h *FFIHotkeyManager) RegisterHotkey(binding, action string, handler func()) error {
	if h.handle == nil {
		return fmt.Errorf("hotkey manager not initialized")
	}

	h.mu.Lock()
	h.handlers[action] = handler
	h.mu.Unlock()

	cBinding := C.CString(binding)
	cAction := C.CString(action)
	defer C.free(unsafe.Pointer(cBinding))
	defer C.free(unsafe.Pointer(cAction))

	// Register with Rust FFI
	result := C.myreviser_hotkey_register(
		h.handle,
		cBinding,
		cAction,
		C.HotkeyCallback(C.hotkeyCallbackGateway),
	)

	if result != 0 {
		return fmt.Errorf("failed to register hotkey '%s': %s", binding, getLastError())
	}

	logger.Info("FFI: Hotkey registered", "binding", binding, "action", action)
	return nil
}

// Start starts listening for hotkeys
func (h *FFIHotkeyManager) Start() error {
	h.mu.Lock()
	if h.active {
		h.mu.Unlock()
		return fmt.Errorf("hotkey manager already active")
	}
	h.mu.Unlock()

	if h.handle == nil {
		return fmt.Errorf("hotkey manager not initialized")
	}

	logger.Info("FFI: Starting hotkey manager")

	result := C.myreviser_hotkey_start(h.handle)
	if result != 0 {
		return fmt.Errorf("failed to start hotkey manager: %s", getLastError())
	}

	h.mu.Lock()
	h.active = true
	h.mu.Unlock()

	logger.Info("FFI: Hotkey manager started")
	return nil
}

// Stop stops listening for hotkeys
func (h *FFIHotkeyManager) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.active {
		return
	}

	if h.handle == nil {
		return
	}

	logger.Info("FFI: Stopping hotkey manager")

	result := C.myreviser_hotkey_stop(h.handle)
	if result != 0 {
		logger.Error("FFI: Failed to stop hotkey manager", "error", getLastError())
	}

	h.active = false
	logger.Info("FFI: Hotkey manager stopped")
}

// IsActive returns whether the hotkey manager is active
func (h *FFIHotkeyManager) IsActive() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.active
}

// Close frees the hotkey manager resources
func (h *FFIHotkeyManager) Close() {
	h.Stop()

	if h.handle != nil {
		C.myreviser_hotkey_manager_free(h.handle)
		h.handle = nil
	}

	// Clear global reference
	globalFFIMu.Lock()
	if globalFFIHotkeyManager == h {
		globalFFIHotkeyManager = nil
	}
	globalFFIMu.Unlock()
}

// Disable and Enable are not supported in FFI version yet
// They would require additional Rust FFI functions
func (h *FFIHotkeyManager) Disable() {
	logger.Warn("FFI: Disable() not yet implemented in FFI version")
}

func (h *FFIHotkeyManager) Enable() {
	logger.Warn("FFI: Enable() not yet implemented in FFI version")
}

// hotkeyCallbackGateway is called from Rust when a hotkey is triggered
// This function must be exported for C
//
//export hotkeyCallbackGateway
func hotkeyCallbackGateway(action *C.char) {
	globalFFIMu.Lock()
	manager := globalFFIHotkeyManager
	globalFFIMu.Unlock()

	if manager == nil {
		logger.Warn("FFI: Hotkey triggered but no global manager set")
		return
	}

	actionStr := C.GoString(action)

	manager.mu.RLock()
	handler, exists := manager.handlers[actionStr]
	manager.mu.RUnlock()

	if exists && handler != nil {
		logger.Info("FFI: Hotkey triggered, executing handler", "action", actionStr)
		go handler() // Run in goroutine to avoid blocking Rust
	} else {
		logger.Warn("FFI: No handler for action", "action", actionStr)
	}
}
