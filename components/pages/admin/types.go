package admin

type DashboardStats struct {
	Users         int
	Organizations int
	Subscriptions int
	WebhookEvents int
}

type PagerState struct {
	Query   string
	Page    int
	PerPage int
	HasPrev bool
	HasNext bool
}

type UserRow struct {
	ID        string
	Email     string
	Name      string
	Verified  bool
	CreatedAt string
	OrgCount  int
}

type UserListData struct {
	Rows   []UserRow
	Pager  PagerState
	Base   string
	Sort   string
	Filter string
}

type UserDetailData struct {
	ID            string
	Email         string
	Name          string
	Verified      bool
	CreatedAt     string
	Organizations []MembershipRow
}

type MembershipRow struct {
	OrgID   string
	OrgSlug string
	Role    string
}

type OrganizationRow struct {
	ID                 string
	Slug               string
	Name               string
	OwnerID            string
	MemberCount        int
	InvitationCount    int
	Plan               string
	SubscriptionStatus string
}

type OrganizationListData struct {
	Rows  []OrganizationRow
	Pager PagerState
	Base  string
}

type OrganizationDetailData struct {
	ID                 string
	Slug               string
	Name               string
	OwnerID            string
	MemberCount        int
	InvitationCount    int
	Plan               string
	SubscriptionStatus string
}

type AuditRow struct {
	Provider   string
	EventType  string
	Family     string
	Status     string
	OccurredAt string
}

type AuditListData struct {
	Rows  []AuditRow
	Pager PagerState
	Base  string
}
