package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Frimurare/Sharecare/internal/auth"
	"github.com/Frimurare/Sharecare/internal/cleanup"
	"github.com/Frimurare/Sharecare/internal/config"
	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
	"github.com/Frimurare/Sharecare/internal/server"
)

const (
	Version = "1.31.0"
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

	fmt.Printf("Sharecare File Sharing System v%s\n", Version)
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

	// Start file expiration cleanup scheduler (runs every 6 hours)
	cleanup.StartCleanupScheduler(*uploadsDir, 6*time.Hour)

	// Load or create configuration
	cfg, err := config.LoadOrCreate(*dataDir)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command-line flags
	cfg.Port = *port
	cfg.ServerURL = *serverURL
	cfg.UploadsDir = *uploadsDir

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
