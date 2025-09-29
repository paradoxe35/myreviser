package input

import (
	"fmt"
	"strings"
	"sync"

	"github.com/paradoxe35/myreviser-go/internal/logger"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

// HotkeyManager manages global system-wide hotkeys for the application
type HotkeyManager struct {
	mu               sync.RWMutex
	selectAllBinding string
	selectionBinding string
	handlers         map[string]func()
	hotkeys          map[string]*hotkey.Hotkey
	active           bool
	stopChan         chan struct{}
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	return &HotkeyManager{
		handlers: make(map[string]func()),
		hotkeys:  make(map[string]*hotkey.Hotkey),
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
	if h.active {
		h.mu.Unlock()
		return fmt.Errorf("hotkey manager already active")
	}
	h.active = true
	h.mu.Unlock()

	logger.Info("Starting hotkey manager")

	// Register and listen for hotkeys
	go h.registerAndListen()

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

	// Unregister all hotkeys
	for action, hk := range h.hotkeys {
		if err := hk.Unregister(); err != nil {
			logger.Error("Failed to unregister hotkey", "action", action, "error", err)
		} else {
			logger.Info("Hotkey unregistered", "action", action)
		}
	}

	logger.Info("Hotkey manager stopped")
}

// registerAndListen registers and listens for all hotkeys
func (h *HotkeyManager) registerAndListen() {
	// Initialize mainthread for hotkey library
	mainthread.Init(func() {
		h.mu.RLock()
		selectAllBinding := h.selectAllBinding
		selectionBinding := h.selectionBinding
		h.mu.RUnlock()

		// Register select all hotkey
		if selectAllBinding != "" {
			h.registerHotkey(selectAllBinding, "select_all")
		}

		// Register selection hotkey
		if selectionBinding != "" {
			h.registerHotkey(selectionBinding, "selection")
		}

		logger.Info("Hotkeys registered, listening for events")
	})
}

// registerHotkey registers a single system-wide hotkey
func (h *HotkeyManager) registerHotkey(binding, action string) {
	logger.Info("Registering system-wide hotkey", "binding", binding, "action", action)

	modifiers, key, err := parseHotkeyBinding(binding)
	if err != nil {
		logger.Error("Failed to parse hotkey binding", "binding", binding, "error", err)
		return
	}

	logger.Info("Parsed hotkey", "binding", binding, "modifiers", modifiers, "key", key)

	// Create hotkey
	hk := hotkey.New(modifiers, key)

	// Register the hotkey
	if err := hk.Register(); err != nil {
		logger.Error("Failed to register hotkey", "binding", binding, "error", err)
		return
	}

	// Store hotkey for later unregistration
	h.mu.Lock()
	h.hotkeys[action] = hk
	h.mu.Unlock()

	logger.Info("Hotkey registered successfully", "binding", binding, "action", action)

	// Listen for hotkey events in a goroutine
	go func() {
		for {
			select {
			case <-h.stopChan:
				return
			case <-hk.Keydown():
				logger.Info("Hotkey triggered", "action", action, "binding", binding)
				h.triggerHandler(action)
			}
		}
	}()
}

// triggerHandler triggers a registered handler
func (h *HotkeyManager) triggerHandler(action string) {
	h.mu.RLock()
	handler, exists := h.handlers[action]
	h.mu.RUnlock()

	if exists && handler != nil {
		logger.Info("Executing handler", "action", action)
		go handler() // Run handler in goroutine to avoid blocking
	} else {
		logger.Warn("No handler registered", "action", action)
	}
}

// parseHotkeyBinding parses a hotkey string like "ctrl+alt+space" into modifiers and key
func parseHotkeyBinding(binding string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(binding)), "+")
	if len(parts) == 0 {
		return nil, 0, fmt.Errorf("empty hotkey binding")
	}

	var modifiers []hotkey.Modifier
	var key hotkey.Key
	var keyFound bool

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if it's a modifier
		mod, isModifier := parseModifier(part)
		if isModifier {
			modifiers = append(modifiers, mod)
		} else {
			// It's a key
			if keyFound {
				return nil, 0, fmt.Errorf("multiple keys specified: %s", binding)
			}
			parsedKey, err := parseKey(part)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid key '%s': %w", part, err)
			}
			key = parsedKey
			keyFound = true
		}
	}

	if !keyFound {
		return nil, 0, fmt.Errorf("no key specified in binding: %s", binding)
	}

	return modifiers, key, nil
}

