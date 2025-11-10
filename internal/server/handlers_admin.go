package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Frimurare/Sharecare/internal/auth"
	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

// handleAdminDashboard renders the admin dashboard
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get statistics
	totalUsers, _ := database.DB.GetTotalUsers()
	activeUsers, _ := database.DB.GetActiveUsers()
	totalDownloads, _ := database.DB.GetTotalDownloads()
	downloadsToday, _ := database.DB.GetDownloadsToday()

	s.renderAdminDashboard(w, user, totalUsers, activeUsers, totalDownloads, downloadsToday)
}

// handleAdminUsers lists all users
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.DB.GetAllUsers()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	s.renderAdminUsers(w, users)
}

// handleAdminUserCreate creates a new user
func (s *Server) handleAdminUserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderAdminUserForm(w, nil, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderAdminUserForm(w, nil, "Invalid form data")
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	quotaMB, _ := strconv.ParseInt(r.FormValue("quota_mb"), 10, 64)
	userLevel, _ := strconv.Atoi(r.FormValue("user_level"))

	// Validate
	if name == "" || email == "" || password == "" {
		s.renderAdminUserForm(w, nil, "All fields are required")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		s.renderAdminUserForm(w, nil, "Failed to hash password")
		return
	}

	// Create user
	newUser := &models.User{
		Name:           name,
		Email:          email,
		Password:       hashedPassword,
		UserLevel:      models.UserRank(userLevel),
		Permissions:    models.UserPermissionNone,
		StorageQuotaMB: quotaMB,
		StorageUsedMB:  0,
		IsActive:       true,
	}

	// Set permissions based on user level
	if newUser.UserLevel == models.UserLevelAdmin {
		newUser.Permissions = models.UserPermissionAll
	}

	if err := database.DB.CreateUser(newUser); err != nil {
		s.renderAdminUserForm(w, nil, "Failed to create user: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handleAdminUserEdit edits a user
func (s *Server) handleAdminUserEdit(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	existingUser, err := database.DB.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		s.renderAdminUserForm(w, existingUser, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderAdminUserForm(w, existingUser, "Invalid form data")
		return
	}

	existingUser.Name = r.FormValue("name")
	existingUser.Email = r.FormValue("email")
	existingUser.StorageQuotaMB, _ = strconv.ParseInt(r.FormValue("quota_mb"), 10, 64)
	existingUser.UserLevel = models.UserRank(mustParseInt(r.FormValue("user_level")))
	existingUser.IsActive = r.FormValue("is_active") == "1"

	// Update password if provided
	newPassword := r.FormValue("password")
	if newPassword != "" {
		hashedPassword, err := auth.HashPassword(newPassword)
		if err != nil {
			s.renderAdminUserForm(w, existingUser, "Failed to hash password")
			return
		}
		existingUser.Password = hashedPassword
	}

	if err := database.DB.UpdateUser(existingUser); err != nil {
		s.renderAdminUserForm(w, existingUser, "Failed to update user: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handleAdminUserDelete deletes a user
func (s *Server) handleAdminUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := strconv.Atoi(r.FormValue("id"))
	if userID == 0 {
		s.sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := database.DB.DeleteUser(userID); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "User deleted"})
}

// handleAdminFiles lists all files in the system
func (s *Server) handleAdminFiles(w http.ResponseWriter, r *http.Request) {
	files, err := database.DB.GetAllFiles()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch files")
		return
	}

	// Calculate total storage
	var totalStorage int64
	for _, f := range files {
		totalStorage += f.SizeBytes
	}

	s.renderAdminFiles(w, files, totalStorage)
}

// handleAdminBranding handles branding settings
func (s *Server) handleAdminBranding(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderAdminBranding(w, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		s.renderAdminBranding(w, "Failed to parse form: "+err.Error())
		return
	}

	// Get form values
	companyName := r.FormValue("company_name")
	primaryColor := r.FormValue("primary_color")
	secondaryColor := r.FormValue("secondary_color")

	// Handle logo upload
	logoData := ""
	file, _, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		// Read file data
		buf := make([]byte, 10<<20) // 10MB max
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			s.renderAdminBranding(w, "Failed to read logo file: "+err.Error())
			return
		}
		// Convert to base64 data URL
		contentType := http.DetectContentType(buf[:n])
		logoData = "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(buf[:n])
	}

	// Save to database
	if companyName != "" {
		database.DB.SetConfigValue("branding_company_name", companyName)
	}
	if primaryColor != "" {
		database.DB.SetConfigValue("branding_primary_color", primaryColor)
	}
	if secondaryColor != "" {
		database.DB.SetConfigValue("branding_secondary_color", secondaryColor)
	}
	if logoData != "" {
		database.DB.SetConfigValue("branding_logo", logoData)
	}

	// Reload config
	s.loadBrandingConfig()

	s.renderAdminBranding(w, "Branding settings updated successfully!")
}

// handleAdminSettings handles general settings
func (s *Server) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderAdminSettings(w, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderAdminSettings(w, "Error: Invalid form data")
		return
	}

	// Update settings in database
	serverURL := r.FormValue("server_url")
	if serverURL != "" {
		database.DB.SetConfigValue("server_url", serverURL)
		s.config.ServerURL = serverURL
	}

	maxFileSizeMB := r.FormValue("max_file_size_mb")
	if maxFileSizeMB != "" {
		database.DB.SetConfigValue("max_file_size_mb", maxFileSizeMB)
	}

	defaultQuotaMB := r.FormValue("default_quota_mb")
	if defaultQuotaMB != "" {
		database.DB.SetConfigValue("default_quota_mb", defaultQuotaMB)
	}

	s.renderAdminSettings(w, "Settings updated successfully!")
}

