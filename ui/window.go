package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/myreviser/internal/ai"
	"github.com/paradoxe35/myreviser/internal/config"
	"github.com/paradoxe35/myreviser/internal/input"
	"github.com/paradoxe35/myreviser/internal/logger"
	"github.com/paradoxe35/myreviser/internal/permissions"
	"github.com/paradoxe35/myreviser/internal/platform"
	"github.com/paradoxe35/myreviser/internal/version"
)

const (
	// UI spacing constants
	PaddingSmall    = 5
	PaddingMedium   = 10
	PaddingLarge    = 20
	MaxWindowWidth  = 800
	MaxWindowHeight = 600
)

type MainWindow struct {
	fyne.Window
	app                 fyne.App
	config              *config.Config
	hotkeyManager       *input.FFIHotkeyManager
	permissionPrompt    *permissionPrompt
	rootContainer       *fyne.Container
	mainContent         fyne.CanvasObject
	permissionContainer fyne.CanvasObject

	// Data bindings
	providerBinding        binding.String
	apiKeyBinding          binding.String
	modelBinding           binding.String
	temperatureBinding     binding.Float
	baseURLBinding         binding.String
	hotkeySelectBinding    binding.String
	hotkeySelectionBinding binding.String
	promptBinding          binding.String
	charLimitBinding       binding.String
	statusBinding          binding.String
	startMinimizedBinding  binding.Bool
	startOnLoginBinding    binding.Bool
	themeBinding           binding.String

	// UI containers for dynamic visibility
	baseURLContainer *fyne.Container
	baseURLEntry     *widget.Entry

	// Custom provider UI components
	providerSelect       *widget.Select
	deleteProviderButton *widget.Button

	// Platform-specific callbacks
	onShowCallback func()
	onHideCallback func()
}

func NewMainWindow(app fyne.App, cfg *config.Config, hotkeyManager *input.FFIHotkeyManager) *MainWindow {
	window := app.NewWindow("MyReviser Settings")

	// Set fixed window size
	window.Resize(fyne.NewSize(650, 550))
	window.SetFixedSize(true)

	window.CenterOnScreen()
	window.SetIcon(app.Icon())

	prompt := newPermissionPrompt()

	mw := &MainWindow{
		Window:           window,
		app:              app,
		config:           cfg,
		hotkeyManager:    hotkeyManager,
		permissionPrompt: prompt,
	}

	// Set restart button callback
	prompt.restartButton.OnTapped = mw.restartApplication

	// Initialize data bindings
	mw.initBindings()

	// Apply theme from config
	themeName := cfg.Appearance.Theme
	if themeName == "" {
		themeName = "auto"
	}
	mw.applyTheme(themeName)

	// Create and set content containers
	mw.mainContent = mw.createContent()
	mw.permissionContainer = mw.permissionPrompt.canvasObject()
	mw.rootContainer = container.NewStack(mw.mainContent, mw.permissionContainer)
	window.SetContent(mw.rootContainer)
	mw.showMainContent()

	// Hide window if start minimized and not first run
	if cfg.Appearance.StartMinimized && !cfg.Meta.FirstRun {
		window.Hide()
	}

	return mw
}

func (w *MainWindow) initBindings() {
	w.providerBinding = binding.NewString()
	w.apiKeyBinding = binding.NewString()
	w.modelBinding = binding.NewString()
	w.temperatureBinding = binding.NewFloat()
	w.baseURLBinding = binding.NewString()
	w.hotkeySelectBinding = binding.NewString()
	w.hotkeySelectionBinding = binding.NewString()
	w.promptBinding = binding.NewString()
	w.charLimitBinding = binding.NewString()
	w.statusBinding = binding.NewString()
	w.startMinimizedBinding = binding.NewBool()
	w.startOnLoginBinding = binding.NewBool()
	w.themeBinding = binding.NewString()

	// Set initial values from config
	currentProvider := w.config.GetCurrentProvider()
	w.providerBinding.Set(currentProvider)

	// Load settings for current provider
	w.loadProviderSettings(currentProvider)

	w.hotkeySelectBinding.Set(w.config.Hotkeys.SelectAll)
	w.hotkeySelectionBinding.Set(w.config.Hotkeys.Selection)
	w.promptBinding.Set(w.config.Revision.SystemPrompt)
	w.charLimitBinding.Set(strconv.Itoa(w.config.Revision.CharacterLimit))
	w.statusBinding.Set("Ready")
	w.startMinimizedBinding.Set(w.config.Appearance.StartMinimized)

	// Sync StartOnLogin with actual system state
	// The user might have removed the login item outside the app
	autoStart := platform.GetAutoStart()
	actualStartOnLogin := autoStart.IsEnabled()
	w.startOnLoginBinding.Set(actualStartOnLogin)

	// Update config if out of sync
	if w.config.Appearance.StartOnLogin != actualStartOnLogin {
		w.config.Appearance.StartOnLogin = actualStartOnLogin
		w.config.Save()
		logger.Info("Synced StartOnLogin with system state", "enabled", actualStartOnLogin)
	}

	// Set theme, default to "auto" if empty
	theme := w.config.Appearance.Theme
	if theme == "" {
		theme = "auto"
	}
	w.themeBinding.Set(theme)
}

