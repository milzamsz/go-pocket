package kanban

import (
	"context"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestKanbanService_CRUD_And_Move(t *testing.T) {
	app := testutil.NewTestApp(t)
	svc := New(app)
	ctx := context.Background()

	// Create test users first to satisfy organization owner constraint
	usersColl, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(usersColl)
	user.Set("email", "owner@example.com")
	user.Set("password", "test-password-123")
	user.Set("passwordConfirm", "test-password-123")
	require.NoError(t, app.Save(user))

	// Create target organization record first to satisfy required relation constraints
	orgColl, err := app.FindCollectionByNameOrId("organizations")
	require.NoError(t, err)

	orgA := core.NewRecord(orgColl)
	orgA.Set("id", "orga12345678901")
	orgA.Set("slug", "org-a")
	orgA.Set("name", "Org A")
	orgA.Set("owner", user.Id)
	require.NoError(t, app.Save(orgA))

	orgB := core.NewRecord(orgColl)
	orgB.Set("id", "orgb12345678901")
	orgB.Set("slug", "org-b")
	orgB.Set("name", "Org B")
	orgB.Set("owner", user.Id)
	require.NoError(t, app.Save(orgB))

	// 1. Create columns in orgA
	col1, err := svc.CreateColumn(ctx, orgA.Id, domain.KanbanColumn{Name: "Todo"})
	require.NoError(t, err)
	require.NotEmpty(t, col1.ID)

	col2, err := svc.CreateColumn(ctx, orgA.Id, domain.KanbanColumn{Name: "In Progress"})
	require.NoError(t, err)
	require.NotEmpty(t, col2.ID)

	// 2. Create cards in orgA
	card1, err := svc.CreateCard(ctx, orgA.Id, domain.KanbanCard{
		Title:    "Task 1",
		ColumnID: col1.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, card1.ID)

	card2, err := svc.CreateCard(ctx, orgA.Id, domain.KanbanCard{
		Title:    "Task 2",
		ColumnID: col1.ID,
	})
	require.NoError(t, err)

	card3, err := svc.CreateCard(ctx, orgA.Id, domain.KanbanCard{
		Title:    "Task 3",
		ColumnID: col2.ID,
	})
	require.NoError(t, err)

	// 3. Get Board
	cols, cards, err := svc.GetBoard(ctx, orgA.Id)
	require.NoError(t, err)
	require.Len(t, cols, 2)
	require.Len(t, cards, 3)

	// 4. Update Card
	card1.Description = "New details"
	updated, err := svc.UpdateCard(ctx, orgA.Id, card1)
	require.NoError(t, err)
	require.Equal(t, "New details", updated.Description)

	// 5. Move card (between columns and reorder)
	err = svc.MoveCard(ctx, orgA.Id, card2.ID, col2.ID, 0) // move Task 2 to the beginning of col2
	require.NoError(t, err)

	// Verify move
	_, cardsAfterMove, err := svc.GetBoard(ctx, orgA.Id)
	require.NoError(t, err)
	var foundMoved bool
	for _, c := range cardsAfterMove {
		if c.ID == card2.ID {
			require.Equal(t, col2.ID, c.ColumnID)
			foundMoved = true
		}
	}
	require.True(t, foundMoved)

	// 6. Delete Card
	err = svc.DeleteCard(ctx, orgA.Id, card3.ID)
	require.NoError(t, err)

	// Verify deletion
	_, cardsAfterDelete, err := svc.GetBoard(ctx, orgA.Id)
	require.NoError(t, err)
	require.Len(t, cardsAfterDelete, 2) // card3 deleted
}
