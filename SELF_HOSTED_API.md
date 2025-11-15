# Self-Hosted Telegram Bot API Server - Free Solution for Large Files

## What is This?

Run Telegram's **official Bot API server** yourself. This allows:
- ✅ Files up to **2GB** (no 50MB limit!)
- ✅ **100% FREE** (just hosting costs)
- ✅ Same Bot API code (no changes needed!)
- ✅ Works with your existing bot

## Why This Works

Telegram's public Bot API has a 50MB limit, but their **Bot API server code** supports 2GB when you run it yourself!

## Requirements

1. **VPS/Server** (can use Render, Railway, or any VPS)
2. **API Credentials** (free from https://my.telegram.org)
3. **Docker** (easiest way)

## Quick Setup (5 minutes)

### Step 1: Get API Credentials

1. Go to https://my.telegram.org
2. Login with phone number
3. Go to "API development tools"
4. Create app → Get:
   - `api_id` (number)
   - `api_hash` (string)

### Step 2: Run Bot API Server

**Option A: Docker (Easiest)**

```bash
docker run -d \
  --name telegram-bot-api \
  -p 8081:8081 \
  -v /path/to/data:/var/lib/telegram-bot-api \
  -e TELEGRAM_API_ID=YOUR_API_ID \
  -e TELEGRAM_API_HASH=YOUR_API_HASH \
  aiogram/telegram-bot-api:latest \
  --local
```

**Option B: Build from Source**

```bash
git clone https://github.com/tdlib/telegram-bot-api.git
cd telegram-bot-api
mkdir build && cd build
cmake ..
cmake --build .
./telegram-bot-api --api-id=YOUR_API_ID --api-hash=YOUR_API_HASH --local
```

### Step 3: Update Your Bot

Change your bot to use your server:

```go
// Instead of:
api, err := tgbotapi.NewBotAPI(token)

// Use:
api, err := tgbotapi.NewBotAPIWithClient(
    token,
    "http://localhost:8081", // Your Bot API server
    http.DefaultClient,
)
```

### Step 4: Deploy

**On Render:**
1. Add Bot API server as separate service
2. Update bot to use: `http://telegram-bot-api:8081`
3. Done!

## Cost

**100% FREE:**
- Bot API server code: Free (open source)
- API credentials: Free
- Hosting: Free (Render free tier works!)

**Total Cost: $0** ✅

## Benefits

- ✅ Files up to 2GB
- ✅ No code changes (just config)
- ✅ Official Telegram code
- ✅ Same API, no limits

## Resources

- **Official Repo:** https://github.com/tdlib/telegram-bot-api
- **Docker Image:** `aiogram/telegram-bot-api`
- **API Credentials:** https://my.telegram.org

## Recommendation

**This is the BEST solution:**
- Free ✅
- Easy ✅
- Works with your code ✅
- Supports 2GB files ✅

