package bot

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/storage"
	"hafton-movie-bot/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	db       *database.DB
	storage  *storage.Storage
	config   *config.Config
	domain   string
}

func New(cfg *config.Config, db *database.DB, storage *storage.Storage, domain string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
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

	// Get Telegram file info (this works for files up to 2GB) - FAST, no download
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: telegramFileID})
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		b.sendError(msg.Chat.ID, "Failed to get file info from Telegram")
		return
	}

	// Get Telegram file URL immediately (no download, instant response)
	telegramFileURL := file.Link(b.api.Token)
	
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

	// Store in database (async - don't wait for this)
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

	// Insert in background (don't block response)
	go func() {
		if err := b.db.InsertFile(record); err != nil {
			log.Printf("Error inserting file record: %v", err)
		}
	}()

	// Send reply IMMEDIATELY with links (don't wait for DB insert)
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

