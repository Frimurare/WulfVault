// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Frimurare/WulfVault/internal/config"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

func TestParseChunkRequest(t *testing.T) {
	const fileSize = int64(100)

	tests := []struct {
		name        string
		query       string
		rangeHeader string
		fileSize    int64
		wantOffset  int64
		wantSize    int64
		wantErr     error
	}{
		{name: "defaults to the first chunk", fileSize: fileSize, wantSize: fileSize},
		{name: "explicit offset and size", query: "offset=10&size=20", fileSize: fileSize, wantOffset: 10, wantSize: 20},
		{name: "size is clamped to the end of the file", query: "offset=90&size=50", fileSize: fileSize, wantOffset: 90, wantSize: 10},
		{name: "last byte", query: "offset=99&size=1", fileSize: fileSize, wantOffset: 99, wantSize: 1},
		{name: "size above the cap is clamped", query: "offset=0&size=999999999", fileSize: 4 * maxDownloadChunkSize, wantSize: maxDownloadChunkSize},
		{name: "negative offset", query: "offset=-1", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "negative size", query: "offset=0&size=-5", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "zero size", query: "offset=0&size=0", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "non numeric offset", query: "offset=abc", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "non numeric size", query: "size=10MB", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "offset at end of file", query: "offset=100", fileSize: fileSize, wantErr: errChunkOutOfRange},
		{name: "offset past end of file", query: "offset=101", fileSize: fileSize, wantErr: errChunkOutOfRange},
		{name: "offset far past end of file", query: "offset=999999999999", fileSize: fileSize, wantErr: errChunkOutOfRange},
		{name: "empty file allows only offset zero", fileSize: 0, wantOffset: 0, wantSize: 0},
		{name: "empty file rejects a non zero offset", query: "offset=1", fileSize: 0, wantErr: errChunkOutOfRange},

		{name: "range header", rangeHeader: "bytes=10-19", fileSize: fileSize, wantOffset: 10, wantSize: 10},
		{name: "range header without end", rangeHeader: "bytes=40-", fileSize: fileSize, wantOffset: 40, wantSize: 60},
		{name: "range header clamped to the file", rangeHeader: "bytes=90-500", fileSize: fileSize, wantOffset: 90, wantSize: 10},
		{name: "range header wins over the query", query: "offset=0&size=5", rangeHeader: "bytes=50-59", fileSize: fileSize, wantOffset: 50, wantSize: 10},
		{name: "range header past the end", rangeHeader: "bytes=200-300", fileSize: fileSize, wantErr: errChunkOutOfRange},
		{name: "multi range is rejected", rangeHeader: "bytes=0-9,20-29", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "suffix range is rejected", rangeHeader: "bytes=-500", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "reversed range is rejected", rangeHeader: "bytes=50-10", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "unknown range unit is rejected", rangeHeader: "items=0-9", fileSize: fileSize, wantErr: errChunkInvalid},
		{name: "malformed range is rejected", rangeHeader: "bytes=abc-def", fileSize: fileSize, wantErr: errChunkInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := url.ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("bad test query %q: %v", tt.query, err)
			}

			offset, size, err := parseChunkRequest(query, tt.rangeHeader, tt.fileSize)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if offset != tt.wantOffset || size != tt.wantSize {
				t.Fatalf("got offset=%d size=%d, want offset=%d size=%d", offset, size, tt.wantOffset, tt.wantSize)
			}
			if offset+size > tt.fileSize {
				t.Fatalf("interval %d..%d reaches outside a file of %d bytes", offset, offset+size, tt.fileSize)
			}
		})
	}
}

