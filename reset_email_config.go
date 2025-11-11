package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./data/sharecare.db?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Delete all email provider configs
	result, err := db.Exec("DELETE FROM EmailProviderConfig")
	if err != nil {
		log.Fatal("Failed to delete email configs:", err)
	}

	rows, _ := result.RowsAffected()
	fmt.Printf("✅ Deleted %d email provider configuration(s)\n", rows)
	fmt.Println("You can now configure Brevo from scratch through the UI")
}
