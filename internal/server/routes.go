package server

import (
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/handlers"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(se *core.ServeEvent, deps *Deps) {
	// Static assets (filesystem first, embedded fallback).
	se.Router.GET("/assets/{path...}", handlers.AssetFile())

	// Public marketing/docs surface.
	se.Router.GET("/", handlers.Page("home"))
	se.Router.GET("/pricing", handlers.Page("pricing"))
	se.Router.GET("/about", handlers.Page("about"))
	se.Router.GET("/contact", handlers.Page("contact"))
	se.Router.GET("/blog", handlers.Page("blog"))
	se.Router.GET("/blog/{slug}", handlers.Page("blog-post"))
	se.Router.GET("/docs", handlers.Page("docs"))
	se.Router.GET("/docs/{path...}", handlers.Page("docs-page"))
	se.Router.GET("/help", handlers.Page("help"))
	se.Router.GET("/sitemap.xml", handlers.PlainXML("<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset></urlset>"))
	se.Router.GET("/robots.txt", handlers.PlainText("User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /app\nDisallow: /api"))
	se.Router.GET("/feed.xml", handlers.PlainXML("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\"></rss>"))

	// Auth routes.
	authGroup := se.Router.Group("/auth")
	authGroup.GET("/login", handlers.Page("auth-login"))
	authGroup.POST("/login", handlers.AuthLogin(deps.Auth))
	authGroup.GET("/signup", handlers.Page("auth-signup"))
	authGroup.POST("/signup", handlers.AuthSignup(deps.Auth))
	authGroup.GET("/logout", handlers.AuthLogout(deps.Auth))
	authGroup.GET("/forgot-password", handlers.Page("auth-forgot-password"))
	authGroup.POST("/forgot-password", handlers.AuthForgotPassword(deps.Auth))
	authGroup.GET("/reset-password", handlers.Page("auth-reset-password"))
	authGroup.POST("/reset-password", handlers.AuthResetPassword(deps.Auth))
	authGroup.GET("/verify-email", handlers.AuthVerifyEmail(deps.Auth))
	authGroup.GET("/oauth/{provider}", handlers.AuthOAuthStart(deps.Auth))
	authGroup.GET("/oauth/{provider}/callback", handlers.AuthOAuthCallback(deps.Auth))

	// App shell routes (user-level).
	appGroup := se.Router.Group("/app")
	appGroup.BindFunc(
		middleware.BindAuthContext(),
		middleware.RequireAuthenticated(),
	)
	appGroup.GET("/{$}", handlers.Page("app-dashboard"))
	appGroup.GET("/onboarding", handlers.Page("app-onboarding"))
	appGroup.POST("/onboarding", handlers.Accepted("app-onboarding"))
	appGroup.POST("/switch-org", handlers.Accepted("app-switch-org"))
	appGroup.GET("/settings/profile", handlers.Page("app-settings-profile"))
	appGroup.POST("/settings/profile", handlers.AppSettingsProfile(deps.Auth))
	appGroup.GET("/settings/security", handlers.Page("app-settings-security"))
	appGroup.POST("/settings/security/password", handlers.AppSettingsSecurityPassword(deps.Auth))
	appGroup.POST("/settings/security/2fa", handlers.AppSettingsSecurity2FA(deps.Auth))
	appGroup.GET("/settings/account", handlers.Page("app-settings-account"))

	// Org-scoped routes.
	orgGroup := se.Router.Group("/org/{slug}")
	orgGroup.BindFunc(
		middleware.BindAuthContext(),
		middleware.RequireAuthenticated(),
		middleware.RequireOrgRole(tenancy.PermissionMembersRead, domain.OrgRoleViewer),
	)
	orgGroup.GET("/{$}", handlers.Page("org-overview"))
	se.Router.GET("/healthz", handlers.Health)
	orgGroup.GET("/members", handlers.ListOrgMembers(deps.Tenancy))
	orgGroup.GET("/members/{userID}", handlers.ShowOrgMemberProfile(deps.Tenancy))

	// Products CRUD routes
	orgGroup.GET("/products", handlers.ListProducts(deps.Products))
	orgGroup.GET("/products/{id}", handlers.ShowProductDetail(deps.Products))
	orgGroup.POST("/products", handlers.CreateProduct(deps.Products)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionProductsWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/products/{id}", handlers.UpdateProduct(deps.Products)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionProductsWrite, domain.OrgRoleMember),
	)
	orgGroup.DELETE("/products/{id}", handlers.DeleteProduct(deps.Products)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionProductsWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/products/bulk-delete", handlers.BulkDeleteProducts(deps.Products)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionProductsWrite, domain.OrgRoleMember),
	)

	// Kanban Board routes
	orgGroup.GET("/kanban", handlers.ShowKanbanBoard(deps.Kanban))
	orgGroup.POST("/kanban/columns", handlers.CreateKanbanColumn(deps.Kanban)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionKanbanWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/kanban/cards", handlers.CreateKanbanCard(deps.Kanban)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionKanbanWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/kanban/cards/{id}", handlers.UpdateKanbanCard(deps.Kanban)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionKanbanWrite, domain.OrgRoleMember),
	)
	orgGroup.DELETE("/kanban/cards/{id}", handlers.DeleteKanbanCard(deps.Kanban)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionKanbanWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/kanban/reorder", handlers.ReorderKanbanCard(deps.Kanban)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionKanbanWrite, domain.OrgRoleMember),
	)
	orgGroup.POST("/members/invite", handlers.OrgMembersInvite(deps.Tenancy, deps.Email)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionInvitesCreate, domain.OrgRoleAdmin),
	)
	orgGroup.DELETE("/members/{userID}", handlers.OrgMembersRemove(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionMembersWrite, domain.OrgRoleAdmin),
	)
	orgGroup.PATCH("/members/{userID}/role", handlers.OrgMembersChangeRole(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionMembersWrite, domain.OrgRoleAdmin),
	)
	orgGroup.GET("/invitations", handlers.Page("org-invitations"))
	orgGroup.POST("/invitations/{id}/resend", handlers.OrgInvitationResend(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionInvitesCreate, domain.OrgRoleAdmin),
	)
	orgGroup.POST("/invitations/{id}/revoke", handlers.OrgInvitationRevoke(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionInvitesCreate, domain.OrgRoleAdmin),
	)
	orgGroup.GET("/billing", handlers.Page("org-billing"))
	orgGroup.POST("/billing/checkout", handlers.OrgBillingCheckout(deps.Billing)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionBillingWrite, domain.OrgRoleAdmin),
	)
	orgGroup.POST("/billing/portal", handlers.OrgBillingPortal(deps.Billing)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionBillingWrite, domain.OrgRoleAdmin),
	)
	orgGroup.GET("/billing/invoices", handlers.Page("org-billing-invoices"))
	orgGroup.GET("/settings", handlers.Page("org-settings"))
	orgGroup.POST("/settings", handlers.OrgSettingsUpdate(deps.Tenancy, deps.Email)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionMembersWrite, domain.OrgRoleAdmin),
	)
	orgGroup.GET("/settings/danger", handlers.Page("org-settings-danger"))
	orgGroup.POST("/settings/transfer", handlers.OrgSettingsTransfer(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionMembersWrite, domain.OrgRoleOwner),
	)
	orgGroup.POST("/settings/delete", handlers.OrgSettingsDelete(deps.Tenancy)).BindFunc(
		middleware.RequireOrgRole(tenancy.PermissionMembersWrite, domain.OrgRoleOwner),
	)
	orgGroup.GET("/audit", handlers.Page("org-audit"))

	// Invite token routes.
	se.Router.GET("/invite/{token}", handlers.Page("invite-accept"))
	se.Router.POST("/invite/{token}", handlers.InviteAccept(deps.Tenancy)).BindFunc(
		middleware.BindAuthContext(),
		middleware.RequireAuthenticated(),
	)
	se.Router.POST("/invite/{token}/decline", handlers.InviteDecline(deps.Tenancy)).BindFunc(
		middleware.BindAuthContext(),
		middleware.RequireAuthenticated(),
	)

	// Platform admin routes.
	adminGroup := se.Router.Group("/admin")
	adminGroup.BindFunc(
		middleware.BindAuthContext(),
		middleware.RequireAuthenticated(),
		middleware.RequireSuperuser(),
	)
	adminGroup.GET("/", handlers.AdminDashboard())
	adminGroup.GET("/users", handlers.AdminUsers())
	adminGroup.GET("/users/{id}", handlers.AdminUserDetail())
	adminGroup.GET("/organizations", handlers.AdminOrganizations())
	adminGroup.GET("/organizations/{id}", handlers.AdminOrganizationDetail())
	adminGroup.GET("/analytics", handlers.AdminAnalytics())
	adminGroup.GET("/audit", handlers.AdminAudit())
	adminGroup.GET("/settings", handlers.Page("admin-settings"))

	// Webhooks.
	se.Router.POST("/webhooks/polar", handlers.PolarWebhook(deps.Billing))
	se.Router.POST("/webhooks/resend", handlers.ResendWebhook(deps.Email))
}
