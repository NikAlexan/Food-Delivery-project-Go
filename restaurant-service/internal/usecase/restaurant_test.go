package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
)

// newTestUsecase builds a usecase backed by a mock repo, nil NATS, and a
// Redis client pointing at an address that will always fail (cache errors are
// tolerated, so tests stay self-contained without a running Redis).
func newTestUsecase(t *testing.T, repo repository.RestaurantRepository) RestaurantUsecase {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6399"}) // intentionally unreachable
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return NewRestaurantUsecase(repo, rdb, nil, logger)
}

// ── CreateRestaurant ──────────────────────────────────────────────────────────

func TestCreateRestaurant_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	input := &model.Restaurant{Name: "Burger Palace", Address: "123 Main St", Phone: "555-1234", CategoryID: 1}

	mockRepo.On("CreateRestaurant", ctx, input).
		Run(func(args mock.Arguments) {
			r := args.Get(1).(*model.Restaurant)
			r.ID = 42
			r.IsActive = true
			r.CreatedAt = time.Now()
			r.UpdatedAt = time.Now()
		}).
		Return(nil)

	result, err := uc.CreateRestaurant(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), result.ID)
	assert.Equal(t, "Burger Palace", result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCreateRestaurant_EmptyName(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)

	_, err := uc.CreateRestaurant(context.Background(), &model.Restaurant{Address: "123 Main St"})

	assert.ErrorIs(t, err, ErrInvalidInput)
	mockRepo.AssertNotCalled(t, "CreateRestaurant")
}

func TestCreateRestaurant_RepoError(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()
	repoErr := errors.New("db error")

	input := &model.Restaurant{Name: "Pizza Hub", Address: "999 Oak Ave", CategoryID: 2}
	mockRepo.On("CreateRestaurant", ctx, input).Return(repoErr)

	_, err := uc.CreateRestaurant(ctx, input)

	assert.ErrorIs(t, err, repoErr)
	mockRepo.AssertExpectations(t)
}

// ── GetRestaurant ─────────────────────────────────────────────────────────────

func TestGetRestaurant_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	want := &model.Restaurant{ID: 1, Name: "Sushi World", IsActive: true}
	mockRepo.On("GetRestaurantByID", ctx, int64(1)).Return(want, nil)

	got, err := uc.GetRestaurant(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, want.Name, got.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetRestaurant_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	mockRepo.On("GetRestaurantByID", ctx, int64(99)).Return(nil, repository.ErrNotFound)

	_, err := uc.GetRestaurant(ctx, 99)

	assert.ErrorIs(t, err, ErrNotFound)
	mockRepo.AssertExpectations(t)
}

// ── UpdateRestaurant ──────────────────────────────────────────────────────────

func TestUpdateRestaurant_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	update := &model.Restaurant{ID: 5, Name: "New Name", Address: "456 Elm St", IsActive: true}
	updated := &model.Restaurant{ID: 5, Name: "New Name", Address: "456 Elm St", IsActive: true}

	mockRepo.On("UpdateRestaurant", ctx, update).Return(nil)
	mockRepo.On("GetRestaurantByID", ctx, int64(5)).Return(updated, nil)

	got, err := uc.UpdateRestaurant(ctx, update)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", got.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateRestaurant_EmptyName(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)

	_, err := uc.UpdateRestaurant(context.Background(), &model.Restaurant{ID: 1, Name: ""})

	assert.ErrorIs(t, err, ErrInvalidInput)
}

// ── DeleteRestaurant ──────────────────────────────────────────────────────────

func TestDeleteRestaurant_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	mockRepo.On("DeleteRestaurant", ctx, int64(7)).Return(nil)

	err := uc.DeleteRestaurant(ctx, 7)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteRestaurant_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	mockRepo.On("DeleteRestaurant", ctx, int64(99)).Return(repository.ErrNotFound)

	err := uc.DeleteRestaurant(ctx, 99)

	assert.ErrorIs(t, err, ErrNotFound)
	mockRepo.AssertExpectations(t)
}

// ── ListRestaurants ───────────────────────────────────────────────────────────

func TestListRestaurants_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	filter := model.ListFilter{Page: 1, PageSize: 10}
	list := []*model.Restaurant{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}
	mockRepo.On("ListRestaurants", ctx, filter).Return(list, 2, nil)

	got, total, err := uc.ListRestaurants(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

// ── SearchRestaurants ─────────────────────────────────────────────────────────

func TestSearchRestaurants_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	filter := model.SearchFilter{Query: "pizza", Page: 1, PageSize: 10}
	list := []*model.Restaurant{{ID: 3, Name: "Pizza Town"}}
	mockRepo.On("SearchRestaurants", ctx, filter).Return(list, 1, nil)

	got, total, err := uc.SearchRestaurants(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 1, total)
}

func TestSearchRestaurants_EmptyQuery(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)

	_, _, err := uc.SearchRestaurants(context.Background(), model.SearchFilter{Query: ""})

	assert.ErrorIs(t, err, ErrInvalidInput)
}
