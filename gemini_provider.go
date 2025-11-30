package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	os.Setenv("GEMINI_API_KEY", apiKey)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &GeminiProvider{
		client: client,
	}, nil
}

func (g *GeminiProvider) GetResponse(ctx context.Context, prompt string) (string, error) {
	safetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
	}

	config := &genai.GenerateContentConfig{
		SafetySettings: safetySettings,
	}

	var result *genai.GenerateContentResponse
	const maxRetries = 5
	baseDelay := 1 * time.Second
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = g.client.Models.GenerateContent(
			ctx,
			"gemini-3-pro",
			genai.Text(prompt),
			config,
		)
		if err == nil {
			break
		}

		if googleapiErr, ok := err.(*googleapi.Error); ok {
			if googleapiErr.Code == 429 {
				delay := baseDelay * time.Duration(1<<uint(i))
				fmt.Printf("Rate limit exceeded (HTTP 429). Retrying in %v... (attempt %d/%d)\n", delay, i+1, maxRetries)
				time.Sleep(delay)
				continue
			}
			if googleapiErr.Code == 503 {
				return "", fmt.Errorf("service unavailable (HTTP 503): gemini api is temporarily overloaded or down. Try again in a few moments")
			}
			if googleapiErr.Code == 401 || googleapiErr.Code == 403 {
				return "", fmt.Errorf("authentication failed (HTTP %d): Invalid API key. Check your GEMINI_API_KEY in .env", googleapiErr.Code)
			}
		}

		return "", fmt.Errorf("failed to call gemini api: %v", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to call gemini api after %d retries: %v. Free tier models may have capacity limits during high demand. Try again later or use a different model", maxRetries, err)
	}

	response := result.Text()
	if response == "" {
		return "Sorry, I cannot respond to this.", nil
	}

	return response, nil
}

func (g *GeminiProvider) GetName() string {
	return "Gemini"
}

func (g *GeminiProvider) GetAvailableModels() []string {
	return []string{
		"gemini-3-pro",
		"gemini-2.5-flash",
		"gemini-2.5-pro",
	}
}
