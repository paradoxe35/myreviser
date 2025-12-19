package revision

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/paradoxe35/myreviser/internal/ai"
	"github.com/paradoxe35/myreviser/internal/config"
	"github.com/paradoxe35/myreviser/internal/input"
	"github.com/paradoxe35/myreviser/internal/logger"
)

// Processor handles text revision operations
type Processor struct {
	mu               sync.Mutex
	config           *config.Config
	providerFactory  *ai.ProviderFactory
	clipboardManager *input.FFIClipboardManager
	processing       bool
}

// NewProcessor creates a new revision processor
func NewProcessor(cfg *config.Config) (*Processor, error) {
	clipManager, err := input.NewFFIClipboardManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard manager: %w", err)
	}

	p := &Processor{
		config:           cfg,
		providerFactory:  ai.NewProviderFactory(),
		clipboardManager: clipManager,
	}

	// Initialize AI providers (don't fail startup if not configured)
	if err := p.initializeProviders(); err != nil {
		logger.Warn("AI provider not configured at startup", "error", err)
	}

	// Register config change listener
	config.RegisterListener(func(newCfg *config.Config) {
		p.mu.Lock()
		p.config = newCfg
		p.mu.Unlock()

		// Reinitialize providers with new config
		if err := p.initializeProviders(); err != nil {
			logger.Warn("AI provider not configured", "error", err)
		}
	})

	return p, nil
}

func (p *Processor) initializeProviders() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cfg := p.config

	currentProvider := cfg.GetCurrentProvider()
	if currentProvider == "" {
		return fmt.Errorf("no provider configured")
	}

	apiKey, err := cfg.GetCurrentAPIKey()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("API key is empty for provider: %s", currentProvider)
	}

	settings := cfg.GetProviderSettings(currentProvider)

	var provider ai.Provider

	if settings.IsCustom {
		customProvider, err := ai.NewCustomProvider(
			currentProvider,
			settings.ProviderType,
			apiKey,
			settings.BaseURL,
			settings.Model,
			settings.Temperature,
		)
		if err != nil {
			return fmt.Errorf("failed to create custom provider: %w", err)
		}
		provider = customProvider
	} else {
		switch currentProvider {
		case config.BuiltInOpenAI:
			provider = ai.NewOpenAIProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		case config.BuiltInClaude:
			provider = ai.NewAnthropicProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		case config.BuiltInGemini:
			provider = ai.NewGeminiProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		default:
			return fmt.Errorf("unknown provider: %s", currentProvider)
		}
	}

	p.providerFactory.Register(currentProvider, provider)
	p.providerFactory.SetCurrent(currentProvider)

	logger.Info("AI provider initialized", "provider", currentProvider, "custom", settings.IsCustom)
	return nil
}

// ProcessSelectAll processes text with select all
func (p *Processor) ProcessSelectAll() error {
	p.mu.Lock()
	if p.processing {
		p.mu.Unlock()
		logger.Warn("Already processing a revision")
		return fmt.Errorf("already processing a revision, please wait")
	}
	p.processing = true
	p.mu.Unlock()

	// Ensure we clear processing flag on exit
	defer func() {
		p.mu.Lock()
		p.processing = false
		p.mu.Unlock()
	}()

	logger.Info("Starting select all revision")

	// Wait for hotkey processing and clipboard operations
	time.Sleep(250 * time.Millisecond)

	if err := p.clipboardManager.SaveCurrent(); err != nil {
		return fmt.Errorf("failed to save clipboard: %w", err)
	}

	time.Sleep(75 * time.Millisecond)

	if err := input.FFISimulateSelectAll(); err != nil {
		return fmt.Errorf("failed to select all: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := input.FFISimulateCopy(); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Wait for clipboard synchronization
	time.Sleep(150 * time.Millisecond)

	// Get text from clipboard
	text, err := p.clipboardManager.GetText()
	if err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to get clipboard text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		p.clipboardManager.Restore()
		return fmt.Errorf("no text to revise (clipboard is empty or contains only whitespace)")
	}

	revisedText, err := p.reviseText(text)
	if err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to revise text: %w", err)
	}

	if err := input.FFISimulateSelectAll(); err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to select all for paste: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := p.clipboardManager.ReplaceSelectedText(revisedText); err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to replace text: %w", err)
	}

	if err := p.clipboardManager.Restore(); err != nil {
		logger.Error("Failed to restore clipboard", "error", err)
	}

	logger.Info("Select all revision completed")
	return nil
}

// ProcessSelection processes selected text
func (p *Processor) ProcessSelection() error {
	p.mu.Lock()
	if p.processing {
		p.mu.Unlock()
		logger.Warn("Already processing a revision")
		return fmt.Errorf("already processing a revision, please wait")
	}
	p.processing = true
	p.mu.Unlock()

	// Ensure we clear processing flag on exit
	defer func() {
		p.mu.Lock()
		p.processing = false
		p.mu.Unlock()
	}()

	logger.Info("Starting selection revision")

	time.Sleep(250 * time.Millisecond)

	text, err := p.clipboardManager.CaptureSelectedText()
	if err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to capture selected text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		p.clipboardManager.Restore()
		return fmt.Errorf("no text selected or selection is empty")
	}

	revisedText, err := p.reviseText(text)
	if err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to revise text: %w", err)
	}

	if err := p.clipboardManager.ReplaceSelectedText(revisedText); err != nil {
		return fmt.Errorf("failed to replace selected text: %w", err)
	}

	if err := p.clipboardManager.Restore(); err != nil {
		logger.Warn("Failed to restore clipboard after paste", "error", err)
	}

	logger.Info("Selection revision completed")
	return nil
}

