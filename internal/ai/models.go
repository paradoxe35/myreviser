package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/paradoxe35/myreviser/internal/config"
)

type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Created int64  `json:"created,omitempty"`
}

func FetchModelsOpenAI(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	url := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, truncateString(string(body), 200))
	}

	var response struct {
		Data []ModelInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].ID < response.Data[j].ID
	})

	return response.Data, nil
}

func FetchModelsAnthropic(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error) {
	return []ModelInfo{
		{ID: "claude-3-5-sonnet-20241022"},
		{ID: "claude-3-5-haiku-20241022"},
		{ID: "claude-3-opus-20240229"},
		{ID: "claude-3-sonnet-20240229"},
		{ID: "claude-3-haiku-20240307"},
	}, nil
}

func FetchModels(ctx context.Context, providerType, apiKey, baseURL string) ([]ModelInfo, error) {
	switch providerType {
	case config.ProviderTypeOpenAICompatible:
		return FetchModelsOpenAI(ctx, apiKey, baseURL)
	case config.ProviderTypeAnthropicCompatible:
		return FetchModelsAnthropic(ctx, apiKey, baseURL)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
