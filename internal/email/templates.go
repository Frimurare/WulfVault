// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"fmt"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
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
				<p>This is an automated message from WulfVault.</p>
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
This is an automated message from WulfVault.
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
</head>
<body style="margin: 0; padding: 0; font-family: Arial, Helvetica, sans-serif;">
	<table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f0f0f0; padding: 20px 0;">
		<tr>
			<td align="center">
				<table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
					<!-- Header -->
					<tr>
						<td style="background-color: #1e3a5f; padding: 30px; text-align: center;">
							<h1 style="color: #ffffff; margin: 0; font-size: 24px;">⬇️ File Downloaded</h1>
							<p style="color: #a0c4e8; margin: 10px 0 0 0; font-size: 14px;">Download Notification</p>
						</td>
					</tr>

					<!-- Main Content -->
					<tr>
						<td style="padding: 40px 30px;">
							<!-- What is this -->
							<div style="background-color: #d1fae5; border-left: 4px solid #10b981; padding: 15px; margin-bottom: 25px;">
								<p style="margin: 0; color: #065f46; font-size: 16px;">
									<strong>Good news!</strong><br>
									Someone has downloaded your file. Here are the details:
								</p>
							</div>

							<!-- File Info Box -->
							<div style="background-color: #f8fafc; border: 2px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 20px;">
								<h3 style="margin: 0 0 15px 0; color: #1e3a5f; font-size: 18px;">📄 %s</h3>
								<table width="100%%" cellpadding="0" cellspacing="0">
									<tr>
										<td style="padding: 8px 0; color: #64748b; font-size: 14px;"><strong>Size:</strong></td>
										<td style="padding: 8px 0; color: #334155; font-size: 14px;">%s</td>
									</tr>
									<tr>
										<td style="padding: 8px 0; color: #64748b; font-size: 14px;"><strong>Downloaded:</strong></td>
										<td style="padding: 8px 0; color: #334155; font-size: 14px;">%s</td>
									</tr>
									<tr>
										<td style="padding: 8px 0; color: #64748b; font-size: 14px;"><strong>IP Address:</strong></td>
										<td style="padding: 8px 0; color: #334155; font-size: 14px;">%s</td>
									</tr>
									<tr>
										<td style="padding: 8px 0; color: #64748b; font-size: 14px;"><strong>Downloads remaining:</strong></td>
										<td style="padding: 8px 0; color: #334155; font-size: 14px;">%s</td>
									</tr>
								</table>
							</div>

							<!-- Dashboard Button -->
							<table width="100%%" cellpadding="0" cellspacing="0" style="margin: 30px 0;">
								<tr>
									<td align="center">
										<a href="%s/dashboard" style="display: inline-block; background-color: #2563eb; color: #ffffff; padding: 16px 40px; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: bold; border: 3px solid #1d4ed8; box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4);">
											VIEW IN DASHBOARD
										</a>
									</td>
								</tr>
							</table>
						</td>
					</tr>

					<!-- Footer -->
					<tr>
						<td style="background-color: #1e3a5f; padding: 20px; text-align: center;">
							<p style="margin: 0; color: #a0c4e8; font-size: 12px;">
								This is an automated download notification from WulfVault
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// GenerateDownloadNotificationText creates text version of download notification
func GenerateDownloadNotificationText(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`Your file has been downloaded!

Someone has downloaded one of your files:

Filename: %s
Size: %s
Downloaded: %s
IP Address: %s
Downloads remaining: %s

Log in to see details:
%s/dashboard

---
This is an automated message from WulfVault.
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// GenerateSplashLinkHTML creates HTML version of splash link email
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
			<h2>📎 Someone Shared a File with You</h2>
		</div>
		<div class="content">
			%s

			<div class="file-info">
				<p><strong>Filename:</strong> %s</p>
				<p><strong>Size:</strong> %s</p>
			</div>

			<center>
				<a href="%s" class="button">📥 Download File</a>
			</center>

			<div class="link-text">
				Or copy this link:<br/>
				<code>%s</code>
			</div>

			<div class="footer">
				<p>This is an automated message from WulfVault.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, getMessageHTML(message), file.Name, file.Size, splashLink, splashLink)
}

// GenerateSplashLinkText creates text version of splash link email
func GenerateSplashLinkText(splashLink string, file *database.FileInfo, message string) string {
	return fmt.Sprintf(`Someone Shared a File with You

%s
Filename: %s
Size: %s

Download the file here: %s

---
This is an automated message from WulfVault.
`, getMessageText(message), file.Name, file.Size, splashLink)
}

// Helper functions

func getDownloadsRemainingText(file *database.FileInfo) string {
	if file.UnlimitedDownloads {
		return "Unlimited"
	}
	if file.DownloadsRemaining <= 0 {
		return "0 (no one can download the file anymore)"
	}
	return fmt.Sprintf("%d", file.DownloadsRemaining)
}

func getMessageHTML(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="message-box"><strong>Message:</strong><br/>%s</div>`, message)
}

