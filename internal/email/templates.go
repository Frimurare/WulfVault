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

// ---------------------------------------------------------------------------
// Prudencia brand constants
// ---------------------------------------------------------------------------

const (
	brandPrimary   = "#004155" // Dark teal — headers, CTA buttons
	brandSecondary = "#5f828c" // Muted teal — accents, borders, secondary text
	brandDark      = "#323e48" // Dark charcoal — card backgrounds, footer
	brandWhite     = "#ffffff"
	brandLightBg   = "#f4f7f8" // Very light teal-grey for page background
	brandCardBg    = "#ffffff"
	brandTextDark  = "#1a2a32"
	brandTextMuted = "#5f828c"
	brandSuccess   = "#0d9488" // Teal-green for success states
	brandWarning   = "#d97706" // Amber for warnings
	brandDanger    = "#b91c1c" // Red for danger/deletion
)

// ServerURL is the public base URL of the WulfVault instance.
// Set at startup (e.g. "https://wulfvault.prudcloud.se").
// Used by email templates to build absolute image URLs.
var ServerURL string

// prudenciaLogoPath is the path to the hosted Prudencia logotype (white on transparent).
// Hosted as a static file on the WulfVault server so all email clients can
// render it reliably (base64 data URIs get stripped by Microsoft Graph relay).
var prudenciaLogoPath = "/static/img/prudencia_logo.png"

// prudenciaLogoImg returns an <img> tag referencing the hosted Prudencia logotype.
// serverURL is the public base URL (e.g. "https://wulfvault.prudcloud.se").
func prudenciaLogoImg() string {
	src := prudenciaLogoPath
	if ServerURL != "" {
		src = ServerURL + prudenciaLogoPath
	}
	return fmt.Sprintf(`<img src="%s" alt="Prudencia Security" width="200" height="117" style="display:block;margin:0 auto 15px auto;" />`, src)
}

// emailHeader returns the standard Prudencia email header
func emailHeader(title, subtitle string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;font-family:Arial,Helvetica,sans-serif;background-color:%s;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:%s;padding:30px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background-color:%s;border-radius:12px;overflow:hidden;box-shadow:0 8px 30px rgba(0,0,0,0.12);">
<!-- Header -->
<tr><td style="background-color:%s;padding:35px 30px 25px 30px;text-align:center;">
%s
<h1 style="color:%s;margin:0;font-size:22px;font-weight:700;letter-spacing:0.5px;">%s</h1>
<p style="color:%s;margin:8px 0 0 0;font-size:13px;letter-spacing:0.5px;">%s</p>
</td></tr>`,
		brandLightBg, brandLightBg, brandCardBg,
		brandPrimary, prudenciaLogoImg(),
		brandWhite, title,
		brandSecondary, subtitle)
}

// emailFooter returns the standard Prudencia email footer
func emailFooter() string {
	return fmt.Sprintf(`<!-- Footer -->
<tr><td style="background-color:%s;padding:25px 30px;text-align:center;">
<p style="margin:0 0 5px 0;color:%s;font-size:11px;letter-spacing:0.5px;">Prudencia Security &middot; Secure File Transfer</p>
<p style="margin:0;color:%s;font-size:10px;opacity:0.7;">This is an automated message. Do not reply to this email.</p>
</td></tr>
</table></td></tr></table></body></html>`, brandDark, brandSecondary, brandSecondary)
}

// ctaButton returns a styled CTA button
func ctaButton(text, href, bgColor string) string {
	if bgColor == "" {
		bgColor = brandPrimary
	}
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" style="margin:25px 0;">
<tr><td align="center">
<a href="%s" style="display:inline-block;background-color:%s;color:#ffffff;padding:14px 40px;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;letter-spacing:0.5px;box-shadow:0 4px 12px rgba(0,65,85,0.3);">%s</a>
</td></tr></table>`, href, bgColor, text)
}

// infoRow returns a key-value row for file info tables
func infoRow(label, value string) string {
	return fmt.Sprintf(`<tr>
<td style="padding:10px 15px;color:%s;font-size:13px;font-weight:600;white-space:nowrap;border-bottom:1px solid #eef2f3;">%s</td>
<td style="padding:10px 15px;color:%s;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td>
</tr>`, brandSecondary, label, brandTextDark, value)
}

// fileInfoTable returns a styled file info card
func fileInfoTable(rows string) string {
	return fmt.Sprintf(`<div style="background:%s;border:1px solid #dce4e6;border-left:4px solid %s;border-radius:8px;overflow:hidden;margin:20px 0;">
<table width="100%%" cellpadding="0" cellspacing="0">%s</table>
</div>`, brandLightBg, brandSecondary, rows)
}

