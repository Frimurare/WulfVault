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
