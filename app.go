package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/myreviser/internal/config"
	"github.com/paradoxe35/myreviser/internal/input"
	"github.com/paradoxe35/myreviser/internal/logger"
	"github.com/paradoxe35/myreviser/internal/revision"
	"github.com/paradoxe35/myreviser/ui"
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

	// Listen for config changes to reload hotkeys
	config.RegisterListener(func(newCfg *config.Config) {
		logger.Info("Config changed, reloading hotkeys")
		// Update application config reference
		application.config = newCfg
		// Reload hotkeys with new bindings
		application.reloadHotkeysFromConfig()
	})

	// Set platform-specific show/hide callbacks (for macOS Dock behavior)
	mainWindow.SetShowHideCallbacks(showInDock, hideFromDock)

	// Setup system tray if available
	if desk, ok := app.(desktop.App); ok {
		ui.SetupSystemTray(desk, mainWindow, func() error {
			application.Stop()
			return nil
		})
	}

	// Setup window close intercept
	mainWindow.SetCloseIntercept(func() {
		mainWindow.HideWindow() // Use HideWindow to trigger callbacks
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

// reloadHotkeysFromConfig stops current hotkeys and re-registers with new config
func (a *Application) reloadHotkeysFromConfig() {
	logger.Info("Reloading hotkeys from updated config")

	// Stop current hotkey manager
	if a.hotkeyManager != nil {
		a.hotkeyManager.Stop()
		a.hotkeyManager.Close()
	}

	// Create new hotkey manager
	newManager := input.NewFFIHotkeyManager()
	if newManager == nil {
		logger.Error("Failed to create new hotkey manager")
		a.notifications.ShowError("Hotkey Reload Failed", "Failed to reload hotkeys")
		return
	}

	a.hotkeyManager = newManager

	// Re-register hotkeys with new config
	a.setupHotkeys()

	// Start the new hotkey manager
	if err := a.hotkeyManager.Start(); err != nil {
		logger.Error("Failed to start new hotkey manager", "error", err)
		a.notifications.ShowError("Hotkey Reload Failed", "Failed to start hotkeys: "+err.Error())
		return
	}

	logger.Info("Hotkeys reloaded successfully",
		"select_all", a.config.Hotkeys.SelectAll,
		"selection", a.config.Hotkeys.Selection)
}

// Start starts the application
func (a *Application) Start() error {
	// Start hotkey manager
	if err := a.hotkeyManager.Start(); err != nil {
		return fmt.Errorf("failed to start hotkey manager: %w", err)
	}

	logger.Info("Application started successfully")

	// Show window if first run, or if not set to start minimized
	if a.config.Meta.FirstRun || !a.config.Appearance.StartMinimized {
		a.mainWindow.ShowWindow() // Shows window and Dock icon
		// Mark first run complete and persist
		if a.config.Meta.FirstRun {
			a.config.Meta.FirstRun = false
			if err := a.config.Save(); err != nil {
				logger.Error("Failed to persist first-run flag", "error", err)
			}
			logger.Info("First run complete, window shown")
		}
	} else {
		hideFromDock() // Hide from Dock when starting minimized
		logger.Info("Starting minimized to tray")
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
