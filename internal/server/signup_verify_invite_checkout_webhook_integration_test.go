package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestSignupVerifyInviteCheckoutAndWebhookUpdates_PersistsExpectedState(t *testing.T) {
	app, mux := newE2ETestServer(t)

	signupForm := url.Values{
		"name":     {"Worker Three"},
		"email":    {"worker3@example.com"},
		"password": {"pass-1234-ABCD"},
	}
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(signupForm.Encode()))
	signupReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	signupRec := httptest.NewRecorder()
	mux.ServeHTTP(signupRec, signupReq)
	require.Equal(t, http.StatusSeeOther, signupRec.Code)
	require.Equal(t, "/app", signupRec.Header().Get("Location"))

	signupCookies := signupRec.Result().Cookies()
	require.NotEmpty(t, signupCookies)
	authToken := findCookieValue(signupCookies, "pb_auth")
	require.NotEmpty(t, authToken)

	userRecord, err := app.FindAuthRecordByEmail("users", "worker3@example.com")
	require.NoError(t, err)
	require.False(t, userRecord.Verified())

	verifyTokenRecord, err := app.FindFirstRecordByFilter(
		"auth_tokens",
		"email = {:email} && kind = {:kind}",
		dbx.Params{"email": "worker3@example.com", "kind": "verify"},
	)
	require.NoError(t, err)
	verifyToken := verifyTokenRecord.GetString("token")
	require.True(t, strings.HasPrefix(verifyToken, "verify_"))
	require.True(t, verifyTokenRecord.GetDateTime("consumed_at").IsZero())

	verifyReq := httptest.NewRequest(http.MethodGet, "/auth/verify-email?token="+verifyToken, nil)
	verifyRec := httptest.NewRecorder()
	mux.ServeHTTP(verifyRec, verifyReq)
	require.Equal(t, http.StatusSeeOther, verifyRec.Code)
	require.Equal(t, "/app", verifyRec.Header().Get("Location"))

	userRecord, err = app.FindAuthRecordByEmail("users", "worker3@example.com")
	require.NoError(t, err)
	require.True(t, userRecord.Verified())

	verifyTokenRecord, err = app.FindFirstRecordByFilter(
		"auth_tokens",
		"token = {:token} && kind = {:kind}",
		dbx.Params{"token": verifyToken, "kind": "verify"},
	)
	require.NoError(t, err)
	require.False(t, verifyTokenRecord.GetDateTime("consumed_at").IsZero())

	orgID := createOrganization(t, app, "owner@example.com", "acme")
	membersCollection, err := app.FindCollectionByNameOrId("organization_members")
	require.NoError(t, err)
	memberRecord := core.NewRecord(membersCollection)
	memberRecord.Set("organization", orgID)
	memberRecord.Set("user", userRecord.Id)
	memberRecord.Set("role", string(domain.OrgRoleAdmin))
	require.NoError(t, app.Save(memberRecord))

	inviteForm := url.Values{
		"email": {"newhire@example.com"},
		"role":  {string(domain.OrgRoleMember)},
	}
	inviteReq := httptest.NewRequest(http.MethodPost, "/org/acme/members/invite", strings.NewReader(inviteForm.Encode()))
	inviteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	inviteReq.Header.Set("Authorization", "Bearer "+authToken)
	inviteRec := httptest.NewRecorder()
	mux.ServeHTTP(inviteRec, inviteReq)
	require.Equal(t, http.StatusSeeOther, inviteRec.Code)
	require.Equal(t, "/org/acme/invitations", inviteRec.Header().Get("Location"))

	inviteRecord, err := app.FindFirstRecordByFilter(
		"invitations",
		"organization = {:orgID} && email = {:email}",
		dbx.Params{"orgID": orgID, "email": "newhire@example.com"},
	)
	require.NoError(t, err)
	require.Equal(t, string(domain.OrgRoleMember), inviteRecord.GetString("role"))
	require.NotEmpty(t, strings.TrimSpace(inviteRecord.GetString("token")))

	checkoutReq := httptest.NewRequest(http.MethodPost, "/org/acme/billing/checkout", nil)
	checkoutReq.Header.Set("Authorization", "Bearer "+authToken)
	checkoutRec := httptest.NewRecorder()
	mux.ServeHTTP(checkoutRec, checkoutReq)
	require.Equal(t, http.StatusSeeOther, checkoutRec.Code)
	require.Equal(t, "https://app.example.test/org/acme/billing?provider=unconfigured", checkoutRec.Header().Get("Location"))

	polarPayload := []byte(`{
		"type":"subscription.updated",
		"data":{
			"id":"evt_polar_e2e_001",
			"status":"active",
			"customer":{"id":"cus_e2e_001","external_id":"acme"},
			"subscription":{"id":"sub_e2e_001"},
			"product":{"id":"prod_e2e_001"},
			"price":{"id":"price_e2e_001"}
		}
	}`)
	polarReq := httptest.NewRequest(http.MethodPost, "/webhooks/polar", bytes.NewReader(polarPayload))
	polarRec := httptest.NewRecorder()
	mux.ServeHTTP(polarRec, polarReq)
	require.Equal(t, http.StatusOK, polarRec.Code)

	polarRetryReq := httptest.NewRequest(http.MethodPost, "/webhooks/polar", bytes.NewReader(polarPayload))
	polarRetryRec := httptest.NewRecorder()
	mux.ServeHTTP(polarRetryRec, polarRetryReq)
	require.Equal(t, http.StatusOK, polarRetryRec.Code)

	resendPayload := []byte(`{
		"type":"email.delivered",
		"created_at":"2026-05-27T08:45:00Z",
		"data":{"email_id":"msg_e2e_001","to":["worker3@example.com"]}
	}`)
	resendReq := httptest.NewRequest(http.MethodPost, "/webhooks/resend", bytes.NewReader(resendPayload))
	resendRec := httptest.NewRecorder()
	mux.ServeHTTP(resendRec, resendReq)
	require.Equal(t, http.StatusOK, resendRec.Code)

	resendRetryReq := httptest.NewRequest(http.MethodPost, "/webhooks/resend", bytes.NewReader(resendPayload))
	resendRetryRec := httptest.NewRecorder()
	mux.ServeHTTP(resendRetryRec, resendRetryReq)
	require.Equal(t, http.StatusOK, resendRetryRec.Code)

	polarWebhookRecord, err := app.FindFirstRecordByFilter(
		"webhook_events",
		"provider = {:provider} && event_type = {:eventType}",
		dbx.Params{"provider": "polar", "eventType": "subscription.updated"},
	)
	require.NoError(t, err)
	require.Equal(t, "subscription", polarWebhookRecord.GetString("family"))
	require.True(t, polarWebhookRecord.GetBool("handled"))

	resendWebhookRecord, err := app.FindFirstRecordByFilter(
		"webhook_events",
		"provider = {:provider} && event_type = {:eventType}",
		dbx.Params{"provider": "resend", "eventType": "email.delivered"},
	)
	require.NoError(t, err)
	require.Equal(t, "delivered", resendWebhookRecord.GetString("family"))
	require.True(t, resendWebhookRecord.GetBool("handled"))
	require.Equal(t, "delivered", resendWebhookRecord.GetString("status"))
	require.Equal(t, "msg_e2e_001", resendWebhookRecord.GetString("message_id"))
	require.Equal(t, "worker3@example.com", resendWebhookRecord.GetString("recipient"))

	updatedOrg, err := app.FindRecordById("organizations", orgID)
	require.NoError(t, err)
	require.Equal(t, "cus_e2e_001", updatedOrg.GetString("polar_customer_id"))
	require.Equal(t, "sub_e2e_001", updatedOrg.GetString("polar_subscription_id"))
	require.Equal(t, "prod_e2e_001", updatedOrg.GetString("polar_product_id"))
	require.Equal(t, "price_e2e_001", updatedOrg.GetString("polar_price_id"))
	require.Equal(t, "prod_e2e_001", updatedOrg.GetString("plan"))
	require.Equal(t, "active", updatedOrg.GetString("subscription_status"))

	polarEventCount, err := app.CountRecords(
		"webhook_events",
		dbx.HashExp{"provider": "polar", "event_type": "subscription.updated"},
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), polarEventCount)

	resendEventCount, err := app.CountRecords(
		"webhook_events",
		dbx.HashExp{"provider": "resend", "event_type": "email.delivered"},
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), resendEventCount)
}

func newE2ETestServer(t *testing.T) (core.App, http.Handler) {
	t.Helper()

	app := testutil.NewTestApp(t)

	pbRouter, err := apis.NewRouter(app)
	require.NoError(t, err)
	serveEvent := &core.ServeEvent{App: app, Router: pbRouter}

	deps := &Deps{
		Config: config.Config{
			BaseURL: "https://app.example.test",
		},
		Auth: auth.New(app),
		Billing: billing.New(config.Config{
			BaseURL: "https://app.example.test",
		}),
		Email:   email.New(config.Config{}),
		Tenancy: tenancy.New(tenancy.NewPocketBaseRepository(app)),
	}
	RegisterRoutes(serveEvent, deps)

	mux, err := serveEvent.Router.BuildMux()
	require.NoError(t, err)
	return app, mux
}

func findCookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
