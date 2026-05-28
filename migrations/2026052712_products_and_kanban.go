package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		orgs, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			return err
		}

		// 1. Products Collection
		products := core.NewBaseCollection("products")
		products.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.SelectField{Name: "category", Required: true, Values: []string{"electronics", "apparel", "home", "toys"}, MaxSelect: 1},
			&core.NumberField{Name: "price", Required: true}, // represented in cents
			&core.NumberField{Name: "stock", Required: true},
			&core.BoolField{Name: "active"},
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		products.AddIndex("idx_products_org", false, "organization", "")
		products.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		products.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		products.CreateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		products.UpdateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		products.DeleteRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin'")

		if err := app.Save(products); err != nil {
			return err
		}

		// 2. Kanban Columns Collection
		cols := core.NewBaseCollection("kanban_columns")
		cols.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 64},
			&core.NumberField{Name: "position", Required: true},
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		cols.AddIndex("idx_kanban_cols_org", false, "organization", "")
		cols.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		cols.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		cols.CreateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		cols.UpdateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		cols.DeleteRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")

		if err := app.Save(cols); err != nil {
			return err
		}

		// 3. Kanban Cards Collection
		cards := core.NewBaseCollection("kanban_cards")
		cards.Fields.Add(
			&core.TextField{Name: "title", Required: true, Max: 256},
			&core.TextField{Name: "description", Max: 1024},
			&core.TextField{Name: "badge", Max: 64},
			&core.NumberField{Name: "position", Required: true},
			&core.RelationField{Name: "column", Required: true, CollectionId: cols.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "assignee", CollectionId: "_pb_users_auth_", MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		cards.AddIndex("idx_kanban_cards_org", false, "organization", "")
		cards.AddIndex("idx_kanban_cards_col", false, "column", "")
		cards.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		cards.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		cards.CreateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		cards.UpdateRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")
		cards.DeleteRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin|member'")

		return app.Save(cards)
	}, func(app core.App) error {
		for _, name := range []string{"kanban_cards", "kanban_columns", "products"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(c); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
