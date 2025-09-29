package ui

import (
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HotkeyCapture is a custom widget for capturing keyboard shortcuts using Fyne's keyboard events
type HotkeyCapture struct {
	widget.BaseWidget
	binding    binding.String
	label      *widget.Label
	captureBtn *widget.Button
	clearBtn   *widget.Button
	container  *fyne.Container
	entry      *captureEntry // Hidden entry to capture keyboard events

	isCapturing bool
	mu          sync.Mutex

	pressedKeys map[fyne.KeyName]bool
	modifiers   map[fyne.KeyModifier]bool
}

// captureEntry is a hidden entry widget that captures keyboard events
type captureEntry struct {
	widget.Entry
	parent *HotkeyCapture
}

func (e *captureEntry) TypedKey(key *fyne.KeyEvent) {
	if e.parent != nil {
		e.parent.handleKeyPress(key)
	}
}

// NewHotkeyCapture creates a new hotkey capture widget
func NewHotkeyCapture(binding binding.String, placeholder string) *HotkeyCapture {
	h := &HotkeyCapture{
		binding:     binding,
		pressedKeys: make(map[fyne.KeyName]bool),
		modifiers:   make(map[fyne.KeyModifier]bool),
	}

	// Create label to display current/captured hotkey
	h.label = widget.NewLabel(placeholder)
	h.label.TextStyle.Monospace = true

	// Update label when binding changes
	currentValue, _ := binding.Get()
	if currentValue != "" {
		h.label.SetText(currentValue)
	}

	// Create hidden entry for keyboard capture
	h.entry = &captureEntry{parent: h}
	h.entry.PlaceHolder = "Press keys when capturing..."
	h.entry.Disable()

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
		h.entry, // Hidden entry in center for keyboard capture
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

	// Reset pressed keys
	h.pressedKeys = make(map[fyne.KeyName]bool)
	h.modifiers = make(map[fyne.KeyModifier]bool)

	// Update UI
	h.label.SetText("Press keys... (ESC to cancel)")
	h.captureBtn.SetText("Listening...")
	h.captureBtn.Disable()

	// Enable and focus the entry to capture keys
	h.entry.Enable()
	h.entry.SetText("")
	// Note: Canvas focus would be set by the window
}

// stopCapture stops listening for keys
func (h *HotkeyCapture) stopCapture() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isCapturing {
		return
	}

	h.isCapturing = false

	// Update UI
	h.entry.Disable()
	h.captureBtn.SetText("Capture")
	h.captureBtn.Enable()
}

// handleKeyPress processes key press events from Fyne
func (h *HotkeyCapture) handleKeyPress(key *fyne.KeyEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isCapturing {
		return
	}

	// Handle ESC to cancel
	if key.Name == fyne.KeyEscape {
		h.mu.Unlock()
		h.stopCapture()
		h.mu.Lock()
		currentValue, _ := h.binding.Get()
		if currentValue != "" {
			h.label.SetText(currentValue)
		} else {
			h.label.SetText("Press 'Capture' to set hotkey")
		}
		return
	}

	// Handle Enter to save
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		h.mu.Unlock()
		h.saveHotkey()
		h.stopCapture()
		h.mu.Lock()
		return
	}

	// Check if this is a modifier key press and track it
	switch key.Name {
	case desktop.KeyShiftLeft, desktop.KeyShiftRight:
		h.modifiers[fyne.KeyModifierShift] = true
	case desktop.KeyControlLeft, desktop.KeyControlRight:
		h.modifiers[fyne.KeyModifierControl] = true
	case desktop.KeyAltLeft, desktop.KeyAltRight:
		h.modifiers[fyne.KeyModifierAlt] = true
	case desktop.KeySuperLeft, desktop.KeySuperRight:
		h.modifiers[fyne.KeyModifierSuper] = true
	default:
		// Track the actual key (not modifiers)
		if !isModifierKey(key.Name) {
			h.pressedKeys[key.Name] = true
		}
	}

	h.updateDisplay()
}

// updateDisplay updates the label with current pressed keys
func (h *HotkeyCapture) updateDisplay() {
	parts := []string{}

	// Add modifiers in standard order
	if h.modifiers[fyne.KeyModifierControl] {
		parts = append(parts, "ctrl")
	}
	if h.modifiers[fyne.KeyModifierAlt] {
		parts = append(parts, getAltName())
	}
	if h.modifiers[fyne.KeyModifierShift] {
		parts = append(parts, "shift")
	}
	if h.modifiers[fyne.KeyModifierSuper] {
		parts = append(parts, getSuperName())
	}

	// Add regular keys
	for keyName := range h.pressedKeys {
		parts = append(parts, keyNameToString(keyName))
	}

	if len(parts) > 0 {
		h.label.SetText(strings.Join(parts, "+"))
	}
}

// saveHotkey saves the captured hotkey combination
func (h *HotkeyCapture) saveHotkey() {
	parts := []string{}

	// Add modifiers in standard order
	if h.modifiers[fyne.KeyModifierControl] {
		parts = append(parts, "ctrl")
	}
	if h.modifiers[fyne.KeyModifierAlt] {
		parts = append(parts, getAltName())
	}
	if h.modifiers[fyne.KeyModifierShift] {
		parts = append(parts, "shift")
	}
	if h.modifiers[fyne.KeyModifierSuper] {
		parts = append(parts, getSuperName())
	}

	// Add regular keys
	for keyName := range h.pressedKeys {
		parts = append(parts, keyNameToString(keyName))
	}

	// Must have at least one modifier and one regular key
	if len(h.modifiers) == 0 || len(h.pressedKeys) == 0 {
		h.label.SetText("Invalid combination (need modifier + key)")
		return
	}

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

// isModifierKey checks if the key is a modifier key
func isModifierKey(key fyne.KeyName) bool {
	return key == desktop.KeyShiftLeft || key == desktop.KeyShiftRight ||
		key == desktop.KeyControlLeft || key == desktop.KeyControlRight ||
		key == desktop.KeyAltLeft || key == desktop.KeyAltRight ||
		key == desktop.KeySuperLeft || key == desktop.KeySuperRight
}

// keyNameToString converts Fyne KeyName to string representation
func keyNameToString(key fyne.KeyName) string {
	// Convert key name to lowercase
	keyStr := strings.ToLower(string(key))

	// Handle special keys
	switch key {
	case fyne.KeySpace:
		return "space"
	case fyne.KeyEscape:
		return "esc"
	case fyne.KeyReturn, fyne.KeyEnter:
		return "enter"
	case fyne.KeyTab:
		return "tab"
	case fyne.KeyBackspace:
		return "backspace"
	case fyne.KeyDelete:
		return "delete"
	case fyne.KeyUp:
		return "up"
	case fyne.KeyDown:
		return "down"
	case fyne.KeyLeft:
		return "left"
	case fyne.KeyRight:
		return "right"
	default:
		// For regular keys, just return lowercase version
		if len(keyStr) == 1 {
			return keyStr
		}
		return keyStr
	}
}

// getAltName returns platform-specific alt key name
func getAltName() string {
	switch runtime.GOOS {
	case "darwin":
		return "option"
	default:
		return "alt"
	}
}

// getSuperName returns platform-specific super key name
func getSuperName() string {
	switch runtime.GOOS {
	case "darwin":
		return "cmd"
	case "windows":
		return "win"
	default:
		return "super"
	}
}