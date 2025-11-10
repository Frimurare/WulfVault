package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/Frimurare/Sharecare/internal/models"
)

// CreateFileRequest creates a new file upload request
func (d *Database) CreateFileRequest(req *models.FileRequest) error {
	if req.CreatedAt == 0 {
		req.CreatedAt = time.Now().Unix()
	}
	if req.RequestToken == "" {
		token, err := generateRequestToken()
		if err != nil {
			return err
		}
		req.RequestToken = token
	}

	result, err := d.db.Exec(`
		INSERT INTO FileRequests (UserId, RequestToken, Title, Message, CreatedAt, ExpiresAt, IsActive, MaxFileSize, AllowedFileTypes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.UserId, req.RequestToken, req.Title, req.Message, req.CreatedAt, req.ExpiresAt, boolToInt(req.IsActive), req.MaxFileSize, req.AllowedFileTypes,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	req.Id = int(id)
	return nil
}

// GetFileRequestByToken retrieves a file request by its token
func (d *Database) GetFileRequestByToken(token string) (*models.FileRequest, error) {
	req := &models.FileRequest{}
	var isActive int

	err := d.db.QueryRow(`
		SELECT Id, UserId, RequestToken, Title, Message, CreatedAt, ExpiresAt, IsActive, MaxFileSize, AllowedFileTypes
		FROM FileRequests WHERE RequestToken = ?`, token).Scan(
		&req.Id, &req.UserId, &req.RequestToken, &req.Title, &req.Message,
		&req.CreatedAt, &req.ExpiresAt, &isActive, &req.MaxFileSize, &req.AllowedFileTypes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("file request not found")
		}
		return nil, err
	}

	req.IsActive = isActive == 1
	return req, nil
}

// GetFileRequestsByUser retrieves all file requests for a user
func (d *Database) GetFileRequestsByUser(userId int) ([]*models.FileRequest, error) {
	rows, err := d.db.Query(`
		SELECT Id, UserId, RequestToken, Title, Message, CreatedAt, ExpiresAt, IsActive, MaxFileSize, AllowedFileTypes
		FROM FileRequests WHERE UserId = ? ORDER BY CreatedAt DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.FileRequest
	for rows.Next() {
		req := &models.FileRequest{}
		var isActive int

		err := rows.Scan(&req.Id, &req.UserId, &req.RequestToken, &req.Title, &req.Message,
			&req.CreatedAt, &req.ExpiresAt, &isActive, &req.MaxFileSize, &req.AllowedFileTypes)
		if err != nil {
			return nil, err
		}

		req.IsActive = isActive == 1
		requests = append(requests, req)
	}

	return requests, nil
}

// UpdateFileRequest updates an existing file request
func (d *Database) UpdateFileRequest(req *models.FileRequest) error {
	_, err := d.db.Exec(`
		UPDATE FileRequests SET Title = ?, Message = ?, ExpiresAt = ?, IsActive = ?, MaxFileSize = ?, AllowedFileTypes = ?
		WHERE Id = ?`,
		req.Title, req.Message, req.ExpiresAt, boolToInt(req.IsActive), req.MaxFileSize, req.AllowedFileTypes, req.Id,
	)
	return err
}

// DeleteFileRequest deletes a file request
func (d *Database) DeleteFileRequest(id int) error {
	_, err := d.db.Exec("DELETE FROM FileRequests WHERE Id = ?", id)
	return err
}

// generateRequestToken generates a unique token for file requests
func generateRequestToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CleanupExpiredFileRequests deletes file requests that have been expired for more than 10 days
// This keeps the expired message visible for 10 days, then removes the request entirely
func (d *Database) CleanupExpiredFileRequests() error {
	// Delete requests that expired more than 10 days ago
	// This means: CreatedAt + 24 hours (expiration) + 10 days < now
	// Or: ExpiresAt + 10 days < now
	cutoffTime := time.Now().Add(-10 * 24 * time.Hour).Unix()

	result, err := d.db.Exec("DELETE FROM FileRequests WHERE ExpiresAt > 0 AND ExpiresAt < ?", cutoffTime)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired file requests", rowsAffected)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