// parseKey converts a key string to hotkey.Key
func parseKey(keyStr string) (hotkey.Key, error) {
	switch keyStr {
	// Letters
	case "a":
		return hotkey.KeyA, nil
	case "b":
		return hotkey.KeyB, nil
	case "c":
		return hotkey.KeyC, nil
	case "d":
		return hotkey.KeyD, nil
	case "e":
		return hotkey.KeyE, nil
	case "f":
		return hotkey.KeyF, nil
	case "g":
		return hotkey.KeyG, nil
	case "h":
		return hotkey.KeyH, nil
	case "i":
		return hotkey.KeyI, nil
	case "j":
		return hotkey.KeyJ, nil
	case "k":
		return hotkey.KeyK, nil
	case "l":
		return hotkey.KeyL, nil
	case "m":
		return hotkey.KeyM, nil
	case "n":
		return hotkey.KeyN, nil
	case "o":
		return hotkey.KeyO, nil
	case "p":
		return hotkey.KeyP, nil
	case "q":
		return hotkey.KeyQ, nil
	case "r":
		return hotkey.KeyR, nil
	case "s":
		return hotkey.KeyS, nil
	case "t":
		return hotkey.KeyT, nil
	case "u":
		return hotkey.KeyU, nil
	case "v":
		return hotkey.KeyV, nil
	case "w":
		return hotkey.KeyW, nil
	case "x":
		return hotkey.KeyX, nil
	case "y":
		return hotkey.KeyY, nil
	case "z":
		return hotkey.KeyZ, nil

	// Numbers
	case "0":
		return hotkey.Key0, nil
	case "1":
		return hotkey.Key1, nil
	case "2":
		return hotkey.Key2, nil
	case "3":
		return hotkey.Key3, nil
	case "4":
		return hotkey.Key4, nil
	case "5":
		return hotkey.Key5, nil
	case "6":
		return hotkey.Key6, nil
	case "7":
		return hotkey.Key7, nil
	case "8":
		return hotkey.Key8, nil
	case "9":
		return hotkey.Key9, nil

	// Special keys
	case "space":
		return hotkey.KeySpace, nil
	case "enter", "return":
		return hotkey.KeyReturn, nil
	case "tab":
		return hotkey.KeyTab, nil
	case "delete", "backspace":
		return hotkey.KeyDelete, nil
	case "escape", "esc":
		return hotkey.KeyEscape, nil
	case "up":
		return hotkey.KeyUp, nil
	case "down":
		return hotkey.KeyDown, nil
	case "left":
		return hotkey.KeyLeft, nil
	case "right":
		return hotkey.KeyRight, nil

	// Function keys
	case "f1":
		return hotkey.KeyF1, nil
	case "f2":
		return hotkey.KeyF2, nil
	case "f3":
		return hotkey.KeyF3, nil
	case "f4":
		return hotkey.KeyF4, nil
	case "f5":
		return hotkey.KeyF5, nil
	case "f6":
		return hotkey.KeyF6, nil
	case "f7":
		return hotkey.KeyF7, nil
	case "f8":
		return hotkey.KeyF8, nil
	case "f9":
		return hotkey.KeyF9, nil
	case "f10":
		return hotkey.KeyF10, nil
	case "f11":
		return hotkey.KeyF11, nil
	case "f12":
		return hotkey.KeyF12, nil

	default:
		return 0, fmt.Errorf("unknown key: %s", keyStr)
	}
}

// IsActive returns whether the hotkey manager is active
func (h *HotkeyManager) IsActive() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.active
}