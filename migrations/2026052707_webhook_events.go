package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		webhookEvents := core.NewBaseCollection("webhook_events")
		webhookEvents.Fields.Add(
			&core.TextField{Name: "provider", Required: true, Max: 32},
			&core.TextField{Name: "event_type", Required: true, Max: 128},
			&core.TextField{Name: "family", Max: 64},
			&core.BoolField{Name: "handled"},
			&core.TextField{Name: "payload_hash", Required: true, Min: 64, Max: 64},
			&core.TextField{Name: "payload_excerpt", Max: 512},
			&core.DateField{Name: "received_at", Required: true},
		)
		webhookEvents.AddIndex("idx_webhook_events_provider_received", false, "provider,received_at", "")
		webhookEvents.AddIndex("idx_webhook_events_event_type", false, "event_type", "")
		webhookEvents.ListRule = pointer("@request.auth.id != ''")
		webhookEvents.ViewRule = pointer("@request.auth.id != ''")
		webhookEvents.CreateRule = nil
		webhookEvents.UpdateRule = nil
		webhookEvents.DeleteRule = nil
		return app.Save(webhookEvents)
	}, func(app core.App) error {
		c, err := app.FindCollectionByNameOrId("webhook_events")
		if err == nil {
			return app.Delete(c)
		}
		return nil
	})
}
