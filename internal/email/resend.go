// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// ResendProvider implementerar EmailProvider för Resend
type ResendProvider struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewResendProvider skapar en ny Resend provider
func NewResendProvider(apiKey, fromEmail, fromName string) *ResendProvider {
	if fromName == "" {
		fromName = "WulfVault"
	}

	return &ResendProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// ResendEmailRequest representerar Resend API email request
type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text,omitempty"`
	Html    string   `json:"html,omitempty"`
}

// SendEmail skickar ett e-postmeddelande via Resend
func (rp *ResendProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	log.Printf("📧 Sending email via Resend to %s", to)

	// Prepare request body
	from := rp.fromEmail
	if rp.fromName != "" {
		from = fmt.Sprintf("%s <%s>", rp.fromName, rp.fromEmail)
	}

	reqBody := ResendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    textBody,
		Html:    htmlBody,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ Failed to marshal Resend request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create Resend request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+rp.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Log request details
	log.Printf("🔍 Resend API Request:")
	log.Printf("   URL: %s", req.URL.String())
	log.Printf("   Method: %s", req.Method)
	log.Printf("   From: %s", from)
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Resend request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📩 Resend Response Status: %d %s", resp.StatusCode, resp.Status)
	if len(respBody) > 0 {
		log.Printf("📩 Resend Response Body: %s", string(respBody))
	}

	// Resend returns 200 OK on success
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("resend API error: %d %s - %v", resp.StatusCode, resp.Status, errResp)
	}

	log.Printf("✓ Email sent successfully via Resend to %s", to)
	return nil
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (rp *ResendProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	subject := "Ny fil uppladdad: " + request.Title
	htmlBody := GenerateUploadNotificationHTML(request, file, uploaderIP, serverURL)
	textBody := GenerateUploadNotificationText(request, file, uploaderIP, serverURL)

	return rp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func (rp *ResendProvider) SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	subject := "Din fil har laddats ner: " + file.Name
	htmlBody := GenerateDownloadNotificationHTML(file, downloaderIP, serverURL)
	textBody := GenerateDownloadNotificationText(file, downloaderIP, serverURL)

	return rp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendSplashLinkEmail skickar splash link via e-post
func (rp *ResendProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	subject := "Delad fil: " + file.Name
	htmlBody := GenerateSplashLinkHTML(splashLink, file, message)
	textBody := GenerateSplashLinkText(splashLink, file, message)

	return rp.SendEmail(to, subject, htmlBody, textBody)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func (rp *ResendProvider) SendAccountDeletionConfirmation(to, accountName string) error {
	subject := "Bekräftelse: Ditt konto har raderats"
	htmlBody := GenerateAccountDeletionHTML(accountName)
	textBody := GenerateAccountDeletionText(accountName)

	return rp.SendEmail(to, subject, htmlBody, textBody)
}
