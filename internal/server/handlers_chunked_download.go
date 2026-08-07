// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

const (
	// defaultDownloadChunkSize is what the browser uses unless it asks for
	// something else. It matches the chunk size of the upload path.
	defaultDownloadChunkSize = 25 * 1024 * 1024

	// maxDownloadChunkSize caps a single chunk request so that a client cannot
	// make the server stream an arbitrary amount of data per request.
	maxDownloadChunkSize = 64 * 1024 * 1024

	// sha256CacheLimit bounds the in-memory checksum cache. Entries are tiny,
	// but the map must not grow without limit on a long-running server.
	sha256CacheLimit = 512
)

// downloadCredentials is the proof of access a caller presented for one file.
// Collecting it needs database lookups, judging it does not - splitting the two
// keeps the access rules in one small, testable function.
type downloadCredentials struct {
	PasswordVerified bool
	User             *models.User
	Account          *models.DownloadAccount
}

// accessDenial describes why a download was refused.
type accessDenial struct {
	Status  int
	Code    string
	Message string
}

// evaluateDownloadAccess applies the same rules, in the same order, as
// handleDownload/handlePasswordProtectedDownload/handleAuthenticatedDownload.
// The browser-facing handlers answer with HTML pages, the API answers with
// JSON, but a file that is refused on /d/ must be refused here as well.
func evaluateDownloadAccess(fileInfo *database.FileInfo, creds downloadCredentials, now time.Time) *accessDenial {
	if !fileInfo.UnlimitedTime && fileInfo.ExpireAt > 0 && now.Unix() > fileInfo.ExpireAt {
		return &accessDenial{http.StatusGone, "expired", "File has expired"}
	}

	if !fileInfo.UnlimitedDownloads && fileInfo.DownloadsRemaining <= 0 {
		return &accessDenial{http.StatusGone, "download_limit_reached", "Download limit reached"}
	}

	if fileInfo.FilePasswordPlain != "" && !creds.PasswordVerified {
		return &accessDenial{http.StatusUnauthorized, "password_required", "Password required"}
	}

	if fileInfo.RequireAuth && creds.User == nil && creds.Account == nil {
		return &accessDenial{http.StatusUnauthorized, "auth_required", "Authentication required"}
	}

	return nil
}

// downloadAPICookiePath is the cookie scope of the chunked download API.
// The cookies issued by the /d/ flow are scoped to /d/<id> and are therefore
// never sent to /api/v1/download/<id>/..., so the same proof is issued a second
// time for this scope. Nothing is granted here that /d/ would not grant.
func downloadAPICookiePath(fileID string) string {
	return "/api/v1/download/" + fileID
}

// resolveDownloadCredentials reads the same cookies and sessions that the /d/
// handlers read. A download account only counts when it is active.
func (s *Server) resolveDownloadCredentials(r *http.Request, fileInfo *database.FileInfo) downloadCredentials {
	var creds downloadCredentials

	if cookie, err := r.Cookie("password_verified_" + fileInfo.Id); err == nil && cookie.Value == "true" {
		creds.PasswordVerified = true
	}

	if user, err := s.getUserFromSession(r); err == nil && user != nil {
		creds.User = user
	}

	if cookie, err := r.Cookie("download_session_" + fileInfo.Id); err == nil {
		account, err := database.DB.GetDownloadAccountByEmail(cookie.Value)
		if err == nil && account.IsActive {
			creds.Account = account
		}
	}

	return creds
}

// handleAPIDownloadRoutes dispatches /api/v1/download/<id>[/action].
func (s *Server) handleAPIDownloadRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/download/")
	fileID, action, _ := strings.Cut(rest, "/")

	switch action {
	case "info":
		s.handleDownloadInfo(w, r, fileID)
	case "chunk":
		s.handleDownloadChunk(w, r, fileID)
	case "verify":
		s.handleDownloadVerify(w, r, fileID)
	default:
		s.handleAPIDownload(w, r)
	}
}

// authorizeChunkedDownload resolves the file and enforces the access rules.
// On refusal it has already written the response.
func (s *Server) authorizeChunkedDownload(w http.ResponseWriter, r *http.Request, fileID string) (*database.FileInfo, *models.DownloadAccount, bool) {
	if r.Method != http.MethodGet {
		s.sendDownloadError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", "")
		return nil, nil, false
	}

	if fileID == "" {
		s.sendDownloadError(w, http.StatusNotFound, "not_found", "File not found", "")
		return nil, nil, false
	}

	fileInfo, err := database.DB.GetFileByID(fileID)
	if err != nil {
		s.sendDownloadError(w, http.StatusNotFound, "not_found", "File not found", "")
		return nil, nil, false
	}

	creds := s.resolveDownloadCredentials(r, fileInfo)
	if denial := evaluateDownloadAccess(fileInfo, creds, time.Now()); denial != nil {
		log.Printf("⚠️  Chunked download denied: %s | File: %s | Reason: %s | IP: %s",
			r.URL.Path, fileID, denial.Code, getClientIP(r))
		s.sendDownloadError(w, denial.Status, denial.Code, denial.Message, "/d/"+fileID)
		return nil, nil, false
	}

	return fileInfo, creds.Account, true
}

