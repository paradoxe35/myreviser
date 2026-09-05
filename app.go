package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"
	"github.com/paradoxe35/myreviser/internal/config"
	"github.com/paradoxe35/myreviser/internal/input"
	"github.com/paradoxe35/myreviser/internal/logger"
	"github.com/paradoxe35/myreviser/internal/permissions"
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

	permissionMonitorCancel    context.CancelFunc
	permissionsMissingOnLaunch bool

	reloadMutex    sync.Mutex
	lastReloadTime time.Time
	reloadDebounce time.Duration
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
	mainWindow.SetIcon(resourceIconPng)

	application := &Application{
		app:            app,
		mainWindow:     mainWindow,
		config:         cfg,
		hotkeyManager:  hotkeyManager,
		processor:      processor,
		notifications:  notifications,
		reloadDebounce: 500 * time.Millisecond, // Debounce rapid reloads
	}

	// Setup permission monitoring before hotkeys to update the UI early
	application.setupPermissions()

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

	if desk, ok := app.(desktop.App); ok {
		desk.SetSystemTrayIcon(resourceIcon32x32Png)
		ui.SetupSystemTray(desk, mainWindow, func() error {
			application.Stop()
			return nil
		})
	}

	// Set tray tooltip
	app.Lifecycle().SetOnStarted(func() {
		systray.SetTooltip("MyReviser - AI Text Revision Tool")
		installReopenHandler(application.ShowWindow)
	})

	// Setup window close intercept
	mainWindow.SetCloseIntercept(func() {
		mainWindow.HideWindow() // Use HideWindow to trigger callbacks
	})

	return application, nil
}

// setupHotkeys configures the hotkey handlers
func (a *Application) setupHotkeys() {
	// Register select_all handler with FFI
	err := a.hotkeyManager.RegisterHotkey(a.config.Hotkeys.SelectAll, "select_all", func() {
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
	a.reportBindingFailure(a.config.Hotkeys.SelectAll, err)

	// Register selection handler with FFI
	err = a.hotkeyManager.RegisterHotkey(a.config.Hotkeys.Selection, "selection", func() {
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
	a.reportBindingFailure(a.config.Hotkeys.Selection, err)
}

// reportBindingFailure says so when a shortcut could not be registered.
//
// Silence here is the worst outcome: an unregistered binding is indistinguishable from one the
// system never delivers, and the settings screen goes on showing it as if it worked.
func (a *Application) reportBindingFailure(binding string, err error) {
	if err == nil {
		return
	}
	logger.Error("Could not register shortcut", "binding", binding, "error", err)
	a.notifications.ShowError("Shortcut not registered", binding+": "+err.Error())
}

// reloadHotkeysFromConfig stops current hotkeys and re-registers with new config
func (a *Application) reloadHotkeysFromConfig() {
	// Debounce rapid reload calls
	a.reloadMutex.Lock()
	now := time.Now()
	if now.Sub(a.lastReloadTime) < a.reloadDebounce {
		logger.Info("Skipping duplicate reload (debounced)")
		a.reloadMutex.Unlock()
		return
	}
	a.lastReloadTime = now
	a.reloadMutex.Unlock()

	logger.Info("Reloading hotkeys from updated config")

	if a.hotkeyManager == nil {
		logger.Error("Hotkey manager not initialized")
		return
	}

	// Clear all existing bindings (keep listener running to avoid spawning new threads)
	logger.Info("Clearing existing hotkey bindings")
	if err := a.hotkeyManager.ClearBindings(); err != nil {
		logger.Error("Failed to clear bindings", "error", err)
		a.notifications.ShowError("Hotkey Reload Failed", "Failed to clear old hotkeys")
		return
	}

	// Re-register hotkeys with new config
	logger.Info("Re-registering hotkeys with new config")
	a.setupHotkeys()

	// No need to stop/start - the running listener will use the updated bindings
	logger.Info("Hotkeys reloaded successfully",
		"select_all", a.config.Hotkeys.SelectAll,
		"selection", a.config.Hotkeys.Selection)
}

// setupPermissions initialises macOS permission handling and keeps the UI in sync.
func (a *Application) setupPermissions() {
	state := permissions.CurrentState()
	supported := permissions.IsSupported()
	missingOnLaunch := supported && !state.AllGranted()

	a.permissionsMissingOnLaunch = missingOnLaunch
	a.mainWindow.SetPermissionState(state, missingOnLaunch)

	if !supported {
		return
	}

	if missingOnLaunch {
		logger.Warn("macOS permissions required for full functionality")
		for _, perm := range state.Missing() {
			logger.Warn("permission pending", "name", perm.DisplayName())
		}

		ctx, cancel := context.WithCancel(context.Background())
		a.permissionMonitorCancel = cancel
		go a.monitorPermissions(ctx, state)
	} else {
		logger.Info("All required macOS permissions granted")
	}
}

func (a *Application) monitorPermissions(ctx context.Context, previous permissions.State) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			state := permissions.CurrentState()

			if state.AccessibilityGranted != previous.AccessibilityGranted {
				if state.AccessibilityGranted {
					logger.Info("Accessibility permission granted")
				} else {
					logger.Warn("Accessibility permission revoked")
				}
			}

			if state.InputMonitoringGranted != previous.InputMonitoringGranted {
				if state.InputMonitoringGranted {
					logger.Info("Input Monitoring permission granted")
				} else {
					logger.Warn("Input Monitoring permission revoked")
				}
			}

			if state != previous {
				fyne.Do(func() {
					a.mainWindow.SetPermissionState(state, a.permissionsMissingOnLaunch)
				})
				previous = state
			}

			if state.AllGranted() {
				return
			}
		}
	}
}

// ShowWindow brings the settings window up. Safe to call from any goroutine, which is what the
// instance handover needs.
func (a *Application) ShowWindow() {
	fyne.Do(a.mainWindow.ShowWindow)
}

// Start starts the application
func (a *Application) Start() error {
	// Start hotkey manager
	if err := a.hotkeyManager.Start(); err != nil {
		return fmt.Errorf("failed to start hotkey manager: %w", err)
	}

	logger.Info("Application started successfully")

	// Starting only spawns the listener thread; the system refuses the key tap after that, on that
	// thread. Without this the app reports itself healthy while no shortcut can ever fire.
	go func() {
		time.Sleep(2 * time.Second)
		if reason := a.hotkeyManager.ListenError(); reason != "" {
			logger.Error("Shortcuts will not fire", "reason", reason)
			fyne.Do(func() {
				a.notifications.ShowError("Shortcuts are not listening", reason)
			})
		}
	}()

	// Show window if permissions are pending, it's first run, or not set to start minimized
	if a.permissionsMissingOnLaunch {
		a.mainWindow.ShowWindow()
		logger.Info("Showing permissions screen", "permissions_pending", true)
	} else if a.config.Meta.FirstRun || !a.config.Appearance.StartMinimized {
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

	if a.permissionMonitorCancel != nil {
		a.permissionMonitorCancel()
	}

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
