package logger

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenLogFile opens the latest log file in the system's default text editor
func OpenLogFile() error {
	logFile := GetCurrentLogFile()

	// Platform-specific command to open file in default editor
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", logFile)
	case "darwin":
		cmd = exec.Command("open", logFile)
	case "linux":
		cmd = exec.Command("xdg-open", logFile)
	default:
		// Fallback: try to open directory
		return OpenLogDirectory()
	}

	// Try to open the log file
	if err := cmd.Start(); err != nil {
		// If failed, open the logs directory instead
		return OpenLogDirectory()
	}

	return nil
}

// OpenLogDirectory opens the logs directory in the system's file manager
func OpenLogDirectory() error {
	logDir := GetLogDirectory()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", logDir)
	case "darwin":
		cmd = exec.Command("open", logDir)
	case "linux":
		cmd = exec.Command("xdg-open", logDir)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}