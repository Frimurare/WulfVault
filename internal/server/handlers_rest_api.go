// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ===========================
// ROUTE HANDLERS
// ===========================

// handleRESTUserRoutes routes user-related requests by method and path
func (s *Server) handleRESTUserRoutes(w http.ResponseWriter, r *http.Request) {
	// Extract path after /api/v1/users/
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		if r.Method == "POST" {
			s.handleAPICreateUser(w, r)
		} else {
			s.handleAPIGetUsers(w, r)
		}
		return
	}

	// Check if it's a numeric ID (user operations) or special path
	userId := parts[0]

	if len(parts) == 1 {
		// /api/v1/users/{id}
		switch r.Method {
		case "GET":
			s.handleAPIGetUser(w, r)
		case "PUT":
			s.handleAPIUpdateUser(w, r)
		case "DELETE":
			s.handleAPIDeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 {
		// /api/v1/users/{id}/{action}
		switch parts[1] {
		case "files":
			s.handleAPIGetUserFiles(w, r)
		case "storage":
			s.handleAPIGetUserStorage(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleRESTFileRoutes routes file-related requests by method and path
func (s *Server) handleRESTFileRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	fileId := parts[0]

	if len(parts) == 1 {
		// /api/v1/files/{id}
		switch r.Method {
		case "GET":
			s.handleAPIGetFile(w, r)
		case "PUT":
			s.handleAPIUpdateFile(w, r)
		case "DELETE":
			s.handleAPIDeleteFile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 {
		// /api/v1/files/{id}/{action}
		switch parts[1] {
		case "downloads":
			if r.Method == "GET" {
				s.handleAPIGetFileDownloads(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "password":
			if r.Method == "POST" {
				s.handleAPISetFilePassword(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleRESTDownloadAccountRoutes routes download account requests
func (s *Server) handleRESTDownloadAccountRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/download-accounts/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		if r.Method == "POST" {
			s.handleAPICreateDownloadAccount(w, r)
		} else {
			s.handleAPIGetDownloadAccounts(w, r)
		}
		return
	}

	accountId := parts[0]

	if len(parts) == 1 {
		// /api/v1/download-accounts/{id}
		switch r.Method {
		case "PUT":
			s.handleAPIUpdateDownloadAccount(w, r)
		case "DELETE":
			s.handleAPIDeleteDownloadAccount(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 && parts[1] == "toggle" {
		// /api/v1/download-accounts/{id}/toggle
		if r.Method == "POST" {
			s.handleAPIToggleDownloadAccount(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleRESTFileRequestRoutes routes file request operations
func (s *Server) handleRESTFileRequestRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/file-requests/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		if r.Method == "POST" {
			s.handleAPICreateFileRequest(w, r)
		} else {
			s.handleAPIGetFileRequests(w, r)
		}
		return
	}

	// Check if it's token-based access
	if parts[0] == "token" && len(parts) == 2 {
		s.handleAPIGetFileRequestByToken(w, r)
		return
	}

	// Otherwise it's ID-based
	requestId := parts[0]

	switch r.Method {
	case "PUT":
		s.handleAPIUpdateFileRequest(w, r)
	case "DELETE":
		s.handleAPIDeleteFileRequest(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRESTTrashRoutes routes trash management requests
func (s *Server) handleRESTTrashRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/trash/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	fileId := parts[0]

	if len(parts) == 1 {
		// /api/v1/trash/{id}
		if r.Method == "DELETE" {
			s.handleAPIPermanentDeleteFile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 && parts[1] == "restore" {
		// /api/v1/trash/{id}/restore
		if r.Method == "POST" {
			s.handleAPIRestoreFile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleRESTBrandingRoutes routes branding configuration requests
func (s *Server) handleRESTBrandingRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleAPIGetBranding(w, r)
	case "POST", "PUT":
		s.handleAPIUpdateBranding(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRESTSettingsRoutes routes system settings requests
func (s *Server) handleRESTSettingsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleAPIGetSettings(w, r)
	case "POST", "PUT":
		s.handleAPIUpdateSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ===========================
// USER MANAGEMENT REST API
// ===========================

// handleAPIGetUsers returns all users (Admin only)
// GET /api/v1/users
func (s *Server) handleAPIGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := database.DB.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Remove sensitive data
	for _, user := range users {
		user.Password = ""
		user.TOTPSecret = ""
		user.BackupCodes = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
		"count":   len(users),
	})
}

// handleAPIGetUser returns a specific user by ID (Admin only)
// GET /api/v1/users/{id}
func (s *Server) handleAPIGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	idStr := r.URL.Path[len("/api/v1/users/"):]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := database.DB.GetUserByID(userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Remove sensitive data
	user.Password = ""
	user.TOTPSecret = ""
	user.BackupCodes = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// handleAPICreateUser creates a new user (Admin only)
// POST /api/v1/users
func (s *Server) handleAPICreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		UserLevel      int    `json:"userLevel"`
		Permissions    int    `json:"permissions"`
		StorageQuotaMB int64  `json:"storageQuotaMB"`
		IsActive       bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Set defaults
	if req.StorageQuotaMB == 0 {
		req.StorageQuotaMB = 10240 // 10GB default
	}
	if req.UserLevel == 0 {
		req.UserLevel = models.UserLevelUser
	}

	user := &models.User{
		Name:           req.Name,
		Email:          req.Email,
		Password:       string(hashedPassword),
		UserLevel:      req.UserLevel,
		Permissions:    req.Permissions,
		StorageQuotaMB: req.StorageQuotaMB,
		IsActive:       req.IsActive,
		CreatedAt:      time.Now().Unix(),
	}

	if err := database.DB.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user (email may already exist)", http.StatusInternalServerError)
		return
	}

	// Remove sensitive data
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// handleAPIUpdateUser updates a user (Admin only)
// PUT /api/v1/users/{id}
func (s *Server) handleAPIUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	idStr := r.URL.Path[len("/api/v1/users/"):]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password,omitempty"`
		UserLevel      int    `json:"userLevel"`
		Permissions    int    `json:"permissions"`
		StorageQuotaMB int64  `json:"storageQuotaMB"`
		IsActive       bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := database.DB.GetUserByID(userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update fields
	user.Name = req.Name
	user.Email = req.Email
	user.UserLevel = req.UserLevel
	user.Permissions = req.Permissions
	user.StorageQuotaMB = req.StorageQuotaMB
	user.IsActive = req.IsActive

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := database.DB.UpdateUser(user); err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Remove sensitive data
	user.Password = ""
	user.TOTPSecret = ""
	user.BackupCodes = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// handleAPIDeleteUser deletes a user (Admin only)
// DELETE /api/v1/users/{id}
func (s *Server) handleAPIDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	idStr := r.URL.Path[len("/api/v1/users/"):]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentUser, _ := userFromContext(r.Context())

	// Prevent deleting yourself
	if userId == currentUser.Id {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	if err := database.DB.DeleteUser(userId); err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

// handleAPIGetUserFiles returns all files for a specific user (Admin only)
// GET /api/v1/users/{id}/files
func (s *Server) handleAPIGetUserFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	pathPrefix := "/api/v1/users/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	// Find first slash to extract user ID
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	idStr := remainingPath[:slashIdx]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	files, err := database.DB.GetFilesByUser(userId)
	if err != nil {
		log.Printf("Error fetching user files: %v", err)
		http.Error(w, "Error fetching files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   files,
		"count":   len(files),
	})
}

// handleAPIGetUserStorage returns storage usage for a user
// GET /api/v1/users/{id}/storage
func (s *Server) handleAPIGetUserStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	pathPrefix := "/api/v1/users/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	idStr := remainingPath[:slashIdx]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := database.DB.GetUserByID(userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	files, _ := database.DB.GetFilesByUser(userId)
	fileCount := len(files)

	percentage := 0
	if user.StorageQuotaMB > 0 {
		percentage = int((user.StorageUsedMB * 100) / user.StorageQuotaMB)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"userId":         userId,
		"storageUsedMB":  user.StorageUsedMB,
		"storageQuotaMB": user.StorageQuotaMB,
		"percentage":     percentage,
		"fileCount":      fileCount,
	})
}

// ===========================
// FILE MANAGEMENT REST API
// ===========================

// handleAPIUpdateFile updates file metadata
// PUT /api/v1/files/{id}
func (s *Server) handleAPIUpdateFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	// Extract file ID from path
	fileId := r.URL.Path[len("/api/v1/files/"):]
	if fileId == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	file, err := database.DB.GetFileByID(fileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if file.UserId != user.Id && !user.HasPermissionEditOtherUploads() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		DownloadsRemaining int    `json:"downloadsRemaining"`
		ExpireAt           int64  `json:"expireAt"`
		ExpireAtString     string `json:"expireAtString"`
		UnlimitedDownloads bool   `json:"unlimitedDownloads"`
		UnlimitedTime      bool   `json:"unlimitedTime"`
		Password           string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update file settings
	if err := database.DB.UpdateFileSettings(fileId, req.DownloadsRemaining, req.ExpireAt,
		req.ExpireAtString, req.UnlimitedDownloads, req.UnlimitedTime); err != nil {
		log.Printf("Error updating file: %v", err)
		http.Error(w, "Error updating file", http.StatusInternalServerError)
		return
	}

	// Update password if provided
	if req.Password != "" {
		if err := database.DB.UpdateFilePassword(fileId, req.Password); err != nil {
			log.Printf("Error updating file password: %v", err)
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}
	}

	// Get updated file
	file, _ = database.DB.GetFileByID(fileId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"file":    file,
	})
}

// handleAPIDeleteFile deletes a file
// DELETE /api/v1/files/{id}
func (s *Server) handleAPIDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	// Extract file ID from path
	fileId := r.URL.Path[len("/api/v1/files/"):]
	if fileId == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	file, err := database.DB.GetFileByID(fileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if file.UserId != user.Id && !user.HasPermissionEditOtherUploads() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := database.DB.DeleteFile(fileId, user.Id); err != nil {
		log.Printf("Error deleting file: %v", err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "File deleted successfully",
	})
}

// handleAPIGetFile returns file details
// GET /api/v1/files/{id}
func (s *Server) handleAPIGetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	// Extract file ID from path
	fileId := r.URL.Path[len("/api/v1/files/"):]
	if fileId == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	file, err := database.DB.GetFileByID(fileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if file.UserId != user.Id && !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"file":    file,
	})
}

// handleAPIGetFileDownloads returns download history for a file
// GET /api/v1/files/{id}/downloads
func (s *Server) handleAPIGetFileDownloads(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	// Extract file ID from path
	pathPrefix := "/api/v1/files/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fileId := remainingPath[:slashIdx]

	file, err := database.DB.GetFileByID(fileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if file.UserId != user.Id && !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	downloads, err := database.DB.GetDownloadLogsByFileID(fileId)
	if err != nil {
		log.Printf("Error fetching downloads: %v", err)
		http.Error(w, "Error fetching downloads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"downloads": downloads,
		"count":     len(downloads),
	})
}

// handleAPISetFilePassword sets or updates a file password
// POST /api/v1/files/{id}/password
func (s *Server) handleAPISetFilePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	// Extract file ID from path
	pathPrefix := "/api/v1/files/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fileId := remainingPath[:slashIdx]

	file, err := database.DB.GetFileByID(fileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if file.UserId != user.Id && !user.HasPermissionEditOtherUploads() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := database.DB.UpdateFilePassword(fileId, req.Password); err != nil {
		log.Printf("Error updating file password: %v", err)
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// ===========================
// DOWNLOAD ACCOUNTS REST API
// ===========================

// handleAPIGetDownloadAccounts returns all download accounts (Admin only)
// GET /api/v1/download-accounts
func (s *Server) handleAPIGetDownloadAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accounts, err := database.DB.GetAllDownloadAccounts()
	if err != nil {
		log.Printf("Error fetching download accounts: %v", err)
		http.Error(w, "Error fetching download accounts", http.StatusInternalServerError)
		return
	}

	// Remove passwords
	for _, acc := range accounts {
		acc.Password = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"accounts": accounts,
		"count":    len(accounts),
	})
}

// handleAPICreateDownloadAccount creates a download account (Admin only)
// POST /api/v1/download-accounts
func (s *Server) handleAPICreateDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		IsActive bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	account := &models.DownloadAccount{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		IsActive:  req.IsActive,
		CreatedAt: time.Now().Unix(),
	}

	if err := database.DB.CreateDownloadAccount(account); err != nil {
		log.Printf("Error creating download account: %v", err)
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	account.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"account": account,
	})
}

// handleAPIUpdateDownloadAccount updates a download account (Admin only)
// PUT /api/v1/download-accounts/{id}
func (s *Server) handleAPIUpdateDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/api/v1/download-accounts/"):]
	accountId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password,omitempty"`
		IsActive bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	account, err := database.DB.GetDownloadAccountByID(accountId)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	account.Name = req.Name
	account.Email = req.Email
	account.IsActive = req.IsActive

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		account.Password = string(hashedPassword)
	}

	if err := database.DB.UpdateDownloadAccount(account); err != nil {
		log.Printf("Error updating download account: %v", err)
		http.Error(w, "Error updating account", http.StatusInternalServerError)
		return
	}

	account.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"account": account,
	})
}

// handleAPIDeleteDownloadAccount deletes a download account (Admin only)
// DELETE /api/v1/download-accounts/{id}
func (s *Server) handleAPIDeleteDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/api/v1/download-accounts/"):]
	accountId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	if err := database.DB.DeleteDownloadAccount(accountId); err != nil {
		log.Printf("Error deleting download account: %v", err)
		http.Error(w, "Error deleting account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Download account deleted successfully",
	})
}

// handleAPIToggleDownloadAccount toggles download account active status (Admin only)
// POST /api/v1/download-accounts/{id}/toggle
func (s *Server) handleAPIToggleDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathPrefix := "/api/v1/download-accounts/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	idStr := remainingPath[:slashIdx]
	accountId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	account, err := database.DB.GetDownloadAccountByID(accountId)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	account.IsActive = !account.IsActive

	if err := database.DB.UpdateDownloadAccount(account); err != nil {
		log.Printf("Error toggling download account: %v", err)
		http.Error(w, "Error updating account", http.StatusInternalServerError)
		return
	}

	account.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"account": account,
	})
}

// ===========================
// FILE REQUESTS REST API
// ===========================

// handleAPIGetFileRequests returns all file requests
// GET /api/v1/file-requests
func (s *Server) handleAPIGetFileRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var requests []*models.FileRequest
	var err error

	if user.IsAdmin() {
		requests, err = database.DB.GetAllFileRequests()
	} else {
		requests, err = database.DB.GetFileRequestsByUser(user.Id)
	}

	if err != nil {
		log.Printf("Error fetching file requests: %v", err)
		http.Error(w, "Error fetching file requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"requests": requests,
		"count":    len(requests),
	})
}

// handleAPICreateFileRequest creates a new file request
// POST /api/v1/file-requests
func (s *Server) handleAPICreateFileRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		Password       string `json:"password,omitempty"`
		MaxSizeMB      int64  `json:"maxSizeMB"`
		ExpiresAt      int64  `json:"expiresAt"`
		MaxUploads     int    `json:"maxUploads"`
		NotifyOnUpload bool   `json:"notifyOnUpload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	fileRequest := &models.FileRequest{
		Title:          req.Title,
		Description:    req.Description,
		Password:       req.Password,
		MaxSizeMB:      req.MaxSizeMB,
		ExpiresAt:      req.ExpiresAt,
		MaxUploads:     req.MaxUploads,
		NotifyOnUpload: req.NotifyOnUpload,
		UserId:         user.Id,
		CreatedAt:      time.Now().Unix(),
	}

	if err := database.DB.CreateFileRequest(fileRequest); err != nil {
		log.Printf("Error creating file request: %v", err)
		http.Error(w, "Error creating file request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": fileRequest,
	})
}

// handleAPIUpdateFileRequest updates a file request
// PUT /api/v1/file-requests/{id}
func (s *Server) handleAPIUpdateFileRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	idStr := r.URL.Path[len("/api/v1/file-requests/"):]
	requestId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	fileRequest, err := database.DB.GetFileRequestByID(requestId)
	if err != nil {
		http.Error(w, "File request not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if fileRequest.UserId != user.Id && !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		Password       string `json:"password,omitempty"`
		MaxSizeMB      int64  `json:"maxSizeMB"`
		ExpiresAt      int64  `json:"expiresAt"`
		MaxUploads     int    `json:"maxUploads"`
		NotifyOnUpload bool   `json:"notifyOnUpload"`
		IsActive       bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fileRequest.Title = req.Title
	fileRequest.Description = req.Description
	fileRequest.Password = req.Password
	fileRequest.MaxSizeMB = req.MaxSizeMB
	fileRequest.ExpiresAt = req.ExpiresAt
	fileRequest.MaxUploads = req.MaxUploads
	fileRequest.NotifyOnUpload = req.NotifyOnUpload
	fileRequest.IsActive = req.IsActive

	if err := database.DB.UpdateFileRequest(fileRequest); err != nil {
		log.Printf("Error updating file request: %v", err)
		http.Error(w, "Error updating file request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": fileRequest,
	})
}

// handleAPIDeleteFileRequest deletes a file request
// DELETE /api/v1/file-requests/{id}
func (s *Server) handleAPIDeleteFileRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	idStr := r.URL.Path[len("/api/v1/file-requests/"):]
	requestId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	fileRequest, err := database.DB.GetFileRequestByID(requestId)
	if err != nil {
		http.Error(w, "File request not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if fileRequest.UserId != user.Id && !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := database.DB.DeleteFileRequest(requestId); err != nil {
		log.Printf("Error deleting file request: %v", err)
		http.Error(w, "Error deleting file request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "File request deleted successfully",
	})
}

// handleAPIGetFileRequestByToken returns a file request by token (public)
// GET /api/v1/file-requests/token/{token}
func (s *Server) handleAPIGetFileRequestByToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Path[len("/api/v1/file-requests/token/"):]
	if token == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	fileRequest, err := database.DB.GetFileRequestByToken(token)
	if err != nil {
		http.Error(w, "File request not found", http.StatusNotFound)
		return
	}

	// Don't expose password in response
	fileRequest.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": fileRequest,
	})
}

// ===========================
// TRASH MANAGEMENT REST API
// ===========================

// handleAPIGetTrash returns all deleted files (Admin only)
// GET /api/v1/trash
func (s *Server) handleAPIGetTrash(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := database.DB.GetDeletedFiles()
	if err != nil {
		log.Printf("Error fetching trash: %v", err)
		http.Error(w, "Error fetching trash", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   files,
		"count":   len(files),
	})
}

// handleAPIRestoreFile restores a file from trash (Admin only)
// POST /api/v1/trash/{id}/restore
func (s *Server) handleAPIRestoreFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathPrefix := "/api/v1/trash/"
	if len(r.URL.Path) <= len(pathPrefix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	remainingPath := r.URL.Path[len(pathPrefix):]
	slashIdx := -1
	for i, c := range remainingPath {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fileId := remainingPath[:slashIdx]

	if err := database.DB.RestoreFile(fileId); err != nil {
		log.Printf("Error restoring file: %v", err)
		http.Error(w, "Error restoring file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "File restored successfully",
	})
}

// handleAPIPermanentDeleteFile permanently deletes a file (Admin only)
// DELETE /api/v1/trash/{id}
func (s *Server) handleAPIPermanentDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileId := r.URL.Path[len("/api/v1/trash/"):]
	if fileId == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	if err := database.DB.PermanentDeleteFile(fileId); err != nil {
		log.Printf("Error permanently deleting file: %v", err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "File permanently deleted",
	})
}

// ===========================
// ADMIN/SYSTEM REST API
// ===========================

// handleAPIGetStats returns system statistics (Admin only)
// GET /api/v1/admin/stats
func (s *Server) handleAPIGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, _ := database.DB.GetAllUsers()
	files, _ := database.DB.GetAllFiles()
	deletedFiles, _ := database.DB.GetDeletedFiles()
	teams, _ := database.DB.GetAllTeams()

	// Calculate total storage
	var totalStorage int64
	for _, file := range files {
		totalStorage += file.SizeBytes
	}

	// Count active users
	activeUsers := 0
	for _, user := range users {
		if user.IsActive {
			activeUsers++
		}
	}

	// Count downloads
	totalDownloads := 0
	for _, file := range files {
		totalDownloads += file.DownloadCount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats": map[string]interface{}{
			"userCount":        len(users),
			"activeUserCount":  activeUsers,
			"fileCount":        len(files),
			"deletedFileCount": len(deletedFiles),
			"teamCount":        len(teams),
			"totalStorageBytes": totalStorage,
			"totalDownloads":   totalDownloads,
		},
	})
}

// handleAPIGetBranding returns branding configuration
// GET /api/v1/admin/branding
func (s *Server) handleAPIGetBranding(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := database.DB.GetBrandingConfig()
	if err != nil {
		log.Printf("Error fetching branding config: %v", err)
		http.Error(w, "Error fetching branding", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"branding": config,
	})
}

// handleAPIUpdateBranding updates branding configuration (Admin only)
// POST /api/v1/admin/branding
func (s *Server) handleAPIUpdateBranding(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CompanyName    string `json:"companyName"`
		PrimaryColor   string `json:"primaryColor"`
		SecondaryColor string `json:"secondaryColor"`
		LogoURL        string `json:"logoUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update branding config
	if err := database.DB.SetConfigValue("branding_company_name", req.CompanyName); err != nil {
		log.Printf("Error updating company name: %v", err)
		http.Error(w, "Error updating branding", http.StatusInternalServerError)
		return
	}
	if err := database.DB.SetConfigValue("branding_primary_color", req.PrimaryColor); err != nil {
		log.Printf("Error updating primary color: %v", err)
		http.Error(w, "Error updating branding", http.StatusInternalServerError)
		return
	}
	if err := database.DB.SetConfigValue("branding_secondary_color", req.SecondaryColor); err != nil {
		log.Printf("Error updating secondary color: %v", err)
		http.Error(w, "Error updating branding", http.StatusInternalServerError)
		return
	}
	if err := database.DB.SetConfigValue("logo_url", req.LogoURL); err != nil {
		log.Printf("Error updating logo URL: %v", err)
		http.Error(w, "Error updating branding", http.StatusInternalServerError)
		return
	}

	// Reload branding config in server
	s.loadBrandingConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Branding updated successfully",
	})
}

// handleAPIGetSettings returns system settings (Admin only)
// GET /api/v1/admin/settings
func (s *Server) handleAPIGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings := map[string]interface{}{
		"serverUrl":        s.config.ServerURL,
		"port":             s.config.Port,
		"companyName":      s.config.CompanyName,
		"maxUploadSizeMB":  s.config.MaxUploadSizeMB,
		"defaultQuotaMB":   10240,
		"trashRetentionDays": 30,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"settings": settings,
	})
}

// handleAPIUpdateSettings updates system settings (Admin only)
// POST /api/v1/admin/settings
func (s *Server) handleAPIUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update each setting in database
	for key, value := range req {
		strValue := ""
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			strValue = strconv.Itoa(v)
		case bool:
			if v {
				strValue = "true"
			} else {
				strValue = "false"
			}
		}

		if err := database.DB.SetConfigValue(key, strValue); err != nil {
			log.Printf("Error updating setting %s: %v", key, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Settings updated successfully",
	})
}
