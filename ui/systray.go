package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// SetupSystemTray sets up the system tray icon and menu
func SetupSystemTray(desk desktop.App, mainWindow *MainWindow, icon fyne.Resource) {
	logger.Info("Setting up system tray")

	// Set system tray icon FIRST - REQUIRED on Windows for menu to appear
	if icon != nil {
		desk.SetSystemTrayIcon(icon)
		logger.Info("System tray icon set")
	} else {
		logger.Warn("No icon provided for system tray")
	}

	// Set system tray menu
	menu := fyne.NewMenu("MyReviser",
		fyne.NewMenuItem("Show", func() {
			logger.Info("System tray 'Show' clicked")
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() {
			logger.Info("System tray 'Settings' clicked")
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItem("View Logs", func() {
			logger.Info("System tray 'View Logs' clicked")
			if err := logger.OpenLogFile(); err != nil {
				logger.Error("Failed to open log file", "error", err)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			logger.Info("System tray 'Quit' clicked")
			mainWindow.app.Quit()
		}),
	)

	desk.SetSystemTrayMenu(menu)
	logger.Info("System tray menu set successfully")
}
