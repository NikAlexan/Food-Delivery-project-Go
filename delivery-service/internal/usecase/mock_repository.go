package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"food-delivery/delivery-service/internal/model"
)

type mockDeliveryRepo struct {
	mock.Mock
}

func (m *mockDeliveryRepo) AssignDriver(ctx context.Context, orderID, userID int64, userEmail, address string) (*model.Delivery, bool, error) {
	args := m.Called(ctx, orderID, userID, userEmail, address)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *mockDeliveryRepo) GetByID(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	args := m.Called(ctx, deliveryID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockDeliveryRepo) GetByOrderID(ctx context.Context, orderID int64) (*model.Delivery, error) {
	args := m.Called(ctx, orderID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockDeliveryRepo) UpdateDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error {
	args := m.Called(ctx, driverID, latitude, longitude)
	return args.Error(0)
}

func (m *mockDeliveryRepo) MarkInTransitByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error) {
	args := m.Called(ctx, orderID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *mockDeliveryRepo) CompleteDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error) {
	args := m.Called(ctx, deliveryID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *mockDeliveryRepo) CancelDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error) {
	args := m.Called(ctx, deliveryID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *mockDeliveryRepo) CancelByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error) {
	args := m.Called(ctx, orderID)
	if delivery, ok := args.Get(0).(*model.Delivery); ok {
		return delivery, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *mockDeliveryRepo) ListDriverDeliveries(ctx context.Context, driverID int64) ([]model.Delivery, error) {
	args := m.Called(ctx, driverID)
	if deliveries, ok := args.Get(0).([]model.Delivery); ok {
		return deliveries, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockDeliveryRepo) GetDeliveryHistory(ctx context.Context, userID int64) ([]model.Delivery, error) {
	args := m.Called(ctx, userID)
	if deliveries, ok := args.Get(0).([]model.Delivery); ok {
		return deliveries, args.Error(1)
	}
	return nil, args.Error(1)
}
