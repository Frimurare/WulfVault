// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	serverLogFile   *os.File
	serverLogMutex  sync.Mutex
	serverLogPath   string
	serverLogMaxMB  int64 = 50 // Default 50MB
)

// InitServerLog initializes the server log file with rotation support
func InitServerLog(dataDir string, maxSizeMB int) error {
	serverLogMutex.Lock()
	defer serverLogMutex.Unlock()

	serverLogPath = filepath.Join(dataDir, "server.log")
	serverLogMaxMB = int64(maxSizeMB)

	// Open log file in append mode
	f, err := os.OpenFile(serverLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open server log file: %w", err)
	}

	serverLogFile = f

	// Set up dual output: both file and stdout
	log.SetOutput(io.MultiWriter(os.Stdout, serverLogFile))

	return nil
}

// CloseServerLog closes the server log file
func CloseServerLog() {
	serverLogMutex.Lock()
	defer serverLogMutex.Unlock()

	if serverLogFile != nil {
		serverLogFile.Close()
		serverLogFile = nil
	}
}

// rotateServerLogIfNeeded checks and rotates the log file if it exceeds max size
func rotateServerLogIfNeeded() error {
	serverLogMutex.Lock()
	defer serverLogMutex.Unlock()

	if serverLogFile == nil || serverLogPath == "" {
		return nil
	}

	// Check file size
	info, err := serverLogFile.Stat()
	if err != nil {
		return err
	}

	maxBytes := serverLogMaxMB * 1024 * 1024
	if info.Size() < maxBytes {
		return nil // No rotation needed
	}

	// Close current file
	serverLogFile.Close()

	// Rename current log to .old
	oldPath := serverLogPath + ".old"
	os.Remove(oldPath) // Remove old backup if it exists
	if err := os.Rename(serverLogPath, oldPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Create new log file
	f, err := os.OpenFile(serverLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	serverLogFile = f
	log.SetOutput(io.MultiWriter(os.Stdout, serverLogFile))

	log.Printf("📝 Server log rotated (old log saved to %s)", oldPath)

	return nil
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// loggingMiddleware logs all HTTP requests with detailed information
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Check if log rotation is needed (check every 100 requests approximately)
		// This is a lightweight check that happens occasionally
		if time.Now().Unix()%100 == 0 {
			rotateServerLogIfNeeded()
		}

		// Wrap the response writer to capture status and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
			bytes:          0,
		}

		// Get request size
		contentLength := r.ContentLength
		if contentLength < 0 {
			contentLength = 0
		}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Format the log message
		statusCode := wrapped.statusCode
		if statusCode == 0 {
			statusCode = 200 // Default if WriteHeader was never called
		}

		// Format sizes in human-readable form
		requestSize := formatBytesHTTP(contentLength)
		responseSize := formatBytesHTTP(int64(wrapped.bytes))

		// Determine log level based on status code
		statusEmoji := "✅"
		if statusCode >= 400 && statusCode < 500 {
			statusEmoji = "⚠️"
		} else if statusCode >= 500 {
			statusEmoji = "❌"
		}

		// Log format: [STATUS] METHOD PATH | Duration: Xms | Req: Xbytes | Res: Xbytes | IP: x.x.x.x
		log.Printf("%s [%d] %s %s | Duration: %v | Req: %s | Res: %s | IP: %s | UA: %s",
			statusEmoji,
			statusCode,
			r.Method,
			r.URL.Path,
			duration.Round(time.Millisecond),
			requestSize,
			responseSize,
			getClientIPFromRequest(r),
			getUserAgentFromRequest(r))
	})
}

// formatBytesHTTP converts bytes to human-readable format
func formatBytesHTTP(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getClientIPFromRequest extracts the real client IP from the request
func getClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getUserAgentFromRequest extracts the User-Agent from the request
func getUserAgentFromRequest(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	if len(ua) > 100 {
		return ua[:97] + "..."
	}
	return ua
}
