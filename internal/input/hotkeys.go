package input

import (
	"fmt"
	"strings"
	"sync"

	"github.com/paradoxe35/myreviser-go/internal/logger"
	gohook "github.com/robotn/gohook"
)

// HotkeyManager manages global hotkeys for the application
type HotkeyManager struct {
	mu               sync.RWMutex
	selectAllBinding string
	selectionBinding string
	handlers         map[string]func()
	hook             chan gohook.Event
	active           bool
	stopChan         chan struct{}
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	return &HotkeyManager{
		handlers: make(map[string]func()),
		stopChan: make(chan struct{}),
	}
}

// SetBindings sets the hotkey bindings
func (h *HotkeyManager) SetBindings(selectAll, selection string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.selectAllBinding = selectAll
	h.selectionBinding = selection

	logger.Info("Hotkey bindings set",
		"select_all", selectAll,
		"selection", selection)
}

// RegisterHandler registers a handler for a specific action
func (h *HotkeyManager) RegisterHandler(action string, handler func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[action] = handler
	logger.Info("Handler registered", "action", action)
}

// Start starts listening for hotkeys
func (h *HotkeyManager) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.active {
		return fmt.Errorf("hotkey manager already active")
	}

	h.active = true
	go h.listenForHotkeys()

	logger.Info("Hotkey manager started")
	return nil
}

// Stop stops listening for hotkeys
func (h *HotkeyManager) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.active {
		return
	}

	h.active = false
	close(h.stopChan)
	gohook.End()

	logger.Info("Hotkey manager stopped")
}

// listenForHotkeys listens for configured hotkeys
func (h *HotkeyManager) listenForHotkeys() {
	// Add panic recovery for gohook failures (e.g., X11 display issues)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Hotkey listener panic recovered", "error", r)
			h.mu.Lock()
			h.active = false
			h.mu.Unlock()
		}
	}()

	// Register hotkeys BEFORE starting the hook
	logger.Info("Registering hotkeys before starting listener")
	h.registerHotkey(h.selectAllBinding, "select_all")
	h.registerHotkey(h.selectionBinding, "selection")

	// Start the hook
	logger.Info("Starting gohook event listener")
	h.hook = gohook.Start()
	defer gohook.End()

	logger.Info("Hotkey listener initialized successfully, processing events")

	// Process events - read from channel directly
	for {
		select {
		case <-h.stopChan:
			logger.Info("Hotkey listener stopped by stop signal")
			return
		case ev := <-h.hook:
			// Log all events for debugging
			logger.Debug("Event received",
				"kind", ev.Kind,
				"rawcode", ev.Rawcode,
				"keychar", string(rune(ev.Keychar)),
				"button", ev.Button,
				"mask", ev.Mask)
			// Note: Registered callbacks are triggered automatically by gohook
		}
	}
}

// registerHotkey registers a single hotkey
func (h *HotkeyManager) registerHotkey(binding, action string) {
	if binding == "" {
		logger.Warn("Empty hotkey binding", "action", action)
		return
	}

	keys := parseHotkeyString(binding)
	if len(keys) == 0 {
		logger.Error("Invalid hotkey binding", "binding", binding)
		return
	}

	logger.Info("Registering hotkey", "binding", binding, "action", action, "keys", keys)

	// Register with gohook
	gohook.Register(gohook.KeyDown, keys, func(e gohook.Event) {
		logger.Info("Hotkey callback triggered", "action", action, "binding", binding)
		h.triggerHandler(action)
	})

	logger.Info("Hotkey registered successfully", "binding", binding, "action", action, "parsed_keys", keys)
}

// triggerHandler triggers a registered handler
func (h *HotkeyManager) triggerHandler(action string) {
	h.mu.RLock()
	handler, exists := h.handlers[action]
	h.mu.RUnlock()

	if exists && handler != nil {
		logger.Info("Triggering handler", "action", action)
		go handler() // Run handler in goroutine to avoid blocking
	}
}

// parseHotkeyString parses a hotkey string like "ctrl+alt+space" into keys array
func parseHotkeyString(binding string) []string {
	parts := strings.Split(strings.ToLower(binding), "+")
	keys := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Map common names to gohook key names
		switch part {
		case "ctrl", "control":
			keys = append(keys, "ctrl")
		case "alt":
			keys = append(keys, "alt")
		case "shift":
			keys = append(keys, "shift")
		case "cmd", "command", "win", "windows", "super":
			keys = append(keys, "cmd")
		case "space":
			keys = append(keys, "space")
		default:
			// For single letters and other keys
			keys = append(keys, part)
		}
	}

	return keys
}

// IsActive returns whether the hotkey manager is active
func (h *HotkeyManager) IsActive() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.active
}