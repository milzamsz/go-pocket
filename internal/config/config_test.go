package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsProduction(t *testing.T) {
	t.Parallel()

	require.True(t, Config{AppEnv: "production"}.IsProduction())
	require.True(t, Config{AppEnv: "PROD"}.IsProduction())
	require.False(t, Config{AppEnv: "development"}.IsProduction())
	require.False(t, Config{}.IsProduction())
}

func TestValidate_NoWarningsOutsideProduction(t *testing.T) {
	t.Parallel()

	require.Empty(t, Config{AppEnv: "development"}.Validate())
}

func TestValidate_FlagsMissingProductionSecrets(t *testing.T) {
	t.Parallel()

	warnings := Config{AppEnv: "production", BaseURL: "http://example.com"}.Validate()
	require.NotEmpty(t, warnings)

	joined := ""
	for _, w := range warnings {
		joined += w + "\n"
	}
	require.Contains(t, joined, "POLAR_WEBHOOK_SECRET")
	require.Contains(t, joined, "RESEND_WEBHOOK_SECRET")
	require.Contains(t, joined, "https")
}

func TestValidate_CleanProductionConfigHasNoWarnings(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:  "production",
		BaseURL: "https://app.example.com",
		Polar:   PolarConfig{AccessToken: "polar-token", WebhookSecret: "whsec_x"},
		Resend:  ResendConfig{APIKey: "re_x", WebhookSecret: "whsec_y"},
	}
	require.Empty(t, cfg.Validate())
}
