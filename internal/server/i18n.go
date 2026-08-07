// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/i18n"
)

// ConfigKeyDefaultLanguage is the Configuration row holding the server-wide
// default language. A missing row reads as "" and falls back to English, so
// existing installations need no migration.
const ConfigKeyDefaultLanguage = "default_language"

// tr returns the translator for the current request. Handlers use it as
//
//	tr := s.tr(r)
//	... + tr.T("login.submit") + ...
//
// It never returns nil, so it is safe to call before any middleware has run.
func (s *Server) tr(r *http.Request) *i18n.Translator {
	return i18n.FromRequest(r)
}

// loadLanguageConfig reads the server default language from the database into
// the i18n package at startup.
func (s *Server) loadLanguageConfig() {
	value, err := database.DB.GetConfigValue(ConfigKeyDefaultLanguage)
	if err != nil {
		log.Printf("Warning: Failed to load default language: %v", err)
		return
	}
	i18n.SetServerDefault(value)
	log.Printf("Default language: %s", i18n.ServerDefault())
}

// handleSetLanguage switches the interface language. It always stores the
// choice in a cookie, and additionally persists it on the profile when a user
// is logged in, so the choice follows the account to other devices.
func (s *Server) handleSetLanguage(w http.ResponseWriter, r *http.Request) {
	lang, ok := i18n.Normalize(r.URL.Query().Get("lang"))
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	i18n.SetCookie(w, lang)

	if user, err := s.getUserFromSession(r); err == nil && user != nil {
		if err := database.DB.UpdateUserLanguage(user.Id, string(lang)); err != nil {
			log.Printf("Warning: Failed to save language for user %d: %v", user.Id, err)
		}
	}

	http.Redirect(w, r, languageReturnPath(r), http.StatusSeeOther)
}

// languageReturnPath works out where to send the visitor after a language
// switch. The explicit "return" parameter wins; otherwise we reuse the page the
// switcher was clicked on. Only same-site paths are accepted, so the parameter
// cannot be used as an open redirect.
func languageReturnPath(r *http.Request) string {
	if path, ok := safeReturnPath(r.URL.Query().Get("return")); ok {
		return path
	}
	if referer, err := url.Parse(r.Referer()); err == nil {
		if path, ok := safeReturnPath(referer.RequestURI()); ok {
			return path
		}
	}
	return "/"
}

// safeReturnPath accepts only root-relative paths ("/files?x=1"). Anything
// starting with "//" or containing a scheme would point off-site.
func safeReturnPath(candidate string) (string, bool) {
	if !strings.HasPrefix(candidate, "/") || strings.HasPrefix(candidate, "//") {
		return "", false
	}
	if candidate == "/lang" || strings.HasPrefix(candidate, "/lang?") {
		return "", false // avoid bouncing back into the switcher
	}
	return candidate, true
}

// getLanguageSwitcherHTML renders the flag icons shown in the header, to the
// left of the version badge. The active language is fully opaque, the other one
// dimmed until hovered.
func (s *Server) getLanguageSwitcherHTML(tr *i18n.Translator) string {
	html := `<div class="lang-switch">`
	for _, lang := range i18n.Supported() {
		title := template.HTMLEscapeString(tr.T("lang.switch_to", "language", tr.T("lang."+string(lang))))
		active := ""
		if tr.Is(lang) {
			active = " active"
			title = template.HTMLEscapeString(tr.T("lang." + string(lang)))
		}
		html += `
                <a class="lang-flag` + active + `" href="/lang?lang=` + string(lang) + `" title="` + title + `" aria-label="` + title + `" hreflang="` + string(lang) + `">` +
			languageFlagSVG(lang) + `</a>`
	}
	return html + `
            </div>`
}

// languageFlagSVG returns a small inline flag. Inline SVG keeps the header
// self-contained: no extra request, no emoji font differences between systems.
func languageFlagSVG(lang i18n.Lang) string {
	switch lang {
	case i18n.LangSV:
		return `<svg viewBox="0 0 16 10" width="22" height="14" role="img" aria-hidden="true">` +
			`<rect width="16" height="10" fill="#006AA7"/>` +
			`<rect y="4" width="16" height="2" fill="#FECC00"/>` +
			`<rect x="5" width="2" height="10" fill="#FECC00"/>` +
			`</svg>`
	default:
		return `<svg viewBox="0 0 60 30" width="22" height="14" role="img" aria-hidden="true">` +
			`<rect width="60" height="30" fill="#012169"/>` +
			`<path d="M0,0 L60,30 M60,0 L0,30" stroke="#FFFFFF" stroke-width="6"/>` +
			`<path d="M0,0 L60,30 M60,0 L0,30" stroke="#C8102E" stroke-width="2"/>` +
			`<path d="M30,0 V30 M0,15 H60" stroke="#FFFFFF" stroke-width="10"/>` +
			`<path d="M30,0 V30 M0,15 H60" stroke="#C8102E" stroke-width="6"/>` +
			`</svg>`
	}
}

// languageSwitcherCSS styles the flag icons. It is appended to the header CSS
// that every page already inlines.
const languageSwitcherCSS = `
        .header nav .lang-switch {
            display: flex;
            align-items: center;
            gap: 6px;
            padding: 0 4px;
        }
        .header nav .lang-switch .lang-flag {
            display: flex;
            align-items: center;
            padding: 3px;
            background: transparent;
            border-radius: 3px;
            opacity: 0.55;
            line-height: 0;
            transition: opacity 0.2s;
        }
        .header nav .lang-switch .lang-flag:hover {
            background: transparent;
            opacity: 1;
        }
        .header nav .lang-switch .lang-flag.active {
            opacity: 1;
            box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.7);
        }
        .header nav .lang-switch .lang-flag svg {
            display: block;
            border-radius: 1px;
        }
        @media screen and (max-width: 768px) {
            .header nav .lang-switch {
                width: 100%;
                padding: 15px 20px !important;
                gap: 12px;
            }
            .header nav .lang-switch .lang-flag svg {
                width: 30px;
                height: 19px;
            }
        }`

// languageSelectOptionsHTML renders <option> elements for a language picker.
// selected is the currently stored value; an empty string selects the
// "follow the server default" option, which is offered when allowFollow is set.
func languageSelectOptionsHTML(tr *i18n.Translator, selected string, allowFollow bool) string {
	html := ""
	if allowFollow {
		serverLang := tr.T("lang." + string(i18n.ServerDefault()))
		label := template.HTMLEscapeString(tr.T("settings.language_follow_server", "language", serverLang))
		attr := ""
		if !i18n.IsSupported(selected) {
			attr = " selected"
		}
		html += `<option value=""` + attr + `>` + label + `</option>`
	}
	for _, lang := range i18n.Supported() {
		attr := ""
		if selected == string(lang) {
			attr = " selected"
		}
		html += `<option value="` + string(lang) + `"` + attr + `>` +
			template.HTMLEscapeString(tr.T("lang."+string(lang))) + `</option>`
	}
	return html
}
