package server

import (
	"fmt"
	"net/http"
	"strconv"

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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.renderAdminBranding(w, "Failed to parse form")
		return
	}

	// Update branding configuration
	s.config.CompanyName = r.FormValue("company_name")
	s.config.PrimaryColor = r.FormValue("primary_color")
	s.config.SecondaryColor = r.FormValue("secondary_color")
	s.config.FooterText = r.FormValue("footer_text")
	s.config.WelcomeMessage = r.FormValue("welcome_message")

	// Save configuration to database
	if err := database.DB.UpdateConfiguration(s.config); err != nil {
		s.renderAdminBranding(w, "Failed to save branding: "+err.Error())
		return
	}

	s.renderAdminBranding(w, "✅ Branding updated successfully!")
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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.renderAdminSettings(w, "Failed to parse form")
		return
	}

	// Update system settings
	s.config.ServerURL = r.FormValue("server_url")
	s.config.Port = r.FormValue("port")
	s.config.MaxUploadSizeMB = mustParseInt(r.FormValue("max_upload_mb"))
	s.config.DefaultQuotaMB = mustParseInt(r.FormValue("default_quota_mb"))
	s.config.SessionTimeoutHours = mustParseInt(r.FormValue("session_timeout_hours"))

	// Save configuration
	if err := database.DB.UpdateConfiguration(s.config); err != nil {
		s.renderAdminSettings(w, "Failed to save settings: "+err.Error())
		return
	}

	s.renderAdminSettings(w, "✅ Settings updated successfully!")
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

