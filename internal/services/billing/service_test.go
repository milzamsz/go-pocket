package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	checkoutURL string
	portalURL   string
}

func (f fakeClient) CreateCheckoutSessionURL(_ context.Context, _ string) (string, error) {
	return f.checkoutURL, nil
}

func (f fakeClient) CreatePortalSessionURL(_ context.Context, _ string) (string, error) {
	return f.portalURL, nil
}

type fakeVerifier struct {
	err error
}

func (f fakeVerifier) Verify(_ http.Header, _ []byte, _ string) error {
	return f.err
}

func TestCreateCheckoutSession_ReturnsFallbackWhenProviderNotConfigured(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
	}, fakeClient{checkoutURL: "https://polar.example.com/c"}, nil)

	url, err := svc.CreateCheckoutSession(context.Background(), "acme")
	require.NoError(t, err)
	require.Equal(t, "https://app.example.com/org/acme/billing?provider=unconfigured", url)
}

func TestCreateCheckoutSession_UsesClientURLWhenConfigured(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
		Polar:   config.PolarConfig{AccessToken: "polar-token"},
	}, fakeClient{checkoutURL: "https://polar.example.com/checkout"}, nil)

	url, err := svc.CreateCheckoutSession(context.Background(), "acme")
	require.NoError(t, err)
	require.Equal(t, "https://polar.example.com/checkout", url)
}

func TestCreatePortalSession_ReturnsFallbackWhenProviderNotConfigured(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
	}, fakeClient{portalURL: "https://polar.example.com/p"}, nil)

	url, err := svc.CreatePortalSession(context.Background(), "acme")
	require.NoError(t, err)
	require.Equal(t, "https://app.example.com/org/acme/billing?provider=unconfigured", url)
}

func TestVerifyAndDispatchWebhook_InvalidSignature(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{
		Polar: config.PolarConfig{WebhookSecret: "secret"},
	}, nil, fakeVerifier{err: ErrInvalidWebhookSignature})

	err := svc.VerifyAndDispatchWebhook(context.Background(), http.Header{}, []byte("{}"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWebhookSignature)
}

func TestVerifyAndDispatchWebhook_ValidPolarSignature(t *testing.T) {
	t.Parallel()

	secret := "polar-secret"
	payload := []byte(`{"type":"order.paid"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	sig := signPolarPayloadForTest(secret, timestamp, payload)

	headers := http.Header{}
	headers.Set("Polar-Signature", fmt.Sprintf("t=%s,v1=%s", timestamp, sig))

	svc := NewWithDependencies(config.Config{
		Polar: config.PolarConfig{WebhookSecret: secret},
	}, nil, nil)

	err := svc.VerifyAndDispatchWebhook(context.Background(), headers, payload)
	require.NoError(t, err)
}

func TestVerifyAndDispatchWebhook_InvalidPolarSignature(t *testing.T) {
	t.Parallel()

	secret := "polar-secret"
	payload := []byte(`{"type":"order.paid"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	headers := http.Header{}
	headers.Set("Polar-Signature", fmt.Sprintf("t=%s,v1=%s", timestamp, "invalid"))

	svc := NewWithDependencies(config.Config{
		Polar: config.PolarConfig{WebhookSecret: secret},
	}, nil, nil)

	err := svc.VerifyAndDispatchWebhook(context.Background(), headers, payload)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWebhookSignature)
}

func TestVerifyAndDispatchWebhook_NoSecretSkipsVerification(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{}, nil, nil)
	err := svc.VerifyAndDispatchWebhook(context.Background(), http.Header{}, []byte("{}"))
	require.NoError(t, err)
}

func TestDispatchWebhookEvent_KnownBillingFamily(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{}, nil, nil).(*service)
	result, err := svc.dispatchWebhookEvent(context.Background(), []byte(`{"type":"subscription.updated"}`))
	require.NoError(t, err)
	require.True(t, result.Handled)
	require.Equal(t, "subscription.updated", result.EventType)
	require.Equal(t, "subscription", result.Family)
	require.Equal(t, 1, svc.handledCountForFamily("subscription"))
	require.Equal(t, "subscription.updated", svc.effects.lastHandledType)
	require.False(t, svc.effects.lastHandledAt.IsZero())
}

func TestDispatchWebhookEvent_UnknownBillingFamilyIsIgnored(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{}, nil, nil).(*service)
	result, err := svc.dispatchWebhookEvent(context.Background(), []byte(`{"type":"invoice.created"}`))
	require.NoError(t, err)
	require.False(t, result.Handled)
	require.Equal(t, "invoice.created", result.EventType)
	require.Equal(t, "invoice", result.Family)
	require.Equal(t, 0, svc.handledCountForFamily("subscription"))
	require.Equal(t, 0, svc.handledCountForFamily("order"))
	require.Equal(t, 0, svc.handledCountForFamily("customer"))
	require.Equal(t, "", svc.effects.lastHandledType)
	require.True(t, svc.effects.lastHandledAt.IsZero())
}

