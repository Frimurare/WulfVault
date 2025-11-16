// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
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

	// Get branding config for logo
	brandingConfig, _ := database.DB.GetBrandingConfig()
	logoData := brandingConfig["branding_logo"]

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Settings - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px 40px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header .logo {
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .header .logo img {
            max-height: 50px;
            max-width: 180px;
        }
        .header h1 {
            color: white;
            font-size: 24px;
            font-weight: 600;
        }
        .header nav {
            display: flex;
            align-items: center;
            gap: 20px;
        }
        .header nav a {
            color: rgba(255, 255, 255, 0.9);
            text-decoration: none;
            font-weight: 500;
            transition: color 0.3s;
        }
        .header nav a:hover {
            color: white;
        }
        .container {
            max-width: 1200px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
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

        /* Hamburger menu styles */
        .hamburger {
            display: none;
            flex-direction: column;
            background: none;
            border: none;
            cursor: pointer;
            padding: 8px;
            z-index: 1001;
            transition: transform 0.3s ease;
        }
        .hamburger span {
            width: 25px;
            height: 3px;
            background: white;
            margin: 3px 0;
            transition: all 0.3s ease;
            border-radius: 3px;
        }
        .hamburger.active span:nth-child(1) {
            transform: rotate(45deg) translate(8px, 8px);
        }
        .hamburger.active span:nth-child(2) {
            opacity: 0;
        }
        .hamburger.active span:nth-child(3) {
            transform: rotate(-45deg) translate(7px, -7px);
        }

        /* Mobile navigation overlay */
        .mobile-nav-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            z-index: 999;
            opacity: 0;
            transition: opacity 0.3s ease;
        }
        .mobile-nav-overlay.active {
            display: block;
            opacity: 1;
        }

        /* Mobile responsive styles */
        @media (max-width: 768px) {
            .header {
                padding: 15px 20px;
                flex-wrap: wrap;
            }

            .header h1 {
                font-size: 18px;
                order: 1;
                flex: 1;
            }

            .header .logo img {
                max-height: 40px;
                max-width: 150px;
            }

            .hamburger {
                display: flex;
                order: 3;
                margin-left: auto;
            }

            .header nav {
                position: fixed;
                top: 0;
                right: -100%;
                width: 280px;
                height: 100vh;
                background: linear-gradient(180deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
                flex-direction: column;
                align-items: flex-start;
                padding: 80px 20px 20px;
                box-shadow: -2px 0 10px rgba(0,0,0,0.1);
                z-index: 1000;
                transition: right 0.3s ease;
                overflow-y: auto;
            }

            .header nav.active {
                right: 0;
            }

            .header nav a {
                color: rgba(255, 255, 255, 0.9);
                padding: 15px 20px;
                width: 100%;
                border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                margin: 0;
            }

            .header nav a:hover {
                background: rgba(255, 255, 255, 0.1);
                color: white;
            }

            .container {
                margin: 20px auto;
                padding: 0 15px;
            }

            .card {
                padding: 20px;
                border-radius: 8px;
            }

            .card h2 {
                font-size: 20px;
                margin-bottom: 15px;
            }

            .setting-item {
                flex-direction: column;
                align-items: flex-start;
                padding: 15px;
                gap: 15px;
            }

            .setting-item h3 {
                font-size: 16px;
            }

            .setting-item p {
                font-size: 13px;
            }

            .setting-item > div {
                width: 100%;
            }

            .setting-item button {
                width: 100%;
                margin: 5px 0 !important;
                padding: 12px 20px !important;
                font-size: 14px !important;
            }

            .modal-content {
                width: 95%;
                padding: 20px;
                margin: 10px;
            }

            .modal-content h3 {
                font-size: 18px;
            }

            .form-group input {
                padding: 14px;
                font-size: 16px;
                min-height: 48px;
            }

            .btn {
                width: 100%;
                padding: 14px 24px;
                font-size: 16px;
                min-height: 48px;
                margin: 5px 0 !important;
            }

            .qr-code img {
                max-width: 100%;
                height: auto;
            }

            .backup-codes {
                font-size: 12px;
            }

            .close-btn {
                font-size: 28px;
                padding: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">`

	if logoData != "" {
		html += `
            <img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		html += `
            <h1>` + s.config.CompanyName + `</h1>`
	}

	html += `
        </div>
        <button class="hamburger" aria-label="Toggle navigation" aria-expanded="false">
            <span></span>
            <span></span>
            <span></span>
        </button>
        <nav>`

	// Different navigation for admin vs regular user
	if user.IsAdmin() {
		html += `
            <a href="/admin">Admin Dashboard</a>
            <a href="/dashboard">My Files</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/teams">Teams</a>
            <a href="/admin/files">All Files</a>
            <a href="/admin/trash">Trash</a>
            <a href="/admin/branding">Branding</a>
            <a href="/admin/email-settings">Email</a>
            <a href="/admin/settings">Server</a>
            <a href="/settings">My Account</a>
            <a href="/logout" style="margin-left: auto;">Logout</a>`
	} else {
		html += `
            <a href="/dashboard">Dashboard</a>
            <a href="/teams">Teams</a>
            <a href="/settings">Settings</a>
            <a href="/logout" style="margin-left: auto;">Logout</a>`
	}

	html += `
        </nav>
    </div>
    <div class="mobile-nav-overlay"></div>

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
                    <h3>Password</h3>
                    <p>Change your account password</p>
                </div>
                <div>
                    <button onclick="changePassword()" style="background: ` + s.getPrimaryColor() + `; color: white; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600;">
                        Change Password
                    </button>
                </div>
            </div>

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

    <!-- Change Password Modal -->
    <div id="changePasswordModal" class="modal">
        <div class="modal-content">
            <span class="close-btn" onclick="closeModal('changePasswordModal')">&times;</span>
            <h3>Change Password</h3>
            <div id="changePasswordMessage"></div>
            <div class="form-group">
                <label for="current-password">Current Password</label>
                <input type="password" id="current-password" required autocomplete="current-password">
            </div>
            <div class="form-group">
                <label for="new-password">New Password</label>
                <input type="password" id="new-password" required autocomplete="new-password">
            </div>
            <div class="form-group">
                <label for="confirm-password">Confirm New Password</label>
                <input type="password" id="confirm-password" required autocomplete="new-password">
            </div>
            <button onclick="confirmChangePassword()" class="btn btn-primary">Change Password</button>
            <button onclick="closeModal('changePasswordModal')" class="btn btn-secondary" style="margin-left: 10px;">Cancel</button>
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
        function changePassword() {
            document.getElementById('changePasswordModal').style.display = 'flex';
            document.getElementById('changePasswordMessage').innerHTML = '';
            document.getElementById('current-password').value = '';
            document.getElementById('new-password').value = '';
            document.getElementById('confirm-password').value = '';
        }

        async function confirmChangePassword() {
            const currentPassword = document.getElementById('current-password').value;
            const newPassword = document.getElementById('new-password').value;
            const confirmPassword = document.getElementById('confirm-password').value;
            const messageDiv = document.getElementById('changePasswordMessage');

            // Validation
            if (!currentPassword || !newPassword || !confirmPassword) {
                messageDiv.innerHTML = '<div class="alert alert-error">All fields are required</div>';
                return;
            }

            if (newPassword.length < 8) {
                messageDiv.innerHTML = '<div class="alert alert-error">New password must be at least 8 characters</div>';
                return;
            }

            if (newPassword !== confirmPassword) {
                messageDiv.innerHTML = '<div class="alert alert-error">New passwords do not match</div>';
                return;
            }

            if (currentPassword === newPassword) {
                messageDiv.innerHTML = '<div class="alert alert-error">New password must be different from current password</div>';
                return;
            }

            try {
                const response = await fetch('/change-password', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: 'current_password=' + encodeURIComponent(currentPassword) +
                          '&new_password=' + encodeURIComponent(newPassword),
                    credentials: 'same-origin'
                });
                const data = await response.json();

                if (data.success) {
                    messageDiv.innerHTML = '<div class="alert alert-success">' + data.message + '</div>';
                    setTimeout(() => {
                        closeModal('changePasswordModal');
                    }, 2000);
                } else {
                    messageDiv.innerHTML = '<div class="alert alert-error">' + data.error + '</div>';
                }
            } catch (error) {
                messageDiv.innerHTML = '<div class="alert alert-error">Error: ' + error.message + '</div>';
            }
        }

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
    <script>
    (function() {
        'use strict';
        function initMobileNav() {
            const header = document.querySelector('.header');
            if (!header) return;
            const nav = header.querySelector('nav');
            if (!nav) return;
            const hamburger = header.querySelector('.hamburger');
            if (!hamburger) return;
            let overlay = document.querySelector('.mobile-nav-overlay');
            if (!overlay) {
                overlay = document.createElement('div');
                overlay.className = 'mobile-nav-overlay';
                document.body.appendChild(overlay);
            }
            function toggleNav() {
                const isActive = nav.classList.contains('active');
                if (isActive) {
                    nav.classList.remove('active');
                    hamburger.classList.remove('active');
                    overlay.classList.remove('active');
                    hamburger.setAttribute('aria-expanded', 'false');
                    document.body.style.overflow = '';
                } else {
                    nav.classList.add('active');
                    hamburger.classList.add('active');
                    overlay.classList.add('active');
                    hamburger.setAttribute('aria-expanded', 'true');
                    document.body.style.overflow = 'hidden';
                }
            }
            hamburger.addEventListener('click', toggleNav);
            overlay.addEventListener('click', toggleNav);
            const navLinks = nav.querySelectorAll('a');
            navLinks.forEach(link => {
                link.addEventListener('click', () => {
                    if (window.innerWidth <= 768) {
                        toggleNav();
                    }
                });
            });
            let resizeTimer;
            window.addEventListener('resize', () => {
                clearTimeout(resizeTimer);
                resizeTimer = setTimeout(() => {
                    if (window.innerWidth > 768 && nav.classList.contains('active')) {
                        toggleNav();
                    }
                }, 250);
            });
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && nav.classList.contains('active')) {
                    toggleNav();
                }
            });
            const tables = document.querySelectorAll('table');
            tables.forEach(table => {
                const headers = table.querySelectorAll('th');
                const headerTexts = Array.from(headers).map(th => th.textContent.trim());
                const rows = table.querySelectorAll('tbody tr');
                rows.forEach(row => {
                    const cells = row.querySelectorAll('td');
                    cells.forEach((cell, index) => {
                        if (headerTexts[index]) {
                            cell.setAttribute('data-label', headerTexts[index]);
                        }
                    });
                });
            });
        }
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', initMobileNav);
        } else {
            initMobileNav();
        }
    })();
    </script>
    <div style="text-align:center; font-size: 0.8em; margin-top: 2em; padding: 1em; color:#777;">
        Powered by WulfVault © Ulf Holmström – AGPL-3.0
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// handleChangePassword handles password change for users and admins
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid form data",
		})
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")

	// Validate inputs
	if currentPassword == "" || newPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "All fields are required",
		})
		return
	}

	if len(newPassword) < 8 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "New password must be at least 8 characters",
		})
		return
	}

	if currentPassword == newPassword {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "New password must be different from current password",
		})
		return
	}

	// Verify current password
	_, err = auth.AuthenticateUser(user.Email, currentPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to hash password",
		})
		return
	}

	// Update password in database
	if err := database.DB.UpdateUserPassword(user.Id, hashedPassword); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to update password",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	})
}
