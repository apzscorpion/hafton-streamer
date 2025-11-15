package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type FileRecord struct {
	ID             string
	TelegramFileID string
	TelegramFileURL string // For large files, we store Telegram's direct URL
	FilePath       string
	FileName       string
	FileSize       int64
	FileType       string
	UploadedAt     time.Time
	ExpiresAt      time.Time
	TelegramUserID int64
	IsProxied      bool // True if file is proxied from Telegram, false if downloaded
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		telegram_file_id TEXT NOT NULL,
		telegram_file_url TEXT,
		file_path TEXT,
		file_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		file_type TEXT NOT NULL,
		uploaded_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL,
		telegram_user_id INTEGER NOT NULL,
		is_proxied INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_expires_at ON files(expires_at);
	CREATE INDEX IF NOT EXISTS idx_telegram_file_id ON files(telegram_file_id);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return err
	}

	// Migrate existing tables - add new columns if they don't exist
	migrationQueries := []string{
		"ALTER TABLE files ADD COLUMN telegram_file_url TEXT",
		"ALTER TABLE files ADD COLUMN is_proxied INTEGER DEFAULT 0",
	}
	
	for _, migrationQuery := range migrationQueries {
		db.conn.Exec(migrationQuery) // Ignore errors if column already exists
	}
	
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) InsertFile(record *FileRecord) error {
	query := `
	INSERT INTO files (
		id, telegram_file_id, telegram_file_url, file_path, file_name, file_size, 
		file_type, uploaded_at, expires_at, telegram_user_id, is_proxied
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isProxied := 0
	if record.IsProxied {
		isProxied = 1
	}

	_, err := db.conn.Exec(
		query,
		record.ID,
		record.TelegramFileID,
		record.TelegramFileURL,
		record.FilePath,
		record.FileName,
		record.FileSize,
		record.FileType,
		record.UploadedAt.Format("2006-01-02 15:04:05"),
		record.ExpiresAt.Format("2006-01-02 15:04:05"),
		record.TelegramUserID,
		isProxied,
	)

	return err
}

func (db *DB) GetFileByID(id string) (*FileRecord, error) {
	query := `
	SELECT id, telegram_file_id, telegram_file_url, file_path, file_name, file_size, 
	       file_type, uploaded_at, expires_at, telegram_user_id, is_proxied
	FROM files
	WHERE id = ?
	`

	row := db.conn.QueryRow(query, id)
	record := &FileRecord{}

	var uploadedAt, expiresAt sql.NullString
	var isProxiedInt int
	err := row.Scan(
		&record.ID,
		&record.TelegramFileID,
		&record.TelegramFileURL,
		&record.FilePath,
		&record.FileName,
		&record.FileSize,
		&record.FileType,
		&uploadedAt,
		&expiresAt,
		&record.TelegramUserID,
		&isProxiedInt,
	)

	if err != nil {
		return nil, err
	}

	// Parse uploaded_at
	if uploadedAt.Valid && uploadedAt.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
			record.UploadedAt = t
		} else {
			return nil, fmt.Errorf("invalid uploaded_at format: %v", err)
		}
	} else {
		record.UploadedAt = time.Now()
	}

	// Parse expires_at
	if expiresAt.Valid && expiresAt.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", expiresAt.String); err == nil {
			record.ExpiresAt = t
		} else {
			return nil, fmt.Errorf("invalid expires_at format: %v", err)
		}
	} else {
		// If expires_at is NULL, set it to 5 days from now
		record.ExpiresAt = time.Now().AddDate(0, 0, 5)
	}

	record.IsProxied = isProxiedInt == 1

	return record, nil
}

func (db *DB) GetExpiredFiles() ([]*FileRecord, error) {
	query := `
	SELECT id, telegram_file_id, telegram_file_url, file_path, file_name, file_size, 
	       file_type, uploaded_at, expires_at, telegram_user_id, is_proxied
	FROM files
	WHERE expires_at < datetime('now')
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*FileRecord
	for rows.Next() {
		record := &FileRecord{}
		var uploadedAt, expiresAt sql.NullString
		var isProxiedInt int

		err := rows.Scan(
			&record.ID,
			&record.TelegramFileID,
			&record.TelegramFileURL,
			&record.FilePath,
			&record.FileName,
			&record.FileSize,
			&record.FileType,
			&uploadedAt,
			&expiresAt,
			&record.TelegramUserID,
			&isProxiedInt,
		)
		if err != nil {
			continue
		}

		// Parse uploaded_at
		if uploadedAt.Valid && uploadedAt.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
				record.UploadedAt = t
			}
		}

		// Parse expires_at
		if expiresAt.Valid && expiresAt.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", expiresAt.String); err == nil {
				record.ExpiresAt = t
			}
		}

		record.IsProxied = isProxiedInt == 1
		records = append(records, record)
	}

	return records, rows.Err()
}

func (db *DB) DeleteFile(id string) error {
	query := `DELETE FROM files WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

