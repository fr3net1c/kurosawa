package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenRouterProvider struct {
	apiKey string
	model  string
	client *http.Client
}

type OpenRouterRequest struct {
	Model    string                 `json:"model"`
	Messages []OpenRouterMessage    `json:"messages"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenRouterChoice `json:"choices"`
	Usage   OpenRouterUsage    `json:"usage"`
}

type OpenRouterChoice struct {
	Index        int               `json:"index"`
	Message      OpenRouterMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type OpenRouterUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost,omitempty"`
}

func NewOpenRouterProvider(apiKey, model string) (*OpenRouterProvider, error) {
	if model == "" {
		model = "openai/gpt-5.1"
	}
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (o *OpenRouterProvider) GetResponse(ctx context.Context, prompt string) (string, error) {
	reqBody := OpenRouterRequest{
		Model: o.model,
		Messages: []OpenRouterMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	// OpenRouter requires a referer header
	req.Header.Set("HTTP-Referer", "https://yourapp.com")
	req.Header.Set("X-Title", "Kurosawa Bot")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection failed: openrouter api is unreachable. Check your internet connection and try again. Error: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Handle HTTP error status codes with helpful messages
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusTooManyRequests: // 429
			return "", fmt.Errorf("rate limited (HTTP 429): openrouter api received too many requests. Your account or API key is under heavy load. Wait a few minutes and try again")
		case http.StatusServiceUnavailable: // 503
			return "", fmt.Errorf("service unavailable (HTTP 503): OpenRouter or the underlying model service is temporarily down. Try again in a few moments")
		case http.StatusUnauthorized: // 401
			return "", fmt.Errorf("authentication failed (HTTP 401): Invalid API key. Check your OPENROUTER_API_KEY in .env")
		case http.StatusForbidden: // 403
			return "", fmt.Errorf("permission denied (HTTP 403): Your API key may not have access to this model. Check your OpenRouter account permissions")
		case http.StatusBadRequest: // 400
			return "", fmt.Errorf("bad request (HTTP 400): Invalid model or request format. Check that model '%s' exists on OpenRouter", o.model)
		default:
			return "", fmt.Errorf("openrouter api returned status %d: %s", resp.StatusCode, string(respBytes))
		}
	}

	var respData OpenRouterResponse
	if err := json.Unmarshal(respBytes, &respData); err != nil {
		return "", err
	}

	if len(respData.Choices) == 0 {
		return "Sorry, I cannot respond to this.", nil
	}

	response := respData.Choices[0].Message.Content
	if response == "" {
		return "Sorry, I cannot respond to this.", nil
	}

	return response, nil
}

func (o *OpenRouterProvider) GetName() string {
	return "OpenRouter"
}

func (o *OpenRouterProvider) GetAvailableModels() []string {
	return []string{
		"openai/gpt-5.1",
		"openai/gpt-4o",
		"openai/gpt-4.1",
		"google/gemini-3-pro",
		"mistral/mistral-large",
		"anthropic/claude-opus-4.5",
		"anthropic/claude-sonnet-4.5",
	}
}

func (o *OpenRouterProvider) SetModel(model string) {
	o.model = model
}
