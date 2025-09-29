package ui

import (
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	gohook "github.com/robotn/gohook"
)

// HotkeyCapture is a custom widget for capturing keyboard shortcuts
type HotkeyCapture struct {
	widget.BaseWidget
	binding    binding.String
	label      *widget.Label
	captureBtn *widget.Button
	clearBtn   *widget.Button
	container  *fyne.Container

	isCapturing bool
	mu          sync.Mutex
	stopChan    chan struct{}
	hook        chan gohook.Event

	pressedKeys map[uint16]string // Track pressed keys by rawcode
	modifiers   []string
}

// NewHotkeyCapture creates a new hotkey capture widget
func NewHotkeyCapture(binding binding.String, placeholder string) *HotkeyCapture {
	h := &HotkeyCapture{
		binding:     binding,
		pressedKeys: make(map[uint16]string),
		stopChan:    make(chan struct{}),
	}

	// Create label to display current/captured hotkey
	h.label = widget.NewLabel(placeholder)
	h.label.TextStyle.Monospace = true

	// Update label when binding changes
	currentValue, _ := binding.Get()
	if currentValue != "" {
		h.label.SetText(currentValue)
	}

	// Create capture button
	h.captureBtn = widget.NewButtonWithIcon("Capture", theme.MediaRecordIcon(), func() {
		h.startCapture()
	})

	// Create clear button
	h.clearBtn = widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		h.clearHotkey()
	})
	h.clearBtn.Importance = widget.LowImportance

	// Create container
	h.container = container.NewBorder(
		nil, nil,
		h.label,
		container.NewHBox(h.clearBtn, h.captureBtn),
		nil,
	)

	h.ExtendBaseWidget(h)
	return h
}

// CreateRenderer implements fyne.Widget
func (h *HotkeyCapture) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.container)
}

// startCapture begins listening for key combinations
func (h *HotkeyCapture) startCapture() {
	h.mu.Lock()
	if h.isCapturing {
		h.mu.Unlock()
		return
	}
	h.isCapturing = true
	h.mu.Unlock()

	// Update UI
	h.label.SetText("Press keys... (ESC to cancel)")
	h.captureBtn.SetText("Listening...")
	h.captureBtn.Disable()

	// Reset pressed keys
	h.pressedKeys = make(map[uint16]string)
	h.modifiers = []string{}

	// Start listening in goroutine
	go h.listenForKeys()
}

// stopCapture stops listening for keys
func (h *HotkeyCapture) stopCapture() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isCapturing {
		return
	}

	h.isCapturing = false

	// Stop the hook
	if h.hook != nil {
		gohook.End()
		h.hook = nil
	}

	// Update UI
	h.captureBtn.SetText("Capture")
	h.captureBtn.Enable()
}

// listenForKeys listens for keyboard events
func (h *HotkeyCapture) listenForKeys() {
	// Start hook
	h.hook = gohook.Start()
	defer func() {
		gohook.End()
		h.hook = nil
	}()

	for {
		select {
		case <-h.stopChan:
			return
		case evt := <-h.hook:
			if !h.isCapturing {
				return
			}

			h.handleKeyEvent(evt)
		}
	}
}

// handleKeyEvent processes individual key events
func (h *HotkeyCapture) handleKeyEvent(evt gohook.Event) {
	// Handle ESC to cancel
	if evt.Keychar == 27 && evt.Kind == gohook.KeyDown { // ESC
		h.stopCapture()
		currentValue, _ := h.binding.Get()
		if currentValue != "" {
			h.label.SetText(currentValue)
		} else {
			h.label.SetText("Press 'Capture' to set hotkey")
		}
		return
	}

	keyName := h.rawcodeToKeyName(evt.Rawcode, uint16(evt.Keychar))

	if evt.Kind == gohook.KeyDown {
		// Key pressed
		if keyName != "" && keyName != "unknown" {
			h.pressedKeys[evt.Rawcode] = keyName
			h.updateDisplay()
		}
	} else if evt.Kind == gohook.KeyUp {
		// Key released - save the combination
		if len(h.pressedKeys) > 0 {
			h.saveHotkey()
			h.stopCapture()
		}
	}
}

