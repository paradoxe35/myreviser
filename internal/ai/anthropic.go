package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AnthropicProvider implements the Provider interface for Claude
type AnthropicProvider struct {
	APIKey  string
	BaseURL string
	Model   string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic/Claude provider
func NewAnthropicProvider(apiKey, baseURL, model string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-3-5-haiku-20241022"
	}
	return &AnthropicProvider{
		APIKey:  apiKey,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		client:  &http.Client{},
	}
}

// AnthropicRequest represents the request structure for Claude API
type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
}

// AnthropicMessage represents a message in the Claude API
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response from Claude API
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ReviseText sends text to Claude for revision
func (p *AnthropicProvider) ReviseText(ctx context.Context, text, systemPrompt string) (string, error) {
	if err := p.ValidateConfig(); err != nil {
		return "", err
	}

	messages := []AnthropicMessage{
		{Role: "user", Content: text},
	}

	requestBody := AnthropicRequest{
		Model:     p.Model,
		Messages:  messages,
		MaxTokens: 4096,
		System:    systemPrompt,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	// Concatenate all text content
	var result strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			result.WriteString(content.Text)
		}
	}

	return result.String(), nil
}

// ValidateConfig validates the provider configuration
func (p *AnthropicProvider) ValidateConfig() error {
	if p.APIKey == "" {
		return fmt.Errorf("anthropic API key is required")
	}
	if p.BaseURL == "" {
		return fmt.Errorf("anthropic base URL is required")
	}
	return nil
}

// GetName returns the provider name
func (p *AnthropicProvider) GetName() string {
	return "claude"
}

// GetModel returns the model being used
func (p *AnthropicProvider) GetModel() string {
	return p.Model
}
