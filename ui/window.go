package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/myreviser-go/internal/ai"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/input"
	"github.com/paradoxe35/myreviser-go/internal/logger"
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
	app           fyne.App
	config        *config.Config
	hotkeyManager *input.FFIHotkeyManager

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
	themeBinding           binding.String

	// UI containers for dynamic visibility
	baseURLContainer *fyne.Container
	baseURLEntry     *widget.Entry
}

func NewMainWindow(app fyne.App, cfg *config.Config, hotkeyManager *input.FFIHotkeyManager) *MainWindow {
	window := app.NewWindow("MyReviser Settings")

	// Set fixed window size
	window.Resize(fyne.NewSize(650, 550))
	window.SetFixedSize(true)

	window.CenterOnScreen()

	mw := &MainWindow{
		Window:        window,
		app:           app,
		config:        cfg,
		hotkeyManager: hotkeyManager,
	}

	// Initialize data bindings
	mw.initBindings()

	// Apply theme from config
	themeName := cfg.Appearance.Theme
	if themeName == "" {
		themeName = "auto"
	}
	mw.applyTheme(themeName)

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

	// Set theme, default to "auto" if empty
	theme := w.config.Appearance.Theme
	if theme == "" {
		theme = "auto"
	}
	w.themeBinding.Set(theme)
}

