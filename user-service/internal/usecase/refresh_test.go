package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

// mockRefreshToken returns a testify matcher that checks only UserID,
// since the token string is generated at runtime.
func mockRefreshToken(userID int64) interface{} {
	return mock.MatchedBy(func(rt *model.RefreshToken) bool {
		return rt.UserID == userID && rt.Token != "" && rt.ExpiresAt.After(time.Now())
	})
}

func TestRefreshToken_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	// First: login to get a real refresh token
	user := newHashedUser(2, "r@example.com", "pass")
	repo.On("GetByEmail", ctx, "r@example.com").Return(user, nil)
	repo.On("SaveRefreshToken", ctx, mockRefreshToken(2)).Return(nil).Once()

	_, oldRefresh, err := uc.Login(ctx, "r@example.com", "pass")
	assert.NoError(t, err)

	// Now refresh
	hashedOld := hashToken(oldRefresh)
	storedRT := &model.RefreshToken{
		UserID:    2,
		Token:     hashedOld,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	repo.On("GetRefreshToken", ctx, hashedOld).Return(storedRT, nil)
	repo.On("DeleteRefreshToken", ctx, hashedOld).Return(nil)
	repo.On("SaveRefreshToken", ctx, mockRefreshToken(2)).Return(nil).Once()

	newAccess, newRefresh, err := uc.RefreshToken(ctx, oldRefresh)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	repo.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	repo.On("GetRefreshToken", ctx, hashToken("bad-token")).Return(nil, repository.ErrNotFound)

	_, _, err := uc.RefreshToken(ctx, "bad-token")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRefreshToken_Expired(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUsecase(repo, "secret", nil, nil)
	ctx := context.Background()

	expired := &model.RefreshToken{
		UserID:    1,
		Token:     hashToken("expired-token"),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	repo.On("GetRefreshToken", ctx, hashToken("expired-token")).Return(expired, nil)

	_, _, err := uc.RefreshToken(ctx, "expired-token")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}