package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/milzamsz/go-pocket/internal/config"
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

func TestWebhookEndpoints_InvalidSignatures_ReturnUnauthorizedAndNoStateMutation(t *testing.T) {
	t.Parallel()

	app, mux := newWebhookTestServer(t, config.Config{
		Polar:  config.PolarConfig{WebhookSecret: "polar-secret"},
		Resend: config.ResendConfig{WebhookSecret: "whsec_bXktdGVzdC1zZWNyZXQ="},
	})

	polarPayload := []byte(`{"type":"subscription.updated","data":{"id":"evt_unauth_1"}}`)
	polarReq := httptest.NewRequest(http.MethodPost, "/webhooks/polar", bytes.NewReader(polarPayload))
	polarReq.Header.Set("Polar-Signature", "t=1234567890,v1=invalid")
	polarRec := httptest.NewRecorder()
	mux.ServeHTTP(polarRec, polarReq)
	require.Equal(t, http.StatusUnauthorized, polarRec.Code)

	resendPayload := []byte(`{"type":"email.delivered","data":{"email_id":"msg_unauth_1"}}`)
	resendReq := httptest.NewRequest(http.MethodPost, "/webhooks/resend", bytes.NewReader(resendPayload))
	resendReq.Header.Set("svix-id", "msg_unauth_1")
	resendReq.Header.Set("svix-timestamp", "1234567890")
	resendReq.Header.Set("svix-signature", "v1,invalid")
	resendRec := httptest.NewRecorder()
	mux.ServeHTTP(resendRec, resendReq)
	require.Equal(t, http.StatusUnauthorized, resendRec.Code)

	polarCount, err := app.CountRecords(
		"webhook_events",
		dbx.HashExp{"provider": "polar", "event_type": "subscription.updated"},
	)
	require.NoError(t, err)
	require.Equal(t, int64(0), polarCount)

	resendCount, err := app.CountRecords(
		"webhook_events",
		dbx.HashExp{"provider": "resend", "event_type": "email.delivered"},
	)
	require.NoError(t, err)
	require.Equal(t, int64(0), resendCount)
}

func newWebhookTestServer(t *testing.T, cfg config.Config) (core.App, http.Handler) {
	t.Helper()

	app := testutil.NewTestApp(t)

	pbRouter, err := apis.NewRouter(app)
	require.NoError(t, err)
	serveEvent := &core.ServeEvent{App: app, Router: pbRouter}

	deps := &Deps{
		Config:  cfg,
		Auth:    auth.NewWithConfig(app, cfg),
		Billing: billing.New(cfg),
		Email:   email.New(cfg),
		Tenancy: tenancy.New(tenancy.NewPocketBaseRepository(app)),
	}
	RegisterRoutes(serveEvent, deps)

	mux, err := serveEvent.Router.BuildMux()
	require.NoError(t, err)
	return app, mux
}
