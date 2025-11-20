// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	emailpkg "github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// handleAdminDashboard renders the admin dashboard
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get basic statistics
	totalUsers, _ := database.DB.GetTotalUsers()
	activeUsers, _ := database.DB.GetActiveUsers()
	totalDownloads, _ := database.DB.GetTotalDownloads()
	downloadsToday, _ := database.DB.GetDownloadsToday()

	// Get data transfer statistics (downloaded)
	bytesDownloadedToday, _ := database.DB.GetBytesSentToday()
	bytesDownloadedWeek, _ := database.DB.GetBytesSentThisWeek()
	bytesDownloadedMonth, _ := database.DB.GetBytesSentThisMonth()
	bytesDownloadedYear, _ := database.DB.GetBytesSentThisYear()

	// Get upload statistics
	bytesUploadedToday, _ := database.DB.GetBytesUploadedToday()
	bytesUploadedWeek, _ := database.DB.GetBytesUploadedThisWeek()
	bytesUploadedMonth, _ := database.DB.GetBytesUploadedThisMonth()
	bytesUploadedYear, _ := database.DB.GetBytesUploadedThisYear()

	// Get user growth statistics
	usersAdded, _ := database.DB.GetUsersAddedThisMonth()
	usersRemoved, _ := database.DB.GetUsersRemovedThisMonth()
	userGrowth, _ := database.DB.GetUserGrowthPercentage()

	// Get usage statistics
	activeFiles7Days, _ := database.DB.GetActiveFilesLast7Days()
	activeFiles30Days, _ := database.DB.GetActiveFilesLast30Days()
	avgFileSize, _ := database.DB.GetAverageFileSize()
	avgDownloadsPerFile, _ := database.DB.GetAverageDownloadsPerFile()

	// Get security statistics
	twoFAAdoption, _ := database.DB.Get2FAAdoptionRate()
	avgBackupCodes, _ := database.DB.GetAverageBackupCodesRemaining()

	// Get file statistics
	largestFileName, largestFileSize, _ := database.DB.GetLargestFile()
	mostActiveUser, userFileCount, _ := database.DB.GetMostActiveUser()

	// Get trend data
	topFileTypes, fileTypeCounts, _ := database.DB.GetTopFileTypes()
	topWeekday, weekdayCount, _ := database.DB.GetMostActiveWeekday()
	storagePast, storageNow, _ := database.DB.GetStorageTrendLastMonth()

	// Get fun fact
	mostDownloadedFile, downloadCount, _ := database.DB.GetMostDownloadedFile()

	s.renderAdminDashboard(w, user, totalUsers, activeUsers, totalDownloads, downloadsToday,
		bytesDownloadedToday, bytesDownloadedWeek, bytesDownloadedMonth, bytesDownloadedYear,
		bytesUploadedToday, bytesUploadedWeek, bytesUploadedMonth, bytesUploadedYear,
		usersAdded, usersRemoved, userGrowth,
		activeFiles7Days, activeFiles30Days, avgFileSize, avgDownloadsPerFile,
		twoFAAdoption, avgBackupCodes,
		largestFileName, largestFileSize, mostActiveUser, userFileCount,
		topFileTypes, fileTypeCounts, topWeekday, weekdayCount, storagePast, storageNow,
		mostDownloadedFile, downloadCount)
}

// handleAdminUsers lists all users and download accounts with pagination
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	// Parse user filter parameters
	userFilter := &database.UserFilter{}

	userFilter.SearchTerm = r.URL.Query().Get("search")

	if levelStr := r.URL.Query().Get("level"); levelStr != "" {
		if level, err := strconv.Atoi(levelStr); err == nil {
			userFilter.UserLevel = level
		}
	}

	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			userFilter.IsActive = &active
		}
	}

	userFilter.SortBy = r.URL.Query().Get("sort_by")
	userFilter.SortOrder = r.URL.Query().Get("sort_order")

	// Pagination for users
	userLimit := 50
	if limitStr := r.URL.Query().Get("user_limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			userLimit = l
		}
	}
	userFilter.Limit = userLimit

	userOffset := 0
	if offsetStr := r.URL.Query().Get("user_offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			userOffset = o
		}
	}
	userFilter.Offset = userOffset

	// Get users with pagination
	users, err := database.DB.GetUsers(userFilter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Get total user count
	userCount, err := database.DB.GetUserCount(userFilter)
	if err != nil {
		log.Printf("Warning: Failed to get user count: %v", err)
		userCount = 0
	}

	// Parse download account filter parameters
	downloadFilter := &database.DownloadAccountFilter{}

	downloadFilter.SearchTerm = r.URL.Query().Get("dl_search")

	if dlActiveStr := r.URL.Query().Get("dl_active"); dlActiveStr != "" {
		if active, err := strconv.ParseBool(dlActiveStr); err == nil {
			downloadFilter.IsActive = &active
		}
	}

	downloadFilter.SortBy = r.URL.Query().Get("dl_sort_by")
	downloadFilter.SortOrder = r.URL.Query().Get("dl_sort_order")

	// Pagination for download accounts
	dlLimit := 50
	if limitStr := r.URL.Query().Get("dl_limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			dlLimit = l
		}
	}
	downloadFilter.Limit = dlLimit

	dlOffset := 0
	if offsetStr := r.URL.Query().Get("dl_offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			dlOffset = o
		}
	}
	downloadFilter.Offset = dlOffset

	// Get download accounts with pagination
	downloadAccounts, err := database.DB.GetDownloadAccounts(downloadFilter)
	if err != nil {
		log.Printf("Warning: Failed to fetch download accounts: %v", err)
		downloadAccounts = []*models.DownloadAccount{}
	}

	// Get total download account count
	downloadCount, err := database.DB.GetDownloadAccountCount(downloadFilter)
	if err != nil {
		log.Printf("Warning: Failed to get download account count: %v", err)
		downloadCount = 0
	}

	s.renderAdminUsers(w, users, downloadAccounts, userFilter, userCount, downloadFilter, downloadCount)
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
	sendWelcomeEmail := r.FormValue("send_welcome_email") == "1"

	// Validate
	if name == "" || email == "" {
		s.renderAdminUserForm(w, nil, "Name and email are required")
		return
	}

	// If not sending welcome email, password is required
	if !sendWelcomeEmail && password == "" {
		s.renderAdminUserForm(w, nil, "Password is required (or check 'Send welcome email')")
		return
	}

	// Hash password (use temporary password if sending welcome email)
	var err error
	if sendWelcomeEmail {
		// Generate temporary random password that will be replaced via email
		tempBytes := make([]byte, 32)
		if _, err := rand.Read(tempBytes); err != nil {
			s.renderAdminUserForm(w, nil, "Failed to generate temporary password")
			return
		}
		tempPassword := hex.EncodeToString(tempBytes)
		password, err = auth.HashPassword(tempPassword)
		if err != nil {
			s.renderAdminUserForm(w, nil, "Failed to generate temporary password")
			return
		}
	} else {
		password, err = auth.HashPassword(password)
		if err != nil {
			s.renderAdminUserForm(w, nil, "Failed to hash password")
			return
		}
	}

	// Create user
	newUser := &models.User{
		Name:           name,
		Email:          email,
		Password:       password,
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

	// Log the action
	admin, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(admin.Id),
		UserEmail:  admin.Email,
		Action:     "USER_CREATED",
		EntityType: "User",
		EntityID:   fmt.Sprintf("%d", newUser.Id),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"user_level\":%d,\"quota_mb\":%d}", newUser.Email, newUser.Name, newUser.UserLevel, newUser.StorageQuotaMB),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

	// Send welcome email if requested
	if sendWelcomeEmail {
		// Create password reset token
		resetToken, err := database.DB.CreatePasswordResetToken(email, database.AccountTypeUser)
		if err != nil {
			log.Printf("Failed to create reset token for welcome email: %v", err)
			// Don't fail user creation, just log the error
		} else {
			// Get admin info from context
			admin, ok := userFromContext(r.Context())
			if !ok {
				log.Printf("Failed to get admin info from context for welcome email")
				return
			}

			// Get branding info
			companyName := s.config.CompanyName

			// Fix server URL - ensure it uses http:// if running on port 8080 without SSL
			emailServerURL := s.config.ServerURL
			// Replace https:// with http:// if present (since we don't have SSL on port 8080)
			if len(emailServerURL) > 8 && emailServerURL[:8] == "https://" {
				emailServerURL = "http://" + emailServerURL[8:]
				log.Printf("Corrected server URL from HTTPS to HTTP for email: %s", emailServerURL)
			}

			// Send welcome email with admin info
			if err := emailpkg.SendWelcomeEmail(email, resetToken, emailServerURL, companyName, admin.Name, admin.Email); err != nil {
				log.Printf("Failed to send welcome email to %s: %v", email, err)
				// Don't fail user creation, just log the error
			} else {
				log.Printf("Welcome email sent to new user: %s (%s) by admin %s", name, email, admin.Name)
			}
		}
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

	// Log the action
	admin, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(admin.Id),
		UserEmail:  admin.Email,
		Action:     "USER_UPDATED",
		EntityType: "User",
		EntityID:   fmt.Sprintf("%d", existingUser.Id),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"user_level\":%d,\"is_active\":%t}", existingUser.Email, existingUser.Name, existingUser.UserLevel, existingUser.IsActive),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handleAdminUserDelete deletes a user
