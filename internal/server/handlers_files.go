package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Frimurare/Sharecare/internal/auth"
	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/email"
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

	// Get expiration settings from form
	expireDate := r.FormValue("expire_date")
	downloadsLimit, _ := strconv.Atoi(r.FormValue("downloads_limit"))
	requireAuth := r.FormValue("require_auth") == "true"
	unlimitedTime := r.FormValue("unlimited_time") == "true"
	unlimitedDownloads := r.FormValue("unlimited_downloads") == "true"
	filePassword := r.FormValue("file_password")
	sendToEmail := r.FormValue("send_to_email")

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

	// Calculate expiration from date
	var expireAt int64
	var expireAtString string

	if !unlimitedTime && expireDate != "" {
		// Parse date from calendar (format: YYYY-MM-DD)
		expireTime, err := time.Parse("2006-01-02", expireDate)
		if err == nil {
			// Set to end of day (23:59:59)
			expireTime = expireTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			expireAt = expireTime.Unix()
			expireAtString = expireTime.Format("2006-01-02 15:04")
		} else {
			log.Printf("Warning: Could not parse expiration date '%s': %v", expireDate, err)
			// Default to 7 days if parse fails
			expireTime := time.Now().Add(7 * 24 * time.Hour)
			expireAt = expireTime.Unix()
			expireAtString = expireTime.Format("2006-01-02 15:04")
		}
	}

	// Handle downloads limit
	if unlimitedDownloads {
		downloadsLimit = 999999 // Set high value for unlimited
	} else if downloadsLimit <= 0 {
		downloadsLimit = 10 // Default to 10 if not specified
	}

	// Save file metadata to database
	fileInfo := &database.FileInfo{
		Id:                 fileID,
		Name:               header.Filename,
		Size:               database.FormatFileSize(fileSize),
		SHA1:               sha1Hash,
		FilePasswordPlain:  filePassword,
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

	// Generate share and download links
	splashLink := s.getPublicURL() + "/s/" + fileID
	downloadLink := s.getPublicURL() + "/d/" + fileID

	log.Printf("File uploaded: %s (%s) by user %d", header.Filename, database.FormatFileSize(fileSize), user.Id)

	// Send email with download link if recipient email is provided
	if sendToEmail != "" && strings.TrimSpace(sendToEmail) != "" {
		go func() {
			subject := "File ready for download"

			htmlBody := fmt.Sprintf(`
				<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
					<h2 style="color: #333;">File Shared With You</h2>
					<p>A file has been shared with you:</p>
					<div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
						<h3 style="margin-top: 0; color: #2563eb;">%s</h3>
						<p><strong>Size:</strong> %s</p>
						%s
						%s
					</div>
					<div style="margin: 30px 0;">
						<a href="%s" style="background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">View & Download File</a>
					</div>
					<p style="color: #666; font-size: 14px;">
						<strong>Direct download link:</strong> <a href="%s">%s</a>
					</p>
					<hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
					<p style="color: #999; font-size: 12px;">This file was sent to you via Sharecare.</p>
				</div>
			`, html.EscapeString(header.Filename),
				database.FormatFileSize(fileSize),
				func() string {
					if fileInfo.ExpireAtString != "" && !fileInfo.UnlimitedTime {
						return fmt.Sprintf("<p><strong>Expires:</strong> %s</p>", html.EscapeString(fileInfo.ExpireAtString))
					}
					return ""
				}(),
				func() string {
					if !fileInfo.UnlimitedDownloads {
						return fmt.Sprintf("<p><strong>Download limit:</strong> %d downloads</p>", fileInfo.DownloadsRemaining)
					}
					return ""
				}(),
				splashLink, downloadLink, downloadLink)

			textBody := fmt.Sprintf(`File Shared With You

File: %s
Size: %s
%s%s

View and download here: %s

Direct download link: %s

This file was sent to you via Sharecare.`,
				header.Filename,
				database.FormatFileSize(fileSize),
				func() string {
					if fileInfo.ExpireAtString != "" && !fileInfo.UnlimitedTime {
						return fmt.Sprintf("\nExpires: %s\n", fileInfo.ExpireAtString)
					}
					return ""
				}(),
				func() string {
					if !fileInfo.UnlimitedDownloads {
						return fmt.Sprintf("\nDownload limit: %d downloads\n", fileInfo.DownloadsRemaining)
					}
					return ""
				}(),
				splashLink, downloadLink)

			provider, err := email.GetActiveProvider(database.DB)
			if err != nil {
				log.Printf("Failed to get email provider: %v", err)
				return
			}

			err = provider.SendEmail(sendToEmail, subject, htmlBody, textBody)
			if err != nil {
				log.Printf("Failed to send file download link email to %s: %v", sendToEmail, err)
			} else {
				log.Printf("File download link email sent to %s", sendToEmail)

				// Log email to database
				err = database.DB.LogEmailSent(fileID, user.Id, sendToEmail, "", header.Filename, fileSize)
				if err != nil {
					log.Printf("Failed to log email to database: %v", err)
				}
			}
		}()
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"file_id":         fileID,
		"file_name":       header.Filename,
		"share_url":       splashLink,
		"download_url":    downloadLink,
		"size":            fileSize,
		"size_formatted":  database.FormatFileSize(fileSize),
		"expire_at":       expireAtString,
		"downloads_limit": downloadsLimit,
		"require_auth":    requireAuth,
		"has_password":    filePassword != "",
	})
}

