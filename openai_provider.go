package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, model string) (*OpenAIProvider, error) {
	if model == "" {
		model = "gpt-5.1"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIProvider{
		client: &client,
		model:  model,
	}, nil
}

func (o *OpenAIProvider) GetResponse(ctx context.Context, prompt string) (string, error) {
	message, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: shared.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		// Provide more helpful error messages
		errMsg := err.Error()
		if errMsg == "429" || contains(errMsg, "rate limit") {
			return "", fmt.Errorf("rate limited (HTTP 429): openai api received too many requests. Your account or API key may be under heavy load. Wait a few minutes and try again")
		}
		if errMsg == "503" || contains(errMsg, "service unavailable") {
			return "", fmt.Errorf("service unavailable (HTTP 503): openai api is temporarily down. Check https://status.openai.com")
		}
		if errMsg == "401" || contains(errMsg, "unauthorized") {
			return "", fmt.Errorf("authentication failed (HTTP 401): Invalid API key. Check your OPENAI_API_KEY in .env")
		}
		if errMsg == "403" || contains(errMsg, "forbidden") {
			return "", fmt.Errorf("permission denied (HTTP 403): Your API key may not have access to this model. Check your OpenAI account permissions")
		}
		if contains(errMsg, "timeout") || contains(errMsg, "connection") {
			return "", fmt.Errorf("connection timeout: Network request failed. Check your internet connection and try again")
		}
		return "", fmt.Errorf("failed to call openai api: %v", err)
	}

	if len(message.Choices) == 0 {
		return "Sorry, I cannot respond to this.", nil
	}

	response := message.Choices[0].Message.Content
	if response == "" {
		return "Sorry, I cannot respond to this.", nil
	}

	return response, nil
}

func (o *OpenAIProvider) GetName() string {
	return "OpenAI"
}

func (o *OpenAIProvider) GetAvailableModels() []string {
	return []string{
		"gpt-5.1",
		"gpt-5",
		"gpt-4.1",
		"o3",
	}
}

func (o *OpenAIProvider) SetModel(model string) {
	o.model = model
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
