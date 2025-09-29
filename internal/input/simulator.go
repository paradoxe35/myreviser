package input

import (
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// SimulateSelectAll simulates the select all keyboard shortcut
func SimulateSelectAll() error {
	logger.Debug("Simulating Select All")

	switch runtime.GOOS {
	case "darwin":
		// macOS: Cmd+A
		robotgo.KeySleep = 100
		robotgo.KeyTap("a", "cmd")
	default:
		// Windows/Linux: Ctrl+A
		robotgo.KeySleep = 100
		robotgo.KeyTap("a", "ctrl")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// SimulateCopy simulates the copy keyboard shortcut
func SimulateCopy() error {
	logger.Debug("Simulating Copy")

	switch runtime.GOOS {
	case "darwin":
		// macOS: Cmd+C
		robotgo.KeySleep = 100
		robotgo.KeyTap("c", "cmd")
	default:
		// Windows/Linux: Ctrl+C
		robotgo.KeySleep = 100
		robotgo.KeyTap("c", "ctrl")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// SimulatePaste simulates the paste keyboard shortcut
func SimulatePaste() error {
	logger.Debug("Simulating Paste")

	switch runtime.GOOS {
	case "darwin":
		// macOS: Cmd+V
		robotgo.KeySleep = 100
		robotgo.KeyTap("v", "cmd")
	default:
		// Windows/Linux: Ctrl+V
		robotgo.KeySleep = 100
		robotgo.KeyTap("v", "ctrl")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// SimulateCut simulates the cut keyboard shortcut
func SimulateCut() error {
	logger.Debug("Simulating Cut")

	switch runtime.GOOS {
	case "darwin":
		// macOS: Cmd+X
		robotgo.KeySleep = 100
		robotgo.KeyTap("x", "cmd")
	default:
		// Windows/Linux: Ctrl+X
		robotgo.KeySleep = 100
		robotgo.KeyTap("x", "ctrl")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// TypeText types text character by character
func TypeText(text string) error {
	logger.Debug("Typing text", "length", len(text))

	robotgo.TypeStr(text)
	time.Sleep(50 * time.Millisecond)

	return nil
}
