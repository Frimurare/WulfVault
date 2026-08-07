// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

func newAuditTestServer(t *testing.T) *Server {
	t.Helper()
	if err := database.Initialize(t.TempDir()); err != nil {
		t.Fatalf("Initialize database: %v", err)
	}
	return New(&config.Config{ServerURL: "https://vault.example.com", Port: "443"})
}

// auditEntries returns the logged entries for one action.
func auditEntries(t *testing.T, action string) []*database.AuditLogEntry {
	t.Helper()
	entries, err := database.DB.GetAuditLogs(&database.AuditLogFilter{Action: action, Limit: 10})
	if err != nil {
		t.Fatalf("GetAuditLogs(%s): %v", action, err)
	}
	return entries
}

func TestServerHasAuditLogger(t *testing.T) {
	s := New(&config.Config{})
	if s.audit == nil {
		t.Fatal("Server.audit is nil, audit logging is not wired up")
	}
}

func TestPasswordChangedIsAudited(t *testing.T) {
	s := newAuditTestServer(t)

	user := &models.User{Name: "Pwd User", Email: "pwd.user@example.com", Password: "hash", IsActive: true}
	if err := database.DB.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/change-password", nil)
	if err := s.audit.LogPasswordChanged(user, req); err != nil {
		t.Fatalf("LogPasswordChanged: %v", err)
	}

	entries := auditEntries(t, database.ActionPasswordChanged)
	if len(entries) != 1 {
		t.Fatalf("got %d PASSWORD_CHANGED entries, want 1", len(entries))
	}
	if entries[0].UserEmail != user.Email {
		t.Errorf("UserEmail = %q, want %q", entries[0].UserEmail, user.Email)
	}
}

// adminEditRequest builds the POST that the admin user form sends.
func adminEditRequest(userID int, name, email string, isActive bool) *http.Request {
	form := url.Values{
		"name":       {name},
		"email":      {email},
		"quota_mb":   {"1000"},
		"user_level": {"2"},
	}
	if isActive {
		form.Set("is_active", "1")
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/users/edit?id="+strconv.Itoa(userID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestAdminUserEditAuditsActivationChanges(t *testing.T) {
	s := newAuditTestServer(t)

	admin := &models.User{Name: "Admin", Email: "admin@example.com", Password: "hash", IsActive: true}
	if err := database.DB.CreateUser(admin); err != nil {
		t.Fatalf("CreateUser admin: %v", err)
	}
	target := &models.User{Name: "Target", Email: "target@example.com", Password: "hash", IsActive: true}
	if err := database.DB.CreateUser(target); err != nil {
		t.Fatalf("CreateUser target: %v", err)
	}

	// Deactivate.
	req := adminEditRequest(target.Id, target.Name, target.Email, false)
	req = req.WithContext(contextWithUser(req.Context(), admin))
	rec := httptest.NewRecorder()
	s.handleAdminUserEdit(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("deactivate status = %d, want %d (body: %s)", rec.Code, http.StatusSeeOther, rec.Body.String())
	}

	entries := auditEntries(t, database.ActionUserDeactivated)
	if len(entries) != 1 {
		t.Fatalf("got %d USER_DEACTIVATED entries, want 1", len(entries))
	}
	if entries[0].UserEmail != admin.Email {
		t.Errorf("actor = %q, want the admin %q", entries[0].UserEmail, admin.Email)
	}

	// Reactivate.
	req = adminEditRequest(target.Id, target.Name, target.Email, true)
	req = req.WithContext(contextWithUser(req.Context(), admin))
	rec = httptest.NewRecorder()
	s.handleAdminUserEdit(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("activate status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	if entries := auditEntries(t, database.ActionUserActivated); len(entries) != 1 {
		t.Fatalf("got %d USER_ACTIVATED entries, want 1", len(entries))
	}

	// An edit that leaves the activation state alone must not add entries.
	req = adminEditRequest(target.Id, "Renamed", target.Email, true)
	req = req.WithContext(contextWithUser(req.Context(), admin))
	rec = httptest.NewRecorder()
	s.handleAdminUserEdit(rec, req)

	if entries := auditEntries(t, database.ActionUserActivated); len(entries) != 1 {
		t.Errorf("got %d USER_ACTIVATED entries after an unrelated edit, want 1", len(entries))
	}
}
