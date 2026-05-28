package config

import "os"

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
