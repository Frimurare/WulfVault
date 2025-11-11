package email

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"

	"github.com/Frimurare/Sharecare/internal/database"
)

// GetOrCreateMasterKey hämtar eller skapar krypteringsnyckeln från databasen
func GetOrCreateMasterKey(db *database.Database) ([]byte, error) {
	keyHex, err := db.GetConfigValue("email_encryption_key")
	if err != nil || keyHex == "" {
		// Skapa ny 32-byte nyckel för AES-256
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		keyHex = hex.EncodeToString(key)
		if err := db.SetConfigValue("email_encryption_key", keyHex); err != nil {
			return nil, err
		}
		log.Printf("Created new email encryption master key")
		return key, nil
	}
	return hex.DecodeString(keyHex)
}

// EncryptAPIKey krypterar en API-nyckel med AES-256-GCM
func EncryptAPIKey(plaintext string, masterKey []byte) (string, error) {
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

// DecryptAPIKey dekrypterar en krypterad API-nyckel
func DecryptAPIKey(ciphertext string, masterKey []byte) (string, error) {
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
