package products

import (
	"context"
	"errors"
	"testing"

	"github.com/milzamsz/go-pocket/internal/domain"
	"github.com/milzamsz/go-pocket/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestProductService_CRUD_And_Tenancy(t *testing.T) {
	app := testutil.NewTestApp(t)
	svc := New(app)
	ctx := context.Background()

	// Create test users first to satisfy organization owner constraint
	usersColl, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	userA := core.NewRecord(usersColl)
	userA.Set("email", "owner-a@example.com")
	userA.Set("password", "test-password-123")
	userA.Set("passwordConfirm", "test-password-123")
	require.NoError(t, app.Save(userA))

	userB := core.NewRecord(usersColl)
	userB.Set("email", "owner-b@example.com")
	userB.Set("password", "test-password-123")
	userB.Set("passwordConfirm", "test-password-123")
	require.NoError(t, app.Save(userB))

	// Create target organization records first to satisfy required relation constraints
	orgColl, err := app.FindCollectionByNameOrId("organizations")
	require.NoError(t, err)

	orgA := core.NewRecord(orgColl)
	orgA.Set("id", "orga12345678901")
	orgA.Set("slug", "org-a")
	orgA.Set("name", "Org A")
	orgA.Set("owner", userA.Id)
	require.NoError(t, app.Save(orgA))

	orgB := core.NewRecord(orgColl)
	orgB.Set("id", "orgb12345678901")
	orgB.Set("slug", "org-b")
	orgB.Set("name", "Org B")
	orgB.Set("owner", userB.Id)
	require.NoError(t, app.Save(orgB))

	// Create some products for orgA
	p1, err := svc.Create(ctx, orgA.Id, domain.Product{
		Name:     "Awesome Widget",
		Category: "electronics",
		Price:    12900,
		Stock:    50,
		Active:   true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, p1.ID)
	require.Equal(t, orgA.Id, p1.OrganizationID)

	p2, err := svc.Create(ctx, orgA.Id, domain.Product{
		Name:     "Comfy Hoodie",
		Category: "apparel",
		Price:    5900,
		Stock:    100,
		Active:   true,
	})
	require.NoError(t, err)

	// Create a product for orgB to test isolation
	p3, err := svc.Create(ctx, orgB.Id, domain.Product{
		Name:     "Secret Gadget",
		Category: "electronics",
		Price:    45000,
		Stock:    5,
		Active:   true,
	})
	require.NoError(t, err)

	// List products for orgA
	listA, totalA, err := svc.List(ctx, orgA.Id, "", "", nil, "", 1, 10)
	require.NoError(t, err)
	require.Equal(t, 2, totalA)
	require.Len(t, listA, 2)

	// List products with query search filter
	listSearch, totalSearch, err := svc.List(ctx, orgA.Id, "Awesome", "", nil, "", 1, 10)
	require.NoError(t, err)
	require.Equal(t, 1, totalSearch)
	require.Len(t, listSearch, 1)
	require.Equal(t, p1.ID, listSearch[0].ID)

	// List products for orgB (should only see p3)
	listB, totalB, err := svc.List(ctx, orgB.Id, "", "", nil, "", 1, 10)
	require.NoError(t, err)
	require.Equal(t, 1, totalB)
	require.Len(t, listB, 1)
	require.Equal(t, p3.ID, listB[0].ID)

	// Get product
	fetched, err := svc.Get(ctx, orgA.Id, p1.ID)
	require.NoError(t, err)
	require.Equal(t, p1.Name, fetched.Name)

	// Attempt to get orgA's product from orgB (should fail/return not found)
	_, err = svc.Get(ctx, orgB.Id, p1.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	// Update product
	p1.Name = "Awesome Widget Pro"
	p1.Price = 14900
	updated, err := svc.Update(ctx, orgA.Id, p1)
	require.NoError(t, err)
	require.Equal(t, "Awesome Widget Pro", updated.Name)
	require.Equal(t, int64(14900), updated.Price)

	// Attempt to update orgA's product using orgB context (should fail/not found)
	_, err = svc.Update(ctx, orgB.Id, p1)
	require.Error(t, err)

	// Delete product
	err = svc.Delete(ctx, orgA.Id, p1.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = svc.Get(ctx, orgA.Id, p1.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))

	// Bulk delete
	err = svc.BulkDelete(ctx, orgA.Id, []string{p2.ID})
	require.NoError(t, err)

	// Verify deletion
	_, err = svc.Get(ctx, orgA.Id, p2.ID)
	require.Error(t, err)
}
