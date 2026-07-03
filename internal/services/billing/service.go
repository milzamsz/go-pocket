package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
	"github.com/polarsource/polar-go/models/operations"
)

type Service interface {
	CreateCheckoutSession(ctx context.Context, orgSlug string) (string, error)
	CreatePortalSession(ctx context.Context, orgSlug string) (string, error)
	VerifyAndDispatchWebhook(ctx context.Context, headers http.Header, payload []byte) error
}

type service struct {
	cfg        config.Config
	client     Client
	verifier   WebhookVerifier
	store      webhookEventStore
	stateStore billingStateStore
	storeMu    sync.Mutex
	mu         sync.Mutex
	effects    webhookSideEffects
}

func New(cfg config.Config) Service {
	var client Client
	if strings.TrimSpace(cfg.Polar.AccessToken) != "" {
		client = newPolarClientAdapter(cfg)
	}
	return newServiceWithDependencies(cfg, client, nil, nil)
}

func NewWithDependencies(cfg config.Config, client Client, verifier WebhookVerifier) Service {
	return newServiceWithDependencies(cfg, client, verifier, nil)
}

func newServiceWithDependencies(cfg config.Config, client Client, verifier WebhookVerifier, store webhookEventStore) Service {
	if client == nil {
		client = noopClient{}
	}
	if verifier == nil {
		verifier = polarWebhookVerifier{}
	}
	if store == nil {
		store = noopWebhookEventStore{}
	}
	return &service{
		cfg:        cfg,
		client:     client,
		verifier:   verifier,
		store:      store,
		stateStore: noopBillingStateStore{},
	}
}

type Client interface {
	CreateCheckoutSessionURL(ctx context.Context, orgSlug string) (string, error)
	CreatePortalSessionURL(ctx context.Context, orgSlug string) (string, error)
}

type polarClientAdapter struct {
	client            *polargo.Polar
	defaultProductIDs []string
	baseURL           string
}

func newPolarClientAdapter(cfg config.Config) Client {
	opts := []polargo.SDKOption{
		polargo.WithSecurity(cfg.Polar.AccessToken),
	}
	polarServer := strings.TrimSpace(os.Getenv("POLAR_SERVER"))
	if polarServer != "" {
		opts = append(opts, polargo.WithServer(polarServer))
	}
	return polarClientAdapter{
		client:            polargo.New(opts...),
		defaultProductIDs: configuredPolarProductIDs(),
		baseURL:           strings.TrimRight(cfg.BaseURL, "/"),
	}
}

func configuredPolarProductIDs() []string {
	keys := []string{
		"POLAR_PRODUCT_ID",
		"POLAR_PRICE_PRO_MONTHLY",
		"POLAR_PRICE_PRO_YEARLY",
		"POLAR_PRICE_TEAM_MONTHLY",
		"POLAR_PRICE_TEAM_YEARLY",
		"POLAR_PRICE_ENTERPRISE_MONTHLY",
		"POLAR_PRICE_ENTERPRISE_YEARLY",
	}
	ids := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ids = append(ids, value)
	}
	return ids
}

func (c polarClientAdapter) CreateCheckoutSessionURL(ctx context.Context, orgSlug string) (string, error) {
	if len(c.defaultProductIDs) == 0 {
		return "", nil
	}
	successURL := c.baseURL + "/org/" + orgSlug + "/billing"
	response, err := c.client.Checkouts.Create(ctx, components.CheckoutCreate{
		ExternalCustomerID: polargo.String(orgSlug),
		SuccessURL:         polargo.String(successURL),
		Products:           c.defaultProductIDs,
	})
	if err != nil {
		return "", err
	}
	if response == nil || response.Checkout == nil {
		return "", nil
	}
	return strings.TrimSpace(response.Checkout.URL), nil
}

func (c polarClientAdapter) CreatePortalSessionURL(ctx context.Context, orgSlug string) (string, error) {
	request := operations.CreateCustomerSessionsCreateCustomerSessionCreateCustomerSessionCustomerExternalIDCreate(
		components.CustomerSessionCustomerExternalIDCreate{ExternalCustomerID: orgSlug},
	)
	response, err := c.client.CustomerSessions.Create(ctx, request)
	if err != nil {
		return "", err
	}
	if response == nil || response.CustomerSession == nil {
		return "", nil
	}
	return strings.TrimSpace(response.CustomerSession.CustomerPortalURL), nil
}

