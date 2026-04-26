package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"food-delivery/user-service/internal/model"
	"food-delivery/user-service/internal/repository"
	"food-delivery/user-service/internal/usecase"
	pb "food-delivery/user-service/proto/pb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	uc usecase.UserUsecase
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func internal(ctx context.Context, op string, err error) error {
	log.Printf("[%s] internal error: %v", op, err)
	return status.Error(codes.Internal, "internal server error")
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	var addr *model.Address
	if req.Address != nil {
		addr = &model.Address{
			Street:    req.Address.Street,
			City:      req.Address.City,
			Zip:       req.Address.Zip,
			IsDefault: req.Address.IsDefault,
		}
	}

	user, err := h.uc.Register(ctx, req.Email, req.Password, req.Name, req.Phone, addr)
	if errors.Is(err, usecase.ErrEmailTaken) {
		return nil, status.Error(codes.AlreadyExists, "email already taken")
	}
	if err != nil {
		return nil, internal(ctx, "RegisterUser", err)
	}

	return &pb.UserResponse{UserId: user.ID, Email: user.Email, Name: user.Name, Phone: user.Phone}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.TokenResponse, error) {
	access, refresh, err := h.uc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err != nil {
		return nil, internal(ctx, "LoginUser", err)
	}

	return &pb.TokenResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *pb.UserIdRequest) (*pb.UserProfile, error) {
	user, err := h.uc.GetProfile(ctx, req.UserId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, internal(ctx, "GetProfile", err)
	}

	return &pb.UserProfile{
		UserId:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserResponse, error) {
	user, err := h.uc.UpdateProfile(ctx, req.UserId, req.Name, req.Phone)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, internal(ctx, "UpdateProfile", err)
	}

	return &pb.UserResponse{UserId: user.ID, Email: user.Email, Name: user.Name, Phone: user.Phone}, nil
}

func (h *UserHandler) AddAddress(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
	addr := &model.Address{
		UserID:    req.UserId,
		Street:    req.Street,
		City:      req.City,
		Zip:       req.Zip,
		IsDefault: req.IsDefault,
	}

	result, err := h.uc.AddAddress(ctx, addr)
	if err != nil {
		return nil, internal(ctx, "AddAddress", err)
	}

	return &pb.AddressResponse{Address: &pb.Address{
		AddressId: result.ID,
		UserId:    result.UserID,
		Street:    result.Street,
		City:      result.City,
		Zip:       result.Zip,
		IsDefault: result.IsDefault,
	}}, nil
}

func (h *UserHandler) GetAddresses(ctx context.Context, req *pb.UserIdRequest) (*pb.AddressList, error) {
	addrs, err := h.uc.GetAddresses(ctx, req.UserId)
	if err != nil {
		return nil, internal(ctx, "GetAddresses", err)
	}

	var pbAddrs []*pb.Address
	for _, a := range addrs {
		pbAddrs = append(pbAddrs, &pb.Address{
			AddressId: a.ID,
			UserId:    a.UserID,
			Street:    a.Street,
			City:      a.City,
			Zip:       a.Zip,
			IsDefault: a.IsDefault,
		})
	}
	return &pb.AddressList{Addresses: pbAddrs}, nil
}

func (h *UserHandler) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.TokenResponse, error) {
	access, refresh, err := h.uc.RefreshToken(ctx, req.RefreshToken)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}
	if err != nil {
		return nil, internal(ctx, "RefreshToken", err)
	}

	return &pb.TokenResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.UserIdRequest) (*pb.Empty, error) {
	if err := h.uc.DeleteUser(ctx, req.UserId); err != nil {
		return nil, internal(ctx, "DeleteUser", err)
	}
	return &pb.Empty{}, nil
}