// handleDownloadInfo returns the metadata a chunked download needs to start.
func (s *Server) handleDownloadInfo(w http.ResponseWriter, r *http.Request, fileID string) {
	fileInfo, _, ok := s.authorizeChunkedDownload(w, r, fileID)
	if !ok {
		return
	}

	filePath := filepath.Join(s.config.UploadsDir, fileInfo.Id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		s.sendDownloadError(w, http.StatusNotFound, "not_found", "File not found on disk", "")
		return
	}

	// The checksum of a multi-gigabyte file takes a while to read through, so it
	// is never computed here - only reported when a previous request has already
	// produced it. The download itself starts the calculation.
	checksum, checksumReady := peekFileSHA256(fileInfo.Id)

	expireAt := int64(0)
	if !fileInfo.UnlimitedTime {
		expireAt = fileInfo.ExpireAt
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"file_id":             fileInfo.Id,
		"name":                fileInfo.Name,
		"size":                fileInfo.Size,
		"size_bytes":          fileInfo.SizeBytes,
		"content_type":        fileInfo.ContentType,
		"upload_date":         fileInfo.UploadDate,
		"upload_date_string":  time.Unix(fileInfo.UploadDate, 0).Format("2006-01-02 15:04"),
		"expire_at":           expireAt,
		"expire_at_string":    fileInfo.ExpireAtString,
		"unlimited_time":      fileInfo.UnlimitedTime,
		"downloads_remaining": fileInfo.DownloadsRemaining,
		"unlimited_downloads": fileInfo.UnlimitedDownloads,
		"sha256":              checksum,
		"sha256_ready":        checksumReady,
		"chunk_size":          defaultDownloadChunkSize,
		"max_chunk_size":      maxDownloadChunkSize,
		"direct_url":          "/d/" + fileInfo.Id,
	})
}

var (
	errChunkInvalid    = errors.New("invalid chunk request")
	errChunkOutOfRange = errors.New("requested range is outside the file")
)

// parseChunkRequest works out which byte interval to serve. A Range header
// takes precedence over the offset/size query parameters; if neither is given
// the first default-sized chunk is served. The returned interval is always
// inside the file.
func parseChunkRequest(query url.Values, rangeHeader string, fileSize int64) (offset int64, size int64, err error) {
	rangeHeader = strings.TrimSpace(rangeHeader)

	switch {
	case rangeHeader != "":
		offset, size, err = parseRangeHeader(rangeHeader, fileSize)
		if err != nil {
			return 0, 0, err
		}
	default:
		offset, err = parseChunkInt(query.Get("offset"), 0)
		if err != nil {
			return 0, 0, err
		}
		size, err = parseChunkInt(query.Get("size"), defaultDownloadChunkSize)
		if err != nil {
			return 0, 0, err
		}
	}

	if offset < 0 || size < 0 {
		return 0, 0, errChunkInvalid
	}
	if size > maxDownloadChunkSize {
		size = maxDownloadChunkSize
	}

	if fileSize == 0 {
		if offset != 0 {
			return 0, 0, errChunkOutOfRange
		}
		return 0, 0, nil
	}

	if offset >= fileSize {
		return 0, 0, errChunkOutOfRange
	}
	if size == 0 {
		return 0, 0, errChunkInvalid
	}
	if offset+size > fileSize {
		size = fileSize - offset
	}

	return offset, size, nil
}

func parseChunkInt(raw string, fallback int64) (int64, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errChunkInvalid
	}
	return value, nil
}

// parseRangeHeader understands the single-interval "bytes=start-end" form that
// a resuming download uses. Multi-range and suffix-only requests are rejected.
func parseRangeHeader(header string, fileSize int64) (int64, int64, error) {
	spec, ok := strings.CutPrefix(header, "bytes=")
	if !ok || strings.Contains(spec, ",") {
		return 0, 0, errChunkInvalid
	}

	startRaw, endRaw, ok := strings.Cut(spec, "-")
	if !ok {
		return 0, 0, errChunkInvalid
	}

	startRaw = strings.TrimSpace(startRaw)
	endRaw = strings.TrimSpace(endRaw)
	if startRaw == "" {
		return 0, 0, errChunkInvalid
	}

	start, err := strconv.ParseInt(startRaw, 10, 64)
	if err != nil || start < 0 {
		return 0, 0, errChunkInvalid
	}

	if endRaw == "" {
		if fileSize == 0 {
			return start, 0, nil
		}
		return start, fileSize - start, nil
	}

	end, err := strconv.ParseInt(endRaw, 10, 64)
	if err != nil || end < start {
		return 0, 0, errChunkInvalid
	}

	return start, end - start + 1, nil
}