// rawcodeToKeyName converts a rawcode to a readable key name
func (h *HotkeyCapture) rawcodeToKeyName(rawcode uint16, keychar uint16) string {
	// Common key mappings (platform-specific may vary)
	switch rawcode {
	// Modifiers
	case 29, 157: // Ctrl
		return "ctrl"
	case 56, 184: // Alt
		return "alt"
	case 42, 54: // Shift
		return "shift"
	case 125, 126: // Windows/Super/Command
		return getSystemModifierName()

	// Special keys
	case 57: // Space
		return "space"
	case 1: // ESC
		return "esc"
	case 28: // Enter
		return "enter"
	case 15: // Tab
		return "tab"
	case 14: // Backspace
		return "backspace"
	case 211: // Delete
		return "delete"

	// Function keys
	case 59, 60, 61, 62, 63, 64, 65, 66, 67, 68:
		return fmt.Sprintf("f%d", rawcode-58)
	case 87, 88:
		return fmt.Sprintf("f%d", rawcode-76)

	// Arrow keys
	case 72, 200: // Up
		return "up"
	case 80, 208: // Down
		return "down"
	case 75, 203: // Left
		return "left"
	case 77, 205: // Right
		return "right"

	// Letters (a-z)
	case 30: return "a"
	case 48: return "b"
	case 46: return "c"
	case 32: return "d"
	case 18: return "e"
	case 33: return "f"
	case 34: return "g"
	case 35: return "h"
	case 23: return "i"
	case 36: return "j"
	case 37: return "k"
	case 38: return "l"
	case 50: return "m"
	case 49: return "n"
	case 24: return "o"
	case 25: return "p"
	case 16: return "q"
	case 19: return "r"
	case 31: return "s"
	case 20: return "t"
	case 22: return "u"
	case 47: return "v"
	case 17: return "w"
	case 45: return "x"
	case 21: return "y"
	case 44: return "z"

	// Numbers
	case 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
		return fmt.Sprintf("%d", rawcode-1)

	default:
		// Try to use keychar if available
		if keychar != 0 && keychar < 128 {
			return strings.ToLower(string(rune(keychar)))
		}
		return "unknown"
	}
}

// getSystemModifierName returns the platform-specific modifier name
func getSystemModifierName() string {
	// This should match the GOOS
	// For simplicity, we'll use "super" which works on Linux
	// Can be enhanced to detect platform
	return "super"
}

// updateDisplay updates the label with current pressed keys
func (h *HotkeyCapture) updateDisplay() {
	// Separate modifiers and regular keys
	modifiers := []string{}
	regularKeys := []string{}

	for _, keyName := range h.pressedKeys {
		if isModifier(keyName) {
			if !contains(modifiers, keyName) {
				modifiers = append(modifiers, keyName)
			}
		} else {
			if !contains(regularKeys, keyName) {
				regularKeys = append(regularKeys, keyName)
			}
		}
	}

	// Sort modifiers in standard order: ctrl, alt, shift, super
	sortedModifiers := sortModifiers(modifiers)

	// Build display string
	parts := append(sortedModifiers, regularKeys...)
	display := strings.Join(parts, "+")

	if display != "" {
		h.label.SetText(display)
	}
}

// saveHotkey saves the captured hotkey combination
func (h *HotkeyCapture) saveHotkey() {
	// Build final hotkey string
	modifiers := []string{}
	regularKeys := []string{}

	for _, keyName := range h.pressedKeys {
		if isModifier(keyName) {
			if !contains(modifiers, keyName) {
				modifiers = append(modifiers, keyName)
			}
		} else {
			if !contains(regularKeys, keyName) && keyName != "unknown" {
				regularKeys = append(regularKeys, keyName)
			}
		}
	}

	// Must have at least one modifier and one regular key
	if len(modifiers) == 0 || len(regularKeys) == 0 {
		h.label.SetText("Invalid combination (need modifier + key)")
		return
	}

	// Sort modifiers in standard order
	sortedModifiers := sortModifiers(modifiers)

	// Build final string
	parts := append(sortedModifiers, regularKeys...)
	hotkeyStr := strings.Join(parts, "+")

	// Save to binding
	h.binding.Set(hotkeyStr)
	h.label.SetText(hotkeyStr)
}

// clearHotkey clears the current hotkey
func (h *HotkeyCapture) clearHotkey() {
	h.binding.Set("")
	h.label.SetText("Press 'Capture' to set hotkey")
}

// Helper functions

func isModifier(key string) bool {
	return key == "ctrl" || key == "alt" || key == "shift" ||
		   key == "super" || key == "cmd" || key == "win"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sortModifiers(mods []string) []string {
	order := map[string]int{
		"ctrl":  1,
		"alt":   2,
		"shift": 3,
		"super": 4,
		"cmd":   4,
		"win":   4,
	}

	sorted := make([]string, len(mods))
	copy(sorted, mods)

	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if order[sorted[i]] > order[sorted[j]] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}