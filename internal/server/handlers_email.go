// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// EmailConfigRequest represents a request for email configuration
type EmailConfigRequest struct {
	Provider     string `json:"provider"`     // "brevo" or "smtp"
	ApiKey       string `json:"apiKey"`       // For Brevo
	SMTPHost     string `json:"smtpHost"`     // For SMTP
	SMTPPort     int    `json:"smtpPort"`     // For SMTP
	SMTPUsername string `json:"smtpUsername"` // For SMTP
	SMTPPassword string `json:"smtpPassword"` // For SMTP
	SMTPUseTLS   bool   `json:"smtpUseTLS"`   // For SMTP
	FromEmail    string `json:"fromEmail"`    // Common
	FromName     string `json:"fromName"`     // Common
}

// handleEmailConfigure handles configuration of email settings
func (s *Server) handleEmailConfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req EmailConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Trim whitespace from all inputs (security + UX improvement)
	req.ApiKey = strings.TrimSpace(req.ApiKey)
	req.FromEmail = strings.TrimSpace(req.FromEmail)
	req.FromName = strings.TrimSpace(req.FromName)
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)
	req.SMTPPassword = strings.TrimSpace(req.SMTPPassword)

	// Log received API key for debugging
	if req.Provider == "brevo" {
		if req.ApiKey != "" {
			log.Printf("🔑 Received NEW Brevo API key in request: length=%d, starts='%s...', ends='...%s'",
				len(req.ApiKey),
				req.ApiKey[:min(15, len(req.ApiKey))],
				req.ApiKey[max(0, len(req.ApiKey)-15):])
		} else {
			log.Printf("⚠️  NO API key in request body (keeping existing)")
		}
	}

	// Validate provider
	if req.Provider != "brevo" && req.Provider != "smtp" {
		s.sendError(w, http.StatusBadRequest, "Invalid provider. Must be 'brevo' or 'smtp'")
		return
	}

	// Validate from address
	if req.FromEmail == "" {
		s.sendError(w, http.StatusBadRequest, "From email is required")
		return
	}

	// Get encryption key
	masterKey, err := email.GetOrCreateMasterKey(database.DB)
	if err != nil {
		log.Printf("Failed to get encryption key: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Encryption error")
		return
	}

	// Encrypt sensitive data
	var apiKeyEncrypted, passwordEncrypted string

	if req.Provider == "brevo" {
		// Brevo: encrypt API key if provided
		if req.ApiKey != "" {
			apiKeyEncrypted, err = email.EncryptAPIKey(req.ApiKey, masterKey)
			if err != nil {
				log.Printf("Failed to encrypt Brevo API key: %v", err)
				s.sendError(w, http.StatusInternalServerError, "Encryption failed")
				return
			}
		}
	} else {
		// SMTP: encrypt password if provided
		if req.SMTPPassword != "" {
			passwordEncrypted, err = email.EncryptAPIKey(req.SMTPPassword, masterKey)
			if err != nil {
				log.Printf("Failed to encrypt SMTP password: %v", err)
				s.sendError(w, http.StatusInternalServerError, "Encryption failed")
				return
			}
		}
	}

	// Deactivate all other providers
	result, err := database.DB.Exec("UPDATE EmailProviderConfig SET IsActive = 0")
	if err != nil {
		log.Printf("Failed to deactivate providers: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Database error")
		return
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Deactivated %d existing provider(s)", rowsAffected)

	// Save or update configuration
	now := time.Now().Unix()

	if req.Provider == "brevo" {
		// Check if Brevo config already exists
		var existingId int
		err := database.DB.QueryRow("SELECT Id FROM EmailProviderConfig WHERE Provider = ?", "brevo").Scan(&existingId)

		if err == nil {
			// Update existing
			log.Printf("Updating existing Brevo config (ID: %d)", existingId)
			updateSQL := `UPDATE EmailProviderConfig SET IsActive = 1, FromEmail = ?, FromName = ?, UpdatedAt = ?`
			args := []interface{}{req.FromEmail, req.FromName, now}

			if apiKeyEncrypted != "" {
				updateSQL += ", ApiKeyEncrypted = ?"
				args = append(args, apiKeyEncrypted)
				log.Printf("Updating with new API key")
			} else {
				log.Printf("Keeping existing API key")
			}

			updateSQL += " WHERE Provider = ?"
			args = append(args, "brevo")

			result, err = database.DB.Exec(updateSQL, args...)
			if err == nil {
				rowsAffected, _ := result.RowsAffected()
				log.Printf("UPDATE affected %d row(s)", rowsAffected)
			}
		} else {
			// Create new
			log.Printf("Creating new Brevo config (no existing found: %v)", err)
			result, err = database.DB.Exec(`
				INSERT INTO EmailProviderConfig
					(Provider, IsActive, ApiKeyEncrypted, FromEmail, FromName, CreatedAt, UpdatedAt)
				VALUES (?, 1, ?, ?, ?, ?, ?)
			`, "brevo", apiKeyEncrypted, req.FromEmail, req.FromName, now, now)
			if err == nil {
				lastId, _ := result.LastInsertId()
				log.Printf("INSERT created row with ID: %d", lastId)
			}
		}
	} else {
		// SMTP
		var existingId int
		err := database.DB.QueryRow("SELECT Id FROM EmailProviderConfig WHERE Provider = ?", "smtp").Scan(&existingId)

		if err == nil {
			// Update existing
			updateSQL := `UPDATE EmailProviderConfig SET IsActive = 1, SMTPHost = ?, SMTPPort = ?,
						  SMTPUsername = ?, SMTPUseTLS = ?, FromEmail = ?, FromName = ?, UpdatedAt = ?`
			args := []interface{}{req.SMTPHost, req.SMTPPort, req.SMTPUsername, btoi(req.SMTPUseTLS), req.FromEmail, req.FromName, now}

			if passwordEncrypted != "" {
				updateSQL += ", SMTPPasswordEncrypted = ?"
				args = append(args, passwordEncrypted)
			}

			updateSQL += " WHERE Provider = ?"
			args = append(args, "smtp")

			_, err = database.DB.Exec(updateSQL, args...)
		} else {
			// Create new
			_, err = database.DB.Exec(`
				INSERT INTO EmailProviderConfig
					(Provider, IsActive, SMTPHost, SMTPPort, SMTPUsername, SMTPPasswordEncrypted,
					 SMTPUseTLS, FromEmail, FromName, CreatedAt, UpdatedAt)
				VALUES (?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, "smtp", req.SMTPHost, req.SMTPPort, req.SMTPUsername, passwordEncrypted,
				btoi(req.SMTPUseTLS), req.FromEmail, req.FromName, now, now)
		}
	}

	if err != nil {
		log.Printf("Failed to save email configuration: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	// Verify the configuration was saved by querying it back
	var verifyId int
	var verifyActive int
	verifyErr := database.DB.QueryRow(`
		SELECT Id, IsActive FROM EmailProviderConfig
		WHERE Provider = ? AND IsActive = 1
	`, req.Provider).Scan(&verifyId, &verifyActive)

	if verifyErr != nil {
		log.Printf("⚠️ WARNING: Configuration save reported success, but verification query failed: %v", verifyErr)
	} else {
		log.Printf("✅ Verified: Configuration saved with ID=%d, IsActive=%d", verifyId, verifyActive)
	}

	log.Printf("Email provider configured: %s", req.Provider)
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleEmailTest tests the email configuration
// EmailTestRequest represents a request to test email settings
type EmailTestRequest struct {
	Provider     string `json:"provider"`
	ApiKey       string `json:"apiKey"`
	FromEmail    string `json:"fromEmail"`
	FromName     string `json:"fromName"`
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"smtpPassword"`
	SMTPUseTLS   bool   `json:"smtpUseTLS"`
}

func (s *Server) handleEmailTest(w http.ResponseWriter, r *http.Request) {
	// Accept both GET and POST
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get user from context (middleware ensures this exists)
	user, ok := userFromContext(r.Context())
	if !ok || user == nil {
		s.sendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var provider email.EmailProvider
	var err error

	if r.Method == http.MethodPost {
		// Test with provided settings (without saving)
		var req EmailTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Create temporary provider for testing
		switch req.Provider {
		case "brevo":
			if req.ApiKey == "" {
				s.sendError(w, http.StatusBadRequest, "API key is required")
				return
			}
			if req.FromEmail == "" {
				s.sendError(w, http.StatusBadRequest, "From email is required")
				return
			}
			provider = email.NewBrevoProvider(req.ApiKey, req.FromEmail, req.FromName)
			log.Printf("Testing Brevo with API key: %s...", req.ApiKey[:min(10, len(req.ApiKey))])

		case "smtp":
			if req.SMTPHost == "" || req.SMTPUsername == "" || req.SMTPPassword == "" {
				s.sendError(w, http.StatusBadRequest, "SMTP host, username and password are required")
				return
			}
			if req.FromEmail == "" {
				s.sendError(w, http.StatusBadRequest, "From email is required")
				return
			}
			provider = email.NewSMTPProvider(req.SMTPHost, req.SMTPPort, req.SMTPUsername, req.SMTPPassword, req.FromEmail, req.FromName, req.SMTPUseTLS)

		default:
			s.sendError(w, http.StatusBadRequest, "Invalid provider")
			return
		}
	} else {
		// GET method - use saved configuration
		provider, err = email.GetActiveProvider(database.DB)
		if err != nil {
			log.Printf("No email provider configured: %v", err)
			s.sendError(w, http.StatusBadRequest, "No email provider configured")
			return
		}
	}

	// Send test email
	err = provider.SendEmail(
		user.Email,
		"WulfVault Email Test",
		"<h1>Test successful!</h1><p>Your email configuration is working correctly.</p>",
		"Test successful! Your email configuration is working correctly.",
	)

	if err != nil {
		log.Printf("Email test failed: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Test failed: "+err.Error())
		return
	}

	log.Printf("Test email sent to: %s", user.Email)
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Test email sent successfully!"})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SendSplashLinkRequest represents a request to send a splash link
type SendSplashLinkRequest struct {
	FileId  string `json:"fileId"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// handleSendSplashLink sends a splash link via email
func (s *Server) handleSendSplashLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SendSplashLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.FileId == "" || req.Email == "" {
		s.sendError(w, http.StatusBadRequest, "File ID and email are required")
		return
	}

	// Get user from context
	user := r.Context().Value("user").(*models.User)

	// Get file
	fileInfo, err := database.DB.GetFileByID(req.FileId)
	if err != nil {
		log.Printf("File not found: %s", req.FileId)
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Check that the user owns the file
	if fileInfo.UserId != user.Id {
		log.Printf("User %d tried to share file %s owned by user %d", user.Id, req.FileId, fileInfo.UserId)
		s.sendError(w, http.StatusForbidden, "You can only share your own files")
		return
	}

	// Generate splash link
	splashLink := s.getPublicURL() + "/s/" + fileInfo.Id

	// Send email
	err = email.SendSplashLinkEmail(req.Email, splashLink, fileInfo, req.Message)
	if err != nil {
		log.Printf("Failed to send splash link email: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	log.Printf("Splash link sent to %s for file %s", req.Email, fileInfo.Name)
	s.sendJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Email sent to " + req.Email,
	})
}

// handleEmailSettings renders the email settings page
func (s *Server) handleEmailSettings(w http.ResponseWriter, r *http.Request) {
	// Check if any provider is configured
	var brevoConfigured, smtpConfigured bool
	var brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName string
	var isBrevoActive, isSMTPActive bool

	// Check Brevo
	row := database.DB.QueryRow("SELECT FromEmail, FromName, IsActive FROM EmailProviderConfig WHERE Provider = 'brevo'")
	err := row.Scan(&brevoFromEmail, &brevoFromName, &isBrevoActive)
	brevoConfigured = (err == nil && brevoFromEmail != "")

	// Check SMTP
	row = database.DB.QueryRow("SELECT FromEmail, FromName, IsActive FROM EmailProviderConfig WHERE Provider = 'smtp'")
	err = row.Scan(&smtpFromEmail, &smtpFromName, &isSMTPActive)
	smtpConfigured = (err == nil && smtpFromEmail != "")

	// Render page
	s.renderEmailSettingsPage(w, brevoConfigured, smtpConfigured, isBrevoActive, isSMTPActive, brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName)
}

// renderEmailSettingsPage renders the email settings page
func (s *Server) renderEmailSettingsPage(w http.ResponseWriter, brevoConfigured, smtpConfigured, isBrevoActive, isSMTPActive bool, brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	activeTab := "brevo"
	if isSMTPActive {
		activeTab = "smtp"
	}

	brevoStatus := "Not configured"
	brevoButtonText := "Test connection"
	brevoButtonDisabled := ""
	if brevoConfigured {
		brevoStatus = "Configured"
		if !isBrevoActive {
			brevoStatus += " (inactive)"
		}
	} else {
		brevoButtonDisabled = "disabled"
	}

	smtpStatus := "Not configured"
	smtpButtonText := "Test connection"
	smtpButtonDisabled := ""
	if smtpConfigured {
		smtpStatus = "Configured"
		if !isSMTPActive {
			smtpStatus += " (inactive)"
		}
	} else {
		smtpButtonDisabled = "disabled"
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Email Settings - WulfVault</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header .logo {
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .header .logo img {
            max-height: 50px;
            max-width: 180px;
        }
        .header h1 {
            color: white;
            font-size: 24px;
            font-weight: 600;
        }
        .header nav {
            display: flex;
            align-items: center;
            gap: 20px;
        }
        .header nav a {
            color: rgba(255, 255, 255, 0.9);
            text-decoration: none;
            font-weight: 500;
            transition: color 0.3s;
        }
        .header nav a:hover {
            color: white;
        }
        .container {
            max-width: 1000px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .settings-card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08);
            padding: 30px;
        }
        h1 {
            color: #1a1a2e;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 15px;
        }
        .tab-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }
        .tab-btn {
            padding: 12px 24px;
            background: none;
            border: none;
            border-bottom: 3px solid transparent;
            cursor: pointer;
            font-size: 16px;
            color: #666;
            transition: all 0.3s;
        }
        .tab-btn.active {
            color: #2563eb;
            border-bottom-color: #2563eb;
        }
        .tab-btn:hover {
            color: #2563eb;
        }
        .provider-config {
            display: none;
            animation: fadeIn 0.3s;
        }
        .provider-config.active {
            display: block;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            font-weight: 600;
            margin-bottom: 8px;
            color: #333;
        }
        .form-group input[type="text"],
        .form-group input[type="email"],
        .form-group input[type="password"],
        .form-group input[type="number"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .form-group small {
            display: block;
            margin-top: 5px;
            color: #666;
            font-size: 12px;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .checkbox-group input[type="checkbox"] {
            width: 18px;
            height: 18px;
        }
        .status-indicator {
            margin: 20px 0;
            padding: 15px;
            background: #f9f9f9;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .status-active {
            color: #28a745;
            font-weight: 600;
        }
        .status-inactive {
            color: #666;
        }
        .btn-primary {
            background: #2563eb;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
        }
        .btn-primary:hover {
            background: #1e40af;
        }
        .btn-primary:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .btn-secondary:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .info-box {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-box h3 {
            color: #1976d2;
            margin-bottom: 10px;
        }
        .info-box ul {
            margin-left: 20px;
            color: #555;
        }
        .info-box li {
            margin: 5px 0;
        }
        .success-message {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            padding: 12px;
            border-radius: 4px;
            margin-bottom: 20px;
            display: none;
        }
        .error-message {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
            padding: 12px;
            border-radius: 4px;
            margin-bottom: 20px;
            display: none;
        }

        /* Mobile Responsive Styles */
        @media (max-width: 768px) {
            .container {
                margin: 20px auto;
                padding: 0 15px;
            }
            .settings-card {
                padding: 20px 15px;
            }
            h1 {
                font-size: 24px;
            }
            .tab-buttons {
                flex-direction: column;
                gap: 0;
            }
            .tab-btn {
                width: 100%;
                text-align: left;
                padding: 15px;
                border-bottom: 1px solid #e0e0e0;
                border-left: 3px solid transparent;
            }
            .tab-btn.active {
                border-bottom-color: #e0e0e0;
                border-left-color: #2563eb;
                background: #f8fafc;
            }
            .form-group input[type="text"],
            .form-group input[type="email"],
            .form-group input[type="password"],
            .form-group input[type="number"] {
                padding: 14px;
                font-size: 16px;
                min-height: 48px;
            }
            .checkbox-group {
                padding: 10px 0;
            }
            .checkbox-group input[type="checkbox"] {
                width: 24px;
                height: 24px;
            }
            .checkbox-group label {
                font-size: 16px;
            }
            .status-indicator {
                flex-direction: column;
                gap: 15px;
                align-items: stretch;
            }
            .btn-primary,
            .btn-secondary {
                width: 100%;
                padding: 14px 24px;
                font-size: 16px;
                min-height: 48px;
            }
            .info-box {
                padding: 12px;
            }
            .info-box h3 {
                font-size: 16px;
            }
            .info-box ul {
                font-size: 14px;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `

    <div class="container">
        <div class="settings-card">
            <h1>Email Settings</h1>
            <p class="subtitle">Configure email provider to send notifications</p>

            <div id="success-message" class="success-message"></div>
            <div id="error-message" class="error-message"></div>

            ` + getActiveProviderBanner(isBrevoActive, isSMTPActive) + `

        <div class="tab-buttons">
            <button class="tab-btn ` + activeTabClass("brevo", activeTab) + `" data-provider="brevo">
                Brevo (Sendinblue) ` + getActiveProviderBadge(isBrevoActive) + `
            </button>
            <button class="tab-btn ` + activeTabClass("smtp", activeTab) + `" data-provider="smtp">
                SMTP Server ` + getActiveProviderBadge(isSMTPActive) + `
            </button>
        </div>

        <!-- Brevo Configuration -->
        <div id="brevo-config" class="provider-config ` + activeConfigClass("brevo", activeTab) + `">
            <form id="brevo-form">
                <div class="form-group">
                    <label>Brevo API Key (REST API, not SMTP)</label>
                    <input type="password"
                           id="brevo-api-key"
                           placeholder="` + placeholderText(brevoConfigured, "xkeysib-...") + `"
                           autocomplete="off">
                    <small><strong>Important:</strong> Use an API key (starts with <code>xkeysib-</code>), NOT an SMTP API key (<code>xsmtpsib-</code>). Create at Settings → SMTP & API → API Keys in Brevo dashboard.</small>
                </div>

                <div class="form-group">
                    <label>From Email Address *</label>
                    <input type="email"
                           id="brevo-from-email"
                           placeholder="no-reply@yourdomain.com"
                           value="` + brevoFromEmail + `"
                           required>
                    <small>Must be verified in your Brevo account.</small>
                </div>

                <div class="form-group">
                    <label>From Name (optional)</label>
                    <input type="text"
                           id="brevo-from-name"
                           placeholder="WulfVault"
                           value="` + brevoFromName + `">
                </div>

                <div class="status-indicator">
                    <span class="` + statusClass(brevoConfigured) + `">` + brevoStatus + `</span>
                    <button type="button" class="btn-secondary" id="test-brevo" ` + brevoButtonDisabled + `>` + brevoButtonText + `</button>
                </div>

                <button type="submit" class="btn-primary">Save Brevo Settings</button>
            </form>
        </div>

        <!-- SMTP Configuration -->
        <div id="smtp-config" class="provider-config ` + activeConfigClass("smtp", activeTab) + `">
            <form id="smtp-form">
                <div class="form-group">
                    <label>SMTP Server *</label>
                    <input type="text"
                           id="smtp-host"
                           placeholder="smtp.gmail.com"
                           value="` + getSMTPHost() + `"
                           required>
                </div>

                <div class="form-group">
                    <label>Port *</label>
                    <input type="number"
                           id="smtp-port"
                           placeholder="587"
                           value="` + getSMTPPort() + `"
                           required>
                    <small>Common ports: 587 (TLS), 465 (SSL), 25 (no encryption)</small>
                </div>

                <div class="form-group">
                    <label>Username *</label>
                    <input type="text"
                           id="smtp-username"
                           placeholder="your-email@gmail.com"
                           value="` + getSMTPUsername() + `"
                           required>
                </div>

                <div class="form-group">
                    <label>Password</label>
                    <input type="password"
                           id="smtp-password"
                           placeholder="` + placeholderText(smtpConfigured, "••••••••") + `"
                           autocomplete="off">
                    <small>Password is encrypted and hidden after saving.</small>
                </div>

                <div class="form-group">
                    <label>From Email Address *</label>
                    <input type="email"
                           id="smtp-from-email"
                           placeholder="no-reply@yourdomain.com"
                           value="` + smtpFromEmail + `"
                           required>
                </div>

                <div class="form-group">
                    <label>From Name (optional)</label>
                    <input type="text"
                           id="smtp-from-name"
                           placeholder="WulfVault"
                           value="` + smtpFromName + `">
                </div>

                <div class="form-group checkbox-group">
                    <input type="checkbox" id="smtp-use-tls" checked>
                    <label for="smtp-use-tls" style="margin-bottom: 0;">Use TLS/STARTTLS</label>
                </div>

                <div class="status-indicator">
                    <span class="` + statusClass(smtpConfigured) + `">` + smtpStatus + `</span>
                    <button type="button" class="btn-secondary" id="test-smtp" ` + smtpButtonDisabled + `>` + smtpButtonText + `</button>
                </div>

                <button type="submit" class="btn-primary">Save SMTP Settings</button>
            </form>
        </div>

        <div class="info-box">
            <h3>Security Information</h3>
            <ul>
                <li>API keys and passwords are encrypted with AES-256-GCM before storage</li>
                <li>Encrypted values are hidden in the interface after saving</li>
                <li>Only you can decrypt and view these values by entering them again</li>
            </ul>
        </div>
        </div>
    </div>

    <script>
        // Mobile Navigation
        (function() {
            'use strict';
            function initMobileNav() {
                const header = document.querySelector('.header');
                if (!header) return;

                const nav = header.querySelector('nav');
                const hamburger = header.querySelector('.hamburger');
                const overlay = document.querySelector('.mobile-nav-overlay');

                if (!nav || !hamburger || !overlay) return;

                function toggleMenu() {
                    const isActive = nav.classList.contains('active');
                    nav.classList.toggle('active');
                    hamburger.classList.toggle('active');
                    overlay.classList.toggle('active');
                    hamburger.setAttribute('aria-expanded', !isActive);

                    if (!isActive) {
                        document.body.style.overflow = 'hidden';
                    } else {
                        document.body.style.overflow = '';
                    }
                }

                function closeMenu() {
                    nav.classList.remove('active');
                    hamburger.classList.remove('active');
                    overlay.classList.remove('active');
                    hamburger.setAttribute('aria-expanded', 'false');
                    document.body.style.overflow = '';
                }

                hamburger.addEventListener('click', function(e) {
                    e.stopPropagation();
                    toggleMenu();
                });

                overlay.addEventListener('click', closeMenu);

                document.addEventListener('keydown', function(e) {
                    if (e.key === 'Escape' && nav.classList.contains('active')) {
                        closeMenu();
                    }
                });

                window.addEventListener('resize', function() {
                    if (window.innerWidth > 768 && nav.classList.contains('active')) {
                        closeMenu();
                    }
                });
            }

            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', initMobileNav);
            } else {
                initMobileNav();
            }
        })();

        // Tab switching
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const provider = this.dataset.provider;

                // Update tabs
                document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                this.classList.add('active');

                // Show correct config
                document.querySelectorAll('.provider-config').forEach(c => c.classList.remove('active'));
                document.getElementById(provider + '-config').classList.add('active');
            });
        });

        // Brevo form submission
        document.getElementById('brevo-form').addEventListener('submit', async function(e) {
            e.preventDefault();

            const apiKey = document.getElementById('brevo-api-key').value.trim();
            const fromEmail = document.getElementById('brevo-from-email').value.trim();
            const fromName = document.getElementById('brevo-from-name').value.trim();

            try {
                const response = await fetch('/api/email/configure', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        provider: 'brevo',
                        apiKey: apiKey || undefined,
                        fromEmail: fromEmail,
                        fromName: fromName
                    }),
                    signal: AbortSignal.timeout(30000)
                });

                if (response.ok) {
                    showSuccess('Brevo settings saved successfully! You can now test the connection.');
                    // Clear the API key field and update placeholder
                    document.getElementById('brevo-api-key').value = '';
                    document.getElementById('brevo-api-key').placeholder = '••••••••••••••••';
                    document.getElementById('test-brevo').disabled = false;
                    // Don't reload - it can cause issues with form state
                } else {
                    const error = await response.json();
                    showError('Error: ' + error.error);
                }
            } catch (err) {
                if (err.name === 'TimeoutError') {
                    showError('Request timed out. Please try again.');
                } else if (err.name === 'AbortError') {
                    showError('Request was aborted. Please try again.');
                } else {
                    showError('Error: ' + err.message);
                }
            }
        });

        // SMTP form submission
        document.getElementById('smtp-form').addEventListener('submit', async function(e) {
            e.preventDefault();

            const host = document.getElementById('smtp-host').value.trim();
            const port = parseInt(document.getElementById('smtp-port').value);
            const username = document.getElementById('smtp-username').value.trim();
            const password = document.getElementById('smtp-password').value.trim();
            const fromEmail = document.getElementById('smtp-from-email').value.trim();
            const fromName = document.getElementById('smtp-from-name').value.trim();
            const useTLS = document.getElementById('smtp-use-tls').checked;

            try {
                const response = await fetch('/api/email/configure', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        provider: 'smtp',
                        smtpHost: host,
                        smtpPort: port,
                        smtpUsername: username,
                        smtpPassword: password || undefined,
                        smtpUseTLS: useTLS,
                        fromEmail: fromEmail,
                        fromName: fromName
                    }),
                    signal: AbortSignal.timeout(30000)
                });

                if (response.ok) {
                    showSuccess('SMTP settings saved successfully! You can now test the connection.');
                    // Clear the password field and update placeholder
                    document.getElementById('smtp-password').value = '';
                    document.getElementById('smtp-password').placeholder = '••••••••••••••••';
                    document.getElementById('test-smtp').disabled = false;
                    // Don't reload - it can cause issues with form state
                } else {
                    const error = await response.json();
                    showError('Error: ' + error.error);
                }
            } catch (err) {
                if (err.name === 'TimeoutError') {
                    showError('Request timed out. Please try again.');
                } else if (err.name === 'AbortError') {
                    showError('Request was aborted. Please try again.');
                } else {
                    showError('Error: ' + err.message);
                }
            }
        });

        // Test Brevo
        document.getElementById('test-brevo')?.addEventListener('click', async function() {
            const btn = this;
            const apiKey = document.getElementById('brevo-api-key').value.trim();
            const apiKeyPlaceholder = document.getElementById('brevo-api-key').placeholder;
            const fromEmail = document.getElementById('brevo-from-email').value.trim();
            const fromName = document.getElementById('brevo-from-name').value.trim();

            btn.disabled = true;
            btn.textContent = 'Testing...';

            try {
                let response;

                // If API key field has value, test with provided config (before save)
                if (apiKey && fromEmail) {
                    response = await fetch('/api/email/test', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        credentials: 'same-origin',
                        body: JSON.stringify({
                            provider: 'brevo',
                            apiKey: apiKey,
                            fromEmail: fromEmail,
                            fromName: fromName
                        }),
                        signal: AbortSignal.timeout(30000)
                    });
                }
                // If API key field is empty but placeholder shows saved (bullets), test saved config
                else if (apiKeyPlaceholder && apiKeyPlaceholder.includes('•')) {
                    response = await fetch('/api/email/test', {
                        method: 'GET',
                        credentials: 'same-origin',
                        signal: AbortSignal.timeout(30000)
                    });
                }
                // Neither provided nor saved config exists
                else {
                    showError('Please save your Brevo settings first, or enter API key and from email to test before saving');
                    btn.disabled = false;
                    btn.textContent = 'Test connection';
                    return;
                }

                if (response.ok) {
                    const result = await response.json();
                    showSuccess(result.message || 'Connection to Brevo successful! Test email sent.');
                } else {
                    const error = await response.json();
                    showError('Test failed: ' + error.error);
                }
            } catch (err) {
                if (err.name === 'TimeoutError') {
                    showError('Test timed out. Please check your connection and try again.');
                } else if (err.name === 'AbortError') {
                    showError('Test was aborted. Please try again.');
                } else {
                    showError('Test failed: ' + err.message);
                }
            } finally {
                btn.disabled = false;
                btn.textContent = 'Test connection';
            }
        });

        // Test SMTP
        document.getElementById('test-smtp')?.addEventListener('click', async function() {
            const btn = this;
            const host = document.getElementById('smtp-host').value.trim();
            const port = parseInt(document.getElementById('smtp-port').value) || 587;
            const username = document.getElementById('smtp-username').value.trim();
            const password = document.getElementById('smtp-password').value.trim();
            const passwordPlaceholder = document.getElementById('smtp-password').placeholder;
            const fromEmail = document.getElementById('smtp-from-email').value.trim();
            const fromName = document.getElementById('smtp-from-name').value.trim();
            const useTLS = document.getElementById('smtp-use-tls').checked;

            btn.disabled = true;
            btn.textContent = 'Testing...';

            try {
                let response;

                // If all fields have values, test with provided config (before save)
                if (host && username && password && fromEmail) {
                    response = await fetch('/api/email/test', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        credentials: 'same-origin',
                        body: JSON.stringify({
                            provider: 'smtp',
                            smtpHost: host,
                            smtpPort: port,
                            smtpUsername: username,
                            smtpPassword: password,
                            fromEmail: fromEmail,
                            fromName: fromName,
                            smtpUseTLS: useTLS
                        }),
                        signal: AbortSignal.timeout(30000)
                    });
                }
                // If password field is empty but placeholder shows saved (bullets), test saved config
                else if (passwordPlaceholder && passwordPlaceholder.includes('•')) {
                    response = await fetch('/api/email/test', {
                        method: 'GET',
                        credentials: 'same-origin',
                        signal: AbortSignal.timeout(30000)
                    });
                }
                // Neither provided nor saved config exists
                else {
                    showError('Please save your SMTP settings first, or fill in all required fields to test before saving');
                    btn.disabled = false;
                    btn.textContent = 'Test connection';
                    return;
                }

                if (response.ok) {
                    const result = await response.json();
                    showSuccess(result.message || 'Connection to SMTP server successful! Test email sent.');
                } else {
                    const error = await response.json();
                    showError('Test failed: ' + error.error);
                }
            } catch (err) {
                if (err.name === 'TimeoutError') {
                    showError('Test timed out. Please check your connection and try again.');
                } else if (err.name === 'AbortError') {
                    showError('Test was aborted. Please try again.');
                } else {
                    showError('Test failed: ' + err.message);
                }
            } finally {
                btn.disabled = false;
                btn.textContent = 'Test connection';
            }
        });

        function showSuccess(message) {
            const el = document.getElementById('success-message');
            el.textContent = message;
            el.style.display = 'block';
            setTimeout(() => el.style.display = 'none', 5000);
        }

        function showError(message) {
            const el = document.getElementById('error-message');
            el.textContent = message;
            el.style.display = 'block';
            setTimeout(() => el.style.display = 'none', 5000);
        }
    </script>