// handleAdminTrash lists all deleted files (trash)
func (s *Server) handleAdminTrash(w http.ResponseWriter, r *http.Request) {
	files, err := database.DB.GetDeletedFiles()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch trash")
		return
	}

	s.renderAdminTrash(w, files)
}

// handleAdminPermanentDelete permanently deletes a file
func (s *Server) handleAdminPermanentDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fileID := r.FormValue("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	// Get file info before deletion
	// We need to use a special query to get deleted files
	rows, err := database.DB.GetDeletedFiles()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch file info")
		return
	}

	var fileInfo *database.FileInfo
	for _, f := range rows {
		if f.Id == fileID {
			fileInfo = f
			break
		}
	}

	if fileInfo == nil {
		s.sendError(w, http.StatusNotFound, "File not found in trash")
		return
	}

	// Delete from disk
	filePath := filepath.Join(s.config.UploadsDir, fileID)
	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Could not delete file from disk: %v", err)
		}
	}

	// Permanently delete from database
	if err := database.DB.PermanentDeleteFile(fileID); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	log.Printf("File permanently deleted by admin: %s (ID: %s)", fileInfo.Name, fileID)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "File permanently deleted",
	})
}

// handleAdminRestoreFile restores a file from trash
func (s *Server) handleAdminRestoreFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fileID := r.FormValue("file_id")
	if fileID == "" {
		s.sendError(w, http.StatusBadRequest, "Missing file_id")
		return
	}

	if err := database.DB.RestoreFile(fileID); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to restore file")
		return
	}

	log.Printf("File restored from trash by admin: %s", fileID)

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "File restored successfully",
	})
}

// Render functions

