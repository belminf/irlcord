.PHONY: setup install update run lint format test clean help docker-build docker-run docker-stop docker-logs

# Default target
.DEFAULT_GOAL := help

# Variables
PYTHON := python3
POETRY := poetry
CONFIG_FILE := config.yaml
SCHEMA_FILE := schema.sql
DOCKER_COMPOSE := docker compose

help:
	@echo "Available commands:"
	@echo "  make setup      - Install Poetry and project dependencies"
	@echo "  make install    - Install project dependencies"
	@echo "  make update     - Update project dependencies"
	@echo "  make run        - Run the Discord bot"
	@echo "  make lint       - Run linters (flake8)"
	@echo "  make format     - Format code (black, isort)"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean up temporary files"
	@echo "  make init-db    - Initialize the database with schema"
	@echo "  make config     - Create a sample config file if it doesn't exist"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-stop  - Stop Docker container"
	@echo "  make docker-logs  - View Docker container logs"

setup:
	@echo "Installing Poetry..."
	@curl -sSL https://install.python-poetry.org | $(PYTHON) -
	@$(POETRY) install
	@echo "Setup complete!"

install:
	@echo "Installing dependencies..."
	@$(POETRY) install
	@echo "Installation complete!"

update:
	@echo "Updating dependencies..."
	@$(POETRY) update
	@echo "Update complete!"

run: config
	@echo "Running Discord bot..."
	@$(POETRY) run python -m irlcord.main

lint:
	@echo "Running linters..."
	@$(POETRY) run flake8 irlcord
	@echo "Linting complete!"

format:
	@echo "Formatting code..."
	@$(POETRY) run isort irlcord
	@$(POETRY) run black irlcord
	@echo "Formatting complete!"

test:
	@echo "Running tests..."
	@$(POETRY) run pytest
	@echo "Tests complete!"

clean:
	@echo "Cleaning up..."
	@rm -rf __pycache__
	@rm -rf .pytest_cache
	@rm -rf dist
	@rm -rf *.egg-info
	@find . -type d -name __pycache__ -exec rm -rf {} +
	@find . -type f -name "*.pyc" -delete
	@echo "Cleanup complete!"

init-db:
	@echo "Initializing database..."
	@if [ ! -f $(SCHEMA_FILE) ]; then \
		echo "Error: $(SCHEMA_FILE) not found!"; \
		exit 1; \
	fi
	@$(POETRY) run python -c "import sqlite3; conn = sqlite3.connect('irlcord.db'); conn.executescript(open('$(SCHEMA_FILE)').read()); conn.commit(); conn.close(); print('Database initialized successfully!')"

config:
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "Creating sample config file..."; \
		echo "general:" > $(CONFIG_FILE); \
		echo "  bot_token: \"YOUR_BOT_TOKEN_HERE\"" >> $(CONFIG_FILE); \
		echo "  admin_user_ids: [\"YOUR_DISCORD_USER_ID\"]" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "terminology:" >> $(CONFIG_FILE); \
		echo "  group_plural: \"Circles\"" >> $(CONFIG_FILE); \
		echo "  group_singular: \"Circle\"" >> $(CONFIG_FILE); \
		echo "  member_plural: \"Folks\"" >> $(CONFIG_FILE); \
		echo "  member_singular: \"Person\"" >> $(CONFIG_FILE); \
		echo "  leader_plural: \"Leaders\"" >> $(CONFIG_FILE); \
		echo "  leader_singular: \"Leader\"" >> $(CONFIG_FILE); \
		echo "  event_plural: \"Events\"" >> $(CONFIG_FILE); \
		echo "  event_singular: \"Event\"" >> $(CONFIG_FILE); \
		echo "  contributor_plural: \"Adventurers\"" >> $(CONFIG_FILE); \
		echo "  contributor_singular: \"Adventurer\"" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "commands:" >> $(CONFIG_FILE); \
		echo "  # Group Management" >> $(CONFIG_FILE); \
		echo "  group_create: \"circle new\"" >> $(CONFIG_FILE); \
		echo "  group_join: \"circle join\"" >> $(CONFIG_FILE); \
		echo "  group_leave: \"circle leave\"" >> $(CONFIG_FILE); \
		echo "  group_info: \"circle info\"" >> $(CONFIG_FILE); \
		echo "  group_modify: \"circle modify\"" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "  # Event Management" >> $(CONFIG_FILE); \
		echo "  event_create: \"event new\"" >> $(CONFIG_FILE); \
		echo "  event_modify: \"event modify\"" >> $(CONFIG_FILE); \
		echo "  event_confirm: \"event confirm\"" >> $(CONFIG_FILE); \
		echo "  event_unconfirm: \"event unconfirm\"" >> $(CONFIG_FILE); \
		echo "  event_waitlist: \"event waitlist\"" >> $(CONFIG_FILE); \
		echo "  event_info: \"event info\"" >> $(CONFIG_FILE); \
		echo "  event_change_host: \"event change host\"" >> $(CONFIG_FILE); \
		echo "Sample config file created! Please update with your bot token and admin user ID."; \
	else \
		echo "Config file already exists."; \
	fi

docker-build:
	@echo "Building Docker image..."
	@$(DOCKER_COMPOSE) build
	@echo "Docker image built successfully!"

docker-run: config
	@echo "Running Docker container..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Docker container started successfully!"

docker-stop:
	@echo "Stopping Docker container..."
	@$(DOCKER_COMPOSE) down
	@echo "Docker container stopped successfully!"

docker-logs:
	@echo "Viewing Docker container logs..."
	@$(DOCKER_COMPOSE) logs -f 