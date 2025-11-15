package cleanup

import (
	"log"
	"time"

	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/storage"
)

type Cleanup struct {
	db      *database.DB
	storage *storage.Storage
	interval time.Duration
}

func New(db *database.DB, storage *storage.Storage, interval time.Duration) *Cleanup {
	return &Cleanup{
		db:       db,
		storage:  storage,
		interval: interval,
	}
}

func (c *Cleanup) Start() {
	// Run immediately on start
	c.runCleanup()

	// Then run on interval
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.runCleanup()
	}
}

func (c *Cleanup) runCleanup() {
	log.Println("Running cleanup for expired files...")

	expiredFiles, err := c.db.GetExpiredFiles()
	if err != nil {
		log.Printf("Error fetching expired files: %v", err)
		return
	}

	if len(expiredFiles) == 0 {
		log.Println("No expired files to clean up")
		return
	}

	log.Printf("Found %d expired files to delete", len(expiredFiles))

	for _, record := range expiredFiles {
		// Delete file from storage
		if err := c.storage.DeleteFileDir(record.ID); err != nil {
			log.Printf("Error deleting file directory for %s: %v", record.ID, err)
		} else {
			log.Printf("Deleted file directory: %s", record.ID)
		}

		// Delete record from database
		if err := c.db.DeleteFile(record.ID); err != nil {
			log.Printf("Error deleting database record for %s: %v", record.ID, err)
		} else {
			log.Printf("Deleted database record: %s", record.ID)
		}
	}

	log.Printf("Cleanup completed. Deleted %d expired files", len(expiredFiles))
}

