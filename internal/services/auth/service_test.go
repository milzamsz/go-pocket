package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

type fakeEmailSender struct {
	resetRecipient  string
	resetToken      string
	verifyRecipient string
	verifyToken     string
}

func (f *fakeEmailSender) SendPasswordReset(_ context.Context, to string, token string) error {
	f.resetRecipient = to
	f.resetToken = token
	return nil
}

func (f *fakeEmailSender) SendEmailVerification(_ context.Context, to string, token string) error {
	f.verifyRecipient = to
	f.verifyToken = token
	return nil
}

func TestRequestPasswordReset_CreatesTokenAndSendsEmail(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	sender := &fakeEmailSender{}
	svc := NewWithDependencies(app, func() (string, error) { return "abcDEF12", nil }, func() time.Time {
		return time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	}, sender)

	err := svc.RequestPasswordReset(context.Background(), "User@Example.com")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", sender.resetRecipient)
	require.Equal(t, "reset_abcDEF12", sender.resetToken)

	record, err := app.FindFirstRecordByFilter(
		"auth_tokens",
		"token = {:token} && kind = {:kind}",
		dbx.Params{"token": "reset_abcDEF12", "kind": authTokenKindReset},
	)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", record.GetString("email"))
}

func TestSignup_CreatesVerificationTokenAndSendsEmail(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	sender := &fakeEmailSender{}
	svc := NewWithDependencies(app, func() (string, error) { return "XYZabc12", nil }, func() time.Time {
		return time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	}, sender)

	_, err := svc.Signup(context.Background(), "Alice", "User@Example.com", "secret123")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", sender.verifyRecipient)
	require.Equal(t, "verify_XYZabc12", sender.verifyToken)
}

func TestLogin_IssuesAuthTokenForValidCredentials(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	svc := NewWithDependencies(app, nil, nil, nil)

	session, err := svc.Login(context.Background(), "user@example.com", "secret123")
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, "user@example.com", session.User.Email)
	require.NotEmpty(t, session.Token)
}

func TestLogin_IssuesAuthTokenForValidSuperuserCredentials(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedSuperuser(t, app, "admin@test.com", "secret123")
	svc := NewWithDependencies(app, nil, nil, nil)

	session, err := svc.Login(context.Background(), "admin@test.com", "secret123")
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, "admin@test.com", session.User.Email)
	require.NotEmpty(t, session.Token)
}

func TestRequestPasswordReset_RequiresEmail(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := NewWithDependencies(app, nil, nil, nil)

	err := svc.RequestPasswordReset(context.Background(), "   ")
	require.EqualError(t, err, "email is required")
}

func TestRequestPasswordReset_ReturnsGeneratorError(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	svc := NewWithDependencies(app, func() (string, error) { return "", errors.New("boom") }, nil, nil)

	err := svc.RequestPasswordReset(context.Background(), "user@example.com")
	require.EqualError(t, err, "generate reset token: boom")
}

func TestResetPassword_FailsWithInvalidToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := NewWithDependencies(app, nil, nil, nil)

	err := svc.ResetPassword(context.Background(), "reset_bad/token", "secret123", "secret123")
	require.EqualError(t, err, "invalid reset token")
}

func TestVerifyEmail_FailsWithInvalidToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	svc := NewWithDependencies(app, nil, nil, nil)

	err := svc.VerifyEmail(context.Background(), "verify_bad/token")
	require.EqualError(t, err, "invalid verification token")
}

func TestResetPassword_ConsumesTokenOnSuccess(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "oldsecret")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	seedAuthToken(t, app, "reset_abcDEF12", authTokenKindReset, "user@example.com", now.Add(time.Hour), nil)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.ResetPassword(context.Background(), "reset_abcDEF12", "secret123", "secret123")
	require.NoError(t, err)

	record := findAuthToken(t, app, "reset_abcDEF12", authTokenKindReset)
	require.Equal(t, now.Unix(), record.GetDateTime("consumed_at").Time().Unix())
}

func TestResetPassword_FailsWithExpiredToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "oldsecret")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	seedAuthToken(t, app, "reset_abcDEF12", authTokenKindReset, "user@example.com", now.Add(-time.Minute), nil)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.ResetPassword(context.Background(), "reset_abcDEF12", "secret123", "secret123")
	require.EqualError(t, err, "invalid reset token")
}

func TestResetPassword_FailsWithConsumedToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "oldsecret")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	consumedAt := now.Add(-time.Minute)
	seedAuthToken(t, app, "reset_abcDEF12", authTokenKindReset, "user@example.com", now.Add(time.Hour), &consumedAt)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.ResetPassword(context.Background(), "reset_abcDEF12", "secret123", "secret123")
	require.EqualError(t, err, "invalid reset token")
}

func TestVerifyEmail_ConsumesTokenOnSuccess(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	seedAuthToken(t, app, "verify_abcDEF12", authTokenKindVerify, "user@example.com", now.Add(time.Hour), nil)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.VerifyEmail(context.Background(), "verify_abcDEF12")
	require.NoError(t, err)

	record := findAuthToken(t, app, "verify_abcDEF12", authTokenKindVerify)
	require.Equal(t, now.Unix(), record.GetDateTime("consumed_at").Time().Unix())
}

func TestVerifyEmail_FailsWithExpiredToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	seedAuthToken(t, app, "verify_abcDEF12", authTokenKindVerify, "user@example.com", now.Add(-time.Minute), nil)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.VerifyEmail(context.Background(), "verify_abcDEF12")
	require.EqualError(t, err, "invalid verification token")
}

func TestVerifyEmail_FailsWithConsumedToken(t *testing.T) {
	t.Parallel()

	app := testutil.NewTestApp(t)
	seedAuthUser(t, app, "user@example.com", "secret123")
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	consumedAt := now.Add(-time.Minute)
	seedAuthToken(t, app, "verify_abcDEF12", authTokenKindVerify, "user@example.com", now.Add(time.Hour), &consumedAt)
	svc := NewWithDependencies(app, nil, func() time.Time { return now }, nil)

	err := svc.VerifyEmail(context.Background(), "verify_abcDEF12")
	require.EqualError(t, err, "invalid verification token")
}

func findAuthToken(t *testing.T, app core.App, token string, kind string) *core.Record {
	t.Helper()
	record, err := app.FindFirstRecordByFilter(
		"auth_tokens",
		"token = {:token} && kind = {:kind}",
		dbx.Params{"token": token, "kind": kind},
	)
	require.NoError(t, err)
	return record
}

func seedAuthToken(t *testing.T, app core.App, token string, kind string, email string, expiresAt time.Time, consumedAt *time.Time) {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("auth_tokens")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("token", token)
	record.Set("kind", kind)
	record.Set("email", email)
	record.Set("expires_at", expiresAt.UTC())
	if consumedAt != nil {
		record.Set("consumed_at", consumedAt.UTC())
	}
	require.NoError(t, app.Save(record))
}

func seedAuthUser(t *testing.T, app core.App, email string, password string) *core.Record {
	t.Helper()
	users, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)
	record := core.NewRecord(users)
	record.Set("email", email)
	record.Set("password", password)
	record.Set("passwordConfirm", password)
	require.NoError(t, app.Save(record))
	return record
}

func seedSuperuser(t *testing.T, app core.App, email string, password string) *core.Record {
	t.Helper()
	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	require.NoError(t, err)
	record := core.NewRecord(superusers)
	record.Set("email", email)
	record.Set("password", password)
	record.Set("passwordConfirm", password)
	require.NoError(t, app.Save(record))
	return record
}
