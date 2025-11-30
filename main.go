package main

import (
	"context"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/joho/godotenv"
)

var (
	mlService *MLService
	dbManager *DatabaseManager
	channelID discord.ChannelID
)

const (
	MaxMessageLength = 2000
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("No DISCORD_TOKEN provided in .env file")
	}

	channelIDStr := os.Getenv("CHANNEL_ID")
	if channelIDStr == "" {
		log.Fatal("No CHANNEL_ID provided in .env file")
	}

	channelIDInt, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid CHANNEL_ID format")
	}

	channelID = discord.ChannelID(channelIDInt)

	dbManager, err = NewDatabaseManager("user_data")
	if err != nil {
		log.Fatal("Cannot initialize database manager:", err)
	}
	defer dbManager.CloseAll()

	providerFactory, err := NewProviderFactory()
	if err != nil {
		log.Fatal("Cannot initialize provider factory:", err)
	}

	mlService, err = NewMLService(dbManager, providerFactory.GetProviders())
	if err != nil {
		log.Fatal("Cannot initialize ML service:", err)
	}

	bot := state.New("Bot " + token)
	bot.AddIntents(gateway.IntentGuildMessages | gateway.IntentMessageContent | gateway.IntentGuilds)

	bot.AddHandler(func(m *gateway.MessageCreateEvent) {
		if m.Author.Bot {
			return
		}

		ch, err := bot.Channel(m.ChannelID)
		if err != nil {
			log.Printf("Error getting channel: %v", err)
			return
		}

		//nolint:SA4006
		isAIChannel := strings.HasPrefix(ch.Name, "ai-")

		if isAIChannel || (m.ChannelID == channelID && strings.HasPrefix(m.Content, "!m ")) {
			handleAIMessage(bot, m, isAIChannel)
		}
	})

	bot.AddInteractionHandler(&InteractionHandler{bot: bot})

	router := cmdroute.NewRouter()
	RegisterCommands(router, bot, dbManager)
	bot.AddInteractionHandler(router)

	if err := bot.Open(context.Background()); err != nil {
		log.Fatal("Cannot open:", err)
	}
	defer bot.Close()

	guildIDStr := os.Getenv("GUILD_ID")
	if guildIDStr == "" {
		log.Fatal("No GUILD_ID provided in .env file")
	}

	guildIDInt, err := strconv.ParseUint(guildIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid GUILD_ID format")
	}

	guildID := discord.GuildID(guildIDInt)

	if err := initTicketSystem(); err != nil {
		log.Fatal("Cannot initialize ticket system:", err)
	}

	if err := RegisterSlashCommands(bot, guildID); err != nil {
		log.Fatal("Cannot register slash commands:", err)
	}

	log.Println("Bot is running! Press CTRL+C to exit.")
	log.Println("AI channel ID:", channelID)
	log.Println("Slash commands registered for guild:", guildID)

	select {}
}

type InteractionHandler struct {
	bot *state.State
}

func (h *InteractionHandler) HandleInteraction(e *discord.InteractionEvent) *api.InteractionResponse {
	switch data := e.Data.(type) {
	case *discord.ButtonInteraction:
		switch data.CustomID {
		case "create_ticket":
			createTicketChannel(h.bot, e)
		case "close_ticket":
			closeTicketChannel(h.bot, e)
		}
	}
	return nil
}

func handleAIMessage(bot *state.State, m *gateway.MessageCreateEvent, isPrivate bool) {
	var message string
	if isPrivate {
		message = m.Content
	} else {
		message = strings.TrimPrefix(m.Content, "!m ")
	}
	message = strings.TrimSpace(message)

	if message == "" {
		return
	}

	urls := findURLs(message)
	if len(urls) > 0 {
		for _, url := range urls {
			content, err := GetContentFromURL(url)
			if err != nil {
				log.Printf("Error getting content from URL: %v", err)
				continue
			}
			message = content + "\n" + message
		}
	}

	userName := m.Author.Username
	if m.Member != nil && m.Member.Nick != "" {
		userName = m.Member.Nick
	}

	response, err := mlService.GetResponse(m.Author.ID.String(), userName, message)
	if err != nil {
		log.Printf("Error getting AI response: %v", err)
		bot.SendMessage(m.ChannelID, "An error occurred while contacting the AI.")
		return
	}

	err = sendLongMessage(bot, m.ChannelID, response)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func findURLs(text string) []string {
	r := regexp.MustCompile(`\bhttps?://\S+\b`)
	return r.FindAllString(text, -1)
}

func sendLongMessage(bot *state.State, channelID discord.ChannelID, message string) error {
	if len(message) <= MaxMessageLength {
		_, err := bot.SendMessage(channelID, message)
		return err
	}

	parts := splitMessage(message, MaxMessageLength)

	for i, part := range parts {
		messageToSend := part
		if i < len(parts)-1 {
			messageToSend += "\n\n"
		}

		_, err := bot.SendMessage(channelID, messageToSend)
		if err != nil {
			return err
		}
	}

	return nil
}

func splitMessage(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var parts []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= maxLength {
			parts = append(parts, remaining)
			break
		}

		splitIndex := findBestSplitPoint(remaining, maxLength)

		part := strings.TrimSpace(remaining[:splitIndex])
		if part != "" {
			parts = append(parts, part)
		}

		remaining = strings.TrimSpace(remaining[splitIndex:])
	}

	return parts
}

func findBestSplitPoint(text string, maxLength int) int {
	if len(text) <= maxLength {
		return len(text)
	}

	separators := []string{"\n\n", ". ", "! ", "? ", "\n", ", ", " "}

	for _, sep := range separators {
		searchText := text[:maxLength]
		if lastIndex := strings.LastIndex(searchText, sep); lastIndex != -1 {
			return lastIndex + len(sep)
		}
	}

	return maxLength
}
