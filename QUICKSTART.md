# Quick Start Guide

Get your streaming bot running in 5 minutes!

## Prerequisites

- Go 1.21+ installed
- Telegram bot token (get from @BotFather)
- A domain name (optional for local testing)

## Local Development (5 minutes)

### Option A: Automated Setup (Recommended)

```bash
# Run the local setup script (works on macOS and Linux)
./setup-local.sh
```

### Option B: Manual Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Configure Bot Token

Edit `config/config.yaml`:

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"
```

### 3. Run Bot (Terminal 1)

```bash
go run cmd/bot/main.go
```

### 4. Run Server (Terminal 2)

```bash
go run cmd/server/main.go
```

### 5. Test

1. Open Telegram
2. Find your bot
3. Send a video file
4. Bot replies with streaming link
5. Open link in VLC or browser!

**Note:** For local testing, use `http://localhost:8080/stream/<id>` instead of HTTPS.

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment guide.

### Quick Deploy (Oracle Cloud)

```bash
# 1. Create VPS on Oracle Cloud (free)
# 2. Copy files to VPS
scp -r . ubuntu@YOUR_VPS_IP:/opt/hafton-movie-bot

# 3. SSH into VPS
ssh ubuntu@YOUR_VPS_IP

# 4. Run deployment
cd /opt/hafton-movie-bot
sudo ./deploy.sh

# 5. Configure domain in config.yaml and nginx config
# 6. Set up SSL
sudo ./scripts/setup-ssl.sh yourdomain.com your@email.com

# 7. Start services
sudo systemctl start bot
sudo systemctl start server
sudo systemctl restart nginx
```

## Using Makefile

```bash
# Build everything
make build

# Run bot
make run-bot

# Run server
make run-server

# Clean build artifacts
make clean
```

## Docker Deployment

```bash
# Set domain
export DOMAIN=yourdomain.com

# Start services
docker-compose up -d

# View logs
docker-compose logs -f
```

## Troubleshooting

**Bot not starting?**
- Check bot token in config.yaml
- Verify Go is installed: `go version`

**Server not starting?**
- Check port 8080 is available
- Verify config file exists

**Streaming not working?**
- Check server is running
- Verify file exists in storage directory
- Check file permissions

## Next Steps

- Read [README.md](README.md) for full documentation
- Read [DEPLOYMENT.md](DEPLOYMENT.md) for production setup
- Customize retention period in config.yaml
- Add your domain for HTTPS streaming

