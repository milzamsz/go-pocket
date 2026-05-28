package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		orgs := core.NewBaseCollection("organizations")
		orgs.Fields.Add(
			&core.TextField{Name: "slug", Required: true, Max: 64},
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.RelationField{Name: "owner", Required: true, CollectionId: "_pb_users_auth_", MaxSelect: 1},
		)
		orgs.AddIndex("idx_organizations_slug", true, "slug", "")
		if err := app.Save(orgs); err != nil {
			return err
		}

		members := core.NewBaseCollection("organization_members")
		members.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "user", Required: true, CollectionId: "_pb_users_auth_", MaxSelect: 1, CascadeDelete: true},
			&core.SelectField{Name: "role", Required: true, Values: []string{"owner", "admin", "member", "viewer"}, MaxSelect: 1},
		)
		members.AddIndex("idx_org_member_unique", true, "organization,user", "")
		members.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		members.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		if err := app.Save(members); err != nil {
			return err
		}

		invitations := core.NewBaseCollection("invitations")
		invitations.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.EmailField{Name: "email", Required: true},
			&core.SelectField{Name: "role", Required: true, Values: []string{"admin", "member", "viewer"}, MaxSelect: 1},
			&core.TextField{Name: "token", Required: true, Max: 96},
		)
		invitations.AddIndex("idx_invitation_token", true, "token", "")
		invitations.ListRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		invitations.ViewRule = pointer("organization.organization_members_via_organization.user ?= @request.auth.id")
		return app.Save(invitations)
	}, func(app core.App) error {
		for _, name := range []string{"invitations", "organization_members", "organizations"} {
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
