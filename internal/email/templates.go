package email

import (
	"fmt"
	"time"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

// GenerateUploadNotificationHTML skapar HTML-version av uppladdningsnotifiering
func GenerateUploadNotificationHTML(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #2563eb; color: white; padding: 20px; border-radius: 5px 5px 0 0; text-align: center; }
		.header h2 { margin: 0; }
		.content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
		.file-info { background: white; padding: 15px; margin: 15px 0; border-left: 4px solid #2563eb; }
		.file-info p { margin: 5px 0; }
		.button {
			display: inline-block;
			padding: 12px 24px;
			background: #28a745;
			color: white;
			text-decoration: none;
			border-radius: 5px;
			margin: 20px 0;
		}
		.footer { margin-top: 20px; font-size: 12px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>✓ New File Uploaded</h2>
		</div>
		<div class="content">
			<p>Someone has uploaded a file via your upload request:</p>

			<div class="file-info">
				<p><strong>Request:</strong> %s</p>
				<p><strong>Filename:</strong> %s</p>
				<p><strong>Size:</strong> %s</p>
				<p><strong>Uploaded:</strong> %s</p>
				<p><strong>IP Address:</strong> %s</p>
			</div>

			<a href="%s/dashboard" class="button">View in Dashboard</a>

			<div class="footer">
				<p>The file is now available in your dashboard and can be downloaded.</p>
				<p>This is an automated message from Sharecare.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, request.Title, file.Name, file.Size, uploadTime, uploaderIP, serverURL)
}

// GenerateUploadNotificationText skapar text-version av uppladdningsnotifiering
func GenerateUploadNotificationText(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`New File Uploaded!

Someone has uploaded a file via your upload request:

Request: %s
Filename: %s
Size: %s
Uploaded: %s
IP Address: %s

Log in to view and download the file:
%s/dashboard

---
This is an automated message from Sharecare.
`, request.Title, file.Name, file.Size, uploadTime, uploaderIP, serverURL)
}

// GenerateDownloadNotificationHTML skapar HTML-version av nedladdningsnotifiering
func GenerateDownloadNotificationHTML(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #2563eb; color: white; padding: 20px; border-radius: 5px 5px 0 0; text-align: center; }
		.header h2 { margin: 0; }
		.content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
		.file-info { background: white; padding: 15px; margin: 15px 0; border-left: 4px solid #2563eb; }
		.file-info p { margin: 5px 0; }
		.button {
			display: inline-block;
			padding: 12px 24px;
			background: #2563eb;
			color: white;
			text-decoration: none;
			border-radius: 5px;
			margin: 20px 0;
		}
		.footer { margin-top: 20px; font-size: 12px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>⬇️ Din fil har laddats ner</h2>
		</div>
		<div class="content">
			<p>Någon har laddat ner en av dina filer:</p>

			<div class="file-info">
				<p><strong>Filnamn:</strong> %s</p>
				<p><strong>Storlek:</strong> %s</p>
				<p><strong>Nedladdad:</strong> %s</p>
				<p><strong>IP-adress:</strong> %s</p>
				<p><strong>Nedladdningar kvar:</strong> %s</p>
			</div>

			<a href="%s/dashboard" class="button">Visa i Dashboard</a>

			<div class="footer">
				<p>Detta är ett automatiskt meddelande från Sharecare.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// GenerateDownloadNotificationText skapar text-version av nedladdningsnotifiering
func GenerateDownloadNotificationText(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`Din fil har laddats ner!

Någon har laddat ner en av dina filer:

Filnamn: %s
Storlek: %s
Nedladdad: %s
IP-adress: %s
Nedladdningar kvar: %s

Logga in för att se detaljer:
%s/dashboard

---
Detta är ett automatiskt meddelande från Sharecare.
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// GenerateSplashLinkHTML skapar HTML-version av splash link e-post
func GenerateSplashLinkHTML(splashLink string, file *database.FileInfo, message string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #2563eb; color: white; padding: 20px; border-radius: 5px 5px 0 0; text-align: center; }
		.header h2 { margin: 0; }
		.content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
		.message-box { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 15px 0; }
		.file-info { background: white; padding: 15px; margin: 15px 0; border-left: 4px solid #2563eb; }
		.file-info p { margin: 5px 0; }
		.button {
			display: inline-block;
			padding: 12px 24px;
			background: #28a745;
			color: white !important;
			text-decoration: none;
			border-radius: 5px;
			margin: 20px 0;
			font-weight: bold;
		}
		.link-text { font-size: 12px; color: #666; word-break: break-all; margin-top: 10px; }
		.footer { margin-top: 20px; font-size: 12px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>📎 Någon har delat en fil med dig</h2>
		</div>
		<div class="content">
			%s

			<div class="file-info">
				<p><strong>Filnamn:</strong> %s</p>
				<p><strong>Storlek:</strong> %s</p>
			</div>

			<center>
				<a href="%s" class="button">📥 Ladda ner fil</a>
			</center>

			<div class="link-text">
				Eller kopiera denna länk:<br/>
				<code>%s</code>
			</div>

			<div class="footer">
				<p>Detta är ett automatiskt meddelande från Sharecare.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, getMessageHTML(message), file.Name, file.Size, splashLink, splashLink)
}

// GenerateSplashLinkText skapar text-version av splash link e-post
func GenerateSplashLinkText(splashLink string, file *database.FileInfo, message string) string {
	return fmt.Sprintf(`Någon har delat en fil med dig

%s
Filnamn: %s
Storlek: %s

Ladda ner filen här: %s

---
Detta är ett automatiskt meddelande från Sharecare.
`, getMessageText(message), file.Name, file.Size, splashLink)
}

// Helper-funktioner

func getDownloadsRemainingText(file *database.FileInfo) string {
	if file.UnlimitedDownloads {
		return "Obegränsat"
	}
	if file.DownloadsRemaining <= 0 {
		return "0 (ingen kan ladda ner filen längre)"
	}
	return fmt.Sprintf("%d", file.DownloadsRemaining)
}

func getMessageHTML(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="message-box"><strong>Meddelande:</strong><br/>%s</div>`, message)
}

func getMessageText(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf("Meddelande: %s\n\n", message)
}

// GenerateAccountDeletionHTML skapar HTML-version av bekräftelse på kontoradering
func GenerateAccountDeletionHTML(accountName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #c53030; color: white; padding: 20px; border-radius: 5px 5px 0 0; text-align: center; }
		.header h2 { margin: 0; }
		.content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
		.info-box { background: #fff5f5; border-left: 4px solid #c53030; padding: 15px; margin: 15px 0; }
		.info-box p { margin: 5px 0; }
		.footer { margin-top: 20px; font-size: 12px; color: #666; text-align: center; }
		.checkmark {
			width: 60px;
			height: 60px;
			background: #d4edda;
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			margin: 20px auto;
			font-size: 32px;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>✓ Ditt konto har raderats</h2>
		</div>
		<div class="content">
			<div class="checkmark">✓</div>

			<p>Hej %s,</p>

			<p>Detta är en bekräftelse på att ditt nedladdningskonto har raderats från vårt system enligt GDPR.</p>

			<div class="info-box">
				<p><strong>Vad har hänt:</strong></p>
				<ul>
					<li>Din personliga information har anonymiserats permanent</li>
					<li>Du kan inte längre ladda ner filer med detta konto</li>
					<li>Om du vill ladda ner filer igen måste du registrera ett nytt konto</li>
				</ul>
			</div>

			<p>Vi respekterar din rätt till radering enligt GDPR och bekräftar att all din personliga information har hanterats i enlighet med dataskyddsförordningen.</p>

			<div class="footer">
				<p>Detta är ett automatiskt meddelande från Sharecare.</p>
				<p>Om du har frågor, vänligen kontakta oss.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, accountName)
}

// GenerateAccountDeletionText skapar text-version av bekräftelse på kontoradering
func GenerateAccountDeletionText(accountName string) string {
	return fmt.Sprintf(`Ditt konto har raderats

Hej %s,

Detta är en bekräftelse på att ditt nedladdningskonto har raderats från vårt system enligt GDPR.

Vad har hänt:
- Din personliga information har anonymiserats permanent
- Du kan inte längre ladda ner filer med detta konto
- Om du vill ladda ner filer igen måste du registrera ett nytt konto

Vi respekterar din rätt till radering enligt GDPR och bekräftar att all din personliga information har hanterats i enlighet med dataskyddsförordningen.

---
Detta är ett automatiskt meddelande från Sharecare.
Om du har frågor, vänligen kontakta oss.
`, accountName)
}
