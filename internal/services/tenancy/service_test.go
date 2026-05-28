package tenancy

import (
	"context"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestListMembers_IsolatedByOrganization(t *testing.T) {
	repo := NewMemoryRepository()
	svc := New(repo)
	ctx := context.Background()

	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-a", UserID: "u1", Role: domain.OrgRoleOwner})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: "org-b", UserID: "u2", Role: domain.OrgRoleOwner})
	require.NoError(t, err)

	got, err := svc.ListMembers(ctx, "org-a", domain.OrgRoleOwner)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "org-a", got[0].OrganizationID)
}

func TestAcceptAndDeclineInvitation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := New(repo)
	ctx := context.Background()

	invite, err := repo.InviteMember(ctx, "org-a", "user@example.com", domain.OrgRoleMember)
	require.NoError(t, err)
	require.NotEmpty(t, invite.Token)

	accepted, err := svc.AcceptInvitation(ctx, invite.Token, "user-1")
	require.NoError(t, err)
	require.Equal(t, invite.ID, accepted.ID)

	members, err := repo.ListMembersByOrg(ctx, "org-a")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, "user-1", members[0].UserID)

	invite2, err := repo.InviteMember(ctx, "org-a", "user2@example.com", domain.OrgRoleViewer)
	require.NoError(t, err)
	declined, err := svc.DeclineInvitation(ctx, invite2.Token)
	require.NoError(t, err)
	require.Equal(t, invite2.ID, declined.ID)
}
