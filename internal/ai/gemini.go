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

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	APIKey      string
	BaseURL     string
	Model       string
	Temperature float64
	client      *http.Client
}

// NewGeminiProvider creates a new Google Gemini provider
func NewGeminiProvider(apiKey, baseURL, model string, temperature float64) *GeminiProvider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	if temperature == 0 {
		temperature = 1.0
	}
	return &GeminiProvider{
		APIKey:      apiKey,
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Model:       model,
		Temperature: temperature,
		client:      &http.Client{},
	}
}

type ThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type GenerationConfig struct {
	ThinkingConfig ThinkingConfig `json:"thinkingConfig"`
	Temperature    float64        `json:"temperature"`
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent  `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

// GeminiContent represents content in the Gemini API
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// ReviseText sends text to Gemini for revision
func (p *GeminiProvider) ReviseText(ctx context.Context, text, systemPrompt string) (string, error) {
	if err := p.ValidateConfig(); err != nil {
		return "", err
	}

	// Combine system prompt and user text for Gemini
	fullText := fmt.Sprintf("%s\n\n%s", systemPrompt, text)

	contents := []GeminiContent{
		{
			Parts: []GeminiPart{
				{Text: fullText},
			},
		},
	}

	requestBody := GeminiRequest{
		Contents: contents,
		GenerationConfig: GenerationConfig{
			ThinkingConfig: ThinkingConfig{
				ThinkingBudget: 0,
			},
			Temperature: p.Temperature,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", p.BaseURL, p.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", ParseAPIError(resp.StatusCode, body, "gemini")
	}

	var response GeminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", ParseUnmarshalError(err, body, resp.StatusCode, "gemini")
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	// Extract text from response
	var result strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		result.WriteString(part.Text)
	}

	return result.String(), nil
}

// ValidateConfig validates the provider configuration
func (p *GeminiProvider) ValidateConfig() error {
	if p.APIKey == "" {
		return fmt.Errorf("gemini API key is required")
	}
	if p.BaseURL == "" {
		return fmt.Errorf("gemini base URL is required")
	}
	return nil
}

// GetName returns the provider name
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// GetModel returns the model being used
func (p *GeminiProvider) GetModel() string {
	return p.Model
}

// GetModel returns the temperature being used
func (p *GeminiProvider) GetTemperature() float64 {
	return p.Temperature
}
