// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Frimurare/WulfVault/internal/models"
)

// Config holds the application configuration
type Config struct {
	ServerURL           string `json:"serverUrl"`
	Port                string `json:"port"`
	DataDir             string `json:"dataDir"`
	UploadsDir          string `json:"uploadsDir"`
	MaxFileSizeMB           int    `json:"maxFileSizeMB"`
	MaxUploadSizeMB         int    `json:"maxUploadSizeMB"`
	DefaultQuotaMB          int64  `json:"defaultQuotaMB"`
	SessionTimeoutHours     int    `json:"sessionTimeoutHours"`
	TrashRetentionDays      int    `json:"trashRetentionDays"`
	AuditLogRetentionDays   int    `json:"auditLogRetentionDays"`   // Days to keep audit logs (default: 90)
	AuditLogMaxSizeMB       int    `json:"auditLogMaxSizeMB"`       // Auto-cleanup if log exceeds this size (default: 100MB)
	SaveIP                  bool   `json:"saveIp"`
	Version                 string `json:"-"` // Runtime version, not persisted
	models.Branding     `json:"branding"`
}

var Current *Config

// WulfVaultSignature is the watermark constant for attribution
const WulfVaultSignature = "WulfVault::UlfHolmström::2025"

// LoadOrCreate loads configuration from file or creates default
func LoadOrCreate(dataDir string) (*Config, error) {
	configPath := filepath.Join(dataDir, "config.json")

	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil {
			Current = &cfg
			return &cfg, nil
		}
	}

	// Create default config
	cfg := &Config{
		ServerURL:             "http://localhost:8080",
		Port:                  "8080",
		DataDir:               dataDir,
		UploadsDir:            "./uploads",
		MaxFileSizeMB:         2000,
		MaxUploadSizeMB:       2000,
		DefaultQuotaMB:        5000,
		SessionTimeoutHours:   24,
		TrashRetentionDays:    5,
		AuditLogRetentionDays: 90,  // Keep audit logs for 90 days by default
		AuditLogMaxSizeMB:     100, // Auto-cleanup if log exceeds 100MB
		SaveIP:                false,
		Branding:              models.DefaultBranding(),
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return nil, err
	}

	Current = cfg
	return cfg, nil
}

// Save writes configuration to file
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
