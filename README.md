# Kurosawa Bot

A Discord bot powered by Google's Gemini AI, written in Go.

## Features

*   **AI Chat:** Interact with a powerful language model directly in your Discord server.
*   **Slash Commands:** Easy-to-use commands for interacting with the bot.
*   **Conversation History:** The bot remembers the context of your conversations.

## Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/kurosawa.git
    cd kurosawa
    ```

2.  **Create a `.env` file** in the root of the project with the following content:
    ```env
    DISCORD_TOKEN=your_discord_bot_token
    AI_TOKEN=your_gemini_api_key
    CHANNEL_ID=your_discord_channel_id
    GUILD_ID=your_discord_server_id
    ```

3.  **Run the bot:**
    ```bash
    go run .
    ```

## Usage

*   Use the `/ping` command to check if the bot is running.
*   Use the `/clear` command to delete the last 500 messages in a channel.
*   In the designated AI channel, start your message with `!m` to chat with the bot (e.g., `!m Hello, how are you?`).
