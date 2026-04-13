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

	subject := fmt.Sprintf("[WulfVault] %s: File '%s' expires in %d day(s)", urgencyText, file.Name, daysLeft)

	splashLink := serverURL + "/s/" + file.Id
	downloadLink := serverURL + "/d/" + file.Id

	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: %s;">%s — File Expiration Notice</h2>
			<p>The following shared file will be <strong>permanently deleted in %d day(s)</strong>:</p>
			<div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid %s;">
				<p><strong>Filename:</strong> %s</p>
				<p><strong>Size:</strong> %s</p>
				<p><strong>Expires:</strong> %s</p>
				<p><strong>Downloads:</strong> %d</p>
			</div>
			%s
			<div style="margin: 30px 0;">
				<a href="%s" style="background: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">Download Now Before It Expires</a>
			</div>
			<hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
			<p style="color: #999; font-size: 12px;">This is an automated expiration reminder from WulfVault Secure File Transfer.</p>
		</div>
	`,
		getUrgencyColor(urgent), urgencyText, daysLeft,
		getUrgencyColor(urgent),
		file.Name, file.Size, file.ExpireAtString, file.DownloadCount,
		getCommentHTML(file.Comment),
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
This is an automated expiration reminder from WulfVault Secure File Transfer.
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

	// TODO: Send to original recipient if tracked (requires storing send_to_email in FileInfo)
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

func getUrgencyColor(urgent bool) string {
	if urgent {
		return "#dc3545" // red
	}
	return "#ffc107" // yellow
}

func getCommentHTML(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf(`<p><strong>Description:</strong> %s</p>`, comment)
}

func getCommentText(comment string) string {
	if comment == "" {
		return ""
	}
	return "Description: " + comment + "\n"
}
