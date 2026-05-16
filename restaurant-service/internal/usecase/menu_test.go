package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
)

// ── CreateMenuItem ────────────────────────────────────────────────────────────

func TestCreateMenuItem_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	item := &model.MenuItem{RestaurantID: 1, Name: "Margherita", Price: 9.99, Category: "Pizza"}

	mockRepo.On("CreateMenuItem", ctx, item).
		Run(func(args mock.Arguments) {
			i := args.Get(1).(*model.MenuItem)
			i.ID = 10
			i.IsAvailable = true
			i.CreatedAt = time.Now()
			i.UpdatedAt = time.Now()
		}).
		Return(nil)

	result, err := uc.CreateMenuItem(ctx, item)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), result.ID)
	assert.True(t, result.IsAvailable)
	mockRepo.AssertExpectations(t)
}

func TestCreateMenuItem_EmptyName(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)

	_, err := uc.CreateMenuItem(context.Background(), &model.MenuItem{RestaurantID: 1, Price: 5.0})

	assert.ErrorIs(t, err, ErrInvalidInput)
	mockRepo.AssertNotCalled(t, "CreateMenuItem")
}

func TestCreateMenuItem_ZeroPrice(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)

	_, err := uc.CreateMenuItem(context.Background(), &model.MenuItem{RestaurantID: 1, Name: "Fries", Price: 0})

	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestCreateMenuItem_RestaurantNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	item := &model.MenuItem{RestaurantID: 999, Name: "Soda", Price: 2.50}
	mockRepo.On("CreateMenuItem", ctx, item).Return(repository.ErrNotFound)

	_, err := uc.CreateMenuItem(ctx, item)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ── UpdateMenuItem ────────────────────────────────────────────────────────────

func TestUpdateMenuItem_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	item := &model.MenuItem{ID: 5, RestaurantID: 1, Name: "Pepperoni", Price: 12.50, Category: "Pizza", IsAvailable: true}
	updated := &model.MenuItem{ID: 5, RestaurantID: 1, Name: "Pepperoni", Price: 12.50, IsAvailable: true}

	mockRepo.On("UpdateMenuItem", ctx, item).Return(nil)
	mockRepo.On("GetMenuItemByID", ctx, int64(5)).Return(updated, nil)

	got, err := uc.UpdateMenuItem(ctx, item)

	assert.NoError(t, err)
	assert.Equal(t, "Pepperoni", got.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateMenuItem_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	item := &model.MenuItem{ID: 99, RestaurantID: 1, Name: "Ghost", Price: 5.00}
	mockRepo.On("UpdateMenuItem", ctx, item).Return(repository.ErrNotFound)

	_, err := uc.UpdateMenuItem(ctx, item)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ── DeleteMenuItem ────────────────────────────────────────────────────────────

func TestDeleteMenuItem_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	existing := &model.MenuItem{ID: 3, RestaurantID: 1, Name: "Sprite"}
	mockRepo.On("GetMenuItemByID", ctx, int64(3)).Return(existing, nil)
	mockRepo.On("DeleteMenuItem", ctx, int64(3)).Return(nil)

	err := uc.DeleteMenuItem(ctx, 3)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteMenuItem_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	mockRepo.On("GetMenuItemByID", ctx, int64(77)).Return(nil, repository.ErrNotFound)

	err := uc.DeleteMenuItem(ctx, 77)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ── GetMenu ───────────────────────────────────────────────────────────────────

func TestGetMenu_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	items := []*model.MenuItem{
		{ID: 1, RestaurantID: 2, Name: "Burger", Price: 8.99},
		{ID: 2, RestaurantID: 2, Name: "Fries", Price: 2.99},
	}
	mockRepo.On("GetMenuByRestaurant", ctx, int64(2)).Return(items, nil)

	got, err := uc.GetMenu(ctx, 2)

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetMenu_EmptyMenu(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	mockRepo.On("GetMenuByRestaurant", ctx, int64(5)).Return([]*model.MenuItem{}, nil)

	got, err := uc.GetMenu(ctx, 5)

	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetMenu_RepoError(t *testing.T) {
	mockRepo := &MockRepository{}
	uc := newTestUsecase(t, mockRepo)
	ctx := context.Background()

	dbErr := errors.New("connection refused")
	mockRepo.On("GetMenuByRestaurant", ctx, int64(1)).Return(nil, dbErr)

	_, err := uc.GetMenu(ctx, 1)

	assert.ErrorIs(t, err, dbErr)
}
