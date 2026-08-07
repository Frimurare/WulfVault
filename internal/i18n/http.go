// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package i18n

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// contextKey is unexported so no other package can collide with it.
type contextKey struct{}

var translatorContextKey contextKey

// cookieMaxAge keeps an explicit language choice for a year.
const cookieMaxAge = 365 * 24 * 60 * 60

// Resolve picks the language for a request following WulfVault's documented
// priority order:
//
//  1. the logged-in user's profile setting
//  2. the wv_lang cookie (an explicit click on the flag switcher)
//  3. the browser's Accept-Language header
//  4. the server default configured in the admin panel
//  5. DefaultLang (English)
//
// Each argument may be empty or hold an unsupported tag; the first one that
// resolves to a supported language wins.
func Resolve(userPref, cookieValue, acceptLanguage, serverDefaultCode string) Lang {
	for _, code := range []string{userPref, cookieValue} {
		if lang, ok := Normalize(code); ok {
			return lang
		}
	}
	if lang, ok := FromAcceptLanguage(acceptLanguage); ok {
		return lang
	}
	if lang, ok := Normalize(serverDefaultCode); ok {
		return lang
	}
	return DefaultLang
}

// FromAcceptLanguage returns the best supported language named in an
// Accept-Language header, honouring the q weights. It reports false when the
// header is empty or names no supported language.
func FromAcceptLanguage(header string) (Lang, bool) {
	if strings.TrimSpace(header) == "" {
		return "", false
	}

	type candidate struct {
		lang    Lang
		quality float64
		order   int
	}
	var candidates []candidate

	for i, part := range strings.Split(header, ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		lang, ok := Normalize(fields[0])
		if !ok {
			continue
		}
		quality := 1.0
		for _, field := range fields[1:] {
			field = strings.TrimSpace(field)
			if !strings.HasPrefix(field, "q=") {
				continue
			}
			if parsed, err := strconv.ParseFloat(field[2:], 64); err == nil {
				quality = parsed
			}
		}
		if quality <= 0 {
			continue // q=0 means "explicitly not acceptable"
		}
		candidates = append(candidates, candidate{lang: lang, quality: quality, order: i})
	}

	if len(candidates) == 0 {
		return "", false
	}

	// Highest q wins; ties keep the order the client listed them in.
	sort.SliceStable(candidates, func(a, b int) bool {
		if candidates[a].quality != candidates[b].quality {
			return candidates[a].quality > candidates[b].quality
		}
		return candidates[a].order < candidates[b].order
	})
	return candidates[0].lang, true
}

// NewContext returns ctx carrying tr.
func NewContext(ctx context.Context, tr *Translator) context.Context {
	return context.WithValue(ctx, translatorContextKey, tr)
}

// FromContext returns the translator stored in ctx. It never returns nil: if no
// translator was stored, the server default is used, so a handler that is
// reached outside the middleware still renders readable text.
func FromContext(ctx context.Context) *Translator {
	if tr, ok := ctx.Value(translatorContextKey).(*Translator); ok && tr != nil {
		return tr
	}
	return For(ServerDefault())
}

// FromRequest is FromContext for the request's context.
func FromRequest(r *http.Request) *Translator {
	if r == nil {
		return For(ServerDefault())
	}
	return FromContext(r.Context())
}

// Middleware resolves the language for every request from the cookie, the
// Accept-Language header and the server default, and stores the resulting
// translator in the request context. Authentication middleware runs later and
// refines the choice with the user's profile setting via WithUserLanguage.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieValue := ""
		if cookie, err := r.Cookie(CookieName); err == nil {
			cookieValue = cookie.Value
		}
		lang := Resolve("", cookieValue, r.Header.Get("Accept-Language"), string(ServerDefault()))
		next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), For(lang))))
	})
}

// WithUserLanguage returns r with the user's profile language applied. An empty
// or unsupported value leaves the request untouched, so users who never picked
// a language keep following the cookie, the browser and the server default.
func WithUserLanguage(r *http.Request, userLang string) *http.Request {
	lang, ok := Normalize(userLang)
	if !ok {
		return r
	}
	return r.WithContext(NewContext(r.Context(), For(lang)))
}

// ClearCookie removes an explicit language choice, so the visitor falls back to
// the browser's Accept-Language header and the server default again.
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetCookie stores an explicit language choice in the browser.
func SetCookie(w http.ResponseWriter, lang Lang) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    string(lang),
		Path:     "/",
		Expires:  time.Now().Add(cookieMaxAge * time.Second),
		MaxAge:   cookieMaxAge,
		HttpOnly: false, // read by nothing today, but harmless and debuggable
		SameSite: http.SameSiteLaxMode,
	})
}
