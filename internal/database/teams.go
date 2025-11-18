// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Frimurare/WulfVault/internal/models"
)

// CreateTeam inserts a new team into the database
func (d *Database) CreateTeam(team *models.Team) error {
	if team.CreatedAt == 0 {
		team.CreatedAt = time.Now().Unix()
	}

	isActive := 1
	if !team.IsActive {
		isActive = 0
	}

	result, err := d.db.Exec(`
		INSERT INTO Teams (Name, Description, CreatedBy, CreatedAt, StorageQuotaMB, StorageUsedMB, IsActive)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		team.Name, team.Description, team.CreatedBy, team.CreatedAt,
		team.StorageQuotaMB, team.StorageUsedMB, isActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	team.Id = int(id)

	// Automatically add the creator as owner
	member := &models.TeamMember{
		TeamId:   team.Id,
		UserId:   team.CreatedBy,
		Role:     models.TeamRoleOwner,
		JoinedAt: team.CreatedAt,
		AddedBy:  team.CreatedBy,
	}
	return d.AddTeamMember(member)
}

// GetTeamByID retrieves a team by ID
func (d *Database) GetTeamByID(id int) (*models.Team, error) {
	team := &models.Team{}
	var isActive int

	err := d.db.QueryRow(`
		SELECT Id, Name, Description, CreatedBy, CreatedAt, StorageQuotaMB, StorageUsedMB, IsActive
		FROM Teams WHERE Id = ?`, id).Scan(
		&team.Id, &team.Name, &team.Description, &team.CreatedBy, &team.CreatedAt,
		&team.StorageQuotaMB, &team.StorageUsedMB, &isActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}

	team.IsActive = isActive == 1
	return team, nil
}

// GetAllTeams returns all active teams
func (d *Database) GetAllTeams() ([]*models.Team, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Description, CreatedBy, CreatedAt, StorageQuotaMB, StorageUsedMB, IsActive
		FROM Teams WHERE IsActive = 1 ORDER BY Name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		var isActive int

		err := rows.Scan(&team.Id, &team.Name, &team.Description, &team.CreatedBy,
			&team.CreatedAt, &team.StorageQuotaMB, &team.StorageUsedMB, &isActive)
		if err != nil {
			return nil, err
		}

		team.IsActive = isActive == 1
		teams = append(teams, team)
	}

	return teams, rows.Err()
}

// GetTeamsByUser returns all teams that a user is a member of
func (d *Database) GetTeamsByUser(userId int) ([]*models.TeamWithMembers, error) {
	rows, err := d.db.Query(`
		SELECT t.Id, t.Name, t.Description, t.CreatedBy, t.CreatedAt,
		       t.StorageQuotaMB, t.StorageUsedMB, t.IsActive,
		       tm.Role,
		       (SELECT COUNT(*) FROM TeamMembers WHERE TeamId = t.Id) as MemberCount
		FROM Teams t
		INNER JOIN TeamMembers tm ON t.Id = tm.TeamId
		WHERE tm.UserId = ? AND t.IsActive = 1
		ORDER BY t.Name ASC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.TeamWithMembers
	for rows.Next() {
		team := &models.TeamWithMembers{}
		var isActive int

		err := rows.Scan(
			&team.Id, &team.Name, &team.Description, &team.CreatedBy,
			&team.CreatedAt, &team.StorageQuotaMB, &team.StorageUsedMB, &isActive,
			&team.UserRole, &team.MemberCount,
		)
		if err != nil {
			return nil, err
		}

		team.IsActive = isActive == 1
		teams = append(teams, team)
	}

	return teams, rows.Err()
}

// UpdateTeam updates team information
func (d *Database) UpdateTeam(team *models.Team) error {
	isActive := 1
	if !team.IsActive {
		isActive = 0
	}

	_, err := d.db.Exec(`
		UPDATE Teams
		SET Name = ?, Description = ?, StorageQuotaMB = ?, IsActive = ?
		WHERE Id = ?`,
		team.Name, team.Description, team.StorageQuotaMB, isActive, team.Id,
	)
	return err
}

// UpdateTeamStorage updates the storage used by a team
func (d *Database) UpdateTeamStorage(teamId int, storageUsedMB int64) error {
	_, err := d.db.Exec(`
		UPDATE Teams
		SET StorageUsedMB = ?
		WHERE Id = ?`,
		storageUsedMB, teamId,
	)
	return err
}

// DeleteTeam soft-deletes a team (sets IsActive to false)
func (d *Database) DeleteTeam(teamId int) error {
	_, err := d.db.Exec("UPDATE Teams SET IsActive = 0 WHERE Id = ?", teamId)
	return err
}

// AddTeamMember adds a user to a team
func (d *Database) AddTeamMember(member *models.TeamMember) error {
	if member.JoinedAt == 0 {
		member.JoinedAt = time.Now().Unix()
	}

	result, err := d.db.Exec(`
		INSERT INTO TeamMembers (TeamId, UserId, Role, JoinedAt, AddedBy)
		VALUES (?, ?, ?, ?, ?)`,
		member.TeamId, member.UserId, member.Role, member.JoinedAt, member.AddedBy,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	member.Id = int(id)
	return nil
}

// RemoveTeamMember removes a user from a team
func (d *Database) RemoveTeamMember(teamId, userId int) error {
	_, err := d.db.Exec("DELETE FROM TeamMembers WHERE TeamId = ? AND UserId = ?", teamId, userId)
	return err
}

// GetTeamMembers returns all members of a team
func (d *Database) GetTeamMembers(teamId int) ([]*models.TeamMember, error) {
	rows, err := d.db.Query(`
		SELECT tm.Id, tm.TeamId, tm.UserId, tm.Role, tm.JoinedAt, tm.AddedBy,
		       u.Name, u.Email
		FROM TeamMembers tm
		INNER JOIN Users u ON tm.UserId = u.Id
		WHERE tm.TeamId = ? AND u.IsActive = 1
		ORDER BY tm.Role ASC, u.Name ASC`, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.TeamMember
	for rows.Next() {
		member := &models.TeamMember{}
		err := rows.Scan(
			&member.Id, &member.TeamId, &member.UserId, &member.Role,
			&member.JoinedAt, &member.AddedBy, &member.UserName, &member.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// GetTeamMember returns a specific member's role in a team
func (d *Database) GetTeamMember(teamId, userId int) (*models.TeamMember, error) {
	member := &models.TeamMember{}
	err := d.db.QueryRow(`
		SELECT tm.Id, tm.TeamId, tm.UserId, tm.Role, tm.JoinedAt, tm.AddedBy,
		       u.Name, u.Email
		FROM TeamMembers tm
		INNER JOIN Users u ON tm.UserId = u.Id
		WHERE tm.TeamId = ? AND tm.UserId = ?`, teamId, userId).Scan(
		&member.Id, &member.TeamId, &member.UserId, &member.Role,
		&member.JoinedAt, &member.AddedBy, &member.UserName, &member.UserEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}

	return member, nil
}

// IsTeamMember checks if a user is a member of a team
func (d *Database) IsTeamMember(teamId, userId int) (bool, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM TeamMembers WHERE TeamId = ? AND UserId = ?",
		teamId, userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateTeamMemberRole updates a member's role in a team
func (d *Database) UpdateTeamMemberRole(teamId, userId int, role models.TeamRole) error {
	_, err := d.db.Exec(`
		UPDATE TeamMembers
		SET Role = ?
		WHERE TeamId = ? AND UserId = ?`,
		role, teamId, userId,
	)
	return err
}

// ShareFileToTeam shares a file with a team
func (d *Database) ShareFileToTeam(fileId string, teamId, sharedBy int) error {
	_, err := d.db.Exec(`
		INSERT INTO TeamFiles (FileId, TeamId, SharedBy, SharedAt)
		VALUES (?, ?, ?, ?)`,
		fileId, teamId, sharedBy, time.Now().Unix(),
	)
	return err
}

// UnshareFileFromTeam removes a file from a team
func (d *Database) UnshareFileFromTeam(fileId string, teamId int) error {
	_, err := d.db.Exec("DELETE FROM TeamFiles WHERE FileId = ? AND TeamId = ?", fileId, teamId)
	return err
}

// GetTeamFiles returns all files shared with a team
func (d *Database) GetTeamFiles(teamId int) ([]*models.TeamFile, error) {
	rows, err := d.db.Query(`
		SELECT Id, FileId, TeamId, SharedBy, SharedAt
		FROM TeamFiles
		WHERE TeamId = ?
		ORDER BY SharedAt DESC`, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.TeamFile
	for rows.Next() {
		file := &models.TeamFile{}
		err := rows.Scan(&file.Id, &file.FileId, &file.TeamId, &file.SharedBy, &file.SharedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

// GetFileTeams returns all teams a file is shared with
func (d *Database) GetFileTeams(fileId string) ([]*models.Team, error) {
	rows, err := d.db.Query(`
		SELECT t.Id, t.Name, t.Description, t.CreatedBy, t.CreatedAt,
		       t.StorageQuotaMB, t.StorageUsedMB, t.IsActive
		FROM Teams t
		INNER JOIN TeamFiles tf ON t.Id = tf.TeamId
		WHERE tf.FileId = ?`, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		var isActive int

		err := rows.Scan(&team.Id, &team.Name, &team.Description, &team.CreatedBy,
			&team.CreatedAt, &team.StorageQuotaMB, &team.StorageUsedMB, &isActive)
		if err != nil {
			return nil, err
		}

		team.IsActive = isActive == 1
		teams = append(teams, team)
	}

	return teams, rows.Err()
}

// CanUserAccessFile checks if a user can access a file (either owns it or is in a team it's shared with)
func (d *Database) CanUserAccessFile(fileId string, userId int) (bool, error) {
	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT 1 FROM Files WHERE Id = ? AND UserId = ?
			UNION
			SELECT 1 FROM TeamFiles tf
			INNER JOIN TeamMembers tm ON tf.TeamId = tm.TeamId
			WHERE tf.FileId = ? AND tm.UserId = ?
		)`, fileId, userId, fileId, userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFileTeamNames returns a map of fileId -> list of team names for batch lookup
func (d *Database) GetFileTeamNames(fileIds []string) (map[string][]string, error) {
	if len(fileIds) == 0 {
		return make(map[string][]string), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(fileIds))
	args := make([]interface{}, len(fileIds))
	for i, id := range fileIds {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT tf.FileId, t.Name
		FROM TeamFiles tf
		INNER JOIN Teams t ON tf.TeamId = t.Id
		WHERE tf.FileId IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY t.Name ASC`

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var fileId, teamName string
		if err := rows.Scan(&fileId, &teamName); err != nil {
			return nil, err
		}
		result[fileId] = append(result[fileId], teamName)
	}

	return result, rows.Err()
}

// GetFilesByUserWithTeams returns all files the user can access (own files + team files)
func (d *Database) GetFilesByUserWithTeams(userId int) ([]*FileInfo, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT f.Id, f.Name, f.Size, f.SHA1, f.PasswordHash, f.FilePasswordPlain, f.HotlinkId,
		       f.ContentType, f.AwsBucket, f.ExpireAtString, f.ExpireAt, f.PendingDeletion,
		       f.SizeBytes, f.UploadDate, f.DownloadsRemaining, f.DownloadCount, f.UserId, f.Comment,
		       f.UnlimitedDownloads, f.UnlimitedTime, f.RequireAuth, f.DeletedAt, f.DeletedBy
		FROM Files f
		LEFT JOIN TeamFiles tf ON f.Id = tf.FileId
		LEFT JOIN TeamMembers tm ON tf.TeamId = tm.TeamId
		WHERE f.DeletedAt = 0 AND (f.UserId = ? OR tm.UserId = ?)
		ORDER BY f.UploadDate DESC`, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FileInfo
	for rows.Next() {
		file := &FileInfo{}
		var passwordHash, filePasswordPlain, hotlinkId, awsBucket, expireAtString, comment sql.NullString
		var pendingDeletion, expireAt, deletedAt, deletedBy sql.NullInt64
		var unlimitedDownloads, unlimitedTime, requireAuth int

		err := rows.Scan(
			&file.Id, &file.Name, &file.Size, &file.SHA1, &passwordHash, &filePasswordPlain,
			&hotlinkId, &file.ContentType, &awsBucket, &expireAtString,
			&expireAt, &pendingDeletion, &file.SizeBytes, &file.UploadDate,
			&file.DownloadsRemaining, &file.DownloadCount, &file.UserId, &comment,
			&unlimitedDownloads, &unlimitedTime, &requireAuth, &deletedAt, &deletedBy,
		)
		if err != nil {
			return nil, err
		}

		file.PasswordHash = passwordHash.String
		file.FilePasswordPlain = filePasswordPlain.String
		file.HotlinkId = hotlinkId.String
		file.AwsBucket = awsBucket.String
		file.ExpireAtString = expireAtString.String
		file.ExpireAt = expireAt.Int64
		file.PendingDeletion = pendingDeletion.Int64
		file.Comment = comment.String
		file.UnlimitedDownloads = unlimitedDownloads == 1
		file.UnlimitedTime = unlimitedTime == 1
		file.RequireAuth = requireAuth == 1
		file.DeletedAt = deletedAt.Int64
		file.DeletedBy = int(deletedBy.Int64)

		files = append(files, file)
	}

	return files, rows.Err()
}

// GetTeamsForFile returns all teams that have access to a specific file
func (d *Database) GetTeamsForFile(fileId string) ([]*models.Team, error) {
	query := `
		SELECT t.Id, t.Name, t.Description, t.CreatedBy, t.CreatedAt,
		       t.StorageQuotaMB, t.StorageUsedMB, t.IsActive
		FROM Teams t
		INNER JOIN TeamFiles tf ON t.Id = tf.TeamId
		WHERE tf.FileId = ?
		ORDER BY t.Name ASC`

	rows, err := d.db.Query(query, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		var isActive int
		err := rows.Scan(
			&team.Id, &team.Name, &team.Description, &team.CreatedBy,
			&team.CreatedAt, &team.StorageQuotaMB, &team.StorageUsedMB, &isActive,
		)
		if err != nil {
			return nil, err
		}
		team.IsActive = isActive == 1
		teams = append(teams, team)
	}

	return teams, rows.Err()
}
