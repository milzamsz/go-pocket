package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	adminpage "github.com/milzamsz/go-pocket/components/pages/admin"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func AdminDashboard() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		users := countRecords(e, "users")
		orgs := countRecords(e, "organizations")
		subs := countRecords(e, "subscriptions")
		events := countRecords(e, "webhook_events")
		return renderHTML(e, http.StatusOK, adminpage.Dashboard(adminpage.DashboardStats{
			Users:         users,
			Organizations: orgs,
			Subscriptions: subs,
			WebhookEvents: events,
		}))
	}
}

func AdminUsers() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		pager := parseAdminPager(e, 20)
		filter := ""
		params := dbx.Params{}
		if pager.Query != "" {
			filter = "email ~ {:q} || name ~ {:q}"
			params["q"] = pager.Query
		}

		records, err := e.App.FindRecordsByFilter("users", filter, "-created", pager.PerPage+1, (pager.Page-1)*pager.PerPage, params)
		if err != nil {
			return e.InternalServerError("failed to list users", err)
		}
		pager.HasNext = len(records) > pager.PerPage
		records = records[:min(len(records), pager.PerPage)]
		rows := make([]adminpage.UserRow, 0, len(records))
		for _, user := range records {
			orgCount, _ := e.App.CountRecords("organization_members", dbx.HashExp{"user": user.Id})
			rows = append(rows, adminpage.UserRow{
				ID:        user.Id,
				Email:     user.Email(),
				Name:      user.GetString("name"),
				Verified:  user.Verified(),
				CreatedAt: toRFC3339(user.GetDateTime("created").Time()),
				OrgCount:  int(orgCount),
			})
		}
		return renderHTML(e, http.StatusOK, adminpage.Users(adminpage.UserListData{
			Rows:   rows,
			Pager:  pager,
			Base:   "/admin/users",
			Sort:   "-created",
			Filter: filter,
		}))
	}
}

func AdminUserDetail() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := strings.TrimSpace(e.Request.PathValue("id"))
		record, err := e.App.FindRecordById("users", id)
		if err != nil {
			return e.NotFoundError("user not found", err)
		}
		memberships, _ := e.App.FindRecordsByFilter("organization_members", "user = {:user}", "", 200, 0, dbx.Params{"user": id})
		entries := make([]adminpage.MembershipRow, 0, len(memberships))
		for _, m := range memberships {
			org, findErr := e.App.FindRecordById("organizations", m.GetString("organization"))
			if findErr != nil {
				continue
			}
			entries = append(entries, adminpage.MembershipRow{
				OrgID:   org.Id,
				OrgSlug: org.GetString("slug"),
				Role:    m.GetString("role"),
			})
		}
		return renderHTML(e, http.StatusOK, adminpage.UserDetail(adminpage.UserDetailData{
			ID:            record.Id,
			Email:         record.Email(),
			Name:          record.GetString("name"),
			Verified:      record.Verified(),
			CreatedAt:     toRFC3339(record.GetDateTime("created").Time()),
			Organizations: entries,
		}))
	}
}

func AdminOrganizations() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		pager := parseAdminPager(e, 20)
		rows, hasNext, err := adminOrganizationRows(e.App, pager)
		if err != nil {
			return e.InternalServerError("failed to list organizations", err)
		}
		pager.HasNext = hasNext
		return renderHTML(e, http.StatusOK, adminpage.Organizations(adminpage.OrganizationListData{
			Rows:  rows,
			Pager: pager,
			Base:  "/admin/organizations",
		}))
	}
}

func AdminOrganizationDetail() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := strings.TrimSpace(e.Request.PathValue("id"))
		org, err := e.App.FindRecordById("organizations", id)
		if err != nil {
			return e.NotFoundError("organization not found", err)
		}
		memberCount, _ := e.App.CountRecords("organization_members", dbx.HashExp{"organization": id})
		inviteCount, _ := e.App.CountRecords("invitations", dbx.HashExp{"organization": id})

		return renderHTML(e, http.StatusOK, adminpage.OrganizationDetail(adminpage.OrganizationDetailData{
			ID:                 org.Id,
			Slug:               org.GetString("slug"),
			Name:               org.GetString("name"),
			OwnerID:            org.GetString("owner"),
			MemberCount:        int(memberCount),
			InvitationCount:    int(inviteCount),
			Plan:               org.GetString("plan"),
			SubscriptionStatus: org.GetString("subscription_status"),
		}))
	}
}

