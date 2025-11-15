# Fix Bot Conflict - Multiple Instances Running

## Problem
You're seeing: `Conflict: terminated by other getUpdates request`

This means **multiple bot instances** are using the same bot token. Telegram only allows ONE active connection per bot token.

## Solution: Find and Stop All Other Instances

### Step 1: Check Local Machine

**On macOS/Linux:**
```bash
# Check for running Go processes
ps aux | grep -E "go run|cmd/bot|cmd/combined|hafton"

# If found, kill them
pkill -f "cmd/bot"
pkill -f "cmd/combined"
pkill -f "hafton-movie-bot"
```

**On Windows:**
```powershell
# Check for running processes
tasklist | findstr "go.exe"

# Kill if found
taskkill /F /IM go.exe
```

### Step 2: Check Docker

```bash
# List all containers
docker ps -a

# Stop any bot containers
docker stop $(docker ps -q --filter "name=hafton")
docker rm $(docker ps -aq --filter "name=hafton")

# Or if using docker-compose
docker-compose down
```

### Step 3: Check Other Deployments

**Railway:**
1. Go to https://railway.app
2. Check all your projects
3. Find any `hafton-streamer` or bot services
4. **Stop or delete** them

**Other VPS/Servers:**
1. SSH into each server
2. Check for running processes:
   ```bash
   ps aux | grep bot
   systemctl status bot
   ```
3. Stop them:
   ```bash
   sudo systemctl stop bot
   sudo systemctl stop hafton-movie-bot
   ```

**Local Development:**
- Make sure you're not running `go run cmd/bot/main.go` or `go run cmd/combined/main.go` locally
- Close any terminal windows running the bot

### Step 4: Verify Only Render is Running

1. Go to Render dashboard: https://dashboard.render.com
2. Check your `hafton-streamer` service
3. Make sure it's **Running** (not stopped)
4. Check logs - should see:
   ```
   ✅ Bot and server are running!
   📡 HTTP server listening on port 10000
   🤖 Telegram bot is active
   ```

### Step 5: Test

1. Send a file to your bot
2. Should work without conflicts
3. Large files (>50MB) should work now

## Quick Fix Script

Save this as `stop-all-bots.sh`:

```bash
#!/bin/bash
echo "Stopping all bot instances..."

# Kill local Go processes
pkill -f "cmd/bot" 2>/dev/null
pkill -f "cmd/combined" 2>/dev/null
pkill -f "hafton-movie-bot" 2>/dev/null

# Stop Docker containers
docker stop $(docker ps -q --filter "name=hafton") 2>/dev/null
docker-compose down 2>/dev/null

# Stop systemd services (if on Linux server)
sudo systemctl stop bot 2>/dev/null
sudo systemctl stop hafton-movie-bot 2>/dev/null

echo "Done! Only Render should be running now."
```

Run it:
```bash
chmod +x stop-all-bots.sh
./stop-all-bots.sh
```

## After Fixing

Once you've stopped all other instances:
- Only Render will handle bot requests
- No more conflicts
- Large files will work
- Instant responses

## Still Having Issues?

If conflicts persist:
1. Check Render logs - is it actually running?
2. Wait 30 seconds after stopping other instances
3. Restart Render service (in Render dashboard)
4. Test again

