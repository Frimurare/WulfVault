// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"bytes"
	"html/template"
	"log"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

// appShellData is shared by the overview and every authenticated subpage.
type appShellData struct {
	Lang          string
	Company       string
	Logo          template.URL
	Version       string
	Name          string
	PageTitle     string
	CurrentPath   string
	AccountPath   string
	LogoutPath    string
	LanguageLinks []dashboardLink
	Navigation    []dashboardLink
	FileLinks     []dashboardLink
	ServerLinks   []dashboardLink
}

func (s *Server) newAppShellData(tr *i18n.Translator, user *models.User, role string) appShellData {
	branding, _ := database.DB.GetBrandingConfig()
	data := appShellData{
		Lang:        string(tr.Lang()),
		Company:     s.config.CompanyName,
		Logo:        dashboardLogoURL(branding["branding_logo"]),
		Version:     s.config.Version,
		PageTitle:   tr.T("nav.dashboard"),
		AccountPath: "/settings",
		LogoutPath:  "/logout",
	}
	if user != nil {
		data.Name = user.Name
	}
	link := func(key, href, icon string) dashboardLink {
		return dashboardLink{Label: tr.T(key), Href: href, Icon: icon}
	}
	for _, lang := range i18n.Supported() {
		data.LanguageLinks = append(data.LanguageLinks, dashboardLink{
			Label: tr.T("lang." + string(lang)),
			Href:  "/lang?lang=" + string(lang),
			Icon:  string(lang),
		})
	}
	switch role {
	case "admin":
		data.PageTitle = tr.T("admin.overview.title")
		data.Navigation = []dashboardLink{
			link("admin.overview.title", "/admin", "home"),
			link("nav.my_files", "/dashboard", "folder"),
			link("nav.users", "/admin/users", "users"),
			link("nav.teams", "/admin/teams", "team"),
		}
		data.FileLinks = []dashboardLink{
			link("nav.all_files", "/admin/files", "folder"),
			link("nav.duplicate_files", "/admin/duplicates", "folder"),
			link("nav.trash", "/admin/trash", "folder"),
		}
		data.ServerLinks = []dashboardLink{
			link("nav.server_settings", "/admin/settings", "server"),
			link("nav.branding", "/admin/branding", "server"),
			link("nav.email", "/admin/email-settings", "server"),
			link("nav.identity_providers", "/admin/identity-providers", "shield"),
			link("nav.audit_logs", "/admin/audit-logs", "shield"),
			link("nav.server_logs", "/admin/server-logs", "server"),
			link("nav.sysmonitor_logs", "/admin/sysmonitor-logs", "server"),
		}
	case "download":
		data.AccountPath = "/download/account-settings"
		data.LogoutPath = "/download/logout"
		data.Navigation = []dashboardLink{
			link("nav.dashboard", "/download/dashboard", "home"),
			link("dlaccount.change_password_title", "/download/change-password", "shield"),
			link("nav.account_settings", "/download/account-settings", "account"),
		}
	default:
		data.Navigation = []dashboardLink{
			link("nav.dashboard", "/dashboard", "home"),
			link("nav.teams", "/teams", "team"),
		}
	}
	return data
}

func (s *Server) renderAppShell(tr *i18n.Translator, user *models.User, role string) string {
	data := s.newAppShellData(tr, user, role)
	page, err := template.New("app-shell").Funcs(template.FuncMap{"t": tr.T, "icon": dashboardIcon}).Parse(appShellTemplate)
	if err != nil {
		log.Printf("Parse application shell: %v", err)
		return ""
	}
	var out bytes.Buffer
	if err := page.ExecuteTemplate(&out, "app-shell", data); err != nil {
		log.Printf("Render application shell: %v", err)
		return ""
	}
	return out.String()
}

const appShellAssets = `<link rel="stylesheet" href="/static/css/admin-overview.css?v=wolf-overhaul-6">
<link rel="stylesheet" href="/static/css/app-shell.css?v=wolf-overhaul-6">
<script src="/static/js/admin-overview.js?v=wolf-overhaul-6" defer></script>`

const appShellTemplate = `{{define "app-shell"}}
<a class="wv-skip" href="#main">{{t "admin.overview.skip"}}</a>
<aside class="wv-sidebar" aria-label="{{t "admin.overview.navigation"}}">
  <a class="wv-brand" href="{{with index .Navigation 0}}{{.Href}}{{end}}">
    {{if .Logo}}<span class="wv-logo"><img src="{{.Logo}}" alt="{{.Company}}"></span>{{else}}<span class="wv-brand-icon">{{icon "shield"}}</span><span class="wv-company">{{.Company}}</span>{{end}}
  </a>
  <details class="wv-navigation" open>
    <summary>{{icon "folder"}}<span>{{t "admin.overview.navigation"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary>
    <nav aria-label="{{t "admin.overview.navigation"}}">
      {{range .Navigation}}<a class="wv-nav-link" href="{{.Href}}"{{if eq .Href $.CurrentPath}} aria-current="page"{{end}}>{{icon .Icon}}<span>{{.Label}}</span></a>{{end}}
      {{if .FileLinks}}<details class="wv-nav-group"><summary>{{icon "folder"}}<span>{{t "nav.files"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary><div>{{range .FileLinks}}<a class="wv-nav-link" href="{{.Href}}">{{.Label}}</a>{{end}}</div></details>{{end}}
      {{if .ServerLinks}}<details class="wv-nav-group"><summary>{{icon "server"}}<span>{{t "nav.server"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary><div>{{range .ServerLinks}}<a class="wv-nav-link" href="{{.Href}}">{{.Label}}</a>{{end}}</div></details>{{end}}
      <div class="wv-nav-account">
        <a class="wv-nav-link" href="{{.AccountPath}}">{{icon "account"}}<span>{{t "nav.my_account"}}</span></a>
        <a class="wv-nav-link" href="{{.LogoutPath}}">{{icon "logout"}}<span>{{t "nav.logout"}}</span></a>
      </div>
    </nav>
  </details>
  <div class="wv-sidebar-version"><span>WulfVault © 2025 Ulf Holmström</span>{{if .Version}}<span>v{{.Version}}</span>{{end}}</div>
</aside>
  <header class="wv-topbar">
    <p class="wv-breadcrumb"><span class="wv-breadcrumb-context">{{.Company}}</span> <span aria-hidden="true">/</span> <strong class="wv-current-page">{{.PageTitle}}</strong></p>
    <div class="wv-topbar-tools">
      <nav class="wv-languages" aria-label="{{t "admin.overview.language"}}">
        {{range .LanguageLinks}}<a href="{{.Href}}" hreflang="{{.Icon}}" lang="{{.Icon}}"{{if eq .Icon $.Lang}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
      </nav>
      {{if .Name}}<span class="wv-user">{{icon "account"}}<span>{{.Name}}</span></span>{{end}}
    </div>
  </header>
{{end}}`
