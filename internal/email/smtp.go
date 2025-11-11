package email

import (
	"crypto/tls"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
	"gopkg.in/gomail.v2"
)

// SMTPProvider implementerar EmailProvider för SMTP-servrar
type SMTPProvider struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string
	useTLS    bool
}

// NewSMTPProvider skapar en ny SMTP provider
func NewSMTPProvider(host string, port int, username, password, fromEmail, fromName string, useTLS bool) *SMTPProvider {
	if fromName == "" {
		fromName = "Sharecare"
	}

	return &SMTPProvider{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
		useTLS:    useTLS,
	}
}

// SendEmail skickar ett e-postmeddelande via SMTP
func (sp *SMTPProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(sp.fromEmail, sp.fromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(sp.host, sp.port, sp.username, sp.password)

	if sp.useTLS {
		d.TLSConfig = &tls.Config{
			ServerName:         sp.host,
			InsecureSkipVerify: false,
		}
	} else {
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return d.DialAndSend(m)
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (sp *SMTPProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	subject := "Ny fil uppladdad: " + request.Title
	htmlBody := GenerateUploadNotificationHTML(request, file, uploaderIP, serverURL)
	textBody := GenerateUploadNotificationText(request, file, uploaderIP, serverURL)

	return sp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func (sp *SMTPProvider) SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	subject := "Din fil har laddats ner: " + file.Name
	htmlBody := GenerateDownloadNotificationHTML(file, downloaderIP, serverURL)
	textBody := GenerateDownloadNotificationText(file, downloaderIP, serverURL)

	return sp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendSplashLinkEmail skickar splash link via e-post
func (sp *SMTPProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	subject := "Delad fil: " + file.Name
	htmlBody := GenerateSplashLinkHTML(splashLink, file, message)
	textBody := GenerateSplashLinkText(splashLink, file, message)

	return sp.SendEmail(to, subject, htmlBody, textBody)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func (sp *SMTPProvider) SendAccountDeletionConfirmation(to, accountName string) error {
	subject := "Bekräftelse: Ditt konto har raderats"
	htmlBody := GenerateAccountDeletionHTML(accountName)
	textBody := GenerateAccountDeletionText(accountName)

	return sp.SendEmail(to, subject, htmlBody, textBody)
}
