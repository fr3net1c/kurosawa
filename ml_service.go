package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

type MLService struct {
	client *genai.Client
	db     *DBService
}

type Message struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Time     string `json:"time"`
	UserName string `json:"user_name"`
}

func NewMLService(apiKey string, db *DBService) (*MLService, error) {
	os.Setenv("GEMINI_API_KEY", apiKey)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &MLService{
		client: client,
		db:     db,
	}, nil
}

func (ml *MLService) GetResponse(userID, userName, message string) (string, error) {

	if err := ml.db.AddMessage(userID, userName, "user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %v", err)
	}

	history, err := ml.db.GetMessages(userID)
	if err != nil {
		return "", fmt.Errorf("failed to load user memory: %v", err)
	}

	prompt := ml.buildPrompt(history)

	ctx := context.Background()

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

	for i := 0; i < maxRetries; i++ {
		result, err = ml.client.Models.GenerateContent(
			ctx,
			"gemini-2.5-pro",
			genai.Text(prompt),
			config,
		)
		if err == nil {
			break
		}

		if googleapiErr, ok := err.(*googleapi.Error); ok && googleapiErr.Code == 429 {
			delay := baseDelay * time.Duration(1<<i)
			fmt.Printf("Rate limit exceeded. Retrying in %v...\n", delay)
			time.Sleep(delay)
			continue
		}

		return "", fmt.Errorf("failed to call Gemini API: %v", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API after %d retries: %v", maxRetries, err)
	}

	response := result.Text()
	if response == "" {
		return "Sorry, I cannot respond to this.", nil
	}

	if err := ml.db.AddMessage(userID, "Kurosawa", "assistant", response); err != nil {
		fmt.Printf("Warning: failed to save assistant message: %v\n", err)
	}

	if err := ml.db.TrimHistory(userID, 20); err != nil {
		fmt.Printf("Warning: failed to trim history: %v\n", err)
	}

	return response, nil
}

func (ml *MLService) buildPrompt(messages []Message) string {

	prompt := SystemPrompt + "\n\nConversation history:\n"

	for _, msg := range messages {

		if msg.Role == "user" {

			prompt += fmt.Sprintf("%s: %s\n", msg.UserName, msg.Content)

		} else {

			prompt += fmt.Sprintf("%s: %s\n", msg.UserName, msg.Content)

		}

	}

	prompt += "\nRespond as Kurosawa:"

	return prompt

}