// SetPermissionState updates the permission prompt visibility and messaging.
func (w *MainWindow) SetPermissionState(state permissions.State, showRestart bool) {
	if w.permissionPrompt == nil {
		return
	}

	w.permissionPrompt.update(state, showRestart)

	if !state.AllGranted() || showRestart {
		w.showPermissionContent()
	} else {
		w.showMainContent()
	}
}

func (w *MainWindow) showPermissionContent() {
	if w.permissionContainer != nil {
		w.permissionContainer.Show()
	}
	if w.mainContent != nil {
		w.mainContent.Hide()
	}
	if w.rootContainer != nil {
		w.rootContainer.Objects = []fyne.CanvasObject{w.mainContent, w.permissionContainer}
		w.rootContainer.Refresh()
	}
}

func (w *MainWindow) showMainContent() {
	if w.mainContent != nil {
		w.mainContent.Show()
	}
	if w.permissionContainer != nil {
		w.permissionContainer.Hide()
	}
	if w.rootContainer != nil {
		w.rootContainer.Objects = []fyne.CanvasObject{w.permissionContainer, w.mainContent}
		w.rootContainer.Refresh()
	}
}

func (w *MainWindow) loadProviderSettings(provider string) {
	settings := w.config.GetProviderSettings(provider)

	apiKey, _ := w.config.GetAPIKey(provider)
	w.apiKeyBinding.Set(apiKey)
	w.modelBinding.Set(settings.Model)
	w.temperatureBinding.Set(settings.Temperature)

	isCustom := w.config.IsCustomProvider(provider)
	if isCustom {
		w.baseURLBinding.Set(settings.BaseURL)
	} else {
		w.baseURLBinding.Set("")
	}

	w.updateProviderUI(provider)
}

func (w *MainWindow) createContent() fyne.CanvasObject {
	// AI Provider Section
	providerSection := w.createProviderSection()

	// Hotkeys Section
	hotkeySection := w.createHotkeySection()

	// Revision Settings Section
	revisionSection := w.createRevisionSection()

	// System Settings Section
	systemSection := w.createSystemSection()

	// Status Bar
	statusBar := w.createStatusBar()

	// Main content with tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("AI Provider", theme.ComputerIcon(), providerSection),
		container.NewTabItemWithIcon("Hotkeys", theme.SettingsIcon(), hotkeySection),
		container.NewTabItemWithIcon("Revision", theme.DocumentIcon(), revisionSection),
		container.NewTabItemWithIcon("System", theme.ViewFullScreenIcon(), systemSection),
	)

	// Fix for AppTabs layout width issue on Windows
	// See: https://github.com/fyne-io/fyne/issues/5338
	tabs.OnSelected = func(tab *container.TabItem) {
		go func() {
			time.Sleep(50 * time.Millisecond)
			fyne.Do(func() {
				tab.Content.Refresh()
			})
		}()
	}

	// Save button
	saveBtn := widget.NewButtonWithIcon("Save Settings", theme.DocumentSaveIcon(), w.saveSettings)
	saveBtn.Importance = widget.HighImportance

	// Main layout
	content := container.NewBorder(
		nil, // top
		container.NewBorder(nil, nil, nil, saveBtn, statusBar), // bottom
		nil,  // left
		nil,  // right
		tabs, // center
	)

	return container.NewPadded(content)
}

func (w *MainWindow) createProviderSection() fyne.CanvasObject {
	selected, _ := w.providerBinding.Get()

	providerSection := w.createProviderSelectionSection(selected)
	configSection := w.createProviderConfigSection()
	testSection := w.createConnectionTestSection()

	w.updateProviderUI(selected)

	content := container.NewVBox(
		container.NewPadded(providerSection),
		widget.NewSeparator(),
		container.NewPadded(configSection),
		widget.NewSeparator(),
		container.NewPadded(testSection),
	)

	return content
}

