// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/cleanup"
	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
	"github.com/Frimurare/WulfVault/internal/server"
)

const (
	Version = "3.6.0-beta2"
)

var (
	port       = flag.String("port", getEnv("PORT", "8080"), "Server port")
	dataDir    = flag.String("data", getEnv("DATA_DIR", "./data"), "Data directory")
	uploadsDir = flag.String("uploads", getEnv("UPLOADS_DIR", "./uploads"), "Uploads directory")
	serverURL  = flag.String("url", getEnv("SERVER_URL", "http://localhost:8080"), "Server URL")
	setup      = flag.Bool("setup", false, "Run initial setup")
)

func main() {
	flag.Parse()

	fmt.Printf("WulfVault File Sharing System v%s\n", Version)
	fmt.Println("Based on Gokapi - https://github.com/Forceu/Gokapi")
	fmt.Println("---")

	// Initialize database
	log.Printf("Initializing database in %s...", *dataDir)
	if err := database.Initialize(*dataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	// Ensure uploads directory exists
	if err := os.MkdirAll(*uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Run setup if requested or if no users exist
	if *setup || needsSetup() {
		if err := runSetup(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
	}

	// Cleanup expired sessions periodically
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := auth.CleanupExpiredSessions(); err != nil {
				log.Printf("Error cleaning up sessions: %v", err)
			}
		}
	}()

	// Load or create configuration
	cfg, err := config.LoadOrCreate(*dataDir)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set runtime version
	cfg.Version = Version

	// Load server URL from database first (highest priority)
	// This allows admin panel settings to override environment variables
	if dbServerURL, err := database.DB.GetConfigValue("server_url"); err == nil && dbServerURL != "" {
		// Add port if it's stored separately
		if dbPort, portErr := database.DB.GetConfigValue("port"); portErr == nil && dbPort != "" {
			cfg.ServerURL = dbServerURL + ":" + dbPort
		} else {
			cfg.ServerURL = dbServerURL + ":" + cfg.Port
		}
	} else {
		// If not in database, check environment variable or command-line flag
		serverURLFromEnv := getEnv("SERVER_URL", "")
		if serverURLFromEnv != "" || isFlagPassed("url") {
			cfg.ServerURL = *serverURL
		}
	}

	// Override config with command-line flags ONLY if they were explicitly set
	// Check if port flag was explicitly provided (not just default value)
	portFromEnv := getEnv("PORT", "")
	if portFromEnv != "" || isFlagPassed("port") {
		cfg.Port = *port
	}

	// Always override uploads dir if provided
	cfg.UploadsDir = *uploadsDir

	// Load trash retention setting from database if available
	if trashRetentionStr, err := database.DB.GetConfigValue("trash_retention_days"); err == nil && trashRetentionStr != "" {
		if days, parseErr := strconv.Atoi(trashRetentionStr); parseErr == nil && days > 0 {
			cfg.TrashRetentionDays = days
		}
	}
	if cfg.TrashRetentionDays <= 0 {
		cfg.TrashRetentionDays = 5 // default fallback
	}

	// Start file expiration cleanup scheduler (runs every 6 hours)
	cleanup.StartCleanupScheduler(*uploadsDir, 6*time.Hour, cfg.TrashRetentionDays)

	// Cleanup expired file requests periodically (runs every 24 hours)
	// File requests expire after 24 hours, then show "expired" message for 10 days, then are deleted
	go func() {
		// Run immediately on startup
		if err := database.DB.CleanupExpiredFileRequests(); err != nil {
			log.Printf("Error cleaning up expired file requests: %v", err)
		}

		// Then run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := database.DB.CleanupExpiredFileRequests(); err != nil {
				log.Printf("Error cleaning up expired file requests: %v", err)
			}
		}
	}()

	// Cleanup old soft-deleted accounts (runs daily, deletes accounts soft-deleted for 90+ days)
	go func() {
		// Run immediately on startup
		log.Println("Starting 90-day soft delete cleanup...")
		userCount, err := database.DB.PermanentlyDeleteOldUsers(90)
		if err != nil {
			log.Printf("Error permanently deleting old users: %v", err)
		} else if userCount > 0 {
			log.Printf("Permanently deleted %d users that were soft-deleted 90+ days ago", userCount)
		}

		downloadAccountCount, err := database.DB.PermanentlyDeleteOldDownloadAccounts(90)
		if err != nil {
			log.Printf("Error permanently deleting old download accounts: %v", err)
		} else if downloadAccountCount > 0 {
			log.Printf("Permanently deleted %d download accounts that were soft-deleted 90+ days ago", downloadAccountCount)
		}

		// Then run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("Running daily 90-day soft delete cleanup...")
			userCount, err := database.DB.PermanentlyDeleteOldUsers(90)
			if err != nil {
				log.Printf("Error permanently deleting old users: %v", err)
			} else if userCount > 0 {
				log.Printf("Permanently deleted %d users that were soft-deleted 90+ days ago", userCount)
			}

			downloadAccountCount, err := database.DB.PermanentlyDeleteOldDownloadAccounts(90)
			if err != nil {
				log.Printf("Error permanently deleting old download accounts: %v", err)
			} else if downloadAccountCount > 0 {
				log.Printf("Permanently deleted %d download accounts that were soft-deleted 90+ days ago", downloadAccountCount)
			}
		}
	}()

	log.Printf("Server configuration:")
	log.Printf("  - URL: %s", cfg.ServerURL)
	log.Printf("  - Port: %s", cfg.Port)
	log.Printf("  - Data: %s", *dataDir)
	log.Printf("  - Uploads: %s", cfg.UploadsDir)
	log.Printf("  - Company: %s", cfg.CompanyName)

	// Create static directory
	os.MkdirAll("web/static", 0755)

	// Start web server
	srv := server.New(cfg)
	log.Fatal(srv.Start())
}

func needsSetup() bool {
	count, err := database.DB.GetTotalUsers()
	if err != nil {
		return true
	}
	return count == 0
}

func runSetup() error {
	log.Println("Running initial setup...")

	// Check if admin already exists
	existing, _ := database.DB.GetTotalUsers()
	if existing > 0 {
		log.Println("Users already exist, skipping setup")
		return nil
	}

	// Get admin credentials
	adminEmail := getEnv("ADMIN_EMAIL", "admin@localhost")
	adminPassword := getEnv("ADMIN_PASSWORD", generateRandomPassword())

	// Hash password
	hashedPassword, err := auth.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create super admin user
	admin := &models.User{
		Name:           "admin",
		Email:          adminEmail,
		Password:       hashedPassword,
		UserLevel:      models.UserLevelSuperAdmin,
		Permissions:    models.UserPermissionAll,
		StorageQuotaMB: 100000, // 100 GB for admin
		StorageUsedMB:  0,
		IsActive:       true,
	}

	if err := database.DB.CreateUser(admin); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Set default branding
	// TODO: Save branding to configuration
	// _ = models.DefaultBranding()

	log.Println("✅ Setup complete!")
	log.Printf("   Admin Email: %s", adminEmail)
	if os.Getenv("ADMIN_PASSWORD") == "" {
		log.Printf("   Admin Password: %s", adminPassword)
		log.Printf("   ⚠️  SAVE THIS PASSWORD - it won't be shown again!")
	}

	return nil
}

func generateRandomPassword() string {
	// Simple random password for demo
	return fmt.Sprintf("admin-%d", time.Now().Unix())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isFlagPassed checks if a command-line flag was explicitly set
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
