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

# Verify files are copied correctly and module is recognized
RUN echo "=== Root directory ===" && \
    ls -la && \
    echo "=== cmd/ directory ===" && \
    ls -la cmd/ && \
    echo "=== internal/ directory ===" && \
    ls -la internal/ && \
    echo "=== internal/storage/ directory ===" && \
    ls -la internal/storage/ && \
    echo "=== go.mod content ===" && \
    cat go.mod && \
    echo "=== Module name ===" && \
    go list -m && \
    echo "=== Testing import resolution ===" && \
    go list ./internal/storage

# Ensure we're in module mode and build from module root
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Build server only (ensure we're building from the module root)
WORKDIR /app
RUN go build -v -o bin/server ./cmd/server

# Final stage
FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/server /app/bin/server

# Create directories
RUN mkdir -p /app/data /app/storage

EXPOSE 8080

# Use PORT environment variable (Railway provides this)
ENV PORT=8080
CMD ["./bin/server"]

