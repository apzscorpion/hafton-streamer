package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/storage"

	"github.com/gorilla/mux"
)

type Server struct {
	db      *database.DB
	storage *storage.Storage
	config  *config.Config
	domain  string
}

func New(cfg *config.Config, db *database.DB, storage *storage.Storage, domain string) *Server {
	return &Server{
		db:      db,
		storage: storage,
		config:  cfg,
		domain:  domain,
	}
}

func (s *Server) Start() error {
	r := mux.NewRouter()

	r.HandleFunc("/stream/{id}", s.handleStream).Methods("GET")
	r.HandleFunc("/file/{id}", s.handleDownload).Methods("GET")
	r.HandleFunc("/health", s.handleHealth).Methods("GET")

	port := s.config.Server.Port
	if port == 0 {
		port = 8080
	}
	
	// Railway provides PORT env variable
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	record, err := s.db.GetFileByID(fileID)
	if err != nil {
		log.Printf("File not found in database: %s, error: %v", fileID, err)
		s.serveExpiredPage(w, r)
		return
	}

	// Check if expiration date is valid (not zero time)
	if record.ExpiresAt.IsZero() {
		log.Printf("File %s has invalid expiration date (zero time), setting to 5 days from now", fileID)
		record.ExpiresAt = time.Now().AddDate(0, 0, 5)
		// Update database with correct expiration
		// (This is a fallback - shouldn't happen if insert worked correctly)
	}

	// Check if file is expired
	now := time.Now()
	if now.After(record.ExpiresAt) {
		log.Printf("File expired: %s, expires at: %v, now: %v", fileID, record.ExpiresAt, now)
		s.serveExpiredPage(w, r)
		return
	}

	// Check if file exists (only for downloaded files, proxied files don't need this check)
	if !record.IsProxied && !s.storage.FileExists(fileID, record.FileName) {
		log.Printf("File not found on disk: %s/%s", fileID, record.FileName)
		s.serveExpiredPage(w, r)
		return
	}

	log.Printf("Serving file: %s/%s", fileID, record.FileName)
	// Serve file with byte-range support
	s.serveFileWithRange(w, r, record)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	record, err := s.db.GetFileByID(fileID)
	if err != nil {
		s.serveExpiredPage(w, r)
		return
	}

	// Check if file is expired
	if time.Now().After(record.ExpiresAt) {
		s.serveExpiredPage(w, r)
		return
	}

	// If proxied, proxy from Telegram
	if record.IsProxied && record.TelegramFileURL != "" {
		log.Printf("Proxying download from Telegram: %s", record.TelegramFileURL)
		s.proxyTelegramFile(w, r, record)
		return
	}

	// Check if file exists (only for downloaded files)
	if !record.IsProxied && !s.storage.FileExists(fileID, record.FileName) {
		s.serveExpiredPage(w, r)
		return
	}

	// Serve file for download
	filePath := s.storage.GetFilePath(fileID, record.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", record.FileType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(record.FileSize, 10))

	io.Copy(w, file)
}

func (s *Server) serveFileWithRange(w http.ResponseWriter, r *http.Request, record *database.FileRecord) {
	// If file is proxied from Telegram, proxy the request to Telegram
	if record.IsProxied && record.TelegramFileURL != "" {
		log.Printf("Proxying file from Telegram: %s", record.TelegramFileURL)
		s.proxyTelegramFile(w, r, record)
		return
	}

	// Otherwise, serve from local storage
	filePath := s.storage.GetFilePath(record.ID, record.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fileSize := fileInfo.Size()

	// Parse Range header
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// No range requested, serve entire file
		w.Header().Set("Content-Type", record.FileType)
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}

	// Parse range
	ranges := parseRange(rangeHeader, fileSize)
	if len(ranges) == 0 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Handle single range (most common case)
	start := ranges[0].start
	end := ranges[0].end

	// Set headers for partial content
	w.Header().Set("Content-Type", record.FileType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to start position
	file.Seek(start, io.SeekStart)

	// Copy requested range
	io.CopyN(w, file, end-start+1)
}

func (s *Server) proxyTelegramFile(w http.ResponseWriter, r *http.Request, record *database.FileRecord) {
	// Create request to Telegram
	req, err := http.NewRequest("GET", record.TelegramFileURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Forward Range header if present (for byte-range support)
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	// Make request to Telegram
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch file from Telegram", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers from Telegram response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set content type if not set
	if resp.Header.Get("Content-Type") == "" {
		w.Header().Set("Content-Type", record.FileType)
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Stream response
	io.Copy(w, resp.Body)
}

type byteRange struct {
	start int64
	end   int64
}

func parseRange(rangeHeader string, fileSize int64) []byteRange {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeHeader, ",")
	if len(parts) == 0 {
		return nil
	}

	var ranges []byteRange
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		dashIndex := strings.Index(part, "-")
		if dashIndex == -1 {
			continue
		}

		startStr := part[:dashIndex]
		endStr := part[dashIndex+1:]

		var start, end int64
		var err error

		if startStr == "" {
			// Suffix range: -500 means last 500 bytes
			if endStr == "" {
				continue
			}
			suffixLen, err := strconv.ParseInt(endStr, 10, 64)
			if err != nil || suffixLen <= 0 {
				continue
			}
			start = fileSize - suffixLen
			if start < 0 {
				start = 0
			}
			end = fileSize - 1
		} else if endStr == "" {
			// Prefix range: 500- means from byte 500 to end
			start, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil || start < 0 {
				continue
			}
			if start >= fileSize {
				continue
			}
			end = fileSize - 1
		} else {
			// Full range: 500-1000
			start, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil || start < 0 {
				continue
			}
			end, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil || end < start {
				continue
			}
			if start >= fileSize {
				continue
			}
			if end >= fileSize {
				end = fileSize - 1
			}
		}

		ranges = append(ranges, byteRange{start: start, end: end})
	}

	return ranges
}

func (s *Server) serveExpiredPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusGone)

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Link Expired</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            max-width: 500px;
            width: 100%;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            text-align: center;
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            margin-bottom: 15px;
            font-size: 28px;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 20px;
            font-size: 16px;
        }
        .instructions {
            background: #f5f5f5;
            border-radius: 10px;
            padding: 20px;
            margin-top: 30px;
            text-align: left;
        }
        .instructions h2 {
            color: #333;
            font-size: 18px;
            margin-bottom: 10px;
        }
        .instructions ol {
            color: #666;
            padding-left: 20px;
        }
        .instructions li {
            margin-bottom: 8px;
        }
        .bot-link {
            color: #667eea;
            text-decoration: none;
            font-weight: 600;
        }
        .bot-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⏰</div>
        <h1>Link Expired</h1>
        <p>This streaming link has expired. Files are automatically deleted after 5 days to save storage space.</p>
        <div class="instructions">
            <h2>To create a new link:</h2>
            <ol>
                <li>Open your Telegram bot</li>
                <li>Forward or upload the file again</li>
                <li>You'll receive a new streaming link</li>
            </ol>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

