package testutil

import (
	"testing"

	_ "github.com/milzamsz/go-pocket/migrations"
	"github.com/pocketbase/pocketbase"
)

func NewTestApp(tb testing.TB) *pocketbase.PocketBase {
	tb.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: tb.TempDir(),
	})

	if err := app.Bootstrap(); err != nil {
		tb.Fatalf("bootstrap pocketbase test app: %v", err)
	}

	if err := app.RunAllMigrations(); err != nil {
		tb.Fatalf("run migrations for pocketbase test app: %v", err)
	}

	tb.Cleanup(func() {
		_ = app.ResetBootstrapState()
	})

	return app
}
