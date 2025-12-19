package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/paradoxe35/myreviser/internal/utils"
)

const (
	ProviderTypeOpenAICompatible = "openai-compatible"

	BuiltInOpenAI = "openai"
	BuiltInClaude = "claude"
	BuiltInGemini = "gemini"
)

type Config struct {
	mu         sync.RWMutex
	AIProvider AIProviderConfig `json:"ai_provider"`
	Hotkeys    HotkeyConfig     `json:"hotkeys"`
	Revision   RevisionConfig   `json:"revision"`
	Appearance AppearanceConfig `json:"appearance"`
	Meta       MetaConfig       `json:"meta"`
}

type ProviderSettings struct {
	APIKey       string  `json:"api_key"`
	BaseURL      string  `json:"base_url,omitempty"`
	Model        string  `json:"model,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	IsCustom     bool    `json:"is_custom,omitempty"`
	ProviderType string  `json:"provider_type,omitempty"`
}

type AIProviderConfig struct {
	Provider  string                      `json:"provider"` // "openai" | "claude" | "gemini"
	Providers map[string]ProviderSettings `json:"providers"`
}

type HotkeyConfig struct {
	SelectAll string `json:"select_all"`
	Selection string `json:"selection"`
}

type RevisionConfig struct {
	CharacterLimit         int    `json:"character_limit"`
	SystemPrompt           string `json:"system_prompt"`
	TimeoutSeconds         int    `json:"timeout_seconds"`
	EnableProviderMentions bool   `json:"enable_provider_mentions"`
}

type AppearanceConfig struct {
	Theme          string `json:"theme"` // "auto" | "light" | "dark"
	StartMinimized bool   `json:"start_minimized"`
	StartOnLogin   bool   `json:"start_on_login"`
}

type MetaConfig struct {
	FirstRun bool `json:"first_run"`
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
	listeners     []func(*Config)
	listenerMutex sync.RWMutex
)

const APP_ID = "me.pngwasi.myreviser"

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	return utils.AppHomeDir("config.json")
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		AIProvider: AIProviderConfig{
			Provider: "openai",
			Providers: map[string]ProviderSettings{
				"openai": {
					BaseURL:     "https://api.openai.com/v1",
					Model:       "gpt-4o",
					Temperature: 1.0,
				},
				"claude": {
					BaseURL:     "https://api.anthropic.com",
					Model:       "claude-3-5-haiku-20241022",
					Temperature: 1.0,
				},
				"gemini": {
					BaseURL:     "https://generativelanguage.googleapis.com",
					Model:       "gemini-2.5-flash",
					Temperature: 1.0,
				},
			},
		},
		Hotkeys:    GetPlatformHotkeys(),
		Revision:   GetDefaultRevision(),
		Appearance: GetDefaultAppearance(),
		Meta:       MetaConfig{FirstRun: true},
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
		TimeoutSeconds:         30,
		EnableProviderMentions: true,
	}
}

// GetDefaultAppearance returns default appearance settings
func GetDefaultAppearance() AppearanceConfig {
	return AppearanceConfig{
		Theme:          "auto",
		StartMinimized: false,
		StartOnLogin:   false,
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

	if cfg.AIProvider.Providers != nil {
		for name, settings := range cfg.AIProvider.Providers {
			if settings.Temperature == 0 {
				settings.Temperature = 1.0
				cfg.AIProvider.Providers[name] = settings
			}
		}
	}

	currentConfig = cfg
	return cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	c.mu.Lock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to write config file: %w", err)
	}
	c.mu.Unlock()

	// Update global config and notify listeners
	configMutex.Lock()
	currentConfig = c
	configMutex.Unlock()

	notifyListeners(c)

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

func RegisterListener(listener func(*Config)) {
	listenerMutex.Lock()
	defer listenerMutex.Unlock()
	listeners = append(listeners, listener)
}

func notifyListeners(cfg *Config) {
	listenerMutex.RLock()
	defer listenerMutex.RUnlock()

	for _, listener := range listeners {
		go listener(cfg)
	}
}

// Reload reloads the configuration from disk and notifies listeners
func Reload() (*Config, error) {
	cfg, err := Load()
	if err == nil {
		notifyListeners(cfg)
	}
	return cfg, err
}

// GetProviderSettings returns settings for a specific provider
func (c *Config) GetProviderSettings(provider string) ProviderSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.AIProvider.Providers == nil {
		return ProviderSettings{}
	}

	settings, ok := c.AIProvider.Providers[provider]
	if !ok {
		// Return defaults for this provider
		defaults := Default()
		if defaultSettings, ok := defaults.AIProvider.Providers[provider]; ok {
			return defaultSettings
		}
		return ProviderSettings{}
	}

	if settings.Temperature == 0 {
		settings.Temperature = 1.0
	}

	return settings
}

// SetProviderSettings updates settings for a specific provider
func (c *Config) SetProviderSettings(provider string, settings ProviderSettings) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AIProvider.Providers == nil {
		c.AIProvider.Providers = make(map[string]ProviderSettings)
	}

	if settings.Temperature == 0 {
		settings.Temperature = 1.0
	}

	c.AIProvider.Providers[provider] = settings
}

// GetCurrentProvider returns the currently selected provider name
func (c *Config) GetCurrentProvider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AIProvider.Provider
}

// SetCurrentProvider sets the currently selected provider
func (c *Config) SetCurrentProvider(provider string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AIProvider.Provider = provider
}

func BuiltInProviders() []string {
	return []string{BuiltInOpenAI, BuiltInClaude, BuiltInGemini}
}

func IsBuiltInProvider(name string) bool {
	nameLower := strings.ToLower(name)
	for _, p := range BuiltInProviders() {
		if strings.ToLower(p) == nameLower {
			return true
		}
	}
	return false
}

func (c *Config) GetAllProviderNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0)
	for _, name := range BuiltInProviders() {
		names = append(names, name)
	}

	customNames := make([]string, 0)
	for name, settings := range c.AIProvider.Providers {
		if settings.IsCustom {
			customNames = append(customNames, name)
		}
	}
	sort.Strings(customNames)
	names = append(names, customNames...)
	return names
}

func (c *Config) AddCustomProvider(name string, settings ProviderSettings) error {
	if IsBuiltInProvider(name) {
		return fmt.Errorf("cannot use built-in provider name: %s", name)
	}
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AIProvider.Providers == nil {
		c.AIProvider.Providers = make(map[string]ProviderSettings)
	}

	nameLower := strings.ToLower(name)
	for existingName := range c.AIProvider.Providers {
		if strings.ToLower(existingName) == nameLower {
			return fmt.Errorf("provider with name '%s' already exists", existingName)
		}
	}

	settings.IsCustom = true
	c.AIProvider.Providers[name] = settings
	return nil
}

func (c *Config) DeleteCustomProvider(name string) error {
	if IsBuiltInProvider(name) {
		return fmt.Errorf("cannot delete built-in provider: %s", name)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	settings, exists := c.AIProvider.Providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}
	if !settings.IsCustom {
		return fmt.Errorf("cannot delete non-custom provider: %s", name)
	}

	delete(c.AIProvider.Providers, name)

	if c.AIProvider.Provider == name {
		c.AIProvider.Provider = BuiltInOpenAI
	}

	return nil
}

func (c *Config) IsCustomProvider(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if settings, exists := c.AIProvider.Providers[name]; exists {
		return settings.IsCustom
	}
	return false
}
