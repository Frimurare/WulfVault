package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/email"
	"github.com/Frimurare/Sharecare/internal/models"
)

// handleUserDashboard renders the user dashboard
func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	s.renderUserDashboard(w, user)
}

// handleUserFiles returns the user's files as JSON
func (s *Server) handleUserFiles(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// TODO: Get files from database
	files := []map[string]interface{}{}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"files": files,
		"user":  user,
	})
}

// handleFileEdit edits a file's settings
func (s *Server) handleFileEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	fileID := r.FormValue("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	expirationDays, _ := strconv.Atoi(r.FormValue("expiration_days"))
	downloadsLimit, _ := strconv.Atoi(r.FormValue("downloads_limit"))

	// Get file to verify ownership
	fileInfo, err := database.DB.GetFileByID(fileID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check ownership (unless admin)
	if fileInfo.UserId != user.Id && !user.IsAdmin() {
		s.sendError(w, http.StatusForbidden, "Not authorized to edit this file")
		return
	}

	// Update expiration
	var newExpireAt int64
	var newExpireAtString string
	unlimitedTime := expirationDays == 0

	if expirationDays > 0 {
		expireTime := time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour)
		newExpireAt = expireTime.Unix()
		newExpireAtString = expireTime.Format("2006-01-02 15:04")
	}

	unlimitedDownloads := downloadsLimit == 0
	if downloadsLimit == 0 {
		downloadsLimit = 999999
	}

	// Update in database
	if err := database.DB.UpdateFileSettings(fileID, downloadsLimit, newExpireAt, newExpireAtString, unlimitedDownloads, unlimitedTime); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to update file: "+err.Error())
		return
	}

	log.Printf("File settings updated: %s by user %d", fileInfo.Name, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "File updated successfully",
	})
}

// handleFileDownloadHistory returns download logs for a file
func (s *Server) handleFileDownloadHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	fileID := r.URL.Query().Get("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	// Get file to verify ownership
	fileInfo, err := database.DB.GetFileByID(fileID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check ownership (unless admin)
	if fileInfo.UserId != user.Id && !user.IsAdmin() {
		s.sendError(w, http.StatusForbidden, "Not authorized to view this file's download history")
		return
	}

	// Get download logs
	downloadLogs, err := database.DB.GetDownloadLogsByFileID(fileID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get download logs")
		return
	}

	// Get email logs
	emailLogs, err := database.DB.GetEmailLogsByFileID(fileID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get email logs")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"downloadLogs": downloadLogs,
		"emailLogs":    emailLogs,
	})
}

// handleFileDelete deletes a file
func (s *Server) handleFileDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fileID := r.FormValue("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	// Get file to verify ownership
	fileInfo, err := database.DB.GetFileByID(fileID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check ownership (unless admin)
	if fileInfo.UserId != user.Id && !user.IsAdmin() {
		s.sendError(w, http.StatusForbidden, "Not authorized to delete this file")
		return
	}

	// Soft delete (move to trash)
	if err := database.DB.DeleteFile(fileID, user.Id); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	// Recalculate user storage
	newStorage, _ := database.DB.CalculateUserStorage(user.Id)
	database.DB.UpdateUserStorage(user.Id, newStorage)

	log.Printf("File deleted: %s by user %d", fileInfo.Name, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "File deleted successfully",
	})
}