func (w *MainWindow) createProviderSelectionSection(selected string) fyne.CanvasObject {
	providerLabel := widget.NewLabel("AI Provider:")
	providerLabel.TextStyle.Bold = true

	providerNames := w.config.GetAllProviderNames()
	providerOptions := append(providerNames, "-- Add Custom Provider --")

	w.providerSelect = widget.NewSelect(
		providerOptions,
		func(value string) {
			if value == "-- Add Custom Provider --" {
				w.showAddCustomProviderDialog()
				w.providerSelect.SetSelected(w.config.GetCurrentProvider())
				return
			}

			w.providerBinding.Set(value)
			w.loadProviderSettings(value)
		},
	)
	w.providerSelect.SetSelected(selected)

	w.deleteProviderButton = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		w.showDeleteProviderConfirmation()
	})
	w.deleteProviderButton.Importance = widget.DangerImportance

	providerRow := container.NewBorder(nil, nil, nil, w.deleteProviderButton, w.providerSelect)

	content := container.NewVBox(
		providerLabel,
		providerRow,
	)

	return content
}

func (w *MainWindow) createProviderConfigSection() fyne.CanvasObject {
	// API Key section
	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyLabel.TextStyle.Bold = true
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.Bind(w.apiKeyBinding)
	apiKeyEntry.PlaceHolder = "Enter your API key"
	apiKeyEntry.Validator = nil // Disable validation icon

	// Model section
	modelLabel := widget.NewLabel("Model:")
	modelEntry := widget.NewEntry()
	modelEntry.Bind(w.modelBinding)
	modelEntry.PlaceHolder = "e.g., gpt-4o"
	modelEntry.Validator = nil // Disable validation icon

	// Temperature section
	tempLabel := widget.NewLabel("Temp:")
	tempValue := widget.NewLabel("1.0")
	tempSlider := widget.NewSlider(0, 2)
	tempSlider.Step = 0.1
	tempSlider.Bind(w.temperatureBinding)
	tempSlider.OnChanged = func(value float64) {
		tempValue.SetText(fmt.Sprintf("%.1f", value))
	}
	// Initialize display from binding
	temp, _ := w.temperatureBinding.Get()
	tempValue.SetText(fmt.Sprintf("%.1f", temp))

	// Model and Temperature on same row
	modelTempRow := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, modelLabel, nil, modelEntry),
		container.NewBorder(nil, nil, tempLabel, tempValue, tempSlider),
	)

	// Base URL section (for custom endpoints - only for OpenAI)
	baseURLLabel := widget.NewLabel("Base URL:")
	baseURLLabel.TextStyle.Bold = true
	w.baseURLEntry = widget.NewEntry()
	w.baseURLEntry.Bind(w.baseURLBinding)
	w.baseURLEntry.PlaceHolder = "https://api.openai.com/v1 (optional)"
	w.baseURLEntry.Validator = nil // Disable validation icon

	// Create Base URL container for visibility control
	w.baseURLContainer = container.NewVBox(
		baseURLLabel,
		w.baseURLEntry,
	)

	// Configuration form with proper spacing
	configForm := container.NewVBox(
		apiKeyLabel,
		apiKeyEntry,
		widget.NewSeparator(),
		modelTempRow,
		widget.NewSeparator(),
		w.baseURLContainer,
	)

	return configForm
}

func (w *MainWindow) createConnectionTestSection() fyne.CanvasObject {
	// Test connection button
	testBtn := widget.NewButtonWithIcon("Test Connection", theme.ConfirmIcon(), func() {
		w.testAPIConnection()
	})
	testBtn.Importance = widget.MediumImportance

	// Layout with proper spacing
	return container.NewVBox(
		widget.NewLabel("Test your settings:"),
		testBtn,
	)
}

