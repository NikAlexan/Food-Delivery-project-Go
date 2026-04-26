package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"food-delivery/user-service/internal/model"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUserWithAddress(ctx context.Context, user *model.User, addr *model.Address) error {
	args := m.Called(ctx, user, addr)
	return args.Error(0)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*model.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) AddAddress(ctx context.Context, addr *model.Address) error {
	args := m.Called(ctx, addr)
	return args.Error(0)
}

func (m *mockUserRepo) GetAddresses(ctx context.Context, userID int64) ([]model.Address, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Address), args.Error(1)
}

func (m *mockUserRepo) SaveRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	args := m.Called(ctx, rt)
	return args.Error(0)
}

func (m *mockUserRepo) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	args := m.Called(ctx, token)
	if rt, ok := args.Get(0).(*model.RefreshToken); ok {
		return rt, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}