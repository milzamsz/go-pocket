package tenancy

import (
	"context"
	"errors"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestPocketBaseRepository_ListMembersByOrg_IsolatedByOrganization(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	userA := createTestUser(t, app, "owner-a@example.com")
	userB := createTestUser(t, app, "owner-b@example.com")

	orgAID := "orga12345678901"
	orgBID := "orgb12345678901"

	_, err := repo.CreateOrganization(ctx, domain.Organization{
		ID:      orgAID,
		Slug:    "org-a",
		Name:    "Organization A",
		OwnerID: userA.Id,
	})
	require.NoError(t, err)
	_, err = repo.CreateOrganization(ctx, domain.Organization{
		ID:      orgBID,
		Slug:    "org-b",
		Name:    "Organization B",
		OwnerID: userB.Id,
	})
	require.NoError(t, err)

	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: orgAID, UserID: userA.Id, Role: domain.OrgRoleOwner})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: orgBID, UserID: userB.Id, Role: domain.OrgRoleOwner})
	require.NoError(t, err)

	members, err := repo.ListMembersByOrg(ctx, orgAID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, orgAID, members[0].OrganizationID)
	require.Equal(t, userA.Id, members[0].UserID)
}

func TestPocketBaseRepository_InviteMember_CreatesInvitation(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	owner := createTestUser(t, app, "owner@example.com")
	orgID := createTestOrganization(t, repo, ctx, owner.Id, "acme")

	invite, err := repo.InviteMember(ctx, orgID, "new@example.com", domain.OrgRoleMember)
	require.NoError(t, err)
	require.Equal(t, orgID, invite.OrganizationID)
	require.Equal(t, "new@example.com", invite.Email)
	require.Equal(t, domain.OrgRoleMember, invite.Role)
	require.NotEmpty(t, invite.ID)
	require.NotEmpty(t, invite.Token)
}

func TestPocketBaseRepository_RemoveMember_EnforcesOrganizationIsolation(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	ownerA := createTestUser(t, app, "owner-a@example.com")
	ownerB := createTestUser(t, app, "owner-b@example.com")
	orgAID := createTestOrganization(t, repo, ctx, ownerA.Id, "org-a")
	orgBID := createTestOrganization(t, repo, ctx, ownerB.Id, "org-b")

	member := createTestUser(t, app, "member@example.com")
	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: orgAID, UserID: member.Id, Role: domain.OrgRoleMember})
	require.NoError(t, err)

	err = repo.RemoveMember(ctx, orgBID, member.Id)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.RemoveMember(ctx, orgAID, member.Id)
	require.NoError(t, err)

	members, err := repo.ListMembersByOrg(ctx, orgAID)
	require.NoError(t, err)
	require.Len(t, members, 1) // owner membership remains
	require.Equal(t, ownerA.Id, members[0].UserID)
}

func TestPocketBaseRepository_ChangeMemberRole_EnforcesOrganizationIsolation(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	ownerA := createTestUser(t, app, "owner-a@example.com")
	ownerB := createTestUser(t, app, "owner-b@example.com")
	orgAID := createTestOrganization(t, repo, ctx, ownerA.Id, "org-a")
	orgBID := createTestOrganization(t, repo, ctx, ownerB.Id, "org-b")

	member := createTestUser(t, app, "member@example.com")
	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: orgAID, UserID: member.Id, Role: domain.OrgRoleMember})
	require.NoError(t, err)

	err = repo.ChangeMemberRole(ctx, orgBID, member.Id, domain.OrgRoleAdmin)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.ChangeMemberRole(ctx, orgAID, member.Id, domain.OrgRoleAdmin)
	require.NoError(t, err)

	memberRecord := findOrgMemberRecord(t, app, orgAID, member.Id)
	require.Equal(t, string(domain.OrgRoleAdmin), memberRecord.GetString("role"))
}