func (s *Server) renderAdminDashboard(w http.ResponseWriter, user *models.User, totalUsers, activeUsers, totalDownloads, downloadsToday int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Dashboard - ` + s.config.CompanyName + `</title>
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
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        h2 {
            margin-bottom: 24px;
            color: #333;
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
            font-size: 36px;
            font-weight: 700;
            color: ` + s.config.PrimaryColor + `;
        }
        .quick-actions {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 40px;
        }
        .action-btn {
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            text-align: center;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: transform 0.2s;
        }
        .action-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + ` - Admin</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/admin/branding">Branding</a>
            <a href="/admin/settings">Settings</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <h2>Dashboard Overview</h2>

        <div class="stats">
            <div class="stat-card">
                <h3>Total Users</h3>
                <div class="value">` + fmt.Sprintf("%d", totalUsers) + `</div>
            </div>
            <div class="stat-card">
                <h3>Active Users</h3>
                <div class="value">` + fmt.Sprintf("%d", activeUsers) + `</div>
            </div>
            <div class="stat-card">
                <h3>Total Downloads</h3>
                <div class="value">` + fmt.Sprintf("%d", totalDownloads) + `</div>
            </div>
            <div class="stat-card">
                <h3>Downloads Today</h3>
                <div class="value">` + fmt.Sprintf("%d", downloadsToday) + `</div>
            </div>
        </div>

        <h2>Quick Actions</h2>
        <div class="quick-actions">
            <a href="/admin/users/create" class="action-btn">➕ Create User</a>
            <a href="/admin/users" class="action-btn">👥 Manage Users</a>
            <a href="/admin/files" class="action-btn">📁 View All Files</a>
            <a href="/admin/branding" class="action-btn">🎨 Customize Branding</a>
            <a href="/admin/settings" class="action-btn">⚙️ System Settings</a>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminUsers(w http.ResponseWriter, users []*models.User) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Manage Users - ` + s.config.CompanyName + `</title>
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
        .header h1 { color: ` + s.config.PrimaryColor + `; font-size: 24px; }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .actions {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
        }
        .btn {
            padding: 10px 20px;
            background: ` + s.config.PrimaryColor + `;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
        }
        table {
            width: 100%;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 16px;
            text-align: left;
        }
        th {
            background: #f9f9f9;
            font-weight: 600;
            color: #666;
        }
        tr:not(:last-child) td {
            border-bottom: 1px solid #e0e0e0;
        }
        .badge {
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge-admin { background: #e3f2fd; color: #1976d2; }
        .badge-user { background: #f3e5f5; color: #7b1fa2; }
        .action-links a {
            margin-right: 12px;
            color: ` + s.config.PrimaryColor + `;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <div class="actions">
            <h2>Manage Users</h2>
            <a href="/admin/users/create" class="btn">+ Create User</a>
        </div>

        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Level</th>
                    <th>Quota</th>
                    <th>Used</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>`

	for _, u := range users {
		levelBadge := `<span class="badge badge-user">User</span>`
		if u.UserLevel == models.UserLevelAdmin || u.UserLevel == models.UserLevelSuperAdmin {
			levelBadge = `<span class="badge badge-admin">Admin</span>`
		}

		status := "Active"
		if !u.IsActive {
			status = "Inactive"
		}

		html += fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%d GB</td>
                    <td>%d MB</td>
                    <td>%s</td>
                    <td class="action-links">
                        <a href="/admin/users/edit?id=%d">Edit</a>
                        <a href="#" onclick="deleteUser(%d); return false;">Delete</a>
                    </td>
                </tr>`,
			u.Name, u.Email, levelBadge, u.StorageQuotaMB/1000, u.StorageUsedMB, status, u.Id, u.Id)
	}

	html += `
            </tbody>
        </table>
    </div>

    <script>
        function deleteUser(id) {
            if (!confirm('Are you sure you want to delete this user?')) return;

            fetch('/admin/users/delete', {
                method: 'POST',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                body: 'id=' + id
            })
            .then(() => window.location.reload())
            .catch(err => alert('Error deleting user'));
        }
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminUserForm(w http.ResponseWriter, user *models.User, errorMsg string) {
	// Simple form implementation
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	isEdit := user != nil
	title := "Create User"
	action := "/admin/users/create"

	if isEdit {
		title = "Edit User"
		action = fmt.Sprintf("/admin/users/edit?id=%d", user.Id)
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>` + title + `</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        input, select { width: 100%; padding: 8px; margin: 8px 0; }
        button { padding: 10px 20px; background: ` + s.config.PrimaryColor + `; color: white; border: none; cursor: pointer; }
        .error { background: #fee; padding: 10px; margin: 10px 0; border-radius: 4px; color: #c33; }
    </style>
</head>
<body>
    <h1>` + title + `</h1>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
	}

	nameVal, emailVal, quotaVal := "", "", "5000"
	userLevelVal := "2"

	if isEdit {
		nameVal = user.Name
		emailVal = user.Email
		quotaVal = fmt.Sprintf("%d", user.StorageQuotaMB)
		userLevelVal = fmt.Sprintf("%d", user.UserLevel)
	}

	html += `
    <form method="POST" action="` + action + `">
        <label>Name:</label>
        <input type="text" name="name" value="` + nameVal + `" required>

        <label>Email:</label>
        <input type="email" name="email" value="` + emailVal + `" required>

        <label>Password` + func() string {
		if isEdit {
			return " (leave empty to keep current)"
		}
		return ""
	}() + `:</label>
        <input type="password" name="password">

        <label>Storage Quota (MB):</label>
        <input type="number" name="quota_mb" value="` + quotaVal + `" required>

        <label>User Level:</label>
        <select name="user_level">
            <option value="2"` + func() string {
		if userLevelVal == "2" {
			return " selected"
		}
		return ""
	}() + `>Regular User</option>
            <option value="1"` + func() string {
		if userLevelVal == "1" {
			return " selected"
		}
		return ""
	}() + `>Admin</option>
        </select>

        <br><br>
        <button type="submit">Save</button>
        <a href="/admin/users">Cancel</a>
    </form>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminFiles(w http.ResponseWriter, files []*database.FileInfo, totalStorage int64) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	totalStorageGB := fmt.Sprintf("%.2f GB", float64(totalStorage)/(1024*1024*1024))

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>All Files - ` + s.config.CompanyName + `</title>
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
        .header h1 { color: ` + s.config.PrimaryColor + `; font-size: 24px; }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .header nav a:hover { color: ` + s.config.PrimaryColor + `; }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .stats-bar {
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 24px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .stat-item {
            text-align: center;
        }
        .stat-item h3 {
            color: #666;
            font-size: 14px;
            margin-bottom: 8px;
        }
        .stat-item .value {
            font-size: 28px;
            font-weight: 700;
            color: ` + s.config.PrimaryColor + `;
        }
        table {
            width: 100%;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 16px;
            text-align: left;
        }
        th {
            background: #f9f9f9;
            font-weight: 600;
            color: #666;
            font-size: 14px;
        }
        tr:not(:last-child) td {
            border-bottom: 1px solid #e0e0e0;
        }
        tr:hover {
            background: #f9f9f9;
        }
        .badge {
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 500;
            display: inline-block;
        }
        .badge-active { background: #e8f5e9; color: #2e7d32; }
        .badge-expired { background: #ffebee; color: #c62828; }
        .badge-auth { background: #e3f2fd; color: #1976d2; }
        .btn {
            padding: 6px 12px;
            border: none;
            border-radius: 6px;
            font-size: 13px;
            font-weight: 500;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-right: 4px;
        }
        .btn-primary { background: ` + s.config.PrimaryColor + `; color: white; }
        .btn-secondary { background: #e0e0e0; color: #333; }
        .btn:hover { opacity: 0.8; }
        .file-name {
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/admin/branding">Branding</a>
            <a href="/admin/settings">Settings</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <h2 style="margin-bottom: 20px;">All Files</h2>

        <div class="stats-bar">
            <div class="stat-item">
                <h3>Total Files</h3>
                <div class="value">` + fmt.Sprintf("%d", len(files)) + `</div>
            </div>
            <div class="stat-item">
                <h3>Total Storage</h3>
                <div class="value">` + totalStorageGB + `</div>
            </div>
            <div class="stat-item">
                <h3>Total Downloads</h3>
                <div class="value">` + fmt.Sprintf("%d", calculateTotalDownloads(files)) + `</div>
            </div>
        </div>

        <table>
            <thead>
                <tr>
                    <th>File Name</th>
                    <th>User</th>
                    <th>Size</th>
                    <th>Downloads</th>
                    <th>Expiration</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>`

	if len(files) == 0 {
		html += `
                <tr>
                    <td colspan="7" style="text-align: center; padding: 40px; color: #999;">
                        No files in the system yet.
                    </td>
                </tr>`
	}

	for _, f := range files {
		// Get user info
		userName := "Unknown"
		user, err := database.DB.GetUserByID(f.UserId)
		if err == nil {
			userName = user.Name
		}

		// Status
		status := `<span class="badge badge-active">Active</span>`
		if !f.UnlimitedDownloads && f.DownloadsRemaining <= 0 {
			status = `<span class="badge badge-expired">Expired</span>`
		} else if !f.UnlimitedTime && f.ExpireAt > 0 && f.ExpireAt < time.Now().Unix() {
			status = `<span class="badge badge-expired">Expired</span>`
		}

		// Auth badge
		authBadge := ""
		if f.RequireAuth {
			authBadge = ` <span class="badge badge-auth">🔒 Auth</span>`
		}

		// Expiration info
		expiryInfo := "Never"
		if !f.UnlimitedTime && f.ExpireAtString != "" {
			expiryInfo = f.ExpireAtString
		}
		if !f.UnlimitedDownloads {
			expiryInfo += fmt.Sprintf(" (%d left)", f.DownloadsRemaining)
		}

		downloadURL := s.config.ServerURL + "/d/" + f.Id

		html += fmt.Sprintf(`
                <tr>
                    <td>
                        <div class="file-name" title="%s">📄 %s%s</div>
                    </td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%d</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>
                        <button class="btn btn-primary" onclick="copyToClipboard('%s', this)" title="Copy link">
                            📋 Copy
                        </button>
                        <button class="btn btn-secondary" onclick="deleteFile('%s')" title="Delete file">
                            🗑️ Delete
                        </button>
                    </td>
                </tr>`,
			f.Name, f.Name, authBadge,
			userName,
			f.Size,
			f.DownloadCount,
			expiryInfo,
			status,
			downloadURL,
			f.Id)
	}

	html += `
            </tbody>
        </table>
    </div>

    <script>
        // Copy to clipboard function
        function copyToClipboard(url, button) {
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(url).then(() => {
                    const originalText = button.innerHTML;
                    button.innerHTML = '✓ Copied!';
                    button.style.background = '#4caf50';
                    setTimeout(() => {
                        button.innerHTML = originalText;
                        button.style.background = '';
                    }, 2000);
                }).catch(() => {
                    fallbackCopyToClipboard(url);
                });
            } else {
                fallbackCopyToClipboard(url);
            }
        }

        function fallbackCopyToClipboard(text) {
            const textArea = document.createElement("textarea");
            textArea.value = text;
            textArea.style.position = "fixed";
            textArea.style.left = "-999999px";
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();
            try {
                document.execCommand('copy');
                alert('✓ Link copied to clipboard!');
            } catch (err) {
                prompt('Copy this link manually:', text);
            }
            document.body.removeChild(textArea);
        }

        async function deleteFile(fileId) {
            if (!confirm('Are you sure you want to delete this file? This action cannot be undone.')) return;

            try {
                const response = await fetch('/file/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: 'file_id=' + fileId
                });

                if (response.ok) {
                    location.reload();
                } else {
                    const result = await response.json();
                    alert('Delete failed: ' + (result.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Delete failed: ' + error.message);
            }
        }
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminBranding(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get current branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Branding Settings - ` + s.config.CompanyName + `</title>
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
        .header h1 { color: ` + s.config.PrimaryColor + `; font-size: 24px; }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .header nav a:hover { color: ` + s.config.PrimaryColor + `; }
        .container {
            max-width: 800px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        h2 {
            margin-bottom: 20px;
            color: #333;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
        }
        input[type="text"], input[type="color"], input[type="file"] {
            width: 100%;
            padding: 10px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
        }
        input:focus {
            outline: none;
            border-color: ` + s.config.PrimaryColor + `;
        }
        .color-input {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .color-input input[type="color"] {
            width: 60px;
            height: 40px;
            padding: 2px;
            cursor: pointer;
        }
        .btn {
            padding: 12px 24px;
            background: ` + s.config.PrimaryColor + `;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
        }
        .btn:hover {
            opacity: 0.9;
        }
        .message {
            background: #4caf50;
            color: white;
            padding: 12px 20px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
        .logo-preview {
            margin-top: 10px;
            max-width: 300px;
        }
        .logo-preview img {
            max-width: 100%;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            padding: 10px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/admin/trash">Trash</a>
            <a href="/admin/branding">Branding</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <h2>Branding Settings</h2>`

	if message != "" {
		html += `<div class="message">` + message + `</div>`
	}

	html += `
        <div class="card">
            <form method="POST" enctype="multipart/form-data">
                <div class="form-group">
                    <label>Company Name</label>
                    <input type="text" name="company_name" value="` + brandingConfig["branding_company_name"] + `" placeholder="Sharecare">
                </div>

                <div class="form-group">
                    <label>Logo (PNG, JPG, SVG - Max 10MB)</label>
                    <input type="file" name="logo" accept="image/*">
                    `
	if brandingConfig["branding_logo"] != "" {
		html += `
                    <div class="logo-preview">
                        <p style="margin-top: 10px; color: #666; font-size: 14px;">Current Logo:</p>
                        <img src="` + brandingConfig["branding_logo"] + `" alt="Current Logo">
                    </div>`
	}
	html += `
                </div>

                <div class="form-group">
                    <label>Primary Color</label>
                    <div class="color-input">
                        <input type="color" name="primary_color" value="` + brandingConfig["branding_primary_color"] + `">
                        <input type="text" value="` + brandingConfig["branding_primary_color"] + `" readonly>
                    </div>
                </div>

                <div class="form-group">
                    <label>Secondary Color</label>
                    <div class="color-input">
                        <input type="color" name="secondary_color" value="` + brandingConfig["branding_secondary_color"] + `">
                        <input type="text" value="` + brandingConfig["branding_secondary_color"] + `" readonly>
                    </div>
                </div>

                <button type="submit" class="btn">Save Changes</button>
            </form>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminSettings(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get current settings
	serverURL, _ := database.DB.GetConfigValue("server_url")
	if serverURL == "" {
		serverURL = s.config.ServerURL
	}
	maxFileSizeMB, _ := database.DB.GetConfigValue("max_file_size_mb")
	if maxFileSizeMB == "" {
		maxFileSizeMB = "2000"
	}
	defaultQuotaMB, _ := database.DB.GetConfigValue("default_quota_mb")
	if defaultQuotaMB == "" {
		defaultQuotaMB = "5000"
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Settings - ` + s.config.CompanyName + `</title>
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
        .header h1 { color: ` + s.config.PrimaryColor + `; font-size: 24px; }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .header nav a:hover { color: ` + s.config.PrimaryColor + `; }
        .container {
            max-width: 800px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .card h2 {
            color: #333;
            margin-bottom: 20px;
            font-size: 20px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="text"], input[type="number"], input[type="url"] {
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
        .help-text {
            color: #666;
            font-size: 12px;
            margin-top: 4px;
        }
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .btn-primary {
            background: ` + s.config.PrimaryColor + `;
            color: white;
        }
        .btn-primary:hover {
            opacity: 0.9;
        }
        .success {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
        .error {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/admin/trash">Trash</a>
            <a href="/admin/branding">Branding</a>
            <a href="/admin/settings">Settings</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <div class="card">
            <h2>System Settings</h2>`

	if message != "" {
		if message[:5] == "Error" {
			html += `<div class="error">` + message + `</div>`
		} else {
			html += `<div class="success">` + message + `</div>`
		}
	}

	html += `
            <form method="POST" action="/admin/settings">
                <div class="form-group">
                    <label for="server_url">Server URL</label>
                    <input type="url" id="server_url" name="server_url" value="` + serverURL + `" required>
                    <p class="help-text">The public URL where this server is accessible (e.g., https://files.manvarg.se)</p>
                </div>

                <div class="form-group">
                    <label for="max_file_size_mb">Max File Size (MB)</label>
                    <input type="number" id="max_file_size_mb" name="max_file_size_mb" value="` + maxFileSizeMB + `" min="1" required>
                    <p class="help-text">Maximum file size users can upload</p>
                </div>

                <div class="form-group">
                    <label for="default_quota_mb">Default User Quota (MB)</label>
                    <input type="number" id="default_quota_mb" name="default_quota_mb" value="` + defaultQuotaMB + `" min="100" required>
                    <p class="help-text">Default storage quota for new users</p>
                </div>

                <button type="submit" class="btn btn-primary">Save Settings</button>
                <a href="/admin" class="btn" style="background: #e0e0e0; margin-left: 10px;">Cancel</a>
            </form>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminTrash(w http.ResponseWriter, files []*database.FileInfo) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Trash - ` + s.config.CompanyName + `</title>
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
        .header h1 { color: ` + s.config.PrimaryColor + `; font-size: 24px; }
        .header nav a {
            margin-left: 20px;
            color: #666;
            text-decoration: none;
            font-weight: 500;
        }
        .header nav a:hover { color: ` + s.config.PrimaryColor + `; }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .info-box {
            background: #fff3cd;
            border: 1px solid #ffc107;
            color: #856404;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        table {
            width: 100%;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 16px;
            text-align: left;
        }
        th {
            background: #f9f9f9;
            font-weight: 600;
            color: #666;
            font-size: 14px;
        }
        tr:not(:last-child) td {
            border-bottom: 1px solid #e0e0e0;
        }
        tr:hover {
            background: #f9f9f9;
        }
        .btn {
            padding: 6px 12px;
            border: none;
            border-radius: 6px;
            font-size: 13px;
            font-weight: 500;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-right: 4px;
        }
        .btn-restore { background: #4caf50; color: white; }
        .btn-delete { background: #f44336; color: white; }
        .btn:hover { opacity: 0.8; }
    </style>
</head>
<body>
    <div class="header">
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="/admin">Dashboard</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/files">Files</a>
            <a href="/admin/trash">Trash</a>
            <a href="/admin/branding">Branding</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <h2 style="margin-bottom: 20px;">Trash (Deleted Files)</h2>

        <div class="info-box">
            ⚠️ Files in trash will be automatically deleted after 5 days. You can restore or permanently delete them here.
        </div>

        <table>
            <thead>
                <tr>
                    <th>File Name</th>
                    <th>Owner</th>
                    <th>Size</th>
                    <th>Deleted At</th>
                    <th>Deleted By</th>
                    <th>Days Left</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>`

	if len(files) == 0 {
		html += `
                <tr>
                    <td colspan="7" style="text-align: center; padding: 40px; color: #999;">
                        Trash is empty
                    </td>
                </tr>`
	}

	for _, f := range files {
		// Get user info
		userName := "Unknown"
		user, err := database.DB.GetUserByID(f.UserId)
		if err == nil {
			userName = user.Name
		}

		// Get deleted by
		deletedByName := "System"
		if f.DeletedBy > 0 {
			deletedBy, err := database.DB.GetUserByID(f.DeletedBy)
			if err == nil {
				deletedByName = deletedBy.Name
			}
		}

		// Calculate days left
		deletedAt := time.Unix(f.DeletedAt, 0)
		deleteAfter := deletedAt.Add(5 * 24 * time.Hour)
		daysLeft := int(time.Until(deleteAfter).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		html += fmt.Sprintf(`
                <tr>
                    <td>📄 %s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%d days</td>
                    <td>
                        <button class="btn btn-restore" onclick="restoreFile('%s')">
                            ♻️ Restore
                        </button>
                        <button class="btn btn-delete" onclick="permanentDelete('%s')">
                            🗑️ Delete Forever
                        </button>
                    </td>
                </tr>`,
			f.Name,
			userName,
			f.Size,
			deletedAt.Format("2006-01-02 15:04"),
			deletedByName,
			daysLeft,
			f.Id,
			f.Id)
	}

	html += `
            </tbody>
        </table>
    </div>

    <script>
        async function restoreFile(fileId) {
            if (!confirm('Are you sure you want to restore this file?')) return;

            try {
                const response = await fetch('/admin/trash/restore', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: 'file_id=' + fileId
                });

                if (response.ok) {
                    location.reload();
                } else {
                    const result = await response.json();
                    alert('Restore failed: ' + (result.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Restore failed: ' + error.message);
            }
        }

        async function permanentDelete(fileId) {
            if (!confirm('⚠️ WARNING: This will PERMANENTLY delete the file. This action cannot be undone. Are you sure?')) return;

            try {
                const response = await fetch('/admin/trash/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: 'file_id=' + fileId
                });

                if (response.ok) {
                    location.reload();
                } else {
                    const result = await response.json();
                    alert('Delete failed: ' + (result.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Delete failed: ' + error.message);
            }
        }
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func calculateTotalDownloads(files []*database.FileInfo) int {
	total := 0
	for _, f := range files {
		total += f.DownloadCount
	}
	return total
}

func mustParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
