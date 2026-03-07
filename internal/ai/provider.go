package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/dev-report/dev-report/internal/config"
	"github.com/dev-report/dev-report/internal/constants"
)

// Provider is the interface every AI backend must implement.
type Provider interface {
	// Generate sends a prompt and returns the raw text response.
	Generate(ctx context.Context, prompt string) (string, error)
	// Name returns a human-readable provider name.
	Name() string
}

// New creates the appropriate Provider based on config.
// Falls back to Groq if the requested provider is unknown.
func New(cfg *config.Config, providerOverride string) (Provider, error) {
	provider := cfg.AIProvider
	if providerOverride != "" {
		provider = providerOverride
	}

	switch provider {
	case constants.ProviderGroq:
		key := cfg.GroqAPIKey
		if key == "" {
			return nil, fmt.Errorf(
				"Groq API key not set.\n" +
					"  Set it with: export " + constants.EnvGroqAPIKey + "=<your-key>\n" +
					"  Or add \"groqApiKey\" to dev-report.config.json\n" +
					"  Get a free key at: " + constants.GroqConsoleURL,
			)
		}
		model := cfg.GroqModel
		if model == "" {
			model = constants.DefaultGroqModel
		}
		return NewOpenAICompat(constants.ProviderGroq, "https://api.groq.com/openai/v1", key, model), nil

	case constants.ProviderGemini:
		key := cfg.GeminiAPIKey
		if key == "" {
			return nil, fmt.Errorf(
				"Gemini API key not set.\n" +
					"  Set it with: export " + constants.EnvGeminiAPIKey + "=<your-key>\n" +
					"  Or add \"geminiApiKey\" to dev-report.config.json\n" +
					"  Get a free key at: " + constants.GeminiConsoleURL,
			)
		}
		return NewGemini(key), nil

	case constants.ProviderOllama:
		url := cfg.OllamaURL
		if url == "" {
			url = constants.DefaultOllamaURL
		}
		model := cfg.OllamaModel
		if model == "" {
			model = constants.DefaultOllamaModel
		}
		return NewOpenAICompat(constants.ProviderOllama, url+"/v1", "", model), nil

	case constants.ProviderOpenRouter:
		key := cfg.OpenRouterKey
		if key == "" {
			return nil, fmt.Errorf(
				"OpenRouter API key not set.\n" +
					"  Set it with: export " + constants.EnvOpenRouterAPIKey + "=<your-key>\n" +
					"  Or add \"openRouterApiKey\" to dev-report.config.json\n" +
					"  Get a free key at: " + constants.OpenRouterKeysURL,
			)
		}
		return NewOpenRouter(key), nil

	default:
		return nil, fmt.Errorf(
			"unknown AI provider %q — supported: %s", provider, strings.Join(constants.SupportedProviders, ", "),
		)
	}
}
