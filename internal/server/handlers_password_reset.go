// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"log"
	"net/http"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
)

// handleForgotPassword shows the forgot password page or handles the request
func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderForgotPasswordPage(w, r, "", false)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderForgotPasswordPage(w, r, "forgot.error.invalid_form", false)
		return
	}

	emailAddress := r.FormValue("email")
	if emailAddress == "" {
		s.renderForgotPasswordPage(w, r, "forgot.error.email_required", false)
		return
	}

	// Determine account type by checking both tables
	var accountType string
	var accountExists bool

	// Check regular users
	user, err := database.DB.GetUserByEmail(emailAddress)
	if err == nil && user.IsActive {
		accountType = database.AccountTypeUser
		accountExists = true
	}

	// Check download accounts if not found as regular user
	if !accountExists {
		downloadAccount, err := database.DB.GetDownloadAccountByEmail(emailAddress)
		if err == nil && downloadAccount.IsActive {
			accountType = database.AccountTypeDownloadAccount
			accountExists = true
		}
	}

	// Always show the same message for security (don't reveal if email exists)
	const sentMessageKey = "forgot.sent"

	// If account exists, create token and send email
	if accountExists {
		token, err := database.DB.CreatePasswordResetToken(emailAddress, accountType)
		if err != nil {
			log.Printf("Failed to create reset token: %v", err)
			s.renderForgotPasswordPage(w, r, sentMessageKey, true)
			return
		}

		// Send email asynchronously
		go func() {
			err := email.SendPasswordResetEmail(emailAddress, token, s.getPublicURL())
			if err != nil {
				log.Printf("Failed to send password reset email to %s: %v", emailAddress, err)
			} else {
				log.Printf("Password reset email sent to %s", emailAddress)
			}
		}()
	}

	// Always show success message
	s.renderForgotPasswordPage(w, r, sentMessageKey, true)
}

