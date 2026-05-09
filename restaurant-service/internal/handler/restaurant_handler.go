package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
	"food-delivery/restaurant-service/internal/usecase"
	pb "food-delivery/restaurant-service/proto/pb"
)

// RestaurantHandler implements the gRPC RestaurantServiceServer interface.
type RestaurantHandler struct {
	pb.UnimplementedRestaurantServiceServer
	uc usecase.RestaurantUsecase
}

// NewRestaurantHandler creates a new handler with the given usecase.
func NewRestaurantHandler(uc usecase.RestaurantUsecase) *RestaurantHandler {
	return &RestaurantHandler{uc: uc}
}

func internalErr(ctx context.Context, op string, err error) error {
	log.Printf("[%s] internal error: %v", op, err)
	return status.Error(codes.Internal, "internal server error")
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, usecase.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// ── Restaurant RPCs ───────────────────────────────────────────────────────────

func (h *RestaurantHandler) CreateRestaurant(ctx context.Context, req *pb.CreateRestaurantRequest) (*pb.RestaurantResponse, error) {
	r := &model.Restaurant{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryId,
		Address:     req.Address,
		Phone:       req.Phone,
		ImageURL:    req.ImageUrl,
	}
	result, err := h.uc.CreateRestaurant(ctx, r)
	if err != nil {
		return nil, mapErr(err)
	}
	return toRestaurantProto(result), nil
}

func (h *RestaurantHandler) GetRestaurant(ctx context.Context, req *pb.RestaurantIdRequest) (*pb.RestaurantResponse, error) {
	r, err := h.uc.GetRestaurant(ctx, req.RestaurantId)
	if err != nil {
		return nil, mapErr(err)
	}
	return toRestaurantProto(r), nil
}

func (h *RestaurantHandler) ListRestaurants(ctx context.Context, req *pb.ListRestaurantsRequest) (*pb.RestaurantList, error) {
	f := model.ListFilter{
		CategoryID: req.CategoryId,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	list, total, err := h.uc.ListRestaurants(ctx, f)
	if err != nil {
		return nil, internalErr(ctx, "ListRestaurants", err)
	}
	return toRestaurantListProto(list, total), nil
}

func (h *RestaurantHandler) SearchRestaurants(ctx context.Context, req *pb.SearchRequest) (*pb.RestaurantList, error) {
	f := model.SearchFilter{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	list, total, err := h.uc.SearchRestaurants(ctx, f)
	if err != nil {
		return nil, mapErr(err)
	}
	return toRestaurantListProto(list, total), nil
}

func (h *RestaurantHandler) UpdateRestaurant(ctx context.Context, req *pb.UpdateRestaurantRequest) (*pb.RestaurantResponse, error) {
	r := &model.Restaurant{
		ID:          req.RestaurantId,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		ImageURL:    req.ImageUrl,
		IsActive:    req.IsActive,
	}
	result, err := h.uc.UpdateRestaurant(ctx, r)
	if err != nil {
		return nil, mapErr(err)
	}
	return toRestaurantProto(result), nil
}

func (h *RestaurantHandler) DeleteRestaurant(ctx context.Context, req *pb.RestaurantIdRequest) (*pb.Empty, error) {
	if err := h.uc.DeleteRestaurant(ctx, req.RestaurantId); err != nil {
		return nil, mapErr(err)
	}
	return &pb.Empty{}, nil
}

// ── Menu RPCs ─────────────────────────────────────────────────────────────────

func (h *RestaurantHandler) CreateMenuItem(ctx context.Context, req *pb.CreateMenuItemRequest) (*pb.MenuItemResponse, error) {
	item := &model.MenuItem{
		RestaurantID: req.RestaurantId,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		Category:     req.Category,
		ImageURL:     req.ImageUrl,
	}
	result, err := h.uc.CreateMenuItem(ctx, item)
	if err != nil {
		return nil, mapErr(err)
	}
	return toMenuItemProto(result), nil
}

func (h *RestaurantHandler) UpdateMenuItem(ctx context.Context, req *pb.UpdateMenuItemRequest) (*pb.MenuItemResponse, error) {
	item := &model.MenuItem{
		ID:           req.ItemId,
		RestaurantID: req.RestaurantId,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		Category:     req.Category,
		ImageURL:     req.ImageUrl,
		IsAvailable:  req.IsAvailable,
	}
	result, err := h.uc.UpdateMenuItem(ctx, item)
	if err != nil {
		return nil, mapErr(err)
	}
	return toMenuItemProto(result), nil
}

func (h *RestaurantHandler) DeleteMenuItem(ctx context.Context, req *pb.MenuItemIdRequest) (*pb.Empty, error) {
	if err := h.uc.DeleteMenuItem(ctx, req.ItemId); err != nil {
		return nil, mapErr(err)
	}
	return &pb.Empty{}, nil
}

func (h *RestaurantHandler) GetMenu(ctx context.Context, req *pb.RestaurantIdRequest) (*pb.MenuList, error) {
	items, err := h.uc.GetMenu(ctx, req.RestaurantId)
	if err != nil {
		return nil, internalErr(ctx, "GetMenu", err)
	}
	protoItems := make([]*pb.MenuItemResponse, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, toMenuItemProto(item))
	}
	return &pb.MenuList{Items: protoItems}, nil
}

// ── Converters ────────────────────────────────────────────────────────────────

func toRestaurantProto(r *model.Restaurant) *pb.RestaurantResponse {
	return &pb.RestaurantResponse{
		RestaurantId: r.ID,
		Name:         r.Name,
		Description:  r.Description,
		CategoryId:   r.CategoryID,
		CategoryName: r.CategoryName,
		Address:      r.Address,
		Phone:        r.Phone,
		ImageUrl:     r.ImageURL,
		IsActive:     r.IsActive,
		Rating:       r.Rating,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
	}
}

func toRestaurantListProto(list []*model.Restaurant, total int) *pb.RestaurantList {
	items := make([]*pb.RestaurantResponse, 0, len(list))
	for _, r := range list {
		items = append(items, toRestaurantProto(r))
	}
	return &pb.RestaurantList{Restaurants: items, Total: int32(total)}
}

func toMenuItemProto(item *model.MenuItem) *pb.MenuItemResponse {
	return &pb.MenuItemResponse{
		ItemId:       item.ID,
		RestaurantId: item.RestaurantID,
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		Category:     item.Category,
		ImageUrl:     item.ImageURL,
		IsAvailable:  item.IsAvailable,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339),
	}
}
