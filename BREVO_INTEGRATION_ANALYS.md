# Brevo & E-post Integration - Implementeringsanalys

**Datum:** 2025-11-10
**Repository:** Sharecare v2.0.0
**Språk:** Go 1.23+
**Databas:** SQLite

---

## Sammanfattning

Denna analys beskriver vad som krävs för att implementera Brevo-integration och alternativt SMTP-stöd i Sharecare för att skicka e-postnotifieringar när:
1. Någon laddar upp en fil via en upload request-länk
2. När man vill dela en "Splash Link" via e-post

## Nuläge - Kritiska Fynd

### ✅ Vad som finns idag:
- **Upload Request-funktionalitet** finns och fungerar:
  - Skapar unika länkar: `http://localhost:8080/upload-request/{TOKEN}`
  - Single-use länkar (används endast av en IP-adress)
  - 24 timmars utgångsdatum
  - Spårar vem som använt länken och när
- **Fildelning** med splash links:
  - Splash page: `/s/{FILE_ID}`
  - Direkt nedladdning: `/d/{FILE_ID}`
- **Databas Configuration-tabell** för nyckel-värde inställningar
- **Säker autentisering** med bcrypt för lösenord
- **API Key-system** för programmatisk åtkomst

### ❌ Vad som INTE finns:
- **INGEN e-postfunktionalitet överhuvudtaget**
- Inget SMTP-bibliotek i `go.mod`
- Inga e-postnotifieringar när filer laddas upp
- Inga e-postnotifieringar vid nedladdning
- Ingen konfiguration för e-postserver

**Detta betyder att användare idag måste:**
- Manuellt kopiera och skicka upload request-länkar
- Manuellt kontrollera dashboarden för att se om någon laddat upp filer

---

## Implementation - Detaljerad Plan

### Alternativ 1: Brevo (Sendinblue) API Integration

#### A. Databas-tillägg

**Ny tabell för säker lagring av API-nycklar:**

```sql
CREATE TABLE IF NOT EXISTS EmailProviderConfig (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Provider TEXT NOT NULL,              -- 'brevo' eller 'smtp'
    IsActive INTEGER DEFAULT 0,          -- Endast en kan vara aktiv

    -- För Brevo
    ApiKeyEncrypted TEXT,                -- AES-256 krypterad nyckel

    -- För SMTP
    SMTPHost TEXT,
    SMTPPort INTEGER,
    SMTPUsername TEXT,
    SMTPPasswordEncrypted TEXT,          -- AES-256 krypterad
    SMTPUseTLS INTEGER DEFAULT 1,

    -- Gemensamt
    FromEmail TEXT NOT NULL,
    FromName TEXT,

    CreatedAt INTEGER NOT NULL,
    UpdatedAt INTEGER NOT NULL
);
```

**Uppdatera Configuration-tabellen för krypteringsnyckel:**

```sql
-- Lagra master encryption key (genereras vid första körningen)
INSERT INTO Configuration (Key, Value)
VALUES ('email_encryption_key', '{RANDOM_32_BYTE_HEX}');
```

#### B. Go-paket som behövs

**Lägg till i `go.mod`:**

```go
require (
    // Befintliga dependencies...

    // För Brevo API
    github.com/sendinblue/APIv3-go-library/v2 v2.1.2

    // För SMTP alternativet
    gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df

    // För AES kryptering av API-nycklar
    golang.org/x/crypto v0.31.0  // Redan inkluderad, behöver bara användas
)
```

#### C. Ny katalogstruktur

```
internal/
├── email/
│   ├── email.go              # Interface och huvudlogik
│   ├── brevo.go              # Brevo API implementation
│   ├── smtp.go               # SMTP implementation
│   ├── encryption.go         # Kryptering av API-nycklar
│   └── templates.go          # E-postmallar
```

#### D. Krypteringsimplementation

**Fil: `internal/email/encryption.go`**

```go
package email

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
    "errors"
    "io"
)

// GetOrCreateMasterKey hämtar eller skapar krypteringsnyckeln
func GetOrCreateMasterKey(db *database.Database) ([]byte, error) {
    keyHex, err := db.GetConfigValue("email_encryption_key")
    if err != nil || keyHex == "" {
        // Skapa ny 32-byte nyckel för AES-256
        key := make([]byte, 32)
        if _, err := rand.Read(key); err != nil {
            return nil, err
        }
        keyHex = hex.EncodeToString(key)
        db.SetConfigValue("email_encryption_key", keyHex)
        return key, nil
    }
    return hex.DecodeString(keyHex)
}

// EncryptAPIKey krypterar en API-nyckel med AES-256-GCM
func EncryptAPIKey(plaintext string, masterKey []byte) (string, error) {
    block, err := aes.NewCipher(masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAPIKey dekrypterar en krypterad API-nyckel
func DecryptAPIKey(ciphertext string, masterKey []byte) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

#### E. E-post Service Interface

**Fil: `internal/email/email.go`**

```go
package email

