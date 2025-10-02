package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/input"
	"github.com/paradoxe35/myreviser-go/internal/logger"
	"github.com/paradoxe35/myreviser-go/internal/revision"
	"github.com/paradoxe35/myreviser-go/ui"
)

// Application represents the main application
type Application struct {
	app           fyne.App
	mainWindow    *ui.MainWindow
	config        *config.Config
	hotkeyManager *input.FFIHotkeyManager
	processor     *revision.Processor
	notifications *ui.NotificationManager
}

// NewApplication creates a new application instance
func NewApplication(app fyne.App, cfg *config.Config) (*Application, error) {
	// Create revision processor
	processor, err := revision.NewProcessor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	// Create FFI hotkey manager
	hotkeyManager := input.NewFFIHotkeyManager()
	if hotkeyManager == nil {
		return nil, fmt.Errorf("failed to create FFI hotkey manager")
	}

	// Create notification manager
	notifications := ui.NewNotificationManager(app)

	// Create main window with hotkey manager reference
	mainWindow := ui.NewMainWindow(app, cfg, hotkeyManager)

	application := &Application{
		app:           app,
		mainWindow:    mainWindow,
		config:        cfg,
		hotkeyManager: hotkeyManager,
		processor:     processor,
		notifications: notifications,
	}

	// Setup hotkeys
	application.setupHotkeys()

	// Setup system tray if available
	if desk, ok := app.(desktop.App); ok {
		ui.SetupSystemTray(desk, mainWindow)
	}

	// Setup window close intercept
	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})

	return application, nil
}

// setupHotkeys configures the hotkey handlers
func (a *Application) setupHotkeys() {
	// Register select_all handler with FFI
	a.hotkeyManager.RegisterHotkey(a.config.Hotkeys.SelectAll, "select_all", func() {
		logger.Info("Select all hotkey triggered")

		// Check if already processing
		if a.processor.IsProcessing() {
			a.notifications.ShowInfo("Please Wait", "A revision is already in progress")
			return
		}

		// Process without showing notification unless error occurs
		if err := a.processor.ProcessSelectAll(); err != nil {
			logger.Error("Failed to process select all", "error", err)
			a.notifications.ShowError("Revision Failed", err.Error())
		} else {
			logger.Info("Text revised successfully")
		}
	})

	// Register selection handler with FFI
	a.hotkeyManager.RegisterHotkey(a.config.Hotkeys.Selection, "selection", func() {
		logger.Info("Selection hotkey triggered")

		// Check if already processing
		if a.processor.IsProcessing() {
			a.notifications.ShowInfo("Please Wait", "A revision is already in progress")
			return
		}

		// Process without showing notification unless error occurs
		if err := a.processor.ProcessSelection(); err != nil {
			logger.Error("Failed to process selection", "error", err)
			a.notifications.ShowError("Revision Failed", err.Error())
		} else {
			logger.Info("Text revised successfully")
		}
	})
}

// Start starts the application
func (a *Application) Start() error {
	// Start hotkey manager
	if err := a.hotkeyManager.Start(); err != nil {
		return fmt.Errorf("failed to start hotkey manager: %w", err)
	}

	logger.Info("Application started successfully")

	// Show window if not set to start minimized
	if !a.config.Appearance.StartMinimized {
		a.mainWindow.Show()
	}

	// Run the application
	a.app.Run()
	return nil
}

// Stop stops the application
func (a *Application) Stop() {
	logger.Info("Stopping application")

	// Stop and close FFI hotkey manager
	if a.hotkeyManager != nil {
		a.hotkeyManager.Stop()
		a.hotkeyManager.Close()
	}

	// Close processor resources
	if a.processor != nil {
		a.processor.Close()
	}

	// Quit the app
	a.app.Quit()
}
