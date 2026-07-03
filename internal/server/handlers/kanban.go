package handlers

import (
	"net/http"
	"strconv"
	"strings"

	orgpage "github.com/milzamsz/go-pocket/components/pages/org"
	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/milzamsz/go-pocket/internal/services/kanban"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase/core"
)

func ShowKanbanBoard(kanbanSvc kanban.Service, tenancySvc tenancy.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}
		shell, err := orgShellState(e, tenancySvc)
		if err != nil {
			return e.ForbiddenError("failed to resolve shell state", err)
		}

		cols, cards, err := kanbanSvc.GetBoard(e.Request.Context(), orgCtx.OrgID)
		if err != nil {
			return e.BadRequestError("failed to load kanban board", err)
		}

		return renderHTML(e, http.StatusOK, orgpage.Kanban(shell, orgCtx.Slug, cols, cards))
	}
}

func CreateKanbanColumn(kanbanSvc kanban.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		name := strings.TrimSpace(e.Request.FormValue("name"))
		if name == "" {
			return e.BadRequestError("column name is required", nil)
		}

		_, err := kanbanSvc.CreateColumn(e.Request.Context(), orgCtx.OrgID, domain.KanbanColumn{Name: name})
		if err != nil {
			return e.BadRequestError("failed to create column", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/kanban")
	}
}

func CreateKanbanCard(kanbanSvc kanban.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		title := strings.TrimSpace(e.Request.FormValue("title"))
		description := strings.TrimSpace(e.Request.FormValue("description"))
		badge := strings.TrimSpace(e.Request.FormValue("badge"))
		columnID := strings.TrimSpace(e.Request.FormValue("column"))

		if title == "" || columnID == "" {
			return e.BadRequestError("card title and column are required", nil)
		}

		_, err := kanbanSvc.CreateCard(e.Request.Context(), orgCtx.OrgID, domain.KanbanCard{
			Title:       title,
			Description: description,
			Badge:       badge,
			ColumnID:    columnID,
		})
		if err != nil {
			return e.BadRequestError("failed to create card", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/kanban")
	}
}

func UpdateKanbanCard(kanbanSvc kanban.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		id := e.Request.PathValue("id")
		title := strings.TrimSpace(e.Request.FormValue("title"))
		description := strings.TrimSpace(e.Request.FormValue("description"))
		badge := strings.TrimSpace(e.Request.FormValue("badge"))

		if title == "" {
			return e.BadRequestError("card title is required", nil)
		}

		_, err := kanbanSvc.UpdateCard(e.Request.Context(), orgCtx.OrgID, domain.KanbanCard{
			ID:          id,
			Title:       title,
			Description: description,
			Badge:       badge,
		})
		if err != nil {
			return e.BadRequestError("failed to update card", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/kanban")
	}
}

func DeleteKanbanCard(kanbanSvc kanban.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		id := e.Request.PathValue("id")
		if err := kanbanSvc.DeleteCard(e.Request.Context(), orgCtx.OrgID, id); err != nil {
			return e.BadRequestError("failed to delete card", err)
		}

		return e.Redirect(http.StatusSeeOther, "/org/"+orgCtx.Slug+"/kanban")
	}
}

func ReorderKanbanCard(kanbanSvc kanban.Service) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		orgCtx, ok := middleware.GetOrgContext(e)
		if !ok {
			return e.ForbiddenError("organization context missing", domain.ErrForbidden)
		}

		cardID := strings.TrimSpace(e.Request.FormValue("card_id"))
		toColumnID := strings.TrimSpace(e.Request.FormValue("to_column"))
		toIndexStr := strings.TrimSpace(e.Request.FormValue("to_index"))

		toIndex, err := strconv.Atoi(toIndexStr)
		if err != nil {
			return e.BadRequestError("invalid toIndex value", err)
		}

		if cardID == "" || toColumnID == "" {
			return e.BadRequestError("card_id and to_column are required", nil)
		}

		if err := kanbanSvc.MoveCard(e.Request.Context(), orgCtx.OrgID, cardID, toColumnID, toIndex); err != nil {
			return e.BadRequestError("failed to move card", err)
		}

		// HTMX request will expect a success response, DOM is already moved by SortableJS,
		// but we can return 200 OK
		return e.NoContent(http.StatusOK)
	}
}
