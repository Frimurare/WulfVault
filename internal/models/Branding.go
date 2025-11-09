package models

import "encoding/json"

// Branding contains the configurable branding settings
type Branding struct {
	CompanyName    string `json:"companyName"`    // Company name displayed throughout the app
	PrimaryColor   string `json:"primaryColor"`   // Primary brand color (hex, e.g., #0066CC)
	SecondaryColor string `json:"secondaryColor"` // Secondary color for accents
	LogoPath       string `json:"logoPath"`       // Path to uploaded logo file
	LogoBase64     string `json:"logoBase64"`     // Base64 encoded logo (for easy embedding)
	FaviconPath    string `json:"faviconPath"`    // Path to favicon
	FooterText     string `json:"footerText"`     // Footer text (e.g., "© 2024 Company Name")
	WelcomeMessage string `json:"welcomeMessage"` // Welcome message on login page
	CustomCSS      string `json:"customCSS"`      // Optional custom CSS
}

// DefaultBranding returns the default branding configuration
func DefaultBranding() Branding {
	return Branding{
		CompanyName:    "Sharecare",
		PrimaryColor:   "#0066CC",
		SecondaryColor: "#333333",
		FooterText:     "Secure File Sharing",
		WelcomeMessage: "Welcome to Sharecare - Secure File Sharing",
	}
}

// ToJson returns the branding as a JSON object
func (b *Branding) ToJson() string {
	result, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(result)
}

// IsLogoUploaded returns true if a logo has been uploaded
func (b *Branding) IsLogoUploaded() bool {
	return b.LogoPath != "" || b.LogoBase64 != ""
}
