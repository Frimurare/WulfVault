package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
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

	// Calculate SHA1
	sha1Hash, err := database.CalculateFileSHA1(uploadPath)
	if err != nil {
		log.Printf("Warning: Could not calculate SHA1: %v", err)
		sha1Hash = ""
	}

	// Calculate expiration
	var expireAt int64
	var expireAtString string
	unlimitedTime := expirationDays == 0

	if expirationDays > 0 {
		expireTime := time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour)
		expireAt = expireTime.Unix()
		expireAtString = expireTime.Format("2006-01-02 15:04")
	}

	unlimitedDownloads := downloadsLimit == 0
	if downloadsLimit == 0 {
		downloadsLimit = 999999 // Set high value for unlimited
	}

	// Save file metadata to database
	fileInfo := &database.FileInfo{
		Id:                 fileID,
		Name:               header.Filename,
		Size:               database.FormatFileSize(fileSize),
		SHA1:               sha1Hash,
		ContentType:        header.Header.Get("Content-Type"),
		ExpireAtString:     expireAtString,
		ExpireAt:           expireAt,
		SizeBytes:          fileSize,
		UploadDate:         time.Now().Unix(),
		DownloadsRemaining: downloadsLimit,
		DownloadCount:      0,
		UserId:             user.Id,
		UnlimitedDownloads: unlimitedDownloads,
		UnlimitedTime:      unlimitedTime,
		RequireAuth:        requireAuth,
	}

	if err := database.DB.SaveFile(fileInfo); err != nil {
		os.Remove(uploadPath)
		s.sendError(w, http.StatusInternalServerError, "Failed to save file metadata: "+err.Error())
		return
	}

	// Update user storage
	newStorageUsed := user.StorageUsedMB + fileSizeMB
	if err := database.DB.UpdateUserStorage(user.Id, newStorageUsed); err != nil {
		log.Printf("Warning: Could not update user storage: %v", err)
	}

	// Generate download link
	downloadLink := s.config.ServerURL + "/d/" + fileID

	log.Printf("File uploaded: %s (%s) by user %d", header.Filename, database.FormatFileSize(fileSize), user.Id)

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"file_id":         fileID,
		"file_name":       header.Filename,
		"download_url":    downloadLink,
		"size":            fileSize,
		"size_formatted":  database.FormatFileSize(fileSize),
		"expire_at":       expireAtString,
		"downloads_limit": downloadsLimit,
		"require_auth":    requireAuth,
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

	// Get file from database
	fileInfo, err := database.DB.GetFileByID(fileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check if file has expired by time
	if !fileInfo.UnlimitedTime && fileInfo.ExpireAt > 0 && time.Now().Unix() > fileInfo.ExpireAt {
		http.Error(w, "File has expired", http.StatusGone)
		return
	}

	// Check if download limit is reached
	if !fileInfo.UnlimitedDownloads && fileInfo.DownloadsRemaining <= 0 {
		http.Error(w, "Download limit reached", http.StatusGone)
		return
	}

	// Check if authentication is required
	if fileInfo.RequireAuth {
		s.handleAuthenticatedDownload(w, r, fileInfo)
		return
	}

	// Direct download (no auth required)
	s.performDownload(w, r, fileInfo, nil)
}

// handleAuthenticatedDownload handles downloads that require authentication
func (s *Server) handleAuthenticatedDownload(w http.ResponseWriter, r *http.Request, fileInfo *database.FileInfo) {
	// Check if user has download session
	cookie, err := r.Cookie("download_session_" + fileInfo.Id)
	if err == nil {
		// User has session, check if valid
		account, err := database.DB.GetDownloadAccountByEmail(cookie.Value)
		if err == nil && account.IsActive {
			// Valid session, perform download
			s.performDownload(w, r, fileInfo, account)
			return
		}
	}

	// No valid session, show login/create account page
	if r.Method == http.MethodPost {
		s.handleDownloadAccountCreation(w, r, fileInfo)
		return
	}

	// Show download auth page
	s.renderDownloadAuthPage(w, fileInfo, "")
}

