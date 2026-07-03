package email

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
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
	"github.com/resend/resend-go/v3"
	"github.com/stretchr/testify/require"
)

type fakeEmailClient struct {
	calls    int
	requests []*resend.SendEmailRequest
}

func (f *fakeEmailClient) SendWithContext(_ context.Context, req *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	f.calls++
	f.requests = append(f.requests, req)
	return &resend.SendEmailResponse{}, nil
}

type fakeWebhookVerifier struct {
	err error
}

func (f fakeWebhookVerifier) Verify(_ http.Header, _ []byte, _ string) error {
	return f.err
}

func TestSendInvite_NoClientConfiguredIsNoop(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{Resend: config.ResendConfig{From: "noreply@example.com"}}, nil, nil)
	require.NoError(t, svc.SendInvite(context.Background(), "user@example.com", "acme"))
}

func TestSendInvite_UsesTemplateAndTrackingTags(t *testing.T) {
	t.Parallel()

	client := &fakeEmailClient{}
	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
		Resend:  config.ResendConfig{From: "noreply@example.com"},
	}, client, nil)

	err := svc.SendInvite(context.Background(), "user@example.com", "acme")
	require.NoError(t, err)
	require.Equal(t, 1, client.calls)
	require.Equal(t, "go-pocket.invite", client.requests[0].Template.Id)
	require.Equal(t, "invite", client.requests[0].Headers["X-GoPocket-Event"])
	require.Equal(t, "acme", client.requests[0].Template.Variables["organization_name"])
	require.Equal(t, "https://app.example.com/auth/signup", client.requests[0].Template.Variables["invite_url"])
}

func TestSendOrgSettingsUpdated_SendsWhenClientConfigured(t *testing.T) {
	t.Parallel()

	client := &fakeEmailClient{}
	svc := NewWithDependencies(config.Config{Resend: config.ResendConfig{From: "noreply@example.com"}}, client, nil)

	err := svc.SendOrgSettingsUpdated(context.Background(), "user@example.com", "acme")
	require.NoError(t, err)
	require.Equal(t, 1, client.calls)
	require.Equal(t, "go-pocket.settings-updated", client.requests[0].Template.Id)
	require.Equal(t, "settings_updated", client.requests[0].Headers["X-GoPocket-Event"])
	require.Equal(t, "go-pocket.settings-updated", client.requests[0].Headers["X-GoPocket-Template"])
	require.Equal(t, "settings_updated", findTag(client.requests[0].Tags, "event"))
	require.Equal(t, "go-pocket", findTag(client.requests[0].Tags, "app"))
}

func TestSendPasswordReset_ComposesResetLinkWithToken(t *testing.T) {
	t.Parallel()

	client := &fakeEmailClient{}
	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
		Resend:  config.ResendConfig{From: "noreply@example.com"},
	}, client, nil)

	err := svc.SendPasswordReset(context.Background(), "user@example.com", "reset_abcDEF12")
	require.NoError(t, err)
	require.Equal(t, 1, client.calls)
	require.Equal(t, []string{"user@example.com"}, client.requests[0].To)
	require.Contains(t, client.requests[0].Text, "https://app.example.com/auth/reset-password?token=reset_abcDEF12")
	require.Equal(t, "go-pocket.password-reset", client.requests[0].Template.Id)
	require.Equal(t, "password_reset", client.requests[0].Headers["X-GoPocket-Event"])
	require.Equal(t, "https://app.example.com/auth/reset-password?token=reset_abcDEF12", client.requests[0].Template.Variables["reset_url"])
}

func TestSendEmailVerification_ComposesVerifyLinkWithToken(t *testing.T) {
	t.Parallel()

	client := &fakeEmailClient{}
	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
		Resend:  config.ResendConfig{From: "noreply@example.com"},
	}, client, nil)

	err := svc.SendEmailVerification(context.Background(), "user@example.com", "verify_abcDEF12")
	require.NoError(t, err)
	require.Equal(t, 1, client.calls)
	require.Equal(t, []string{"user@example.com"}, client.requests[0].To)
	require.Contains(t, client.requests[0].Text, "https://app.example.com/auth/verify-email?token=verify_abcDEF12")
	require.Equal(t, "go-pocket.verify-email", client.requests[0].Template.Id)
	require.Equal(t, "verify_email", client.requests[0].Headers["X-GoPocket-Event"])
	require.Equal(t, "https://app.example.com/auth/verify-email?token=verify_abcDEF12", client.requests[0].Template.Variables["verification_url"])
}

