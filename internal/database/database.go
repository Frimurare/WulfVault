package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

var DB *Database

// Initialize creates and initializes the database connection
func Initialize(dataDir string) error {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(dataDir, "sharecare.db")

	// Open SQLite database
	sqliteDb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := sqliteDb.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set pragmas for better performance
	if _, err := sqliteDb.Exec("PRAGMA journal_mode = WAL"); err != nil {
		log.Printf("Warning: Could not set WAL mode: %v", err)
	}

	DB = &Database{db: sqliteDb}

	// Create tables
	if err := DB.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

// createTables executes the schema creation SQL
func (d *Database) createTables() error {
	_, err := d.db.Exec(CreateTablesSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Run migrations for existing databases
	if err := d.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations handles database schema migrations
func (d *Database) runMigrations() error {
	// Migration 1: Add DeletedAt and DeletedBy columns to Files table if they don't exist
	var count int
	row := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('Files') WHERE name='DeletedAt'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Adding DeletedAt and DeletedBy columns to Files table")

		// Add DeletedAt column
		if _, err := d.db.Exec("ALTER TABLE Files ADD COLUMN DeletedAt INTEGER DEFAULT 0"); err != nil {
			log.Printf("Migration warning for DeletedAt (may be safe to ignore): %v", err)
		}

		// Add DeletedBy column
		if _, err := d.db.Exec("ALTER TABLE Files ADD COLUMN DeletedBy INTEGER DEFAULT 0"); err != nil {
			log.Printf("Migration warning for DeletedBy (may be safe to ignore): %v", err)
		}

		log.Printf("Migration completed: DeletedAt and DeletedBy columns added")
	}

	// Migration 2: Add FilePasswordPlain to Files table
	row = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('Files') WHERE name='FilePasswordPlain'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Adding FilePasswordPlain column to Files table")
		if _, err := d.db.Exec("ALTER TABLE Files ADD COLUMN FilePasswordPlain TEXT"); err != nil {
			log.Printf("Migration warning for FilePasswordPlain: %v", err)
		}
		log.Printf("Migration completed: FilePasswordPlain column added")
	}

	// Migration 3: Add Name to DownloadAccounts table
	row = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('DownloadAccounts') WHERE name='Name'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Adding Name column to DownloadAccounts table")
		if _, err := d.db.Exec("ALTER TABLE DownloadAccounts ADD COLUMN Name TEXT NOT NULL DEFAULT ''"); err != nil {
			log.Printf("Migration warning for DownloadAccounts Name: %v", err)
		}
		log.Printf("Migration completed: Name column added to DownloadAccounts")
	}

	// Migration 4: Create FileRequests table
	row = d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='FileRequests'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Creating FileRequests table")
		if _, err := d.db.Exec(`
			CREATE TABLE IF NOT EXISTS FileRequests (
				Id INTEGER PRIMARY KEY AUTOINCREMENT,
				UserId INTEGER NOT NULL,
				RequestToken TEXT NOT NULL UNIQUE,
				Title TEXT NOT NULL,
				Message TEXT,
				CreatedAt INTEGER NOT NULL,
				ExpiresAt INTEGER,
				IsActive INTEGER DEFAULT 1,
				MaxFileSize INTEGER DEFAULT 0,
				AllowedFileTypes TEXT,
				FOREIGN KEY (UserId) REFERENCES Users(Id)
			)
		`); err != nil {
			log.Printf("Migration error for FileRequests table: %v", err)
		}
		d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_filerequests_userid ON FileRequests(UserId)`)
		d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_filerequests_token ON FileRequests(RequestToken)`)
		log.Printf("Migration completed: FileRequests table created")
	}

	// Migration 5: Add UsedByIP and UsedAt columns to FileRequests for single-use tracking
	row = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('FileRequests') WHERE name='UsedByIP'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Adding UsedByIP and UsedAt columns to FileRequests table")

		// Add UsedByIP column
		if _, err := d.db.Exec("ALTER TABLE FileRequests ADD COLUMN UsedByIP TEXT"); err != nil {
			log.Printf("Migration warning for UsedByIP: %v", err)
		}

		// Add UsedAt column
		if _, err := d.db.Exec("ALTER TABLE FileRequests ADD COLUMN UsedAt INTEGER DEFAULT 0"); err != nil {
			log.Printf("Migration warning for UsedAt: %v", err)
		}

		log.Printf("Migration completed: UsedByIP and UsedAt columns added to FileRequests")
	}

	// Migration 6: Create EmailProviderConfig table
	row = d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='EmailProviderConfig'")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Printf("Running migration: Creating EmailProviderConfig table")
		if _, err := d.db.Exec(`
			CREATE TABLE IF NOT EXISTS EmailProviderConfig (
				Id INTEGER PRIMARY KEY AUTOINCREMENT,
				Provider TEXT NOT NULL UNIQUE,
				IsActive INTEGER DEFAULT 0,
				ApiKeyEncrypted TEXT,
				SMTPHost TEXT,
				SMTPPort INTEGER,
				SMTPUsername TEXT,
				SMTPPasswordEncrypted TEXT,
				SMTPUseTLS INTEGER DEFAULT 1,
				FromEmail TEXT NOT NULL,
				FromName TEXT,
				CreatedAt INTEGER NOT NULL,
				UpdatedAt INTEGER NOT NULL
			)
		`); err != nil {
			log.Printf("Migration error for EmailProviderConfig table: %v", err)
		}
		log.Printf("Migration completed: EmailProviderConfig table created")
	}

	// Migration 7: Add soft delete columns to Users and DownloadAccounts (via migrations.go)
	if err := d.RunMigrations(); err != nil {
		log.Printf("Migration error for soft delete columns: %v", err)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB returns the underlying sql.DB for direct queries
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Exec executes a query without returning rows
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// UpdateConfiguration saves configuration (placeholder - config is saved to file)
func (d *Database) UpdateConfiguration(cfg interface{}) error {
	// Configuration is saved to config.json file, not database
	// This is a placeholder method for compatibility
	return nil
}
