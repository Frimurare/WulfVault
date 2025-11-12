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

// GetBytesSentToday returns total bytes transferred today (includes deleted files for historical accuracy)
func (d *Database) GetBytesSentToday() (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(Files.SizeBytes), 0)
		FROM DownloadLogs
		JOIN Files ON DownloadLogs.FileId = Files.Id
		WHERE DownloadLogs.DownloadedAt >= ?
	`, startOfDay).Scan(&total)
	return total, err
}

// GetBytesSentThisWeek returns total bytes transferred this week (includes deleted files for historical accuracy)
func (d *Database) GetBytesSentThisWeek() (int64, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(Files.SizeBytes), 0)
		FROM DownloadLogs
		JOIN Files ON DownloadLogs.FileId = Files.Id
		WHERE DownloadLogs.DownloadedAt >= ?
	`, startOfWeek.Unix()).Scan(&total)
	return total, err
}

// GetBytesSentThisMonth returns total bytes transferred this month (includes deleted files for historical accuracy)
func (d *Database) GetBytesSentThisMonth() (int64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(Files.SizeBytes), 0)
		FROM DownloadLogs
		JOIN Files ON DownloadLogs.FileId = Files.Id
		WHERE DownloadLogs.DownloadedAt >= ?
	`, startOfMonth).Scan(&total)
	return total, err
}

// GetBytesSentThisYear returns total bytes transferred this year (includes deleted files for historical accuracy)
func (d *Database) GetBytesSentThisYear() (int64, error) {
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(Files.SizeBytes), 0)
		FROM DownloadLogs
		JOIN Files ON DownloadLogs.FileId = Files.Id
		WHERE DownloadLogs.DownloadedAt >= ?
	`, startOfYear).Scan(&total)
	return total, err
}

// GetBytesUploadedToday returns total bytes uploaded today (includes deleted files for historical accuracy)
func (d *Database) GetBytesUploadedToday() (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE UploadDate >= ?
	`, startOfDay).Scan(&total)
	return total, err
}

// GetBytesUploadedThisWeek returns total bytes uploaded this week (includes deleted files for historical accuracy)
func (d *Database) GetBytesUploadedThisWeek() (int64, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE UploadDate >= ?
	`, startOfWeek.Unix()).Scan(&total)
	return total, err
}

// GetBytesUploadedThisMonth returns total bytes uploaded this month (includes deleted files for historical accuracy)
func (d *Database) GetBytesUploadedThisMonth() (int64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE UploadDate >= ?
	`, startOfMonth).Scan(&total)
	return total, err
}

// GetBytesUploadedThisYear returns total bytes uploaded this year (includes deleted files for historical accuracy)
func (d *Database) GetBytesUploadedThisYear() (int64, error) {
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Unix()

	var total int64
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE UploadDate >= ?
	`, startOfYear).Scan(&total)
	return total, err
}

// ============================================================================
// USAGE STATISTICS
// ============================================================================

// GetActiveFilesLast7Days returns the count of files that have been downloaded in the last 7 days
func (d *Database) GetActiveFilesLast7Days() (int, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()

	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(DISTINCT FileId)
		FROM DownloadLogs
		WHERE DownloadedAt >= ?
	`, sevenDaysAgo).Scan(&count)
	return count, err
}

// GetActiveFilesLast30Days returns the count of files that have been downloaded in the last 30 days
func (d *Database) GetActiveFilesLast30Days() (int, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()

	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(DISTINCT FileId)
		FROM DownloadLogs
		WHERE DownloadedAt >= ?
	`, thirtyDaysAgo).Scan(&count)
	return count, err
}

// GetAverageFileSize returns the average file size in bytes (excluding deleted files)
func (d *Database) GetAverageFileSize() (int64, error) {
	var avg int64
	err := d.db.QueryRow(`
		SELECT COALESCE(AVG(SizeBytes), 0)
		FROM Files
		WHERE DeletedAt = 0
	`).Scan(&avg)
	return avg, err
}

// GetAverageDownloadsPerFile returns the average number of downloads per file
func (d *Database) GetAverageDownloadsPerFile() (float64, error) {
	var avg float64
	err := d.db.QueryRow(`
		SELECT COALESCE(AVG(download_count), 0)
		FROM (
			SELECT FileId, COUNT(*) as download_count
			FROM DownloadLogs
			GROUP BY FileId
		)
	`).Scan(&avg)
	return avg, err
}

