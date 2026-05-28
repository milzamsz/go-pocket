package middleware

import (
	"database/sql"
	"errors"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type OrgContext struct {
	OrgID string
	Slug  string
	Role  domain.OrgRole
}

const orgContextStoreKey = "orgContext"

type orgContextResolver func(*core.RequestEvent, ActorContext) (OrgContext, error)

func RequireOrgRole(permission tenancy.Permission, minRole domain.OrgRole) func(e *core.RequestEvent) error {
	return RequireOrgRoleWithResolver(permission, minRole, resolveOrgContextFromMembership)
}

func RequireOrgRoleWithResolver(permission tenancy.Permission, minRole domain.OrgRole, resolver orgContextResolver) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := CurrentActor(e)
		if err != nil {
			return e.UnauthorizedError("authentication required", err)
		}

		orgCtx, ok := GetOrgContext(e)
		if !ok {
			orgCtx, err = resolver(e, actor)
			if err != nil {
				if errors.Is(err, domain.ErrForbidden) {
					return e.ForbiddenError("organization access denied", err)
				}
				return e.InternalServerError("failed to resolve organization membership", err)
			}
			SetOrgContext(e, orgCtx)
		}

		if !roleAtLeast(orgCtx.Role, minRole) || !tenancy.Can(orgCtx.Role, permission) {
			return e.ForbiddenError("insufficient organization role", domain.ErrForbidden)
		}

		return e.Next()
	}
}

func SetOrgContext(e *core.RequestEvent, orgCtx OrgContext) {
	e.Set(orgContextStoreKey, orgCtx)
}

func GetOrgContext(e *core.RequestEvent) (OrgContext, bool) {
	orgCtx, ok := e.Get(orgContextStoreKey).(OrgContext)
	return orgCtx, ok
}

func resolveOrgContextFromMembership(e *core.RequestEvent, actor ActorContext) (OrgContext, error) {
	orgSlug := e.Request.PathValue("slug")
	if orgSlug == "" {
		return OrgContext{}, domain.ErrForbidden
	}

	org, err := e.App.FindFirstRecordByFilter("organizations", "slug = {:slug}", dbx.Params{"slug": orgSlug})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrgContext{}, domain.ErrForbidden
		}
		return OrgContext{}, err
	}

	member, err := e.App.FindFirstRecordByFilter(
		"organization_members",
		"organization = {:org} && user = {:user}",
		dbx.Params{"org": org.Id, "user": actor.UserID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrgContext{}, domain.ErrForbidden
		}
		return OrgContext{}, err
	}

	return OrgContext{
		OrgID: org.Id,
		Slug:  orgSlug,
		Role:  domain.OrgRole(member.GetString("role")),
	}, nil
}

var rolePriority = map[domain.OrgRole]int{
	domain.OrgRoleViewer: 1,
	domain.OrgRoleMember: 2,
	domain.OrgRoleAdmin:  3,
	domain.OrgRoleOwner:  4,
}

func roleAtLeast(actual domain.OrgRole, min domain.OrgRole) bool {
	return rolePriority[actual] >= rolePriority[min]
}
