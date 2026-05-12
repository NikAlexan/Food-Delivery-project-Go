package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"food-delivery/order-service/internal/model"
)

// ── mock repo ─────────────────────────────────────────────────────────────────

type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) CreateOrderWithItems(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if o, ok := args.Get(0).(*model.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) ListByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus) (*model.Order, error) {
	args := m.Called(ctx, orderID, status)
	if o, ok := args.Get(0).(*model.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) CreatePayment(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockOrderRepo) GetPaymentByOrderID(ctx context.Context, orderID int64) (*model.Payment, error) {
	args := m.Called(ctx, orderID)
	if p, ok := args.Get(0).(*model.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOrderRepo) UpdatePaymentStatus(ctx context.Context, orderID int64, status model.PaymentStatus) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

// ── mock publisher ────────────────────────────────────────────────────────────

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, subject string, payload any) error {
	args := m.Called(ctx, subject, payload)
	return args.Error(0)
}

func (m *mockPublisher) Close() {}
