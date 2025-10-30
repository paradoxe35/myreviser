package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser/internal/logger"
)

func SetupSystemTray(desk desktop.App, mainWindow *MainWindow, onQuit func() error) {
	menu := fyne.NewMenu("MyReviser - AI Text Revision Tool",
		fyne.NewMenuItem("Show", func() {
			mainWindow.ShowWindow()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() {
			mainWindow.ShowWindow()
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
