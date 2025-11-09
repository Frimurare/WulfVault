package cleanup

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
)

// CleanupExpiredFiles removes expired files from database and disk
func CleanupExpiredFiles(uploadsDir string) error {
	files, err := database.DB.GetExpiredFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	log.Printf("Cleaning up %d expired files...", len(files))

	cleaned := 0
	for _, file := range files {
		// Delete from disk
		filePath := filepath.Join(uploadsDir, file.Id)
		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: Could not delete file %s from disk: %v", file.Name, err)
			}
		}

		// Delete from database
		if err := database.DB.DeleteFile(file.Id); err != nil {
			log.Printf("Warning: Could not delete file %s from database: %v", file.Name, err)
			continue
		}

		// Recalculate user storage
		newStorage, _ := database.DB.CalculateUserStorage(file.UserId)
		database.DB.UpdateUserStorage(file.UserId, newStorage)

		cleaned++
		log.Printf("Cleaned up expired file: %s (ID: %s)", file.Name, file.Id)
	}

	log.Printf("Cleanup complete: %d files removed", cleaned)
	return nil
}

// StartCleanupScheduler starts a background cleanup scheduler
func StartCleanupScheduler(uploadsDir string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		if err := CleanupExpiredFiles(uploadsDir); err != nil {
			log.Printf("Error during initial cleanup: %v", err)
		}

		// Then run on schedule
		for range ticker.C {
			if err := CleanupExpiredFiles(uploadsDir); err != nil {
				log.Printf("Error during cleanup: %v", err)
			}
		}
	}()

	log.Printf("Cleanup scheduler started (interval: %v)", interval)
}