// handleSplashPage shows the splash page with download button
func (s *Server) handleSplashPage(w http.ResponseWriter, r *http.Request) {
	// Extract file ID from URL (/s/ABC123)
	fileID := r.URL.Path[len("/s/"):]

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

	// Check if file has expired
	if !fileInfo.UnlimitedTime && fileInfo.ExpireAt > 0 && time.Now().Unix() > fileInfo.ExpireAt {
		s.renderSplashPageExpired(w, fileInfo)
		return
	}

	// Check if download limit is reached
	if !fileInfo.UnlimitedDownloads && fileInfo.DownloadsRemaining <= 0 {
		s.renderSplashPageExpired(w, fileInfo)
		return
	}

	// Render splash page
	s.renderSplashPage(w, fileInfo)
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

	// Check if this is a direct download request (from iframe redirect)
	isDirect := r.URL.Query().Get("direct") == "1"

	// If direct download and user has session, just download
	if isDirect && fileInfo.RequireAuth {
		cookie, err := r.Cookie("download_session_" + fileInfo.Id)
		if err == nil {
			account, err := database.DB.GetDownloadAccountByEmail(cookie.Value)
			if err == nil && account.IsActive {
				s.performDownload(w, r, fileInfo, account)
				return
			}
		}
	}

	// Check if file password is required
	if fileInfo.FilePasswordPlain != "" {
		s.handlePasswordProtectedDownload(w, r, fileInfo)
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

// handlePasswordProtectedDownload handles downloads that require a password
func (s *Server) handlePasswordProtectedDownload(w http.ResponseWriter, r *http.Request, fileInfo *database.FileInfo) {
	// Check if password has been verified (via session cookie)
	cookie, err := r.Cookie("password_verified_" + fileInfo.Id)
	if err == nil && cookie.Value == "true" {
		// Password already verified, check if also requires auth
		if fileInfo.RequireAuth {
			s.handleAuthenticatedDownload(w, r, fileInfo)
			return
		}
		// Just password, no auth required
		s.performDownload(w, r, fileInfo, nil)
		return
	}

	// Check if password provided via POST
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.renderPasswordPromptPage(w, fileInfo, "Invalid form data")
			return
		}

		providedPassword := r.FormValue("file_password")
		if providedPassword == "" {
			s.renderPasswordPromptPage(w, fileInfo, "Password required")
			return
		}

		// Verify password
		if providedPassword != fileInfo.FilePasswordPlain {
			s.renderPasswordPromptPage(w, fileInfo, "Incorrect password")
			return
		}

		// Password correct, set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "password_verified_" + fileInfo.Id,
			Value:    "true",
			Path:     "/d/" + fileInfo.Id,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Check if also requires authentication
		if fileInfo.RequireAuth {
			s.handleAuthenticatedDownload(w, r, fileInfo)
			return
		}

		// Just password, proceed with download
		s.performDownload(w, r, fileInfo, nil)
		return
	}

	// Show password prompt page
	s.renderPasswordPromptPage(w, fileInfo, "")
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

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		s.renderDownloadAuthPage(w, fileInfo, "Email and password required")
		return
	}

	// Check if account exists
	account, err := database.DB.GetDownloadAccountByEmail(email)
	isNewAccount := false
	if err != nil {
		// Create new account - name is required for new accounts
		if name == "" {
			s.renderDownloadAuthPage(w, fileInfo, "Name is required for new accounts")
			return
		}
		account, err = createDownloadAccount(name, email, password)
		if err != nil {
			s.renderDownloadAuthPage(w, fileInfo, "Failed to create account: "+err.Error())
			return
		}
		isNewAccount = true
		log.Printf("Download account created: %s (%s)", email, name)
	} else {
		// Verify password
		if !checkDownloadPassword(password, account.Password) {
			s.renderDownloadAuthPage(w, fileInfo, "Invalid credentials")
			return
		}
	}

	// Set file-specific download session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "download_session_" + fileInfo.Id,
		Value:    email,
		Path:     "/d/" + fileInfo.Id,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// If this is a new account, also set global download session for dashboard access
	if isNewAccount {
		log.Printf("🔐 Setting up global session for new download account: %s", email)
		sessionEmail, err := auth.CreateDownloadAccountSession(account.Id)
		if err != nil {
			log.Printf("❌ Warning: Could not create global session for new account: %v", err)
		} else {
			log.Printf("✅ Global download_session cookie set for: %s", sessionEmail)
			http.SetCookie(w, &http.Cookie{
				Name:     "download_session",
				Value:    sessionEmail,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
	}

	// For new accounts, render a special page that downloads the file and redirects to dashboard
	if isNewAccount {
		s.performDownloadWithRedirect(w, r, fileInfo, account)
	} else {
		// Perform normal download
		s.performDownload(w, r, fileInfo, account)
	}
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

	// Send email notification to file owner
	go func() {
		owner, err := database.DB.GetUserByID(fileInfo.UserId)
		if err != nil {
			log.Printf("Could not get file owner for download notification: %v", err)
			return
		}

		clientIP := getClientIP(r)
		err = email.SendFileDownloadNotification(fileInfo, clientIP, s.getPublicURL(), owner.Email)
		if err != nil {
			log.Printf("Failed to send download notification email: %v", err)
		} else {
			log.Printf("Download notification email sent to %s", owner.Email)
		}
	}()

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
			"download_url":        s.getPublicURL() + "/d/" + f.Id,
			"upload_date":         f.UploadDate,
			"expire_at":           f.ExpireAtString,
			"downloads_remaining": f.DownloadsRemaining,
			"download_count":      f.DownloadCount,
			"require_auth":        f.RequireAuth,
			"unlimited_downloads": f.UnlimitedDownloads,
			"unlimited_time":      f.UnlimitedTime,
			"has_password":        f.FilePasswordPlain != "",
			"file_password":       f.FilePasswordPlain,
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

func createDownloadAccount(name, email, password string) (*models.DownloadAccount, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	account := &models.DownloadAccount{
		Name:     name,
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
	return auth.CheckPasswordHash(password, hash)
}

func hashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

func getDownloaderInfo(account *models.DownloadAccount, ip string) string {
	if account != nil {
		return account.Email
	}
	return "anonymous (" + ip + ")"
}

// renderPasswordPromptPage renders the password prompt page for password-protected files
func (s *Server) renderPasswordPromptPage(w http.ResponseWriter, fileInfo *database.FileInfo, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Required - ` + s.config.CompanyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .password-container {
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
            color: ` + s.getPrimaryColor() + `;
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
        .password-section {
            margin-bottom: 20px;
        }
        .password-section h3 {
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
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .btn {
            width: 100%;
            padding: 14px;
            background: ` + s.getPrimaryColor() + `;
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
            background: #fff3cd;
            border: 1px solid #ffc107;
            color: #856404;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 13px;
        }
    </style>
</head>
<body>
    <div class="password-container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
        </div>

        <div class="file-info">
            <h2>🔒 ` + fileInfo.Name + `</h2>
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
            🔐 This file is password protected. Please enter the password to download.
        </div>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
	}

	html += `
        <div class="password-section">
            <form method="POST">
                <div class="form-group">
                    <label for="file_password">Password</label>
                    <input type="password" id="file_password" name="file_password" required autofocus>
                </div>
                <button type="submit" class="btn">
                    <span style="font-size: 18px; margin-right: 8px;">🔓</span>
                    <span style="font-size: 16px; font-weight: 700;">Unlock & Download</span>
                </button>
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
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
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
            color: ` + s.getPrimaryColor() + `;
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
        input[type="text"], input[type="email"], input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .btn {
            width: 100%;
            padding: 14px;
            background: ` + s.getPrimaryColor() + `;
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
                    <label for="name">Name</label>
                    <input type="text" id="name" name="name" required autofocus placeholder="Your name">
                    <p style="font-size: 12px; color: #999; margin-top: 4px;">
                        Required for new accounts only
                    </p>
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required minlength="4">
                    <p style="font-size: 12px; color: #999; margin-top: 4px;">
                        New user? Your account will be created automatically
                    </p>
                </div>
                <button type="submit" class="btn">
                    <span style="font-size: 18px; margin-right: 8px;">🔓</span>
                    <span style="font-size: 16px; font-weight: 700;">Login / Create Account & Download</span>
                </button>
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

// renderSplashPage renders the splash page with download button
func (s *Server) renderSplashPage(w http.ResponseWriter, fileInfo *database.FileInfo) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()
	companyName := brandingConfig["branding_company_name"]
	primaryColor := s.getPrimaryColor()
	secondaryColor := s.getSecondaryColor()
	logoData := brandingConfig["branding_logo"]

	downloadURL := s.getPublicURL() + "/d/" + fileInfo.Id

	// Get poem of the day
	poem := models.GetPoemOfTheDay()

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Download File - ` + companyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + primaryColor + ` 0%, ` + secondaryColor + ` 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .splash-container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px;
            max-width: 700px;
            width: 100%;
            text-align: center;
        }
        .logo {
            margin-bottom: 30px;
        }
        .logo img {
            max-width: 200px;
            max-height: 80px;
        }
        .logo h1 {
            color: ` + primaryColor + `;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .file-icon {
            font-size: 80px;
            margin-bottom: 20px;
        }
        .file-info {
            margin-bottom: 30px;
        }
        .file-info h2 {
            color: #333;
            font-size: 24px;
            margin-bottom: 10px;
            word-break: break-word;
        }
        .file-details {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin: 30px 0;
        }
        .detail-item {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 10px;
        }
        .detail-item h3 {
            color: #999;
            font-size: 12px;
            text-transform: uppercase;
            margin-bottom: 5px;
            font-weight: 500;
        }
        .detail-item p {
            color: #333;
            font-size: 18px;
            font-weight: 600;
        }
        .download-btn {
            display: inline-block;
            padding: 18px 40px;
            background: ` + primaryColor + `;
            color: white;
            text-decoration: none;
            border-radius: 10px;
            font-size: 18px;
            font-weight: 600;
            transition: all 0.3s;
            box-shadow: 0 4px 15px rgba(0,0,0,0.2);
        }
        .download-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(0,0,0,0.3);
        }
        .footer {
            margin-top: 30px;
            color: #999;
            font-size: 14px;
        }
        .badge {
            display: inline-block;
            padding: 5px 15px;
            background: #e3f2fd;
            color: #1976d2;
            border-radius: 20px;
            font-size: 13px;
            font-weight: 500;
            margin-top: 10px;
        }
        .poem-section {
            margin: 30px 0;
            padding: 25px;
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
            border-radius: 15px;
            border-left: 4px solid ` + primaryColor + `;
        }
        .poem-title {
            color: ` + primaryColor + `;
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 15px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .poem-text {
            color: #2c3e50;
            font-size: 16px;
            line-height: 1.8;
            font-style: italic;
            white-space: pre-line;
            margin-bottom: 12px;
        }
        .poem-author {
            color: #7f8c8d;
            font-size: 13px;
            font-weight: 500;
            text-align: right;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="splash-container">
        <div class="logo">`

	if logoData != "" {
		html += `<img src="` + logoData + `" alt="` + companyName + `">`
	} else {
		html += `<h1>` + companyName + `</h1>`
	}

	html += `
        </div>

        <div class="file-icon">📦</div>

        <div class="file-info">
            <h2>` + fileInfo.Name + `</h2>
        </div>

        <div class="file-details">
            <div class="detail-item">
                <h3>File Size</h3>
                <p>` + fileInfo.Size + `</p>
            </div>
            <div class="detail-item">
                <h3>Downloads</h3>
                <p>` + fmt.Sprintf("%d", fileInfo.DownloadCount) + `</p>
            </div>`

	if !fileInfo.UnlimitedDownloads {
		html += `
            <div class="detail-item">
                <h3>Remaining</h3>
                <p>` + fmt.Sprintf("%d", fileInfo.DownloadsRemaining) + `</p>
            </div>`
	}

	if fileInfo.ExpireAtString != "" && !fileInfo.UnlimitedTime {
		html += `
            <div class="detail-item">
                <h3>Expires</h3>
                <p style="font-size: 14px;">` + fileInfo.ExpireAtString + `</p>
            </div>`
	}

	html += `
        </div>`

	if fileInfo.RequireAuth {
		html += `<div class="badge">🔒 Authentication Required</div>`
	}

	// Add Poem of the Day section
	html += `
        <div class="poem-section">
            <div class="poem-title">📖 While waiting, here is Poem of the Day</div>
            <div class="poem-text">` + poem.Text + `</div>
            <div class="poem-author">— ` + poem.Author + `</div>
        </div>

        <a href="` + downloadURL + `" class="download-btn">
            <span style="font-size: 24px; margin-right: 10px;">⬇️</span>
            <span style="font-size: 20px; font-weight: 700;">Download File</span>
        </a>

        <div class="footer">
            Powered by ` + companyName + `
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderSplashPageExpired renders expired file splash page
func (s *Server) renderSplashPageExpired(w http.ResponseWriter, fileInfo *database.FileInfo) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()
	companyName := brandingConfig["branding_company_name"]
	primaryColor := s.getPrimaryColor()
	secondaryColor := s.getSecondaryColor()
	logoData := brandingConfig["branding_logo"]

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Expired - ` + companyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + primaryColor + ` 0%, ` + secondaryColor + ` 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .splash-container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px;
            max-width: 600px;
            width: 100%;
            text-align: center;
        }
        .logo {
            margin-bottom: 30px;
        }
        .logo img {
            max-width: 200px;
            max-height: 80px;
        }
        .logo h1 {
            color: ` + primaryColor + `;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .expired-icon {
            font-size: 80px;
            margin-bottom: 20px;
        }
        h2 {
            color: #f44336;
            font-size: 28px;
            margin-bottom: 15px;
        }
        p {
            color: #666;
            font-size: 16px;
            line-height: 1.6;
        }
        .footer {
            margin-top: 30px;
            color: #999;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="splash-container">
        <div class="logo">`

	if logoData != "" {
		html += `<img src="` + logoData + `" alt="` + companyName + `">`
	} else {
		html += `<h1>` + companyName + `</h1>`
	}

	html += `
        </div>

        <div class="expired-icon">⏰</div>

        <h2>File No Longer Available</h2>
        <p>This file has expired and is no longer available for download.</p>

        <div class="footer">
            Powered by ` + companyName + `
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// performDownloadWithRedirect performs a download and redirects to dashboard (for new accounts)
func (s *Server) performDownloadWithRedirect(w http.ResponseWriter, r *http.Request, fileInfo *database.FileInfo, account *models.DownloadAccount) {
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
		IsAuthenticated: true,
		DownloadAccountId: account.Id,
		Email:           account.Email,
	}

	if err := database.DB.CreateDownloadLog(downloadLog); err != nil {
		log.Printf("Warning: Could not create download log: %v", err)
	}

	// Update account last used
	database.DB.UpdateDownloadAccountLastUsed(account.Id)

	// Send email notification to file owner
	go func() {
		owner, err := database.DB.GetUserByID(fileInfo.UserId)
		if err != nil {
			log.Printf("Could not get file owner for download notification: %v", err)
			return
		}

		clientIP := getClientIP(r)
		err = email.SendFileDownloadNotification(fileInfo, clientIP, s.getPublicURL(), owner.Email)
		if err != nil {
			log.Printf("Failed to send download notification email: %v", err)
		} else {
			log.Printf("Download notification email sent to %s", owner.Email)
		}
	}()

	log.Printf("File downloaded: %s (%s) by %s (new account - redirecting to dashboard)", fileInfo.Name, fileInfo.Size, account.Email)

	// Render HTML page that downloads file and redirects to dashboard
	downloadURL := "/d/" + fileInfo.Id

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Created - ` + s.config.CompanyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .success-container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px;
            max-width: 600px;
            width: 100%;
            text-align: center;
        }
        .success-icon {
            width: 80px;
            height: 80px;
            background: #d4edda;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 30px;
            font-size: 40px;
        }
        h1 {
            color: #155724;
            margin-bottom: 20px;
            font-size: 28px;
        }
        p {
            color: #666;
            line-height: 1.8;
            margin-bottom: 15px;
            font-size: 16px;
        }
        .info-box {
            background: #e3f2fd;
            border-left: 4px solid ` + s.getPrimaryColor() + `;
            padding: 20px;
            margin: 20px 0;
            text-align: left;
        }
        .info-box p {
            margin: 8px 0;
            color: #333;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid ` + s.getPrimaryColor() + `;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        .redirect-text {
            color: #999;
            font-size: 14px;
            margin-top: 20px;
        }
    </style>
    <script>
        // Start download immediately
        window.onload = function() {
            // Create hidden iframe to trigger download
            var iframe = document.createElement('iframe');
            iframe.style.display = 'none';
            iframe.src = '` + downloadURL + `?direct=1';
            document.body.appendChild(iframe);

            // Redirect to dashboard after 3 seconds
            setTimeout(function() {
                window.location.href = '/download/dashboard';
            }, 3000);
        };
    </script>
</head>
<body>
    <div class="success-container">
        <div class="success-icon">✓</div>

        <h1>Account Created Successfully!</h1>

        <div class="info-box">
            <p><strong>✓</strong> Your download account has been created</p>
            <p><strong>✓</strong> You are now logged in</p>
            <p><strong>✓</strong> Your file download is starting...</p>
        </div>

        <p>Welcome <strong>` + account.Name + `</strong>!</p>
        <p>Your file <strong>` + fileInfo.Name + `</strong> is being downloaded.</p>

        <div class="spinner"></div>

        <p class="redirect-text">Redirecting you to your dashboard in 3 seconds...</p>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
