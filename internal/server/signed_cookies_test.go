// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"strings"
	"testing"
)

func TestPasswordProofRoundTrip(t *testing.T) {
	newTestServer(t)

	proof := passwordProofValue("somefile")
	if proof == "" {
		t.Fatal("no proof issued")
	}
	if !verifyPasswordProof("somefile", proof) {
		t.Fatal("genuine proof rejected")
	}
	if verifyPasswordProof("otherfile", proof) {
		t.Fatal("proof for one file accepted for another")
	}
	if verifyPasswordProof("somefile", "true") {
		t.Fatal("legacy literal accepted")
	}
}

func TestAccountSessionRememberFlag(t *testing.T) {
	newTestServer(t)

	remembered := accountSessionValue("anna@example.com", true)
	standard := accountSessionValue("anna@example.com", false)

	if email, remember := accountSessionEmail(remembered); email != "anna@example.com" || !remember {
		t.Fatalf("remembered session parsed as (%q, %v)", email, remember)
	}
	if email, remember := accountSessionEmail(standard); email != "anna@example.com" || remember {
		t.Fatalf("standard session parsed as (%q, %v)", email, remember)
	}

	// Flipping the flag without re-signing must invalidate the cookie
	tampered := strings.Replace(standard, "|s|", "|r|", 1)
	if email, _ := accountSessionEmail(tampered); email != "" {
		t.Fatal("tampered remember flag accepted")
	}

	if email, _ := accountSessionEmail("anna@example.com"); email != "" {
		t.Fatal("bare email accepted")
	}
}