// ============================================================================
// FILE STATISTICS
// ============================================================================

// GetLargestFile returns the name and size of the largest file
func (d *Database) GetLargestFile() (string, int64, error) {
	var name string
	var size int64

	err := d.db.QueryRow(`
		SELECT FileName, SizeBytes
		FROM Files
		WHERE DeletedAt = 0
		ORDER BY SizeBytes DESC
		LIMIT 1
	`).Scan(&name, &size)

	if err == sql.ErrNoRows {
		return "N/A", 0, nil
	}
	return name, size, err
}

// GetMostActiveUser returns the username who has uploaded the most files
func (d *Database) GetMostActiveUser() (string, int, error) {
	var username string
	var fileCount int

	err := d.db.QueryRow(`
		SELECT u.Name, COUNT(f.Id) as file_count
		FROM Files f
		JOIN Users u ON f.CreatedBy = u.Id
		WHERE f.DeletedAt = 0
		GROUP BY f.CreatedBy, u.Name
		ORDER BY file_count DESC
		LIMIT 1
	`).Scan(&username, &fileCount)

	if err == sql.ErrNoRows {
		return "N/A", 0, nil
	}
	return username, fileCount, err
}

// ============================================================================
// TREND DATA
// ============================================================================

// GetTopFileTypes returns the top 3 most common file types (by extension)
func (d *Database) GetTopFileTypes() ([]string, []int, error) {
	rows, err := d.db.Query(`
		SELECT
			CASE
				WHEN INSTR(FileName, '.') > 0
				THEN SUBSTR(FileName, INSTR(FileName, '.') + 1)
				ELSE 'no extension'
			END as extension,
			COUNT(*) as count
		FROM Files
		WHERE DeletedAt = 0
		GROUP BY extension
		ORDER BY count DESC
		LIMIT 3
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var extensions []string
	var counts []int

	for rows.Next() {
		var ext string
		var count int
		if err := rows.Scan(&ext, &count); err != nil {
			return nil, nil, err
		}
		extensions = append(extensions, ext)
		counts = append(counts, count)
	}

	if len(extensions) == 0 {
		return []string{"N/A"}, []int{0}, nil
	}

	return extensions, counts, rows.Err()
}

// GetMostActiveWeekday returns the weekday with the most downloads
func (d *Database) GetMostActiveWeekday() (string, int, error) {
	// SQLite doesn't have built-in day name function, so we'll get day number and convert
	rows, err := d.db.Query(`
		SELECT strftime('%w', datetime(DownloadedAt, 'unixepoch')) as day_num, COUNT(*) as count
		FROM DownloadLogs
		GROUP BY day_num
		ORDER BY count DESC
		LIMIT 1
	`)
	if err != nil {
		return "N/A", 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return "N/A", 0, nil
	}

	var dayNum string
	var count int
	if err := rows.Scan(&dayNum, &count); err != nil {
		return "N/A", 0, err
	}

	// Convert day number to name (0=Sunday, 1=Monday, etc.)
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	dayIndex := 0
	if dayNum != "" && dayNum >= "0" && dayNum <= "6" {
		dayIndex = int(dayNum[0] - '0')
	}

	return dayNames[dayIndex], count, nil
}

// GetStorageTrendLastMonth returns total storage used 30 days ago vs now (in bytes)
func (d *Database) GetStorageTrendLastMonth() (int64, int64, error) {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30).Unix()

	var storageThirtyDaysAgo, storageNow int64

	// Storage 30 days ago (files uploaded before that time and not yet deleted)
	err := d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE UploadDate < ? AND (DeletedAt = 0 OR DeletedAt >= ?)
	`, thirtyDaysAgo, thirtyDaysAgo).Scan(&storageThirtyDaysAgo)
	if err != nil {
		return 0, 0, err
	}

	// Storage now (files not deleted)
	err = d.db.QueryRow(`
		SELECT COALESCE(SUM(SizeBytes), 0)
		FROM Files
		WHERE DeletedAt = 0
	`).Scan(&storageNow)
	if err != nil {
		return 0, 0, err
	}

	return storageThirtyDaysAgo, storageNow, nil
}