</body>
</html>`

	fmt.Fprint(w, html)
}

// Helper functions

func activeTabClass(provider, activeTab string) string {
	if provider == activeTab {
		return "active"
	}
	return ""
}

func activeConfigClass(provider, activeTab string) string {
	if provider == activeTab {
		return "active"
	}
	return ""
}

func statusClass(configured bool) string {
	if configured {
		return "status-active"
	}
	return "status-inactive"
}

func placeholderText(configured bool, defaultText string) string {
	if configured {
		return "••••••••••••••••"
	}
	return defaultText
}

func getSMTPHost() string {
	// TODO: Retrieve from database if exists
	return ""
}

func getSMTPPort() string {
	// TODO: Retrieve from database if exists
	return "587"
}

func getSMTPUsername() string {
	// TODO: Retrieve from database if exists
	return ""
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func getActiveProviderBanner(isBrevoActive, isSMTPActive bool) string {
	if isBrevoActive {
		return `
        <div style="background: #d4edda; border: 1px solid #c3e6cb; color: #155724; padding: 16px; border-radius: 4px; margin-bottom: 20px;">
            <strong>✓ Active Provider:</strong> Brevo (Sendinblue) - Email notifications are enabled
        </div>`
	} else if isSMTPActive {
		return `
        <div style="background: #d4edda; border: 1px solid #c3e6cb; color: #155724; padding: 16px; border-radius: 4px; margin-bottom: 20px;">
            <strong>✓ Active Provider:</strong> SMTP Server - Email notifications are enabled
        </div>`
	}
	return `
        <div style="background: #fff3cd; border: 1px solid #ffc107; color: #856404; padding: 16px; border-radius: 4px; margin-bottom: 20px;">
            <strong>⚠ No Active Provider:</strong> Configure Brevo or SMTP to enable email notifications
        </div>`
}

func getActiveProviderBadge(isActive bool) string {
	if isActive {
		return `<span style="display: inline-block; background: #28a745; color: white; padding: 2px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; margin-left: 8px;">ACTIVE</span>`
	}
	return ""
}