// statusBanner returns a colored banner for status messages
func statusBanner(text, bgColor, textColor, borderColor string) string {
	return fmt.Sprintf(`<div style="background-color:%s;border-left:4px solid %s;padding:16px 20px;margin:20px 0;border-radius:6px;">
<p style="margin:0;color:%s;font-size:14px;">%s</p>
</div>`, bgColor, borderColor, textColor, text)
}

// ---------------------------------------------------------------------------
// Upload notification (file uploaded via file request)
// ---------------------------------------------------------------------------

func GenerateUploadNotificationHTML(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

	return emailHeader("New File Uploaded", "Upload Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
<p style="text-align:center;color:%s;font-size:12px;margin-top:5px;">Or go directly to: <code style="background:#eef2f3;padding:2px 6px;border-radius:3px;font-size:11px;">%s/dashboard</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Good news!</strong> Someone has uploaded a file via your request: <strong>%s</strong>`, request.Title),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size)+
					infoRow("Uploaded", uploadTime)+
					infoRow("IP Address", uploaderIP),
			),
			ctaButton("VIEW IN DASHBOARD", serverURL+"/dashboard", ""),
			brandTextMuted, serverURL,
		) +
		emailFooter()
}

func GenerateUploadNotificationText(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`New File Uploaded

Someone has uploaded a file via your upload request:

Request: %s
Filename: %s
Size: %s
Uploaded: %s
IP Address: %s

Log in to view and download the file:
%s/dashboard

---
Prudencia Security — Secure File Transfer
`, request.Title, file.Name, file.Size, uploadTime, uploaderIP, serverURL)
}

// ---------------------------------------------------------------------------
// Download notification (someone downloaded your file)
// ---------------------------------------------------------------------------

func GenerateDownloadNotificationHTML(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")

	return emailHeader("File Downloaded", "Download Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
</td></tr>`,
			statusBanner(
				`<strong>Good news!</strong> Someone has downloaded your file. Here are the details:`,
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size)+
					infoRow("Downloaded", downloadTime)+
					infoRow("IP Address", downloaderIP)+
					infoRow("Downloads remaining", getDownloadsRemainingText(file)),
			),
			ctaButton("VIEW IN DASHBOARD", serverURL+"/dashboard", ""),
		) +
		emailFooter()
}

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
Prudencia Security — Secure File Transfer
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// ---------------------------------------------------------------------------
// Splash link email (file shared with someone — THE main customer-facing email)
// ---------------------------------------------------------------------------

func GenerateSplashLinkHTML(splashLink string, file *database.FileInfo, message, senderEmail string) string {
	senderBlock := ""
	if senderEmail != "" {
		senderBlock = fmt.Sprintf(`<div style="background:%s;border-radius:8px;padding:14px 20px;margin:0 0 20px 0;">
<p style="margin:0;color:%s;font-size:13px;"><strong style="color:%s;">Sent by:</strong> %s</p>
</div>`, brandLightBg, brandTextDark, brandSecondary, senderEmail)
	}

	messageBlock := ""
	if message != "" {
		messageBlock = fmt.Sprintf(`<div style="background:#fefce8;border-left:4px solid %s;padding:16px 20px;margin:0 0 20px 0;border-radius:6px;">
<p style="margin:0 0 5px 0;color:%s;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:1px;">Message</p>
<p style="margin:0;color:%s;font-size:14px;">%s</p>
</div>`, brandWarning, brandSecondary, brandTextDark, message)
	}

	return emailHeader("A file has been shared with you", "Secure File Transfer") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:15px;">Or copy this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:11px;word-break:break-all;color:%s;">%s</code></p>
</td></tr>`,
			senderBlock,
			messageBlock,
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size),
			),
			ctaButton("DOWNLOAD FILE", splashLink, brandSuccess),
			brandTextMuted, brandTextDark, splashLink,
		) +
		emailFooter()
}

func GenerateSplashLinkText(splashLink string, file *database.FileInfo, message, senderEmail string) string {
	senderLine := ""
	if senderEmail != "" {
		senderLine = "Sent by: " + senderEmail + "\n"
	}
	return fmt.Sprintf(`A file has been shared with you

%s%s
Filename: %s
Size: %s

Download: %s

---
Prudencia Security — Secure File Transfer
`, senderLine, getMessageText(message), file.Name, file.Size, splashLink)
}

// ---------------------------------------------------------------------------
// Account deletion (GDPR)
// ---------------------------------------------------------------------------

func GenerateAccountDeletionHTML(accountName string) string {
	return emailHeader("Account Deleted", "GDPR Compliance") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
<div style="text-align:center;margin:0 0 20px 0;">
<div style="width:60px;height:60px;background:#e6f5f0;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;font-size:28px;line-height:60px;">&#10003;</div>
</div>
<p style="color:%s;font-size:15px;">Hi %s,</p>
<p style="color:%s;font-size:14px;">This is a confirmation that your download account has been deleted from our system in accordance with GDPR.</p>
%s
<p style="color:%s;font-size:14px;">We respect your right to erasure under GDPR and confirm that all your personal information has been handled in accordance with data protection regulations.</p>
</td></tr>`,
			brandTextDark, accountName,
			brandTextDark,
			statusBanner(
				`<strong>What happened:</strong><br/>
&bull; Your personal information has been permanently anonymized<br/>
&bull; You can no longer download files with this account<br/>
&bull; To download files again you must register a new account`,
				"#fef2f2", brandDanger, brandDanger,
			),
			brandTextDark,
		) +
		emailFooter()
}

