#!/bin/bash

# Deployment script for Oracle Cloud Free Tier (Ubuntu 22.04)
# Run this script on your VPS (Linux/Ubuntu), NOT on your local macOS/Windows machine

set -e

echo "=== Hafton Movie Bot Deployment Script ==="

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "❌ ERROR: This script is designed for Linux/Ubuntu VPS servers."
    echo ""
    echo "You're running macOS. This script should be run on your VPS server, not locally."
    echo ""
    echo "To deploy:"
    echo "1. Copy files to your VPS: scp -r . user@your-vps-ip:/opt/hafton-movie-bot"
    echo "2. SSH into VPS: ssh user@your-vps-ip"
    echo "3. Then run this script on the VPS"
    echo ""
    echo "For local development, use:"
    echo "  make run-bot    # Run bot locally"
    echo "  make run-server # Run server locally"
    exit 1
fi

# Check if running on Linux with apt-get
if ! command -v apt-get &> /dev/null; then
    echo "❌ ERROR: This script requires apt-get (Ubuntu/Debian Linux)."
    echo ""
    echo "This script is designed for Ubuntu/Debian-based VPS servers."
    echo "If you're on a different Linux distribution, you may need to adapt the package manager commands."
    exit 1
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ ERROR: This script must be run as root (use sudo)"
    exit 1
fi

# Update system
echo "Updating system packages..."
apt-get update
apt-get upgrade -y

# Install required packages
echo "Installing required packages..."
apt-get install -y \
    golang-go \
    sqlite3 \
    nginx \
    certbot \
    python3-certbot-nginx \
    git \
    build-essential

# Create application directory
APP_DIR="/opt/hafton-movie-bot"
echo "Creating application directory at ${APP_DIR}..."
mkdir -p ${APP_DIR}
mkdir -p ${APP_DIR}/data
mkdir -p ${APP_DIR}/storage
mkdir -p ${APP_DIR}/config

# Copy application files (assuming you're running this from the project directory)
echo "Copying application files..."
cp -r . ${APP_DIR}/

# Build the application
echo "Building application..."
cd ${APP_DIR}
go mod download
CGO_ENABLED=1 go build -o ${APP_DIR}/bin/bot ./cmd/bot
CGO_ENABLED=1 go build -o ${APP_DIR}/bin/server ./cmd/server

# Set permissions
chmod +x ${APP_DIR}/bin/bot
chmod +x ${APP_DIR}/bin/server
chmod +x ${APP_DIR}/scripts/setup-ssl.sh

# Copy systemd service files
echo "Setting up systemd services..."
cp ${APP_DIR}/systemd/bot.service /etc/systemd/system/
cp ${APP_DIR}/systemd/server.service /etc/systemd/system/

# Reload systemd
systemctl daemon-reload

# Configure firewall
echo "Configuring firewall..."
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Set up nginx (user needs to configure domain)
echo "Setting up nginx..."
cp ${APP_DIR}/nginx/streaming.conf /etc/nginx/sites-available/streaming
ln -sf /etc/nginx/sites-available/streaming /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Test nginx configuration
nginx -t

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Next steps:"
echo "1. Edit ${APP_DIR}/config/config.yaml and set your domain"
echo "2. Edit /etc/nginx/sites-available/streaming and replace 'yourdomain.com' with your actual domain"
echo "3. Point your domain DNS A record to this server's IP address"
echo "4. Run: ${APP_DIR}/scripts/setup-ssl.sh yourdomain.com your@email.com"
echo "5. Start services:"
echo "   systemctl start bot"
echo "   systemctl start server"
echo "   systemctl restart nginx"
echo "6. Enable services to start on boot:"
echo "   systemctl enable bot"
echo "   systemctl enable server"
echo ""

