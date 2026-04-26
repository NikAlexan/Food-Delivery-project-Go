package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

func TestGetProfile_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	expected := &model.User{ID: 1, Email: "u@example.com", Name: "Nikita"}
	repo.On("GetByID", ctx, int64(1)).Return(expected, nil)

	user, err := uc.GetProfile(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	repo.On("GetByID", ctx, int64(99)).Return(nil, repository.ErrNotFound)

	_, err := uc.GetProfile(ctx, 99)

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	existing := &model.User{ID: 1, Email: "u@example.com", Name: "Old", Phone: "0"}
	repo.On("GetByID", ctx, int64(1)).Return(existing, nil)
	repo.On("UpdateProfile", ctx, mock.AnythingOfType("*model.User")).Return(nil)

	updated, err := uc.UpdateProfile(ctx, 1, "New Name", "+77009998877")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "+77009998877", updated.Phone)
}

func TestDeleteUser_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	repo.On("DeleteUser", ctx, int64(1)).Return(nil)

	err := uc.DeleteUser(ctx, 1)

	assert.NoError(t, err)
}

func TestGetAddresses_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	addrs := []model.Address{
		{ID: 1, UserID: 1, Street: "Abay 1", City: "Almaty"},
		{ID: 2, UserID: 1, Street: "Dostyk 5", City: "Almaty"},
	}
	repo.On("GetAddresses", ctx, int64(1)).Return(addrs, nil)

	result, err := uc.GetAddresses(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestAddAddress_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	addr := &model.Address{UserID: 1, Street: "Abay 1", City: "Almaty", Zip: "050000", IsDefault: true}
	repo.On("AddAddress", ctx, addr).Return(nil)

	result, err := uc.AddAddress(ctx, addr)

	assert.NoError(t, err)
	assert.Equal(t, addr, result)
}