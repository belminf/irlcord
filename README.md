# IRLCord Discord Bot

IRLCord is a Discord bot designed to help communities organize and manage real-life events and groups. It provides features for creating and managing groups, scheduling events, handling RSVPs, and more.

## Features

- **Group Management**: Create, join, and manage groups with customizable settings
- **Event Management**: Schedule events, manage attendees, and handle RSVPs
- **Customizable Terminology**: Customize the terminology used by the bot (e.g., "Group" vs "Circle")
- **Role-Based Permissions**: Different permissions for group leaders and members

## Installation

### Prerequisites

- Go 1.21 or higher
- SQLite3

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/azlyth/irlcord.git
   cd irlcord
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the bot:
   ```
   go build
   ```

4. Run the bot:
   ```
   ./irlcord
   ```

   On first run, a default `config.json` file will be created. Edit this file to add your Discord bot token.

## Configuration

The bot is configured using a JSON file. By default, it looks for `config.json` in the current directory, but you can specify a different path using the `-config` flag.

Example configuration:

```json
{
  "discord_token": "YOUR_DISCORD_BOT_TOKEN",
  "database_path": "irlcord.db",
  "prefix": "!",
  "admin_ids": ["YOUR_DISCORD_USER_ID"],
  "guild_id": "YOUR_DISCORD_SERVER_ID",
  "terminology": {
    "group_singular": "Group",
    "group_plural": "Groups",
    "event_singular": "Event",
    "event_plural": "Events"
  },
  "channels": {
    "log_channel": "",
    "admin_channel": "",
    "events_channel": ""
  }
}
```

## Usage

### Basic Commands

- `!help` - Show help message
- `!group create name="Group Name" description="Group Description"` - Create a new group
- `!group list` - List all groups
- `!group info id=1` - Show information about a group
- `!group join id=1` - Join a group
- `!group leave id=1` - Leave a group
- `!event create group=1 name="Event Name" date="2023-01-01" time="18:00" location="Location Name" address="Location Address" description="Event Description"` - Create a new event
- `!event list` - List upcoming events
- `!event info id=1` - Show information about an event
- `!rsvp yes id=1` - RSVP yes to an event
- `!rsvp no id=1` - RSVP no to an event
- `!settings group id=1 open=true` - Update group settings
- `!settings terminology group="Crew" event="Hangout"` - Update terminology

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API library for Go
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite3 driver for Go 