func GenerateAccountDeletionText(accountName string) string {
	return fmt.Sprintf(`Account Deleted — GDPR Compliance

Hi %s,

This is a confirmation that your download account has been deleted from our system in accordance with GDPR.

What happened:
- Your personal information has been permanently anonymized
- You can no longer download files with this account
- To download files again you must register a new account

We respect your right to erasure under GDPR.

---
Prudencia Security — Secure File Transfer
`, accountName)
}

// ---------------------------------------------------------------------------
// Welcome email (new user account)
// ---------------------------------------------------------------------------

func SendWelcomeEmail(email, resetToken, serverURL, companyName, adminName, adminEmail string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)
	subject := fmt.Sprintf("Welcome to %s — Set Your Password", companyName)

	htmlBody := emailHeader(fmt.Sprintf("Welcome to %s!", companyName), "Your account has been created") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<p style="color:%s;font-size:14px;">To get started, you need to set your password and log in to your account.</p>
<div style="background:%s;border:2px solid %s;border-radius:10px;padding:25px;margin:25px 0;text-align:center;">
<h2 style="color:%s;margin:0 0 10px 0;font-size:18px;">Set Your Password</h2>
<p style="color:%s;margin:0 0 20px 0;font-size:14px;">Click the button below to create your password and access your account.</p>
%s
<p style="font-size:12px;color:%s;margin:15px 0 0 0;">This link is valid for 1 hour</p>
</div>
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">If the button doesn't work, copy and paste this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:10px;word-break:break-all;">%s</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Congratulations!</strong> <strong>%s</strong> (%s) has added you to <strong>%s</strong>. You can now share, receive, and request files securely.`, adminName, adminEmail, companyName),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			brandTextDark,
			brandCardBg, brandSecondary,
			brandPrimary, brandTextDark,
			ctaButton("SET PASSWORD &amp; LOGIN", resetLink, ""),
			brandTextMuted,
			fileInfoTable(infoRow("Your Login Email", email)),
			brandTextMuted, resetLink,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Welcome to %s!

%s (%s) has added you to %s. You can now share, receive, and request files securely.

Your login email: %s

Set your password by visiting this link:
%s

This link is valid for 1 hour.

---
Prudencia Security — Secure File Transfer
`, companyName, adminName, adminEmail, companyName, email, resetLink)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Team invitation email
// ---------------------------------------------------------------------------

