// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
)

// ChunkedUpload represents an ongoing chunked upload session
type ChunkedUpload struct {
	ID             string
	UserID         int
	Filename       string
	TotalSize      int64
	ChunksReceived int64
	File           *os.File
	LastActivity   time.Time
	Metadata       map[string]string
	mu             sync.Mutex
}

var (
	activeUploads   = make(map[string]*ChunkedUpload)
	activeUploadsMu sync.RWMutex
)

// handleChunkedUploadInit initializes a new chunked upload session
func (s *Server) handleChunkedUploadInit(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		Filename          string            `json:"filename"`
		TotalSize         int64             `json:"total_size"`
		Metadata          map[string]string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate upload ID
	uploadID := generateUploadID()

	// Create temp file for upload
	tempPath := filepath.Join(s.config.UploadsDir, ".chunks", uploadID)
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		log.Printf("Failed to create chunks directory: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	file, err := os.Create(tempPath)
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Store upload session
	upload := &ChunkedUpload{
		ID:             uploadID,
		UserID:         user.Id,
		Filename:       req.Filename,
		TotalSize:      req.TotalSize,
		ChunksReceived: 0,
		File:           file,
		LastActivity:   time.Now(),
		Metadata:       req.Metadata,
	}

	activeUploadsMu.Lock()
	activeUploads[uploadID] = upload
	activeUploadsMu.Unlock()

	log.Printf("✅ Chunked upload initialized: %s (%s, %d bytes) by user %d", uploadID, req.Filename, req.TotalSize, user.Id)

	// Return upload ID
	json.NewEncoder(w).Encode(map[string]string{
		"upload_id": uploadID,
	})
}

// handleChunkedUploadChunk receives a chunk of data
func (s *Server) handleChunkedUploadChunk(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.URL.Query().Get("upload_id")
	chunkIndexStr := r.URL.Query().Get("chunk_index")

	if uploadID == "" || chunkIndexStr == "" {
		http.Error(w, "Missing upload_id or chunk_index", http.StatusBadRequest)
		return
	}

	chunkIndex, err := strconv.ParseInt(chunkIndexStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chunk_index", http.StatusBadRequest)
		return
	}

	// Get upload session
	activeUploadsMu.RLock()
	upload, exists := activeUploads[uploadID]
	activeUploadsMu.RUnlock()

	if !exists {
		http.Error(w, "Upload session not found", http.StatusNotFound)
		return
	}

	// Verify user owns this upload
	if upload.UserID != user.Id {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Lock upload for writing
	upload.mu.Lock()
	defer upload.mu.Unlock()

	// Read chunk data
	chunkData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read chunk: %v", err)
		http.Error(w, "Failed to read chunk", http.StatusInternalServerError)
		return
	}

	// Write chunk to file
	n, err := upload.File.Write(chunkData)
	if err != nil {
		log.Printf("Failed to write chunk: %v", err)
		http.Error(w, "Failed to write chunk", http.StatusInternalServerError)
		return
	}

	upload.ChunksReceived += int64(n)
	upload.LastActivity = time.Now()

	log.Printf("📦 Chunk %d received for upload %s (%d/%d bytes)", chunkIndex, uploadID, upload.ChunksReceived, upload.TotalSize)

	// Return current status
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bytes_received": upload.ChunksReceived,
		"total_size":     upload.TotalSize,
		"complete":       upload.ChunksReceived >= upload.TotalSize,
	})
}

