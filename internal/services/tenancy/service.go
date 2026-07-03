package tenancy

import (
	"context"
	"errors"
	"fmt"

	"github.com/milzamsz/go-pocket/internal/domain"
)

type Repository interface {
	CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)
	CreateMembership(ctx context.Context, membership domain.OrganizationMember) (domain.OrganizationMember, error)
	ListMembersByOrg(ctx context.Context, orgID string) ([]domain.OrganizationMember, error)
	FindMemberProfile(ctx context.Context, orgID string, userID string) (domain.OrganizationMemberProfile, error)
	GetOrganizationShell(ctx context.Context, orgID string) (OrganizationShell, error)
	InviteMember(ctx context.Context, orgID string, email string, role domain.OrgRole) (domain.Invitation, error)
	RemoveMember(ctx context.Context, orgID string, userID string) error
	ChangeMemberRole(ctx context.Context, orgID string, userID string, role domain.OrgRole) error
	ResendInvitation(ctx context.Context, orgID string, invitationID string) error
	RevokeInvitation(ctx context.Context, orgID string, invitationID string) error
	UpdateOrganizationSettings(ctx context.Context, orgID string, name string) error
	TransferOrganizationOwnership(ctx context.Context, orgID string, newOwnerUserID string) error
	DeleteOrganization(ctx context.Context, orgID string) error
	FindInvitationByToken(ctx context.Context, token string) (domain.Invitation, error)
	AcceptInvitation(ctx context.Context, invitation domain.Invitation, userID string) error
	DeclineInvitation(ctx context.Context, invitation domain.Invitation) error
}

type Service interface {
	CreateOrganization(ctx context.Context, org domain.Organization, ownerUserID string) (domain.Organization, error)
	ListMembers(ctx context.Context, orgID string, actorRole domain.OrgRole) ([]domain.OrganizationMember, error)
	GetMemberProfile(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string) (domain.OrganizationMemberProfile, error)
	GetOrganizationShell(ctx context.Context, orgID string, actorRole domain.OrgRole) (OrganizationShell, error)
	InviteMember(ctx context.Context, orgID string, actorRole domain.OrgRole, email string, role domain.OrgRole) (domain.Invitation, error)
	RemoveMember(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string) error
	ChangeMemberRole(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string, role domain.OrgRole) error
	ResendInvitation(ctx context.Context, orgID string, actorRole domain.OrgRole, invitationID string) error
	RevokeInvitation(ctx context.Context, orgID string, actorRole domain.OrgRole, invitationID string) error
	UpdateSettings(ctx context.Context, orgID string, actorRole domain.OrgRole, name string) error
	TransferOwnership(ctx context.Context, orgID string, actorRole domain.OrgRole, newOwnerUserID string) error
	DeleteOrganization(ctx context.Context, orgID string, actorRole domain.OrgRole) error
	AcceptInvitation(ctx context.Context, token string, userID string) (domain.Invitation, error)
	DeclineInvitation(ctx context.Context, token string) (domain.Invitation, error)
}

type service struct {
	repo Repository
}

type OrganizationShell struct {
	Name               string
	Plan               string
	SubscriptionStatus string
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateOrganization(ctx context.Context, org domain.Organization, ownerUserID string) (domain.Organization, error) {
	if ownerUserID == "" {
		return domain.Organization{}, fmt.Errorf("owner user id is required")
	}
	createdOrg, err := s.repo.CreateOrganization(ctx, org)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("create organization: %w", err)
	}
	_, err = s.repo.CreateMembership(ctx, domain.OrganizationMember{
		OrganizationID: createdOrg.ID,
		UserID:         ownerUserID,
		Role:           domain.OrgRoleOwner,
	})
	if err != nil {
		return domain.Organization{}, fmt.Errorf("create owner membership: %w", err)
	}
	return createdOrg, nil
}

func (s *service) ListMembers(ctx context.Context, orgID string, actorRole domain.OrgRole) ([]domain.OrganizationMember, error) {
	if orgID == "" {
		return nil, errors.New("orgID is required")
	}
	if !Can(actorRole, PermissionMembersRead) {
		return nil, domain.ErrForbidden
	}
	members, err := s.repo.ListMembersByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members by org: %w", err)
	}
	return members, nil
}

func (s *service) GetMemberProfile(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string) (domain.OrganizationMemberProfile, error) {
	if orgID == "" || userID == "" {
		return domain.OrganizationMemberProfile{}, errors.New("orgID and userID are required")
	}
	if !Can(actorRole, PermissionMembersRead) {
		return domain.OrganizationMemberProfile{}, domain.ErrForbidden
	}
	profile, err := s.repo.FindMemberProfile(ctx, orgID, userID)
	if err != nil {
		return domain.OrganizationMemberProfile{}, fmt.Errorf("find member profile: %w", err)
	}
	return profile, nil
}

