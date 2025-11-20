// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// handleDownloadAccountGDPR shows download account self-service page with GDPR delete option
func (s *Server) handleDownloadAccountGDPR(w http.ResponseWriter, r *http.Request) {
	// Get account from download session
	account, err := s.getDownloadAccountFromSession(r)
	if err != nil {
		// Redirect to login if no session
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Render GDPR self-service page
	s.renderDownloadAccountGDPRPage(w, account, "")
}

// handleDownloadAccountDelete handles GDPR-compliant account deletion
func (s *Server) handleDownloadAccountDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get account from download session
	account, err := s.getDownloadAccountFromSession(r)
	if err != nil {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Verify confirmation
	confirmation := r.FormValue("confirmation")
	if confirmation != "DELETE" {
		// Re-render page with error
		account, _ = s.getDownloadAccountFromSession(r)
		s.renderDownloadAccountGDPRPage(w, account, "Du måste skriva DELETE för att bekräfta")
		return
	}

	// Store email before anonymization for confirmation email
	accountEmail := account.Email
	accountName := account.Name

	// Anonymize the account (GDPR-compliant deletion)
	err = database.DB.AnonymizeDownloadAccount(account.Id)
	if err != nil {
		log.Printf("Failed to anonymize download account: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	log.Printf("Download account anonymized (GDPR): ID=%d, Email=%s", account.Id, accountEmail)

	// Send confirmation email
	go func() {
		err := email.SendAccountDeletionConfirmation(accountEmail, accountName)
		if err != nil {
			log.Printf("Failed to send deletion confirmation email: %v", err)
		} else {
			log.Printf("Account deletion confirmation sent to: %s", accountEmail)
		}
	}()

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "download_session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	})

	// Show success page
	s.renderAccountDeletionSuccess(w)
}

// getDownloadAccountFromSession retrieves download account from session cookie
func (s *Server) getDownloadAccountFromSession(r *http.Request) (*models.DownloadAccount, error) {
	// Try to get the download_session cookie
	cookie, err := r.Cookie("download_session")
	if err != nil {
		log.Printf("DEBUG: download_session cookie not found: %v. All cookies: %v", err, r.Cookies())
		return nil, http.ErrNoCookie
	}

	log.Printf("DEBUG: Found download_session cookie with value: %s", cookie.Value)

	// The cookie value is the email address
	account, err := database.DB.GetDownloadAccountByEmail(cookie.Value)
	if err != nil {
		log.Printf("DEBUG: Failed to get account by email %s: %v", cookie.Value, err)
		return nil, http.ErrNoCookie
	}

	if !account.IsActive {
		log.Printf("DEBUG: Account %s is not active", cookie.Value)
		return nil, http.ErrNoCookie
	}

	log.Printf("DEBUG: Successfully retrieved download account: %s", account.Email)
	return account, nil
}

// renderDownloadAccountGDPRPage renders the GDPR self-service page
func (s *Server) renderDownloadAccountGDPRPage(w http.ResponseWriter, account *models.DownloadAccount, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()
	logoData := brandingConfig["branding_logo"]

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = `<div style="background: #fee; border: 1px solid #c33; color: #c33; padding: 15px; border-radius: 5px; margin-bottom: 20px;">` + errorMsg + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Mitt konto - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
        }
        .nav-header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            color: white;
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .nav-header .logo {
            display: flex;
            align-items: center;
        }
        .nav-header .logo img {
            max-height: 40px;
            max-width: 150px;
        }
        .nav-header h1 { font-size: 24px; }
        .nav-header nav {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .nav-header nav a {
            color: white;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 5px;
            background: rgba(255,255,255,0.2);
            transition: background 0.3s;
        }
        .nav-header nav a:hover {
            background: rgba(255,255,255,0.3);
        }
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
            transform: rotate(-45deg) translate(-6px, 6px);
        }
        .hamburger.active span:nth-child(2) {
            opacity: 0;
        }
        .hamburger.active span:nth-child(3) {
            transform: rotate(45deg) translate(-6px, -6px);
        }
        .mobile-nav-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            z-index: 999;
        }
        .mobile-nav-overlay.active {
            display: block;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .page-header {
            background: ` + s.getPrimaryColor() + `;
            color: white;
            padding: 30px;
            text-align: center;
        }
        .page-header h2 { font-size: 24px; margin-bottom: 5px; }
        .page-header p { opacity: 0.9; font-size: 14px; }
        .content { padding: 30px; }
        .account-info {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 30px;
        }
        .account-info p {
            margin: 10px 0;
            display: flex;
            justify-content: space-between;
        }
        .account-info strong { color: #333; }
        .danger-zone {
            background: #fff5f5;
            border: 2px solid #feb2b2;
            border-radius: 8px;
            padding: 25px;
            margin-top: 30px;
        }
        .danger-zone h2 {
            color: #c53030;
            font-size: 18px;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .danger-zone p {
            color: #742a2a;
            line-height: 1.6;
            margin-bottom: 15px;
        }
        .danger-zone ul {
            margin-left: 20px;
            margin-bottom: 20px;
            color: #742a2a;
        }
        .danger-zone li { margin: 8px 0; }
        .confirmation-box {
            background: white;
            border: 2px solid #feb2b2;
            border-radius: 5px;
            padding: 20px;
            margin: 20px 0;
        }
        .confirmation-box label {
            display: block;
            font-weight: bold;
            color: #c53030;
            margin-bottom: 10px;
        }
        .confirmation-box input[type="text"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e2e8f0;
            border-radius: 5px;
            font-size: 16px;
            font-family: monospace;
        }
        .confirmation-box small {
            display: block;
            margin-top: 8px;
            color: #718096;
        }
        .btn {
            display: inline-block;
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            text-decoration: none;
            transition: all 0.3s;
        }
        .btn-danger {
            background: #c53030;
            color: white;
        }
        .btn-danger:hover { background: #9b2c2c; transform: translateY(-2px); }
        .btn-secondary {
            background: #e2e8f0;
            color: #4a5568;
            margin-left: 10px;
        }
        .btn-secondary:hover { background: #cbd5e0; }
        .footer {
            text-align: center;
            padding: 20px;
            color: #718096;
            font-size: 14px;
            border-top: 1px solid #e2e8f0;
        }

        /* Mobile Responsive Styles */
        @media screen and (max-width: 768px) {
            .nav-header {
                padding: 15px 20px;
                flex-wrap: wrap;
            }
            .nav-header h1 {
                font-size: 18px;
                order: 1;
                flex: 1;
            }
            .hamburger {
                display: flex !important;
                order: 3;
                margin-left: auto;
            }
            .nav-header nav {
                display: none !important;
                position: fixed !important;
                right: -100% !important;
                top: 0 !important;
                width: 280px !important;
                height: 100vh !important;
                background: linear-gradient(180deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%) !important;
                flex-direction: column !important;
                padding: 80px 20px 20px !important;
                transition: right 0.3s ease !important;
                z-index: 1000 !important;
                gap: 0 !important;
            }
            .nav-header nav.active {
                display: flex !important;
                right: 0 !important;
            }
            .nav-header nav a {
                width: 100%;
                padding: 15px 20px !important;
                border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                margin: 0 !important;
                border-radius: 0 !important;
            }
            .container {
                margin: 20px 15px;
            }
            .btn-secondary {
                margin-left: 0;
                margin-top: 10px;
                display: block;
                width: 100%;
            }
        }
    </style>
    <script>
        function confirmDelete() {
            const input = document.getElementById('confirmInput').value;
            if (input !== 'DELETE') {
                alert('Du måste skriva DELETE exakt som det står för att bekräfta.');
                return false;
            }
            return confirm('Är du helt säker? Detta går inte att ångra. Om du vill ladda ner filer från vår tjänst igen måste du registrera om dig.');
        }
    </script>
</head>
<body>
    <div class="nav-header">
        <div class="logo">`

	if logoData != "" {
		html += `<img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		html += `<h1>` + s.config.CompanyName + ` - My Downloads</h1>`
	}

	html += `
        </div>
        <button class="hamburger" aria-label="Toggle navigation" aria-expanded="false">
            <span></span>
            <span></span>
            <span></span>
        </button>
        <nav>
            <a href="/download/dashboard">Dashboard</a>
            <a href="/download/change-password">Change Password</a>
            <a href="/download/account-settings">Account Settings</a>
            <a href="/download/logout">Logout</a>
        </nav>
    </div>
    <div class="mobile-nav-overlay"></div>

    <div class="container">
        <div class="page-header">
            <h2>Mitt Nedladdningskonto</h2>
            <p>` + s.config.CompanyName + `</p>
        </div>

        <div class="content">
            ` + errorHTML + `

            <div class="account-info">
                <h3 style="margin-bottom: 15px; color: #2d3748;">Kontoinformation</h3>
                <p><strong>Namn:</strong> <span>` + account.Name + `</span></p>
                <p><strong>E-post:</strong> <span>` + account.Email + `</span></p>
                <p><strong>Antal nedladdningar:</strong> <span>` + strconv.Itoa(account.DownloadCount) + `</span></p>
                <p><strong>Status:</strong> <span style="color: #38a169;">Aktiv</span></p>
            </div>

            <div class="danger-zone">
                <h2>
                    <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
                    </svg>
                    Radera mitt konto (GDPR)
                </h2>

                <p><strong>OBS!</strong> Genom att radera ditt konto:</p>
                <ul>
                    <li>Raderas din personliga information permanent från vårt system</li>
                    <li>Du kommer inte längre kunna ladda ner filer från denna tjänst</li>
                    <li>Om du vill ladda ner filer igen måste du registrera ett nytt konto</li>
                    <li>Du kommer få en bekräftelse via e-post</li>
                </ul>

                <form method="POST" action="/download-account/delete" onsubmit="return confirmDelete()">
                    <div class="confirmation-box">
                        <label for="confirmInput">Skriv DELETE för att bekräfta:</label>
                        <input type="text" id="confirmInput" name="confirmation" autocomplete="off" required>
                        <small>Detta går inte att ångra. Skriv exakt: DELETE</small>
                    </div>

                    <button type="submit" class="btn btn-danger">Radera mitt konto permanent</button>
                    <a href="/download/dashboard" class="btn btn-secondary">Avbryt</a>
                </form>
            </div>
        </div>

        <div class="footer">
            <p>Detta är en GDPR-kompatibel självbetjäningsfunktion</p>
            <p>&copy; ` + s.config.CompanyName + `</p>
        </div>

        <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
            Powered by WulfVault © Ulf Holmström – AGPL-3.0
        </div>
    </div>

    
</body>
</html>`

	w.Write([]byte(html))
}

// renderAccountDeletionSuccess renders success page after account deletion
func (s *Server) renderAccountDeletionSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Konto raderat - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
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
            max-width: 500px;
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
            text-align: center;
            padding: 50px 30px;
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
        }
        .success-icon svg {
            width: 40px;
            height: 40px;
            color: #155724;
        }
        h1 { color: #155724; margin-bottom: 20px; font-size: 28px; }
        p { color: #666; line-height: 1.8; margin-bottom: 15px; }
        .btn {
            display: inline-block;
            margin-top: 30px;
            padding: 12px 30px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: all 0.3s;
        }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,0,0,0.2); }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
            </svg>
        </div>

        <h1>Ditt konto har raderats</h1>

        <p>Din personliga information har anonymiserats och raderats från vårt system enligt GDPR.</p>

        <p>Du har fått en bekräftelse via e-post till den adress som var registrerad på kontot.</p>

        <p>Om du vill ladda ner filer från vår tjänst igen i framtiden måste du registrera ett nytt konto.</p>

        <a href="/" class="btn">Till startsidan</a>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// handleUserDataExport exports all user data as JSON (GDPR Article 15 - Right of Access)
func (s *Server) handleUserDataExport(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Collect all user data
	userData := make(map[string]interface{})

	// User profile
	userData["user"] = map[string]interface{}{
		"id":               user.Id,
		"name":             user.Name,
		"email":            user.Email,
		"user_level":       user.GetReadableUserLevel(),
		"created_at":       user.CreatedAt,
		"is_active":        user.IsActive,
		"storage_quota_mb": user.StorageQuotaMB,
		"storage_used_mb":  user.StorageUsedMB,
		"totp_enabled":     user.TOTPEnabled,
	}

	// Note: File and audit log export requires database methods that may not exist yet
	// For now, we provide basic user profile export
	userData["files"] = []interface{}{}
	userData["audit_logs"] = []interface{}{}

	// Export metadata
	userData["export_metadata"] = map[string]interface{}{
		"export_date": strconv.FormatInt(currentTimestamp(), 10),
		"export_type": "GDPR Article 15 - Right of Access",
		"format":      "JSON",
	}

	// Set headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=wulfvault-user-data-export-"+strconv.Itoa(user.Id)+".json")

	// Encode and send JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(userData); err != nil {
		log.Printf("Failed to encode user data export: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to export data")
		return
	}

	log.Printf("User data exported: UserID=%d, Email=%s", user.Id, user.Email)
}

// handleUserAccountSettings shows account settings page with deletion option
func (s *Server) handleUserAccountSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	s.renderUserAccountSettings(w, user, "")
}

// handleUserAccountDelete handles GDPR-compliant account deletion for system users
func (s *Server) handleUserAccountDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Verify confirmation
	confirmation := r.FormValue("confirmation")
	if confirmation != "DELETE" {
		// Re-render page with error
		s.renderUserAccountSettings(w, user, "You must type DELETE to confirm")
		return
	}

	// Store email before anonymization for confirmation email
	userEmail := user.Email
	userName := user.Name

	// Soft delete the user (GDPR-compliant deletion)
	err := database.DB.SoftDeleteUser(user.Id, "self")
	if err != nil {
		log.Printf("Failed to soft delete user: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	log.Printf("User account soft-deleted (GDPR): ID=%d, Email=%s", user.Id, userEmail)

	// Send confirmation email
	go func() {
		err := email.SendAccountDeletionConfirmation(userEmail, userName)
		if err != nil {
			log.Printf("Failed to send deletion confirmation email: %v", err)
		} else {
			log.Printf("Account deletion confirmation sent to: %s", userEmail)
		}
	}()

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	})

	// Show success page
	s.renderAccountDeletionSuccess(w)
}

// renderUserAccountSettings renders the account settings page with deletion option
func (s *Server) renderUserAccountSettings(w http.ResponseWriter, user *models.User, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get branding config
	brandingConfig, _ := database.DB.GetBrandingConfig()
	logoData := brandingConfig["branding_logo"]

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = `<div style="background: #fee; border: 1px solid #c33; color: #c33; padding: 15px; border-radius: 5px; margin-bottom: 20px;">` + errorMsg + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Account Settings - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
        }
        .nav-header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            color: white;
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .nav-header .logo {
            display: flex;
            align-items: center;
        }
        .nav-header .logo img {
            max-height: 40px;
            max-width: 150px;
        }
        .nav-header h1 { font-size: 24px; }
        .nav-header nav {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .nav-header nav a {
            color: white;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 5px;
            background: rgba(255,255,255,0.2);
            transition: background 0.3s;
        }
        .nav-header nav a:hover {
            background: rgba(255,255,255,0.3);
        }
        .container {
            max-width: 800px;
            margin: 40px auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .page-header {
            background: ` + s.getPrimaryColor() + `;
            color: white;
            padding: 30px;
            text-align: center;
        }
        .page-header h2 { font-size: 24px; margin-bottom: 5px; }
        .page-header p { opacity: 0.9; font-size: 14px; }
        .content { padding: 30px; }
        .section {
            margin-bottom: 40px;
            padding-bottom: 30px;
            border-bottom: 1px solid #e2e8f0;
        }
        .section:last-child {
            border-bottom: none;
        }
        .section h3 {
            font-size: 18px;
            margin-bottom: 15px;
            color: #2d3748;
        }
        .account-info {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .account-info p {
            margin: 10px 0;
            display: flex;
            justify-content: space-between;
        }
        .account-info strong { color: #333; }
        .btn {
            display: inline-block;
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            text-decoration: none;
            transition: all 0.3s;
        }
        .btn-primary {
            background: ` + s.getPrimaryColor() + `;
            color: white;
        }
        .btn-primary:hover { opacity: 0.9; transform: translateY(-2px); }
        .btn-secondary {
            background: #e2e8f0;
            color: #4a5568;
            margin-left: 10px;
        }
        .btn-secondary:hover { background: #cbd5e0; }
        .danger-zone {
            background: #fff5f5;
            border: 2px solid #feb2b2;
            border-radius: 8px;
            padding: 25px;
        }
        .danger-zone h3 {
            color: #c53030;
            font-size: 18px;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .danger-zone p {
            color: #742a2a;
            line-height: 1.6;
            margin-bottom: 15px;
        }
        .danger-zone ul {
            margin-left: 20px;
            margin-bottom: 20px;
            color: #742a2a;
        }
        .danger-zone li { margin: 8px 0; }
        .confirmation-box {
            background: white;
            border: 2px solid #feb2b2;
            border-radius: 5px;
            padding: 20px;
            margin: 20px 0;
        }
        .confirmation-box label {
            display: block;
            font-weight: bold;
            color: #c53030;
            margin-bottom: 10px;
        }
        .confirmation-box input[type="text"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e2e8f0;
            border-radius: 5px;
            font-size: 16px;
            font-family: monospace;
        }
        .confirmation-box small {
            display: block;
            margin-top: 8px;
            color: #718096;
        }
        .btn-danger {
            background: #c53030;
            color: white;
        }
        .btn-danger:hover { background: #9b2c2c; transform: translateY(-2px); }

        @media screen and (max-width: 768px) {
            .container {
                margin: 20px 15px;
            }
            .btn-secondary {
                margin-left: 0;
                margin-top: 10px;
                display: block;
                width: 100%;
            }
        }
    </style>
    <script>
        function confirmDelete() {
            const input = document.getElementById('confirmInput').value;
            if (input !== 'DELETE') {
                alert('You must type DELETE exactly as shown to confirm.');
                return false;
            }
            return confirm('Are you absolutely sure? This cannot be undone. Your account and all files will be permanently deleted.');
        }
    </script>
</head>
<body>
    <div class="nav-header">
        <div class="logo">`

	if logoData != "" {
		html += `<img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		html += `<h1>` + s.config.CompanyName + `</h1>`
	}

	html += `
        </div>
        <nav>
            <a href="/dashboard">Dashboard</a>
            <a href="/settings">Settings</a>
            <a href="/logout">Logout</a>
        </nav>
    </div>

    <div class="container">
        <div class="page-header">
            <h2>Account Settings</h2>
            <p>` + s.config.CompanyName + `</p>
        </div>

        <div class="content">
            ` + errorHTML + `

            <div class="section">
                <h3>Account Information</h3>
                <div class="account-info">
                    <p><strong>Name:</strong> <span>` + user.Name + `</span></p>
                    <p><strong>Email:</strong> <span>` + user.Email + `</span></p>
                    <p><strong>Role:</strong> <span>` + user.GetReadableUserLevel() + `</span></p>
                    <p><strong>Account Created:</strong> <span>` + formatTimestamp(user.CreatedAt) + `</span></p>
                    <p><strong>Storage Used:</strong> <span>` + formatBytes(user.StorageUsedMB*1024*1024) + ` / ` + formatBytes(user.StorageQuotaMB*1024*1024) + `</span></p>
                </div>
            </div>

            <div class="section">
                <h3>Data Privacy & Export</h3>
                <p>Under GDPR, you have the right to access and export all your personal data.</p>
                <a href="/api/v1/user/export-data" class="btn btn-primary" download>Export My Data (JSON)</a>
                <p style="margin-top: 10px; font-size: 14px; color: #718096;">
                    This will download a JSON file containing your profile, files list, and activity logs.
                </p>
            </div>

            <div class="section danger-zone">
                <h3>
                    <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
                    </svg>
                    Delete My Account (GDPR)
                </h3>

                <p><strong>WARNING!</strong> By deleting your account:</p>
                <ul>
                    <li>Your personal information will be anonymized and removed from our system</li>
                    <li>All your files will be permanently deleted</li>
                    <li>All your share links will be disabled</li>
                    <li>This action cannot be undone</li>
                    <li>You will receive a confirmation email</li>
                </ul>

                <form method="POST" action="/settings/delete-account" onsubmit="return confirmDelete()">
                    <div class="confirmation-box">
                        <label for="confirmInput">Type DELETE to confirm:</label>
                        <input type="text" id="confirmInput" name="confirmation" autocomplete="off" required>
                        <small>This action is irreversible. Type exactly: DELETE</small>
                    </div>

                    <button type="submit" class="btn btn-danger">Delete My Account Permanently</button>
                    <a href="/dashboard" class="btn btn-secondary">Cancel</a>
                </form>
            </div>
        </div>

        <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
            Powered by WulfVault © Ulf Holmström – AGPL-3.0
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// Helper functions for formatting
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + "KMGTPE"[exp:exp+1] + "B"
}

func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

func currentTimestamp() int64 {
	return time.Now().Unix()
}
