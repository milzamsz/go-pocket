package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type Service interface {
	CurrentUser(ctx context.Context) (*domain.User, error)
	Login(ctx context.Context, email string, password string) (*AuthSession, error)
	Signup(ctx context.Context, name string, email string, password string) (*AuthSession, error)
	Logout(ctx context.Context, userID string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, password string, confirmPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	UpdateProfile(ctx context.Context, userID string, name string, email string) error
	ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string, confirmPassword string) error
	UpdateTwoFactor(ctx context.Context, userID string, enabled bool) error
	OAuthRedirectURL(ctx context.Context, provider string, state string) (string, error)
	OAuthCallback(ctx context.Context, provider string, code string) (*AuthSession, error)
}

type AuthSession struct {
	User  *domain.User
	Token string
}

type emailSender interface {
	SendPasswordReset(ctx context.Context, to string, token string) error
	SendEmailVerification(ctx context.Context, to string, token string) error
}

const (
	authTokenKindReset  = "reset"
	authTokenKindVerify = "verify"
)

var (
	errInvalidResetToken  = errors.New("invalid reset token")
	errInvalidVerifyToken = errors.New("invalid verification token")

	// ErrIncorrectPassword is returned when a password change is attempted with
	// an incorrect current password.
	ErrIncorrectPassword = errors.New("current password is incorrect")
)

func New(app core.App) Service {
	return NewWithConfig(app, config.Config{})
}

func NewWithConfig(app core.App, cfg config.Config) Service {
	return NewWithConfigAndDependencies(app, cfg, nil, time.Now, nil, nil)
}

type tokenGenerator func() (string, error)
type nowFunc func() time.Time

type service struct {
	app           core.App
	generateToken tokenGenerator
	now           nowFunc
	email         emailSender
	oauth         map[string]oauthProvider
}

func NewWithDependencies(app core.App, generator tokenGenerator, now nowFunc, sender emailSender) Service {
	return NewWithConfigAndDependencies(app, config.Config{}, generator, now, sender, nil)
}

func NewWithConfigAndDependencies(
	app core.App,
	cfg config.Config,
	generator tokenGenerator,
	now nowFunc,
	sender emailSender,
	httpClient oauthHTTPClient,
) Service {
	if generator == nil {
		generator = defaultTokenGenerator
	}
	if now == nil {
		now = time.Now
	}
	return &service{
		app:           app,
		generateToken: generator,
		now:           now,
		email:         sender,
		oauth:         newOAuthProviders(cfg, httpClient),
	}
}

func (s *service) CurrentUser(_ context.Context) (*domain.User, error) {
	// TODO: Resolve current user from PocketBase auth context.
	return nil, domain.ErrUnauthenticated
}

func (s *service) Login(_ context.Context, email string, password string) (*AuthSession, error) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email and password are required")
	}
	if s.app == nil {
		return nil, errors.New("auth store is not configured")
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	record, err := s.app.FindAuthRecordByEmail("users", normalizedEmail)
	if err != nil || record == nil {
		record, err = s.app.FindAuthRecordByEmail(core.CollectionNameSuperusers, normalizedEmail)
	}
	if err != nil || record == nil {
		return nil, domain.ErrUnauthenticated
	}
	if !record.ValidatePassword(password) {
		return nil, domain.ErrUnauthenticated
	}
	token, err := record.NewAuthToken()
	if err != nil {
		return nil, fmt.Errorf("issue auth token: %w", err)
	}
	return &AuthSession{
		User:  mapRecordToDomainUser(record),
		Token: token,
	}, nil
}

