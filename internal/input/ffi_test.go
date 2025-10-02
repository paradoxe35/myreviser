//go:build linux || darwin || windows
// +build linux darwin windows

package input

import (
	"testing"
	"time"
)

// TestFFIClipboard tests the FFI clipboard manager
func TestFFIClipboard(t *testing.T) {
	clipboard, err := NewFFIClipboardManager()
	if err != nil {
		t.Fatalf("Failed to create clipboard: %v", err)
	}
	defer clipboard.Close()

	// Test set and get
	testText := "Hello from Rust FFI!"
	err = clipboard.SetText(testText)
	if err != nil {
		t.Fatalf("Failed to set text: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	retrievedText, err := clipboard.GetText()
	if err != nil {
		t.Fatalf("Failed to get text: %v", err)
	}

	if retrievedText != testText {
		t.Errorf("Text mismatch: expected '%s', got '%s'", testText, retrievedText)
	}

	t.Logf("✓ Clipboard test passed")
}

// TestFFISimulator tests the FFI key simulator
func TestFFISimulator(t *testing.T) {
	simulator, err := NewFFIKeySimulator()
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}
	defer simulator.Close()

	// Note: Actual key simulation requires a GUI context
	// We just test that the functions don't crash
	t.Logf("✓ Simulator created successfully")
	t.Logf("  (Actual key simulation requires GUI context)")
}

// TestFFIHotkeys tests the FFI hotkey manager
func TestFFIHotkeys(t *testing.T) {
	hotkeyMgr := NewFFIHotkeyManager()
	if hotkeyMgr == nil {
		t.Fatal("Failed to create hotkey manager")
	}
	defer hotkeyMgr.Close()

	// Register a test hotkey
	err := hotkeyMgr.RegisterHotkey("ctrl+alt+t", "test", func() {
		t.Log("Hotkey callback called!")
	})

	if err != nil {
		t.Fatalf("Failed to register hotkey: %v", err)
	}

	// Start listening
	err = hotkeyMgr.Start()
	if err != nil {
		t.Fatalf("Failed to start hotkey manager: %v", err)
	}

	// Stop immediately (we don't wait for actual keypresses)
	hotkeyMgr.Stop()

	t.Logf("✓ Hotkey manager test passed")
	t.Logf("  (Hotkey registered and manager started/stopped successfully)")
}
