package main

import "context"

type AIProvider interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
	GetName() string
	GetAvailableModels() []string
}
