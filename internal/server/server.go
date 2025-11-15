// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

type Server struct {
	config           *config.Config
	templates        *template.Template
	activeTransfers  map[string]bool // sessionId -> has active transfer
	transfersMutex   sync.RWMutex
}

// New creates a new web server instance
func New(cfg *config.Config) *Server {
	return &Server{
		config:          cfg,
		activeTransfers: make(map[string]bool),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Load templates
	if err := s.loadTemplates(); err != nil {
		return err
	}

	// Setup routes
	mux := http.NewServeMux()

	// Load branding configuration
	s.loadBrandingConfig()

	// Public routes
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/forgot-password", s.handleForgotPassword)
	mux.HandleFunc("/reset-password", s.handleResetPassword)
	mux.HandleFunc("/s/", s.handleSplashPage)
	mux.HandleFunc("/d/", s.handleDownload)
	mux.HandleFunc("/health", s.handleHealth)

	// 2FA routes
	mux.HandleFunc("/2fa/verify", s.handle2FAVerify)
	mux.HandleFunc("/2fa/setup", s.requireAuth(s.handle2FASetup))
	mux.HandleFunc("/2fa/enable", s.requireAuth(s.handle2FAEnable))
	mux.HandleFunc("/2fa/disable", s.requireAuth(s.handle2FADisable))
	mux.HandleFunc("/2fa/regenerate-backup-codes", s.requireAuth(s.handle2FARegenerateBackupCodes))

	// Public file request routes
	mux.HandleFunc("/upload-request/", s.handleUploadRequest)

	// Download account GDPR self-service (legacy route - kept for compatibility)
	mux.HandleFunc("/download-account/gdpr", s.handleDownloadAccountGDPR)
	mux.HandleFunc("/download-account/delete", s.handleDownloadAccountDelete)

	// Download user routes (require download account authentication)
	mux.HandleFunc("/download/dashboard", s.requireDownloadAuth(s.handleDownloadDashboard))
	mux.HandleFunc("/download/change-password", s.requireDownloadAuth(s.handleDownloadChangePassword))
	mux.HandleFunc("/download/account-settings", s.requireDownloadAuth(s.handleDownloadAccountSettings))
	mux.HandleFunc("/download/delete-account", s.requireDownloadAuth(s.handleDownloadAccountDeleteSelf))
	mux.HandleFunc("/download/logout", s.handleDownloadLogout)
	mux.HandleFunc("/download/deleted-success", s.handleDownloadDeletedSuccess)

	// User routes (require authentication)
	mux.HandleFunc("/dashboard", s.requireAuth(s.handleUserDashboard))
	mux.HandleFunc("/settings", s.requireAuth(s.handleUserSettings))
	mux.HandleFunc("/change-password", s.requireAuth(s.handleChangePassword))
	mux.HandleFunc("/upload", s.requireAuth(s.handleUpload))
	mux.HandleFunc("/files", s.requireAuth(s.handleUserFiles))
	mux.HandleFunc("/file/delete", s.requireAuth(s.handleFileDelete))
	mux.HandleFunc("/file/edit", s.requireAuth(s.handleFileEdit))
	mux.HandleFunc("/file/downloads", s.requireAuth(s.handleFileDownloadHistory))
	mux.HandleFunc("/file/email", s.requireAuth(s.handleFileEmail))
	mux.HandleFunc("/file-request/create", s.requireAuth(s.handleFileRequestCreate))
	mux.HandleFunc("/file-request/list", s.requireAuth(s.handleFileRequestList))
	mux.HandleFunc("/file-request/delete", s.requireAuth(s.handleFileRequestDelete))

	// Teams routes (require authentication)
	mux.HandleFunc("/teams", s.requireAuth(s.handleUserTeams))

	// Admin routes (require admin authentication)
	mux.HandleFunc("/admin", s.requireAdmin(s.handleAdminDashboard))
	mux.HandleFunc("/admin/users", s.requireAdmin(s.handleAdminUsers))
	mux.HandleFunc("/admin/users/create", s.requireAdmin(s.handleAdminUserCreate))
	mux.HandleFunc("/admin/users/edit", s.requireAdmin(s.handleAdminUserEdit))
	mux.HandleFunc("/admin/users/delete", s.requireAdmin(s.handleAdminUserDelete))
	mux.HandleFunc("/admin/download-accounts/toggle", s.requireAdmin(s.handleAdminToggleDownloadAccount))
	mux.HandleFunc("/admin/download-accounts/create", s.requireAdmin(s.handleAdminCreateDownloadAccount))
	mux.HandleFunc("/admin/download-accounts/edit", s.requireAdmin(s.handleAdminEditDownloadAccount))
	mux.HandleFunc("/admin/download-accounts/delete", s.requireAdmin(s.handleAdminDeleteDownloadAccount))
	mux.HandleFunc("/admin/files", s.requireAdmin(s.handleAdminFiles))
	mux.HandleFunc("/admin/trash", s.requireAdmin(s.handleAdminTrash))
	mux.HandleFunc("/admin/trash/restore", s.requireAdmin(s.handleAdminRestoreFile))
	mux.HandleFunc("/admin/trash/delete", s.requireAdmin(s.handleAdminPermanentDelete))
	mux.HandleFunc("/admin/branding", s.requireAdmin(s.handleAdminBranding))
	mux.HandleFunc("/admin/settings", s.requireAdmin(s.handleAdminSettings))
	mux.HandleFunc("/admin/email-settings", s.requireAdmin(s.handleEmailSettings))
	mux.HandleFunc("/admin/teams", s.requireAdmin(s.handleAdminTeams))
	mux.HandleFunc("/admin/reboot", s.requireAdmin(s.handleAdminReboot))

	// Teams API routes (require authentication)
	mux.HandleFunc("/api/teams/my", s.requireAuth(s.handleAPIMyTeams))
	mux.HandleFunc("/api/teams/members", s.requireAuth(s.handleAPITeamMembers))
	mux.HandleFunc("/api/teams/files", s.requireAuth(s.handleAPITeamFiles))
	mux.HandleFunc("/api/teams/add-member", s.requireAuth(s.handleAPITeamAddMember))
	mux.HandleFunc("/api/teams/remove-member", s.requireAuth(s.handleAPITeamRemoveMember))
	mux.HandleFunc("/api/teams/share-file", s.requireAuth(s.handleAPIShareFileToTeam))
	mux.HandleFunc("/api/teams/unshare-file", s.requireAuth(s.handleAPIUnshareFileFromTeam))

	// Teams Admin API routes (require admin)
	mux.HandleFunc("/api/admin/teams/create", s.requireAdmin(s.handleAPITeamCreate))
	mux.HandleFunc("/api/admin/teams/update", s.requireAdmin(s.handleAPITeamUpdate))
	mux.HandleFunc("/api/admin/teams/delete", s.requireAdmin(s.handleAPITeamDelete))
	mux.HandleFunc("/api/admin/users/list", s.requireAdmin(s.handleAPIUsersList))

	// Email API routes
	mux.HandleFunc("/api/email/configure", s.requireAuth(s.requireAdmin(s.handleEmailConfigure)))
	mux.HandleFunc("/api/email/test", s.requireAuth(s.requireAdmin(s.handleEmailTest)))
	mux.HandleFunc("/api/email/send-splash-link", s.requireAuth(s.handleSendSplashLink))

	// API routes (legacy)
	mux.HandleFunc("/api/v1/upload", s.requireAuth(s.handleAPIUpload))
	mux.HandleFunc("/api/v1/files", s.requireAuth(s.handleAPIFiles))
	mux.HandleFunc("/api/v1/download/", s.handleAPIDownload)

	// User Management REST API (Admin only)
	mux.HandleFunc("/api/v1/users", s.requireAdmin(s.handleAPIGetUsers))
	mux.HandleFunc("/api/v1/users/", s.requireAdmin(s.handleRESTUserRoutes))

	// File Management REST API
	mux.HandleFunc("/api/v1/files/", s.requireAuth(s.handleRESTFileRoutes))

	// Download Accounts REST API (Admin only)
	mux.HandleFunc("/api/v1/download-accounts", s.requireAdmin(s.handleAPIGetDownloadAccounts))
	mux.HandleFunc("/api/v1/download-accounts/", s.requireAdmin(s.handleRESTDownloadAccountRoutes))

	// File Requests REST API
	mux.HandleFunc("/api/v1/file-requests", s.requireAuth(s.handleAPIGetFileRequests))
	mux.HandleFunc("/api/v1/file-requests/", s.requireAuth(s.handleRESTFileRequestRoutes))

	// Trash Management REST API (Admin only)
	mux.HandleFunc("/api/v1/trash", s.requireAdmin(s.handleAPIGetTrash))
	mux.HandleFunc("/api/v1/trash/", s.requireAdmin(s.handleRESTTrashRoutes))

	// Admin/System REST API
	mux.HandleFunc("/api/v1/admin/stats", s.requireAdmin(s.handleAPIGetStats))
	mux.HandleFunc("/api/v1/admin/branding", s.requireAdmin(s.handleRESTBrandingRoutes))
	mux.HandleFunc("/api/v1/admin/settings", s.requireAdmin(s.handleRESTSettingsRoutes))

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Server configuration
	addr := ":" + s.config.Port
	server := &http.Server{
		Addr:              addr,
		Handler:           s.loggingMiddleware(mux),
		ReadHeaderTimeout: 60 * time.Second,  // Time to read request headers only (not body)
		WriteTimeout:      8 * time.Hour,     // Extended for very large file uploads on slow connections (up to 8 hours)
		IdleTimeout:       120 * time.Second, // Keep-alive timeout
	}

	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📍 Server URL: %s", s.config.ServerURL)
	return server.ListenAndServe()
}

// loadTemplates loads all HTML templates
func (s *Server) loadTemplates() error {
	// For now, we'll use embedded templates
	// Later we can load from web/templates/
	s.templates = template.New("")
	return nil
}

// Middleware: Logging
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// Middleware: Require authentication
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		user, err := s.getUserFromSession(r)
		if err != nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Check for inactivity timeout (10 minutes), but only if no active transfer
		if !s.hasActiveTransfer(cookie.Value) {
			timeSinceLastActivity := time.Since(time.Unix(user.LastOnline, 0))
			if timeSinceLastActivity > auth.InactivityTimeout {
				// Force logout due to inactivity
				auth.DeleteSession(cookie.Value)
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Redirect(w, r, "/login?timeout=1", http.StatusSeeOther)
				return
			}
		}

		// Store user in context (simple approach: we'll pass it via request context)
		r = r.WithContext(contextWithUser(r.Context(), user))
		next(w, r)
	}
}

