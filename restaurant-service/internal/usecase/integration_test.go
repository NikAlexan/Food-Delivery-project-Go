// Package usecase_test contains end-to-end tests that run against real
// infrastructure (PostgreSQL + Redis + NATS).  They are skipped automatically
// when the TEST_INTEGRATION environment variable is not set, so the regular
// unit-test suite stays fast.
package usecase_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
	"food-delivery/restaurant-service/internal/usecase"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("set TEST_INTEGRATION=1 to run integration tests")
	}
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/restaurantdb_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { db.Close() })
	return db
}

func setupUsecase(t *testing.T, db *sql.DB) usecase.RestaurantUsecase {
	t.Helper()
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	t.Cleanup(func() { rdb.Close() })

	repo := repository.NewPostgresRepo(db)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Pass nil JetStream – NATS publishing is best-effort and won't block tests.
	return usecase.NewRestaurantUsecase(repo, rdb, nil, logger)
}

// ── Restaurant CRUD ───────────────────────────────────────────────────────────

func TestIntegration_RestaurantCRUD(t *testing.T) {
	requireIntegration(t)
	db := setupDB(t)
	uc := setupUsecase(t, db)
	ctx := context.Background()

	// Ensure the category exists.
	_, err := db.ExecContext(ctx, `INSERT INTO categories (name) VALUES ('TestCat') ON CONFLICT (name) DO NOTHING`)
	require.NoError(t, err)

	var catID int64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM categories WHERE name='TestCat'`).Scan(&catID))

	// Create
	r := &model.Restaurant{
		Name:        "Integration Bistro",
		Description: "A test restaurant",
		CategoryID:  catID,
		Address:     "1 Integration Ave",
		Phone:       "000-0000",
	}
	created, err := uc.CreateRestaurant(ctx, r)
	require.NoError(t, err)
	assert.Positive(t, created.ID)
	assert.True(t, created.IsActive)

	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM restaurants WHERE id=$1`, created.ID)
	})

	// Get
	got, err := uc.GetRestaurant(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Integration Bistro", got.Name)

	// Update
	update := &model.Restaurant{
		ID:       created.ID,
		Name:     "Integration Bistro V2",
		Address:  "2 Integration Ave",
		IsActive: true,
	}
	updated, err := uc.UpdateRestaurant(ctx, update)
	require.NoError(t, err)
	assert.Equal(t, "Integration Bistro V2", updated.Name)

	// List
	list, total, err := uc.ListRestaurants(ctx, model.ListFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Positive(t, total)
	assert.NotEmpty(t, list)

	// Search
	results, _, err := uc.SearchRestaurants(ctx, model.SearchFilter{Query: "Integration", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// Delete
	require.NoError(t, uc.DeleteRestaurant(ctx, created.ID))
	_, err = uc.GetRestaurant(ctx, created.ID)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

// ── Menu CRUD ─────────────────────────────────────────────────────────────────

func TestIntegration_MenuCRUD(t *testing.T) {
	requireIntegration(t)
	db := setupDB(t)
	uc := setupUsecase(t, db)
	ctx := context.Background()

	// Seed a restaurant.
	var restaurantID int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO restaurants (name, address) VALUES ('Menu Test', '5 Menu Rd') RETURNING id`,
	).Scan(&restaurantID)
	require.NoError(t, err)
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM restaurants WHERE id=$1`, restaurantID) })

	// Create menu item.
	item := &model.MenuItem{
		RestaurantID: restaurantID,
		Name:         "Veggie Wrap",
		Price:        7.50,
		Category:     "Wraps",
	}
	created, err := uc.CreateMenuItem(ctx, item)
	require.NoError(t, err)
	assert.Positive(t, created.ID)
	assert.True(t, created.IsAvailable)

	// Get menu.
	menu, err := uc.GetMenu(ctx, restaurantID)
	require.NoError(t, err)
	assert.Len(t, menu, 1)
	assert.Equal(t, "Veggie Wrap", menu[0].Name)

	// Update menu item.
	update := &model.MenuItem{
		ID:           created.ID,
		RestaurantID: restaurantID,
		Name:         "Veggie Wrap Deluxe",
		Price:        9.00,
		Category:     "Wraps",
		IsAvailable:  true,
	}
	updated, err := uc.UpdateMenuItem(ctx, update)
	require.NoError(t, err)
	assert.Equal(t, "Veggie Wrap Deluxe", updated.Name)
	assert.Equal(t, 9.00, updated.Price)

	// Delete menu item.
	require.NoError(t, uc.DeleteMenuItem(ctx, created.ID))
	menu, err = uc.GetMenu(ctx, restaurantID)
	require.NoError(t, err)
	assert.Empty(t, menu)
}
