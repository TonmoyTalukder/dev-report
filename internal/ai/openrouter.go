package ai

import "github.com/dev-report/dev-report/internal/constants"

// OpenRouter free models (append :free to use free tier).
// Get a free API key at: https://openrouter.ai/keys
// NewOpenRouter creates a Provider for the OpenRouter API.
// OpenRouter uses the OpenAI-compatible API format with extra headers.
func NewOpenRouter(apiKey string) Provider {
	return NewOpenAICompat(
		constants.ProviderOpenRouter,
		"https://openrouter.ai/api/v1",
		apiKey,
		constants.DefaultOpenRouterModel,
		map[string]string{
			"HTTP-Referer": constants.RepoURL,
			"X-Title":      constants.AppName,
		},
	)
}
