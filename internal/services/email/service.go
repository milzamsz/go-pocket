package email

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/pocketbase/pocketbase/core"
	"github.com/resend/resend-go/v3"
)

type Service interface {
	SendInvite(ctx context.Context, to string, orgName string) error
	SendWelcome(ctx context.Context, to string) error
	SendOrgSettingsUpdated(ctx context.Context, to string, orgName string) error
	SendPasswordReset(ctx context.Context, to string, token string) error
	SendEmailVerification(ctx context.Context, to string, token string) error
	VerifyAndDispatchWebhook(ctx context.Context, headers http.Header, payload []byte) error
}

type service struct {
	client   EmailClient
	from     string
	baseURL  string
	verifier WebhookVerifier
	secret   string
	store    webhookEventStore
	storeMu  sync.RWMutex
	mu       sync.Mutex
	effects  webhookSideEffects
}

const (
	templateWelcome        = "go-pocket.welcome"
	templateInvite         = "go-pocket.invite"
	templatePasswordReset  = "go-pocket.password-reset"
	templateVerifyEmail    = "go-pocket.verify-email"
	templateSettingsUpdate = "go-pocket.settings-updated"
)

func New(cfg config.Config) Service {
	var client EmailClient
	if strings.TrimSpace(cfg.Resend.APIKey) != "" {
		client = resendClientAdapter{client: resend.NewClient(cfg.Resend.APIKey)}
	}

	return newServiceWithDependencies(cfg, client, nil, nil)
}

func NewWithDependencies(cfg config.Config, client EmailClient, verifier WebhookVerifier) Service {
	return newServiceWithDependencies(cfg, client, verifier, nil)
}

func newServiceWithDependencies(cfg config.Config, client EmailClient, verifier WebhookVerifier, store webhookEventStore) Service {
	if verifier == nil {
		verifier = resendWebhookVerifier{}
	}
	if store == nil {
		store = noopWebhookEventStore{}
	}
	return &service{
		client:   client,
		from:     cfg.Resend.From,
		baseURL:  cfg.BaseURL,
		verifier: verifier,
		secret:   cfg.Resend.WebhookSecret,
		store:    store,
	}
}

type EmailClient interface {
	SendWithContext(ctx context.Context, params *resend.SendEmailRequest) (*resend.SendEmailResponse, error)
}

type resendClientAdapter struct {
	client *resend.Client
}

func (c resendClientAdapter) SendWithContext(ctx context.Context, params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	return c.client.Emails.SendWithContext(ctx, params)
}

type WebhookVerifier interface {
	Verify(headers http.Header, payload []byte, secret string) error
}

type resendWebhookVerifier struct{}

func (v resendWebhookVerifier) Verify(headers http.Header, payload []byte, secret string) error {
	if secret == "" {
		return nil
	}
	err := (&resend.WebhooksSvcImpl{}).Verify(&resend.VerifyWebhookOptions{
		Payload: string(payload),
		Headers: resend.WebhookHeaders{
			Id:        strings.TrimSpace(headers.Get("svix-id")),
			Timestamp: strings.TrimSpace(headers.Get("svix-timestamp")),
			Signature: strings.TrimSpace(headers.Get("svix-signature")),
		},
		WebhookSecret: secret,
	})
	if err != nil {
		return ErrInvalidWebhookSignature
	}
	return nil
}

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

func (s *service) SendInvite(ctx context.Context, to string, orgName string) error {
	if s.client == nil {
		return nil
	}
	inviteURL := s.composeURL("/auth/signup", "")
	_, err := s.client.SendWithContext(ctx, s.buildTemplateRequest(templateInvite, "invite", "Organization invite", to, map[string]any{
		"organization_name": orgName,
		"invite_url":        inviteURL,
	}, fmt.Sprintf("You were invited to join %s. Open %s to continue.", orgName, inviteURL)))
	if err != nil {
		return fmt.Errorf("send invite email: %w", err)
	}
	return nil
}

