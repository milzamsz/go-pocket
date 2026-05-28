package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			return fmt.Errorf("find organizations collection: %w", err)
		}

		collection.Fields.Add(
			&core.TextField{Name: "plan", Max: 64},
			&core.SelectField{Name: "subscription_status", MaxSelect: 1, Values: []string{"trialing", "active", "past_due", "canceled"}},
		)
		collection.AddIndex("idx_organizations_subscription_status", false, "subscription_status", "")

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("plan")
		collection.Fields.RemoveByName("subscription_status")
		collection.RemoveIndex("idx_organizations_subscription_status")
		return app.Save(collection)
	})
}
