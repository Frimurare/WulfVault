// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// MailgunProvider implementerar EmailProvider för Mailgun
type MailgunProvider struct {
	apiKey    string
	domain    string
	fromEmail string
	fromName  string
	region    string // "us" eller "eu"
}

// NewMailgunProvider skapar en ny Mailgun provider
func NewMailgunProvider(apiKey, domain, fromEmail, fromName, region string) *MailgunProvider {
	if fromName == "" {
		fromName = "WulfVault"
	}
	if region == "" {
		region = "us" // Default till US region
	}

	return &MailgunProvider{
		apiKey:    apiKey,
		domain:    domain,
		fromEmail: fromEmail,
		fromName:  fromName,
		region:    region,
	}
}

// getAPIBase returnerar rätt API-bas baserat på region
func (mp *MailgunProvider) getAPIBase() string {
	if mp.region == "eu" {
		return "https://api.eu.mailgun.net/v3"
	}
	return "https://api.mailgun.net/v3"
}

// SendEmail skickar ett e-postmeddelande via Mailgun
func (mp *MailgunProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	log.Printf("📧 Sending email via Mailgun to %s (domain: %s, region: %s)", to, mp.domain, mp.region)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("from", fmt.Sprintf("%s <%s>", mp.fromName, mp.fromEmail))
	_ = writer.WriteField("to", to)
	_ = writer.WriteField("subject", subject)
	_ = writer.WriteField("text", textBody)
	_ = writer.WriteField("html", htmlBody)

	err := writer.Close()
	if err != nil {
		log.Printf("❌ Failed to create multipart form: %v", err)
		return fmt.Errorf("failed to create form data: %w", err)
	}

	// Prepare request
	apiURL := fmt.Sprintf("%s/%s/messages", mp.getAPIBase(), mp.domain)
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		log.Printf("❌ Failed to create Mailgun request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("api", mp.apiKey)

	// Log request details
	log.Printf("🔍 Mailgun API Request:")
	log.Printf("   URL: %s", apiURL)
	log.Printf("   Method: %s", req.Method)
	log.Printf("   Domain: %s", mp.domain)
	log.Printf("   Region: %s", mp.region)
	log.Printf("   From: %s <%s>", mp.fromName, mp.fromEmail)
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Mailgun request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📩 Mailgun Response Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("📩 Mailgun Response Body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mailgun API error: %d %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	log.Printf("✓ Email sent successfully via Mailgun to %s", to)
	return nil
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (mp *MailgunProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	subject := "Ny fil uppladdad: " + request.Title
	htmlBody := GenerateUploadNotificationHTML(request, file, uploaderIP, serverURL)
	textBody := GenerateUploadNotificationText(request, file, uploaderIP, serverURL)

	return mp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func (mp *MailgunProvider) SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	subject := "Din fil har laddats ner: " + file.Name
	htmlBody := GenerateDownloadNotificationHTML(file, downloaderIP, serverURL)
	textBody := GenerateDownloadNotificationText(file, downloaderIP, serverURL)

	return mp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendSplashLinkEmail skickar splash link via e-post
func (mp *MailgunProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	subject := "Delad fil: " + file.Name
	htmlBody := GenerateSplashLinkHTML(splashLink, file, message)
	textBody := GenerateSplashLinkText(splashLink, file, message)

	return mp.SendEmail(to, subject, htmlBody, textBody)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func (mp *MailgunProvider) SendAccountDeletionConfirmation(to, accountName string) error {
	subject := "Bekräftelse: Ditt konto har raderats"
	htmlBody := GenerateAccountDeletionHTML(accountName)
	textBody := GenerateAccountDeletionText(accountName)

	return mp.SendEmail(to, subject, htmlBody, textBody)
}
