// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/secrets"
)

// GetOrCreateMasterKey hämtar eller skapar krypteringsnyckeln från databasen.
// Delegerar till internal/secrets så samma nyckel används av alla
// secret-konsumenter (e-postleverantörer, OIDC m.fl.).
func GetOrCreateMasterKey(db *database.Database) ([]byte, error) {
	return secrets.GetOrCreateMasterKey(db)
}

// EncryptAPIKey krypterar en API-nyckel med AES-256-GCM.
func EncryptAPIKey(plaintext string, masterKey []byte) (string, error) {
	return secrets.Encrypt(plaintext, masterKey)
}

// DecryptAPIKey dekrypterar en krypterad API-nyckel.
func DecryptAPIKey(ciphertext string, masterKey []byte) (string, error) {
	return secrets.Decrypt(ciphertext, masterKey)
}
