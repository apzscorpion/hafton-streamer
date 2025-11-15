# Render Setup - Add Bot API Server for Large Files

## Current Setup (What You Have Now)

You have **ONE service** on Render:
- **Service Name:** `hafton-streamer`
- **What it does:** Runs your bot + HTTP server
- **Current limitation:** Can't handle files >50MB (Telegram's public API limit)

## What We Need to Add

Add a **SECOND service** on Render:
- **Service Name:** `telegram-bot-api` (new service)
- **What it does:** Runs Telegram's Bot API server (supports 2GB files)
- **Why:** Your bot will connect to this instead of Telegram's public API

## Step-by-Step Setup

### Step 1: Create Bot API Server Service (NEW)

1. Go to Render dashboard: https://dashboard.render.com
2. Click **"New +"** → **"Web Service"**
3. **Connect your GitHub repo:** `apzscorpion/hafton-streamer` ✅ **SAME REPO!**
4. Configure the NEW service:

   **Basic Settings:**
   - **Name:** `telegram-bot-api` (or any name you like)
   - **Region:** Same as your current service
   - **Branch:** `main`
   - **Root Directory:** Leave empty
   - **Runtime:** `Docker`

   **Build & Deploy:**
   - **Dockerfile Path:** `Dockerfile.bot-api` ✅ **Different Dockerfile!**
   - **Docker Context:** Leave empty
   
   **Important:** Same repo, but different Dockerfile!
   - Your bot service uses: `Dockerfile` (builds your Go bot)
   - Bot API service uses: `Dockerfile.bot-api` (runs Telegram's Bot API server)

   **Environment Variables:**
   - `TELEGRAM_API_ID` = `33608323`
   - `TELEGRAM_API_HASH` = `339982c3dc6fa78474ea07d77a9b0d7b`

   **Advanced:**
   - **Port:** `8081`
   - **Plan:** Free

5. Click **"Create Web Service"**

### Step 2: Create Dockerfile for Bot API Server

Create a new file `Dockerfile.bot-api` in your repo:

```dockerfile
FROM aiogram/telegram-bot-api:latest

ENV TELEGRAM_API_ID=33608323
ENV TELEGRAM_API_HASH=339982c3dc6fa78474ea07d77a9b0d7b

CMD ["--local", "--http-port=8081"]
```

### Step 3: Update Your Existing Bot Service

1. Go to your **existing** `hafton-streamer` service
2. Go to **"Environment"** tab
3. Add new environment variable:
   - **Key:** `TELEGRAM_BOT_API_URL`
   - **Value:** `http://telegram-bot-api.onrender.com` 
     (Replace `telegram-bot-api` with whatever you named your new service)
4. Click **"Save Changes"**
5. Render will auto-redeploy

### Step 4: Verify It Works

1. Check logs of `telegram-bot-api` service - should see it starting
2. Check logs of `hafton-streamer` service - should see:
   ```
   Using custom Bot API server: http://telegram-bot-api.onrender.com/bot
   ```
3. Test with a large file (>50MB) - should work now! ✅

## Visual Overview

```
Before (Current):
┌─────────────────────┐
│  hafton-streamer    │
│  (Bot + Server)     │───→ Telegram Public API (50MB limit ❌)
└─────────────────────┘

After (New Setup):
┌─────────────────────┐         ┌──────────────────────┐
│  hafton-streamer    │────────→│  telegram-bot-api    │
│  (Bot + Server)     │         │  (Bot API Server)    │───→ Telegram (2GB limit ✅)
└─────────────────────┘         └──────────────────────┘
```

## Important Notes

1. **Two separate services:**
   - `hafton-streamer` = Your bot (existing)
   - `telegram-bot-api` = Bot API server (new)

2. **Both use FREE tier:**
   - Both can run on Render free tier
   - No extra cost

3. **Internal communication:**
   - Render services can talk to each other
   - Use the service URL Render provides

## Troubleshooting

**Can't find service URL?**
- In Render dashboard, click on `telegram-bot-api` service
- Look for "Internal URL" or service name
- Format: `http://telegram-bot-api.onrender.com`

**Bot still using public API?**
- Check `TELEGRAM_BOT_API_URL` is set correctly
- Check logs for "Using custom Bot API server"
- Restart `hafton-streamer` service

**Bot API server not starting?**
- Check API credentials are correct
- Check Dockerfile.bot-api exists
- Check logs for errors

## Summary

1. ✅ Keep your current `hafton-streamer` service (don't delete it!)
2. ✅ Add NEW `telegram-bot-api` service
3. ✅ Update `hafton-streamer` to use the new service
4. ✅ Test with large files

That's it! Your current service stays, we just add one more service.

