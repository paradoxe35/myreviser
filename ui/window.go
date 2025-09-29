package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/myreviser-go/internal/ai"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

type MainWindow struct {
	fyne.Window
	app    fyne.App
	config *config.Config

	// Data bindings
	providerBinding       binding.String
	apiKeyBinding         binding.String
	modelBinding          binding.String
	hotkeySelectBinding   binding.String
	hotkeySelectionBinding binding.String
	promptBinding         binding.String
	charLimitBinding      binding.String
	statusBinding         binding.String
}

func NewMainWindow(app fyne.App, cfg *config.Config) *MainWindow {
	window := app.NewWindow("MyReviser Settings")
	window.Resize(fyne.NewSize(700, 500))
	window.CenterOnScreen()

	mw := &MainWindow{
		Window: window,
		app:    app,
		config: cfg,
	}

	// Initialize data bindings
	mw.initBindings()

	// Create and set content
	content := mw.createContent()
	window.SetContent(content)

	// Hide window if start minimized
	if cfg.Appearance.StartMinimized {
		window.Hide()
	}

	return mw
}

func (w *MainWindow) initBindings() {
	w.providerBinding = binding.NewString()
	w.apiKeyBinding = binding.NewString()
	w.modelBinding = binding.NewString()
	w.hotkeySelectBinding = binding.NewString()
	w.hotkeySelectionBinding = binding.NewString()
	w.promptBinding = binding.NewString()
	w.charLimitBinding = binding.NewString()
	w.statusBinding = binding.NewString()

	// Set initial values from config
	w.providerBinding.Set(w.config.AIProvider.Provider)

	// Decrypt API key for display
	apiKey, _ := w.config.GetAPIKey()
	w.apiKeyBinding.Set(apiKey)

	w.modelBinding.Set(w.config.AIProvider.Model)
	w.hotkeySelectBinding.Set(w.config.Hotkeys.SelectAll)
	w.hotkeySelectionBinding.Set(w.config.Hotkeys.Selection)
	w.promptBinding.Set(w.config.Revision.SystemPrompt)
	w.charLimitBinding.Set(string(rune(w.config.Revision.CharacterLimit)))
	w.statusBinding.Set("Ready")
}

func (w *MainWindow) createContent() fyne.CanvasObject {
	// AI Provider Section
	providerSection := w.createProviderSection()

	// Hotkeys Section
	hotkeySection := w.createHotkeySection()

	// Revision Settings Section
	revisionSection := w.createRevisionSection()

	// Status Bar
	statusBar := w.createStatusBar()

	// Main content with tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("AI Provider", theme.ComputerIcon(), providerSection),
		container.NewTabItemWithIcon("Hotkeys", theme.SettingsIcon(), hotkeySection),
		container.NewTabItemWithIcon("Revision", theme.DocumentIcon(), revisionSection),
	)

	// Save button
	saveBtn := widget.NewButtonWithIcon("Save Settings", theme.DocumentSaveIcon(), w.saveSettings)
	saveBtn.Importance = widget.HighImportance

	// Main layout
	content := container.NewBorder(
		nil, // top
		container.NewBorder(nil, nil, nil, saveBtn, statusBar), // bottom
		nil, // left
		nil, // right
		tabs, // center
	)

	return container.NewPadded(content)
}

