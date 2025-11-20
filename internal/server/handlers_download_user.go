// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// requireDownloadAuth is middleware that requires download account authentication
func (s *Server) requireDownloadAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("download_session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		account, err := s.getDownloadAccountFromSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check for inactivity timeout (10 minutes), but only if no active transfer
		// Use email as session identifier for download accounts
		if !s.hasActiveTransfer(cookie.Value) {
			timeSinceLastActivity := time.Since(time.Unix(account.LastUsed, 0))
			if timeSinceLastActivity > auth.InactivityTimeout {
				// Force logout due to inactivity
				http.SetCookie(w, &http.Cookie{
					Name:     "download_session",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Redirect(w, r, "/login?timeout=1", http.StatusSeeOther)
				return
			}
		}

		// Store account in context
		r = r.WithContext(contextWithDownloadAccount(r.Context(), account))
		next(w, r)
	}
}

// handleDownloadDashboard shows the download account dashboard
func (s *Server) handleDownloadDashboard(w http.ResponseWriter, r *http.Request) {
	account, ok := downloadAccountFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get download history
	downloadLogs, err := database.DB.GetDownloadLogsByAccountID(account.Id)
	if err != nil {
		log.Printf("Error fetching download logs: %v", err)
		downloadLogs = []*models.DownloadLog{}
	}

	s.renderDownloadDashboard(w, account, downloadLogs)
}

// handleDownloadChangePassword allows download users to change their password
func (s *Server) handleDownloadChangePassword(w http.ResponseWriter, r *http.Request) {
	account, ok := downloadAccountFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		s.renderDownloadChangePasswordPage(w, account, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderDownloadChangePasswordPage(w, account, "Invalid form data")
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate current password
	if !auth.CheckPasswordHash(currentPassword, account.Password) {
		s.renderDownloadChangePasswordPage(w, account, "Current password is incorrect")
		return
	}

	// Validate new password
	if newPassword == "" || len(newPassword) < 6 {
		s.renderDownloadChangePasswordPage(w, account, "New password must be at least 6 characters")
		return
	}

	if newPassword != confirmPassword {
		s.renderDownloadChangePasswordPage(w, account, "Passwords do not match")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		s.renderDownloadChangePasswordPage(w, account, "Failed to hash password")
		return
	}

	// Update password
	account.Password = hashedPassword
	if err := database.DB.UpdateDownloadAccount(account); err != nil {
		s.renderDownloadChangePasswordPage(w, account, "Failed to update password")
		return
	}

	log.Printf("Password changed for download account: %s", account.Email)

	// Redirect back to dashboard with success message
	s.renderDownloadChangePasswordPage(w, account, "SUCCESS:Password changed successfully!")
}

// handleDownloadLogout logs out a download user
func (s *Server) handleDownloadLogout(w http.ResponseWriter, r *http.Request) {
	// Clear download session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "download_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleDownloadAccountDelete handles GDPR self-service deletion
func (s *Server) handleDownloadAccountDeleteSelf(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	account, ok := downloadAccountFromContext(r.Context())
	if !ok {
		s.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Verify confirmation
	confirmation := r.FormValue("confirmation")
	if confirmation != "DELETE" {
		s.sendError(w, http.StatusBadRequest, "Confirmation required")
		return
	}

	// Soft delete the account
	err := database.DB.SoftDeleteDownloadAccount(account.Id, "user")
	if err != nil {
		log.Printf("Failed to soft delete download account: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	log.Printf("Download account soft deleted (GDPR): ID=%d, Email=%s", account.Id, account.Email)

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(account.Id),
		UserEmail:  account.Email,
		Action:     "DOWNLOAD_ACCOUNT_DELETED",
		EntityType: "DownloadAccount",
		EntityID:   fmt.Sprintf("%d", account.Id),
		Details:    fmt.Sprintf("{\"email\":\"%s\",\"name\":\"%s\",\"soft_delete\":true,\"self_deleted\":true}", account.Email, account.Name),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	// Clear session
	http.SetCookie(w, &http.Cookie{
		Name:     "download_session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect to success page
	http.Redirect(w, r, "/download/deleted-success", http.StatusSeeOther)
}

// renderDownloadDashboard renders the download user dashboard
func (s *Server) renderDownloadDashboard(w http.ResponseWriter, account *models.DownloadAccount, downloadLogs []*models.DownloadLog) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>My Downloads - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .account-info {
            background: white;
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        .account-info h2 { color: #333; margin-bottom: 20px; }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
        }
        .info-item { padding: 15px; background: #f8f9fa; border-radius: 8px; }
        .info-item strong { display: block; color: #666; font-size: 12px; margin-bottom: 5px; }
        .info-item span { font-size: 18px; color: #333; }
        table {
            width: 100%;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td { padding: 16px; text-align: left; }
        th {
            background: #f9f9f9;
            font-weight: 600;
            color: #666;
            font-size: 14px;
        }
        tr:not(:last-child) td { border-bottom: 1px solid #e0e0e0; }
        tr:hover { background: #f9f9f9; }
        .btn {
            padding: 10px 20px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: 500;
            transition: all 0.3s;
            display: inline-block;
        }
        .btn-primary {
            background: ` + s.getPrimaryColor() + `;
            color: white;
        }
        .btn-primary:hover { opacity: 0.9; }

        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .info-grid {
                grid-template-columns: 1fr !important;
            }
            table thead {
                display: none;
            }
            table, table tbody, table tr {
                display: block;
                width: 100%;
            }
            table tr {
                margin-bottom: 15px;
                border: 1px solid #e0e0e0;
                border-radius: 8px;
                padding: 15px;
                background: white;
            }
            table td {
                display: block;
                text-align: right;
                padding: 8px 0;
                border: none;
                position: relative;
                padding-left: 50%;
            }
            table td::before {
                content: attr(data-label);
                position: absolute;
                left: 0;
                width: 45%;
                padding-right: 10px;
                text-align: left;
                font-weight: 600;
                color: #666;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `

    <div class="container">
        <div class="account-info">
            <h2>Account Information</h2>
            <div class="info-grid">
                <div class="info-item">
                    <strong>NAME</strong>
                    <span>` + account.Name + `</span>
                </div>
                <div class="info-item">
                    <strong>EMAIL</strong>
                    <span>` + account.Email + `</span>
                </div>
                <div class="info-item">
                    <strong>DOWNLOADS</strong>
                    <span>` + strconv.Itoa(account.DownloadCount) + `</span>
                </div>
                <div class="info-item">
                    <strong>LAST USED</strong>
                    <span>` + account.GetLastUsedDate() + `</span>
                </div>
            </div>
        </div>

        <h2 style="margin-bottom: 20px; color: #333;">Download History</h2>
        <table>
            <thead>
                <tr>
                    <th>File Name</th>
                    <th>Downloaded At</th>
                    <th>Size</th>
                </tr>
            </thead>
            <tbody>`

	if len(downloadLogs) == 0 {
		html += `
                <tr>
                    <td colspan="3" style="text-align: center; padding: 40px; color: #999;">
                        No downloads yet
                    </td>
                </tr>`
	} else {
		for _, log := range downloadLogs {
			sizeStr := fmt.Sprintf("%.2f MB", float64(log.FileSize)/(1024*1024))
			html += fmt.Sprintf(`
                <tr>
                    <td data-label="File Name">%s</td>
                    <td data-label="Downloaded At">%s</td>
                    <td data-label="Size">%s</td>
                </tr>`, log.FileName, log.GetReadableDownloadDate(), sizeStr)
		}
	}

	html += `
            </tbody>
        </table>
    </div>

    <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
        Powered by WulfVault © Ulf Holmström – AGPL-3.0
    </div>

    
</body>
</html>`

	w.Write([]byte(html))
}

// renderDownloadChangePasswordPage renders the password change page
func (s *Server) renderDownloadChangePasswordPage(w http.ResponseWriter, account *models.DownloadAccount, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	messageHTML := ""
	if message != "" {
		if len(message) > 8 && message[:8] == "SUCCESS:" {
			messageHTML = `<div style="background: #d4edda; border: 1px solid #c3e6cb; color: #155724; padding: 15px; border-radius: 5px; margin-bottom: 20px;">` + message[8:] + `</div>`
		} else {
			messageHTML = `<div style="background: #fee; border: 1px solid #c33; color: #c33; padding: 15px; border-radius: 5px; margin-bottom: 20px;">` + message + `</div>`
		}
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Change Password - ` + s.config.CompanyName + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .form-group { margin-bottom: 20px; }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
        }
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
        }
        input:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            cursor: pointer;
            font-weight: 600;
            width: 100%;
        }
        .btn-primary {
            background: ` + s.getPrimaryColor() + `;
            color: white;
        }
        .btn-primary:hover { opacity: 0.9; }

        /* Mobile Responsive Styles */
        @media screen and (max-width: 768px) {
            .container {
                margin: 20px auto;
                padding: 0 15px;
            }
            .card {
                padding: 25px 20px;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `

    <div class="container">
        <div class="card">
            ` + messageHTML + `
            <form method="POST" action="/download/change-password">
                <div class="form-group">
                    <label>Current Password</label>
                    <input type="password" name="current_password" required>
                </div>
                <div class="form-group">
                    <label>New Password</label>
                    <input type="password" name="new_password" required minlength="6">
                </div>
                <div class="form-group">
                    <label>Confirm New Password</label>
                    <input type="password" name="confirm_password" required minlength="6">
                </div>
                <button type="submit" class="btn btn-primary">Change Password</button>
            </form>
        </div>
    </div>

    <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
        Powered by WulfVault © Ulf Holmström – AGPL-3.0
    </div>

    
</body>
</html>`

	w.Write([]byte(html))
}

// handleDownloadAccountSettings shows account settings with GDPR delete option
func (s *Server) handleDownloadAccountSettings(w http.ResponseWriter, r *http.Request) {
	account, ok := downloadAccountFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	s.renderDownloadAccountGDPRPage(w, account, "")
}

// handleDownloadDeletedSuccess shows success page after account deletion
func (s *Server) handleDownloadDeletedSuccess(w http.ResponseWriter, r *http.Request) {
	s.renderAccountDeletionSuccess(w)
}
