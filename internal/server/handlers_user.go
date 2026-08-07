// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

// handleUserDashboard renders the user dashboard
func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	s.renderUserDashboard(w, s.tr(r), user)
}

// handleWhoAmI returns JSON with the currently authenticated user's info.
// Designed for API clients (like external API client) to reliably
// verify that their session cookie is valid without any side effects.
// Added in v6.2.7.
//
// GET /api/whoami
//
// Returns:
//
//	200 { "authenticated": true, "id": 123, "email": "...", "name": "...",
//	      "role": "user", "storage_used_mb": 42, "storage_quota_mb": 1000,
//	      "server_version": "6.2.7", "two_factor_enabled": false }
//	401 { "authenticated": false, "error": "Not authenticated" }
//
// No form data, no side effects, fast. Does NOT use the requireAuth
// middleware because that middleware redirects unauthenticated browser
// requests to /login (HTTP 303). API clients need a proper JSON 401.
func (s *Server) handleWhoAmI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	// Manually resolve the session since this handler is NOT wrapped in requireAuth.
	// We must return JSON 401 instead of the HTML redirect that requireAuth does.
	_, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
			"error":         "Not authenticated",
		})
		return
	}
	user, err := s.getUserFromSession(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
			"error":         "Session expired or invalid",
		})
		return
	}

	role := "user"
	if user.IsAdmin() {
		role = "admin"
	}

	response := map[string]interface{}{
		"authenticated":      true,
		"id":                 user.Id,
		"email":              user.Email,
		"name":               user.Name,
		"role":               role,
		"storage_used_mb":    user.StorageUsedMB,
		"storage_quota_mb":   user.StorageQuotaMB,
		"server_version":     s.config.Version,
		"two_factor_enabled": user.TOTPEnabled,
	}
	_ = json.NewEncoder(w).Encode(response)
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

	// Parse multipart form (since FormData sends multipart/form-data)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// Fallback to regular form parsing
		if err := r.ParseForm(); err != nil {
			log.Printf("ERROR: Failed to parse form: %v", err)
			s.sendError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
	}

	fileID := r.FormValue("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	expirationDays, _ := strconv.Atoi(r.FormValue("expiration_days"))
	downloadsLimit, _ := strconv.Atoi(r.FormValue("downloads_limit"))
	teamIDStr := r.FormValue("team_id")
	fileComment := r.FormValue("file_comment")
	requireAuth := r.FormValue("require_auth") == "true"
	filePassword := r.FormValue("file_password")

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

	// Update comment if provided
	if err := database.DB.UpdateFileComment(fileID, fileComment); err != nil {
		log.Printf("Warning: Failed to update file comment: %v", err)
		// Don't fail the request, just log the error
	}

	// Update require auth setting
	if err := database.DB.UpdateFileRequireAuth(fileID, requireAuth); err != nil {
		log.Printf("Warning: Failed to update require auth: %v", err)
		// Don't fail the request, just log the error
	}

	// Update password (empty string will clear the password)
	if err := database.DB.UpdateFilePassword(fileID, filePassword); err != nil {
		log.Printf("Warning: Failed to update file password: %v", err)
		// Don't fail the request, just log the error
	}

	// Share to team if team_id is provided
	if teamIDStr != "" {
		teamID, err := strconv.Atoi(teamIDStr)
		if err == nil {
			// Check if user is team member
			isMember, err := database.DB.IsTeamMember(teamID, user.Id)
			if err == nil && isMember {
				// Share file to team
				if err := database.DB.ShareFileToTeam(fileID, teamID, user.Id); err != nil {
					// Log error but don't fail the request (file was already updated)
					log.Printf("Warning: Failed to share file to team: %v", err)
				} else {
					log.Printf("File %s shared to team %d by user %d", fileInfo.Name, teamID, user.Id)
				}
			} else {
				log.Printf("Warning: User %d is not a member of team %d, skipping team share", user.Id, teamID)
			}
		}
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

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "FILE_DELETED",
		EntityType: "File",
		EntityID:   fileID,
		Details:    fmt.Sprintf("{\"file_name\":\"%s\",\"size\":%d}", fileInfo.Name, fileInfo.SizeBytes),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

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

	// Get branding config for email styling
	brandingConfig, _ := database.DB.GetBrandingConfig()
	primaryColor := brandingConfig["branding_primary_color"]
	if primaryColor == "" {
		primaryColor = "#2563eb"
	}
	companyName := brandingConfig["branding_company_name"]
	if companyName == "" {
		companyName = s.config.CompanyName
	}

	// Construct email content
	subject := fmt.Sprintf("%s has shared a file with you via %s", user.Name, companyName)

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
									<strong>%s</strong> has sent you a file. Click the blue button below to download it.
								</p>
							</div>

							<!-- File Info Box -->
							<div style="background-color: #f8fafc; border: 2px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 20px;">
								<h3 style="margin: 0 0 10px 0; color: #1e3a5f; font-size: 18px;">📄 %s</h3>
								<p style="margin: 0; color: #64748b; font-size: 14px;">Size: %.2f MB</p>
							</div>

							%s
							%s

							<!-- BIG BLUE DOWNLOAD BUTTON -->
							<table width="100%%" cellpadding="0" cellspacing="0" style="margin: 30px 0;">
								<tr>
									<td align="center">
										<a href="%s" style="display: inline-block; background-color: #2563eb; color: #ffffff; padding: 20px 50px; text-decoration: none; border-radius: 8px; font-size: 20px; font-weight: bold; border: 3px solid #1d4ed8; box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4); text-transform: uppercase; letter-spacing: 1px;">
											⬇️ DOWNLOAD FILE
										</a>
									</td>
								</tr>
							</table>

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
	`, companyName, user.Name, fileInfo.Name, float64(fileInfo.SizeBytes)/(1024*1024),
		func() string {
			if fileInfo.Comment != "" {
				return fmt.Sprintf(`
							<!-- File Description -->
							<div style="background-color: #f0f9ff; border-left: 4px solid #2563eb; padding: 15px; margin-bottom: 15px; border-radius: 0 8px 8px 0;">
								<p style="margin: 0 0 8px 0; color: #1d4ed8; font-weight: 600; font-size: 14px;">📝 File Description:</p>
								<p style="margin: 0; color: #334155; font-size: 14px; line-height: 1.5;">%s</p>
							</div>`, template.HTMLEscapeString(fileInfo.Comment))
			}
			return ""
		}(),
		func() string {
			if request.Message != "" {
				return fmt.Sprintf(`
							<!-- Message from sender -->
							<div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin-bottom: 15px; border-radius: 0 8px 8px 0;">
								<p style="margin: 0 0 8px 0; color: #92400e; font-weight: 600; font-size: 14px;">💬 Message from %s:</p>
								<p style="margin: 0; color: #78350f; font-size: 14px; line-height: 1.5;">%s</p>
							</div>`, user.Name, template.HTMLEscapeString(request.Message))
			}
			return ""
		}(),
		fileURL, fileURL, fileURL, companyName)

	textBody := fmt.Sprintf(
		`FILE SHARED WITH YOU
====================

WHAT IS THIS?
%s has sent you a file. Use the link below to download it.

FILE: %s
SIZE: %.2f MB

%s%sDOWNLOAD YOUR FILE:
%s

---
This is an automated message from %s`,
		user.Name, fileInfo.Name, float64(fileInfo.SizeBytes)/(1024*1024),
		func() string {
			if fileInfo.Comment != "" {
				return fmt.Sprintf("FILE DESCRIPTION:\n%s\n\n", fileInfo.Comment)
			}
			return ""
		}(),
		func() string {
			if request.Message != "" {
				return fmt.Sprintf("MESSAGE FROM %s:\n%s\n\n", user.Name, request.Message)
			}
			return ""
		}(),
		fileURL, companyName,
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

	// Audit log for email sent
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     database.ActionEmailSent,
		EntityType: database.EntityFile,
		EntityID:   fileInfo.Id,
		Details: database.CreateAuditDetails(map[string]interface{}{
			"recipient":   request.Recipient,
			"file_name":   fileInfo.Name,
			"file_size":   fileInfo.SizeBytes,
			"has_message": request.Message != "",
		}),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Success:   true,
	})

	log.Printf("File link emailed: %s to %s by user %d", fileInfo.Name, request.Recipient, user.Id)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "Email sent successfully",
	})
}

// renderUserDashboard renders the user dashboard HTML
func (s *Server) renderUserDashboard(w http.ResponseWriter, tr *i18n.Translator, userModel interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	user := userModel.(*models.User)

	// Get dashboard style preference
	dashboardStyle, _ := database.DB.GetConfigValue("dashboard_style")
	if dashboardStyle == "" {
		dashboardStyle = "colorful" // Default to colorful
	}

	// Get joke of the day
	joke := models.GetJokeOfTheDay()

	// Get user's files (including team files)
	files, err := database.DB.GetFilesByUserWithTeams(user.Id)
	if err != nil {
		log.Printf("Warning: Failed to get files with teams for user %d: %v", user.Id, err)
		// Fallback to user's own files only
		files, _ = database.DB.GetFilesByUser(user.Id)
	}

	// Get team names for all files
	fileIds := make([]string, len(files))
	for i, f := range files {
		fileIds[i] = f.Id
	}
	fileTeams, err := database.DB.GetFileTeamNames(fileIds)
	if err != nil {
		log.Printf("Warning: Failed to get team names for files: %v", err)
		fileTeams = make(map[string][]string) // Empty map as fallback
	}

	// Collect all unique team names for the team filter dropdown
	allTeamNames := make(map[string]bool)
	for _, teams := range fileTeams {
		for _, teamName := range teams {
			allTeamNames[teamName] = true
		}
	}
	// Convert to sorted slice
	var uniqueTeamNames []string
	for teamName := range allTeamNames {
		uniqueTeamNames = append(uniqueTeamNames, teamName)
	}
	// Sort alphabetically
	for i := 0; i < len(uniqueTeamNames); i++ {
		for j := i + 1; j < len(uniqueTeamNames); j++ {
			if uniqueTeamNames[i] > uniqueTeamNames[j] {
				uniqueTeamNames[i], uniqueTeamNames[j] = uniqueTeamNames[j], uniqueTeamNames[i]
			}
		}
	}

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

	showPasswordLabel := "👁️ " + tr.T("dashboard.show_password")

	html := `<!DOCTYPE html>
<html lang="` + string(tr.Lang()) + `">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>` + tr.T("dashboard.page_title") + ` - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: ` + func() string {
		if dashboardStyle == "plain" {
			return "#ffffff"
		}
		return "#f5f5f5"
	}() + `;
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
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08);
            border-left: 4px solid ` + s.getPrimaryColor() + `;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .stat-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.12);
        }
        .stat-card h3 {
            color: #888;
            font-size: 13px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 12px;
        }
        .stat-card .value {
            font-size: 36px;
            font-weight: 700;
            color: #1a1a2e;
        }
        .stat-card .progress {
            margin-top: 12px;
            height: 6px;
            background: #f0f0f0;
            border-radius: 3px;
            overflow: hidden;
        }
        .stat-card .progress-bar {
            height: 100%;
            background: linear-gradient(90deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            transition: width 0.3s;
            border-radius: 3px;
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
            border-bottom: 3px solid ` + s.getPrimaryColor() + `;
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
            flex-wrap: wrap;
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
        .joke-section {
            margin: 30px 0;
            padding: 25px 30px;
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            border-radius: 15px;
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.15);
        }
        .joke-title {
            color: rgba(255,255,255,0.9);
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 12px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
        }
        .joke-text {
            color: white;
            font-size: 17px;
            line-height: 1.6;
            font-weight: 500;
            font-style: italic;
        }

        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .stats {
                grid-template-columns: 1fr !important;
            }
            .upload-zone {
                padding: 40px 20px !important;
            }
            .file-item {
                flex-direction: column;
                align-items: flex-start !important;
                gap: 12px;
            }
            .file-actions {
                width: 100%;
                justify-content: stretch;
            }
            .file-actions button {
                flex: 1;
            }
        }
    </style>
