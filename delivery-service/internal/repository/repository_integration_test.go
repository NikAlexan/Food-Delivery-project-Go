package repository_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	natsgo "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"food-delivery/delivery-service/internal/cache"
	"food-delivery/delivery-service/internal/email"
	"food-delivery/delivery-service/internal/model"
	natsclient "food-delivery/delivery-service/internal/nats"
	"food-delivery/delivery-service/internal/repository"
	"food-delivery/delivery-service/internal/usecase"
)

func TestDeliveryFlow_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dsn := startPostgres(t, ctx)
	redisAddr := startRedis(t, ctx)
	natsAddr := startNATS(t, ctx)
	mailhogSMTPAddr, mailhogAPIURL := startMailHog(t, ctx)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, applyMigrations(ctx, db))

	locationCache, err := cache.NewRedisLocationCache(redisAddr, 5*time.Minute)
	require.NoError(t, err)
	t.Cleanup(func() { _ = locationCache.Close() })

	smtpHost, smtpPort := splitHostPort(t, mailhogSMTPAddr)
	mailer, err := email.NewSMTPClient(smtpHost, "", "", "delivery@example.com", smtpPort)
	require.NoError(t, err)

	natsConn, err := natsclient.NewClient(natsAddr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = natsConn.Close() })

	repo := repository.NewPostgresDeliveryRepo(db)
	uc := usecase.NewDeliveryUsecase(repo, mailer, natsConn, locationCache)
	require.NoError(t, natsConn.StartOrderConsumers(uc))

	seedDriver(t, ctx, db)

	nc, err := natsgo.Connect(natsAddr)
	require.NoError(t, err)
	t.Cleanup(nc.Close)

	js, err := nc.JetStream()
	require.NoError(t, err)

	completedSub, err := nc.SubscribeSync("delivery.completed")
	require.NoError(t, err)
	require.NoError(t, nc.Flush())

	orderEvent := model.OrderEvent{
		OrderID:         1001,
		UserID:          42,
		UserEmail:       "user@example.com",
		DeliveryAddress: "Abay 10, Almaty",
	}

	_, err = js.Publish("order.created", mustJSON(t, orderEvent))
	require.NoError(t, err)

	delivery := waitForDeliveryStatus(t, ctx, repo, 1001, model.StatusAssigned)
	assert.Equal(t, int64(42), delivery.UserID)
	assert.Equal(t, "user@example.com", delivery.UserEmail)

	waitForMailHogMessage(t, mailhogAPIURL, 1, "Order #1001 confirmed")

	require.NoError(t, uc.UpdateDriverLocation(ctx, delivery.DriverID, 43.238949, 76.889709))
	_, err = db.ExecContext(ctx, `
		UPDATE drivers
		SET current_latitude = 0, current_longitude = 0
		WHERE id = $1`,
		delivery.DriverID,
	)
	require.NoError(t, err)

	tracked, err := uc.TrackDelivery(ctx, delivery.ID)
	require.NoError(t, err)
	assert.Equal(t, 43.238949, tracked.CurrentLatitude)
	assert.Equal(t, 76.889709, tracked.CurrentLongitude)

	_, err = js.Publish("order.paid", mustJSON(t, orderEvent))
	require.NoError(t, err)

	delivery = waitForDeliveryStatus(t, ctx, repo, 1001, model.StatusInTransit)
	assert.Equal(t, model.StatusInTransit, delivery.Status)
	waitForMailHogMessage(t, mailhogAPIURL, 2, "Order #1001 is on the way")

	completed, err := uc.CompleteDelivery(ctx, delivery.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusCompleted, completed.Status)

	msg, err := completedSub.NextMsg(5 * time.Second)
	require.NoError(t, err)

	var completedEvent model.DeliveryCompletedEvent
	require.NoError(t, json.Unmarshal(msg.Data, &completedEvent))
	assert.Equal(t, delivery.ID, completedEvent.DeliveryID)
	assert.Equal(t, delivery.OrderID, completedEvent.OrderID)

	waitForMailHogMessage(t, mailhogAPIURL, 3, "Order #1001 delivered")

	_, err = uc.CompleteDelivery(ctx, delivery.ID)
	require.NoError(t, err)

	_, err = completedSub.NextMsg(1 * time.Second)
	assert.Error(t, err)

	raw, total := fetchMailHogSnapshot(t, mailhogAPIURL)
	assert.Equal(t, 3, total)
	assert.Contains(t, raw, "Order #1001 delivered")
}

func startPostgres(t *testing.T, ctx context.Context) string {
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
	return dsn
}

func startRedis(t *testing.T, ctx context.Context) string {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:8.4-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("%s:%s", host, port.Port())
}

func startNATS(t *testing.T, ctx context.Context) string {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2.12-alpine",
			ExposedPorts: []string{"4222/tcp"},
			Cmd:          []string{"-js"},
			WaitingFor:   wait.ForListeningPort("4222/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "4222/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("nats://%s:%s", host, port.Port())
}

func startMailHog(t *testing.T, ctx context.Context) (string, string) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mailhog/mailhog:v1.0.1",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForListeningPort("1025/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)

	smtpPort, err := container.MappedPort(ctx, "1025/tcp")
	require.NoError(t, err)
	httpPort, err := container.MappedPort(ctx, "8025/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("%s:%s", host, smtpPort.Port()), fmt.Sprintf("http://%s:%s", host, httpPort.Port())
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
		query, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}

	return nil
}

func seedDriver(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO drivers (name, email, phone, is_available)
		VALUES ('Courier One', 'driver@example.com', '+77010000000', TRUE)`,
	)
	require.NoError(t, err)
}

func waitForDeliveryStatus(t *testing.T, ctx context.Context, repo repository.DeliveryRepository, orderID int64, expected string) *model.Delivery {
	t.Helper()

	var delivery *model.Delivery
	require.Eventually(t, func() bool {
		var err error
		delivery, err = repo.GetByOrderID(ctx, orderID)
		return err == nil && delivery.Status == expected
	}, 20*time.Second, 250*time.Millisecond)

	return delivery
}

func waitForMailHogMessage(t *testing.T, apiURL string, expectedCount int, expectedSubject string) {
	t.Helper()

	require.Eventually(t, func() bool {
		raw, total := fetchMailHogSnapshot(t, apiURL)
		return total >= expectedCount && strings.Contains(raw, expectedSubject)
	}, 20*time.Second, 250*time.Millisecond)
}

func fetchMailHogSnapshot(t *testing.T, apiURL string) (string, int) {
	t.Helper()

	resp, err := http.Get(apiURL + "/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload struct {
		Total int `json:"total"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	return string(body), payload.Total
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	payload, err := json.Marshal(value)
	require.NoError(t, err)
	return payload
}

func splitHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, port, found := strings.Cut(addr, ":")
	require.True(t, found)

	parsed, err := strconv.Atoi(port)
	require.NoError(t, err)
	return host, parsed
}
