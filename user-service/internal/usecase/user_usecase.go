package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already taken")
)

type UserUsecase interface {
	Register(ctx context.Context, email, password, name, phone string, addr *model.Address) (*model.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	UpdateProfile(ctx context.Context, userID int64, name, phone string) (*model.User, error)
	AddAddress(ctx context.Context, addr *model.Address) (*model.Address, error)
	GetAddresses(ctx context.Context, userID int64) ([]model.Address, error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
	DeleteUser(ctx context.Context, userID int64) error
}

type userUsecase struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewUserUsecase(repo repository.UserRepository, jwtSecret string) UserUsecase {
	return &userUsecase{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (u *userUsecase) Register(ctx context.Context, email, password, name, phone string, addr *model.Address) (*model.User, error) {
	existing, err := u.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{Email: email, PasswordHash: string(hash), Name: name, Phone: phone}
	if err := u.repo.CreateUserWithAddress(ctx, user, addr); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", "", ErrInvalidCredentials
	}
	if err != nil {
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	access, err := u.generateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refresh, err := u.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (u *userUsecase) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return u.repo.GetByID(ctx, userID)
}

func (u *userUsecase) UpdateProfile(ctx context.Context, userID int64, name, phone string) (*model.User, error) {
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Phone = phone
	if err := u.repo.UpdateProfile(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) AddAddress(ctx context.Context, addr *model.Address) (*model.Address, error) {
	if err := u.repo.AddAddress(ctx, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

func (u *userUsecase) GetAddresses(ctx context.Context, userID int64) ([]model.Address, error) {
	return u.repo.GetAddresses(ctx, userID)
}

func (u *userUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	hashed := hashToken(refreshToken)
	rt, err := u.repo.GetRefreshToken(ctx, hashed)
	if errors.Is(err, repository.ErrNotFound) {
		return "", "", ErrInvalidCredentials
	}
	if err != nil {
		return "", "", err
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", ErrInvalidCredentials
	}

	if err := u.repo.DeleteRefreshToken(ctx, hashed); err != nil {
		return "", "", err
	}

	access, err := u.generateAccessToken(rt.UserID)
	if err != nil {
		return "", "", err
	}
	newRefresh, err := u.generateRefreshToken(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}
	return access, newRefresh, nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, userID int64) error {
	return u.repo.DeleteUser(ctx, userID)
}

func (u *userUsecase) generateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"type": "access",
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.jwtSecret)
}

func (u *userUsecase) generateRefreshToken(ctx context.Context, userID int64) (string, error) {
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
		"jti":  jti,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.jwtSecret)
	if err != nil {
		return "", err
	}

	rt := &model.RefreshToken{
		UserID:    userID,
		Token:     hashToken(token),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := u.repo.SaveRefreshToken(ctx, rt); err != nil {
		return "", err
	}
	return token, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
