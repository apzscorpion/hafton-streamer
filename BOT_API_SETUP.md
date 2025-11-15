# Self-Hosted Bot API Server Setup - Complete Guide

## Your API Credentials

✅ **api_id:** `33608323`  
✅ **api_hash:** `339982c3dc6fa78474ea07d77a9b0d7b`

## What This Does

Running your own Bot API server allows:
- ✅ Files up to **2GB** (no 50MB limit!)
- ✅ Same Bot API code (minimal changes)
- ✅ **100% FREE** (just hosting)

## Option 1: Docker Compose (Easiest - Recommended)

### Step 1: Update docker-compose.bot-api.yml

The file is already configured with your credentials!

### Step 2: Set Environment Variables

```bash
export TELEGRAM_BOT_TOKEN="7529698346:AAGwnFvdpVVlmEBSCgIu61OrXnaOBWhfTVY"
export DOMAIN="hafton-streamer.onrender.com"
```

### Step 3: Start Services

```bash
docker-compose -f docker-compose.bot-api.yml up -d
```

### Step 4: Verify

```bash
# Check Bot API server is running
curl http://localhost:8081/bot<Token>/getMe

# Check logs
docker-compose -f docker-compose.bot-api.yml logs -f
```

## Option 2: Render Deployment (Free Hosting)

### Step 1: Create Bot API Server Service

1. Go to Render dashboard
2. Click "New +" → "Web Service"
3. Connect your GitHub repo
4. Configure:
   - **Name:** `telegram-bot-api`
   - **Runtime:** Docker
   - **Dockerfile:** Create new file (see below)
   - **Environment Variables:**
     - `TELEGRAM_API_ID=33608323`
     - `TELEGRAM_API_HASH=339982c3dc6fa78474ea07d77a9b0d7b`
   - **Port:** `8081`
   - **Plan:** Free

### Step 2: Create Dockerfile for Bot API Server

Create `Dockerfile.bot-api`:

```dockerfile
FROM aiogram/telegram-bot-api:latest

ENV TELEGRAM_API_ID=33608323
ENV TELEGRAM_API_HASH=339982c3dc6fa78474ea07d77a9b0d7b

CMD ["--local", "--http-port=8081"]
```

### Step 3: Update Your Bot Service

In Render, update your `hafton-streamer` service:

**Add Environment Variable:**
- **Key:** `TELEGRAM_BOT_API_URL`
- **Value:** `http://telegram-bot-api:8081` (or the internal URL Render provides)

**Or use Render's internal service URL:**
- Render provides internal URLs like: `http://telegram-bot-api.onrender.com:8081`
- Use that as `TELEGRAM_BOT_API_URL`

### Step 4: Redeploy

1. Both services will redeploy
2. Bot will now use your Bot API server
3. Large files (>50MB) will work! 🎉

## Option 3: Separate VPS (Most Control)

### Step 1: Install Docker

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

### Step 2: Run Bot API Server

```bash
docker run -d \
  --name telegram-bot-api \
  --restart unless-stopped \
  -p 8081:8081 \
  -e TELEGRAM_API_ID=33608323 \
  -e TELEGRAM_API_HASH=339982c3dc6fa78474ea07d77a9b0d7b \
  -v bot-api-data:/var/lib/telegram-bot-api \
  aiogram/telegram-bot-api:latest \
  --local --http-port=8081
```

### Step 3: Update Bot Config

In your bot's `config/config.yaml`:

```yaml
telegram:
  bot_token: "7529698346:AAGwnFvdpVVlmEBSCgIu61OrXnaOBWhfTVY"
  bot_api_url: "http://your-vps-ip:8081"  # Your Bot API server URL
```

## Testing

### Test Bot API Server

```bash
# Replace <TOKEN> with your bot token
curl "http://localhost:8081/bot<TOKEN>/getMe"
```

Should return your bot info.

### Test Large File

1. Send a file >50MB to your bot
2. Should work now! ✅
3. No more "file too big" errors

## Troubleshooting

**Bot API server not starting?**
- Check API credentials are correct
- Check port 8081 is available
- Check Docker logs: `docker logs telegram-bot-api`

**Bot can't connect?**
- Verify `TELEGRAM_BOT_API_URL` is correct
- Check network connectivity between services
- On Render, use internal service URLs

**Still getting 50MB errors?**
- Make sure bot is using custom Bot API URL
- Check environment variable is set
- Restart bot service

## Benefits

✅ **Files up to 2GB** (Telegram's limit)  
✅ **No code changes** (just config)  
✅ **100% FREE** (just hosting)  
✅ **Same API** (no learning curve)

## Next Steps

1. Deploy Bot API server (choose option above)
2. Update bot to use custom URL
3. Test with large file
4. Enjoy unlimited file sizes! 🎉

