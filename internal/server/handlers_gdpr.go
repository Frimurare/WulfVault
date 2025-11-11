package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/email"
	"github.com/Frimurare/Sharecare/internal/models"
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

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = `<div style="background: #fee; border: 1px solid #c33; color: #c33; padding: 15px; border-radius: 5px; margin-bottom: 20px;">` + errorMsg + `</div>`
	}

	html := `<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mitt konto - ` + s.config.CompanyName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 50px auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: ` + s.getPrimaryColor() + `;
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 { font-size: 24px; margin-bottom: 5px; }
        .header p { opacity: 0.9; font-size: 14px; }
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
    <div class="container">
        <div class="header">
            <h1>Mitt Nedladdningskonto</h1>
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
    <title>Konto raderat - ` + s.config.CompanyName + `</title>
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
