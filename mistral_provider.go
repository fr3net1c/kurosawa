package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MistralProvider struct {
	apiKey string
	model  string
	client *http.Client
}

type MistralRequest struct {
	Model    string                 `json:"model"`
	Messages []MistralMessage       `json:"messages"`
	SafeMode bool                   `json:"safe_mode"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

type MistralMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MistralResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []MistralChoice `json:"choices"`
	Usage   MistralUsage    `json:"usage"`
}

type MistralChoice struct {
	Index        int            `json:"index"`
	Message      MistralMessage `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type MistralUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewMistralProvider(apiKey, model string) (*MistralProvider, error) {
	if model == "" {
		model = "mistral-large-latest"
	}
	return &MistralProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (m *MistralProvider) GetResponse(ctx context.Context, prompt string) (string, error) {
	reqBody := MistralRequest{
		Model: m.model,
		Messages: []MistralMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		SafeMode: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection failed: mistral api is unreachable. Check your internet connection and try again. Error: %v", err)
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
			return "", fmt.Errorf("rate limited (HTTP 429): mistral api received too many requests. Your account or API key is under heavy load. Wait a few minutes and try again")
		case http.StatusServiceUnavailable: // 503
			return "", fmt.Errorf("service unavailable (HTTP 503): mistral api is temporarily down. Check https://status.mistral.ai")
		case http.StatusUnauthorized: // 401
			return "", fmt.Errorf("authentication failed (HTTP 401): Invalid API key. Check your MISTRAL_API_KEY in .env")
		case http.StatusForbidden: // 403
			return "", fmt.Errorf("permission denied (HTTP 403): Your API key may not have access to this model. Check your Mistral account permissions")
		case http.StatusBadRequest: // 400
			return "", fmt.Errorf("bad request (HTTP 400): Invalid model or request format. Check that model '%s' is valid", m.model)
		default:
			//nolint:ST1005
			return "", fmt.Errorf("mistral api returned status %d: %s", resp.StatusCode, string(respBytes))
		}
	}

	var respData MistralResponse
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

func (m *MistralProvider) GetName() string {
	return "Mistral"
}

func (m *MistralProvider) GetAvailableModels() []string {
	return []string{
		"mistral-large-latest",
		"mistral-medium-latest",
		"devstral-small-latest",
	}
}

func (m *MistralProvider) SetModel(model string) {
	m.model = model
}