func (s *service) GetOrganizationShell(ctx context.Context, orgID string, actorRole domain.OrgRole) (OrganizationShell, error) {
	if orgID == "" {
		return OrganizationShell{}, errors.New("orgID is required")
	}
	if !Can(actorRole, PermissionMembersRead) {
		return OrganizationShell{}, domain.ErrForbidden
	}
	shell, err := s.repo.GetOrganizationShell(ctx, orgID)
	if err != nil {
		return OrganizationShell{}, fmt.Errorf("get organization shell: %w", err)
	}
	return shell, nil
}

func (s *service) InviteMember(ctx context.Context, orgID string, actorRole domain.OrgRole, email string, role domain.OrgRole) (domain.Invitation, error) {
	if !Can(actorRole, PermissionInvitesCreate) {
		return domain.Invitation{}, domain.ErrForbidden
	}
	if orgID == "" || email == "" {
		return domain.Invitation{}, errors.New("orgID and email are required")
	}
	if role == "" {
		role = domain.OrgRoleMember
	}
	invite, err := s.repo.InviteMember(ctx, orgID, email, role)
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("invite member: %w", err)
	}
	return invite, nil
}

func (s *service) RemoveMember(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string) error {
	if !Can(actorRole, PermissionMembersWrite) {
		return domain.ErrForbidden
	}
	if orgID == "" || userID == "" {
		return errors.New("orgID and userID are required")
	}
	if err := s.repo.RemoveMember(ctx, orgID, userID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (s *service) ChangeMemberRole(ctx context.Context, orgID string, actorRole domain.OrgRole, userID string, role domain.OrgRole) error {
	if !Can(actorRole, PermissionMembersWrite) {
		return domain.ErrForbidden
	}
	if orgID == "" || userID == "" || role == "" {
		return errors.New("orgID, userID, and role are required")
	}
	if err := s.repo.ChangeMemberRole(ctx, orgID, userID, role); err != nil {
		return fmt.Errorf("change member role: %w", err)
	}
	return nil
}

func (s *service) ResendInvitation(ctx context.Context, orgID string, actorRole domain.OrgRole, invitationID string) error {
	if !Can(actorRole, PermissionInvitesCreate) {
		return domain.ErrForbidden
	}
	if orgID == "" || invitationID == "" {
		return errors.New("orgID and invitationID are required")
	}
	if err := s.repo.ResendInvitation(ctx, orgID, invitationID); err != nil {
		return fmt.Errorf("resend invitation: %w", err)
	}
	return nil
}

func (s *service) RevokeInvitation(ctx context.Context, orgID string, actorRole domain.OrgRole, invitationID string) error {
	if !Can(actorRole, PermissionInvitesCreate) {
		return domain.ErrForbidden
	}
	if orgID == "" || invitationID == "" {
		return errors.New("orgID and invitationID are required")
	}
	if err := s.repo.RevokeInvitation(ctx, orgID, invitationID); err != nil {
		return fmt.Errorf("revoke invitation: %w", err)
	}
	return nil
}

func (s *service) UpdateSettings(ctx context.Context, orgID string, actorRole domain.OrgRole, name string) error {
	if !Can(actorRole, PermissionMembersWrite) {
		return domain.ErrForbidden
	}
	if orgID == "" {
		return errors.New("orgID is required")
	}
	if err := s.repo.UpdateOrganizationSettings(ctx, orgID, name); err != nil {
		return fmt.Errorf("update organization settings: %w", err)
	}
	return nil
}

func (s *service) TransferOwnership(ctx context.Context, orgID string, actorRole domain.OrgRole, newOwnerUserID string) error {
	if actorRole != domain.OrgRoleOwner {
		return domain.ErrForbidden
	}
	if orgID == "" || newOwnerUserID == "" {
		return errors.New("orgID and newOwnerUserID are required")
	}
	if err := s.repo.TransferOrganizationOwnership(ctx, orgID, newOwnerUserID); err != nil {
		return fmt.Errorf("transfer organization ownership: %w", err)
	}
	return nil
}

func (s *service) DeleteOrganization(ctx context.Context, orgID string, actorRole domain.OrgRole) error {
	if actorRole != domain.OrgRoleOwner {
		return domain.ErrForbidden
	}
	if orgID == "" {
		return errors.New("orgID is required")
	}
	if err := s.repo.DeleteOrganization(ctx, orgID); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

func (s *service) AcceptInvitation(ctx context.Context, token string, userID string) (domain.Invitation, error) {
	if token == "" || userID == "" {
		return domain.Invitation{}, errors.New("token and userID are required")
	}
	invitation, err := s.repo.FindInvitationByToken(ctx, token)
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("find invitation by token: %w", err)
	}
	if err := s.repo.AcceptInvitation(ctx, invitation, userID); err != nil {
		return domain.Invitation{}, fmt.Errorf("accept invitation: %w", err)
	}
	return invitation, nil
}

func (s *service) DeclineInvitation(ctx context.Context, token string) (domain.Invitation, error) {
	if token == "" {
		return domain.Invitation{}, errors.New("token is required")
	}
	invitation, err := s.repo.FindInvitationByToken(ctx, token)
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("find invitation by token: %w", err)
	}
	if err := s.repo.DeclineInvitation(ctx, invitation); err != nil {
		return domain.Invitation{}, fmt.Errorf("decline invitation: %w", err)
	}
	return invitation, nil
}
