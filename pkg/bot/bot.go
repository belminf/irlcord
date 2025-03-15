package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/azlyth/irlcord/pkg/config"
	"github.com/azlyth/irlcord/pkg/db"
	"github.com/azlyth/irlcord/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot
type Bot struct {
	Config  *config.Config
	DB      *db.Database
	Session *discordgo.Session
}

// New creates a new bot instance
func New(cfg *config.Config, db *db.Database) (*Bot, error) {
	// Create a new Discord session
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	// Create the bot
	bot := &Bot{
		Config:  cfg,
		DB:      db,
		Session: session,
	}

	// Register handlers
	session.AddHandler(bot.handleReady)
	session.AddHandler(bot.handleMessageCreate)
	session.AddHandler(bot.handleInteractionCreate)

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsDirectMessages

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start() error {
	// Open the websocket connection to Discord
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	return nil
}

// Stop stops the bot
func (b *Bot) Stop() error {
	// Close the websocket connection to Discord
	return b.Session.Close()
}

// handleReady handles the ready event
func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %s#%s", s.State.User.Username, s.State.User.Discriminator)
	log.Printf("Bot is ready and serving %d guilds", len(s.State.Guilds))

	// Set the bot's status
	err := s.UpdateGameStatus(0, "Type !help for commands")
	if err != nil {
		log.Printf("Error setting status: %v", err)
	}
}

// handleMessageCreate handles incoming messages
func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message starts with the command prefix
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// Parse the command
	parts := strings.SplitN(m.Content, " ", 2)
	command := strings.ToLower(strings.TrimPrefix(parts[0], "!"))

	// Handle commands
	switch command {
	case "help":
		b.handleHelpCommand(s, m)
	case "group":
		b.handleGroupCommand(s, m)
	case "event":
		b.handleEventCommand(s, m)
	case "rsvp":
		b.handleRSVPCommand(s, m)
	case "settings":
		b.handleSettingsCommand(s, m)
	}
}

// handleInteractionCreate handles interactions (buttons, select menus, etc.)
func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Handle different interaction types
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		// Handle message components (buttons, select menus)
		b.handleMessageComponent(s, i)
	case discordgo.InteractionApplicationCommand:
		// Handle slash commands
		b.handleApplicationCommand(s, i)
	}
}

// handleMessageComponent handles message components (buttons, select menus)
func (b *Bot) handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the custom ID
	customID := i.MessageComponentData().CustomID

	// Handle different components based on their custom ID
	if strings.HasPrefix(customID, "rsvp_") {
		b.handleRSVPButton(s, i)
	} else if strings.HasPrefix(customID, "group_join_") {
		b.handleGroupJoinButton(s, i)
	} else if strings.HasPrefix(customID, "event_approve_") {
		b.handleEventApprovalButton(s, i)
	}
}

// handleApplicationCommand handles slash commands
func (b *Bot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the command name
	commandName := i.ApplicationCommandData().Name

	// Handle different commands
	switch commandName {
	case "help":
		b.handleHelpSlashCommand(s, i)
	case "group":
		b.handleGroupSlashCommand(s, i)
	case "event":
		b.handleEventSlashCommand(s, i)
	case "rsvp":
		b.handleRSVPSlashCommand(s, i)
	case "settings":
		b.handleSettingsSlashCommand(s, i)
	}
}

// handleHelpCommand handles the !help command
func (b *Bot) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Create help message
	helpMsg := "**IRLCord Bot Commands**\n\n" +
		"**Group Commands**\n" +
		"`!group create name=\"Group Name\" description=\"Group Description\"` - Create a new group\n" +
		"`!group list` - List all groups\n" +
		"`!group info id=1` - Show information about a group\n" +
		"`!group join id=1` - Join a group\n" +
		"`!group leave id=1` - Leave a group\n\n" +
		"**Event Commands**\n" +
		"`!event create group=1 name=\"Event Name\" date=\"2023-01-01\" time=\"18:00\" location=\"Location Name\" address=\"Location Address\" description=\"Event Description\"` - Create a new event\n" +
		"`!event list` - List upcoming events\n" +
		"`!event info id=1` - Show information about an event\n\n" +
		"**RSVP Commands**\n" +
		"`!rsvp yes id=1` - RSVP yes to an event\n" +
		"`!rsvp no id=1` - RSVP no to an event\n\n" +
		"**Settings Commands**\n" +
		"`!settings group id=1 open=true` - Update group settings\n" +
		"`!settings terminology group=\"Crew\" event=\"Hangout\"` - Update terminology"

	// Send the help message
	_, err := discord.SendMessage(s, m.ChannelID, helpMsg)
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

// handleGroupCommand handles the !group command
func (b *Bot) handleGroupCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse the subcommand
	parts := strings.SplitN(m.Content, " ", 3)
	if len(parts) < 2 {
		discord.SendErrorMessage(s, m.ChannelID, "Invalid group command. Use `!help` for usage information.")
		return
	}

	subcommand := strings.ToLower(parts[1])

	// Handle different subcommands
	switch subcommand {
	case "create":
		b.handleGroupCreate(s, m)
	case "list":
		b.handleGroupList(s, m)
	case "info":
		b.handleGroupInfo(s, m)
	case "join":
		b.handleGroupJoin(s, m)
	case "leave":
		b.handleGroupLeave(s, m)
	default:
		discord.SendErrorMessage(s, m.ChannelID, "Invalid group subcommand. Use `!help` for usage information.")
	}
}

