package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/Frimurare/Sharecare/internal/auth"
	"github.com/Frimurare/Sharecare/internal/config"
	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

type Server struct {
	config    *config.Config
	templates *template.Template
}

// New creates a new web server instance
func New(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
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
	mux.HandleFunc("/s/", s.handleSplashPage)
	mux.HandleFunc("/d/", s.handleDownload)
	mux.HandleFunc("/health", s.handleHealth)

	// User routes (require authentication)
	mux.HandleFunc("/dashboard", s.requireAuth(s.handleUserDashboard))
	mux.HandleFunc("/upload", s.requireAuth(s.handleUpload))
	mux.HandleFunc("/files", s.requireAuth(s.handleUserFiles))
	mux.HandleFunc("/file/delete", s.requireAuth(s.handleFileDelete))
	mux.HandleFunc("/file/edit", s.requireAuth(s.handleFileEdit))

	// Admin routes (require admin authentication)
	mux.HandleFunc("/admin", s.requireAdmin(s.handleAdminDashboard))
	mux.HandleFunc("/admin/users", s.requireAdmin(s.handleAdminUsers))
	mux.HandleFunc("/admin/users/create", s.requireAdmin(s.handleAdminUserCreate))
	mux.HandleFunc("/admin/users/edit", s.requireAdmin(s.handleAdminUserEdit))
	mux.HandleFunc("/admin/users/delete", s.requireAdmin(s.handleAdminUserDelete))
	mux.HandleFunc("/admin/files", s.requireAdmin(s.handleAdminFiles))
	mux.HandleFunc("/admin/trash", s.requireAdmin(s.handleAdminTrash))
	mux.HandleFunc("/admin/trash/restore", s.requireAdmin(s.handleAdminRestoreFile))
	mux.HandleFunc("/admin/trash/delete", s.requireAdmin(s.handleAdminPermanentDelete))
	mux.HandleFunc("/admin/branding", s.requireAdmin(s.handleAdminBranding))
	mux.HandleFunc("/admin/settings", s.requireAdmin(s.handleAdminSettings))

	// API routes
	mux.HandleFunc("/api/v1/upload", s.requireAuth(s.handleAPIUpload))
	mux.HandleFunc("/api/v1/files", s.requireAuth(s.handleAPIFiles))
	mux.HandleFunc("/api/v1/download/", s.handleAPIDownload)

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Server configuration
	addr := ":" + s.config.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
		user, err := s.getUserFromSession(r)
		if err != nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Store user in context (simple approach: we'll pass it via request context)
		r = r.WithContext(contextWithUser(r.Context(), user))
		next(w, r)
	}
}

// Middleware: Require admin authentication
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromSession(r)
		if err != nil || !user.IsAdmin() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
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

// handleHealth is a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "1.31",
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
