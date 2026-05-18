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

// SendExpirationReminders sends ONE expiration reminder per file, only to
// files that have not yet been downloaded and have not already been
// reminded. Triggered when a file is within 1 day of expiration — the
// "last chance" window. This guarantees every file gets exactly one
// reminder regardless of its total expiration period, and it stops
// retransmissions on every 6 h cleanup tick.
//
// Behavior note: a file that has been downloaded at least once is
// considered "delivered" and is silently skipped. The previous
// implementation hammered both downloaded files and not-yet-downloaded
// files with a reminder every 6 h, up to four times per day.
func SendExpirationReminders(serverURL string) {
	files, err := database.DB.GetFilesNeedingReminder(1)
	if err != nil {
		log.Printf("Reminder: Error getting files needing reminder: %v", err)
		return
	}
	for _, file := range files {
		sendReminder(file, serverURL, true)
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
	sent := false
	if owner != nil && owner.Email != "" {
		if err := sendReminderEmail(owner.Email, subject, htmlBody, textBody); err != nil {
			log.Printf("Reminder: Failed to send to owner %s: %v", owner.Email, err)
		} else {
			log.Printf("Reminder: Sent to owner %s for file %s (%d days left)", owner.Email, file.Name, daysLeft)
			sent = true
		}
	}

	// Stamp RemindedAt so this file is not picked up again. Only stamp on a
	// successful send — a transient SMTP failure should retry on the next tick.
	if sent {
		if err := database.DB.MarkFileReminded(file.Id); err != nil {
			log.Printf("Reminder: Failed to mark file %s as reminded: %v", file.Id, err)
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
