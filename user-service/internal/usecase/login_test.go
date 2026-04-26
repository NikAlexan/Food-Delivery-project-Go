package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

func newHashedUser(id int64, email, password string) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &model.User{ID: id, Email: email, PasswordHash: string(hash)}
}

func TestLogin_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	user := newHashedUser(1, "user@example.com", "correct")
	repo.On("GetByEmail", ctx, "user@example.com").Return(user, nil)
	repo.On("SaveRefreshToken", ctx, mockRefreshToken(1)).Return(nil)

	access, refresh, err := uc.Login(ctx, "user@example.com", "correct")

	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	user := newHashedUser(1, "user@example.com", "correct")
	repo.On("GetByEmail", ctx, "user@example.com").Return(user, nil)

	_, _, err := uc.Login(ctx, "user@example.com", "wrong")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret")
	ctx := context.Background()

	repo.On("GetByEmail", ctx, "ghost@example.com").Return(nil, repository.ErrNotFound)

	_, _, err := uc.Login(ctx, "ghost@example.com", "pass")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}