func getMessageText(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf("Message: %s\n\n", message)
}

// GenerateAccountDeletionHTML creates HTML version of account deletion confirmation
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
			<h2>✓ Your Account Has Been Deleted</h2>
		</div>
		<div class="content">
			<div class="checkmark">✓</div>

			<p>Hello %s,</p>

			<p>This is confirmation that your download account has been deleted from our system in accordance with GDPR.</p>

			<div class="info-box">
				<p><strong>What happened:</strong></p>
				<ul>
					<li>Your personal information has been permanently anonymized</li>
					<li>You can no longer download files with this account</li>
					<li>If you want to download files again, you must register a new account</li>
				</ul>
			</div>

			<p>We respect your right to deletion under GDPR and confirm that all your personal information has been handled in accordance with data protection regulations.</p>

			<div class="footer">
				<p>This is an automated message from WulfVault.</p>
				<p>If you have questions, please contact us.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, accountName)
}

// GenerateAccountDeletionText creates text version of account deletion confirmation
func GenerateAccountDeletionText(accountName string) string {
	return fmt.Sprintf(`Your Account Has Been Deleted

Hello %s,

This is confirmation that your download account has been deleted from our system in accordance with GDPR.

What happened:
- Your personal information has been permanently anonymized
- You can no longer download files with this account
- If you want to download files again, you must register a new account

We respect your right to deletion under GDPR and confirm that all your personal information has been handled in accordance with data protection regulations.

---
This is an automated message from WulfVault.
If you have questions, please contact us.
`, accountName)
}

