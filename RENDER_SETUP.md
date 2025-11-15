# Render Setup Complete - Next Steps

Your app is live at: **https://hafton-streamer.onrender.com**

## Step 1: Set Domain Environment Variable

1. Go to your Render dashboard
2. Click on your `hafton-streamer` service
3. Go to "Environment" tab
4. Add a new environment variable:
   - **Key:** `DOMAIN`
   - **Value:** `hafton-streamer.onrender.com`
5. Click "Save Changes"
6. Render will automatically redeploy

**OR** you can also use:
   - **Key:** `RAILWAY_PUBLIC_DOMAIN` (the code checks this too)
   - **Value:** `hafton-streamer.onrender.com`

## Step 2: Verify Environment Variables

Make sure these are set in Render:
- ✅ `TELEGRAM_BOT_TOKEN` = `7529698346:AAGwnFvdpVVlmEBSCgIu61OrXnaOBWhfTVY`
- ✅ `DOMAIN` = `hafton-streamer.onrender.com` (add this)

## Step 3: Test Your Service

1. **Health Check:**
   ```
   https://hafton-streamer.onrender.com/health
   ```
   Should return: `OK`

2. **Test Streaming Endpoint:**
   - First, upload a file to your Telegram bot
   - Bot will reply with a streaming link
   - Click the link to test streaming

## Step 4: Test Your Telegram Bot

1. Open Telegram
2. Find your bot: `@MovieHubStreamerbot` or search for it
3. Send a video file or document
4. Bot should reply with:
   - Streaming link: `https://hafton-streamer.onrender.com/stream/ABC12345`
   - Download link: `https://hafton-streamer.onrender.com/file/ABC12345`

## Step 5: Keep Service Awake (Optional)

To prevent the 50-second wake-up delay:

1. Go to https://uptimerobot.com
2. Sign up (free)
3. Add a monitor:
   - **Monitor Type:** HTTP(s)
   - **Friendly Name:** Hafton Streamer
   - **URL:** `https://hafton-streamer.onrender.com/health`
   - **Monitoring Interval:** 5 minutes
4. Save

This will ping your service every 5 minutes, keeping it awake 24/7.

## Your Bot is Ready! 🎉

- ✅ Server is running
- ✅ Domain is set
- ✅ Bot token is configured
- ✅ Ready to stream files!

## Troubleshooting

**Bot not responding?**
- Check Render logs for errors
- Verify `TELEGRAM_BOT_TOKEN` is correct
- Make sure service is running (not sleeping)

**Streaming links not working?**
- Make sure `DOMAIN` environment variable is set
- Check Render logs for errors
- Verify the file was uploaded successfully

**Service keeps sleeping?**
- Normal for free tier
- Set up UptimeRobot to keep it awake
- Or accept the 50-second wake-up delay

## What's Working:

✅ HTTP server running on port 10000
✅ Cleanup system running
✅ Environment variables working
✅ Service is live and accessible

Just need to set the DOMAIN variable and you're all set!