func (p *Processor) reviseText(text string) (string, error) {
	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	mentionedProvider, cleanedText, hasMention := p.parseProviderMention(text)

	textToRevise := text
	if hasMention && mentionedProvider != "" {
		textToRevise = cleanedText
	}

	trimmedText := strings.TrimSpace(textToRevise)
	if trimmedText == "" {
		return "", fmt.Errorf("text is empty or contains only whitespace")
	}

	if len(textToRevise) > cfg.Revision.CharacterLimit {
		return "", fmt.Errorf("text exceeds character limit (%d > %d)",
			len(textToRevise), cfg.Revision.CharacterLimit)
	}

	if len(trimmedText) < 1 {
		return "", fmt.Errorf("text too short to revise")
	}

	var provider ai.Provider
	if hasMention && mentionedProvider != "" {
		var err error
		provider, err = p.getOrCreateProvider(mentionedProvider)
		if err != nil {
			logger.Warn("Failed to use mentioned provider, using default",
				"mentioned", mentionedProvider, "error", err)
			provider = p.providerFactory.GetCurrent()
			textToRevise = text
		} else {
			logger.Info("Using mentioned provider", "provider", mentionedProvider)
		}
	} else {
		provider = p.providerFactory.GetCurrent()
	}

	if provider == nil {
		return "", fmt.Errorf("no AI provider configured - please configure your API key in Settings")
	}

	timeout := time.Duration(p.config.Revision.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("Sending text to AI provider",
		"provider", provider.GetName(),
		"model", provider.GetModel(),
		"temperature", provider.GetTemperature(),
		"text", textToRevise,
		"text_length", len(textToRevise),
	)

	revised, err := provider.ReviseText(ctx, textToRevise, p.config.Revision.SystemPrompt)
	if err != nil {
		return "", fmt.Errorf("AI revision failed: %w", err)
	}

	revisedTrimmed := strings.TrimSpace(revised)
	if revisedTrimmed == "" {
		return "", fmt.Errorf("AI provider returned empty response")
	}

	logger.Info("Text revised successfully",
		"revised", revised,
		"original_length", len(textToRevise),
		"revised_length", len(revised),
	)

	return revised, nil
}

// UpdateProvider updates the AI provider
func (p *Processor) UpdateProvider(provider string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.AIProvider.Provider = provider
	return p.initializeProviders()
}

// IsProcessing returns whether the processor is currently processing
func (p *Processor) IsProcessing() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.processing
}

// Close releases the processor resources
func (p *Processor) Close() {
	if p.clipboardManager != nil {
		p.clipboardManager.Close()
	}
}

func (p *Processor) parseProviderMention(text string) (string, string, bool) {
	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	if !cfg.Revision.EnableProviderMentions {
		return "", text, false
	}

	trimmedText := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmedText, "@") {
		return "", text, false
	}

	parts := strings.SplitN(trimmedText, " ", 2)
	if len(parts) == 0 {
		return "", text, false
	}

	providerName := strings.ToLower(strings.TrimPrefix(parts[0], "@"))

	if !p.isValidProvider(providerName) {
		return "", text, false
	}

	cleanedText := ""
	if len(parts) > 1 {
		cleanedText = strings.TrimSpace(parts[1])
	}

	return providerName, cleanedText, true
}

func (p *Processor) isValidProvider(name string) bool {
	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	if cfg.AIProvider.Providers == nil {
		return false
	}

	_, exists := cfg.AIProvider.Providers[name]
	return exists
}

func (p *Processor) getOrCreateProvider(name string) (ai.Provider, error) {
	if provider, err := p.providerFactory.Get(name); err == nil {
		return provider, nil
	}

	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	apiKey, err := cfg.GetAPIKey(name)
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("no API key configured for %s", name)
	}

	settings := cfg.GetProviderSettings(name)

	var provider ai.Provider

	if settings.IsCustom {
		customProvider, err := ai.NewCustomProvider(name, settings.ProviderType, apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		if err != nil {
			return nil, err
		}
		provider = customProvider
	} else {
		switch name {
		case config.BuiltInOpenAI:
			provider = ai.NewOpenAIProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		case config.BuiltInClaude:
			provider = ai.NewAnthropicProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		case config.BuiltInGemini:
			provider = ai.NewGeminiProvider(apiKey, settings.BaseURL, settings.Model, settings.Temperature)
		default:
			return nil, fmt.Errorf("unknown provider: %s", name)
		}
	}

	p.providerFactory.Register(name, provider)
	return provider, nil
}
