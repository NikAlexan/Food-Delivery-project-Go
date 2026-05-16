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

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

func TestRegisterUser_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := startPostgres(t, ctx)

	require.NoError(t, applyMigrations(ctx, db))

	repo := repository.NewPostgresUserRepo(db)

	t.Run("create user without address", func(t *testing.T) {
		user := &model.User{
			Email:        "alice@example.com",
			PasswordHash: "hashed",
			Name:         "Alice",
			Phone:        "+70000000001",
		}
		err := repo.CreateUserWithAddress(ctx, user, nil)
		require.NoError(t, err)
		assert.Positive(t, user.ID)

		found, err := repo.GetByEmail(ctx, "alice@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "Alice", found.Name)
	})

	t.Run("create user with address", func(t *testing.T) {
		user := &model.User{
			Email:        "bob@example.com",
			PasswordHash: "hashed",
			Name:         "Bob",
			Phone:        "+70000000002",
		}
		addr := &model.Address{
			Street:    "Lenin St 1",
			City:      "Almaty",
			Zip:       "050000",
			IsDefault: true,
		}
		err := repo.CreateUserWithAddress(ctx, user, addr)
		require.NoError(t, err)
		assert.Positive(t, user.ID)
		assert.Equal(t, user.ID, addr.UserID)

		addrs, err := repo.GetAddresses(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, addrs, 1)
		assert.Equal(t, "Lenin St 1", addrs[0].Street)
	})

	t.Run("get by id", func(t *testing.T) {
		user := &model.User{
			Email:        "carol@example.com",
			PasswordHash: "hashed",
			Name:         "Carol",
			Phone:        "+70000000003",
		}
		require.NoError(t, repo.CreateUserWithAddress(ctx, user, nil))

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "carol@example.com", found.Email)
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		user := &model.User{
			Email:        "dup@example.com",
			PasswordHash: "hashed",
			Name:         "Dup",
			Phone:        "+70000000004",
		}
		require.NoError(t, repo.CreateUserWithAddress(ctx, user, nil))

		user2 := &model.User{
			Email:        "dup@example.com",
			PasswordHash: "hashed2",
			Name:         "Dup2",
			Phone:        "+70000000005",
		}
		err := repo.CreateUserWithAddress(ctx, user2, nil)
		assert.Error(t, err)
	})
}

func startPostgres(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("userdb"),
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
		upSQL := extractUpSection(string(raw))
		if _, err := db.ExecContext(ctx, upSQL); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

// extractUpSection returns only the SQL between "-- +goose Up" and "-- +goose Down".
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
