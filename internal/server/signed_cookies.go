// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
	"sync"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/secrets"
)

// The download-flow cookies used to hold bare values that any client could
// fabricate: password_verified_<id> held the literal string "true" and the
// session cookies held a plain email address. Every value is now bound to the
// install's master key with an HMAC, so a cookie only counts when this server
// issued it. Old-format cookies fail verification and the visitor simply
// authenticates again.

var (
	cookieKeyOnce sync.Once
	cookieKey     []byte
)

func cookieSigningKey() []byte {
	cookieKeyOnce.Do(func() {
		key, err := secrets.GetOrCreateMasterKey(database.DB)
		if err != nil {
			log.Printf("Warning: cookie signing key unavailable, download cookies disabled: %v", err)
			return
		}
		cookieKey = key
	})
	return cookieKey
}

// cookieMAC computes an HMAC over the given parts. Parts are length-separated
// so ("ab","c") and ("a","bc") can never collide. Returns "" when no signing
// key is available, which makes every verification fail closed.
func cookieMAC(parts ...string) string {
	key := cookieSigningKey()
	if len(key) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, key)
	for _, p := range parts {
		mac.Write([]byte(p))
		mac.Write([]byte{0})
	}
	return hex.EncodeToString(mac.Sum(nil))
}

// passwordProofValue is the value stored in password_verified_<fileID> after a
// correct file password has been posted.
func passwordProofValue(fileID string) string {
	return cookieMAC("password", fileID)
}

func verifyPasswordProof(fileID, value string) bool {
	want := passwordProofValue(fileID)
	return want != "" && hmac.Equal([]byte(value), []byte(want))
}

// downloadAccountScope is the scope of the site-wide download_session cookie;
// the per-file download_session_<id> cookies use the file ID as scope.
const downloadAccountScope = "account"

// sessionCookieValue builds a signed "email|mac" session cookie value.
func sessionCookieValue(scope, email string) string {
	mac := cookieMAC("session", scope, email)
	if mac == "" {
		return ""
	}
	return email + "|" + mac
}

// sessionCookieEmail extracts the email from a signed session cookie value.
// Returns "" when the value was not issued by this server for that scope.
func sessionCookieEmail(scope, value string) string {
	email, mac, ok := strings.Cut(value, "|")
	if !ok {
		return ""
	}
	want := cookieMAC("session", scope, email)
	if want == "" || !hmac.Equal([]byte(mac), []byte(want)) {
		return ""
	}
	return email
}

// accountSessionValue builds the signed value of the site-wide
// download_session cookie: "email|flag|mac", where flag is "r" for a
// Remember-Me login and "s" for a standard one. The flag is covered by the
// MAC, so a client cannot upgrade its own session to Remember-Me.
func accountSessionValue(email string, remember bool) string {
	flag := "s"
	if remember {
		flag = "r"
	}
	mac := cookieMAC("session", downloadAccountScope, flag, email)
	if mac == "" {
		return ""
	}
	return email + "|" + flag + "|" + mac
}

// accountSessionEmail verifies a download_session value and returns the email
// and whether the session was issued as Remember-Me. Returns "" when the
// value was not issued by this server.
func accountSessionEmail(value string) (string, bool) {
	parts := strings.SplitN(value, "|", 3)
	if len(parts) != 3 {
		return "", false
	}
	email, flag, mac := parts[0], parts[1], parts[2]
	want := cookieMAC("session", downloadAccountScope, flag, email)
	if want == "" || !hmac.Equal([]byte(mac), []byte(want)) {
		return "", false
	}
	return email, flag == "r"
}
