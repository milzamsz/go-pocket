package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv  string
	BaseURL string
	Polar   PolarConfig
	Resend  ResendConfig
	OAuth   OAuthConfig
}

type PolarConfig struct {
	AccessToken   string
	WebhookSecret string
}

type ResendConfig struct {
	APIKey        string
	From          string
	WebhookSecret string
}

type OAuthConfig struct {
	Google OAuthProviderConfig
	GitHub OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
}

func Load() Config {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = getenv("APP_URL", "http://localhost:8090")
	}

	return Config{
		AppEnv:  getenv("APP_ENV", "development"),
		BaseURL: baseURL,
		Polar: PolarConfig{
			AccessToken:   os.Getenv("POLAR_ACCESS_TOKEN"),
			WebhookSecret: os.Getenv("POLAR_WEBHOOK_SECRET"),
		},
		Resend: ResendConfig{
			APIKey:        os.Getenv("RESEND_API_KEY"),
			From:          getenv("RESEND_FROM", "noreply@example.com"),
			WebhookSecret: os.Getenv("RESEND_WEBHOOK_SECRET"),
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
			},
			GitHub: OAuthProviderConfig{
				ClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
			},
		},
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsProduction reports whether the app is running in a production-like
// environment. Used to switch fail-open development conveniences (e.g. skipping
// webhook signature checks when no secret is configured) into fail-closed
// behavior.
func (c Config) IsProduction() bool {
	switch strings.ToLower(strings.TrimSpace(c.AppEnv)) {
	case "production", "prod":
		return true
	default:
		return false
	}
}

// Validate returns a list of human-readable warnings about configuration that
// is unsafe or incomplete for the current environment. It never fails the boot;
// callers are expected to log the warnings so operators can act on them.
func (c Config) Validate() []string {
	var warnings []string
	if !c.IsProduction() {
		return warnings
	}
	if strings.TrimSpace(c.Polar.WebhookSecret) == "" {
		warnings = append(warnings, "POLAR_WEBHOOK_SECRET is not set; Polar webhooks will be rejected in production")
	}
	if strings.TrimSpace(c.Resend.WebhookSecret) == "" {
		warnings = append(warnings, "RESEND_WEBHOOK_SECRET is not set; Resend webhooks will be rejected in production")
	}
	if strings.TrimSpace(c.Polar.AccessToken) == "" {
		warnings = append(warnings, "POLAR_ACCESS_TOKEN is not set; billing falls back to unconfigured URLs")
	}
	if strings.TrimSpace(c.Resend.APIKey) == "" {
		warnings = append(warnings, "RESEND_API_KEY is not set; outbound email is disabled")
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(c.BaseURL)), "https://") {
		warnings = append(warnings, "APP_BASE_URL should use https in production")
	}
	return warnings
}