func AdminAnalytics() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		pager := parseAdminPager(e, 20)
		rows, _, err := adminOrganizationRows(e.App, pager)
		if err != nil {
			return e.InternalServerError("failed to load analytics", err)
		}
		return renderHTML(e, http.StatusOK, adminpage.Analytics(rows))
	}
}

func AdminAudit() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		pager := parseAdminPager(e, 20)
		filter := ""
		params := dbx.Params{}
		if pager.Query != "" {
			filter = "provider ~ {:q} || event_type ~ {:q} || family ~ {:q}"
			params["q"] = pager.Query
		}
		records, err := e.App.FindRecordsByFilter("webhook_events", filter, "-received_at", pager.PerPage+1, (pager.Page-1)*pager.PerPage, params)
		if err != nil {
			return e.InternalServerError("failed to load audit events", err)
		}
		pager.HasNext = len(records) > pager.PerPage
		records = records[:min(len(records), pager.PerPage)]
		rows := make([]adminpage.AuditRow, 0, len(records))
		for _, record := range records {
			occurred := record.GetDateTime("occurred_at").Time()
			if occurred.IsZero() {
				occurred = record.GetDateTime("received_at").Time()
			}
			rows = append(rows, adminpage.AuditRow{
				Provider:   record.GetString("provider"),
				EventType:  record.GetString("event_type"),
				Family:     record.GetString("family"),
				Status:     record.GetString("status"),
				OccurredAt: toRFC3339(occurred),
			})
		}
		return renderHTML(e, http.StatusOK, adminpage.Audit(adminpage.AuditListData{
			Rows:  rows,
			Pager: pager,
			Base:  "/admin/audit",
		}))
	}
}

func adminOrganizationRows(app core.App, pager adminpage.PagerState) ([]adminpage.OrganizationRow, bool, error) {
	filter := ""
	params := dbx.Params{}
	if pager.Query != "" {
		filter = "slug ~ {:q} || name ~ {:q}"
		params["q"] = pager.Query
	}

	orgs, err := app.FindRecordsByFilter("organizations", filter, "-created", pager.PerPage+1, (pager.Page-1)*pager.PerPage, params)
	if err != nil {
		return nil, false, err
	}
	hasNext := len(orgs) > pager.PerPage
	orgs = orgs[:min(len(orgs), pager.PerPage)]
	rows := make([]adminpage.OrganizationRow, 0, len(orgs))
	for _, org := range orgs {
		memberCount, _ := app.CountRecords("organization_members", dbx.HashExp{"organization": org.Id})
		inviteCount, _ := app.CountRecords("invitations", dbx.HashExp{"organization": org.Id})
		rows = append(rows, adminpage.OrganizationRow{
			ID:                 org.Id,
			Slug:               org.GetString("slug"),
			Name:               org.GetString("name"),
			OwnerID:            org.GetString("owner"),
			MemberCount:        int(memberCount),
			InvitationCount:    int(inviteCount),
			Plan:               org.GetString("plan"),
			SubscriptionStatus: org.GetString("subscription_status"),
		})
	}
	return rows, hasNext, nil
}

func countRecords(e *core.RequestEvent, collection string) int {
	count, err := e.App.CountRecords(collection)
	if err != nil {
		return 0
	}
	return int(count)
}

func toRFC3339(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func parseAdminPager(e *core.RequestEvent, perPage int) adminpage.PagerState {
	q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
	page := 1
	if raw := strings.TrimSpace(e.Request.URL.Query().Get("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}
	return adminpage.PagerState{
		Query:   q,
		Page:    page,
		PerPage: perPage,
		HasPrev: page > 1,
	}
}
