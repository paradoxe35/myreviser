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

	// Start the hook
	logger.Info("Starting hotkey listener")
	h.hook = gohook.Start()
	defer gohook.End()

	logger.Info("Hotkey listener initialized successfully")

	// Register hotkeys
	h.registerHotkey(h.selectAllBinding, "select_all")
	h.registerHotkey(h.selectionBinding, "selection")

	// Process events
	for {
		select {
		case <-h.stopChan:
			logger.Info("Hotkey listener stopped")
			return
		case evt := <-h.hook:
			if evt.Kind == gohook.KeyDown || evt.Kind == gohook.KeyHold {
				h.handleEvent(evt)
			}
		}
	}
}

// registerHotkey registers a single hotkey
func (h *HotkeyManager) registerHotkey(binding, action string) {
	if binding == "" {
		return
	}

	keys := parseHotkeyString(binding)
	if len(keys) == 0 {
		logger.Error("Invalid hotkey binding", "binding", binding)
		return
	}

	// Register with gohook
	gohook.Register(gohook.KeyDown, keys, func(e gohook.Event) {
		h.triggerHandler(action)
	})

	logger.Info("Hotkey registered", "binding", binding, "action", action)
}

// handleEvent processes keyboard events
func (h *HotkeyManager) handleEvent(evt gohook.Event) {
	// Event handling is done via registered callbacks
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