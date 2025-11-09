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

	_, err := d.db.Exec(`
		INSERT INTO Files (
			Id, Name, Size, SHA1, PasswordHash, HotlinkId, ContentType,
			AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
			UploadDate, DownloadsRemaining, DownloadCount, UserId,
			UnlimitedDownloads, UnlimitedTime, RequireAuth
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.Id, file.Name, file.Size, file.SHA1, file.PasswordHash, file.HotlinkId,
		file.ContentType, file.AwsBucket, file.ExpireAtString, file.ExpireAt,
		file.PendingDeletion, file.SizeBytes, file.UploadDate, file.DownloadsRemaining,
		file.DownloadCount, file.UserId, unlimitedDownloads, unlimitedTime, requireAuth,
	)
	return err
}

// GetFileByID retrieves a file by its ID
func (d *Database) GetFileByID(id string) (*FileInfo, error) {
	file := &FileInfo{}
	var unlimitedDownloads, unlimitedTime, requireAuth int

	err := d.db.QueryRow(`
		SELECT Id, Name, Size, SHA1, PasswordHash, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth
		FROM Files WHERE Id = ?`, id).Scan(
		&file.Id, &file.Name, &file.Size, &file.SHA1, &file.PasswordHash,
		&file.HotlinkId, &file.ContentType, &file.AwsBucket, &file.ExpireAtString,
		&file.ExpireAt, &file.PendingDeletion, &file.SizeBytes, &file.UploadDate,
		&file.DownloadsRemaining, &file.DownloadCount, &file.UserId,
		&unlimitedDownloads, &unlimitedTime, &requireAuth,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	file.UnlimitedDownloads = unlimitedDownloads == 1
	file.UnlimitedTime = unlimitedTime == 1
	file.RequireAuth = requireAuth == 1

	return file, nil
}

// GetFilesByUser returns all files for a user
func (d *Database) GetFilesByUser(userId int) ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth
		FROM Files WHERE UserId = ? ORDER BY UploadDate DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// GetAllFiles returns all files
func (d *Database) GetAllFiles() ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth
		FROM Files ORDER BY UploadDate DESC`)
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

// DeleteFile deletes a file from the database
func (d *Database) DeleteFile(fileId string) error {
	_, err := d.db.Exec("DELETE FROM Files WHERE Id = ?", fileId)
	return err
}

// GetExpiredFiles returns files that should be deleted
func (d *Database) GetExpiredFiles() ([]*FileInfo, error) {
	now := time.Now().Unix()

	rows, err := d.db.Query(`
		SELECT Id, Name, Size, SHA1, PasswordHash, HotlinkId, ContentType,
		       AwsBucket, ExpireAtString, ExpireAt, PendingDeletion, SizeBytes,
		       UploadDate, DownloadsRemaining, DownloadCount, UserId,
		       UnlimitedDownloads, UnlimitedTime, RequireAuth
		FROM Files
		WHERE (ExpireAt > 0 AND ExpireAt < ? AND UnlimitedTime = 0)
		   OR (DownloadsRemaining <= 0 AND UnlimitedDownloads = 0)`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiles(rows)
}

// CalculateUserStorage calculates total storage used by a user
func (d *Database) CalculateUserStorage(userId int) (int64, error) {
	var totalBytes sql.NullInt64

	err := d.db.QueryRow(`
		SELECT SUM(SizeBytes) FROM Files WHERE UserId = ?`, userId).Scan(&totalBytes)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if !totalBytes.Valid {
		return 0, nil
	}

	// Convert bytes to MB
	return totalBytes.Int64 / (1024 * 1024), nil
}

// GetTotalFiles returns the count of all files
func (d *Database) GetTotalFiles() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Files").Scan(&count)
	return count, err
}

// GetActiveFiles returns count of non-expired files
func (d *Database) GetActiveFiles() (int, error) {
	now := time.Now().Unix()
	var count int

	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM Files
		WHERE (ExpireAt = 0 OR ExpireAt > ? OR UnlimitedTime = 1)
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

		err := rows.Scan(
			&file.Id, &file.Name, &file.Size, &file.SHA1, &file.PasswordHash,
			&file.HotlinkId, &file.ContentType, &file.AwsBucket, &file.ExpireAtString,
			&file.ExpireAt, &file.PendingDeletion, &file.SizeBytes, &file.UploadDate,
			&file.DownloadsRemaining, &file.DownloadCount, &file.UserId,
			&unlimitedDownloads, &unlimitedTime, &requireAuth,
		)
		if err != nil {
			return nil, err
		}

		file.UnlimitedDownloads = unlimitedDownloads == 1
		file.UnlimitedTime = unlimitedTime == 1
		file.RequireAuth = requireAuth == 1

		files = append(files, file)
	}

	return files, nil
}