func TestDispatchWebhookEvent_PersistsHandledAuditRecord(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
	payload := []byte(`{"type":"subscription.updated"}`)

	result, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.True(t, result.Handled)

	record, err := app.FindFirstRecordByFilter("webhook_events", "provider = {:provider}", dbx.Params{"provider": "polar"})
	require.NoError(t, err)
	require.Equal(t, "subscription.updated", record.GetString("event_type"))
	require.Equal(t, "subscription", record.GetString("family"))
	require.True(t, record.GetBool("handled"))
	require.Equal(t, 64, len(record.GetString("payload_hash")))
	require.Equal(t, string(payload), record.GetString("payload_excerpt"))
}

func TestDispatchWebhookEvent_PersistsUnhandledAuditRecord(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
	payload := []byte(`{"type":"invoice.created"}`)

	result, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.False(t, result.Handled)

	record, err := app.FindFirstRecordByFilter("webhook_events", "provider = {:provider}", dbx.Params{"provider": "polar"})
	require.NoError(t, err)
	require.Equal(t, "invoice.created", record.GetString("event_type"))
	require.Equal(t, "invoice", record.GetString("family"))
	require.False(t, record.GetBool("handled"))
}

func TestDispatchWebhookEvent_UpsertsSubscriptionState(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	org := createOrgRecord(t, app, "acme")
	svc := New(config.Config{}).(*service)
	svc.BindWebhookEventStore(app)

	payload := []byte(`{"type":"subscription.updated","data":{"id":"sub_123","organization_id":"` + org.Id + `","status":"active"}}`)
	_, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)

	record, err := app.FindFirstRecordByFilter("subscriptions", "provider_subscription_id = {:id}", dbx.Params{"id": "sub_123"})
	require.NoError(t, err)
	require.Equal(t, org.Id, record.GetString("organization"))
	require.Equal(t, "polar", record.GetString("provider"))
	require.Equal(t, "active", record.GetString("status"))
}

func TestDispatchWebhookEvent_UpsertsInvoiceStateAndLinksSubscription(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	org := createOrgRecord(t, app, "acme")
	subscription := createSubscriptionRecord(t, app, org.Id, "sub_abc")

	svc := New(config.Config{}).(*service)
	svc.BindWebhookEventStore(app)

	payload := []byte(`{"type":"order.paid","data":{"id":"inv_987","organization_id":"` + org.Id + `","status":"paid","amount":1234,"subscription":{"id":"sub_abc"}}}`)
	_, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)

	record, err := app.FindFirstRecordByFilter("invoices", "provider_invoice_id = {:id}", dbx.Params{"id": "inv_987"})
	require.NoError(t, err)
	require.Equal(t, org.Id, record.GetString("organization"))
	require.Equal(t, "paid", record.GetString("status"))
	require.Equal(t, float64(1234), record.GetFloat("amount_cents"))
	require.Equal(t, subscription.Id, record.GetString("subscription"))
}

func TestDispatchWebhookEvent_Contract_SubscriptionUpdated_MapsNestedFieldsToOrgDenorm(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	org := createOrgRecord(t, app, "acme")
	svc := New(config.Config{}).(*service)
	svc.BindWebhookEventStore(app)

	payload := []byte(`{
		"type":"subscription.updated",
		"data":{
			"status":"active",
			"customer":{"id":"cus_001","external_id":"acme"},
			"subscription":{"id":"sub_001"},
			"product":{"id":"prod_001"},
			"price":{"id":"price_001"}
		}
	}`)
	_, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)

	updatedOrg, err := app.FindRecordById("organizations", org.Id)
	require.NoError(t, err)
	require.Equal(t, "cus_001", updatedOrg.GetString("polar_customer_id"))
	require.Equal(t, "sub_001", updatedOrg.GetString("polar_subscription_id"))
	require.Equal(t, "prod_001", updatedOrg.GetString("polar_product_id"))
	require.Equal(t, "price_001", updatedOrg.GetString("polar_price_id"))
	require.Equal(t, "prod_001", updatedOrg.GetString("plan"))
	require.Equal(t, "active", updatedOrg.GetString("subscription_status"))
}

func TestDispatchWebhookEvent_Contract_OrderPaid_UsesInvoiceDotIDAndAmountCents(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	org := createOrgRecord(t, app, "acme")
	createSubscriptionRecord(t, app, org.Id, "sub_abc")

	svc := New(config.Config{}).(*service)
	svc.BindWebhookEventStore(app)

	payload := []byte(`{
		"type":"order.paid",
		"data":{
			"organization_id":"` + org.Id + `",
			"status":"paid",
			"amount_cents":1999,
			"invoice":{"id":"inv_dot_001"},
			"subscription":{"id":"sub_abc"}
		}
	}`)
	_, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)

	record, err := app.FindFirstRecordByFilter("invoices", "provider_invoice_id = {:id}", dbx.Params{"id": "inv_dot_001"})
	require.NoError(t, err)
	require.Equal(t, org.Id, record.GetString("organization"))
	require.Equal(t, float64(1999), record.GetFloat("amount_cents"))
	require.Equal(t, "paid", record.GetString("status"))
}

