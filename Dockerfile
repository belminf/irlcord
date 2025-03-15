FROM golang:1.23-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o irlcord-bot .

# Use a smaller image for the final build
FROM alpine:latest

# Install SQLite and other dependencies
RUN apk add --no-cache sqlite ca-certificates libc6-compat

WORKDIR /app

# Copy the binary and config example from the builder stage
COPY --from=builder /app/irlcord-bot .
COPY --from=builder /app/config.json.example .

# Create a directory for the database
RUN mkdir -p /app/data

# Set environment variables
ENV CONFIG_PATH=/app/config.json
ENV DATABASE_PATH=/app/data/irlcord.db

# Expose any necessary ports
EXPOSE 8080

# Run the application (default command, can be overridden by docker-compose)
CMD ["./irlcord-bot", "-config", "/app/config.json"] 