func (s *service) Signup(ctx context.Context, name string, email string, password string) (*AuthSession, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("name, email, and password are required")
	}
	if s.app == nil {
		return nil, errors.New("auth store is not configured")
	}
	trimmedName := strings.TrimSpace(name)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	users, err := s.app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, fmt.Errorf("find users collection: %w", err)
	}
	record := core.NewRecord(users)
	record.Set("name", trimmedName)
	record.Set("email", normalizedEmail)
	record.Set("password", password)
	record.Set("passwordConfirm", password)
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	token, err := record.NewAuthToken()
	if err != nil {
		return nil, fmt.Errorf("issue auth token: %w", err)
	}

	if !record.Verified() {
		verifySeed, err := s.generateToken()
		if err != nil {
			return nil, fmt.Errorf("generate verification token: %w", err)
		}
		verifyToken := "verify_" + verifySeed
		expiresAt := s.now().UTC().Add(24 * time.Hour)
		if err := s.createAuthToken(ctx, verifyToken, authTokenKindVerify, normalizedEmail, expiresAt); err != nil {
			return nil, fmt.Errorf("create verification token: %w", err)
		}
		if s.email != nil {
			if err := s.email.SendEmailVerification(ctx, normalizedEmail, verifyToken); err != nil {
				return nil, fmt.Errorf("send verification email: %w", err)
			}
		}
	}

	return &AuthSession{
		User:  mapRecordToDomainUser(record),
		Token: token,
	}, nil
}

func (s *service) Logout(_ context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return domain.ErrUnauthenticated
	}
	return nil
}

func (s *service) RequestPasswordReset(ctx context.Context, email string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return errors.New("email is required")
	}
	if s.app == nil {
		return errors.New("auth token store is not configured")
	}
	_, err := s.app.FindAuthRecordByEmail("users", normalizedEmail)
	if err != nil {
		return nil
	}
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	resetToken := "reset_" + token
	expiresAt := s.now().UTC().Add(time.Hour)
	if err := s.createAuthToken(ctx, resetToken, authTokenKindReset, normalizedEmail, expiresAt); err != nil {
		return err
	}
	if s.email != nil {
		if err := s.email.SendPasswordReset(ctx, normalizedEmail, resetToken); err != nil {
			return fmt.Errorf("send password reset email: %w", err)
		}
	}
	return nil
}

func (s *service) ResetPassword(ctx context.Context, token string, password string, confirmPassword string) error {
	normalizedToken := strings.TrimSpace(token)
	if normalizedToken == "" {
		return errors.New("reset token is required")
	}
	if s.app == nil {
		return errors.New("auth token store is not configured")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	if password != confirmPassword {
		return errors.New("password confirmation does not match")
	}

	if !isValidToken(normalizedToken, "reset_") {
		return errInvalidResetToken
	}
	record, err := s.findTokenRecord(normalizedToken, authTokenKindReset)
	if err != nil {
		return err
	}
	if !isTokenUsable(record, s.now()) {
		return errInvalidResetToken
	}
	record.Set("consumed_at", s.now().UTC())
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("consume reset token record: %w", err)
	}
	userRecord, err := s.app.FindAuthRecordByEmail("users", strings.ToLower(strings.TrimSpace(record.GetString("email"))))
	if err != nil {
		return fmt.Errorf("find user for reset token: %w", err)
	}
	userRecord.Set("password", password)
	userRecord.Set("passwordConfirm", confirmPassword)
	if err := s.app.SaveWithContext(ctx, userRecord); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) error {
	normalizedToken := strings.TrimSpace(token)
	if normalizedToken == "" {
		return errors.New("verification token is required")
	}
	if s.app == nil {
		return errors.New("auth token store is not configured")
	}

	if !isValidToken(normalizedToken, "verify_") {
		return errInvalidVerifyToken
	}
	record, err := s.findTokenRecord(normalizedToken, authTokenKindVerify)
	if err != nil {
		return err
	}
	if !isTokenUsable(record, s.now()) {
		return errInvalidVerifyToken
	}
	record.Set("consumed_at", s.now().UTC())
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("consume verification token record: %w", err)
	}
	userRecord, err := s.app.FindAuthRecordByEmail("users", strings.ToLower(strings.TrimSpace(record.GetString("email"))))
	if err != nil {
		return fmt.Errorf("find user for verification token: %w", err)
	}
	userRecord.Set("verified", true)
	if err := s.app.SaveWithContext(ctx, userRecord); err != nil {
		return fmt.Errorf("mark user as verified: %w", err)
	}
	return nil
}