func (s *service) SendWelcome(ctx context.Context, to string) error {
	if s.client == nil {
		return nil
	}
	dashboardURL := s.composeURL("/app", "")
	_, err := s.client.SendWithContext(ctx, s.buildTemplateRequest(templateWelcome, "welcome", "Welcome", to, map[string]any{
		"dashboard_url": dashboardURL,
	}, fmt.Sprintf("Welcome to go-pocket. Open %s to get started.", dashboardURL)))
	if err != nil {
		return fmt.Errorf("send welcome email: %w", err)
	}
	return nil
}

func (s *service) SendOrgSettingsUpdated(ctx context.Context, to string, orgName string) error {
	if s.client == nil {
		return nil
	}
	settingsURL := s.composeURL("/org/"+orgName+"/settings", "")
	_, err := s.client.SendWithContext(ctx, s.buildTemplateRequest(templateSettingsUpdate, "settings_updated", "Organization settings updated", to, map[string]any{
		"organization_name": orgName,
		"settings_url":      settingsURL,
	}, fmt.Sprintf("Settings were updated for %s. Review at %s.", orgName, settingsURL)))
	if err != nil {
		return fmt.Errorf("send settings update email: %w", err)
	}
	return nil
}

func (s *service) SendPasswordReset(ctx context.Context, to string, token string) error {
	if s.client == nil {
		return nil
	}
	resetURL := s.composeURL("/auth/reset-password", token)
	_, err := s.client.SendWithContext(ctx, s.buildTemplateRequest(templatePasswordReset, "password_reset", "Reset your password", to, map[string]any{
		"reset_url": resetURL,
	}, fmt.Sprintf("Use this link to reset your password: %s", resetURL)))
	if err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}

func (s *service) SendEmailVerification(ctx context.Context, to string, token string) error {
	if s.client == nil {
		return nil
	}
	verifyURL := s.composeURL("/auth/verify-email", token)
	_, err := s.client.SendWithContext(ctx, s.buildTemplateRequest(templateVerifyEmail, "verify_email", "Verify your email", to, map[string]any{
		"verification_url": verifyURL,
	}, fmt.Sprintf("Use this link to verify your email: %s", verifyURL)))
	if err != nil {
		return fmt.Errorf("send email verification email: %w", err)
	}
	return nil
}

func (s *service) composeURL(routePath string, token string) string {
	base, err := url.Parse(strings.TrimSpace(s.baseURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		base = &url.URL{Scheme: "http", Host: "localhost:8090"}
	}
	base.Path = path.Join(strings.TrimSuffix(base.Path, "/"), strings.TrimPrefix(routePath, "/"))
	if strings.TrimSpace(token) != "" {
		query := base.Query()
		query.Set("token", token)
		base.RawQuery = query.Encode()
	}
	return base.String()
}

func (s *service) buildTemplateRequest(templateID string, eventName string, subject string, to string, variables map[string]any, fallbackText string) *resend.SendEmailRequest {
	recipientHash := hashRecipient(to)
	return &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Text:    fallbackText,
		Template: &resend.EmailTemplate{
			Id:        templateID,
			Variables: variables,
		},
		Tags: []resend.Tag{
			{Name: "app", Value: "go-pocket"},
			{Name: "event", Value: eventName},
			{Name: "recipient_hash", Value: recipientHash},
		},
		Headers: map[string]string{
			"X-GoPocket-Template":         templateID,
			"X-GoPocket-Event":            eventName,
			"X-GoPocket-Recipient-SHA256": recipientHash,
		},
	}
}

func (s *service) VerifyAndDispatchWebhook(ctx context.Context, headers http.Header, payload []byte) error {
	if err := s.verifier.Verify(headers, payload, s.secret); err != nil {
		return fmt.Errorf("verify resend webhook: %w", err)
	}
	if _, err := s.dispatchWebhookEvent(ctx, payload); err != nil {
		return fmt.Errorf("dispatch resend webhook: %w", err)
	}
	return nil
}

type emailWebhookEvent struct {
	Type      string         `json:"type"`
	CreatedAt string         `json:"created_at"`
	Data      map[string]any `json:"data"`
}

