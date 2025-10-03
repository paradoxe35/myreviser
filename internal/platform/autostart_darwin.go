//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/paradoxe35/myreviser/internal/logger"
)

type autoStart struct{}

// Enable creates a LaunchAgent plist file for macOS
func (a *autoStart) Enable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistPath := filepath.Join(launchAgentsDir, "me.pngwasi.myreviser.plist")

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>me.pngwasi.myreviser</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
</dict>
</plist>`, executable)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load the launch agent
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		logger.Warn("Failed to load launch agent (may already be loaded)", "error", err)
	}

	logger.Info("Auto-start enabled for macOS", "plist", plistPath)
	return nil
}

// Disable removes the LaunchAgent plist file
func (a *autoStart) Disable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "me.pngwasi.myreviser.plist")

	// Unload the launch agent if it exists
	if _, err := os.Stat(plistPath); err == nil {
		cmd := exec.Command("launchctl", "unload", plistPath)
		if err := cmd.Run(); err != nil {
			logger.Warn("Failed to unload launch agent", "error", err)
		}
	}

	// Remove the plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	logger.Info("Auto-start disabled for macOS")
	return nil
}

// IsEnabled checks if the LaunchAgent plist exists
func (a *autoStart) IsEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "me.pngwasi.myreviser.plist")
	_, err = os.Stat(plistPath)
	return err == nil
}
