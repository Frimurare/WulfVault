package email

import (
	"database/sql"
	"errors"
	"log"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

// EmailProvider definierar interfacet för e-postleverantörer
type EmailProvider interface {
	SendEmail(to, subject, htmlBody, textBody string) error
	SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error
	SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error
	SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error
	SendAccountDeletionConfirmation(to, accountName string) error
}

// EmailService hanterar e-posttjänster
type EmailService struct {
	provider EmailProvider
	db       *database.Database
}

// NewEmailService skapar en ny e-posttjänst
func NewEmailService(provider EmailProvider, db *database.Database) *EmailService {
	return &EmailService{
		provider: provider,
		db:       db,
	}
}

// GetActiveProvider hämtar den aktiva e-postleverantören från databasen
func GetActiveProvider(db *database.Database) (EmailProvider, error) {
	var provider string
	var apiKeyEncrypted, smtpHost, smtpUsername, smtpPasswordEncrypted, fromEmail, fromName sql.NullString
	var smtpPort, smtpUseTLS sql.NullInt64

	row := db.QueryRow(`
		SELECT Provider, ApiKeyEncrypted, SMTPHost, SMTPPort, SMTPUsername,
		       SMTPPasswordEncrypted, SMTPUseTLS, FromEmail, FromName
		FROM EmailProviderConfig
		WHERE IsActive = 1
		LIMIT 1
	`)

	err := row.Scan(&provider, &apiKeyEncrypted, &smtpHost, &smtpPort,
		&smtpUsername, &smtpPasswordEncrypted, &smtpUseTLS, &fromEmail, &fromName)
	if err != nil {
		log.Printf("GetActiveProvider scan error: %v", err)
		return nil, errors.New("no active email provider configured")
	}

	log.Printf("GetActiveProvider found: provider=%s, hasApiKey=%v, fromEmail=%v",
		provider, apiKeyEncrypted.Valid, fromEmail.Valid)

	// Hämta master key för dekryptering
	masterKey, err := GetOrCreateMasterKey(db)
	if err != nil {
		return nil, err
	}

	switch provider {
	case "brevo":
		if !apiKeyEncrypted.Valid || apiKeyEncrypted.String == "" {
			return nil, errors.New("brevo API key not configured")
		}
		apiKey, err := DecryptAPIKey(apiKeyEncrypted.String, masterKey)
		if err != nil {
			log.Printf("Failed to decrypt Brevo API key: %v", err)
			return nil, err
		}
		prefix := apiKey
		if len(apiKey) > 10 {
			prefix = apiKey[:10]
		}
		log.Printf("Decrypted API key length: %d chars, starts with: %s...", len(apiKey), prefix)
		return NewBrevoProvider(apiKey, fromEmail.String, fromName.String), nil

	case "smtp":
		if !smtpPasswordEncrypted.Valid || smtpPasswordEncrypted.String == "" {
			return nil, errors.New("SMTP password not configured")
		}
		password, err := DecryptAPIKey(smtpPasswordEncrypted.String, masterKey)
		if err != nil {
			return nil, err
		}
		useTLS := smtpUseTLS.Valid && smtpUseTLS.Int64 == 1
		return NewSMTPProvider(smtpHost.String, int(smtpPort.Int64), smtpUsername.String, password, fromEmail.String, fromName.String, useTLS), nil

	default:
		return nil, errors.New("unknown email provider: " + provider)
	}
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		log.Printf("Email not configured, skipping upload notification: %v", err)
		return nil // Don't fail the upload if email fails
	}

	return provider.SendFileUploadNotification(request, file, uploaderIP, serverURL, recipientEmail)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		log.Printf("Email not configured, skipping download notification: %v", err)
		return nil // Don't fail the download if email fails
	}

	return provider.SendFileDownloadNotification(file, downloaderIP, serverURL, recipientEmail)
}

// SendSplashLinkEmail skickar splash link via e-post
func SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}

	return provider.SendSplashLinkEmail(to, splashLink, file, message)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func SendAccountDeletionConfirmation(to, accountName string) error {
	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		log.Printf("Email not configured, skipping deletion confirmation: %v", err)
		return nil // Don't fail the deletion if email fails
	}

	return provider.SendAccountDeletionConfirmation(to, accountName)
}