// handleDownloadChunk serves one byte interval of the file.
func (s *Server) handleDownloadChunk(w http.ResponseWriter, r *http.Request, fileID string) {
	fileInfo, account, ok := s.authorizeChunkedDownload(w, r, fileID)
	if !ok {
		return
	}

	filePath := filepath.Join(s.config.UploadsDir, fileInfo.Id)
	file, err := os.Open(filePath)
	if err != nil {
		s.sendDownloadError(w, http.StatusNotFound, "not_found", "File not found on disk", "")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("Warning: Could not stat file %s: %v", fileInfo.Id, err)
		s.sendDownloadError(w, http.StatusInternalServerError, "stat_failed", "Could not read file", "")
		return
	}
	fileSize := stat.Size()

	offset, size, err := parseChunkRequest(r.URL.Query(), r.Header.Get("Range"), fileSize)
	if err != nil {
		if errors.Is(err, errChunkOutOfRange) {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			s.sendDownloadError(w, http.StatusRequestedRangeNotSatisfiable, "out_of_range", "Requested range is outside the file", "")
			return
		}
		s.sendDownloadError(w, http.StatusBadRequest, "invalid_range", "Invalid offset or size", "")
		return
	}

	// Keep the session alive while the chunk is on the wire, exactly as
	// performDownload does for a whole-file download.
	var sessionId string
	if cookie, err := r.Cookie("session"); err == nil {
		sessionId = cookie.Value
	} else if cookie, err := r.Cookie("download_session"); err == nil {
		sessionId = cookie.Value
	}
	if sessionId != "" {
		s.markTransferActive(sessionId)
		defer s.markTransferInactive(sessionId)
	}

	isFirstChunk := offset == 0
	isLastChunk := offset+size >= fileSize

	if isFirstChunk {
		log.Printf("📥 Chunked download started: %s (%s) by %s", fileInfo.Name, fileInfo.Size, getDownloaderInfo(account, r.RemoteAddr))
		// Hash in the background so the checksum is ready by the time the client
		// has finished downloading and asks for it.
		startFileSHA256(fileInfo.Id, filePath)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, offset+size-1, fileSize))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusPartialContent)

	written, err := io.Copy(w, io.NewSectionReader(file, offset, size))
	if err != nil {
		log.Printf("Chunked download interrupted: %s | offset %d | %v", fileInfo.Id, offset, err)
		return
	}
	if written != size {
		log.Printf("Chunked download short write: %s | offset %d | %d of %d bytes", fileInfo.Id, offset, written, size)
		return
	}

	// The download counter is spent when the last chunk has been delivered, not
	// when the first one is requested. Counting up front would push a file with
	// one remaining download over its limit halfway through its own transfer.
	if isLastChunk {
		s.finalizeChunkedDownload(r, fileInfo, account)
	}
}

// finalizeChunkedDownload records a completed chunked download with the same
// bookkeeping that performDownload does for a direct download.
func (s *Server) finalizeChunkedDownload(r *http.Request, fileInfo *database.FileInfo, account *models.DownloadAccount) {
	if err := database.DB.UpdateFileDownloadCount(fileInfo.Id); err != nil {
		log.Printf("Warning: Could not update download count: %v", err)
	}

	downloadLog := &models.DownloadLog{
		FileId:          fileInfo.Id,
		FileName:        fileInfo.Name,
		FileSize:        fileInfo.SizeBytes,
		DownloadedAt:    time.Now().Unix(),
		IpAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
		IsAuthenticated: account != nil,
	}

	if account != nil {
		downloadLog.DownloadAccountId = account.Id
		downloadLog.Email = account.Email
		database.DB.UpdateDownloadAccountLastUsed(account.Id)
	}

	if err := database.DB.CreateDownloadLog(downloadLog); err != nil {
		log.Printf("Warning: Could not create download log: %v", err)
	}

	// Send email notification to file owner
	go func() {
		owner, err := database.DB.GetUserByID(fileInfo.UserId)
		if err != nil {
			log.Printf("Could not get file owner for download notification: %v", err)
			return
		}

		clientIP := getClientIP(r)
		err = email.SendFileDownloadNotification(fileInfo, clientIP, s.getPublicURL(), owner.Email)
		if err != nil {
			log.Printf("Failed to send download notification email: %v", err)
		} else {
			log.Printf("Download notification email sent to %s", owner.Email)
		}
	}()

	var userID int64
	userEmail := "anonymous"
	if account != nil {
		userID = int64(account.Id)
		userEmail = account.Email
	}

	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     userID,
		UserEmail:  userEmail,
		Action:     "FILE_DOWNLOADED",
		EntityType: "File",
		EntityID:   fileInfo.Id,
		Details:    fmt.Sprintf("{\"file_name\":\"%s\",\"size\":%d,\"authenticated\":%v,\"chunked\":true}", fileInfo.Name, fileInfo.SizeBytes, account != nil),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	log.Printf("File download completed: %s (%s) by %s - chunked", fileInfo.Name, fileInfo.Size, getDownloaderInfo(account, r.RemoteAddr))
}

