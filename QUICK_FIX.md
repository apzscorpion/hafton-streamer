# Quick Fix - Bot API URL Error

## The Error

```
parse "http://hafton-streamer-2.onrender.com/bot%!(EXTRA ...)": invalid URL escape
```

## Solution

Use **HTTPS** instead of HTTP for the Bot API URL.

### Update Environment Variable

1. Go to your `hafton-streamer` service in Render
2. Environment tab
3. Update `TELEGRAM_BOT_API_URL`:
   - **Change from:** `http://hafton-streamer-2.onrender.com`
   - **Change to:** `https://hafton-streamer-2.onrender.com` ✅
4. Save → Redeploy

## Why?

Render's external URLs use HTTPS. The Bot API server is accessible via HTTPS, so use that.

## Alternative: Use Internal URL (If Same Project)

If both services are in the same Render project, you can use internal URL:
- **Value:** `http://hafton-streamer-2:8081`

But HTTPS external URL is simpler and works across projects.

