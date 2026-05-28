package tenancy

import (
	"context"
	"fmt"
	"sync"

	"github.com/milzamsz/go-pocket/internal/domain"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	orgs    map[string]domain.Organization
	members []domain.OrganizationMember
	invites map[string]domain.Invitation
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		orgs:    make(map[string]domain.Organization),
		invites: make(map[string]domain.Invitation),
	}
}

func (r *MemoryRepository) CreateOrganization(_ context.Context, org domain.Organization) (domain.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if org.ID == "" {
		return domain.Organization{}, fmt.Errorf("organization id is required")
	}
	r.orgs[org.ID] = org
	return org, nil
}

func (r *MemoryRepository) CreateMembership(_ context.Context, membership domain.OrganizationMember) (domain.OrganizationMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members = append(r.members, membership)
	return membership, nil
}

func (r *MemoryRepository) ListMembersByOrg(_ context.Context, orgID string) ([]domain.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.OrganizationMember, 0)
	for _, member := range r.members {
		if member.OrganizationID == orgID {
			out = append(out, member)
		}
	}
	return out, nil
}

func (r *MemoryRepository) FindMemberProfile(_ context.Context, orgID string, userID string) (domain.OrganizationMemberProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, member := range r.members {
		if member.OrganizationID == orgID && member.UserID == userID {
			return domain.OrganizationMemberProfile{
				OrganizationID: orgID,
				UserID:         userID,
				Role:           member.Role,
				Name:           userID,
				Email:          userID + "@example.test",
			}, nil
		}
	}
	return domain.OrganizationMemberProfile{}, domain.ErrNotFound
}

func (r *MemoryRepository) InviteMember(_ context.Context, orgID string, email string, role domain.OrgRole) (domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, err := generateInvitationToken()
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("generate invitation token: %w", err)
	}

	inv := domain.Invitation{
		ID:             fmt.Sprintf("inv-%d", len(r.invites)+1),
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		Token:          token,
	}
	r.invites[inv.ID] = inv
	return inv, nil
}

func (r *MemoryRepository) RemoveMember(_ context.Context, orgID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	memberIndex := r.findMemberIndexByOrgAndUser(orgID, userID)
	if memberIndex == -1 {
		return domain.ErrNotFound
	}

	r.members = append(r.members[:memberIndex], r.members[memberIndex+1:]...)
	return nil
}

func (r *MemoryRepository) ChangeMemberRole(_ context.Context, orgID string, userID string, role domain.OrgRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	memberIndex := r.findMemberIndexByOrgAndUser(orgID, userID)
	if memberIndex == -1 {
		return domain.ErrNotFound
	}

	r.members[memberIndex].Role = role
	return nil
}

func (r *MemoryRepository) ResendInvitation(_ context.Context, orgID string, invitationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, ok := r.invites[invitationID]
	if !ok || invite.OrganizationID != orgID {
		return domain.ErrNotFound
	}

	token, err := generateInvitationToken()
	if err != nil {
		return fmt.Errorf("generate invitation token: %w", err)
	}

	invite.Token = token
	r.invites[invitationID] = invite
	return nil
}

func (r *MemoryRepository) RevokeInvitation(_ context.Context, orgID string, invitationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, ok := r.invites[invitationID]
	if !ok || invite.OrganizationID != orgID {
		return domain.ErrNotFound
	}

	delete(r.invites, invitationID)
	return nil
}

func (r *MemoryRepository) UpdateOrganizationSettings(_ context.Context, orgID string, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, ok := r.orgs[orgID]
	if !ok {
		return domain.ErrNotFound
	}
	org.Name = name
	r.orgs[orgID] = org
	return nil
}

func (r *MemoryRepository) TransferOrganizationOwnership(_ context.Context, orgID string, newOwnerUserID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, ok := r.orgs[orgID]
	if !ok {
		return domain.ErrNotFound
	}

	newOwnerIndex := r.findMemberIndexByOrgAndUser(orgID, newOwnerUserID)
	if newOwnerIndex == -1 {
		return domain.ErrNotFound
	}

	prevOwnerUserID := org.OwnerID
	if prevOwnerUserID != "" && prevOwnerUserID != newOwnerUserID {
		prevOwnerIndex := r.findMemberIndexByOrgAndUser(orgID, prevOwnerUserID)
		if prevOwnerIndex == -1 {
			return domain.ErrNotFound
		}
		r.members[prevOwnerIndex].Role = domain.OrgRoleAdmin
	}

	r.members[newOwnerIndex].Role = domain.OrgRoleOwner
	org.OwnerID = newOwnerUserID
	r.orgs[orgID] = org
	return nil
}

func (r *MemoryRepository) DeleteOrganization(_ context.Context, orgID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orgs[orgID]; !ok {
		return domain.ErrNotFound
	}
	delete(r.orgs, orgID)

	filteredMembers := make([]domain.OrganizationMember, 0, len(r.members))
	for _, member := range r.members {
		if member.OrganizationID != orgID {
			filteredMembers = append(filteredMembers, member)
		}
	}
	r.members = filteredMembers

	for inviteID, invite := range r.invites {
		if invite.OrganizationID == orgID {
			delete(r.invites, inviteID)
		}
	}

	return nil
}

func (r *MemoryRepository) FindInvitationByToken(_ context.Context, token string) (domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, invite := range r.invites {
		if invite.Token == token {
			return invite, nil
		}
	}
	return domain.Invitation{}, domain.ErrNotFound
}

func (r *MemoryRepository) AcceptInvitation(_ context.Context, invitation domain.Invitation, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.invites[invitation.ID]; !ok {
		return domain.ErrNotFound
	}
	r.members = append(r.members, domain.OrganizationMember{
		OrganizationID: invitation.OrganizationID,
		UserID:         userID,
		Role:           invitation.Role,
	})
	delete(r.invites, invitation.ID)
	return nil
}

func (r *MemoryRepository) DeclineInvitation(_ context.Context, invitation domain.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.invites[invitation.ID]; !ok {
		return domain.ErrNotFound
	}
	delete(r.invites, invitation.ID)
	return nil
}

func (r *MemoryRepository) findMemberIndexByOrgAndUser(orgID string, userID string) int {
	for i := range r.members {
		member := r.members[i]
		if member.OrganizationID == orgID && member.UserID == userID {
			return i
		}
	}
	return -1
}
