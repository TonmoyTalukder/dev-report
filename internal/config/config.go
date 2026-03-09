package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dev-report/dev-report/internal/constants"
)

// Config holds all tool configuration, merged from file + environment variables.
type Config struct {
	User           string `json:"user"`
	GitHubUsername string `json:"githubUsername"`
	AIProvider     string `json:"aiProvider"`
	GroqAPIKey     string `json:"groqApiKey"`
	GeminiAPIKey   string `json:"geminiApiKey"`
	OpenRouterKey  string `json:"openRouterApiKey"`
	OllamaURL      string `json:"ollamaUrl"`
	OllamaModel    string `json:"ollamaModel"`
	GroqModel      string `json:"groqModel"`
	DefaultOutput  string `json:"defaultOutput"`
}

// Defaults returns a Config with sensible defaults applied.
func Defaults() *Config {
	return &Config{
		AIProvider:    constants.DefaultAIProvider,
		DefaultOutput: constants.DefaultOutput,
		OllamaURL:     constants.DefaultOllamaURL,
		OllamaModel:   constants.DefaultOllamaModel,
		GroqModel:     constants.DefaultGroqModel,
	}
}

// Load reads dev-report.config.json from dir (if it exists) and merges
// environment variables on top. Missing file is not an error.
func Load(dir string) (*Config, error) {
	cfg := Defaults()

	path := filepath.Join(dir, constants.ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnv(cfg)
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnv(cfg)
	return cfg, nil
}

// applyEnv overlays environment variable values onto cfg.
func applyEnv(cfg *Config) {
	if v := os.Getenv(constants.EnvGroqAPIKey); v != "" {
		cfg.GroqAPIKey = v
	}
	if v := os.Getenv(constants.EnvGeminiAPIKey); v != "" {
		cfg.GeminiAPIKey = v
	}
	if v := os.Getenv(constants.EnvOpenRouterAPIKey); v != "" {
		cfg.OpenRouterKey = v
	}
	if v := os.Getenv(constants.EnvOllamaURL); v != "" {
		cfg.OllamaURL = v
	}
	if v := os.Getenv(constants.EnvOllamaModel); v != "" {
		cfg.OllamaModel = v
	}
	if v := os.Getenv(constants.EnvGroqModel); v != "" {
		cfg.GroqModel = v
	}
	if v := os.Getenv(constants.EnvDevReportAPIKey); v != "" {
		// Generic fallback: applies to whichever provider is active
		if cfg.GroqAPIKey == "" {
			cfg.GroqAPIKey = v
		}
	}
	if v := os.Getenv(constants.EnvDevReportAI); v != "" {
		cfg.AIProvider = v
	}
	if v := os.Getenv(constants.EnvDevReportOutput); v != "" {
		cfg.DefaultOutput = v
	}
}

// APIKeyForProvider returns the API key relevant to the given provider name.
func (c *Config) APIKeyForProvider(provider string) string {
	switch provider {
	case constants.ProviderGroq:
		return c.GroqAPIKey
	case constants.ProviderGemini:
		return c.GeminiAPIKey
	case constants.ProviderOpenRouter:
		return c.OpenRouterKey
	case constants.ProviderOllama:
		return "" // no key needed
	default:
		return c.GroqAPIKey
	}
}

// Write saves the config to dev-report.config.json in dir.
func Write(dir string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, constants.ConfigFileName)
	return os.WriteFile(path, data, 0644)
}