// SendWelcomeEmail sends a welcome email to newly created users with password setup link
func SendWelcomeEmail(email, resetToken, serverURL, companyName, adminName, adminEmail string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)

	subject := fmt.Sprintf("Welcome to %s - Set Your Password", companyName)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header {
			background: #2563eb;
			color: white;
			padding: 35px;
			border-radius: 10px 10px 0 0;
			text-align: center;
		}
		.header h1 { margin: 0; font-size: 32px; }
		.header p { margin: 10px 0 0 0; opacity: 0.9; }
		.content {
			background: #f9f9f9;
			padding: 30px;
			border-radius: 0 0 10px 10px;
		}
		.welcome-box {
			background: #d4edda;
			border-left: 4px solid #28a745;
			padding: 20px;
			margin: 20px 0;
			border-radius: 5px;
		}
		.welcome-box h2 {
			color: #155724;
			margin-top: 0;
		}
		.setup-box {
			background: white;
			padding: 30px;
			margin: 25px 0;
			border-radius: 8px;
			border: 2px solid #2563eb;
			text-align: center;
		}
		.button {
			display: inline-block;
			padding: 18px 50px;
			background: #2563eb;
			color: white !important;
			text-decoration: none;
			border-radius: 8px;
			margin: 20px 0;
			font-weight: bold;
			font-size: 18px;
		}
		.footer {
			margin-top: 30px;
			padding-top: 20px;
			border-top: 2px solid #ddd;
			font-size: 12px;
			color: #666;
			text-align: center;
		}
		.info-box {
			background: #e3f2fd;
			padding: 15px;
			margin: 20px 0;
			border-radius: 5px;
			border-left: 4px solid #2196f3;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🎉 Welcome to %s!</h1>
			<p>Your account has been created</p>
		</div>

		<div class="content">
			<div class="welcome-box">
				<h2>Congratulations!</h2>
				<p><strong>%s</strong> (%s) has added you to <strong>%s</strong>. You can now share, receive, and request both small and huge files securely.</p>
			</div>

			<p>To get started, you need to set your password and log in to your account.</p>

			<div class="setup-box">
				<h2 style="color: #2563eb; margin-bottom: 15px;">Set Your Password</h2>
				<p style="margin-bottom: 25px;">Click the button below to create your password and access your account.</p>

				<a href="%s" class="button">SET PASSWORD &amp; LOGIN</a>

				<p style="font-size: 13px; color: #999; margin-top: 20px;">
					This link is valid for 1 hour
				</p>
			</div>

			<div class="info-box">
				<p style="margin: 0;"><strong>📧 Your Login Email:</strong></p>
				<p style="margin: 5px 0 0 0; font-size: 16px; font-weight: bold;">%s</p>
			</div>

			<p style="text-align: center; color: #666; margin-top: 30px;">
				If the button doesn't work, copy and paste this link into your browser:
			</p>
			<p style="text-align: center; word-break: break-all; font-size: 12px; color: #999;">
				%s
			</p>
		</div>

		<div class="footer">
			<p>This is an automated message from %s.</p>
			<p>Do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`, companyName, adminName, adminEmail, companyName, resetLink, email, resetLink, companyName)

	textBody := fmt.Sprintf(`Welcome to %s!

Congratulations! %s (%s) has added you to %s. You can now share, receive, and request both small and huge files securely.

To get started, you need to set your password and log in to your account.

Your login email: %s

Set your password by visiting this link:
%s

This link is valid for 1 hour.

---
This is an automated message from %s.
Do not reply to this email.`, companyName, adminName, adminEmail, companyName, email, resetLink, companyName)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}

	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// SendTeamInvitationEmail sends an invitation email when a user is added to a team
func SendTeamInvitationEmail(email, teamName, serverURL, companyName string) error {
	subject := fmt.Sprintf("Welcome to teamshare group %s in the %s fileshare", teamName, companyName)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header {
			background: #2563eb;
			color: white;
			padding: 35px;
			border-radius: 10px 10px 0 0;
			text-align: center;
		}
		.header h1 { margin: 0; font-size: 32px; }
		.header p { margin: 10px 0 0 0; opacity: 0.9; }
		.content {
			background: #f9f9f9;
			padding: 30px;
			border-radius: 0 0 10px 10px;
		}
		.team-box {
			background: #d4edda;
			border-left: 4px solid #28a745;
			padding: 20px;
			margin: 20px 0;
			border-radius: 5px;
		}
		.team-box h2 {
			color: #155724;
			margin-top: 0;
		}
		.login-box {
			background: white;
			padding: 30px;
			margin: 25px 0;
			border-radius: 8px;
			border: 2px solid #2563eb;
			text-align: center;
		}
		.button {
			display: inline-block;
			padding: 18px 50px;
			background: #2563eb;
			color: white !important;
			text-decoration: none;
			border-radius: 8px;
			margin: 20px 0;
			font-weight: bold;
			font-size: 18px;
		}
		.footer {
			margin-top: 30px;
			padding-top: 20px;
			border-top: 2px solid #ddd;
			font-size: 12px;
			color: #666;
			text-align: center;
		}
		.info-box {
			background: #e3f2fd;
			padding: 15px;
			margin: 20px 0;
			border-radius: 5px;
			border-left: 4px solid #2196f3;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🎉 Welcome to Team: %s</h1>
			<p>You've been added to a collaborative team</p>
		</div>

		<div class="content">
			<div class="team-box">
				<h2>Congratulations!</h2>
				<p>You have been added to the teamshare group <strong>"%s"</strong> in the <strong>%s</strong> fileshare platform.</p>
			</div>

			<p>As a team member, you can now:</p>
			<ul>
				<li>📁 Access all files shared with the team</li>
				<li>⬆️ Upload files to share with team members</li>
				<li>👥 Collaborate with other team members</li>
				<li>🔒 Securely transfer files within your team</li>
			</ul>

			<div class="login-box">
				<h2 style="color: #2563eb; margin-bottom: 15px;">Get Started</h2>
				<p style="margin-bottom: 25px;">Click the button below to log in and access your team workspace.</p>

				<a href="%s/login" class="button">LOG IN TO YOUR TEAM</a>
			</div>

			<div class="info-box">
				<p style="margin: 0;"><strong>📧 Your Login Email:</strong></p>
				<p style="margin: 5px 0 0 0; font-size: 16px; font-weight: bold;">%s</p>
			</div>

			<p style="text-align: center; color: #666; margin-top: 30px;">
				If the button doesn't work, copy and paste this link into your browser:
			</p>
			<p style="text-align: center; word-break: break-all; font-size: 12px; color: #999;">
				%s/login
			</p>
		</div>

		<div class="footer">
			<p>This is an automated message from %s.</p>
			<p>Do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`, teamName, teamName, companyName, serverURL, email, serverURL, companyName)

	textBody := fmt.Sprintf(`Welcome to teamshare group %s in the %s fileshare

Congratulations! You have been added to the teamshare group "%s" in the %s fileshare platform.

As a team member, you can now:
- Access all files shared with the team
- Upload files to share with team members
- Collaborate with other team members
- Securely transfer files within your team

Your login email: %s

Log in here: %s/login

---
This is an automated message from %s.
Do not reply to this email.`, teamName, companyName, teamName, companyName, email, serverURL, companyName)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}

	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// SendPasswordResetEmail sends a password reset email with a humoristic/ironic tone
