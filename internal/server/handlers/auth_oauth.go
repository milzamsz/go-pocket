package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/pocketbase/pocketbase/core"
)

const oauthStateCookieName = "oauth_state"

func AuthOAuthStart(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		provider := strings.ToLower(strings.TrimSpace(e.Request.PathValue("provider")))
		state, err := generateOAuthState()
		if err != nil {
			return e.InternalServerError("failed to start oauth flow", err)
		}
		redirectURL, err := authSvc.OAuthRedirectURL(e.Request.Context(), provider, state)
		if err != nil {
			return e.BadRequestError("unsupported oauth provider", err)
		}
		http.SetCookie(e.Response, &http.Cookie{
			Name:     oauthStateCookieName,
			Value:    provider + ":" + state,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   strings.EqualFold(e.Request.URL.Scheme, "https") || e.Request.TLS != nil,
			MaxAge:   int((10 * time.Minute).Seconds()),
		})
		return e.Redirect(http.StatusFound, redirectURL)
	}
}

func AuthOAuthCallback(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		provider := strings.ToLower(strings.TrimSpace(e.Request.PathValue("provider")))
		code := strings.TrimSpace(e.Request.URL.Query().Get("code"))
		state := strings.TrimSpace(e.Request.URL.Query().Get("state"))
		if err := verifyOAuthState(e, provider, state); err != nil {
			return e.UnauthorizedError("invalid oauth state", err)
		}
		clearOAuthStateCookie(e)
		session, err := authSvc.OAuthCallback(e.Request.Context(), provider, code)
		if err != nil {
			return renderHTML(e, http.StatusUnauthorized, simpleAcceptedPage("auth-oauth-callback: failed"))
		}
		setAuthCookie(e, session.Token)
		return e.Redirect(http.StatusSeeOther, "/app")
	}
}

func generateOAuthState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func verifyOAuthState(e *core.RequestEvent, provider string, state string) error {
	if strings.TrimSpace(state) == "" {
		return errors.New("missing state")
	}
	cookie, err := e.Request.Cookie(oauthStateCookieName)
	if err != nil || cookie == nil {
		return errors.New("missing oauth state cookie")
	}
	value := strings.TrimSpace(cookie.Value)
	expectedPrefix := provider + ":"
	if !strings.HasPrefix(value, expectedPrefix) {
		return errors.New("provider mismatch")
	}
	if strings.TrimPrefix(value, expectedPrefix) != state {
		return errors.New("state mismatch")
	}
	return nil
}

func clearOAuthStateCookie(e *core.RequestEvent) {
	http.SetCookie(e.Response, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.EqualFold(e.Request.URL.Scheme, "https") || e.Request.TLS != nil,
		MaxAge:   -1,
	})
}
