// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"strings"
	"testing"

	"github.com/Frimurare/WulfVault/internal/database"
)

// markers that must only show up when the file requires authentication.
// Compared against whitespace-normalized bodies, since the templates wrap the
// same sentences differently in HTML and plain text.
var authMarkers = []string{
	"You will be asked to confirm your email address",
	"Open the download link in this email.",
	"Enter the email address this message was sent to and choose a password of your own.",
	"Download the file.",
	"data protection rules such as the GDPR ask for",
	"you can delete the account right away in your account settings",
	"You will not receive newsletters, marketing or any other email from us.",
}

// normalize collapses all whitespace runs to a single space so a marker can be
// matched regardless of how the template happens to wrap or indent it.
func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func testFile(requireAuth bool) *database.FileInfo {
	return &database.FileInfo{
		Id:          "abc123",
		Name:        "quarterly-report.pdf",
		Size:        "2.4 MB",
		RequireAuth: requireAuth,
	}
}

func TestGenerateSplashLinkHTMLIncludesAuthInstructions(t *testing.T) {
	body := GenerateSplashLinkHTML("https://vault.example.com/s/abc123", testFile(true), "", "sender@example.com")

	for _, marker := range authMarkers {
		if !strings.Contains(normalize(body), marker) {
			t.Errorf("expected HTML body to contain %q when authentication is required", marker)
		}
	}
}

func TestGenerateSplashLinkHTMLOmitsAuthInstructions(t *testing.T) {
	body := GenerateSplashLinkHTML("https://vault.example.com/s/abc123", testFile(false), "", "sender@example.com")

	for _, marker := range authMarkers {
		if strings.Contains(normalize(body), marker) {
			t.Errorf("expected HTML body NOT to contain %q for a direct link", marker)
		}
	}

	// The rest of the email must be untouched by the conditional block.
	for _, marker := range []string{"quarterly-report.pdf", "2.4 MB", "https://vault.example.com/s/abc123", "📥 Download file"} {
		if !strings.Contains(normalize(body), marker) {
			t.Errorf("expected HTML body to still contain %q", marker)
		}
	}
}

func TestGenerateSplashLinkTextIncludesAuthInstructions(t *testing.T) {
	body := GenerateSplashLinkText("https://vault.example.com/s/abc123", testFile(true), "", "sender@example.com")

	for _, marker := range authMarkers {
		if !strings.Contains(normalize(body), marker) {
			t.Errorf("expected text body to contain %q when authentication is required", marker)
		}
	}

	if strings.Contains(body, "<") {
		t.Error("plain text body must not contain HTML markup")
	}
}

func TestGenerateSplashLinkTextOmitsAuthInstructions(t *testing.T) {
	body := GenerateSplashLinkText("https://vault.example.com/s/abc123", testFile(false), "", "")

	for _, marker := range authMarkers {
		if strings.Contains(normalize(body), marker) {
			t.Errorf("expected text body NOT to contain %q for a direct link", marker)
		}
	}

	// Layout for direct links must stay byte-identical to the pre-change output.
	want := `A file has been shared with you


Filename: quarterly-report.pdf
Size: 2.4 MB

Download: https://vault.example.com/s/abc123

---
This is an automated message from WulfVault Secure File Transfer.
`
	if body != want {
		t.Errorf("direct-link text body changed.\ngot:\n%q\nwant:\n%q", body, want)
	}
}

func TestAuthInstructionsHTML(t *testing.T) {
	if got := AuthInstructionsHTML(false); got != "" {
		t.Errorf("expected empty string when authentication is not required, got %q", got)
	}

	got := AuthInstructionsHTML(true)
	if got == "" {
		t.Fatal("expected a block when authentication is required")
	}
	if !strings.HasPrefix(got, "<div") || !strings.HasSuffix(got, "</div>") {
		t.Error("expected the block to be a single self-contained div")
	}
	if strings.Contains(got, "class=") {
		t.Error("block must use inline styles so it renders in emails without a <style> section")
	}
}

func TestAuthInstructionsText(t *testing.T) {
	if got := AuthInstructionsText(false); got != "" {
		t.Errorf("expected empty string when authentication is not required, got %q", got)
	}

	got := AuthInstructionsText(true)
	if got == "" {
		t.Fatal("expected a block when authentication is required")
	}
	if !strings.HasPrefix(got, "\n") || !strings.HasSuffix(got, "\n") {
		t.Error("expected the block to be padded with a blank line on each side")
	}
}

func TestGenerateSplashLinkKeepsSenderAndMessageBlocks(t *testing.T) {
	// The new block must not interfere with the existing sender/comment blocks.
	htmlBody := GenerateSplashLinkHTML("https://vault.example.com/s/abc123", testFile(true), "Here is the report", "sender@example.com")
	for _, marker := range []string{"<strong>Sent by:</strong> sender@example.com", "<strong>Message:</strong>", "Here is the report"} {
		if !strings.Contains(htmlBody, marker) {
			t.Errorf("expected HTML body to contain %q", marker)
		}
	}

	textBody := GenerateSplashLinkText("https://vault.example.com/s/abc123", testFile(true), "Here is the report", "sender@example.com")
	for _, marker := range []string{"Sent by: sender@example.com", "Meddelande: Here is the report"} {
		if !strings.Contains(textBody, marker) {
			t.Errorf("expected text body to contain %q", marker)
		}
	}
}
