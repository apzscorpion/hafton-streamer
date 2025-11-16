package bot

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/storage"
	"hafton-movie-bot/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// urlFixerClient wraps an HTTP client to fix malformed URLs from the telegram-bot-api library
type urlFixerClient struct {
	baseClient *http.Client
	baseURL    string
	token      string
}

func (c *urlFixerClient) Do(req *http.Request) (*http.Response, error) {
	// ALWAYS rewrite the URL to use our custom Bot API server
	// The library constructs paths like: /bot{token}/{method}
	// We need to rewrite to: {baseURL}/bot{token}/{method}
	
	// Log the original request for debugging
	log.Printf("Intercepting request: %s %s", req.Method, req.URL.String())
	
	path := req.URL.Path
	
	// Extract method from path
	// Format is usually: /bot{token}/{method}
	method := ""
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 2 {
		// Format: bot{token}/{method} - get the method (last part)
		method = parts[len(parts)-1]
	} else if len(parts) == 1 && parts[0] != "" && !strings.HasPrefix(parts[0], "bot") {
		// Just the method
		method = parts[0]
	}
	
	// Reconstruct proper URL
	// Bot API server expects: https://server.com:8081/bot{token}/{method}
	// But Render routes to port 10000, so we need to check if port is needed
	if method != "" {
		// Try with port 8081 first (Bot API server's internal port)
		// But Render might route to 10000, so we'll try without port first
		fixedURL := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
		parsedURL, err := url.Parse(fixedURL)
		if err == nil {
			req.URL = parsedURL
			log.Printf("Rewrote URL: %s -> %s", path, fixedURL)
		} else {
			log.Printf("Failed to parse fixed URL %s: %v", fixedURL, err)
			return nil, fmt.Errorf("failed to construct URL: %w", err)
		}
	} else {
		// Fallback: use base URL + original path
		fixedURL := c.baseURL + path
		parsedURL, err := url.Parse(fixedURL)
		if err == nil {
			req.URL = parsedURL
			log.Printf("Rewrote URL (fallback): %s -> %s", path, fixedURL)
		} else {
			log.Printf("Failed to parse fallback URL %s: %v", fixedURL, err)
			return nil, fmt.Errorf("failed to construct fallback URL: %w", err)
		}
	}
	
	// Log the response to debug what we're getting back
	// We'll add this after the request
	
	// Make the request
	resp, err := c.baseClient.Do(req)
	if err != nil {
		return resp, err
	}
	
	// Log ALL responses for debugging (not just errors)
	log.Printf("Bot API server returned status %d for %s", resp.StatusCode, req.URL.String())
	
	// Always log response body preview to debug issues
	if resp.Body != nil {
		// Read first 500 chars to see what we're getting
		bodyBytes := make([]byte, 500)
		n, _ := resp.Body.Read(bodyBytes)
		if n > 0 {
			preview := string(bodyBytes[:n])
			// Check if it's HTML (error page) or JSON
			if strings.HasPrefix(strings.TrimSpace(preview), "<") {
				log.Printf("⚠️ Got HTML response (error page): %s", preview)
			} else {
				log.Printf("✅ Got JSON response preview: %s", preview)
			}
		}
		// Create a new reader that includes the bytes we already read
		resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(bodyBytes[:n])), resp.Body))
	}
	
	return resp, err
}

type Bot struct {
	api      *tgbotapi.BotAPI
	db       *database.DB
	storage  *storage.Storage
	config   *config.Config
	domain   string
}

