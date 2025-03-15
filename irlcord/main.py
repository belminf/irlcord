import os
import logging
import discord
from discord.ext import commands
import yaml
import asyncio
import sqlite3
from pathlib import Path

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("irlcord.log"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger("irlcord")

# Load configuration
def load_config():
    config_path = Path(__file__).parent.parent / "config.yaml"
    with open(config_path, "r") as f:
        config = yaml.safe_load(f)
    
    # Override bot token with environment variable if available
    env_token = os.environ.get("DISCORD_BOT_TOKEN")
    if env_token:
        logger.info("Using Discord bot token from environment variable")
        if "general" not in config:
            config["general"] = {}
        config["general"]["bot_token"] = env_token
    
    return config

# Initialize the bot
class IRLCordBot(commands.Bot):
    def __init__(self, config):
        intents = discord.Intents.default()
        intents.message_content = True
        intents.members = True
        
        self.config = config
        self.db = None
        
        super().__init__(
            command_prefix=commands.when_mentioned_or("!"),
            intents=intents,
            help_command=None
        )
    
    async def setup_hook(self):
        # Connect to database
        await self.connect_db()
        
        # Load cogs
        for cog_file in Path(__file__).parent.glob("cogs/*.py"):
            if cog_file.stem.startswith("_"):
                continue
            
            try:
                await self.load_extension(f"irlcord.cogs.{cog_file.stem}")
                logger.info(f"Loaded extension: {cog_file.stem}")
            except Exception as e:
                logger.error(f"Failed to load extension {cog_file.stem}: {e}")
    
    async def connect_db(self):
        """Connect to the SQLite database"""
        db_path = Path(__file__).parent.parent / "irlcord.db"
        self.db = sqlite3.connect(db_path)
        self.db.row_factory = sqlite3.Row
        
        # Initialize database if it doesn't exist
        schema_path = Path(__file__).parent.parent / "schema.sql"
        if schema_path.exists():
            with open(schema_path, "r") as f:
                self.db.executescript(f.read())
                self.db.commit()
            logger.info("Database initialized")
    
    async def on_ready(self):
        logger.info(f"Logged in as {self.user} (ID: {self.user.id})")
        logger.info(f"Connected to {len(self.guilds)} guilds")
        
        # Set bot status
        await self.change_presence(
            activity=discord.Activity(
                type=discord.ActivityType.listening,
                name=f"{self.command_prefix}help"
            )
        )

async def start_bot():
    # Load config
    config = load_config()
    
    # Check if bot token is available
    if not config.get("general", {}).get("bot_token"):
        logger.error("No Discord bot token found in config or environment variables")
        return
    
    # Create and start the bot
    bot = IRLCordBot(config)
    async with bot:
        await bot.start(config["general"]["bot_token"])

def main():
    """Entry point for the application."""
    try:
        asyncio.run(start_bot())
    except KeyboardInterrupt:
        logger.info("Bot stopped by user")
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        raise

if __name__ == "__main__":
    main() 