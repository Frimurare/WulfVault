// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SysMonitorLogEntry represents a parsed sysmonitor log entry
type SysMonitorLogEntry struct {
	Timestamp int64  `json:"timestamp"`
	RawLog    string `json:"raw_log"`
}

// handleAdminSysMonitorLogs renders the sysmonitor logs page
func (s *Server) handleAdminSysMonitorLogs(w http.ResponseWriter, r *http.Request) {
	s.renderAdminSysMonitorLogsPage(w)
}

// handleAPIGetSysMonitorLogs returns sysmonitor logs with filtering and pagination
func (s *Server) handleAPIGetSysMonitorLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	searchTerm := r.URL.Query().Get("search")

	// Pagination
	limit := 100 // Default to 100 for sysmonitor (more detailed)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Read and parse sysmonitor logs
	logPath := filepath.Join(s.config.DataDir, "sysmonitor.log")
	entries, totalCount, err := s.parseSysMonitorLogs(logPath, searchTerm, limit, offset)
	if err != nil {
		http.Error(w, "Error reading sysmonitor logs", http.StatusInternalServerError)
		return
	}

	// Get file info
	fileInfo, _ := os.Stat(logPath)
	var fileSize string
	var lastModified string
	if fileInfo != nil {
		fileSize = formatBytesHTTP(fileInfo.Size())
		lastModified = fileInfo.ModTime().Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success":true,"logs":%s,"total_count":%d,"offset":%d,"limit":%d,"file_size":"%s","last_modified":"%s"}`,
		s.serializeSysMonitorLogEntries(entries), totalCount, offset, limit, fileSize, lastModified)
}

// parseSysMonitorLogs reads and parses sysmonitor log file
func (s *Server) parseSysMonitorLogs(logPath string, searchTerm string, limit int, offset int) ([]SysMonitorLogEntry, int, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var allEntries []SysMonitorLogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Apply search filter
		if searchTerm != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(searchTerm)) {
			continue
		}

		// Parse timestamp (2025/12/10 09:31:45 format)
		parts := strings.SplitN(line, " ", 3)
		var timestamp int64
		if len(parts) >= 2 {
			timeStr := parts[0] + " " + parts[1]
			if t, err := time.Parse("2006/01/02 15:04:05", timeStr); err == nil {
				timestamp = t.Unix()
			}
		}

		entry := SysMonitorLogEntry{
			Timestamp: timestamp,
			RawLog:    line,
		}

		allEntries = append(allEntries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	// Reverse to show newest first
	for i, j := 0, len(allEntries)-1; i < j; i, j = i+1, j-1 {
		allEntries[i], allEntries[j] = allEntries[j], allEntries[i]
	}

	totalCount := len(allEntries)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	return allEntries[start:end], totalCount, nil
}

// serializeSysMonitorLogEntries converts log entries to JSON
func (s *Server) serializeSysMonitorLogEntries(entries []SysMonitorLogEntry) string {
	if len(entries) == 0 {
		return "[]"
	}

	var result strings.Builder
	result.WriteString("[")

	for i, entry := range entries {
		if i > 0 {
			result.WriteString(",")
		}

		// Escape raw log for JSON
		escapedLog := strings.ReplaceAll(entry.RawLog, "\\", "\\\\")
		escapedLog = strings.ReplaceAll(escapedLog, "\"", "\\\"")
		escapedLog = strings.ReplaceAll(escapedLog, "\n", "\\n")
		escapedLog = strings.ReplaceAll(escapedLog, "\r", "\\r")
		escapedLog = strings.ReplaceAll(escapedLog, "\t", "\\t")

		result.WriteString(fmt.Sprintf(`{"timestamp":%d,"raw_log":"%s"}`,
			entry.Timestamp, escapedLog))
	}

	result.WriteString("]")
	return result.String()
}

// renderAdminSysMonitorLogsPage renders the sysmonitor logs UI (simplified version of server logs)
func (s *Server) renderAdminSysMonitorLogsPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	companyName := s.config.CompanyName
	if companyName == "" {
		companyName = "WulfVault"
	}

	headerHTML := s.getAdminHeaderHTML("SysMonitor Logs")
	faviconHTML := s.getFaviconHTML()

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SysMonitor Logs - ` + companyName + `</title>
    ` + faviconHTML + `
</head>
<body>
` + headerHTML + `
`
	html += `
    <style>
        .log-viewer {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .log-entry {
            padding: 8px 12px;
            margin: 4px 0;
            background: #f8f9fa;
            border-left: 3px solid #dee2e6;
            font-family: 'Courier New', monospace;
            font-size: 12px;
            word-wrap: break-word;
            white-space: pre-wrap;
        }
        .search-box {
            width: 100%;
            padding: 10px;
            margin-bottom: 20px;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .log-info {
            color: #666;
            margin-bottom: 20px;
        }
    </style>

    <div class="container" style="margin-top: 30px;">
        <h2>SysMonitor Logs</h2>
        <p class="log-info">Detailed system monitoring logs including all chunk uploads (max 10MB, auto-rotated)</p>

        <input type="text" id="searchBox" class="search-box" placeholder="Search logs..." />

        <div id="logInfo" class="log-info"></div>

        <div id="logViewer" class="log-viewer">
            <p>Loading logs...</p>
        </div>

        <div style="margin-top: 20px; text-align: center;">
            <button onclick="loadMore()" id="loadMoreBtn">Load More</button>
        </div>
    </div>

    <script>
        let currentOffset = 0;
        let currentSearch = '';
        const limit = 100;

        function loadLogs(append = false) {
            if (!append) {
                currentOffset = 0;
                document.getElementById('logViewer').innerHTML = '<p>Loading...</p>';
            }

            const search = document.getElementById('searchBox').value;
            currentSearch = search;

            fetch('/api/v1/admin/sysmonitor-logs?limit=' + limit + '&offset=' + currentOffset + '&search=' + encodeURIComponent(search))
                .then(r => r.json())
                .then(data => {
                    const viewer = document.getElementById('logViewer');

                    if (!append) {
                        viewer.innerHTML = '';
                    }

                    if (data.logs.length === 0) {
                        viewer.innerHTML = '<p>No logs found</p>';
                        return;
                    }

                    data.logs.forEach(log => {
                        const entry = document.createElement('div');
                        entry.className = 'log-entry';
                        entry.textContent = log.raw_log;
                        viewer.appendChild(entry);
                    });

                    document.getElementById('logInfo').textContent =
                        'Showing ' + (currentOffset + 1) + '-' + (currentOffset + data.logs.length) +
                        ' of ' + data.total_count + ' entries | File size: ' + data.file_size;

                    currentOffset += data.logs.length;

                    // Hide load more button if no more logs
                    document.getElementById('loadMoreBtn').style.display =
                        currentOffset >= data.total_count ? 'none' : 'inline-block';
                });
        }

        function loadMore() {
            loadLogs(true);
        }

        // Search on typing
        let searchTimeout;
        document.getElementById('searchBox').addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => loadLogs(), 500);
        });

        // Initial load
        loadLogs();

        // Refresh every 5 seconds
        setInterval(() => {
            if (currentOffset <= limit) { // Only auto-refresh first page
                loadLogs(false);
            }
        }, 5000);
    </script>
</body>
</html>
`

	w.Write([]byte(html))
}