// handleDownloadAccountCreation handles creation of download account
func (s *Server) handleDownloadAccountCreation(w http.ResponseWriter, r *http.Request, fileInfo *database.FileInfo) {
	if err := r.ParseForm(); err != nil {
		s.renderDownloadAuthPage(w, fileInfo, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		s.renderDownloadAuthPage(w, fileInfo, "Email and password required")
		return
	}

	// Check if account exists
	account, err := database.DB.GetDownloadAccountByEmail(email)
	if err != nil {
		// Create new account
		account, err = createDownloadAccount(email, password)
		if err != nil {
			s.renderDownloadAuthPage(w, fileInfo, "Failed to create account: "+err.Error())
			return
		}
		log.Printf("Download account created: %s", email)
	} else {
		// Verify password
		if !checkDownloadPassword(password, account.Password) {
			s.renderDownloadAuthPage(w, fileInfo, "Invalid credentials")
			return
		}
	}

	// Set download session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "download_session_" + fileInfo.Id,
		Value:    email,
		Path:     "/d/" + fileInfo.Id,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Perform download
	s.performDownload(w, r, fileInfo, account)
}

// performDownload performs the actual file download
func (s *Server) performDownload(w http.ResponseWriter, r *http.Request, fileInfo *database.FileInfo, account *models.DownloadAccount) {
	filePath := filepath.Join(s.config.UploadsDir, fileInfo.Id)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	// Update download count
	if err := database.DB.UpdateFileDownloadCount(fileInfo.Id); err != nil {
		log.Printf("Warning: Could not update download count: %v", err)
	}

	// Create download log
	downloadLog := &models.DownloadLog{
		FileId:          fileInfo.Id,
		FileName:        fileInfo.Name,
		FileSize:        fileInfo.SizeBytes,
		DownloadedAt:    time.Now().Unix(),
		IpAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
		IsAuthenticated: account != nil,
	}

	if account != nil {
		downloadLog.DownloadAccountId = account.Id
		downloadLog.Email = account.Email
		// Update account last used
		database.DB.UpdateDownloadAccountLastUsed(account.Id)
	}

	if err := database.DB.CreateDownloadLog(downloadLog); err != nil {
		log.Printf("Warning: Could not create download log: %v", err)
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name))
	w.Header().Set("Content-Type", fileInfo.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.SizeBytes, 10))

	log.Printf("File downloaded: %s (%s) by %s", fileInfo.Name, fileInfo.Size, getDownloaderInfo(account, r.RemoteAddr))

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

	// Get files from database
	files, err := database.DB.GetFilesByUser(user.Id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch files")
		return
	}

	// Format files for JSON response
	var fileList []map[string]interface{}
	for _, f := range files {
		fileList = append(fileList, map[string]interface{}{
			"id":                  f.Id,
			"name":                f.Name,
			"size":                f.Size,
			"size_bytes":          f.SizeBytes,
			"download_url":        s.config.ServerURL + "/d/" + f.Id,
			"upload_date":         f.UploadDate,
			"expire_at":           f.ExpireAtString,
			"downloads_remaining": f.DownloadsRemaining,
			"download_count":      f.DownloadCount,
			"require_auth":        f.RequireAuth,
			"unlimited_downloads": f.UnlimitedDownloads,
			"unlimited_time":      f.UnlimitedTime,
		})
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"files": fileList,
		"total": len(fileList),
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

// Helper functions

func createDownloadAccount(email, password string) (*models.DownloadAccount, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	account := &models.DownloadAccount{
		Email:    email,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := database.DB.CreateDownloadAccount(account); err != nil {
		return nil, err
	}

	return account, nil
}

func checkDownloadPassword(password, hash string) bool {
	// Simple bcrypt check (reuse auth package would be better)
	return password == hash // TODO: Use proper bcrypt
}

func hashPassword(password string) (string, error) {
	// TODO: Use bcrypt from auth package
	return password, nil // Temporary
}

func getDownloaderInfo(account *models.DownloadAccount, ip string) string {
	if account != nil {
		return account.Email
	}
	return "anonymous (" + ip + ")"
}

// renderDownloadAuthPage renders the download authentication page
func (s *Server) renderDownloadAuthPage(w http.ResponseWriter, fileInfo *database.FileInfo, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Download File - ` + s.config.CompanyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + s.config.PrimaryColor + ` 0%, ` + s.config.SecondaryColor + ` 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .download-container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 500px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo h1 {
            color: ` + s.config.PrimaryColor + `;
            font-size: 28px;
            margin-bottom: 8px;
        }
        .file-info {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 24px;
        }
        .file-info h2 {
            color: #333;
            font-size: 18px;
            margin-bottom: 12px;
            word-break: break-all;
        }
        .file-info p {
            color: #666;
            font-size: 14px;
            margin: 4px 0;
        }
        .auth-section {
            margin-bottom: 20px;
        }
        .auth-section h3 {
            color: #333;
            font-size: 16px;
            margin-bottom: 16px;
        }
        .form-group {
            margin-bottom: 16px;
        }
        label {
            display: block;
            margin-bottom: 6px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="email"], input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: ` + s.config.PrimaryColor + `;
        }
        .btn {
            width: 100%;
            padding: 14px;
            background: ` + s.config.PrimaryColor + `;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: opacity 0.3s;
        }
        .btn:hover {
            opacity: 0.9;
        }
        .error {
            background: #fee;
            border: 1px solid #fcc;
            color: #c33;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .info {
            background: #e3f2fd;
            border: 1px solid #90caf9;
            color: #1976d2;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 13px;
        }
    </style>
</head>
<body>
    <div class="download-container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
        </div>

        <div class="file-info">
            <h2>📁 ` + fileInfo.Name + `</h2>
            <p><strong>Size:</strong> ` + fileInfo.Size + `</p>
            <p><strong>Downloads:</strong> ` + fmt.Sprintf("%d", fileInfo.DownloadCount) + `</p>`

	if !fileInfo.UnlimitedDownloads {
		html += `<p><strong>Remaining:</strong> ` + fmt.Sprintf("%d", fileInfo.DownloadsRemaining) + `</p>`
	}

	if fileInfo.ExpireAtString != "" {
		html += `<p><strong>Expires:</strong> ` + fileInfo.ExpireAtString + `</p>`
	}

	html += `
        </div>

        <div class="info">
            🔒 This file requires authentication. Create an account or login to download.
        </div>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
	}

	html += `
        <div class="auth-section">
            <h3>Create Account / Login</h3>
            <form method="POST">
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required autofocus>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required minlength="4">
                    <p style="font-size: 12px; color: #999; margin-top: 4px;">
                        New user? Your account will be created automatically
                    </p>
                </div>
                <button type="submit" class="btn">Download File</button>
            </form>
        </div>

        <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
            ` + s.config.FooterText + `
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
