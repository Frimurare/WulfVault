// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

// dashboardMetric is presentation data; values are escaped by html/template.
type dashboardMetric struct {
	Label string
	Value string
	Note  string
	Icon  string
}

type dashboardLink struct {
	Label string
	Href  string
	Icon  string
	Note  string
}

type adminOverviewData struct {
	Lang           string
	Company        string
	Logo           string
	Version        string
	Name           string
	Plain          bool
	LanguageLinks  []dashboardLink
	Navigation     []dashboardLink
	FileLinks      []dashboardLink
	ServerLinks    []dashboardLink
	Metrics        []dashboardMetric
	Transfers      []dashboardMetric
	StorageUsed    string
	StorageFree    string
	StoragePercent string
	StorageKnown   bool
	UserGrowth     []dashboardMetric
	Usage          []dashboardMetric
	Security       []dashboardMetric
	Files          []dashboardMetric
	Trends         []dashboardMetric
	TopUsers       []dashboardMetric
	Duplicates     []DuplicateFile
}

func dashboardFormatBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}
	size := float64(value)
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	unit := 0
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}
	return fmt.Sprintf("%.1f %s", size, units[unit])
}

// Storage combines bytes in the upload directory with available space, matching
// the existing dashboard counters. It is not a claim about total disk capacity:
// other applications, filesystem reserves and quotas may also use this disk.
func dashboardStoragePercent(used, available int64) (string, bool) {
	if used < 0 || available < 0 || (used == 0 && available == 0) {
		return "0", false
	}
	percentage := float64(used) / (float64(used) + float64(available)) * 100
	return fmt.Sprintf("%.1f", percentage), true
}

