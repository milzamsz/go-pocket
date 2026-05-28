package products

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
	List(ctx context.Context, orgID string, q string, category string, active *bool, sort string, page, pageSize int) ([]domain.Product, int, error)
	Get(ctx context.Context, orgID string, id string) (domain.Product, error)
	Create(ctx context.Context, orgID string, product domain.Product) (domain.Product, error)
	Update(ctx context.Context, orgID string, product domain.Product) (domain.Product, error)
	Delete(ctx context.Context, orgID string, id string) error
	BulkDelete(ctx context.Context, orgID string, ids []string) error
}

type service struct {
	app core.App
}

func New(app core.App) Service {
	return &service{app: app}
}

func (s *service) List(ctx context.Context, orgID string, q string, category string, active *bool, sort string, page, pageSize int) ([]domain.Product, int, error) {
	collection, err := s.app.FindCollectionByNameOrId("products")
	if err != nil {
		return nil, 0, fmt.Errorf("find products collection: %w", err)
	}

	filter := "organization = {:orgID}"
	params := dbx.Params{"orgID": orgID}

	if q != "" {
		filter += " AND name LIKE {:q}"
		params["q"] = "%" + q + "%"
	}
	if category != "" {
		filter += " AND category = {:category}"
		params["category"] = category
	}
	if active != nil {
		filter += " AND active = {:active}"
		params["active"] = *active
	}

	query := s.app.RecordQuery(collection).
		AndWhere(dbx.NewExp(filter, params))

	// Get total count
	var total int
	countQuery := s.app.RecordQuery(collection).
		Select("count(*)").
		AndWhere(dbx.NewExp(filter, params))
	if err := countQuery.Row(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	// Sorting
	allowedSorts := map[string]string{
		"name":     "name asc",
		"-name":    "name desc",
		"price":    "price asc",
		"-price":   "price desc",
		"stock":    "stock asc",
		"-stock":   "stock desc",
		"created":  "created asc",
		"-created": "created desc",
	}
	dbSort, ok := allowedSorts[sort]
	if !ok {
		dbSort = "created desc"
	}
	query = query.OrderBy(dbSort)

	// Pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	query = query.Limit(int64(pageSize)).Offset(int64(offset))

	records := make([]*core.Record, 0)
	if err := query.WithContext(ctx).All(&records); err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}

	out := make([]domain.Product, 0, len(records))
	for _, record := range records {
		out = append(out, domain.Product{
			ID:             record.Id,
			Name:           record.GetString("name"),
			Category:       record.GetString("category"),
			Price:          int64(record.GetInt("price")),
			Stock:          record.GetInt("stock"),
			Active:         record.GetBool("active"),
			OrganizationID: record.GetString("organization"),
			CreatedAt:      record.GetDateTime("created").Time(),
			UpdatedAt:      record.GetDateTime("updated").Time(),
		})
	}

	return out, total, nil
}

func (s *service) Get(ctx context.Context, orgID string, id string) (domain.Product, error) {
	record, err := s.app.FindFirstRecordByFilter(
		"products",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": id, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, fmt.Errorf("find product: %w", err)
	}

	return domain.Product{
		ID:             record.Id,
		Name:           record.GetString("name"),
		Category:       record.GetString("category"),
		Price:          int64(record.GetInt("price")),
		Stock:          record.GetInt("stock"),
		Active:         record.GetBool("active"),
		OrganizationID: record.GetString("organization"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) Create(ctx context.Context, orgID string, product domain.Product) (domain.Product, error) {
	collection, err := s.app.FindCollectionByNameOrId("products")
	if err != nil {
		return domain.Product{}, fmt.Errorf("find products collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", product.Name)
	record.Set("category", product.Category)
	record.Set("price", product.Price)
	record.Set("stock", product.Stock)
	record.Set("active", product.Active)
	record.Set("organization", orgID)

	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return domain.Product{}, fmt.Errorf("create product: %w", err)
	}

	return domain.Product{
		ID:             record.Id,
		Name:           record.GetString("name"),
		Category:       record.GetString("category"),
		Price:          int64(record.GetInt("price")),
		Stock:          record.GetInt("stock"),
		Active:         record.GetBool("active"),
		OrganizationID: record.GetString("organization"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) Update(ctx context.Context, orgID string, product domain.Product) (domain.Product, error) {
	record, err := s.app.FindFirstRecordByFilter(
		"products",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": product.ID, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, fmt.Errorf("find product for update: %w", err)
	}

	record.Set("name", product.Name)
	record.Set("category", product.Category)
	record.Set("price", product.Price)
	record.Set("stock", product.Stock)
	record.Set("active", product.Active)

	if err := s.app.SaveWithContext(ctx, record); err != nil {
		return domain.Product{}, fmt.Errorf("update product: %w", err)
	}

	return domain.Product{
		ID:             record.Id,
		Name:           record.GetString("name"),
		Category:       record.GetString("category"),
		Price:          int64(record.GetInt("price")),
		Stock:          record.GetInt("stock"),
		Active:         record.GetBool("active"),
		OrganizationID: record.GetString("organization"),
		CreatedAt:      record.GetDateTime("created").Time(),
		UpdatedAt:      record.GetDateTime("updated").Time(),
	}, nil
}

func (s *service) Delete(ctx context.Context, orgID string, id string) error {
	record, err := s.app.FindFirstRecordByFilter(
		"products",
		"id = {:id} && organization = {:orgID}",
		dbx.Params{"id": id, "orgID": orgID},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("find product for delete: %w", err)
	}

	if err := s.app.DeleteWithContext(ctx, record); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}

func (s *service) BulkDelete(ctx context.Context, orgID string, ids []string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := s.Delete(ctx, orgID, id); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
	}
	return nil
}