</head>
<body data-user-id="` + fmt.Sprintf("%d", user.Id) + `">
    ` + s.getHeaderHTML(user, user.IsAdmin(), tr) + `
    <div class="container">
        <div class="joke-section">
            <div class="joke-title">💡 ` + tr.T("dashboard.joke_heading") + `</div>
            <div class="joke-text">` + joke.Text + `</div>
        </div>

        <div class="stats">
            <div class="stat-card">
                <h3>` + tr.T("dashboard.storage_used_heading") + `</h3>
                <div class="value">` + storageUsedGB + ` GB</div>
                <div class="progress">
                    <div class="progress-bar" style="width: ` + fmt.Sprintf("%d", storagePercent) + `%"></div>
                </div>
                <p style="margin-top: 8px; color: #999; font-size: 14px;">` + tr.T("dashboard.storage_of_gb", "used", storageUsedGB, "quota", storageQuotaGB) + `</p>
            </div>
            <div class="stat-card">
                <h3>` + tr.T("dashboard.active_files_heading") + `</h3>
                <div class="value">` + fmt.Sprintf("%d", activeFileCount) + `</div>
            </div>
            <div class="stat-card">
                <h3>` + tr.T("dashboard.total_downloads_heading") + `</h3>
                <div class="value">` + fmt.Sprintf("%d", totalDownloads) + `</div>
            </div>
        </div>

        <!-- Upload Form -->
        <div class="upload-section">
            <h2 style="margin-bottom: 20px; color: #333;">` + tr.T("dashboard.upload_file_heading") + `</h2>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="upload-zone" id="uploadZone" onclick="document.getElementById('fileInput').click()">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <h3>` + tr.T("dashboard.drop_zone_heading") + `</h3>
                    <p>` + tr.T("dashboard.max_file_size") + `</p>
                    <input type="file" id="fileInput" name="file">
                </div>

                <div class="upload-options" id="uploadOptions" style="display: none;">
                    <h3 style="margin-bottom: 16px; color: #333;">` + tr.T("dashboard.upload_settings_heading") + `</h3>

                    <div class="form-group" style="background: #f0f9ff; padding: 15px; border-radius: 8px; border: 2px solid #3b82f6; margin-bottom: 20px;">
                        <label for="fileComment" style="color: #1d4ed8; font-weight: 600;">💬 ` + tr.T("dashboard.description_label_recommended") + `</label>
                        <textarea id="fileComment" name="file_comment" rows="3" maxlength="1000" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.description_placeholder")) + `" style="width: 100%; padding: 10px; border: 2px solid #93c5fd; border-radius: 6px; font-size: 14px; font-family: inherit; resize: vertical; margin-top: 8px;"></textarea>
                        <p style="color: #1e40af; font-size: 12px; margin-top: 4px;">
                            ` + tr.T("dashboard.description_help") + `
                        </p>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="expireDate">📅 ` + tr.T("dashboard.expiration_date_label") + `</label>
                            <input type="date" id="expireDate" name="expire_date">
                            <label style="margin-top: 8px;">
                                <input type="checkbox" id="unlimitedTime" name="unlimited_time" checked> ` + tr.T("dashboard.never_expire_time_label") + `
                            </label>
                        </div>

                        <div class="form-group">
                            <label for="downloadsLimit">⬇️ ` + tr.T("dashboard.download_limit_label") + `</label>
                            <input type="number" id="downloadsLimit" name="downloads_limit" min="1" value="10">
                            <label style="margin-top: 8px;">
                                <input type="checkbox" id="unlimitedDownloads" name="unlimited_downloads" checked> ` + tr.T("dashboard.unlimited_downloads_label") + `
                            </label>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>🔗 ` + tr.T("dashboard.link_type_label") + `</label>
                        <div style="display: flex; gap: 16px; margin-top: 8px;">
                            <label style="display: flex; align-items: center; gap: 8px;">
                                <input type="radio" name="link_type" value="splash" checked>
                                <span>` + tr.T("dashboard.splash_page_option") + `</span>
                            </label>
                            <label style="display: flex; align-items: center; gap: 8px;">
                                <input type="radio" name="link_type" value="direct">
                                <span>` + tr.T("dashboard.direct_download_option") + `</span>
                            </label>
                        </div>
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">
                            ` + tr.T("dashboard.splash_page_help") + `
                        </p>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="requireAuth" name="require_auth">
                            🔒 ` + tr.T("dashboard.require_auth_label") + `
                        </label>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="enablePassword" onchange="togglePasswordField()">
                            🔐 ` + tr.T("dashboard.password_protect_label") + `
                        </label>
                        <div id="passwordFieldContainer" style="display: none; margin-top: 12px;">
                            <input type="text" id="filePassword" name="file_password" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.password_placeholder")) + `" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                            <p style="color: #666; font-size: 12px; margin-top: 4px;">
                                ` + tr.T("dashboard.password_help") + `
                            </p>
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="sendToEmail">📧 ` + tr.T("dashboard.send_email_label") + `</label>
                        <input type="email" id="sendToEmail" name="send_to_email" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.recipient_email_placeholder")) + `" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">
                            ` + tr.T("dashboard.send_email_help") + `
                        </p>
                    </div>

                    <div class="form-group">
                        <label>👥 ` + tr.T("dashboard.share_teams_label") + `</label>
                        <div id="teamSelectContainer" style="border: 2px solid #e0e0e0; border-radius: 6px; padding: 12px; max-height: 150px; overflow-y: auto; background: #fafafa;">
                            <div style="color: #999; font-style: italic;">` + tr.T("dashboard.loading_teams") + `</div>
                        </div>
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">
                            ` + tr.T("dashboard.share_teams_help") + `
                        </p>
                    </div>

                    <button type="submit" id="uploadButton" style="
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        gap: 12px;
                        padding: 20px 40px;
                        font-size: 20px;
                        min-width: 250px;
                        background: linear-gradient(135deg, #10b981 0%, #059669 100%);
                        color: white;
                        border: none;
                        border-radius: 12px;
                        cursor: pointer;
                        box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
                        transition: all 0.3s ease;
                        margin-bottom: 15px;
                    " onmouseover="this.style.transform='translateY(-2px)'; this.style.boxShadow='0 6px 16px rgba(16, 185, 129, 0.4)'" onmouseout="this.style.transform='translateY(0)'; this.style.boxShadow='0 4px 12px rgba(16, 185, 129, 0.3)'">
                        <span style="font-size: 32px;">📤</span>
                        <span style="font-size: 20px; font-weight: 700; letter-spacing: 0.5px;">` + tr.T("dashboard.upload_file_heading") + `</span>
                    </button>
                    <button type="button" onclick="resetUploadForm()" style="
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        gap: 10px;
                        padding: 15px 30px;
                        font-size: 18px;
                        background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
                        color: white;
                        border: none;
                        border-radius: 10px;
                        cursor: pointer;
                        box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
                        transition: all 0.3s ease;
                    " onmouseover="this.style.transform='translateY(-2px)'; this.style.boxShadow='0 6px 16px rgba(239, 68, 68, 0.4)'" onmouseout="this.style.transform='translateY(0)'; this.style.boxShadow='0 4px 12px rgba(239, 68, 68, 0.3)'">
                        ✖️ ` + tr.T("common.cancel") + `
                    </button>
                </div>
            </form>
        </div>

        <!-- File Request Section -->
        <div class="file-request-section" style="background: white; padding: 30px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); margin-bottom: 40px;">
            <h2 style="margin-bottom: 16px; color: #333;">📥 ` + tr.T("dashboard.file_request_heading") + `</h2>
            <p style="color: #666; margin-bottom: 12px;">` + tr.T("dashboard.file_request_description") + `</p>
            <div style="background: #e3f2fd; border-left: 4px solid #2196f3; padding: 12px 16px; margin-bottom: 20px; border-radius: 4px;">
                <p style="color: #1976d2; font-size: 13px; margin: 0;">
                    🔒 ` + tr.T("dashboard.file_request_security_notice") + `
                </p>
            </div>
            <button onclick="showCreateRequestModal()" style="padding: 12px 24px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                ➕ ` + tr.T("dashboard.create_upload_request") + `
            </button>
            <div id="requestsList" style="margin-top: 20px;"></div>
        </div>

        <!-- File Request Modal -->
        <div id="fileRequestModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 10000; align-items: center; justify-content: center;">
            <div style="background: white; border-radius: 12px; padding: 32px; max-width: 500px; width: 90%; box-shadow: 0 8px 32px rgba(0,0,0,0.2);">
                <h2 style="margin-bottom: 24px; color: #333;">` + tr.T("dashboard.create_upload_request") + `</h2>
                <form id="fileRequestForm" onsubmit="submitFileRequest(event)">
                    <div style="margin-bottom: 20px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">` + tr.T("dashboard.request_title_label") + `</label>
                        <input type="text" id="requestTitle" required placeholder="` + template.HTMLEscapeString(tr.T("dashboard.request_title_placeholder")) + `" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">` + tr.T("dashboard.request_title_help") + `</p>
                    </div>

                    <div style="margin-bottom: 20px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">` + tr.T("dashboard.message_optional_label") + `</label>
                        <textarea id="requestMessage" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.request_message_placeholder")) + `" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px; min-height: 80px; resize: vertical;"></textarea>
                    </div>

                    <div style="margin-bottom: 20px; padding: 14px; background: #fff3cd; border: 2px solid #ffc107; border-radius: 8px;">
                        <p style="color: #856404; font-size: 13px; margin: 0; line-height: 1.5;">
                            ⏰ ` + tr.T("dashboard.request_security_notice") + `
                        </p>
                    </div>

                    <div style="margin-bottom: 24px;">
                        <label style="display: block; margin-bottom: 8px; color: #333; font-weight: 600;">` + tr.T("dashboard.request_max_size_label") + `</label>
                        <input type="number" id="requestMaxSize" min="0.1" max="15" step="0.1" value="1" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                        <p style="color: #666; font-size: 12px; margin-top: 4px;">` + tr.T("dashboard.request_max_size_help") + `</p>
                    </div>

                    <div style="margin-bottom: 24px; background: #fff9e6; padding: 16px; border-radius: 8px; border: 3px solid #ff9800;">
                        <label style="display: block; margin-bottom: 8px; color: #e65100; font-weight: 700; font-size: 16px;">📧 ` + tr.T("dashboard.request_recipient_email_label") + `</label>
                        <input type="email" id="requestRecipientEmail" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.recipient_email_placeholder")) + `" style="width: 100%; padding: 12px; border: 3px solid #ff9800; border-radius: 6px; font-size: 14px; background: white;">
                        <p style="color: #e65100; font-size: 13px; margin-top: 8px; font-weight: 600;">` + tr.T("dashboard.request_recipient_email_help") + `</p>
                    </div>

                    <div style="display: flex; gap: 12px;">
                        <button type="submit" style="flex: 1; padding: 12px 24px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                            ` + tr.T("dashboard.create_request_button") + `
                        </button>
                        <button type="button" onclick="closeFileRequestModal()" style="flex: 1; padding: 12px 24px; background: #f5f5f5; color: #333; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;">
                            ` + tr.T("common.cancel") + `
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <div class="files-section">
            <div class="files-header">
                <h2>` + tr.T("dashboard.my_files") + `</h2>
                <div class="file-tabs" style="margin-top: 16px; display: flex; gap: 12px; border-bottom: 2px solid #e0e0e0; padding-bottom: 8px; align-items: center;">
                    <button class="file-tab active" onclick="filterFiles('all')" data-filter="all" style="background: none; border: none; padding: 8px 16px; font-size: 14px; font-weight: 600; cursor: pointer; border-bottom: 3px solid ` + s.getPrimaryColor() + `; color: ` + s.getPrimaryColor() + `;">` + tr.T("dashboard.tab_all_files") + `</button>
                    <button class="file-tab" onclick="filterFiles('my')" data-filter="my" style="background: none; border: none; padding: 8px 16px; font-size: 14px; font-weight: 500; cursor: pointer; border-bottom: 3px solid transparent; color: #666;">` + tr.T("dashboard.my_files") + `</button>
                    <button class="file-tab" onclick="filterFiles('team')" data-filter="team" style="background: none; border: none; padding: 8px 16px; font-size: 14px; font-weight: 500; cursor: pointer; border-bottom: 3px solid transparent; color: #666;">` + tr.T("dashboard.tab_team_files") + `</button>
                    <select id="teamFilter" onchange="filterByTeam(this.value)" style="display: none; margin-left: auto; padding: 6px 12px; border: 2px solid ` + s.getPrimaryColor() + `; border-radius: 6px; font-size: 13px; background: white; cursor: pointer;">
                        <option value="">` + tr.T("dashboard.all_teams_option") + `</option>` + func() string {
		teamOptionsHTML := ""
		for _, teamName := range uniqueTeamNames {
			teamOptionsHTML += fmt.Sprintf(`<option value="%s">%s</option>`, template.HTMLEscapeString(teamName), template.HTMLEscapeString(teamName))
		}
		return teamOptionsHTML
	}() + `
                    </select>
                </div>
                <!-- Search and Sort Controls -->
                <div style="margin-top: 20px; display: flex; gap: 12px; flex-wrap: wrap; align-items: center;">
                    <input type="text" id="fileSearch" placeholder="🔍 ` + template.HTMLEscapeString(tr.T("dashboard.search_files_placeholder")) + `" onkeyup="searchAndSortFiles()" style="flex: 1; min-width: 250px; padding: 10px 15px; border: 2px solid #e0e0e0; border-radius: 8px; font-size: 14px; transition: border-color 0.3s;">
                    <select id="fileSort" onchange="searchAndSortFiles()" style="padding: 10px 15px; border: 2px solid ` + s.getPrimaryColor() + `; border-radius: 8px; font-size: 14px; background: white; cursor: pointer; font-weight: 500;">
                        <option value="name-asc">📝 ` + tr.T("dashboard.sort_name_asc") + `</option>
                        <option value="name-desc">📝 ` + tr.T("dashboard.sort_name_desc") + `</option>
                        <option value="date-desc" selected>📅 ` + tr.T("dashboard.sort_date_desc") + `</option>
                        <option value="date-asc">📅 ` + tr.T("dashboard.sort_date_asc") + `</option>
                        <option value="downloads-desc">📊 ` + tr.T("dashboard.sort_downloads_desc") + `</option>
                        <option value="downloads-asc">📊 ` + tr.T("dashboard.sort_downloads_asc") + `</option>
                        <option value="size-desc">📦 ` + tr.T("dashboard.sort_size_desc") + `</option>
                        <option value="size-asc">📦 ` + tr.T("dashboard.sort_size_asc") + `</option>
                    </select>
                    <select id="perPageSelect" onchange="changePerPage()" style="padding: 10px 15px; border: 2px solid ` + s.getPrimaryColor() + `; border-radius: 8px; font-size: 14px; background: white; cursor: pointer; font-weight: 500;">
                        <option value="5">` + tr.T("dashboard.per_page", "count", "5") + `</option>
                        <option value="25" selected>` + tr.T("dashboard.per_page", "count", "25") + `</option>
                        <option value="50">` + tr.T("dashboard.per_page", "count", "50") + `</option>
                        <option value="100">` + tr.T("dashboard.per_page", "count", "100") + `</option>
                        <option value="200">` + tr.T("dashboard.per_page", "count", "200") + `</option>
                        <option value="250">` + tr.T("dashboard.per_page", "count", "250") + `</option>
                    </select>
                </div>
                <!-- File counter and pagination -->
                <div style="margin-top: 15px; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px;">
                    <div id="fileCounter" style="font-weight: 600; color: #333; font-size: 14px;">
                        ` + tr.T("dashboard.file_counter", "visible", `<span id="visibleCount">0</span>`, "total", `<span id="totalCount">0</span>`) + `
                    </div>
                    <div id="paginationControls" style="display: flex; gap: 8px; align-items: center;">
                        <button onclick="prevPage()" id="prevBtn" style="padding: 6px 12px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 13px;">← ` + tr.T("dashboard.prev_button") + `</button>
                        <span id="pageInfo" style="font-size: 14px; color: #666; min-width: 80px; text-align: center;">` + tr.T("dashboard.page_info", "page", "1", "total", "1") + `</span>
                        <button onclick="nextPage()" id="nextBtn" style="padding: 6px 12px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 13px;">` + tr.T("dashboard.next_button") + ` →</button>
                    </div>
                </div>
            </div>`

	if len(files) == 0 {
		html += `
            <div class="empty-state">
                ` + tr.T("dashboard.no_files") + `
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
			status := tr.T("common.active")
			statusColor := "#4caf50"

			if !f.UnlimitedDownloads && f.DownloadsRemaining <= 0 {
				status = tr.T("dashboard.status_expired_downloads")
				statusColor = "#f44336"
			} else if !f.UnlimitedTime && f.ExpireAt > 0 && f.ExpireAt < time.Now().Unix() {
				status = tr.T("dashboard.status_expired_time")
				statusColor = "#f44336"
			}

			expiryInfo := ""
			if f.UnlimitedTime && f.UnlimitedDownloads {
				expiryInfo = tr.T("dashboard.never_expires")
			} else if f.UnlimitedTime {
				expiryInfo = tr.T("dashboard.downloads_remaining", "count", fmt.Sprintf("%d", f.DownloadsRemaining))
			} else if f.UnlimitedDownloads {
				expiryInfo = tr.T("dashboard.expires_colon", "date", f.ExpireAtString)
			} else {
				expiryInfo = tr.T("dashboard.downloads_left_expires", "count", fmt.Sprintf("%d", f.DownloadsRemaining), "date", f.ExpireAtString)
			}

			authBadge := ""
			if f.RequireAuth {
				authBadge = `<span style="background: #2196f3; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">🔒 ` + tr.T("dashboard.auth_required_badge") + `</span>`
			}

			passwordBadge := ""
			if f.FilePasswordPlain != "" {
				passwordBadge = `<span style="background: #9c27b0; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">🔐 ` + tr.T("dashboard.password_protected_badge") + `</span>`
			}

			// Team badges
			teamBadges := ""
			isTeamFile := false
			if teams, ok := fileTeams[f.Id]; ok && len(teams) > 0 {
				isTeamFile = true
				if len(teams) == 1 {
					// Single team - show name directly
					teamBadges = fmt.Sprintf(`<span style="background: #ff9800; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">👥 %s</span>`, template.HTMLEscapeString(teams[0]))
				} else {
					// Multiple teams - show count with tooltip
					teamsListHTML := ""
					for i, teamName := range teams {
						if i > 0 {
							teamsListHTML += ", "
						}
						teamsListHTML += template.HTMLEscapeString(teamName)
					}
					teamBadges = fmt.Sprintf(`<span style="background: #ff9800; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px; cursor: help;" title="%s">👥 %s</span>`,
						tr.T("dashboard.shared_with_tooltip", "teams", teamsListHTML),
						tr.T("dashboard.teams_count", "count", fmt.Sprintf("%d", len(teams))))
				}
			}

			// Determine file type (my file vs team file)
			fileType := "my"
			if isTeamFile && f.UserId != user.Id {
				fileType = "team"
			} else if isTeamFile && f.UserId == user.Id {
				fileType = "both" // Own file shared with team
			}

			passwordDisplay := ""
			if f.FilePasswordPlain != "" {
				passwordDisplay = fmt.Sprintf(`<p style="margin-top: 8px;"><strong>🔐 %s:</strong> <span id="password-%s" style="cursor: pointer; color: #9c27b0; text-decoration: underline;" onclick="togglePasswordVisibility('%s', '%s')">%s</span></p>`,
					tr.T("common.password"), f.Id, f.Id, template.JSEscapeString(f.FilePasswordPlain), showPasswordLabel)
			}

			commentDisplay := ""
			if f.Comment != "" {
				commentDisplay = fmt.Sprintf(`<p style="margin-top: 8px; padding: 12px; background: #fff3cd; border-left: 4px solid %s; border-radius: 4px; color: #333; font-weight: 500;"><strong style="font-weight: 700;">📝 %s</strong> %s</p>`,
					s.getPrimaryColor(), tr.T("dashboard.note_label"), template.HTMLEscapeString(f.Comment))
			}

			// Create data-teams attribute for filtering
			dataTeamsAttr := ""
			if teams, ok := fileTeams[f.Id]; ok && len(teams) > 0 {
				// Join team names with comma for the attribute
				teamsJSON := ""
				for i, t := range teams {
					if i > 0 {
						teamsJSON += ","
					}
					teamsJSON += template.HTMLEscapeString(t)
				}
				dataTeamsAttr = teamsJSON
			}

			// Get file extension
			fileExt := filepath.Ext(f.Name)
			if len(fileExt) > 0 && fileExt[0] == '.' {
				fileExt = fileExt[1:] // Remove leading dot
			}

			downloadedTimesText := tr.T("dashboard.downloaded_times", "count", fmt.Sprintf("%d", f.DownloadCount))

			html += fmt.Sprintf(`
                <li class="file-item" data-file-type="%s" data-teams="%s" data-filename="%s" data-extension="%s" data-size="%d" data-timestamp="%d" data-downloads="%d" data-comment="%s">
                    <div class="file-info">
                        <h3 title="%s">
                            <span style="display: inline-block; max-width: 600px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: bottom;">📄 %s</span>%s%s%s
                        </h3>
                        %s
                        <p>%s • %s • %s</p>
                        <p style="color: %s;">%s: %s</p>
                        %s
                        <div class="link-display">
                            <h4>🌐 %s</h4>
                            <div class="link-box">
                                <a href="%s" target="_blank">%s</a>
                                <button class="btn btn-primary" onclick="copyToClipboard('%s', this)" style="font-size: 11px; padding: 4px 8px;">📋 %s</button>
                            </div>
                            <h4>⬇️ %s</h4>
                            <div class="link-box">
                                <a href="%s" target="_blank">%s</a>
                                <button class="btn btn-primary" onclick="copyToClipboard('%s', this)" style="font-size: 11px; padding: 4px 8px;">📋 %s</button>
                            </div>
                        </div>
                        <div class="file-actions" style="margin-top: 16px; display: flex; gap: 8px; flex-wrap: wrap;">
                            <button class="btn btn-secondary" onclick="showDownloadHistory('%s', '%s')" title="%s" style="flex: 0 0 auto;">
                                📊 %s
                            </button>
                            <button class="btn btn-primary" onclick="showEmailModal('%s', '%s', '%s')" title="%s" style="background: #007bff; flex: 0 0 auto;">
                                📧 %s
                            </button>
                            <button class="btn btn-secondary" onclick="showEditModal('%s', '%s', %d, %d, %t, %t, '%s', %t, '%s')" title="%s" style="flex: 0 0 auto;">
                                ✏️ %s
                            </button>
                            <button class="btn btn-danger" onclick="deleteFile('%s', '%s')" style="flex: 0 0 auto; background: #dc3545 !important; color: white;">
                                🗑️ %s
                            </button>
                        </div>
                    </div>
                </li>`, fileType, dataTeamsAttr, template.HTMLEscapeString(f.Name), fileExt, f.SizeBytes, f.UploadDate, f.DownloadCount, template.HTMLEscapeString(f.Comment), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), authBadge, passwordBadge, teamBadges, commentDisplay, f.Size, downloadedTimesText, expiryInfo, statusColor, tr.T("common.status"), status, passwordDisplay,
				tr.T("dashboard.splash_link_heading"),
				splashURL, splashURL, splashURLEscaped, tr.T("dashboard.copy_button"),
				tr.T("dashboard.direct_link_heading"),
				directURL, directURL, directURLEscaped, tr.T("dashboard.copy_button"),
				f.Id, template.JSEscapeString(f.Name), tr.T("dashboard.view_history_title"), tr.T("dashboard.history_button"),
				f.Id, template.JSEscapeString(f.Name), template.JSEscapeString(splashURL), tr.T("dashboard.send_email_title"), tr.T("dashboard.email_button"),
				f.Id, template.JSEscapeString(f.Name), f.DownloadsRemaining, f.ExpireAt, f.UnlimitedDownloads, f.UnlimitedTime, template.JSEscapeString(f.Comment), f.RequireAuth, template.JSEscapeString(f.FilePasswordPlain), tr.T("dashboard.edit_file_title"), tr.T("common.edit"),
				f.Id, template.JSEscapeString(f.Name), tr.T("common.delete"))
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
            <h2 style="margin-bottom: 24px; color: #333;">` + tr.T("dashboard.email_modal_heading") + `</h2>
            <input type="hidden" id="emailFileId">
            <p style="margin-bottom: 20px; color: #666;">` + tr.T("dashboard.email_modal_sending_for") + ` <strong id="emailFileName"></strong></p>
            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; color: #555; font-weight: 500;">` + tr.T("dashboard.recipient_email_label") + `</label>
                <input type="email" id="emailRecipient" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.recipient_email_placeholder")) + `" style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px;">
            </div>
            <div style="margin-bottom: 24px;">
                <label style="display: block; margin-bottom: 8px; color: #555; font-weight: 500;">` + tr.T("dashboard.message_optional_label") + `</label>
                <textarea id="emailMessage" rows="4" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.personal_message_placeholder")) + `" style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px; resize: vertical;"></textarea>
            </div>
            <div style="display: flex; gap: 12px; justify-content: flex-end;">
                <button onclick="closeEmailModal()" class="btn btn-secondary" style="padding: 10px 20px;">` + tr.T("common.cancel") + `</button>
                <button onclick="sendEmailLink()" class="btn btn-primary" style="padding: 10px 20px; background: #007bff;">` + tr.T("dashboard.send_email_button") + `</button>
            </div>
        </div>
    </div>

    <!-- Edit File Modal -->
    <div id="editModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 500px; width: 90%;">
            <h2 style="margin-bottom: 24px; color: #333;">` + tr.T("dashboard.edit_modal_heading") + `</h2>

            <input type="hidden" id="editFileId">

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.file_label_colon") + `</label>
                <p id="editFileName" style="color: #666; font-weight: 600;"></p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editUnlimitedTime" onchange="toggleEditTimeLimit()">
                    ` + tr.T("dashboard.never_expire_forever_label") + `
                </label>
            </div>

            <div id="editTimeLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.days_until_expiration_label") + `</label>
                <input type="number" id="editExpirationDays" value="7" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">` + tr.T("dashboard.days_until_expiration_help") + `</p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editUnlimitedDownloads" onchange="toggleEditDownloadLimit()">
                    ` + tr.T("dashboard.unlimited_downloads_label") + `
                </label>
            </div>

            <div id="editDownloadLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.downloads_remaining_label") + `</label>
                <input type="number" id="editDownloadsLimit" value="5" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">💬 ` + tr.T("dashboard.description_label_colon") + `</label>
                <textarea id="editFileComment" rows="3" maxlength="1000" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.description_placeholder_short")) + `" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px; font-family: inherit; resize: vertical;"></textarea>
                <p style="font-size: 12px; color: #999; margin-top: 4px;">` + tr.T("dashboard.description_help_short") + `</p>
            </div>

            <div style="margin-bottom: 20px; padding-top: 20px; border-top: 2px solid #e0e0e0;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editRequireAuth">
                    🔒 ` + tr.T("dashboard.require_auth_download_label") + `
                </label>
                <p style="font-size: 12px; color: #999; margin-top: 4px; margin-left: 24px;">` + tr.T("dashboard.require_auth_help") + `</p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="editEnablePassword" onchange="toggleEditPasswordField()">
                    🔐 ` + tr.T("dashboard.password_protect_label") + `
                </label>
                <div id="editPasswordFieldContainer" style="display: none; margin-top: 12px; margin-left: 24px;">
                    <label style="display: block; margin-bottom: 8px; color: #555; font-weight: 500;">` + tr.T("common.password") + `:</label>
                    <input type="text" id="editFilePassword" placeholder="` + template.HTMLEscapeString(tr.T("dashboard.password_placeholder")) + `" maxlength="100" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px;">
                    <p style="font-size: 12px; color: #999; margin-top: 4px;">` + tr.T("dashboard.password_help") + `</p>
                </div>
            </div>

            <div style="margin-bottom: 20px; padding-top: 20px; border-top: 2px solid #e0e0e0;">
                <label style="display: block; margin-bottom: 12px; font-weight: 500;">👥 ` + tr.T("dashboard.team_sharing_label") + `</label>

                <!-- Current teams -->
                <div id="editCurrentTeams" style="margin-bottom: 16px;">
                    <div style="color: #999; font-style: italic; font-size: 14px;">` + tr.T("dashboard.loading_current_teams") + `</div>
                </div>

                <!-- Add new team -->
                <div style="background: #f5f5f5; padding: 12px; border-radius: 6px;">
                    <label style="display: block; margin-bottom: 8px; font-size: 14px; font-weight: 500;">` + tr.T("dashboard.add_to_team_label") + `</label>
                    <select id="editTeamSelect" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; background: white;">
                        <option value="">` + tr.T("dashboard.select_team_option") + `</option>
                    </select>
                    <button onclick="addTeamToFile()" style="margin-top: 8px; padding: 8px 16px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 4px; font-size: 13px; cursor: pointer; width: 100%;">
                        ➕ ` + tr.T("dashboard.add_team_button") + `
                    </button>
                </div>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="saveFileEdit()" style="flex: 1; padding: 14px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    ` + tr.T("dashboard.save_changes_button") + `
                </button>
                <button onclick="closeEditModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    ` + tr.T("common.cancel") + `
                </button>
            </div>
        </div>
    </div>

    <!-- Upload Settings Modal -->
    <div id="uploadModal" style="display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;">
        <div style="background: white; padding: 40px; border-radius: 12px; max-width: 500px; width: 90%;">
            <h2 style="margin-bottom: 24px; color: #333;">` + tr.T("dashboard.upload_settings_heading") + `</h2>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.file_label_colon") + `</label>
                <p id="selectedFileName" style="color: #666;"></p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="requireAuth" style="margin-right: 8px;">
                    ` + tr.T("dashboard.require_auth_download_label") + `
                </label>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="unlimitedTime" onchange="toggleTimeLimit()">
                    ` + tr.T("dashboard.never_expire_forever_label") + `
                </label>
            </div>

            <div id="timeLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.expiration_days_label") + `</label>
                <input type="number" id="expirationDays" value="7" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">` + tr.T("dashboard.expiration_days_help") + `</p>
            </div>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">
                    <input type="checkbox" id="unlimitedDownloads" onchange="toggleDownloadLimit()">
                    ` + tr.T("dashboard.unlimited_downloads_label") + `
                </label>
            </div>

            <div id="downloadLimitSection" style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.download_limit_colon_label") + `</label>
                <input type="number" id="downloadsLimit" value="5" min="0" style="width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px;">
                <p style="font-size: 12px; color: #999; margin-top: 4px;">` + tr.T("dashboard.download_limit_help") + `</p>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="performUpload()" style="flex: 1; padding: 14px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    ` + tr.T("dashboard.upload_file_heading") + `
                </button>
                <button onclick="closeUploadModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    ` + tr.T("common.cancel") + `
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
            <h2 style="margin-bottom: 24px; color: #333;">📊 ` + tr.T("dashboard.download_history_heading") + `</h2>

            <div style="margin-bottom: 20px;">
                <label style="display: block; margin-bottom: 8px; font-weight: 500;">` + tr.T("dashboard.file_label_colon") + `</label>
                <p id="historyFileName" style="color: #666; font-weight: 600;"></p>
            </div>

            <div id="downloadHistoryContent" style="margin-top: 20px;">
                <p style="text-align: center; color: #999;">` + tr.T("common.loading") + `</p>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 24px;">
                <button onclick="closeDownloadHistoryModal()" style="flex: 1; padding: 14px; background: #e0e0e0; color: #333; border: none; border-radius: 6px; font-weight: 600; cursor: pointer;">
                    ` + tr.T("common.close") + `
                </button>
            </div>
        </div>
    </div>

    <script src="/static/js/dashboard.js?v=6.1.6"></script>
    <script>
        const wvShowPasswordLabel = '` + template.JSEscapeString(showPasswordLabel) + `';
        const I18N = {
            loading: '` + template.JSEscapeString(tr.T("common.loading")) + `',
            noActivityYet: '` + template.JSEscapeString(tr.T("dashboard.no_activity_yet")) + `',
            downloads: '` + template.JSEscapeString(tr.T("dashboard.downloads")) + `',
            dateTime: '` + template.JSEscapeString(tr.T("dashboard.date_time_header")) + `',
            downloadedBy: '` + template.JSEscapeString(tr.T("dashboard.downloaded_by_header")) + `',
            ipAddress: '` + template.JSEscapeString(tr.T("dashboard.ip_address_header")) + `',
            anonymous: '` + template.JSEscapeString(tr.T("dashboard.anonymous_label")) + `',
            na: '` + template.JSEscapeString(tr.T("dashboard.na_label")) + `',
            authShort: '` + template.JSEscapeString(tr.T("dashboard.auth_short_badge")) + `',
            emailsSent: '` + template.JSEscapeString(tr.T("dashboard.emails_sent_heading")) + `',
            recipient: '` + template.JSEscapeString(tr.T("dashboard.recipient_header")) + `',
            message: '` + template.JSEscapeString(tr.T("dashboard.message_header")) + `',
            noMessage: '` + template.JSEscapeString(tr.T("dashboard.no_message_label")) + `',
            errorLoadingHistory: '` + template.JSEscapeString(tr.T("dashboard.error_loading_history")) + `',
            errRecipientEmailRequired: '` + template.JSEscapeString(tr.T("dashboard.err_recipient_email_required")) + `',
            sending: '` + template.JSEscapeString(tr.T("dashboard.sending_label")) + `',
            sendEmailButton: '` + template.JSEscapeString(tr.T("dashboard.send_email_button")) + `',
            emailSentSuccess: '` + template.JSEscapeString(tr.T("notice.email_sent")) + `',
            errorPrefix: '` + template.JSEscapeString(tr.T("dashboard.error_prefix")) + `',
            failedToSendEmail: '` + template.JSEscapeString(tr.T("dashboard.failed_to_send_email")) + `',
            errorSendingEmail: '` + template.JSEscapeString(tr.T("dashboard.error_sending_email")) + `',
            editFormNotLoaded: '` + template.JSEscapeString(tr.T("dashboard.edit_form_not_loaded")) + `',
            currentlySharedWith: '` + template.JSEscapeString(tr.T("dashboard.currently_shared_with")) + `',
            notSharedWithTeams: '` + template.JSEscapeString(tr.T("dashboard.not_shared_with_teams")) + `',
            failedToLoadTeams: '` + template.JSEscapeString(tr.T("dashboard.failed_to_load_teams")) + `',
            removeButton: '` + template.JSEscapeString(tr.T("dashboard.remove_button")) + `',
            pleaseSelectTeam: '` + template.JSEscapeString(tr.T("dashboard.please_select_team")) + `',
            failedToAddTeam: '` + template.JSEscapeString(tr.T("dashboard.failed_to_add_team")) + `',
            unknownError: '` + template.JSEscapeString(tr.T("dashboard.unknown_error")) + `',
            removeTeamConfirm: '` + template.JSEscapeString(tr.T("dashboard.remove_team_confirm")) + `',
            failedToRemoveTeam: '` + template.JSEscapeString(tr.T("dashboard.failed_to_remove_team")) + `',
            fileIdMissing: '` + template.JSEscapeString(tr.T("dashboard.file_id_missing")) + `',
            passwordRequiredOrUncheck: '` + template.JSEscapeString(tr.T("dashboard.password_required_or_uncheck")) + `',
            errorSavingChanges: '` + template.JSEscapeString(tr.T("dashboard.error_saving_changes")) + `',
            pageOf: '` + template.JSEscapeString(tr.T("dashboard.page_info")) + `',
            selectTeamOption: '` + template.JSEscapeString(tr.T("dashboard.select_team_option")) + `'
        };

        function showDownloadHistory(fileId, fileName) {
            document.getElementById('historyFileName').textContent = fileName;
            document.getElementById('downloadHistoryModal').style.display = 'flex';
            document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">' + I18N.loading + '</p>';

            fetch('/file/downloads?file_id=' + encodeURIComponent(fileId))
                .then(response => response.json())
                .then(data => {
                    const downloadLogs = data.downloadLogs || [];
                    const emailLogs = data.emailLogs || [];

                    if (downloadLogs.length === 0 && emailLogs.length === 0) {
                        document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">' + I18N.noActivityYet + '</p>';
                        return;
                    }

                    let html = '';

                    // Show download logs
                    if (downloadLogs.length > 0) {
                        html += '<h3 style="margin-top: 0; margin-bottom: 15px; color: #333; font-size: 16px;">📥 ' + I18N.downloads + ' (' + downloadLogs.length + ')</h3>';
                        html += '<table style="width: 100%; border-collapse: collapse; margin-bottom: 30px;">';
                        html += '<thead><tr style="background: #f5f5f5; border-bottom: 2px solid #ddd;">';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.dateTime + '</th>';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.downloadedBy + '</th>';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.ipAddress + '</th>';
                        html += '</tr></thead><tbody>';

                        downloadLogs.forEach(log => {
                            const date = new Date(log.downloadedAt * 1000);
                            const dateStr = date.toLocaleString('sv-SE');
                            const downloader = log.email || I18N.anonymous;
                            const ip = log.ipAddress || I18N.na;
                            const authBadge = log.isAuthenticated ? ' <span style="background: #2196f3; color: white; padding: 2px 6px; border-radius: 3px; font-size: 11px;">🔒 ' + I18N.authShort + '</span>' : '';

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
                        html += '<h3 style="margin-top: 0; margin-bottom: 15px; color: #333; font-size: 16px;">📧 ' + I18N.emailsSent + ' (' + emailLogs.length + ')</h3>';
                        html += '<table style="width: 100%; border-collapse: collapse;">';
                        html += '<thead><tr style="background: #f5f5f5; border-bottom: 2px solid #ddd;">';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.dateTime + '</th>';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.recipient + '</th>';
                        html += '<th style="padding: 12px; text-align: left;">' + I18N.message + '</th>';
                        html += '</tr></thead><tbody>';

                        emailLogs.forEach(log => {
                            const date = new Date(log.sentAt * 1000);
                            const dateStr = date.toLocaleString('sv-SE');
                            const message = log.message || '<em style="color: #999;">' + I18N.noMessage + '</em>';

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
                    document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #f44336;">' + I18N.errorLoadingHistory + '</p>';
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
            if (element.textContent === wvShowPasswordLabel) {
                element.textContent = password;
                element.style.fontFamily = 'monospace';
            } else {
                element.textContent = wvShowPasswordLabel;
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
                alert(I18N.errRecipientEmailRequired);
                return;
            }

            const btn = event.target;
            btn.disabled = true;
            btn.textContent = I18N.sending;

            try {
                const response = await fetch('/file/email', {
                    method: 'POST',
                    credentials: 'include',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ fileId, recipient, message })
                });

                const result = await response.json();
                if (response.ok) {
                    alert(I18N.emailSentSuccess);
                    closeEmailModal();
                } else {
                    alert(I18N.errorPrefix + ' ' + (result.error || I18N.failedToSendEmail));
                }
            } catch (error) {
                alert(I18N.errorSendingEmail + ' ' + error.message);
            } finally {
                btn.disabled = false;
                btn.textContent = I18N.sendEmailButton;
            }
        }

        // Edit File Modal Functions
        function showEditModal(fileId, fileName, downloadsRemaining, expireAt, unlimitedDownloads, unlimitedTime, fileComment, requireAuth, filePassword) {
            // Store file info
            const fileIdInput = document.getElementById('editFileId');
            if (!fileIdInput) {
                console.error('ERROR: editFileId input element not found!');
                alert(I18N.editFormNotLoaded);
                return;
            }
            fileIdInput.value = fileId;
            document.getElementById('editFileName').textContent = fileName;

            // Set comment/note
            document.getElementById('editFileComment').value = fileComment || '';

            // Set unlimited checkboxes
            document.getElementById('editUnlimitedTime').checked = unlimitedTime;
            document.getElementById('editUnlimitedDownloads').checked = unlimitedDownloads;

            // Set require auth checkbox
            document.getElementById('editRequireAuth').checked = requireAuth;

            // Set password protection
            const hasPassword = filePassword && filePassword.length > 0;
            document.getElementById('editEnablePassword').checked = hasPassword;
            document.getElementById('editFilePassword').value = filePassword || '';
            toggleEditPasswordField();

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

            // Load user's teams and current file teams
            loadUserTeamsForEdit();
            loadCurrentFileTeams(fileId);

            // Show modal
            document.getElementById('editModal').style.display = 'flex';
        }

        function loadUserTeamsForEdit() {
            fetch('/api/teams/my', {
                credentials: 'same-origin'
            })
            .then(response => response.json())
            .then(data => {
                const select = document.getElementById('editTeamSelect');
                select.innerHTML = '<option value="">' + I18N.selectTeamOption + '</option>';

                if (data.success && data.teams && data.teams.length > 0) {
                    data.teams.forEach(team => {
                        const option = document.createElement('option');
                        option.value = team.id;
                        option.textContent = team.name;
                        select.appendChild(option);
                    });
                }
            })
            .catch(error => {
                console.error('Error loading teams:', error);
            });
        }

        function loadCurrentFileTeams(fileId) {
            fetch('/api/teams/file-teams?file_id=' + encodeURIComponent(fileId), {
                credentials: 'same-origin'
            })
            .then(response => response.json())
            .then(data => {
                const container = document.getElementById('editCurrentTeams');
                if (data.success && data.teams && data.teams.length > 0) {
                    container.innerHTML = '<div style="margin-bottom: 8px; font-size: 13px; color: #666; font-weight: 500;">' + I18N.currentlySharedWith + '</div>';
                    data.teams.forEach(team => {
                        const teamDiv = document.createElement('div');
                        teamDiv.style.cssText = 'display: flex; align-items: center; justify-content: space-between; padding: 10px; background: #fff3cd; border: 1px solid #ffc107; border-radius: 4px; margin-bottom: 6px;';
                        teamDiv.innerHTML = '<span style="display: flex; align-items: center; gap: 6px;">' +
                            '<span>👥</span>' +
                            '<span style="font-weight: 500;">' + escapeHtml(team.name) + '</span>' +
                            '</span>' +
                            '<button onclick="removeTeamFromFile(\'' + fileId + '\', ' + team.id + ', \'' + escapeHtml(team.name).replace(/'/g, "\\'") + '\')" ' +
                            'style="padding: 4px 12px; background: #dc3545; color: white; border: none; border-radius: 4px; font-size: 12px; cursor: pointer; font-weight: 500;">' +
                            '✕ ' + I18N.removeButton +
                            '</button>';
                        container.appendChild(teamDiv);
                    });
                } else {
                    container.innerHTML = '<div style="color: #999; font-style: italic; font-size: 14px;">' + I18N.notSharedWithTeams + '</div>';
                }
            })
            .catch(error => {
                console.error('Error loading file teams:', error);
                document.getElementById('editCurrentTeams').innerHTML = '<div style="color: #f44336; font-size: 14px;">' + I18N.failedToLoadTeams + '</div>';
            });
        }

        function addTeamToFile() {
            const fileId = document.getElementById('editFileId').value;
            const select = document.getElementById('editTeamSelect');
            const teamId = select.value;
            const teamName = select.options[select.selectedIndex].text;

            if (!teamId) {
                alert(I18N.pleaseSelectTeam);
                return;
            }

            fetch('/api/teams/share-file', {
                method: 'POST',
                credentials: 'same-origin',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ file_id: fileId, team_id: parseInt(teamId) })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    loadCurrentFileTeams(fileId);
                    select.value = '';
                    location.reload(); // Reload to update badges
                } else {
                    alert(I18N.failedToAddTeam + ': ' + (data.message || I18N.unknownError));
                }
            })
            .catch(error => {
                console.error('Error adding team:', error);
                alert(I18N.failedToAddTeam);
            });
        }

        function removeTeamFromFile(fileId, teamId, teamName) {
            if (!confirm(I18N.removeTeamConfirm.replace('{{team}}', teamName))) {
                return;
            }

            fetch('/api/teams/unshare-file', {
                method: 'POST',
                credentials: 'same-origin',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ file_id: fileId, team_id: teamId })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    loadCurrentFileTeams(fileId);
                    location.reload(); // Reload to update badges
                } else {
                    alert(I18N.failedToRemoveTeam + ': ' + (data.message || I18N.unknownError));
                }
            })
            .catch(error => {
                console.error('Error removing team:', error);
                alert(I18N.failedToRemoveTeam);
            });
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

        function toggleEditPasswordField() {
            const checkbox = document.getElementById('editEnablePassword');
            const container = document.getElementById('editPasswordFieldContainer');
            const passwordInput = document.getElementById('editFilePassword');

            if (checkbox.checked) {
                container.style.display = 'block';
                passwordInput.required = true;
            } else {
                container.style.display = 'none';
                passwordInput.required = false;
                passwordInput.value = '';
            }
        }

        function saveFileEdit() {
            const fileId = document.getElementById('editFileId').value;
            const unlimitedTime = document.getElementById('editUnlimitedTime').checked;
            const unlimitedDownloads = document.getElementById('editUnlimitedDownloads').checked;
            const teamId = document.getElementById('editTeamSelect').value;
            const fileComment = document.getElementById('editFileComment').value;
            const requireAuth = document.getElementById('editRequireAuth').checked;
            const enablePassword = document.getElementById('editEnablePassword').checked;
            const filePassword = document.getElementById('editFilePassword').value;

            if (!fileId || fileId === '') {
                alert(I18N.fileIdMissing);
                return;
            }

            // Validate password if enabled
            if (enablePassword && (!filePassword || filePassword.trim() === '')) {
                alert(I18N.passwordRequiredOrUncheck);
                return;
            }

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
            formData.append('file_comment', fileComment);
            formData.append('require_auth', requireAuth ? 'true' : 'false');

            // Only send password if checkbox is enabled
            if (enablePassword) {
                formData.append('file_password', filePassword);
            } else {
                formData.append('file_password', ''); // Clear password
            }

            if (teamId) {
                formData.append('team_id', teamId);
            }

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
                    alert(I18N.errorPrefix + ' ' + result.error);
                }
            })
            .catch(error => {
                alert(I18N.errorSavingChanges + ' ' + error);
            });
        }

        // File filtering function
        function filterFiles(type) {
            const fileItems = document.querySelectorAll('.file-item');
            const tabs = document.querySelectorAll('.file-tab');
            const teamFilter = document.getElementById('teamFilter');

            // Update active tab
            tabs.forEach(tab => {
                const filter = tab.getAttribute('data-filter');
                if (filter === type) {
                    tab.classList.add('active');
                    tab.style.fontWeight = '600';
                    tab.style.borderBottomColor = '` + s.getPrimaryColor() + `';
                    tab.style.color = '` + s.getPrimaryColor() + `';
                } else {
                    tab.classList.remove('active');
                    tab.style.fontWeight = '500';
                    tab.style.borderBottomColor = 'transparent';
                    tab.style.color = '#666';
                }
            });

            // Show/hide team filter dropdown
            if (type === 'team') {
                teamFilter.style.display = 'block';
            } else {
                teamFilter.style.display = 'none';
                teamFilter.value = ''; // Reset selection when switching away
            }

            // Filter files - use data-filter-hidden attribute
            fileItems.forEach(item => {
                const fileType = item.getAttribute('data-file-type');
                if (type === 'all') {
                    item.setAttribute('data-filter-hidden', 'false');
                } else if (type === 'my') {
                    // Show my files and files I shared with teams (both)
                    const shouldShow = (fileType === 'my' || fileType === 'both');
                    item.setAttribute('data-filter-hidden', shouldShow ? 'false' : 'true');
                } else if (type === 'team') {
                    // Show team files and my files shared with teams (both)
                    const shouldShow = (fileType === 'team' || fileType === 'both');
                    item.setAttribute('data-filter-hidden', shouldShow ? 'false' : 'true');
                }
            });

            // Update pagination after filtering
            currentPage = 1;
            updatePagination();
        }

        // Filter by specific team
        function filterByTeam(teamName) {
            const fileItems = document.querySelectorAll('.file-item');

            fileItems.forEach(item => {
                const fileType = item.getAttribute('data-file-type');
                const teams = item.getAttribute('data-teams') || '';

                // Only filter team files (team or both)
                if (fileType !== 'team' && fileType !== 'both') {
                    item.setAttribute('data-filter-hidden', 'true');
                    return;
                }

                if (!teamName) {
                    // Show all team files
                    item.setAttribute('data-filter-hidden', 'false');
                } else {
                    // Check if file belongs to selected team
                    const teamList = teams.split(',').map(t => t.trim());
                    if (teamList.includes(teamName)) {
                        item.setAttribute('data-filter-hidden', 'false');
                    } else {
                        item.setAttribute('data-filter-hidden', 'true');
                    }
                }
            });

            // Update pagination after team filtering
            currentPage = 1;
            updatePagination();
        }

        // Search and sort files function
        function searchAndSortFiles() {
            const searchTerm = document.getElementById('fileSearch').value.toLowerCase();
            const sortValue = document.getElementById('fileSort').value;
            const fileList = document.querySelector('.file-list');
            const fileItems = Array.from(document.querySelectorAll('.file-item'));

            // First, apply search filter separately
            fileItems.forEach(item => {
                const filename = item.getAttribute('data-filename').toLowerCase();
                const extension = item.getAttribute('data-extension').toLowerCase();
                const comment = (item.getAttribute('data-comment') || '').toLowerCase();

                // Search in filename, extension, and comment/description
                if (searchTerm === '' || filename.includes(searchTerm) || extension.includes(searchTerm) || comment.includes(searchTerm)) {
                    item.setAttribute('data-search-hidden', 'false');
                } else {
                    item.setAttribute('data-search-hidden', 'true');
                }
            });

            // Get items that pass both tab/team filter AND search filter
            const visibleItems = fileItems.filter(item => {
                const tabFilterHidden = item.getAttribute('data-filter-hidden') === 'true';
                const searchFilterHidden = item.getAttribute('data-search-hidden') === 'true';
                return !tabFilterHidden && !searchFilterHidden;
            });

            // Sort visible items
            visibleItems.sort((a, b) => {
                let aVal, bVal;

                switch(sortValue) {
                    case 'name-asc':
                        aVal = a.getAttribute('data-filename').toLowerCase();
                        bVal = b.getAttribute('data-filename').toLowerCase();
                        return aVal.localeCompare(bVal);

                    case 'name-desc':
                        aVal = a.getAttribute('data-filename').toLowerCase();
                        bVal = b.getAttribute('data-filename').toLowerCase();
                        return bVal.localeCompare(aVal);

                    case 'date-asc':
                        aVal = parseInt(a.getAttribute('data-timestamp'));
                        bVal = parseInt(b.getAttribute('data-timestamp'));
                        return aVal - bVal;

                    case 'date-desc':
                        aVal = parseInt(a.getAttribute('data-timestamp'));
                        bVal = parseInt(b.getAttribute('data-timestamp'));
                        return bVal - aVal;

                    case 'downloads-asc':
                        aVal = parseInt(a.getAttribute('data-downloads'));
                        bVal = parseInt(b.getAttribute('data-downloads'));
                        return aVal - bVal;

                    case 'downloads-desc':
                        aVal = parseInt(a.getAttribute('data-downloads'));
                        bVal = parseInt(b.getAttribute('data-downloads'));
                        return bVal - aVal;

                    case 'size-asc':
                        aVal = parseInt(a.getAttribute('data-size'));
                        bVal = parseInt(b.getAttribute('data-size'));
                        return aVal - bVal;

                    case 'size-desc':
                        aVal = parseInt(a.getAttribute('data-size'));
                        bVal = parseInt(b.getAttribute('data-size'));
                        return bVal - aVal;

                    default:
                        return 0;
                }
            });

            // Reorder DOM elements
            visibleItems.forEach(item => {
                fileList.appendChild(item);
            });

            // Append hidden items at the end
            fileItems.filter(item => item.style.display === 'none').forEach(item => {
                fileList.appendChild(item);
            });

            // Update pagination after filtering/sorting
            updatePagination();
        }

        // Pagination variables
        let currentPage = 1;
        let perPage = 25;

        // Initialize pagination on page load
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize all items as visible
            const fileItems = document.querySelectorAll('.file-item');
            fileItems.forEach(item => {
                item.setAttribute('data-filter-hidden', 'false');
                item.setAttribute('data-search-hidden', 'false');
            });
            updatePagination();
        });

        function changePerPage() {
            perPage = parseInt(document.getElementById('perPageSelect').value);
            currentPage = 1; // Reset to first page
            updatePagination();
        }

        function prevPage() {
            const visibleItems = getVisibleItems();
            const totalPages = Math.ceil(visibleItems.length / perPage);
            if (currentPage > 1) {
                currentPage--;
                updatePagination();
            }
        }

        function nextPage() {
            const visibleItems = getVisibleItems();
            const totalPages = Math.ceil(visibleItems.length / perPage);
            if (currentPage < totalPages) {
                currentPage++;
                updatePagination();
            }
        }

        function getVisibleItems() {
            const fileItems = Array.from(document.querySelectorAll('.file-item'));
            // Get items that are NOT hidden by filters (tab filters, team filters, search)
            // Use a custom attribute to track filter state separately from pagination
            return fileItems.filter(item => {
                // If item has been marked as filter-hidden, exclude it
                return item.getAttribute('data-filter-hidden') !== 'true';
            });
        }

        function updatePagination() {
            const allItems = Array.from(document.querySelectorAll('.file-item'));

            // First, determine which items are visible based on current filters (not pagination)
            // Check both tab/team filter AND search filter
            const visibleItems = allItems.filter(item => {
                const tabFilterHidden = item.getAttribute('data-filter-hidden') === 'true';
                const searchFilterHidden = item.getAttribute('data-search-hidden') === 'true';
                return !tabFilterHidden && !searchFilterHidden;
            });

            const totalFiltered = visibleItems.length;
            const totalPages = Math.ceil(totalFiltered / perPage);

            // Ensure current page is valid
            if (currentPage > totalPages && totalPages > 0) {
                currentPage = totalPages;
            }
            if (currentPage < 1) {
                currentPage = 1;
            }

            // Calculate start and end indices for current page
            const startIdx = (currentPage - 1) * perPage;
            const endIdx = Math.min(startIdx + perPage, totalFiltered);

            // First hide all items
            allItems.forEach(item => {
                item.style.display = 'none';
            });

            // Show only items for current page (among the filtered visible items)
            for (let i = startIdx; i < endIdx; i++) {
                if (visibleItems[i]) {
                    visibleItems[i].style.display = '';
                }
            }

            // Update counter
            const visibleCount = endIdx - startIdx;
            const totalCountEl = document.getElementById('totalCount');
            const visibleCountEl = document.getElementById('visibleCount');
            if (totalCountEl) totalCountEl.textContent = totalFiltered;
            if (visibleCountEl) visibleCountEl.textContent = visibleCount;

            // Update pagination controls
            const pageInfoEl = document.getElementById('pageInfo');
            const prevBtnEl = document.getElementById('prevBtn');
            const nextBtnEl = document.getElementById('nextBtn');

            if (pageInfoEl) pageInfoEl.textContent = I18N.pageOf.replace('{{page}}', currentPage).replace('{{total}}', Math.max(1, totalPages));

            if (prevBtnEl) {
                prevBtnEl.disabled = currentPage === 1;
                prevBtnEl.style.opacity = currentPage === 1 ? '0.5' : '1';
                prevBtnEl.style.cursor = currentPage === 1 ? 'not-allowed' : 'pointer';
            }

            if (nextBtnEl) {
                nextBtnEl.disabled = currentPage >= totalPages || totalPages === 0;
                nextBtnEl.style.opacity = (currentPage >= totalPages || totalPages === 0) ? '0.5' : '1';
                nextBtnEl.style.cursor = (currentPage >= totalPages || totalPages === 0) ? 'not-allowed' : 'pointer';
            }
        }

        // Note: loadFileRequests, deleteFileRequest, escapeHtml, and copyToClipboard
        // are defined in dashboard.js and loaded automatically on page load
    </script>
    <div style="text-align: center; padding: 40px 20px 20px; color: #999; font-size: 12px;">
        ` + tr.T("dashboard.powered_by", "version", s.config.Version) + `
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
