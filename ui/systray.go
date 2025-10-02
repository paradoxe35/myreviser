package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser/internal/logger"
)

// SetupSystemTray sets up the system tray icon and menu
func SetupSystemTray(desk desktop.App, mainWindow *MainWindow, onQuit func() error) {
	// Set system tray menu
	menu := fyne.NewMenu("MyReviser",
		fyne.NewMenuItem("Show", func() {
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() {
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItem("View Logs", func() {
			if err := logger.OpenLogFile(); err != nil {
				logger.Error("Failed to open log file", "error", err)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			if err := onQuit(); err != nil {
				mainWindow.app.Quit()
			}
		}),
	)

	desk.SetSystemTrayMenu(menu)
}