func (s *service) createAuthToken(ctx context.Context, token string, kind string, email string, expiresAt time.Time) error {
	collection, err := s.app.FindCollectionByNameOrId("auth_tokens")
	if err != nil {
		return fmt.Errorf("find auth_tokens collection: %w", err)
	}
	record := core.NewRecord(collection)
	record.Set("token", hashAuthToken(token))
	record.Set("kind", kind)
	record.Set("email", email)
	record.Set("expires_at", expiresAt)
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("save auth token record: %w", err)
	}
	return nil
}

func (s *service) findTokenRecord(token string, kind string) (*core.Record, error) {
	record, err := s.app.FindFirstRecordByFilter(
		"auth_tokens",
		"token = {:token} && kind = {:kind}",
		dbx.Params{"token": hashAuthToken(token), "kind": kind},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if kind == authTokenKindVerify {
				return nil, errInvalidVerifyToken
			}
			return nil, errInvalidResetToken
		}
		return nil, fmt.Errorf("find auth token record: %w", err)
	}
	return record, nil
}

func isTokenUsable(record *core.Record, now time.Time) bool {
	if record.GetDateTime("consumed_at").Time().Unix() > 0 {
		return false
	}
	expiresAt := record.GetDateTime("expires_at").Time()
	return expiresAt.After(now.UTC())
}

func (s *service) UpdateProfile(ctx context.Context, userID string, name string, email string) error {
	if strings.TrimSpace(userID) == "" {
		return domain.ErrUnauthenticated
	}
	trimmedName := strings.TrimSpace(name)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if trimmedName == "" || normalizedEmail == "" {
		return errors.New("name and email are required")
	}
	if s.app == nil {
		return errors.New("auth store is not configured")
	}
	record, err := s.app.FindRecordById("users", userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	record.Set("name", trimmedName)
	record.SetEmail(normalizedEmail)
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string, confirmPassword string) error {
	if strings.TrimSpace(userID) == "" {
		return domain.ErrUnauthenticated
	}
	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
		return errors.New("current and new passwords are required")
	}
	if newPassword != confirmPassword {
		return errors.New("password confirmation does not match")
	}
	if s.app == nil {
		return errors.New("auth store is not configured")
	}
	record, err := s.app.FindRecordById("users", userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if !record.ValidatePassword(currentPassword) {
		return ErrIncorrectPassword
	}
	record.Set("password", newPassword)
	record.Set("passwordConfirm", newPassword)
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (s *service) UpdateTwoFactor(ctx context.Context, userID string, enabled bool) error {
	if strings.TrimSpace(userID) == "" {
		return domain.ErrUnauthenticated
	}
	if s.app == nil {
		return errors.New("auth store is not configured")
	}
	record, err := s.app.FindRecordById("users", userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	record.Set("two_factor_enabled", enabled)
	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return fmt.Errorf("update two-factor setting: %w", err)
	}
	return nil
}

// hashAuthToken derives the at-rest representation of a single-use auth token.
// Only the SHA-256 hash is persisted, so a leak of the auth_tokens table does
// not expose usable reset/verification tokens. The plaintext token is only ever
// delivered to the user via email.
func hashAuthToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func defaultTokenGenerator() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func isValidToken(token string, requiredPrefix string) bool {
	if !strings.HasPrefix(token, requiredPrefix) {
		return false
	}
	value := strings.TrimPrefix(token, requiredPrefix)
	if len(value) < 8 {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

func mapRecordToDomainUser(record *core.Record) *domain.User {
	if record == nil {
		return nil
	}
	return &domain.User{
		ID:    record.Id,
		Name:  strings.TrimSpace(record.GetString("name")),
		Email: strings.TrimSpace(record.Email()),
	}
}
