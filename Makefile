.PHONY: build run clean docker-build docker-run docker-stop docker-logs help

# Variables
BINARY_NAME=irlcord-bot
CONFIG_FILE=config.json

# Default target
all: build

# Help
help:
	@echo "IRLCord Discord Bot - Makefile commands:"
	@echo "  make build       - Build the bot"
	@echo "  make run         - Run the bot"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run  - Run Docker container"
	@echo "  make docker-stop - Stop Docker container"
	@echo "  make docker-logs - View Docker container logs"
	@echo "  make help        - Show this help message"

# Build the bot
build:
	@echo "Building IRLCord bot..."
	@go build -o $(BINARY_NAME) .

# Run the bot
run: build
	@echo "Running IRLCord bot..."
	@./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@go clean

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker compose build

docker-run:
	@echo "Running Docker container..."
	@docker compose up -d

docker-stop:
	@echo "Stopping Docker container..."
	@docker compose down

docker-logs:
	@echo "Viewing Docker container logs..."
	@docker compose logs -f 