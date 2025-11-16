FROM golang:1.21-bullseye AS builder

WORKDIR /app

# Install build dependencies for CGO (Debian has better SQLite support)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files first
COPY go.mod go.sum ./

# Verify module and download dependencies
RUN go mod verify && go mod download

# Copy all source code (copy everything except what's in .dockerignore)
COPY . .

# Ensure we're in module mode
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Set CFLAGS for SQLite compilation (fix pread64/pwrite64 issues)
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

# Build combined bot+server
RUN go build -o bin/combined ./cmd/combined

# Final stage - include Bot API server
FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Telegram Bot API server
RUN curl -L https://github.com/tdlib/telegram-bot-api/releases/latest/download/telegram-bot-api_Linux_x86_64.tar.gz -o /tmp/bot-api.tar.gz && \
    tar -xzf /tmp/bot-api.tar.gz -C /usr/local/bin/ telegram-bot-api && \
    chmod +x /usr/local/bin/telegram-bot-api && \
    rm /tmp/bot-api.tar.gz

# Copy binary
COPY --from=builder /app/bin/combined /app/bin/combined

# Create directories
RUN mkdir -p /app/data /app/storage /var/lib/telegram-bot-api

# Set environment variables for Bot API server
ENV TELEGRAM_API_ID=33608323
ENV TELEGRAM_API_HASH=339982c3dc6fa78474ea07d77a9b0d7b

EXPOSE 8080 8081

# Use PORT environment variable (Render/Railway provides this)
ENV PORT=8080

# Start both Bot API server and combined bot+server
CMD /usr/local/bin/telegram-bot-api --local --http-port=8081 --dir=/var/lib/telegram-bot-api & \
    sleep 2 && \
    ./bin/combined

