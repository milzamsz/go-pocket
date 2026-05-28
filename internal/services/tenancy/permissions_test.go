package tenancy

import (
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestPermissionMatrix(t *testing.T) {
	tests := []struct {
		role       domain.OrgRole
		permission Permission
		allowed    bool
	}{
		{role: domain.OrgRoleOwner, permission: PermissionMembersRead, allowed: true},
		{role: domain.OrgRoleOwner, permission: PermissionMembersWrite, allowed: true},
		{role: domain.OrgRoleOwner, permission: PermissionBillingRead, allowed: true},
		{role: domain.OrgRoleOwner, permission: PermissionBillingWrite, allowed: true},
		{role: domain.OrgRoleOwner, permission: PermissionInvitesCreate, allowed: true},
		{role: domain.OrgRoleAdmin, permission: PermissionMembersWrite, allowed: true},
		{role: domain.OrgRoleAdmin, permission: PermissionBillingWrite, allowed: true},
		{role: domain.OrgRoleAdmin, permission: PermissionInvitesCreate, allowed: true},
		{role: domain.OrgRoleMember, permission: PermissionMembersRead, allowed: true},
		{role: domain.OrgRoleMember, permission: PermissionMembersWrite, allowed: false},
		{role: domain.OrgRoleMember, permission: PermissionBillingRead, allowed: true},
		{role: domain.OrgRoleMember, permission: PermissionBillingWrite, allowed: false},
		{role: domain.OrgRoleViewer, permission: PermissionMembersRead, allowed: true},
		{role: domain.OrgRoleViewer, permission: PermissionBillingRead, allowed: true},
		{role: domain.OrgRoleViewer, permission: PermissionInvitesCreate, allowed: false},
	}

	for _, tc := range tests {
		require.Equal(t, tc.allowed, Can(tc.role, tc.permission), "%s:%s", tc.role, tc.permission)
	}
}
