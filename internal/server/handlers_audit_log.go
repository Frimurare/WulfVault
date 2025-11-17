// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)

package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
)

// handleAdminAuditLogs displays the audit log admin page
func (s *Server) handleAdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	s.renderAdminAuditLogsPage(w)
}

// handleAPIGetAuditLogs returns audit logs with filtering and pagination
func (s *Server) handleAPIGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	filter := &database.AuditLogFilter{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filter.UserID = userID
		}
	}

	filter.Action = r.URL.Query().Get("action")
	filter.EntityType = r.URL.Query().Get("entity_type")
	filter.SearchTerm = r.URL.Query().Get("search")

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			filter.EndDate = endDate
		}
	}

	// Pagination
	limit := 200
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}
	filter.Limit = limit

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	filter.Offset = offset

	// Get logs
	logs, err := database.DB.GetAuditLogs(filter)
	if err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		http.Error(w, "Error fetching audit logs", http.StatusInternalServerError)
		return
	}

	// Get total count
	totalCount, err := database.DB.GetAuditLogCount(filter)
	if err != nil {
		log.Printf("Error getting audit log count: %v", err)
		totalCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"logs":        logs,
		"total_count": totalCount,
		"offset":      offset,
		"limit":       limit,
	})
}

// handleAPIExportAuditLogs exports audit logs to CSV
func (s *Server) handleAPIExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all logs (with same filtering as API)
	filter := &database.AuditLogFilter{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filter.UserID = userID
		}
	}

	filter.Action = r.URL.Query().Get("action")
	filter.EntityType = r.URL.Query().Get("entity_type")
	filter.SearchTerm = r.URL.Query().Get("search")

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			filter.EndDate = endDate
		}
	}

	// No limit for export - get all matching logs
	filter.Limit = 0

	logs, err := database.DB.GetAuditLogs(filter)
	if err != nil {
		log.Printf("Error fetching audit logs for export: %v", err)
		http.Error(w, "Error fetching audit logs", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Create CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	csvWriter.Write([]string{
		"ID",
		"Timestamp",
		"Date/Time",
		"User ID",
		"User Email",
		"Action",
		"Entity Type",
		"Entity ID",
		"Details",
		"IP Address",
		"User Agent",
		"Success",
		"Error Message",
	})

	// Write data rows
	for _, entry := range logs {
		timestamp := time.Unix(entry.Timestamp, 0).Format("2006-01-02 15:04:05")
		successStr := "Yes"
		if !entry.Success {
			successStr = "No"
		}

		csvWriter.Write([]string{
			fmt.Sprintf("%d", entry.ID),
			fmt.Sprintf("%d", entry.Timestamp),
			timestamp,
			fmt.Sprintf("%d", entry.UserID),
			entry.UserEmail,
			entry.Action,
			entry.EntityType,
			entry.EntityID,
			entry.Details,
			entry.IPAddress,
			entry.UserAgent,
			successStr,
			entry.ErrorMsg,
		})
	}
}

// renderAdminAuditLogsPage renders the audit logs admin page
func (s *Server) renderAdminAuditLogsPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	companyName := s.config.CompanyName
	if companyName == "" {
		companyName = "WulfVault"
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Audit Logs - ` + companyName + `</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            background: white;
            padding: 20px 30px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .header h1 {
            color: #333;
            font-size: 24px;
        }

        .nav-links a {
            color: #667eea;
            text-decoration: none;
            margin-left: 20px;
            font-weight: 500;
        }

        .nav-links a:hover {
            text-decoration: underline;
        }

        .filters-card {
            background: white;
            padding: 25px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .filters-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 15px;
        }

        .filter-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #333;
            font-size: 14px;
        }

        .filter-group input,
        .filter-group select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
        }

        .filter-buttons {
            display: flex;
            gap: 10px;
            margin-top: 15px;
        }

        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 600;
            font-size: 14px;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .btn-secondary {
            background: #6c757d;
            color: white;
        }

        .btn-success {
            background: #28a745;
            color: white;
        }

        .btn:hover {
            opacity: 0.9;
            transform: translateY(-2px);
            transition: all 0.3s ease;
        }

        .logs-card {
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }

        .logs-header {
            padding: 20px 25px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .logs-stats {
            font-size: 14px;
            color: #666;
        }

        .table-container {
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        thead {
            background: #f8f9fa;
        }

        th {
            padding: 12px;
            text-align: left;
            font-weight: 600;
            color: #333;
            font-size: 13px;
            border-bottom: 2px solid #e0e0e0;
        }

        td {
            padding: 12px;
            border-bottom: 1px solid #f0f0f0;
            font-size: 13px;
        }

        tbody tr:hover {
            background-color: #f8f9fa;
        }

        .badge {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 600;
            display: inline-block;
        }

        .badge-success {
            background: #d4edda;
            color: #155724;
        }

        .badge-danger {
            background: #f8d7da;
            color: #721c24;
        }

        .badge-info {
            background: #d1ecf1;
            color: #0c5460;
        }

        .badge-warning {
            background: #fff3cd;
            color: #856404;
        }

        .pagination {
            padding: 20px 25px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-top: 1px solid #e0e0e0;
        }

        .pagination-info {
            color: #666;
            font-size: 14px;
        }

        .pagination-buttons {
            display: flex;
            gap: 10px;
        }

        .details-cell {
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            cursor: pointer;
            color: #667eea;
        }

        .details-cell:hover {
            text-decoration: underline;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }

        .no-data {
            text-align: center;
            padding: 40px;
            color: #999;
        }

        @media (max-width: 768px) {
            .header {
                flex-direction: column;
                align-items: flex-start;
            }

            .nav-links {
                margin-top: 15px;
            }

            .filters-grid {
                grid-template-columns: 1fr;
            }

            .table-container {
                font-size: 12px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 Audit Logs</h1>
            <div class="nav-links">
                <a href="/admin">Dashboard</a>
                <a href="/admin/settings">Settings</a>
                <a href="/logout">Logout</a>
            </div>
        </div>

        <div class="filters-card">
            <h3 style="margin-bottom: 15px; color: #333;">Filters</h3>
            <div class="filters-grid">
                <div class="filter-group">
                    <label for="action">Action</label>
                    <select id="action">
                        <option value="">All Actions</option>
                        <option value="LOGIN_SUCCESS">Login Success</option>
                        <option value="LOGIN_FAILED">Login Failed</option>
                        <option value="LOGOUT">Logout</option>
                        <option value="FILE_UPLOADED">File Uploaded</option>
                        <option value="FILE_DOWNLOADED">File Downloaded</option>
                        <option value="FILE_DELETED">File Deleted</option>
                        <option value="FILE_RESTORED">File Restored</option>
                        <option value="FILE_PERMANENTLY_DELETED">File Permanently Deleted</option>
                        <option value="USER_CREATED">User Created</option>
                        <option value="USER_UPDATED">User Updated</option>
                        <option value="USER_DELETED">User Deleted</option>
                        <option value="USER_ACTIVATED">User Activated</option>
                        <option value="USER_DEACTIVATED">User Deactivated</option>
                        <option value="TEAM_CREATED">Team Created</option>
                        <option value="TEAM_MEMBER_ADDED">Team Member Added</option>
                        <option value="TEAM_MEMBER_REMOVED">Team Member Removed</option>
                        <option value="PASSWORD_CHANGED">Password Changed</option>
                        <option value="2FA_ENABLED">2FA Enabled</option>
                        <option value="2FA_DISABLED">2FA Disabled</option>
                        <option value="SETTINGS_UPDATED">Settings Updated</option>
                    </select>
                </div>
                <div class="filter-group">
                    <label for="entity_type">Entity Type</label>
                    <select id="entity_type">
                        <option value="">All Types</option>
                        <option value="User">User</option>
                        <option value="File">File</option>
                        <option value="Team">Team</option>
                        <option value="Session">Session</option>
                        <option value="Settings">Settings</option>
                        <option value="System">System</option>
                    </select>
                </div>
                <div class="filter-group">
                    <label for="start_date">Start Date</label>
                    <input type="date" id="start_date">
                </div>
                <div class="filter-group">
                    <label for="end_date">End Date</label>
                    <input type="date" id="end_date">
                </div>
                <div class="filter-group">
                    <label for="items_per_page">Items Per Page</label>
                    <select id="items_per_page" onchange="updateLimit()">
                        <option value="20" selected>20</option>
                        <option value="50">50</option>
                        <option value="100">100</option>
                        <option value="200">200</option>
                    </select>
                </div>
                <div class="filter-group">
                    <label for="search">Search (Email/Details)</label>
                    <input type="text" id="search" placeholder="Search...">
                </div>
            </div>
            <div class="filter-buttons">
                <button class="btn btn-primary" onclick="applyFilters()">Apply Filters</button>
                <button class="btn btn-secondary" onclick="resetFilters()">Reset</button>
                <button class="btn btn-success" onclick="exportCSV()">📥 Export CSV</button>
            </div>
        </div>

        <div class="logs-card">
            <div class="logs-header">
                <h3 style="color: #333;">Audit Log Entries</h3>
                <div class="logs-stats" id="logs-stats">Loading...</div>
            </div>

            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Timestamp</th>
                            <th>User</th>
                            <th>Action</th>
                            <th>Entity</th>
                            <th>Details</th>
                            <th>IP</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody id="logs-tbody">
                        <tr>
                            <td colspan="8" class="loading">Loading audit logs...</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <div class="pagination">
                <div class="pagination-info" id="pagination-info"></div>
                <div class="pagination-buttons">
                    <button class="btn btn-secondary" id="prev-btn" onclick="prevPage()" disabled>Previous</button>
                    <button class="btn btn-secondary" id="next-btn" onclick="nextPage()" disabled>Next</button>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentOffset = 0;
        let limit = 20;
        let totalCount = 0;

        function formatTimestamp(timestamp) {
            const date = new Date(timestamp * 1000);
            return date.toLocaleString();
        }

        function getActionBadgeClass(action) {
            if (action.includes('LOGIN_SUCCESS') || action.includes('CREATED') || action.includes('ENABLED') || action.includes('ACTIVATED')) {
                return 'badge-success';
            } else if (action.includes('FAILED') || action.includes('DELETED') || action.includes('DISABLED') || action.includes('DEACTIVATED')) {
                return 'badge-danger';
            } else if (action.includes('UPDATED') || action.includes('CHANGED')) {
                return 'badge-warning';
            }
            return 'badge-info';
        }

        function showDetails(details) {
            alert(details);
        }

        async function loadLogs() {
            const params = new URLSearchParams();
            params.append('offset', currentOffset);
            params.append('limit', limit);

            const action = document.getElementById('action').value;
            const entityType = document.getElementById('entity_type').value;
            const search = document.getElementById('search').value;
            const startDate = document.getElementById('start_date').value;
            const endDate = document.getElementById('end_date').value;

            if (action) params.append('action', action);
            if (entityType) params.append('entity_type', entityType);
            if (search) params.append('search', search);
            if (startDate) params.append('start_date', new Date(startDate).getTime() / 1000);
            if (endDate) params.append('end_date', new Date(endDate + ' 23:59:59').getTime() / 1000);

            try {
                const response = await fetch('/api/v1/admin/audit-logs?' + params.toString());
                const data = await response.json();

                if (data.success) {
                    totalCount = data.total_count;
                    renderLogs(data.logs);
                    updatePagination();
                } else {
                    document.getElementById('logs-tbody').innerHTML = '<tr><td colspan="8" class="no-data">Error loading logs</td></tr>';
                }
            } catch (error) {
                console.error('Error loading logs:', error);
                document.getElementById('logs-tbody').innerHTML = '<tr><td colspan="8" class="no-data">Error loading logs</td></tr>';
            }
        }

        function renderLogs(logs) {
            const tbody = document.getElementById('logs-tbody');

            if (!logs || logs.length === 0) {
                tbody.innerHTML = '<tr><td colspan="8" class="no-data">No audit logs found</td></tr>';
                document.getElementById('logs-stats').textContent = 'No entries';
                return;
            }

            let html = '';
            logs.forEach(log => {
                const badgeClass = getActionBadgeClass(log.action);
                const statusBadge = log.success
                    ? '<span class="badge badge-success">Success</span>'
                    : '<span class="badge badge-danger">Failed</span>';

                html += '<tr>' +
                    '<td>' + log.id + '</td>' +
                    '<td>' + formatTimestamp(log.timestamp) + '</td>' +
                    '<td>' + log.user_email + '</td>' +
                    '<td><span class="badge ' + badgeClass + '">' + log.action + '</span></td>' +
                    '<td>' + log.entity_type + (log.entity_id ? ' #' + log.entity_id : '') + '</td>' +
                    '<td class="details-cell" onclick="showDetails(\'' + log.details.replace(/'/g, "\\'") + '\')">' + log.details + '</td>' +
                    '<td>' + log.ip_address + '</td>' +
                    '<td>' + statusBadge + '</td>' +
                    '</tr>';
            });

            tbody.innerHTML = html;
            document.getElementById('logs-stats').textContent = 'Showing ' + (currentOffset + 1) + '-' + Math.min(currentOffset + limit, totalCount) + ' of ' + totalCount + ' entries';
        }

        function updatePagination() {
            document.getElementById('prev-btn').disabled = currentOffset === 0;
            document.getElementById('next-btn').disabled = currentOffset + limit >= totalCount;
            document.getElementById('pagination-info').textContent = 'Page ' + (Math.floor(currentOffset / limit) + 1) + ' of ' + Math.ceil(totalCount / limit);
        }

        function updateLimit() {
            limit = parseInt(document.getElementById('items_per_page').value);
            currentOffset = 0;
            loadLogs();
        }

        function prevPage() {
            if (currentOffset > 0) {
                currentOffset -= limit;
                loadLogs();
            }
        }

        function nextPage() {
            if (currentOffset + limit < totalCount) {
                currentOffset += limit;
                loadLogs();
            }
        }

        function applyFilters() {
            currentOffset = 0;
            loadLogs();
        }

        function resetFilters() {
            document.getElementById('action').value = '';
            document.getElementById('entity_type').value = '';
            document.getElementById('search').value = '';
            document.getElementById('start_date').value = '';
            document.getElementById('end_date').value = '';
            currentOffset = 0;
            loadLogs();
        }

        function exportCSV() {
            const params = new URLSearchParams();
            const action = document.getElementById('action').value;
            const entityType = document.getElementById('entity_type').value;
            const search = document.getElementById('search').value;
            const startDate = document.getElementById('start_date').value;
            const endDate = document.getElementById('end_date').value;

            if (action) params.append('action', action);
            if (entityType) params.append('entity_type', entityType);
            if (search) params.append('search', search);
            if (startDate) params.append('start_date', new Date(startDate).getTime() / 1000);
            if (endDate) params.append('end_date', new Date(endDate + ' 23:59:59').getTime() / 1000);

            window.location.href = '/api/v1/admin/audit-logs/export?' + params.toString();
        }

        // Load logs on page load
        loadLogs();
    </script>
</body>
</html>`

	fmt.Fprint(w, html)
}
