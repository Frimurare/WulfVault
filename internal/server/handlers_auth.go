// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
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
	rememberMe := r.FormValue("remember_me") == "on"

	// Log login attempt start (for debugging double-submit issues)
	log.Printf("🔐 Login attempt: %s | IP: %s | RememberMe: %v", email, getClientIP(r), rememberMe)

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
			// Store user ID and remember_me preference in temporary cookie
			pendingData := map[string]interface{}{
				"user_id":     user.Id,
				"created_at":  time.Now().Unix(),
				"remember_me": rememberMe,
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

		// No 2FA, create session directly with appropriate duration
		sessionDuration := 24 * time.Hour
		if rememberMe {
			sessionDuration = 30 * 24 * time.Hour // 30 days
		}

		sessionID, err := auth.CreateSession(user.Id, sessionDuration)
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

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			Expires:  time.Now().Add(sessionDuration),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false,
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

		// Set download account session cookie with appropriate expiration
		sessionDuration := 24 * time.Hour
		if rememberMe {
			sessionDuration = 30 * 24 * time.Hour // 30 days
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "download_session",
			Value:    accountSessionValue(sessionEmail, rememberMe),
			Path:     "/",
			Expires:  time.Now().Add(sessionDuration),
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
	inviteEmail := template.HTMLEscapeString(r.URL.Query().Get("invite"))
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
        .btn:hover:not(:disabled) {
            opacity: 0.9;
        }
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
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
    <script>
        // Prevent double form submission and provide feedback
        document.addEventListener('DOMContentLoaded', function() {
            const loginForm = document.querySelector('form');
            const submitBtn = document.querySelector('.btn');
            let isSubmitting = false;

            if (loginForm && submitBtn) {
                loginForm.addEventListener('submit', function(e) {
                    // Prevent double-submit
                    if (isSubmitting) {
                        e.preventDefault();
                        return false;
                    }

                    // Validate form fields
                    const email = document.getElementById('email').value.trim();
                    const password = document.getElementById('password').value;

                    if (!email || !password) {
                        return true; // Let browser validation handle it
                    }

                    // Mark as submitting
                    isSubmitting = true;
                    submitBtn.disabled = true;
                    submitBtn.textContent = 'Logging in...';
                    submitBtn.style.opacity = '0.7';

                    // Safety timeout to re-enable after 5 seconds if no response
                    setTimeout(function() {
                        if (isSubmitting) {
                            isSubmitting = false;
                            submitBtn.disabled = false;
                            submitBtn.textContent = 'Login';
                            submitBtn.style.opacity = '1';
                        }
                    }, 5000);
                });
            }
        });
    </script>
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

	if inviteEmail != "" {
		// Use the configured provider display name in the banner so the
		// invitee sees the right brand ("Sign in with Google" etc).
		bannerProvider := "your SSO provider"
		if ssoCfg, err := auth.LoadIdentityProviderConfig(database.DB); err == nil && ssoCfg.ShouldShowSSOButton() {
			bannerProvider = ssoCfg.EffectiveDisplayName()
		}
		html += `<div style="background: #e8f4f8; border: 1px solid #b3e0ed; color: #0c5460; padding: 12px; border-radius: 6px; margin-bottom: 20px; font-size: 14px;">
            <strong>You've been invited.</strong> Sign in with ` + template.HTMLEscapeString(bannerProvider) + ` below to activate your account.
        </div>`
	}

	emailValueAttr := ""
	emailAutofocus := " autofocus"
	if inviteEmail != "" {
		emailValueAttr = ` value="` + inviteEmail + `"`
		emailAutofocus = "" // focus the Microsoft button area instead, since password is irrelevant for invited users
	}

	html += `
        <form method="POST" action="/login">
            <div class="form-group">
                <label for="email">Email or Username</label>
                <input type="text" id="email" name="email"` + emailValueAttr + ` required` + emailAutofocus + `>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <div class="form-group" style="display: flex; align-items: center; margin-bottom: 20px;">
                <input type="checkbox" id="remember_me" name="remember_me" style="width: auto; margin-right: 8px;">
                <label for="remember_me" style="margin: 0; font-weight: normal; cursor: pointer;">Keep me logged in (30 days)</label>
            </div>
            <button type="submit" class="btn">Login</button>
        </form>`

	// Render the SSO button only if a provider is enabled and not overridden
	// by the local-only escape hatch. Logo + label adapt to the provider type:
	// the Microsoft 4-tile for Entra, a neutral lock icon for Generic OIDC.
	if ssoCfg, err := auth.LoadIdentityProviderConfig(database.DB); err == nil && ssoCfg.ShouldShowSSOButton() {
		providerName := template.HTMLEscapeString(ssoCfg.EffectiveDisplayName())
		var providerIcon string
		if ssoCfg.IsEntra() {
			providerIcon = `<svg width="18" height="18" viewBox="0 0 23 23" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                <rect x="1"  y="1"  width="10" height="10" fill="#F35325"/>
                <rect x="12" y="1"  width="10" height="10" fill="#81BC06"/>
                <rect x="1"  y="12" width="10" height="10" fill="#05A6F0"/>
                <rect x="12" y="12" width="10" height="10" fill="#FFBA08"/>
            </svg>`
		} else {
			providerIcon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>`
		}
		html += `
        <div style="display:flex; align-items:center; margin: 20px 0 16px; color:#999; font-size:12px;">
            <div style="flex:1; height:1px; background:#e5e7eb;"></div>
            <div style="padding: 0 12px;">OR</div>
            <div style="flex:1; height:1px; background:#e5e7eb;"></div>
        </div>
        <a href="/auth/oidc/login" style="display:flex; align-items:center; justify-content:center; gap:10px; width:100%; padding:12px; background:white; color:#1f1f1f; border:1px solid #d2d2d2; border-radius:6px; font-size:14px; font-weight:600; text-decoration:none; transition: background 0.2s;" onmouseover="this.style.background='#f5f5f5'" onmouseout="this.style.background='white'">
            ` + providerIcon + `
            Sign in with ` + providerName + `
        </a>`
	}

	html += `
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