// handleChunkedUploadComplete finalizes the upload
func (s *Server) handleChunkedUploadComplete(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		http.Error(w, "Missing upload_id", http.StatusBadRequest)
		return
	}

	// Get upload session
	activeUploadsMu.Lock()
	upload, exists := activeUploads[uploadID]
	if exists {
		delete(activeUploads, uploadID)
	}
	activeUploadsMu.Unlock()

	if !exists {
		http.Error(w, "Upload session not found", http.StatusNotFound)
		return
	}

	// Verify user owns this upload
	if upload.UserID != user.Id {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Close temp file
	upload.File.Close()

	// Move file to final location
	tempPath := filepath.Join(s.config.UploadsDir, ".chunks", uploadID)
	finalPath := filepath.Join(s.config.UploadsDir, uploadID)

	if err := os.Rename(tempPath, finalPath); err != nil {
		log.Printf("Failed to move file: %v", err)
		http.Error(w, "Failed to finalize upload", http.StatusInternalServerError)
		return
	}

	// Calculate SHA1
	sha1Hash, err := database.CalculateFileSHA1(finalPath)
	if err != nil {
		log.Printf("Failed to calculate SHA1: %v", err)
		sha1Hash = ""
	}

	// Parse metadata
	expireAt := int64(0)
	expireAtString := ""
	if expireDate, ok := upload.Metadata["expire_date"]; ok && expireDate != "" {
		expireTime, err := time.Parse("2006-01-02", expireDate)
		if err == nil {
			expireTime = expireTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			expireAt = expireTime.Unix()
			expireAtString = expireTime.Format("2006-01-02 15:04")
		}
	}

	downloadsLimit := 10
	if limitStr, ok := upload.Metadata["downloads_limit"]; ok && limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			downloadsLimit = limit
		}
	}

	requireAuth := upload.Metadata["require_auth"] == "true"
	unlimitedTime := upload.Metadata["unlimited_time"] == "true"
	unlimitedDownloads := upload.Metadata["unlimited_downloads"] == "true"
	filePassword := upload.Metadata["file_password"]
	fileComment := upload.Metadata["file_comment"]

	// Create file entry in database
	fileInfo := &database.FileInfo{
		Id:                 uploadID,
		Name:               upload.Filename,
		Size:               database.FormatFileSize(upload.TotalSize),
		SHA1:               sha1Hash,
		FilePasswordPlain:  filePassword,
		ContentType:        upload.Metadata["filetype"],
		ExpireAtString:     expireAtString,
		ExpireAt:           expireAt,
		SizeBytes:          upload.TotalSize,
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
		log.Printf("Failed to save file metadata: %v", err)
		os.Remove(finalPath)
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	// Update user storage
	fileSizeMB := upload.TotalSize / (1024 * 1024)
	newStorageUsed := user.StorageUsedMB + fileSizeMB
	if err := database.DB.UpdateUserStorage(user.Id, newStorageUsed); err != nil {
		log.Printf("Warning: Could not update user storage: %v", err)
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "FILE_UPLOADED_CHUNKED",
		EntityType: "File",
		EntityID:   uploadID,
		Details:    fmt.Sprintf(`{"file_name":"%s","size":%d,"chunked":true}`, upload.Filename, upload.TotalSize),
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	// Send email notification for large files (>5GB)
	fileSizeGB := float64(upload.TotalSize) / (1024 * 1024 * 1024)
	if fileSizeGB > 5.0 {
		go s.sendLargeFileUploadNotification(user, upload.Filename, upload.TotalSize, uploadID, sha1Hash)
	}

	log.Printf("✅ Chunked upload completed: %s (%s) by user %d", upload.Filename, database.FormatFileSize(upload.TotalSize), user.Id)

	// Return success
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"file_id": uploadID,
	})
}

// generateUploadID generates a unique upload ID
func generateUploadID() string {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())))
	return hex.EncodeToString(hash.Sum(nil))
}

// cleanupStaleUploads removes upload sessions that have been inactive for too long
func cleanupStaleUploads() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		activeUploadsMu.Lock()
		for id, upload := range activeUploads {
			if time.Since(upload.LastActivity) > 1*time.Hour {
				upload.File.Close()
				os.Remove(filepath.Join(upload.File.Name()))
				delete(activeUploads, id)
				log.Printf("🧹 Cleaned up stale upload: %s", id)
			}
		}
		activeUploadsMu.Unlock()
	}
}

func init() {
	// Start cleanup goroutine
	go cleanupStaleUploads()
}
