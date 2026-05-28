package migrations_test

import (
	"testing"

	_ "github.com/milzamsz/go-pocket/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestTenancyCollectionsHaveHardenedWriteRules(t *testing.T) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: t.TempDir(),
	})

	require.NoError(t, app.Bootstrap())
	t.Cleanup(func() {
		_ = app.ResetBootstrapState()
	})
	require.NoError(t, app.RunAllMigrations())

	assertWriteRules := func(collectionName string) {
		t.Helper()
		collection, err := app.FindCollectionByNameOrId(collectionName)
		require.NoError(t, err)
		require.NotNil(t, collection.CreateRule, "create rule missing for %s", collectionName)
		require.NotNil(t, collection.UpdateRule, "update rule missing for %s", collectionName)
		require.NotNil(t, collection.DeleteRule, "delete rule missing for %s", collectionName)
	}

	assertWriteRules("organizations")
	assertWriteRules("organization_members")
	assertWriteRules("invitations")
	assertWriteRules("subscriptions")
	assertWriteRules("invoices")
}
