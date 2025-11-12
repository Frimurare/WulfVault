// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU General Public License v3.0 (GPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

// DbConnection is a struct that contains the database configuration for connecting
type DbConnection struct {
	HostUrl     string
	RedisPrefix string
	Username    string
	Password    string
	RedisUseSsl bool
	Type        int
}