type WebhookVerifier interface {
	Verify(headers http.Header, payload []byte, secret string) error
}

type noopClient struct{}

func (noopClient) CreateCheckoutSessionURL(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (noopClient) CreatePortalSessionURL(_ context.Context, _ string) (string, error) { return "", nil }

type polarWebhookVerifier struct{}

// Verify implements the Standard Webhooks specification (https://www.standardwebhooks.com),
// which Polar uses for webhook signing. The signed content is
// "{id}.{timestamp}.{payload}", the signature is a base64-encoded HMAC-SHA256
// digest, and the secret is a base64 value optionally prefixed with "whsec_".
// Both the canonical "webhook-*" headers and the legacy "svix-*" aliases are
// accepted.
func (v polarWebhookVerifier) Verify(headers http.Header, payload []byte, secret string) error {
	if secret == "" {
		return nil
	}

	msgID := standardWebhookHeader(headers, "webhook-id", "svix-id")
	timestamp := standardWebhookHeader(headers, "webhook-timestamp", "svix-timestamp")
	signatureHeader := standardWebhookHeader(headers, "webhook-signature", "svix-signature")
	if msgID == "" || timestamp == "" || signatureHeader == "" {
		return ErrInvalidWebhookSignature
	}

	if !withinTimestampTolerance(timestamp, 5*time.Minute) {
		return ErrInvalidWebhookSignature
	}

	key, err := decodeStandardWebhookSecret(secret)
	if err != nil {
		return ErrInvalidWebhookSignature
	}

	expected := computeStandardWebhookSignature(key, msgID, timestamp, payload)
	for _, candidate := range strings.Fields(signatureHeader) {
		// Each space-separated entry is "<version>,<base64 signature>".
		_, sig, found := strings.Cut(candidate, ",")
		if !found {
			sig = candidate
		}
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return nil
		}
	}

	return ErrInvalidWebhookSignature
}

func standardWebhookHeader(headers http.Header, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(headers.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

// decodeStandardWebhookSecret returns the raw HMAC key for a Standard Webhooks
// secret. The canonical form is base64 with an optional "whsec_" prefix; if the
// remainder is not valid base64 (e.g. an operator set a plain custom secret) the
// raw bytes are used instead.
func decodeStandardWebhookSecret(secret string) ([]byte, error) {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return nil, errors.New("empty webhook secret")
	}
	trimmed = strings.TrimPrefix(trimmed, "whsec_")
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		return decoded, nil
	}
	return []byte(trimmed), nil
}

func withinTimestampTolerance(timestamp string, tolerance time.Duration) bool {
	unixSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	diff := time.Since(time.Unix(unixSeconds, 0))
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func computeStandardWebhookSignature(key []byte, msgID, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(msgID + "." + timestamp + "."))
	_, _ = mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

func (s *service) CreateCheckoutSession(ctx context.Context, orgSlug string) (string, error) {
	if strings.TrimSpace(s.cfg.Polar.AccessToken) == "" {
		return fallbackBillingURL(s.cfg.BaseURL, orgSlug), nil
	}
	url, err := s.client.CreateCheckoutSessionURL(ctx, orgSlug)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}
	if strings.TrimSpace(url) == "" {
		return fallbackBillingURL(s.cfg.BaseURL, orgSlug), nil
	}
	return url, nil
}

func (s *service) CreatePortalSession(ctx context.Context, orgSlug string) (string, error) {
	if strings.TrimSpace(s.cfg.Polar.AccessToken) == "" {
		return fallbackBillingURL(s.cfg.BaseURL, orgSlug), nil
	}
	url, err := s.client.CreatePortalSessionURL(ctx, orgSlug)
	if err != nil {
		return "", fmt.Errorf("create billing portal session: %w", err)
	}
	if strings.TrimSpace(url) == "" {
		return fallbackBillingURL(s.cfg.BaseURL, orgSlug), nil
	}
	return url, nil
}