func SendPasswordResetEmail(email, resetToken, serverURL string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)

	subject := "Forgot your password... again? 🤔"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header {
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			color: white;
			padding: 30px;
			border-radius: 10px 10px 0 0;
			text-align: center;
		}
		.header h1 { margin: 0; font-size: 28px; }
		.header p { margin: 10px 0 0 0; opacity: 0.9; }
		.content {
			background: #f9f9f9;
			padding: 30px;
			border-radius: 0 0 10px 10px;
		}
		.message-box {
			background: #fff3cd;
			border-left: 4px solid #ffc107;
			padding: 15px;
			margin: 20px 0;
			border-radius: 5px;
		}
		.reset-box {
			background: white;
			padding: 25px;
			margin: 25px 0;
			border-radius: 8px;
			border: 2px solid #667eea;
			text-align: center;
		}
		.reset-box h2 {
			color: #667eea;
			margin-top: 0;
		}
		.button {
			display: inline-block;
			padding: 15px 35px;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			color: white;
			text-decoration: none;
			border-radius: 25px;
			margin: 20px 0;
			font-weight: bold;
			font-size: 16px;
			transition: transform 0.2s;
		}
		.button:hover {
			transform: scale(1.05);
		}
		.tips {
			background: #e3f2fd;
			padding: 15px;
			margin: 20px 0;
			border-radius: 5px;
			border-left: 4px solid #2196f3;
		}
		.tips h3 {
			margin-top: 0;
			color: #1976d2;
		}
		.footer {
			margin-top: 30px;
			padding-top: 20px;
			border-top: 2px solid #ddd;
			font-size: 12px;
			color: #666;
			text-align: center;
		}
		.warning {
			background: #ffebee;
			border-left: 4px solid #f44336;
			padding: 15px;
			margin: 20px 0;
			border-radius: 5px;
			color: #c62828;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🔐 Reset Password</h1>
			<p>We've all been there...</p>
		</div>

		<div class="content">
			<div class="message-box">
				<p style="margin: 0;"><strong>Hey there!</strong></p>
				<p style="margin: 10px 0 0 0;">
					We received a request to reset the password for your account.
					No panic – it happens to the best of us!
					(Though maybe not <em>quite</em> as often... 😉)
				</p>
			</div>

			<div class="reset-box">
				<h2>Reset Your Password</h2>
				<p>Click the button below to create a new password.</p>
				<p style="font-size: 14px; color: #666;">
					(And maybe... write it down this time? 📝)
				</p>

				<a href="%s" class="button">Reset Password</a>

				<p style="font-size: 13px; color: #999; margin-top: 20px;">
					This link is valid for 1 hour
				</p>
			</div>

			<div class="tips">
				<h3>💡 Pro Tips for the Future:</h3>
				<ul style="margin: 10px 0; padding-left: 20px;">
					<li>Use a password manager (like LastPass, 1Password, Bitwarden)</li>
					<li>Make passwords unique for each site</li>
					<li>Think of a sentence and take the first letter of each word</li>
					<li>Or just... write it down somewhere safe? 🤷</li>
				</ul>
			</div>

			<div class="warning">
				<p style="margin: 0;"><strong>⚠️ Important information:</strong></p>
				<ul style="margin: 10px 0 0 0; padding-left: 20px;">
					<li>If you did NOT request this reset – ignore this email</li>
					<li>NEVER share this link with anyone else</li>
					<li>We will NEVER ask for your password via email</li>
				</ul>
			</div>

			<p style="text-align: center; color: #666; margin-top: 30px;">
				Button not working? Copy and paste this link into your browser:
			</p>
			<p style="text-align: center; word-break: break-all; font-size: 12px; color: #999;">
				%s
			</p>
		</div>

		<div class="footer">
			<p>This is an automated message from WulfVault.</p>
			<p>Do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`, resetLink, resetLink)

	textBody := fmt.Sprintf(`Reset Your Password

Hey!

We received a request to reset the password for your account.

Click the link below to reset your password:
%s

This link is valid for 1 hour.

If you did not request this reset, please ignore this email.

Tip: Consider using a password manager to avoid this in the future! 😊

---
This is an automated message from WulfVault.
Do not reply to this email.`, resetLink)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}

	return provider.SendEmail(email, subject, htmlBody, textBody)
}
