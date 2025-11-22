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

// SendGridProvider implementerar EmailProvider för SendGrid
type SendGridProvider struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewSendGridProvider skapar en ny SendGrid provider
func NewSendGridProvider(apiKey, fromEmail, fromName string) *SendGridProvider {
	if fromName == "" {
		fromName = "WulfVault"
	}

	return &SendGridProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendGridEmailRequest representerar SendGrid API v3 email request
type SendGridEmailRequest struct {
	Personalizations []struct {
		To []struct {
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		} `json:"to"`
	} `json:"personalizations"`
	From struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	} `json:"from"`
	Subject string `json:"subject"`
	Content []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"content"`
}

// SendEmail skickar ett e-postmeddelande via SendGrid
func (sp *SendGridProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	log.Printf("📧 Sending email via SendGrid to %s", to)

	// Prepare request body
	reqBody := SendGridEmailRequest{
		Subject: subject,
	}

	// Set sender
	reqBody.From.Email = sp.fromEmail
	reqBody.From.Name = sp.fromName

	// Set recipient
	reqBody.Personalizations = []struct {
		To []struct {
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		} `json:"to"`
	}{
		{
			To: []struct {
				Email string `json:"email"`
				Name  string `json:"name,omitempty"`
			}{
				{Email: to},
			},
		},
	}

	// Add both text and HTML content
	reqBody.Content = []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		{
			Type:  "text/plain",
			Value: textBody,
		},
		{
			Type:  "text/html",
			Value: htmlBody,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ Failed to marshal SendGrid request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create SendGrid request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+sp.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Log request details
	log.Printf("🔍 SendGrid API Request:")
	log.Printf("   URL: %s", req.URL.String())
	log.Printf("   Method: %s", req.Method)
	log.Printf("   From: %s <%s>", sp.fromName, sp.fromEmail)
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ SendGrid request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📩 SendGrid Response Status: %d %s", resp.StatusCode, resp.Status)
	if len(respBody) > 0 {
		log.Printf("📩 SendGrid Response Body: %s", string(respBody))
	}

	// SendGrid returns 202 Accepted on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("sendgrid API error: %d %s - %v", resp.StatusCode, resp.Status, errResp)
	}

	log.Printf("✓ Email sent successfully via SendGrid to %s", to)
	return nil
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (sp *SendGridProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	subject := "Ny fil uppladdad: " + request.Title
	htmlBody := GenerateUploadNotificationHTML(request, file, uploaderIP, serverURL)
	textBody := GenerateUploadNotificationText(request, file, uploaderIP, serverURL)

	return sp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func (sp *SendGridProvider) SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	subject := "Din fil har laddats ner: " + file.Name
	htmlBody := GenerateDownloadNotificationHTML(file, downloaderIP, serverURL)
	textBody := GenerateDownloadNotificationText(file, downloaderIP, serverURL)

	return sp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendSplashLinkEmail skickar splash link via e-post
func (sp *SendGridProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	subject := "Delad fil: " + file.Name
	htmlBody := GenerateSplashLinkHTML(splashLink, file, message)
	textBody := GenerateSplashLinkText(splashLink, file, message)

	return sp.SendEmail(to, subject, htmlBody, textBody)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func (sp *SendGridProvider) SendAccountDeletionConfirmation(to, accountName string) error {
	subject := "Bekräftelse: Ditt konto har raderats"
	htmlBody := GenerateAccountDeletionHTML(accountName)
	textBody := GenerateAccountDeletionText(accountName)

	return sp.SendEmail(to, subject, htmlBody, textBody)
}
