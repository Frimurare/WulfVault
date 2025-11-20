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
		s.renderForgotPasswordPage(w, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderForgotPasswordPage(w, "Invalid form data")
		return
	}

	emailAddress := r.FormValue("email")
	if emailAddress == "" {
		s.renderForgotPasswordPage(w, "Email is required")
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
	successMessage := "If you have an account, a password reset email has been sent!"

	// If account exists, create token and send email
	if accountExists {
		token, err := database.DB.CreatePasswordResetToken(emailAddress, accountType)
		if err != nil {
			log.Printf("Failed to create reset token: %v", err)
			s.renderForgotPasswordPage(w, successMessage)
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
	s.renderForgotPasswordPage(w, "SUCCESS:"+successMessage)
}

// handleResetPassword shows the reset password page or handles the reset
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		s.renderResetPasswordPage(w, "", "Invalid reset link")
		return
	}

	if r.Method == http.MethodGet {
		// Verify token is valid
		_, err := database.DB.GetPasswordResetToken(token)
		if err != nil {
			s.renderResetPasswordPage(w, "", "Invalid or expired reset link")
			return
		}
		s.renderResetPasswordPage(w, token, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderResetPasswordPage(w, token, "Invalid form data")
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate passwords
	if password == "" || confirmPassword == "" {
		s.renderResetPasswordPage(w, token, "Both password fields are required")
		return
	}

	if len(password) < 6 {
		s.renderResetPasswordPage(w, token, "Password must be at least 6 characters")
		return
	}

	if password != confirmPassword {
		s.renderResetPasswordPage(w, token, "Passwords do not match")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		s.renderResetPasswordPage(w, token, "Failed to process password")
		return
	}

	// Reset password
	err = database.DB.ResetPasswordWithToken(token, hashedPassword)
	if err != nil {
		log.Printf("Failed to reset password: %v", err)
		s.renderResetPasswordPage(w, token, "Failed to reset password. The link may be expired or already used.")
		return
	}

	log.Printf("Password reset successful for token: %s", token)

	// Show success page
	s.renderPasswordResetSuccessPage(w)
}

// renderForgotPasswordPage renders the forgot password form
func (s *Server) renderForgotPasswordPage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	messageHTML := ""
	if message != "" {
		if len(message) > 8 && message[:8] == "SUCCESS:" {
			messageHTML = `<div class="success-message">` + message[8:] + `</div>`
		} else {
			messageHTML = `<div class="error-message">` + message + `</div>`
		}
	}

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Glömt Lösenord - ` + s.config.CompanyName + `</title>
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
    <div class="container">
        <div class="logo">
            <h1>🔐 Glömt Lösenord?</h1>
            <p>` + s.config.CompanyName + `</p>
        </div>

        ` + messageHTML + `

        <div class="info-box">
            <p>Ange din e-postadress så skickar vi dig en länk för att återställa ditt lösenord.</p>
        </div>

        <form method="POST" action="/forgot-password">
            <div class="form-group">
                <label for="email">E-postadress</label>
                <input type="email" id="email" name="email" required autofocus>
            </div>
            <button type="submit" class="btn">Skicka Återställningslänk</button>
        </form>

        <div class="back-link">
            <a href="/login">← Tillbaka till inloggning</a>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderResetPasswordPage renders the reset password form with password visibility toggle
func (s *Server) renderResetPasswordPage(w http.ResponseWriter, token, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = `<div class="error-message">` + errorMsg + `</div>`
	}

	// If no token, show error page
	if token == "" {
		html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="author" content="Ulf Holmström">
    <title>Felaktig Länk - ` + s.config.CompanyName + `</title>
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
    <div class="container">
        <div class="error-icon">⚠️</div>
        <h1>Felaktig Återställningslänk</h1>
        <p>` + errorMsg + `</p>
        <p>Länken kan vara utgången eller felaktig. Försök begära en ny återställningslänk.</p>
        <a href="/forgot-password" class="btn">Begär Ny Länk</a>
    </div>
</body>
</html>`
		w.Write([]byte(html))
		return
	}

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Återställ Lösenord - ` + s.config.CompanyName + `</title>
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
                alert('Lösenordet måste vara minst 6 tecken långt');
                return false;
            }

            if (password !== confirmPassword) {
                alert('Lösenorden matchar inte!');
                return false;
            }

            return true;
        }
    </script>
</head>
<body>
    <div class="container">
        <div class="logo">
            <h1>🔐 Nytt Lösenord</h1>
            <p>` + s.config.CompanyName + `</p>
        </div>

        ` + errorHTML + `

        <div class="info-box">
            <p><strong>Tips:</strong></p>
            <p>• Minst 6 tecken</p>
            <p>• Håll in ögat-ikonen för att se lösenordet</p>
            <p>• Se till att båda fälten matchar</p>
        </div>

        <form method="POST" action="/reset-password?token=` + token + `" onsubmit="return validateForm()">
            <div class="form-group">
                <label for="password">Nytt Lösenord</label>
                <input type="password" id="password" name="password" required minlength="6" autofocus>
                <span class="password-toggle" id="password_icon"
                      onmousedown="togglePassword('password')"
                      onmouseup="togglePassword('password')"
                      onmouseleave="if(document.getElementById('password').type === 'text') togglePassword('password')">👁️</span>
            </div>
            <div class="form-group">
                <label for="confirm_password">Bekräfta Nytt Lösenord</label>
                <input type="password" id="confirm_password" name="confirm_password" required minlength="6">
                <span class="password-toggle" id="confirm_password_icon"
                      onmousedown="togglePassword('confirm_password')"
                      onmouseup="togglePassword('confirm_password')"
                      onmouseleave="if(document.getElementById('confirm_password').type === 'text') togglePassword('confirm_password')">👁️</span>
            </div>
            <button type="submit" class="btn">Återställ Lösenord</button>
        </form>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// renderPasswordResetSuccessPage shows success after password reset
func (s *Server) renderPasswordResetSuccessPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Lösenord Återställt - ` + s.config.CompanyName + `</title>
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
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Lösenord Återställt!</h1>
        <p>Ditt lösenord har uppdaterats framgångsrikt.</p>
        <p>Du kan nu logga in med ditt nya lösenord.</p>
        <p style="font-size: 14px; color: #999; margin-top: 20px;">
            Du omdirigeras automatiskt till inloggningssidan om 5 sekunder...
        </p>
        <a href="/login" class="btn">Logga In Nu</a>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
