package server

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// handleUpload handles file upload
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse multipart form (max 10GB)
	err := r.ParseMultipartForm(10 << 30)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Get expiration settings
	expirationDays, _ := strconv.Atoi(r.FormValue("expiration_days"))
	downloadsLimit, _ := strconv.Atoi(r.FormValue("downloads_limit"))
	requireAuth := r.FormValue("require_auth") == "true"

	// Check file size
	fileSize := header.Size
	fileSizeMB := fileSize / (1024 * 1024)

	// Check quota
	if !user.HasStorageSpace(fileSizeMB) {
		s.sendError(w, http.StatusBadRequest, "Insufficient storage quota")
		return
	}

	// Generate file ID
	fileID, err := generateFileID()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to generate file ID")
		return
	}

	// Save file to disk
	uploadPath := filepath.Join(s.config.UploadsDir, fileID)
	dst, err := os.Create(uploadPath)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(uploadPath)
		s.sendError(w, http.StatusInternalServerError, "Failed to write file")
		return
	}

	// Calculate expiration
	var expireAt int64
	if expirationDays > 0 {
		expireAt = time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour).Unix()
	}

	// TODO: Save file metadata to database
	_ = requireAuth
	_ = downloadsLimit
	_ = expireAt

	// Update user storage
	// database.DB.UpdateUserStorage(user.Id, user.StorageUsedMB + fileSizeMB)

	// Generate download link
	downloadLink := s.config.ServerURL + "/d/" + fileID

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"file_id":      fileID,
		"download_url": downloadLink,
		"size":         fileSize,
	})
}

// handleDownload handles file download
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	// Extract file ID from URL (/d/ABC123)
	fileID := r.URL.Path[len("/d/"):]

	if fileID == "" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// TODO: Check if file exists in database
	// TODO: Check if authentication is required
	// TODO: Check if download limit is reached
	// TODO: Create download log

	// Serve file
	filePath := filepath.Join(s.config.UploadsDir, fileID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", "attachment; filename=\"download\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

// API Handlers

// handleAPIUpload handles API file upload
func (s *Server) handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	// Reuse the same logic as handleUpload
	s.handleUpload(w, r)
}

// handleAPIFiles returns list of files for authenticated user
func (s *Server) handleAPIFiles(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// TODO: Get files from database
	_ = user

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"files": []interface{}{},
	})
}

// handleAPIDownload handles API file download
func (s *Server) handleAPIDownload(w http.ResponseWriter, r *http.Request) {
	// Reuse the same logic as handleDownload
	s.handleDownload(w, r)
}

// generateFileID generates a random file ID
func generateFileID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