func (w *MainWindow) createHotkeySection() fyne.CanvasObject {
	// Select All Hotkey with capture widget
	selectAllLabel := widget.NewLabel("Select All & Revise:")
	selectAllLabel.TextStyle.Bold = true
	selectAllCapture := NewHotkeyCapture(w.hotkeySelectBinding, "Click 'Capture' to set hotkey")
	selectAllCapture.window = w.Window // Set window reference for focus
	// Connect callbacks to disable/enable global hotkeys during capture
	// Global hotkey library requires at least one non-modifier key on all platforms
	selectAllCapture.SetAllowModifierOnly(true)
	selectAllCapture.onCaptureStart = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Disable()
		}
	}
	selectAllCapture.onCaptureStop = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Enable()
		}
	}

	// Selection Hotkey with capture widget
	selectionLabel := widget.NewLabel("Revise Selection:")
	selectionLabel.TextStyle.Bold = true
	selectionCapture := NewHotkeyCapture(w.hotkeySelectionBinding, "Click 'Capture' to set hotkey")
	selectionCapture.window = w.Window // Set window reference for focus
	// Require at least one non-modifier key; modifier-only (e.g., Ctrl+Win) is not supported by OS/global hotkey APIs
	selectionCapture.SetAllowModifierOnly(true)

	// Connect callbacks to disable/enable global hotkeys during capture
	selectionCapture.onCaptureStart = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Disable()
		}
	}
	selectionCapture.onCaptureStop = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Enable()
		}
	}

	// Link capture widgets as siblings so only one can capture at a time
	selectAllCapture.SetSiblings(selectionCapture)
	selectionCapture.SetSiblings(selectAllCapture)

	// Hotkey instructions - collapsible accordion
	instructionsLabel := widget.NewLabel(
		"• Click 'Capture', press keys one at a time, then Enter to save\n" +
			"• Requires at least one modifier (Ctrl/Alt/Shift/Super)\n" +
			"• Press ESC to cancel")
	instructionsLabel.Wrapping = fyne.TextWrapWord

	instructions := widget.NewAccordion(
		widget.NewAccordionItem("How to Capture Hotkeys", instructionsLabel),
	)

	// Reset to defaults button
	resetBtn := widget.NewButton("Reset to Defaults", func() {
		// Stop any active captures
		selectAllCapture.StopCapture()
		selectionCapture.StopCapture()

		// Set default values
		defaults := config.GetPlatformHotkeys()
		w.hotkeySelectBinding.Set(defaults.SelectAll)
		w.hotkeySelectionBinding.Set(defaults.Selection)

		// Update displays
		selectAllCapture.UpdateFromBinding()
		selectionCapture.UpdateFromBinding()
	})

	form := container.NewVBox(
		container.NewPadded(container.NewVBox(selectAllLabel, selectAllCapture)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(selectionLabel, selectionCapture)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(instructions, resetBtn)),
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createRevisionSection() fyne.CanvasObject {
	// System Prompt
	promptLabel := widget.NewLabel("System Prompt:")
	promptLabel.TextStyle.Bold = true
	promptEntry := widget.NewMultiLineEntry()
	promptEntry.Bind(w.promptBinding)
	promptEntry.SetMinRowsVisible(5)
	promptEntry.Wrapping = fyne.TextWrapWord
	promptEntry.Validator = nil

	// Character Limit - using number entry
	charLimitLabel := widget.NewLabel("Character Limit:")
	charLimitEntry := widget.NewEntry()
	charLimitEntry.Bind(w.charLimitBinding)
	charLimitEntry.PlaceHolder = "1000"
	charLimitEntry.Validator = func(s string) error {
		if _, err := strconv.Atoi(s); err != nil && s != "" {
			return fmt.Errorf("must be a number")
		}
		return nil
	}

	// Timeout
	timeoutLabel := widget.NewLabel("Timeout (seconds):")
	timeoutSlider := widget.NewSlider(5, 60)
	timeoutSlider.SetValue(float64(w.config.Revision.TimeoutSeconds))
	timeoutValue := widget.NewLabel(fmt.Sprintf("%ds", w.config.Revision.TimeoutSeconds))
	timeoutSlider.OnChanged = func(value float64) {
		timeoutValue.SetText(fmt.Sprintf("%ds", int(value)))
	}
	timeoutSlider.OnChangeEnded = func(value float64) {
		w.config.Revision.TimeoutSeconds = int(value)
	}

	// Provider Mentions toggle
	providerMentionsLabel := widget.NewLabel("Provider Mentions:")
	providerMentionsLabel.TextStyle.Bold = true
	providerMentionsCheck := widget.NewCheck("Enable @provider mentions", func(enabled bool) {
		w.config.Revision.EnableProviderMentions = enabled
	})
	providerMentionsCheck.SetChecked(w.config.Revision.EnableProviderMentions)

	providerMentionsHelp := widget.NewLabel(
		"Add @provider before text to revise with that AI (e.g., @claude Fix this)")
	providerMentionsHelp.Wrapping = fyne.TextWrapWord
	providerMentionsHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Reset prompt button
	resetPromptBtn := widget.NewButton("Reset to Default Prompt", func() {
		w.promptBinding.Set(config.GetDefaultRevision().SystemPrompt)
	})

	form := container.NewVBox(
		container.NewPadded(container.NewVBox(promptLabel, promptEntry)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(
			container.NewBorder(nil, nil, charLimitLabel, nil, charLimitEntry),
			container.NewBorder(nil, nil, timeoutLabel, timeoutValue, timeoutSlider),
		)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(
			providerMentionsLabel,
			providerMentionsCheck,
			providerMentionsHelp,
		)),
		widget.NewSeparator(),
		container.NewPadded(resetPromptBtn),
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createSystemSection() fyne.CanvasObject {
	// Theme selection
	themeLabel := widget.NewLabel("Theme:")
	themeLabel.TextStyle.Bold = true

	themeSelect := widget.NewSelect(
		[]string{"auto", "light", "dark"},
		func(value string) {
			w.themeBinding.Set(value)
			w.applyTheme(value)
		},
	)

	// Set initial selection
	currentTheme, _ := w.themeBinding.Get()
	themeSelect.SetSelected(currentTheme)

	// Theme description
	themeDesc := widget.NewLabel("Auto: Follow system theme\nLight: Always use light theme\nDark: Always use dark theme")
	themeDesc.Wrapping = fyne.TextWrapWord

	// Start Minimized checkbox
	startMinimizedCheck := widget.NewCheck("Start minimized to system tray", func(checked bool) {
		w.startMinimizedBinding.Set(checked)
	})
	startMinimizedCheck.Bind(w.startMinimizedBinding)

	// Start on Login checkbox
	startOnLoginCheck := widget.NewCheck("Start on login", func(checked bool) {
		w.startOnLoginBinding.Set(checked)
	})
	startOnLoginCheck.Bind(w.startOnLoginBinding)

	// Version display (only show for production builds)
	var versionContainer *fyne.Container
	if version.IsProduction(w.app) {
		versionLabel := widget.NewLabel(fmt.Sprintf("Version: %s", version.GetVersion(w.app)))
		versionLabel.TextStyle.Italic = true
		// Use a subtle gray color for the version
		versionLabel.Importance = widget.LowImportance

		versionContainer = container.NewVBox(
			widget.NewSeparator(),
			container.NewPadded(versionLabel),
		)
	}

	formItems := []fyne.CanvasObject{
		container.NewPadded(container.NewVBox(themeLabel, themeSelect, themeDesc)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(startMinimizedCheck, startOnLoginCheck)),
	}

	// Add version if production build
	if versionContainer != nil {
		formItems = append(formItems, versionContainer)
	}

	form := container.NewVBox(formItems...)

	return container.NewScroll(form)
}

func (w *MainWindow) createStatusBar() fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	statusLabel.Bind(w.statusBinding)
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.Truncation = fyne.TextTruncateEllipsis

	// Create a scrollable container for long status messages
	statusScroll := container.NewScroll(statusLabel)
	statusScroll.SetMinSize(fyne.NewSize(550, 40))

	return container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.InfoIcon()),
		nil,
		statusScroll,
	)
}

