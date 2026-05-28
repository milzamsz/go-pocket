package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.Add(
			&core.TextField{Name: "name", Max: 120},
			&core.SelectField{Name: "system_role", Values: []string{"user", "admin"}, MaxSelect: 1},
		)
		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.RemoveByName("name")
		users.Fields.RemoveByName("system_role")
		return app.Save(users)
	})
}
