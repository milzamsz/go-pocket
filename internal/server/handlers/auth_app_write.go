package handlers

import (
	"net/http"
	"strings"

	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/pocketbase/pocketbase/core"
)

func AuthForgotPassword(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := authSvc.RequestPasswordReset(e.Request.Context(), strings.TrimSpace(e.Request.FormValue("email"))); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("auth-forgot-password: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}
}

func AuthResetPassword(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.FormValue("token"))
		if token == "" {
			token = strings.TrimSpace(e.Request.URL.Query().Get("token"))
		}
		if err := authSvc.ResetPassword(
			e.Request.Context(),
			token,
			e.Request.FormValue("password"),
			e.Request.FormValue("confirm_password"),
		); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("auth-reset-password: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}
}

func AuthVerifyEmail(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.URL.Query().Get("token"))
		if err := authSvc.VerifyEmail(e.Request.Context(), token); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("auth-verify-email: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/app")
	}
}

func AppSettingsProfile(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := middleware.CurrentActor(e)
		if err != nil {
			return e.Redirect(http.StatusSeeOther, "/auth/login")
		}
		if err := authSvc.UpdateProfile(
			e.Request.Context(),
			actor.UserID,
			strings.TrimSpace(e.Request.FormValue("name")),
			strings.TrimSpace(e.Request.FormValue("email")),
		); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("app-settings-profile: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/app/settings/profile")
	}
}

func AppSettingsSecurityPassword(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := middleware.CurrentActor(e)
		if err != nil {
			return e.Redirect(http.StatusSeeOther, "/auth/login")
		}
		if err := authSvc.ChangePassword(
			e.Request.Context(),
			actor.UserID,
			e.Request.FormValue("current_password"),
			e.Request.FormValue("new_password"),
			e.Request.FormValue("confirm_password"),
		); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("app-settings-security-password: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/app/settings/security")
	}
}

func AppSettingsSecurity2FA(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := middleware.CurrentActor(e)
		if err != nil {
			return e.Redirect(http.StatusSeeOther, "/auth/login")
		}

		enabledValue := strings.TrimSpace(strings.ToLower(e.Request.FormValue("enabled")))
		enabled := enabledValue == "1" || enabledValue == "true" || enabledValue == "on"
		if err := authSvc.UpdateTwoFactor(e.Request.Context(), actor.UserID, enabled); err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("app-settings-security-2fa: failed"))
		}
		return e.Redirect(http.StatusSeeOther, "/app/settings/security")
	}
}