// handleEventCommand handles the !event command
func (b *Bot) handleEventCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse the subcommand
	parts := strings.SplitN(m.Content, " ", 3)
	if len(parts) < 2 {
		discord.SendErrorMessage(s, m.ChannelID, "Invalid event command. Use `!help` for usage information.")
		return
	}

	subcommand := strings.ToLower(parts[1])

	// Handle different subcommands
	switch subcommand {
	case "create":
		b.handleEventCreate(s, m)
	case "list":
		b.handleEventList(s, m)
	case "info":
		b.handleEventInfo(s, m)
	default:
		discord.SendErrorMessage(s, m.ChannelID, "Invalid event subcommand. Use `!help` for usage information.")
	}
}

// handleRSVPCommand handles the !rsvp command
func (b *Bot) handleRSVPCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse the subcommand
	parts := strings.SplitN(m.Content, " ", 3)
	if len(parts) < 2 {
		discord.SendErrorMessage(s, m.ChannelID, "Invalid RSVP command. Use `!help` for usage information.")
		return
	}

	subcommand := strings.ToLower(parts[1])

	// Handle different subcommands
	switch subcommand {
	case "yes":
		b.handleRSVPYes(s, m)
	case "no":
		b.handleRSVPNo(s, m)
	default:
		discord.SendErrorMessage(s, m.ChannelID, "Invalid RSVP subcommand. Use `!help` for usage information.")
	}
}

// handleSettingsCommand handles the !settings command
func (b *Bot) handleSettingsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse the subcommand
	parts := strings.SplitN(m.Content, " ", 3)
	if len(parts) < 2 {
		discord.SendErrorMessage(s, m.ChannelID, "Invalid settings command. Use `!help` for usage information.")
		return
	}

	subcommand := strings.ToLower(parts[1])

	// Handle different subcommands
	switch subcommand {
	case "group":
		b.handleSettingsGroup(s, m)
	case "terminology":
		b.handleSettingsTerminology(s, m)
	default:
		discord.SendErrorMessage(s, m.ChannelID, "Invalid settings subcommand. Use `!help` for usage information.")
	}
}

// Placeholder methods for command handlers
func (b *Bot) handleGroupCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group creation
	discord.SendSuccessMessage(s, m.ChannelID, "Group creation not yet implemented")
}

func (b *Bot) handleGroupList(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group listing
	discord.SendSuccessMessage(s, m.ChannelID, "Group listing not yet implemented")
}

func (b *Bot) handleGroupInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group info
	discord.SendSuccessMessage(s, m.ChannelID, "Group info not yet implemented")
}

func (b *Bot) handleGroupJoin(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group joining
	discord.SendSuccessMessage(s, m.ChannelID, "Group joining not yet implemented")
}

func (b *Bot) handleGroupLeave(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group leaving
	discord.SendSuccessMessage(s, m.ChannelID, "Group leaving not yet implemented")
}

func (b *Bot) handleEventCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement event creation
	discord.SendSuccessMessage(s, m.ChannelID, "Event creation not yet implemented")
}

func (b *Bot) handleEventList(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement event listing
	discord.SendSuccessMessage(s, m.ChannelID, "Event listing not yet implemented")
}

func (b *Bot) handleEventInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement event info
	discord.SendSuccessMessage(s, m.ChannelID, "Event info not yet implemented")
}

func (b *Bot) handleRSVPYes(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement RSVP yes
	discord.SendSuccessMessage(s, m.ChannelID, "RSVP yes not yet implemented")
}

func (b *Bot) handleRSVPNo(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement RSVP no
	discord.SendSuccessMessage(s, m.ChannelID, "RSVP no not yet implemented")
}

func (b *Bot) handleSettingsGroup(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement group settings
	discord.SendSuccessMessage(s, m.ChannelID, "Group settings not yet implemented")
}

func (b *Bot) handleSettingsTerminology(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Implement terminology settings
	discord.SendSuccessMessage(s, m.ChannelID, "Terminology settings not yet implemented")
}

func (b *Bot) handleRSVPButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement RSVP button handling
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "RSVP button handling not yet implemented",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) handleGroupJoinButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement group join button handling
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Group join button handling not yet implemented",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) handleEventApprovalButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement event approval button handling
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Event approval button handling not yet implemented",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) handleHelpSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement help slash command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Help slash command not yet implemented",
		},
	})
}

func (b *Bot) handleGroupSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement group slash command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Group slash command not yet implemented",
		},
	})
}

func (b *Bot) handleEventSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement event slash command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Event slash command not yet implemented",
		},
	})
}

func (b *Bot) handleRSVPSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement RSVP slash command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "RSVP slash command not yet implemented",
		},
	})
}

func (b *Bot) handleSettingsSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement settings slash command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Settings slash command not yet implemented",
		},
	})
} 