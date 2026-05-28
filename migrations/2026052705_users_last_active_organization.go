package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return fmt.Errorf("find users collection: %w", err)
		}
		orgs, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			return fmt.Errorf("find organizations collection: %w", err)
		}

		if users.Fields.GetByName("last_active_organization") == nil {
			users.Fields.Add(&core.RelationField{Name: "last_active_organization", CollectionId: orgs.Id, MaxSelect: 1})
			if err := app.Save(users); err != nil {
				return fmt.Errorf("save users collection relation field: %w", err)
			}
		}
		return nil
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return fmt.Errorf("find users collection: %w", err)
		}
		users.Fields.RemoveByName("last_active_organization")
		if err := app.Save(users); err != nil {
			return fmt.Errorf("remove users relation field: %w", err)
		}
		return nil
	})
}