func (s *Server) handleAdminUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get admin user from context
	admin, ok := userFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := strconv.Atoi(r.FormValue("id"))
	if userID == 0 {
		s.sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user info before deletion for audit log
	userToDelete, err := database.DB.GetUserByID(userID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	// Delete user (this will also soft-delete all their files to trash)
	if err := database.DB.DeleteUser(userID, admin.Id); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(admin.Id),
		UserEmail:  admin.Email,
		Action:     "USER_DELETED",
		EntityType: "User",
		EntityID:   fmt.Sprintf("%d", userID),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"user_level\":%d}", userToDelete.Email, userToDelete.Name, userToDelete.UserLevel),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "User deleted, files moved to trash"})
}

// handleAdminToggleDownloadAccount toggles download account active status
func (s *Server) handleAdminToggleDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountID, _ := strconv.Atoi(r.FormValue("id"))
	if accountID == 0 {
		s.sendError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	// Get account
	account, err := database.DB.GetDownloadAccountByID(accountID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Download account not found")
		return
	}

	// Toggle active status
	account.IsActive = !account.IsActive
	if err := database.DB.UpdateDownloadAccount(account); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "Account updated"})
}

// handleAdminCreateDownloadAccount creates a new download account
func (s *Server) handleAdminCreateDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderAdminDownloadAccountForm(w, nil, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderAdminDownloadAccountForm(w, nil, "Invalid form data")
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate
	if name == "" || email == "" || password == "" {
		s.renderAdminDownloadAccountForm(w, nil, "All fields are required")
		return
	}

	// Check if account already exists
	existing, _ := database.DB.GetDownloadAccountByEmail(email)
	if existing != nil {
		s.renderAdminDownloadAccountForm(w, nil, "Account with this email already exists")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		s.renderAdminDownloadAccountForm(w, nil, "Failed to hash password")
		return
	}

	// Create account
	account := &models.DownloadAccount{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := database.DB.CreateDownloadAccount(account); err != nil {
		s.renderAdminDownloadAccountForm(w, nil, "Failed to create account: "+err.Error())
		return
	}

	log.Printf("Admin created download account: %s", email)

	// Log the action
	admin, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(admin.Id),
		UserEmail:  admin.Email,
		Action:     "DOWNLOAD_ACCOUNT_CREATED",
		EntityType: "DownloadAccount",
		EntityID:   fmt.Sprintf("%d", account.Id),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"admin_created\":true}", account.Email, account.Name),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handleAdminEditDownloadAccount edits a download account
func (s *Server) handleAdminEditDownloadAccount(w http.ResponseWriter, r *http.Request) {
	accountID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if accountID == 0 {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	existingAccount, err := database.DB.GetDownloadAccountByID(accountID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		s.renderAdminDownloadAccountForm(w, existingAccount, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderAdminDownloadAccountForm(w, existingAccount, "Invalid form data")
		return
	}

	existingAccount.Name = r.FormValue("name")
	existingAccount.Email = r.FormValue("email")
	existingAccount.IsActive = r.FormValue("is_active") == "1"

	// Update password if provided
	newPassword := r.FormValue("password")
	if newPassword != "" {
		hashedPassword, err := auth.HashPassword(newPassword)
		if err != nil {
			s.renderAdminDownloadAccountForm(w, existingAccount, "Failed to hash password")
			return
		}
		existingAccount.Password = hashedPassword
	}

	if err := database.DB.UpdateDownloadAccount(existingAccount); err != nil {
		s.renderAdminDownloadAccountForm(w, existingAccount, "Failed to update account: "+err.Error())
		return
	}

	log.Printf("Admin updated download account: ID=%d, Email=%s", accountID, existingAccount.Email)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handleAdminDeleteDownloadAccount soft deletes a download account
func (s *Server) handleAdminDeleteDownloadAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountID, _ := strconv.Atoi(r.FormValue("id"))
	if accountID == 0 {
		s.sendError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	// Get account for logging
	account, err := database.DB.GetDownloadAccountByID(accountID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Account not found")
		return
	}

	// Soft delete the account
	if err := database.DB.SoftDeleteDownloadAccount(accountID, "admin"); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	log.Printf("Admin soft deleted download account: ID=%d, Email=%s", accountID, account.Email)

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "DOWNLOAD_ACCOUNT_DELETED",
		EntityType: "DownloadAccount",
		EntityID:   fmt.Sprintf("%d", accountID),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"soft_delete\":true,\"admin_deleted\":true}", account.Email, account.Name),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "Account deleted"})
}

// renderAdminDownloadAccountForm renders the download account form
func (s *Server) renderAdminDownloadAccountForm(w http.ResponseWriter, account *models.DownloadAccount, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	isEdit := account != nil
	title := "Create Download Account"
	action := "/admin/download-accounts/create"

	if isEdit {
		title = "Edit Download Account"
		action = fmt.Sprintf("/admin/download-accounts/edit?id=%d", account.Id)
	}

	nameVal, emailVal := "", ""
	if isEdit {
		nameVal = account.Name
		emailVal = account.Email
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <meta name="author" content="Ulf Holmström">
    <title>` + title + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 600px; margin: 40px auto; padding: 20px; background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h2 { margin-bottom: 24px; color: #333; }
        input { width: 100%; padding: 8px; margin: 8px 0; }
        button { padding: 10px 20px; background: ` + s.getPrimaryColor() + `; color: white; border: none; cursor: pointer; border-radius: 6px; }
        .error { background: #fee; padding: 10px; margin: 10px 0; border-radius: 4px; color: #c33; }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <h2>` + title + `</h2>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
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
            <input type="password" name="password"` + func() string {
		if !isEdit {
			return " required"
		}
		return ""
	}() + `>

            <br><br>
            <label style="display: flex; align-items: center; cursor: pointer;">
                <input type="checkbox" name="is_active" value="1"` + func() string {
		if isEdit && account.IsActive {
			return " checked"
		} else if !isEdit {
			return " checked"
		}
		return ""
	}() + ` style="width: auto; margin-right: 8px;">
                <span>Active (account can log in)</span>
            </label>

            <br><br>
            <button type="submit">Save</button>
            <a href="/admin/users">Cancel</a>
        </form>
    </div>
</body>
</html>`

	w.Write([]byte(html))
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

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "BRANDING_UPDATED",
		EntityType: "Settings",
		EntityID:   "branding",
		Details:    fmt.Sprintf("{\"company_name\":\"%s\",\"has_logo\":%v}", companyName, logoData != ""),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

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

	// Update settings in database and config
	serverURL := r.FormValue("server_url")
	if serverURL != "" {
		// Strip port from URL if present (port is configured separately)
		serverURL = stripPortFromURL(serverURL)
		database.DB.SetConfigValue("server_url", serverURL)
		s.config.ServerURL = serverURL
	}

	// Handle port change - ONLY if port actually changed
	port := r.FormValue("port")
	currentPort, _ := database.DB.GetConfigValue("port")
	if currentPort == "" {
		currentPort = s.config.Port
	}

	portChanged := false
	if port != "" && port != currentPort {
		// Validate port number
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			s.renderAdminSettings(w, "Error: Invalid port number (must be 1-65535)")
			return
		}

		// Warn if port < 1024 (requires root privileges)
		if portNum < 1024 {
			s.renderAdminSettings(w, "Warning: Ports below 1024 require root/administrator privileges. Change saved but may fail to bind.")
			// Still allow the change but show warning
		}

		// Update config.json file
		if err := s.updateConfigJSON("port", port); err != nil {
			log.Printf("Error updating config.json: %v", err)
			s.renderAdminSettings(w, "Error: Failed to save port to config file")
			return
		}

		// Store in database for reference
		database.DB.SetConfigValue("port", port)
		portChanged = true
	}

	maxFileSizeMB := r.FormValue("max_file_size_mb")
	if maxFileSizeMB != "" {
		database.DB.SetConfigValue("max_file_size_mb", maxFileSizeMB)
	}

	defaultQuotaMB := r.FormValue("default_quota_mb")
	if defaultQuotaMB != "" {
		database.DB.SetConfigValue("default_quota_mb", defaultQuotaMB)
	}

	trashRetentionDays := r.FormValue("trash_retention_days")
	if trashRetentionDays != "" {
		database.DB.SetConfigValue("trash_retention_days", trashRetentionDays)
		if days, err := strconv.Atoi(trashRetentionDays); err == nil {
			s.config.TrashRetentionDays = days
		}
	}

	auditLogRetentionDays := r.FormValue("audit_log_retention_days")
	if auditLogRetentionDays != "" {
		database.DB.SetConfigValue("audit_log_retention_days", auditLogRetentionDays)
		if days, err := strconv.Atoi(auditLogRetentionDays); err == nil {
			s.config.AuditLogRetentionDays = days
		}
	}

	auditLogMaxSizeMB := r.FormValue("audit_log_max_size_mb")
	if auditLogMaxSizeMB != "" {
		database.DB.SetConfigValue("audit_log_max_size_mb", auditLogMaxSizeMB)
		if sizeMB, err := strconv.Atoi(auditLogMaxSizeMB); err == nil {
			s.config.AuditLogMaxSizeMB = sizeMB
		}
	}

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "SETTINGS_UPDATED",
		EntityType: "Settings",
		EntityID:   "general",
		Details:    fmt.Sprintf("{\"server_url\":\"%s\",\"port_changed\":%v}", serverURL, portChanged),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	// Show appropriate success message
	if portChanged {
		s.renderAdminSettings(w, fmt.Sprintf("Port changed to %s. ⚠️ RESTART REQUIRED: Stop and start the server for changes to take effect.", port))
	} else {
		s.renderAdminSettings(w, "Settings updated successfully!")
	}
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

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "FILE_PERMANENTLY_DELETED",
		EntityType: "File",
		EntityID:   fileID,
		Details:    fmt.Sprintf("{\"filename\":\"%s\",\"size\":\"%s\"}", fileInfo.Name, fileInfo.Size),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

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

	// Get file info before restore for audit log
	deletedFiles, err := database.DB.GetDeletedFiles()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	var fileInfo *database.FileInfo
	for _, f := range deletedFiles {
		if f.Id == fileID {
			fileInfo = f
			break
		}
	}

	if fileInfo == nil {
		s.sendError(w, http.StatusNotFound, "File not found in trash")
		return
	}

	if err := database.DB.RestoreFile(fileID); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to restore file")
		return
	}

	log.Printf("File restored from trash by admin: %s", fileID)

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "FILE_RESTORED",
		EntityType: "File",
		EntityID:   fileID,
		Details:    fmt.Sprintf("{\"filename\":\"%s\",\"size\":\"%s\"}", fileInfo.Name, fileInfo.Size),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
	})

	s.sendJSON(w, http.StatusOK, map[string]string{
		"message": "File restored successfully",
	})
}