// handleDownloadVerify returns the server-side checksum for the client to
// compare its own against.
func (s *Server) handleDownloadVerify(w http.ResponseWriter, r *http.Request, fileID string) {
	fileInfo, _, ok := s.authorizeChunkedDownload(w, r, fileID)
	if !ok {
		return
	}

	filePath := filepath.Join(s.config.UploadsDir, fileInfo.Id)
	checksum, err := fileSHA256(fileInfo.Id, filePath)
	if err != nil {
		log.Printf("Warning: Could not calculate SHA-256 for %s: %v", fileInfo.Id, err)
		s.sendDownloadError(w, http.StatusInternalServerError, "checksum_failed", "Could not calculate checksum", "")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"file_id":    fileInfo.Id,
		"algorithm":  "sha256",
		"sha256":     checksum,
		"size_bytes": fileInfo.SizeBytes,
	})
}

// sendDownloadError answers a download API request with a JSON error. The
// fallback URL lets the browser fall back to the plain /d/ link.
func (s *Server) sendDownloadError(w http.ResponseWriter, status int, code, message, fallbackURL string) {
	payload := map[string]string{
		"error": message,
		"code":  code,
	}
	if fallbackURL != "" {
		payload["fallback_url"] = fallbackURL
	}
	s.sendJSON(w, status, payload)
}

// fileDigest is one entry in the checksum cache. Readers wait on ready and may
// then read value and err without further locking.
type fileDigest struct {
	value      string
	err        error
	ready      chan struct{}
	computedAt time.Time
}

var (
	digestCacheMu sync.Mutex
	digestCache   = make(map[string]*fileDigest)
)

// peekFileSHA256 reports the checksum only if it is already known.
func peekFileSHA256(fileID string) (string, bool) {
	digestCacheMu.Lock()
	entry, exists := digestCache[fileID]
	digestCacheMu.Unlock()

	if !exists {
		return "", false
	}

	select {
	case <-entry.ready:
		if entry.err != nil {
			return "", false
		}
		return entry.value, true
	default:
		return "", false
	}
}

// fileSHA256 returns the checksum of an uploaded file, reading through it at
// most once per file for as long as the server is running. Uploaded files are
// immutable, so a cached checksum never goes stale.
func fileSHA256(fileID, path string) (string, error) {
	digestCacheMu.Lock()
	entry, exists := digestCache[fileID]
	if !exists {
		entry = &fileDigest{ready: make(chan struct{}), computedAt: time.Now()}
		digestCache[fileID] = entry
		evictOldestDigests()
	}
	digestCacheMu.Unlock()

	if !exists {
		started := time.Now()
		entry.value, entry.err = calculateFileSHA256(path)
		close(entry.ready)

		if entry.err != nil {
			digestCacheMu.Lock()
			delete(digestCache, fileID)
			digestCacheMu.Unlock()
		} else {
			log.Printf("🔐 SHA-256 calculated for %s in %v", fileID, time.Since(started).Round(time.Millisecond))
		}
	}

	<-entry.ready
	return entry.value, entry.err
}

// startFileSHA256 warms the cache without blocking the caller.
func startFileSHA256(fileID, path string) {
	digestCacheMu.Lock()
	_, exists := digestCache[fileID]
	digestCacheMu.Unlock()
	if exists {
		return
	}

	go func() {
		if _, err := fileSHA256(fileID, path); err != nil {
			log.Printf("Warning: Could not calculate SHA-256 for %s: %v", fileID, err)
		}
	}()
}

// evictOldestDigests keeps the cache bounded. Callers must hold digestCacheMu.
func evictOldestDigests() {
	for len(digestCache) > sha256CacheLimit {
		var oldestID string
		var oldest time.Time
		for id, entry := range digestCache {
			if oldestID == "" || entry.computedAt.Before(oldest) {
				oldestID = id
				oldest = entry.computedAt
			}
		}
		delete(digestCache, oldestID)
	}
}

// calculateFileSHA256 calculates the SHA-256 hash of a file
func calculateFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
