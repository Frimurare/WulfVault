package database

import (
	"time"

	"github.com/Frimurare/Sharecare/internal/models"
)

// LogEmailSent creates a new email log entry
func (d *Database) LogEmailSent(fileId string, senderUserId int, recipientEmail, message, fileName string, fileSize int64) error {
	_, err := d.db.Exec(`
		INSERT INTO EmailLogs (FileId, SenderUserId, RecipientEmail, Message, SentAt, FileName, FileSize)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fileId, senderUserId, recipientEmail, message, time.Now().Unix(), fileName, fileSize)
	return err
}

// GetEmailLogsByFileID retrieves all email logs for a specific file
func (d *Database) GetEmailLogsByFileID(fileId string) ([]*models.EmailLog, error) {
	rows, err := d.db.Query(`
		SELECT Id, FileId, SenderUserId, RecipientEmail, Message, SentAt, FileName, FileSize
		FROM EmailLogs WHERE FileId = ? ORDER BY SentAt DESC`, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.EmailLog
	for rows.Next() {
		log := &models.EmailLog{}
		err := rows.Scan(&log.Id, &log.FileId, &log.SenderUserId, &log.RecipientEmail,
			&log.Message, &log.SentAt, &log.FileName, &log.FileSize)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
