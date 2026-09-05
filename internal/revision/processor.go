package revision

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

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

	settings := cfg.GetProviderSettings(currentProvider)
	if settings.RequiresAPIKey() && strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("API key is empty for provider: %s", currentProvider)
	}

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

	applyReasoning(provider, settings)

	p.providerFactory.Register(currentProvider, provider)
	p.providerFactory.SetCurrent(currentProvider)

	logger.Info("AI provider initialized", "provider", currentProvider, "custom", settings.IsCustom)
	return nil
}

// begin takes the one-at-a-time guard. Two overlapping runs would fight over the clipboard, a
// single global resource each of them saves and restores.
func (p *Processor) begin() (func(), error) {
	p.mu.Lock()
	if p.processing {
		p.mu.Unlock()
		logger.Warn("Already processing a revision")
		return nil, fmt.Errorf("already processing a revision, please wait")
	}
	p.processing = true
	p.mu.Unlock()

	return func() {
		p.mu.Lock()
		p.processing = false
		p.mu.Unlock()
	}, nil
}

func outcomeError(outcome input.CaptureOutcome) error {
	switch outcome {
	case input.CaptureNothingSelected:
		return fmt.Errorf("no text selected")
	case input.CaptureCopyFailed:
		return fmt.Errorf("could not copy from that window - some applications block it")
	default:
		return nil
	}
}

// ProcessSelectAll selects the whole field, revises it, and writes the result back.
func (p *Processor) ProcessSelectAll() error {
	release, err := p.begin()
	if err != nil {
		return err
	}
	defer release()

	logger.Info("Starting select all revision")

	text, outcome, err := p.clipboardManager.CaptureAll()
	if err != nil {
		return err
	}
	if err := outcomeError(outcome); err != nil {
		return err
	}

	revisedText, err := p.reviseText(text)
	if err != nil {
		p.clipboardManager.Abandon()
		return fmt.Errorf("failed to revise text: %w", err)
	}

	// Selected again: a field can drop its selection while the model is working, and pasting
	// without one inserts the revision beside the original instead of replacing it.
	if err := input.FFISimulateSelectAll(); err != nil {
		p.clipboardManager.Abandon()
		return fmt.Errorf("failed to select all for paste: %w", err)
	}

	if err := p.clipboardManager.ReplaceSelectedText(revisedText); err != nil {
		return fmt.Errorf("failed to replace text: %w", err)
	}

	logger.Info("Select all revision completed")
	return nil
}

// ProcessSelection revises whatever the user has selected, leaving the rest of the field alone.
func (p *Processor) ProcessSelection() error {
	release, err := p.begin()
	if err != nil {
		return err
	}
	defer release()

	logger.Info("Starting selection revision")

	text, outcome, err := p.clipboardManager.CaptureSelection()
	if err != nil {
		return err
	}
	if err := outcomeError(outcome); err != nil {
		return err
	}

	revisedText, err := p.reviseText(text)
	if err != nil {
		p.clipboardManager.Abandon()
		return fmt.Errorf("failed to revise text: %w", err)
	}

	if err := p.clipboardManager.ReplaceSelectedText(revisedText); err != nil {
		return fmt.Errorf("failed to replace selected text: %w", err)
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

	// Counted in characters, not bytes: an accented letter is two bytes in UTF-8, so len() halved
	// the limit for exactly the text this app exists to correct.
	if characters := utf8.RuneCountInString(trimmedText); characters > cfg.Revision.CharacterLimit {
		return "", fmt.Errorf("text exceeds character limit (%d > %d)",
			characters, cfg.Revision.CharacterLimit)
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
		"characters", utf8.RuneCountInString(trimmedText),
	)

	revised, err := provider.ReviseText(ctx, trimmedText, p.config.Revision.SystemPrompt)
	if err != nil {
		return "", fmt.Errorf("AI revision failed: %w", err)
	}

	cleaned := ai.CleanResponse(revised)
	if cleaned == "" {
		return "", fmt.Errorf("AI provider returned empty response")
	}

	logger.Info("Text revised successfully",
		"original_characters", utf8.RuneCountInString(trimmedText),
		"revised_characters", utf8.RuneCountInString(cleaned),
	)

	// The reply replaces the selection as it was, so the selection's own edges go back on. Without
	// them "word " returns as "word" and runs into the next one.
	return leadingWhitespace(textToRevise) + cleaned + trailingWhitespace(textToRevise), nil
}

func leadingWhitespace(text string) string {
	return text[:len(text)-len(strings.TrimLeftFunc(text, unicode.IsSpace))]
}

func trailingWhitespace(text string) string {
	return text[len(strings.TrimRightFunc(text, unicode.IsSpace)):]
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

	mentionedName := strings.TrimPrefix(parts[0], "@")

	// Case-insensitive lookup returns the actual stored provider name
	actualProviderName, found := p.findProviderByName(mentionedName)
	if !found {
		return "", text, false
	}

	cleanedText := ""
	if len(parts) > 1 {
		cleanedText = strings.TrimSpace(parts[1])
	}

	return actualProviderName, cleanedText, true
}

// findProviderByName does a case-insensitive lookup and returns the actual stored provider name
func (p *Processor) findProviderByName(name string) (string, bool) {
	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	if cfg.AIProvider.Providers == nil {
		return "", false
	}

	nameLower := strings.ToLower(name)
	for storedName := range cfg.AIProvider.Providers {
		if strings.ToLower(storedName) == nameLower {
			return storedName, true
		}
	}
	return "", false
}

func (p *Processor) getOrCreateProvider(name string) (ai.Provider, error) {
	if provider, err := p.providerFactory.Get(name); err == nil {
		return provider, nil
	}

	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	apiKey, err := cfg.GetAPIKey(name)
	if err != nil {
		return nil, fmt.Errorf("no API key configured for %s", name)
	}

	settings := cfg.GetProviderSettings(name)
	if settings.RequiresAPIKey() && strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("no API key configured for %s", name)
	}

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

	applyReasoning(provider, settings)

	p.providerFactory.Register(name, provider)
	return provider, nil
}

// applyReasoning reaches only the providers that have a reasoning parameter to send.
func applyReasoning(provider ai.Provider, settings config.ProviderSettings) {
	if aware, ok := provider.(ai.ReasoningAware); ok {
		aware.SetLowReasoning(settings.LowReasoning)
	}
}
