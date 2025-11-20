// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
)

// handleLogin handles user and download account login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderLoginPage(w, r, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderLoginPage(w, r, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Try to authenticate as any account type (User or DownloadAccount)
	authResult, err := auth.AuthenticateAnyAccount(email, password)
	if err != nil {
		// Log failed login attempt
		database.DB.LogAction(&database.AuditLogEntry{
			UserID:     0,
			UserEmail:  email,
			Action:     "LOGIN_FAILED",
			EntityType: "Session",
			EntityID:   "",
			Details:    fmt.Sprintf("{\"email\":\"%s\",\"success\":false,\"reason\":\"invalid_credentials\"}", email),
			IPAddress:  getClientIP(r),
			UserAgent:  r.UserAgent(),
			Success:    false,
			ErrorMsg:   "Invalid credentials",
		})
		s.renderLoginPage(w, r, "Invalid credentials")
		return
	}

	// Handle based on account type
	if authResult.AccountType == auth.AccountTypeUser {
		// Regular user login
		user := authResult.User

		// Check if 2FA is enabled for this user
		if user.TOTPEnabled {
			// Store user ID in temporary cookie and redirect to 2FA verification
			pendingData := map[string]interface{}{
				"user_id":    user.Id,
				"created_at": time.Now().Unix(),
			}
			pendingJSON, _ := json.Marshal(pendingData)

			http.SetCookie(w, &http.Cookie{
				Name:     "totp_pending",
				Value:    base64.StdEncoding.EncodeToString(pendingJSON),
				Path:     "/",
				Expires:  time.Now().Add(5 * time.Minute),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			http.Redirect(w, r, "/2fa/verify", http.StatusSeeOther)
			return
		}

		// No 2FA, create session directly
		sessionID, err := auth.CreateSession(user.Id)
		if err != nil {
			s.renderLoginPage(w, r, "Failed to create session")
			return
		}

		// Log successful login
		database.DB.LogAction(&database.AuditLogEntry{
			UserID:     int64(user.Id),
			UserEmail:  user.Email,
			Action:     "LOGIN_SUCCESS",
			EntityType: "Session",
			EntityID:   sessionID,
			Details:    fmt.Sprintf("{\"email\":\"%s\",\"success\":true}", user.Email),
			IPAddress:  getClientIP(r),
			UserAgent:  r.UserAgent(),
			Success:    true,
			ErrorMsg:   "",
		})

		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Redirect
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			if user.IsAdmin() {
				redirect = "/admin"
			} else {
				redirect = "/dashboard"
			}
		}

		http.Redirect(w, r, redirect, http.StatusSeeOther)

	} else if authResult.AccountType == auth.AccountTypeDownloadAccount {
		// Download account login
		downloadAccount := authResult.DownloadAccount

		// Create session (using email as session identifier)
		sessionEmail, err := auth.CreateDownloadAccountSession(downloadAccount.Id)
		if err != nil {
			s.renderLoginPage(w, r, "Failed to create session")
			return
		}

		// Log successful download account login
		database.DB.LogAction(&database.AuditLogEntry{
			UserID:     int64(downloadAccount.Id),
			UserEmail:  downloadAccount.Email,
			Action:     "DOWNLOAD_ACCOUNT_LOGIN_SUCCESS",
			EntityType: "DownloadSession",
			EntityID:   fmt.Sprintf("%d", downloadAccount.Id),
			Details:    fmt.Sprintf("{\"email\":\"%s\",\"success\":true,\"account_type\":\"download\"}", downloadAccount.Email),
			IPAddress:  getClientIP(r),
			UserAgent:  r.UserAgent(),
			Success:    true,
			ErrorMsg:   "",
		})

		// Set download account session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "download_session",
			Value:    sessionEmail,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to download user dashboard
		http.Redirect(w, r, "/download/dashboard", http.StatusSeeOther)
	} else {
		s.renderLoginPage(w, r, "Unknown account type")
	}
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Try to get user info before deleting session for audit log
	var userEmail string
	var userID int64

	// Get session cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Try to get user from session before deleting it
		if user, err := auth.GetUserBySession(cookie.Value); err == nil && user != nil {
			userEmail = user.Email
			userID = int64(user.Id)
		}

		// Delete session from database
		auth.DeleteSession(cookie.Value)

		// Log logout
		if userEmail != "" {
			database.DB.LogAction(&database.AuditLogEntry{
				UserID:     userID,
				UserEmail:  userEmail,
				Action:     "LOGOUT",
				EntityType: "Session",
				EntityID:   cookie.Value,
				Details:    fmt.Sprintf("{\"email\":\"%s\"}", userEmail),
				IPAddress:  getClientIP(r),
				UserAgent:  r.UserAgent(),
				Success:    true,
				ErrorMsg:   "",
			})
		}
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// renderLoginPage renders the login page
func (s *Server) renderLoginPage(w http.ResponseWriter, r *http.Request, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Login - ` + s.config.CompanyName + `</title>
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
        .login-container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 400px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo img {
            max-width: 200px;
            max-height: 80px;
            margin-bottom: 10px;
        }
        .logo h1 {
            color: ` + s.getPrimaryColor() + `;
            font-size: 28px;
            margin-bottom: 8px;
        }
        .logo p {
            color: #666;
            font-size: 14px;
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
        input[type="email"], input[type="password"], input[type="text"] {
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
        .error {
            background: #fee;
            border: 1px solid #fcc;
            color: #c33;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .footer {
            text-align: center;
            margin-top: 20px;
            color: #999;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">`

	// Get branding config for logo
	brandingConfig, _ := database.DB.GetBrandingConfig()
	if logoData, ok := brandingConfig["branding_logo"]; ok && logoData != "" {
		html += `
            <img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		html += `
            <h1>` + s.config.CompanyName + `</h1>`
	}

	html += `
            <p>Secure File Sharing</p>
        </div>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
	}

	html += `
        <form method="POST" action="/login">
            <div class="form-group">
                <label for="email">Email or Username</label>
                <input type="text" id="email" name="email" required autofocus>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit" class="btn">Login</button>
        </form>
        <div style="text-align: center; margin-top: 15px;">
            <a href="/forgot-password" style="color: ` + s.getPrimaryColor() + `; text-decoration: none; font-size: 14px;">Forgot Password?</a>
        </div>
        <div class="footer">
            ` + s.config.FooterText + `
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