// handleFileEmail sends a file link via email
func (s *Server) handleFileEmail(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse JSON request body
	var request struct {
		FileID    string `json:"fileId"`
		Recipient string `json:"recipient"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if request.FileID == "" || request.Recipient == "" {
		s.sendError(w, http.StatusBadRequest, "Missing fileId or recipient")
		return
	}

	// Get file to verify ownership
	fileInfo, err := database.DB.GetFileByID(request.FileID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check ownership (unless admin)
	if fileInfo.UserId != user.Id && !user.IsAdmin() {
		s.sendError(w, http.StatusForbidden, "Not authorized to share this file")
		return
	}

	// Construct file URL
	fileURL := fmt.Sprintf("%s/s/%s", s.getPublicURL(), fileInfo.Id)

	// Get active email provider
	provider, err := email.GetActiveProvider(database.DB)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "No active email provider configured")
		return
	}

	// Construct email content
	subject := fmt.Sprintf("%s has shared a file with you", user.Name)

	htmlBody := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">You've received a file</h2>
			<p><strong>%s</strong> has shared the following file with you:</p>
			<div style="background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 20px 0;">
				<h3 style="margin: 0 0 10px 0; color: #2563eb;">%s</h3>
				<p style="color: #666; margin: 0;">Size: %.2f MB</p>
			</div>
			%s
			<p>
				<a href="%s" style="display: inline-block; background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Download File</a>
			</p>
			<p style="color: #999; font-size: 12px; margin-top: 30px;">
				This link was sent from %s
			</p>
		</body>
		</html>
	`, user.Name, fileInfo.Name, float64(fileInfo.SizeBytes)/(1024*1024),
		func() string {
			if request.Message != "" {
				return fmt.Sprintf(`<p style="color: #666;"><em>Message from sender:</em><br>%s</p>`, request.Message)
			}
			return ""
		}(),
		fileURL, s.config.CompanyName)

	textBody := fmt.Sprintf(
		"%s has shared a file with you\n\nFile: %s\nSize: %.2f MB\n\n%sDownload: %s\n\nThis link was sent from %s",
		user.Name, fileInfo.Name, float64(fileInfo.SizeBytes)/(1024*1024),
		func() string {
			if request.Message != "" {
				return fmt.Sprintf("Message: %s\n\n", request.Message)
			}
			return ""
		}(),
		fileURL, s.config.CompanyName,
	)

	// Send email
	if err := provider.SendEmail(request.Recipient, subject, htmlBody, textBody); err != nil {
		log.Printf("Failed to send email to %s: %v", request.Recipient, err)
		s.sendError(w, http.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	// Log the email send to database
	if err := database.DB.LogEmailSent(fileInfo.Id, user.Id, request.Recipient, request.Message, fileInfo.Name, fileInfo.SizeBytes); err != nil {
		log.Printf("Warning: Failed to log email send: %v", err)
		// Don't fail the request if logging fails
	}

	log.Printf("File link emailed: %s to %s by user %d", fileInfo.Name, request.Recipient, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "Email sent successfully",
	})
}

// renderUserDashboard renders the user dashboard HTML
func (s *Server) renderUserDashboard(w http.ResponseWriter, userModel interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	user := userModel.(*models.User)

	// Get branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()
	logoData := brandingConfig["branding_logo"]

	// Get user's files
	files, _ := database.DB.GetFilesByUser(user.Id)

	// Calculate storage
	storageUsed := user.StorageUsedMB
	storageQuota := user.StorageQuotaMB
	storagePercent := 0
	if storageQuota > 0 {
		storagePercent = int((storageUsed * 100) / storageQuota)
	}

	activeFileCount := 0
	totalDownloads := 0
	for _, f := range files {
		// Count active files
		if f.DownloadsRemaining > 0 || f.UnlimitedDownloads {
			activeFileCount++
		}
		totalDownloads += f.DownloadCount
	}

	// Stats with real data
	storageUsedGB := fmt.Sprintf("%.1f", float64(storageUsed)/1000)
	storageQuotaGB := fmt.Sprintf("%.1f", float64(storageQuota)/1000)

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - ` + s.config.CompanyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header .logo {
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .header .logo img {
            max-height: 50px;
            max-width: 180px;
        }
        .header h1 {
            color: white;
            font-size: 24px;
            font-weight: 600;
        }
        .header nav {
            display: flex;
            align-items: center;
            gap: 20px;
        }
        .header nav a {
            color: rgba(255, 255, 255, 0.9);
            text-decoration: none;
            font-weight: 500;
            transition: color 0.3s;
        }
        .header nav a:hover {
            color: white;
        }
        .container {
            max-width: 1200px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .stat-card {
            background: white;
            padding: 24px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .stat-card h3 {
            color: #666;
            font-size: 14px;
            font-weight: 500;
            margin-bottom: 8px;
        }
        .stat-card .value {
            font-size: 32px;
            font-weight: 700;
            color: ` + s.getPrimaryColor() + `;
        }
        .stat-card .progress {
            margin-top: 12px;
            height: 8px;
            background: #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
        }
        .stat-card .progress-bar {
            height: 100%;
            background: ` + s.getPrimaryColor() + `;
            transition: width 0.3s;
        }
        .upload-zone {
            background: white;
            border: 3px dashed #ddd;
            border-radius: 12px;
            padding: 60px 20px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
            margin-bottom: 40px;
        }
        .upload-zone:hover {
            border-color: ` + s.getPrimaryColor() + `;
            background: #f9f9f9;
        }
        .upload-zone.drag-over {
            border-color: ` + s.getPrimaryColor() + `;
            background: #f0f8ff;
        }
        .upload-zone svg {
            width: 64px;
            height: 64px;
            margin-bottom: 16px;
            color: #999;
        }
        .upload-zone h2 {
            color: #333;
            margin-bottom: 8px;
        }
        .upload-zone p {
            color: #999;
        }
        .files-section {
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .files-header {
            padding: 24px;
            border-bottom: 1px solid #e0e0e0;
        }
        .files-header h2 {
            color: #333;
        }
        .file-list {
            list-style: none;
        }
        .file-item {
            padding: 20px 24px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .file-item:last-child {
            border-bottom: none;
        }
        .file-info h3 {
            color: #333;
            font-size: 16px;
            margin-bottom: 4px;
        }
        .file-info p {
            color: #999;
            font-size: 14px;
        }
        .file-actions {
            display: flex;
            gap: 12px;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: opacity 0.3s;
        }
        .btn:hover {
            opacity: 0.8;
        }
        .btn-primary {
            background: ` + s.getPrimaryColor() + `;
            color: white;
        }
        .btn-secondary {
            background: #e0e0e0;
            color: #333;
        }
        .btn-danger {
            background: #dc3545;
            color: white;
        }
        .empty-state {
            padding: 60px 20px;
            text-align: center;
            color: #999;
        }
        input[type="file"] {
            display: none;
        }
        .upload-section {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 40px;
        }
        .upload-options {
            margin-top: 24px;
            padding-top: 24px;
            border-top: 2px solid #e0e0e0;
        }
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
        }
        .form-group input[type="text"],
        .form-group input[type="date"],
        .form-group input[type="number"] {
            width: 100%;
            padding: 10px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
        }
        .form-group input:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .btn-large {
            padding: 12px 32px;
            font-size: 16px;
            margin-right: 12px;
        }
        .link-display {
            margin-top: 12px;
            padding: 12px;
            background: #f0f8ff;
            border: 1px solid #b3d9ff;
            border-radius: 6px;
        }
        .link-display h4 {
            color: #333;
            font-size: 13px;
            margin-bottom: 8px;
            font-weight: 600;
        }
        .link-box {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px;
            background: white;
            border-radius: 4px;
            font-family: monospace;
            font-size: 12px;
            margin-bottom: 8px;
        }
        .link-box a {
            flex: 1;
            color: #1976d2;
            text-decoration: none;
            word-break: break-all;
        }
        .link-box button {
            flex-shrink: 0;
        }
        @media (max-width: 768px) {
            .form-row {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">`

	if logoData != "" {
		html += `
            <img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		html += `
            <h1>` + s.config.CompanyName + `</h1>`
	}

	html += `
        </div>
        <nav>`

	// Add admin link if user is admin
	if user.IsAdmin() {
		html += `
            <a href="/admin">Admin Panel</a>`
	}

	html += `
            <a href="/dashboard">Dashboard</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <div class="stats">
            <div class="stat-card">
                <h3>Storage Used</h3>
                <div class="value">` + storageUsedGB + ` GB</div>
                <div class="progress">
                    <div class="progress-bar" style="width: ` + fmt.Sprintf("%d", storagePercent) + `%"></div>
                </div>
                <p style="margin-top: 8px; color: #999; font-size: 14px;">` + storageUsedGB + ` GB of ` + storageQuotaGB + ` GB</p>
            </div>
            <div class="stat-card">
                <h3>Active Files</h3>
                <div class="value">` + fmt.Sprintf("%d", activeFileCount) + `</div>
            </div>
            <div class="stat-card">
                <h3>Total Downloads</h3>
                <div class="value">` + fmt.Sprintf("%d", totalDownloads) + `</div>
            </div>
        </div>

        <!-- Upload Form -->
        <div class="upload-section">
            <h2 style="margin-bottom: 20px; color: #333;">Upload File</h2>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="upload-zone" id="uploadZone" onclick="document.getElementById('fileInput').click()">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <h3>Drop files here or click to select</h3>
                    <p>Maximum file size: 5 GB</p>
                    <input type="file" id="fileInput" name="file">
                </div>

                <div class="upload-options" id="uploadOptions" style="display: none;">
                    <h3 style="margin-bottom: 16px; color: #333;">Upload Settings</h3>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="expireDate">📅 Expiration Date</label>
                            <input type="date" id="expireDate" name="expire_date">
                            <label style="margin-top: 8px;">
                                <input type="checkbox" id="unlimitedTime" name="unlimited_time"> Never expire (by time)
                            </label>
                        </div>

                        <div class="form-group">
                            <label for="downloadsLimit">⬇️ Download Limit</label>
                            <input type="number" id="downloadsLimit" name="downloads_limit" min="1" value="10">
                            <label style="margin-top: 8px;">
                                <input type="checkbox" id="unlimitedDownloads" name="unlimited_downloads"> Unlimited downloads
                            </label>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>🔗 Link Type</label>
                        <div style="display: flex; gap: 16px; margin-top: 8px;">
                            <label style="display: flex; align-items: center; gap: 8px;">
                                <input type="radio" name="link_type" value="splash" checked>
                                <span>Splash Page (recommended)</span>
                            </label>
                            <label style="display: flex; align-items: center; gap: 8px;">
                                <input type="radio" name="link_type" value="direct">
                                <span>Direct Download</span>
                            </label>
                        </div>
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">
                            Splash page shows branding and file info before download
                        </p>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="requireAuth" name="require_auth">
                            🔒 Require recipient authentication (email + password)
                        </label>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="enablePassword" onchange="togglePasswordField()">
                            🔐 Password protect this file
                        </label>
                        <div id="passwordFieldContainer" style="display: none; margin-top: 12px;">
                            <input type="text" id="filePassword" name="file_password" placeholder="Enter password" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                            <p style="color: #666; font-size: 12px; margin-top: 4px;">
                                Recipients will need this password to download the file
                            </p>
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="sendToEmail">📧 Send link to email (optional)</label>
                        <input type="email" id="sendToEmail" name="send_to_email" placeholder="recipient@example.com" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">
                            After upload, send the download link via email to this address
                        </p>
                    </div>

                    <button type="submit" class="btn btn-primary btn-large" id="uploadButton" style="display: flex; align-items: center; justify-content: center; gap: 10px;">
                        <span style="font-size: 24px;">📤</span>
                        <span style="font-size: 18px; font-weight: 700;">Upload File</span>
                    </button>
                    <button type="button" class="btn btn-secondary" onclick="resetUploadForm()">
                        ✖️ Cancel
                    </button>
                </div>
            </form>
        </div>

        <!-- File Request Section -->
        <div class="file-request-section" style="background: white; padding: 30px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); margin-bottom: 40px;">
            <h2 style="margin-bottom: 16px; color: #333;">📥 Request Files from Others</h2>
            <p style="color: #666; margin-bottom: 20px;">Create a link that allows others to upload files directly to you. Perfect for collecting files from clients or colleagues.</p>
            <button onclick="showCreateRequestModal()" style="padding: 12px 24px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                ➕ Create Upload Request
            </button>
            <div id="requestsList" style="margin-top: 20px;"></div>
        </div>

        <!-- File Request Modal -->
        <div id="fileRequestModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 10000; align-items: center; justify-content: center;">
            <div style="background: white; border-radius: 12px; padding: 32px; max-width: 500px; width: 90%; box-shadow: 0 8px 32px rgba(0,0,0,0.2);">
                <h2 style="margin-bottom: 24px; color: #333;">Create Upload Request</h2>
                <form id="fileRequestForm" onsubmit="submitFileRequest(event)">
                    <div style="margin-bottom: 20px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">Title *</label>
                        <input type="text" id="requestTitle" required placeholder="e.g., Upload Documents" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">Short description of what you're requesting</p>
                    </div>

                    <div style="margin-bottom: 20px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">Message (optional)</label>
                        <textarea id="requestMessage" placeholder="Additional instructions for the uploader..." style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px; min-height: 80px; resize: vertical;"></textarea>
                    </div>

                    <div style="margin-bottom: 20px; padding: 12px; background: #fff3cd; border: 1px solid #ffc107; border-radius: 6px;">
                        <p style="color: #856404; font-size: 13px; margin: 0;">
                            ⏰ <strong>Note:</strong> The upload link will automatically expire after 24 hours for security purposes.
                        </p>
                    </div>

                    <div style="margin-bottom: 24px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">Max file size (MB)</label>
                        <input type="number" id="requestMaxSize" min="1" max="5000" value="100" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">Maximum size per file (0 = no limit)</p>
                    </div>

                    <div style="margin-bottom: 24px; background: #fff9e6; padding: 16px; border-radius: 8px; border: 3px solid #ff9800;">
                        <label style="display: block; margin-bottom: 8px; color: #e65100; font-weight: 700; font-size: 16px;">📧 Send upload request to email (optional)</label>
                        <input type="email" id="requestRecipientEmail" placeholder="recipient@example.com" style="width: 100%; padding: 12px; border: 3px solid #ff9800; border-radius: 6px; font-size: 14px; background: white;">
                        <p style="color: #e65100; font-size: 13px; margin-top: 8px; font-weight: 600;">Send the upload link directly to this email address</p>
                    </div>

                    <div style="display: flex; gap: 12px;">
                        <button type="submit" style="flex: 1; padding: 12px 24px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                            Create Request
                        </button>
                        <button type="button" onclick="closeFileRequestModal()" style="flex: 1; padding: 12px 24px; background: #f5f5f5; color: #333; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <div class="files-section">
            <div class="files-header">
                <h2>My Files</h2>
            </div>`

	if len(files) == 0 {
		html += `
            <div class="empty-state">
                No files uploaded yet. Start by uploading your first file!
            </div>`
	} else {
		html += `
            <ul class="file-list">`
		for _, f := range files {
			// Both URL types
			splashURL := s.getPublicURL() + "/s/" + f.Id
			directURL := s.getPublicURL() + "/d/" + f.Id
			// Escape URLs for safe use in JavaScript
			splashURLEscaped := template.HTMLEscapeString(splashURL)
			directURLEscaped := template.HTMLEscapeString(directURL)
			status := "Active"
			statusColor := "#4caf50"

			if !f.UnlimitedDownloads && f.DownloadsRemaining <= 0 {
				status = "Expired (downloads)"
				statusColor = "#f44336"
			} else if !f.UnlimitedTime && f.ExpireAt > 0 && f.ExpireAt < time.Now().Unix() {
				status = "Expired (time)"
				statusColor = "#f44336"
			}

			expiryInfo := ""
			if f.UnlimitedTime && f.UnlimitedDownloads {
				expiryInfo = "Never expires"
			} else if f.UnlimitedTime {
				expiryInfo = fmt.Sprintf("%d downloads remaining", f.DownloadsRemaining)
			} else if f.UnlimitedDownloads {
				expiryInfo = fmt.Sprintf("Expires: %s", f.ExpireAtString)
			} else {
				expiryInfo = fmt.Sprintf("%d downloads left, expires %s", f.DownloadsRemaining, f.ExpireAtString)
			}

			authBadge := ""
			if f.RequireAuth {
				authBadge = `<span style="background: #2196f3; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">🔒 Auth Required</span>`
			}

			passwordBadge := ""
			if f.FilePasswordPlain != "" {
				passwordBadge = `<span style="background: #9c27b0; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">🔐 Password Protected</span>`
			}

			passwordDisplay := ""
			if f.FilePasswordPlain != "" {
				passwordDisplay = fmt.Sprintf(`<p style="margin-top: 8px;"><strong>🔐 Password:</strong> <span id="password-%s" style="cursor: pointer; color: #9c27b0; text-decoration: underline;" onclick="togglePasswordVisibility('%s', '%s')">👁️ Show</span></p>`,
					f.Id, f.Id, template.JSEscapeString(f.FilePasswordPlain))
			}

			html += fmt.Sprintf(`
                <li class="file-item">
                    <div class="file-info">
                        <h3>📄 %s %s%s</h3>
                        <p>%s • Downloaded %d times • %s</p>
                        <p style="color: %s;">Status: %s</p>
                        %s
                        <div class="link-display">
                            <h4>🌐 Splash Page (Recommended - Shows branding)</h4>
                            <div class="link-box">
                                <a href="%s" target="_blank">%s</a>
                                <button class="btn btn-primary" onclick="copyToClipboard('%s', this)" style="font-size: 11px; padding: 4px 8px;">📋 Copy</button>
                            </div>
                            <h4>⬇️ Direct Download Link</h4>
                            <div class="link-box">
                                <a href="%s" target="_blank">%s</a>
                                <button class="btn btn-primary" onclick="copyToClipboard('%s', this)" style="font-size: 11px; padding: 4px 8px;">📋 Copy</button>
                            </div>
                        </div>
                    </div>
                    <div class="file-actions">
                        <button class="btn btn-secondary" onclick="showDownloadHistory('%s', '%s')" title="View download history">
                            📊 History
                        </button>
                        <button class="btn btn-primary" onclick="showEmailModal('%s', '%s', '%s')" title="Send file link via email" style="background: #007bff;">
                            📧 Email
                        </button>
                        <button class="btn btn-secondary" onclick="showEditModal('%s', '%s', %d, %d, %t, %t)" title="Edit file settings">
                            ✏️ Edit
                        </button>
                        <button class="btn btn-danger" onclick="deleteFile('%s', '%s')">
                            🗑️ Delete
                        </button>
                    </div>
                </li>`, template.HTMLEscapeString(f.Name), authBadge, passwordBadge, f.Size, f.DownloadCount, expiryInfo, statusColor, status, passwordDisplay,
				splashURL, splashURL, splashURLEscaped,
				directURL, directURL, directURLEscaped,
				f.Id, template.JSEscapeString(f.Name), f.Id, template.JSEscapeString(f.Name), template.JSEscapeString(splashURL), f.Id, template.JSEscapeString(f.Name), f.DownloadsRemaining, f.ExpireAt, f.UnlimitedDownloads, f.UnlimitedTime, f.Id, template.JSEscapeString(f.Name))
		}
		html += `
            </ul>`
	}

	html += `
        </div>
    </div>

    <!-- Email File Modal -->
    <div id="emailModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 500px; width: 90%;">
            <h2 style="margin-bottom: 24px; color: #333;">Send File Link via Email</h2>
            <input type="hidden" id="emailFileId">
            <p style="margin-bottom: 20px; color: #666;">Sending link for: <strong id="emailFileName"></strong></p>
            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; color: #555; font-weight: 500;">Recipient Email:</label>
                <input type="email" id="emailRecipient" placeholder="recipient@example.com" style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px;">
            </div>
            <div style="margin-bottom: 24px;">
                <label style="display: block; margin-bottom: 8px; color: #555; font-weight: 500;">Message (optional):</label>
                <textarea id="emailMessage" rows="4" placeholder="Add a personal message..." style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px; resize: vertical;"></textarea>
            </div>
            <div style="display: flex; gap: 12px; justify-content: flex-end;">
                <button onclick="closeEmailModal()" class="btn btn-secondary" style="padding: 10px 20px;">Cancel</button>
                <button onclick="sendEmailLink()" class="btn btn-primary" style="padding: 10px 20px; background: #007bff;">Send Email</button>
            </div>
        </div>
    </div>

    <!-- Edit File Modal -->
    <div id="editModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 500px; width: 90%;">
            <h2 style="margin-bottom: 24px; color: #333;">Edit File Settings</h2>

            <input type="hidden" id="editFileId">

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">File:</label>
                <p id="editFileName" style="color: #666; font-weight: 600;"></p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editUnlimitedTime" onchange="toggleEditTimeLimit()">
                    Never expire (keep forever)
                </label>
            </div>

            <div id="editTimeLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">Days Until Expiration:</label>
                <input type="number" id="editExpirationDays" value="7" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">Days from now until file expires</p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editUnlimitedDownloads" onchange="toggleEditDownloadLimit()">
                    Unlimited downloads
                </label>
            </div>

            <div id="editDownloadLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">Downloads Remaining:</label>
                <input type="number" id="editDownloadsLimit" value="5" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="saveFileEdit()" style="flex: 1; padding: 14px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    Save Changes
                </button>
                <button onclick="closeEditModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    Cancel
                </button>
            </div>
        </div>
    </div>

    <!-- Upload Settings Modal -->
    <div id="uploadModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 500px; width: 90%;">
            <h2 style="margin-bottom: 24px; color: #333;">Upload Settings</h2>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">File:</label>
                <p id="selectedFileName" style="color: #666;"></p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="requireAuth" style="margin-right: 8px;">
                    Require authentication to download
                </label>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="unlimitedTime" onchange="toggleTimeLimit()">
                    Never expire (keep forever)
                </label>
            </div>

            <div id="timeLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">Expiration Days:</label>
                <input type="number" id="expirationDays" value="7" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">Set to 0 for no time limit</p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="unlimitedDownloads" onchange="toggleDownloadLimit()">
                    Unlimited downloads
                </label>
            </div>

            <div id="downloadLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">Download Limit:</label>
                <input type="number" id="downloadsLimit" value="5" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">Set to 0 for unlimited downloads</p>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="performUpload()" style="flex: 1; padding: 14px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    Upload File
                </button>
                <button onclick="closeUploadModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    Cancel
                </button>
            </div>

            <div id="uploadProgress" style="display: none; margin-top: 20px;">
                <div style="background: #e0e0e0; border-radius: 4px; overflow: hidden; height: 8px;">
                    <div id="progressBar" style="height: 100%; background: ` + s.getPrimaryColor() + `; width: 0%; transition: width 0.3s;"></div>
                </div>
                <p id="uploadStatus" style="text-align: center; margin-top: 8px; color: #666;"></p>
            </div>
        </div>
    </div>

    <!-- Download History Modal -->
    <div id="downloadHistoryModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 800px; width: 90%; max-height: 80vh; overflow-y: auto;">
            <h2 style="margin-bottom: 24px; color: #333;">📊 Download History</h2>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">File:</label>
                <p id="historyFileName" style="color: #666; font-weight: 600;"></p>
            </div>

            <div id="downloadHistoryContent" style="margin-top: 20px;">
                <p style="text-align: center; color: #999;">Loading...</p>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="closeDownloadHistoryModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    Close
                </button>
            </div>
        </div>
    </div>

    <script src="/static/js/dashboard.js"></script>
    <script>
        function showDownloadHistory(fileId, fileName) {
            document.getElementById('historyFileName').textContent = fileName;
            document.getElementById('downloadHistoryModal').style.display = 'flex';
            document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">Loading...</p>';

            fetch('/file/downloads?file_id=' + encodeURIComponent(fileId))
                .then(response => response.json())
                .then(data => {
                    const downloadLogs = data.downloadLogs || [];
                    const emailLogs = data.emailLogs || [];

                    if (downloadLogs.length === 0 && emailLogs.length === 0) {
                        document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">No activity yet</p>';
                        return;
                    }

                    let html = '';

                    // Show download logs
                    if (downloadLogs.length > 0) {
                        html += '<h3 style="margin-top: 0; margin-bottom: 15px; color: #333; font-size: 16px;">📥 Downloads (' + downloadLogs.length + ')</h3>';
                        html += '<table style="width: 100%; border-collapse: collapse; margin-bottom: 30px;">';
                        html += '<thead><tr style="background: #f5f5f5; border-bottom: 2px solid #ddd;">';
                        html += '<th style="padding: 12px; text-align: left;">Date & Time</th>';
                        html += '<th style="padding: 12px; text-align: left;">Downloaded By</th>';
                        html += '<th style="padding: 12px; text-align: left;">IP Address</th>';
                        html += '</tr></thead><tbody>';

                        downloadLogs.forEach(log => {
                            const date = new Date(log.downloadedAt * 1000);
                            const dateStr = date.toLocaleString('sv-SE');
                            const downloader = log.email || 'Anonymous';
                            const ip = log.ipAddress || 'N/A';
                            const authBadge = log.isAuthenticated ? ' <span style="background: #2196f3; color: white; padding: 2px 6px; border-radius: 3px; font-size: 11px;">🔒 Auth</span>' : '';

                            html += '<tr style="border-bottom: 1px solid #eee;">';
                            html += '<td style="padding: 12px;">' + dateStr + '</td>';
                            html += '<td style="padding: 12px;">' + downloader + authBadge + '</td>';
                            html += '<td style="padding: 12px; font-family: monospace; font-size: 12px;">' + ip + '</td>';
                            html += '</tr>';
                        });

                        html += '</tbody></table>';
                    }

                    // Show email logs
                    if (emailLogs.length > 0) {
                        html += '<h3 style="margin-top: 0; margin-bottom: 15px; color: #333; font-size: 16px;">📧 Emails Sent (' + emailLogs.length + ')</h3>';
                        html += '<table style="width: 100%; border-collapse: collapse;">';
                        html += '<thead><tr style="background: #f5f5f5; border-bottom: 2px solid #ddd;">';
                        html += '<th style="padding: 12px; text-align: left;">Date & Time</th>';
                        html += '<th style="padding: 12px; text-align: left;">Recipient</th>';
                        html += '<th style="padding: 12px; text-align: left;">Message</th>';
                        html += '</tr></thead><tbody>';

                        emailLogs.forEach(log => {
                            const date = new Date(log.sentAt * 1000);
                            const dateStr = date.toLocaleString('sv-SE');
                            const message = log.message || '<em style="color: #999;">No message</em>';

                            html += '<tr style="border-bottom: 1px solid #eee;">';
                            html += '<td style="padding: 12px;">' + dateStr + '</td>';
                            html += '<td style="padding: 12px;">' + log.recipientEmail + '</td>';
                            html += '<td style="padding: 12px; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="' + (log.message || '') + '">' + message + '</td>';
                            html += '</tr>';
                        });

                        html += '</tbody></table>';
                    }

                    document.getElementById('downloadHistoryContent').innerHTML = html;
                })
                .catch(error => {
                    document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #f44336;">Error loading history</p>';
                    console.error('Error:', error);
                });
        }

        function closeDownloadHistoryModal() {
            document.getElementById('downloadHistoryModal').style.display = 'none';
        }

        function togglePasswordField() {
            const checkbox = document.getElementById('enablePassword');
            const container = document.getElementById('passwordFieldContainer');
            const passwordInput = document.getElementById('filePassword');

            if (checkbox.checked) {
                container.style.display = 'block';
                passwordInput.required = true;
            } else {
                container.style.display = 'none';
                passwordInput.required = false;
                passwordInput.value = '';
            }
        }

        function togglePasswordVisibility(fileId, password) {
            const element = document.getElementById('password-' + fileId);
            if (element.textContent === '👁️ Show') {
                element.textContent = password;
                element.style.fontFamily = 'monospace';
            } else {
                element.textContent = '👁️ Show';
                element.style.fontFamily = 'inherit';
            }
        }

        // Email Modal Functions
        function showEmailModal(fileId, fileName, fileUrl) {
            document.getElementById('emailFileId').value = fileId;
            document.getElementById('emailFileName').textContent = fileName;
            document.getElementById('emailModal').style.display = 'flex';
        }

        function closeEmailModal() {
            document.getElementById('emailModal').style.display = 'none';
            document.getElementById('emailRecipient').value = '';
            document.getElementById('emailMessage').value = '';
        }

        async function sendEmailLink() {
            const fileId = document.getElementById('emailFileId').value;
            const recipient = document.getElementById('emailRecipient').value;
            const message = document.getElementById('emailMessage').value;

            if (!recipient) {
                alert('Please enter a recipient email address');
                return;
            }

            const btn = event.target;
            btn.disabled = true;
            btn.textContent = 'Sending...';

            try {
                const response = await fetch('/file/email', {
                    method: 'POST',
                    credentials: 'include',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ fileId, recipient, message })
                });

                const result = await response.json();
                if (response.ok) {
                    alert('Email sent successfully!');
                    closeEmailModal();
                } else {
                    alert('Error: ' + (result.error || 'Failed to send email'));
                }
            } catch (error) {
                alert('Error sending email: ' + error.message);
            } finally {
                btn.disabled = false;
                btn.textContent = 'Send Email';
            }
        }

        // Edit File Modal Functions
        function showEditModal(fileId, fileName, downloadsRemaining, expireAt, unlimitedDownloads, unlimitedTime) {
            // Store file info
            document.getElementById('editFileId').value = fileId;
            document.getElementById('editFileName').textContent = fileName;

            // Set unlimited checkboxes
            document.getElementById('editUnlimitedTime').checked = unlimitedTime;
            document.getElementById('editUnlimitedDownloads').checked = unlimitedDownloads;

            // Calculate days until expiration
            if (expireAt > 0 && !unlimitedTime) {
                const now = Math.floor(Date.now() / 1000);
                const daysRemaining = Math.max(0, Math.ceil((expireAt - now) / (24 * 60 * 60)));
                document.getElementById('editExpirationDays').value = daysRemaining;
            } else {
                document.getElementById('editExpirationDays').value = 7;
            }

            // Set downloads limit
            if (unlimitedDownloads) {
                document.getElementById('editDownloadsLimit').value = 5;
            } else {
                document.getElementById('editDownloadsLimit').value = downloadsRemaining;
            }

            // Toggle sections based on unlimited flags
            toggleEditTimeLimit();
            toggleEditDownloadLimit();

            // Show modal
            document.getElementById('editModal').style.display = 'flex';
        }

        function closeEditModal() {
            document.getElementById('editModal').style.display = 'none';
        }

        function toggleEditTimeLimit() {
            const checkbox = document.getElementById('editUnlimitedTime');
            const section = document.getElementById('editTimeLimitSection');
            section.style.display = checkbox.checked ? 'none' : 'block';
        }

        function toggleEditDownloadLimit() {
            const checkbox = document.getElementById('editUnlimitedDownloads');
            const section = document.getElementById('editDownloadLimitSection');
            section.style.display = checkbox.checked ? 'none' : 'block';
        }

        function saveFileEdit() {
            const fileId = document.getElementById('editFileId').value;
            const unlimitedTime = document.getElementById('editUnlimitedTime').checked;
            const unlimitedDownloads = document.getElementById('editUnlimitedDownloads').checked;

            let expirationDays = 0;
            if (!unlimitedTime) {
                expirationDays = parseInt(document.getElementById('editExpirationDays').value) || 0;
            }

            let downloadsLimit = 0;
            if (!unlimitedDownloads) {
                downloadsLimit = parseInt(document.getElementById('editDownloadsLimit').value) || 0;
            }

            const formData = new FormData();
            formData.append('file_id', fileId);
            formData.append('expiration_days', expirationDays);
            formData.append('downloads_limit', downloadsLimit);

            fetch('/file/edit', {
                method: 'POST',
                body: formData,
                credentials: 'same-origin'
            })
            .then(response => response.json())
            .then(result => {
                if (result.message) {
                    closeEditModal();
                    location.reload();
                } else if (result.error) {
                    alert('Error: ' + result.error);
                }
            })
            .catch(error => {
                alert('Error saving changes: ' + error);
            });
        }

        // Note: loadFileRequests, deleteFileRequest, escapeHtml, and copyToClipboard
        // are defined in dashboard.js and loaded automatically on page load
    </script>
</body>
</html>`

	w.Write([]byte(html))
}