func (w *MainWindow) createProviderSection() fyne.CanvasObject {
	// Provider selection
	providerLabel := widget.NewLabel("AI Provider:")
	providerSelect := widget.NewSelect(
		[]string{"openai", "claude", "gemini"},
		func(value string) {
			w.providerBinding.Set(value)
			w.updateProviderDefaults(value)
		},
	)
	selected, _ := w.providerBinding.Get()
	providerSelect.SetSelected(selected)

	// API Key
	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.Bind(w.apiKeyBinding)
	apiKeyEntry.PlaceHolder = "Enter your API key"

	// Model
	modelLabel := widget.NewLabel("Model (optional):")
	modelEntry := widget.NewEntry()
	modelEntry.Bind(w.modelBinding)
	modelEntry.PlaceHolder = "e.g., gpt-4o-mini"

	// Base URL (for custom endpoints)
	baseURLLabel := widget.NewLabel("Base URL (optional):")
	baseURLEntry := widget.NewEntry()
	baseURLEntry.SetText(w.config.AIProvider.BaseURL)
	baseURLEntry.PlaceHolder = "Default API endpoint"

	// Test connection button
	testBtn := widget.NewButtonWithIcon("Test Connection", theme.ConfirmIcon(), func() {
		w.testAPIConnection()
	})

	form := container.NewVBox(
		container.NewBorder(nil, nil, providerLabel, nil, providerSelect),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, apiKeyLabel, nil, apiKeyEntry),
		container.NewBorder(nil, nil, modelLabel, nil, modelEntry),
		container.NewBorder(nil, nil, baseURLLabel, nil, baseURLEntry),
		widget.NewSeparator(),
		testBtn,
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createHotkeySection() fyne.CanvasObject {
	// Select All Hotkey
	selectAllLabel := widget.NewLabel("Select All & Revise:")
	selectAllEntry := widget.NewEntry()
	selectAllEntry.Bind(w.hotkeySelectBinding)
	selectAllEntry.PlaceHolder = "e.g., ctrl+alt+space"

	// Selection Hotkey
	selectionLabel := widget.NewLabel("Revise Selection:")
	selectionEntry := widget.NewEntry()
	selectionEntry.Bind(w.hotkeySelectionBinding)
	selectionEntry.PlaceHolder = "e.g., ctrl+win"

	// Hotkey instructions
	instructions := widget.NewCard("", "Hotkey Format",
		widget.NewLabel("Use modifiers: ctrl, alt, shift, cmd/win/super\n"+
			"Examples: 'ctrl+alt+space', 'cmd+shift+r'\n\n"+
			"Note: Some combinations may be reserved by the OS."))

	// Reset to defaults button
	resetBtn := widget.NewButton("Reset to Defaults", func() {
		defaults := config.GetPlatformHotkeys()
		w.hotkeySelectBinding.Set(defaults.SelectAll)
		w.hotkeySelectionBinding.Set(defaults.Selection)
	})

	form := container.NewVBox(
		container.NewBorder(nil, nil, selectAllLabel, nil, selectAllEntry),
		container.NewBorder(nil, nil, selectionLabel, nil, selectionEntry),
		widget.NewSeparator(),
		instructions,
		resetBtn,
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createRevisionSection() fyne.CanvasObject {
	// System Prompt
	promptLabel := widget.NewLabel("System Prompt:")
	promptEntry := widget.NewMultiLineEntry()
	promptEntry.Bind(w.promptBinding)
	promptEntry.SetMinRowsVisible(5)

	// Character Limit
	charLimitLabel := widget.NewLabel("Character Limit:")
	charLimitEntry := widget.NewEntry()
	charLimitEntry.Bind(w.charLimitBinding)
	charLimitEntry.PlaceHolder = "1000"

	// Timeout
	timeoutLabel := widget.NewLabel("Timeout (seconds):")
	timeoutSlider := widget.NewSlider(5, 60)
	timeoutSlider.SetValue(float64(w.config.Revision.TimeoutSeconds))
	timeoutValue := widget.NewLabel("30s")
	timeoutSlider.OnChanged = func(value float64) {
		timeoutValue.SetText(fmt.Sprintf("%ds", int(value)))
	}

	// Reset prompt button
	resetPromptBtn := widget.NewButton("Reset to Default Prompt", func() {
		w.promptBinding.Set(config.GetDefaultRevision().SystemPrompt)
	})

	form := container.NewVBox(
		promptLabel,
		promptEntry,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, charLimitLabel, nil, charLimitEntry),
		container.NewBorder(nil, nil, timeoutLabel, timeoutValue, timeoutSlider),
		widget.NewSeparator(),
		resetPromptBtn,
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createStatusBar() fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	statusLabel.Bind(w.statusBinding)

	return container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		statusLabel,
	)
}

func (w *MainWindow) updateProviderDefaults(provider string) {
	switch provider {
	case "openai":
		w.modelBinding.Set("gpt-4o-mini")
		w.config.AIProvider.BaseURL = "https://api.openai.com/v1"
	case "claude":
		w.modelBinding.Set("claude-3-5-haiku-20241022")
		w.config.AIProvider.BaseURL = "https://api.anthropic.com"
	case "gemini":
		w.modelBinding.Set("gemini-2.5-flash-lite")
		w.config.AIProvider.BaseURL = "https://generativelanguage.googleapis.com"
	}
}

func (w *MainWindow) testAPIConnection() {
	w.statusBinding.Set("Testing API connection...")

	go func() {
		// Get current provider settings
		provider, _ := w.providerBinding.Get()
		apiKey, _ := w.apiKeyBinding.Get()
		model, _ := w.modelBinding.Get()

		if apiKey == "" {
			w.statusBinding.Set("Error: API key is required")
			return
		}

		// Create a test provider
		var testProvider ai.Provider
		switch provider {
		case "openai":
			testProvider = ai.NewOpenAIProvider(apiKey, w.config.AIProvider.BaseURL, model)
		case "claude":
			testProvider = ai.NewAnthropicProvider(apiKey, w.config.AIProvider.BaseURL, model)
		case "gemini":
			testProvider = ai.NewGeminiProvider(apiKey, w.config.AIProvider.BaseURL, model)
		default:
			w.statusBinding.Set("Error: Unknown provider")
			return
		}

		// Test the connection with a simple prompt
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		testText := "Hello"
		testPrompt := "Reply with 'Connection successful' if you receive this message."

		_, err := testProvider.ReviseText(ctx, testText, testPrompt)
		if err != nil {
			w.statusBinding.Set(fmt.Sprintf("Connection failed: %v", err))
			logger.Error("API connection test failed", "error", err)
		} else {
			w.statusBinding.Set("Connection successful!")
			logger.Info("API connection test successful")
		}
	}()
}

func (w *MainWindow) saveSettings() {
	// Update config from bindings
	provider, _ := w.providerBinding.Get()
	apiKey, _ := w.apiKeyBinding.Get()
	model, _ := w.modelBinding.Get()
	hotkeySelect, _ := w.hotkeySelectBinding.Get()
	hotkeySelection, _ := w.hotkeySelectionBinding.Get()
	prompt, _ := w.promptBinding.Get()

	w.config.AIProvider.Provider = provider
	w.config.AIProvider.Model = model
	w.config.SaveAPIKey(provider, apiKey)
	w.config.Hotkeys.SelectAll = hotkeySelect
	w.config.Hotkeys.Selection = hotkeySelection
	w.config.Revision.SystemPrompt = prompt

	// Save to disk
	if err := w.config.Save(); err != nil {
		w.statusBinding.Set("Error saving settings: " + err.Error())
		logger.Error("Failed to save settings", "error", err)
	} else {
		w.statusBinding.Set("Settings saved successfully")
		logger.Info("Settings saved")
	}
}

func (w *MainWindow) ShowAndRun() {
	w.Window.ShowAndRun()
}