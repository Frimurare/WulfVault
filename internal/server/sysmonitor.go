// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	sysMonitorLog     *os.File
	sysMonitorMutex   sync.Mutex
	sysMonitorLogPath string
	sysMonitorMaxSize int64 = 10 * 1024 * 1024 // 10MB max size
)

// InitSysMonitorLog initializes the system monitor log file
func InitSysMonitorLog(dataDir string) error {
	sysMonitorLogPath = filepath.Join(dataDir, "sysmonitor.log")

	file, err := os.OpenFile(sysMonitorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open sysmonitor log: %w", err)
	}

	sysMonitorLog = file
	log.Printf("📊 System monitor log initialized at %s (max size: 10MB)", sysMonitorLogPath)
	return nil
}

// CloseSysMonitorLog closes the sysmonitor log file
func CloseSysMonitorLog() {
	if sysMonitorLog != nil {
		sysMonitorLog.Close()
	}
}

// LogSysMonitor writes a log entry to the sysmonitor log with automatic rotation
func LogSysMonitor(format string, args ...interface{}) {
	sysMonitorMutex.Lock()
	defer sysMonitorMutex.Unlock()

	if sysMonitorLog == nil {
		return // Log not initialized
	}

	// Check file size and rotate if needed
	if fileInfo, err := sysMonitorLog.Stat(); err == nil {
		if fileInfo.Size() >= sysMonitorMaxSize {
			rotateSysMonitorLog()
		}
	}

	// Write log entry with timestamp
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	logEntry := fmt.Sprintf("%s %s\n", timestamp, fmt.Sprintf(format, args...))
	sysMonitorLog.WriteString(logEntry)
}

// rotateSysMonitorLog rotates the sysmonitor log file when it exceeds max size
func rotateSysMonitorLog() {
	if sysMonitorLog == nil {
		return
	}

	// Close current log
	sysMonitorLog.Close()

	// Rename old log to .old
	oldPath := sysMonitorLogPath + ".old"
	os.Remove(oldPath) // Remove old backup if exists
	os.Rename(sysMonitorLogPath, oldPath)

	// Open new log file
	file, err := os.OpenFile(sysMonitorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("⚠️  Failed to rotate sysmonitor log: %v", err)
		return
	}

	sysMonitorLog = file
	log.Printf("🔄 Sysmonitor log rotated (previous log saved as sysmonitor.log.old)")
}
