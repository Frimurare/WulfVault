// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// The file list API must report whether a file is password protected without
// handing out the password itself.
func TestAPIFilesDoesNotExposeFilePassword(t *testing.T) {
	if err := database.Initialize(t.TempDir()); err != nil {
		t.Fatalf("Initialize database: %v", err)
	}

	user := &models.User{Name: "API User", Email: "api.user@example.com", Password: "hash", IsActive: true}
	if err := database.DB.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	const filePassword = "s3cret-file-password"
	file := &database.FileInfo{
		Id:                "file-api-1",
		Name:              "protected.txt",
		Size:              "1 B",
		SHA1:              "sha",
		FilePasswordPlain: filePassword,
		UserId:            user.Id,
	}
	if err := database.DB.SaveFile(file); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	s := New(&config.Config{ServerURL: "https://vault.example.com", Port: "443"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	s.handleAPIFiles(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if strings.Contains(body, filePassword) {
		t.Errorf("response contains the file password: %s", body)
	}
	if strings.Contains(body, "file_password") {
		t.Errorf("response contains a file_password field: %s", body)
	}

	var payload struct {
		Files []map[string]interface{} `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Files) != 1 {
		t.Fatalf("got %d files, want 1", len(payload.Files))
	}
	if _, present := payload.Files[0]["file_password"]; present {
		t.Error("file entry still has a file_password key")
	}
	if hasPassword, _ := payload.Files[0]["has_password"].(bool); !hasPassword {
		t.Error("has_password = false, want true for a password protected file")
	}
}
