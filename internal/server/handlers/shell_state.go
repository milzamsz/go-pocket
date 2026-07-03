package handlers

import (
	"strings"
	"unicode"

	"github.com/milzamsz/go-pocket/components/layouts"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func appShellState(e *core.RequestEvent) (layouts.ShellState, error) {
	shell, err := actorShellState(e)
	if err != nil {
		return layouts.ShellState{}, err
	}
	shell.ContextName = "Application"
	shell.ProfileHref = "/app/settings/profile"
	return shell, nil
}

func adminShellState(e *core.RequestEvent) (layouts.ShellState, error) {
	shell, err := actorShellState(e)
	if err != nil {
		return layouts.ShellState{}, err
	}
	shell.ContextName = "Platform Admin"
	shell.ProfileHref = "/app/settings/profile"
	return shell, nil
}

func orgShellState(e *core.RequestEvent, tenancySvc tenancy.Service) (layouts.ShellState, error) {
	shell, err := actorShellState(e)
	if err != nil {
		return layouts.ShellState{}, err
	}
	orgCtx, ok := middleware.GetOrgContext(e)
	if !ok {
		return layouts.ShellState{}, domain.ErrForbidden
	}
	orgShell, err := tenancySvc.GetOrganizationShell(e.Request.Context(), orgCtx.OrgID, orgCtx.Role)
	if err != nil {
		return layouts.ShellState{}, err
	}

	shell.ContextName = firstNonEmpty(orgShell.Name, orgCtx.Slug, "Organization")
	shell.ProfileHref = "/app/settings/profile"
	shell.PlanLabel = normalizePlanLabel(orgShell.Plan)
	shell.SubscriptionStatus = normalizeStatusLabel(orgShell.SubscriptionStatus)
	shell.ShowTrialBanner = false
	return shell, nil
}

func actorShellState(e *core.RequestEvent) (layouts.ShellState, error) {
	actor, err := middleware.CurrentActor(e)
	if err != nil {
		return layouts.ShellState{}, err
	}
	name := strings.TrimSpace(actor.Email)
	if e.Auth != nil {
		if authName := strings.TrimSpace(e.Auth.GetString("name")); authName != "" {
			name = authName
		}
	}
	return layouts.ShellState{
		UserName:     name,
		UserEmail:    strings.TrimSpace(actor.Email),
		UserInitials: userInitials(name, actor.Email),
	}, nil
}

func userInitials(name string, email string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		parts := strings.Fields(name)
		if len(parts) == 1 {
			r := []rune(parts[0])
			if len(r) >= 2 {
				return strings.ToUpper(string(r[:2]))
			}
			return strings.ToUpper(parts[0])
		}
		return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[1])[0]))
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	local, _, found := strings.Cut(email, "@")
	if !found || local == "" {
		local = email
	}
	r := []rune(local)
	if len(r) >= 2 {
		return strings.ToUpper(string(r[:2]))
	}
	return strings.ToUpper(local)
}

func normalizePlanLabel(plan string) string {
	plan = strings.TrimSpace(plan)
	if plan == "" {
		return ""
	}
	if strings.HasPrefix(plan, "prod_") {
		plan = strings.TrimPrefix(plan, "prod_")
	}
	plan = strings.ReplaceAll(plan, "_", " ")
	plan = strings.TrimSpace(plan)
	if plan == "" {
		return ""
	}
	return titleWords(plan)
}

func normalizeStatusLabel(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "_", " ")
	return titleWords(status)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func titleWords(value string) string {
	var out []rune
	upperNext := true
	for _, r := range []rune(strings.ToLower(value)) {
		if unicode.IsSpace(r) {
			upperNext = true
			out = append(out, r)
			continue
		}
		if upperNext {
			out = append(out, unicode.ToUpper(r))
			upperNext = false
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