// Middleware: Require admin authentication
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		user, err := s.getUserFromSession(r)
		if err != nil || !user.IsAdmin() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check for inactivity timeout (10 minutes), but only if no active transfer
		if !s.hasActiveTransfer(cookie.Value) {
			timeSinceLastActivity := time.Since(time.Unix(user.LastOnline, 0))
			if timeSinceLastActivity > auth.InactivityTimeout {
				// Force logout due to inactivity
				auth.DeleteSession(cookie.Value)
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Redirect(w, r, "/login?timeout=1", http.StatusSeeOther)
				return
			}
		}

		r = r.WithContext(contextWithUser(r.Context(), user))
		next(w, r)
	}
}

// getUserFromSession retrieves user from session cookie
func (s *Server) getUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	user, err := auth.GetUserBySession(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// loadBrandingConfig loads branding configuration from database
func (s *Server) loadBrandingConfig() {
	brandingConfig, err := database.DB.GetBrandingConfig()
	if err != nil {
		log.Printf("Warning: Failed to load branding config: %v", err)
		return
	}

	// Update server config with branding
	if companyName, ok := brandingConfig["branding_company_name"]; ok && companyName != "" {
		s.config.CompanyName = companyName
	}
	if primaryColor, ok := brandingConfig["branding_primary_color"]; ok && primaryColor != "" {
		s.config.PrimaryColor = primaryColor
	}
	if secondaryColor, ok := brandingConfig["branding_secondary_color"]; ok && secondaryColor != "" {
		s.config.SecondaryColor = secondaryColor
	}

	log.Printf("Branding config loaded: %s", s.config.CompanyName)
}

// Helper functions for color fallbacks
func (s *Server) getPrimaryColor() string {
	color := s.config.PrimaryColor
	// Reject empty, white, or near-white colors
	if color == "" || color == "#ffffff" || color == "#fff" ||
		color == "#FFFFFF" || color == "#FFF" ||
		color == "white" || color == "White" || color == "WHITE" ||
		color == "#fefefe" || color == "#FEFEFE" {
		return "#2563eb" // Default blue
	}
	return color
}

func (s *Server) getSecondaryColor() string {
	color := s.config.SecondaryColor
	// Reject empty, white, or near-white colors
	if color == "" || color == "#ffffff" || color == "#fff" ||
		color == "#FFFFFF" || color == "#FFF" ||
		color == "white" || color == "White" || color == "WHITE" ||
		color == "#fefefe" || color == "#FEFEFE" {
		return "#1e40af" // Default darker blue
	}
	return color
}

// getPublicURL returns the full server URL including port
func (s *Server) getPublicURL() string {
	serverURL := s.config.ServerURL
	port := s.config.Port

	// If port is standard (80 for http, 443 for https), don't add it
	if port == "80" || port == "443" {
		return serverURL
	}

	// Check if URL already has a port
	if len(serverURL) > 0 && serverURL[len(serverURL)-1:] != "/" {
		// Check if already has ":port" suffix
		for i := len(serverURL) - 1; i >= 0; i-- {
			if serverURL[i] == ':' {
				// Already has port, return as is
				return serverURL
			}
			if serverURL[i] == '/' {
				// Found / before :, no port in URL
				break
			}
		}
	}

	// Add port to URL
	return serverURL + ":" + port
}

// handleHealth is a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": s.config.Version,
	})
}

// handleHome serves the homepage
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user, err := s.getUserFromSession(r)
	if err != nil {
		// Not logged in, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Logged in, redirect to dashboard
	if user.IsAdmin() {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// renderTemplate is a helper to render templates (we'll implement simple HTML for now)
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// For MVP, we'll send simple HTML
	// Later we can use proper templates
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response
func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, map[string]string{"error": message})
}

// hasActiveTransfer checks if a session has an active file transfer
func (s *Server) hasActiveTransfer(sessionId string) bool {
	s.transfersMutex.RLock()
	defer s.transfersMutex.RUnlock()
	return s.activeTransfers[sessionId]
}

// markTransferActive marks a session as having an active transfer
func (s *Server) markTransferActive(sessionId string) {
	s.transfersMutex.Lock()
	defer s.transfersMutex.Unlock()
	s.activeTransfers[sessionId] = true
}

// markTransferInactive removes the active transfer flag from a session
func (s *Server) markTransferInactive(sessionId string) {
	s.transfersMutex.Lock()
	defer s.transfersMutex.Unlock()
	delete(s.activeTransfers, sessionId)
}
