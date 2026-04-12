// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)

package cleanup

import (
	"fmt"
	"log"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
)

// SendExpirationReminders checks for files about to expire and sends reminder emails.
// Sends at two points: halfway through the expiration period, and 1 day before expiration.
// Reminders go to BOTH the recipient (send_to_email) and the file owner.
func SendExpirationReminders(serverURL string) {
	// Files expiring within 1 day — URGENT reminder
	urgentFiles, err := database.DB.GetFilesExpiringInDays(1)
	if err != nil {
		log.Printf("Reminder: Error getting urgent expiring files: %v", err)
	} else {
		for _, file := range urgentFiles {
			sendReminder(file, serverURL, true)
		}
	}

	// Files expiring within 3 days (for ~5 day expiration = roughly halfway)
	halfwayFiles, err := database.DB.GetFilesExpiringInDays(3)
	if err != nil {
		log.Printf("Reminder: Error getting halfway expiring files: %v", err)
	} else {
		for _, file := range halfwayFiles {
			// Don't double-send for files already in the 1-day window
			daysLeft := daysUntilExpiry(file)
			if daysLeft > 1 && daysLeft <= 3 {
				sendReminder(file, serverURL, false)
			}
		}
	}
}

func sendReminder(file *database.FileInfo, serverURL string, urgent bool) {
	daysLeft := daysUntilExpiry(file)
	if daysLeft < 0 {
		return
	}

	// Get file owner email
	owner, err := database.DB.GetUserByID(file.UserId)
	if err != nil {
		log.Printf("Reminder: Could not get owner for file %s: %v", file.Name, err)
		return
	}

	urgencyText := "Reminder"
	if urgent {
		urgencyText = "URGENT — Last day"
	}

	subject := fmt.Sprintf("[Prudencia] %s: File '%s' expires in %d day(s)", urgencyText, file.Name, daysLeft)

	splashLink := serverURL + "/s/" + file.Id
	downloadLink := serverURL + "/d/" + file.Id

	// Use Prudencia brand colors
	bannerBg := "#fefce8"
	bannerBorder := "#d97706"
	bannerTextColor := "#1a2a32"
	if urgent {
		bannerBg = "#fef2f2"
		bannerBorder = "#b91c1c"
		bannerTextColor = "#b91c1c"
	}

	urgencyMsg := fmt.Sprintf("Your file will be <strong>permanently deleted in %d day(s)</strong>.", daysLeft)
	if urgent {
		urgencyMsg = fmt.Sprintf("<strong>URGENT:</strong> Your file expires <strong>tomorrow</strong> and will be permanently deleted!")
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;font-family:Arial,Helvetica,sans-serif;background-color:#f4f7f8;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7f8;padding:30px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 8px 30px rgba(0,0,0,0.12);">
<tr><td style="background-color:#004155;padding:35px 30px 25px 30px;text-align:center;">
<h1 style="color:#ffffff;margin:0;font-size:22px;font-weight:700;">File Expiration Reminder</h1>
<p style="color:#5f828c;margin:8px 0 0 0;font-size:13px;">Action Required</p>
</td></tr>
<tr><td style="padding:30px;">
<div style="background-color:%s;border-left:4px solid %s;padding:16px 20px;margin:0 0 20px 0;border-radius:6px;">
<p style="margin:0;color:%s;font-size:14px;">%s</p>
</div>
<div style="background:#f4f7f8;border:1px solid #dce4e6;border-left:4px solid #5f828c;border-radius:8px;overflow:hidden;margin:20px 0;">
<table width="100%%" cellpadding="0" cellspacing="0">
<tr><td style="padding:10px 15px;color:#5f828c;font-size:13px;font-weight:600;border-bottom:1px solid #eef2f3;">Filename</td><td style="padding:10px 15px;color:#1a2a32;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td></tr>
<tr><td style="padding:10px 15px;color:#5f828c;font-size:13px;font-weight:600;border-bottom:1px solid #eef2f3;">Size</td><td style="padding:10px 15px;color:#1a2a32;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td></tr>
<tr><td style="padding:10px 15px;color:#5f828c;font-size:13px;font-weight:600;border-bottom:1px solid #eef2f3;">Expires</td><td style="padding:10px 15px;color:#1a2a32;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td></tr>
<tr><td style="padding:10px 15px;color:#5f828c;font-size:13px;font-weight:600;border-bottom:1px solid #eef2f3;">Downloads</td><td style="padding:10px 15px;color:#1a2a32;font-size:13px;border-bottom:1px solid #eef2f3;">%d</td></tr>
%s
</table></div>
<table width="100%%" cellpadding="0" cellspacing="0" style="margin:25px 0;">
<tr><td align="center">
<a href="%s" style="display:inline-block;background-color:#004155;color:#ffffff;padding:14px 40px;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;letter-spacing:0.5px;box-shadow:0 4px 12px rgba(0,65,85,0.3);">DOWNLOAD NOW</a>
</td></tr></table>
</td></tr>
<tr><td style="background-color:#323e48;padding:25px 30px;text-align:center;">
<p style="margin:0 0 5px 0;color:#5f828c;font-size:11px;">Prudencia Security &middot; Secure File Transfer</p>
<p style="margin:0;color:#5f828c;font-size:10px;opacity:0.7;">This is an automated message. Do not reply to this email.</p>
</td></tr>
</table></td></tr></table></body></html>`,
		bannerBg, bannerBorder, bannerTextColor, urgencyMsg,
		file.Name, file.Size, file.ExpireAtString, file.DownloadCount,
		getCommentRow(file.Comment),
		splashLink)

	textBody := fmt.Sprintf(`%s — File Expiration Notice

The following shared file will be permanently deleted in %d day(s):

Filename: %s
Size: %s
Expires: %s
Downloads: %d
%s
Download now: %s

---
Prudencia Security — Secure File Transfer
`,
		urgencyText, daysLeft,
		file.Name, file.Size, file.ExpireAtString, file.DownloadCount,
		getCommentText(file.Comment),
		downloadLink)

	// Send to file owner
	if owner != nil && owner.Email != "" {
		if err := sendReminderEmail(owner.Email, subject, htmlBody, textBody); err != nil {
			log.Printf("Reminder: Failed to send to owner %s: %v", owner.Email, err)
		} else {
			log.Printf("Reminder: Sent to owner %s for file %s (%d days left)", owner.Email, file.Name, daysLeft)
		}
	}
}

func sendReminderEmail(to, subject, htmlBody, textBody string) error {
	provider, err := email.GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(to, subject, htmlBody, textBody)
}

func daysUntilExpiry(file *database.FileInfo) int {
	if file.ExpireAt <= 0 {
		return 999
	}
	expireTime := time.Unix(file.ExpireAt, 0)
	return int(time.Until(expireTime).Hours() / 24)
}

func getCommentRow(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf(`<tr><td style="padding:10px 15px;color:#5f828c;font-size:13px;font-weight:600;border-bottom:1px solid #eef2f3;">Description</td><td style="padding:10px 15px;color:#1a2a32;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td></tr>`, comment)
}

func getCommentText(comment string) string {
	if comment == "" {
		return ""
	}
	return "Description: " + comment + "\n"
}
