package tenancy

import "github.com/milzamsz/go-pocket/internal/domain"

type Permission string

const (
	PermissionMembersRead   Permission = "members:read"
	PermissionMembersWrite  Permission = "members:write"
	PermissionBillingRead   Permission = "billing:read"
	PermissionBillingWrite  Permission = "billing:write"
	PermissionInvitesCreate Permission = "invites:create"
	PermissionProductsRead  Permission = "products:read"
	PermissionProductsWrite Permission = "products:write"
	PermissionKanbanRead    Permission = "kanban:read"
	PermissionKanbanWrite   Permission = "kanban:write"
)

var permissionMatrix = map[domain.OrgRole]map[Permission]bool{
	domain.OrgRoleOwner: {
		PermissionMembersRead:   true,
		PermissionMembersWrite:  true,
		PermissionBillingRead:   true,
		PermissionBillingWrite:  true,
		PermissionInvitesCreate: true,
		PermissionProductsRead:  true,
		PermissionProductsWrite: true,
		PermissionKanbanRead:    true,
		PermissionKanbanWrite:   true,
	},
	domain.OrgRoleAdmin: {
		PermissionMembersRead:   true,
		PermissionMembersWrite:  true,
		PermissionBillingRead:   true,
		PermissionBillingWrite:  true,
		PermissionInvitesCreate: true,
		PermissionProductsRead:  true,
		PermissionProductsWrite: true,
		PermissionKanbanRead:    true,
		PermissionKanbanWrite:   true,
	},
	domain.OrgRoleMember: {
		PermissionMembersRead:   true,
		PermissionBillingRead:   true,
		PermissionMembersWrite:  false,
		PermissionBillingWrite:  false,
		PermissionInvitesCreate: false,
		PermissionProductsRead:  true,
		PermissionProductsWrite: true,
		PermissionKanbanRead:    true,
		PermissionKanbanWrite:   true,
	},
	domain.OrgRoleViewer: {
		PermissionMembersRead:   true,
		PermissionBillingRead:   true,
		PermissionMembersWrite:  false,
		PermissionBillingWrite:  false,
		PermissionInvitesCreate: false,
		PermissionProductsRead:  true,
		PermissionProductsWrite: false,
		PermissionKanbanRead:    true,
		PermissionKanbanWrite:   false,
	},
}

func Can(role domain.OrgRole, permission Permission) bool {
	if perms, ok := permissionMatrix[role]; ok {
		return perms[permission]
	}
	return false
}
