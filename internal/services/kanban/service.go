package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type Service interface {
	GetBoard(ctx context.Context, orgID string) ([]domain.KanbanColumn, []domain.KanbanCard, error)
	CreateColumn(ctx context.Context, orgID string, col domain.KanbanColumn) (domain.KanbanColumn, error)
	CreateCard(ctx context.Context, orgID string, card domain.KanbanCard) (domain.KanbanCard, error)
	UpdateCard(ctx context.Context, orgID string, card domain.KanbanCard) (domain.KanbanCard, error)
	DeleteCard(ctx context.Context, orgID string, id string) error
	MoveCard(ctx context.Context, orgID string, cardID string, toColumnID string, toIndex int) error
}

type service struct {
	app core.App
}

func New(app core.App) Service {
	return &service{app: app}
}

func (s *service) GetBoard(ctx context.Context, orgID string) ([]domain.KanbanColumn, []domain.KanbanCard, error) {
	colColl, err := s.app.FindCollectionByNameOrId("kanban_columns")
	if err != nil {
		return nil, nil, fmt.Errorf("find kanban_columns collection: %w", err)
	}

	cardColl, err := s.app.FindCollectionByNameOrId("kanban_cards")
	if err != nil {
		return nil, nil, fmt.Errorf("find kanban_cards collection: %w", err)
	}

	// 1. Fetch columns sorted by position
	colQuery := s.app.RecordQuery(colColl).
		AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgID})).
		OrderBy("position asc").
		WithContext(ctx)

	colRecords := make([]*core.Record, 0)
	if err := colQuery.All(&colRecords); err != nil {
		return nil, nil, fmt.Errorf("query kanban columns: %w", err)
	}

	cols := make([]domain.KanbanColumn, 0, len(colRecords))
	for _, r := range colRecords {
		cols = append(cols, domain.KanbanColumn{
			ID:             r.Id,
			Name:           r.GetString("name"),
			Position:       r.GetInt("position"),
			OrganizationID: r.GetString("organization"),
			CreatedAt:      r.GetDateTime("created").Time(),
			UpdatedAt:      r.GetDateTime("updated").Time(),
		})
	}

	// 2. Fetch cards sorted by position
	cardQuery := s.app.RecordQuery(cardColl).
		AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgID})).
		OrderBy("position asc").
		WithContext(ctx)

	cardRecords := make([]*core.Record, 0)
	if err := cardQuery.All(&cardRecords); err != nil {
		return nil, nil, fmt.Errorf("query kanban cards: %w", err)
	}

	cards := make([]domain.KanbanCard, 0, len(cardRecords))
	for _, r := range cardRecords {
		cards = append(cards, domain.KanbanCard{
			ID:             r.Id,
			Title:          r.GetString("title"),
			Description:    r.GetString("description"),
			Badge:          r.GetString("badge"),
			Position:       r.GetInt("position"),
			ColumnID:       r.GetString("column"),
			OrganizationID: r.GetString("organization"),
			AssigneeID:     r.GetString("assignee"),
			CreatedAt:      r.GetDateTime("created").Time(),
			UpdatedAt:      r.GetDateTime("updated").Time(),
		})
	}

	return cols, cards, nil
}

