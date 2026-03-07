package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// openAICompatClient is a reusable client for any OpenAI-compatible API
// (Groq, Ollama /v1, OpenRouter, etc.).
type openAICompatClient struct {
	name    string
	baseURL string
	apiKey  string
	model   string
	headers map[string]string
}

// NewOpenAICompat creates a new OpenAI-compatible provider.
// Pass an empty apiKey for providers that don't require one (e.g. Ollama).
func NewOpenAICompat(name, baseURL, apiKey, model string, extraHeaders ...map[string]string) Provider {
	headers := map[string]string{}
	for _, h := range extraHeaders {
		for k, v := range h {
			headers[k] = v
		}
	}
	return &openAICompatClient{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		headers: headers,
	}
}

func (c *openAICompatClient) Name() string { return c.name }

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openAICompatClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   2048,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s request failed: %w", c.name, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s API error %d: %s", c.name, resp.StatusCode, truncate(string(body), 300))
	}

	var result chatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("%s response parse error: %w", c.name, err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("%s error: %s", c.name, result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("%s returned no choices", c.name)
	}

	return result.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