func (w *MainWindow) testAPIConnection() {
	w.statusBinding.Set("Testing API connection...")

	go func() {
		provider, _ := w.providerBinding.Get()
		settings := w.config.GetProviderSettings(provider)
		apiKey, _ := w.apiKeyBinding.Get()
		model, _ := w.modelBinding.Get()
		baseURL, _ := w.baseURLBinding.Get()

		if apiKey == "" {
			fyne.Do(func() {
				w.statusBinding.Set("Error: API key is required")
			})
			return
		}

		var testProvider ai.Provider
		isCustom := w.config.IsCustomProvider(provider)

		if isCustom {
			if baseURL == "" {
				fyne.Do(func() {
					w.statusBinding.Set("Error: Base URL is required for custom providers")
				})
				return
			}
			customProvider, err := ai.NewCustomProvider(provider, settings.ProviderType, apiKey, baseURL, model, settings.Temperature)
			if err != nil {
				fyne.Do(func() {
					w.statusBinding.Set(fmt.Sprintf("Error: %s", err.Error()))
				})
				return
			}
			testProvider = customProvider
		} else {
			switch provider {
			case config.BuiltInOpenAI:
				testProvider = ai.NewOpenAIProvider(apiKey, baseURL, model, settings.Temperature)
			case config.BuiltInClaude:
				testProvider = ai.NewAnthropicProvider(apiKey, baseURL, model, settings.Temperature)
			case config.BuiltInGemini:
				testProvider = ai.NewGeminiProvider(apiKey, baseURL, model, settings.Temperature)
			default:
				fyne.Do(func() {
					w.statusBinding.Set("Error: Unknown provider")
				})
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		testText := "Hello"
		testPrompt := "Reply with 'Connection successful' if you receive this message."

		_, err := testProvider.ReviseText(ctx, testText, testPrompt)
		if err != nil {
			errMsg := err.Error()
			if len(errMsg) > 80 {
				errMsg = errMsg[:77] + "..."
			}
			fyne.Do(func() {
				w.statusBinding.Set(fmt.Sprintf("Connection failed: %s", errMsg))
			})
			logger.Error("API connection test failed", "error", err)
		} else {
			fyne.Do(func() {
				w.statusBinding.Set("Connection successful!")
			})
			logger.Info("API connection test successful",
				"provider", provider,
				"model", model,
				"temperature", settings.Temperature,
				"custom", isCustom,
			)
		}
	}()
}

func (w *MainWindow) applyTheme(themeName string) {
	switch themeName {
	case "light":
		w.app.Settings().SetTheme(&forceLight{})
	case "dark":
		w.app.Settings().SetTheme(&forceDark{})
	case "auto":
		w.app.Settings().SetTheme(theme.DefaultTheme())
	}
}

func (w *MainWindow) saveSettings() {
	provider, _ := w.providerBinding.Get()
	apiKey, _ := w.apiKeyBinding.Get()
	model, _ := w.modelBinding.Get()
	temperature, _ := w.temperatureBinding.Get()
	hotkeySelect, _ := w.hotkeySelectBinding.Get()
	hotkeySelection, _ := w.hotkeySelectionBinding.Get()
	prompt, _ := w.promptBinding.Get()
	charLimitStr, _ := w.charLimitBinding.Get()
	startMinimized, _ := w.startMinimizedBinding.Get()
	startOnLogin, _ := w.startOnLoginBinding.Get()
	themeSetting, _ := w.themeBinding.Get()
	baseURL, _ := w.baseURLBinding.Get()

	charLimit, err := strconv.Atoi(charLimitStr)
	if err != nil || charLimit <= 0 {
		w.statusBinding.Set("Error: Invalid character limit")
		return
	}

	existingSettings := w.config.GetProviderSettings(provider)
	isCustom := w.config.IsCustomProvider(provider)

	if isCustom && baseURL == "" {
		w.statusBinding.Set("Error: Base URL is required for custom providers")
		return
	}

	settings := config.ProviderSettings{
		Model:        model,
		Temperature:  temperature,
		IsCustom:     existingSettings.IsCustom,
		ProviderType: existingSettings.ProviderType,
	}

	if isCustom {
		settings.BaseURL = baseURL
	}

	w.config.SetProviderSettings(provider, settings)

	if err := w.config.SaveAPIKey(provider, apiKey); err != nil {
		errMsg := err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		w.statusBinding.Set("Error saving API key: " + errMsg)
		logger.Error("Failed to save API key", "error", err)
		return
	}

	w.config.SetCurrentProvider(provider)

	w.config.Hotkeys.SelectAll = hotkeySelect
	w.config.Hotkeys.Selection = hotkeySelection
	w.config.Revision.SystemPrompt = prompt
	w.config.Revision.CharacterLimit = charLimit
	w.config.Appearance.StartMinimized = startMinimized
	w.config.Appearance.StartOnLogin = startOnLogin
	w.config.Appearance.Theme = themeSetting

	w.applyAutoStartSetting(startOnLogin)

	if err := w.config.Save(); err != nil {
		errMsg := err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		w.statusBinding.Set("Error saving settings: " + errMsg)
		logger.Error("Failed to save settings", "error", err)
	} else {
		w.statusBinding.Set("Settings saved successfully")
		logger.Info("Settings saved")
	}
}

func (w *MainWindow) ShowAndRun() {
	w.Window.ShowAndRun()
}

// ShowWindow shows the window and handles platform-specific behavior (e.g., macOS Dock)
func (w *MainWindow) ShowWindow() {
	// Call the platform-specific show handler if set
	if w.onShowCallback != nil {
		w.onShowCallback()
	}
	w.Show()
	w.RequestFocus()

	// Force layout refresh after show to fix Windows minimize/restore sizing issue
	// See: https://github.com/fyne-io/fyne/issues/300
	w.Resize(w.Canvas().Size())
	w.Content().Refresh()
}

// HideWindow hides the window and handles platform-specific behavior (e.g., macOS Dock)
func (w *MainWindow) HideWindow() {
	w.Hide()
	// Call the platform-specific hide handler if set
	if w.onHideCallback != nil {
		w.onHideCallback()
	}
}

// SetShowHideCallbacks sets the callbacks for platform-specific show/hide behavior
func (w *MainWindow) SetShowHideCallbacks(onShow func(), onHide func()) {
	w.onShowCallback = onShow
	w.onHideCallback = onHide
}

// applyAutoStartSetting enables or disables auto-start based on the setting
func (w *MainWindow) applyAutoStartSetting(enabled bool) {
	autoStart := platform.GetAutoStart()

	if enabled {
		if err := autoStart.Enable(); err != nil {
			logger.Error("Failed to enable auto-start", "error", err)
			w.statusBinding.Set("✗ Failed to enable auto-start")
		} else {
			logger.Info("Auto-start enabled")
		}
	} else {
		if err := autoStart.Disable(); err != nil {
			logger.Error("Failed to disable auto-start", "error", err)
			w.statusBinding.Set("✗ Failed to disable auto-start")
		} else {
			logger.Info("Auto-start disabled")
		}
	}
}

// restartApplication restarts the application
func (w *MainWindow) restartApplication() {
	logger.Info("User requested application restart")

	go func() {
		time.Sleep(200 * time.Millisecond)

		err := platform.RestartApplication()
		if err != nil {
			logger.Error("Failed to restart application", "error", err)
			return
		}

		logger.Info("New instance started, quitting current instance")

		fyne.Do(func() {
			w.app.Quit()
		})
	}()
}

func (w *MainWindow) updateProviderUI(provider string) {
	isCustom := w.config.IsCustomProvider(provider)

	if w.deleteProviderButton != nil {
		if isCustom {
			w.deleteProviderButton.Show()
		} else {
			w.deleteProviderButton.Hide()
		}
	}

	if w.baseURLContainer != nil {
		if isCustom {
			w.baseURLContainer.Show()
			if w.baseURLEntry != nil {
				w.baseURLEntry.PlaceHolder = "Required for custom providers"
			}
		} else {
			w.baseURLContainer.Hide()
		}
	}

}

func (w *MainWindow) refreshProviderList() {
	if w.providerSelect == nil {
		return
	}
	providerNames := w.config.GetAllProviderNames()
	providerOptions := append(providerNames, "-- Add Custom Provider --")
	w.providerSelect.Options = providerOptions
	w.providerSelect.Refresh()
}

func (w *MainWindow) showAddCustomProviderDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.PlaceHolder = "my-custom-provider"

	baseURLEntry := widget.NewEntry()
	baseURLEntry.PlaceHolder = "https://api.example.com/v1"

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.PlaceHolder = "Enter API key"

	modelEntry := widget.NewEntry()
	modelEntry.PlaceHolder = "e.g., gpt-4 or llama3"

	errorLabel := widget.NewLabel("")
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.Importance = widget.DangerImportance

	fetchModelsBtn := widget.NewButton("Fetch Models", func() {
		w.fetchModelsForCustomProvider(apiKeyEntry.Text, baseURLEntry.Text, modelEntry, errorLabel)
	})

	form := container.NewVBox(
		widget.NewLabel("Provider Name:"),
		nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Base URL (required):"),
		baseURLEntry,
		widget.NewSeparator(),
		widget.NewLabel("API Key:"),
		apiKeyEntry,
		widget.NewSeparator(),
		widget.NewLabel("Model:"),
		container.NewBorder(nil, nil, nil, fetchModelsBtn, modelEntry),
		errorLabel,
	)

	scrollContainer := container.NewScroll(form)
	scrollContainer.SetMinSize(fyne.NewSize(400, 350))

	var d dialog.Dialog

	addBtn := widget.NewButton("Add", func() {
		// Clear previous error
		errorLabel.SetText("")

		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			errorLabel.SetText("Provider name is required")
			return
		}

		baseURL := strings.TrimSpace(baseURLEntry.Text)
		if baseURL == "" {
			errorLabel.SetText("Base URL is required")
			return
		}

		settings := config.ProviderSettings{
			BaseURL:      baseURL,
			Model:        strings.TrimSpace(modelEntry.Text),
			Temperature:  1.0,
			IsCustom:     true,
			ProviderType: config.ProviderTypeOpenAICompatible,
		}

		if err := w.config.AddCustomProvider(name, settings); err != nil {
			errorLabel.SetText(err.Error())
			return
		}

		if apiKey := strings.TrimSpace(apiKeyEntry.Text); apiKey != "" {
			if err := w.config.SaveAPIKey(name, apiKey); err != nil {
				errorLabel.SetText(fmt.Sprintf("Error saving API key: %s", err.Error()))
				return
			}
		}

		if err := w.config.Save(); err != nil {
			errorLabel.SetText(fmt.Sprintf("Error saving: %s", err.Error()))
			return
		}

		// Success - close dialog and update UI
		d.Hide()
		w.refreshProviderList()
		w.providerSelect.SetSelected(name)
		w.providerBinding.Set(name)
		w.loadProviderSettings(name)
		w.statusBinding.Set(fmt.Sprintf("Custom provider '%s' added", name))
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})

	buttons := container.NewHBox(cancelBtn, addBtn)
	content := container.NewBorder(nil, buttons, nil, nil, scrollContainer)

	d = dialog.NewCustomWithoutButtons("Add Custom Provider", content, w.Window)
	d.Resize(fyne.NewSize(450, 480))
	d.Show()
}

