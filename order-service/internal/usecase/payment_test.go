package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/order-service/internal/model"
	"food-delivery/order-service/internal/repository"
)

func TestProcessPayment_Success(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub)
	ctx := context.Background()

	existing := &model.Order{ID: 1, Status: model.StatusPending, Total: 3500}
	paid := &model.Order{ID: 1, Status: model.StatusPaid}

	repo.On("GetByID", ctx, int64(1)).Return(existing, nil)
	repo.On("CreatePayment", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)
	repo.On("UpdateStatus", ctx, int64(1), model.StatusPaid).Return(paid, nil)
	repo.On("UpdatePaymentStatus", ctx, int64(1), model.PaymentSuccess).Return(nil)
	pub.On("Publish", ctx, "order.paid", mock.Anything).Return(nil)

	payment, err := uc.ProcessPayment(ctx, 1, "card", 3500)

	assert.NoError(t, err)
	assert.Equal(t, model.PaymentSuccess, payment.Status)
	repo.AssertExpectations(t)
}

func TestProcessPayment_OrderNotFound(t *testing.T) {
	repo := &mockOrderRepo{}
	pub := &mockPublisher{}
	uc := NewOrderUsecase(repo, pub)
	ctx := context.Background()

	repo.On("GetByID", ctx, int64(99)).Return(nil, repository.ErrNotFound)

	_, err := uc.ProcessPayment(ctx, 99, "card", 1000)

	assert.ErrorIs(t, err, ErrOrderNotFound)
}