// Render functions

// getAdminHeaderHTML returns branded header HTML for admin pages

func (s *Server) renderAdminDashboard(w http.ResponseWriter, user *models.User, totalUsers, activeUsers, totalDownloads, downloadsToday int,
	bytesDownloadedToday, bytesDownloadedWeek, bytesDownloadedMonth, bytesDownloadedYear int64,
	bytesUploadedToday, bytesUploadedWeek, bytesUploadedMonth, bytesUploadedYear int64,
	usersAdded, usersRemoved int, userGrowth float64,
	activeFiles7Days, activeFiles30Days int, avgFileSize int64, avgDownloadsPerFile float64,
	twoFAAdoption, avgBackupCodes float64,
	largestFileName string, largestFileSize int64, mostActiveUser string, userFileCount int,
	topFileTypes []string, fileTypeCounts []int, topWeekday string, weekdayCount int, storagePast, storageNow int64,
	mostDownloadedFile string, downloadCount int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get joke of the day
	joke := models.GetJokeOfTheDay()

	// Helper function to format bytes
	formatBytes := func(bytes int64) string {
		const unit = 1024
		if bytes < unit {
			return fmt.Sprintf("%d B", bytes)
		}
		div, exp := int64(unit), 0
		for n := bytes / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
	}

	// Format all byte statistics (downloaded)
	bytesDownloadedTodayStr := formatBytes(bytesDownloadedToday)
	bytesDownloadedWeekStr := formatBytes(bytesDownloadedWeek)
	bytesDownloadedMonthStr := formatBytes(bytesDownloadedMonth)
	bytesDownloadedYearStr := formatBytes(bytesDownloadedYear)

	// Format upload statistics
	bytesUploadedTodayStr := formatBytes(bytesUploadedToday)
	bytesUploadedWeekStr := formatBytes(bytesUploadedWeek)
	bytesUploadedMonthStr := formatBytes(bytesUploadedMonth)
	bytesUploadedYearStr := formatBytes(bytesUploadedYear)

	// Format usage statistics
	avgFileSizeStr := formatBytes(avgFileSize)

	// Format file statistics
	largestFileSizeStr := formatBytes(largestFileSize)

	// Format trend data - file types string
	fileTypesStr := ""
	if len(topFileTypes) > 0 {
		for i, ext := range topFileTypes {
			if i > 0 {
				fileTypesStr += ", "
			}
			fileTypesStr += fmt.Sprintf(".%s (%d)", ext, fileTypeCounts[i])
		}
	}

	// Format storage trend
	storagePastStr := formatBytes(storagePast)
	storageNowStr := formatBytes(storageNow)
	storageGrowth := 0.0
	if storagePast > 0 {
		storageGrowth = float64(storageNow-storagePast) / float64(storagePast) * 100
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Admin Dashboard - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 1400px; margin: 40px auto; padding: 0 20px; }
        h2 { margin-bottom: 24px; color: #333; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .stat-card { background: white; padding: 24px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .stat-card h3 { color: #888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; }
        .stat-card .value { font-size: 36px; font-weight: 700; color: #1a1a2e; }
        .quick-actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 40px; }
        .action-btn { background: white; padding: 20px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); text-align: center; text-decoration: none; color: #333; font-weight: 500; transition: transform 0.2s; }
        .action-btn:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
        .joke-section { margin: 30px 0; padding: 25px 30px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border-radius: 15px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3); }
        .joke-title { color: rgba(255,255,255,0.9); font-size: 14px; font-weight: 600; margin-bottom: 12px; text-transform: uppercase; letter-spacing: 1.5px; }
        .joke-text { color: white; font-size: 17px; line-height: 1.6; font-weight: 500; font-style: italic; }
        @media screen and (max-width: 768px) {
            .container { padding: 0 15px !important; margin: 20px auto !important; }
            .stats { grid-template-columns: 1fr !important; gap: 15px !important; }
            .stat-card { padding: 20px !important; }
            .stat-card .value { font-size: 28px !important; }
            .quick-actions { grid-template-columns: 1fr !important; }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <div class="joke-section">
            <div class="joke-title">💡 File Sharing Wisdom</div>
            <div class="joke-text">` + joke.Text + `</div>
        </div>

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

        <h2 style="margin-top: 40px;">📥 Downloaded Data</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #3b82f6;">
                <h3>Today</h3>
                <div class="value">` + bytesDownloadedTodayStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #3b82f6;">
                <h3>This Week</h3>
                <div class="value">` + bytesDownloadedWeekStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #3b82f6;">
                <h3>This Month</h3>
                <div class="value">` + bytesDownloadedMonthStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #3b82f6;">
                <h3>This Year</h3>
                <div class="value">` + bytesDownloadedYearStr + `</div>
            </div>
        </div>

        <h2 style="margin-top: 40px;">📤 Uploaded Data</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #10b981;">
                <h3>Today</h3>
                <div class="value">` + bytesUploadedTodayStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #10b981;">
                <h3>This Week</h3>
                <div class="value">` + bytesUploadedWeekStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #10b981;">
                <h3>This Month</h3>
                <div class="value">` + bytesUploadedMonthStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #10b981;">
                <h3>This Year</h3>
                <div class="value">` + bytesUploadedYearStr + `</div>
            </div>
        </div>

        <h2 style="margin-top: 40px;">👥 User Growth (This Month)</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #4CAF50;">
                <h3>Users Added</h3>
                <div class="value" style="color: #4CAF50;">` + fmt.Sprintf("%d", usersAdded) + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #f44336;">
                <h3>Users Removed</h3>
                <div class="value" style="color: #f44336;">` + fmt.Sprintf("%d", usersRemoved) + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #2196F3;">
                <h3>Growth</h3>
                <div class="value" style="color: #2196F3;">` + fmt.Sprintf("%.1f%%", userGrowth) + `</div>
            </div>
        </div>

        <h2 style="margin-top: 40px;">📈 Usage Statistics</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #9C27B0;">
                <h3>Active Files (7 days)</h3>
                <div class="value" style="color: #9C27B0;">` + fmt.Sprintf("%d", activeFiles7Days) + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #673AB7;">
                <h3>Active Files (30 days)</h3>
                <div class="value" style="color: #673AB7;">` + fmt.Sprintf("%d", activeFiles30Days) + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #3F51B5;">
                <h3>Avg File Size</h3>
                <div class="value" style="color: #3F51B5;">` + avgFileSizeStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #2196F3;">
                <h3>Avg Downloads/File</h3>
                <div class="value" style="color: #2196F3;">` + fmt.Sprintf("%.1f", avgDownloadsPerFile) + `</div>
            </div>
        </div>

        <h2 style="margin-top: 40px;">🔐 Security Overview</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #8b5cf6; grid-column: span 2;">
                <h3>2FA Adoption Rate</h3>
                <div class="value">` + fmt.Sprintf("%.1f%%", twoFAAdoption) + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 14px;">Percentage of Users/Admins with 2FA enabled</p>
            </div>
            <div class="stat-card" style="border-left: 4px solid #8b5cf6; grid-column: span 2;">
                <h3>Avg Backup Codes Remaining</h3>
                <div class="value">` + fmt.Sprintf("%.1f", avgBackupCodes) + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 14px;">Average per user with 2FA enabled</p>
            </div>
        </div>

        <h2 style="margin-top: 40px;">📁 File Statistics</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #f59e0b; grid-column: span 2;">
                <h3>Largest File</h3>
                <div class="value" style="font-size: 24px; word-break: break-word;">` + largestFileName + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 16px;">` + largestFileSizeStr + `</p>
            </div>
            <div class="stat-card" style="border-left: 4px solid #f59e0b; grid-column: span 2;">
                <h3>Most Active User</h3>
                <div class="value" style="font-size: 24px;">` + mostActiveUser + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 16px;">` + fmt.Sprintf("%d files uploaded", userFileCount) + `</p>
            </div>
        </div>

        <h2 style="margin-top: 40px;">⚡ Trend Data</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #64748b;">
                <h3>Top File Types</h3>
                <div class="value" style="font-size: 16px; word-break: break-word;">` + fileTypesStr + `</div>
            </div>
            <div class="stat-card" style="border-left: 4px solid #64748b;">
                <h3>Most Active Day</h3>
                <div class="value" style="font-size: 20px;">` + topWeekday + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 14px;">` + fmt.Sprintf("%d downloads", weekdayCount) + `</p>
            </div>
            <div class="stat-card" style="border-left: 4px solid #64748b; grid-column: span 2;">
                <h3>Storage Trend (Last 30 Days)</h3>
                <div class="value" style="font-size: 20px;">` + fmt.Sprintf("%+.1f%%", storageGrowth) + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 14px;">` + storagePastStr + ` → ` + storageNowStr + `</p>
            </div>
        </div>

        <h2 style="margin-top: 40px;">🎯 Fun Fact</h2>
        <div class="stats">
            <div class="stat-card" style="border-left: 4px solid #ec4899; grid-column: span 2;">
                <h3>Most Downloaded File</h3>
                <div class="value" style="font-size: 24px; word-break: break-word;">` + mostDownloadedFile + `</div>
                <p style="color: #666; margin-top: 10px; font-size: 16px;">` + fmt.Sprintf("%d downloads", downloadCount) + `</p>
            </div>
        </div>
    </div>
    <div style="text-align: center; padding: 40px 20px 20px; color: #999; font-size: 12px;">
        Powered by WulfVault Version ` + s.config.Version + `
    </div>
    
</body>
</html>`

	w.Write([]byte(html))
}

func (s *Server) renderAdminUsers(w http.ResponseWriter, users []*models.User, downloadAccounts []*models.DownloadAccount,
	userFilter *database.UserFilter, userCount int, dlFilter *database.DownloadAccountFilter, dlCount int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Manage Users - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
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
            background: ` + s.getPrimaryColor() + `;
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
            margin-bottom: 30px;
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
        .badge-download { background: #fff3e0; color: #e65100; }
        .action-links a {
            margin-right: 12px;
            color: ` + s.getPrimaryColor() + `;
            text-decoration: none;
        }
        h3 {
            margin: 30px 0 16px 0;
            color: #333;
        }
        .filters {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .filters-row {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
            align-items: flex-end;
        }
        .filter-group {
            flex: 1;
            min-width: 200px;
        }
        .filter-group label {
            display: block;
            margin-bottom: 6px;
            font-size: 14px;
            color: #666;
            font-weight: 500;
        }
        .filter-group input, .filter-group select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
        }
        .filter-group input:focus, .filter-group select:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .filter-btn {
            padding: 10px 20px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            white-space: nowrap;
        }
        .filter-btn:hover {
            opacity: 0.9;
        }
        .clear-btn {
            padding: 10px 20px;
            background: #f0f0f0;
            color: #333;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            white-space: nowrap;
        }
        .clear-btn:hover {
            background: #e0e0e0;
        }
        .pagination {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin: 20px 0;
            padding: 16px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .pagination-info {
            color: #666;
            font-size: 14px;
        }
        .pagination-controls {
            display: flex;
            gap: 12px;
        }
        .pagination-controls button {
            padding: 8px 16px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
        }
        .pagination-controls button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .pagination-controls button:not(:disabled):hover {
            opacity: 0.9;
        }

        /* Mobile Responsive Styles */
        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .actions {
                flex-direction: column;
                align-items: stretch !important;
                gap: 15px;
            }
            .actions h2 {
                font-size: 20px;
            }
            .btn {
                width: 100%;
                text-align: center;
            }
            table {
                border: 0;
                display: block;
                overflow-x: auto;
            }
            table thead {
                display: none;
            }
            table tbody {
                display: block;
            }
            table tr {
                display: block;
                margin-bottom: 20px;
                border: 1px solid #ddd;
                border-radius: 8px;
                padding: 15px;
                background: white;
            }
            table td {
                display: block;
                text-align: left;
                padding: 12px 0;
                border-bottom: 1px solid #eee;
            }
            table td:last-child {
                border-bottom: none;
            }
            table td::before {
                content: attr(data-label);
                display: block;
                font-weight: 600;
                color: #666;
                margin-bottom: 4px;
                font-size: 13px;
            }
            table td:last-child::before {
                display: none;
            }
            .action-links {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }
            .action-links a {
                margin: 0 !important;
                padding: 8px 12px;
                background: #f0f0f0;
                border-radius: 4px;
                text-align: center;
                display: block;
            }
            .filters-row {
                flex-direction: column;
            }
            .filter-group {
                width: 100%;
                min-width: unset;
            }
            .pagination {
                flex-direction: column;
                gap: 12px;
            }
            .pagination-info {
                text-align: center;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <div class="actions">
            <h2>Manage Users</h2>
            <a href="/admin/users/create" class="btn">+ Create User</a>
        </div>

        <!-- User Filters -->
        <div class="filters">
            <form method="GET" action="/admin/users">
                <div class="filters-row">
                    <div class="filter-group">
                        <label for="search">Search</label>
                        <input type="text" id="search" name="search" placeholder="Search name or email..." value="` + userFilter.SearchTerm + `">
                    </div>
                    <div class="filter-group">
                        <label for="level">User Level</label>
                        <select id="level" name="level">
                            <option value="0"` + func() string {
		if userFilter.UserLevel == 0 {
			return " selected"
		}
		return ""
	}() + `>All Users</option>
                            <option value="1"` + func() string {
		if userFilter.UserLevel == 1 {
			return " selected"
		}
		return ""
	}() + `>Regular Users</option>
                            <option value="2"` + func() string {
		if userFilter.UserLevel == 2 {
			return " selected"
		}
		return ""
	}() + `>Admins</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label for="active">Status</label>
                        <select id="active" name="active">
                            <option value="">All</option>
                            <option value="true"` + func() string {
		if userFilter.IsActive != nil && *userFilter.IsActive {
			return " selected"
		}
		return ""
	}() + `>Active</option>
                            <option value="false"` + func() string {
		if userFilter.IsActive != nil && !*userFilter.IsActive {
			return " selected"
		}
		return ""
	}() + `>Inactive</option>
                        </select>
                    </div>
                    <button type="submit" class="filter-btn">Filter</button>
                    <button type="button" class="clear-btn" onclick="window.location.href='/admin/users'">Clear</button>
                </div>
            </form>
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

	// Regular users
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
                    <td data-label="Name">%s</td>
                    <td data-label="Email">%s</td>
                    <td data-label="Level">%s</td>
                    <td data-label="Quota">%d GB</td>
                    <td data-label="Used">%d MB</td>
                    <td data-label="Status">%s</td>
                    <td data-label="Actions" class="action-links">
                        <a href="/admin/users/edit?id=%d">Edit</a>
                        <a href="#" onclick="deleteUser(%d); return false;">Delete</a>
                    </td>
                </tr>`,
			u.Name, u.Email, levelBadge, u.StorageQuotaMB/1000, u.StorageUsedMB, status, u.Id, u.Id)
	}

	html += `
            </tbody>
        </table>

        <!-- User Pagination -->`

	// Calculate pagination info for users
	userStart := userFilter.Offset + 1
	userEnd := userFilter.Offset + len(users)
	hasPrevUser := userFilter.Offset > 0
	hasNextUser := userFilter.Offset+userFilter.Limit < userCount

	html += `
        <div class="pagination">
            <div class="pagination-info">
                Showing ` + fmt.Sprintf("%d-%d", userStart, userEnd) + ` of ` + fmt.Sprintf("%d", userCount) + ` users
            </div>
            <div class="pagination-controls">
                <button onclick="changePage(-1, 'user')" ` + func() string {
		if !hasPrevUser {
			return "disabled"
		}
		return ""
	}() + `>Previous</button>
                <button onclick="changePage(1, 'user')" ` + func() string {
		if !hasNextUser {
			return "disabled"
		}
		return ""
	}() + `>Next</button>
            </div>
        </div>

        <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 40px; margin-bottom: 16px;">
            <h3>Download Accounts (` + fmt.Sprintf("%d", dlCount) + `)</h3>
            <a href="/admin/download-accounts/create" class="btn">+ Create Download Account</a>
        </div>

        <!-- Download Account Filters -->
        <div class="filters">
            <form method="GET" action="/admin/users">
                <div class="filters-row">
                    <div class="filter-group">
                        <label for="dl_search">Search</label>
                        <input type="text" id="dl_search" name="dl_search" placeholder="Search name or email..." value="` + dlFilter.SearchTerm + `">
                    </div>
                    <div class="filter-group">
                        <label for="dl_active">Status</label>
                        <select id="dl_active" name="dl_active">
                            <option value="">All</option>
                            <option value="true"` + func() string {
		if dlFilter.IsActive != nil && *dlFilter.IsActive {
			return " selected"
		}
		return ""
	}() + `>Active</option>
                            <option value="false"` + func() string {
		if dlFilter.IsActive != nil && !*dlFilter.IsActive {
			return " selected"
		}
		return ""
	}() + `>Inactive</option>
                        </select>
                    </div>
                    <!-- Preserve user filters -->
                    <input type="hidden" name="search" value="` + userFilter.SearchTerm + `">
                    <input type="hidden" name="level" value="` + fmt.Sprintf("%d", userFilter.UserLevel) + `">` + func() string {
		if userFilter.IsActive != nil {
			return `<input type="hidden" name="active" value="` + fmt.Sprintf("%t", *userFilter.IsActive) + `">`
		}
		return ""
	}() + `
                    <input type="hidden" name="user_offset" value="` + fmt.Sprintf("%d", userFilter.Offset) + `">
                    <button type="submit" class="filter-btn">Filter</button>
                    <button type="button" class="clear-btn" onclick="clearDownloadFilters()">Clear</button>
                </div>
            </form>
        </div>

        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Level</th>
                    <th>Downloads</th>
                    <th>Last Used</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>`

	// Download accounts
	for _, da := range downloadAccounts {
		status := "Active"
		if !da.IsActive {
			status = "Inactive"
		}

		lastUsed := "Never"
		if da.LastUsed > 0 {
			lastUsed = time.Unix(da.LastUsed, 0).Format("2006-01-02 15:04")
		}

		html += fmt.Sprintf(`
                <tr>
                    <td data-label="Name">%s</td>
                    <td data-label="Email">%s</td>
                    <td data-label="Level"><span class="badge badge-download">Download Account</span></td>
                    <td data-label="Downloads">%d</td>
                    <td data-label="Last Used">%s</td>
                    <td data-label="Status">%s</td>
                    <td data-label="Actions" class="action-links">
                        <a href="/admin/download-accounts/edit?id=%d">Edit</a>
                        <a href="#" onclick="toggleDownloadAccount(%d, %t); return false;">%s</a>
                        <a href="#" onclick="deleteDownloadAccount(%d); return false;">Delete</a>
                    </td>
                </tr>`,
			da.Name, da.Email, da.DownloadCount, lastUsed, status,
			da.Id, da.Id, da.IsActive,
			func() string {
				if da.IsActive {
					return "Deactivate"
				}
				return "Activate"
			}(), da.Id)
	}

	html += `
            </tbody>
        </table>

        <!-- Download Account Pagination -->`

	// Calculate pagination info for download accounts
	dlStart := dlFilter.Offset + 1
	dlEnd := dlFilter.Offset + len(downloadAccounts)
	hasPrevDL := dlFilter.Offset > 0
	hasNextDL := dlFilter.Offset+dlFilter.Limit < dlCount

	html += `
        <div class="pagination">
            <div class="pagination-info">
                Showing ` + fmt.Sprintf("%d-%d", dlStart, dlEnd) + ` of ` + fmt.Sprintf("%d", dlCount) + ` download accounts
            </div>
            <div class="pagination-controls">
                <button onclick="changePage(-1, 'dl')" ` + func() string {
		if !hasPrevDL {
			return "disabled"
		}
		return ""
	}() + `>Previous</button>
                <button onclick="changePage(1, 'dl')" ` + func() string {
		if !hasNextDL {
			return "disabled"
		}
		return ""
	}() + `>Next</button>
            </div>
        </div>
    </div>

    <script>
        function changePage(direction, type) {
            const url = new URL(window.location.href);
            const params = new URLSearchParams(url.search);

            if (type === 'user') {
                const currentOffset = parseInt(params.get('user_offset') || '0');
                const limit = parseInt(params.get('user_limit') || '50');
                const newOffset = Math.max(0, currentOffset + (direction * limit));
                params.set('user_offset', newOffset);
            } else if (type === 'dl') {
                const currentOffset = parseInt(params.get('dl_offset') || '0');
                const limit = parseInt(params.get('dl_limit') || '50');
                const newOffset = Math.max(0, currentOffset + (direction * limit));
                params.set('dl_offset', newOffset);
            }

            window.location.href = '/admin/users?' + params.toString();
        }

        function clearDownloadFilters() {
            const url = new URL(window.location.href);
            const params = new URLSearchParams(url.search);

            // Remove download filter params
            params.delete('dl_search');
            params.delete('dl_active');
            params.delete('dl_offset');

            window.location.href = '/admin/users?' + params.toString();
        }

        async function deleteUser(id) {
            if (!confirm('Are you sure you want to delete this user?\n\nIf you choose yes, the account will be deleted and all the user\'s uploaded files will be available in the trash for 5 days if not deleted manually.')) return;

            try {
                const response = await fetch('/admin/users/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: 'id=' + id
                });

                if (response.ok) {
                    window.location.reload();
                } else {
                    const result = await response.json();
                    alert('Delete failed: ' + (result.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Error deleting user: ' + error.message);
            }
        }

        function toggleDownloadAccount(id, isActive) {
            const action = isActive ? 'deactivate' : 'activate';
            if (!confirm('Are you sure you want to ' + action + ' this download account?')) return;

            fetch('/admin/download-accounts/toggle', {
                method: 'POST',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                body: 'id=' + id
            })
            .then(() => window.location.reload())
            .catch(err => alert('Error toggling download account'));
        }

        function deleteDownloadAccount(id) {
            if (!confirm('Are you sure you want to soft delete this download account? The account will be marked as deleted and fully removed after 90 days.')) return;

            fetch('/admin/download-accounts/delete', {
                method: 'POST',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                body: 'id=' + id
            })
            .then(() => window.location.reload())
            .catch(err => alert('Error deleting download account'));
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
    <meta name="author" content="Ulf Holmström">
    <title>` + title + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 600px; margin: 40px auto; padding: 20px; background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h2 { margin-bottom: 24px; color: #333; }
        input, select { width: 100%; padding: 8px; margin: 8px 0; }
        button { padding: 10px 20px; background: ` + s.getPrimaryColor() + `; color: white; border: none; cursor: pointer; border-radius: 6px; }
        .error { background: #fee; padding: 10px; margin: 10px 0; border-radius: 4px; color: #c33; }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <h2>` + title + `</h2>`

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
    <form method="POST" action="` + action + `" onsubmit="return validatePasswords()">
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
        <div style="position: relative;">
            <input type="password" id="password" name="password"` + func() string {
		if !isEdit {
			return " required"
		}
		return ""
	}() + `>
            <button type="button" onclick="togglePassword('password')" style="position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: transparent; border: none; cursor: pointer; font-size: 20px; padding: 0; width: 30px; height: 30px;">👁️</button>
        </div>` + func() string {
		if !isEdit {
			return `

        <label>Confirm Password:</label>
        <div style="position: relative;">
            <input type="password" id="password_confirm" required>
            <button type="button" onclick="togglePassword('password_confirm')" style="position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: transparent; border: none; cursor: pointer; font-size: 20px; padding: 0; width: 30px; height: 30px;">👁️</button>
        </div>
        <div id="password-error" style="color: #c33; font-size: 14px; margin-top: 4px; display: none;">Passwords do not match</div>`
		}
		return ""
	}() + `

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
        <label style="display: flex; align-items: center; cursor: pointer;">
            <input type="checkbox" name="is_active" value="1"` + func() string {
		if isEdit && user.IsActive {
			return " checked"
		} else if !isEdit {
			return " checked"
		}
		return ""
	}() + ` style="width: auto; margin-right: 8px;">
            <span>Active (user can log in)</span>
        </label>` + func() string {
		if !isEdit {
			return `

        <br>
        <label style="display: flex; align-items: center; cursor: pointer; background: #e3f2fd; padding: 12px; border-radius: 6px; border-left: 4px solid #2196f3;">
            <input type="checkbox" name="send_welcome_email" value="1" checked style="width: auto; margin-right: 8px;">
            <span style="font-weight: 500;">📧 Send welcome email with password setup link</span>
        </label>
        <div style="font-size: 13px; color: #666; margin-top: 8px; margin-left: 28px;">
            User will receive an email to set their own password. The password entered above will be ignored if checked.
        </div>`
		}
		return ""
	}() + `

        <br><br>
        <button type="submit">Save</button>
        <a href="/admin/users">Cancel</a>
    </form>
    </div>

    <script>
        function togglePassword(fieldId) {
            const field = document.getElementById(fieldId);
            if (field.type === 'password') {
                field.type = 'text';
            } else {
                field.type = 'password';
            }
        }

        function validatePasswords() {` + func() string {
		if !isEdit {
			return `
            const password = document.getElementById('password').value;
            const passwordConfirm = document.getElementById('password_confirm').value;
            const errorDiv = document.getElementById('password-error');

            if (password !== passwordConfirm) {
                errorDiv.style.display = 'block';
                return false;
            }
            errorDiv.style.display = 'none';`
		}
		return ""
	}() + `
            return true;
        }
    </script>
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
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>All Files - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1100px;
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
            color: ` + s.getPrimaryColor() + `;
        }
        /* Card-based file list */
        .file-list {
            display: flex;
            flex-direction: column;
            gap: 16px;
        }
        .file-card {
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border: 2px solid #e0e0e0;
            overflow: hidden;
        }
        .file-card:hover {
            border-color: ` + s.getPrimaryColor() + `;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .file-card-header {
            background: ` + s.getPrimaryColor() + `;
            color: white;
            padding: 12px 16px;
            font-weight: 600;
            font-size: 14px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .file-card-body {
            padding: 16px;
        }
        .file-info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 12px;
            margin-bottom: 12px;
        }
        .file-info-item {
            font-size: 13px;
        }
        .file-info-item label {
            display: block;
            color: #666;
            font-size: 11px;
            text-transform: uppercase;
            margin-bottom: 4px;
        }
        .file-info-item span {
            font-weight: 500;
            color: #333;
        }
        .file-note {
            background: #f8f9fa;
            border-left: 4px solid ` + s.getPrimaryColor() + `;
            padding: 10px 12px;
            margin-top: 12px;
            border-radius: 4px;
            font-size: 13px;
        }
        .file-note strong {
            font-weight: 700;
        }
        .files-section {
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .file-list {
            list-style: none;
            margin: 0;
            padding: 0;
        }
        .file-item {
            padding: 20px 24px;
            border-bottom: 3px solid ` + s.getPrimaryColor() + `;
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
            margin: 0 0 4px 0;
        }
        .file-info p {
            color: #666;
            font-size: 14px;
            margin: 0;
        }
        .file-note {
            background: #f8f9fa;
            border-left: 4px solid ` + s.getPrimaryColor() + `;
            padding: 10px 12px;
            margin-top: 12px;
            border-radius: 4px;
            font-size: 13px;
        }
        .file-note strong {
            font-weight: 700;
        }
        .file-actions {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
        .empty-state {
            padding: 60px 20px;
            text-align: center;
            color: #999;
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
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-right: 10px;
            transition: all 0.3s ease;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .btn-primary {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            color: white;
            border: 2px solid ` + s.getPrimaryColor() + `;
        }
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0, 102, 204, 0.3);
        }
        .btn-secondary {
            background: linear-gradient(135deg, #6c757d 0%, #5a6268 100%);
            color: white;
            border: 2px solid #6c757d;
        }
        .btn-secondary:hover {
            background: linear-gradient(135deg, #5a6268 0%, #4e555b 100%);
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(108, 117, 125, 0.3);
        }
        .btn:active {
            transform: translateY(0);
        }
        .file-name {
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        /* Mobile Navigation Styles */
        .hamburger {
            display: none;
            flex-direction: column;
            cursor: pointer;
            padding: 8px;
            background: none;
            border: none;
            z-index: 1001;
        }
        .hamburger span {
            width: 25px;
            height: 3px;
            background: white;
            margin: 3px 0;
            transition: 0.3s;
            border-radius: 2px;
        }
        .hamburger.active span:nth-child(1) {
            transform: rotate(-45deg) translate(-5px, 6px);
        }
        .hamburger.active span:nth-child(2) {
            opacity: 0;
        }
        .hamburger.active span:nth-child(3) {
            transform: rotate(45deg) translate(-5px, -6px);
        }
        .mobile-nav-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.5);
            z-index: 999;
            opacity: 0;
            transition: opacity 0.3s ease;
        }
        .mobile-nav-overlay.active {
            display: block;
            opacity: 1;
        }

        /* Mobile Responsive Styles */
        @media (max-width: 768px) {
            .container {
                margin: 20px auto;
                padding: 0 12px;
            }

            .stats-bar {
                flex-direction: column;
                gap: 16px;
                padding: 16px;
            }

            .stat-item {
                width: 100%;
                padding: 12px;
                background: #f9f9f9;
                border-radius: 8px;
            }

            /* Card-based table layout for mobile */
            table {
                border-radius: 0;
                box-shadow: none;
                background: transparent;
            }

            thead {
                display: none;
            }

            tbody {
                display: block;
            }

            tr {
                display: block;
                background: white;
                border-radius: 12px;
                margin-bottom: 16px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                padding: 16px;
            }

            tr:hover {
                background: white;
            }

            td {
                display: block;
                text-align: left;
                padding: 12px 0;
                border: none !important;
                position: relative;
                min-height: 35px;
            }

            td::before {
                content: attr(data-label);
                display: block;
                font-weight: 600;
                color: #666;
                margin-bottom: 4px;
                font-size: 13px;
            }

            td:last-child {
                padding-top: 16px;
                margin-top: 12px;
                border-top: 1px solid #e0e0e0 !important;
                text-align: left;
                padding-left: 0;
            }

            td:last-child::before {
                display: none;
            }

            /* Stack action buttons horizontally on mobile with better spacing */
            td:last-child .btn {
                display: inline-block;
                width: auto;
                min-width: 45px;
                margin: 4px 2px;
                text-align: center;
                padding: 10px 14px;
                font-size: 16px;
            }

            .file-name {
                max-width: 100%;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
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

        <div class="files-section">
            <ul class="file-list">`

	if len(files) == 0 {
		html += `
                <li class="empty-state">
                    No files in the system yet.
                </li>`
	}

	for _, f := range files {
		// Get user info
		userName := "Deleted user"
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

		downloadURL := s.getPublicURL() + "/d/" + f.Id

		// Note display
		noteDisplay := ""
		if f.Comment != "" {
			noteDisplay = fmt.Sprintf(`<p class="file-note"><strong>📝 Note:</strong> %s</p>`,
				template.HTMLEscapeString(f.Comment))
		}

		html += fmt.Sprintf(`
                <li class="file-item">
                    <div class="file-info">
                        <h3>📄 %s %s %s</h3>
                        <p>%s • %s • %d downloads • Expires: %s</p>
                        %s
                    </div>
                    <div class="file-actions">
                        <button class="btn btn-secondary" onclick="showDownloadHistory('%s', '%s')">📊 History</button>
                        <button class="btn btn-primary" onclick="copyToClipboard('%s', this)">📋 Copy</button>
                        <button class="btn btn-danger" onclick="deleteFile('%s')">🗑️ Delete</button>
                    </div>
                </li>`,
			f.Name, authBadge, status,
			userName, f.Size, f.DownloadCount, expiryInfo,
			noteDisplay,
			f.Id, f.Name,
			downloadURL,
			f.Id)
	}

	html += `
            </ul>
        </div>
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

        function showDownloadHistory(fileId, fileName) {
            document.getElementById('historyFileName').textContent = fileName;
            document.getElementById('downloadHistoryModal').style.display = 'flex';
            document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">Loading...</p>';

            fetch('/file/downloads?file_id=' + encodeURIComponent(fileId))
                .then(response => response.json())
                .then(data => {
                    if (data.logs && data.logs.length > 0) {
                        let html = '<table style="width: 100%; border-collapse: collapse;">';
                        html += '<thead><tr style="background: #f5f5f5; border-bottom: 2px solid #ddd;">';
                        html += '<th style="padding: 12px; text-align: left;">Date & Time</th>';
                        html += '<th style="padding: 12px; text-align: left;">Downloaded By</th>';
                        html += '<th style="padding: 12px; text-align: left;">IP Address</th>';
                        html += '</tr></thead><tbody>';

                        data.logs.forEach(log => {
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
                        html += '<p style="margin-top: 16px; color: #666; font-size: 14px;">Total downloads: ' + data.logs.length + '</p>';
                        document.getElementById('downloadHistoryContent').innerHTML = html;
                    } else {
                        document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #999;">No downloads yet</p>';
                    }
                })
                .catch(error => {
                    document.getElementById('downloadHistoryContent').innerHTML = '<p style="text-align: center; color: #f44336;">Error loading download history</p>';
                    console.error('Error:', error);
                });
        }

        function closeDownloadHistoryModal() {
            document.getElementById('downloadHistoryModal').style.display = 'none';
        }
    </script>

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
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Branding Settings - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
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
            border-color: ` + s.getPrimaryColor() + `;
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
            background: ` + s.getPrimaryColor() + `;
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
    ` + s.getAdminHeaderHTML("") + `
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
                    <input type="text" name="company_name" value="` + brandingConfig["branding_company_name"] + `" placeholder="WulfVault">
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
	// Strip port from URL for display
	serverURL = stripPortFromURL(serverURL)
	maxFileSizeMB, _ := database.DB.GetConfigValue("max_file_size_mb")
	if maxFileSizeMB == "" {
		maxFileSizeMB = "2000"
	}
	defaultQuotaMB, _ := database.DB.GetConfigValue("default_quota_mb")
	if defaultQuotaMB == "" {
		defaultQuotaMB = "5000"
	}
	trashRetentionDays, _ := database.DB.GetConfigValue("trash_retention_days")
	if trashRetentionDays == "" {
		if s.config.TrashRetentionDays > 0 {
			trashRetentionDays = fmt.Sprintf("%d", s.config.TrashRetentionDays)
		} else {
			trashRetentionDays = "5"
		}
	}
	auditLogRetentionDays, _ := database.DB.GetConfigValue("audit_log_retention_days")
	if auditLogRetentionDays == "" {
		if s.config.AuditLogRetentionDays > 0 {
			auditLogRetentionDays = fmt.Sprintf("%d", s.config.AuditLogRetentionDays)
		} else {
			auditLogRetentionDays = "90"
		}
	}
	auditLogMaxSizeMB, _ := database.DB.GetConfigValue("audit_log_max_size_mb")
	if auditLogMaxSizeMB == "" {
		if s.config.AuditLogMaxSizeMB > 0 {
			auditLogMaxSizeMB = fmt.Sprintf("%d", s.config.AuditLogMaxSizeMB)
		} else {
			auditLogMaxSizeMB = "100"
		}
	}
	port, _ := database.DB.GetConfigValue("port")
	if port == "" {
		port = s.config.Port
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Settings - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
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
            border-color: ` + s.getPrimaryColor() + `;
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
            background: ` + s.getPrimaryColor() + `;
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
    ` + s.getAdminHeaderHTML("") + `
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

	// Build full public URL for display
	fullPublicURL := serverURL + ":" + port

	html += `
            <div style="background: #fff3cd; border: 2px solid #ffc107; border-radius: 8px; padding: 20px; margin-bottom: 30px;">
                <h3 style="color: #856404; margin-bottom: 10px; font-size: 16px;">📋 Current Public URL</h3>
                <p style="color: #856404; font-size: 13px; margin-bottom: 12px;">Share this URL with your users to access the system:</p>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <input type="text" id="publicUrl" value="` + fullPublicURL + `" readonly
                           style="flex: 1; padding: 12px; background: white; border: 2px solid #ffc107; font-family: monospace; font-size: 14px; font-weight: 600; color: #d32f2f;">
                    <button type="button" onclick="copyPublicURL()" class="btn"
                            style="background: #ffc107; color: #856404; font-weight: 700; white-space: nowrap;">
                        📋 COPY URL
                    </button>
                </div>
                <p id="copyStatus" style="color: #28a745; font-size: 12px; margin-top: 8px; display: none;">✓ Copied to clipboard!</p>
            </div>

            <form method="POST" action="/admin/settings">
                <div class="form-group">
                    <label for="server_url">Server URL</label>
                    <input type="url" id="server_url" name="server_url" value="` + serverURL + `" required>
                    <p class="help-text">The public URL where this server is accessible (e.g., https://files.manvarg.se). Do not include the port - it's configured separately below.</p>
                </div>

                <div class="form-group">
                    <label for="port">Server Port</label>
                    <input type="number" id="port" name="port" value="` + port + `" min="1" max="65535" required>
                    <p class="help-text">Port number for the server to listen on. Ports below 1024 require administrator privileges.</p>
                    <p class="help-text" style="color: #ff6b00; font-weight: 600;">⚠️ Changes require server restart to take effect</p>
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

                <div class="form-group">
                    <label for="trash_retention_days">Trash Retention Period (Days)</label>
                    <input type="number" id="trash_retention_days" name="trash_retention_days" value="` + trashRetentionDays + `" min="1" max="365" required>
                    <p class="help-text">Number of days to keep deleted files in trash before permanent deletion</p>
                </div>

                <div class="form-group">
                    <label for="audit_log_retention_days">Audit Log Retention (Days)</label>
                    <input type="number" id="audit_log_retention_days" name="audit_log_retention_days" value="` + auditLogRetentionDays + `" min="1" max="3650" required>
                    <p class="help-text">Number of days to keep audit logs before automatic cleanup (default: 90 days)</p>
                </div>

                <div class="form-group">
                    <label for="audit_log_max_size_mb">Audit Log Max Size (MB)</label>
                    <input type="number" id="audit_log_max_size_mb" name="audit_log_max_size_mb" value="` + auditLogMaxSizeMB + `" min="10" max="10000" required>
                    <p class="help-text">Maximum database size for audit logs before automatic cleanup of oldest entries (default: 100 MB)</p>
                </div>

                <button type="submit" class="btn btn-primary">Save Settings</button>
                <a href="/admin" class="btn" style="background: #e0e0e0; margin-left: 10px;">Cancel</a>
            </form>
        </div>

        <!-- RESTART SERVER BUTTON - DISABLED UNTIL SYSTEMD IS INSTALLED
             To enable: Uncomment this section after installing systemd service
             See README.md section "Server Restart Feature" for details

        <div class="card" style="margin-top: 30px; border: 2px solid #f44336;">
            <h2 style="color: #f44336;">⚙️ Server Management</h2>
            <p style="color: #666; margin-bottom: 20px;">
                Restart the server to apply configuration changes or recover from issues.
            </p>
            <button onclick="confirmReboot()" class="btn" style="background: #f44336; color: white;">
                🔄 Restart Server
            </button>
            <p style="color: #999; font-size: 12px; margin-top: 10px;">
                Note: Requires systemd service to be installed. See DEPLOYMENT.md for setup.
            </p>
        </div>
        -->
    </div>

    <script>
        function copyPublicURL() {
            const urlInput = document.getElementById('publicUrl');
            const statusMsg = document.getElementById('copyStatus');

            // Select the text
            urlInput.select();
            urlInput.setSelectionRange(0, 99999); // For mobile devices

            // Try multiple methods to ensure compatibility
            let success = false;

            // Method 1: Try modern clipboard API first (works on HTTPS)
            if (navigator.clipboard && window.isSecureContext) {
                navigator.clipboard.writeText(urlInput.value)
                    .then(() => {
                        showCopySuccess();
                    })
                    .catch(() => {
                        // Fallback to execCommand
                        fallbackCopy();
                    });
            } else {
                // Method 2: Fallback to execCommand (works on HTTP)
                fallbackCopy();
            }

            function fallbackCopy() {
                try {
                    // Use the older but more compatible execCommand
                    success = document.execCommand('copy');
                    if (success) {
                        showCopySuccess();
                    } else {
                        alert('Failed to copy URL. Please copy manually: ' + urlInput.value);
                    }
                } catch (err) {
                    console.error('Copy failed:', err);
                    alert('Failed to copy URL. Please copy manually: ' + urlInput.value);
                }
            }

            function showCopySuccess() {
                statusMsg.style.display = 'block';
                setTimeout(() => {
                    statusMsg.style.display = 'none';
                }, 2000);
            }
        }

        /* RESTART SERVER FUNCTION - Uncomment when systemd is installed
        function confirmReboot() {
            if (confirm('Are you sure you want to restart the server?\n\nThis will briefly interrupt service. Continue?')) {
                fetch('/admin/reboot', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('Server is restarting... Please wait a moment and refresh the page.');
                        setTimeout(() => window.location.reload(), 3000);
                    })
                    .catch(err => console.error('Reboot error:', err));
            }
        }
        */
    </script>
    <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
        Powered by WulfVault © Ulf Holmström – AGPL-3.0
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
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Trash - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
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
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-right: 8px;
            transition: all 0.3s ease;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            min-width: 150px;
            text-align: center;
        }
        .btn-restore {
            background: linear-gradient(135deg, #4caf50 0%, #45a049 100%);
            color: white;
            border: 2px solid #45a049;
        }
        .btn-restore:hover {
            background: linear-gradient(135deg, #45a049 0%, #3d8b40 100%);
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(76, 175, 80, 0.3);
        }
        .btn-delete {
            background: linear-gradient(135deg, #f44336 0%, #da190b 100%);
            color: white;
            border: 2px solid #da190b;
        }
        .btn-delete:hover {
            background: linear-gradient(135deg, #da190b 0%, #c41408 100%);
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(244, 67, 54, 0.3);
        }

        /* Mobile Responsive Styles */
        @media screen and (max-width: 768px) {
            .container {
                margin: 20px auto !important;
                padding: 0 10px !important;
            }
            .info-box {
                padding: 12px !important;
                font-size: 14px !important;
            }
            /* Hide table headers on mobile */
            table thead {
                display: none;
            }
            /* Make table, tbody, tr, td block elements */
            table, table tbody, table tr, table td {
                display: block;
                width: 100%;
            }
            /* Style each row as a card */
            table tr {
                margin-bottom: 15px;
                border: 1px solid #e0e0e0;
                border-radius: 8px;
                padding: 15px;
                background: white;
            }
            table tr:hover {
                background: white;
                box-shadow: 0 2px 8px rgba(0,0,0,0.15);
            }
            /* Style table cells with labels */
            table td {
                padding: 8px 0 !important;
                border: none !important;
                position: relative;
                padding-left: 50% !important;
                text-align: right;
            }
            table td:before {
                content: attr(data-label);
                position: absolute;
                left: 0;
                width: 45%;
                padding-right: 10px;
                font-weight: 600;
                color: #666;
                text-align: left;
            }
            /* Make action buttons stack vertically */
            table td:last-child {
                text-align: center;
                padding-left: 0 !important;
            }
            table td:last-child:before {
                display: none;
            }
            .btn {
                display: block;
                width: 100%;
                margin: 5px 0 !important;
                padding: 12px 16px !important;
                font-size: 15px !important;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <h2 style="margin-bottom: 20px;">Trash (Deleted Files)</h2>

        <div class="info-box">
            ⚠️ Files in trash will be automatically deleted after ` + fmt.Sprintf("%d", s.config.TrashRetentionDays) + ` days. You can restore or permanently delete them here.
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
		userName := "Deleted user"
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

		// Calculate days left using configured retention period
		deletedAt := time.Unix(f.DeletedAt, 0)
		retentionDays := s.config.TrashRetentionDays
		if retentionDays <= 0 {
			retentionDays = 5 // fallback to 5 days if not configured
		}
		deleteAfter := deletedAt.Add(time.Duration(retentionDays) * 24 * time.Hour)
		daysLeft := int(time.Until(deleteAfter).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		html += fmt.Sprintf(`
                <tr>
                    <td data-label="File Name">📄 %s</td>
                    <td data-label="Owner">%s</td>
                    <td data-label="Size">%s</td>
                    <td data-label="Deleted At">%s</td>
                    <td data-label="Deleted By">%s</td>
                    <td data-label="Days Left">%d days</td>
                    <td data-label="Actions">
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

// stripPortFromURL removes the port from a URL if present
func stripPortFromURL(urlStr string) string {
	if urlStr == "" {
		return urlStr
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// If parsing fails, return as-is
		return urlStr
	}

	// Remove the port by setting Host to Hostname()
	parsedURL.Host = parsedURL.Hostname()

	return parsedURL.String()
}

// updateConfigJSON updates a single field in config.json
func (s *Server) updateConfigJSON(key, value string) error {
	configPath := filepath.Join(s.config.DataDir, "config.json")

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Update field
	config[key] = value

	// Write back with pretty print
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// handleAdminReboot handles server restart request
func (s *Server) handleAdminReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("⚠️  Server restart requested by admin")

	// Send response before shutting down
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Server is restarting..."}`))

	// Flush the response
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Try to restart using systemctl if running as service
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Println("🔄 Attempting graceful server restart...")

		// Try systemctl restart first
		cmd := exec.Command("systemctl", "restart", "wulfvault")
		if err := cmd.Run(); err != nil {
			// If systemctl doesn't work, just exit (process manager will restart)
			log.Println("systemctl not available, exiting for process manager restart...")
			os.Exit(0)
		}
	}()
}

// handleAPIUsersList returns all users for team member selection
func (s *Server) handleAPIUsersList(w http.ResponseWriter, r *http.Request) {
	users, err := database.DB.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Filter out deleted users and format for response
	type UserInfo struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var userList []UserInfo
	for _, user := range users {
		if !user.IsActive {
			continue
		}
		userList = append(userList, UserInfo{
			ID:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   userList,
	})
}
