package models

import (
	"encoding/json"
	"time"
)

// DownloadAccount represents a temporary account created when someone downloads a file with authentication
type DownloadAccount struct {
	Id            int    `json:"id" redis:"id"`
	Name          string `json:"name" redis:"Name"`
	Email         string `json:"email" redis:"Email"`
	Password      string `json:"-" redis:"Password"` // Hashed password
	CreatedAt     int64  `json:"createdAt" redis:"CreatedAt"`
	LastUsed      int64  `json:"lastUsed" redis:"LastUsed"`
	DownloadCount int    `json:"downloadCount" redis:"DownloadCount"`
	IsActive      bool   `json:"isActive" redis:"IsActive"`
}

// DownloadLog tracks individual download events
type DownloadLog struct {
	Id                int    `json:"id"`
	FileId            string `json:"fileId"`            // The file that was downloaded
	DownloadAccountId int    `json:"downloadAccountId"` // If authenticated download, the account ID
	Email             string `json:"email"`             // Email of downloader (if authenticated)
	IpAddress         string `json:"ipAddress"`         // Optional IP tracking
	UserAgent         string `json:"userAgent"`         // Optional browser tracking
	DownloadedAt      int64  `json:"downloadedAt"`      // Unix timestamp
	FileSize          int64  `json:"fileSize"`          // Size in bytes
	FileName          string `json:"fileName"`          // Name of file downloaded
	IsAuthenticated   bool   `json:"isAuthenticated"`   // True if download required authentication
}

// GetReadableDate returns the date as YYYY-MM-DD HH:MM
func (d *DownloadAccount) GetReadableDate() string {
	if d.CreatedAt == 0 {
		return "Unknown"
	}
	return time.Unix(d.CreatedAt, 0).Format("2006-01-02 15:04")
}

// GetLastUsedDate returns the last used date as YYYY-MM-DD HH:MM
func (d *DownloadAccount) GetLastUsedDate() string {
	if d.LastUsed == 0 {
		return "Never"
	}
	if time.Now().Unix()-d.LastUsed < 120 {
		return "Just now"
	}
	return time.Unix(d.LastUsed, 0).Format("2006-01-02 15:04")
}

// ToJson returns the download account as a JSON object
func (d *DownloadAccount) ToJson() string {
	result, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(result)
}

// GetReadableDownloadDate returns the download timestamp as YYYY-MM-DD HH:MM
func (d *DownloadLog) GetReadableDownloadDate() string {
	if d.DownloadedAt == 0 {
		return "Unknown"
	}
	return time.Unix(d.DownloadedAt, 0).Format("2006-01-02 15:04")
}

// ToJson returns the download log as a JSON object
func (d *DownloadLog) ToJson() string {
	result, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(result)
}