func TestSendWelcome_UsesTemplateAndTrackingTags(t *testing.T) {
	t.Parallel()

	client := &fakeEmailClient{}
	svc := NewWithDependencies(config.Config{
		BaseURL: "https://app.example.com",
		Resend:  config.ResendConfig{From: "noreply@example.com"},
	}, client, nil)

	err := svc.SendWelcome(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.Equal(t, 1, client.calls)
	require.Equal(t, "go-pocket.welcome", client.requests[0].Template.Id)
	require.Equal(t, "welcome", client.requests[0].Headers["X-GoPocket-Event"])
	require.Equal(t, hashRecipient("user@example.com"), client.requests[0].Headers["X-GoPocket-Recipient-SHA256"])
	require.Equal(t, hashRecipient("user@example.com"), findTag(client.requests[0].Tags, "recipient_hash"))
}

func TestVerifyAndDispatchWebhook_InvalidSignature(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(
		config.Config{Resend: config.ResendConfig{WebhookSecret: "secret"}},
		nil,
		fakeWebhookVerifier{err: ErrInvalidWebhookSignature},
	)

	err := svc.VerifyAndDispatchWebhook(context.Background(), http.Header{}, []byte("{}"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWebhookSignature)
}

func TestVerifyAndDispatchWebhook_ValidSignature(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(
		config.Config{Resend: config.ResendConfig{WebhookSecret: "secret"}},
		nil,
		fakeWebhookVerifier{},
	)

	err := svc.VerifyAndDispatchWebhook(context.Background(), http.Header{}, []byte("{}"))
	require.NoError(t, err)
	require.False(t, errors.Is(err, ErrInvalidWebhookSignature))
}

func TestVerifyAndDispatchWebhook_ValidResendSignature(t *testing.T) {
	t.Parallel()

	secret := "whsec_bXktdGVzdC1zZWNyZXQ="
	payload := []byte(`{"type":"email.delivered"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	id := "msg_123"
	signature := signSvixPayloadForTest(secret, id, timestamp, payload)

	headers := http.Header{}
	headers.Set("svix-id", id)
	headers.Set("svix-timestamp", timestamp)
	headers.Set("svix-signature", "v1,"+signature)

	svc := NewWithDependencies(
		config.Config{Resend: config.ResendConfig{WebhookSecret: secret}},
		nil,
		nil,
	)

	err := svc.VerifyAndDispatchWebhook(context.Background(), headers, payload)
	require.NoError(t, err)
}

func TestVerifyAndDispatchWebhook_InvalidResendSignature(t *testing.T) {
	t.Parallel()

	secret := "whsec_bXktdGVzdC1zZWNyZXQ="
	payload := []byte(`{"type":"email.delivered"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	headers := http.Header{}
	headers.Set("svix-id", "msg_123")
	headers.Set("svix-timestamp", timestamp)
	headers.Set("svix-signature", "v1,invalid")

	svc := NewWithDependencies(
		config.Config{Resend: config.ResendConfig{WebhookSecret: secret}},
		nil,
		nil,
	)

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

func TestVerifyAndDispatchWebhook_ProductionRequiresSecret(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{AppEnv: "production"}, nil, nil)
	err := svc.VerifyAndDispatchWebhook(context.Background(), http.Header{}, []byte("{}"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWebhookSignature)
}

func TestDispatchWebhookEvent_KnownEmailFamily(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{}, nil, nil).(*service)
	result, err := svc.dispatchWebhookEvent(context.Background(), []byte(`{"type":"email.delivered"}`))
	require.NoError(t, err)
	require.True(t, result.Handled)
	require.Equal(t, "email.delivered", result.EventType)
	require.Equal(t, "delivered", result.Family)
	require.Equal(t, 1, svc.handledCountForFamily("delivered"))
	require.Equal(t, "email.delivered", svc.effects.lastHandledType)
	require.False(t, svc.effects.lastHandledAt.IsZero())
}

func TestDispatchWebhookEvent_UnknownEmailFamilyIsIgnored(t *testing.T) {
	t.Parallel()

	svc := NewWithDependencies(config.Config{}, nil, nil).(*service)
	result, err := svc.dispatchWebhookEvent(context.Background(), []byte(`{"type":"email.opened"}`))
	require.NoError(t, err)
	require.False(t, result.Handled)
	require.Equal(t, "email.opened", result.EventType)
	require.Equal(t, "opened", result.Family)
	require.Equal(t, 0, svc.handledCountForFamily("sent"))
	require.Equal(t, 0, svc.handledCountForFamily("delivered"))
	require.Equal(t, 0, svc.handledCountForFamily("bounced"))
	require.Equal(t, 0, svc.handledCountForFamily("complained"))
	require.Equal(t, "", svc.effects.lastHandledType)
	require.True(t, svc.effects.lastHandledAt.IsZero())
}

func TestDispatchWebhookEvent_PersistsHandledAuditRecord(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
	payload := []byte(`{"type":"email.delivered","created_at":"2026-05-27T07:30:00Z","data":{"email_id":"msg_123","to":["user@example.com"]}}`)

	result, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.True(t, result.Handled)

	record, err := app.FindFirstRecordByFilter("webhook_events", "provider = {:provider}", dbx.Params{"provider": "resend"})
	require.NoError(t, err)
	require.Equal(t, "email.delivered", record.GetString("event_type"))
	require.Equal(t, "delivered", record.GetString("family"))
	require.True(t, record.GetBool("handled"))
	require.Equal(t, "delivered", record.GetString("status"))
	require.Equal(t, "msg_123", record.GetString("message_id"))
	require.Equal(t, "user@example.com", record.GetString("recipient"))
	require.Equal(t, "2026-05-27 07:30:00.000Z", record.GetString("occurred_at"))
	require.Equal(t, 64, len(record.GetString("payload_hash")))
	require.Equal(t, string(payload), record.GetString("payload_excerpt"))
}

func TestDispatchWebhookEvent_PersistsUnhandledAuditRecord(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
	payload := []byte(`{"type":"email.opened"}`)

	result, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.False(t, result.Handled)

	record, err := app.FindFirstRecordByFilter("webhook_events", "provider = {:provider}", dbx.Params{"provider": "resend"})
	require.NoError(t, err)
	require.Equal(t, "email.opened", record.GetString("event_type"))
	require.Equal(t, "opened", record.GetString("family"))
	require.False(t, record.GetBool("handled"))
	require.Equal(t, "unknown", record.GetString("status"))
}

func TestDispatchWebhookEvent_FixtureContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fixture        string
		expectedType   string
		expectedFamily string
		expectedStatus string
		expectedMsgID  string
		expectedTo     string
		expectedAt     string
	}{
		{
			name:           "delivered fixture persists delivered status metadata",
			fixture:        "resend_email_delivered.json",
			expectedType:   "email.delivered",
			expectedFamily: "delivered",
			expectedStatus: "delivered",
			expectedMsgID:  "msg_delivered_001",
			expectedTo:     "delivered@example.com",
			expectedAt:     "2026-05-27 08:00:00.000Z",
		},
		{
			name:           "bounced fixture persists bounced status metadata",
			fixture:        "resend_email_bounced.json",
			expectedType:   "email.bounced",
			expectedFamily: "bounced",
			expectedStatus: "bounced",
			expectedMsgID:  "msg_bounced_001",
			expectedTo:     "bounced@example.com",
			expectedAt:     "2026-05-27 08:15:00.000Z",
		},
		{
			name:           "complained fixture persists complained status metadata",
			fixture:        "resend_email_complained.json",
			expectedType:   "email.complained",
			expectedFamily: "complained",
			expectedStatus: "complained",
			expectedMsgID:  "msg_complained_001",
			expectedTo:     "complained@example.com",
			expectedAt:     "2026-05-27 08:30:00.000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := testutil.NewTestApp(t)
			svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
			payload := loadEmailFixture(t, tt.fixture)

			result, err := svc.dispatchWebhookEvent(context.Background(), payload)
			require.NoError(t, err)
			require.True(t, result.Handled)
			require.Equal(t, tt.expectedType, result.EventType)
			require.Equal(t, tt.expectedFamily, result.Family)

			record, err := app.FindFirstRecordByFilter(
				"webhook_events",
				"provider = {:provider} && event_type = {:eventType}",
				dbx.Params{"provider": "resend", "eventType": tt.expectedType},
			)
			require.NoError(t, err)
			require.Equal(t, tt.expectedFamily, record.GetString("family"))
			require.True(t, record.GetBool("handled"))
			require.Equal(t, tt.expectedStatus, record.GetString("status"))
			require.Equal(t, tt.expectedMsgID, record.GetString("message_id"))
			require.Equal(t, tt.expectedTo, record.GetString("recipient"))
			require.Equal(t, tt.expectedAt, record.GetString("occurred_at"))
		})
	}
}

func TestDispatchWebhookEvent_DuplicateHandledEventIsIgnored(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := newServiceWithDependencies(config.Config{}, nil, nil, newPocketBaseWebhookEventStore(app)).(*service)
	payload := loadEmailFixture(t, "resend_email_delivered.json")

	first, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.True(t, first.Handled)
	require.Equal(t, 1, svc.handledCountForFamily("delivered"))

	second, err := svc.dispatchWebhookEvent(context.Background(), payload)
	require.NoError(t, err)
	require.False(t, second.Handled)
	require.Equal(t, 1, svc.handledCountForFamily("delivered"))

	count, err := app.CountRecords("webhook_events", dbx.HashExp{"provider": "resend"})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func loadEmailFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	return raw
}

func signSvixPayloadForTest(secret, id, timestamp string, payload []byte) string {
	decodedSecret, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		return ""
	}
	content := fmt.Sprintf("%s.%s.%s", id, timestamp, payload)
	mac := hmac.New(sha256.New, decodedSecret)
	_, _ = mac.Write([]byte(content))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func findTag(tags []resend.Tag, name string) string {
	for _, tag := range tags {
		if tag.Name == name {
			return tag.Value
		}
	}
	return ""
}
