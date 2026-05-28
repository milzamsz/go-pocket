package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		writeRuleByOrg := "organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?~ 'owner|admin'"
		ownerOnlyByOrg := "organization.organization_members_via_organization.user ?= @request.auth.id && organization.organization_members_via_organization.role ?= 'owner'"
		organizationCreateRule := "@request.auth.id != ''"
		organizationOwnerRule := "owner ?= @request.auth.id"

		type ruleSet struct {
			create *string
			update *string
			delete *string
		}

		rules := map[string]ruleSet{
			"organizations": {
				create: pointer(organizationCreateRule),
				update: pointer(organizationOwnerRule),
				delete: pointer(organizationOwnerRule),
			},
			"organization_members": {
				create: pointer(writeRuleByOrg),
				update: pointer(writeRuleByOrg),
				delete: pointer(writeRuleByOrg),
			},
			"invitations": {
				create: pointer(writeRuleByOrg),
				update: pointer(writeRuleByOrg),
				delete: pointer(writeRuleByOrg),
			},
			"subscriptions": {
				create: pointer(writeRuleByOrg),
				update: pointer(writeRuleByOrg),
				delete: pointer(ownerOnlyByOrg),
			},
			"invoices": {
				create: pointer(writeRuleByOrg),
				update: pointer(writeRuleByOrg),
				delete: pointer(ownerOnlyByOrg),
			},
		}

		for collectionName, collectionRules := range rules {
			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				return fmt.Errorf("find %s collection: %w", collectionName, err)
			}

			collection.CreateRule = collectionRules.create
			collection.UpdateRule = collectionRules.update
			collection.DeleteRule = collectionRules.delete

			if err := app.Save(collection); err != nil {
				return fmt.Errorf("save %s rule hardening: %w", collectionName, err)
			}
		}

		return nil
	}, func(app core.App) error {
		for _, collectionName := range []string{"organizations", "organization_members", "invitations", "subscriptions", "invoices"} {
			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				return fmt.Errorf("find %s collection: %w", collectionName, err)
			}

			collection.CreateRule = nil
			collection.UpdateRule = nil
			collection.DeleteRule = nil

			if err := app.Save(collection); err != nil {
				return fmt.Errorf("revert %s rule hardening: %w", collectionName, err)
			}
		}
		return nil
	})
}
