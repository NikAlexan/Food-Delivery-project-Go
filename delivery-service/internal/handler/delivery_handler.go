package handler

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"food-delivery/delivery-service/internal/model"
	"food-delivery/delivery-service/internal/repository"
	"food-delivery/delivery-service/internal/usecase"
	pb "food-delivery/delivery-service/proto/pb"
)

type DeliveryHandler struct {
	pb.UnimplementedDeliveryServiceServer
	uc usecase.DeliveryUsecase
}

func NewDeliveryHandler(uc usecase.DeliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{uc: uc}
}

func internal(ctx context.Context, op string, err error) error {
	log.Printf("[%s] internal error: %v", op, err)
	return status.Error(codes.Internal, "internal server error")
}

func authenticatedUserID(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing auth context")
	}

	values := md.Get("x-user-id")
	if len(values) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing auth context")
	}

	userID, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, status.Error(codes.Unauthenticated, "invalid auth context")
	}

	return userID, nil
}

func (h *DeliveryHandler) AssignDriver(ctx context.Context, req *pb.AssignRequest) (*pb.Delivery, error) {
	delivery, err := h.uc.AssignDriver(ctx, usecase.AssignInput{
		OrderID:         req.OrderId,
		UserID:          req.UserId,
		UserEmail:       req.UserEmail,
		DeliveryAddress: req.DeliveryAddress,
	})
	if errors.Is(err, usecase.ErrInvalidDelivery) {
		return nil, status.Error(codes.InvalidArgument, "invalid delivery request")
	}
	if errors.Is(err, repository.ErrNoAvailableDriver) {
		return nil, status.Error(codes.ResourceExhausted, "no available driver")
	}
	if err != nil {
		return nil, internal(ctx, "AssignDriver", err)
	}
	return toPBDelivery(delivery), nil
}

func (h *DeliveryHandler) GetDelivery(ctx context.Context, req *pb.DeliveryIdRequest) (*pb.Delivery, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	delivery, err := h.uc.GetDelivery(ctx, req.DeliveryId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "delivery not found")
	}
	if err != nil {
		return nil, internal(ctx, "GetDelivery", err)
	}
	if delivery.UserID != userID {
		driver, err := h.uc.GetMyDriver(ctx, userID)
		if err != nil || driver.ID != delivery.DriverID {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}
	}
	return toPBDelivery(delivery), nil
}

func (h *DeliveryHandler) UpdateDriverLocation(ctx context.Context, req *pb.LocationRequest) (*pb.Empty, error) {
	err := h.uc.UpdateDriverLocation(ctx, req.DriverId, req.Latitude, req.Longitude)
	if errors.Is(err, usecase.ErrInvalidCoordinates) {
		return nil, status.Error(codes.InvalidArgument, "invalid coordinates")
	}
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	if err != nil {
		return nil, internal(ctx, "UpdateDriverLocation", err)
	}
	return &pb.Empty{}, nil
}

func (h *DeliveryHandler) TrackDelivery(ctx context.Context, req *pb.DeliveryIdRequest) (*pb.DeliveryStatus, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	delivery, err := h.uc.TrackDelivery(ctx, req.DeliveryId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "delivery not found")
	}
	if err != nil {
		return nil, internal(ctx, "TrackDelivery", err)
	}
	if delivery.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return &pb.DeliveryStatus{
		DeliveryId:       delivery.ID,
		Status:           delivery.Status,
		DriverName:       delivery.DriverName,
		CurrentLatitude:  delivery.CurrentLatitude,
		CurrentLongitude: delivery.CurrentLongitude,
		UpdatedAt:        delivery.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *DeliveryHandler) CompleteDelivery(ctx context.Context, req *pb.DeliveryIdRequest) (*pb.Delivery, error) {
	delivery, err := h.uc.CompleteDelivery(ctx, req.DeliveryId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "delivery not found")
	}
	if err != nil {
		return nil, internal(ctx, "CompleteDelivery", err)
	}
	return toPBDelivery(delivery), nil
}

