package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

const authCookieName = "pb_auth"

func AuthLogin(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		session, err := authSvc.Login(e.Request.Context(), e.Request.FormValue("email"), e.Request.FormValue("password"))
		if err != nil {
			return renderHTML(e, http.StatusUnauthorized, simpleAcceptedPage("auth-login: failed"))
		}
		setAuthCookie(e, session.Token)
		return e.Redirect(http.StatusSeeOther, "/app")
	}
}

func AuthSignup(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		session, err := authSvc.Signup(e.Request.Context(), e.Request.FormValue("name"), e.Request.FormValue("email"), e.Request.FormValue("password"))
		if err != nil {
			return renderHTML(e, http.StatusBadRequest, simpleAcceptedPage("auth-signup: failed"))
		}
		setAuthCookie(e, session.Token)
		return e.Redirect(http.StatusSeeOther, "/app")
	}
}

func AuthLogout(authSvc auth.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, _ := middleware.CurrentActor(e)
		_ = authSvc.Logout(e.Request.Context(), actor.UserID)
		clearAuthCookie(e)
		return e.Redirect(http.StatusSeeOther, "/")
	}
}

func setAuthCookie(e *core.RequestEvent, token string) {
	if strings.TrimSpace(token) == "" {
		return
	}
	http.SetCookie(e.Response, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.EqualFold(e.Request.URL.Scheme, "https") || e.Request.TLS != nil,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
}

func clearAuthCookie(e *core.RequestEvent) {
	http.SetCookie(e.Response, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.EqualFold(e.Request.URL.Scheme, "https") || e.Request.TLS != nil,
		MaxAge:   -1,
	})
}

func OrgMembersInvite(tenancySvc tenancy.Service, emailSvc email.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		role := domain.OrgRole(strings.TrimSpace(e.Request.FormValue("role")))
		invite, err := tenancySvc.InviteMember(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, strings.TrimSpace(e.Request.FormValue("email")), role)
		if err != nil {
			return e.BadRequestError("failed to invite member", err)
		}
		if err := emailSvc.SendInvite(e.Request.Context(), invite.Email, orgCtx.Slug); err != nil {
			return e.InternalServerError("failed to send invite email", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/invitations")
	}
}

func OrgMembersRemove(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.RemoveMember(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, e.Request.PathValue("userID")); err != nil {
			return e.BadRequestError("failed to remove member", err)
		}
		return e.NoContent(http.StatusNoContent)
	}
}

func OrgMembersChangeRole(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		role := domain.OrgRole(strings.TrimSpace(e.Request.FormValue("role")))
		if err := tenancySvc.ChangeMemberRole(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, e.Request.PathValue("userID"), role); err != nil {
			return e.BadRequestError("failed to change role", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/members")
	}
}

func OrgInvitationResend(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.ResendInvitation(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, e.Request.PathValue("id")); err != nil {
			return e.BadRequestError("failed to resend invitation", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/invitations")
	}
}

func OrgInvitationRevoke(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.RevokeInvitation(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, e.Request.PathValue("id")); err != nil {
			return e.BadRequestError("failed to revoke invitation", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/invitations")
	}
}

func OrgSettingsTransfer(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.TransferOwnership(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, strings.TrimSpace(e.Request.FormValue("new_owner_user_id"))); err != nil {
			return e.BadRequestError("failed to transfer ownership", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/settings/danger")
	}
}

func OrgSettingsDelete(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.DeleteOrganization(e.Request.Context(), orgCtx.OrgID, orgCtx.Role); err != nil {
			return e.BadRequestError("failed to delete organization", err)
		}
		return e.Redirect(http.StatusSeeOther, "/app")
	}
}

func InviteAccept(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := middleware.CurrentActor(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		token := strings.TrimSpace(e.Request.PathValue("token"))
		invitation, err := tenancySvc.AcceptInvitation(e.Request.Context(), token, actor.UserID)
		if err != nil {
			return e.BadRequestError("failed to accept invitation", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+invitation.OrganizationID+"/")
	}
}

func InviteDecline(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.PathValue("token"))
		invitation, err := tenancySvc.DeclineInvitation(e.Request.Context(), token)
		if err != nil {
			return e.BadRequestError("failed to decline invitation", err)
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+invitation.OrganizationID+"/invitations")
	}
}
