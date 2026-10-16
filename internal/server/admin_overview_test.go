// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

func renderOverviewTestPage(t *testing.T, lang i18n.Lang) string {
	t.Helper()
	s := newAuditTestServer(t)
	s.config.CompanyName = "Example Vault"
	s.config.Version = "test"
	w := httptest.NewRecorder()
	s.renderAdminDashboard(w, i18n.For(lang), &models.User{Name: "<script>user</script>"},
		5, 4, 102, 0,
		0, 1024, 2048, 4096,
		512, 1024, 2048, 8192,
		2, 1, 25,
		3, 4, 1024, 2.5,
		80, 4,
		"<script>largest</script>", 2048, []string{"<b>user</b>", "unmatched"}, []int{3},
		[]string{"pdf", "unmatched"}, []int{2}, "Monday", 12, 1024, 2048,
		"<script>downloaded</script>", 8, 15*1024*1024*1024, 120*1024*1024*1024,
		[]DuplicateFile{{Name: "<img src=x onerror=alert(1)>", Size: "1 KB", Count: 2, FileIds: []string{"safe", "<script>id</script>"}}})
	if w.Code != 200 {
		t.Fatalf("render status = %d: %s", w.Code, w.Body.String())
	}
	return w.Body.String()
}

func TestAdminOverviewUsesSelectedLanguage(t *testing.T) {
	for _, tc := range []struct {
		lang i18n.Lang
		want []string
		omit []string
	}{
		{i18n.LangEN, []string{"Overview", "Data transfer", "Uploaded", "More statistics", "Monday", "Review duplicate files"}, []string{"Översikt", "Dataöverföring", "Granska dubbletter"}},
		{i18n.LangSV, []string{"Översikt", "Dataöverföring", "Uppladdat", "Mer statistik", "Måndag", "Granska dubbletter"}, []string{"Data transfer", "More statistics", "Monday"}},
	} {
		t.Run(string(tc.lang), func(t *testing.T) {
			page := renderOverviewTestPage(t, tc.lang)
			assertLanguage(t, page, string(tc.lang), tc.want, tc.omit)
			if strings.Contains(page, "admin.overview.") {
				t.Fatal("untranslated overview key rendered")
			}
			for _, path := range []string{"/lang?lang=en", "/lang?lang=sv", "/dashboard#uploadForm", "/admin/teams", "/admin/trash", "/admin/identity-providers", "/admin/sysmonitor-logs", "/logout"} {
				if !strings.Contains(page, path) {
					t.Errorf("missing navigation destination %q", path)
				}
			}
		})
	}
}

func TestAdminOverviewEscapesStoredContent(t *testing.T) {
	page := renderOverviewTestPage(t, i18n.LangEN)
	for _, unsafe := range []string{"<script>user", "<script>largest", "<script>downloaded", "<b>user</b>", "<img src=x", "<script>id"} {
		if strings.Contains(page, unsafe) {
			t.Errorf("stored content rendered as markup: %s", unsafe)
		}
	}
	for _, escaped := range []string{"&lt;script&gt;user", "&lt;script&gt;largest", "&lt;b&gt;user&lt;/b&gt;", "&lt;img src=x"} {
		if !strings.Contains(page, escaped) {
			t.Errorf("escaped content missing: %s", escaped)
		}
	}
	if strings.Contains(page, "cdn.tailwindcss.com") || strings.Contains(page, "unpkg.com") {
		t.Error("admin overview must not depend on third-party scripts")
	}
}

func TestDashboardStoragePercent(t *testing.T) {
	for _, tc := range []struct {
		used, available int64
		value           string
		known           bool
	}{
		{0, 0, "0", false},
		{1, -1, "0", false},
		{-1, 10, "0", false},
		{0, 100, "0.0", true},
		{100, 0, "100.0", true},
		{15, 120, "11.1", true},
		{1<<63 - 1, 1<<63 - 1, "50.0", true},
	} {
		value, known := dashboardStoragePercent(tc.used, tc.available)
		if value != tc.value || known != tc.known {
			t.Errorf("dashboardStoragePercent(%d, %d) = (%q, %v), want (%q, %v)", tc.used, tc.available, value, known, tc.value, tc.known)
		}
	}
}

// CI sets WV_UI_PREVIEW_DIR to render disposable, non-production fixtures for
// browser checks. Without it this test does not write persistent artifacts.
func TestAdminOverviewPreviewFixtures(t *testing.T) {
	dir := os.Getenv("WV_UI_PREVIEW_DIR")
	if dir == "" {
		t.Skip("browser preview output is not requested")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, lang := range []i18n.Lang{i18n.LangEN, i18n.LangSV} {
		page := renderOverviewTestPage(t, lang)
		page = strings.ReplaceAll(page, "&lt;script&gt;user&lt;/script&gt;", "Demo Admin")
		if err := os.WriteFile(filepath.Join(dir, string(lang)+".html"), []byte(page), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
