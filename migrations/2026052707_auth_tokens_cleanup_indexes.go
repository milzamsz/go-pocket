package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("auth_tokens")
		if err != nil {
			return err
		}

		collection.AddIndex("idx_auth_tokens_expires_at", false, "expires_at", "")
		collection.AddIndex("idx_auth_tokens_consumed_at", false, "consumed_at", "")
		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("auth_tokens")
		if err != nil {
			return err
		}

		collection.RemoveIndex("idx_auth_tokens_expires_at")
		collection.RemoveIndex("idx_auth_tokens_consumed_at")
		return app.Save(collection)
	})
}