func New(cfg *config.Config, db *database.DB, storage *storage.Storage, domain string) (*Bot, error) {
	var api *tgbotapi.BotAPI
	var err error
	
	// Use custom Bot API server if configured (for large file support)
	if cfg.Telegram.BotAPIURL != "" {
		// Clean up URL - remove trailing slashes and ensure proper format
		apiEndpoint := strings.TrimSuffix(cfg.Telegram.BotAPIURL, "/")
		// Ensure URL is valid
		if !strings.HasPrefix(apiEndpoint, "http://") && !strings.HasPrefix(apiEndpoint, "https://") {
			return nil, fmt.Errorf("Bot API URL must start with http:// or https://")
		}
		
		// Note: Render's internal service discovery might not work as expected
		// For now, keep using external URL but we'll handle Cloudflare issues differently
		// If internal URL is provided directly (http://service-name:port), use it
		// Otherwise, use external URL and hope Cloudflare allows POST requests
		originalEndpoint := apiEndpoint
		
		// Check if it's already an internal URL (no .onrender.com and has port)
		if !strings.Contains(apiEndpoint, ".onrender.com") && strings.Contains(apiEndpoint, ":") {
			log.Printf("ℹ️ Using provided internal URL: %s", apiEndpoint)
		} else if strings.Contains(apiEndpoint, ".onrender.com") {
			// External URL - keep it for now since internal DNS might not work
			// The issue is Cloudflare blocking POST, but we'll try anyway
			log.Printf("⚠️ Using external URL: %s (Cloudflare may block POST requests)", apiEndpoint)
			log.Printf("💡 Tip: If you get 403 errors, both services must be in the same Render project")
			log.Printf("💡 Then use internal URL format: http://hafton-streamer-2:8081")
		}
		
		log.Printf("Using custom Bot API server: %s", apiEndpoint)
		
		// Create a custom HTTP client that intercepts requests and fixes URLs
		// This works around a bug in the library's URL construction
		baseClient := &http.Client{
			Timeout: 30 * time.Second,
		}
		
		// Wrap the client to intercept and fix URLs
		client := &urlFixerClient{
			baseClient: baseClient,
			baseURL:    apiEndpoint,
			token:      cfg.Telegram.BotToken,
		}
		
		// DON'T use SetAPIEndpoint or NewBotAPIWithAPIEndpoint - they have bugs
		// Instead, create bot with default API and override the client
		// The client will intercept and rewrite all URLs
		api, err = tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create bot API: %w", err)
		}
		
		// Set custom client with URL fixer - this will rewrite all URLs
		api.Client = client
		
		log.Printf("Using custom Bot API server: %s (URLs will be rewritten by client)", apiEndpoint)
	} else {
		// Use default Telegram Bot API
		api, err = tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create bot API: %w", err)
		}
	}

	bot := &Bot{
		api:     api,
		db:      db,
		storage: storage,
		config:  cfg,
		domain:  domain,
	}

	return bot, nil
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go b.handleMessage(update.Message)
	}

	return nil
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	// Handle forwarded messages - check original message for files
	actualMsg := msg
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		// It's a forwarded message, but we still process the current message's files
		// Telegram forwards include the file in the forwarded message
	}

	// Handle document, video, and audio messages
	var telegramFileID string
	var fileName string
	var fileSize int64
	var fileType string

	if actualMsg.Document != nil {
		telegramFileID = actualMsg.Document.FileID
		fileName = actualMsg.Document.FileName
		fileSize = int64(actualMsg.Document.FileSize)
		fileType = actualMsg.Document.MimeType
	} else if actualMsg.Video != nil {
		telegramFileID = actualMsg.Video.FileID
		fileName = actualMsg.Video.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%s.mp4", actualMsg.Video.FileID)
		}
		fileSize = int64(actualMsg.Video.FileSize)
		fileType = actualMsg.Video.MimeType
		if fileType == "" {
			fileType = "video/mp4"
		}
	} else if actualMsg.Audio != nil {
		telegramFileID = actualMsg.Audio.FileID
		fileName = actualMsg.Audio.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("audio_%s.mp3", actualMsg.Audio.FileID)
		}
		fileSize = int64(actualMsg.Audio.FileSize)
		fileType = actualMsg.Audio.MimeType
		if fileType == "" {
			fileType = "audio/mpeg"
		}
	} else {
		// Not a file message
		return
	}

	// Generate unique ID
	fileID, err := utils.GenerateID()
	if err != nil {
		b.sendError(msg.Chat.ID, "Failed to generate file ID")
		return
	}

	// Get Telegram file info - try GetFile first
	// For large files (>50MB), GetFile may fail, so we'll construct URL manually
	var telegramFileURL string
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: telegramFileID})
	if err != nil {
		// Check if error is about file being too big
		if strings.Contains(err.Error(), "too big") || strings.Contains(err.Error(), "file is too big") {
			// For large files (>50MB), Telegram's GetFile API fails
			// Unfortunately, we can't get the file_path without GetFile
			// We'll store the file_id and construct a URL that might work
			// The actual file path format varies, but we'll try common patterns
			log.Printf("GetFile failed for large file %s (%d bytes), using file_id workaround", fileName, fileSize)
			
			// Store a placeholder URL - we'll need to handle this specially
			// Format will be: file_id:<file_id> to indicate we need special handling
			telegramFileURL = fmt.Sprintf("file_id:%s", telegramFileID)
			
			log.Printf("Large file detected, stored file_id: %s", telegramFileID)
			
			// Inform user about the limitation
			b.sendError(msg.Chat.ID, fmt.Sprintf("Files larger than 50MB cannot be streamed due to Telegram Bot API limitations. Your file (%d MB) exceeds this limit. Please use files smaller than 50MB or download directly from Telegram.", fileSize/(1024*1024)))
			return
		} else {
			log.Printf("Error getting file info for %s (size: %d bytes): %v", fileName, fileSize, err)
			b.sendError(msg.Chat.ID, fmt.Sprintf("Failed to get file info: %v", err))
			return
		}
	} else {
		// Get Telegram file URL from GetFile response (works for files <50MB)
		telegramFileURL = file.Link(b.api.Token)
	}
	
	// Check file size - Telegram allows up to 2GB
	const maxTelegramSize = 2 * 1024 * 1024 * 1024 // 2GB
	if fileSize > maxTelegramSize {
		log.Printf("File %s is too large (%d bytes > %d bytes). Telegram limit is 2GB.", fileName, fileSize, maxTelegramSize)
		b.sendError(msg.Chat.ID, fmt.Sprintf("File too large (%d GB). Telegram limit is 2GB.", fileSize/(1024*1024*1024)))
		return
	}

	// Ensure file type is set
	if fileType == "" {
		ext := filepath.Ext(fileName)
		fileType = mime.TypeByExtension(ext)
		if fileType == "" {
			fileType = "application/octet-stream"
		}
	}

	// Calculate expiration
	uploadedAt := time.Now()
	expiresAt := uploadedAt.AddDate(0, 0, b.config.Retention.Days)

	// Store in database (must complete before sending links to avoid race condition)
	record := &database.FileRecord{
		ID:              fileID,
		TelegramFileID:  telegramFileID,
		TelegramFileURL: telegramFileURL,
		FilePath:        "", // Not needed for proxied files
		FileName:        fileName,
		FileSize:        fileSize,
		FileType:        fileType,
		UploadedAt:      uploadedAt,
		ExpiresAt:       expiresAt,
		TelegramUserID:  msg.From.ID,
		IsProxied:       true, // Always proxy - instant response
	}

	// Insert BEFORE sending links (must be in DB when user clicks link)
	if err := b.db.InsertFile(record); err != nil {
		log.Printf("Error inserting file record: %v", err)
		b.sendError(msg.Chat.ID, "Failed to save file metadata")
		return
	}

	log.Printf("File %s (%d bytes) ready, ID: %s, expires: %v", fileName, fileSize, fileID, expiresAt)

	// Send reply with links
	b.sendFileLinks(msg.Chat.ID, record)
}

