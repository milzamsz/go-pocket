package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/require"
)

func TestBindAuthContext_ExtractsActorFromRequestAuth(t *testing.T) {
	e := newRequestEvent(t, "/org/acme/members")
	e.Auth = newAuthRecord(t, "user-1", "user1@example.com", false)

	err := BindAuthContext()(e)

	require.NoError(t, err)

	actor, err := CurrentActor(e)
	require.NoError(t, err)
	require.Equal(t, "user-1", actor.UserID)
	require.Equal(t, "user1@example.com", actor.Email)
	require.False(t, actor.IsSuperuser)
}

func TestRequireAuthenticated_RejectsMissingActor(t *testing.T) {
	e := newRequestEvent(t, "/org/acme")

	err := RequireAuthenticated()(e)

	apiErr := &router.ApiError{}
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 401, apiErr.Status)
}

func TestRequireOrgRole_UsesOrgContextAndAllowsWhenPermissionMatches(t *testing.T) {
	e := newRequestEvent(t, "/org/acme/members")
	SetActorContext(e, ActorContext{UserID: "user-1"})
	SetOrgContext(e, OrgContext{OrgID: "org-1", Slug: "acme", Role: domain.OrgRoleAdmin})

	err := RequireOrgRoleWithResolver(
		tenancy.PermissionMembersWrite,
		domain.OrgRoleAdmin,
		func(_ *core.RequestEvent, _ ActorContext) (OrgContext, error) {
			t.Fatal("resolver should not run when org context already exists")
			return OrgContext{}, nil
		},
	)(e)

	require.NoError(t, err)
}

func TestRequireOrgRole_DeniesWhenRoleInsufficient(t *testing.T) {
	e := newRequestEvent(t, "/org/acme/members")
	SetActorContext(e, ActorContext{UserID: "user-1"})
	SetOrgContext(e, OrgContext{OrgID: "org-1", Slug: "acme", Role: domain.OrgRoleViewer})

	err := RequireOrgRoleWithResolver(
		tenancy.PermissionMembersWrite,
		domain.OrgRoleAdmin,
		func(_ *core.RequestEvent, _ ActorContext) (OrgContext, error) {
			return OrgContext{}, nil
		},
	)(e)

	apiErr := &router.ApiError{}
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 403, apiErr.Status)
}

func TestRequireSuperuser_DeniesNonSuperuser(t *testing.T) {
	e := newRequestEvent(t, "/admin")
	SetActorContext(e, ActorContext{UserID: "user-1", IsSuperuser: false})

	err := RequireSuperuser()(e)

	apiErr := &router.ApiError{}
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 403, apiErr.Status)
}

func TestReadAuthTokenFromRequest_PrefersBearerToken(t *testing.T) {
	e := newRequestEvent(t, "/org/acme")
	e.Request.Header.Set("Authorization", "Bearer header-token")
	e.Request.AddCookie(&http.Cookie{Name: "pb_auth", Value: "cookie-token"})

	token := readAuthTokenFromRequest(e)

	require.Equal(t, "header-token", token)
}

func TestReadAuthTokenFromRequest_UsesCookieWhenBearerMissing(t *testing.T) {
	e := newRequestEvent(t, "/org/acme")
	e.Request.AddCookie(&http.Cookie{Name: "pb_auth", Value: "cookie-token"})

	token := readAuthTokenFromRequest(e)

	require.Equal(t, "cookie-token", token)
}

func newRequestEvent(t *testing.T, path string) *core.RequestEvent {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	req.SetPathValue("slug", "acme")

	return &core.RequestEvent{
		Event: router.Event{
			Response: httptest.NewRecorder(),
			Request:  req,
		},
	}
}

func newAuthRecord(t *testing.T, id string, email string, superuser bool) *core.Record {
	t.Helper()
	collectionName := "users"
	if superuser {
		collectionName = core.CollectionNameSuperusers
	}
	collection := core.NewAuthCollection(collectionName)
	record := core.NewRecord(collection)
	record.Id = id
	record.SetEmail(email)
	return record
}
