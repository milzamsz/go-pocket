package domain

import "time"

type User struct {
	ID    string
	Email string
	Name  string
}

type Organization struct {
	ID        string
	Slug      string
	Name      string
	OwnerID   string
	CreatedAt time.Time
}

type OrganizationMember struct {
	ID             string
	OrganizationID string
	UserID         string
	Role           OrgRole
}

type OrganizationMemberProfile struct {
	OrganizationID string
	UserID         string
	Role           OrgRole
	Name           string
	Email          string
}

type Invitation struct {
	ID             string
	OrganizationID string
	Email          string
	Role           OrgRole
	Token          string
	ExpiresAt      time.Time
}

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
	OrgRoleViewer OrgRole = "viewer"
)

type Product struct {
	ID             string
	Name           string
	Category       string
	Price          int64 // represented in cents
	Stock          int
	Active         bool
	OrganizationID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type KanbanColumn struct {
	ID             string
	Name           string
	Position       int
	OrganizationID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type KanbanCard struct {
	ID             string
	Title          string
	Description    string
	Badge          string
	Position       int
	ColumnID       string
	OrganizationID string
	AssigneeID     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
