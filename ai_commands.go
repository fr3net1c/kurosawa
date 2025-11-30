package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// aiCommandHandler encapsulates dependencies for AI commands
type aiCommandHandler struct {
	mlService *MLService
	dbManager *DatabaseManager
}

// RegisterAICommands registers commands for working with providers and models
func RegisterAICommands(router *cmdroute.Router, dbMgr *DatabaseManager, mlSvc *MLService) {
	handler := &aiCommandHandler{
		mlService: mlSvc,
		dbManager: dbMgr,
	}

	router.AddFunc("provider", handler.providerCommand)
	router.AddFunc("model", handler.modelCommand)
	router.AddFunc("aiconfig", handler.aiConfigCommand)
}

// providerCommand allows user to view and select a provider
func (h *aiCommandHandler) providerCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	userID := data.Event.SenderID()

	// Get provider name from arguments (if provided)
	var selectedProvider string
	if len(data.Options) > 0 && data.Options[0].Name == "name" {
		selectedProvider = strings.ToLower(data.Options[0].String())
	}

	// If no name provided - show list of available providers
	if selectedProvider == "" {
		return h.listProvidersResponse()
	}

	// Try to set the selected provider
	return h.setProviderResponse(userID.String(), selectedProvider)
}

// modelCommand allows user to view and select a model
func (h *aiCommandHandler) modelCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	userID := data.Event.SenderID()

	// Get user's current provider
	userDB, err := h.dbManager.GetUserDB(userID.String())
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot access database: %v", err))
	}

	providerName, _, err := userDB.GetUserPreference(userID.String())
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot get preferences: %v", err))
	}

	// Check that provider is selected
	if providerName == "none" || providerName == "" {
		return h.errorResponse("First select a provider using `/provider name:<provider>`")
	}

	// Get model name from arguments (if provided)
	var selectedModel string
	if len(data.Options) > 0 && data.Options[0].Name == "name" {
		selectedModel = data.Options[0].String()
	}

	// If no name provided - show list of models for current provider
	if selectedModel == "" {
		return h.listModelsResponse(providerName)
	}

	// Try to set the selected model
	return h.setModelResponse(userID.String(), providerName, selectedModel)
}

// aiConfigCommand shows user's current AI configuration
func (h *aiCommandHandler) aiConfigCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	userID := data.Event.SenderID()

	userDB, err := h.dbManager.GetUserDB(userID.String())
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot access database: %v", err))
	}

	providerName, modelName, err := userDB.GetUserPreference(userID.String())
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot get preferences: %v", err))
	}

	// Build message with current configuration
	response := "**Your AI Configuration:**\n"

	if providerName == "none" || providerName == "" {
		response += "Provider: Not selected\n"
	} else {
		response += fmt.Sprintf("Provider: %s\n", providerName)
	}

	if modelName == "none" || modelName == "" {
		response += "Model: Not selected\n"
	} else {
		response += fmt.Sprintf("Model: %s\n", modelName)
	}

	response += "\n**Available commands:**\n"
	response += "• `/provider` - View and set your AI provider\n"
	response += "• `/model` - View and set your AI model\n"

	return &api.InteractionResponseData{
		Content: option.NewNullableString(response),
		Flags:   discord.EphemeralMessage,
	}
}

// === HELPER METHODS ===

// listProvidersResponse returns a list of available providers
func (h *aiCommandHandler) listProvidersResponse() *api.InteractionResponseData {
	providers := h.mlService.GetAvailableProviders()
	if len(providers) == 0 {
		return h.errorResponse("No AI providers are configured")
	}

	response := "**Available AI Providers:**\n"
	for _, p := range providers {
		response += fmt.Sprintf("• %s\n", p)
	}
	response += "\nUsage: `/provider name:<provider>`"

	return &api.InteractionResponseData{
		Content: option.NewNullableString(response),
		Flags:   discord.EphemeralMessage,
	}
}

// setProviderResponse sets a provider for the user
func (h *aiCommandHandler) setProviderResponse(userID, providerName string) *api.InteractionResponseData {
	// Check that provider exists
	provider := h.mlService.GetProvider(providerName)
	if provider == nil {
		availableProviders := h.mlService.GetAvailableProviders()
		return h.errorResponse(
			fmt.Sprintf("Provider '%s' not found. Available: %s",
				providerName,
				strings.Join(availableProviders, ", ")),
		)
	}

	// Save user's choice
	userDB, err := h.dbManager.GetUserDB(userID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot access database: %v", err))
	}

	err = userDB.SetUserPreference(userID, providerName, "none")
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot save preference: %v", err))
	}

	// Show available models for this provider
	models := provider.GetAvailableModels()
	modelsStr := strings.Join(models, ", ")

	response := fmt.Sprintf(
		"Selected provider: **%s**\n\n"+
			"**Available models:**\n%s\n\n"+
			"Next, use `/model name:<model>` to choose a model.",
		providerName,
		modelsStr,
	)

	return &api.InteractionResponseData{
		Content: option.NewNullableString(response),
		Flags:   discord.EphemeralMessage,
	}
}

// listModelsResponse returns a list of models for the current provider
func (h *aiCommandHandler) listModelsResponse(providerName string) *api.InteractionResponseData {
	provider := h.mlService.GetProvider(providerName)
	if provider == nil {
		return h.errorResponse(fmt.Sprintf("Provider '%s' not found", providerName))
	}

	models := provider.GetAvailableModels()
	response := fmt.Sprintf("**Available models for %s:**\n", providerName)
	for _, m := range models {
		response += fmt.Sprintf("• %s\n", m)
	}
	response += "\nUsage: `/model name:<model>`"

	return &api.InteractionResponseData{
		Content: option.NewNullableString(response),
		Flags:   discord.EphemeralMessage,
	}
}

// setModelResponse sets a model for the user
func (h *aiCommandHandler) setModelResponse(userID, providerName, modelName string) *api.InteractionResponseData {
	userDB, err := h.dbManager.GetUserDB(userID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot access database: %v", err))
	}

	err = userDB.SetUserPreference(userID, providerName, modelName)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Cannot save preference: %v", err))
	}

	response := fmt.Sprintf("Model set to **%s** for **%s**", modelName, providerName)
	return &api.InteractionResponseData{
		Content: option.NewNullableString(response),
		Flags:   discord.EphemeralMessage,
	}
}

// errorResponse returns a standard error message
func (h *aiCommandHandler) errorResponse(message string) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Error: " + message),
		Flags:   discord.EphemeralMessage,
	}
}
