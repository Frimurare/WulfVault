// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU General Public License v3.0 (GPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

// AuthenticationConfig holds configuration on how to authenticate to Gokapi admin menu
type AuthenticationConfig struct {
	Method               int      `json:"Method"`
	SaltAdmin            string   `json:"SaltAdmin"`
	SaltFiles            string   `json:"SaltFiles"`
	Username             string   `json:"Username"`
	HeaderKey            string   `json:"HeaderKey"`
	OAuthProvider        string   `json:"OauthProvider"`
	OAuthClientId        string   `json:"OAuthClientId"`
	OAuthClientSecret    string   `json:"OAuthClientSecret"`
	OAuthGroupScope      string   `json:"OauthGroupScope"`
	OAuthRecheckInterval int      `json:"OAuthRecheckInterval"`
	OAuthGroups          []string `json:"OAuthGroups"`
	OnlyRegisteredUsers  bool     `json:"OnlyRegisteredUsers"`
}

const (
	// AuthenticationInternal authentication method uses a user / password combination handled by Gokapi
	AuthenticationInternal = iota

	// AuthenticationOAuth2 authentication retrieves the users email with Open Connect ID
	AuthenticationOAuth2

	// AuthenticationHeader authentication relies on a header from a reverse proxy to parse the username
	AuthenticationHeader

	// AuthenticationDisabled authentication ignores all internal authentication procedures. A reverse proxy needs to restrict access
	AuthenticationDisabled
)
