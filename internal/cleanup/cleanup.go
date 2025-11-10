package cleanup

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
)

// CleanupExpiredFiles moves expired files to trash (soft delete)
func CleanupExpiredFiles(uploadsDir string) error {
	files, err := database.DB.GetExpiredFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	log.Printf("Moving %d expired files to trash...", len(files))

	cleaned := 0
	for _, file := range files {
		// Soft delete (move to trash) - use system user ID (0) for automated cleanup
		if err := database.DB.DeleteFile(file.Id, 0); err != nil {
			log.Printf("Warning: Could not move file %s to trash: %v", file.Name, err)
			continue
		}

		// Recalculate user storage (soft-deleted files don't count toward quota)
		newStorage, _ := database.DB.CalculateUserStorage(file.UserId)
		database.DB.UpdateUserStorage(file.UserId, newStorage)

		cleaned++
		log.Printf("Moved expired file to trash: %s (ID: %s)", file.Name, file.Id)
	}

	log.Printf("Expiration cleanup complete: %d files moved to trash", cleaned)
	return nil
}

// CleanupTrash permanently deletes files that have been in trash for 5+ days
func CleanupTrash(uploadsDir string) error {
	files, err := database.DB.GetOldDeletedFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	log.Printf("Permanently deleting %d files from trash...", len(files))

	deleted := 0
	for _, file := range files {
		// Delete from disk
		filePath := filepath.Join(uploadsDir, file.Id)
		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: Could not delete file %s from disk: %v", file.Name, err)
			}
		}

		// Permanently delete from database
		if err := database.DB.PermanentDeleteFile(file.Id); err != nil {
			log.Printf("Warning: Could not delete file %s from database: %v", file.Name, err)
			continue
		}

		deleted++
		log.Printf("Permanently deleted file: %s (ID: %s)", file.Name, file.Id)
	}

	log.Printf("Trash cleanup complete: %d files permanently deleted", deleted)
	return nil
}

// StartCleanupScheduler starts a background cleanup scheduler
func StartCleanupScheduler(uploadsDir string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		if err := CleanupExpiredFiles(uploadsDir); err != nil {
			log.Printf("Error during expired files cleanup: %v", err)
		}
		if err := CleanupTrash(uploadsDir); err != nil {
			log.Printf("Error during trash cleanup: %v", err)
		}

		// Then run on schedule
		for range ticker.C {
			if err := CleanupExpiredFiles(uploadsDir); err != nil {
				log.Printf("Error during expired files cleanup: %v", err)
			}
			if err := CleanupTrash(uploadsDir); err != nil {
				log.Printf("Error during trash cleanup: %v", err)
			}
		}
	}()

	log.Printf("Cleanup scheduler started (interval: %v)", interval)
}
