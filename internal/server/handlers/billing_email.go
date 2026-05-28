package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

type emailWebhookStoreBinder interface {
	BindWebhookEventStore(app core.App)
}

type billingWebhookStoreBinder interface {
	BindWebhookEventStore(app core.App)
}

func OrgBillingCheckout(billingSvc billing.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgSlug := e.Request.PathValue("slug")
		redirectURL, err := billingSvc.CreateCheckoutSession(e.Request.Context(), orgSlug)
		if err != nil {
			return e.InternalServerError("failed to create checkout session", err)
		}
		return e.Redirect(http.StatusSeeOther, redirectURL)
	}
}

func OrgBillingPortal(billingSvc billing.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgSlug := e.Request.PathValue("slug")
		redirectURL, err := billingSvc.CreatePortalSession(e.Request.Context(), orgSlug)
		if err != nil {
			return e.InternalServerError("failed to create billing portal session", err)
		}
		return e.Redirect(http.StatusSeeOther, redirectURL)
	}
}

func OrgSettingsUpdate(tenancySvc tenancy.Service, emailSvc email.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		if err := tenancySvc.UpdateSettings(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, strings.TrimSpace(e.Request.FormValue("name"))); err != nil {
			return e.BadRequestError("failed to update organization settings", err)
		}
		actor, err := middleware.CurrentActor(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		if strings.TrimSpace(actor.Email) != "" {
			if err := emailSvc.SendOrgSettingsUpdated(e.Request.Context(), actor.Email, orgCtx.Slug); err != nil {
				return e.InternalServerError("failed to send settings update email", err)
			}
		}
		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/settings")
	}
}

func PolarWebhook(billingSvc billing.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if binder, ok := billingSvc.(billingWebhookStoreBinder); ok {
			binder.BindWebhookEventStore(e.App)
		}
		payload, err := io.ReadAll(e.Request.Body)
		if err != nil {
			return e.BadRequestError("failed to read request body", err)
		}
		if err := billingSvc.VerifyAndDispatchWebhook(e.Request.Context(), e.Request.Header, payload); err != nil {
			if errors.Is(err, billing.ErrInvalidWebhookSignature) {
				return e.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
			}
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

func ResendWebhook(emailSvc email.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if binder, ok := emailSvc.(emailWebhookStoreBinder); ok {
			binder.BindWebhookEventStore(e.App)
		}
		payload, err := io.ReadAll(e.Request.Body)
		if err != nil {
			return e.BadRequestError("failed to read request body", err)
		}
		if err := emailSvc.VerifyAndDispatchWebhook(e.Request.Context(), e.Request.Header, payload); err != nil {
			if errors.Is(err, email.ErrInvalidWebhookSignature) {
				return e.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
			}
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}
