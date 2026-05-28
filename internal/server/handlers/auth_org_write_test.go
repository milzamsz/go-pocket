package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/stretchr/testify/require"
)

type authOrgService struct {
	loginErr  error
	signupErr error
	logoutErr error
	token     string
	logoutID  string
}

func (m *authOrgService) CurrentUser(context.Context) (*domain.User, error) { return nil, nil }
func (m *authOrgService) Login(context.Context, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{ID: "user-1"}, Token: m.token}, m.loginErr
}
func (m *authOrgService) Signup(context.Context, string, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{ID: "user-1"}, Token: m.token}, m.signupErr
}
func (m *authOrgService) Logout(_ context.Context, userID string) error {
	m.logoutID = userID
	return m.logoutErr
}
func (m *authOrgService) RequestPasswordReset(context.Context, string) error          { return nil }
func (m *authOrgService) ResetPassword(context.Context, string, string, string) error { return nil }
func (m *authOrgService) VerifyEmail(context.Context, string) error                   { return nil }
func (m *authOrgService) UpdateProfile(context.Context, string, string, string) error { return nil }
func (m *authOrgService) ChangePassword(context.Context, string, string, string, string) error {
	return nil
}
func (m *authOrgService) UpdateTwoFactor(context.Context, string, bool) error { return nil }
func (m *authOrgService) OAuthRedirectURL(context.Context, string, string) (string, error) {
	return "", nil
}
func (m *authOrgService) OAuthCallback(context.Context, string, string) (*auth.AuthSession, error) {
	return nil, nil
}

func TestAuthLogin_SetsAuthCookieAndRedirects(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/login", "email=user%40example.com&password=secret123")
	svc := &authOrgService{token: "token-abc"}

	err := AuthLogin(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app", rec.Header().Get("Location"))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, authCookieName, cookies[0].Name)
	require.Equal(t, "token-abc", cookies[0].Value)
	require.True(t, cookies[0].HttpOnly)
}

func TestAuthSignup_SetsAuthCookieAndRedirects(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/signup", "name=Jane&email=jane%40example.com&password=secret123")
	svc := &authOrgService{token: "token-signup"}

	err := AuthSignup(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app", rec.Header().Get("Location"))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, authCookieName, cookies[0].Name)
	require.Equal(t, "token-signup", cookies[0].Value)
}

func TestAuthLogout_ClearsCookieAndRedirects(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/logout", "")
	middleware.SetActorContext(e, middleware.ActorContext{UserID: "user-1"})
	svc := &authOrgService{}

	err := AuthLogout(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
	require.Equal(t, "user-1", svc.logoutID)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, authCookieName, cookies[0].Name)
	require.Equal(t, "", cookies[0].Value)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestAuthLogin_Failed_DoesNotSetCookie(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/login", "email=user%40example.com&password=wrong")
	svc := &authOrgService{loginErr: errors.New("invalid credentials"), token: "ignored"}

	err := AuthLogin(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Empty(t, rec.Result().Cookies())
}
