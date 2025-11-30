package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

const (
	// These should be set in your .env file
	ticketCategoryIDEnv = "TICKET_CATEGORY_ID"
	moderatorRoleIDEnv  = "MODERATOR_ROLE_ID"
)

var (
	ticketCategoryID discord.ChannelID
	moderatorRoleID  discord.RoleID
)

func initTicketSystem() error {
	categoryIDStr := os.Getenv(ticketCategoryIDEnv)
	if categoryIDStr == "" {
		return fmt.Errorf("%s not set in .env file", ticketCategoryIDEnv)
	}
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", ticketCategoryIDEnv, err)
	}
	ticketCategoryID = discord.ChannelID(categoryID)

	modRoleIDStr := os.Getenv(moderatorRoleIDEnv)
	if modRoleIDStr == "" {
		return fmt.Errorf("%s not set in .env file", moderatorRoleIDEnv)
	}
	modRoleID, err := strconv.ParseUint(modRoleIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", moderatorRoleIDEnv, err)
	}
	moderatorRoleID = discord.RoleID(modRoleID)

	return nil
}

func ticketCommand(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	p, err := botState.Permissions(data.Event.ChannelID, data.Event.SenderID())
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error checking permissions."),
			Flags:   discord.EphemeralMessage,
		}
	}

	if !p.Has(discord.PermissionManageChannels) {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("You don't have permission to use this command."),
			Flags:   discord.EphemeralMessage,
		}
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Click the button below to create a new ticket."),
		Components: &discord.ContainerComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "Create Ticket",
					Style:    discord.PrimaryButtonStyle(),
					CustomID: "create_ticket",
				},
			},
		},
	}
}

func createTicketChannel(s *state.State, e *discord.InteractionEvent) {
	overwrites := []discord.Overwrite{
		{
			ID:   discord.Snowflake(e.GuildID),
			Type: discord.OverwriteRole,
			Deny: discord.PermissionViewChannel,
		},
		{
			ID:    discord.Snowflake(e.Member.User.ID),
			Type:  discord.OverwriteMember,
			Allow: discord.PermissionViewChannel,
		},
		{
			ID:    discord.Snowflake(moderatorRoleID),
			Type:  discord.OverwriteRole,
			Allow: discord.PermissionViewChannel,
		},
	}

	err := s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags: discord.EphemeralMessage,
		},
	})
	if err != nil {
		log.Printf("Error deferring interaction: %v", err)
		return
	}

	ch, err := s.CreateChannel(e.GuildID, api.CreateChannelData{
		Name:       fmt.Sprintf("ticket-%s", e.Member.User.Username),
		Type:       discord.GuildText,
		CategoryID: ticketCategoryID,
		Overwrites: overwrites,
	})
	if err != nil {
		log.Printf("Error creating ticket channel: %v", err)
		s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
			Content: option.NewNullableString("Failed to create ticket channel."),
		})
		return
	}

	_, err = s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
		Content: option.NewNullableString(fmt.Sprintf("Ticket channel created: %s", ch.Mention())),
	})
	if err != nil {
		log.Printf("Error editing interaction response: %v", err)
	}

	_, err = s.SendMessageComplex(ch.ID, api.SendMessageData{
		Content: fmt.Sprintf("Welcome %s! A moderator will be with you shortly.", e.Member.User.Mention()),
		Components: discord.ContainerComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "Close Ticket",
					Style:    discord.DangerButtonStyle(),
					CustomID: "close_ticket",
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error sending message to ticket channel: %v", err)
	}
}

func createAIChannel(s *state.State, e *discord.InteractionEvent) {
	overwrites := []discord.Overwrite{
		{
			ID:   discord.Snowflake(e.GuildID),
			Type: discord.OverwriteRole,
			Deny: discord.PermissionViewChannel,
		},
		{
			ID:    discord.Snowflake(e.Member.User.ID),
			Type:  discord.OverwriteMember,
			Allow: discord.PermissionViewChannel,
		},
		{
			ID:    discord.Snowflake(moderatorRoleID),
			Type:  discord.OverwriteRole,
			Allow: discord.PermissionViewChannel,
		},
	}

	ch, err := s.CreateChannel(e.GuildID, api.CreateChannelData{
		Name:       fmt.Sprintf("ai-%s", e.Member.User.Username),
		Type:       discord.GuildText,
		CategoryID: ticketCategoryID,
		Overwrites: overwrites,
	})
	if err != nil {
		log.Printf("Error creating AI channel: %v", err)
		return
	}

	_, err = s.SendMessage(ch.ID, fmt.Sprintf("Welcome %s! You can start your private conversation with the AI here.", e.Member.User.Mention()))
	if err != nil {
		log.Printf("Error sending message to AI channel: %v", err)
	}
}

func closeTicketChannel(s *state.State, e *discord.InteractionEvent) {
	isModerator := false
	for _, roleID := range e.Member.RoleIDs {
		if roleID == moderatorRoleID {
			isModerator = true
			break
		}
	}

	if !isModerator {
		s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("You do not have permission to close this ticket."),
				Flags:   discord.EphemeralMessage,
			},
		})
		return
	}

	err := s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
	})
	if err != nil {
		log.Printf("failed to acknowledge interaction: %v", err)
		return
	}

	err = s.DeleteChannel(e.ChannelID, "Ticket closed by moderator")
	if err != nil {
		log.Printf("Error deleting ticket channel: %v", err)
		s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
			Content: option.NewNullableString("Failed to close ticket."),
		})
	}
}
