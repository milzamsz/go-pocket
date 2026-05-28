package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("webhook_events")
		if err != nil {
			return fmt.Errorf("find webhook_events collection: %w", err)
		}

		collection.Fields.Add(&core.TextField{Name: "dedupe_key", Max: 200})
		collection.AddIndex("idx_webhook_events_provider_dedupe_key", true, "provider,dedupe_key", "dedupe_key != ''")

		if err := app.Save(collection); err != nil {
			return fmt.Errorf("save webhook_events collection: %w", err)
		}
		return nil
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("webhook_events")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("dedupe_key")
		collection.RemoveIndex("idx_webhook_events_provider_dedupe_key")

		return app.Save(collection)
	})
}