func (s *service) CreateColumn(ctx context.Context, orgID string, col domain.KanbanColumn) (domain.KanbanColumn, error) {
	collection, err := s.app.FindCollectionByNameOrId("kanban_columns")
	if err != nil {
		return domain.KanbanColumn{}, fmt.Errorf("find kanban_columns collection: %w", err)
	}

	// Find max position
	var maxPos int
	maxQuery := s.app.RecordQuery(collection).
		Select("max(position)").
		AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgID}))
	_ = maxQuery.Row(&maxPos) // ignore error if no columns exist yet

	record := core.NewRecord(collection)
	record.Set("name", col.Name)
	record.Set("position", maxPos+1000)
	record.Set("organization", orgID)

	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return domain.KanbanColumn{}, fmt.Errorf("create column: %w", err)
	}

	return domain.KanbanColumn{
		ID:             record.Id,
		Name:           record.GetString("name"),
		Position:       record.GetInt("position"),
		OrganizationID: record.GetString("organization"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) CreateCard(ctx context.Context, orgID string, card domain.KanbanCard) (domain.KanbanCard, error) {
	collection, err := s.app.FindCollectionByNameOrId("kanban_cards")
	if err != nil {
		return domain.KanbanCard{}, fmt.Errorf("find kanban_cards collection: %w", err)
	}

	// Find max position in the target column
	var maxPos int
	maxQuery := s.app.RecordQuery(collection).
		Select("max(position)").
		AndWhere(dbx.NewExp("organization = {:orgID} AND column = {:colID}", dbx.Params{"orgID": orgID, "colID": card.ColumnID}))
	_ = maxQuery.Row(&maxPos)

	record := core.NewRecord(collection)
	record.Set("title", card.Title)
	record.Set("description", card.Description)
	record.Set("badge", card.Badge)
	record.Set("position", maxPos+1000)
	record.Set("column", card.ColumnID)
	record.Set("organization", orgID)
	if card.AssigneeID != "" {
		record.Set("assignee", card.AssigneeID)
	}

	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return domain.KanbanCard{}, fmt.Errorf("create card: %w", err)
	}

	return domain.KanbanCard{
		ID:             record.Id,
		Title:          record.GetString("title"),
		Description:    record.GetString("description"),
		Badge:          record.GetString("badge"),
		Position:       record.GetInt("position"),
		ColumnID:       record.GetString("column"),
		OrganizationID: record.GetString("organization"),
		AssigneeID:     record.GetString("assignee"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) UpdateCard(ctx context.Context, orgID string, card domain.KanbanCard) (domain.KanbanCard, error) {
	record, err := s.app.FindFirstRecordByFilter(
		"kanban_cards",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": card.ID, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.KanbanCard{}, domain.ErrNotFound
		}
		return domain.KanbanCard{}, fmt.Errorf("find card for update: %w", err)
	}

	record.Set("title", card.Title)
	record.Set("description", card.Description)
	record.Set("badge", card.Badge)
	if card.AssigneeID != "" {
		record.Set("assignee", card.AssigneeID)
	} else {
		record.Set("assignee", nil)
	}

	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return domain.KanbanCard{}, fmt.Errorf("update card: %w", err)
	}

	return domain.KanbanCard{
		ID:             record.Id,
		Title:          record.GetString("title"),
		Description:    record.GetString("description"),
		Badge:          record.GetString("badge"),
		Position:       record.GetInt("position"),
		ColumnID:       record.GetString("column"),
		OrganizationID: record.GetString("organization"),
		AssigneeID:     record.GetString("assignee"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) DeleteCard(ctx context.Context, orgID string, id string) error {
	record, err := s.app.FindFirstRecordByFilter(
		"kanban_cards",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": id, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("find card for delete: %w", err)
	}

	if err := s.app.DeleteWithContext(ctx, record); err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	return nil
}

func (s *service) MoveCard(ctx context.Context, orgID string, cardID string, toColumnID string, toIndex int) error {
	// 1. Find the card to move
	cardRecord, err := s.app.FindFirstRecordByFilter(
		"kanban_cards",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": cardID, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("find card for move: %w", err)
	}

	// 2. Query target column's existing cards sorted by position asc
	cardsColl, err := s.app.FindCollectionByNameOrId("kanban_cards")
	if err != nil {
		return fmt.Errorf("find kanban_cards collection: %w", err)
	}

	query := s.app.RecordQuery(cardsColl).
		AndWhere(dbx.NewExp("organization = {:orgID} AND column = {:colID} AND id != {:cardID}", dbx.Params{
			"orgID":  orgID,
			"colID":  toColumnID,
			"cardID": cardID,
		})).
		OrderBy("position asc").
		WithContext(ctx)

	targetRecords := make([]*core.Record, 0)
	if err := query.All(&targetRecords); err != nil {
		return fmt.Errorf("query target column cards: %w", err)
	}

	var newPos int

	if len(targetRecords) == 0 {
		newPos = 1000
	} else if toIndex <= 0 {
		newPos = targetRecords[0].GetInt("position") / 2
		if newPos < 1 {
			// fallback/reindex if positions collapsed
			newPos = 500
		}
	} else if toIndex >= len(targetRecords) {
		newPos = targetRecords[len(targetRecords)-1].GetInt("position") + 1000
	} else {
		prevPos := targetRecords[toIndex-1].GetInt("position")
		nextPos := targetRecords[toIndex].GetInt("position")
		newPos = (prevPos + nextPos) / 2

		if newPos == prevPos || newPos == nextPos {
			// Collapse detected. We need to re-index all cards in the column in increments of 1000
			newPos = (toIndex * 1000) + 500
			for idx, r := range targetRecords {
				curIdx := idx
				if idx >= toIndex {
					curIdx++
				}
				r.Set("position", curIdx*1000+1000)
				if err := s.app.SaveWithContext(ctx, r); err != nil {
					return fmt.Errorf("failed to reindex card %s: %w", r.Id, err)
				}
			}
		}
	}

	cardRecord.Set("column", toColumnID)
	cardRecord.Set("position", newPos)

	if err := s.app.SaveWithContext(ctx, cardRecord); err != nil {
		return fmt.Errorf("save moved card: %w", err)
	}

	return nil
}
