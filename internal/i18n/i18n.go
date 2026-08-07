// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// Package i18n provides the translation catalogues and the per-request
// language resolution used by the WulfVault web interface.
//
// The catalogues are flat key/value JSON files under locales/ and are compiled
// into the binary with go:embed, so a deployment never needs extra files on
// disk. English is the reference locale: any key missing from another locale
// falls back to English, and a key missing everywhere is rendered as the key
// itself so untranslated strings are obvious rather than blank.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

// Lang is a supported language code (ISO 639-1, lowercase).
type Lang string

const (
	// LangEN is English, the reference locale used for fallback.
	LangEN Lang = "en"
	// LangSV is Swedish.
	LangSV Lang = "sv"
)

// DefaultLang is used when nothing else resolves to a supported language.
const DefaultLang = LangEN

// CookieName is the cookie holding an explicit language choice.
const CookieName = "wv_lang"

// supported lists the languages in the order they are offered in the UI.
var supported = []Lang{LangEN, LangSV}

// catalogs holds the loaded translations, keyed by language.
var catalogs = map[Lang]map[string]string{}

// translators caches one Translator per language; they are immutable and safe
// for concurrent use.
var translators = map[Lang]*Translator{}

// serverDefault is the administrator-configured fallback language. It is read
// on every request, so it is guarded by a mutex.
var (
	serverDefaultMutex sync.RWMutex
	serverDefault      = DefaultLang
)

func init() {
	for _, lang := range supported {
		data, err := localeFS.ReadFile("locales/" + string(lang) + ".json")
		if err != nil {
			// The files are embedded at build time, so this can only happen if
			// a locale was added to supported without adding its file.
			panic(fmt.Sprintf("i18n: missing embedded locale %s: %v", lang, err))
		}
		var catalog map[string]string
		if err := json.Unmarshal(data, &catalog); err != nil {
			panic(fmt.Sprintf("i18n: invalid locale file %s.json: %v", lang, err))
		}
		catalogs[lang] = catalog
		translators[lang] = &Translator{lang: lang}
	}
}

// Supported returns the languages offered by the UI, in display order.
func Supported() []Lang {
	out := make([]Lang, len(supported))
	copy(out, supported)
	return out
}

// IsSupported reports whether code names a language WulfVault ships.
func IsSupported(code string) bool {
	_, ok := Normalize(code)
	return ok
}

// Normalize maps a language tag to a supported language. It accepts full tags
// such as "sv-SE" or "en_GB" and ignores surrounding whitespace and case.
// The second return value is false if the tag is empty or unsupported.
func Normalize(code string) (Lang, bool) {
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return "", false
	}
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	for _, lang := range supported {
		if Lang(code) == lang {
			return lang, true
		}
	}
	return "", false
}

// SetServerDefault sets the instance-wide fallback language. An empty or
// unsupported value resets it to DefaultLang.
func SetServerDefault(code string) {
	lang, ok := Normalize(code)
	if !ok {
		lang = DefaultLang
	}
	serverDefaultMutex.Lock()
	serverDefault = lang
	serverDefaultMutex.Unlock()
}

// ServerDefault returns the instance-wide fallback language.
func ServerDefault() Lang {
	serverDefaultMutex.RLock()
	defer serverDefaultMutex.RUnlock()
	return serverDefault
}

// Translator renders keys in one language. It is immutable and safe for
// concurrent use; obtain one with For, FromContext or FromRequest.
type Translator struct {
	lang Lang
}

// For returns the translator for lang, falling back to DefaultLang if lang is
// not supported.
func For(lang Lang) *Translator {
	if t, ok := translators[lang]; ok {
		return t
	}
	return translators[DefaultLang]
}

// ForCode is For with a plain string, accepting tags such as "sv-SE".
func ForCode(code string) *Translator {
	if lang, ok := Normalize(code); ok {
		return For(lang)
	}
	return For(DefaultLang)
}

// Lang returns the translator's language code.
func (t *Translator) Lang() Lang {
	if t == nil {
		return DefaultLang
	}
	return t.lang
}

// Is reports whether the translator renders the given language.
func (t *Translator) Is(lang Lang) bool {
	return t.Lang() == lang
}

// T looks up key and interpolates the given placeholders.
//
// args are alternating placeholder names and values, so
//
//	tr.T("files.expires_in", "days", "7")
//
// replaces {{days}} in the translated string with "7". A trailing odd argument
// is ignored, and placeholders without a matching argument are left in place so
// the mistake is visible in the UI.
//
// The returned string is not HTML-escaped: the catalogue is trusted content and
// may contain markup. Escape untrusted values yourself, or use TH.
func (t *Translator) T(key string, args ...string) string {
	return interpolate(t.lookup(key), args)
}

// TH is T with the interpolated values HTML-escaped. The translated string
// itself is still trusted, so markup in the catalogue keeps working.
func (t *Translator) TH(key string, args ...string) string {
	escaped := make([]string, len(args))
	for i, arg := range args {
		if i%2 == 1 {
			arg = html.EscapeString(arg)
		}
		escaped[i] = arg
	}
	return interpolate(t.lookup(key), escaped)
}

// TN picks between a singular and a plural key based on n and interpolates
// {{count}} with n. Swedish and English share the same one/other split, which
// is all WulfVault needs.
func (t *Translator) TN(singularKey, pluralKey string, n int, args ...string) string {
	key := pluralKey
	if n == 1 {
		key = singularKey
	}
	return t.T(key, append(args, "count", strconv.Itoa(n))...)
}

// Has reports whether key exists in this translator's own catalogue, ignoring
// the English fallback.
func (t *Translator) Has(key string) bool {
	_, ok := catalogs[t.Lang()][key]
	return ok
}

// lookup resolves key in the translator's catalogue, then in English, and
// finally returns the key itself.
func (t *Translator) lookup(key string) string {
	if value, ok := catalogs[t.Lang()][key]; ok {
		return value
	}
	if value, ok := catalogs[DefaultLang][key]; ok {
		return value
	}
	return key
}

// Keys returns the sorted keys of a locale, for tooling and tests.
func Keys(lang Lang) []string {
	catalog := catalogs[lang]
	keys := make([]string, 0, len(catalog))
	for key := range catalog {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// interpolate replaces {{name}} placeholders using alternating name/value args.
func interpolate(s string, args []string) string {
	if len(args) < 2 || !strings.Contains(s, "{{") {
		return s
	}
	pairs := make([]string, 0, len(args))
	for i := 0; i+1 < len(args); i += 2 {
		pairs = append(pairs, "{{"+args[i]+"}}", args[i+1])
	}
	return strings.NewReplacer(pairs...).Replace(s)
}
