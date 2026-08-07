// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

// splashRequest runs a request through the language middleware the way the real
// server does, so the test covers the s.tr(r) wiring and not just the template.
func splashRequest(t *testing.T, handler http.HandlerFunc, path, acceptLanguage string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Accept-Language", acceptLanguage)
	rec := httptest.NewRecorder()
	i18n.Middleware(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s returned %d, want %d (body: %s)", path, rec.Code, http.StatusOK, rec.Body.String())
	}
	return rec.Body.String()
}

// assertLanguage checks that a page speaks one language and not the other.
func assertLanguage(t *testing.T, page, lang string, wanted, unwanted []string) {
	t.Helper()

	if !strings.Contains(page, `<html lang="`+lang+`">`) {
		t.Errorf(`page is missing <html lang="%s">`, lang)
	}
	for _, marker := range wanted {
		if !strings.Contains(page, marker) {
			t.Errorf("expected the %s page to contain %q", lang, marker)
		}
	}
	for _, marker := range unwanted {
		if strings.Contains(page, marker) {
			t.Errorf("expected the %s page NOT to contain %q", lang, marker)
		}
	}
}

// splashTestFile stores a downloadable file, owned by a freshly created user,
// and returns it.
func splashTestFile(t *testing.T, id string, requireAuth bool, expireAt int64, password string) *database.FileInfo {
	t.Helper()

	owner := &models.User{Name: "Owner", Email: id + "@example.com", Password: "hash", IsActive: true}
	if err := database.DB.CreateUser(owner); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	file := &database.FileInfo{
		Id:                 id,
		Name:               "report.pdf",
		Size:               "2.4 MB",
		SizeBytes:          2400000,
		UploadDate:         time.Now().Unix(),
		DownloadsRemaining: 3,
		ExpireAt:           expireAt,
		RequireAuth:        requireAuth,
		FilePasswordPlain:  password,
		UserId:             owner.Id,
	}
	if err := database.DB.SaveFile(file); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	return file
}

func TestSplashPageRendersInTheRequestLanguage(t *testing.T) {
	s := newAuditTestServer(t)
	splashTestFile(t, "splash-sv-en", true, 0, "")

	swedish := splashRequest(t, s.handleSplashPage, "/s/splash-sv-en", "sv-SE,sv;q=0.9")
	assertLanguage(t, swedish, "sv",
		[]string{"Ladda ner filen", "Inloggning krävs", "Medan du väntar", "Storlek", "Nedladdningar", "Drivs av"},
		[]string{"Download file", "Authentication required", "Powered by"},
	)

	english := splashRequest(t, s.handleSplashPage, "/s/splash-sv-en", "en-GB,en;q=0.9")
	assertLanguage(t, english, "en",
		[]string{"Download file", "Authentication required", "While you wait", "Powered by"},
		[]string{"Ladda ner filen", "Inloggning krävs", "Drivs av"},
	)
}

func TestSplashPageShipsTheDownloadScriptTranslations(t *testing.T) {
	s := newAuditTestServer(t)
	splashTestFile(t, "splash-js", false, 0, "")

	swedish := splashRequest(t, s.handleSplashPage, "/s/splash-js", "sv")
	for _, marker := range []string{"window.WV_DOWNLOAD_I18N", `"downloading":"LADDAR NER`, `"close":"Stäng"`} {
		if !strings.Contains(swedish, marker) {
			t.Errorf("expected the Swedish splash page to contain %q", marker)
		}
	}

	english := splashRequest(t, s.handleSplashPage, "/s/splash-js", "en")
	if !strings.Contains(english, `"close":"Close"`) {
		t.Error("expected the English splash page to ship the English download strings")
	}
}

func TestExpiredSplashPageRendersInTheRequestLanguage(t *testing.T) {
	s := newAuditTestServer(t)
	splashTestFile(t, "splash-expired", false, time.Now().Add(-time.Hour).Unix(), "")

	swedish := splashRequest(t, s.handleSplashPage, "/s/splash-expired", "sv")
	assertLanguage(t, swedish, "sv",
		[]string{"Länken har gått ut", "Be avsändaren om en ny länk"},
		[]string{"The link has expired"},
	)

	english := splashRequest(t, s.handleSplashPage, "/s/splash-expired", "en")
	assertLanguage(t, english, "en",
		[]string{"The link has expired", "Ask the sender for a new link"},
		[]string{"Länken har gått ut"},
	)
}

func TestPasswordPromptPageRendersInTheRequestLanguage(t *testing.T) {
	s := newAuditTestServer(t)

	splashTestFile(t, "splash-password", false, 0, "hunter2")

	swedish := splashRequest(t, s.handleDownload, "/d/splash-password", "sv")
	assertLanguage(t, swedish, "sv",
		[]string{"Filen är skyddad med lösenord", "Lås upp", "Lösenord"},
		[]string{"This file is password protected"},
	)

	english := splashRequest(t, s.handleDownload, "/d/splash-password", "en")
	assertLanguage(t, english, "en",
		[]string{"This file is password protected", "Unlock"},
		[]string{"Filen är skyddad med lösenord"},
	)
}

func TestDownloadAuthPageRendersInTheRequestLanguage(t *testing.T) {
	s := newAuditTestServer(t)
	splashTestFile(t, "splash-auth", true, 0, "")

	swedish := splashRequest(t, s.handleDownload, "/d/splash-auth", "sv")
	assertLanguage(t, swedish, "sv",
		[]string{"Filen kräver inloggning", "Skapa konto eller logga in", "E-postadress", "Ditt namn"},
		[]string{"This file requires authentication"},
	)

	english := splashRequest(t, s.handleDownload, "/d/splash-auth", "en")
	assertLanguage(t, english, "en",
		[]string{"This file requires authentication", "Create account / log in", "Your name"},
		[]string{"Filen kräver inloggning"},
	)
}
