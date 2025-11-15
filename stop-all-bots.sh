#!/bin/bash
echo "🔍 Checking for running bot instances..."

# Check local Go processes
echo "Checking local Go processes..."
LOCAL=$(ps aux | grep -E "go run.*cmd/(bot|combined)|hafton-movie-bot" | grep -v grep)
if [ ! -z "$LOCAL" ]; then
    echo "Found local processes:"
    echo "$LOCAL"
    echo "Killing local processes..."
    pkill -f "cmd/bot" 2>/dev/null
    pkill -f "cmd/combined" 2>/dev/null
    pkill -f "hafton-movie-bot" 2>/dev/null
else
    echo "✅ No local processes found"
fi

# Check Docker
echo ""
echo "Checking Docker containers..."
if command -v docker &> /dev/null; then
    DOCKER=$(docker ps -q --filter "name=hafton" 2>/dev/null)
    if [ ! -z "$DOCKER" ]; then
        echo "Found Docker containers, stopping..."
        docker stop $DOCKER 2>/dev/null
    else
        echo "✅ No Docker containers found"
    fi
else
    echo "Docker not installed, skipping..."
fi

echo ""
echo "✅ Done! Only Render should be running now."
echo "Check Render dashboard to confirm: https://dashboard.render.com"
