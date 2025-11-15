# Hafton Movie Bot - Telegram Streaming Bot

A self-hosted Telegram bot that downloads files and provides unlimited streaming links with byte-range support for seeking, rewinding, and resuming playback.

## Features

- 🔥 **Unlimited file size** - No 2GB limits
- 🎬 **Byte-range streaming** - Full seeking support (VLC, browsers, smart TVs)
- ⚡ **Fast streaming** - Your own server, your own speed
- 🔒 **Auto-expiration** - Files auto-delete after 5 days
- 📱 **Universal compatibility** - Works on all devices and players
- 🆓 **Free hosting** - Designed for Oracle Cloud Free Tier

## Architecture

- **Telegram Bot** (`cmd/bot`) - Receives files, downloads from Telegram, generates links
- **HTTP Streaming Server** (`cmd/server`) - Serves files with HTTP 206 Partial Content support
- **SQLite Database** - Tracks files and expiration dates
- **Auto-cleanup** - Background goroutine deletes expired files hourly

## Quick Start

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Configure:**
   Edit `config/config.yaml` with your bot token:
   ```yaml
   telegram:
     bot_token: "YOUR_BOT_TOKEN"
   ```

3. **Run bot:**
   ```bash
   go run cmd/bot/main.go
   ```

4. **Run server (in another terminal):**
   ```bash
   go run cmd/server/main.go
   ```

5. **Test:**
   - Send a file to your Telegram bot
   - Bot will reply with streaming and download links
   - Open the streaming link in VLC or browser

### Docker Deployment

1. **Set environment variable:**
   ```bash
   export DOMAIN=yourdomain.com
   ```

2. **Start services:**
   ```bash
   docker-compose up -d
   ```

### VPS Deployment (Oracle Cloud Free Tier)

1. **Create VPS instance:**
   - Sign up at [Oracle Cloud](https://www.oracle.com/cloud/free/)
   - Create Ubuntu 22.04 instance (1 vCPU, 1GB RAM)
   - Open ports 22, 80, 443

2. **Get a free domain:**
   - Use [Freenom](https://www.freenom.com) (.tk, .ml, .ga, .cf, .gq)
   - Or use a subdomain from [NoIP](https://www.noip.com) or [DuckDNS](https://www.duckdns.org)

3. **Deploy:**
   ```bash
   # On your local machine, copy files to VPS
   scp -r . user@your-vps-ip:/opt/hafton-movie-bot
   
   # SSH into VPS
   ssh user@your-vps-ip
   
   # Run deployment script
   cd /opt/hafton-movie-bot
   chmod +x deploy.sh
   sudo ./deploy.sh
   ```

4. **Configure domain:**
   ```bash
   # Edit config
   sudo nano /opt/hafton-movie-bot/config/config.yaml
   # Set domain: yourdomain.com
   
   # Edit nginx config
   sudo nano /etc/nginx/sites-available/streaming
   # Replace 'yourdomain.com' with your actual domain
   ```

5. **Set up DNS:**
   - Point your domain's A record to your VPS IP address
   - Wait for DNS propagation (5-60 minutes)

6. **Set up SSL:**
   ```bash
   sudo /opt/hafton-movie-bot/scripts/setup-ssl.sh yourdomain.com your@email.com
   ```

7. **Start services:**
   ```bash
   sudo systemctl start bot
   sudo systemctl start server
   sudo systemctl restart nginx
   sudo systemctl enable bot
   sudo systemctl enable server
   ```

## Configuration

Edit `config/config.yaml`:

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN"

server:
  port: 8080
  domain: "yourdomain.com"  # Your domain name
  storage_path: "./storage"

database:
  path: "./data/bot.db"

retention:
  days: 5  # Files expire after 5 days
```

## Usage

1. **Forward or upload a file** to your Telegram bot
2. **Bot replies** with:
   - Streaming link: `https://yourdomain.com/stream/ABC12345`
   - Download link: `https://yourdomain.com/file/ABC12345`
3. **Open streaming link** in:
   - VLC Media Player
   - Web browser
   - Smart TV
   - Mobile apps
   - Any media player that supports HTTP streaming

## File Expiration

- Files automatically expire after 5 days (configurable)
- Expired links show a friendly UI asking to re-upload
- Background cleanup runs every hour
- No manual intervention needed

## API Endpoints

- `GET /stream/<id>` - Stream file with byte-range support
- `GET /file/<id>` - Direct download
- `GET /health` - Health check

## Requirements

- Go 1.21+
- SQLite3
- 100GB+ storage (depending on usage)
- VPS with 1GB+ RAM (Oracle Cloud Free Tier works)

## Free Hosting Options

### VPS (Recommended)
- **Oracle Cloud Free Tier** - 2 VMs, 1 vCPU, 1GB RAM each, 10TB bandwidth/month
- **Google Cloud Free Tier** - f1-micro instance
- **AWS Free Tier** - t2.micro (750 hours/month)

### Free Domains
- **Freenom** - .tk, .ml, .ga, .cf, .gq domains
- **DuckDNS** - Free subdomain
- **NoIP** - Free subdomain

## Troubleshooting

### Bot not receiving files
- Check bot token in `config/config.yaml`
- Ensure bot is running: `systemctl status bot`

### Streaming not working
- Check server is running: `systemctl status server`
- Verify domain is set correctly
- Check nginx logs: `tail -f /var/log/nginx/error.log`

### SSL certificate issues
- Ensure DNS A record points to VPS IP
- Wait for DNS propagation
- Check firewall allows ports 80 and 443

### Files not deleting
- Check cleanup logs in server output
- Verify database path is correct
- Check file permissions

## Security

- Files stored with 600 permissions (owner read/write only)
- Unique IDs prevent guessing file paths
- Input validation on all endpoints
- SQL injection protection via parameterized queries

## License

MIT License - Feel free to use and modify as needed.

## Support

For issues or questions, check the logs:
- Bot: `journalctl -u bot -f`
- Server: `journalctl -u server -f`
- Nginx: `tail -f /var/log/nginx/error.log`

# hafton-streamer
