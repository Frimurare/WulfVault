// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

func TestApplicationShellKeepsRoleNavigationSeparate(t *testing.T) {
	s := newAuditTestServer(t)
	for _, lang := range []i18n.Lang{i18n.LangEN, i18n.LangSV} {
		tr := i18n.For(lang)
		admin := s.getAdminHeaderHTML("", tr)
		member := s.getHeaderHTML(&models.User{UserLevel: models.UserLevelUser, Name: "<script>test</script>"}, false, tr)
		download := s.getDownloadUserHeaderHTML(tr)
		if !strings.Contains(admin, "/admin/users") || !strings.Contains(admin, "/admin/identity-providers") {
			t.Fatal("admin destinations are missing")
		}
		for _, page := range []string{member, download} {
			if strings.Contains(page, `href="/admin`) {
				t.Fatal("non-admin shell exposes admin destinations")
			}
			if !strings.Contains(page, `class="wv-navigation"`) || !strings.Contains(page, `hreflang="sv"`) {
				t.Fatal("shared navigation or language switcher is missing")
			}
		}
		if strings.Contains(member, "<script>test</script>") || !strings.Contains(member, "&lt;script&gt;test&lt;/script&gt;") {
			t.Fatal("shared account name was not escaped")
		}
		if !strings.Contains(download, "/download/logout") || strings.Contains(download, `href="/logout"`) {
			t.Fatal("download account must use its own logout endpoint")
		}
	}
}

