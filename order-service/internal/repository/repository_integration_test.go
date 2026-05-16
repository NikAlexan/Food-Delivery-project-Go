package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"food-delivery/order-service/internal/model"
	"food-delivery/order-service/internal/repository"
)

func TestOrderRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := startPostgres(t, ctx)
	require.NoError(t, applyMigrations(ctx, db))

	repo := repository.NewPostgresOrderRepo(db)

	t.Run("create order with items", func(t *testing.T) {
		order := &model.Order{
			UserID:       1,
			RestaurantID: 10,
			Status:       model.StatusPending,
			Total:        1500,
			Items: []model.OrderItem{
				{MenuItemID: 101, Name: "Burger", Quantity: 2, Price: 500},
				{MenuItemID: 102, Name: "Fries", Quantity: 1, Price: 500},
			},
		}
		require.NoError(t, repo.CreateOrderWithItems(ctx, order))
		assert.Positive(t, order.ID)
		assert.Equal(t, order.ID, order.Items[0].OrderID)
		assert.Equal(t, order.ID, order.Items[1].OrderID)
	})

	t.Run("get by id", func(t *testing.T) {
		order := makeOrder(1, 10)
		require.NoError(t, repo.CreateOrderWithItems(ctx, order))

		found, err := repo.GetByID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
		assert.Equal(t, model.StatusPending, found.Status)
		assert.Len(t, found.Items, 1)
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999999)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("list by user id", func(t *testing.T) {
		order1 := makeOrder(42, 10)
		order2 := makeOrder(42, 11)
		require.NoError(t, repo.CreateOrderWithItems(ctx, order1))
		require.NoError(t, repo.CreateOrderWithItems(ctx, order2))

		list, err := repo.ListByUserID(ctx, 42)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("update status", func(t *testing.T) {
		order := makeOrder(2, 10)
		require.NoError(t, repo.CreateOrderWithItems(ctx, order))

		updated, err := repo.UpdateStatus(ctx, order.ID, model.StatusPaid)
		require.NoError(t, err)
		assert.Equal(t, model.StatusPaid, updated.Status)
	})

	t.Run("payment full cycle", func(t *testing.T) {
		order := makeOrder(3, 10)
		require.NoError(t, repo.CreateOrderWithItems(ctx, order))

		payment := &model.Payment{
			OrderID:   order.ID,
			Status:    model.PaymentPending,
			Amount:    1000,
			Method:    "card",
			CreatedAt: time.Now(),
		}
		require.NoError(t, repo.CreatePayment(ctx, payment))
		assert.Positive(t, payment.ID)

		require.NoError(t, repo.UpdatePaymentStatus(ctx, order.ID, model.PaymentSuccess))

		found, err := repo.GetPaymentByOrderID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, model.PaymentSuccess, found.Status)
	})
}

func makeOrder(userID, restaurantID int64) *model.Order {
	return &model.Order{
		UserID:       userID,
		RestaurantID: restaurantID,
		Status:       model.StatusPending,
		Total:        1000,
		Items:        []model.OrderItem{{MenuItemID: 1, Name: "Item", Quantity: 1, Price: 1000}},
	}
}

func startPostgres(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("orderdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(2*time.Minute)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	return db
}

func applyMigrations(ctx context.Context, db *sql.DB) error {
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	entries, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	slices.Sort(entries)

	for _, path := range entries {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, extractUpSection(string(raw))); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func extractUpSection(content string) string {
	const upMarker = "-- +goose Up"
	const downMarker = "-- +goose Down"

	start := strings.Index(content, upMarker)
	if start == -1 {
		return content
	}
	content = content[start+len(upMarker):]
	if end := strings.Index(content, downMarker); end != -1 {
		content = content[:end]
	}
	return strings.TrimSpace(content)
}
