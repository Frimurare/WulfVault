package email

import (
	"context"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

// BrevoProvider implementerar EmailProvider för Brevo (Sendinblue)
type BrevoProvider struct {
	client    *sendinblue.APIClient
	fromEmail string
	fromName  string
}

// NewBrevoProvider skapar en ny Brevo provider
func NewBrevoProvider(apiKey, fromEmail, fromName string) *BrevoProvider {
	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)

	if fromName == "" {
		fromName = "Sharecare"
	}

	return &BrevoProvider{
		client:    sendinblue.NewAPIClient(cfg),
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendEmail skickar ett e-postmeddelande via Brevo
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
