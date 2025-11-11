package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Frimurare/Sharecare/internal/models"
)

// CreateDownloadAccount creates a new download account
func (d *Database) CreateDownloadAccount(account *models.DownloadAccount) error {
	if account.CreatedAt == 0 {
		account.CreatedAt = time.Now().Unix()
	}

	isActive := 1
	if !account.IsActive {
		isActive = 0
	}

	result, err := d.db.Exec(`
		INSERT INTO DownloadAccounts (Name, Email, Password, CreatedAt, LastUsed, DownloadCount, IsActive)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		account.Name, account.Email, account.Password, account.CreatedAt, account.LastUsed, account.DownloadCount, isActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	account.Id = int(id)
	return nil
}

// GetDownloadAccountByEmail retrieves a download account by email (excluding soft-deleted accounts)
func (d *Database) GetDownloadAccountByEmail(email string) (*models.DownloadAccount, error) {
	account := &models.DownloadAccount{}
	var isActive int
	var deletedAt int64
	var deletedBy, originalEmail string

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, CreatedAt, LastUsed, DownloadCount, IsActive,
		       COALESCE(DeletedAt, 0), COALESCE(DeletedBy, ''), COALESCE(OriginalEmail, '')
		FROM DownloadAccounts
		WHERE Email = ? AND (DeletedAt = 0 OR DeletedAt IS NULL)`, email).Scan(
		&account.Id, &account.Name, &account.Email, &account.Password, &account.CreatedAt,
		&account.LastUsed, &account.DownloadCount, &isActive,
		&deletedAt, &deletedBy, &originalEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	account.IsActive = isActive == 1
	account.DeletedAt = deletedAt
	account.DeletedBy = deletedBy
	account.OriginalEmail = originalEmail
	return account, nil
}

// GetDownloadAccountByID retrieves a download account by ID (including soft-deleted for admin purposes)
func (d *Database) GetDownloadAccountByID(id int) (*models.DownloadAccount, error) {
	account := &models.DownloadAccount{}
	var isActive int
	var deletedAt int64
	var deletedBy, originalEmail string

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, CreatedAt, LastUsed, DownloadCount, IsActive,
		       COALESCE(DeletedAt, 0), COALESCE(DeletedBy, ''), COALESCE(OriginalEmail, '')
		FROM DownloadAccounts WHERE Id = ?`, id).Scan(
		&account.Id, &account.Name, &account.Email, &account.Password, &account.CreatedAt,
		&account.LastUsed, &account.DownloadCount, &isActive,
		&deletedAt, &deletedBy, &originalEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	account.IsActive = isActive == 1
	account.DeletedAt = deletedAt
	account.DeletedBy = deletedBy
	account.OriginalEmail = originalEmail
	return account, nil
}

// UpdateDownloadAccount updates an existing download account
func (d *Database) UpdateDownloadAccount(account *models.DownloadAccount) error {
	isActive := 1
	if !account.IsActive {
		isActive = 0
	}

	_, err := d.db.Exec(`
		UPDATE DownloadAccounts SET Email = ?, Password = ?, LastUsed = ?, DownloadCount = ?, IsActive = ?
		WHERE Id = ?`,
		account.Email, account.Password, account.LastUsed, account.DownloadCount, isActive, account.Id,
	)
	return err
}

// UpdateDownloadAccountLastUsed updates the last used timestamp and increments download count
func (d *Database) UpdateDownloadAccountLastUsed(id int) error {
	_, err := d.db.Exec(`
		UPDATE DownloadAccounts
		SET LastUsed = ?, DownloadCount = DownloadCount + 1
		WHERE Id = ?`,
		time.Now().Unix(), id,
	)
	return err
}

// GetAllDownloadAccounts returns all download accounts (excluding soft-deleted)
func (d *Database) GetAllDownloadAccounts() ([]*models.DownloadAccount, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Email, Password, CreatedAt, LastUsed, DownloadCount, IsActive,
		       COALESCE(DeletedAt, 0), COALESCE(DeletedBy, ''), COALESCE(OriginalEmail, '')
		FROM DownloadAccounts
		WHERE (DeletedAt = 0 OR DeletedAt IS NULL)
		ORDER BY LastUsed DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.DownloadAccount
	for rows.Next() {
		account := &models.DownloadAccount{}
		var isActive int
		var deletedAt int64
		var deletedBy, originalEmail string

		err := rows.Scan(&account.Id, &account.Name, &account.Email, &account.Password, &account.CreatedAt,
			&account.LastUsed, &account.DownloadCount, &isActive,
			&deletedAt, &deletedBy, &originalEmail)
		if err != nil {
			return nil, err
		}

		account.IsActive = isActive == 1
		account.DeletedAt = deletedAt
		account.DeletedBy = deletedBy
		account.OriginalEmail = originalEmail
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// CreateDownloadLog creates a new download log entry
func (d *Database) CreateDownloadLog(log *models.DownloadLog) error {
	if log.DownloadedAt == 0 {
		log.DownloadedAt = time.Now().Unix()
	}

	isAuth := 0
	if log.IsAuthenticated {
		isAuth = 1
	}

	var downloadAccountId sql.NullInt64
	if log.DownloadAccountId > 0 {
		downloadAccountId = sql.NullInt64{Int64: int64(log.DownloadAccountId), Valid: true}
	}

	result, err := d.db.Exec(`
		INSERT INTO DownloadLogs (FileId, DownloadAccountId, Email, IpAddress, UserAgent,
		                          DownloadedAt, FileSize, FileName, IsAuthenticated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.FileId, downloadAccountId, log.Email, log.IpAddress, log.UserAgent,
		log.DownloadedAt, log.FileSize, log.FileName, isAuth,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.Id = int(id)
	return nil
}

// GetDownloadLogsByFileID retrieves all download logs for a specific file
func (d *Database) GetDownloadLogsByFileID(fileId string) ([]*models.DownloadLog, error) {
	rows, err := d.db.Query(`
		SELECT Id, FileId, DownloadAccountId, Email, IpAddress, UserAgent,
		       DownloadedAt, FileSize, FileName, IsAuthenticated
		FROM DownloadLogs WHERE FileId = ? ORDER BY DownloadedAt DESC`, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDownloadLogs(rows)
}

// GetDownloadLogsByAccountID retrieves all download logs for a specific download account
func (d *Database) GetDownloadLogsByAccountID(accountId int) ([]*models.DownloadLog, error) {
	rows, err := d.db.Query(`
		SELECT Id, FileId, DownloadAccountId, Email, IpAddress, UserAgent,
		       DownloadedAt, FileSize, FileName, IsAuthenticated
		FROM DownloadLogs WHERE DownloadAccountId = ? ORDER BY DownloadedAt DESC`, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDownloadLogs(rows)
}

// GetAllDownloadLogs retrieves all download logs
func (d *Database) GetAllDownloadLogs(limit int) ([]*models.DownloadLog, error) {
	query := `
		SELECT Id, FileId, DownloadAccountId, Email, IpAddress, UserAgent,
		       DownloadedAt, FileSize, FileName, IsAuthenticated
		FROM DownloadLogs ORDER BY DownloadedAt DESC`

	if limit > 0 {
		query += " LIMIT ?"
		rows, err := d.db.Query(query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanDownloadLogs(rows)
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDownloadLogs(rows)
}

// scanDownloadLogs is a helper function to scan download log rows
func scanDownloadLogs(rows *sql.Rows) ([]*models.DownloadLog, error) {
	var logs []*models.DownloadLog
	for rows.Next() {
		log := &models.DownloadLog{}
		var accountId sql.NullInt64
		var isAuth int

		err := rows.Scan(&log.Id, &log.FileId, &accountId, &log.Email, &log.IpAddress,
			&log.UserAgent, &log.DownloadedAt, &log.FileSize, &log.FileName, &isAuth)
		if err != nil {
			return nil, err
		}

		if accountId.Valid {
			log.DownloadAccountId = int(accountId.Int64)
		}
		log.IsAuthenticated = isAuth == 1
		logs = append(logs, log)
	}

	return logs, nil
}

// GetTotalDownloads returns the total number of downloads
func (d *Database) GetTotalDownloads() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM DownloadLogs").Scan(&count)
	return count, err
}

// GetDownloadsToday returns the number of downloads today
func (d *Database) GetDownloadsToday() (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM DownloadLogs WHERE DownloadedAt >= ?", startOfDay).Scan(&count)
	return count, err
}

// AnonymizeDownloadAccount anonymizes a download account for GDPR compliance
// This uses soft delete approach - marks as deleted but keeps data for 90 days
// NOTE: This is now handled by SoftDeleteDownloadAccount in migrations.go
// Kept for backward compatibility
func (d *Database) AnonymizeDownloadAccount(id int) error {
	return d.SoftDeleteDownloadAccount(id, "user")
}

// DeleteDownloadAccount permanently deletes a download account (use with caution)
func (d *Database) DeleteDownloadAccount(id int) error {
	// First delete all download logs for this account
	_, err := d.db.Exec("DELETE FROM DownloadLogs WHERE DownloadAccountId = ?", id)
	if err != nil {
		return err
	}

	// Then delete the account
	_, err = d.db.Exec("DELETE FROM DownloadAccounts WHERE Id = ?", id)
	return err
}
