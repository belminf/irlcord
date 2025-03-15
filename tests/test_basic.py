import pytest
import os
import sys
from pathlib import Path

# Add the parent directory to sys.path
sys.path.insert(0, str(Path(__file__).parent.parent))

def test_imports():
    """Test that all modules can be imported without errors."""
    from irlcord import main
    from irlcord.utils import db, discord_helpers
    
    assert main is not None
    assert db is not None
    assert discord_helpers is not None

def test_config_path():
    """Test that the config path is correctly resolved."""
    from irlcord.main import load_config
    
    # Create a temporary config file for testing
    config_path = Path(__file__).parent.parent / "config.yaml"
    if not config_path.exists():
        with open(config_path, "w") as f:
            f.write("general:\n  bot_token: 'test_token'\n")
    
    try:
        config = load_config()
        assert config is not None
        assert "general" in config
        assert "bot_token" in config["general"]
    finally:
        # Clean up if we created a test config
        if config_path.exists() and config_path.read_text().strip() == "general:\n  bot_token: 'test_token'":
            os.remove(config_path)

def test_bot_initialization():
    """Test that the bot can be initialized."""
    from irlcord.main import IRLCordBot
    
    config = {"general": {"bot_token": "test_token"}}
    bot = IRLCordBot(config)
    
    assert bot is not None
    assert bot.config == config 