// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// Package secrets provides AES-256-GCM symmetric encryption for at-rest secrets
// (email API keys, OIDC client secrets, etc.) stored in the Configuration table.
// One master key per install, lazily generated and persisted via configKey.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
)

const configKey = "secret_master_key"

// LegacyConfigKey is the pre-secrets-package key used by internal/email.
// Kept so existing email API keys keep decrypting after the refactor.
const LegacyConfigKey = "email_encryption_key"

// ConfigStore is the minimal interface secrets needs from the database layer.
// Implemented by *database.Database.
type ConfigStore interface {
	GetConfigValue(key string) (string, error)
	SetConfigValue(key, value string) error
}

// GetOrCreateMasterKey returns the install's master encryption key.
// Honors LegacyConfigKey first so previously-encrypted email keys keep working.
func GetOrCreateMasterKey(db ConfigStore) ([]byte, error) {
	// Prefer the legacy key if present — pre-existing installs already have
	// data encrypted under it.
	if keyHex, err := db.GetConfigValue(LegacyConfigKey); err == nil && keyHex != "" {
		return hex.DecodeString(keyHex)
	}
	if keyHex, err := db.GetConfigValue(configKey); err == nil && keyHex != "" {
		return hex.DecodeString(keyHex)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	keyHex := hex.EncodeToString(key)
	if err := db.SetConfigValue(configKey, keyHex); err != nil {
		return nil, err
	}
	log.Printf("Created new secrets master key")
	return key, nil
}

// Encrypt encrypts plaintext with AES-256-GCM. Empty input returns empty output.
func Encrypt(plaintext string, masterKey []byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Empty input returns an error.
func Decrypt(ciphertext string, masterKey []byte) (string, error) {
	if ciphertext == "" {
		return "", errors.New("no encrypted data provided")
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
