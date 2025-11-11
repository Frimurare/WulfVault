package database

import (
	"log"
	"time"
)

// RunMigrations applies any pending database migrations
func (d *Database) RunMigrations() error {
	// Add soft delete columns to Users table if they don't exist
	if err := d.addColumnIfNotExists("Users", "DeletedAt", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "DeletedBy", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "OriginalEmail", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	// Add soft delete columns to DownloadAccounts table if they don't exist
	if err := d.addColumnIfNotExists("DownloadAccounts", "DeletedAt", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("DownloadAccounts", "DeletedBy", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("DownloadAccounts", "OriginalEmail", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// addColumnIfNotExists adds a column to a table if it doesn't already exist
func (d *Database) addColumnIfNotExists(tableName, columnName, columnDef string) error {
	// Check if column exists
	query := `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	var count int
	err := d.db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return err
	}

	// If column doesn't exist, add it
	if count == 0 {
		alterSQL := "ALTER TABLE " + tableName + " ADD COLUMN " + columnName + " " + columnDef
		_, err := d.db.Exec(alterSQL)
		if err != nil {
			return err
		}
		log.Printf("Added column %s to table %s", columnName, tableName)
	}

	return nil
}

// SoftDeleteUser marks a user as deleted without removing data (GDPR-compliant soft delete)
func (d *Database) SoftDeleteUser(userId int, deletedBy string) error {
	// Get user to store original email
	user, err := d.GetUserByID(userId)
	if err != nil {
		return err
	}

	// If already deleted, skip
	if user.DeletedAt > 0 {
		return nil
	}

	// Anonymize email but keep original
	anonymizedEmail := "deleted_user_" + user.Email + "@deleted.local"

	_, err = d.db.Exec(`
		UPDATE Users
		SET Email = ?, OriginalEmail = ?, DeletedAt = ?, DeletedBy = ?, IsActive = 0
		WHERE Id = ?`,
		anonymizedEmail, user.Email, currentTimestamp(), deletedBy, userId)

	return err
}

// SoftDeleteDownloadAccount marks a download account as deleted (GDPR-compliant soft delete)
func (d *Database) SoftDeleteDownloadAccount(accountId int, deletedBy string) error {
	// Get account to store original email
	account, err := d.GetDownloadAccountByID(accountId)
	if err != nil {
		return err
	}

	// If already deleted, skip
	if account.DeletedAt > 0 {
		return nil
	}

	// Anonymize email but keep original
	anonymizedEmail := "deleted_download_" + account.Email + "@deleted.local"

	_, err = d.db.Exec(`
		UPDATE DownloadAccounts
		SET Email = ?, OriginalEmail = ?, DeletedAt = ?, DeletedBy = ?, IsActive = 0
		WHERE Id = ?`,
		anonymizedEmail, account.Email, currentTimestamp(), deletedBy, accountId)

	// Also anonymize download logs
	_, _ = d.db.Exec(`
		UPDATE DownloadLogs
		SET Email = ?
		WHERE DownloadAccountId = ?`,
		anonymizedEmail, accountId)

	return err
}

// PermanentlyDeleteOldUsers permanently deletes users that have been soft-deleted for more than 90 days
func (d *Database) PermanentlyDeleteOldUsers(daysOld int) (int, error) {
	if daysOld <= 0 {
		daysOld = 90
	}

	cutoffTime := currentTimestamp() - int64(daysOld*24*60*60)

	// Get list of users to delete for logging
	rows, err := d.db.Query(`
		SELECT Id, Email, OriginalEmail, DeletedBy
		FROM Users
		WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var id int
		var email, originalEmail, deletedBy string
		if err := rows.Scan(&id, &email, &originalEmail, &deletedBy); err == nil {
			log.Printf("Permanently deleting user: ID=%d, OriginalEmail=%s, DeletedBy=%s", id, originalEmail, deletedBy)
			deletedCount++
		}
	}

	// Permanently delete users
	result, err := d.db.Exec(`DELETE FROM Users WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// PermanentlyDeleteOldDownloadAccounts permanently deletes download accounts that have been soft-deleted for more than 90 days
func (d *Database) PermanentlyDeleteOldDownloadAccounts(daysOld int) (int, error) {
	if daysOld <= 0 {
		daysOld = 90
	}

	cutoffTime := currentTimestamp() - int64(daysOld*24*60*60)

	// Get list of accounts to delete for logging
	rows, err := d.db.Query(`
		SELECT Id, Email, OriginalEmail, DeletedBy
		FROM DownloadAccounts
		WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var id int
		var email, originalEmail, deletedBy string
		if err := rows.Scan(&id, &email, &originalEmail, &deletedBy); err == nil {
			log.Printf("Permanently deleting download account: ID=%d, OriginalEmail=%s, DeletedBy=%s", id, originalEmail, deletedBy)
			deletedCount++
		}
	}

	// Delete download logs for these accounts first
	_, _ = d.db.Exec(`
		DELETE FROM DownloadLogs
		WHERE DownloadAccountId IN (
			SELECT Id FROM DownloadAccounts WHERE DeletedAt > 0 AND DeletedAt < ?
		)`, cutoffTime)

	// Permanently delete accounts
	result, err := d.db.Exec(`DELETE FROM DownloadAccounts WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// currentTimestamp returns the current Unix timestamp
func currentTimestamp() int64 {
	return time.Now().Unix()
}
