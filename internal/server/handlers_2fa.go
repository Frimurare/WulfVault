// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/totp"
)

// handle2FASetup initiates 2FA setup by generating a secret and QR code
func (s *Server) handle2FASetup(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		// Generate new TOTP secret
		key, err := totp.GenerateSecret(user.Email, s.config.CompanyName)
		if err != nil {
			http.Error(w, "Failed to generate secret", http.StatusInternalServerError)
			return
		}

		// Generate QR code
		qrCode, err := totp.GenerateQRCode(key)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			return
		}

		// Generate backup codes
		backupCodes, err := totp.GenerateBackupCodes()
		if err != nil {
			http.Error(w, "Failed to generate backup codes", http.StatusInternalServerError)
			return
		}

		// Store secret temporarily in session (not in database yet)
		// We'll only save it after the user verifies it works
		sessionData := map[string]interface{}{
			"totp_secret":   key.Secret(),
			"backup_codes":  backupCodes,
			"user_id":       user.Id,
			"setup_started": time.Now().Unix(),
		}

		sessionJSON, _ := json.Marshal(sessionData)

		// Store in a temporary setup cookie (5 minutes expiry)
		http.SetCookie(w, &http.Cookie{
			Name:     "totp_setup",
			Value:    base64.StdEncoding.EncodeToString(sessionJSON),
			Path:     "/",
			Expires:  time.Now().Add(5 * time.Minute),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Return QR code and backup codes as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"qr_code":      base64.StdEncoding.EncodeToString(qrCode),
			"secret":       key.Secret(),
			"backup_codes": backupCodes,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handle2FAEnable verifies the TOTP code and enables 2FA
func (s *Server) handle2FAEnable(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get setup data from cookie
	cookie, err := r.Cookie("totp_setup")
	if err != nil {
		http.Error(w, "Setup session not found", http.StatusBadRequest)
		return
	}

	sessionData := make(map[string]interface{})
	decodedData, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid setup session", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(decodedData, &sessionData); err != nil {
		http.Error(w, "Invalid setup session", http.StatusBadRequest)
		return
	}

	secret, ok := sessionData["totp_secret"].(string)
	if !ok {
		http.Error(w, "Invalid setup session", http.StatusBadRequest)
		return
	}

	// Get verification code from request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Verification code is required",
		})
		return
	}

	// Validate the code
	if !totp.ValidateCode(code, secret) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid verification code",
		})
		return
	}

	// Code is valid, enable 2FA
	backupCodesRaw, _ := sessionData["backup_codes"].([]interface{})
	backupCodes := make([]string, len(backupCodesRaw))
	for i, v := range backupCodesRaw {
		backupCodes[i] = v.(string)
	}

	if err := database.DB.EnableTOTP(user.Id, secret, backupCodes); err != nil {
		http.Error(w, "Failed to enable 2FA", http.StatusInternalServerError)
		return
	}

	// Clear setup cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "totp_setup",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Two-factor authentication enabled successfully",
	})
}

// handle2FADisable disables 2FA for the user
func (s *Server) handle2FADisable(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify password before disabling
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Password is required",
		})
		return
	}

	// Authenticate with password
	_, err = auth.AuthenticateUser(user.Email, password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid password",
		})
		return
	}

	// Disable 2FA
	if err := database.DB.DisableTOTP(user.Id); err != nil {
		http.Error(w, "Failed to disable 2FA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Two-factor authentication disabled",
	})
}