func (s *Server) renderAdminBranding(w http.ResponseWriter, message string) {
	msgHTML := ""
	if message != "" {
		msgClass := "success"
		if len(message) > 2 && message[:2] != "✅" {
			msgClass = "error"
		}
		msgHTML = `<div class="message ` + msgClass + `">` + message + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Branding Settings - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { color: ` + s.config.PrimaryColor + `; margin-bottom: 30px; }
        .form-group { margin-bottom: 24px; }
        label { display: block; margin-bottom: 8px; font-weight: 500; color: #333; }
        input[type="text"], input[type="color"], textarea { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px; }
        textarea { min-height: 100px; font-family: inherit; }
        .color-input { display: flex; gap: 12px; align-items: center; }
        .color-input input[type="color"] { width: 80px; height: 48px; padding: 4px; cursor: pointer; }
        .color-input input[type="text"] { flex: 1; }
        .button-group { display: flex; gap: 12px; margin-top: 32px; }
        button { padding: 12px 24px; border: none; border-radius: 6px; font-size: 14px; font-weight: 500; cursor: pointer; }
        button[type="submit"] { background: ` + s.config.PrimaryColor + `; color: white; }
        button[type="submit"]:hover { opacity: 0.9; }
        .btn-secondary { background: #e0e0e0; color: #333; text-decoration: none; display: inline-block; padding: 12px 24px; }
        .btn-secondary:hover { background: #d0d0d0; }
        .message { padding: 16px; border-radius: 6px; margin-bottom: 24px; }
        .message.success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .message.error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        .preview { margin-top: 12px; padding: 20px; border: 2px dashed #ddd; border-radius: 6px; text-align: center; }
        .preview h3 { margin: 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎨 Branding Settings</h1>
        ` + msgHTML + `
        <form method="POST">
            <div class="form-group">
                <label for="company_name">Company Name</label>
                <input type="text" id="company_name" name="company_name" value="` + s.config.CompanyName + `" required>
                <small style="color: #666;">This appears in the header and page titles</small>
            </div>

            <div class="form-group">
                <label for="primary_color">Primary Color</label>
                <div class="color-input">
                    <input type="color" id="primary_color_picker" value="` + s.config.PrimaryColor + `" onchange="document.getElementById('primary_color').value = this.value; updatePreview()">
                    <input type="text" id="primary_color" name="primary_color" value="` + s.config.PrimaryColor + `" pattern="#[0-9A-Fa-f]{6}" required>
                </div>
                <small style="color: #666;">Main brand color for headers, buttons, and accents</small>
            </div>

            <div class="form-group">
                <label for="secondary_color">Secondary Color</label>
                <div class="color-input">
                    <input type="color" id="secondary_color_picker" value="` + s.config.SecondaryColor + `" onchange="document.getElementById('secondary_color').value = this.value; updatePreview()">
                    <input type="text" id="secondary_color" name="secondary_color" value="` + s.config.SecondaryColor + `" pattern="#[0-9A-Fa-f]{6}" required>
                </div>
                <small style="color: #666;">Secondary accent color</small>
            </div>

            <div class="form-group">
                <label for="welcome_message">Welcome Message</label>
                <textarea id="welcome_message" name="welcome_message" placeholder="Welcome to our secure file sharing platform">` + s.config.WelcomeMessage + `</textarea>
                <small style="color: #666;">Shown on the login page</small>
            </div>

            <div class="form-group">
                <label for="footer_text">Footer Text</label>
                <input type="text" id="footer_text" name="footer_text" value="` + s.config.FooterText + `" placeholder="© 2025 Company Name. All rights reserved.">
            </div>

            <div class="form-group">
                <label>Preview</label>
                <div class="preview" id="preview">
                    <h3 id="preview_name" style="color: ` + s.config.PrimaryColor + `">` + s.config.CompanyName + `</h3>
                    <p>Sample text with branding colors</p>
                </div>
            </div>

            <div class="button-group">
                <button type="submit">💾 Save Branding</button>
                <a href="/admin" class="btn-secondary">Cancel</a>
            </div>
        </form>
    </div>

    <script>
        function updatePreview() {
            const name = document.getElementById('company_name').value;
            const color = document.getElementById('primary_color').value;
            document.getElementById('preview_name').textContent = name;
            document.getElementById('preview_name').style.color = color;
        }
        document.getElementById('company_name').addEventListener('input', updatePreview);
        document.getElementById('primary_color').addEventListener('input', function() {
            document.getElementById('primary_color_picker').value = this.value;
            updatePreview();
        });
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminSettings(w http.ResponseWriter, message string) {
	msgHTML := ""
	if message != "" {
		msgClass := "success"
		if len(message) > 2 && message[:2] != "✅" {
			msgClass = "error"
		}
		msgHTML = `<div class="message ` + msgClass + `">` + message + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>System Settings - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { color: ` + s.config.PrimaryColor + `; margin-bottom: 30px; }
        h2 { color: #333; font-size: 18px; margin-top: 32px; margin-bottom: 16px; padding-bottom: 8px; border-bottom: 2px solid #e0e0e0; }
        .form-group { margin-bottom: 24px; }
        label { display: block; margin-bottom: 8px; font-weight: 500; color: #333; }
        input[type="text"], input[type="number"], input[type="url"] { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px; }
        small { color: #666; font-size: 13px; }
        .button-group { display: flex; gap: 12px; margin-top: 32px; }
        button { padding: 12px 24px; border: none; border-radius: 6px; font-size: 14px; font-weight: 500; cursor: pointer; }
        button[type="submit"] { background: ` + s.config.PrimaryColor + `; color: white; }
        button[type="submit"]:hover { opacity: 0.9; }
        .btn-secondary { background: #e0e0e0; color: #333; text-decoration: none; display: inline-block; padding: 12px 24px; }
        .btn-secondary:hover { background: #d0d0d0; }
        .message { padding: 16px; border-radius: 6px; margin-bottom: 24px; }
        .message.success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .message.error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        .input-group { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
        @media (max-width: 600px) { .input-group { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚙️ System Settings</h1>
        ` + msgHTML + `
        <form method="POST">
            <h2>Server Configuration</h2>

            <div class="form-group">
                <label for="server_url">Server URL</label>
                <input type="url" id="server_url" name="server_url" value="` + s.config.ServerURL + `" required>
                <small>Base URL for download links (e.g., https://share.example.com)</small>
            </div>

            <div class="input-group">
                <div class="form-group">
                    <label for="port">Port</label>
                    <input type="text" id="port" name="port" value="` + s.config.Port + `" required>
                    <small>Server listening port</small>
                </div>

                <div class="form-group">
                    <label for="max_upload_mb">Max Upload Size (MB)</label>
                    <input type="number" id="max_upload_mb" name="max_upload_mb" value="` + fmt.Sprintf("%d", s.config.MaxUploadSizeMB) + `" min="1" max="10000" required>
                    <small>Maximum file size per upload</small>
                </div>
            </div>

            <h2>User Defaults</h2>

            <div class="input-group">
                <div class="form-group">
                    <label for="default_quota_mb">Default User Quota (MB)</label>
                    <input type="number" id="default_quota_mb" name="default_quota_mb" value="` + fmt.Sprintf("%d", s.config.DefaultQuotaMB) + `" min="100" max="100000" required>
                    <small>Storage quota for new users</small>
                </div>

                <div class="form-group">
                    <label for="session_timeout_hours">Session Timeout (Hours)</label>
                    <input type="number" id="session_timeout_hours" name="session_timeout_hours" value="` + fmt.Sprintf("%d", s.config.SessionTimeoutHours) + `" min="1" max="720" required>
                    <small>Auto-logout after inactivity</small>
                </div>
            </div>

            <div class="button-group">
                <button type="submit">💾 Save Settings</button>
                <a href="/admin" class="btn-secondary">Cancel</a>
            </div>
        </form>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

func mustParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
