// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// FileInfo represents a file in the database
type FileInfo struct {
	Id                 string
	Name               string
	Size               string
	SHA1               string
	PasswordHash       string
	FilePasswordPlain  string
	HotlinkId          string
	ContentType        string
	AwsBucket          string
	ExpireAtString     string
	ExpireAt           int64
	PendingDeletion    int64
	SizeBytes          int64
	UploadDate         int64
	DownloadsRemaining int
	DownloadCount      int
	UserId             int
	UnlimitedDownloads bool
	UnlimitedTime      bool
	RequireAuth        bool
	DeletedAt          int64
	DeletedBy          int
}

// SaveFile saves file metadata to the database
func (d *Database) SaveFile(file *FileInfo) error {
	unlimitedDownloads := 0
	if file.UnlimitedDownloads {
		unlimitedDownloads = 1
	}
	unlimitedTime := 0
	if file.UnlimitedTime {
		unlimitedTime = 1
	}
	requireAuth := 0
	if file.RequireAuth {
		requireAuth = 1
	}

	// Convert empty password to NULL for database storage
	var filePassword interface{}
	if file.FilePasswordPlain == "" {
		filePassword = nil
	} else {
		filePassword = file.FilePasswordPlain
	}

	_, err := d.db.Exec(`
		INSERT INTO Files (
			Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
			AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
			UploadDate, DownloadsRemaining, DownloadCount, UserId,
			UnlimitedDownloads, UnlimitedTime, RequireAuth
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.Id, file.Name, file.Size, file.SHA1, file.PasswordHash, filePassword, file.HotlinkId,
		file.ContentType, file.AwsBucket, file.ExpireAtString, file.ExpireAt,
		file.PendingDeletion, file.SizeBytes, file.UploadDate, file.DownloadsRemaining,
		file.DownloadCount, file.UserId, unlimitedDownloads, unlimitedTime, requireAuth,
	)
	return err
}

// GetFileByID retrieves a file by its ID (only non-deleted files)
func (d *Database) GetFileByID(id string) (*FileInfo, error) {
	file := &FileInfo{}
	var unlimitedDownloads, unlimitedTime, requireAuth int
	var filePassword sql.NullString

	err := d.db.QueryRow(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files WHERE Id = ? AND DeletedAt = 0`, id).Scan(
		&file.Id, &file.Name, &file.Size, &file.SHA1, &file.PasswordHash, &filePassword,
		&file.HotlinkId, &file.ContentType, &file.AwsBucket, &file.ExpireAtString,
		&file.ExpireAt, &file.PendingDeletion, &file.SizeBytes, &file.UploadDate,
		&file.DownloadsRemaining, &file.DownloadCount, &file.UserId,
		&unlimitedDownloads, &unlimitedTime, &requireAuth, &file.DeletedAt, &file.DeletedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	// Handle NULL password
	if filePassword.Valid {
		file.FilePasswordPlain = filePassword.String
	}

	file.UnlimitedDownloads = unlimitedDownloads == 1
	file.UnlimitedTime = unlimitedTime == 1
	file.RequireAuth = requireAuth == 1

	return file, nil
}

// GetFilesByUser returns all non-deleted files for a user
func (d *Database) GetFilesByUser(userId int) ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files WHERE UserId = ? AND DeletedAt = 0 ORDER BY UploadDate DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// GetAllFiles returns all non-deleted files
func (d *Database) GetAllFiles() ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files WHERE DeletedAt = 0 ORDER BY UploadDate DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// UpdateFileDownloadCount increments download count and decrements remaining
func (d *Database) UpdateFileDownloadCount(fileId string) error {
	_, err := d.db.Exec(`
		UPDATE Files
		SET DownloadCount = DownloadCount + 1,
		    DownloadsRemaining = CASE
		        WHEN UnlimitedDownloads = 1 THEN DownloadsRemaining
		        ELSE DownloadsRemaining - 1
		    END
		WHERE Id = ?`, fileId)
	return err
}

// UpdateFileSettings updates a file's expiration and download settings
func (d *Database) UpdateFileSettings(fileId string, downloadsRemaining int, expireAt int64, expireAtString string, unlimitedDownloads, unlimitedTime bool) error {
	unlimitedDownloadsInt := 0
	if unlimitedDownloads {
		unlimitedDownloadsInt = 1
	}
	unlimitedTimeInt := 0
	if unlimitedTime {
		unlimitedTimeInt = 1
	}

	_, err := d.db.Exec(`
		UPDATE Files
		SET DownloadsRemaining = ?,
		    ExpireAt = ?,
		    ExpireAtString = ?,
		    UnlimitedDownloads = ?,
		    UnlimitedTime = ?
		WHERE Id = ?`,
		downloadsRemaining, expireAt, expireAtString, unlimitedDownloadsInt, unlimitedTimeInt, fileId)
	return err
}

// UpdateFilePassword updates a file's password
func (d *Database) UpdateFilePassword(fileId string, password string) error {
	// Convert empty password to NULL for database storage
	var filePassword interface{}
	if password == "" {
		filePassword = nil
	} else {
		filePassword = password
	}

	_, err := d.db.Exec("UPDATE Files SET FilePasswordPlain = ? WHERE Id = ?", filePassword, fileId)
	return err
}

// DeleteFile soft-deletes a file (moves to trash for 5 days)
func (d *Database) DeleteFile(fileId string, userId int) error {
	now := time.Now().Unix()
	_, err := d.db.Exec("UPDATE Files SET DeletedAt = ?, DeletedBy = ? WHERE Id = ?", now, userId, fileId)
	return err
}

// SoftDeleteUserFiles soft-deletes all files belonging to a user (moves to trash)
// This is used when deleting a user account to preserve files in trash
func (d *Database) SoftDeleteUserFiles(userId int, deletedBy int) error {
	now := time.Now().Unix()
	_, err := d.db.Exec("UPDATE Files SET DeletedAt = ?, DeletedBy = ? WHERE UserId = ? AND DeletedAt = 0", now, deletedBy, userId)
	return err
}

// PermanentDeleteFile permanently deletes a file from the database
func (d *Database) PermanentDeleteFile(fileId string) error {
	// First delete associated download logs to avoid foreign key constraint violation
	_, err := d.db.Exec("DELETE FROM DownloadLogs WHERE FileId = ?", fileId)
	if err != nil {
		return fmt.Errorf("failed to delete download logs: %w", err)
	}

	// Then delete the file itself
	_, err = d.db.Exec("DELETE FROM Files WHERE Id = ?", fileId)
	return err
}

// GetDeletedFiles returns all files in trash (admin only)
func (d *Database) GetDeletedFiles() ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files WHERE DeletedAt > 0 ORDER BY DeletedAt DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// GetOldDeletedFiles returns files deleted more than retentionDays ago for cleanup
func (d *Database) GetOldDeletedFiles(retentionDays int) ([]*FileInfo, error) {
	if retentionDays <= 0 {
		retentionDays = 5 // default fallback
	}
	cutoffTime := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour).Unix()

	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// RestoreFile restores a file from trash
func (d *Database) RestoreFile(fileId string) error {
	_, err := d.db.Exec("UPDATE Files SET DeletedAt = 0, DeletedBy = 0 WHERE Id = ?", fileId)
	return err
}

// GetExpiredFiles returns non-deleted files that should be deleted
func (d *Database) GetExpiredFiles() ([]*FileInfo, error) {
	now := time.Now().Unix()

	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, FilePasswordPlain, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth, DeletedAt, DeletedBy
		FROM Files
		WHERE DeletedAt = 0 AND ((ExpireAt > 0 AND ExpireAt < ? AND UnlimitedTime = 0)
		   OR (DownloadsRemaining <= 0 AND UnlimitedDownloads = 0))`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// CalculateUserStorage calculates total storage used by a user (non-deleted files only)
func (d *Database) CalculateUserStorage(userId int) (int64, error) {
	var totalBytes sql.NullInt64

	err := d.db.QueryRow(`
		SELECT SUM(SizeBytes) FROM Files WHERE UserId = ? AND DeletedAt = 0`, userId).Scan(&totalBytes)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if !totalBytes.Valid {
		return 0, nil
	}

	// Convert bytes to MB
	return totalBytes.Int64 / (1024 * 1024), nil
}

// GetTotalFiles returns the count of all non-deleted files
func (d *Database) GetTotalFiles() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Files WHERE DeletedAt = 0").Scan(&count)
	return count, err
}

// GetActiveFiles returns count of non-expired, non-deleted files
func (d *Database) GetActiveFiles() (int, error) {
	now := time.Now().Unix()
	var count int

	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM Files
		WHERE DeletedAt = 0 AND (ExpireAt = 0 OR ExpireAt > ? OR UnlimitedTime = 1)
		  AND (DownloadsRemaining > 0 OR UnlimitedDownloads = 1)`, now).Scan(&count)

	return count, err
}

