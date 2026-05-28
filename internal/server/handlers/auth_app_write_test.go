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

type writeAuthService struct {
	forgotErr     error
	resetErr      error
	verifyErr     error
	profileErr    error
	lastForgot    string
	lastResetTok  string
	lastVerifyTok string
	lastProfileID string
}

func (m *writeAuthService) CurrentUser(context.Context) (*domain.User, error) { return nil, nil }
func (m *writeAuthService) Login(context.Context, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{}, Token: "token"}, nil
}
func (m *writeAuthService) Signup(context.Context, string, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{}, Token: "token"}, nil
}
func (m *writeAuthService) Logout(context.Context, string) error { return nil }
func (m *writeAuthService) RequestPasswordReset(_ context.Context, email string) error {
	m.lastForgot = email
	return m.forgotErr
}
func (m *writeAuthService) ResetPassword(_ context.Context, token string, _ string, _ string) error {
	m.lastResetTok = token
	return m.resetErr
}
func (m *writeAuthService) VerifyEmail(_ context.Context, token string) error {
	m.lastVerifyTok = token
	return m.verifyErr
}
func (m *writeAuthService) UpdateProfile(_ context.Context, userID string, _ string, _ string) error {
	m.lastProfileID = userID
	return m.profileErr
}
func (m *writeAuthService) ChangePassword(context.Context, string, string, string, string) error {
	return nil
}
func (m *writeAuthService) UpdateTwoFactor(context.Context, string, bool) error { return nil }
func (m *writeAuthService) OAuthRedirectURL(context.Context, string, string) (string, error) {
	return "", nil
}
func (m *writeAuthService) OAuthCallback(context.Context, string, string) (*auth.AuthSession, error) {
	return nil, nil
}

func TestAuthForgotPassword_RedirectsOnSuccess(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/forgot-password", "email=user%40example.com")
	svc := &writeAuthService{}

	err := AuthForgotPassword(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/auth/login", rec.Header().Get("Location"))
	require.Equal(t, "user@example.com", svc.lastForgot)
}

func TestAuthResetPassword_UsesTokenAndRedirectsOnSuccess(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/reset-password?token=token-123", "password=secret123&confirm_password=secret123")
	svc := &writeAuthService{}

	err := AuthResetPassword(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/auth/login", rec.Header().Get("Location"))
	require.Equal(t, "token-123", svc.lastResetTok)
}

func TestAppSettingsProfile_RedirectsBackToLoginWhenUnauthenticated(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/app/settings/profile", "name=Jane&email=jane%40example.com")

	err := AppSettingsProfile(&writeAuthService{})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/auth/login", rec.Header().Get("Location"))
}

func TestAppSettingsProfile_CallsServiceWithActorAndRedirects(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/app/settings/profile", "name=Jane&email=jane%40example.com")
	middleware.SetActorContext(e, middleware.ActorContext{UserID: "user-1", Email: "jane@example.com"})
	svc := &writeAuthService{}

	err := AppSettingsProfile(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app/settings/profile", rec.Header().Get("Location"))
	require.Equal(t, "user-1", svc.lastProfileID)
}

func TestAuthForgotPassword_RendersBadRequestOnValidationError(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/forgot-password", "")
	svc := &writeAuthService{forgotErr: errors.New("email is required")}

	err := AuthForgotPassword(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthVerifyEmail_RedirectsOnSuccess(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/verify-email?token=verify_abcDEF12", "")
	svc := &writeAuthService{}

	err := AuthVerifyEmail(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app", rec.Header().Get("Location"))
	require.Equal(t, "verify_abcDEF12", svc.lastVerifyTok)
}

func TestAuthVerifyEmail_RendersBadRequestOnValidationError(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/auth/verify-email?token=bad", "")
	svc := &writeAuthService{verifyErr: errors.New("invalid verification token")}

	err := AuthVerifyEmail(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
