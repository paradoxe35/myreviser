package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// SetupSystemTray sets up the system tray icon and menu
func SetupSystemTray(desk desktop.App, mainWindow *MainWindow) {
	logger.Info("Setting up system tray")

	// Set system tray menu
	menu := fyne.NewMenu("MyReviser",
		fyne.NewMenuItem("Show", func() {
			logger.Info("Show menu item clicked")
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() {
			logger.Info("Settings menu item clicked")
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItem("View Logs", func() {
			logger.Info("View Logs menu item clicked")
			if err := logger.OpenLogFile(); err != nil {
				logger.Error("Failed to open log file", "error", err)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			logger.Info("Quit menu item clicked")
			fyne.CurrentApp().Quit()
		}),
	)

	if menu == nil {
		logger.Error("Failed to create menu")
		return
	}

	desk.SetSystemTrayMenu(menu)
	logger.Info("System tray menu configured successfully")
}