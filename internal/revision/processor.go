package revision

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/paradoxe35/myreviser-go/internal/ai"
	"github.com/paradoxe35/myreviser-go/internal/config"
	"github.com/paradoxe35/myreviser-go/internal/input"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// Processor handles text revision operations
type Processor struct {
	mu               sync.Mutex
	config           *config.Config
	providerFactory  *ai.ProviderFactory
	clipboardManager *input.ClipboardManager
	processing       bool
}

// NewProcessor creates a new revision processor
func NewProcessor(cfg *config.Config) (*Processor, error) {
	clipManager, err := input.NewClipboardManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard manager: %w", err)
	}

	p := &Processor{
		config:           cfg,
		providerFactory:  ai.NewProviderFactory(),
		clipboardManager: clipManager,
	}

	// Initialize AI providers
	if err := p.initializeProviders(); err != nil {
		return nil, err
	}

	// Register config change listener
	config.RegisterListener(func(newCfg *config.Config) {
		p.mu.Lock()
		p.config = newCfg
		p.mu.Unlock()

		// Reinitialize providers with new config
		if err := p.initializeProviders(); err != nil {
			logger.Error("Failed to reinitialize providers on config change", "error", err)
		}
	})

	return p, nil
}

// initializeProviders initializes the AI providers
func (p *Processor) initializeProviders() error {
	cfg := p.config

	// Get current provider name
	currentProvider := cfg.GetCurrentProvider()

	// Get decrypted API key for current provider
	apiKey, err := cfg.GetCurrentAPIKey()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Get provider settings
	settings := cfg.GetProviderSettings(currentProvider)

	// Register providers based on configuration
	switch currentProvider {
	case "openai":
		provider := ai.NewOpenAIProvider(apiKey, settings.BaseURL, settings.Model)
		p.providerFactory.Register("openai", provider)
		p.providerFactory.SetCurrent("openai")

	case "claude":
		provider := ai.NewAnthropicProvider(apiKey, settings.BaseURL, settings.Model)
		p.providerFactory.Register("claude", provider)
		p.providerFactory.SetCurrent("claude")

	case "gemini":
		provider := ai.NewGeminiProvider(apiKey, settings.BaseURL, settings.Model)
		p.providerFactory.Register("gemini", provider)
		p.providerFactory.SetCurrent("gemini")

	default:
		return fmt.Errorf("unknown provider: %s", currentProvider)
	}

	logger.Info("AI provider initialized", "provider", currentProvider)
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

	// Save current clipboard
	if err := p.clipboardManager.SaveCurrent(); err != nil {
		return fmt.Errorf("failed to save clipboard: %w", err)
	}

	// Select all text
	if err := input.SimulateSelectAll(); err != nil {
		return fmt.Errorf("failed to select all: %w", err)
	}

	// Copy to clipboard
	if err := input.SimulateCopy(); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Get text from clipboard
	text, err := p.clipboardManager.GetText()
	if err != nil {
		return fmt.Errorf("failed to get clipboard text: %w", err)
	}

	// Process the text
	revisedText, err := p.reviseText(text)
	if err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to revise text: %w", err)
	}

	// Replace the text
	if err := p.clipboardManager.SetText(revisedText); err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to set revised text: %w", err)
	}

	// Select all again and paste
	if err := input.SimulateSelectAll(); err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to select all for paste: %w", err)
	}

	if err := input.SimulatePaste(); err != nil {
		p.clipboardManager.Restore()
		return fmt.Errorf("failed to paste: %w", err)
	}

	// Restore original clipboard
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

	// Capture selected text
	text, err := p.clipboardManager.CaptureSelectedText()
	if err != nil {
		return fmt.Errorf("failed to capture selected text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("no text selected")
	}

	// Process the text
	revisedText, err := p.reviseText(text)
	if err != nil {
		return fmt.Errorf("failed to revise text: %w", err)
	}

	// Replace selected text
	if err := p.clipboardManager.ReplaceSelectedText(revisedText); err != nil {
		return fmt.Errorf("failed to replace selected text: %w", err)
	}

	logger.Info("Selection revision completed")
	return nil
}

// reviseText sends text to the AI provider for revision
func (p *Processor) reviseText(text string) (string, error) {
	p.mu.Lock()
	cfg := p.config
	p.mu.Unlock()

	// Check character limit
	if len(text) > cfg.Revision.CharacterLimit {
		return "", fmt.Errorf("text exceeds character limit (%d > %d)",
			len(text), cfg.Revision.CharacterLimit)
	}

	// Get current provider
	provider := p.providerFactory.GetCurrent()
	if provider == nil {
		return "", fmt.Errorf("no AI provider configured")
	}

	// Create context with timeout
	timeout := time.Duration(p.config.Revision.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("Sending text to AI provider",
		"provider", provider.GetName(),
		"text_length", len(text))

	// Revise text with latest system prompt
	revised, err := provider.ReviseText(ctx, text, p.config.Revision.SystemPrompt)
	if err != nil {
		return "", fmt.Errorf("AI revision failed: %w", err)
	}

	logger.Info("Text revised successfully",
		"original_length", len(text),
		"revised_length", len(revised))

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