// loadProviderSettings loads settings for a specific provider into the UI
func (w *MainWindow) loadProviderSettings(provider string) {
	// Get provider settings
	settings := w.config.GetProviderSettings(provider)

	// Decrypt and load API key
	apiKey, _ := w.config.GetAPIKey(provider)
	w.apiKeyBinding.Set(apiKey)

	// Load model and base URL
	w.modelBinding.Set(settings.Model)

	// Update base URL entry if it exists
	if w.baseURLEntry != nil {
		w.baseURLEntry.SetText(settings.BaseURL)
	}
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

	// Provider selection
	providerLabel := widget.NewLabel("AI Provider:")
	providerLabel.TextStyle.Bold = true
	providerSelect := widget.NewSelect(
		[]string{"openai", "claude", "gemini"},
		func(value string) {
			w.providerBinding.Set(value)

			// Load settings for the selected provider
			w.loadProviderSettings(value)

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
	apiKeyLabel.TextStyle.Bold = true
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.Bind(w.apiKeyBinding)
	apiKeyEntry.PlaceHolder = "Enter your API key"

	// Model
	modelLabel := widget.NewLabel("Model (optional):")
	modelLabel.TextStyle.Bold = true
	modelEntry := widget.NewEntry()
	modelEntry.Bind(w.modelBinding)
	modelEntry.PlaceHolder = "e.g., gpt-4o"

	// Base URL (for custom endpoints - only for OpenAI)
	baseURLLabel := widget.NewLabel("Base URL (optional):")
	baseURLLabel.TextStyle.Bold = true
	w.baseURLEntry = widget.NewEntry()
	// Base URL will be loaded by loadProviderSettings
	w.baseURLEntry.PlaceHolder = "Default: https://api.openai.com/v1"

	// Create Base URL container for visibility control
	w.baseURLContainer = container.NewVBox(
		container.NewPadded(container.NewBorder(nil, nil, baseURLLabel, nil, w.baseURLEntry)),
	)

	// Test connection button
	testBtn := widget.NewButtonWithIcon("Test Connection", theme.ConfirmIcon(), func() {
		w.testAPIConnection()
	})

	form := container.NewVBox(
		container.NewPadded(providerLabel),
		container.NewPadded(providerSelect),
		widget.NewSeparator(),
		container.NewPadded(apiKeyLabel),
		container.NewPadded(apiKeyEntry),
		container.NewPadded(modelLabel),
		container.NewPadded(modelEntry),
		w.baseURLContainer,
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
	selectAllLabel.TextStyle.Bold = true
	selectAllCapture := NewHotkeyCapture(w.hotkeySelectBinding, "Click 'Capture' to set hotkey")
	selectAllCapture.window = w.Window // Set window reference for focus
	// Connect callbacks to disable/enable global hotkeys during capture
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

	// Hotkey instructions
	instructionsLabel := widget.NewLabel("1. Click the 'Capture' button\n" +
		"2. Press keys ONE AT A TIME to build combination\n" +
		"   (e.g., press Ctrl, then Alt, then Space)\n" +
		"3. Press Enter or click 'Save' to save\n" +
		"4. Press ESC to cancel without saving\n\n" +
		"Requirements:\n" +
		"• At least one modifier (Ctrl, Alt, Shift, Super/Win/Cmd)\n" +
		"• At least one regular key\n" +
		"• Each hotkey must be unique\n" +
		"• Maximum 3 regular keys per combination\n\n" +
		"Note: Only one hotkey can be captured at a time.\n" +
		"Global hotkeys are disabled during capture.")
	instructionsLabel.Wrapping = fyne.TextWrapWord
	instructions := widget.NewCard("", "How to Capture Hotkeys", instructionsLabel)

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

	form := container.NewVBox(
		container.NewPadded(promptLabel),
		container.NewPadded(promptEntry),
		widget.NewSeparator(),
		container.NewPadded(container.NewBorder(nil, nil, charLimitLabel, nil, charLimitEntry)),
		container.NewPadded(container.NewBorder(nil, nil, timeoutLabel, timeoutValue, timeoutSlider)),
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

	form := container.NewVBox(
		container.NewPadded(themeLabel),
		container.NewPadded(themeSelect),
		container.NewPadded(themeDesc),
		widget.NewSeparator(),
		container.NewPadded(startMinimizedCheck),
	)

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
		// Get current provider settings
		provider, _ := w.providerBinding.Get()
		apiKey, _ := w.apiKeyBinding.Get()
		model, _ := w.modelBinding.Get()

		if apiKey == "" {
			w.statusBinding.Set("Error: API key is required")
			return
		}

		// Get Base URL from entry
		baseURL := ""
		if w.baseURLEntry != nil {
			baseURL = w.baseURLEntry.Text
		}

		// Fallback to provider defaults if not set
		if baseURL == "" {
			settings := w.config.GetProviderSettings(provider)
			baseURL = settings.BaseURL
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
			// Truncate error message if too long (limit to 80 chars)
			errMsg := err.Error()
			if len(errMsg) > 80 {
				errMsg = errMsg[:77] + "..."
			}
			w.statusBinding.Set(fmt.Sprintf("✗ Connection failed: %s", errMsg))
			logger.Error("API connection test failed", "error", err)
		} else {
			w.statusBinding.Set("✓ Connection successful!")
			logger.Info("API connection test successful")
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
	// Update config from bindings
	provider, _ := w.providerBinding.Get()
	apiKey, _ := w.apiKeyBinding.Get()
	model, _ := w.modelBinding.Get()
	hotkeySelect, _ := w.hotkeySelectBinding.Get()
	hotkeySelection, _ := w.hotkeySelectionBinding.Get()
	prompt, _ := w.promptBinding.Get()
	charLimitStr, _ := w.charLimitBinding.Get()
	startMinimized, _ := w.startMinimizedBinding.Get()
	themeSetting, _ := w.themeBinding.Get()

	// Parse character limit
	charLimit, err := strconv.Atoi(charLimitStr)
	if err != nil || charLimit <= 0 {
		w.statusBinding.Set("Error: Invalid character limit")
		return
	}

	// Get base URL from entry
	baseURL := ""
	if w.baseURLEntry != nil {
		baseURL = w.baseURLEntry.Text
	}

	// Update provider-specific settings
	settings := config.ProviderSettings{
		Model:   model,
		BaseURL: baseURL,
	}
	w.config.SetProviderSettings(provider, settings)

	// Save API key for this provider
	if err := w.config.SaveAPIKey(provider, apiKey); err != nil {
		errMsg := err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		w.statusBinding.Set("✗ Error saving API key: " + errMsg)
		logger.Error("Failed to save API key", "error", err)
		return
	}

	// Set current provider
	w.config.SetCurrentProvider(provider)

	// Update other settings
	w.config.Hotkeys.SelectAll = hotkeySelect
	w.config.Hotkeys.Selection = hotkeySelection
	w.config.Revision.SystemPrompt = prompt
	w.config.Revision.CharacterLimit = charLimit
	w.config.Appearance.StartMinimized = startMinimized
	w.config.Appearance.Theme = themeSetting

	// Save to disk
	if err := w.config.Save(); err != nil {
		errMsg := err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:57] + "..."
		}
		w.statusBinding.Set("✗ Error saving settings: " + errMsg)
		logger.Error("Failed to save settings", "error", err)
	} else {
		w.statusBinding.Set("✓ Settings saved successfully")
		logger.Info("Settings saved")
	}
}

func (w *MainWindow) ShowAndRun() {
	w.Window.ShowAndRun()
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
