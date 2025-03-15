package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the bot configuration
type Config struct {
	DiscordToken string `json:"discord_token"`
	DatabasePath string `json:"database_path"`
	Prefix       string `json:"prefix"`
	AdminIDs     []string `json:"admin_ids"`
	GuildID      string `json:"guild_id"`
	Terminology  Terminology `json:"terminology"`
	Channels     Channels `json:"channels"`
	Commands     Commands `json:"commands"`
}

// Terminology represents custom terminology for the bot
type Terminology struct {
	GroupSingular string `json:"group_singular"`
	GroupPlural   string `json:"group_plural"`
	EventSingular string `json:"event_singular"`
	EventPlural   string `json:"event_plural"`
}

// Channels represents channel IDs for the bot
type Channels struct {
	LogChannel    string `json:"log_channel"`
	AdminChannel  string `json:"admin_channel"`
	EventsChannel string `json:"events_channel"`
}

// Commands represents command names for the bot
type Commands struct {
	// Group commands
	GroupCreate string `json:"group_create"`
	GroupJoin   string `json:"group_join"`
	GroupLeave  string `json:"group_leave"`
	GroupInfo   string `json:"group_info"`
	GroupModify string `json:"group_modify"`

	// Event commands
	EventCreate     string `json:"event_create"`
	EventModify     string `json:"event_modify"`
	EventConfirm    string `json:"event_confirm"`
	EventUnconfirm  string `json:"event_unconfirm"`
	EventWaitlist   string `json:"event_waitlist"`
	EventInfo       string `json:"event_info"`
	EventChangeHost string `json:"event_change_host"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DiscordToken: "",
		DatabasePath: "irlcord.db",
		Prefix:       "!",
		AdminIDs:     []string{},
		GuildID:      "",
		Terminology: Terminology{
			GroupSingular: "Group",
			GroupPlural:   "Groups",
			EventSingular: "Event",
			EventPlural:   "Events",
		},
		Channels: Channels{
			LogChannel:    "",
			AdminChannel:  "",
			EventsChannel: "",
		},
		Commands: Commands{
			// Group commands
			GroupCreate: "!group create",
			GroupJoin:   "!group join",
			GroupLeave:  "!group leave",
			GroupInfo:   "!group info",
			GroupModify: "!group modify",

			// Event commands
			EventCreate:     "!event create",
			EventModify:     "!event modify",
			EventConfirm:    "!event confirm",
			EventUnconfirm:  "!event unconfirm",
			EventWaitlist:   "!event waitlist",
			EventInfo:       "!event info",
			EventChangeHost: "!event changehost",
		},
	}
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the JSON
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	// Marshal the JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write the file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
} 