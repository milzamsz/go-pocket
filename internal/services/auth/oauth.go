package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
	ghoauth "golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type oauthHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type oauthProvider struct {
	name      string
	oauth     *oauth2.Config
	userInfo  string
	emailPath string
	client    oauthHTTPClient
}

type oauthProfile struct {
	Email string
	Name  string
}

var errOAuthProviderUnsupported = errors.New("oauth provider unsupported")

func newOAuthProviders(cfg config.Config, client oauthHTTPClient) map[string]oauthProvider {
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	providers := map[string]oauthProvider{}

	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		providers["google"] = oauthProvider{
			name: "google",
			oauth: &oauth2.Config{
				ClientID:     cfg.OAuth.Google.ClientID,
				ClientSecret: cfg.OAuth.Google.ClientSecret,
				RedirectURL:  baseURL + "/auth/oauth/google/callback",
				Scopes:       []string{"openid", "profile", "email"},
				Endpoint:     google.Endpoint,
			},
			userInfo: "https://openidconnect.googleapis.com/v1/userinfo",
			client:   client,
		}
	}

	if cfg.OAuth.GitHub.ClientID != "" && cfg.OAuth.GitHub.ClientSecret != "" {
		providers["github"] = oauthProvider{
			name: "github",
			oauth: &oauth2.Config{
				ClientID:     cfg.OAuth.GitHub.ClientID,
				ClientSecret: cfg.OAuth.GitHub.ClientSecret,
				RedirectURL:  baseURL + "/auth/oauth/github/callback",
				Scopes:       []string{"read:user", "user:email"},
				Endpoint:     ghoauth.Endpoint,
			},
			userInfo:  "https://api.github.com/user",
			emailPath: "https://api.github.com/user/emails",
			client:    client,
		}
	}

	return providers
}

func (s *service) OAuthRedirectURL(_ context.Context, provider string, state string) (string, error) {
	p, ok := s.oauth[strings.ToLower(strings.TrimSpace(provider))]
	if !ok || p.oauth == nil {
		return "", errOAuthProviderUnsupported
	}
	return p.oauth.AuthCodeURL(state), nil
}

func (s *service) OAuthCallback(ctx context.Context, provider string, code string) (*AuthSession, error) {
	p, ok := s.oauth[strings.ToLower(strings.TrimSpace(provider))]
	if !ok || p.oauth == nil {
		return nil, errOAuthProviderUnsupported
	}
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("oauth code is required")
	}
	if s.app == nil {
		return nil, errors.New("auth store is not configured")
	}

	token, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}
	profile, err := p.fetchProfile(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("fetch oauth profile: %w", err)
	}
	if strings.TrimSpace(profile.Email) == "" {
		return nil, domain.ErrUnauthenticated
	}
	userRecord, err := s.findOrCreateOAuthUser(ctx, profile)
	if err != nil {
		return nil, err
	}
	authToken, err := userRecord.NewAuthToken()
	if err != nil {
		return nil, fmt.Errorf("issue auth token: %w", err)
	}
	return &AuthSession{User: mapRecordToDomainUser(userRecord), Token: authToken}, nil
}

func (p oauthProvider) fetchProfile(ctx context.Context, accessToken string) (oauthProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfo, nil)
	if err != nil {
		return oauthProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return oauthProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oauthProfile{}, fmt.Errorf("userinfo status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthProfile{}, err
	}

	switch p.name {
	case "google":
		var payload struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return oauthProfile{}, err
		}
		return oauthProfile{Email: strings.ToLower(strings.TrimSpace(payload.Email)), Name: strings.TrimSpace(payload.Name)}, nil
	case "github":
		var payload struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Login string `json:"login"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return oauthProfile{}, err
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		if email == "" && p.emailPath != "" {
			resolved, err := p.fetchGitHubPrimaryEmail(ctx, accessToken)
			if err == nil {
				email = resolved
			}
		}
		name := strings.TrimSpace(payload.Name)
		if name == "" {
			name = strings.TrimSpace(payload.Login)
		}
		return oauthProfile{Email: email, Name: name}, nil
	default:
		return oauthProfile{}, errOAuthProviderUnsupported
	}
}

func (p oauthProvider) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.emailPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github emails status %d", resp.StatusCode)
	}
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, email := range emails {
		if email.Primary && strings.TrimSpace(email.Email) != "" {
			return strings.ToLower(strings.TrimSpace(email.Email)), nil
		}
	}
	for _, email := range emails {
		if strings.TrimSpace(email.Email) != "" {
			return strings.ToLower(strings.TrimSpace(email.Email)), nil
		}
	}
	return "", nil
}

func (s *service) findOrCreateOAuthUser(ctx context.Context, profile oauthProfile) (*core.Record, error) {
	record, err := s.app.FindAuthRecordByEmail("users", profile.Email)
	if err == nil && record != nil {
		if !record.Verified() {
			record.Set("verified", true)
			if err := s.app.SaveWithContext(ctx, record); err != nil {
				return nil, fmt.Errorf("mark oauth user verified: %w", err)
			}
		}
		return record, nil
	}

	users, err := s.app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, fmt.Errorf("find users collection: %w", err)
	}
	user := core.NewRecord(users)
	user.Set("email", strings.ToLower(strings.TrimSpace(profile.Email)))
	user.Set("name", strings.TrimSpace(profile.Name))
	user.Set("verified", true)
	// OAuth-created users still need a local credential placeholder.
	placeholderPassword := "oauth-login-disabled-" + strings.ReplaceAll(strings.TrimSpace(profile.Email), "@", "-")
	user.Set("password", placeholderPassword)
	user.Set("passwordConfirm", placeholderPassword)
	if err := s.app.SaveWithContext(ctx, user); err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	return user, nil
}

func oauthCallbackURL(baseURL string, provider string) string {
	u, _ := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if u == nil {
		return ""
	}
	u.Path = "/auth/oauth/" + provider + "/callback"
	return u.String()
}
