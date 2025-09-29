package logger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// GetLatestLogFile finds the most recent log file in the logs directory
func GetLatestLogFile() (string, error) {
	logDir := GetLogDirectory()

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", fmt.Errorf("failed to read log directory: %w", err)
	}

	// Filter log files matching pattern: myreviser-YYYY-MM-DD.log
	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "myreviser-") && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	if len(logFiles) == 0 {
		return "", fmt.Errorf("no log files found")
	}

	// Sort in reverse order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(logFiles)))

	return filepath.Join(logDir, logFiles[0]), nil
}

// OpenLogFile opens the latest log file in the system's default text editor
func OpenLogFile() error {
	// Try to get the latest log file
	logFile, err := GetLatestLogFile()
	if err != nil {
		Info("Failed to find latest log file, opening directory instead", "error", err)
		return OpenLogDirectory()
	}

	// Check if file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		Info("Log file does not exist, opening directory instead", "file", logFile)
		return OpenLogDirectory()
	}

	Info("Opening log file", "file", logFile)

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
		Warn("Failed to open log file, opening directory instead", "error", err)
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
