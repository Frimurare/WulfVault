// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
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
	DeletedAt     int64  `json:"deletedAt" redis:"DeletedAt"`         // Unix timestamp, 0 = not deleted
	DeletedBy     string `json:"deletedBy" redis:"DeletedBy"`         // "user", "admin", or "system"
	OriginalEmail string `json:"originalEmail" redis:"OriginalEmail"` // SHA-256 fingerprint of the address the row had before deletion
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

// stampLayout is how account and log timestamps are shown in the admin pages.
const stampLayout = "2006-01-02 15:04"

// GetLastUsedDate describes when the download account was last used.
func (d *DownloadAccount) GetLastUsedDate() string {
	switch {
	case d.LastUsed == 0:
		return "Never"
	case time.Since(time.Unix(d.LastUsed, 0)) < 2*time.Minute:
		return "Just now"
	}
	return time.Unix(d.LastUsed, 0).Format(stampLayout)
}

// GetReadableDownloadDate formats the moment the download happened.
func (d *DownloadLog) GetReadableDownloadDate() string {
	if d.DownloadedAt <= 0 {
		return "Unknown"
	}
	return time.Unix(d.DownloadedAt, 0).Format(stampLayout)
}
