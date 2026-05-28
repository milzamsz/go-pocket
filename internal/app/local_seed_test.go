package app

import (
	"context"
	"testing"

	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestSeedLocalAccounts_CreatesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	pb := testutil.NewTestApp(t)
	ctx := context.Background()

	adminCred, userCred, userAliasCred, err := seedLocalAccounts(ctx, pb)
	require.NoError(t, err)
	require.Equal(t, adminEmail, adminCred.Email)
	require.Equal(t, userEmail, userCred.Email)
	require.Equal(t, userAliasEmail, userAliasCred.Email)

	adminCount, err := pb.CountRecords(core.CollectionNameSuperusers, dbx.HashExp{"email": adminEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), adminCount)
	userCount, err := pb.CountRecords("users", dbx.HashExp{"email": userEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), userCount)
	userAliasCount, err := pb.CountRecords("users", dbx.HashExp{"email": userAliasEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), userAliasCount)

	adminCred2, userCred2, userAliasCred2, err := seedLocalAccounts(ctx, pb)
	require.NoError(t, err)
	require.Equal(t, adminCred.Password, adminCred2.Password)
	require.Equal(t, userCred.Password, userCred2.Password)
	require.Equal(t, userAliasCred.Password, userAliasCred2.Password)

	adminCount, err = pb.CountRecords(core.CollectionNameSuperusers, dbx.HashExp{"email": adminEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), adminCount)
	userCount, err = pb.CountRecords("users", dbx.HashExp{"email": userEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), userCount)
	userAliasCount, err = pb.CountRecords("users", dbx.HashExp{"email": userAliasEmail})
	require.NoError(t, err)
	require.Equal(t, int64(1), userAliasCount)
}

func TestUpsertAuthUser_FallbackOnRejectedPassword(t *testing.T) {
	t.Parallel()

	pb := testutil.NewTestApp(t)
	ctx := context.Background()
	cred, err := upsertAuthUser(ctx, pb, "fallback@test.com", "x", "Fallback#123")
	require.NoError(t, err)
	require.True(t, cred.UsedFallback)
	require.Equal(t, "Fallback#123", cred.Password)
}
