# Telegram Userbot Guide - For Large Files (>50MB)

## What is a Userbot?

A **Userbot** uses Telegram's **MTProto** protocol (same as the official Telegram app) instead of the Bot API. This allows:
- ✅ Files up to **2GB** (Telegram's limit)
- ✅ No 50MB restriction
- ✅ Direct file access
- ✅ **100% FREE** (no cost)

## Why Switch to Userbot?

**Current Bot API Limitations:**
- ❌ `GetFile()` fails for files >50MB
- ❌ Can't get file_path for large files
- ❌ Limited to 50MB for streaming

**Userbot Advantages:**
- ✅ Works with files up to 2GB
- ✅ Direct file access
- ✅ No API restrictions
- ✅ Same functionality, better limits

## Implementation Options

### Option 1: Go Userbot Library

**Library:** `github.com/gotd/td` (Go Telegram)

**Pros:**
- Same language (Go)
- Good performance
- Active development

**Cons:**
- More complex than Bot API
- Requires phone number verification
- Need to handle sessions

**Example:**
```go
// Using gotd/td library
import "github.com/gotd/td/telegram"

// Initialize client
client := telegram.NewClient(...)
// Get file directly - no 50MB limit!
```

### Option 2: Python Userbot (Easier)

**Library:** `Pyrogram` or `Telethon`

**Pros:**
- Easier to implement
- Better documentation
- More examples available

**Cons:**
- Different language (Python)
- Need separate service

**Example:**
```python
from pyrogram import Client

app = Client("my_account")
# Get file - works for any size!
file = await app.get_messages(chat_id, message_id)
```

### Option 3: Hybrid Approach (Recommended)

Keep your current Go bot for:
- Small files (<50MB) - instant response
- Bot commands
- Link generation

Add Python Userbot service for:
- Large files (>50MB)
- File streaming

**Architecture:**
```
Telegram → Go Bot (<50MB) → Fast response
Telegram → Python Userbot (>50MB) → Large files
Both → Same database → Same streaming links
```

## Cost

**Userbots are 100% FREE:**
- No API costs
- No subscription fees
- Uses your Telegram account
- Same as using Telegram app

## Requirements

1. **Phone Number:**
   - Need a real phone number
   - Telegram will send verification code
   - Same as creating Telegram account

2. **API Credentials:**
   - Get from https://my.telegram.org
   - Free to get
   - Takes 2 minutes

3. **Session Management:**
   - Store session files securely
   - Handle re-authentication

## Quick Start (Python Userbot)

### 1. Install Pyrogram

```bash
pip install pyrogram
```

### 2. Get API Credentials

1. Go to https://my.telegram.org
2. Login with phone number
3. Go to "API development tools"
4. Create app → Get `api_id` and `api_hash`

### 3. Create Userbot

```python
from pyrogram import Client

app = Client(
    "my_account",
    api_id=YOUR_API_ID,
    api_hash=YOUR_API_HASH
)

@app.on_message()
async def handle_message(client, message):
    if message.document:
        # Get file - works for ANY size!
        file = await client.download_media(message)
        # Process file...
```

### 4. Deploy

- Can run on same Render service
- Or separate service
- Both connect to same database

## Recommendation

**For now:** Keep current bot (works great for <50MB)

**If you need large files:**
1. Add Python Userbot service
2. Route large files to Userbot
3. Keep Go bot for small files

**Best of both worlds:**
- Fast response for small files (Go)
- Large file support (Python Userbot)
- Same database and links

## Resources

- **Pyrogram Docs:** https://docs.pyrogram.org
- **Go Telegram (gotd):** https://github.com/gotd/td
- **API Credentials:** https://my.telegram.org

## Is It Worth It?

**If most files are <50MB:** Current bot is perfect ✅

**If you need 2GB support:** Userbot is the way to go ✅

**Cost:** FREE (just need phone number) ✅