func (s *service) VerifyAndDispatchWebhook(ctx context.Context, headers http.Header, payload []byte) error {
	if strings.TrimSpace(s.cfg.Polar.WebhookSecret) == "" && s.cfg.IsProduction() {
		// Fail closed: never accept unsigned webhooks in production.
		return fmt.Errorf("verify polar webhook: %w", ErrInvalidWebhookSignature)
	}
	if err := s.verifier.Verify(headers, payload, s.cfg.Polar.WebhookSecret); err != nil {
		return fmt.Errorf("verify polar webhook: %w", err)
	}
	if _, err := s.dispatchWebhookEvent(ctx, payload); err != nil {
		return fmt.Errorf("dispatch polar webhook: %w", err)
	}
	return nil
}

func (s *service) BindWebhookEventStore(app core.App) {
	s.storeMu.Lock()
	defer s.storeMu.Unlock()
	s.store = newPocketBaseWebhookEventStore(app)
	s.stateStore = newPocketBaseBillingStateStore(app)
}

type billingWebhookEvent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type webhookDispatchResult struct {
	EventType string
	Family    string
	Handled   bool
}

func (s *service) dispatchWebhookEvent(ctx context.Context, payload []byte) (webhookDispatchResult, error) {
	var event billingWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return webhookDispatchResult{}, fmt.Errorf("parse billing webhook payload: %w", err)
	}

	result := webhookDispatchResult{EventType: event.Type}
	family := billingEventFamily(event.Type)
	result.Family = family

	s.storeMu.Lock()
	store := s.store
	stateStore := s.stateStore
	s.storeMu.Unlock()
	dedupeKey := buildBillingWebhookDedupeKey(event, payload)
	duplicate, err := store.IsDuplicate(ctx, "polar", dedupeKey)
	if err != nil {
		return result, fmt.Errorf("check billing webhook idempotency: %w", err)
	}
	if duplicate {
		return result, nil
	}

	handler, ok := s.billingEventFamilyHandlers()[family]
	if ok {
		if err := handler(event); err != nil {
			return result, err
		}
		result.Handled = true
	}

	if result.Handled {
		if err := stateStore.Apply(ctx, event); err != nil {
			return result, fmt.Errorf("persist billing state changes: %w", err)
		}
	}
	if err := store.Append(ctx, buildWebhookEventAuditRecord("polar", result, event, payload)); err != nil {
		return result, fmt.Errorf("persist billing webhook event: %w", err)
	}
	return result, nil
}

func buildWebhookEventAuditRecord(provider string, result webhookDispatchResult, event billingWebhookEvent, payload []byte) webhookEventAuditRecord {
	excerpt := strings.TrimSpace(string(payload))
	if len(excerpt) > 512 {
		excerpt = excerpt[:512]
	}
	hash := sha256.Sum256(payload)
	return webhookEventAuditRecord{
		Provider:       provider,
		EventType:      result.EventType,
		Family:         result.Family,
		Handled:        result.Handled,
		DedupeKey:      buildBillingWebhookDedupeKey(event, payload),
		PayloadHash:    hex.EncodeToString(hash[:]),
		PayloadExcerpt: excerpt,
		ReceivedAt:     time.Now().UTC(),
	}
}

type webhookEventAuditRecord struct {
	Provider       string
	EventType      string
	Family         string
	Handled        bool
	DedupeKey      string
	PayloadHash    string
	PayloadExcerpt string
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
	entry.Set("dedupe_key", record.DedupeKey)
	entry.Set("payload_hash", record.PayloadHash)
	entry.Set("payload_excerpt", record.PayloadExcerpt)
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

func billingEventFamily(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return ""
	}
	family, _, _ := strings.Cut(eventType, ".")
	return family
}

type billingEventHandler func(event billingWebhookEvent) error

type webhookSideEffects struct {
	handledByFamily map[string]int
	lastHandledType string
	lastHandledAt   time.Time
}

func (s *service) billingEventFamilyHandlers() map[string]billingEventHandler {
	return map[string]billingEventHandler{
		"subscription": s.handleSubscriptionEvent,
		"order":        s.handleOrderEvent,
		"customer":     s.handleCustomerEvent,
	}
}

func (s *service) handleSubscriptionEvent(event billingWebhookEvent) error {
	s.recordHandledEvent("subscription", event.Type)
	return nil
}

func (s *service) handleOrderEvent(event billingWebhookEvent) error {
	s.recordHandledEvent("order", event.Type)
	return nil
}

func (s *service) handleCustomerEvent(event billingWebhookEvent) error {
	s.recordHandledEvent("customer", event.Type)
	return nil
}

type billingStateStore interface {
	Apply(ctx context.Context, event billingWebhookEvent) error
}

