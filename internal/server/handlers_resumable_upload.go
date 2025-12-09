// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/handler"
)

// setupTusHandler configures the tus.io resumable upload handler
func (s *Server) setupTusHandler() (http.Handler, error) {
	// Create uploads directory for tus
	tusUploadsDir := filepath.Join(s.config.UploadsDir, ".tus")
	if err := os.MkdirAll(tusUploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tus uploads directory: %w", err)
	}

	// Create filestore for storing upload chunks
	store := filestore.New(tusUploadsDir)

	// Create composer for upload hooks
	composer := handler.NewStoreComposer()
	store.UseIn(composer)

	// Configure tus handler
	tusHandler, err := handler.NewHandler(handler.Config{
		BasePath:              "/files/",
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tus handler: %w", err)
	}

	// Listen for upload completion events
	go s.handleTusEvents(tusHandler)

	log.Println("✅ tus.io resumable upload handler initialized")
	return tusHandler, nil
}

// handleTusEvents processes tus upload completion events
func (s *Server) handleTusEvents(tusHandler *handler.Handler) {
	for {
		select {
		case info := <-tusHandler.CompleteUploads:
			log.Printf("✅ Resumable upload completed: %s (size: %d bytes)", info.Upload.ID, info.Upload.Size)

			// Process completed upload
			go s.processTusUploadCompleted(info)
		}
	}
}

// processTusUploadCompleted moves the completed tus upload to final location and creates file entry
func (s *Server) processTusUploadCompleted(info handler.HookEvent) {
	// Get upload metadata
	upload := info.Upload
	metadata := upload.MetaData
	uploadID := upload.ID
	fileSize := upload.Size

	// Extract user ID from metadata
	userID, ok := metadata["user_id"]
	if !ok {
		log.Printf("❌ Resumable upload %s missing user_id metadata", uploadID)
		return
	}

	// Get user from database
	user, err := database.DB.GetUserByID(parseInt(userID))
	if err != nil {
		log.Printf("❌ Resumable upload %s: failed to get user %s: %v", uploadID, userID, err)
		return
	}

	// Get filename from metadata
	filename, ok := metadata["filename"]
	if !ok {
		filename = "upload-" + uploadID
	}

	// Move file from tus storage to final location
	tusFilePath := filepath.Join(s.config.UploadsDir, ".tus", uploadID)
	finalPath := filepath.Join(s.config.UploadsDir, uploadID)

	// Copy file to final location
	if err := os.Rename(tusFilePath, finalPath); err != nil {
		log.Printf("❌ Resumable upload %s: failed to move file: %v", uploadID, err)
		return
	}

	// Calculate SHA1
	sha1Hash, err := database.CalculateFileSHA1(finalPath)
	if err != nil {
		log.Printf("⚠️ Resumable upload %s: failed to calculate SHA1: %v", uploadID, err)
		sha1Hash = ""
	}

	// Get expiration from metadata
	expireAt := int64(0)
	expireAtString := ""
	if expireStr, ok := metadata["expire_date"]; ok && expireStr != "" {
		expireTime, err := time.Parse("2006-01-02", expireStr)
		if err == nil {
			expireTime = expireTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			expireAt = expireTime.Unix()
			expireAtString = expireTime.Format("2006-01-02 15:04")
		}
	}

	// Get downloads limit from metadata
	downloadsLimit := 10
	if limitStr, ok := metadata["downloads_limit"]; ok && limitStr != "" {
		downloadsLimit = parseInt(limitStr)
	}

	// Get other metadata
	requireAuth := metadata["require_auth"] == "true"
	unlimitedTime := metadata["unlimited_time"] == "true"
	unlimitedDownloads := metadata["unlimited_downloads"] == "true"
	filePassword := metadata["file_password"]
	fileComment := metadata["file_comment"]

	// Create file entry in database
	fileInfo := &database.FileInfo{
		Id:                 uploadID,
		Name:               filename,
		Size:               database.FormatFileSize(fileSize),
		SHA1:               sha1Hash,
		FilePasswordPlain:  filePassword,
		ContentType:        metadata["filetype"],
		ExpireAtString:     expireAtString,
		ExpireAt:           expireAt,
		SizeBytes:          fileSize,
		UploadDate:         time.Now().Unix(),
		DownloadsRemaining: downloadsLimit,
		DownloadCount:      0,
		UserId:             user.Id,
		Comment:            fileComment,
		UnlimitedDownloads: unlimitedDownloads,
		UnlimitedTime:      unlimitedTime,
		RequireAuth:        requireAuth,
	}

	if err := database.DB.SaveFile(fileInfo); err != nil {
		log.Printf("❌ Resumable upload %s: failed to save file metadata: %v", uploadID, err)
		os.Remove(finalPath)
		return
	}

	// Update user storage
	fileSizeMB := fileSize / (1024 * 1024)
	newStorageUsed := user.StorageUsedMB + fileSizeMB
	if err := database.DB.UpdateUserStorage(user.Id, newStorageUsed); err != nil {
		log.Printf("⚠️ Could not update user storage: %v", err)
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "FILE_UPLOADED_RESUMABLE",
		EntityType: "File",
		EntityID:   uploadID,
		Details:    fmt.Sprintf(`{"file_name":"%s","size":%d,"resumable":true}`, filename, fileSize),
		IPAddress:  metadata["client_ip"],
		UserAgent:  metadata["user_agent"],
		Success:    true,
		ErrorMsg:   "",
	})

	// Send email notification for large files (>5GB)
	fileSizeGB := float64(fileSize) / (1024 * 1024 * 1024)
	if fileSizeGB > 5.0 {
		go s.sendLargeFileUploadNotification(user, filename, fileSize, uploadID, sha1Hash)
	}

	log.Printf("✅ Resumable upload completed and processed: %s (%s) by user %d", filename, database.FormatFileSize(fileSize), user.Id)

	// Cleanup tus metadata files
	os.Remove(tusFilePath + ".info")
}

// parseInt safely parses string to int
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
