package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/milzamsz/go-pocket/components/layouts"
	"github.com/milzamsz/go-pocket/components/pages/app"
	"github.com/milzamsz/go-pocket/components/pages/auth"
	"github.com/milzamsz/go-pocket/components/pages/marketing"
	"github.com/pocketbase/pocketbase/core"
)

type renderable interface {
	Render(ctx context.Context, w io.Writer) error
}

type pageBuilder func(e *core.RequestEvent) renderable

var pageRegistry = map[string]pageBuilder{
	"home":    func(_ *core.RequestEvent) renderable { return marketing.Home() },
	"pricing": simplePlaceholder("Pricing"),
	"about":   simplePlaceholder("About"),
	"contact": simplePlaceholder("Contact"),
	"blog":    func(_ *core.RequestEvent) renderable { return marketing.BlogIndex(nil) },
	"blog-post": func(e *core.RequestEvent) renderable {
		return marketing.BlogPost(e.Request.PathValue("slug"), "Blog summary", "<p class=\"text-sm text-muted-foreground\">Post content scaffold.</p>")
	},
	"docs": func(_ *core.RequestEvent) renderable { return marketing.DocsIndex(nil) },
	"docs-page": func(e *core.RequestEvent) renderable {
		return marketing.DocsPage(e.Request.PathValue("path"), "Documentation page scaffold", "<p class=\"text-sm text-muted-foreground\">Docs content scaffold.</p>")
	},
	"help":                 func(_ *core.RequestEvent) renderable { return marketing.HelpCenter() },
	"auth-login":           func(_ *core.RequestEvent) renderable { return auth.Login() },
	"auth-signup":          func(_ *core.RequestEvent) renderable { return auth.Signup() },
	"auth-forgot-password": func(_ *core.RequestEvent) renderable { return auth.ForgotPassword() },
	"auth-reset-password":  func(_ *core.RequestEvent) renderable { return auth.ResetPassword() },
	"auth-verify-email":    func(_ *core.RequestEvent) renderable { return auth.VerifyEmail() },
	"org-billing-invoices": simplePlaceholder("Billing Invoices"),
	"invite-accept":        simplePlaceholder("Accept Invitation"),
}

func Page(name string) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		builder, ok := pageRegistry[name]
		if !ok {
			return renderHTML(e, http.StatusOK, simplePlaceholder(name)(e))
		}
		return renderHTML(e, http.StatusOK, builder(e))
	}
}

// Accepted is used for scaffold POST/PATCH/DELETE handlers.
func Accepted(action string) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		return renderHTML(e, http.StatusAccepted, simpleAcceptedPage(action))
	}
}

func PlainText(text string) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		return e.String(http.StatusOK, text)
	}
}

func PlainXML(xml string) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		e.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
		return e.String(http.StatusOK, xml)
	}
}

func renderHTML(e *core.RequestEvent, status int, component renderable) error {
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	e.Response.WriteHeader(status)
	if err := component.Render(context.Background(), e.Response); err != nil {
		return fmt.Errorf("render html: %w", err)
	}
	return nil
}

func simplePlaceholder(title string) pageBuilder {
	return func(_ *core.RequestEvent) renderable {
		return layouts.Surface(
			layouts.Meta{Title: title, Description: "Scaffold page"},
			title,
			"This route is scaffolded and now renders HTML.",
			app.Placeholder("Connect service data and actions for this page."),
		)
	}
}

func simpleAcceptedPage(action string) renderable {
	return layouts.Surface(
		layouts.Meta{Title: "Accepted", Description: "Action accepted"},
		"Request Accepted",
		"Your request was accepted and queued.",
		app.Placeholder("Action: "+action+". Backend flow remains scaffolded."),
	)
}
