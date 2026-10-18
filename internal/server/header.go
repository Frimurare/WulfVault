// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"github.com/Frimurare/WulfVault/internal/i18n"
	"github.com/Frimurare/WulfVault/internal/models"
)

// getFaviconHTML also supplies the common theme assets to standalone pages.
func (s *Server) getFaviconHTML() string {
	return `<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🐺</text></svg>">` + appShellAssets
}

func (s *Server) getAdminHeaderHTML(pageTitle string, tr *i18n.Translator) string {
	return s.renderAppShell(tr, nil, "admin")
}

func (s *Server) getDownloadUserHeaderHTML(tr *i18n.Translator) string {
	return s.renderAppShell(tr, nil, "download")
}

func (s *Server) getHeaderHTML(user *models.User, forAdmin bool, tr *i18n.Translator) string {
	if user != nil && user.IsAdmin() && forAdmin {
		return s.renderAppShell(tr, user, "admin")
	}
	return s.renderAppShell(tr, user, "user")
}
