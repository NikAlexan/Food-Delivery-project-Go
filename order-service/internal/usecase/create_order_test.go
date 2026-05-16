package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/order-service/internal/model"
	"food-delivery/order-service/internal/repository"
)

func TestCreateOrder_Success(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	items := []model.OrderItem{
		{MenuItemID: 1, Name: "Burger", Quantity: 2, Price: 1500},
	}

	repo.On("CreateOrderWithItems", ctx, mock.AnythingOfType("*model.Order")).Return(nil)
	pub.On("Publish", ctx, "order.created", mock.Anything).Return(nil)

	order, err := uc.CreateOrder(ctx, 1, 10, items)

	assert.NoError(t, err)
	assert.Equal(t, model.StatusPending, order.Status)
	assert.Equal(t, 3500.0, order.Total) // 2*1500 + 500 delivery
	repo.AssertExpectations(t)
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	_, err := uc.CreateOrder(ctx, 1, 10, nil)

	assert.ErrorIs(t, err, ErrEmptyItems)
}

func TestCancelOrder_Success(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	existing := &model.Order{ID: 1, Status: model.StatusPending}
	cancelled := &model.Order{ID: 1, Status: model.StatusCancelled}

	repo.On("GetByID", ctx, int64(1)).Return(existing, nil)
	repo.On("UpdateStatus", ctx, int64(1), model.StatusCancelled).Return(cancelled, nil)
	pub.On("Publish", ctx, "order.cancelled", mock.Anything).Return(nil)

	order, err := uc.CancelOrder(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, model.StatusCancelled, order.Status)
	repo.AssertExpectations(t)
}

func TestCancelOrder_NotFound(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	repo.On("GetByID", ctx, int64(99)).Return(nil, repository.ErrNotFound)

	_, err := uc.CancelOrder(ctx, 99)

	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	existing := &model.Order{ID: 1, Status: model.StatusCancelled}
	repo.On("GetByID", ctx, int64(1)).Return(existing, nil)

	_, err := uc.CancelOrder(ctx, 1)

	assert.ErrorIs(t, err, ErrAlreadyCancelled)
}

func TestCalculateTotal(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub, nil)
	ctx := context.Background()

	items := []model.OrderItem{
		{Quantity: 1, Price: 2000},
		{Quantity: 2, Price: 500},
	}

	subtotal, fee, total := uc.CalculateTotal(ctx, items)

	assert.Equal(t, 3000.0, subtotal)
	assert.Equal(t, 500.0, fee)
	assert.Equal(t, 3500.0, total)
}
