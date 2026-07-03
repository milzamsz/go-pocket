package handlers

import (
	"net/http"

	orgpage "github.com/milzamsz/go-pocket/components/pages/org"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/products"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func ListOrgMembers(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		members, err := tenancySvc.ListMembers(e.Request.Context(), orgCtx.OrgID, orgCtx.Role)
		if err != nil {
			return e.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return renderHTML(e, http.StatusOK, orgpage.Members(shell, orgCtx.Slug, members))
	}
}

func ShowOrgMemberProfile(tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		profile, err := tenancySvc.GetMemberProfile(e.Request.Context(), orgCtx.OrgID, orgCtx.Role, e.Request.PathValue("userID"))
		if err != nil {
			return e.BadRequestError("failed to load member profile", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.MemberProfile(shell, orgCtx.Slug, profile))
	}
}

func ShowProductDetail(productsSvc products.Service, tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}
		product, err := productsSvc.Get(e.Request.Context(), orgCtx.OrgID, e.Request.PathValue("id"))
		if err != nil {
			return e.BadRequestError("failed to load product", err)
		}
		return renderHTML(e, http.StatusOK, orgpage.ProductDetail(shell, orgCtx.Slug, product))
	}
}