import (
    "errors"
    "sharecare/internal/database"
    "sharecare/internal/models"
)

type EmailProvider interface {
    SendEmail(to, subject, htmlBody, textBody string) error
    SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP string) error
    SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error
}

type EmailService struct {
    provider EmailProvider
    db       *database.Database
}

// GetActiveProvider hämtar den aktiva e-postleverantören
func GetActiveProvider(db *database.Database) (EmailProvider, error) {
    var provider string
    var isActive int
    var apiKeyEncrypted, smtpHost, smtpUsername, smtpPasswordEncrypted, fromEmail string
    var smtpPort int

    row := db.QueryRow(`
        SELECT Provider, ApiKeyEncrypted, SMTPHost, SMTPPort, SMTPUsername,
               SMTPPasswordEncrypted, FromEmail
        FROM EmailProviderConfig
        WHERE IsActive = 1
        LIMIT 1
    `)

    err := row.Scan(&provider, &apiKeyEncrypted, &smtpHost, &smtpPort,
                    &smtpUsername, &smtpPasswordEncrypted, &fromEmail)
    if err != nil {
        return nil, errors.New("no active email provider configured")
    }

    masterKey, err := GetOrCreateMasterKey(db)
    if err != nil {
        return nil, err
    }

    switch provider {
    case "brevo":
        apiKey, err := DecryptAPIKey(apiKeyEncrypted, masterKey)
        if err != nil {
            return nil, err
        }
        return NewBrevoProvider(apiKey, fromEmail), nil

    case "smtp":
        password, err := DecryptAPIKey(smtpPasswordEncrypted, masterKey)
        if err != nil {
            return nil, err
        }
        return NewSMTPProvider(smtpHost, smtpPort, smtpUsername, password, fromEmail), nil

    default:
        return nil, errors.New("unknown email provider: " + provider)
    }
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (es *EmailService) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP string) error {
    // Hämta request-skaparen
    user, err := es.db.GetUser(request.UserId)
    if err != nil {
        return err
    }

    subject := "Ny fil uppladdad: " + request.Title
    htmlBody := generateUploadNotificationHTML(request, file, uploaderIP)
    textBody := generateUploadNotificationText(request, file, uploaderIP)

    return es.provider.SendEmail(user.Email, subject, htmlBody, textBody)
}
```

#### F. Brevo Implementation

**Fil: `internal/email/brevo.go`**

```go
package email

