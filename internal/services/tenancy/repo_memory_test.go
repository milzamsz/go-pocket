package tenancy

import (
	"context"
	"errors"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_RemoveMember_EnforcesOrganizationIsolation(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "user-1", Role: domain.OrgRoleMember})
	require.NoError(t, err)

	err = repo.RemoveMember(ctx, "org-b", "user-1")
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.RemoveMember(ctx, "org-a", "user-1")
	require.NoError(t, err)

	members, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	require.Empty(t, members)
}

func TestMemoryRepository_ChangeMemberRole_EnforcesOrganizationIsolation(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "user-1", Role: domain.OrgRoleMember})
	require.NoError(t, err)

	err = repo.ChangeMemberRole(ctx, "org-b", "user-1", domain.OrgRoleAdmin)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.ChangeMemberRole(ctx, "org-a", "user-1", domain.OrgRoleAdmin)
	require.NoError(t, err)

	members, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, domain.OrgRoleAdmin, members[0].Role)
}

func TestMemoryRepository_ResendAndRevokeInvitation_EnforcesOrganizationIsolation(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	invite, err := repo.InviteMember(ctx, "org-a", "new@example.com", domain.OrgRoleMember)
	require.NoError(t, err)
	require.NotEmpty(t, invite.Token)

	err = repo.ResendInvitation(ctx, "org-b", invite.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.ResendInvitation(ctx, "org-a", invite.ID)
	require.NoError(t, err)
	updated := repo.invites[invite.ID]
	require.NotEqual(t, invite.Token, updated.Token)

	err = repo.RevokeInvitation(ctx, "org-b", invite.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.RevokeInvitation(ctx, "org-a", invite.ID)
	require.NoError(t, err)
	_, ok := repo.invites[invite.ID]
	require.False(t, ok)
}

func TestMemoryRepository_UpdateSettings_TransferOwnership_AndDeleteOrganization(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.CreateOrganization(ctx, domain.Organization{
		ID:      "org-a",
		Slug:    "org-a",
		Name:    "Org A",
		OwnerID: "owner-1",
	})
	require.NoError(t, err)

	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "owner-1", Role: domain.OrgRoleOwner})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "owner-2", Role: domain.OrgRoleAdmin})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-b", UserID: "other", Role: domain.OrgRoleMember})
	require.NoError(t, err)

	err = repo.UpdateOrganizationSettings(ctx, "org-a", "Org A Renamed")
	require.NoError(t, err)
	require.Equal(t, "Org A Renamed", repo.orgs["org-a"].Name)

	err = repo.TransferOrganizationOwnership(ctx, "org-a", "owner-2")
	require.NoError(t, err)
	require.Equal(t, "owner-2", repo.orgs["org-a"].OwnerID)

	members, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	var roleByUser = map[string]domain.OrgRole{}
	for _, m := range members {
		roleByUser[m.UserID] = m.Role
	}
	require.Equal(t, domain.OrgRoleOwner, roleByUser["owner-2"])
	require.Equal(t, domain.OrgRoleAdmin, roleByUser["owner-1"])

	invite, err := repo.InviteMember(ctx, "org-a", "pending@example.com", domain.OrgRoleMember)
	require.NoError(t, err)

	err = repo.DeleteOrganization(ctx, "org-a")
	require.NoError(t, err)
	_, ok := repo.orgs["org-a"]
	require.False(t, ok)

	orgAMembers, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	require.Empty(t, orgAMembers)

	orgBMembers, err := repo.ListMembersByOrg(ctx, "org-b")
	require.NoError(t, err)
	require.Len(t, orgBMembers, 1)

	_, inviteExists := repo.invites[invite.ID]
	require.False(t, inviteExists)
}

func TestMemoryRepository_TransferOwnership_ReturnsNotFoundWhenNewOwnerOutsideOrg(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.CreateOrganization(ctx, domain.Organization{
		ID:      "org-a",
		Slug:    "org-a",
		Name:    "Org A",
		OwnerID: "owner-1",
	})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "owner-1", Role: domain.OrgRoleOwner})
	require.NoError(t, err)

	err = repo.TransferOrganizationOwnership(ctx, "org-a", "outsider")
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestMemoryRepository_FindAcceptDeclineInvitationByToken(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	invite, err := repo.InviteMember(ctx, "org-a", "new@example.com", domain.OrgRoleMember)
	require.NoError(t, err)

	found, err := repo.FindInvitationByToken(ctx, invite.Token)
	require.NoError(t, err)
	require.Equal(t, invite.ID, found.ID)

	err = repo.AcceptInvitation(ctx, found, "user-1")
	require.NoError(t, err)
	members, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	require.Len(t, members, 1)

	invite2, err := repo.InviteMember(ctx, "org-a", "decline@example.com", domain.OrgRoleViewer)
	require.NoError(t, err)
	found2, err := repo.FindInvitationByToken(ctx, invite2.Token)
	require.NoError(t, err)
	err = repo.DeclineInvitation(ctx, found2)
	require.NoError(t, err)
}
