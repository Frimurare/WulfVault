// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/Frimurare/WulfVault/internal/i18n"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ServerLogEntry represents a parsed server log entry
type ServerLogEntry struct {
	Timestamp    int64  `json:"timestamp"`
	Level        string `json:"level"` // success, warning, error, info
	StatusCode   int    `json:"status_code"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Duration     string `json:"duration"`
	RequestSize  string `json:"request_size"`
	ResponseSize string `json:"response_size"`
	IP           string `json:"ip"`
	UserAgent    string `json:"user_agent"`
	RawLog       string `json:"raw_log"`
}

// handleAdminServerLogs renders the server logs page
func (s *Server) handleAdminServerLogs(w http.ResponseWriter, r *http.Request) {
	s.renderAdminServerLogsPage(w, s.tr(r))
}

// handleAPIGetServerLogs returns server logs with filtering and pagination
func (s *Server) handleAPIGetServerLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	searchTerm := r.URL.Query().Get("search")
	levelFilter := r.URL.Query().Get("level")

	var startDate, endDate int64
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if sd, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			startDate = sd
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if ed, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			endDate = ed
		}
	}

	// Pagination
	limit := 50 // Default to 50
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

	// Read and parse server logs
	logPath := filepath.Join(s.config.DataDir, "server.log")
	entries, totalCount, err := s.parseServerLogs(logPath, searchTerm, levelFilter, startDate, endDate, limit, offset)
	if err != nil {
		http.Error(w, "Error reading server logs", http.StatusInternalServerError)
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
		s.serializeLogEntries(entries), totalCount, offset, limit, fileSize, lastModified)
}

// handleAPIExportServerLogs exports server logs to CSV
func (s *Server) handleAPIExportServerLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get filter parameters
	searchTerm := r.URL.Query().Get("search")
	levelFilter := r.URL.Query().Get("level")

	var startDate, endDate int64
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if sd, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			startDate = sd
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if ed, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			endDate = ed
		}
	}

	// Get all matching logs (no limit/offset for export)
	logPath := filepath.Join(s.config.DataDir, "server.log")
	entries, _, err := s.parseServerLogs(logPath, searchTerm, levelFilter, startDate, endDate, 0, 0)
	if err != nil {
		http.Error(w, "Error reading server logs", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	filename := fmt.Sprintf("server_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Create CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	csvWriter.Write([]string{
		"Timestamp",
		"Date/Time",
		"Level",
		"Status Code",
		"Method",
		"Path",
		"Duration",
		"Request Size",
		"Response Size",
		"IP Address",
		"User Agent",
		"Full Log",
	})

	// Write data rows
	for _, entry := range entries {
		timestamp := time.Unix(entry.Timestamp, 0).Format("2006-01-02 15:04:05")
		csvWriter.Write([]string{
			fmt.Sprintf("%d", entry.Timestamp),
			timestamp,
			entry.Level,
			fmt.Sprintf("%d", entry.StatusCode),
			entry.Method,
			entry.Path,
			entry.Duration,
			entry.RequestSize,
			entry.ResponseSize,
			entry.IP,
			entry.UserAgent,
			entry.RawLog,
		})
	}
}

// parseServerLogs reads and parses server log file with filtering
func (s *Server) parseServerLogs(filePath, searchTerm, levelFilter string, startDate, endDate int64, limit, offset int) ([]ServerLogEntry, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var allEntries []ServerLogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entry := s.parseLogLine(line)

		// Skip non-HTTP log entries UNLESS they are important upload/system events
		// Include: HTTP requests, upload events (📤 📦 ✅ UPLOAD), and system events
		isHTTPLog := entry.StatusCode != 0 && entry.Method != ""
		isUploadLog := strings.Contains(line, "UPLOAD STARTED") ||
			strings.Contains(line, "UPLOAD COMPLETED") ||
			strings.Contains(line, "UPLOAD ABANDONED") ||
			strings.Contains(line, "Upload progress:")

		if !isHTTPLog && !isUploadLog {
			continue
		}

		// Apply filters
		if searchTerm != "" && !strings.Contains(strings.ToLower(entry.RawLog), strings.ToLower(searchTerm)) {
			continue
		}

		if levelFilter != "" && entry.Level != levelFilter {
			continue
		}

		if startDate > 0 && entry.Timestamp < startDate {
			continue
		}

		if endDate > 0 && entry.Timestamp > endDate {
			continue
		}

		allEntries = append(allEntries, entry)
	}

	totalCount := len(allEntries)

	// Apply pagination if limit > 0
	if limit > 0 {
		start := offset
		if start > totalCount {
			start = totalCount
		}

		end := start + limit
		if end > totalCount {
			end = totalCount
		}

		// Return most recent entries first
		reversedEntries := make([]ServerLogEntry, len(allEntries))
		for i, entry := range allEntries {
			reversedEntries[len(allEntries)-1-i] = entry
		}

		return reversedEntries[start:end], totalCount, scanner.Err()
	}

	// Return all entries (for export)
	reversedEntries := make([]ServerLogEntry, len(allEntries))
	for i, entry := range allEntries {
		reversedEntries[len(allEntries)-1-i] = entry
	}

	return reversedEntries, totalCount, scanner.Err()
}

// parseLogLine parses a single log line into ServerLogEntry
func (s *Server) parseLogLine(line string) ServerLogEntry {
	entry := ServerLogEntry{
		RawLog: line,
		Level:  "info",
	}

	// Parse timestamp (format: 2025/12/09 11:38:40)
	if len(line) > 19 {
		timeStr := line[0:19]
		if t, err := time.Parse("2006/01/02 15:04:05", timeStr); err == nil {
			entry.Timestamp = t.Unix()
		}
	}

	// Detect level from emoji and keywords
	if strings.Contains(line, "✅") || strings.Contains(line, "UPLOAD COMPLETED") {
		entry.Level = "success"
	} else if strings.Contains(line, "⚠️") || strings.Contains(line, "UPLOAD ABANDONED") {
		entry.Level = "warning"
	} else if strings.Contains(line, "❌") {
		entry.Level = "error"
	} else if strings.Contains(line, "📝") || strings.Contains(line, "🚀") ||
		strings.Contains(line, "📤") || strings.Contains(line, "UPLOAD STARTED") {
		entry.Level = "info"
	} else if strings.Contains(line, "📦") || strings.Contains(line, "Upload progress:") {
		entry.Level = "info"
	}

	// Parse HTTP log format: ✅ [200] GET /path | Duration: 1ms | Req: 0 B | Res: 1 KB | IP: x.x.x.x | UA: ...
	if strings.Contains(line, "[") && strings.Contains(line, "]") {
		// Extract status code
		startIdx := strings.Index(line, "[")
		endIdx := strings.Index(line, "]")
		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			if code, err := strconv.Atoi(line[startIdx+1 : endIdx]); err == nil {
				entry.StatusCode = code
			}
		}

		// Extract method and path
		parts := strings.Split(line, "|")
		if len(parts) > 0 {
			firstPart := strings.TrimSpace(parts[0])
			// Extract method and path from: ✅ [200] GET /path
			methodPathParts := strings.Fields(firstPart)
			if len(methodPathParts) >= 3 {
				entry.Method = methodPathParts[len(methodPathParts)-2]
				entry.Path = methodPathParts[len(methodPathParts)-1]
			}
		}

		// Extract other fields
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "Duration:") {
				entry.Duration = strings.TrimPrefix(part, "Duration: ")
			} else if strings.HasPrefix(part, "Req:") {
				entry.RequestSize = strings.TrimPrefix(part, "Req: ")
			} else if strings.HasPrefix(part, "Res:") {
				entry.ResponseSize = strings.TrimPrefix(part, "Res: ")
			} else if strings.HasPrefix(part, "IP:") {
				entry.IP = strings.TrimPrefix(part, "IP: ")
			} else if strings.HasPrefix(part, "UA:") {
				entry.UserAgent = strings.TrimPrefix(part, "UA: ")
			}
		}
	}

	return entry
}

// serializeLogEntries converts log entries to JSON manually (for better control)
func (s *Server) serializeLogEntries(entries []ServerLogEntry) string {
	if len(entries) == 0 {
		return "[]"
	}

	var result strings.Builder
	result.WriteString("[")
	for i, entry := range entries {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(fmt.Sprintf(`{"timestamp":%d,"level":"%s","status_code":%d,"method":"%s","path":"%s","duration":"%s","request_size":"%s","response_size":"%s","ip":"%s","user_agent":"%s","raw_log":%s}`,
			entry.Timestamp,
			entry.Level,
			entry.StatusCode,
			s.escapeJSON(entry.Method),
			s.escapeJSON(entry.Path),
			s.escapeJSON(entry.Duration),
			s.escapeJSON(entry.RequestSize),
			s.escapeJSON(entry.ResponseSize),
			s.escapeJSON(entry.IP),
			s.escapeJSON(entry.UserAgent),
			fmt.Sprintf(`"%s"`, s.escapeJSON(entry.RawLog)),
		))
	}
	result.WriteString("]")
	return result.String()
}

// escapeJSON escapes special characters for JSON
func (s *Server) escapeJSON(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	str = strings.ReplaceAll(str, "\n", "\\n")
	str = strings.ReplaceAll(str, "\r", "\\r")
	str = strings.ReplaceAll(str, "\t", "\\t")
	return str
}

// renderAdminServerLogsPage renders the server logs page with inline HTML
func (s *Server) renderAdminServerLogsPage(w http.ResponseWriter, tr *i18n.Translator) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	companyName := s.config.CompanyName
	if companyName == "" {
		companyName = "WulfVault"
	}

	headerHTML := s.getAdminHeaderHTML("Server Logs", tr)
	faviconHTML := s.getFaviconHTML()

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Logs - ` + companyName + `</title>
    ` + faviconHTML + `
</head>
<body>
` + headerHTML + `
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
            font-family: 'Courier New', monospace;
            font-size: 12px;
            word-wrap: break-word;
            white-space: pre-wrap;
        }
        .log-entry.success { border-left: 3px solid #28a745; }
        .log-entry.warning { border-left: 3px solid #ffc107; }
        .log-entry.error   { border-left: 3px solid #dc3545; }
        .log-entry.info    { border-left: 3px solid #17a2b8; }
        .filters {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            flex-wrap: wrap;
            align-items: center;
        }
        .filters input, .filters select {
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .filters input[type="text"] { flex: 1; min-width: 200px; }
        .filters select { min-width: 120px; }
        .log-info {
            color: #666;
            margin-bottom: 15px;
            font-size: 13px;
        }
        .load-more-btn {
            display: block;
            margin: 20px auto;
            padding: 10px 30px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .load-more-btn:hover { background: #0056b3; }
        .status-badge {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            font-weight: bold;
            margin-right: 6px;
        }
        .status-badge.s2xx { background: #d4edda; color: #155724; }
        .status-badge.s3xx { background: #cce5ff; color: #004085; }
        .status-badge.s4xx { background: #fff3cd; color: #856404; }
        .status-badge.s5xx { background: #f8d7da; color: #721c24; }
    </style>

    <div class="container" style="margin-top: 30px;">
        <h2>Server Logs</h2>
        <p class="log-info">HTTP request logs with status codes, durations and upload events</p>

        <div class="filters">
            <input type="text" id="searchBox" placeholder="Search logs..." />
            <select id="levelFilter">
                <option value="">All levels</option>
                <option value="success">Success</option>
                <option value="info">Info</option>
                <option value="warning">Warning</option>
                <option value="error">Error</option>
            </select>
        </div>

        <div id="logInfo" class="log-info"></div>

        <div id="logViewer" class="log-viewer">
            <p>Loading logs...</p>
        </div>

        <button onclick="loadMore()" id="loadMoreBtn" class="load-more-btn">Load More</button>
    </div>

    <script>
        let currentOffset = 0;
        const limit = 50;

        function loadLogs(append) {
            if (!append) {
                currentOffset = 0;
                document.getElementById('logViewer').innerHTML = '<p>Loading...</p>';
            }

            const search = document.getElementById('searchBox').value;
            const level = document.getElementById('levelFilter').value;

            let url = '/api/v1/admin/server-logs?limit=' + limit + '&offset=' + currentOffset;
            if (search) url += '&search=' + encodeURIComponent(search);
            if (level) url += '&level=' + encodeURIComponent(level);

            fetch(url)
                .then(function(r) { return r.json(); })
                .then(function(data) {
                    const viewer = document.getElementById('logViewer');

                    if (!append) viewer.innerHTML = '';

                    if (!data.logs || data.logs.length === 0) {
                        if (!append) viewer.innerHTML = '<p>No logs found</p>';
                        document.getElementById('loadMoreBtn').style.display = 'none';
                        return;
                    }

                    data.logs.forEach(function(log) {
                        const entry = document.createElement('div');
                        entry.className = 'log-entry ' + log.level;

                        // Add status badge if we have a status code
                        let html = '';
                        if (log.status_code > 0) {
                            let badgeClass = 's2xx';
                            if (log.status_code >= 300) badgeClass = 's3xx';
                            if (log.status_code >= 400) badgeClass = 's4xx';
                            if (log.status_code >= 500) badgeClass = 's5xx';
                            html += '<span class="status-badge ' + badgeClass + '">' + log.status_code + '</span>';
                        }
                        html += log.raw_log;
                        entry.innerHTML = html;
                        viewer.appendChild(entry);
                    });

                    document.getElementById('logInfo').textContent =
                        'Showing ' + (currentOffset + 1) + '-' + (currentOffset + data.logs.length) +
                        ' of ' + data.total_count + ' entries | File size: ' + data.file_size;

                    currentOffset += data.logs.length;

                    document.getElementById('loadMoreBtn').style.display =
                        currentOffset >= data.total_count ? 'none' : 'block';
                });
        }

        function loadMore() { loadLogs(true); }

        var searchTimeout;
        document.getElementById('searchBox').addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(function() { loadLogs(false); }, 500);
        });
        document.getElementById('levelFilter').addEventListener('change', function() {
            loadLogs(false);
        });

        loadLogs(false);

        setInterval(function() {
            if (currentOffset <= limit) loadLogs(false);
        }, 5000);
    </script>
</body>
</html>` + "\n"

	w.Write([]byte(html))
}
