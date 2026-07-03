package handlers

import (
	"net/http"

	apppage "github.com/milzamsz/go-pocket/components/pages/app"
	orgpage "github.com/milzamsz/go-pocket/components/pages/org"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func AppDashboard() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		shell, err := appShellState(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return renderHTML(e, http.StatusOK, apppage.Dashboard(shell))
	}
}

func AppOnboarding() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		shell, err := appShellState(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return renderHTML(e, http.StatusOK, apppage.Onboarding(shell))
	}
}

func AppSettingsProfilePage() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		shell, err := appShellState(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return renderHTML(e, http.StatusOK, apppage.SettingsProfile(shell))
	}
}

func AppSettingsSecurityPage() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		shell, err := appShellState(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return renderHTML(e, http.StatusOK, apppage.SettingsSecurity(shell))
	}
}

func AppSettingsAccountPage() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		shell, err := appShellState(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}
		return renderHTML(e, http.StatusOK, apppage.SettingsAccount(shell))
	}
}

func OrgOverview(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.Overview(shell, orgCtx.Slug))
	}
}

func OrgInvitations(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.Invitations(shell, orgCtx.Slug))
	}
}

func OrgBilling(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.Billing(shell, orgCtx.Slug))
	}
}

func OrgSettingsPage(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.Settings(shell, orgCtx.Slug))
	}
}

func OrgSettingsDangerPage(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.SettingsDanger(shell, orgCtx.Slug))
	}
}

func OrgAudit(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.Audit(shell, orgCtx.Slug))
	}
}
