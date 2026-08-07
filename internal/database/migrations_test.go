// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Frimurare/WulfVault/internal/models"
)

// newTestDB returns a database backed by a temporary directory.
func newTestDB(t *testing.T) *Database {
	t.Helper()
	if err := Initialize(t.TempDir()); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return DB
}

// insertTestFile creates the Files row that DownloadLogs references.
func (d *Database) insertTestFile(t *testing.T, fileID string) {
	t.Helper()
	owner := &models.User{Name: "File Owner " + fileID, Email: "owner-" + fileID + "@example.com", Password: "hash", IsActive: true}
	if err := d.CreateUser(owner); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := d.db.Exec(`
		INSERT INTO Files (Id, Name, Size, SHA1, UserId) VALUES (?, 'file.txt', '1 B', 'sha', ?)`,
		fileID, owner.Id); err != nil {
		t.Fatalf("insert file: %v", err)
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func (d *Database) rowEmails(t *testing.T, table string, id int) (string, string) {
	t.Helper()
	var email, originalEmail string
	err := d.db.QueryRow(
		"SELECT Email, COALESCE(OriginalEmail, '') FROM "+table+" WHERE Id = ?", id).
		Scan(&email, &originalEmail)
	if err != nil {
		t.Fatalf("read %s row %d: %v", table, id, err)
	}
	return email, originalEmail
}

func TestSoftDeleteUserRemovesAddress(t *testing.T) {
	d := newTestDB(t)

	user := &models.User{
		Name:     "Test User",
		Email:    "Test.User@example.com",
		Password: "hash",
		IsActive: true,
	}
	if err := d.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := d.SoftDeleteUser(user.Id, "admin"); err != nil {
		t.Fatalf("SoftDeleteUser: %v", err)
	}

	email, originalEmail := d.rowEmails(t, "Users", user.Id)

	if want := anonymizedEmail(anonymizedUserPrefix, user.Id); email != want {
		t.Errorf("Email = %q, want %q", email, want)
	}
	if strings.Contains(strings.ToLower(email), "test.user") ||
		strings.Contains(strings.ToLower(email), "example.com") {
		t.Errorf("Email %q still contains the original address", email)
	}
	if want := sha256Hex("test.user@example.com"); originalEmail != want {
		t.Errorf("OriginalEmail = %q, want fingerprint %q", originalEmail, want)
	}
}

func TestSoftDeleteDownloadAccountRemovesAddress(t *testing.T) {
	d := newTestDB(t)

	account := &models.DownloadAccount{
		Name:     "Downloader",
		Email:    "downloader@example.com",
		Password: "hash",
		IsActive: true,
	}
	if err := d.CreateDownloadAccount(account); err != nil {
		t.Fatalf("CreateDownloadAccount: %v", err)
	}

	d.insertTestFile(t, "file-1")
	if _, err := d.db.Exec(`
		INSERT INTO DownloadLogs (FileId, DownloadAccountId, Email, DownloadedAt)
		VALUES ('file-1', ?, ?, 1)`, account.Id, account.Email); err != nil {
		t.Fatalf("insert download log: %v", err)
	}

	if err := d.SoftDeleteDownloadAccount(account.Id, "user"); err != nil {
		t.Fatalf("SoftDeleteDownloadAccount: %v", err)
	}

	email, originalEmail := d.rowEmails(t, "DownloadAccounts", account.Id)
	want := anonymizedEmail(anonymizedDownloadPrefix, account.Id)
	if email != want {
		t.Errorf("Email = %q, want %q", email, want)
	}
	if originalEmail != sha256Hex("downloader@example.com") {
		t.Errorf("OriginalEmail = %q, want fingerprint", originalEmail)
	}

	var logEmail string
	if err := d.db.QueryRow(
		"SELECT Email FROM DownloadLogs WHERE DownloadAccountId = ?", account.Id).Scan(&logEmail); err != nil {
		t.Fatalf("read download log: %v", err)
	}
	if logEmail != want {
		t.Errorf("DownloadLogs.Email = %q, want %q", logEmail, want)
	}
}

func TestRewriteLegacyAnonymizedRows(t *testing.T) {
	d := newTestDB(t)

	// A user and a download account anonymized by the old format, where the
	// address was embedded in the placeholder and copied to OriginalEmail.
	legacyUserEmail := "legacy.user@example.com"
	res, err := d.db.Exec(`
		INSERT INTO Users (Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		                   StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive, OriginalEmail, DeletedAt, DeletedBy)
		VALUES ('Legacy User', ?, '', 0, 2, 0, 0, 0, 0, 1, 0, ?, 1, 'admin')`,
		"deleted_user_"+legacyUserEmail+"@deleted.local", legacyUserEmail)
	if err != nil {
		t.Fatalf("insert legacy user: %v", err)
	}
	legacyUserID, _ := res.LastInsertId()

	legacyAccountEmail := "legacy.downloader@example.com"
	res, err = d.db.Exec(`
		INSERT INTO DownloadAccounts (Name, Email, Password, CreatedAt, OriginalEmail, DeletedAt, DeletedBy)
		VALUES ('Legacy Downloader', ?, '', 1, ?, 1, 'admin')`,
		"deleted_download_"+legacyAccountEmail+"@deleted.local", legacyAccountEmail)
	if err != nil {
		t.Fatalf("insert legacy download account: %v", err)
	}
	legacyAccountID, _ := res.LastInsertId()

	d.insertTestFile(t, "file-1")
	if _, err := d.db.Exec(`
		INSERT INTO DownloadLogs (FileId, DownloadAccountId, Email, DownloadedAt)
		VALUES ('file-1', ?, ?, 1)`,
		legacyAccountID, "deleted_download_"+legacyAccountEmail+"@deleted.local"); err != nil {
		t.Fatalf("insert legacy download log: %v", err)
	}

	// A row that was never anonymized must be left alone.
	activeUser := &models.User{Name: "Active", Email: "active@example.com", Password: "hash", IsActive: true}
	if err := d.CreateUser(activeUser); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := d.rewriteLegacyAnonymizedRows(); err != nil {
		t.Fatalf("rewriteLegacyAnonymizedRows: %v", err)
	}

	email, originalEmail := d.rowEmails(t, "Users", int(legacyUserID))
	if want := anonymizedEmail(anonymizedUserPrefix, int(legacyUserID)); email != want {
		t.Errorf("legacy user Email = %q, want %q", email, want)
	}
	if strings.Contains(email, "legacy.user") || strings.Contains(originalEmail, "legacy.user") {
		t.Errorf("legacy user row still contains the address: %q / %q", email, originalEmail)
	}
	if want := sha256Hex(legacyUserEmail); originalEmail != want {
		t.Errorf("legacy user OriginalEmail = %q, want fingerprint %q", originalEmail, want)
	}

	accountEmail, accountOriginal := d.rowEmails(t, "DownloadAccounts", int(legacyAccountID))
	wantAccount := anonymizedEmail(anonymizedDownloadPrefix, int(legacyAccountID))
	if accountEmail != wantAccount {
		t.Errorf("legacy account Email = %q, want %q", accountEmail, wantAccount)
	}
	if accountOriginal != sha256Hex(legacyAccountEmail) {
		t.Errorf("legacy account OriginalEmail = %q, want fingerprint", accountOriginal)
	}

	var logEmail string
	if err := d.db.QueryRow(
		"SELECT Email FROM DownloadLogs WHERE DownloadAccountId = ?", legacyAccountID).Scan(&logEmail); err != nil {
		t.Fatalf("read download log: %v", err)
	}
	if logEmail != wantAccount {
		t.Errorf("legacy DownloadLogs.Email = %q, want %q", logEmail, wantAccount)
	}

	activeEmail, _ := d.rowEmails(t, "Users", activeUser.Id)
	if activeEmail != "active@example.com" {
		t.Errorf("active user Email was rewritten: %q", activeEmail)
	}

	// Idempotent: a second pass must not change anything.
	if err := d.rewriteLegacyAnonymizedRows(); err != nil {
		t.Fatalf("second rewriteLegacyAnonymizedRows: %v", err)
	}
	secondEmail, secondOriginal := d.rowEmails(t, "Users", int(legacyUserID))
	if secondEmail != email || secondOriginal != originalEmail {
		t.Errorf("second pass changed the row: %q/%q -> %q/%q", email, originalEmail, secondEmail, secondOriginal)
	}
}

func TestRewriteLegacyAnonymizedRowsOnCleanDatabase(t *testing.T) {
	d := newTestDB(t)

	user := &models.User{Name: "Clean", Email: "clean@example.com", Password: "hash", IsActive: true}
	if err := d.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := d.rewriteLegacyAnonymizedRows(); err != nil {
		t.Fatalf("rewriteLegacyAnonymizedRows: %v", err)
	}

	email, originalEmail := d.rowEmails(t, "Users", user.Id)
	if email != "clean@example.com" || originalEmail != "" {
		t.Errorf("clean row modified: %q / %q", email, originalEmail)
	}
}

func TestEmailFingerprintIsStableAndNotReversible(t *testing.T) {
	fingerprint := emailFingerprint(" User@Example.COM ")
	if fingerprint != sha256Hex("user@example.com") {
		t.Errorf("fingerprint = %q, want normalised SHA-256", fingerprint)
	}
	if strings.Contains(fingerprint, "@") || strings.Contains(strings.ToLower(fingerprint), "user") {
		t.Errorf("fingerprint %q leaks the address", fingerprint)
	}
	if emailFingerprint("") != "" {
		t.Errorf("empty address should produce an empty fingerprint")
	}
}
