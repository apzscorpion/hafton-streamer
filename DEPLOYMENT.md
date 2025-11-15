# Deployment Guide - Free Hosting Setup

This guide walks you through deploying the Hafton Movie Bot on free hosting services.

## Option 1: Oracle Cloud Free Tier (Recommended)

### Step 1: Create Oracle Cloud Account

1. Go to [Oracle Cloud Free Tier](https://www.oracle.com/cloud/free/)
2. Sign up for a free account (requires credit card, but won't be charged)
3. Verify your email

### Step 2: Create Compute Instance

1. Log into Oracle Cloud Console
2. Navigate to **Compute** → **Instances**
3. Click **Create Instance**
4. Configure:
   - **Name**: `hafton-movie-bot`
   - **Image**: Ubuntu 22.04
   - **Shape**: VM.Standard.A1.Flex (1 OCPU, 1GB RAM) - FREE
   - **Networking**: Create new VCN or use default
   - **SSH Keys**: Upload your public SSH key
5. Click **Create**

### Step 3: Configure Firewall

1. Go to **Networking** → **Security Lists**
2. Edit the security list for your VCN
3. Add Ingress Rules:
   - **Port 22** (SSH) - Source: 0.0.0.0/0
   - **Port 80** (HTTP) - Source: 0.0.0.0/0
   - **Port 443** (HTTPS) - Source: 0.0.0.0/0

### Step 4: Get Free Domain

**Option A: Freenom (Free .tk, .ml, .ga domains)**
1. Go to [Freenom](https://www.freenom.com)
2. Search for a domain (e.g., `mybot.tk`)
3. Add to cart and checkout (free)
4. Go to **My Domains** → **Manage Domain**
5. Add A record pointing to your VPS IP

**Option B: DuckDNS (Free subdomain)**
1. Go to [DuckDNS](https://www.duckdns.org)
2. Sign up and create a subdomain (e.g., `mybot.duckdns.org`)
3. Update IP address to your VPS IP

**Option C: NoIP (Free subdomain)**
1. Go to [NoIP](https://www.noip.com)
2. Sign up and create a hostname
3. Point to your VPS IP

### Step 5: Deploy Application

**On your local machine:**

```bash
# Clone or copy the project
cd hafton-movie-bot

# Copy to VPS
scp -r . ubuntu@YOUR_VPS_IP:/opt/hafton-movie-bot
```

**SSH into your VPS:**

```bash
ssh ubuntu@YOUR_VPS_IP
```

**On VPS:**

```bash
# Navigate to project directory
cd /opt/hafton-movie-bot

# Make deployment script executable
chmod +x deploy.sh

# Run deployment script
sudo ./deploy.sh
```

### Step 6: Configure Domain

```bash
# Edit config file
sudo nano /opt/hafton-movie-bot/config/config.yaml
# Set domain: yourdomain.com (or yourdomain.duckdns.org)

# Edit nginx config
sudo nano /etc/nginx/sites-available/streaming
# Replace all instances of 'yourdomain.com' with your actual domain
```

### Step 7: Set Up SSL Certificate

```bash
# Run SSL setup script
sudo /opt/hafton-movie-bot/scripts/setup-ssl.sh yourdomain.com your@email.com

# Update nginx config with your domain (if not done already)
sudo nano /etc/nginx/sites-available/streaming
# Replace 'yourdomain.com' with your domain

# Restart nginx
sudo systemctl restart nginx
```

### Step 8: Start Services

```bash
# Start bot and server
sudo systemctl start bot
sudo systemctl start server

# Enable auto-start on boot
sudo systemctl enable bot
sudo systemctl enable server

# Check status
sudo systemctl status bot
sudo systemctl status server
```

### Step 9: Verify

1. Send a file to your Telegram bot
2. Bot should reply with streaming links
3. Open the streaming link in VLC or browser
4. Test seeking/rewinding (should work!)

## Option 2: Google Cloud Free Tier

### Step 1: Create GCP Account

1. Go to [Google Cloud Platform](https://cloud.google.com/)
2. Sign up (get $300 free credit)
3. Create a new project

### Step 2: Create VM Instance

1. Go to **Compute Engine** → **VM Instances**
2. Click **Create Instance**
3. Configure:
   - **Name**: `hafton-movie-bot`
   - **Machine type**: f1-micro (free tier)
   - **Boot disk**: Ubuntu 22.04
   - **Firewall**: Allow HTTP and HTTPS traffic
4. Click **Create**

### Step 3: Deploy

Follow steps 4-9 from Oracle Cloud guide above.

## Option 3: AWS Free Tier

### Step 1: Create AWS Account

1. Go to [AWS Free Tier](https://aws.amazon.com/free/)
2. Sign up (requires credit card)

### Step 2: Create EC2 Instance

1. Go to **EC2** → **Instances**
2. Click **Launch Instance**
3. Configure:
   - **AMI**: Ubuntu Server 22.04 LTS
   - **Instance type**: t2.micro (free tier)
   - **Security group**: Allow SSH (22), HTTP (80), HTTPS (443)
4. Launch instance

### Step 3: Deploy

Follow steps 4-9 from Oracle Cloud guide above.

## Troubleshooting

### Can't connect via SSH
- Check security group/firewall rules
- Verify SSH key is correct
- Check instance is running

### Domain not resolving
- Wait 5-60 minutes for DNS propagation
- Verify A record points to correct IP
- Use `dig yourdomain.com` to check DNS

### SSL certificate fails
- Ensure DNS is pointing to your server
- Check port 80 is open (needed for verification)
- Verify domain is accessible via HTTP

### Services not starting
- Check logs: `journalctl -u bot -n 50`
- Verify config file is correct
- Check file permissions

### Bot not receiving messages
- Verify bot token in config.yaml
- Check bot is running: `systemctl status bot`
- Test bot token: `curl https://api.telegram.org/bot<TOKEN>/getMe`

## Monitoring

### View logs
```bash
# Bot logs
sudo journalctl -u bot -f

# Server logs
sudo journalctl -u server -f

# Nginx logs
sudo tail -f /var/log/nginx/error.log
```

### Check disk space
```bash
df -h
du -sh /opt/hafton-movie-bot/storage/*
```

### Restart services
```bash
sudo systemctl restart bot
sudo systemctl restart server
sudo systemctl restart nginx
```

## Maintenance

### Update application
```bash
cd /opt/hafton-movie-bot
git pull  # if using git
# Or copy new files via scp

# Rebuild
go build -o bin/bot ./cmd/bot
go build -o bin/server ./cmd/server

# Restart services
sudo systemctl restart bot
sudo systemctl restart server
```

### Backup database
```bash
cp /opt/hafton-movie-bot/data/bot.db /backup/bot-$(date +%Y%m%d).db
```

## Cost Breakdown

**Oracle Cloud Free Tier:**
- Compute: FREE (2 VMs, 1 vCPU, 1GB RAM each)
- Storage: FREE (200GB)
- Bandwidth: FREE (10TB/month)
- **Total: $0/month**

**Google Cloud Free Tier:**
- f1-micro: FREE (always free)
- Storage: ~$5/month for 100GB
- Bandwidth: FREE (1GB/month, then paid)
- **Total: ~$5/month**

**AWS Free Tier:**
- t2.micro: FREE (750 hours/month = 1 instance)
- Storage: FREE (30GB)
- Bandwidth: FREE (15GB/month)
- **Total: $0/month** (if within free tier limits)

## Next Steps

1. Set up monitoring (optional)
2. Configure automatic backups
3. Set up log rotation
4. Consider adding rate limiting
5. Add analytics (optional)