func (h *DeliveryHandler) ListDriverDeliveries(ctx context.Context, req *pb.DriverIdRequest) (*pb.DeliveryList, error) {
	deliveries, err := h.uc.ListDriverDeliveries(ctx, req.DriverId)
	if err != nil {
		return nil, internal(ctx, "ListDriverDeliveries", err)
	}
	return toPBList(deliveries), nil
}

func (h *DeliveryHandler) GetDeliveryHistory(ctx context.Context, req *pb.UserIdRequest) (*pb.DeliveryList, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	deliveries, err := h.uc.GetDeliveryHistory(ctx, req.UserId)
	if err != nil {
		return nil, internal(ctx, "GetDeliveryHistory", err)
	}
	return toPBList(deliveries), nil
}

func (h *DeliveryHandler) CancelDelivery(ctx context.Context, req *pb.DeliveryIdRequest) (*pb.Delivery, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	delivery, err := h.uc.GetDelivery(ctx, req.DeliveryId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "delivery not found")
	}
	if err != nil {
		return nil, internal(ctx, "CancelDelivery", err)
	}
	if delivery.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	delivery, err = h.uc.CancelDelivery(ctx, req.DeliveryId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "delivery not found")
	}
	if err != nil {
		return nil, internal(ctx, "CancelDelivery", err)
	}
	return toPBDelivery(delivery), nil
}

func (h *DeliveryHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.DriverProfile, error) {
	uid, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.Name == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "name and email are required")
	}
	driver, err := h.uc.RegisterDriver(ctx, uid, req.Name, req.Email, req.Phone)
	if errors.Is(err, usecase.ErrAlreadyRegistered) {
		return nil, status.Error(codes.AlreadyExists, "already registered as driver")
	}
	if err != nil {
		return nil, internal(ctx, "RegisterDriver", err)
	}
	return toPBDriverProfile(driver), nil
}

func (h *DeliveryHandler) GetMyDriver(ctx context.Context, _ *pb.Empty) (*pb.DriverProfile, error) {
	uid, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	driver, err := h.uc.GetMyDriver(ctx, uid)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "driver profile not found")
	}
	if err != nil {
		return nil, internal(ctx, "GetMyDriver", err)
	}
	return toPBDriverProfile(driver), nil
}

func toPBDriverProfile(d *model.Driver) *pb.DriverProfile {
	return &pb.DriverProfile{
		DriverId:    d.ID,
		Name:        d.Name,
		Email:       d.Email,
		Phone:       d.Phone,
		IsAvailable: d.IsAvailable,
	}
}

func toPBList(deliveries []model.Delivery) *pb.DeliveryList {
	items := make([]*pb.Delivery, 0, len(deliveries))
	for _, delivery := range deliveries {
		item := delivery
		items = append(items, toPBDelivery(&item))
	}
	return &pb.DeliveryList{Deliveries: items}
}

func toPBDelivery(delivery *model.Delivery) *pb.Delivery {
	var assignedAt string
	if delivery.AssignedAt != nil {
		assignedAt = delivery.AssignedAt.Format(time.RFC3339)
	}

	var completedAt string
	if delivery.CompletedAt != nil {
		completedAt = delivery.CompletedAt.Format(time.RFC3339)
	}

	return &pb.Delivery{
		DeliveryId:       delivery.ID,
		OrderId:          delivery.OrderID,
		UserId:           delivery.UserID,
		UserEmail:        delivery.UserEmail,
		DriverId:         delivery.DriverID,
		DriverName:       delivery.DriverName,
		DriverEmail:      delivery.DriverEmail,
		DriverPhone:      delivery.DriverPhone,
		Status:           delivery.Status,
		DeliveryAddress:  delivery.DeliveryAddress,
		CreatedAt:        delivery.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        delivery.UpdatedAt.Format(time.RFC3339),
		AssignedAt:       assignedAt,
		CompletedAt:      completedAt,
		CurrentLatitude:  delivery.CurrentLatitude,
		CurrentLongitude: delivery.CurrentLongitude,
	}
}
