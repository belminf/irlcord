#!/bin/bash

# Exit on error
set -e

# Create directories if they don't exist
mkdir -p scripts

# Check if Poetry is installed
if ! command -v poetry &> /dev/null; then
    echo "Poetry is not installed. Installing Poetry..."
    curl -sSL https://install.python-poetry.org | python3 -
    echo "Poetry installed successfully!"
else
    echo "Poetry is already installed."
fi

# Install dependencies
echo "Installing dependencies..."
poetry install

# Create config file if it doesn't exist
if [ ! -f config.yaml ]; then
    echo "Creating sample config file..."
    make config
    echo "Please update config.yaml with your Discord bot token and admin user ID."
else
    echo "Config file already exists."
fi

# Initialize database
echo "Initializing database..."
make init-db

echo "Development setup complete!"
echo "Run 'make run' to start the bot." 