# IRLCord

A Discord bot for managing IRL events within different groups/circles. The bot helps communities organize and manage in-person events, handle RSVPs, and maintain group-specific settings.

## Features

- **Group Management**
  - Create and manage groups/circles
  - Control group visibility and access
  - Assign group leaders
  - Customize group settings

- **Event Management**
  - Create and schedule events
  - Handle RSVPs and waitlists
  - Set attendance limits
  - Track event details (location, time, description)
  - Event approval workflow

- **Member Management**
  - Track member participation
  - Manage user profiles
  - Handle dietary restrictions and preferences

## Setup

### Using Poetry (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/irlcord.git
   cd irlcord
   ```

2. Install Poetry and dependencies:
   ```bash
   make setup
   ```
   
   Or manually:
   ```bash
   curl -sSL https://install.python-poetry.org | python3 -
   poetry install
   ```

3. Create a configuration file:
   ```bash
   make config
   ```
   
   Then edit the `config.yaml` file with your Discord bot token and admin user ID.

4. Initialize the database:
   ```bash
   make init-db
   ```

5. Run the bot:
   ```bash
   make run
   ```

### Using Docker

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/irlcord.git
   cd irlcord
   ```

2. Create a configuration file:
   ```bash
   make config
   ```
   
   Then edit the `config.yaml` file with your Discord bot token and admin user ID.

3. Build and run the Docker container:
   ```bash
   make docker-build
   make docker-run
   ```

4. View logs:
   ```bash
   make docker-logs
   ```

5. Stop the container:
   ```bash
   make docker-stop
   ```

### Using pip (Alternative)

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/irlcord.git
   cd irlcord
   ```

2. Create a virtual environment and install dependencies:
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install -r requirements.txt
   ```

3. Create a `config.yaml` file in the root directory (see example below).

4. Run the bot:
   ```bash
   python -m irlcord.main
   ```

## Makefile Commands

The project includes a Makefile with several useful commands:

- `make setup` - Install Poetry and project dependencies
- `make install` - Install project dependencies
- `make update` - Update project dependencies
- `make run` - Run the Discord bot
- `make lint` - Run linters (flake8)
- `make format` - Format code (black, isort)
- `make test` - Run tests
- `make clean` - Clean up temporary files
- `make init-db` - Initialize the database with schema
- `make config` - Create a sample config file if it doesn't exist
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-stop` - Stop Docker container
- `make docker-logs` - View Docker container logs
- `make help` - Show available commands

## Configuration

Create a `config.yaml` file in the root directory:

```yaml
general:
  bot_token: "YOUR_BOT_TOKEN_HERE"
  admin_user_ids: ["YOUR_DISCORD_USER_ID"]

terminology:
  group_plural: "Circles"
  group_singular: "Circle"
  member_plural: "Folks"
  member_singular: "Person"
  leader_plural: "Leaders"
  leader_singular: "Leader"
  event_plural: "Events"
  event_singular: "Event"
  contributor_plural: "Adventurers"
  contributor_singular: "Adventurer"

commands:
  # Group Management
  group_create: "circle new"
  group_join: "circle join"
  group_leave: "circle leave"
  group_info: "circle info"
  group_modify: "circle modify"

  # Event Management
  event_create: "event new"
  event_modify: "event modify"
  event_confirm: "event confirm"
  event_unconfirm: "event unconfirm"
  event_waitlist: "event waitlist"
  event_info: "event info"
  event_change_host: "event change host"
```

## Usage

### Group Commands

- Create a new group:
  ```
  circle new name="Group Name" description="Group description"
  ```

- Join a group:
  ```
  circle join name="Group Name"
  ```

- Leave a group:
  ```
  circle leave
  ```

- View group info:
  ```
  circle info
  ```

- Modify group settings:
  ```
  circle modify is_open=false new_members_can_create_events=false
  ```

### Event Commands

- Create a new event:
  ```
  event new name="Event Name" date="2024-03-20" time="19:00" location="Location Name" address="123 Main St" description="Event description" max=10
  ```

- Modify an event:
  ```
  event modify description="New description" date="2024-03-21" time="20:00"
  ```

- Confirm attendance:
  ```
  event confirm
  ```

- Cancel attendance:
  ```
  event unconfirm
  ```

- Join waitlist:
  ```
  event waitlist
  ```

- View event info:
  ```
  event info
  ```

- Change event host:
  ```
  event change host user=@username
  ```

## Development

### Project Structure

```
irlcord/
├── cogs/               # Command modules
│   ├── events.py       # Event management commands
│   └── groups.py       # Group management commands
├── utils/              # Utility modules
│   ├── db.py           # Database operations
│   └── discord_helpers.py  # Discord-specific helpers
└── main.py             # Bot initialization
```

### Adding New Features

1. Create a new cog in the `irlcord/cogs/` directory
2. Implement your commands using the Discord.py framework
3. Register your cog in `irlcord/main.py`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 