func TestDispatchWebhookEvent_FixtureContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		setup   func(t *testing.T, app core.App, org *core.Record)
		assert  func(t *testing.T, app core.App, org *core.Record)
	}{
		{
			name:    "subscription updated fixture updates org denorm and subscription",
			fixture: "polar_subscription_updated.json",
			assert: func(t *testing.T, app core.App, org *core.Record) {
				t.Helper()
				orgRecord, err := app.FindRecordById("organizations", org.Id)
				require.NoError(t, err)
				require.Equal(t, "cus_fixture_001", orgRecord.GetString("polar_customer_id"))
				require.Equal(t, "sub_fixture_001", orgRecord.GetString("polar_subscription_id"))
				require.Equal(t, "prod_fixture_001", orgRecord.GetString("polar_product_id"))
				require.Equal(t, "price_fixture_001", orgRecord.GetString("polar_price_id"))
				require.Equal(t, "prod_fixture_001", orgRecord.GetString("plan"))
				require.Equal(t, "active", orgRecord.GetString("subscription_status"))

				sub, err := app.FindFirstRecordByFilter("subscriptions", "provider_subscription_id = {:id}", dbx.Params{"id": "sub_fixture_001"})
				require.NoError(t, err)
				require.Equal(t, org.Id, sub.GetString("organization"))
				require.Equal(t, "active", sub.GetString("status"))
			},
		},
		{
			name:    "order paid fixture creates invoice and links subscription",
			fixture: "polar_order_paid.json",
			setup: func(t *testing.T, app core.App, org *core.Record) {
				t.Helper()
				createSubscriptionRecord(t, app, org.Id, "sub_fixture_order_001")
			},
			assert: func(t *testing.T, app core.App, org *core.Record) {
				t.Helper()
				invoice, err := app.FindFirstRecordByFilter("invoices", "provider_invoice_id = {:id}", dbx.Params{"id": "inv_fixture_001"})
				require.NoError(t, err)
				require.Equal(t, org.Id, invoice.GetString("organization"))
				require.Equal(t, float64(2599), invoice.GetFloat("amount_cents"))
				require.Equal(t, "paid", invoice.GetString("status"))
			},
		},
		{
			name:    "customer updated fixture updates org customer denorm",
			fixture: "polar_customer_updated.json",
			assert: func(t *testing.T, app core.App, org *core.Record) {
				t.Helper()
				orgRecord, err := app.FindRecordById("organizations", org.Id)
				require.NoError(t, err)
				require.Equal(t, "cus_fixture_999", orgRecord.GetString("polar_customer_id"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := testutil.NewTestApp(t)
			org := createOrgRecord(t, app, "acme")
			if tt.setup != nil {
				tt.setup(t, app, org)
			}

			svc := New(config.Config{}).(*service)
			svc.BindWebhookEventStore(app)

			payload := loadBillingFixture(t, tt.fixture, org.Id, org.GetString("slug"))
			_, err := svc.dispatchWebhookEvent(context.Background(), payload)
			require.NoError(t, err)

			tt.assert(t, app, org)
		})
	}
}

func TestDispatchWebhookEvent_DuplicateHandledEventIsIgnored(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	org := createOrgRecord(t, app, "acme")
	svc := New(config.Config{}).(*service)
	svc.BindWebhookEventStore(app)
	payload := []byte(`{"type":"subscription.updated","data":{"id":"sub_dup_001","organization_id":"` + org.Id + `","status":"active"}}`)

	first, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.True(t, first.Handled)
	require.Equal(t, 1, svc.handledCountForFamily("subscription"))

	second, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.False(t, second.Handled)
	require.Equal(t, 1, svc.handledCountForFamily("subscription"))

	count, err := app.CountRecords("webhook_events", dbx.HashExp{"provider": "polar"})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func loadBillingFixture(t *testing.T, name, orgID, orgSlug string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	replaced := strings.ReplaceAll(string(raw), "__ORG_ID__", orgID)
	replaced = strings.ReplaceAll(replaced, "__ORG_SLUG__", orgSlug)
	return []byte(replaced)
}

func createOrgRecord(t *testing.T, app core.App, slug string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("organizations")
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Set("slug", slug)
	record.Set("name", "Acme")
	record.Set("owner", "user-owner")
	require.NoError(t, app.SaveNoValidate(record))
	return record
}

func createSubscriptionRecord(t *testing.T, app core.App, orgID string, providerSubID string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("subscriptions")
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Set("organization", orgID)
	record.Set("provider", "polar")
	record.Set("provider_subscription_id", providerSubID)
	record.Set("status", "active")
	require.NoError(t, app.SaveNoValidate(record))
	return record
}

func signPolarPayloadForTest(secret, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