func (b *Bot) downloadFile(fileID string) ([]byte, error) {
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		// Check if error is about file being too big
		if strings.Contains(err.Error(), "too big") || strings.Contains(err.Error(), "file is too big") {
			return nil, fmt.Errorf("file too large: Telegram Bot API limit is 50MB. Files larger than 50MB cannot be downloaded via bot API")
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := file.Link(b.api.Token)
	
	// Use http.Get with timeout for large files
	httpClient := &http.Client{
		Timeout: 30 * time.Minute, // Allow up to 30 minutes for large files
	}
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// For very large files, read in chunks to avoid memory issues
	const maxMemorySize = 100 * 1024 * 1024 // 100MB
	if resp.ContentLength > maxMemorySize {
		// Stream to temp file instead of memory
		return b.downloadLargeFile(resp.Body, resp.ContentLength)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	return data, nil
}

func (b *Bot) downloadLargeFile(reader io.Reader, contentLength int64) ([]byte, error) {
	// For files > 100MB, we still read into memory but log a warning
	// In production, you might want to stream directly to disk
	log.Printf("Downloading large file (%d bytes), this may take a while...", contentLength)
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read large file: %w", err)
	}
	
	log.Printf("Successfully downloaded large file (%d bytes)", len(data))
	return data, nil
}

func (b *Bot) sendFileLinks(chatID int64, record *database.FileRecord) {
	streamURL := fmt.Sprintf("https://%s/stream/%s", b.domain, record.ID)
	downloadURL := fmt.Sprintf("https://%s/file/%s", b.domain, record.ID)

	expiresIn := time.Until(record.ExpiresAt)
	expiresInDays := int(expiresIn.Hours() / 24)

	message := fmt.Sprintf(`🎬 Stream Ready

Play Online:
%s

Download:
%s

Type: %s
Size: %.2f GB
Valid for %d days`, 
		streamURL,
		downloadURL,
		record.FileType,
		float64(record.FileSize)/(1024*1024*1024),
		expiresInDays,
	)

	msg := tgbotapi.NewMessage(chatID, message)
	b.api.Send(msg)
}

func (b *Bot) sendError(chatID int64, errorMsg string) {
	msg := tgbotapi.NewMessage(chatID, "❌ Error: "+errorMsg)
	b.api.Send(msg)
}