func dashboardIcon(name string) template.HTML {
	paths := map[string]string{
		"home": `<path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1z"/>`,
		"folder": `<path d="M3 7V5a2 2 0 0 1 2-2h5l2 3h7a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>`,
		"users": `<circle cx="9" cy="7" r="3"/><path d="M3 21v-3a6 6 0 0 1 12 0v3m1-17a3 3 0 0 1 0 6m2 5a5 5 0 0 1 3 5"/>`,
		"team": `<circle cx="12" cy="7" r="3"/><path d="M6 21v-2a6 6 0 0 1 12 0v2M3 9v6m-3-3h6m15-3v6m-3-3h6"/>`,
		"server": `<rect x="3" y="3" width="18" height="7" rx="2"/><rect x="3" y="14" width="18" height="7" rx="2"/><path d="M7 6h1m-1 11h1"/>`,
		"download": `<path d="M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5"/>`,
		"upload": `<path d="M12 16V4m-5 5 5-5 5 5M4 16v5h16v-5"/>`,
		"calendar": `<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M7 3v4m10-4v4M3 11h18"/>`,
		"account": `<circle cx="12" cy="12" r="10"/><circle cx="12" cy="9" r="3"/><path d="M5 19a7 7 0 0 1 14 0"/>`,
		"logout": `<path d="M9 4H4v16h5m0-8h12m-4-4 4 4-4 4"/>`,
		"arrow": `<path d="M4 12h16m-6-6 6 6-6 6"/>`,
		"shield": `<path d="m12 3 9 4v5c0 5-9 10-9 10S3 17 3 12V7z"/><path d="m8 12 3 3 5-6"/>`,
	}
	path, ok := paths[name]
	if !ok {
		path = paths["folder"]
	}
	// Only fixed, internal SVG paths are allowed here.
	return template.HTML(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">` + path + `</svg>`)
}

func (s *Server) renderAdminDashboard(w http.ResponseWriter, tr *i18n.Translator, user *models.User, totalUsers, activeUsers, totalDownloads, downloadsToday int,
	bytesDownloadedToday, bytesDownloadedWeek, bytesDownloadedMonth, bytesDownloadedYear int64,
	bytesUploadedToday, bytesUploadedWeek, bytesUploadedMonth, bytesUploadedYear int64,
	usersAdded, usersRemoved int, userGrowth float64,
	activeFiles7Days, activeFiles30Days int, avgFileSize int64, avgDownloadsPerFile float64,
	twoFAAdoption, avgBackupCodes float64,
	largestFileName string, largestFileSize int64, top5ActiveUsers []string, top5FileCounts []int,
	topFileTypes []string, fileTypeCounts []int, topWeekday string, weekdayCount int, storagePast, storageNow int64,
	mostDownloadedFile string, downloadCount int, uploadsUsed, diskAvailable int64, duplicateFiles []DuplicateFile) {
	t := tr.T
	metric := func(key, value string) dashboardMetric {
		return dashboardMetric{Label: t("admin.overview." + key), Value: value}
	}
	link := func(key, href, icon string) dashboardLink {
		return dashboardLink{Label: t(key), Href: href, Icon: icon}
	}
	branding, _ := database.DB.GetBrandingConfig()
	style, _ := database.DB.GetConfigValue("dashboard_style")
	percentage, storageKnown := dashboardStoragePercent(uploadsUsed, diskAvailable)
	data := adminOverviewData{
		Lang:           string(tr.Lang()),
		Company:        s.config.CompanyName,
		Logo:           branding["branding_logo"],
		Version:        s.config.Version,
		Plain:          style == "plain",
		StorageUsed:    dashboardFormatBytes(uploadsUsed),
		StorageFree:    dashboardFormatBytes(diskAvailable),
		StoragePercent: percentage,
		StorageKnown:   storageKnown,
		Duplicates:     duplicateFiles,
	}
	if diskAvailable < 0 {
		data.StorageFree = t("admin.overview.no_data")
	}
	if user != nil {
		data.Name = user.Name
	}
	for _, lang := range i18n.Supported() {
		data.LanguageLinks = append(data.LanguageLinks, dashboardLink{
			Label: tr.T("lang." + string(lang)),
			Href:  "/lang?lang=" + string(lang) + "&return=/admin",
			Icon:  string(lang),
		})
	}
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
	data.Metrics = []dashboardMetric{
		{Label: t("admin.overview.total_users"), Value: fmt.Sprint(totalUsers), Icon: "users"},
		{Label: t("admin.overview.active_users"), Value: fmt.Sprint(activeUsers), Icon: "account"},
		{Label: t("admin.overview.total_downloads"), Value: fmt.Sprint(totalDownloads), Icon: "download"},
		{Label: t("admin.overview.downloads_today"), Value: fmt.Sprint(downloadsToday), Icon: "calendar"},
	}
	periods := []string{"today", "week", "month", "year"}
	downloaded := []int64{bytesDownloadedToday, bytesDownloadedWeek, bytesDownloadedMonth, bytesDownloadedYear}
	uploaded := []int64{bytesUploadedToday, bytesUploadedWeek, bytesUploadedMonth, bytesUploadedYear}
	for i, period := range periods {
		data.Transfers = append(data.Transfers, dashboardMetric{
			Label: t("admin.overview." + period),
			Value: dashboardFormatBytes(downloaded[i]),
			Note:  dashboardFormatBytes(uploaded[i]),
		})
	}
	data.UserGrowth = []dashboardMetric{
		metric("users_added", fmt.Sprint(usersAdded)),
		metric("users_removed", fmt.Sprint(usersRemoved)),
		metric("growth", fmt.Sprintf("%+.1f%%", userGrowth)),
	}
	data.Usage = []dashboardMetric{
		metric("active_files_7", fmt.Sprint(activeFiles7Days)),
		metric("active_files_30", fmt.Sprint(activeFiles30Days)),
		metric("average_size", dashboardFormatBytes(avgFileSize)),
		metric("average_downloads", fmt.Sprintf("%.1f", avgDownloadsPerFile)),
	}
	data.Security = []dashboardMetric{
		{Label: t("admin.overview.two_fa"), Value: fmt.Sprintf("%.1f%%", twoFAAdoption), Note: t("admin.overview.two_fa_note")},
		{Label: t("admin.overview.backup_codes"), Value: fmt.Sprintf("%.1f", avgBackupCodes), Note: t("admin.overview.backup_codes_note")},
	}
	if largestFileName == "" {
		largestFileName = t("admin.overview.no_data")
	}
	if mostDownloadedFile == "" {
		mostDownloadedFile = t("admin.overview.no_data")
	}
	data.Files = []dashboardMetric{
		{Label: t("admin.overview.largest_file"), Value: largestFileName, Note: dashboardFormatBytes(largestFileSize)},
		{Label: t("admin.overview.most_downloaded"), Value: mostDownloadedFile, Note: t("admin.overview.download_count", "count", fmt.Sprint(downloadCount))},
	}
	for i, name := range top5ActiveUsers {
		if i >= 5 || i >= len(top5FileCounts) {
			break
		}
		data.TopUsers = append(data.TopUsers, dashboardMetric{Label: name, Value: fmt.Sprint(top5FileCounts[i])})
	}
	var fileTypes []string
	for i, ext := range topFileTypes {
		if i >= len(fileTypeCounts) {
			break
		}
		fileTypes = append(fileTypes, fmt.Sprintf(".%s (%d)", ext, fileTypeCounts[i]))
	}
	fileTypesText := strings.Join(fileTypes, ", ")
	if fileTypesText == "" {
		fileTypesText = t("admin.overview.no_data")
	}
	weekday := t("admin.overview.no_data")
	for _, day := range []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"} {
		if strings.EqualFold(topWeekday, day) {
			weekday = t("admin.overview.day_" + strings.ToLower(day))
			break
		}
	}
	storageGrowth := t("admin.overview.no_data")
	if storagePast > 0 {
		storageGrowth = fmt.Sprintf("%+.1f%%", float64(storageNow-storagePast)/float64(storagePast)*100)
	}
	data.Trends = []dashboardMetric{
		metric("file_types", fileTypesText),
		{Label: t("admin.overview.active_day"), Value: weekday, Note: t("admin.overview.download_count", "count", fmt.Sprint(weekdayCount))},
		{Label: t("admin.overview.storage_trend"), Value: storageGrowth, Note: dashboardFormatBytes(storagePast) + " → " + dashboardFormatBytes(storageNow)},
	}
	page, err := template.New("admin-overview").Funcs(template.FuncMap{"t": t, "icon": dashboardIcon}).Parse(adminOverviewTemplate)
	if err != nil {
		log.Printf("Parse admin overview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var output bytes.Buffer
	if err := page.Execute(&output, data); err != nil {
		log.Printf("Render admin overview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(output.Bytes())
}

const adminOverviewTemplate = `<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{t "admin.page_title"}} · {{.Company}}</title>
<link rel="stylesheet" href="/static/css/admin-overview.css">
<script src="/static/js/admin-overview.js" defer></script>
</head>
<body class="wv-overview{{if .Plain}} wv-plain{{end}}">
<a class="wv-skip" href="#main">{{t "admin.overview.skip"}}</a>
<aside class="wv-sidebar" aria-label="{{t "admin.overview.navigation"}}">
  <a class="wv-brand" href="/admin">
    {{if .Logo}}<span class="wv-logo"><img src="{{.Logo}}" alt="{{.Company}}"></span>{{else}}<span class="wv-brand-icon">{{icon "shield"}}</span><span class="wv-company">{{.Company}}</span>{{end}}
    <span class="wv-product">WulfVault</span>
  </a>
  <details class="wv-navigation" open>
    <summary>{{icon "folder"}}<span>{{t "admin.overview.navigation"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary>
    <nav aria-label="{{t "admin.overview.navigation"}}">
      {{range .Navigation}}<a class="wv-nav-link" href="{{.Href}}"{{if eq .Href "/admin"}} aria-current="page"{{end}}>{{icon .Icon}}<span>{{.Label}}</span></a>{{end}}
      <details class="wv-nav-group"><summary>{{icon "folder"}}<span>{{t "nav.files"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary><div>{{range .FileLinks}}<a class="wv-nav-link" href="{{.Href}}">{{.Label}}</a>{{end}}</div></details>
      <details class="wv-nav-group"><summary>{{icon "server"}}<span>{{t "nav.server"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary><div>{{range .ServerLinks}}<a class="wv-nav-link" href="{{.Href}}">{{.Label}}</a>{{end}}</div></details>
      <div class="wv-nav-account">
        <a class="wv-nav-link" href="/settings">{{icon "account"}}<span>{{t "nav.my_account"}}</span></a>
        <a class="wv-nav-link" href="/logout">{{icon "logout"}}<span>{{t "nav.logout"}}</span></a>
      </div>
    </nav>
  </details>
</aside>
<div class="wv-workspace">
  <header class="wv-topbar">
    <p class="wv-breadcrumb">{{t "admin.overview.administration"}} <span aria-hidden="true">/</span> <strong>{{t "admin.overview.title"}}</strong></p>
    <div class="wv-topbar-tools">
      <nav class="wv-languages" aria-label="{{t "admin.overview.language"}}">
        {{range .LanguageLinks}}<a href="{{.Href}}" hreflang="{{.Icon}}" lang="{{.Icon}}"{{if eq .Icon $.Lang}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
      </nav>
      {{if .Name}}<span class="wv-user">{{icon "account"}}<span>{{.Name}}</span></span>{{end}}
    </div>
  </header>
  <main id="main" class="wv-main" tabindex="-1">
    <div class="wv-page-heading">
      <div><p class="wv-eyebrow">{{t "admin.overview.administration"}}</p><h1>{{t "admin.overview.title"}}</h1><p class="wv-subtitle">{{t "admin.overview.subtitle"}}</p></div>
      <div class="wv-actions"><a class="wv-button" href="/admin/users">{{icon "users"}}{{t "admin.overview.manage_users"}}</a><a class="wv-button wv-primary" href="/dashboard#uploadForm">{{icon "upload"}}{{t "dashboard.upload_files"}}</a></div>
    </div>
    <section class="wv-metrics" aria-label="{{t "admin.overview.key_metrics"}}">
      {{range .Metrics}}<article class="wv-card wv-metric"><span class="wv-metric-icon">{{icon .Icon}}</span><div><h2>{{.Label}}</h2><p class="wv-number">{{.Value}}</p></div></article>{{end}}
    </section>
    <div class="wv-summary-grid">
      <section class="wv-card wv-panel">
        <div class="wv-panel-heading"><h2>{{t "admin.overview.transfer"}}</h2>{{icon "download"}}</div>
        <p class="wv-help">{{t "admin.overview.transfer_note"}}</p>
        <div class="wv-table-scroll">
          <table class="wv-transfer-table"><caption class="wv-sr-only">{{t "admin.overview.transfer"}}</caption>
            <thead><tr><th scope="col">{{t "admin.overview.period"}}</th><th scope="col">{{t "admin.overview.downloaded"}}</th><th scope="col">{{t "admin.overview.uploaded"}}</th></tr></thead>
            <tbody>{{range .Transfers}}<tr><th scope="row">{{.Label}}</th><td>{{.Value}}</td><td>{{.Note}}</td></tr>{{end}}</tbody>
          </table>
        </div>
      </section>
      <section class="wv-card wv-panel">
        <div class="wv-panel-heading"><h2>{{t "admin.overview.storage"}}</h2>{{icon "server"}}</div>
        <p class="wv-help">{{t "admin.overview.storage_note"}}</p>
        <div class="wv-storage-content">
          <div class="wv-storage-ring">
            <svg viewBox="0 0 120 120" aria-hidden="true"><circle class="wv-ring-track" cx="60" cy="60" r="50" pathLength="100"/>{{if .StorageKnown}}<circle class="wv-ring-used" cx="60" cy="60" r="50" pathLength="100" stroke-dasharray="{{.StoragePercent}} 100"/>{{end}}</svg>
            <div><strong>{{if .StorageKnown}}{{.StoragePercent}}%{{else}}—{{end}}</strong><span>{{t "admin.overview.used"}}</span></div>
          </div>
          <dl class="wv-storage-values"><div><dt><span class="wv-dot"></span>{{t "admin.overview.upload_storage"}}</dt><dd>{{.StorageUsed}}</dd></div><div><dt><span class="wv-dot wv-dot-muted"></span>{{t "admin.overview.available"}}</dt><dd>{{.StorageFree}}</dd></div></dl>
        </div>
        {{if not .StorageKnown}}<p class="wv-help">{{t "admin.overview.storage_unknown"}}</p>{{end}}
      </section>
    </div>
    <section class="wv-quick-section" aria-labelledby="quick-heading">
      <h2 id="quick-heading">{{t "admin.overview.quick_access"}}</h2>
      <div class="wv-quick-grid">
        <a class="wv-card wv-quick" href="/dashboard"><span class="wv-metric-icon">{{icon "folder"}}</span><span><strong>{{t "nav.my_files"}}</strong><span>{{t "admin.overview.files_note"}}</span></span>{{icon "arrow"}}</a>
        <a class="wv-card wv-quick" href="/admin/users"><span class="wv-metric-icon">{{icon "users"}}</span><span><strong>{{t "nav.users"}}</strong><span>{{t "admin.overview.users_note"}}</span></span>{{icon "arrow"}}</a>
        <a class="wv-card wv-quick" href="/admin/settings"><span class="wv-metric-icon">{{icon "server"}}</span><span><strong>{{t "nav.server"}}</strong><span>{{t "admin.overview.server_note"}}</span></span>{{icon "arrow"}}</a>
      </div>
    </section>
    <p class="wv-tip">{{icon "shield"}}<span>{{t "admin.overview.tip"}}</span></p>
    <details class="wv-more">
      <summary><span>{{t "admin.overview.more_statistics"}}</span><span class="wv-chevron" aria-hidden="true">⌄</span></summary>
      <div class="wv-detail-grid">
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.user_growth"}}</h2>{{template "metric-list" .UserGrowth}}</section>
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.usage"}}</h2>{{template "metric-list" .Usage}}</section>
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.security"}}</h2>{{template "metric-list" .Security}}</section>
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.file_statistics"}}</h2>{{template "metric-list" .Files}}</section>
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.trends"}}</h2>{{template "metric-list" .Trends}}</section>
        <section class="wv-card wv-panel"><h2>{{t "admin.overview.top_users"}}</h2>{{if .TopUsers}}<dl class="wv-stat-list">{{range .TopUsers}}<div><dt>{{.Label}}</dt><dd>{{.Value}} {{t "admin.overview.files_unit"}}</dd></div>{{end}}</dl>{{else}}<p class="wv-help">{{t "admin.overview.no_data"}}</p>{{end}}</section>
      </div>
    </details>
    <details class="wv-more">
      <summary><span>{{t "nav.duplicate_files"}} <span class="wv-count">{{len .Duplicates}}</span></span><span class="wv-chevron" aria-hidden="true">⌄</span></summary>
      <div class="wv-duplicates">{{range .Duplicates}}<article class="wv-card wv-panel"><h3>{{.Name}}</h3><p>{{.Size}} · {{t "admin.overview.copies" "count" (printf "%d" .Count)}}</p><p class="wv-help">{{t "admin.overview.file_ids"}}: {{range $i, $id := .FileIds}}{{if $i}}, {{end}}{{$id}}{{end}}</p></article>{{else}}<p class="wv-help">{{t "admin.overview.no_duplicates"}}</p>{{end}}<a class="wv-text-link" href="/admin/duplicates">{{t "admin.overview.review_duplicates"}} →</a></div>
    </details>
  </main>
  <footer class="wv-footer"><span>{{.Company}} · WulfVault</span><span>v{{.Version}}</span></footer>
</div>
</body>
</html>
{{define "metric-list"}}<dl class="wv-stat-list">{{range .}}<div><dt>{{.Label}}</dt><dd>{{.Value}}{{if .Note}}<small>{{.Note}}</small>{{end}}</dd></div>{{end}}</dl>{{end}}
`