func TestEvaluateDownloadAccess(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	openFile := func() *database.FileInfo {
		return &database.FileInfo{
			Id:                 "abc",
			ExpireAt:           now.Unix() + 3600,
			DownloadsRemaining: 5,
		}
	}

	tests := []struct {
		name     string
		file     func() *database.FileInfo
		creds    downloadCredentials
		wantCode string
	}{
		{
			name: "open file is served",
			file: openFile,
		},
		{
			name: "expired file is refused",
			file: func() *database.FileInfo {
				f := openFile()
				f.ExpireAt = now.Unix() - 1
				return f
			},
			wantCode: "expired",
		},
		{
			name: "expiry is ignored when time is unlimited",
			file: func() *database.FileInfo {
				f := openFile()
				f.ExpireAt = now.Unix() - 1
				f.UnlimitedTime = true
				return f
			},
		},
		{
			name: "spent download counter is refused",
			file: func() *database.FileInfo {
				f := openFile()
				f.DownloadsRemaining = 0
				return f
			},
			wantCode: "download_limit_reached",
		},
		{
			name: "counter is ignored when downloads are unlimited",
			file: func() *database.FileInfo {
				f := openFile()
				f.DownloadsRemaining = 0
				f.UnlimitedDownloads = true
				return f
			},
		},
		{
			name: "password protected file without the cookie is refused",
			file: func() *database.FileInfo {
				f := openFile()
				f.FilePasswordPlain = "hunter2"
				return f
			},
			wantCode: "password_required",
		},
		{
			name: "password protected file with a verified cookie is served",
			file: func() *database.FileInfo {
				f := openFile()
				f.FilePasswordPlain = "hunter2"
				return f
			},
			creds: downloadCredentials{PasswordVerified: true},
		},
		{
			name: "a logged in user does not bypass the password",
			file: func() *database.FileInfo {
				f := openFile()
				f.FilePasswordPlain = "hunter2"
				return f
			},
			creds:    downloadCredentials{User: &models.User{Id: 1}},
			wantCode: "password_required",
		},
		{
			name: "file requiring auth without credentials is refused",
			file: func() *database.FileInfo {
				f := openFile()
				f.RequireAuth = true
				return f
			},
			wantCode: "auth_required",
		},
		{
			name: "file requiring auth is served to a logged in user",
			file: func() *database.FileInfo {
				f := openFile()
				f.RequireAuth = true
				return f
			},
			creds: downloadCredentials{User: &models.User{Id: 1}},
		},
		{
			name: "file requiring auth is served to a download account",
			file: func() *database.FileInfo {
				f := openFile()
				f.RequireAuth = true
				return f
			},
			creds: downloadCredentials{Account: &models.DownloadAccount{Id: 7, IsActive: true}},
		},
		{
			name: "password and auth both have to be satisfied",
			file: func() *database.FileInfo {
				f := openFile()
				f.RequireAuth = true
				f.FilePasswordPlain = "hunter2"
				return f
			},
			creds:    downloadCredentials{Account: &models.DownloadAccount{Id: 7, IsActive: true}},
			wantCode: "password_required",
		},
		{
			name: "expiry is checked before the password",
			file: func() *database.FileInfo {
				f := openFile()
				f.ExpireAt = now.Unix() - 1
				f.FilePasswordPlain = "hunter2"
				return f
			},
			wantCode: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			denial := evaluateDownloadAccess(tt.file(), tt.creds, now)

			if tt.wantCode == "" {
				if denial != nil {
					t.Fatalf("access refused with %q, want it granted", denial.Code)
				}
				return
			}
			if denial == nil {
				t.Fatalf("access granted, want it refused with %q", tt.wantCode)
			}
			if denial.Code != tt.wantCode {
				t.Fatalf("refused with %q, want %q", denial.Code, tt.wantCode)
			}
			if denial.Status < 400 {
				t.Fatalf("refusal carries status %d, want a 4xx or 5xx", denial.Status)
			}
		})
	}
}