// handle2FAVerify verifies a TOTP code during login
func (s *Server) handle2FAVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render2FAVerifyPage(w, r, "")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get pending user ID from temporary cookie
	cookie, err := r.Cookie("totp_pending")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var pendingData struct {
		UserID    int   `json:"user_id"`
		CreatedAt int64 `json:"created_at"`
	}

	decodedData, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := json.Unmarshal(decodedData, &pendingData); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if cookie has expired (5 minutes)
	if time.Now().Unix()-pendingData.CreatedAt > 300 {
		http.Redirect(w, r, "/login?error=Session expired", http.StatusSeeOther)
		return
	}

	// Get user
	user, err := database.DB.GetUserByID(pendingData.UserID)
	if err != nil || !user.TOTPEnabled {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.render2FAVerifyPage(w, r, "Invalid form data")
		return
	}

	code := r.FormValue("code")
	useBackup := r.FormValue("use_backup") == "1"

	var valid bool

	if useBackup {
		// Validate backup code
		valid, err = database.DB.ValidateBackupCode(user.Id, code)
		if err != nil {
			s.render2FAVerifyPage(w, r, "Error validating backup code")
			return
		}
	} else {
		// Validate TOTP code
		valid = totp.ValidateCode(code, user.TOTPSecret)
	}

	if !valid {
		s.render2FAVerifyPage(w, r, "Invalid verification code")
		return
	}

	// Clear pending cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "totp_pending",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Create session
	sessionID, err := auth.CreateSession(user.Id)
	if err != nil {
		s.render2FAVerifyPage(w, r, "Failed to create session")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect to appropriate dashboard
	if user.IsAdmin() {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// render2FAVerifyPage renders the 2FA verification page
func (s *Server) render2FAVerifyPage(w http.ResponseWriter, r *http.Request, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Two-Factor Authentication - ` + s.config.CompanyName + `</title>
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
        .verify-container {
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
        input[type="text"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 18px;
            text-align: center;
            letter-spacing: 0.3em;
            font-family: monospace;
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
        .help-text {
            text-align: center;
            margin-top: 15px;
            color: #666;
            font-size: 13px;
        }
        .backup-link {
            text-align: center;
            margin-top: 15px;
        }
        .backup-link a {
            color: ` + s.getPrimaryColor() + `;
            text-decoration: none;
            font-size: 14px;
        }
        #backup-form {
            display: none;
        }
    </style>
</head>
<body>
    <div class="verify-container">
        <div class="logo">
            <h1>` + s.config.CompanyName + `</h1>
            <p>Two-Factor Authentication</p>
        </div>`

	if errorMsg != "" {
		html += `<div class="error">` + errorMsg + `</div>`
	}

	html += `
        <form method="POST" action="/2fa/verify" id="totp-form">
            <div class="form-group">
                <label for="code">Enter the 6-digit code from your authenticator app</label>
                <input type="text" id="code" name="code" maxlength="6" pattern="[0-9]{6}" required autofocus autocomplete="off">
                <input type="hidden" name="use_backup" value="0">
            </div>
            <button type="submit" class="btn">Verify</button>
        </form>

        <form method="POST" action="/2fa/verify" id="backup-form">
            <div class="form-group">
                <label for="backup-code">Enter a backup code</label>
                <input type="text" id="backup-code" name="code" maxlength="16" required autocomplete="off">
                <input type="hidden" name="use_backup" value="1">
            </div>
            <button type="submit" class="btn">Verify Backup Code</button>
        </form>

        <div class="help-text">
            Enter the code from your authenticator app
        </div>

        <div class="backup-link">
            <a href="#" id="toggle-backup">Use a backup code instead</a>
        </div>
    </div>

    <script>
        const totpForm = document.getElementById('totp-form');
        const backupForm = document.getElementById('backup-form');
        const toggleLink = document.getElementById('toggle-backup');
        let showingBackup = false;

        toggleLink.addEventListener('click', (e) => {
            e.preventDefault();
            showingBackup = !showingBackup;

            if (showingBackup) {
                totpForm.style.display = 'none';
                backupForm.style.display = 'block';
                toggleLink.textContent = 'Use authenticator app instead';
            } else {
                totpForm.style.display = 'block';
                backupForm.style.display = 'none';
                toggleLink.textContent = 'Use a backup code instead';
            }
        });

        // Auto-submit when 6 digits are entered
        document.getElementById('code').addEventListener('input', (e) => {
            if (e.target.value.length === 6) {
                setTimeout(() => totpForm.submit(), 100);
            }
        });
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// handle2FARegenerateBackupCodes regenerates backup codes
func (s *Server) handle2FARegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !user.TOTPEnabled {
		http.Error(w, "2FA is not enabled", http.StatusBadRequest)
		return
	}

	// Regenerate codes
	codes, err := database.DB.RegenerateBackupCodes(user.Id)
	if err != nil {
		http.Error(w, "Failed to regenerate backup codes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"backup_codes": codes,
	})
}
