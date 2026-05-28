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
			&core.TextField{Name: "polar_customer_id", Max: 128},
			&core.TextField{Name: "polar_subscription_id", Max: 128},
			&core.TextField{Name: "polar_product_id", Max: 128},
			&core.TextField{Name: "polar_price_id", Max: 128},
		)
		collection.AddIndex("idx_organizations_polar_customer_id", false, "polar_customer_id", "")
		collection.AddIndex("idx_organizations_polar_subscription_id", false, "polar_subscription_id", "")

		if err := app.Save(collection); err != nil {
			return fmt.Errorf("save organizations collection: %w", err)
		}
		return nil
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			return err
		}

		for _, field := range []string{
			"polar_customer_id",
			"polar_subscription_id",
			"polar_product_id",
			"polar_price_id",
		} {
			collection.Fields.RemoveByName(field)
		}
		collection.RemoveIndex("idx_organizations_polar_customer_id")
		collection.RemoveIndex("idx_organizations_polar_subscription_id")

		return app.Save(collection)
	})
}
