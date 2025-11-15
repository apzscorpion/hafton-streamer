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
	FilePath       string
	FileName       string
	FileSize       int64
	FileType       string
	UploadedAt     time.Time
	ExpiresAt      time.Time
	TelegramUserID int64
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
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		file_type TEXT NOT NULL,
		uploaded_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL,
		telegram_user_id INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_expires_at ON files(expires_at);
	CREATE INDEX IF NOT EXISTS idx_telegram_file_id ON files(telegram_file_id);
	`

	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) InsertFile(record *FileRecord) error {
	query := `
	INSERT INTO files (
		id, telegram_file_id, file_path, file_name, file_size, 
		file_type, uploaded_at, expires_at, telegram_user_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		record.ID,
		record.TelegramFileID,
		record.FilePath,
		record.FileName,
		record.FileSize,
		record.FileType,
		record.UploadedAt.Format("2006-01-02 15:04:05"),
		record.ExpiresAt.Format("2006-01-02 15:04:05"),
		record.TelegramUserID,
	)

	return err
}

func (db *DB) GetFileByID(id string) (*FileRecord, error) {
	query := `
	SELECT id, telegram_file_id, file_path, file_name, file_size, 
	       file_type, uploaded_at, expires_at, telegram_user_id
	FROM files
	WHERE id = ?
	`

	row := db.conn.QueryRow(query, id)
	record := &FileRecord{}

	var uploadedAt, expiresAt string
	err := row.Scan(
		&record.ID,
		&record.TelegramFileID,
		&record.FilePath,
		&record.FileName,
		&record.FileSize,
		&record.FileType,
		&uploadedAt,
		&expiresAt,
		&record.TelegramUserID,
	)

	if err != nil {
		return nil, err
	}

	record.UploadedAt, _ = time.Parse("2006-01-02 15:04:05", uploadedAt)
	record.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)

	return record, nil
}

func (db *DB) GetExpiredFiles() ([]*FileRecord, error) {
	query := `
	SELECT id, telegram_file_id, file_path, file_name, file_size, 
	       file_type, uploaded_at, expires_at, telegram_user_id
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
		var uploadedAt, expiresAt string

		err := rows.Scan(
			&record.ID,
			&record.TelegramFileID,
			&record.FilePath,
			&record.FileName,
			&record.FileSize,
			&record.FileType,
			&uploadedAt,
			&expiresAt,
			&record.TelegramUserID,
		)
		if err != nil {
			continue
		}

		record.UploadedAt, _ = time.Parse("2006-01-02 15:04:05", uploadedAt)
		record.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
		records = append(records, record)
	}

	return records, rows.Err()
}

func (db *DB) DeleteFile(id string) error {
	query := `DELETE FROM files WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