func TestCalculateFileSHA256(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "empty file", content: "", want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{name: "abc", content: "abc", want: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "payload")
			if err := os.WriteFile(path, []byte(tt.content), 0600); err != nil {
				t.Fatalf("could not write test file: %v", err)
			}

			got, err := calculateFileSHA256(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCalculateFileSHA256MissingFile(t *testing.T) {
	if _, err := calculateFileSHA256(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected an error for a file that does not exist")
	}
}

func TestFileSHA256IsCached(t *testing.T) {
	resetDigestCache()

	dir := t.TempDir()
	path := filepath.Join(dir, "payload")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}

	if _, ready := peekFileSHA256("cached-file"); ready {
		t.Fatal("checksum reported as ready before anything was computed")
	}

	first, err := fileSHA256("cached-file", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cached, ready := peekFileSHA256("cached-file")
	if !ready || cached != first {
		t.Fatalf("peek returned (%q, %v), want (%q, true)", cached, ready, first)
	}

	// Uploaded files never change, so a second call must not read the file
	// again - overwriting it here proves the cached value is the one returned.
	if err := os.WriteFile(path, []byte("something else entirely"), 0600); err != nil {
		t.Fatalf("could not rewrite test file: %v", err)
	}

	second, err := fileSHA256("cached-file", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second != first {
		t.Fatalf("got %s on the second call, want the cached %s", second, first)
	}
}

func TestFileSHA256DoesNotCacheFailures(t *testing.T) {
	resetDigestCache()

	dir := t.TempDir()
	path := filepath.Join(dir, "payload")

	if _, err := fileSHA256("late-file", path); err == nil {
		t.Fatal("expected an error for a file that does not exist")
	}

	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}

	got, err := fileSHA256("late-file", path)
	if err != nil {
		t.Fatalf("unexpected error after the file appeared: %v", err)
	}
	if got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("got %s, want the checksum of the file that now exists", got)
	}
}

func TestEvictOldestDigests(t *testing.T) {
	resetDigestCache()

	digestCacheMu.Lock()
	for i := 0; i < sha256CacheLimit+10; i++ {
		ready := make(chan struct{})
		close(ready)
		digestCache[string(rune('a'+i%26))+string(rune('a'+i/26))] = &fileDigest{
			value:      "x",
			ready:      ready,
			computedAt: time.Unix(int64(i), 0),
		}
	}
	evictOldestDigests()
	size := len(digestCache)
	_, oldestStillThere := digestCache["aa"]
	digestCacheMu.Unlock()

	if size > sha256CacheLimit {
		t.Fatalf("cache holds %d entries, want at most %d", size, sha256CacheLimit)
	}
	if oldestStillThere {
		t.Fatal("the oldest entry survived eviction")
	}
}

func resetDigestCache() {
	digestCacheMu.Lock()
	digestCache = make(map[string]*fileDigest)
	digestCacheMu.Unlock()
}

// --- endpoint tests -------------------------------------------------------

type testServer struct {
	*Server
	uploadsDir string
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	resetDigestCache()

	dir := t.TempDir()
	uploadsDir := filepath.Join(dir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("could not create uploads dir: %v", err)
	}

	if err := database.Initialize(filepath.Join(dir, "data")); err != nil {
		t.Fatalf("could not initialise database: %v", err)
	}
	t.Cleanup(func() { database.DB.Close() })

	// Files reference an owner, so there has to be one for the metadata to save.
	if _, err := database.DB.Exec(
		`INSERT INTO Users (Id, Name, Email, Password, CreatedAt) VALUES (1, 'owner', 'owner@example.test', '', ?)`,
		time.Now().Unix(),
	); err != nil {
		t.Fatalf("could not create file owner: %v", err)
	}

	return &testServer{
		Server: New(&config.Config{
			ServerURL:  "http://localhost:8080",
			Port:       "8080",
			DataDir:    dir,
			UploadsDir: uploadsDir,
		}),
		uploadsDir: uploadsDir,
	}
}

// storeFile writes a payload to disk and registers it, the way an upload would.
func (ts *testServer) storeFile(t *testing.T, id string, payload []byte, customise func(*database.FileInfo)) *database.FileInfo {
	t.Helper()

	if err := os.WriteFile(filepath.Join(ts.uploadsDir, id), payload, 0600); err != nil {
		t.Fatalf("could not write payload: %v", err)
	}

	fileInfo := &database.FileInfo{
		Id:                 id,
		Name:               id + ".bin",
		Size:               database.FormatFileSize(int64(len(payload))),
		SHA1:               "",
		ContentType:        "application/octet-stream",
		SizeBytes:          int64(len(payload)),
		UploadDate:         time.Now().Unix(),
		DownloadsRemaining: 5,
		UserId:             1,
	}
	if customise != nil {
		customise(fileInfo)
	}
	if err := database.DB.SaveFile(fileInfo); err != nil {
		t.Fatalf("could not save file metadata: %v", err)
	}
	return fileInfo
}

func (ts *testServer) get(path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	ts.handleAPIDownloadRoutes(recorder, request)
	return recorder
}

func testPayload(size int) []byte {
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	return payload
}

func TestDownloadInfoReturnsMetadata(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(4096)
	ts.storeFile(t, "plainfile", payload, nil)

	recorder := ts.get("/api/v1/download/plainfile/info")
	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var info struct {
		FileID    string `json:"file_id"`
		Name      string `json:"name"`
		SizeBytes int64  `json:"size_bytes"`
		ChunkSize int64  `json:"chunk_size"`
		DirectURL string `json:"direct_url"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &info); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if info.FileID != "plainfile" || info.Name != "plainfile.bin" {
		t.Fatalf("unexpected identity in response: %+v", info)
	}
	if info.SizeBytes != int64(len(payload)) {
		t.Fatalf("got size %d, want %d", info.SizeBytes, len(payload))
	}
	if info.ChunkSize != defaultDownloadChunkSize {
		t.Fatalf("got chunk size %d, want %d", info.ChunkSize, defaultDownloadChunkSize)
	}
	if info.DirectURL != "/d/plainfile" {
		t.Fatalf("got direct url %q, want /d/plainfile", info.DirectURL)
	}
}

func TestDownloadInfoRejectsUnknownFile(t *testing.T) {
	ts := newTestServer(t)

	if code := ts.get("/api/v1/download/nosuchfile/info").Code; code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", code)
	}
}

func TestDownloadChunkServesTheRequestedInterval(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(4096)
	ts.storeFile(t, "plainfile", payload, nil)

	recorder := ts.get("/api/v1/download/plainfile/chunk?offset=1000&size=500")
	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("got status %d, want 206: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Range"); got != "bytes 1000-1499/4096" {
		t.Fatalf("got Content-Range %q, want bytes 1000-1499/4096", got)
	}
	if recorder.Body.String() != string(payload[1000:1500]) {
		t.Fatalf("chunk body does not match the file contents")
	}
}

func TestDownloadChunkHonoursRangeHeader(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(4096)
	ts.storeFile(t, "plainfile", payload, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/download/plainfile/chunk", nil)
	request.Header.Set("Range", "bytes=100-199")
	recorder := httptest.NewRecorder()
	ts.handleAPIDownloadRoutes(recorder, request)

	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("got status %d, want 206", recorder.Code)
	}
	if recorder.Body.String() != string(payload[100:200]) {
		t.Fatalf("range body does not match the file contents")
	}
}

func TestDownloadChunkRejectsBadIntervals(t *testing.T) {
	ts := newTestServer(t)
	ts.storeFile(t, "plainfile", testPayload(4096), nil)

	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "negative offset", path: "/api/v1/download/plainfile/chunk?offset=-1", want: http.StatusBadRequest},
		{name: "negative size", path: "/api/v1/download/plainfile/chunk?offset=0&size=-1", want: http.StatusBadRequest},
		{name: "non numeric offset", path: "/api/v1/download/plainfile/chunk?offset=nope", want: http.StatusBadRequest},
		{name: "offset past the end", path: "/api/v1/download/plainfile/chunk?offset=4096", want: http.StatusRequestedRangeNotSatisfiable},
		{name: "offset far past the end", path: "/api/v1/download/plainfile/chunk?offset=999999999", want: http.StatusRequestedRangeNotSatisfiable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := ts.get(tt.path).Code; code != tt.want {
				t.Fatalf("got status %d, want %d", code, tt.want)
			}
		})
	}
}

func TestDownloadChunkNeverReadsPastTheFile(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(4096)
	ts.storeFile(t, "plainfile", payload, nil)

	recorder := ts.get("/api/v1/download/plainfile/chunk?offset=4000&size=1000000")
	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("got status %d, want 206", recorder.Code)
	}
	if recorder.Body.Len() != 96 {
		t.Fatalf("got %d bytes, want the 96 bytes that are left in the file", recorder.Body.Len())
	}
}

func TestDownloadEndpointsEnforcePassword(t *testing.T) {
	ts := newTestServer(t)
	ts.storeFile(t, "secretfile", testPayload(4096), func(f *database.FileInfo) {
		f.FilePasswordPlain = "hunter2"
	})

	for _, endpoint := range []string{"info", "chunk", "verify"} {
		t.Run(endpoint+" without the cookie", func(t *testing.T) {
			recorder := ts.get("/api/v1/download/secretfile/" + endpoint)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("got status %d, want 401", recorder.Code)
			}

			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("could not decode response: %v", err)
			}
			if body["code"] != "password_required" {
				t.Fatalf("got code %q, want password_required", body["code"])
			}
			if body["fallback_url"] != "/d/secretfile" {
				t.Fatalf("got fallback %q, want /d/secretfile", body["fallback_url"])
			}
		})

		t.Run(endpoint+" with the cookie", func(t *testing.T) {
			cookie := &http.Cookie{Name: "password_verified_secretfile", Value: passwordProofValue("secretfile")}
			if code := ts.get("/api/v1/download/secretfile/"+endpoint, cookie).Code; code >= 400 {
				t.Fatalf("got status %d for an unlocked file, want it served", code)
			}
		})
	}

	t.Run("a cookie for another file does not unlock this one", func(t *testing.T) {
		cookie := &http.Cookie{Name: "password_verified_otherfile", Value: passwordProofValue("otherfile")}
		if code := ts.get("/api/v1/download/secretfile/chunk", cookie).Code; code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", code)
		}
	})

	t.Run("a falsified cookie value does not unlock the file", func(t *testing.T) {
		cookie := &http.Cookie{Name: "password_verified_secretfile", Value: "hunter2"}
		if code := ts.get("/api/v1/download/secretfile/chunk", cookie).Code; code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", code)
		}
	})

	t.Run("the legacy literal true is no longer accepted", func(t *testing.T) {
		cookie := &http.Cookie{Name: "password_verified_secretfile", Value: "true"}
		if code := ts.get("/api/v1/download/secretfile/chunk", cookie).Code; code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", code)
		}
	})
}

func TestDownloadEndpointsEnforceAuthentication(t *testing.T) {
	ts := newTestServer(t)
	ts.storeFile(t, "privatefile", testPayload(4096), func(f *database.FileInfo) {
		f.RequireAuth = true
	})

	for _, endpoint := range []string{"info", "chunk", "verify"} {
		recorder := ts.get("/api/v1/download/privatefile/" + endpoint)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("%s: got status %d, want 401", endpoint, recorder.Code)
		}

		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: could not decode response: %v", endpoint, err)
		}
		if body["code"] != "auth_required" {
			t.Fatalf("%s: got code %q, want auth_required", endpoint, body["code"])
		}
	}

	t.Run("a bare email in the session cookie does not authenticate", func(t *testing.T) {
		cookie := &http.Cookie{Name: "download_session_privatefile", Value: "stranger@example.com"}
		if code := ts.get("/api/v1/download/privatefile/chunk", cookie).Code; code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", code)
		}
	})

	t.Run("a signed session for an unknown account does not authenticate", func(t *testing.T) {
		cookie := &http.Cookie{Name: "download_session_privatefile", Value: sessionCookieValue("privatefile", "stranger@example.com")}
		if code := ts.get("/api/v1/download/privatefile/chunk", cookie).Code; code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", code)
		}
	})
}

func TestDownloadEndpointsEnforceExpiryAndLimit(t *testing.T) {
	ts := newTestServer(t)
	ts.storeFile(t, "expiredfile", testPayload(64), func(f *database.FileInfo) {
		f.ExpireAt = time.Now().Add(-time.Hour).Unix()
	})
	ts.storeFile(t, "usedupfile", testPayload(64), func(f *database.FileInfo) {
		f.DownloadsRemaining = 0
	})

	for _, endpoint := range []string{"info", "chunk", "verify"} {
		if code := ts.get("/api/v1/download/expiredfile/" + endpoint).Code; code != http.StatusGone {
			t.Fatalf("%s on an expired file: got status %d, want 410", endpoint, code)
		}
		if code := ts.get("/api/v1/download/usedupfile/" + endpoint).Code; code != http.StatusGone {
			t.Fatalf("%s on a spent file: got status %d, want 410", endpoint, code)
		}
	}
}

func TestDownloadCounterIsSpentOnTheLastChunkOnly(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(4096)
	ts.storeFile(t, "countedfile", payload, func(f *database.FileInfo) {
		f.DownloadsRemaining = 1
	})

	// Every chunk but the last one must leave the counter alone, otherwise a
	// file with a single download left would expire during its own transfer.
	for offset := 0; offset < 4096-1024; offset += 1024 {
		recorder := ts.get("/api/v1/download/countedfile/chunk?offset=" + strconv.Itoa(offset) + "&size=1024")
		if recorder.Code != http.StatusPartialContent {
			t.Fatalf("chunk at %d: got status %d, want 206", offset, recorder.Code)
		}

		fileInfo, err := database.DB.GetFileByID("countedfile")
		if err != nil {
			t.Fatalf("could not reload file: %v", err)
		}
		if fileInfo.DownloadsRemaining != 1 || fileInfo.DownloadCount != 0 {
			t.Fatalf("counter moved mid-download: remaining=%d count=%d", fileInfo.DownloadsRemaining, fileInfo.DownloadCount)
		}
	}

	if code := ts.get("/api/v1/download/countedfile/chunk?offset=3072&size=1024").Code; code != http.StatusPartialContent {
		t.Fatalf("last chunk: got status %d, want 206", code)
	}

	fileInfo, err := database.DB.GetFileByID("countedfile")
	if err != nil {
		t.Fatalf("could not reload file: %v", err)
	}
	if fileInfo.DownloadsRemaining != 0 || fileInfo.DownloadCount != 1 {
		t.Fatalf("after the last chunk: remaining=%d count=%d, want 0 and 1", fileInfo.DownloadsRemaining, fileInfo.DownloadCount)
	}

	if code := ts.get("/api/v1/download/countedfile/chunk?offset=0&size=1024").Code; code != http.StatusGone {
		t.Fatalf("got status %d for a spent file, want 410", code)
	}
}

func TestDownloadVerifyMatchesTheStoredFile(t *testing.T) {
	ts := newTestServer(t)
	payload := []byte("abc")
	ts.storeFile(t, "hashedfile", payload, nil)

	recorder := ts.get("/api/v1/download/hashedfile/verify")
	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Algorithm string `json:"algorithm"`
		SHA256    string `json:"sha256"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if body.Algorithm != "sha256" {
		t.Fatalf("got algorithm %q, want sha256", body.Algorithm)
	}
	if body.SHA256 != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("got checksum %s, want the SHA-256 of the stored file", body.SHA256)
	}
}

func TestChunkedDownloadReassemblesTheWholeFile(t *testing.T) {
	ts := newTestServer(t)
	payload := testPayload(10_000)
	ts.storeFile(t, "wholefile", payload, nil)

	const chunkSize = 1024
	var assembled []byte
	for offset := 0; offset < len(payload); offset += chunkSize {
		recorder := ts.get("/api/v1/download/wholefile/chunk?offset=" + strconv.Itoa(offset) + "&size=" + strconv.Itoa(chunkSize))
		if recorder.Code != http.StatusPartialContent {
			t.Fatalf("chunk at %d: got status %d, want 206", offset, recorder.Code)
		}
		assembled = append(assembled, recorder.Body.Bytes()...)
	}

	if string(assembled) != string(payload) {
		t.Fatalf("assembled %d bytes that do not match the %d byte source", len(assembled), len(payload))
	}

	recorder := ts.get("/api/v1/download/wholefile/verify")
	if recorder.Code != http.StatusOK {
		t.Fatalf("verify: got status %d, want 200", recorder.Code)
	}
	var body struct {
		SHA256 string `json:"sha256"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	want := sha256.Sum256(assembled)
	if body.SHA256 != hex.EncodeToString(want[:]) {
		t.Fatalf("server checksum %s does not match the assembled bytes", body.SHA256)
	}
}

func TestDownloadEndpointsRejectNonGet(t *testing.T) {
	ts := newTestServer(t)
	ts.storeFile(t, "plainfile", testPayload(64), nil)

	for _, endpoint := range []string{"info", "chunk", "verify"} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/download/plainfile/"+endpoint, nil)
		recorder := httptest.NewRecorder()
		ts.handleAPIDownloadRoutes(recorder, request)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s: got status %d, want 405", endpoint, recorder.Code)
		}
	}
}
