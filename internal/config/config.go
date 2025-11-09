package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Frimurare/Sharecare/internal/models"
)

// Config holds the application configuration
type Config struct {
	ServerURL            string `json:"serverUrl"`
	Port                 string `json:"port"`
	DataDir              string `json:"dataDir"`
	UploadsDir           string `json:"uploadsDir"`
	MaxFileSizeMB        int    `json:"maxFileSizeMB"`
	MaxUploadSizeMB      int    `json:"maxUploadSizeMB"`
	DefaultQuotaMB       int64  `json:"defaultQuotaMB"`
	SessionTimeoutHours  int    `json:"sessionTimeoutHours"`
	SaveIP               bool   `json:"saveIp"`
	models.Branding      `json:"branding"`
}

var Current *Config

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
		ServerURL:           "http://localhost:8080",
		Port:                "8080",
		DataDir:             dataDir,
		UploadsDir:          "./uploads",
		MaxFileSizeMB:       2000,
		MaxUploadSizeMB:     2000,
		DefaultQuotaMB:      5000,
		SessionTimeoutHours: 24,
		SaveIP:              false,
		Branding:            models.DefaultBranding(),
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
