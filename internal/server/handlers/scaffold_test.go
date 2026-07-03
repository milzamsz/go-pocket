package handlers

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPageRegistryIncludesCoreSurfaces(t *testing.T) {
	t.Parallel()

	required := []string{
		"auth-login",
		"auth-signup",
		"auth-forgot-password",
		"auth-reset-password",
		"org-billing-invoices",
		"invite-accept",
	}

	for _, page := range required {
		if _, ok := pageRegistry[page]; !ok {
			t.Fatalf("expected page %q to be registered", page)
		}
	}
}

func TestSimpleAcceptedPageRendersHTML(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	component := simpleAcceptedPage("org-billing-checkout")
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render accepted page: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, "<html") {
		t.Fatalf("expected html document in rendered output")
	}
	if !strings.Contains(rendered, "org-billing-checkout") {
		t.Fatalf("expected action marker in rendered output")
	}
}
