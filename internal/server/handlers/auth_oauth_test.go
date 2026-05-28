package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

type oauthAuthService struct {
	redirectURL string
	redirectErr error
	callbackErr error
	session     *auth.AuthSession
}

func (m *oauthAuthService) CurrentUser(context.Context) (*domain.User, error) { return nil, nil }
func (m *oauthAuthService) Login(context.Context, string, string) (*auth.AuthSession, error) {
	return nil, nil
}
func (m *oauthAuthService) Signup(context.Context, string, string, string) (*auth.AuthSession, error) {
	return nil, nil
}
func (m *oauthAuthService) Logout(context.Context, string) error                        { return nil }
func (m *oauthAuthService) RequestPasswordReset(context.Context, string) error          { return nil }
func (m *oauthAuthService) ResetPassword(context.Context, string, string, string) error { return nil }
func (m *oauthAuthService) VerifyEmail(context.Context, string) error                   { return nil }
func (m *oauthAuthService) UpdateProfile(context.Context, string, string, string) error { return nil }
func (m *oauthAuthService) ChangePassword(context.Context, string, string, string, string) error {
	return nil
}
func (m *oauthAuthService) UpdateTwoFactor(context.Context, string, bool) error { return nil }
func (m *oauthAuthService) OAuthRedirectURL(context.Context, string, string) (string, error) {
	return m.redirectURL, m.redirectErr
}
func (m *oauthAuthService) OAuthCallback(context.Context, string, string) (*auth.AuthSession, error) {
	return m.session, m.callbackErr
}

func TestAuthOAuthStart_RedirectsAndSetsStateCookie(t *testing.T) {
	t.Parallel()

	e, rec := newGetEvent("/auth/oauth/google")
	e.Request.SetPathValue("provider", "google")
	svc := &oauthAuthService{redirectURL: "https://accounts.google.com/o/oauth2/v2/auth?state=abc"}

	err := AuthOAuthStart(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, svc.redirectURL, rec.Header().Get("Location"))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, oauthStateCookieName, cookies[0].Name)
	require.Contains(t, cookies[0].Value, "google:")
}

func TestAuthOAuthCallback_ValidState_SetsAuthCookieAndRedirects(t *testing.T) {
	t.Parallel()

	callbackPath := "/auth/oauth/google/callback?state=state123&code=code123"
	e, rec := newGetEvent(callbackPath)
	e.Request.SetPathValue("provider", "google")
	e.Request.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "google:state123"})
	svc := &oauthAuthService{
		session: &auth.AuthSession{
			User:  &domain.User{ID: "u-1"},
			Token: "oauth-token",
		},
	}

	err := AuthOAuthCallback(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app", rec.Header().Get("Location"))
	allCookies := rec.Result().Cookies()
	require.GreaterOrEqual(t, len(allCookies), 1)
	foundAuth := false
	for _, cookie := range allCookies {
		if cookie.Name == authCookieName && cookie.Value == "oauth-token" {
			foundAuth = true
		}
	}
	require.True(t, foundAuth)
}

func TestAuthOAuthCallback_InvalidState_Unauthorized(t *testing.T) {
	t.Parallel()

	e, _ := newGetEvent("/auth/oauth/google/callback?state=bad&code=code123")
	e.Request.SetPathValue("provider", "google")
	e.Request.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "google:expected"})

	err := AuthOAuthCallback(&oauthAuthService{})(e)
	require.Error(t, err)
}

func TestAuthOAuthStart_UnsupportedProvider_BadRequest(t *testing.T) {
	t.Parallel()

	e, _ := newGetEvent("/auth/oauth/unknown")
	e.Request.SetPathValue("provider", "unknown")
	svc := &oauthAuthService{redirectErr: errors.New("unsupported")}
	err := AuthOAuthStart(svc)(e)
	require.Error(t, err)
}

func newGetEvent(path string) (*core.RequestEvent, *httptest.ResponseRecorder) {
	e, rec := newPostEvent(path, "")
	e.Request.Method = http.MethodGet
	return e, rec
}
