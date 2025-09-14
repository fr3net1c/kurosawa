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
	}

	_, err = bot.BulkOverwriteGuildCommands(app.ID, guildID, commands)
	return err
}
