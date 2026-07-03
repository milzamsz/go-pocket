package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/require"
)

type mockBillingService struct {
	checkoutURL string
	portalURL   string
	webhookErr  error
	bindCalls   int
}

func (m mockBillingService) CreateCheckoutSession(_ context.Context, _ string) (string, error) {
	return m.checkoutURL, nil
}
func (m mockBillingService) CreatePortalSession(_ context.Context, _ string) (string, error) {
	return m.portalURL, nil
}
func (m mockBillingService) VerifyAndDispatchWebhook(_ context.Context, _ http.Header, _ []byte) error {
	return m.webhookErr
}
func (m *mockBillingService) BindWebhookEventStore(_ core.App) {
	m.bindCalls++
}

type mockEmailService struct {
	webhookErr   error
	lastInviteTo string
	boundStore   bool
}

func (m *mockEmailService) SendInvite(_ context.Context, to string, _ string) error {
	m.lastInviteTo = to
	return nil
}
func (m mockEmailService) SendWelcome(_ context.Context, _ string) error { return nil }
func (m mockEmailService) SendPasswordReset(_ context.Context, _ string, _ string) error {
	return nil
}
func (m mockEmailService) SendEmailVerification(_ context.Context, _ string, _ string) error {
	return nil
}
func (m mockEmailService) SendOrgSettingsUpdated(_ context.Context, _ string, _ string) error {
	return nil
}
func (m mockEmailService) VerifyAndDispatchWebhook(_ context.Context, _ http.Header, _ []byte) error {
	return m.webhookErr
}
func (m *mockEmailService) BindWebhookEventStore(_ core.App) {
	m.boundStore = true
}

type mockAuthService struct{ loginErr error }

func (m mockAuthService) CurrentUser(context.Context) (*domain.User, error) { return nil, nil }
func (m mockAuthService) Login(context.Context, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{}, Token: "token"}, m.loginErr
}
func (m mockAuthService) Signup(context.Context, string, string, string) (*auth.AuthSession, error) {
	return &auth.AuthSession{User: &domain.User{}, Token: "token"}, nil
}
func (m mockAuthService) Logout(context.Context, string) error { return nil }
func (m mockAuthService) RequestPasswordReset(context.Context, string) error {
	return nil
}
func (m mockAuthService) ResetPassword(context.Context, string, string, string) error {
	return nil
}
func (m mockAuthService) VerifyEmail(context.Context, string) error { return nil }
func (m mockAuthService) UpdateProfile(context.Context, string, string, string) error {
	return nil
}
func (m mockAuthService) ChangePassword(context.Context, string, string, string, string) error {
	return nil
}
func (m mockAuthService) UpdateTwoFactor(context.Context, string, bool) error { return nil }
func (m mockAuthService) OAuthRedirectURL(context.Context, string, string) (string, error) {
	return "", nil
}
func (m mockAuthService) OAuthCallback(context.Context, string, string) (*auth.AuthSession, error) {
	return nil, nil
}

type mockTenancyService struct{ inviteOrgID string }

func (m *mockTenancyService) CreateOrganization(context.Context, domain.Organization, string) (domain.Organization, error) {
	return domain.Organization{}, nil
}
func (m *mockTenancyService) ListMembers(context.Context, string, domain.OrgRole) ([]domain.OrganizationMember, error) {
	return nil, nil
}
func (m *mockTenancyService) GetMemberProfile(context.Context, string, domain.OrgRole, string) (domain.OrganizationMemberProfile, error) {
	return domain.OrganizationMemberProfile{}, nil
}
func (m *mockTenancyService) GetOrganizationShell(context.Context, string, domain.OrgRole) (tenancy.OrganizationShell, error) {
	return tenancy.OrganizationShell{}, nil
}
func (m *mockTenancyService) InviteMember(_ context.Context, orgID string, _ domain.OrgRole, _ string, role domain.OrgRole) (domain.Invitation, error) {
	m.inviteOrgID = orgID
	return domain.Invitation{OrganizationID: orgID, Email: "user@example.com", Role: role}, nil
}
func (m *mockTenancyService) RemoveMember(context.Context, string, domain.OrgRole, string) error {
	return nil
}
func (m *mockTenancyService) ChangeMemberRole(context.Context, string, domain.OrgRole, string, domain.OrgRole) error {
	return nil
}
func (m *mockTenancyService) ResendInvitation(context.Context, string, domain.OrgRole, string) error {
	return nil
}
func (m *mockTenancyService) RevokeInvitation(context.Context, string, domain.OrgRole, string) error {
	return nil
}
func (m *mockTenancyService) UpdateSettings(context.Context, string, domain.OrgRole, string) error {
	return nil
}
func (m *mockTenancyService) TransferOwnership(context.Context, string, domain.OrgRole, string) error {
	return nil
}
func (m *mockTenancyService) DeleteOrganization(context.Context, string, domain.OrgRole) error {
	return nil
}
func (m *mockTenancyService) AcceptInvitation(context.Context, string, string) (domain.Invitation, error) {
	return domain.Invitation{OrganizationID: "org-1"}, nil
}
func (m *mockTenancyService) DeclineInvitation(context.Context, string) (domain.Invitation, error) {
	return domain.Invitation{OrganizationID: "org-1"}, nil
}

