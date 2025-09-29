package input

import (
	"fmt"
	"sync"
	"time"

	"golang.design/x/clipboard"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// ClipboardManager manages clipboard operations
type ClipboardManager struct {
	mu           sync.Mutex
	initialized  bool
	savedContent []byte
}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager() (*ClipboardManager, error) {
	cm := &ClipboardManager{}
	if err := cm.Init(); err != nil {
		return nil, err
	}
	return cm, nil
}

// Init initializes the clipboard
func (c *ClipboardManager) Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	c.initialized = true
	logger.Info("Clipboard manager initialized")
	return nil
}

// GetText gets text from clipboard
func (c *ClipboardManager) GetText() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return "", fmt.Errorf("clipboard not initialized")
	}

	data := clipboard.Read(clipboard.FmtText)
	return string(data), nil
}

// SetText sets text to clipboard
func (c *ClipboardManager) SetText(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return fmt.Errorf("clipboard not initialized")
	}

	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// SaveCurrent saves the current clipboard content
func (c *ClipboardManager) SaveCurrent() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return fmt.Errorf("clipboard not initialized")
	}

	c.savedContent = clipboard.Read(clipboard.FmtText)
	logger.Debug("Clipboard content saved", "length", len(c.savedContent))
	return nil
}

// Restore restores the saved clipboard content
func (c *ClipboardManager) Restore() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return fmt.Errorf("clipboard not initialized")
	}

	if c.savedContent != nil {
		clipboard.Write(clipboard.FmtText, c.savedContent)
		logger.Debug("Clipboard content restored", "length", len(c.savedContent))
	}
	return nil
}

// CaptureSelectedText captures the currently selected text
func (c *ClipboardManager) CaptureSelectedText() (string, error) {
	// Save current clipboard
	if err := c.SaveCurrent(); err != nil {
		return "", fmt.Errorf("failed to save clipboard: %w", err)
	}

	// Simulate Ctrl+C to copy selected text
	if err := SimulateCopy(); err != nil {
		return "", fmt.Errorf("failed to simulate copy: %w", err)
	}

	// Wait a bit for clipboard to update
	time.Sleep(100 * time.Millisecond)

	// Get the new clipboard content
	text, err := c.GetText()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard text: %w", err)
	}

	// Restore original clipboard
	if err := c.Restore(); err != nil {
		logger.Error("Failed to restore clipboard", "error", err)
	}

	return text, nil
}

// ReplaceSelectedText replaces the selected text with new text
func (c *ClipboardManager) ReplaceSelectedText(newText string) error {
	// Save current clipboard
	if err := c.SaveCurrent(); err != nil {
		return fmt.Errorf("failed to save clipboard: %w", err)
	}

	// Set new text to clipboard
	if err := c.SetText(newText); err != nil {
		return fmt.Errorf("failed to set clipboard text: %w", err)
	}

	// Simulate Ctrl+V to paste
	if err := SimulatePaste(); err != nil {
		return fmt.Errorf("failed to simulate paste: %w", err)
	}

	// Wait a bit for paste to complete
	time.Sleep(100 * time.Millisecond)

	// Restore original clipboard
	if err := c.Restore(); err != nil {
		logger.Error("Failed to restore clipboard", "error", err)
	}

	return nil
}