// CalculateFileSHA1 calculates SHA1 hash of a file
func CalculateFileSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// FormatFileSize formats bytes to human-readable size
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// scanFiles is a helper to scan file rows
func scanFiles(rows *sql.Rows) ([]*FileInfo, error) {
	var files []*FileInfo

	for rows.Next() {
		file := &FileInfo{}
		var unlimitedDownloads, unlimitedTime, requireAuth int
		var filePassword sql.NullString

		err := rows.Scan(
			&file.Id, &file.Name, &file.Size, &file.SHA1, &file.PasswordHash, &filePassword,
			&file.HotlinkId, &file.ContentType, &file.AwsBucket, &file.ExpireAtString,
			&file.ExpireAt, &file.PendingDeletion, &file.SizeBytes, &file.UploadDate,
			&file.DownloadsRemaining, &file.DownloadCount, &file.UserId,
			&unlimitedDownloads, &unlimitedTime, &requireAuth, &file.DeletedAt, &file.DeletedBy,
		)
		if err != nil {
			return nil, err
		}

		// Handle NULL password
		if filePassword.Valid {
			file.FilePasswordPlain = filePassword.String
		}

		file.UnlimitedDownloads = unlimitedDownloads == 1
		file.UnlimitedTime = unlimitedTime == 1
		file.RequireAuth = requireAuth == 1

		files = append(files, file)
	}

	return files, nil
}

// GetMostDownloadedFile returns the file with most downloads and its download count
func (d *Database) GetMostDownloadedFile() (string, int, error) {
	var fileName string
	var downloadCount int

	err := d.db.QueryRow(`
		SELECT Files.Name, COUNT(DownloadLogs.Id) as downloads
		FROM Files
		LEFT JOIN DownloadLogs ON Files.Id = DownloadLogs.FileId
		WHERE Files.DeletedAt = 0
		GROUP BY Files.Id
		ORDER BY downloads DESC
		LIMIT 1
	`).Scan(&fileName, &downloadCount)

	if err == sql.ErrNoRows {
		return "No files yet", 0, nil
	}
	if err != nil {
		return "", 0, err
	}

	return fileName, downloadCount, nil
}