func TestOrgBillingCheckout_RedirectsToSessionURL(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/org/acme/billing/checkout", "")
	err := OrgBillingCheckout(mockBillingService{checkoutURL: "https://polar.example.com/checkout"})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "https://polar.example.com/checkout", rec.Header().Get("Location"))
}

func TestOrgBillingPortal_RedirectsToSessionURL(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/org/acme/billing/portal", "")
	err := OrgBillingPortal(mockBillingService{portalURL: "https://polar.example.com/portal"})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "https://polar.example.com/portal", rec.Header().Get("Location"))
}

func TestPolarWebhook_InvalidSignatureReturnsJSONUnauthorized(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/webhooks/polar", `{"type":"test"}`)
	err := PolarWebhook(&mockBillingService{webhookErr: billing.ErrInvalidWebhookSignature})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), `"error":"invalid signature"`)
}

func TestPolarWebhook_BindsWebhookStoreWhenAvailable(t *testing.T) {
	t.Parallel()

	svc := &mockBillingService{}
	e, rec := newPostEvent("/webhooks/polar", `{"type":"subscription.updated"}`)
	err := PolarWebhook(svc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, svc.bindCalls)
}

func TestResendWebhook_InvalidSignatureReturnsJSONUnauthorized(t *testing.T) {
	t.Parallel()

	e, rec := newPostEvent("/webhooks/resend", `{"type":"email.delivered"}`)
	err := ResendWebhook(&mockEmailService{webhookErr: email.ErrInvalidWebhookSignature})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), `"error":"invalid signature"`)
}

func TestResendWebhook_BindsWebhookStoreWhenSupported(t *testing.T) {
	t.Parallel()

	emailSvc := &mockEmailService{}
	e, rec := newPostEvent("/webhooks/resend", `{"type":"email.delivered"}`)

	err := ResendWebhook(emailSvc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, emailSvc.boundStore)
}

func TestAuthLogin_RedirectsOnSuccess(t *testing.T) {
	t.Parallel()
	e, rec := newPostEvent("/auth/login", "email=a%40b.com&password=secret")

	err := AuthLogin(mockAuthService{})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/app", rec.Header().Get("Location"))
}

func TestOrgMembersInvite_UsesTenancyAndEmail(t *testing.T) {
	t.Parallel()
	e, rec := newPostEvent("/org/acme/members/invite", "email=user%40example.com&role=member")
	middleware.SetOrgContext(e, middleware.OrgContext{OrgID: "org-1", Slug: "acme", Role: domain.OrgRoleAdmin})

	tenancySvc := &mockTenancyService{}
	emailSvc := &mockEmailService{}

	err := OrgMembersInvite(tenancySvc, emailSvc)(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/org/acme/invitations", rec.Header().Get("Location"))
	require.Equal(t, "org-1", tenancySvc.inviteOrgID)
	require.Equal(t, "user@example.com", emailSvc.lastInviteTo)
}

func TestInviteAccept_RedirectsToOrganization(t *testing.T) {
	t.Parallel()
	e, rec := newPostEvent("/invite/token123", "")
	e.Request.SetPathValue("token", "token123")
	middleware.SetActorContext(e, middleware.ActorContext{UserID: "user-1"})

	err := InviteAccept(&mockTenancyService{})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/org/org-1/", rec.Header().Get("Location"))
}

func TestInviteDecline_RedirectsToInvitations(t *testing.T) {
	t.Parallel()
	e, rec := newPostEvent("/invite/token123/decline", "")
	e.Request.SetPathValue("token", "token123")

	err := InviteDecline(&mockTenancyService{})(e)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/org/org-1/invitations", rec.Header().Get("Location"))
}

func newPostEvent(path string, body string) (*core.RequestEvent, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "acme")
	rec := httptest.NewRecorder()
	return &core.RequestEvent{
		Event: router.Event{
			Response: rec,
			Request:  req,
		},
	}, rec
}
