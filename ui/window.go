package ui

import (
	"context"
	"fmt"
	"strconv"
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

const (
	// UI spacing constants
	PaddingSmall  = 5
	PaddingMedium = 10
	PaddingLarge  = 20
	MaxWindowWidth = 800
	MaxWindowHeight = 600
)

type MainWindow struct {
	fyne.Window
	app    fyne.App
	config *config.Config

	// Data bindings
	providerBinding        binding.String
	apiKeyBinding          binding.String
	modelBinding           binding.String
	hotkeySelectBinding    binding.String
	hotkeySelectionBinding binding.String
	promptBinding          binding.String
	charLimitBinding       binding.String
	statusBinding          binding.String
	startMinimizedBinding  binding.Bool

	// UI containers for dynamic visibility
	baseURLContainer   *fyne.Container
	baseURLEntry       *widget.Entry
}

func NewMainWindow(app fyne.App, cfg *config.Config) *MainWindow {
	window := app.NewWindow("MyReviser Settings")

	// Set window size with constraints
	window.Resize(fyne.NewSize(700, 500))
	window.SetFixedSize(false)

	// Set maximum size to prevent overflow
	if desk, ok := window.(interface {
		SetMaxSize(fyne.Size)
	}); ok {
		desk.SetMaxSize(fyne.NewSize(MaxWindowWidth, MaxWindowHeight))
	}

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
	w.startMinimizedBinding = binding.NewBool()

	// Set initial values from config
	w.providerBinding.Set(w.config.AIProvider.Provider)

	// Decrypt API key for display
	apiKey, _ := w.config.GetAPIKey()
	w.apiKeyBinding.Set(apiKey)

	w.modelBinding.Set(w.config.AIProvider.Model)
	w.hotkeySelectBinding.Set(w.config.Hotkeys.SelectAll)
	w.hotkeySelectionBinding.Set(w.config.Hotkeys.Selection)
	w.promptBinding.Set(w.config.Revision.SystemPrompt)
	w.charLimitBinding.Set(strconv.Itoa(w.config.Revision.CharacterLimit))
	w.statusBinding.Set("Ready")
	w.startMinimizedBinding.Set(w.config.Appearance.StartMinimized)
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
	selected, _ := w.providerBinding.Get()

	// Provider selection
	providerLabel := widget.NewLabel("AI Provider:")
	providerSelect := widget.NewSelect(
		[]string{"openai", "claude", "gemini"},
		func(value string) {
			w.providerBinding.Set(value)
			w.updateProviderDefaults(value)

			// Show/hide Base URL based on provider
			if w.baseURLContainer != nil {
				if value == "openai" {
					w.baseURLContainer.Show()
				} else {
					w.baseURLContainer.Hide()
				}
			}
		},
	)

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

	// Base URL (for custom endpoints - only for OpenAI)
	baseURLLabel := widget.NewLabel("Base URL (optional):")
	w.baseURLEntry = widget.NewEntry()
	w.baseURLEntry.SetText(w.config.AIProvider.BaseURL)
	w.baseURLEntry.PlaceHolder = "Default: https://api.openai.com/v1"

	// Create Base URL container for visibility control
	w.baseURLContainer = container.NewVBox(
		container.NewBorder(nil, nil, baseURLLabel, nil, w.baseURLEntry),
	)

	// Test connection button
	testBtn := widget.NewButtonWithIcon("Test Connection", theme.ConfirmIcon(), func() {
		w.testAPIConnection()
	})

	form := container.NewVBox(
		container.NewPadded(container.NewBorder(nil, nil, providerLabel, nil, providerSelect)),
		widget.NewSeparator(),
		container.NewPadded(container.NewBorder(nil, nil, apiKeyLabel, nil, apiKeyEntry)),
		container.NewPadded(container.NewBorder(nil, nil, modelLabel, nil, modelEntry)),
		container.NewPadded(w.baseURLContainer),
		widget.NewSeparator(),
		container.NewPadded(testBtn),
	)

	// Now set the initial provider selection (after container is created)
	providerSelect.SetSelected(selected)

	// Set initial visibility based on provider
	if selected != "openai" {
		w.baseURLContainer.Hide()
	}

	return container.NewScroll(form)
}

func (w *MainWindow) createHotkeySection() fyne.CanvasObject {
	// Select All Hotkey with capture widget
	selectAllLabel := widget.NewLabel("Select All & Revise:")
	selectAllCapture := NewHotkeyCapture(w.hotkeySelectBinding, "Click 'Capture' to set hotkey")

	// Selection Hotkey with capture widget
	selectionLabel := widget.NewLabel("Revise Selection:")
	selectionCapture := NewHotkeyCapture(w.hotkeySelectionBinding, "Click 'Capture' to set hotkey")

	// Hotkey instructions
	instructions := widget.NewCard("", "How to Capture Hotkeys",
		widget.NewLabel("1. Click the 'Capture' button\n"+
			"2. Press your desired key combination\n"+
			"3. Release the keys to save\n"+
			"4. Press ESC to cancel\n\n"+
			"Note: You must use at least one modifier key\n"+
			"(Ctrl, Alt, Shift, or Super/Win/Cmd).\n\n"+
			"Some combinations may be reserved by your OS."))

	// Reset to defaults button
	resetBtn := widget.NewButton("Reset to Defaults", func() {
		defaults := config.GetPlatformHotkeys()
		w.hotkeySelectBinding.Set(defaults.SelectAll)
		w.hotkeySelectionBinding.Set(defaults.Selection)
	})

	form := container.NewVBox(
		container.NewPadded(selectAllLabel),
		container.NewPadded(selectAllCapture),
		widget.NewSeparator(),
		container.NewPadded(selectionLabel),
		container.NewPadded(selectionCapture),
		widget.NewSeparator(),
		container.NewPadded(instructions),
		container.NewPadded(resetBtn),
	)

	return container.NewScroll(form)
}

func (w *MainWindow) createRevisionSection() fyne.CanvasObject {
	// System Prompt
	promptLabel := widget.NewLabel("System Prompt:")
	promptEntry := widget.NewMultiLineEntry()
	promptEntry.Bind(w.promptBinding)
	promptEntry.SetMinRowsVisible(5)
	promptEntry.Wrapping = fyne.TextWrapWord

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
		w.config.Revision.TimeoutSeconds = int(value)
	}

	// Reset prompt button
	resetPromptBtn := widget.NewButton("Reset to Default Prompt", func() {
		w.promptBinding.Set(config.GetDefaultRevision().SystemPrompt)
	})

	// Start Minimized checkbox
	startMinimizedCheck := widget.NewCheck("Start minimized to system tray", func(checked bool) {
		w.startMinimizedBinding.Set(checked)
	})
	startMinimizedCheck.Bind(w.startMinimizedBinding)

	form := container.NewVBox(
		container.NewPadded(promptLabel),
		container.NewPadded(promptEntry),
		widget.NewSeparator(),
		container.NewPadded(container.NewBorder(nil, nil, charLimitLabel, nil, charLimitEntry)),
		container.NewPadded(container.NewBorder(nil, nil, timeoutLabel, timeoutValue, timeoutSlider)),
		widget.NewSeparator(),
		container.NewPadded(startMinimizedCheck),
		container.NewPadded(resetPromptBtn),
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

		// Get Base URL from entry if OpenAI
		baseURL := w.config.AIProvider.BaseURL
		if provider == "openai" && w.baseURLEntry != nil {
			baseURL = w.baseURLEntry.Text
		}

		// Create a test provider
		var testProvider ai.Provider
		switch provider {
		case "openai":
			testProvider = ai.NewOpenAIProvider(apiKey, baseURL, model)
		case "claude":
			testProvider = ai.NewAnthropicProvider(apiKey, baseURL, model)
		case "gemini":
			testProvider = ai.NewGeminiProvider(apiKey, baseURL, model)
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
			// Truncate error message if too long
			errMsg := err.Error()
			if len(errMsg) > 100 {
				errMsg = errMsg[:97] + "..."
			}
			w.statusBinding.Set(fmt.Sprintf("Connection failed: %s", errMsg))
			logger.Error("API connection test failed", "error", err)
		} else {
			w.statusBinding.Set("✓ Connection successful!")
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
	charLimitStr, _ := w.charLimitBinding.Get()
	startMinimized, _ := w.startMinimizedBinding.Get()

	// Parse character limit
	charLimit, err := strconv.Atoi(charLimitStr)
	if err != nil || charLimit <= 0 {
		w.statusBinding.Set("Error: Invalid character limit")
		return
	}

	// Update config
	w.config.AIProvider.Provider = provider
	w.config.AIProvider.Model = model
	w.config.SaveAPIKey(provider, apiKey)

	// Update Base URL if OpenAI
	if provider == "openai" && w.baseURLEntry != nil {
		w.config.AIProvider.BaseURL = w.baseURLEntry.Text
	}

	w.config.Hotkeys.SelectAll = hotkeySelect
	w.config.Hotkeys.Selection = hotkeySelection
	w.config.Revision.SystemPrompt = prompt
	w.config.Revision.CharacterLimit = charLimit
	w.config.Appearance.StartMinimized = startMinimized

	// Save to disk
	if err := w.config.Save(); err != nil {
		w.statusBinding.Set("✗ Error saving settings: " + err.Error())
		logger.Error("Failed to save settings", "error", err)
	} else {
		w.statusBinding.Set("✓ Settings saved successfully")
		logger.Info("Settings saved")
	}
}

func (w *MainWindow) ShowAndRun() {
	w.Window.ShowAndRun()
}