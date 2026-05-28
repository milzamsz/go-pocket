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

		subscriptions := core.NewBaseCollection("subscriptions")
		subscriptions.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.TextField{Name: "provider", Required: true, Max: 32},
			&core.TextField{Name: "provider_subscription_id", Required: true, Max: 128},
			&core.SelectField{Name: "status", Values: []string{"trialing", "active", "past_due", "canceled"}, MaxSelect: 1},
		)
		subscriptions.AddIndex("idx_subscriptions_org_provider", true, "organization,provider", "")
		subscriptions.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		subscriptions.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		if err := app.Save(subscriptions); err != nil {
			return err
		}

		invoices := core.NewBaseCollection("invoices")
		invoices.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "subscription", CollectionId: subscriptions.Id, MaxSelect: 1, CascadeDelete: true},
			&core.TextField{Name: "provider_invoice_id", Required: true, Max: 128},
			&core.NumberField{Name: "amount_cents", Required: true},
			&core.SelectField{Name: "status", Values: []string{"open", "paid", "void"}, MaxSelect: 1},
		)
		invoices.AddIndex("idx_invoices_provider_id", true, "provider_invoice_id", "")
		invoices.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		invoices.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		return app.Save(invoices)
	}, func(app core.App) error {
		for _, name := range []string{"invoices", "subscriptions"} {
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
