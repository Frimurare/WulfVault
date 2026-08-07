// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package i18n

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLookupReturnsLocalisedString(t *testing.T) {
	tests := []struct {
		lang Lang
		key  string
		want string
	}{
		{LangEN, "nav.logout", "Log out"},
		{LangSV, "nav.logout", "Logga ut"},
		{LangEN, "common.upload", "Upload"},
		{LangSV, "common.upload", "Ladda upp"},
		{LangSV, "common.expires", "Utgår"},
		{LangSV, "error.login_required", "Inloggning krävs"},
	}

	for _, tc := range tests {
		if got := For(tc.lang).T(tc.key); got != tc.want {
			t.Errorf("For(%q).T(%q) = %q, want %q", tc.lang, tc.key, got, tc.want)
		}
	}
}

func TestFallbackToEnglishForMissingKey(t *testing.T) {
	const key = "test.only_in_english"
	catalogs[LangEN][key] = "English only"
	defer delete(catalogs[LangEN], key)

	if got := For(LangSV).T(key); got != "English only" {
		t.Errorf("Swedish lookup of English-only key = %q, want %q", got, "English only")
	}
	if For(LangSV).Has(key) {
		t.Error("Has() must report false for a key that only exists in the fallback locale")
	}
}

func TestUnknownKeyRendersAsKey(t *testing.T) {
	const key = "test.does.not.exist.anywhere"
	for _, lang := range Supported() {
		if got := For(lang).T(key); got != key {
			t.Errorf("For(%q).T(%q) = %q, want the key itself", lang, key, got)
		}
	}
}

func TestEveryEnglishKeyIsTranslated(t *testing.T) {
	for _, key := range Keys(LangEN) {
		if _, ok := catalogs[LangSV][key]; !ok {
			t.Errorf("sv.json is missing key %q", key)
		}
	}
	for _, key := range Keys(LangSV) {
		if _, ok := catalogs[LangEN][key]; !ok {
			t.Errorf("en.json is missing key %q (English is the reference locale)", key)
		}
	}
}

func TestPlaceholdersMatchAcrossLocales(t *testing.T) {
	for _, key := range Keys(LangEN) {
		english := catalogs[LangEN][key]
		swedish := catalogs[LangSV][key]
		for _, placeholder := range placeholdersIn(english) {
			if !strings.Contains(swedish, placeholder) {
				t.Errorf("sv.json key %q is missing placeholder %s", key, placeholder)
			}
		}
		for _, placeholder := range placeholdersIn(swedish) {
			if !strings.Contains(english, placeholder) {
				t.Errorf("en.json key %q is missing placeholder %s", key, placeholder)
			}
		}
	}
}

// placeholdersIn extracts the {{name}} tokens from a translation.
func placeholdersIn(s string) []string {
	var found []string
	for {
		start := strings.Index(s, "{{")
		if start < 0 {
			return found
		}
		end := strings.Index(s[start:], "}}")
		if end < 0 {
			return found
		}
		found = append(found, s[start:start+end+2])
		s = s[start+end+2:]
	}
}

