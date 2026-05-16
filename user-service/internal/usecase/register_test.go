package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	repo.On("GetByEmail", ctx, "test@example.com").Return(nil, repository.ErrNotFound)
	repo.On("CreateUserWithAddress", ctx, mock.AnythingOfType("*model.User"), mock.AnythingOfType("*model.Address")).
		Return(nil)

	user, err := uc.Register(ctx, "test@example.com", "password123", "Nikita", "+77001234567",
		&model.Address{Street: "Abay 1", City: "Almaty", Zip: "050000"})

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Nikita", user.Name)
	repo.AssertExpectations(t)
}

func TestRegister_EmailTaken(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	existing := &model.User{ID: 1, Email: "test@example.com"}
	repo.On("GetByEmail", ctx, "test@example.com").Return(existing, nil)

	_, err := uc.Register(ctx, "test@example.com", "password123", "Nikita", "", nil)

	assert.ErrorIs(t, err, ErrEmailTaken)
	repo.AssertExpectations(t)
}

func TestRegister_NoAddress(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	repo.On("GetByEmail", ctx, "noaddr@example.com").Return(nil, repository.ErrNotFound)
	repo.On("CreateUserWithAddress", ctx, mock.AnythingOfType("*model.User"), (*model.Address)(nil)).
		Return(nil)

	user, err := uc.Register(ctx, "noaddr@example.com", "pass", "Test", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, "noaddr@example.com", user.Email)
	repo.AssertExpectations(t)
}