type noopBillingStateStore struct{}

func (noopBillingStateStore) Apply(_ context.Context, _ billingWebhookEvent) error {
	return nil
}

type pocketBaseBillingStateStore struct {
	app core.App
}

func newPocketBaseBillingStateStore(app core.App) billingStateStore {
	if app == nil {
		return noopBillingStateStore{}
	}
	return pocketBaseBillingStateStore{app: app}
}

func (s pocketBaseBillingStateStore) Apply(_ context.Context, event billingWebhookEvent) error {
	eventData := event.Data
	if len(eventData) == 0 {
		return nil
	}
	orgID, orgSlug := s.resolveOrganization(eventData)
	if orgID == "" {
		return nil
	}
	if err := s.applySubscriptionUpsert(orgID, event); err != nil {
		return err
	}
	if err := s.applyInvoiceUpsert(orgID, event); err != nil {
		return err
	}
	if err := s.applyOrganizationBillingDenorm(orgID, eventData); err != nil {
		return err
	}
	_ = orgSlug
	return nil
}

func (s pocketBaseBillingStateStore) resolveOrganization(data map[string]any) (string, string) {
	candidates := []string{
		lookupNestedString(data, "organization_id"),
		lookupNestedString(data, "organization.id"),
		lookupNestedString(data, "customer.external_id"),
		lookupNestedString(data, "external_customer_id"),
		lookupNestedString(data, "customer_external_id"),
		lookupNestedString(data, "metadata.organization_id"),
		lookupNestedString(data, "metadata.org_id"),
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if record, err := s.app.FindRecordById("organizations", candidate); err == nil && record != nil {
			return record.Id, record.GetString("slug")
		}
		record, err := s.app.FindFirstRecordByFilter("organizations", "slug = {:slug}", dbx.Params{"slug": candidate})
		if err == nil && record != nil {
			return record.Id, record.GetString("slug")
		}
	}
	return "", ""
}

func (s pocketBaseBillingStateStore) applySubscriptionUpsert(orgID string, event billingWebhookEvent) error {
	subscriptionID := strings.TrimSpace(firstNonEmpty(
		lookupNestedString(event.Data, "subscription_id"),
		lookupNestedString(event.Data, "subscription.id"),
		lookupNestedString(event.Data, "id"),
	))
	if subscriptionID == "" || !strings.HasPrefix(event.Type, "subscription.") {
		return nil
	}
	collection, err := s.app.FindCollectionByNameOrId("subscriptions")
	if err != nil {
		return fmt.Errorf("find subscriptions collection: %w", err)
	}
	record, err := s.app.FindFirstRecordByFilter("subscriptions", "provider_subscription_id = {:id}", dbx.Params{"id": subscriptionID})
	if err != nil || record == nil {
		record = core.NewRecord(collection)
		record.Set("provider", "polar")
		record.Set("provider_subscription_id", subscriptionID)
	}
	record.Set("organization", orgID)
	if mappedStatus := mapSubscriptionStatus(lookupNestedString(event.Data, "status")); mappedStatus != "" {
		record.Set("status", mappedStatus)
	}
	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("save subscription billing state: %w", err)
	}
	return nil
}

func (s pocketBaseBillingStateStore) applyInvoiceUpsert(orgID string, event billingWebhookEvent) error {
	invoiceID := strings.TrimSpace(firstNonEmpty(
		lookupNestedString(event.Data, "invoice_id"),
		lookupNestedString(event.Data, "invoice.id"),
		lookupNestedString(event.Data, "id"),
	))
	if invoiceID == "" || !strings.HasPrefix(event.Type, "order.") {
		return nil
	}
	collection, err := s.app.FindCollectionByNameOrId("invoices")
	if err != nil {
		return fmt.Errorf("find invoices collection: %w", err)
	}
	record, err := s.app.FindFirstRecordByFilter("invoices", "provider_invoice_id = {:id}", dbx.Params{"id": invoiceID})
	if err != nil || record == nil {
		record = core.NewRecord(collection)
		record.Set("provider_invoice_id", invoiceID)
	}
	record.Set("organization", orgID)
	if amountCents, ok := parseAmountCents(event.Data); ok {
		record.Set("amount_cents", amountCents)
	}
	if mappedStatus := mapInvoiceStatus(lookupNestedString(event.Data, "status")); mappedStatus != "" {
		record.Set("status", mappedStatus)
	}
	if subID := lookupNestedString(event.Data, "subscription.id"); strings.TrimSpace(subID) != "" {
		subRecord, findErr := s.app.FindFirstRecordByFilter("subscriptions", "provider_subscription_id = {:id}", dbx.Params{"id": strings.TrimSpace(subID)})
		if findErr == nil && subRecord != nil {
			record.Set("subscription", subRecord.Id)
		}
	}
	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("save invoice billing state: %w", err)
	}
	return nil
}

