# Kurosawa Bot

A Discord bot for intelligent conversations, written in Go. Kurosawa lets you chat with multiple AI models directly from Discord.

## Features

* Chat with AI models from Discord channels
* Choose from multiple AI providers (Gemini, OpenAI, Mistral, OpenRouter)
* Per-user model selection and preferences
* Full conversation history saved per user
* Slash commands for easy interaction
* Local SQLite storage for data privacy

## Setup

1. Clone the repository:

```bash
git clone https://github.com/your-username/kurosawa.git
cd kurosawa
```

2. Create a `.env` file in the project root:

```env
# Discord Configuration (Required)
DISCORD_TOKEN=your_discord_bot_token
GUILD_ID=your_discord_server_id
CHANNEL_ID=your_discord_channel_id

# AI Models - Add at least ONE

# Google Gemini (Free tier available)
GEMINI_API_KEY=your_gemini_api_key

# OpenAI (Paid)
OPENAI_API_KEY=your_openai_api_key
OPENAI_DEFAULT_MODEL=gpt-5.1

# Mistral (Paid)
MISTRAL_API_KEY=your_mistral_api_key
MISTRAL_DEFAULT_MODEL=mistral-large-latest

# OpenRouter (Paid - aggregates multiple models)
OPENROUTER_API_KEY=your_openrouter_api_key
OPENROUTER_DEFAULT_MODEL=openai/gpt-5.1
```

3. Install dependencies:

```bash
go mod download
```

4. Run the bot:

```bash
go run .
```

## Usage

### Commands

**Provider & Model Selection**
* `/provider` - View available AI providers
* `/provider name:<provider>` - Select a provider (gemini, openai, mistral, openrouter)
* `/model` - View available models for your selected provider
* `/model name:<model>` - Select a specific model
* `/aiconfig` - View your current configuration

**Bot Management**
* `/ping` - Check if bot is running
* `/clear` - Delete last 500 messages in channel
* `/clearhistory` - Clear your conversation history
* `/deletedata` - Delete all your data from the bot

**Chat**
* Send messages in designated AI channels (prefixed with `ai-`)
* Use `/ai` command for private conversations
* Bot responds based on your selected provider and model

### Quick Start

1. Start the bot: `go run .`
2. In Discord, use `/provider` to select an AI provider
3. Use `/model` to choose a model
4. Start chatting!

## Supported AI Models

| Provider | Models | Setup |
|----------|--------|-------|
| Google Gemini | gemini-3-pro, gemini-2.5-flash, gemini-2.5-pro | Set `GEMINI_API_KEY` |
| OpenAI | gpt-5.1, gpt-4o, gpt-4.1, gpt-4 | Set `OPENAI_API_KEY` |
| Mistral | mistral-large-latest, mistral-medium-latest, mistral-small-latest | Set `MISTRAL_API_KEY` |
| OpenRouter | 50+ models (GPT, Gemini, Llama, etc.) | Set `OPENROUTER_API_KEY` |

## Troubleshooting

### "No AI providers configured"

Ensure at least one AI provider API key is set in `.env` file in the project root.

### "Provider not found" or "Model not found"

Use `/provider` and `/model` commands to see available options for your setup.

### "An error occurred while contacting the AI"

This can indicate:

* **Rate limited (HTTP 429)**: High demand. Wait a few minutes and try again.
* **Service unavailable (HTTP 503)**: Provider's service is temporarily down.
* **Invalid API key (HTTP 401/403)**: Check your API key in `.env`.
* **Network timeout**: Check your internet connection.
* **Model overloaded**: Free tier models may have capacity limits.

### "Cannot access database"

Check that `user_data/` directory has read/write permissions.

## Getting API Keys

* **Gemini**: https://ai.google.dev/ (free tier available)
* **OpenAI**: https://platform.openai.com/api-keys (paid)
* **Mistral**: https://console.mistral.ai/ (paid)
* **OpenRouter**: https://openrouter.ai/ (paid)

## License

See LICENSE file for details.
