package tenancy

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type PocketBaseRepository struct {
	app core.App
}

func NewPocketBaseRepository(app core.App) *PocketBaseRepository {
	return &PocketBaseRepository{app: app}
}

func (r *PocketBaseRepository) CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	if org.ID == "" {
		return domain.Organization{}, fmt.Errorf("organization id is required")
	}

	collection, err := r.app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return domain.Organization{}, fmt.Errorf("find organizations collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("id", org.ID)
	record.Set("slug", org.Slug)
	record.Set("name", org.Name)
	record.Set("owner", org.OwnerID)

	if err := r.app.SaveWithContext(ctx, record); err != nil {
		return domain.Organization{}, fmt.Errorf("save organization record: %w", err)
	}

	return domain.Organization{
		ID:      record.Id,
		Slug:    record.GetString("slug"),
		Name:    record.GetString("name"),
		OwnerID: record.GetString("owner"),
	}, nil
}

func (r *PocketBaseRepository) CreateMembership(ctx context.Context, membership domain.OrganizationMember) (domain.OrganizationMember, error) {
	collection, err := r.app.FindCollectionByNameOrId("organization_members")
	if err != nil {
		return domain.OrganizationMember{}, fmt.Errorf("find organization_members collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("organization", membership.OrganizationID)
	record.Set("user", membership.UserID)
	record.Set("role", string(membership.Role))

	if err := r.app.SaveWithContext(ctx, record); err != nil {
		return domain.OrganizationMember{}, fmt.Errorf("save organization member record: %w", err)
	}

	return domain.OrganizationMember{
		ID:             record.Id,
		OrganizationID: record.GetString("organization"),
		UserID:         record.GetString("user"),
		Role:           domain.OrgRole(record.GetString("role")),
	}, nil
}

func (r *PocketBaseRepository) ListMembersByOrg(ctx context.Context, orgID string) ([]domain.OrganizationMember, error) {
	collection, err := r.app.FindCollectionByNameOrId("organization_members")
	if err != nil {
		return nil, fmt.Errorf("find organization_members collection: %w", err)
	}

	query := r.app.RecordQuery(collection).
		AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgID})).
		WithContext(ctx)

	records := make([]*core.Record, 0)
	if err := query.All(&records); err != nil {
		return nil, fmt.Errorf("find organization members by org filter: %w", err)
	}

	out := make([]domain.OrganizationMember, 0, len(records))
	for _, record := range records {
		out = append(out, domain.OrganizationMember{
			ID:             record.Id,
			OrganizationID: record.GetString("organization"),
			UserID:         record.GetString("user"),
			Role:           domain.OrgRole(record.GetString("role")),
		})
	}

	return out, nil
}

func (r *PocketBaseRepository) FindMemberProfile(_ context.Context, orgID string, userID string) (domain.OrganizationMemberProfile, error) {
	member, err := r.app.FindFirstRecordByFilter(
		"organization_members",
		"organization = {:orgID} && user = {:userID}",
		dbx.Params{"orgID": orgID, "userID": userID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OrganizationMemberProfile{}, domain.ErrNotFound
		}
		return domain.OrganizationMemberProfile{}, fmt.Errorf("find organization member: %w", err)
	}

	user, err := r.app.FindRecordById("users", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OrganizationMemberProfile{}, domain.ErrNotFound
		}
		return domain.OrganizationMemberProfile{}, fmt.Errorf("find user for member profile: %w", err)
	}

	return domain.OrganizationMemberProfile{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           domain.OrgRole(member.GetString("role")),
		Name:           user.GetString("name"),
		Email:          user.GetString("email"),
	}, nil
}

func (r *PocketBaseRepository) InviteMember(_ context.Context, orgID string, email string, role domain.OrgRole) (domain.Invitation, error) {
	collection, err := r.app.FindCollectionByNameOrId("invitations")
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("find invitations collection: %w", err)
	}

	token, err := generateInvitationToken()
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("generate invitation token: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("organization", orgID)
	record.Set("email", email)
	record.Set("role", string(role))
	record.Set("token", token)

	if err := r.app.Save(record); err != nil {
		return domain.Invitation{}, fmt.Errorf("save invitation record: %w", err)
	}

	return domain.Invitation{
		ID:             record.Id,
		OrganizationID: record.GetString("organization"),
		Email:          record.GetString("email"),
		Role:           domain.OrgRole(record.GetString("role")),
		Token:          record.GetString("token"),
	}, nil
}

