package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
)

// MockRepository is a testify mock for RestaurantRepository.
type MockRepository struct {
	mock.Mock
}

var _ repository.RestaurantRepository = (*MockRepository)(nil)

func (m *MockRepository) CreateRestaurant(ctx context.Context, r *model.Restaurant) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRepository) GetRestaurantByID(ctx context.Context, id int64) (*model.Restaurant, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Restaurant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateRestaurant(ctx context.Context, r *model.Restaurant) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRepository) DeleteRestaurant(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListRestaurants(ctx context.Context, f model.ListFilter) ([]*model.Restaurant, int, error) {
	args := m.Called(ctx, f)
	if v := args.Get(0); v != nil {
		return v.([]*model.Restaurant), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockRepository) SearchRestaurants(ctx context.Context, f model.SearchFilter) ([]*model.Restaurant, int, error) {
	args := m.Called(ctx, f)
	if v := args.Get(0); v != nil {
		return v.([]*model.Restaurant), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockRepository) CreateMenuItem(ctx context.Context, item *model.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockRepository) GetMenuItemByID(ctx context.Context, id int64) (*model.MenuItem, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.MenuItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateMenuItem(ctx context.Context, item *model.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockRepository) DeleteMenuItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetMenuByRestaurant(ctx context.Context, restaurantID int64) ([]*model.MenuItem, error) {
	args := m.Called(ctx, restaurantID)
	if v := args.Get(0); v != nil {
		return v.([]*model.MenuItem), args.Error(1)
	}
	return nil, args.Error(1)
}
