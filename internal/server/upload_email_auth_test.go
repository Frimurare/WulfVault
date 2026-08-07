// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// authInstructionMarker is one sentence from email.AuthInstructionsHTML/Text.
// Both bodies of the recipient email are checked for it, so the test fails if
// either the HTML or the plain text version loses the block.
const authInstructionMarker = "You will be asked to confirm your email address"

// startFakeSMTP runs a throwaway SMTP server that answers the handful of
// commands the plain-SMTP provider sends and hands every accepted message to
// the returned channel.
func startFakeSMTP(t *testing.T) (host string, port int, messages <-chan string) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	received := make(chan string, 4)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go serveFakeSMTP(conn, received)
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port, received
}

func serveFakeSMTP(conn net.Conn, received chan<- string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	fmt.Fprint(conn, "220 wulfvault-test ESMTP\r\n")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.ToUpper(strings.TrimSpace(line))

		switch {
		case strings.HasPrefix(command, "EHLO"):
			fmt.Fprint(conn, "250-wulfvault-test\r\n250 HELP\r\n")
		case command == "DATA":
			fmt.Fprint(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
			var body strings.Builder
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimRight(dataLine, "\r\n") == "." {
					break
				}
				body.WriteString(dataLine)
			}
			fmt.Fprint(conn, "250 OK\r\n")
			received <- body.String()
		case command == "QUIT":
			fmt.Fprint(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprint(conn, "250 OK\r\n")
		}
	}
}

// newEmailTestServer returns a server with an uploads directory and the fake
// SMTP host wired in as the active email provider.
func newEmailTestServer(t *testing.T) (*Server, *models.User, <-chan string) {
	t.Helper()

	dir := t.TempDir()
	if err := database.Initialize(dir); err != nil {
		t.Fatalf("Initialize database: %v", err)
	}

	host, port, messages := startFakeSMTP(t)
	now := time.Now().Unix()
	_, err := database.DB.Exec(`
		INSERT INTO EmailProviderConfig
			(Provider, IsActive, SMTPHost, SMTPPort, SMTPUseTLS, FromEmail, FromName, CreatedAt, UpdatedAt)
		VALUES ('smtp', 1, ?, ?, 0, 'vault@example.com', 'WulfVault', ?, ?)`,
		host, port, now, now)
	if err != nil {
		t.Fatalf("configure SMTP provider: %v", err)
	}

	user := &models.User{
		Name:           "Uploader",
		Email:          "uploader@example.com",
		Password:       "hash",
		IsActive:       true,
		StorageQuotaMB: 1024,
	}
	if err := database.DB.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	return New(&config.Config{ServerURL: "https://vault.example.com", UploadsDir: dir}), user, messages
}

// waitForMessage returns the next accepted message, or fails the test.
func waitForMessage(t *testing.T, messages <-chan string) string {
	t.Helper()

	select {
	case message := <-messages:
		return message
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the recipient email")
		return ""
	}
}

// legacyUpload posts a small file to the multipart /upload endpoint.
func legacyUpload(t *testing.T, s *Server, user *models.User, requireAuth bool) {
	t.Helper()

	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "report.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	part.Write([]byte("wulfvault"))
	form.WriteField("send_to_email", "recipient@example.com")
	if requireAuth {
		form.WriteField("require_auth", "true")
	}
	form.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	s.handleUpload(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

// chunkedUpload runs the init/chunk/complete sequence of the modern endpoint.
func chunkedUpload(t *testing.T, s *Server, user *models.User, requireAuth bool) {
	t.Helper()

	const payload = "wulfvault"
	metadata := map[string]string{"send_to_email": "recipient@example.com"}
	if requireAuth {
		metadata["require_auth"] = "true"
	}
	initBody, err := json.Marshal(map[string]interface{}{
		"filename":   "report.pdf",
		"total_size": len(payload),
		"metadata":   metadata,
	})
	if err != nil {
		t.Fatalf("marshal init request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload/init", bytes.NewReader(initBody))
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	s.handleChunkedUploadInit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("init status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var initResponse struct {
		UploadID string `json:"upload_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &initResponse); err != nil {
		t.Fatalf("decode init response: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/upload/chunk?upload_id="+initResponse.UploadID+"&chunk_index=0", strings.NewReader(payload))
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec = httptest.NewRecorder()
	s.handleChunkedUploadChunk(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("chunk status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/upload/complete?upload_id="+initResponse.UploadID, nil)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec = httptest.NewRecorder()
	s.handleChunkedUploadComplete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUploadEmailsExplainTheIdentityCheck(t *testing.T) {
	uploads := map[string]func(*testing.T, *Server, *models.User, bool){
		"legacy":  legacyUpload,
		"chunked": chunkedUpload,
	}

	for name, upload := range uploads {
		t.Run(name+"/auth required", func(t *testing.T) {
			s, user, messages := newEmailTestServer(t)
			upload(t, s, user, true)

			message := waitForMessage(t, messages)
			if count := strings.Count(message, authInstructionMarker); count != 2 {
				t.Errorf("found the auth instructions %d times, want 2 (once in HTML, once in plain text)\n%s", count, message)
			}
		})

		t.Run(name+"/direct link", func(t *testing.T) {
			s, user, messages := newEmailTestServer(t)
			upload(t, s, user, false)

			message := waitForMessage(t, messages)
			if strings.Contains(message, authInstructionMarker) {
				t.Errorf("a direct-link email must not explain the identity check\n%s", message)
			}
			if !strings.Contains(message, "A file has been shared with you") {
				t.Errorf("the rest of the email must be untouched\n%s", message)
			}
		})
	}
}
