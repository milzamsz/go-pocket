package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/kanban"
	"github.com/milzamsz/go-pocket/internal/services/products"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestProtectedRoutes_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	app, mux := newRouteTestServer(t)
	createOrganization(t, app, "owner@example.com", "acme")

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "app dashboard", method: http.MethodGet, path: "/app/"},
		{name: "org members", method: http.MethodGet, path: "/org/acme/members"},
		{name: "org member profile", method: http.MethodGet, path: "/org/acme/members/user-1"},
		{name: "org product detail", method: http.MethodGet, path: "/org/acme/products/prod-1"},
		{name: "org invite", method: http.MethodPost, path: "/org/acme/members/invite"},
		{name: "invite accept", method: http.MethodPost, path: "/invite/token-test"},
		{name: "admin users", method: http.MethodGet, path: "/admin/users"},
		{name: "admin audit", method: http.MethodGet, path: "/admin/audit"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			status, _ := performRequest(t, mux, tc.method, tc.path, "")
			require.Equal(t, http.StatusUnauthorized, status)
		})
	}
}

func TestProtectedRoutes_InsufficientRole_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	app, mux := newRouteTestServer(t)
	orgID := createOrganization(t, app, "owner@example.com", "acme")
	viewer := createMembership(t, app, orgID, "viewer@example.com", domain.OrgRoleViewer)
	viewerToken := newAuthToken(t, viewer)
	nonSuperuser := createAuthUser(t, app, "member@example.com")
	nonSuperuserToken := newAuthToken(t, nonSuperuser)

	t.Run("org invite forbidden for viewer", func(t *testing.T) {
		status, _ := performRequest(t, mux, http.MethodPost, "/org/acme/members/invite", viewerToken)
		require.Equal(t, http.StatusForbidden, status)
	})

	t.Run("admin users forbidden for regular user", func(t *testing.T) {
		status, _ := performRequest(t, mux, http.MethodGet, "/admin/users", nonSuperuserToken)
		require.Equal(t, http.StatusForbidden, status)
	})
}

