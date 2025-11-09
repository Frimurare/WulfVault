package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
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

	// Get files from database
	dbFiles, err := database.DB.GetFilesByUser(user.Id)
	if err != nil {
		log.Printf("Error getting files for user %d: %v", user.Id, err)
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch files")
		return
	}

	// Convert to response format
	files := make([]map[string]interface{}, 0, len(dbFiles))
	for _, f := range dbFiles {
		downloadURL := s.config.ServerURL + "/d/" + f.Id

		files = append(files, map[string]interface{}{
			"id":              f.Id,
			"name":            f.Name,
			"size":            f.Size,
			"size_formatted":  database.FormatFileSize(f.SizeBytes),
			"download_url":    downloadURL,
			"downloads":       f.DownloadCount,
			"max_downloads":   f.DownloadsRemaining,
			"expires_at":      f.ExpireAt,
			"created_at":      f.UploadDate,
			"require_auth":    f.RequireAuth,
		})
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"files": files,
		"user":  user,
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

	// Delete file from disk
	filePath := filepath.Join(s.config.UploadsDir, fileID)
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: Could not delete file from disk: %v", err)
	}

	// Delete from database
	if err := database.DB.DeleteFile(fileID); err != nil {
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

// renderUserDashboard renders the user dashboard HTML
func (s *Server) renderUserDashboard(w http.ResponseWriter, userModel interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	user := userModel.(*models.User)

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
            background: white;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            color: ` + s.config.PrimaryColor + `;
            font-size: 24px;
        }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .header nav a:hover {
            color: ` + s.config.PrimaryColor + `;
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
            color: ` + s.config.PrimaryColor + `;
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
            background: ` + s.config.PrimaryColor + `;
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
            border-color: ` + s.config.PrimaryColor + `;
            background: #f9f9f9;
        }
        .upload-zone.drag-over {
            border-color: ` + s.config.PrimaryColor + `;
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
            background: ` + s.config.PrimaryColor + `;
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
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
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

        <div class="upload-zone" id="uploadZone" onclick="document.getElementById('fileInput').click()">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <h2>Drop files here or click to upload</h2>
            <p>Maximum file size: 5 GB</p>
            <input type="file" id="fileInput" multiple>
        </div>

        <div class="files-section">
            <div class="files-header">
                <h2>My Files</h2>
            </div>`

	// Generate file list HTML
	if len(files) == 0 {
		html += `
            <div class="empty-state">
                No files uploaded yet. Start by uploading your first file!
            </div>`
	} else {
		html += `
            <ul class="file-list">`

		for _, f := range files {
			downloadURL := s.config.ServerURL + "/d/" + f.Id

			// Format expiration info
			expiryInfo := ""
			if !f.UnlimitedTime && f.ExpireAt > 0 {
				expiryTime := time.Unix(f.ExpireAt, 0)
				expiryInfo = fmt.Sprintf(" • Expires: %s", expiryTime.Format("2006-01-02 15:04"))
			}

			downloadsInfo := ""
			if f.UnlimitedDownloads {
				downloadsInfo = fmt.Sprintf("%d downloads", f.DownloadCount)
			} else {
				totalDownloads := f.DownloadCount + f.DownloadsRemaining
				downloadsInfo = fmt.Sprintf("%d/%d downloads", f.DownloadCount, totalDownloads)
			}

			authBadge := ""
			if f.RequireAuth {
				authBadge = ` <span style="background: #ffc107; color: #000; padding: 2px 8px; border-radius: 4px; font-size: 12px;">🔒 Auth Required</span>`
			}

			html += `
                <li class="file-item">
                    <div class="file-info">
                        <h3>` + f.Name + authBadge + `</h3>
                        <p>` + database.FormatFileSize(f.SizeBytes) + ` • ` + downloadsInfo + expiryInfo + `</p>
                        <p style="font-size: 12px; color: #999; margin-top: 4px;">` + downloadURL + `</p>
                    </div>
                    <div class="file-actions">
                        <button class="btn btn-primary" onclick="copyToClipboard('` + downloadURL + `', this)">📋 Copy Link</button>
                        <button class="btn btn-danger" onclick="deleteFile('` + f.Id + `', '` + f.Name + `')">🗑️ Delete</button>
                    </div>
                </li>`
		}

		html += `
            </ul>`
	}

	html += `
        </div>
    </div>

    <script src="/static/js/dashboard.js"></script>
</body>
</html>`

	w.Write([]byte(html))
}
