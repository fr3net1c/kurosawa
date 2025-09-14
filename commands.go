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

func RegisterCommands(router *cmdroute.Router, s *state.State) {
	botState = s
	router.AddFunc("ping", pingCommand)
	router.AddFunc("clear", clearCommand)
}

func pingCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
	}
}

func clearCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
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