import (
    "context"

    sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

type BrevoProvider struct {
    client    *sendinblue.APIClient
    fromEmail string
    fromName  string
}

func NewBrevoProvider(apiKey, fromEmail string) *BrevoProvider {
    cfg := sendinblue.NewConfiguration()
    cfg.AddDefaultHeader("api-key", apiKey)

    return &BrevoProvider{
        client:    sendinblue.NewAPIClient(cfg),
        fromEmail: fromEmail,
        fromName:  "Sharecare",
    }
}

func (bp *BrevoProvider) SendEmail(to, subject, htmlBody, textBody string) error {
    ctx := context.Background()

    sendEmail := sendinblue.SendSmtpEmail{
        Sender: &sendinblue.SendSmtpEmailSender{
            Email: bp.fromEmail,
            Name:  bp.fromName,
        },
        To: []sendinblue.SendSmtpEmailTo{
            {Email: to},
        },
        Subject:     subject,
        HtmlContent: htmlBody,
        TextContent: textBody,
    }

    _, _, err := bp.client.TransactionalEmailsApi.SendTransacEmail(ctx, sendEmail)
    return err
}

func (bp *BrevoProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP string) error {
    subject := "Ny fil uppladdad: " + request.Title
    htmlBody := generateUploadNotificationHTML(request, file, uploaderIP)
    textBody := generateUploadNotificationText(request, file, uploaderIP)

    // Hämta request-skaparens e-post från databas
    // ... (implementeras i email.go)

    return bp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

func (bp *BrevoProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
    subject := "Delad fil: " + file.Name

    htmlBody := `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .button {
                display: inline-block;
                padding: 12px 24px;
                background: #007bff;
                color: white;
                text-decoration: none;
                border-radius: 5px;
                margin: 20px 0;
            }
            .footer { margin-top: 30px; font-size: 12px; color: #666; }
        </style>
    </head>
    <body>
        <div class="container">
            <h2>Någon har delat en fil med dig</h2>
            ` + (message != "" ? `<p><strong>Meddelande:</strong><br/>` + message + `</p>` : ``) + `
            <p><strong>Filnamn:</strong> ` + file.Name + `</p>
            <p><strong>Storlek:</strong> ` + file.Size + `</p>
            <a href="` + splashLink + `" class="button">Ladda ner fil</a>
            <p style="font-size: 12px; color: #666;">
                Eller kopiera denna länk: <br/>
                <code>` + splashLink + `</code>
            </p>
            <div class="footer">
                <p>Detta är ett automatiskt meddelande från Sharecare.</p>
            </div>
        </div>
    </body>
    </html>
    `

    textBody := `Någon har delat en fil med dig

` + (message != "" ? `Meddelande: ` + message + "\n\n" : ``) + `
Filnamn: ` + file.Name + `
Storlek: ` + file.Size + `

Ladda ner filen här: ` + splashLink + `

---
Detta är ett automatiskt meddelande från Sharecare.
`

    return bp.SendEmail(to, subject, htmlBody, textBody)
}
```

#### G. SMTP Implementation

**Fil: `internal/email/smtp.go`**

```go
package email

import (
    "crypto/tls"
    "gopkg.in/gomail.v2"
)

type SMTPProvider struct {
    host      string
    port      int
    username  string
    password  string
    fromEmail string
    fromName  string
}

func NewSMTPProvider(host string, port int, username, password, fromEmail string) *SMTPProvider {
    return &SMTPProvider{
        host:      host,
        port:      port,
        username:  username,
        password:  password,
        fromEmail: fromEmail,
        fromName:  "Sharecare",
    }
}

func (sp *SMTPProvider) SendEmail(to, subject, htmlBody, textBody string) error {
    m := gomail.NewMessage()
    m.SetHeader("From", sp.fromEmail)
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/plain", textBody)
    m.AddAlternative("text/html", htmlBody)

    d := gomail.NewDialer(sp.host, sp.port, sp.username, sp.password)
    d.TLSConfig = &tls.Config{InsecureSkipVerify: false}

    return d.DialAndSend(m)
}

func (sp *SMTPProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP string) error {
    // Samma implementation som Brevo
    subject := "Ny fil uppladdad: " + request.Title
    htmlBody := generateUploadNotificationHTML(request, file, uploaderIP)
    textBody := generateUploadNotificationText(request, file, uploaderIP)

    return sp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

func (sp *SMTPProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
    // Samma implementation som Brevo
    subject := "Delad fil: " + file.Name
    htmlBody := generateSplashLinkHTML(splashLink, file, message)
    textBody := generateSplashLinkText(splashLink, file, message)

    return sp.SendEmail(to, subject, htmlBody, textBody)
}
```

#### H. E-postmallar

**Fil: `internal/email/templates.go`**

```go
package email

import (
    "fmt"
    "sharecare/internal/database"
    "sharecare/internal/models"
    "time"
)

func generateUploadNotificationHTML(request *models.FileRequest, file *database.FileInfo, uploaderIP string) string {
    uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

    return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .header { background: #007bff; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
            .content { background: #f9f9f9; padding: 20px; }
            .file-info { background: white; padding: 15px; margin: 15px 0; border-left: 4px solid #007bff; }
            .button {
                display: inline-block;
                padding: 12px 24px;
                background: #28a745;
                color: white;
                text-decoration: none;
                border-radius: 5px;
                margin: 20px 0;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h2>✓ Ny fil uppladdad</h2>
            </div>
            <div class="content">
                <p>Någon har laddat upp en fil via din upload request:</p>

                <div class="file-info">
                    <p><strong>Request:</strong> %s</p>
                    <p><strong>Filnamn:</strong> %s</p>
                    <p><strong>Storlek:</strong> %s</p>
                    <p><strong>Uppladdad:</strong> %s</p>
                    <p><strong>IP-adress:</strong> %s</p>
                </div>

                <a href="%s/dashboard" class="button">Visa filer</a>

                <p style="margin-top: 20px; font-size: 12px; color: #666;">
                    Filen finns nu i din dashboard och kan laddas ner.
                </p>
            </div>
        </div>
    </body>
    </html>
    `, request.Title, file.Name, file.Size, uploadTime, uploaderIP, "{SERVER_URL}")
}

func generateUploadNotificationText(request *models.FileRequest, file *database.FileInfo, uploaderIP string) string {
    uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

    return fmt.Sprintf(`
Ny fil uppladdad!

Någon har laddat upp en fil via din upload request:

Request: %s
Filnamn: %s
Storlek: %s
Uppladdad: %s
IP-adress: %s

Logga in för att se och ladda ner filen:
%s/dashboard

---
Detta är ett automatiskt meddelande från Sharecare.
    `, request.Title, file.Name, file.Size, uploadTime, uploaderIP, "{SERVER_URL}")
}
```

#### I. Uppdatera File Request Handler

**Fil: `internal/server/handlers_file_requests.go`**

Lägg till e-postnotifiering efter uppladdning:

```go
// Rad ~380, efter att filen sparats
func (s *Server) handleFileRequestUpload(w http.ResponseWriter, r *http.Request) {
    // ... befintlig kod för att ladda upp fil ...

    // NYTT: Skicka e-postnotifiering
    emailService, err := email.GetActiveProvider(database.DB)
    if err == nil {
        // E-post är konfigurerat, skicka notifiering
        go func() {
            err := emailService.SendFileUploadNotification(fileRequest, fileInfo, ipAddress)
            if err != nil {
                log.Printf("Failed to send email notification: %v", err)
            }
        }()
    }

    // ... resten av koden ...
}
```

#### J. Settings-sida för E-postkonfiguration

**Ny fil: `web/templates/email-settings.html`**

```html
<!DOCTYPE html>
<html lang="sv">
<head>
    <meta charset="UTF-8">
    <title>E-postinställningar - Sharecare</title>
    <link rel="stylesheet" href="/static/css/admin.css">
</head>
<body>
    <div class="container">
        <h1>E-postinställningar</h1>

        <div class="settings-section">
            <h2>Välj E-postleverantör</h2>

            <div class="provider-tabs">
                <button class="tab-btn active" data-provider="brevo">Brevo (Sendinblue)</button>
                <button class="tab-btn" data-provider="smtp">SMTP Server</button>
            </div>

            <!-- Brevo Configuration -->
            <div id="brevo-config" class="provider-config active">
                <form id="brevo-form">
                    <div class="form-group">
                        <label>Brevo API-nyckel</label>
                        <input type="password"
                               id="brevo-api-key"
                               placeholder="xkeysib-..."
                               autocomplete="off">
                        <small>Din API-nyckel krypteras och döljs efter att den sparats.</small>
                    </div>

                    <div class="form-group">
                        <label>Från e-postadress</label>
                        <input type="email"
                               id="brevo-from-email"
                               placeholder="no-reply@dittdomän.se"
                               required>
                        <small>Måste vara verifierad i ditt Brevo-konto.</small>
                    </div>

                    <div class="form-group">
                        <label>Från namn (valfritt)</label>
                        <input type="text"
                               id="brevo-from-name"
                               placeholder="Sharecare"
                               value="Sharecare">
                    </div>

                    <div class="status-indicator" id="brevo-status">
                        {{if .BrevoConfigured}}
                        <span class="status-active">✓ Konfigurerad</span>
                        <button type="button" class="btn-secondary" id="test-brevo">Testa anslutning</button>
                        {{else}}
                        <span class="status-inactive">○ Inte konfigurerad</span>
                        {{end}}
                    </div>

                    <button type="submit" class="btn-primary">Spara Brevo-inställningar</button>
                </form>
            </div>

            <!-- SMTP Configuration -->
            <div id="smtp-config" class="provider-config">
                <form id="smtp-form">
                    <div class="form-group">
                        <label>SMTP-server</label>
                        <input type="text"
                               id="smtp-host"
                               placeholder="smtp.gmail.com"
                               required>
                    </div>

                    <div class="form-group">
                        <label>Port</label>
                        <input type="number"
                               id="smtp-port"
                               placeholder="587"
                               value="587"
                               required>
                        <small>Vanliga portar: 587 (TLS), 465 (SSL), 25 (utan kryptering)</small>
                    </div>

                    <div class="form-group">
                        <label>Användarnamn</label>
                        <input type="text"
                               id="smtp-username"
                               placeholder="din-email@gmail.com"
                               required>
                    </div>

                    <div class="form-group">
                        <label>Lösenord</label>
                        <input type="password"
                               id="smtp-password"
                               placeholder="••••••••"
                               autocomplete="off">
                        <small>Lösenordet krypteras och döljs efter att det sparats.</small>
                    </div>

                    <div class="form-group">
                        <label>Från e-postadress</label>
                        <input type="email"
                               id="smtp-from-email"
                               placeholder="no-reply@dittdomän.se"
                               required>
                    </div>

                    <div class="form-group">
                        <label>Från namn (valfritt)</label>
                        <input type="text"
                               id="smtp-from-name"
                               placeholder="Sharecare"
                               value="Sharecare">
                    </div>

                    <div class="form-group checkbox-group">
                        <label>
                            <input type="checkbox" id="smtp-use-tls" checked>
                            Använd TLS/STARTTLS
                        </label>
                    </div>

                    <div class="status-indicator" id="smtp-status">
                        {{if .SMTPConfigured}}
                        <span class="status-active">✓ Konfigurerad</span>
                        <button type="button" class="btn-secondary" id="test-smtp">Testa anslutning</button>
                        {{else}}
                        <span class="status-inactive">○ Inte konfigurerad</span>
                        {{end}}
                    </div>

                    <button type="submit" class="btn-primary">Spara SMTP-inställningar</button>
                </form>
            </div>
        </div>

        <div class="settings-section">
            <h2>E-postnotifieringar</h2>
            <div class="notification-settings">
                <label>
                    <input type="checkbox" id="notify-file-upload" checked>
                    Skicka notifiering när någon laddar upp via upload request
                </label>
                <br>
                <label>
                    <input type="checkbox" id="notify-file-download" checked>
                    Skicka notifiering när någon laddar ner dina filer
                </label>
            </div>
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

    <script src="/static/js/email-settings.js"></script>
</body>
</html>
```

**Ny fil: `web/static/js/email-settings.js`**

```javascript
// Tab-switching mellan Brevo och SMTP
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', function() {
        const provider = this.dataset.provider;

        // Uppdatera tabs
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        this.classList.add('active');

        // Visa rätt config
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

    if (!apiKey && !isAlreadyConfigured()) {
        alert('Vänligen ange API-nyckel');
        return;
    }

    const response = await fetch('/api/email/configure', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            provider: 'brevo',
            apiKey: apiKey || undefined,  // Skicka endast om ny nyckel angetts
            fromEmail: fromEmail,
            fromName: fromName
        })
    });

    if (response.ok) {
        alert('Brevo-inställningar sparade!');
        // Dölj API-nyckel efter att den sparats
        document.getElementById('brevo-api-key').value = '';
        document.getElementById('brevo-api-key').placeholder = '••••••••••••••••';
        location.reload();
    } else {
        const error = await response.json();
        alert('Fel: ' + error.message);
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
            smtpPassword: password || undefined,  // Skicka endast om nytt lösenord
            smtpUseTLS: useTLS,
            fromEmail: fromEmail,
            fromName: fromName
        })
    });

    if (response.ok) {
        alert('SMTP-inställningar sparade!');
        document.getElementById('smtp-password').value = '';
        document.getElementById('smtp-password').placeholder = '••••••••••••••••';
        location.reload();
    } else {
        const error = await response.json();
        alert('Fel: ' + error.message);
    }
});

// Test Brevo connection
document.getElementById('test-brevo')?.addEventListener('click', async function() {
    const btn = this;
    btn.disabled = true;
    btn.textContent = 'Testar...';

    const response = await fetch('/api/email/test?provider=brevo');

    if (response.ok) {
        alert('✓ Anslutning till Brevo fungerar!');
    } else {
        const error = await response.json();
        alert('✗ Test misslyckades: ' + error.message);
    }

    btn.disabled = false;
    btn.textContent = 'Testa anslutning';
});

// Test SMTP connection
document.getElementById('test-smtp')?.addEventListener('click', async function() {
    const btn = this;
    btn.disabled = true;
    btn.textContent = 'Testar...';

    const response = await fetch('/api/email/test?provider=smtp');

    if (response.ok) {
        alert('✓ Anslutning till SMTP-server fungerar!');
    } else {
        const error = await response.json();
        alert('✗ Test misslyckades: ' + error.message);
    }

    btn.disabled = false;
    btn.textContent = 'Testa anslutning';
});
```

#### K. API-endpoints för E-postkonfiguration

**Lägg till i `internal/server/server.go`:**

```go
// E-post konfiguration (admin only)
mux.HandleFunc("/api/email/configure", s.requireAuth(s.requireAdmin(s.handleEmailConfigure)))
mux.HandleFunc("/api/email/test", s.requireAuth(s.requireAdmin(s.handleEmailTest)))
mux.HandleFunc("/api/email/send-splash-link", s.requireAuth(s.handleSendSplashLink))
```

**Ny fil: `internal/server/handlers_email.go`**

```go
package server

import (
    "encoding/json"
    "net/http"
    "sharecare/internal/database"
    "sharecare/internal/email"
)

type EmailConfigRequest struct {
    Provider         string `json:"provider"`         // "brevo" eller "smtp"
    ApiKey           string `json:"apiKey"`           // För Brevo
    SMTPHost         string `json:"smtpHost"`
    SMTPPort         int    `json:"smtpPort"`
    SMTPUsername     string `json:"smtpUsername"`
    SMTPPassword     string `json:"smtpPassword"`
    SMTPUseTLS       bool   `json:"smtpUseTLS"`
    FromEmail        string `json:"fromEmail"`
    FromName         string `json:"fromName"`
}

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

    // Hämta krypteringsnyckel
    masterKey, err := email.GetOrCreateMasterKey(database.DB)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Encryption key error")
        return
    }

    // Kryptera känslig data beroende på provider
    var apiKeyEncrypted, passwordEncrypted string

    switch req.Provider {
    case "brevo":
        if req.ApiKey != "" {
            apiKeyEncrypted, err = email.EncryptAPIKey(req.ApiKey, masterKey)
            if err != nil {
                s.sendError(w, http.StatusInternalServerError, "Encryption failed")
                return
            }
        }

    case "smtp":
        if req.SMTPPassword != "" {
            passwordEncrypted, err = email.EncryptAPIKey(req.SMTPPassword, masterKey)
            if err != nil {
                s.sendError(w, http.StatusInternalServerError, "Encryption failed")
                return
            }
        }

    default:
        s.sendError(w, http.StatusBadRequest, "Invalid provider")
        return
    }

    // Deaktivera alla andra providers
    _, err = database.DB.Exec("UPDATE EmailProviderConfig SET IsActive = 0")
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Database error")
        return
    }

    // Infoga eller uppdatera konfiguration
    if req.Provider == "brevo" {
        _, err = database.DB.Exec(`
            INSERT INTO EmailProviderConfig
                (Provider, IsActive, ApiKeyEncrypted, FromEmail, FromName, CreatedAt, UpdatedAt)
            VALUES (?, 1, ?, ?, ?, ?, ?)
            ON CONFLICT(Provider) DO UPDATE SET
                IsActive = 1,
                ApiKeyEncrypted = COALESCE(?, ApiKeyEncrypted),
                FromEmail = ?,
                FromName = ?,
                UpdatedAt = ?
        `, req.Provider, apiKeyEncrypted, req.FromEmail, req.FromName,
           time.Now().Unix(), time.Now().Unix(),
           apiKeyEncrypted, req.FromEmail, req.FromName, time.Now().Unix())
    } else {
        _, err = database.DB.Exec(`
            INSERT INTO EmailProviderConfig
                (Provider, IsActive, SMTPHost, SMTPPort, SMTPUsername,
                 SMTPPasswordEncrypted, SMTPUseTLS, FromEmail, FromName,
                 CreatedAt, UpdatedAt)
            VALUES (?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(Provider) DO UPDATE SET
                IsActive = 1,
                SMTPHost = ?,
                SMTPPort = ?,
                SMTPUsername = ?,
                SMTPPasswordEncrypted = COALESCE(?, SMTPPasswordEncrypted),
                SMTPUseTLS = ?,
                FromEmail = ?,
                FromName = ?,
                UpdatedAt = ?
        `, req.Provider, req.SMTPHost, req.SMTPPort, req.SMTPUsername,
           passwordEncrypted, req.SMTPUseTLS, req.FromEmail, req.FromName,
           time.Now().Unix(), time.Now().Unix(),
           req.SMTPHost, req.SMTPPort, req.SMTPUsername, passwordEncrypted,
           req.SMTPUseTLS, req.FromEmail, req.FromName, time.Now().Unix())
    }

    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Failed to save configuration")
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) handleEmailTest(w http.ResponseWriter, r *http.Request) {
    provider, err := email.GetActiveProvider(database.DB)
    if err != nil {
        s.sendError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Hämta användarens e-post för testmeddelande
    user := r.Context().Value("user").(*models.User)

    err = provider.SendEmail(
        user.Email,
        "Sharecare E-post Test",
        "<h1>Test lyckades!</h1><p>Din e-postkonfiguration fungerar korrekt.</p>",
        "Test lyckades! Din e-postkonfiguration fungerar korrekt.",
    )

    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Test failed: "+err.Error())
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

type SendSplashLinkRequest struct {
    FileId  string `json:"fileId"`
    Email   string `json:"email"`
    Message string `json:"message"`
}

func (s *Server) handleSendSplashLink(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req SendSplashLinkRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid request")
        return
    }

    // Hämta fil
    fileInfo, err := database.DB.GetFile(req.FileId)
    if err != nil {
        s.sendError(w, http.StatusNotFound, "File not found")
        return
    }

    // Kontrollera ägarskap
    user := r.Context().Value("user").(*models.User)
    if fileInfo.UserId != user.Id {
        s.sendError(w, http.StatusForbidden, "Not your file")
        return
    }

    // Hämta e-postprovider
    provider, err := email.GetActiveProvider(database.DB)
    if err != nil {
        s.sendError(w, http.StatusBadRequest, "Email not configured")
        return
    }

    // Skicka e-post
    splashLink := s.getPublicURL() + "/s/" + fileInfo.Id
    err = provider.SendSplashLinkEmail(req.Email, splashLink, fileInfo, req.Message)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Failed to send email: "+err.Error())
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": "Email sent to " + req.Email,
    })
}
```

#### L. Uppdatera File Details-sidan

Lägg till "Skicka via e-post"-knapp på fildetaljsidan (`web/templates/file-details.html`):

```html
<div class="file-actions">
    <!-- Befintliga knappar -->
    <button class="btn-primary" onclick="copyToClipboard('{{.SplashLink}}')">
        Kopiera Splash Link
    </button>

    <!-- NY KNAPP -->
    <button class="btn-secondary" onclick="showEmailModal('{{.File.Id}}')">
        📧 Skicka via e-post
    </button>
</div>

<!-- E-post Modal -->
<div id="email-modal" class="modal">
    <div class="modal-content">
        <span class="close" onclick="closeEmailModal()">&times;</span>
        <h2>Skicka Splash Link via e-post</h2>
        <form id="email-form">
            <div class="form-group">
                <label>Mottagarens e-postadress</label>
                <input type="email" id="recipient-email" required>
            </div>
            <div class="form-group">
                <label>Meddelande (valfritt)</label>
                <textarea id="email-message" rows="4"></textarea>
            </div>
            <button type="submit" class="btn-primary">Skicka</button>
        </form>
    </div>
</div>

<script>
function showEmailModal(fileId) {
    document.getElementById('email-modal').style.display = 'block';
    document.getElementById('email-form').onsubmit = async function(e) {
        e.preventDefault();

        const email = document.getElementById('recipient-email').value;
        const message = document.getElementById('email-message').value;

        const response = await fetch('/api/email/send-splash-link', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                fileId: fileId,
                email: email,
                message: message
            })
        });

        if (response.ok) {
            alert('✓ E-post skickad till ' + email);
            closeEmailModal();
        } else {
            const error = await response.json();
            alert('✗ Fel: ' + error.message);
        }
    };
}

function closeEmailModal() {
    document.getElementById('email-modal').style.display = 'none';
}
</script>
```

---

## Säkerhetsaspekter

### 1. Kryptering av API-nycklar
- **Metod:** AES-256-GCM (Galois/Counter Mode)
- **Nyckelhantering:** Master key lagras i Configuration-tabellen
- **Fördelar:**
  - Authenticated encryption (skyddar mot manipulering)
  - Stark kryptering (256-bit)
  - Per-installation unik nyckel

### 2. Visa/Dölja API-nycklar i UI
- **Vid inmatning:** Visas i password-fält
- **Efter sparande:**
  - Fältet töms omedelbart
  - Placeholder ändras till `••••••••••••••••`
  - För att uppdatera: Ange ny nyckel
  - Befintlig nyckel behålls om fältet lämnas tomt

### 3. Databassäkerhet
```sql
-- Master encryption key (256-bit hex)
Configuration:
  Key='email_encryption_key',
  Value='64-character-hex-string'

-- Krypterad API-nyckel (base64)
EmailProviderConfig:
  ApiKeyEncrypted='base64-encoded-aes-gcm-ciphertext'
```

### 4. RBAC (Role-Based Access Control)
- **E-postkonfiguration:** Endast admins (`requireAdmin` middleware)
- **Skicka splash link:** Alla autentiserade användare (endast egna filer)
- **Testa anslutning:** Endast admins

---

## Tidsuppskattning

### Alternativ 1: Endast Brevo
| Uppgift | Tid (timmar) |
|---------|--------------|
| Databas migrations (schema) | 1h |
| Krypteringsmodul (`encryption.go`) | 2h |
| Brevo integration (`brevo.go`) | 3h |
| E-postmallar (`templates.go`) | 2h |
| Settings-sida (HTML/CSS/JS) | 4h |
| API-endpoints för konfiguration | 3h |
| Integration i upload request flow | 2h |
| Integration i file details (skicka splash link) | 2h |
| Testning och debugging | 3h |
| **TOTALT** | **22 timmar** |

### Alternativ 2: Brevo + SMTP
| Uppgift | Extra tid (timmar) |
|---------|---------------------|
| SMTP provider (`smtp.go`) | 2h |
| Utökad settings-sida med tabs | 1h |
| SMTP-specifika fält i form | 1h |
| Testning av båda providers | 2h |
| **EXTRA TOTALT** | **6 timmar** |

**Total tid för båda alternativen:** **28 timmar**

---

## Funktionalitet - Sammanfattning

### ✅ Vad kommer att fungera efter implementation:

1. **Upload Request Notifieringar**
   - När någon laddar upp via `/upload-request/{TOKEN}`
   - E-post skickas automatiskt till request-skaparen
   - Innehåller: filnamn, storlek, uppladdningstid, IP-adress
   - Länk till dashboard för att hämta filen

2. **Skicka Splash Link via E-post**
   - Från file details-sidan (`/file/{ID}`)
   - Knapp: "📧 Skicka via e-post"
   - Modal för att ange mottagare och meddelande
   - E-post med nedladdningslänk skickas

3. **Brevo Integration**
   - Konfigurera API-nyckel i settings
   - Nyckeln krypteras med AES-256-GCM
   - Döljs i UI efter sparande
   - Kan bytas genom att ange ny nyckel

4. **SMTP Alternativ**
   - Stöd för alla SMTP-servrar (Gmail, Office 365, etc.)
   - Konfigurera host, port, användarnamn, lösenord
   - TLS/STARTTLS-stöd
   - Samma kryptering som Brevo

5. **Testfunktion**
   - "Testa anslutning"-knapp
   - Skickar test-e-post till inloggad admin
   - Verifierar konfiguration innan användning

6. **Säkerhet**
   - AES-256-GCM kryptering
   - Master key per installation
   - Authenticated encryption (skydd mot manipulering)
   - Endast admins kan konfigurera

---

## Beroenden (dependencies)

**Lägg till i `go.mod`:**

```go
require (
    // Brevo
    github.com/sendinblue/APIv3-go-library/v2 v2.1.2

    // SMTP
    gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df

    // Kryptering (redan inkluderad)
    golang.org/x/crypto v0.31.0
)
```

**Kör:**
```bash
go get github.com/sendinblue/APIv3-go-library/v2
go get gopkg.in/gomail.v2
go mod tidy
```

---

## Migrationsplan

### Steg 1: Databas
```sql
-- Migrations körs automatiskt vid uppstart
-- Se internal/database/migrations.go
```

### Steg 2: Konfiguration
Admins behöver:
1. Gå till Settings → Email Configuration
2. Välja provider (Brevo eller SMTP)
3. Ange API-nyckel / SMTP-detaljer
4. Testa anslutning
5. Spara

### Steg 3: Användning
- Upload request-notifieringar skickas automatiskt
- Användare kan skicka splash links från file details

---

## Framtida Förbättringar (ej inkluderade nu)

1. **Notifiering vid nedladdning**
   - Skicka e-post när någon laddar ner en fil
   - Konfigurerbart per fil

2. **E-postmallar**
   - Anpassningsbara mallar i admin-panelen
   - Variabler: `{{filename}}`, `{{uploader}}`, etc.

3. **Webhook-stöd**
   - Slack, Discord, Microsoft Teams
   - Webhook-URL i settings

4. **Batch-notifieringar**
   - Samla flera uppladdningar i en e-post
   - Skickas var X timme

5. **E-postloggar**
   - Spara skickade e-postmeddelanden
   - Visa status (skickad, misslyckad, öppnad)

6. **Rate limiting**
   - Begränsa antal e-postmeddelanden per timme
   - Förhindra spam

---

## Rekommendation

Jag rekommenderar **Alternativ 2 (Brevo + SMTP)** eftersom:

1. **Flexibilitet:** Användare kan välja vad som passar dem bäst
2. **Låg kostnad:** SMTP är gratis för många (Gmail, Office 365 etc.)
3. **Skalbarhet:** Brevo för stora volymer, SMTP för små
4. **Endast 6 timmars extra arbete** (21% mer tid för 2x funktionalitet)
5. **Future-proof:** Om Brevo API ändras finns alternativ

**Total implementationstid:** 28 timmar (ca 3.5 arbetsdagar)

---

## Kontaktpunkter för Integration

### Filer som behöver skapas (nya):
1. `internal/email/email.go`
2. `internal/email/brevo.go`
3. `internal/email/smtp.go`
4. `internal/email/encryption.go`
5. `internal/email/templates.go`
6. `internal/server/handlers_email.go`
7. `web/templates/email-settings.html`
8. `web/static/js/email-settings.js`
9. `web/static/css/email-settings.css`

### Filer som behöver modifieras (befintliga):
1. `internal/database/schema.go` - Lägg till EmailProviderConfig-tabell
2. `internal/database/migrations.go` - Migration för ny tabell
3. `internal/server/server.go` - Lägg till routes
4. `internal/server/handlers_file_requests.go` - Rad ~380, lägg till e-postnotifiering
5. `web/templates/file-details.html` - Lägg till "Skicka via e-post"-knapp
6. `go.mod` - Lägg till dependencies

---

## Slutsats

Implementationen är **relativt enkel** eftersom Sharecare redan har:
- God strukturerad kodbase
- Configuration-system
- Säker autentisering med bcrypt
- Upload request-funktionalitet

**Saknas:**
- E-postfunktionalitet (helt ny)
- Krypteringsmodul för API-nycklar (helt ny)
- Settings-sida för e-postkonfiguration (helt ny)

Med **28 timmars arbete** kan hela integrationen vara klar och testad.