// handleResetPassword shows the reset password page or handles the reset
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		s.renderResetPasswordPage(w, r, "", "reset.error.invalid_link")
		return
	}

	if r.Method == http.MethodGet {
		// Verify token is valid
		_, err := database.DB.GetPasswordResetToken(token)
		if err != nil {
			s.renderResetPasswordPage(w, r, "", "reset.error.expired_link")
			return
		}
		s.renderResetPasswordPage(w, r, token, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderResetPasswordPage(w, r, token, "reset.error.invalid_form")
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate passwords
	if password == "" || confirmPassword == "" {
		s.renderResetPasswordPage(w, r, token, "reset.error.both_required")
		return
	}

	if len(password) < 6 {
		s.renderResetPasswordPage(w, r, token, "reset.error.too_short")
		return
	}

	if password != confirmPassword {
		s.renderResetPasswordPage(w, r, token, "reset.error.mismatch")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		s.renderResetPasswordPage(w, r, token, "reset.error.process_failed")
		return
	}

	// Reset password
	err = database.DB.ResetPasswordWithToken(token, hashedPassword)
	if err != nil {
		log.Printf("Failed to reset password: %v", err)
		s.renderResetPasswordPage(w, r, token, "reset.error.reset_failed")
		return
	}

	log.Printf("Password reset successful for token: %s", token)

	// Show success page
	s.renderPasswordResetSuccessPage(w, r)
}

// renderForgotPasswordPage renders the forgot password form. messageKey is an
// i18n key; success decides whether it is shown as a confirmation or an error.
func (s *Server) renderForgotPasswordPage(w http.ResponseWriter, r *http.Request, messageKey string, success bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tr := s.tr(r)

	messageHTML := ""
	if messageKey != "" {
		class := "error-message"
		if success {
			class = "success-message"
		}
		messageHTML = `<div class="` + class + `">` + tr.T(messageKey) + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="` + string(tr.Lang()) + `">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>` + tr.T("forgot.page_title") + ` - ` + s.config.CompanyName + `</title>
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
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 450px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
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
        input[type="email"] {
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
        .success-message {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .error-message {
            background: #fee;
            border: 1px solid #fcc;
            color: #c33;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .back-link {
            text-align: center;
            margin-top: 20px;
        }
        .back-link a {
            color: ` + s.getPrimaryColor() + `;
            text-decoration: none;
            font-size: 14px;
        }
        .back-link a:hover {
            text-decoration: underline;
        }
        .info-box {
            background: #e3f2fd;
            border-left: 4px solid ` + s.getPrimaryColor() + `;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        .info-box p {
            margin: 0;
            color: #1976d2;
            font-size: 14px;
        }
    </style>
</head>
<body>
    ` + s.getStandaloneLanguageSwitcherHTML(tr) + `
    <div class="container">
        <div class="logo">
            <h1>🔐 ` + tr.T("forgot.title") + `</h1>
            <p>` + s.config.CompanyName + `</p>
        </div>

        ` + messageHTML + `

        <div class="info-box">
            <p>` + tr.T("forgot.info") + `</p>
        </div>

        <form method="POST" action="/forgot-password">
            <div class="form-group">
                <label for="email">` + tr.T("forgot.email_label") + `</label>
                <input type="email" id="email" name="email" required autofocus>
            </div>
            <button type="submit" class="btn">` + tr.T("forgot.submit") + `</button>
        </form>

        <div class="back-link">
            <a href="/login">← ` + tr.T("forgot.back_to_login") + `</a>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderResetPasswordPage renders the reset password form with password
// visibility toggle. errorKey is an i18n key rather than finished text.
func (s *Server) renderResetPasswordPage(w http.ResponseWriter, r *http.Request, token, errorKey string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tr := s.tr(r)

	errorHTML := ""
	if errorKey != "" {
		errorHTML = `<div class="error-message">` + tr.T(errorKey) + `</div>`
	}

	// If no token, show error page
	if token == "" {
		html := `<!DOCTYPE html>
<html lang="` + string(tr.Lang()) + `">
<head>
    <meta charset="UTF-8">
    <meta name="author" content="Ulf Holmström">
    <title>` + tr.T("reset.invalid_page_title") + ` - ` + s.config.CompanyName + `</title>
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
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 450px;
            width: 100%;
            text-align: center;
        }
        .error-icon {
            font-size: 60px;
            margin-bottom: 20px;
        }
        h1 { color: #c33; margin-bottom: 15px; }
        p { color: #666; line-height: 1.6; margin-bottom: 20px; }
        .btn {
            display: inline-block;
            padding: 12px 24px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 600;
        }
    </style>
</head>
<body>
    ` + s.getStandaloneLanguageSwitcherHTML(tr) + `
    <div class="container">
        <div class="error-icon">⚠️</div>
        <h1>` + tr.T("reset.invalid_title") + `</h1>
        <p>` + tr.T(errorKey) + `</p>
        <p>` + tr.T("reset.invalid_body") + `</p>
        <a href="/forgot-password" class="btn">` + tr.T("reset.request_new_link") + `</a>
    </div>
</body>
</html>`
		w.Write([]byte(html))
		return
	}

	html := `<!DOCTYPE html>
<html lang="` + string(tr.Lang()) + `">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>` + tr.T("reset.page_title") + ` - ` + s.config.CompanyName + `</title>
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
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 450px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
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
            position: relative;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
        }
        input[type="password"], input[type="text"] {
            width: 100%;
            padding: 12px 45px 12px 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: ` + s.getPrimaryColor() + `;
        }
        .password-toggle {
            position: absolute;
            right: 12px;
            top: 38px;
            cursor: pointer;
            user-select: none;
            font-size: 20px;
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
        .error-message {
            background: #fee;
            border: 1px solid #fcc;
            color: #c33;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .info-box {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        .info-box p {
            margin: 5px 0;
            color: #856404;
            font-size: 13px;
        }
    </style>
    <script>
        function togglePassword(fieldId) {
            const field = document.getElementById(fieldId);
            const icon = document.getElementById(fieldId + '_icon');
            if (field.type === 'password') {
                field.type = 'text';
                icon.textContent = '🙈';
            } else {
                field.type = 'password';
                icon.textContent = '👁️';
            }
        }

        function validateForm() {
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirm_password').value;

            if (password.length < 6) {
                alert('` + tr.T("reset.js_too_short") + `');
                return false;
            }

            if (password !== confirmPassword) {
                alert('` + tr.T("reset.js_mismatch") + `');
                return false;
            }

            return true;
        }
    </script>
</head>
<body>
    ` + s.getStandaloneLanguageSwitcherHTML(tr) + `
    <div class="container">
        <div class="logo">
            <h1>🔐 ` + tr.T("reset.title") + `</h1>
            <p>` + s.config.CompanyName + `</p>
        </div>

        ` + errorHTML + `

        <div class="info-box">
            <p><strong>` + tr.T("reset.tips_heading") + `</strong></p>
            <p>• ` + tr.T("reset.tip_min_length") + `</p>
            <p>• ` + tr.T("reset.tip_eye") + `</p>
            <p>• ` + tr.T("reset.tip_match") + `</p>
        </div>

        <form method="POST" action="/reset-password?token=` + token + `" onsubmit="return validateForm()">
            <div class="form-group">
                <label for="password">` + tr.T("reset.new_password") + `</label>
                <input type="password" id="password" name="password" required minlength="6" autofocus>
                <span class="password-toggle" id="password_icon"
                      onmousedown="togglePassword('password')"
                      onmouseup="togglePassword('password')"
                      onmouseleave="if(document.getElementById('password').type === 'text') togglePassword('password')">👁️</span>
            </div>
            <div class="form-group">
                <label for="confirm_password">` + tr.T("reset.confirm_password") + `</label>
                <input type="password" id="confirm_password" name="confirm_password" required minlength="6">
                <span class="password-toggle" id="confirm_password_icon"
                      onmousedown="togglePassword('confirm_password')"
                      onmouseup="togglePassword('confirm_password')"
                      onmouseleave="if(document.getElementById('confirm_password').type === 'text') togglePassword('confirm_password')">👁️</span>
            </div>
            <button type="submit" class="btn">` + tr.T("reset.submit") + `</button>
        </form>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderPasswordResetSuccessPage shows success after password reset
func (s *Server) renderPasswordResetSuccessPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tr := s.tr(r)

	html := `<!DOCTYPE html>
<html lang="` + string(tr.Lang()) + `">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>` + tr.T("reset.success_page_title") + ` - ` + s.config.CompanyName + `</title>
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
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 50px 40px;
            max-width: 450px;
            width: 100%;
            text-align: center;
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
            font-size: 40px;
        }
        h1 {
            color: #155724;
            margin-bottom: 20px;
            font-size: 28px;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 15px;
        }
        .btn {
            display: inline-block;
            margin-top: 20px;
            padding: 14px 30px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 600;
            transition: opacity 0.3s;
        }
        .btn:hover {
            opacity: 0.9;
        }
    </style>
    <script>
        setTimeout(function() {
            window.location.href = '/login';
        }, 5000);
    </script>
</head>
<body>
    ` + s.getStandaloneLanguageSwitcherHTML(tr) + `
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>` + tr.T("reset.success_title") + `</h1>
        <p>` + tr.T("reset.success_body") + `</p>
        <p>` + tr.T("reset.success_hint") + `</p>
        <p style="font-size: 14px; color: #999; margin-top: 20px;">
            ` + tr.T("reset.success_redirect") + `
        </p>
        <a href="/login" class="btn">` + tr.T("reset.success_login_now") + `</a>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
