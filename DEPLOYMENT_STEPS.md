# Step-by-Step Deployment Guide

Since you only have the app on your local machine, here's exactly what to do next:

## 🎯 Quick Overview

1. **Get a FREE VPS** (Choose from options below)
2. **Get a FREE domain** (DuckDNS or Freenom)
3. **Copy your app to the VPS**
4. **Run deployment script**
5. **Configure domain & SSL**
6. **Start services**

**Total Cost: $0/month**

---

## Step 1: Choose Your Free VPS Provider

> **💡 No Credit Card?** See `NO_CREDIT_CARD_GUIDE.md` for options that don't require cards (Railway, Render, etc.)

### Option A: Oracle Cloud (Recommended - Best Free Tier) ⭐

**Why choose this:**
- ✅ 2 VMs forever free (1 vCPU, 1GB RAM each)
- ✅ 10TB bandwidth/month (huge!)
- ✅ 200GB storage free
- ✅ No credit card required for free tier
- ✅ Best for long-term projects

**Setup:** See Section 1A below

---

### Option B: Google Cloud Platform (GCP) Free Tier

**Why choose this:**
- ✅ Always free f1-micro instance
- ✅ 0.6GB RAM, 30GB storage
- ✅ 1GB egress/month free
- ⚠️ Requires credit card (won't be charged)
- ⚠️ Less RAM than Oracle

**Setup:** See Section 1B below

---

### Option C: AWS Free Tier

**Why choose this:**
- ✅ 750 hours/month of t2.micro (first 12 months)
- ✅ 1GB RAM, 30GB storage
- ✅ 15GB bandwidth/month
- ⚠️ Only free for 12 months
- ⚠️ Requires credit card

**Setup:** See Section 1C below

---

### Option D: Microsoft Azure Free Tier

**Why choose this:**
- ✅ B1S VM (1 vCPU, 1GB RAM)
- ✅ 10GB storage
- ✅ $200 credit for 30 days
- ⚠️ Requires credit card
- ⚠️ Limited free tier after credit expires

**Setup:** See Section 1D below

---

### Option E: Railway.app (Easiest - No VPS Management)

**Why choose this:**
- ✅ $5 free credit/month
- ✅ No server management needed
- ✅ Auto-deploys from GitHub
- ✅ Built-in SSL
- ⚠️ Requires GitHub account
- ⚠️ Credit expires monthly

**Setup:** See Section 1E below

---

### Option F: Render.com (Also Easy)

**Why choose this:**
- ✅ Free tier available
- ✅ Auto-deploy from GitHub
- ✅ Built-in SSL
- ⚠️ Service sleeps after inactivity
- ⚠️ Limited resources

**Setup:** See Section 1F below

---

## Step 1A: Oracle Cloud Setup (Recommended) - 10 minutes

### 1.1 Sign Up
1. Go to https://www.oracle.com/cloud/free/
2. Click "Start for Free"
3. Sign up (requires credit card but won't be charged)
4. Verify your email

### 1.2 Create Compute Instance
1. Log into Oracle Cloud Console
2. Go to **Menu** → **Compute** → **Instances**
3. Click **Create Instance**
4. Fill in:
   - **Name**: `movie-bot-vps`
   - **Image**: Select **Ubuntu 22.04**
   - **Shape**: **VM.Standard.A1.Flex** (FREE tier)
     - OCPUs: 1
     - Memory: 1 GB
   - **Networking**: Use default VCN
   - **SSH Keys**: 
     - Generate a new key pair OR
     - Upload your existing public SSH key
5. Click **Create**

### 1.3 Get Your VPS IP Address
1. Wait for instance to be **Running** (2-3 minutes)
2. Copy the **Public IP address** (you'll need this)

### 1.4 Configure Firewall
1. Go to **Networking** → **Virtual Cloud Networks**
2. Click on your VCN
3. Go to **Security Lists** → **Default Security List**
4. Click **Add Ingress Rules**:
   - **Port 22** (SSH) - Source: `0.0.0.0/0`
   - **Port 80** (HTTP) - Source: `0.0.0.0/0`
   - **Port 443** (HTTPS) - Source: `0.0.0.0.0/0`
5. Save rules

---

## Step 2: Get a Free Domain - 5 minutes

### Option A: Freenom (Free .tk, .ml, .ga domains)
1. Go to https://www.freenom.com
2. Search for a domain (e.g., `mybot123.tk`)
3. Add to cart → Checkout (FREE)
4. Complete registration
5. Go to **My Domains** → Click **Manage Domain**
6. Go to **Management Tools** → **Nameservers**
7. Add **A Record**:
   - **Name**: `@` (or leave blank)
   - **Type**: `A`
   - **TTL**: `3600`
   - **Target**: Your VPS IP address (from Step 1.3)
8. Save

### Option B: DuckDNS (Free subdomain) - Easier!
1. Go to https://www.duckdns.org
2. Sign in with Google/GitHub
3. Create subdomain (e.g., `mybot.duckdns.org`)
4. Enter your VPS IP address
5. Click **Update IP**
6. Done! (No DNS propagation wait)

**Recommendation**: Use DuckDNS for easier setup, or Freenom if you want a custom domain.

---

## Step 3: Copy Your App to VPS - 5 minutes

### 3.1 On Your Local Machine (macOS)

Open Terminal and run:

```bash
# Navigate to your project
cd /Users/pits/Projects/hafton-movie-bot

# Copy files to VPS (replace with your VPS IP and username)
scp -r . ubuntu@YOUR_VPS_IP:/opt/hafton-movie-bot
```

**Note**: 
- Replace `YOUR_VPS_IP` with the IP from Step 1.3
- Replace `ubuntu` with your VPS username (usually `ubuntu` or `opc`)

### 3.2 SSH Into Your VPS

```bash
ssh ubuntu@YOUR_VPS_IP
```

(Use the SSH key you configured in Step 1.2)

---

## Step 4: Deploy on VPS - 10 minutes

### 4.1 Run Deployment Script

Once SSH'd into your VPS:

```bash
cd /opt/hafton-movie-bot
sudo chmod +x deploy.sh
sudo ./deploy.sh
```

This will:
- Install Go, SQLite, Nginx
- Build your application
- Set up systemd services
- Configure firewall

### 4.2 Configure Domain

```bash
# Edit config file
sudo nano /opt/hafton-movie-bot/config/config.yaml
```

Change:
```yaml
server:
  domain: "yourdomain.duckdns.org"  # Your domain from Step 2
```

Save: `Ctrl+X`, then `Y`, then `Enter`

### 4.3 Configure Nginx

```bash
sudo nano /etc/nginx/sites-available/streaming
```

Replace ALL instances of `yourdomain.com` with your actual domain:
- `yourdomain.duckdns.org` (if using DuckDNS)
- Or your Freenom domain

Save: `Ctrl+X`, then `Y`, then `Enter`

---

## Step 5: Set Up SSL Certificate - 5 minutes

```bash
# Run SSL setup script
sudo /opt/hafton-movie-bot/scripts/setup-ssl.sh yourdomain.duckdns.org your@email.com
```

Replace:
- `yourdomain.duckdns.org` with your domain
- `your@email.com` with your email

**Note**: If using DuckDNS, wait 2-3 minutes after Step 2 for DNS to propagate.

---

## Step 6: Start Services - 2 minutes

```bash
# Start bot and server
sudo systemctl start bot
sudo systemctl start server
sudo systemctl restart nginx

# Enable auto-start on boot
sudo systemctl enable bot
sudo systemctl enable server

# Check status
sudo systemctl status bot
sudo systemctl status server
```

Both should show `active (running)` ✅

---

## Step 7: Test Your Bot! 🎉

1. Open Telegram
2. Find your bot (search for `@MovieHubStreamerbot` or your bot username)
3. Send a video file
4. Bot should reply with streaming links!

**Streaming URL format:**
- `https://yourdomain.duckdns.org/stream/ABC12345`
- `https://yourdomain.duckdns.org/file/ABC12345`

---

## Troubleshooting

### Can't SSH into VPS?
- Check firewall rules (Step 1.4)
- Verify SSH key is correct
- Try: `ssh -i ~/.ssh/your-key ubuntu@YOUR_VPS_IP`

### Domain not working?
- Wait 5-60 minutes for DNS propagation
- Check DNS: `dig yourdomain.duckdns.org`
- Verify A record points to VPS IP

### SSL certificate fails?
- Ensure DNS is pointing to your server
- Check port 80 is open
- Wait for DNS propagation

### Bot not receiving messages?
- Check bot token in `config/config.yaml`
- Verify bot is running: `sudo systemctl status bot`
- Check logs: `sudo journalctl -u bot -f`

### Services not starting?
- Check logs: `sudo journalctl -u bot -n 50`
- Verify config file is correct
- Check file permissions

---

## Quick Commands Reference

```bash
# View bot logs
sudo journalctl -u bot -f

# View server logs
sudo journalctl -u server -f

# Restart services
sudo systemctl restart bot
sudo systemctl restart server
sudo systemctl restart nginx

# Check service status
sudo systemctl status bot
sudo systemctl status server

# Check disk space
df -h
du -sh /opt/hafton-movie-bot/storage/*
```

---

## What You've Built! 🚀

✅ Unlimited file streaming (no 2GB limit)
✅ Byte-range support (seeking, rewinding)
✅ Auto-expiration (5 days)
✅ Your own domain
✅ Free hosting ($0/month)
✅ Production-ready setup

**Total setup time: ~30-40 minutes**
**Monthly cost: $0**

Enjoy your streaming bot! 🎬

