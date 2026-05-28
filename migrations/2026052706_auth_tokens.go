package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		authTokens := core.NewBaseCollection("auth_tokens")
		authTokens.Fields.Add(
			&core.TextField{Name: "token", Required: true, Max: 128},
			&core.SelectField{Name: "kind", Required: true, Values: []string{"reset", "verify"}, MaxSelect: 1},
			&core.EmailField{Name: "email"},
			&core.DateField{Name: "expires_at", Required: true},
			&core.DateField{Name: "consumed_at"},
		)
		authTokens.AddIndex("idx_auth_tokens_token_kind", true, "token,kind", "")
		return app.Save(authTokens)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("auth_tokens")
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	})
}