type webhookDispatchResult struct {
	EventType string
	Family    string
	Handled   bool
}

func (s *service) dispatchWebhookEvent(ctx context.Context, payload []byte) (webhookDispatchResult, error) {
	var event emailWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return webhookDispatchResult{}, fmt.Errorf("parse resend webhook payload: %w", err)
	}

	result := webhookDispatchResult{EventType: event.Type}
	family := emailEventFamily(event.Type)
	result.Family = family

	s.storeMu.RLock()
	store := s.store
	s.storeMu.RUnlock()
	dedupeKey := buildEmailWebhookDedupeKey(event, payload)
	duplicate, err := store.IsDuplicate(ctx, "resend", dedupeKey)
	if err != nil {
		return result, fmt.Errorf("check email webhook idempotency: %w", err)
	}
	if duplicate {
		return result, nil
	}

	handler, ok := s.emailEventFamilyHandlers()[family]
	if ok {
		if err := handler(event); err != nil {
			return result, err
		}
		result.Handled = true
	}

	if err := store.Append(ctx, buildWebhookEventAuditRecord("resend", result, event, payload)); err != nil {
		return result, fmt.Errorf("persist email webhook event: %w", err)
	}
	return result, nil
}

func buildWebhookEventAuditRecord(provider string, result webhookDispatchResult, event emailWebhookEvent, payload []byte) webhookEventAuditRecord {
	excerpt := strings.TrimSpace(string(payload))
	if len(excerpt) > 512 {
		excerpt = excerpt[:512]
	}
	hash := sha256.Sum256(payload)
	occurredAt := parseWebhookOccurredAt(event.CreatedAt)
	messageID := firstString(event.Data, "email_id", "message_id", "id")
	recipient := firstRecipient(event.Data)
	return webhookEventAuditRecord{
		Provider:       provider,
		EventType:      result.EventType,
		Family:         result.Family,
		Handled:        result.Handled,
		Status:         webhookStatusFromFamily(result.Family),
		MessageID:      messageID,
		Recipient:      recipient,
		DedupeKey:      buildEmailWebhookDedupeKey(event, payload),
		PayloadHash:    hex.EncodeToString(hash[:]),
		PayloadExcerpt: excerpt,
		OccurredAt:     occurredAt,
		ReceivedAt:     time.Now().UTC(),
	}
}

type webhookEventAuditRecord struct {
	Provider       string
	EventType      string
	Family         string
	Handled        bool
	Status         string
	MessageID      string
	Recipient      string
	DedupeKey      string
	PayloadHash    string
	PayloadExcerpt string
	OccurredAt     time.Time
	ReceivedAt     time.Time
}

type webhookEventStore interface {
	IsDuplicate(ctx context.Context, provider string, dedupeKey string) (bool, error)
	Append(ctx context.Context, record webhookEventAuditRecord) error
}

type noopWebhookEventStore struct{}

func (noopWebhookEventStore) IsDuplicate(_ context.Context, _ string, _ string) (bool, error) {
	return false, nil
}

func (noopWebhookEventStore) Append(_ context.Context, _ webhookEventAuditRecord) error {
	return nil
}

type pocketBaseWebhookEventStore struct {
	app core.App
}

func newPocketBaseWebhookEventStore(app core.App) webhookEventStore {
	if app == nil {
		return noopWebhookEventStore{}
	}
	return pocketBaseWebhookEventStore{app: app}
}

func (s pocketBaseWebhookEventStore) Append(_ context.Context, record webhookEventAuditRecord) error {
	collection, err := s.app.FindCollectionByNameOrId("webhook_events")
	if err != nil {
		return fmt.Errorf("find webhook_events collection: %w", err)
	}
	entry := core.NewRecord(collection)
	entry.Set("provider", record.Provider)
	entry.Set("event_type", record.EventType)
	entry.Set("family", record.Family)
	entry.Set("handled", record.Handled)
	entry.Set("status", record.Status)
	entry.Set("message_id", record.MessageID)
	entry.Set("recipient", record.Recipient)
	entry.Set("dedupe_key", record.DedupeKey)
	entry.Set("payload_hash", record.PayloadHash)
	entry.Set("payload_excerpt", record.PayloadExcerpt)
	entry.Set("occurred_at", record.OccurredAt)
	entry.Set("received_at", record.ReceivedAt)
	if err := s.app.Save(entry); err != nil {
		return fmt.Errorf("save webhook event audit record: %w", err)
	}
	return nil
}

