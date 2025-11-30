package main

import (
	"fmt"
	"os"
)

type providerConfig struct {
	name        string
	apiKeyEnv   string
	modelEnvKey string
	constructor func(apiKey, model string) (AIProvider, error)
}

type ProviderFactory struct {
	providers map[string]AIProvider
}

func NewProviderFactory() (*ProviderFactory, error) {
	factory := &ProviderFactory{
		providers: make(map[string]AIProvider),
	}

	configs := []providerConfig{
		{
			name:        "gemini",
			apiKeyEnv:   "GEMINI_API_KEY",
			modelEnvKey: "",
			constructor: func(apiKey, _ string) (AIProvider, error) {
				return NewGeminiProvider(apiKey)
			},
		},
		{
			name:        "openai",
			apiKeyEnv:   "OPENAI_API_KEY",
			modelEnvKey: "OPENAI_DEFAULT_MODEL",
			constructor: func(apiKey, model string) (AIProvider, error) {
				provider, err := NewOpenAIProvider(apiKey, model)
				return AIProvider(provider), err
			},
		},
		{
			name:        "mistral",
			apiKeyEnv:   "MISTRAL_API_KEY",
			modelEnvKey: "MISTRAL_DEFAULT_MODEL",
			constructor: func(apiKey, model string) (AIProvider, error) {
				provider, err := NewMistralProvider(apiKey, model)
				return AIProvider(provider), err
			},
		},
		{
			name:        "openrouter",
			apiKeyEnv:   "OPENROUTER_API_KEY",
			modelEnvKey: "OPENROUTER_DEFAULT_MODEL",
			constructor: func(apiKey, model string) (AIProvider, error) {
				provider, err := NewOpenRouterProvider(apiKey, model)
				return AIProvider(provider), err
			},
		},
	}

	for _, cfg := range configs {
		if err := factory.registerProvider(cfg); err != nil {
			fmt.Printf("Warning: %s: %v\n", cfg.name, err)
		}
	}

	if len(factory.providers) == 0 {
		return nil, fmt.Errorf(
			"No AI providers configured. Please set at least ONE API key in .env file: " +
				"GEMINI_API_KEY, OPENAI_API_KEY, MISTRAL_API_KEY, or OPENROUTER_API_KEY",
		)
	}

	return factory, nil
}

func (f *ProviderFactory) registerProvider(cfg providerConfig) error {
	apiKey := os.Getenv(cfg.apiKeyEnv)
	if apiKey == "" {
		return fmt.Errorf("API key not set (%s)", cfg.apiKeyEnv)
	}

	defaultModel := os.Getenv(cfg.modelEnvKey)

	provider, err := cfg.constructor(apiKey, defaultModel)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	f.providers[cfg.name] = provider
	fmt.Printf("Loaded %s provider\n", cfg.name)
	return nil
}

func (f *ProviderFactory) GetProviders() map[string]AIProvider {
	return f.providers
}

func (f *ProviderFactory) GetProvider(name string) AIProvider {
	return f.providers[name]
}

func (f *ProviderFactory) GetAvailableProviders() []string {
	providers := make([]string, 0, len(f.providers))
	for name := range f.providers {
		providers = append(providers, name)
	}
	return providers
}
