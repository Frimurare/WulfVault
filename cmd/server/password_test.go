// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package main

import (
	"strings"
	"testing"
)

func TestGenerateRandomPasswordLengthAndAlphabet(t *testing.T) {
	password, err := generateRandomPassword()
	if err != nil {
		t.Fatalf("generateRandomPassword: %v", err)
	}

	if len(password) != passwordLength {
		t.Errorf("length = %d, want %d", len(password), passwordLength)
	}

	for _, c := range password {
		if !strings.ContainsRune(passwordAlphabet, c) {
			t.Errorf("password contains character %q which is not in the alphabet", c)
		}
	}

	// Easily confused characters must not appear.
	for _, c := range "0O1lI" {
		if strings.ContainsRune(password, c) {
			t.Errorf("password contains ambiguous character %q", c)
		}
	}
}

func TestGenerateRandomPasswordIsNotPredictable(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		password, err := generateRandomPassword()
		if err != nil {
			t.Fatalf("generateRandomPassword: %v", err)
		}
		if seen[password] {
			t.Fatalf("generateRandomPassword repeated a password after %d calls", i+1)
		}
		seen[password] = true
	}
}