func (s pocketBaseWebhookEventStore) IsDuplicate(_ context.Context, provider string, dedupeKey string) (bool, error) {
	record, err := s.app.FindFirstRecordByFilter(
		"webhook_events",
		"provider = {:provider} && dedupe_key = {:dedupeKey}",
		map[string]any{"provider": provider, "dedupeKey": dedupeKey},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return record != nil, nil
}

func (s *service) BindWebhookEventStore(app core.App) {
	s.storeMu.Lock()
	defer s.storeMu.Unlock()
	s.store = newPocketBaseWebhookEventStore(app)
}

func emailEventFamily(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return ""
	}
	first, rest, found := strings.Cut(eventType, ".")
	if !found {
		return first
	}
	if first == "email" {
		second, _, _ := strings.Cut(rest, ".")
		return second
	}
	return first
}

type emailEventHandler func(event emailWebhookEvent) error

type webhookSideEffects struct {
	handledByFamily map[string]int
	lastHandledType string
	lastHandledAt   time.Time
}

func (s *service) emailEventFamilyHandlers() map[string]emailEventHandler {
	return map[string]emailEventHandler{
		"sent":       s.handleSentEvent,
		"delivered":  s.handleDeliveredEvent,
		"bounced":    s.handleBouncedEvent,
		"complained": s.handleComplainedEvent,
	}
}

func (s *service) handleSentEvent(event emailWebhookEvent) error {
	s.recordHandledEvent("sent", event.Type)
	return nil
}

func (s *service) handleDeliveredEvent(event emailWebhookEvent) error {
	s.recordHandledEvent("delivered", event.Type)
	return nil
}

func (s *service) handleBouncedEvent(event emailWebhookEvent) error {
	s.recordHandledEvent("bounced", event.Type)
	return nil
}

func (s *service) handleComplainedEvent(event emailWebhookEvent) error {
	s.recordHandledEvent("complained", event.Type)
	return nil
}

func (s *service) recordHandledEvent(family, eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.effects.handledByFamily == nil {
		s.effects.handledByFamily = make(map[string]int)
	}
	s.effects.handledByFamily[family]++
	s.effects.lastHandledType = eventType
	s.effects.lastHandledAt = time.Now().UTC()
}

func (s *service) handledCountForFamily(family string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.effects.handledByFamily == nil {
		return 0
	}
	return s.effects.handledByFamily[family]
}

func hashRecipient(to string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(to))))
	return hex.EncodeToString(hash[:])
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		v, ok := values[key]
		if !ok {
			continue
		}
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func firstRecipient(values map[string]any) string {
	if s := firstString(values, "recipient", "email"); s != "" {
		return s
	}
	recipients, ok := values["to"].([]any)
	if !ok || len(recipients) == 0 {
		return ""
	}
	first, ok := recipients[0].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(first)
}

func parseWebhookOccurredAt(createdAt string) time.Time {
	ts := strings.TrimSpace(createdAt)
	if ts == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}

func webhookStatusFromFamily(family string) string {
	switch family {
	case "sent", "delivered", "bounced", "complained":
		return family
	default:
		return "unknown"
	}
}

func buildEmailWebhookDedupeKey(event emailWebhookEvent, payload []byte) string {
	seed := firstString(event.Data, "message_id", "email_id", "id")
	if seed == "" {
		hash := sha256.Sum256(payload)
		seed = hex.EncodeToString(hash[:])
	}
	return strings.TrimSpace(event.Type) + ":" + seed
}
