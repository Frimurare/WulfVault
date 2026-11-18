// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"testing"

	"github.com/Frimurare/WulfVault/internal/models"
)

// Rows written by earlier releases (Userlevel / Permissions as plain numbers)
// must load into the current model with the same meaning.
func TestUserRowsFromEarlierReleases(t *testing.T) {
	d := newTestDB(t)
	if _, err := d.db.Exec(`INSERT INTO Users (Name, Email, Password, Permissions, Userlevel, CreatedAt)
		VALUES ('Old Admin', 'old-admin@example.com', 'hash', 255, 1, 1), ('Old User', 'old-user@example.com', 'hash', 0, 2, 1)`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	admin, err := d.GetUserByEmail("old-admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if admin.Role != models.RoleAdmin || !admin.IsAdmin() || !admin.CanEditOthersFiles() {
		t.Errorf("admin row lost rights: %+v", admin)
	}
	user, err := d.GetUserByEmail("old-user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != models.RoleUser || user.IsAdmin() || user.CanEditOthersFiles() {
		t.Errorf("user row gained rights: %+v", user)
	}
}

func TestUserRoundTrip(t *testing.T) {
	d := newTestDB(t)
	u := &models.User{Name: "Round Trip", Email: "rt@example.com", Password: "hash", IsActive: true,
		Role: models.RoleSuperAdmin, Permissions: models.PermAll, StorageQuotaMB: 500, Language: "sv"}
	if err := d.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	var level, perms int
	if err := d.db.QueryRow(`SELECT Userlevel, Permissions FROM Users WHERE Id = ?`, u.Id).Scan(&level, &perms); err != nil {
		t.Fatal(err)
	}
	if level != 0 || perms != 255 {
		t.Fatalf("stored values changed: Userlevel=%d Permissions=%d", level, perms)
	}
	got, err := d.GetUserByID(u.Id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != models.RoleSuperAdmin || got.Permissions != models.PermAll || got.StorageQuotaMB != 500 {
		t.Errorf("round trip mismatch: %+v", got)
	}
}
