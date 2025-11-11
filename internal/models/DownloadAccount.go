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

// EmailLog tracks when files are shared via email
type EmailLog struct {
	Id             int    `json:"id"`
	FileId         string `json:"fileId"`         // The file that was shared
	SenderUserId   int    `json:"senderUserId"`   // User who sent the email
	RecipientEmail string `json:"recipientEmail"` // Email of recipient
	Message        string `json:"message"`        // Optional personal message
	SentAt         int64  `json:"sentAt"`         // Unix timestamp
	FileName       string `json:"fileName"`       // Name of file shared
	FileSize       int64  `json:"fileSize"`       // Size in bytes
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

// GetReadableSentDate returns the sent timestamp as YYYY-MM-DD HH:MM
func (e *EmailLog) GetReadableSentDate() string {
	if e.SentAt == 0 {
		return "Unknown"
	}
	return time.Unix(e.SentAt, 0).Format("2006-01-02 15:04")
}

// ToJson returns the email log as a JSON object
func (e *EmailLog) ToJson() string {
	result, err := json.Marshal(e)
	if err != nil {
		return "{}"
	}
	return string(result)
}
