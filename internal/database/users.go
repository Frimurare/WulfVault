package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Frimurare/Sharecare/internal/models"
)

// CreateUser inserts a new user into the database
func (d *Database) CreateUser(user *models.User) error {
	if user.CreatedAt == 0 {
		user.CreatedAt = time.Now().Unix()
	}

	resetPw := 0
	if user.ResetPassword {
		resetPw = 1
	}
	isActive := 1
	if !user.IsActive {
		isActive = 0
	}

	result, err := d.db.Exec(`
		INSERT INTO Users (Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		                   StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Name, user.Email, user.Password, user.Permissions, user.UserLevel, user.LastOnline,
		resetPw, user.StorageQuotaMB, user.StorageUsedMB, user.CreatedAt, isActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = int(id)
	return nil
}

// GetUserByID retrieves a user by ID
func (d *Database) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive
		FROM Users WHERE Id = ?`, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive
		FROM Users WHERE Email = ?`, email).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	return user, nil
}

// GetUserByName retrieves a user by username
func (d *Database) GetUserByName(name string) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive
		FROM Users WHERE Name = ?`, name).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	return user, nil
}

// GetAllUsers returns all users
func (d *Database) GetAllUsers() ([]*models.User, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive
		FROM Users ORDER BY Userlevel ASC, LastOnline DESC, Name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var resetPw, isActive int

		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions,
			&user.UserLevel, &user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
			&user.CreatedAt, &isActive)
		if err != nil {
			return nil, err
		}

		user.ResetPassword = resetPw == 1
		user.IsActive = isActive == 1
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates an existing user
func (d *Database) UpdateUser(user *models.User) error {
	resetPw := 0
	if user.ResetPassword {
		resetPw = 1
	}
	isActive := 1
	if !user.IsActive {
		isActive = 0
	}

	_, err := d.db.Exec(`
		UPDATE Users SET Name = ?, Email = ?, Password = ?, Permissions = ?, Userlevel = ?,
		                 LastOnline = ?, ResetPassword = ?, StorageQuotaMB = ?, StorageUsedMB = ?,
		                 IsActive = ?
		WHERE Id = ?`,
		user.Name, user.Email, user.Password, user.Permissions, user.UserLevel, user.LastOnline,
		resetPw, user.StorageQuotaMB, user.StorageUsedMB, isActive, user.Id,
	)
	return err
}

// UpdateUserLastOnline updates the last online timestamp
func (d *Database) UpdateUserLastOnline(id int) error {
	_, err := d.db.Exec("UPDATE Users SET LastOnline = ? WHERE Id = ?", time.Now().Unix(), id)
	return err
}

// UpdateUserStorage updates a user's storage usage
func (d *Database) UpdateUserStorage(id int, storageUsedMB int64) error {
	_, err := d.db.Exec("UPDATE Users SET StorageUsedMB = ? WHERE Id = ?", storageUsedMB, id)
	return err
}

// DeleteUser deletes a user by ID
func (d *Database) DeleteUser(id int) error {
	_, err := d.db.Exec("DELETE FROM Users WHERE Id = ?", id)
	return err
}

// GetTotalUsers returns the count of all users
func (d *Database) GetTotalUsers() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users").Scan(&count)
	return count, err
}

// GetActiveUsers returns the count of active users
func (d *Database) GetActiveUsers() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE IsActive = 1").Scan(&count)
	return count, err
}

// GetUsersAddedThisMonth returns count of users (including download accounts) created this month
func (d *Database) GetUsersAddedThisMonth() (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var regularUsers, downloadAccounts int

	// Count regular users
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE CreatedAt >= ?", startOfMonth).Scan(&regularUsers)
	if err != nil {
		return 0, err
	}

	// Count download accounts
	err = d.db.QueryRow("SELECT COUNT(*) FROM DownloadAccounts WHERE CreatedAt >= ? AND DeletedAt = 0", startOfMonth).Scan(&downloadAccounts)
	if err != nil {
		return 0, err
	}

	return regularUsers + downloadAccounts, nil
}

// GetUsersRemovedThisMonth returns count of download accounts deleted this month
func (d *Database) GetUsersRemovedThisMonth() (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM DownloadAccounts WHERE DeletedAt >= ? AND DeletedAt < ?",
		startOfMonth, now.Unix()).Scan(&count)
	return count, err
}

// GetUserGrowthPercentage returns the user growth percentage for this month
func (d *Database) GetUserGrowthPercentage() (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var usersAtStart, usersNow int

	// Count users at start of month
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE CreatedAt < ?", startOfMonth).Scan(&usersAtStart)
	if err != nil {
		return 0, err
	}

	// Count users now
	totalNow, err := d.GetTotalUsers()
	if err != nil {
		return 0, err
	}
	usersNow = totalNow

	// Calculate percentage
	if usersAtStart == 0 {
		return 0, nil
	}

	growth := float64(usersNow-usersAtStart) / float64(usersAtStart) * 100
	return growth, nil
}
