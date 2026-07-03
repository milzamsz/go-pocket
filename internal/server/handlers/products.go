package handlers

import (
	"net/http"
	"strconv"
	"strings"

	orgpage "github.com/milzamsz/go-pocket/components/pages/org"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/products"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func ListProducts(productsSvc products.Service, tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}

		q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
		category := strings.TrimSpace(e.Request.URL.Query().Get("category"))
		sort := strings.TrimSpace(e.Request.URL.Query().Get("sort"))
		if sort == "" {
			sort = "-created"
		}

		page := 1
		if pStr := e.Request.URL.Query().Get("page"); pStr != "" {
			if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
				page = p
			}
		}

		pageSize := 20
		if psStr := e.Request.URL.Query().Get("pageSize"); psStr != "" {
			if ps, err := strconv.Atoi(psStr); err == nil && ps > 0 {
				pageSize = ps
			}
		}

		var active *bool
		if actStr := e.Request.URL.Query().Get("active"); actStr != "" {
			b := actStr == "true"
			active = &b
		}

		rows, total, err := productsSvc.List(e.Request.Context(), orgCtx.OrgID, q, category, active, sort, page, pageSize)
		if err != nil {
			return e.BadRequestError("failed to list products", err)
		}

		isHX := e.Request.Header.Get("HX-Request") == "true"
		if isHX {
			// If it is HTMX request, we can render the inner table body partial
			return renderHTML(e, http.StatusOK, orgpage.ProductsTableBody(orgCtx.Slug, rows, q, category, sort, page, pageSize, total))
		}

		return renderHTML(e, http.StatusOK, orgpage.Products(shell, orgCtx.Slug, rows, q, category, sort, page, pageSize, total))
	}
}

func CreateProduct(productsSvc products.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		name := strings.TrimSpace(e.Request.FormValue("name"))
		category := strings.TrimSpace(e.Request.FormValue("category"))
		priceStr := strings.TrimSpace(e.Request.FormValue("price"))
		stockStr := strings.TrimSpace(e.Request.FormValue("stock"))
		active := e.Request.FormValue("active") == "on" || e.Request.FormValue("active") == "true"

		price, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			return e.BadRequestError("invalid price value", err)
		}

		stock, err := strconv.Atoi(stockStr)
		if err != nil {
			return e.BadRequestError("invalid stock value", err)
		}

		prod := domain.Product{
			Name:     name,
			Category: category,
			Price:    price,
			Stock:    stock,
			Active:   active,
		}

		_, err = productsSvc.Create(e.Request.Context(), orgCtx.OrgID, prod)
		if err != nil {
			return e.BadRequestError("failed to create product", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/products")
	}
}

func UpdateProduct(productsSvc products.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		id := e.Request.PathValue("id")
		name := strings.TrimSpace(e.Request.FormValue("name"))
		category := strings.TrimSpace(e.Request.FormValue("category"))
		priceStr := strings.TrimSpace(e.Request.FormValue("price"))
		stockStr := strings.TrimSpace(e.Request.FormValue("stock"))
		active := e.Request.FormValue("active") == "on" || e.Request.FormValue("active") == "true"

		price, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			return e.BadRequestError("invalid price value", err)
		}

		stock, err := strconv.Atoi(stockStr)
		if err != nil {
			return e.BadRequestError("invalid stock value", err)
		}

		prod := domain.Product{
			ID:       id,
			Name:     name,
			Category: category,
			Price:    price,
			Stock:    stock,
			Active:   active,
		}

		_, err = productsSvc.Update(e.Request.Context(), orgCtx.OrgID, prod)
		if err != nil {
			return e.BadRequestError("failed to update product", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/products")
	}
}

func DeleteProduct(productsSvc products.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		id := e.Request.PathValue("id")
		if err := productsSvc.Delete(e.Request.Context(), orgCtx.OrgID, id); err != nil {
			return e.BadRequestError("failed to delete product", err)
		}

		// Return 200 OK or redirect
		isHX := e.Request.Header.Get("HX-Request") == "true"
		if isHX {
			e.Response.Header().Set("HX-Trigger", "refreshProducts")
			return e.NoContent(http.StatusOK)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/products")
	}
}

func BulkDeleteProducts(productsSvc products.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		_ = e.Request.ParseForm()
		ids := e.Request.Form["ids"]

		if err := productsSvc.BulkDelete(e.Request.Context(), orgCtx.OrgID, ids); err != nil {
			return e.BadRequestError("failed to bulk delete products", err)
		}

		isHX := e.Request.Header.Get("HX-Request") == "true"
		if isHX {
			e.Response.Header().Set("HX-Trigger", "refreshProducts")
			return e.NoContent(http.StatusOK)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/products")
	}
}
