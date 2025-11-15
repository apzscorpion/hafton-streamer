#!/bin/bash

# Local development setup script for macOS/Linux
# This sets up the environment for local development

set -e

echo "=== Local Development Setup ==="

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Detected macOS"
    
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        echo "❌ Homebrew not found. Install it from https://brew.sh"
        exit 1
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "Installing Go via Homebrew..."
        brew install go
    else
        echo "✓ Go is already installed: $(go version)"
    fi
    
    # Check if SQLite is installed
    if ! command -v sqlite3 &> /dev/null; then
        echo "Installing SQLite via Homebrew..."
        brew install sqlite
    else
        echo "✓ SQLite is already installed: $(sqlite3 --version)"
    fi
    
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Detected Linux"
    
    # Check package manager
    if command -v apt-get &> /dev/null; then
        echo "Installing dependencies via apt-get..."
        sudo apt-get update
        sudo apt-get install -y golang-go sqlite3
    elif command -v yum &> /dev/null; then
        echo "Installing dependencies via yum..."
        sudo yum install -y golang sqlite
    elif command -v pacman &> /dev/null; then
        echo "Installing dependencies via pacman..."
        sudo pacman -S --noconfirm go sqlite
    else
        echo "⚠️  Could not detect package manager. Please install Go and SQLite manually."
    fi
else
    echo "⚠️  Unsupported OS: $OSTYPE"
    echo "Please install Go 1.21+ and SQLite3 manually"
fi

# Verify Go installation
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "✓ Go version: $GO_VERSION"
    
    # Check Go version (should be 1.21+)
    GO_MAJOR=$(echo $GO_VERSION | sed 's/go\([0-9]*\)\..*/\1/')
    GO_MINOR=$(echo $GO_VERSION | sed 's/go[0-9]*\.\([0-9]*\).*/\1/')
    
    if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
        echo "⚠️  Warning: Go 1.21+ is recommended. You have $GO_VERSION"
    fi
else
    echo "❌ Go installation failed"
    exit 1
fi

# Download dependencies
echo ""
echo "Downloading Go dependencies..."
go mod download
go mod tidy

# Create necessary directories
echo ""
echo "Creating directories..."
mkdir -p data
mkdir -p storage
mkdir -p bin

# Build binaries
echo ""
echo "Building application..."
CGO_ENABLED=1 go build -o bin/bot ./cmd/bot
CGO_ENABLED=1 go build -o bin/server ./cmd/server

echo ""
echo "✅ Local development setup complete!"
echo ""
echo "To run locally:"
echo "  Terminal 1: go run cmd/bot/main.go"
echo "  Terminal 2: go run cmd/server/main.go"
echo ""
echo "Or use Makefile:"
echo "  make run-bot"
echo "  make run-server"
echo ""