// Render real page templates, not hand-written mockups, for route-wide browser
// validation. All users, files, settings and accounts here are disposable.
func TestApplicationPageFixtures(t *testing.T) {
	s := newAuditTestServer(t)
	s.config.CompanyName = "Example Vault"
	s.config.Version = "test"
	s.config.DataDir = t.TempDir()
	s.config.UploadsDir = t.TempDir()
	user := &models.User{Name: "Demo Admin", Email: "admin@example.com", Password: "hash", IsActive: true, UserLevel: models.UserLevelAdmin, StorageQuotaMB: 5000}
	if err := database.DB.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	member := &models.User{Name: "Demo Member", Email: "member@example.com", Password: "hash", IsActive: true, UserLevel: models.UserLevelUser, StorageQuotaMB: 5000}
	if err := database.DB.CreateUser(member); err != nil {
		t.Fatal(err)
	}
	file := &database.FileInfo{Id: "demo-report", Name: "Quarterly-report-with-a-long-filename.pdf", UserId: user.Id, Size: "4 KB", SizeBytes: 4096, UploadDate: time.Now().Unix(), DownloadsRemaining: -1, UnlimitedDownloads: true, UnlimitedTime: true, RequireAuth: true, Comment: "Shared document for review"}
	if err := database.DB.SaveFile(file); err != nil {
		t.Fatal(err)
	}
	account := &models.DownloadAccount{Id: 1, Name: "Demo Recipient", Email: "recipient@example.com", IsActive: true}
	teamRows := []struct {
		*models.Team
		MemberCount int
	}{
		{Team: &models.Team{Id: 1, Name: "Design team", Description: "Shared documents", IsActive: true, StorageQuotaMB: 10000}, MemberCount: 3},
	}
	outputDir := os.Getenv("WV_UI_PREVIEW_DIR")
	for _, lang := range []i18n.Lang{i18n.LangEN, i18n.LangSV} {
		tr := i18n.For(lang)
		cases := []struct {
			name   string
			public bool
			render func(http.ResponseWriter)
		}{
			{"users", false, func(w http.ResponseWriter) {
				s.renderAdminUsers(w, tr, []*models.User{user, member}, nil, &database.UserFilter{Limit: 50}, 2, &database.DownloadAccountFilter{Limit: 50}, 0)
			}},
			{"user-edit", false, func(w http.ResponseWriter) { s.renderAdminUserForm(w, tr, member, "") }},
			{"recipient-edit", false, func(w http.ResponseWriter) { s.renderAdminDownloadAccountForm(w, tr, account, "") }},
			{"files", false, func(w http.ResponseWriter) { s.renderAdminFiles(w, tr, []*database.FileInfo{file}, 4096) }},
			{"duplicates", false, func(w http.ResponseWriter) { s.renderAdminDuplicates(w, tr, []*database.FileInfo{file}, 1, 50, 0, 1) }},
			{"trash", false, func(w http.ResponseWriter) { s.renderAdminTrash(w, tr, nil) }},
			{"branding", false, func(w http.ResponseWriter) { s.renderAdminBranding(w, tr, "") }},
			{"server-settings", false, func(w http.ResponseWriter) { s.renderAdminSettings(w, tr, "") }},
			{"identity", false, func(w http.ResponseWriter) { s.renderIdentityProviders(w, tr, "", "") }},
			{"email", false, func(w http.ResponseWriter) {
				s.renderEmailSettingsPage(w, tr, false, false, false, false, false, false, false, false, false, false, "", "", "", "", "", "", "", "", "", "")
			}},
			{"audit", false, func(w http.ResponseWriter) { s.renderAdminAuditLogsPage(w, tr) }},
			{"logs", false, func(w http.ResponseWriter) { s.renderAdminServerLogsPage(w, tr) }},
			{"monitor", false, func(w http.ResponseWriter) { s.renderAdminSysMonitorLogsPage(w, tr) }},
			{"admin-teams", false, func(w http.ResponseWriter) { s.renderAdminTeams(w, tr, teamRows) }},
			{"teams", false, func(w http.ResponseWriter) { s.renderUserTeams(w, tr, member, nil) }},
			{"team-files", false, func(w http.ResponseWriter) { s.renderTeamFiles(w, tr, member, &models.Team{Id: 1, Name: "Demo Team"}) }},
			{"my-files", false, func(w http.ResponseWriter) { s.renderUserDashboard(w, tr, user) }},
			{"settings", false, func(w http.ResponseWriter) { s.renderUserSettingsPage(w, tr, user, 0) }},
			{"account", false, func(w http.ResponseWriter) { s.renderUserAccountSettings(w, tr, user, "") }},
			{"downloads", false, func(w http.ResponseWriter) {
				s.renderDownloadDashboard(w, tr, account, nil, []*database.FileInfo{file})
			}},
			{"download-password", false, func(w http.ResponseWriter) { s.renderDownloadChangePasswordPage(w, tr, account, "", false) }},
			{"download-account", false, func(w http.ResponseWriter) { s.renderDownloadAccountGDPRPage(w, tr, account, "") }},
			{"login", true, func(w http.ResponseWriter) {
				r := httptest.NewRequest("GET", "/login", nil)
				r.Header.Set("Accept-Language", string(lang))
				s.renderLoginPage(w, r, "")
			}},
			{"forgot-password", true, func(w http.ResponseWriter) {
				r := httptest.NewRequest("GET", "/forgot-password", nil)
				r.Header.Set("Accept-Language", string(lang))
				s.renderForgotPasswordPage(w, r, "", false)
			}},
			{"download-splash", true, func(w http.ResponseWriter) { s.renderSplashPage(w, tr, file) }},
		}
		for _, tc := range cases {
			t.Run(tc.name+"-"+string(lang), func(t *testing.T) {
				w := httptest.NewRecorder()
				tc.render(w)
				page := w.Body.String()
				if w.Code != http.StatusOK || !strings.Contains(page, "/static/css/app-shell.css") {
					t.Fatalf("page did not render with the shared theme: status %d", w.Code)
				}
				if !tc.public {
					if strings.Count(page, `class="wv-sidebar"`) != 1 || !strings.Contains(page, `id="main"`) {
						t.Fatal("page must have one shared sidebar and a skip-link destination")
					}
					if !strings.Contains(page, `<html lang="`+string(lang)+`">`) {
						t.Fatal("page document language does not match the translator")
					}
				}
				if lang == i18n.LangSV {
					var required []string
					switch tc.name {
					case "admin-teams":
						required = []string{"Hantera team", "Skapa team", "Medlemmar", "Redigera", "Teamnamn"}
					case "files":
						required = []string{"Alla filer", "Antal filer", "Kopiera", "Ta bort", "Nedladdningshistorik"}
					case "server-settings":
						required = []string{"Anslutning och språk", "Lagring och uppladdning", "Lagringstider och loggar", "Spara inställningar"}
					case "my-files":
						required = []string{"Delningslänkar", "Inga uppladdningsförfrågningar ännu"}
					}
					for _, marker := range required {
						if !strings.Contains(page, marker) {
							t.Errorf("Swedish %s page is missing %q", tc.name, marker)
						}
					}
				}
				if strings.Contains(page, "%!") {
					t.Fatal("rendered page contains a formatting error")
				}
				if outputDir != "" {
					if err := os.MkdirAll(outputDir, 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(outputDir, tc.name+"-"+string(lang)+".html"), []byte(page), 0644); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
