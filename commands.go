package main

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

var botState *state.State

func RegisterCommands(router *cmdroute.Router, s *state.State, dbManager *DatabaseManager) {
	botState = s
	router.AddFunc("ping", pingCommand)
	router.AddFunc("clear", clearCommand)
	router.AddFunc("ticket", ticketCommand)
	router.AddFunc("ai", aiCommand)
	router.AddFunc("deletedata", deleteDataCommand)
	router.AddFunc("clearhistory", clearHistoryCommand)
	RegisterAICommands(router, dbManager, mlService)
}

func pingCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
	}
}

func aiCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	go createAIChannel(botState, data.Event)
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Your private AI channel is being created..."),
		Flags:   discord.EphemeralMessage,
	}
}

func clearHistoryCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	userID := data.Event.SenderID()
	err := dbManager.ClearUserHistory(userID.String())
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString(fmt.Sprintf("Error clearing your history: %s", err.Error())),
			Flags:   discord.EphemeralMessage,
		}
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Your conversation history has been successfully cleared."),
		Flags:   discord.EphemeralMessage,
	}
}

func deleteDataCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	userID := data.Event.SenderID()
	err := dbManager.DeleteUserDB(userID.String())
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString(fmt.Sprintf("Error deleting your data: %s", err.Error())),
			Flags:   discord.EphemeralMessage,
		}
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Your data has been successfully deleted."),
		Flags:   discord.EphemeralMessage,
	}
}

func clearCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	p, err := botState.Permissions(data.Event.ChannelID, data.Event.SenderID())
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error checking permissions."),
			Flags:   discord.EphemeralMessage,
		}
	}

	if !p.Has(discord.PermissionManageMessages) && !p.Has(discord.PermissionAdministrator) {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("You don't have permission to use this command."),
			Flags:   discord.EphemeralMessage,
		}
	}

	channelID := data.Event.ChannelID
	go func() {
		totalDeleted := 0
		for i := 0; i < 5; i++ {
			messages, err := botState.Messages(channelID, 100)
			if err != nil {
				botState.SendMessage(channelID, fmt.Sprintf("Error getting messages: %s", err.Error()))
				return
			}
			if len(messages) == 0 {
				break
			}
			var ids []discord.MessageID
			for _, m := range messages {
				if m.ID.Time().After(time.Now().Add(-14 * 24 * time.Hour)) {
					ids = append(ids, m.ID)
				}
			}
			if len(ids) == 0 {
				break
			}
			if len(ids) == 1 {
				err = botState.DeleteMessage(channelID, ids[0], "Cleared by bot")
			} else {
				err = botState.DeleteMessages(channelID, ids, "Bulk delete by bot")
			}
			if err != nil {
				botState.SendMessage(channelID, fmt.Sprintf("Error deleting messages: %s", err.Error()))
				return
			}
			totalDeleted += len(ids)
			time.Sleep(1 * time.Second)
		}
		if totalDeleted > 0 {
			botState.SendMessage(channelID, fmt.Sprintf("Successfully deleted %d messages", totalDeleted))
		} else {
			botState.SendMessage(channelID, "ℹNo messages to delete")
		}
	}()
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Starting to delete messages..."),
	}
}
