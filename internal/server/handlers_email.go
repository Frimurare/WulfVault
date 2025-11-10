package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/email"
	"github.com/Frimurare/Sharecare/internal/models"
)

// EmailConfigRequest representerar en request för e-postkonfiguration
type EmailConfigRequest struct {
	Provider         string `json:"provider"`         // "brevo" eller "smtp"
	ApiKey           string `json:"apiKey"`           // För Brevo
	SMTPHost         string `json:"smtpHost"`         // För SMTP
	SMTPPort         int    `json:"smtpPort"`         // För SMTP
	SMTPUsername     string `json:"smtpUsername"`     // För SMTP
	SMTPPassword     string `json:"smtpPassword"`     // För SMTP
	SMTPUseTLS       bool   `json:"smtpUseTLS"`       // För SMTP
	FromEmail        string `json:"fromEmail"`        // Gemensam
	FromName         string `json:"fromName"`         // Gemensam
}

// handleEmailConfigure hanterar konfiguration av e-postinställningar
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

	// Validera provider
	if req.Provider != "brevo" && req.Provider != "smtp" {
		s.sendError(w, http.StatusBadRequest, "Invalid provider. Must be 'brevo' or 'smtp'")
		return
	}

	// Validera från-adress
	if req.FromEmail == "" {
		s.sendError(w, http.StatusBadRequest, "From email is required")
		return
	}

	// Hämta krypteringsnyckel
	masterKey, err := email.GetOrCreateMasterKey(database.DB)
	if err != nil {
		log.Printf("Failed to get encryption key: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Encryption error")
		return
	}

	// Kryptera känslig data
	var apiKeyEncrypted, passwordEncrypted string

	if req.Provider == "brevo" {
		// Brevo: kryptera API-nyckel om den är angiven
		if req.ApiKey != "" {
			apiKeyEncrypted, err = email.EncryptAPIKey(req.ApiKey, masterKey)
			if err != nil {
				log.Printf("Failed to encrypt Brevo API key: %v", err)
				s.sendError(w, http.StatusInternalServerError, "Encryption failed")
				return
			}
		}
	} else {
		// SMTP: kryptera lösenord om det är angivet
		if req.SMTPPassword != "" {
			passwordEncrypted, err = email.EncryptAPIKey(req.SMTPPassword, masterKey)
			if err != nil {
				log.Printf("Failed to encrypt SMTP password: %v", err)
				s.sendError(w, http.StatusInternalServerError, "Encryption failed")
				return
			}
		}
	}

	// Deaktivera alla andra providers
	_, err = database.DB.Exec("UPDATE EmailProviderConfig SET IsActive = 0")
	if err != nil {
		log.Printf("Failed to deactivate providers: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Spara eller uppdatera konfiguration
	now := time.Now().Unix()

	if req.Provider == "brevo" {
		// Kolla om Brevo-config redan finns
		var existingId int
		err := database.DB.QueryRow("SELECT Id FROM EmailProviderConfig WHERE Provider = ?", "brevo").Scan(&existingId)

		if err == nil {
			// Uppdatera befintlig
			updateSQL := `UPDATE EmailProviderConfig SET IsActive = 1, FromEmail = ?, FromName = ?, UpdatedAt = ?`
			args := []interface{}{req.FromEmail, req.FromName, now}

			if apiKeyEncrypted != "" {
				updateSQL += ", ApiKeyEncrypted = ?"
				args = append(args, apiKeyEncrypted)
			}

			updateSQL += " WHERE Provider = ?"
			args = append(args, "brevo")

			_, err = database.DB.Exec(updateSQL, args...)
		} else {
			// Skapa ny
			_, err = database.DB.Exec(`
				INSERT INTO EmailProviderConfig
					(Provider, IsActive, ApiKeyEncrypted, FromEmail, FromName, CreatedAt, UpdatedAt)
				VALUES (?, 1, ?, ?, ?, ?, ?)
			`, "brevo", apiKeyEncrypted, req.FromEmail, req.FromName, now, now)
		}
	} else {
		// SMTP
		var existingId int
		err := database.DB.QueryRow("SELECT Id FROM EmailProviderConfig WHERE Provider = ?", "smtp").Scan(&existingId)

		if err == nil {
			// Uppdatera befintlig
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
			// Skapa ny
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

	log.Printf("Email provider configured: %s", req.Provider)
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleEmailTest testar e-postkonfigurationen
func (s *Server) handleEmailTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Hämta användare från kontext
	user := r.Context().Value("user").(*models.User)

	// Hämta aktiv provider
	provider, err := email.GetActiveProvider(database.DB)
	if err != nil {
		log.Printf("No email provider configured: %v", err)
		s.sendError(w, http.StatusBadRequest, "No email provider configured")
		return
	}

	// Skicka test-e-post
	err = provider.SendEmail(
		user.Email,
		"Sharecare E-post Test",
		"<h1>✓ Test lyckades!</h1><p>Din e-postkonfiguration fungerar korrekt.</p>",
		"Test lyckades! Din e-postkonfiguration fungerar korrekt.",
	)

	if err != nil {
		log.Printf("Email test failed: %v", err)
		s.sendError(w, http.StatusInternalServerError, "Test failed: "+err.Error())
		return
	}

	log.Printf("Test email sent to: %s", user.Email)
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// SendSplashLinkRequest representerar en request för att skicka splash link
type SendSplashLinkRequest struct {
	FileId  string `json:"fileId"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// handleSendSplashLink skickar en splash link via e-post
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

	// Validera input
	if req.FileId == "" || req.Email == "" {
		s.sendError(w, http.StatusBadRequest, "File ID and email are required")
		return
	}

	// Hämta användare från kontext
	user := r.Context().Value("user").(*models.User)

	// Hämta fil
	fileInfo, err := database.DB.GetFile(req.FileId)
	if err != nil {
		log.Printf("File not found: %s", req.FileId)
		s.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	// Kontrollera att användaren äger filen
	if fileInfo.UserId != user.Id {
		log.Printf("User %d tried to share file %s owned by user %d", user.Id, req.FileId, fileInfo.UserId)
		s.sendError(w, http.StatusForbidden, "You can only share your own files")
		return
	}

	// Generera splash link
	splashLink := s.getPublicURL() + "/s/" + fileInfo.Id

	// Skicka e-post
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

// handleEmailSettings renderar e-postinställningssidan
func (s *Server) handleEmailSettings(w http.ResponseWriter, r *http.Request) {
	// Kolla om någon provider är konfigurerad
	var brevoConfigured, smtpConfigured bool
	var brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName string
	var isBrevoActive, isSMTPActive bool

	// Kolla Brevo
	row := database.DB.QueryRow("SELECT FromEmail, FromName, IsActive FROM EmailProviderConfig WHERE Provider = 'brevo'")
	err := row.Scan(&brevoFromEmail, &brevoFromName, &isBrevoActive)
	brevoConfigured = (err == nil && brevoFromEmail != "")

	// Kolla SMTP
	row = database.DB.QueryRow("SELECT FromEmail, FromName, IsActive FROM EmailProviderConfig WHERE Provider = 'smtp'")
	err = row.Scan(&smtpFromEmail, &smtpFromName, &isSMTPActive)
	smtpConfigured = (err == nil && smtpFromEmail != "")

	// Rendera sidan
	s.renderEmailSettingsPage(w, brevoConfigured, smtpConfigured, isBrevoActive, isSMTPActive, brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName)
}

// renderEmailSettingsPage renderar email settings-sidan
func (s *Server) renderEmailSettingsPage(w http.ResponseWriter, brevoConfigured, smtpConfigured, isBrevoActive, isSMTPActive bool, brevoFromEmail, smtpFromEmail, brevoFromName, smtpFromName string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	activeTab := "brevo"
	if isSMTPActive {
		activeTab = "smtp"
	}

	brevoStatus := "Inte konfigurerad"
	brevoButtonText := "Testa anslutning"
	brevoButtonDisabled := ""
	if brevoConfigured {
		brevoStatus = "Konfigurerad"
		if !isBrevoActive {
			brevoStatus += " (inaktiv)"
		}
	} else {
		brevoButtonDisabled = "disabled"
	}

	smtpStatus := "Inte konfigurerad"
	smtpButtonText := "Testa anslutning"
	smtpButtonDisabled := ""
	if smtpConfigured {
		smtpStatus = "Konfigurerad"
		if !isSMTPActive {
			smtpStatus += " (inaktiv)"
		}
	} else {
		smtpButtonDisabled = "disabled"
	}

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E-postinställningar - Sharecare</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 30px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
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
        .back-link {
            display: inline-block;
            margin-bottom: 20px;
            color: #2563eb;
            text-decoration: none;
        }
        .back-link:hover {
            text-decoration: underline;
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
    </style>
</head>
<body>
    <div class="container">
        <a href="/admin" class="back-link">← Tillbaka till Admin</a>

        <h1>E-postinställningar</h1>
        <p class="subtitle">Konfigurera e-postleverantör för att skicka notifieringar</p>

        <div id="success-message" class="success-message"></div>
        <div id="error-message" class="error-message"></div>

        <div class="tab-buttons">
            <button class="tab-btn ` + activeTabClass("brevo", activeTab) + `" data-provider="brevo">Brevo (Sendinblue)</button>
            <button class="tab-btn ` + activeTabClass("smtp", activeTab) + `" data-provider="smtp">SMTP Server</button>
        </div>

        <!-- Brevo Configuration -->
        <div id="brevo-config" class="provider-config ` + activeConfigClass("brevo", activeTab) + `">
            <form id="brevo-form">
                <div class="form-group">
                    <label>Brevo API-nyckel</label>
                    <input type="password"
                           id="brevo-api-key"
                           placeholder="` + placeholderText(brevoConfigured, "xkeysib-...") + `"
                           autocomplete="off">
                    <small>Din API-nyckel krypteras och döljs efter att den sparats.</small>
                </div>

                <div class="form-group">
                    <label>Från e-postadress *</label>
                    <input type="email"
                           id="brevo-from-email"
                           placeholder="no-reply@dittdomän.se"
                           value="` + brevoFromEmail + `"
                           required>
                    <small>Måste vara verifierad i ditt Brevo-konto.</small>
                </div>

                <div class="form-group">
                    <label>Från namn (valfritt)</label>
                    <input type="text"
                           id="brevo-from-name"
                           placeholder="Sharecare"
                           value="` + brevoFromName + `">
                </div>

                <div class="status-indicator">
                    <span class="` + statusClass(brevoConfigured) + `">` + brevoStatus + `</span>
                    <button type="button" class="btn-secondary" id="test-brevo" ` + brevoButtonDisabled + `>` + brevoButtonText + `</button>
                </div>

                <button type="submit" class="btn-primary">Spara Brevo-inställningar</button>
            </form>
        </div>

        <!-- SMTP Configuration -->
        <div id="smtp-config" class="provider-config ` + activeConfigClass("smtp", activeTab) + `">
            <form id="smtp-form">
                <div class="form-group">
                    <label>SMTP-server *</label>
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
                    <small>Vanliga portar: 587 (TLS), 465 (SSL), 25 (utan kryptering)</small>
                </div>

                <div class="form-group">
                    <label>Användarnamn *</label>
                    <input type="text"
                           id="smtp-username"
                           placeholder="din-email@gmail.com"
                           value="` + getSMTPUsername() + `"
                           required>
                </div>

                <div class="form-group">
                    <label>Lösenord</label>
                    <input type="password"
                           id="smtp-password"
                           placeholder="` + placeholderText(smtpConfigured, "••••••••") + `"
                           autocomplete="off">
                    <small>Lösenordet krypteras och döljs efter att det sparats.</small>
                </div>

                <div class="form-group">
                    <label>Från e-postadress *</label>
                    <input type="email"
                           id="smtp-from-email"
                           placeholder="no-reply@dittdomän.se"
                           value="` + smtpFromEmail + `"
                           required>
                </div>

                <div class="form-group">
                    <label>Från namn (valfritt)</label>
                    <input type="text"
                           id="smtp-from-name"
                           placeholder="Sharecare"
                           value="` + smtpFromName + `">
                </div>

                <div class="form-group checkbox-group">
                    <input type="checkbox" id="smtp-use-tls" checked>
                    <label for="smtp-use-tls" style="margin-bottom: 0;">Använd TLS/STARTTLS</label>
                </div>

                <div class="status-indicator">
                    <span class="` + statusClass(smtpConfigured) + `">` + smtpStatus + `</span>
                    <button type="button" class="btn-secondary" id="test-smtp" ` + smtpButtonDisabled + `>` + smtpButtonText + `</button>
                </div>

                <button type="submit" class="btn-primary">Spara SMTP-inställningar</button>
            </form>
        </div>

        <div class="info-box">
            <h3>ℹ️ Säkerhet</h3>
            <ul>
                <li>API-nycklar och lösenord krypteras med AES-256-GCM före lagring</li>
                <li>Krypterade värden döljs i gränssnittet efter att de sparats</li>
                <li>Endast du kan dekryptera och se dessa värden genom att ange dem på nytt</li>
            </ul>
        </div>
    </div>

    <script>
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

            const apiKey = document.getElementById('brevo-api-key').value;
            const fromEmail = document.getElementById('brevo-from-email').value;
            const fromName = document.getElementById('brevo-from-name').value;

            const response = await fetch('/api/email/configure', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    provider: 'brevo',
                    apiKey: apiKey || undefined,
                    fromEmail: fromEmail,
                    fromName: fromName
                })
            });

            if (response.ok) {
                showSuccess('Brevo-inställningar sparade!');
                document.getElementById('brevo-api-key').value = '';
                document.getElementById('brevo-api-key').placeholder = '••••••••••••••••';
                document.getElementById('test-brevo').disabled = false;
                setTimeout(() => location.reload(), 1500);
            } else {
                const error = await response.json();
                showError('Fel: ' + error.error);
            }
        });

        // SMTP form submission
        document.getElementById('smtp-form').addEventListener('submit', async function(e) {
            e.preventDefault();

            const host = document.getElementById('smtp-host').value;
            const port = parseInt(document.getElementById('smtp-port').value);
            const username = document.getElementById('smtp-username').value;
            const password = document.getElementById('smtp-password').value;
            const fromEmail = document.getElementById('smtp-from-email').value;
            const fromName = document.getElementById('smtp-from-name').value;
            const useTLS = document.getElementById('smtp-use-tls').checked;

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
                })
            });

            if (response.ok) {
                showSuccess('SMTP-inställningar sparade!');
                document.getElementById('smtp-password').value = '';
                document.getElementById('smtp-password').placeholder = '••••••••••••••••';
                document.getElementById('test-smtp').disabled = false;
                setTimeout(() => location.reload(), 1500);
            } else {
                const error = await response.json();
                showError('Fel: ' + error.error);
            }
        });

        // Test Brevo
        document.getElementById('test-brevo')?.addEventListener('click', async function() {
            const btn = this;
            btn.disabled = true;
            btn.textContent = 'Testar...';

            const response = await fetch('/api/email/test?provider=brevo');

            if (response.ok) {
                showSuccess('✓ Anslutning till Brevo fungerar! Testmail skickat.');
            } else {
                const error = await response.json();
                showError('✗ Test misslyckades: ' + error.error);
            }

            btn.disabled = false;
            btn.textContent = 'Testa anslutning';
        });

        // Test SMTP
        document.getElementById('test-smtp')?.addEventListener('click', async function() {
            const btn = this;
            btn.disabled = true;
            btn.textContent = 'Testar...';

            const response = await fetch('/api/email/test?provider=smtp');

            if (response.ok) {
                showSuccess('✓ Anslutning till SMTP-server fungerar! Testmail skickat.');
            } else {
                const error = await response.json();
                showError('✗ Test misslyckades: ' + error.error);
            }

            btn.disabled = false;
            btn.textContent = 'Testa anslutning';
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
