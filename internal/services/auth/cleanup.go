package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// CleanupExpiredOrConsumedTokens removes auth_tokens records that are no longer usable.
func CleanupExpiredOrConsumedTokens(ctx context.Context, app core.App, now time.Time) (int, error) {
	collection, err := app.FindCollectionByNameOrId("auth_tokens")
	if err != nil {
		return 0, fmt.Errorf("find auth_tokens collection: %w", err)
	}

	records := make([]*core.Record, 0)
	query := app.RecordQuery(collection).
		AndWhere(dbx.NewExp("consumed_at != '' OR expires_at <= {:now}", dbx.Params{"now": now.UTC()})).
		WithContext(ctx)
	if err := query.All(&records); err != nil {
		return 0, fmt.Errorf("find stale auth token records: %w", err)
	}

	deleted := 0
	for _, record := range records {
		if err := app.Delete(record); err != nil {
			return deleted, fmt.Errorf("delete stale auth token record: %w", err)
		}
		deleted++
	}

	return deleted, nil
}
