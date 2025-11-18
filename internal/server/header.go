// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// getAdminHeaderHTML returns branded header HTML for admin pages (compatibility wrapper)
func (s *Server) getAdminHeaderHTML(pageTitle string) string {
	// Create a dummy admin user for header rendering
	user := &models.User{UserLevel: models.UserLevelAdmin}
	return s.getHeaderHTML(user, true)
}

// getHeaderHTML generates consistent header HTML for all pages
// forAdmin: true shows admin navigation, false shows user navigation
func (s *Server) getHeaderHTML(user *models.User, forAdmin bool) string {
	brandingConfig, _ := database.DB.GetBrandingConfig()
	logoData := brandingConfig["branding_logo"]

	headerCSS := `
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
            gap: 10px;
        }
        .header nav a, .header nav .dropdown-toggle {
            color: white;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 5px;
            background: rgba(255, 255, 255, 0.2);
            transition: background 0.3s;
        }
        .header nav a:hover, .header nav .dropdown-toggle:hover {
            background: rgba(255, 255, 255, 0.3);
        }
        .header nav span {
            color: rgba(255, 255, 255, 0.6);
            font-size: 11px;
            font-weight: 400;
        }
        /* Dropdown Menu Styles */
        .header nav .dropdown {
            position: relative;
            display: inline-block;
        }
        .header nav .dropdown-toggle {
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .header nav .dropdown-toggle::after {
            content: '▾';
            font-size: 12px;
        }
        .header nav .dropdown-content {
            display: none;
            position: absolute;
            top: calc(100% + 2px);
            left: 0;
            min-width: 180px;
            background: white;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            border-radius: 5px;
            z-index: 1000;
            padding: 4px 0;
        }
        .header nav .dropdown-content a {
            color: #333 !important;
            background: white !important;
            display: block;
            padding: 12px 16px;
            border-radius: 0 !important;
            white-space: nowrap;
        }
        .header nav .dropdown-content a:first-child {
            border-radius: 5px 5px 0 0 !important;
        }
        .header nav .dropdown-content a:last-child {
            border-radius: 0 0 5px 5px !important;
        }
        .header nav .dropdown-content a:hover {
            background: #f5f5f5 !important;
        }
        .header nav .dropdown:hover .dropdown-content {
            display: block;
        }
        /* Extend hover area to prevent gap issues */
        .header nav .dropdown::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: calc(100% + 10px);
        }

        /* Mobile Navigation Styles */
        .hamburger {
            display: none;
            flex-direction: column;
            cursor: pointer;
            padding: 8px;
            background: none;
            border: none;
            z-index: 1001;
        }
        .hamburger span {
            width: 25px;
            height: 3px;
            background: white;
            margin: 3px 0;
            transition: 0.3s;
            border-radius: 2px;
        }
        .hamburger.active span:nth-child(1) {
            transform: rotate(-45deg) translate(-5px, 6px);
        }
        .hamburger.active span:nth-child(2) {
            opacity: 0;
        }
        .hamburger.active span:nth-child(3) {
            transform: rotate(45deg) translate(-5px, -6px);
        }
        .mobile-nav-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            z-index: 999;
        }
        .mobile-nav-overlay.active {
            display: block;
        }

        @media screen and (max-width: 768px) {
            .header {
                padding: 15px 20px !important;
                flex-wrap: wrap;
            }
            .header .logo h1 {
                font-size: 18px !important;
            }
            .header .logo img {
                max-height: 40px !important;
                max-width: 120px !important;
            }
            .hamburger {
                display: flex !important;
                order: 3;
            }
            .header nav {
                display: none !important;
                position: fixed !important;
                top: 0 !important;
                right: -100% !important;
                width: 280px !important;
                height: 100vh !important;
                background: linear-gradient(180deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%) !important;
                flex-direction: column !important;
                align-items: flex-start !important;
                padding: 80px 30px 30px !important;
                gap: 0 !important;
                transition: right 0.3s ease !important;
                z-index: 1000 !important;
                overflow-y: auto !important;
                box-shadow: -5px 0 15px rgba(0,0,0,0.3) !important;
            }
            .header nav.active {
                display: flex !important;
                right: 0 !important;
            }
            .header nav a, .header nav .dropdown-toggle {
                width: 100%;
                padding: 15px 20px !important;
                border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                font-size: 16px !important;
                margin: 0 !important;
            }
            .header nav a:hover, .header nav .dropdown-toggle:hover {
                background: rgba(255, 255, 255, 0.1);
            }
            .header nav span {
                padding: 15px 20px !important;
                margin: 0 !important;
            }
            /* Mobile dropdown styles */
            .header nav .dropdown {
                width: 100%;
                display: block !important;
            }
            .header nav .dropdown-toggle {
                width: 100%;
                display: block !important;
                pointer-events: none !important;
                cursor: default !important;
                opacity: 0.7 !important;
            }
            .header nav .dropdown-toggle::after {
                display: none !important;
            }
            .header nav .dropdown-content {
                position: static !important;
                display: block !important;
                background: rgba(0, 0, 0, 0.2) !important;
                box-shadow: none !important;
                border-radius: 0 !important;
                padding: 0 !important;
                margin: 0 !important;
                z-index: 100 !important;
            }
            .header nav .dropdown-content a {
                color: rgba(255, 255, 255, 0.8) !important;
                background: transparent !important;
                padding: 12px 20px 12px 40px !important;
                border-bottom: 1px solid rgba(255, 255, 255, 0.05) !important;
                font-size: 15px !important;
                pointer-events: auto !important;
                cursor: pointer !important;
                position: relative !important;
                z-index: 101 !important;
                display: block !important;
            }
            .header nav .dropdown-content a:hover,
            .header nav .dropdown-content a:active {
                background: rgba(255, 255, 255, 0.1) !important;
                color: white !important;
            }
        }`

	headerHTML := `
    <div class="header">
        <div class="logo">`

	if logoData != "" {
		headerHTML += `
            <img src="` + logoData + `" alt="` + s.config.CompanyName + `">`
	} else {
		headerHTML += `
            <h1>` + s.config.CompanyName + `</h1>`
	}

	headerHTML += `
        </div>
        <button class="hamburger" aria-label="Toggle navigation" aria-expanded="false">
            <span></span>
            <span></span>
            <span></span>
        </button>
        <nav>`

	// Different navigation based on user type and page context
	if user.IsAdmin() && forAdmin {
		// Full admin navigation
		headerHTML += `
            <a href="/admin">Admin Dashboard</a>
            <a href="/dashboard">My Files</a>
            <a href="/admin/users">Users</a>
            <a href="/admin/teams">Teams</a>
            <div class="dropdown">
                <a class="dropdown-toggle">Files</a>
                <div class="dropdown-content">
                    <a href="/admin/files">All Files</a>
                    <a href="/admin/trash">Trash</a>
                </div>
            </div>
            <div class="dropdown">
                <a class="dropdown-toggle">Server</a>
                <div class="dropdown-content">
                    <a href="/admin/settings">Server Settings</a>
                    <a href="/admin/branding">Branding</a>
                    <a href="/admin/email-settings">Email</a>
                    <a href="/admin/audit-logs">Audit Logs</a>
                </div>
            </div>
            <a href="/settings">My Account</a>
            <a href="/logout" style="margin-left: auto;">Logout</a>
            <span>v` + s.config.Version + `</span>`
	} else {
		// Regular user navigation
		headerHTML += `
            <a href="/dashboard">Dashboard</a>
            <a href="/teams">Teams</a>
            <a href="/settings">Settings</a>
            <a href="/logout" style="margin-left: auto;">Logout</a>
            <span>v` + s.config.Version + `</span>`
	}

	headerHTML += `
        </nav>
    </div>
    <div class="mobile-nav-overlay"></div>
    <script>
        // Mobile navigation toggle
        document.addEventListener('DOMContentLoaded', function() {
            const hamburger = document.querySelector('.hamburger');
            const nav = document.querySelector('.header nav');
            const overlay = document.querySelector('.mobile-nav-overlay');

            if (!hamburger || !nav || !overlay) {
                console.error('Mobile nav elements not found:', {hamburger, nav, overlay});
                return;
            }

            function toggleNav() {
                const isActive = nav.classList.contains('active');

                if (isActive) {
                    nav.classList.remove('active');
                    hamburger.classList.remove('active');
                    overlay.classList.remove('active');
                    hamburger.setAttribute('aria-expanded', 'false');
                } else {
                    nav.classList.add('active');
                    hamburger.classList.add('active');
                    overlay.classList.add('active');
                    hamburger.setAttribute('aria-expanded', 'true');
                }
            }

            hamburger.addEventListener('click', function(e) {
                e.preventDefault();
                e.stopPropagation();
                toggleNav();
            });

            overlay.addEventListener('click', function(e) {
                e.preventDefault();
                e.stopPropagation();
                toggleNav();
            });
        });
    </script>`

	// Wolf emoji favicon as SVG data URI
	faviconSVG := `<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🐺</text></svg>">`

	return faviconSVG + `<link rel="stylesheet" href="/static/css/style.css"><style>` + headerCSS + `</style>` + headerHTML
}

