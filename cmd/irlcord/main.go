package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/azlyth/irlcord/pkg/bot"
	"github.com/azlyth/irlcord/pkg/config"
	"github.com/azlyth/irlcord/pkg/db"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting IRLCord Discord bot...")

	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// If the config file doesn't exist, create a default one
		if os.IsNotExist(err) {
			log.Printf("Config file not found, creating default config at %s", *configPath)
			cfg = config.DefaultConfig()
			err = config.SaveConfig(cfg, *configPath)
			if err != nil {
				log.Fatalf("Error creating default config: %v", err)
			}
		} else {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Check if Discord token is set
	if cfg.DiscordToken == "" {
		log.Fatalf("Discord token not set in config file")
	}

	// Initialize database
	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize the bot
	discordBot, err := bot.New(cfg, database)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Start the bot
	err = discordBot.Start()
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
	
	log.Println("Bot is now running. Press CTRL-C to exit.")

	// Wait for a SIGINT or SIGTERM signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	log.Println("Shutting down...")
	
	// Stop the bot
	err = discordBot.Stop()
	if err != nil {
		log.Printf("Error stopping bot: %v", err)
	}
} 