func TestInterpolation(t *testing.T) {
	tests := []struct {
		name string
		lang Lang
		key  string
		args []string
		want string
	}{
		{
			name: "single placeholder",
			lang: LangSV,
			key:  "lang.switch_to",
			args: []string{"language", "Engelska"},
			want: "Byt till Engelska",
		},
		{
			name: "two placeholders",
			lang: LangSV,
			key:  "dashboard.storage_used",
			args: []string{"used", "1,2 GB", "quota", "5 GB"},
			want: "1,2 GB av 5 GB använt",
		},
		{
			name: "unmatched placeholder is left visible",
			lang: LangEN,
			key:  "lang.switch_to",
			args: nil,
			want: "Switch to {{language}}",
		},
		{
			name: "trailing odd argument is ignored",
			lang: LangEN,
			key:  "lang.switch_to",
			args: []string{"language", "Swedish", "stray"},
			want: "Switch to Swedish",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := For(tc.lang).T(tc.key, tc.args...); got != tc.want {
				t.Errorf("T() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTHEscapesValuesButNotTheTranslation(t *testing.T) {
	got := For(LangEN).TH("login.invite_banner", "provider", `Acme <script>alert("x")</script>`)
	if !strings.Contains(got, "<strong>") {
		t.Errorf("TH() must keep markup from the catalogue, got %q", got)
	}
	if strings.Contains(got, "<script>") {
		t.Errorf("TH() must escape the interpolated value, got %q", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Errorf("TH() should HTML-escape the value, got %q", got)
	}
}

func TestTNChoosesSingularOrPlural(t *testing.T) {
	catalogs[LangEN]["test.one"] = "{{count}} file"
	catalogs[LangEN]["test.many"] = "{{count}} files"
	defer func() {
		delete(catalogs[LangEN], "test.one")
		delete(catalogs[LangEN], "test.many")
	}()

	if got := For(LangEN).TN("test.one", "test.many", 1); got != "1 file" {
		t.Errorf("TN(1) = %q, want %q", got, "1 file")
	}
	if got := For(LangEN).TN("test.one", "test.many", 3); got != "3 files" {
		t.Errorf("TN(3) = %q, want %q", got, "3 files")
	}
	if got := For(LangEN).TN("test.one", "test.many", 0); got != "0 files" {
		t.Errorf("TN(0) = %q, want %q", got, "0 files")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want Lang
		ok   bool
	}{
		{"sv", LangSV, true},
		{"SV", LangSV, true},
		{"sv-SE", LangSV, true},
		{"sv_SE", LangSV, true},
		{"  en-GB  ", LangEN, true},
		{"", "", false},
		{"de", "", false},
		{"-sv", "", false},
	}

	for _, tc := range tests {
		got, ok := Normalize(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("Normalize(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestFromAcceptLanguage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   Lang
		ok     bool
	}{
		{"plain swedish", "sv", LangSV, true},
		{"regional tag", "sv-SE,sv;q=0.9", LangSV, true},
		{"quality decides", "de;q=1.0,en;q=0.4,sv;q=0.8", LangSV, true},
		{"unsupported first, supported later", "de-DE,fr;q=0.9,en;q=0.5", LangEN, true},
		{"equal quality keeps client order", "en;q=0.8,sv;q=0.8", LangEN, true},
		{"q=0 means unacceptable", "sv;q=0,en;q=0.3", LangEN, true},
		{"nothing supported", "de,fr,es", "", false},
		{"empty header", "", "", false},
		{"whitespace only", "   ", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FromAcceptLanguage(tc.header)
			if got != tc.want || ok != tc.ok {
				t.Errorf("FromAcceptLanguage(%q) = (%q, %v), want (%q, %v)", tc.header, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestResolvePriorityOrder(t *testing.T) {
	tests := []struct {
		name           string
		userPref       string
		cookie         string
		acceptLanguage string
		serverDefault  string
		want           Lang
	}{
		{
			name:           "profile beats everything",
			userPref:       "sv",
			cookie:         "en",
			acceptLanguage: "en-GB",
			serverDefault:  "en",
			want:           LangSV,
		},
		{
			name:           "cookie beats header and server default",
			cookie:         "sv",
			acceptLanguage: "en-GB",
			serverDefault:  "en",
			want:           LangSV,
		},
		{
			name:           "header beats server default",
			acceptLanguage: "sv-SE,sv;q=0.9,en;q=0.8",
			serverDefault:  "en",
			want:           LangSV,
		},
		{
			name:           "server default when the header is unusable",
			acceptLanguage: "de-DE,fr;q=0.9",
			serverDefault:  "sv",
			want:           LangSV,
		},
		{
			name: "english when nothing is set",
			want: LangEN,
		},
		{
			name:          "unsupported values are skipped",
			userPref:      "de",
			cookie:        "fr",
			serverDefault: "sv",
			want:          LangSV,
		},
		{
			name:          "unsupported server default falls back to english",
			serverDefault: "de",
			want:          LangEN,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.userPref, tc.cookie, tc.acceptLanguage, tc.serverDefault)
			if got != tc.want {
				t.Errorf("Resolve() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestServerDefaultRoundTrip(t *testing.T) {
	original := ServerDefault()
	defer SetServerDefault(string(original))

	SetServerDefault("sv-SE")
	if got := ServerDefault(); got != LangSV {
		t.Errorf("ServerDefault() = %q, want %q", got, LangSV)
	}

	SetServerDefault("klingon")
	if got := ServerDefault(); got != DefaultLang {
		t.Errorf("unsupported code should reset to %q, got %q", DefaultLang, got)
	}
}

func TestMiddlewareResolvesFromCookieAndHeader(t *testing.T) {
	original := ServerDefault()
	defer SetServerDefault(string(original))
	SetServerDefault("en")

	var seen Lang
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = FromRequest(r).Lang()
	}))

	t.Run("cookie wins", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Header.Set("Accept-Language", "en-GB")
		req.AddCookie(&http.Cookie{Name: CookieName, Value: "sv"})
		handler.ServeHTTP(httptest.NewRecorder(), req)
		if seen != LangSV {
			t.Errorf("got %q, want %q", seen, LangSV)
		}
	})

	t.Run("header used without cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9")
		handler.ServeHTTP(httptest.NewRecorder(), req)
		if seen != LangSV {
			t.Errorf("got %q, want %q", seen, LangSV)
		}
	})

	t.Run("server default when nothing matches", func(t *testing.T) {
		SetServerDefault("sv")
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Header.Set("Accept-Language", "de-DE")
		handler.ServeHTTP(httptest.NewRecorder(), req)
		if seen != LangSV {
			t.Errorf("got %q, want %q", seen, LangSV)
		}
	})
}

func TestWithUserLanguageOverridesTheMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = req.WithContext(NewContext(req.Context(), For(LangEN)))

	if got := WithUserLanguage(req, "sv").Context(); FromContext(got).Lang() != LangSV {
		t.Errorf("profile language did not take effect, got %q", FromContext(got).Lang())
	}

	// An empty or unsupported profile setting must leave the request untouched,
	// so the cookie/header choice keeps applying.
	for _, pref := range []string{"", "  ", "de"} {
		if got := WithUserLanguage(req, pref); FromRequest(got).Lang() != LangEN {
			t.Errorf("WithUserLanguage(%q) changed the language to %q", pref, FromRequest(got).Lang())
		}
	}
}

func TestFromContextNeverReturnsNil(t *testing.T) {
	original := ServerDefault()
	defer SetServerDefault(string(original))
	SetServerDefault("sv")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req).Lang(); got != LangSV {
		t.Errorf("a request without a translator should fall back to the server default, got %q", got)
	}
	if got := FromRequest(nil).Lang(); got != LangSV {
		t.Errorf("FromRequest(nil) should fall back to the server default, got %q", got)
	}
}

func TestSetCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	SetCookie(rec, LangSV)

	result := rec.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one cookie, got %d", len(cookies))
	}
	if cookies[0].Name != CookieName || cookies[0].Value != "sv" {
		t.Errorf("got cookie %s=%s, want %s=sv", cookies[0].Name, cookies[0].Value, CookieName)
	}
	if cookies[0].Path != "/" {
		t.Errorf("cookie path = %q, want %q", cookies[0].Path, "/")
	}
}

func TestForUnsupportedLanguageFallsBackToDefault(t *testing.T) {
	if got := For("de").Lang(); got != DefaultLang {
		t.Errorf("For(\"de\").Lang() = %q, want %q", got, DefaultLang)
	}
	if got := ForCode("sv-SE").Lang(); got != LangSV {
		t.Errorf("ForCode(\"sv-SE\").Lang() = %q, want %q", got, LangSV)
	}
	var nilTranslator *Translator
	if got := nilTranslator.Lang(); got != DefaultLang {
		t.Errorf("nil translator should report %q, got %q", DefaultLang, got)
	}
}
