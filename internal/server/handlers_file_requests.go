// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
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

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// RemoteAddr includes port, strip it
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// handleFileRequestCreate creates a new file upload request
func (s *Server) handleFileRequestCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse multipart form (since FormData sends multipart)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// Fallback to regular form parsing
		if err := r.ParseForm(); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
	}

	title := r.FormValue("title")
	message := r.FormValue("message")
	maxFileSizeMB, _ := strconv.Atoi(r.FormValue("max_file_size_mb"))
	allowedFileTypes := r.FormValue("allowed_file_types")
	recipientEmail := r.FormValue("recipient_email")
	// Note: expires_in_days is for uploaded files, not the request link itself

	// Debug logging
	log.Printf("File request params: title='%s', message='%s', sizeMB=%d", title, message, maxFileSizeMB)

	if title == "" {
		s.sendError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Upload request link ALWAYS expires after 24 hours
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Convert MB to bytes for storage
	maxFileSize := int64(maxFileSizeMB) * 1024 * 1024

	fileRequest := &models.FileRequest{
		UserId:           user.Id,
		Title:            title,
		Message:          message,
		ExpiresAt:        expiresAt,
		IsActive:         true,
		MaxFileSize:      maxFileSize,
		AllowedFileTypes: allowedFileTypes,
	}

	if err := database.DB.CreateFileRequest(fileRequest); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to create file request: "+err.Error())
		return
	}

	uploadURL := fileRequest.GetUploadURL(s.getPublicURL())

	// Send invitation email if recipient email is provided
	if recipientEmail != "" && strings.TrimSpace(recipientEmail) != "" {
		go func() {
			expireTime := time.Unix(fileRequest.ExpiresAt, 0).Format("2006-01-02 15:04")
			subject := "Action Required: Please upload your file"

			// Get branding for company name
			brandingConfig, _ := database.DB.GetBrandingConfig()
			companyName := brandingConfig["branding_company_name"]
			if companyName == "" {
				companyName = s.config.CompanyName
			}

			htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
</head>
<body style="margin: 0; padding: 0; font-family: Arial, Helvetica, sans-serif;">
	<table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f0f0f0; padding: 20px 0;">
		<tr>
			<td align="center">
				<table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
					<!-- Header -->
					<tr>
						<td style="background-color: #1e3a5f; padding: 30px; text-align: center;">
							<h1 style="color: #ffffff; margin: 0; font-size: 24px;">%s</h1>
							<p style="color: #a0c4e8; margin: 10px 0 0 0; font-size: 14px;">Secure File Transfer</p>
						</td>
					</tr>

					<!-- Main Content -->
					<tr>
						<td style="padding: 40px 30px;">
							<!-- What is this -->
							<div style="background-color: #e8f4fd; border-left: 4px solid #2563eb; padding: 15px; margin-bottom: 25px;">
								<p style="margin: 0; color: #1e3a5f; font-size: 16px;">
									<strong>What is this?</strong><br>
									Someone needs you to upload a file to them securely. Click the green button below to upload your file.
								</p>
							</div>

							<!-- Request Title -->
							<h2 style="color: #1e3a5f; margin: 0 0 15px 0; font-size: 20px;">📋 %s</h2>
							%s

							<!-- BIG GREEN UPLOAD BUTTON -->
							<table width="100%%" cellpadding="0" cellspacing="0" style="margin: 30px 0;">
								<tr>
									<td align="center">
										<a href="%s" style="display: inline-block; background-color: #16a34a; color: #ffffff; padding: 20px 50px; text-decoration: none; border-radius: 8px; font-size: 20px; font-weight: bold; border: 3px solid #15803d; box-shadow: 0 4px 12px rgba(22, 163, 74, 0.4); text-transform: uppercase; letter-spacing: 1px;">
											⬆️ UPLOAD FILE HERE
										</a>
									</td>
								</tr>
							</table>

							<!-- Expiration Warning -->
							<div style="background-color: #fef3c7; border: 2px solid #f59e0b; border-radius: 8px; padding: 20px; margin: 25px 0; text-align: center;">
								<p style="margin: 0; color: #92400e; font-size: 16px; font-weight: bold;">
									⏰ IMPORTANT: This link expires
								</p>
								<p style="margin: 10px 0 0 0; color: #78350f; font-size: 18px; font-weight: bold;">
									%s
								</p>
							</div>

							<!-- Backup Link -->
							<div style="background-color: #f3f4f6; padding: 15px; border-radius: 6px; margin-top: 20px;">
								<p style="margin: 0 0 8px 0; color: #374151; font-size: 12px;">
									<strong>If the button doesn't work, copy this link:</strong>
								</p>
								<p style="margin: 0; word-break: break-all; font-size: 11px;">
									<a href="%s" style="color: #2563eb;">%s</a>
								</p>
							</div>
						</td>
					</tr>

					<!-- Footer -->
					<tr>
						<td style="background-color: #1e3a5f; padding: 20px; text-align: center;">
							<p style="margin: 0; color: #a0c4e8; font-size: 12px;">
								This is an automated message from %s
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
			`, companyName,
				html.EscapeString(title),
				func() string {
					if message != "" {
						return fmt.Sprintf(`<p style="color: #374151; font-size: 15px; line-height: 1.6; margin: 0 0 15px 0;">%s</p>`, html.EscapeString(message))
					}
					return ""
				}(),
				uploadURL, expireTime, uploadURL, uploadURL, companyName)

			textBody := fmt.Sprintf(`ACTION REQUIRED: Please Upload Your File
============================================

WHAT IS THIS?
Someone needs you to upload a file to them securely.

REQUEST: %s
%s
UPLOAD YOUR FILE HERE:
%s

⚠️ IMPORTANT: This link expires on %s

---
This is an automated message from %s`,
				title,
				func() string {
					if message != "" {
						return fmt.Sprintf("\nMESSAGE: %s\n", message)
					}
					return ""
				}(),
				uploadURL, expireTime, companyName)

			provider, err := email.GetActiveProvider(database.DB)
			if err != nil {
				log.Printf("Failed to get email provider: %v", err)
				return
			}

			err = provider.SendEmail(recipientEmail, subject, htmlBody, textBody)
			if err != nil {
				log.Printf("Failed to send file request invitation email to %s: %v", recipientEmail, err)
			} else {
				log.Printf("File request invitation email sent to %s", recipientEmail)
			}
		}()
	}

	log.Printf("File request created: %s by user %d", title, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"id":            fileRequest.Id,
		"title":         fileRequest.Title,
		"request_token": fileRequest.RequestToken,
		"upload_url":    uploadURL,
		"expires_at":    fileRequest.ExpiresAt,
	})
}

// handleFileRequestList returns all file requests for the authenticated user
func (s *Server) handleFileRequestList(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	requests, err := database.DB.GetFileRequestsByUser(user.Id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch requests")
		return
	}

	now := time.Now().Unix()
	fiveDaysAgo := now - (5 * 24 * 60 * 60)

	var requestList []map[string]interface{}
	for _, req := range requests {
		// Skip used requests (single-use links that have been consumed)
		if req.IsUsed() {
			continue
		}

		// Skip requests that have been expired for more than 5 days
		if req.IsExpired() && req.ExpiresAt < fiveDaysAgo {
			continue
		}

		requestList = append(requestList, map[string]interface{}{
			"id":                 req.Id,
			"title":              req.Title,
			"message":            req.Message,
			"request_token":      req.RequestToken,
			"upload_url":         req.GetUploadURL(s.getPublicURL()),
			"created_at":         req.CreatedAt,
			"expires_at":         req.ExpiresAt,
			"is_active":          req.IsActive,
			"is_expired":         req.IsExpired(),
			"max_file_size_mb":   req.MaxFileSize / (1024 * 1024),
			"allowed_file_types": req.AllowedFileTypes,
		})
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"requests": requestList,
		"total":    len(requestList),
	})
}

// handleFileRequestDelete deletes a file request
func (s *Server) handleFileRequestDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse multipart form (since FormData sends multipart)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// Fallback to regular form parsing
		if err := r.ParseForm(); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
	}

	requestId, _ := strconv.Atoi(r.FormValue("request_id"))
	log.Printf("Delete request: request_id='%s', parsed=%d", r.FormValue("request_id"), requestId)

	if requestId == 0 {
		s.sendError(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	// Verify the request belongs to the user
	// (we should check this in the database layer, but for now we'll trust it)

	if err := database.DB.DeleteFileRequest(requestId); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete request")
		return
	}

	log.Printf("File request deleted: %d by user %d", requestId, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "File request deleted",
	})
}

// handleUploadRequest routes to either the page or upload handler
func (s *Server) handleUploadRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/upload-request/"):]

	// Check if this is an upload submission (/upload-request/TOKEN/upload)
	if len(path) > 7 && path[len(path)-7:] == "/upload" {
		s.handleUploadRequestSubmit(w, r)
		return
	}

	// Otherwise, show the upload page
	s.handleUploadRequestPage(w, r)
}

// handleUploadRequestPage shows the public upload page for a file request
func (s *Server) handleUploadRequestPage(w http.ResponseWriter, r *http.Request) {
	// Extract token from URL (/upload-request/ABC123)
	token := r.URL.Path[len("/upload-request/"):]

	if token == "" {
		http.Error(w, "Invalid request token", http.StatusNotFound)
		return
	}

	// Get file request from database
	fileRequest, err := database.DB.GetFileRequestByToken(token)
	if err != nil {
		http.Error(w, "File request not found", http.StatusNotFound)
		return
	}

	// Check if already used
	if fileRequest.IsUsed() {
		clientIP := getClientIP(r)
		s.renderUploadRequestUsed(w, fileRequest, clientIP)
		return
	}

	// Check if expired or inactive
	if !fileRequest.IsActive || fileRequest.IsExpired() {
		s.renderUploadRequestExpired(w, fileRequest)
		return
	}

	// Render upload page
	s.renderUploadRequestPage(w, fileRequest)
}

// handleUploadRequestSubmit handles file upload from public upload request
func (s *Server) handleUploadRequestSubmit(w http.ResponseWriter, r *http.Request) {
	// Extract token from URL (/upload-request/ABC123/upload)
	path := r.URL.Path[len("/upload-request/"):]
	token := path[:len(path)-len("/upload")]

	if token == "" {
		s.sendError(w, http.StatusBadRequest, "Invalid request token")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get file request
	fileRequest, err := database.DB.GetFileRequestByToken(token)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "File request not found")
		return
	}

	// Check if already used
	if fileRequest.IsUsed() {
		clientIP := getClientIP(r)
		s.sendError(w, http.StatusGone, fmt.Sprintf("This upload link has already been used from IP: %s", clientIP))
		return
	}

	// Check if expired or inactive
	if !fileRequest.IsActive || fileRequest.IsExpired() {
		s.sendError(w, http.StatusGone, "File request has expired or is inactive")
		return
	}

	// Get the user who created the request
	user, err := database.DB.GetUserByID(fileRequest.UserId)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get request owner")
		return
	}

	// Parse multipart form (32MB max memory buffer, rest spills to disk)
	// This prevents loading entire large files into RAM
	err = r.ParseMultipartForm(32 << 20)
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

	// Get optional comment from uploader
	comment := r.FormValue("comment")
	if len(comment) > 1000 {
		comment = comment[:1000] // Truncate to max length
	}

	// Check file size
	fileSize := header.Size
	if fileRequest.MaxFileSize > 0 && fileSize > fileRequest.MaxFileSize {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("File too large. Max size: %d MB", fileRequest.MaxFileSize/(1024*1024)))
		return
	}

	fileSizeMB := fileSize / (1024 * 1024)

	// Check quota of request owner
	if !user.HasStorageSpace(fileSizeMB) {
		s.sendError(w, http.StatusBadRequest, "Request owner has insufficient storage quota")
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

	// Default expiration: 30 days
	expireTime := time.Now().Add(30 * 24 * time.Hour)
	expireAt := expireTime.Unix()
	expireAtString := expireTime.Format("2006-01-02 15:04")

	// Save file metadata - file belongs to the request owner
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
		DownloadsRemaining: 100, // Default for requested files
		DownloadCount:      0,
		UserId:             user.Id, // File belongs to request owner
		Comment:            comment, // Comment from uploader
		UnlimitedDownloads: false,
		UnlimitedTime:      false,
		RequireAuth:        false,
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

	// Mark file request as used (single-use link)
	clientIP := getClientIP(r)
	if err := database.DB.MarkFileRequestAsUsed(fileRequest.Id, clientIP); err != nil {
		log.Printf("Warning: Could not mark file request as used: %v", err)
	}

	// Send email notification to request owner
	go func() {
		err := email.SendFileUploadNotification(fileRequest, fileInfo, clientIP, s.getPublicURL(), user.Email)
		if err != nil {
			log.Printf("Failed to send upload notification email: %v", err)
		} else {
			log.Printf("Upload notification email sent to %s", user.Email)
		}
	}()

	shareLink := s.getPublicURL() + "/s/" + fileID

	// Audit log for file request upload
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     database.ActionFileRequestUploaded,
		EntityType: database.EntityFileRequest,
		EntityID:   fmt.Sprintf("%d", fileRequest.Id),
		Details: database.CreateAuditDetails(map[string]interface{}{
			"request_title": fileRequest.Title,
			"file_id":       fileID,
			"file_name":     header.Filename,
			"file_size":     fileSize,
			"uploader_ip":   clientIP,
			"has_comment":   comment != "",
		}),
		IPAddress: clientIP,
		UserAgent: r.UserAgent(),
		Success:   true,
	})

	log.Printf("File uploaded via request %s: %s (%s) for user %d - link now consumed by IP %s",
		fileRequest.Title, header.Filename, database.FormatFileSize(fileSize), user.Id, clientIP)

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"file_id":   fileID,
		"file_name": header.Filename,
		"share_url": shareLink,
		"size":      fileSize,
		"message":   "File uploaded successfully",
	})
}

// renderUploadRequestPage renders the public upload page
func (s *Server) renderUploadRequestPage(w http.ResponseWriter, fileRequest *models.FileRequest) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	maxFileSizeMB := fileRequest.MaxFileSize / (1024 * 1024)
	if maxFileSizeMB == 0 {
		maxFileSizeMB = 100 // Default
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Upload File - ` + s.config.CompanyName + `</title>
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
        .upload-container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 600px;
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
        .request-info {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 24px;
        }
        .request-info h2 {
            color: #333;
            font-size: 20px;
            margin-bottom: 12px;
        }
        .request-info p {
            color: #666;
            font-size: 14px;
            margin: 8px 0;
            line-height: 1.6;
        }
        .upload-section {
            margin-bottom: 20px;
        }
        .form-group {
            margin-bottom: 16px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="file"] {
            width: 100%;
            padding: 12px;
            border: 2px dashed #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            cursor: pointer;
        }
        input[type="file"]:hover {
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
        .btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        .success-message {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            display: none;
        }
        .error-message {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            display: none;
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
        .progress {
            width: 100%;
            height: 24px;
            background: #f0f0f0;
            border-radius: 12px;
            overflow: hidden;
            margin: 16px 0;
            display: none;
        }
        .progress-bar {
            height: 100%;
            background: ` + s.getPrimaryColor() + `;
            width: 0%;
            transition: width 0.3s;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 12px;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="upload-container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
        </div>

        <div class="request-info">
            <h2>📤 ` + fileRequest.Title + `</h2>`

	if fileRequest.Message != "" {
		html += `<p>` + fileRequest.Message + `</p>`
	}

	html += `<p style="margin-top: 12px;"><strong>Max file size:</strong> ` + fmt.Sprintf("%d MB", maxFileSizeMB) + `</p>`

	if fileRequest.AllowedFileTypes != "" {
		html += `<p><strong>Allowed types:</strong> ` + fileRequest.AllowedFileTypes + `</p>`
	}

	html += `
        </div>

        <div class="info">
            📁 Upload your file using the form below. The file will be delivered to the requester.
        </div>

        <div class="success-message" id="successMessage"></div>
        <div class="error-message" id="errorMessage"></div>

        <div class="upload-section">
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="form-group">
                    <label for="file">Select File</label>
                    <input type="file" id="file" name="file" required>
                </div>
                <div class="form-group" style="background: #f0f9ff; padding: 15px; border-radius: 8px; border: 2px solid #3b82f6; margin-top: 16px;">
                    <label for="comment" style="color: #1d4ed8; font-weight: 600;">💬 Description/Note (optional)</label>
                    <textarea id="comment" name="comment" rows="3" maxlength="1000" placeholder="Add a description or note about this file (e.g., what it contains, special instructions)" style="width: 100%; padding: 10px; border: 2px solid #93c5fd; border-radius: 6px; font-size: 14px; font-family: inherit; resize: vertical; margin-top: 8px;"></textarea>
                    <p style="color: #1e40af; font-size: 12px; margin-top: 4px;">
                        This message will be shown to the recipient on the download page (max 1000 characters)
                    </p>
                </div>
                <div class="progress" id="progressContainer">
                    <div class="progress-bar" id="progressBar">0%</div>
                </div>
                <button type="submit" class="btn" id="submitBtn">
                    <span style="font-size: 18px; margin-right: 8px;">📤</span>
                    <span style="font-size: 16px; font-weight: 700;">Upload File</span>
                </button>
            </form>
        </div>

        <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
            ` + s.config.FooterText + `
        </div>
    </div>

    <script>
        document.getElementById('uploadForm').addEventListener('submit', async function(e) {
            e.preventDefault();

            const fileInput = document.getElementById('file');
            const submitBtn = document.getElementById('submitBtn');
            const progressContainer = document.getElementById('progressContainer');
            const progressBar = document.getElementById('progressBar');
            const successMsg = document.getElementById('successMessage');
            const errorMsg = document.getElementById('errorMessage');

            if (!fileInput.files[0]) {
                errorMsg.textContent = 'Please select a file';
                errorMsg.style.display = 'block';
                return;
            }

            const formData = new FormData();
            formData.append('file', fileInput.files[0]);
            const commentField = document.getElementById('comment');
            if (commentField && commentField.value) {
                formData.append('comment', commentField.value);
            }

            submitBtn.disabled = true;
            progressContainer.style.display = 'block';
            successMsg.style.display = 'none';
            errorMsg.style.display = 'none';

            try {
                const xhr = new XMLHttpRequest();

                xhr.upload.addEventListener('progress', function(e) {
                    if (e.lengthComputable) {
                        const percentComplete = (e.loaded / e.total) * 100;
                        progressBar.style.width = percentComplete + '%';
                        progressBar.textContent = Math.round(percentComplete) + '%';
                    }
                });

                xhr.addEventListener('load', function() {
                    if (xhr.status === 200) {
                        const response = JSON.parse(xhr.responseText);
                        successMsg.textContent = 'File uploaded successfully! Share link: ' + response.share_url;
                        successMsg.style.display = 'block';
                        fileInput.value = '';
                        progressContainer.style.display = 'none';
                    } else {
                        const response = JSON.parse(xhr.responseText);
                        errorMsg.textContent = 'Upload failed: ' + (response.error || 'Unknown error');
                        errorMsg.style.display = 'block';
                        progressContainer.style.display = 'none';
                    }
                    submitBtn.disabled = false;
                });

                xhr.addEventListener('error', function() {
                    errorMsg.textContent = 'Network error occurred';
                    errorMsg.style.display = 'block';
                    progressContainer.style.display = 'none';
                    submitBtn.disabled = false;
                });

                xhr.open('POST', window.location.pathname + '/upload', true);
                xhr.send(formData);
            } catch (error) {
                errorMsg.textContent = 'Error: ' + error.message;
                errorMsg.style.display = 'block';
                progressContainer.style.display = 'none';
                submitBtn.disabled = false;
            }
        });
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// renderUploadRequestExpired renders the expired request page
func (s *Server) renderUploadRequestExpired(w http.ResponseWriter, fileRequest *models.FileRequest) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Request Expired - ` + s.config.CompanyName + `</title>
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
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px;
            max-width: 600px;
            width: 100%;
            text-align: center;
        }
        .logo h1 {
            color: ` + s.getPrimaryColor() + `;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .expired-icon {
            font-size: 80px;
            margin: 20px 0;
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
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
        </div>
        <div class="expired-icon">⏰</div>
        <h2>Upload Link Expired</h2>
        <p>This upload link has expired and is no longer accepting files.</p>
        <p style="margin-top: 15px;">Please contact the person who sent you this link and ask them to create a new upload request.</p>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderUploadRequestUsed renders the page for already-used upload links
func (s *Server) renderUploadRequestUsed(w http.ResponseWriter, fileRequest *models.FileRequest, clientIP string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Link Already Used - ` + s.config.CompanyName + `</title>
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
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px;
            max-width: 600px;
            width: 100%;
            text-align: center;
        }
        .logo h1 {
            color: ` + s.getPrimaryColor() + `;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .used-icon {
            font-size: 80px;
            margin: 20px 0;
        }
        h2 {
            color: #ff9800;
            font-size: 28px;
            margin-bottom: 15px;
        }
        p {
            color: #666;
            font-size: 16px;
            line-height: 1.6;
            margin: 10px 0;
        }
        .ip-info {
            background: #fff3cd;
            border: 1px solid #ffc107;
            color: #856404;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            font-family: monospace;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
        </div>
        <div class="used-icon">🔒</div>
        <h2>Upload Link Already Used</h2>
        <p>This upload link has already been used and is no longer accepting files.</p>
        <div class="ip-info">
            This link was used from IP: ` + fileRequest.UsedByIP + `
        </div>
        <p style="margin-top: 15px;">Upload request links are single-use for security purposes. Please contact the person who sent you this link and ask them to create a new upload request.</p>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
