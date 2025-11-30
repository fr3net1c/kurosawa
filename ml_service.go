package main

import (
	"context"
	"fmt"
)

// MLService routes requests to the correct AI provider and manages conversation history
type MLService struct {
	providers map[string]AIProvider
	dbManager *DatabaseManager
}

type Message struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Time     string `json:"time"`
	UserName string `json:"user_name"`
}

// NewMLService creates a new instance of MLService
func NewMLService(dbManager *DatabaseManager, providers map[string]AIProvider) (*MLService, error) {
	return &MLService{
		providers: providers,
		dbManager: dbManager,
	}, nil
}

// GetResponse processes a user message:
// 1. Saves the message to history
// 2. Retrieves conversation history
// 3. Checks which provider the user selected
// 4. Sends request to the provider
// 5. Saves the response to history
func (ml *MLService) GetResponse(userID, userName, message string) (string, error) {
	db, err := ml.dbManager.GetUserDB(userID)
	if err != nil {
		return "", fmt.Errorf("could not get user DB: %w", err)
	}

	// Save incoming user message
	if err := db.AddMessage(userID, userName, "user", message); err != nil {
		return "", fmt.Errorf("failed to save user message: %v", err)
	}

	// Load full conversation history for context
	history, err := db.GetMessages()
	if err != nil {
		return "", fmt.Errorf("failed to load conversation history: %v", err)
	}

	// Build full prompt with conversation history
	prompt := ml.buildPrompt(history)

	// Get user settings (which provider they selected)
	providerName, _, err := db.GetUserPreference(userID)
	if err != nil {
		return "", fmt.Errorf("could not get user preferences: %w", err)
	}

	// Check that provider is selected - no fallback to default
	if providerName == "none" || providerName == "" {
		return "Please select an AI provider first using /provider name:<provider>", nil
	}

	// Get provider instance
	provider, exists := ml.providers[providerName]
	if !exists {
		return fmt.Sprintf("Provider '%s' not found. Available providers: %s",
			providerName,
			ml.getAvailableProvidersStr()), nil
	}

	ctx := context.Background()
	response, err := provider.GetResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get response from %s: %v", providerName, err)
	}

	if response == "" {
		return "Sorry, I cannot respond to this.", nil
	}

	// Save assistant response to history
	db, err = ml.dbManager.GetUserDB(userID)
	if err != nil {
		fmt.Printf("Warning: could not get user DB to save assistant message: %v\n", err)
	} else {
		if err := db.AddMessage(userID, "Kurosawa", "assistant", response); err != nil {
			fmt.Printf("Warning: failed to save assistant message: %v\n", err)
		}
	}

	return response, nil
}

// buildPrompt constructs the full prompt from system prompt and conversation history
func (ml *MLService) buildPrompt(messages []Message) string {
	prompt := SystemPrompt + "\n\nConversation history:\n"

	for _, msg := range messages {
		if msg.Role == "user" {
			prompt += fmt.Sprintf("%s: %s\n", msg.UserName, msg.Content)
		} else {
			prompt += fmt.Sprintf("%s: %s\n", msg.UserName, msg.Content)
		}
	}

	prompt += "\n"
	return prompt
}

// GetProvider returns a provider by name
func (ml *MLService) GetProvider(name string) AIProvider {
	return ml.providers[name]
}

// GetAvailableProviders returns a list of available provider names
func (ml *MLService) GetAvailableProviders() []string {
	var providers []string
	for name := range ml.providers {
		providers = append(providers, name)
	}
	return providers
}

// getAvailableProvidersStr returns providers as a string for error messages
func (ml *MLService) getAvailableProvidersStr() string {
	providers := ml.GetAvailableProviders()
	if len(providers) == 0 {
		return "none"
	}

	result := ""
	for i, p := range providers {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
