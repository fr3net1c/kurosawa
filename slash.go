package main

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

func RegisterSlashCommands(bot *state.State, guildID discord.GuildID) error {
	app, err := bot.CurrentApplication()
	if err != nil {
		return err
	}

	commands := []api.CreateCommandData{
		{
			Name:        "ping",
			Description: "Check if the bot is working",
		},
		{
			Name:        "clear",
			Description: "Delete 500 messages in the channel",
		},
		{
			Name:        "ticket",
			Description: "Create a new support ticket",
		},
		{
			Name:        "ai",
			Description: "Start a private conversation with the AI",
		},
		{
			Name:        "deletedata",
			Description: "Delete all your data from the bot's database",
		},
		{
			Name:        "clearhistory",
			Description: "Clear your conversation history with the AI",
		},
		{
			Name:        "provider",
			Description: "Select or view your AI provider",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "name",
					Description: "Provider name (gemini, openai, mistral, openrouter)",
					Required:    false,
				},
			},
		},
		{
			Name:        "model",
			Description: "Select or view your AI model",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "name",
					Description: "Model name",
					Required:    false,
				},
			},
		},
		{
			Name:        "aiconfig",
			Description: "View your current AI configuration",
		},
	}

	_, err = bot.BulkOverwriteGuildCommands(app.ID, guildID, commands)
	return err
}
