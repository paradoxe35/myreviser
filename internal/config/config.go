package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Config struct {
	mu         sync.RWMutex
	AIProvider AIProviderConfig `json:"ai_provider"`
	Hotkeys    HotkeyConfig     `json:"hotkeys"`
	Revision   RevisionConfig   `json:"revision"`
	Appearance AppearanceConfig `json:"appearance"`
}

type AIProviderConfig struct {
	Provider string `json:"provider"` // "openai" | "claude" | "gemini"
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model,omitempty"`
}

type HotkeyConfig struct {
	SelectAll string `json:"select_all"`
	Selection string `json:"selection"`
}

type RevisionConfig struct {
	CharacterLimit int    `json:"character_limit"`
	SystemPrompt   string `json:"system_prompt"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type AppearanceConfig struct {
	Theme          string `json:"theme"` // "auto" | "light" | "dark"
	StartMinimized bool   `json:"start_minimized"`
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
)

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".myreviser", "config.json")
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		AIProvider: AIProviderConfig{
			Provider: "openai",
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4o-mini",
		},
		Hotkeys:    GetPlatformHotkeys(),
		Revision:   GetDefaultRevision(),
		Appearance: GetDefaultAppearance(),
	}
}

// GetPlatformHotkeys returns platform-specific hotkey defaults
func GetPlatformHotkeys() HotkeyConfig {
	switch runtime.GOOS {
	case "darwin":
		return HotkeyConfig{
			SelectAll: "ctrl+option+space",
			Selection: "ctrl+cmd",
		}
	case "windows":
		return HotkeyConfig{
			SelectAll: "ctrl+alt+space",
			Selection: "ctrl+win",
		}
	default: // Linux
		return HotkeyConfig{
			SelectAll: "ctrl+alt+space",
			Selection: "ctrl+super",
		}
	}
}

// GetDefaultRevision returns default revision settings
func GetDefaultRevision() RevisionConfig {
	return RevisionConfig{
		CharacterLimit: 1000,
		SystemPrompt: "You are a multilingual text enhancer: fix errors, improve clarity and quality " +
			"while preserving tone, context, and intent in the original language. " +
			"Return only the enhanced version without additional text.",
		TimeoutSeconds: 30,
	}
}

// GetDefaultAppearance returns default appearance settings
func GetDefaultAppearance() AppearanceConfig {
	return AppearanceConfig{
		Theme:          "auto",
		StartMinimized: true,
	}
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Ensure config directory exists
	configDir := filepath.Dir(ConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		// Create default config
		cfg := Default()
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		currentConfig = cfg
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults for any missing fields
	if cfg.Revision.CharacterLimit == 0 {
		cfg.Revision.CharacterLimit = 1000
	}
	if cfg.Revision.TimeoutSeconds == 0 {
		cfg.Revision.TimeoutSeconds = 30
	}
	if cfg.Revision.SystemPrompt == "" {
		cfg.Revision.SystemPrompt = GetDefaultRevision().SystemPrompt
	}

	currentConfig = cfg
	return cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

// Update updates the configuration with the given function
func Update(fn func(*Config)) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if currentConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	fn(currentConfig)
	return currentConfig.Save()
}