func TestProtectedRoutes_AuthorizedRole_AllowsAccess(t *testing.T) {
	t.Parallel()

	app, mux := newRouteTestServer(t)
	orgID := createOrganization(t, app, "owner@example.com", "acme")
	memberViewer := createMembership(t, app, orgID, "viewer@example.com", domain.OrgRoleViewer)
	memberAdmin := createMembership(t, app, orgID, "admin@example.com", domain.OrgRoleAdmin)
	superuser := createSuperuser(t, app, "root@example.com")

	viewerToken := newAuthToken(t, memberViewer)
	adminToken := newAuthToken(t, memberAdmin)
	superuserToken := newAuthToken(t, superuser)
	inviteToken := createInvitationTokenRecord(t, app, orgID, "viewer@example.com", domain.OrgRoleMember)

	t.Run("org members allowed for viewer", func(t *testing.T) {
		status, body := performRequest(t, mux, http.MethodGet, "/org/acme/members", viewerToken)
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Organization Members")
		require.Contains(t, body, "Members Workspace")
	})
	t.Run("help center is public", func(t *testing.T) {
		status, body := performRequest(t, mux, http.MethodGet, "/help", "")
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Help Center")
	})
	t.Run("app dashboard allowed for authenticated user", func(t *testing.T) {
		status, _ := performRequest(t, mux, http.MethodGet, "/app/", viewerToken)
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("org invite allowed for admin", func(t *testing.T) {
		form := url.Values{
			"email": {"newmember@example.com"},
			"role":  {string(domain.OrgRoleMember)},
		}
		status, _ := performFormRequest(t, mux, http.MethodPost, "/org/acme/members/invite", adminToken, form)
		require.Equal(t, http.StatusSeeOther, status)
	})
	t.Run("org product detail allowed for viewer", func(t *testing.T) {
		productID := createProductRecord(t, app, orgID, "Neon Widget")
		status, body := performRequest(t, mux, http.MethodGet, "/org/acme/products/"+productID, viewerToken)
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Product Detail")
		require.Contains(t, body, "Neon Widget")
	})
	t.Run("org member profile allowed for viewer", func(t *testing.T) {
		status, body := performRequest(t, mux, http.MethodGet, "/org/acme/members/"+memberViewer.Id, viewerToken)
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Member Profile")
		require.Contains(t, body, memberViewer.GetString("email"))
	})

	t.Run("admin users allowed for superuser", func(t *testing.T) {
		status, body := performRequest(t, mux, http.MethodGet, "/admin/users", superuserToken)
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Users")
	})
	t.Run("admin audit allowed for superuser", func(t *testing.T) {
		status, body := performRequest(t, mux, http.MethodGet, "/admin/audit", superuserToken)
		require.Equal(t, http.StatusOK, status)
		require.Contains(t, body, "Admin Audit")
	})
	t.Run("invite accept allowed for authenticated user", func(t *testing.T) {
		status, _ := performRequest(t, mux, http.MethodPost, "/invite/"+inviteToken, viewerToken)
		require.Equal(t, http.StatusSeeOther, status)
	})
}

func newRouteTestServer(t *testing.T) (core.App, http.Handler) {
	t.Helper()

	app := testutil.NewTestApp(t)

	pbRouter, err := apis.NewRouter(app)
	require.NoError(t, err)

	serveEvent := &core.ServeEvent{
		App:    app,
		Router: pbRouter,
	}

	deps := &Deps{
		Config:  config.Config{},
		Auth:    auth.New(app),
		Billing: billing.New(config.Config{}),
		Email:   email.New(config.Config{}),
		Tenancy: tenancy.New(tenancy.NewPocketBaseRepository(app)),
		Products: products.New(app),
		Kanban:   kanban.New(app),
	}
	RegisterRoutes(serveEvent, deps)

	var mux http.Handler
	mux, err = serveEvent.Router.BuildMux()
	require.NoError(t, err)

	return app, mux
}

func performRequest(t *testing.T, handler http.Handler, method string, path string, token string) (int, string) {
	t.Helper()

	reqBody := bytes.NewBuffer(nil)
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut {
		reqBody = bytes.NewBufferString(`{}`)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res.StatusCode, string(respBody)
}

func performFormRequest(t *testing.T, handler http.Handler, method string, path string, token string, form url.Values) (int, string) {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res.StatusCode, string(respBody)
}

func createOrganization(t *testing.T, app core.App, ownerEmail string, slug string) string {
	t.Helper()

	owner := createAuthUser(t, app, ownerEmail)

	orgs, err := app.FindCollectionByNameOrId("organizations")
	require.NoError(t, err)

	org := core.NewRecord(orgs)
	org.Set("slug", slug)
	org.Set("name", "Acme Org")
	org.Set("owner", owner.Id)
	require.NoError(t, app.Save(org))

	return org.Id
}

func createMembership(t *testing.T, app core.App, organizationID string, email string, role domain.OrgRole) *core.Record {
	t.Helper()

	user := createAuthUser(t, app, email)

	members, err := app.FindCollectionByNameOrId("organization_members")
	require.NoError(t, err)

	member := core.NewRecord(members)
	member.Set("organization", organizationID)
	member.Set("user", user.Id)
	member.Set("role", string(role))
	require.NoError(t, app.Save(member))

	return user
}

func createAuthUser(t *testing.T, app core.App, email string) *core.Record {
	t.Helper()

	users, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(users)
	user.Set("email", email)
	user.Set("password", "test-password-123")
	user.Set("passwordConfirm", "test-password-123")
	require.NoError(t, app.Save(user))

	return user
}

func createSuperuser(t *testing.T, app core.App, email string) *core.Record {
	t.Helper()

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	require.NoError(t, err)

	user := core.NewRecord(superusers)
	user.Set("email", email)
	user.Set("password", "test-password-123")
	user.Set("passwordConfirm", "test-password-123")
	require.NoError(t, app.Save(user))

	return user
}

func createInvitationTokenRecord(t *testing.T, app core.App, orgID string, email string, role domain.OrgRole) string {
	t.Helper()
	invitations, err := app.FindCollectionByNameOrId("invitations")
	require.NoError(t, err)
	record := core.NewRecord(invitations)
	record.Set("organization", orgID)
	record.Set("email", email)
	record.Set("role", string(role))
	record.Set("token", "token-test")
	require.NoError(t, app.Save(record))
	return "token-test"
}

func createProductRecord(t *testing.T, app core.App, orgID string, name string) string {
	t.Helper()
	products, err := app.FindCollectionByNameOrId("products")
	require.NoError(t, err)
	record := core.NewRecord(products)
	record.Set("organization", orgID)
	record.Set("name", name)
	record.Set("category", "electronics")
	record.Set("price", 12999)
	record.Set("stock", 20)
	record.Set("active", true)
	require.NoError(t, app.Save(record))
	return record.Id
}

func newAuthToken(t *testing.T, record *core.Record) string {
	t.Helper()

	token, err := record.NewAuthToken()
	require.NoError(t, err)
	return token
}
