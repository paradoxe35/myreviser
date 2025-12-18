package ai

import (
	"context"
	"fmt"

	"github.com/paradoxe35/myreviser/internal/config"
)

type CustomProvider struct {
	name         string
	providerType string
	inner        Provider
}

func NewCustomProvider(name, providerType, apiKey, baseURL, model string, temperature float64) (*CustomProvider, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required for custom providers")
	}
	if providerType == "" {
		providerType = config.ProviderTypeOpenAICompatible
	}

	var inner Provider
	switch providerType {
	case config.ProviderTypeOpenAICompatible:
		inner = NewOpenAIProvider(apiKey, baseURL, model, temperature)
	case config.ProviderTypeAnthropicCompatible:
		inner = NewAnthropicProvider(apiKey, baseURL, model, temperature)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	return &CustomProvider{
		name:         name,
		providerType: providerType,
		inner:        inner,
	}, nil
}

func (p *CustomProvider) ReviseText(ctx context.Context, text, systemPrompt string) (string, error) {
	return p.inner.ReviseText(ctx, text, systemPrompt)
}

func (p *CustomProvider) ValidateConfig() error {
	return p.inner.ValidateConfig()
}

func (p *CustomProvider) GetName() string {
	return p.name
}

func (p *CustomProvider) GetModel() string {
	return p.inner.GetModel()
}

func (p *CustomProvider) GetTemperature() float64 {
	return p.inner.GetTemperature()
}

func (p *CustomProvider) GetProviderType() string {
	return p.providerType
}