func TestPocketBaseRepository_ResendAndRevokeInvitation_EnforcesOrganizationIsolation(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	ownerA := createTestUser(t, app, "owner-a@example.com")
	ownerB := createTestUser(t, app, "owner-b@example.com")
	orgAID := createTestOrganization(t, repo, ctx, ownerA.Id, "org-a")
	orgBID := createTestOrganization(t, repo, ctx, ownerB.Id, "org-b")

	invite, err := repo.InviteMember(ctx, orgAID, "new@example.com", domain.OrgRoleMember)
	require.NoError(t, err)
	originalToken := invite.Token

	err = repo.ResendInvitation(ctx, orgBID, invite.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.ResendInvitation(ctx, orgAID, invite.ID)
	require.NoError(t, err)
	updatedInvite := findInvitationRecord(t, app, orgAID, invite.ID)
	require.NotEqual(t, originalToken, updatedInvite.GetString("token"))

	err = repo.RevokeInvitation(ctx, orgBID, invite.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	err = repo.RevokeInvitation(ctx, orgAID, invite.ID)
	require.NoError(t, err)

	_, err = app.FindFirstRecordByFilter("invitations", "id = {:id}", dbx.Params{"id": invite.ID})
	require.Error(t, err)
}

func TestPocketBaseRepository_UpdateSettings_TransferOwnership_AndDeleteOrganization(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	owner := createTestUser(t, app, "owner@example.com")
	newOwner := createTestUser(t, app, "new-owner@example.com")
	orgID := createTestOrganization(t, repo, ctx, owner.Id, "acme")

	_, err := repo.CreateMembership(ctx, domain.OrganizationMember{OrganizationID: orgID, UserID: newOwner.Id, Role: domain.OrgRoleAdmin})
	require.NoError(t, err)

	err = repo.UpdateOrganizationSettings(ctx, orgID, "Acme Renamed")
	require.NoError(t, err)
	orgRecord := findOrganizationRecord(t, app, orgID)
	require.Equal(t, "Acme Renamed", orgRecord.GetString("name"))

	err = repo.TransferOrganizationOwnership(ctx, orgID, newOwner.Id)
	require.NoError(t, err)

	orgRecord = findOrganizationRecord(t, app, orgID)
	require.Equal(t, newOwner.Id, orgRecord.GetString("owner"))
	require.Equal(t, string(domain.OrgRoleOwner), findOrgMemberRecord(t, app, orgID, newOwner.Id).GetString("role"))
	require.Equal(t, string(domain.OrgRoleAdmin), findOrgMemberRecord(t, app, orgID, owner.Id).GetString("role"))

	err = repo.DeleteOrganization(ctx, orgID)
	require.NoError(t, err)

	_, err = app.FindFirstRecordByFilter("organizations", "id = {:id}", dbx.Params{"id": orgID})
	require.Error(t, err)
}

func TestPocketBaseRepository_TransferOwnership_ReturnsNotFoundWhenNewOwnerOutsideOrg(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	owner := createTestUser(t, app, "owner@example.com")
	outsider := createTestUser(t, app, "outsider@example.com")
	orgID := createTestOrganization(t, repo, ctx, owner.Id, "acme")

	err := repo.TransferOrganizationOwnership(ctx, orgID, outsider.Id)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestPocketBaseRepository_FindAcceptDeclineInvitationByToken(t *testing.T) {
	app := testutil.NewTestApp(t)
	repo := NewPocketBaseRepository(app)
	ctx := context.Background()

	owner := createTestUser(t, app, "owner@example.com")
	member := createTestUser(t, app, "member@example.com")
	orgID := createTestOrganization(t, repo, ctx, owner.Id, "acme")

	invite, err := repo.InviteMember(ctx, orgID, "member@example.com", domain.OrgRoleMember)
	require.NoError(t, err)

	found, err := repo.FindInvitationByToken(ctx, invite.Token)
	require.NoError(t, err)
	require.Equal(t, invite.ID, found.ID)

	err = repo.AcceptInvitation(ctx, found, member.Id)
	require.NoError(t, err)
	acceptedMembership := findOrgMemberRecord(t, app, orgID, member.Id)
	require.Equal(t, string(domain.OrgRoleMember), acceptedMembership.GetString("role"))

	_, err = app.FindFirstRecordByFilter("invitations", "id = {:id}", dbx.Params{"id": invite.ID})
	require.Error(t, err)

	invite2, err := repo.InviteMember(ctx, orgID, "decline@example.com", domain.OrgRoleViewer)
	require.NoError(t, err)
	found2, err := repo.FindInvitationByToken(ctx, invite2.Token)
	require.NoError(t, err)
	err = repo.DeclineInvitation(ctx, found2)
	require.NoError(t, err)
	_, err = app.FindFirstRecordByFilter("invitations", "id = {:id}", dbx.Params{"id": invite2.ID})
	require.Error(t, err)
}

func createTestOrganization(t *testing.T, repo *PocketBaseRepository, ctx context.Context, ownerUserID string, slug string) string {
	t.Helper()
	filtered := make([]rune, 0, len(slug))
	for _, ch := range slug {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			filtered = append(filtered, ch)
		}
	}
	orgID := "org" + string(filtered)
	if len(orgID) < 15 {
		orgID = orgID + "123456789012345"[:15-len(orgID)]
	} else if len(orgID) > 15 {
		orgID = orgID[:15]
	}
	_, err := repo.CreateOrganization(ctx, domain.Organization{
		ID:      orgID,
		Slug:    slug,
		Name:    slug,
		OwnerID: ownerUserID,
	})
	require.NoError(t, err)
	_, err = repo.CreateMembership(ctx, domain.OrganizationMember{
		OrganizationID: orgID,
		UserID:         ownerUserID,
		Role:           domain.OrgRoleOwner,
	})
	require.NoError(t, err)
	return orgID
}

func findOrgMemberRecord(t *testing.T, app core.App, orgID string, userID string) *core.Record {
	t.Helper()
	record, err := app.FindFirstRecordByFilter(
		"organization_members",
		"organization = {:orgID} && user = {:userID}",
		dbx.Params{"orgID": orgID, "userID": userID},
	)
	require.NoError(t, err)
	return record
}

func findInvitationRecord(t *testing.T, app core.App, orgID string, invitationID string) *core.Record {
	t.Helper()
	record, err := app.FindFirstRecordByFilter(
		"invitations",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": invitationID, "orgID": orgID},
	)
	require.NoError(t, err)
	return record
}

func findOrganizationRecord(t *testing.T, app core.App, orgID string) *core.Record {
	t.Helper()
	record, err := app.FindFirstRecordByFilter("organizations", "id = {:id}", dbx.Params{"id": orgID})
	require.NoError(t, err)
	return record
}

func createTestUser(t *testing.T, app core.App, email string) *core.Record {
	t.Helper()

	users, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	record := core.NewRecord(users)
	record.Set("email", email)
	record.Set("password", "test-password-123")
	record.Set("passwordConfirm", "test-password-123")

	require.NoError(t, app.Save(record))
	return record
}