func (w *MainWindow) showDeleteProviderConfirmation() {
	currentProvider, _ := w.providerBinding.Get()

	if !w.config.IsCustomProvider(currentProvider) {
		dialog.ShowError(fmt.Errorf("cannot delete built-in provider"), w.Window)
		return
	}

	dialog.ShowConfirm(
		"Delete Provider",
		fmt.Sprintf("Are you sure you want to delete '%s'?\n\nThis action cannot be undone.", currentProvider),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			if err := w.config.DeleteCustomProvider(currentProvider); err != nil {
				w.statusBinding.Set(fmt.Sprintf("Error: %s", err.Error()))
				return
			}

			if err := w.config.Save(); err != nil {
				w.statusBinding.Set(fmt.Sprintf("Error saving: %s", err.Error()))
				return
			}

			w.refreshProviderList()
			newProvider := w.config.GetCurrentProvider()
			w.providerSelect.SetSelected(newProvider)
			w.providerBinding.Set(newProvider)
			w.loadProviderSettings(newProvider)
			w.statusBinding.Set(fmt.Sprintf("Provider '%s' deleted", currentProvider))
		},
		w.Window,
	)
}

func (w *MainWindow) fetchModelsForCustomProvider(apiKey, baseURL string, modelEntry *widget.Entry, statusLabel *widget.Label) {
	if apiKey == "" || baseURL == "" {
		statusLabel.SetText("API key and Base URL are required")
		return
	}

	statusLabel.SetText("Fetching models...")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		models, err := ai.FetchModels(ctx, apiKey, baseURL)

		fyne.Do(func() {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
				return
			}

			if len(models) == 0 {
				statusLabel.SetText("No models found")
				return
			}

			statusLabel.SetText(fmt.Sprintf("Found %d models", len(models)))

			modelNames := make([]string, len(models))
			for i, m := range models {
				modelNames[i] = m.ID
			}

			modelList := widget.NewList(
				func() int { return len(modelNames) },
				func() fyne.CanvasObject {
					return widget.NewLabel("model-name-placeholder")
				},
				func(id widget.ListItemID, obj fyne.CanvasObject) {
					obj.(*widget.Label).SetText(modelNames[id])
				},
			)

			var modelDialog dialog.Dialog
			modelList.OnSelected = func(id widget.ListItemID) {
				modelEntry.SetText(modelNames[id])
				modelDialog.Hide()
			}

			listContainer := container.NewScroll(modelList)
			listContainer.SetMinSize(fyne.NewSize(350, 250))

			modelDialog = dialog.NewCustom(
				fmt.Sprintf("Select Model (%d available)", len(models)),
				"Close",
				listContainer,
				w.Window,
			)
			modelDialog.Resize(fyne.NewSize(400, 350))
			modelDialog.Show()
		})
	}()
}

// Custom theme types to force light or dark theme

type forceLight struct{}

func (f *forceLight) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (f *forceLight) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (f *forceLight) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (f *forceLight) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type forceDark struct{}

func (f *forceDark) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (f *forceDark) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (f *forceDark) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (f *forceDark) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
