//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paradoxe35/myreviser/internal/logger"
)

type autoStart struct{}

// Enable creates an autostart desktop entry for Linux
func (a *autoStart) Enable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real executable path
	// This is important for AppImages and symlinked installations
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logger.Warn("Failed to resolve symlinks", "error", err)
		// Continue with original path if symlink resolution fails
	}

	autostartDir := filepath.Join(homeDir, ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return fmt.Errorf("failed to create autostart directory: %w", err)
	}

	desktopFilePath := filepath.Join(autostartDir, "myreviser.desktop")

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=MyReviser
Comment=AI-powered text revision tool
Exec="%s"
Icon=myreviser
Terminal=false
X-GNOME-Autostart-enabled=true
Categories=Utility;Office;
`, executable)

	if err := os.WriteFile(desktopFilePath, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	logger.Info("Auto-start enabled for Linux", "desktop_file", desktopFilePath)
	return nil
}

// Disable removes the autostart desktop entry
func (a *autoStart) Disable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	desktopFilePath := filepath.Join(homeDir, ".config", "autostart", "myreviser.desktop")

	if err := os.Remove(desktopFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove desktop file: %w", err)
	}

	logger.Info("Auto-start disabled for Linux")
	return nil
}

// IsEnabled checks if the desktop entry exists
func (a *autoStart) IsEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	desktopFilePath := filepath.Join(homeDir, ".config", "autostart", "myreviser.desktop")
	_, err = os.Stat(desktopFilePath)
	return err == nil
}
