package bot

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
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
	// Handle document, video, and audio messages
	var telegramFileID string
	var fileName string
	var fileSize int64
	var fileType string

	if msg.Document != nil {
		telegramFileID = msg.Document.FileID
		fileName = msg.Document.FileName
		fileSize = int64(msg.Document.FileSize)
		fileType = msg.Document.MimeType
	} else if msg.Video != nil {
		telegramFileID = msg.Video.FileID
		fileName = msg.Video.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%s.mp4", msg.Video.FileID)
		}
		fileSize = int64(msg.Video.FileSize)
		fileType = msg.Video.MimeType
		if fileType == "" {
			fileType = "video/mp4"
		}
	} else if msg.Audio != nil {
		telegramFileID = msg.Audio.FileID
		fileName = msg.Audio.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("audio_%s.mp3", msg.Audio.FileID)
		}
		fileSize = int64(msg.Audio.FileSize)
		fileType = msg.Audio.MimeType
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

	// Download file from Telegram
	log.Printf("Downloading file %s (size: %d bytes)", fileName, fileSize)
	fileData, err := b.downloadFile(telegramFileID)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		b.sendError(msg.Chat.ID, "Failed to download file from Telegram")
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

	// Save file to storage
	if err := b.storage.SaveFile(fileID, fileName, fileData); err != nil {
		log.Printf("Error saving file: %v", err)
		b.sendError(msg.Chat.ID, "Failed to save file")
		return
	}

	// Calculate expiration
	uploadedAt := time.Now()
	expiresAt := uploadedAt.AddDate(0, 0, b.config.Retention.Days)

	// Store in database
	record := &database.FileRecord{
		ID:             fileID,
		TelegramFileID: telegramFileID,
		FilePath:       b.storage.GetFilePath(fileID, fileName),
		FileName:       fileName,
		FileSize:       int64(len(fileData)),
		FileType:       fileType,
		UploadedAt:     uploadedAt,
		ExpiresAt:      expiresAt,
		TelegramUserID: msg.From.ID,
	}

	if err := b.db.InsertFile(record); err != nil {
		log.Printf("Error inserting file record: %v", err)
		b.sendError(msg.Chat.ID, "Failed to save file metadata")
		return
	}

	// Send reply with links
	b.sendFileLinks(msg.Chat.ID, record)
}

func (b *Bot) downloadFile(fileID string) ([]byte, error) {
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := file.Link(b.api.Token)
	
	// Use http.Get directly since api.Client might not have Get method
	httpClient := &http.Client{}
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

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