func (s pocketBaseBillingStateStore) applyOrganizationBillingDenorm(orgID string, data map[string]any) error {
	org, err := s.app.FindRecordById("organizations", orgID)
	if err != nil || org == nil {
		return nil
	}
	collection := org.Collection()
	setIfField := func(name, value string) {
		if collection.Fields.GetByName(name) != nil && strings.TrimSpace(value) != "" {
			org.Set(name, strings.TrimSpace(value))
		}
	}
	setIfField("polar_customer_id", firstNonEmpty(
		lookupNestedString(data, "customer_id"),
		lookupNestedString(data, "customer.id"),
	))
	setIfField("polar_subscription_id", firstNonEmpty(
		lookupNestedString(data, "subscription_id"),
		lookupNestedString(data, "subscription.id"),
	))
	setIfField("polar_product_id", firstNonEmpty(
		lookupNestedString(data, "product_id"),
		lookupNestedString(data, "product.id"),
	))
	setIfField("polar_price_id", firstNonEmpty(
		lookupNestedString(data, "price_id"),
		lookupNestedString(data, "price.id"),
	))
	setIfField("plan", firstNonEmpty(
		lookupNestedString(data, "product.slug"),
		lookupNestedString(data, "product.name"),
		lookupNestedString(data, "product.id"),
		lookupNestedString(data, "plan"),
	))
	setIfField("subscription_status", mapSubscriptionStatus(lookupNestedString(data, "status")))
	return s.app.SaveNoValidate(org)
}

func mapSubscriptionStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "trialing", "active", "past_due", "canceled":
		return strings.ToLower(strings.TrimSpace(raw))
	case "cancelled":
		return "canceled"
	default:
		return ""
	}
}

func mapInvoiceStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "open", "draft", "pending":
		return "open"
	case "paid":
		return "paid"
	case "void", "voided", "cancelled", "canceled", "uncollectible":
		return "void"
	default:
		return ""
	}
}

func parseAmountCents(data map[string]any) (float64, bool) {
	for _, key := range []string{"amount_cents", "amount", "total_amount"} {
		value, ok := lookupNestedValue(data, key)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed, true
		case float32:
			return float64(typed), true
		case int:
			return float64(typed), true
		case int64:
			return float64(typed), true
		case json.Number:
			fv, err := typed.Float64()
			if err == nil {
				return fv, true
			}
		case string:
			fv, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
			if err == nil {
				return fv, true
			}
		}
	}
	return 0, false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func lookupNestedString(payload map[string]any, path string) string {
	value, ok := lookupNestedValue(payload, path)
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return ""
	}
}

func lookupNestedValue(payload map[string]any, path string) (any, bool) {
	current := any(payload)
	for _, segment := range strings.Split(path, ".") {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		value, exists := obj[segment]
		if !exists {
			return nil, false
		}
		current = value
	}
	return current, true
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

func fallbackBillingURL(baseURL, orgSlug string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = ""
	}
	return fmt.Sprintf("%s/org/%s/billing?provider=unconfigured", base, orgSlug)
}

func (s *service) handledCountForFamily(family string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.effects.handledByFamily == nil {
		return 0
	}
	return s.effects.handledByFamily[family]
}

func buildBillingWebhookDedupeKey(event billingWebhookEvent, payload []byte) string {
	seed := firstNonEmpty(
		lookupNestedString(event.Data, "event_id"),
		lookupNestedString(event.Data, "id"),
		lookupNestedString(event.Data, "subscription.id"),
		lookupNestedString(event.Data, "invoice.id"),
		lookupNestedString(event.Data, "customer.id"),
	)
	if seed == "" {
		hash := sha256.Sum256(payload)
		seed = hex.EncodeToString(hash[:])
	}
	return strings.TrimSpace(event.Type) + ":" + seed
}