func SendTeamInvitationEmail(email, teamName, serverURL, companyName string) error {
	subject := fmt.Sprintf("Welcome to team %s — %s", teamName, companyName)

	htmlBody := emailHeader(fmt.Sprintf("Welcome to Team: %s", teamName), "You've been added to a collaborative team") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<p style="color:%s;font-size:14px;">As a team member, you can now:</p>
<div style="margin:15px 0 25px 0;padding:0 0 0 10px;">
<p style="color:%s;font-size:14px;margin:8px 0;line-height:1.6;">&#128193; Access all files shared with the team<br/>
&#11014;&#65039; Upload files to share with team members<br/>
&#128101; Collaborate with other team members<br/>
&#128274; Securely transfer files within your team</p>
</div>
%s
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">Or go directly to: <code style="background:#eef2f3;padding:2px 6px;border-radius:3px;font-size:10px;">%s/login</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Congratulations!</strong> You have been added to the team <strong>"%s"</strong> in the <strong>%s</strong> file sharing platform.`, teamName, companyName),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			brandTextDark,
			brandTextDark,
			fileInfoTable(infoRow("Your Login Email", email)),
			ctaButton("LOG IN TO YOUR TEAM", serverURL+"/login", ""),
			brandTextMuted, serverURL,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Welcome to team %s — %s

You have been added to the team "%s" in the %s file sharing platform.

As a team member, you can now:
- Access all files shared with the team
- Upload files to share with team members
- Collaborate with other team members
- Securely transfer files within your team

Your login email: %s

Log in here: %s/login

---
Prudencia Security — Secure File Transfer
`, teamName, companyName, teamName, companyName, email, serverURL)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Password reset email
// ---------------------------------------------------------------------------

func SendPasswordResetEmail(email, resetToken, serverURL string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)
	subject := "Password Reset Request"

	htmlBody := emailHeader("Password Reset", "Security Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<div style="background:%s;border:2px solid %s;border-radius:10px;padding:25px;margin:25px 0;text-align:center;">
<h2 style="color:%s;margin:0 0 10px 0;font-size:18px;">Reset Your Password</h2>
<p style="color:%s;margin:0 0 20px 0;font-size:14px;">Click the button below to create a new password.</p>
%s
<p style="font-size:12px;color:%s;margin:15px 0 0 0;">This link is valid for 1 hour</p>
</div>
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">If the button doesn't work, copy and paste this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:10px;word-break:break-all;">%s</code></p>
</td></tr>`,
			statusBanner(
				`A password reset has been requested for your account. If you did not request this, please ignore this email.`,
				"#fefce8", brandWarning, brandWarning,
			),
			brandCardBg, brandSecondary,
			brandPrimary, brandTextDark,
			ctaButton("RESET PASSWORD", resetLink, ""),
			brandTextMuted,
			statusBanner(
				`<strong>Security reminder:</strong><br/>
&bull; Never share this link with anyone<br/>
&bull; We will never ask for your password via email<br/>
&bull; If you didn't request this reset, ignore this email`,
				"#fef2f2", brandDanger, brandDanger,
			),
			brandTextMuted, resetLink,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Password Reset Request

A password reset has been requested for your account.

Click the link below to reset your password:
%s

This link is valid for 1 hour.

Security reminder:
- Never share this link with anyone
- We will never ask for your password via email
- If you didn't request this reset, ignore this email

---
Prudencia Security — Secure File Transfer
`, resetLink)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Expiration reminder email (used by cleanup/reminders.go)
// ---------------------------------------------------------------------------

// GenerateExpirationReminderHTML creates a branded expiration reminder email
func GenerateExpirationReminderHTML(fileName, fileSize, splashLink string, daysLeft int, urgent bool) string {
	urgencyText := fmt.Sprintf("Your file will expire in <strong>%d days</strong>.", daysLeft)
	bannerBg := "#fefce8"
	bannerBorder := brandWarning
	bannerText := brandTextDark
	if urgent {
		urgencyText = fmt.Sprintf("<strong>URGENT:</strong> Your file expires <strong>tomorrow</strong>!")
		bannerBg = "#fef2f2"
		bannerBorder = brandDanger
		bannerText = brandDanger
	}

	return emailHeader("File Expiration Reminder", "Action Required") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
<p style="text-align:center;color:%s;font-size:12px;margin-top:10px;">Download or re-share the file before it expires.</p>
</td></tr>`,
			statusBanner(urgencyText, bannerBg, bannerText, bannerBorder),
			fileInfoTable(
				infoRow("Filename", fileName)+
					infoRow("Size", fileSize)+
					infoRow("Expires in", fmt.Sprintf("%d day(s)", daysLeft)),
			),
			ctaButton("VIEW FILE", splashLink, brandPrimary),
			brandTextMuted,
		) +
		emailFooter()
}

// GenerateExpirationReminderText creates a plain text expiration reminder
func GenerateExpirationReminderText(fileName, fileSize, splashLink string, daysLeft int, urgent bool) string {
	prefix := ""
	if urgent {
		prefix = "URGENT: "
	}
	return fmt.Sprintf(`%sFile Expiration Reminder

Your file will expire in %d day(s):

Filename: %s
Size: %s

Download or re-share the file before it expires:
%s

---
Prudencia Security — Secure File Transfer
`, prefix, daysLeft, fileName, fileSize, splashLink)
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func getDownloadsRemainingText(file *database.FileInfo) string {
	if file.UnlimitedDownloads {
		return "Unlimited"
	}
	if file.DownloadsRemaining <= 0 {
		return "0 (no more downloads available)"
	}
	return fmt.Sprintf("%d", file.DownloadsRemaining)
}

func getMessageText(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf("Message: %s\n\n", message)
}
