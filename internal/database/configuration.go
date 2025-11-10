package database

import (
	"database/sql"
	"errors"
)

// GetConfigValue gets a configuration value
func (d *Database) GetConfigValue(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT Value FROM Configuration WHERE Key = ?", key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// SetConfigValue sets a configuration value
func (d *Database) SetConfigValue(key, value string) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO Configuration (Key, Value)
		VALUES (?, ?)`, key, value)
	return err
}

// GetBrandingConfig gets all branding configuration
func (d *Database) GetBrandingConfig() (map[string]string, error) {
	rows, err := d.db.Query("SELECT Key, Value FROM Configuration WHERE Key LIKE 'branding_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		config[key] = value
	}

	// Set defaults if not exists
	if _, ok := config["branding_logo"]; !ok {
		config["branding_logo"] = ""
	}
	if _, ok := config["branding_company_name"]; !ok {
		config["branding_company_name"] = "Manvarg Sharecare"
	}
	if _, ok := config["branding_primary_color"]; !ok {
		config["branding_primary_color"] = "#2563eb"
	}
	if _, ok := config["branding_secondary_color"]; !ok {
		config["branding_secondary_color"] = "#1e40af"
	}

	return config, nil
}