func (r *PocketBaseRepository) RemoveMember(_ context.Context, orgID string, userID string) error {
	record, err := r.findOrgMemberByUser(orgID, userID)
	if err != nil {
		return err
	}
	if err := r.app.Delete(record); err != nil {
		return fmt.Errorf("delete organization member: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) ChangeMemberRole(_ context.Context, orgID string, userID string, role domain.OrgRole) error {
	record, err := r.findOrgMemberByUser(orgID, userID)
	if err != nil {
		return err
	}
	record.Set("role", string(role))
	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("save organization member role: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) ResendInvitation(_ context.Context, orgID string, invitationID string) error {
	record, err := r.findInvitationByID(orgID, invitationID)
	if err != nil {
		return err
	}
	token, err := generateInvitationToken()
	if err != nil {
		return fmt.Errorf("generate invitation token: %w", err)
	}
	record.Set("token", token)
	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("save invitation record: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) RevokeInvitation(_ context.Context, orgID string, invitationID string) error {
	record, err := r.findInvitationByID(orgID, invitationID)
	if err != nil {
		return err
	}
	if err := r.app.Delete(record); err != nil {
		return fmt.Errorf("delete invitation: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) UpdateOrganizationSettings(_ context.Context, orgID string, name string) error {
	record, err := r.findOrganizationByID(orgID)
	if err != nil {
		return err
	}
	record.Set("name", name)
	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("save organization settings: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) TransferOrganizationOwnership(_ context.Context, orgID string, newOwnerUserID string) error {
	org, err := r.findOrganizationByID(orgID)
	if err != nil {
		return err
	}

	newOwnerMembership, err := r.findOrgMemberByUser(orgID, newOwnerUserID)
	if err != nil {
		return err
	}

	prevOwnerUserID := org.GetString("owner")
	if prevOwnerUserID != "" && prevOwnerUserID != newOwnerUserID {
		prevOwnerMembership, err := r.findOrgMemberByUser(orgID, prevOwnerUserID)
		if err != nil {
			return err
		}
		prevOwnerMembership.Set("role", string(domain.OrgRoleAdmin))
		if err := r.app.Save(prevOwnerMembership); err != nil {
			return fmt.Errorf("save previous owner membership role: %w", err)
		}
	}

	newOwnerMembership.Set("role", string(domain.OrgRoleOwner))
	if err := r.app.Save(newOwnerMembership); err != nil {
		return fmt.Errorf("save new owner membership role: %w", err)
	}

	org.Set("owner", newOwnerUserID)
	if err := r.app.Save(org); err != nil {
		return fmt.Errorf("save organization owner: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) DeleteOrganization(_ context.Context, orgID string) error {
	org, err := r.findOrganizationByID(orgID)
	if err != nil {
		return err
	}
	if err := r.app.Delete(org); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) FindInvitationByToken(_ context.Context, token string) (domain.Invitation, error) {
	record, err := r.app.FindFirstRecordByFilter(
		"invitations",
		"token = {:token}",
		dbx.Params{"token": token},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Invitation{}, domain.ErrNotFound
		}
		return domain.Invitation{}, fmt.Errorf("find invitation by token: %w", err)
	}
	return domain.Invitation{
		ID:             record.Id,
		OrganizationID: record.GetString("organization"),
		Email:          record.GetString("email"),
		Role:           domain.OrgRole(record.GetString("role")),
		Token:          record.GetString("token"),
	}, nil
}

func (r *PocketBaseRepository) AcceptInvitation(_ context.Context, invitation domain.Invitation, userID string) error {
	member, err := r.findOrgMemberByUser(invitation.OrganizationID, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if member == nil {
		if _, err := r.CreateMembership(context.Background(), domain.OrganizationMember{
			OrganizationID: invitation.OrganizationID,
			UserID:         userID,
			Role:           invitation.Role,
		}); err != nil {
			return err
		}
	}

	record, err := r.findInvitationByID(invitation.OrganizationID, invitation.ID)
	if err != nil {
		return err
	}
	if err := r.app.Delete(record); err != nil {
		return fmt.Errorf("delete invitation on accept: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) DeclineInvitation(_ context.Context, invitation domain.Invitation) error {
	record, err := r.findInvitationByID(invitation.OrganizationID, invitation.ID)
	if err != nil {
		return err
	}
	if err := r.app.Delete(record); err != nil {
		return fmt.Errorf("delete invitation on decline: %w", err)
	}
	return nil
}

func (r *PocketBaseRepository) findOrgMemberByUser(orgID string, userID string) (*core.Record, error) {
	record, err := r.app.FindFirstRecordByFilter(
		"organization_members",
		"organization = {:orgID} && user = {:userID}",
		dbx.Params{"orgID": orgID, "userID": userID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find organization member: %w", err)
	}
	return record, nil
}

func (r *PocketBaseRepository) findInvitationByID(orgID string, invitationID string) (*core.Record, error) {
	record, err := r.app.FindFirstRecordByFilter(
		"invitations",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": invitationID, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find invitation: %w", err)
	}
	return record, nil
}

func (r *PocketBaseRepository) findOrganizationByID(orgID string) (*core.Record, error) {
	record, err := r.app.FindFirstRecordByFilter(
		"organizations",
		"id = {:orgID}",
		dbx.Params{"orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find organization: %w", err)
	}
	return record, nil
}

func generateInvitationToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
