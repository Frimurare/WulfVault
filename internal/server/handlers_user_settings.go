package server

import (
	"net/http"
	"strconv"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

// handleUserSettings displays user settings including 2FA
func (s *Server) handleUserSettings(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get remaining backup codes count
	backupCodesCount := 0
	if user.TOTPEnabled {
		backupCodesCount, _ = database.DB.GetRemainingBackupCodesCount(user.Id)
	}

	s.renderUserSettingsPage(w, user, backupCodesCount)
}

// renderUserSettingsPage renders the user settings page
func (s *Server) renderUserSettingsPage(w http.ResponseWriter, user *models.User, backupCodesCount int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	totpStatusBadge := ""
	totpActionButton := ""

	if user.TOTPEnabled {
		totpStatusBadge = `<span style="background: #4CAF50; color: white; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600;">ENABLED</span>`
		totpActionButton = `
			<button onclick="disable2FA()" style="background: #f44336; color: white; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600;">
				Disable 2FA
			</button>
			<button onclick="regenerateBackupCodes()" style="background: #2196F3; color: white; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600; margin-left: 10px;">
				Regenerate Backup Codes (` + strconv.Itoa(backupCodesCount) + ` remaining)
			</button>`
	} else {
		totpStatusBadge = `<span style="background: #999; color: white; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600;">DISABLED</span>`
		totpActionButton = `
			<button onclick="enable2FA()" style="background: ` + s.getPrimaryColor() + `; color: white; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600;">
				Enable 2FA
			</button>`
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Settings - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            color: white;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .container {
            max-width: 1200px;
            margin: 30px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .card h2 {
            margin-bottom: 20px;
            color: #333;
        }
        .setting-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 20px;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            margin-bottom: 15px;
        }
        .setting-info h3 {
            margin-bottom: 8px;
            color: #333;
        }
        .setting-info p {
            color: #666;
            font-size: 14px;
        }
        nav {
            display: flex;
            gap: 20px;
            margin-top: 15px;
        }
        nav a {
            color: white;
            text-decoration: none;
            opacity: 0.9;
            transition: opacity 0.3s;
        }
        nav a:hover {
            opacity: 1;
        }
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 1000;
            align-items: center;
            justify-content: center;
        }
        .modal-content {
            background: white;
            padding: 30px;
            border-radius: 12px;
            max-width: 500px;
            width: 90%;
            max-height: 90vh;
            overflow-y: auto;
        }
        .modal-content h3 {
            margin-bottom: 20px;
        }
        .qr-code {
            text-align: center;
            margin: 20px 0;
        }
        .qr-code img {
            max-width: 256px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
        }
        .backup-codes {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 8px;
            margin: 15px 0;
            font-family: monospace;
            font-size: 14px;
        }
        .backup-code {
            margin: 5px 0;
            padding: 8px;
            background: white;
            border-radius: 4px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }
        .form-group input {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 16px;
        }
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: opacity 0.3s;
        }
        .btn-primary {
            background: ` + s.getPrimaryColor() + `;
            color: white;
        }
        .btn-secondary {
            background: #999;
            color: white;
        }
        .btn:hover {
            opacity: 0.9;
        }
        .close-btn {
            float: right;
            font-size: 24px;
            cursor: pointer;
            color: #999;
        }
        .close-btn:hover {
            color: #333;
        }
        .alert {
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 15px;
        }
        .alert-success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .alert-error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .secret-text {
            font-family: monospace;
            background: #f5f5f5;
            padding: 10px;
            border-radius: 6px;
            word-break: break-all;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <header>
        <h1>` + s.config.CompanyName + `</h1>
        <nav>
            <a href="` + (func() string {
		if user.IsAdmin() {
			return "/admin"
		}
		return "/dashboard"
	})() + `">Dashboard</a>
            <a href="/settings">Settings</a>
            <a href="/logout">Logout</a>
        </nav>
    </header>

    <div class="container">
        <div class="card">
            <h2>Account Settings</h2>

            <div class="setting-item">
                <div class="setting-info">
                    <h3>Email</h3>
                    <p>` + user.Email + `</p>
                </div>
            </div>

            <div class="setting-item">
                <div class="setting-info">
                    <h3>Username</h3>
                    <p>` + user.Name + `</p>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>Security Settings</h2>

            <div class="setting-item">
                <div class="setting-info">
                    <h3>Two-Factor Authentication ` + totpStatusBadge + `</h3>
                    <p>Add an extra layer of security to your account using an authenticator app</p>
                </div>
                <div>
                    ` + totpActionButton + `
                </div>
            </div>
        </div>
    </div>

    <!-- Enable 2FA Modal -->
    <div id="enable2FAModal" class="modal">
        <div class="modal-content">
            <span class="close-btn" onclick="closeModal('enable2FAModal')">&times;</span>
            <h3>Enable Two-Factor Authentication</h3>
            <div id="enable2FAContent">
                <p>Click "Generate QR Code" to start setting up 2FA</p>
                <button onclick="generateQRCode()" class="btn btn-primary">Generate QR Code</button>
            </div>
        </div>
    </div>

    <!-- Disable 2FA Modal -->
    <div id="disable2FAModal" class="modal">
        <div class="modal-content">
            <span class="close-btn" onclick="closeModal('disable2FAModal')">&times;</span>
            <h3>Disable Two-Factor Authentication</h3>
            <p>Enter your password to disable 2FA</p>
            <div class="form-group">
                <label for="disable-password">Password</label>
                <input type="password" id="disable-password" required>
            </div>
            <button onclick="confirmDisable2FA()" class="btn btn-primary">Disable 2FA</button>
            <button onclick="closeModal('disable2FAModal')" class="btn btn-secondary" style="margin-left: 10px;">Cancel</button>
        </div>
    </div>

    <!-- Backup Codes Modal -->
    <div id="backupCodesModal" class="modal">
        <div class="modal-content">
            <span class="close-btn" onclick="closeModal('backupCodesModal')">&times;</span>
            <h3>Backup Codes</h3>
            <div id="backupCodesContent"></div>
            <button onclick="closeModal('backupCodesModal')" class="btn btn-primary">Close</button>
        </div>
    </div>

    <script>
        function enable2FA() {
            document.getElementById('enable2FAModal').style.display = 'flex';
        }

        function disable2FA() {
            document.getElementById('disable2FAModal').style.display = 'flex';
        }

        function closeModal(modalId) {
            document.getElementById(modalId).style.display = 'none';
        }

        async function generateQRCode() {
            try {
                const response = await fetch('/2fa/setup', {
                    method: 'POST',
                    credentials: 'same-origin'
                });
                const data = await response.json();

                if (data.success) {
                    const content = document.getElementById('enable2FAContent');
                    content.innerHTML = ` + "`" + `
                        <div class="alert alert-success">
                            QR Code generated successfully! Scan it with your authenticator app.
                        </div>
                        <div class="qr-code">
                            <img src="data:image/png;base64,${data.qr_code}" alt="QR Code">
                        </div>
                        <div class="secret-text">
                            <strong>Manual Entry Key:</strong><br>
                            ${data.secret}
                        </div>
                        <h4>Backup Codes (Save these!)</h4>
                        <div class="backup-codes">
                            ${data.backup_codes.map(code => ` + "`<div class='backup-code'>${code}</div>`" + `).join('')}
                        </div>
                        <div class="form-group">
                            <label for="verify-code">Enter the 6-digit code from your app to verify</label>
                            <input type="text" id="verify-code" maxlength="6" pattern="[0-9]{6}" required>
                        </div>
                        <button onclick="verify2FA()" class="btn btn-primary">Verify and Enable</button>
                    ` + "`" + `;
                } else {
                    alert('Failed to generate QR code');
                }
            } catch (error) {
                alert('Error: ' + error.message);
            }
        }

        async function verify2FA() {
            const code = document.getElementById('verify-code').value;
            if (code.length !== 6) {
                alert('Please enter a 6-digit code');
                return;
            }

            try {
                const response = await fetch('/2fa/enable', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: 'code=' + code,
                    credentials: 'same-origin'
                });
                const data = await response.json();

                if (data.success) {
                    alert('Two-factor authentication enabled successfully!');
                    location.reload();
                } else {
                    alert('Error: ' + data.error);
                }
            } catch (error) {
                alert('Error: ' + error.message);
            }
        }

        async function confirmDisable2FA() {
            const password = document.getElementById('disable-password').value;
            if (!password) {
                alert('Please enter your password');
                return;
            }

            try {
                const response = await fetch('/2fa/disable', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: 'password=' + encodeURIComponent(password),
                    credentials: 'same-origin'
                });
                const data = await response.json();

                if (data.success) {
                    alert('Two-factor authentication disabled');
                    location.reload();
                } else {
                    alert('Error: ' + data.error);
                }
            } catch (error) {
                alert('Error: ' + error.message);
            }
        }

        async function regenerateBackupCodes() {
            if (!confirm('This will invalidate all existing backup codes. Continue?')) {
                return;
            }

            try {
                const response = await fetch('/2fa/regenerate-backup-codes', {
                    method: 'POST',
                    credentials: 'same-origin'
                });
                const data = await response.json();

                if (data.success) {
                    const content = document.getElementById('backupCodesContent');
                    content.innerHTML = ` + "`" + `
                        <div class="alert alert-success">
                            New backup codes generated! Save these in a safe place.
                        </div>
                        <div class="backup-codes">
                            ${data.backup_codes.map(code => ` + "`<div class='backup-code'>${code}</div>`" + `).join('')}
                        </div>
                        <p style="color: #c33; font-weight: 600; margin-top: 15px;">
                            ⚠️ Your old backup codes no longer work. Save these new ones!
                        </p>
                    ` + "`" + `;
                    document.getElementById('backupCodesModal').style.display = 'flex';
                } else {
                    alert('Failed to regenerate backup codes');
                }
            } catch (error) {
                alert('Error: ' + error.message);
            }
        }

        // Close modal when clicking outside
        window.onclick = function(event) {
            if (event.target.classList.contains('modal')) {
                event.target.style.display = 'none';
            }
        }
    </script>
</body>
</html>`

	w.Write([]byte(html))
}
