# Deploy to Render.com (Free Forever)

Railway trial expired? No problem! Render.com is free forever and perfect for your bot.

## Quick Deploy to Render (5 minutes)

### Step 1: Sign Up
1. Go to https://render.com
2. Click "Get Started for Free"
3. Sign up with GitHub (same account you used for Railway)
4. Verify your email

### Step 2: Create Web Service
1. In Render dashboard, click "New +"
2. Select "Web Service"
3. Connect your GitHub account (if not already connected)
4. Select repository: `apzscorpion/hafton-streamer`
5. Click "Connect"

### Step 3: Configure Service
Fill in these settings:

**Name:** `hafton-streamer` (or any name you like)

**Region:** Choose closest to you (e.g., `Oregon (US West)`)

**Branch:** `main`

**Root Directory:** Leave empty (default)

**Runtime:** `Docker`

**Build Command:** Leave empty (uses Dockerfile)

**Start Command:** Leave empty (uses Dockerfile CMD)

**Plan:** Select **Free** (this is the free tier!)

### Step 4: Add Environment Variables
Click "Advanced" → "Add Environment Variable"

Add this variable:
- **Key:** `TELEGRAM_BOT_TOKEN`
- **Value:** `7529698346:AAGwnFvdpVVlmEBSCgIu61OrXnaOBWhfTVY`

Click "Add"

### Step 5: Deploy!
1. Scroll down and click "Create Web Service"
2. Render will start building your Docker image
3. Wait 5-10 minutes for build to complete
4. Your service will be live at: `https://hafton-streamer.onrender.com`

### Step 6: Get Your Domain
1. Once deployed, go to your service settings
2. Find "Custom Domain" section
3. Your free domain is: `hafton-streamer.onrender.com`
4. Copy this domain

### Step 7: Update Bot Domain (Optional)
If you want to use the domain in your bot responses, add another environment variable:
- **Key:** `DOMAIN`
- **Value:** `hafton-streamer.onrender.com`

Then redeploy.

## Important Notes

### Free Tier Limitations:
- **Sleeps after 15 minutes** of inactivity
- **Wakes automatically** when someone visits (takes ~30 seconds)
- **750 hours/month** free (enough for always-on if you get traffic)
- **512MB RAM** (plenty for your bot)

### For Always-On (Optional):
If you want it to never sleep, you can:
1. Set up a free uptime monitor (like UptimeRobot)
2. Ping your service every 10 minutes
3. Keeps it awake 24/7

## Your Bot is Ready!

Once deployed:
- Your streaming server will be at: `https://hafton-streamer.onrender.com`
- Bot will work the same way
- Files expire after 5 days (as configured)
- **Completely free forever!**

## Troubleshooting

**Service keeps sleeping?**
- Normal for free tier
- It wakes automatically when accessed
- Or set up uptime monitor

**Build fails?**
- Check build logs in Render dashboard
- Make sure Dockerfile is correct
- Verify environment variables are set

**Bot not working?**
- Check `TELEGRAM_BOT_TOKEN` is set correctly
- Check service logs in Render dashboard
- Verify service is running (not sleeping)

## Next Steps

1. Deploy to Render (follow steps above)
2. Test your bot by sending a file
3. Get streaming links
4. Enjoy your free streaming bot! 🎉

