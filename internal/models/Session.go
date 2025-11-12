// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

// Session contains cookie parameter
type Session struct {
	RenewAt    int64 `redis:"renew_at"`
	ValidUntil int64 `redis:"valid_until"`
	UserId     int   `redis:"user_id"`
}
