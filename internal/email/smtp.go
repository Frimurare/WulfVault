// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
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
		fromName = "WulfVault"
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
	log.Printf("📧 Sending email via SMTP to %s through %s:%d (TLS: %v)", to, sp.host, sp.port, sp.useTLS)

	// If TLS is disabled, use plain SMTP (for MailHog, test servers, etc.)
	if !sp.useTLS {
		return sp.sendPlainSMTP(to, subject, htmlBody, textBody)
	}

	// Use gomail for TLS connections
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(sp.fromEmail, sp.fromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(sp.host, sp.port, sp.username, sp.password)
	d.TLSConfig = &tls.Config{
		ServerName:         sp.host,
		InsecureSkipVerify: false,
	}
	log.Printf("🔒 TLS enabled with certificate verification")

	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("❌ SMTP failed to %s:%d - %v", sp.host, sp.port, err)
		return fmt.Errorf("SMTP connection failed to %s:%d - %w", sp.host, sp.port, err)
	}

	log.Printf("✓ Email sent successfully via SMTP to %s", to)
	return nil
}

// sendPlainSMTP sends email using plain SMTP without TLS (for MailHog, etc.)
func (sp *SMTPProvider) sendPlainSMTP(to, subject, htmlBody, textBody string) error {
	log.Printf("⚠️  Using plain SMTP (no TLS) - connection may be insecure")

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", sp.host, sp.port)
	c, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("❌ Failed to connect to %s - %v", addr, err)
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer c.Close()

	// Set sender
	if err = c.Mail(sp.fromEmail); err != nil {
		log.Printf("❌ MAIL FROM failed - %v", err)
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Set recipient
	if err = c.Rcpt(to); err != nil {
		log.Printf("❌ RCPT TO failed - %v", err)
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Send email body
	w, err := c.Data()
	if err != nil {
		log.Printf("❌ DATA command failed - %v", err)
		return fmt.Errorf("DATA command failed: %w", err)
	}

	// Build email message
	msg := fmt.Sprintf("From: %s <%s>\r\n", sp.fromName, sp.fromEmail)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n"
	msg += "\r\n"
	msg += "--boundary123\r\n"
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += textBody + "\r\n"
	msg += "--boundary123\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += htmlBody + "\r\n"
	msg += "--boundary123--\r\n"

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Printf("❌ Failed to write message - %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		log.Printf("❌ Failed to close message - %v", err)
		return fmt.Errorf("failed to close message: %w", err)
	}

	// Quit
	err = c.Quit()
	if err != nil {
		log.Printf("❌ QUIT failed - %v", err)
		return fmt.Errorf("QUIT failed: %w", err)
	}

	log.Printf("✓ Email sent successfully via plain SMTP to %s", to)
	return nil
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
