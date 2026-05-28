package auth

import (
	"context"
	"testing"
	"time"

	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/require"
)

func TestCleanupExpiredOrConsumedTokens_RemovesOnlyStaleTokens(t *testing.T) {
	app := testutil.NewTestApp(t)
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)

	seedAuthToken(t, app, "reset_expiredToken", authTokenKindReset, "expired@example.com", now.Add(-time.Minute), nil)
	consumedAt := now.Add(-2 * time.Minute)
	seedAuthToken(t, app, "verify_consumedToken", authTokenKindVerify, "consumed@example.com", now.Add(time.Hour), &consumedAt)
	seedAuthToken(t, app, "reset_activeToken", authTokenKindReset, "active@example.com", now.Add(time.Hour), nil)

	deleted, err := CleanupExpiredOrConsumedTokens(context.Background(), app, now)
	require.NoError(t, err)
	require.Equal(t, 2, deleted)

	_, err = app.FindFirstRecordByFilter("auth_tokens", "token = {:token}", dbx.Params{"token": "reset_expiredToken"})
	require.Error(t, err)

	_, err = app.FindFirstRecordByFilter("auth_tokens", "token = {:token}", dbx.Params{"token": "verify_consumedToken"})
	require.Error(t, err)

	active := findAuthToken(t, app, "reset_activeToken", authTokenKindReset)
	require.Equal(t, "active@example.com", active.GetString("email"))
	require.LessOrEqual(t, active.GetDateTime("consumed_at").Time().Unix(), int64(0))
}
