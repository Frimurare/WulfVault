package server

import (
	"fmt"
	"net/http"
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
	userLevel, _ := strconv.Atoi(r.FormValue("user_level")

)

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

	// TODO: Implement branding update
	s.renderAdminBranding(w, "Branding updated (feature in progress)")
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

	// TODO: Implement settings update
	s.renderAdminSettings(w, "Settings updated (feature in progress)")
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

        <label>Password` + func() string { if isEdit { return " (leave empty to keep current)" }; return "" }() + `:</label>
        <input type="password" name="password">

        <label>Storage Quota (MB):</label>
        <input type="number" name="quota_mb" value="` + quotaVal + `" required>

        <label>User Level:</label>
        <select name="user_level">
            <option value="2"` + func() string { if userLevelVal == "2" { return " selected" }; return "" }() + `>Regular User</option>
            <option value="1"` + func() string { if userLevelVal == "1" { return " selected" }; return "" }() + `>Admin</option>
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

	totalStorageFormatted := database.FormatFileSize(totalStorage)
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
	w.Write([]byte("<h1>Branding Settings (Coming Soon)</h1><p>" + message + "</p><a href='/admin'>Back</a>"))
}

func (s *Server) renderAdminSettings(w http.ResponseWriter, message string) {
	w.Write([]byte("<h1>System Settings (Coming Soon)</h1><p>" + message + "</p><a href='/admin'>Back</a>"))
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
