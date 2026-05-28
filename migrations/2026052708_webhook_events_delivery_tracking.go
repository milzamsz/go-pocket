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

		collection.Fields.Add(
			&core.TextField{Name: "status", Max: 32},
			&core.TextField{Name: "message_id", Max: 128},
			&core.TextField{Name: "recipient", Max: 255},
			&core.DateField{Name: "occurred_at"},
		)
		collection.AddIndex("idx_webhook_events_provider_message_id", false, "provider,message_id", "")
		collection.AddIndex("idx_webhook_events_status_occurred", false, "status,occurred_at", "")

		if err := app.Save(collection); err != nil {
			return fmt.Errorf("save webhook_events collection: %w", err)
		}
		return nil
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("webhook_events")
		if err != nil {
			return nil
		}

		collection.Fields.RemoveByName("status")
		collection.Fields.RemoveByName("message_id")
		collection.Fields.RemoveByName("recipient")
		collection.Fields.RemoveByName("occurred_at")
		collection.RemoveIndex("idx_webhook_events_provider_message_id")
		collection.RemoveIndex("idx_webhook_events_status_occurred")

		if err := app.Save(collection); err != nil {
			return fmt.Errorf("save webhook_events collection rollback: %w", err)
		}
		return nil
	})
}
