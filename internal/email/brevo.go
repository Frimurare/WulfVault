package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

// BrevoProvider implementerar EmailProvider för Brevo (Sendinblue)
type BrevoProvider struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewBrevoProvider skapar en ny Brevo provider
func NewBrevoProvider(apiKey, fromEmail, fromName string) *BrevoProvider {
	if fromName == "" {
		fromName = "Sharecare"
	}

	return &BrevoProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// BrevoEmailRequest representerar Brevo API email request
type BrevoEmailRequest struct {
	Sender struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"sender"`
	To []struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	} `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"htmlContent,omitempty"`
	TextContent string `json:"textContent,omitempty"`
}

// SendEmail skickar ett e-postmeddelande via Brevo
func (bp *BrevoProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	// Prepare request
	reqBody := BrevoEmailRequest{
		Subject: subject,
	}
	reqBody.Sender.Name = bp.fromName
	reqBody.Sender.Email = bp.fromEmail
	reqBody.To = []struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}{
		{Email: to},
	}
	reqBody.HtmlContent = htmlBody
	reqBody.TextContent = textBody

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Brevo API
	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", bp.apiKey)
	req.Header.Set("content-type", "application/json")

	// Log the full request for debugging
	log.Printf("🔍 Brevo API Request:")
	log.Printf("   URL: %s", req.URL.String())
	log.Printf("   Method: %s", req.Method)
	log.Printf("   API Key (full): '%s'", bp.apiKey)
	log.Printf("   From: %s <%s>", bp.fromName, bp.fromEmail)
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)
	log.Printf("   Request Body: %s", string(jsonData))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Brevo request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for logging
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📩 Brevo Response Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("📩 Brevo Response Body: %s", string(respBody))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("%d %s: %v", resp.StatusCode, resp.Status, errResp)
	}

	return nil
}

// SendFileUploadNotification skickar notifiering när fil laddats upp via request
func (bp *BrevoProvider) SendFileUploadNotification(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string, recipientEmail string) error {
	subject := "Ny fil uppladdad: " + request.Title
	htmlBody := GenerateUploadNotificationHTML(request, file, uploaderIP, serverURL)
	textBody := GenerateUploadNotificationText(request, file, uploaderIP, serverURL)

	return bp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendFileDownloadNotification skickar notifiering när fil laddas ner
func (bp *BrevoProvider) SendFileDownloadNotification(file *database.FileInfo, downloaderIP, serverURL string, recipientEmail string) error {
	subject := "Din fil har laddats ner: " + file.Name
	htmlBody := GenerateDownloadNotificationHTML(file, downloaderIP, serverURL)
	textBody := GenerateDownloadNotificationText(file, downloaderIP, serverURL)

	return bp.SendEmail(recipientEmail, subject, htmlBody, textBody)
}

// SendSplashLinkEmail skickar splash link via e-post
func (bp *BrevoProvider) SendSplashLinkEmail(to, splashLink string, file *database.FileInfo, message string) error {
	subject := "Delad fil: " + file.Name
	htmlBody := GenerateSplashLinkHTML(splashLink, file, message)
	textBody := GenerateSplashLinkText(splashLink, file, message)

	return bp.SendEmail(to, subject, htmlBody, textBody)
}

// SendAccountDeletionConfirmation skickar bekräftelse på kontoradering (GDPR)
func (bp *BrevoProvider) SendAccountDeletionConfirmation(to, accountName string) error {
	subject := "Bekräftelse: Ditt konto har raderats"
	htmlBody := GenerateAccountDeletionHTML(accountName)
	textBody := GenerateAccountDeletionText(accountName)

	return bp.SendEmail(to, subject, htmlBody, textBody)
}
