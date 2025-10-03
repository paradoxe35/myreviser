package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseAPIError creates a user-friendly error message from API response
func ParseAPIError(statusCode int, body []byte, providerName string) error {
	// Try to parse as generic structured error
	var errResp struct {
		Error *struct {
			Message string      `json:"message"`
			Type    string      `json:"type"`
			Code    interface{} `json:"code,omitempty"`
			Status  string      `json:"status,omitempty"`
		} `json:"error,omitempty"`
	}

	if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
		return fmt.Errorf("API error (%d): %s", statusCode, errResp.Error.Message)
	}

	// Fallback to status-based messages
	switch statusCode {
	case 401:
		return fmt.Errorf("authentication failed: invalid API key or credentials")
	case 403:
		return fmt.Errorf("access forbidden: check your API key permissions")
	case 404:
		return fmt.Errorf("endpoint not found: verify the base URL is correct")
	case 429:
		return fmt.Errorf("rate limit exceeded: please try again later")
	case 500, 502, 503, 504:
		return fmt.Errorf("API server error (%d): service may be temporarily unavailable", statusCode)
	default:
		// Include first 100 chars of body if available
		preview := string(body)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		if len(preview) > 0 {
			return fmt.Errorf("API request failed (%d): %s", statusCode, preview)
		}
		return fmt.Errorf("API request failed with status code %d", statusCode)
	}
}

// ParseUnmarshalError handles JSON parsing errors with helpful messages
func ParseUnmarshalError(err error, body []byte, statusCode int, providerName string) error {
	bodyPreview := string(body)
	if len(bodyPreview) > 200 {
		bodyPreview = bodyPreview[:200] + "..."
	}

	// Check for common issues
	if len(body) == 0 {
		return fmt.Errorf("received empty response from API: verify the base URL is correct")
	}

	if strings.Contains(string(body), "<!DOCTYPE") || strings.Contains(string(body), "<html") {
		return fmt.Errorf("received HTML instead of JSON: verify the base URL points to the API endpoint")
	}

	if statusCode == 404 {
		var exampleURL string
		switch providerName {
		case "openai":
			exampleURL = "https://api.openai.com/v1"
		case "claude":
			exampleURL = "https://api.anthropic.com"
		case "gemini":
			exampleURL = "https://generativelanguage.googleapis.com"
		}
		if exampleURL != "" {
			return fmt.Errorf("endpoint not found: verify the base URL is correct (e.g., %s)", exampleURL)
		}
		return fmt.Errorf("endpoint not found: verify the base URL is correct")
	}

	// Generic unmarshal error with context
	return fmt.Errorf("invalid API response format: %v (status: %d, response: %s)", err, statusCode, bodyPreview)
}
