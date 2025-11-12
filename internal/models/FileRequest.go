// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"time"
)

// FileRequest represents a request for someone to upload files
type FileRequest struct {
	Id               int    `json:"id"`
	UserId           int    `json:"userId"`
	RequestToken     string `json:"requestToken"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	CreatedAt        int64  `json:"createdAt"`
	ExpiresAt        int64  `json:"expiresAt"`
	IsActive         bool   `json:"isActive"`
	MaxFileSize      int64  `json:"maxFileSize"`      // in MB
	AllowedFileTypes string `json:"allowedFileTypes"` // comma-separated
	UsedByIP         string `json:"usedByIP"`         // IP address that used this link
	UsedAt           int64  `json:"usedAt"`           // Unix timestamp when link was used
}

// IsExpired checks if the request has expired
func (fr *FileRequest) IsExpired() bool {
	if fr.ExpiresAt == 0 {
		return false // No expiration
	}
	return time.Now().Unix() > fr.ExpiresAt
}

// IsUsed checks if the upload link has been used
func (fr *FileRequest) IsUsed() bool {
	return fr.UsedAt > 0
}

// GetUploadURL returns the public upload URL for this request
func (fr *FileRequest) GetUploadURL(serverURL string) string {
	return serverURL + "/upload-request/" + fr.RequestToken
}
