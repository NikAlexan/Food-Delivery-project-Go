package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"food-delivery/order-service/internal/model"
	"food-delivery/order-service/internal/usecase"
	pb "food-delivery/order-service/proto/pb"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	uc usecase.OrderUsecase
}

func NewOrderHandler(uc usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func internal(ctx context.Context, op string, err error) error {
	log.Printf("[%s] internal error: %v", op, err)
	return status.Error(codes.Internal, "internal server error")
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.Order, error) {
	if req.UserId == 0 || req.RestaurantId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and restaurant_id are required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items cannot be empty")
	}

	items := protoItemsToModel(req.Items)
	order, err := h.uc.CreateOrder(ctx, req.UserId, req.RestaurantId, items)
	if errors.Is(err, usecase.ErrEmptyItems) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internal(ctx, "CreateOrder", err)
	}
	return modelToProto(order), nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.OrderIdRequest) (*pb.Order, error) {
	order, err := h.uc.GetOrder(ctx, req.OrderId)
	if errors.Is(err, usecase.ErrOrderNotFound) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return nil, internal(ctx, "GetOrder", err)
	}
	return modelToProto(order), nil
}

func (h *OrderHandler) ListUserOrders(ctx context.Context, req *pb.UserIdRequest) (*pb.OrderList, error) {
	orders, err := h.uc.ListUserOrders(ctx, req.UserId)
	if err != nil {
		return nil, internal(ctx, "ListUserOrders", err)
	}
	return &pb.OrderList{Orders: modelsToProto(orders)}, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.Order, error) {
	order, err := h.uc.UpdateOrderStatus(ctx, req.OrderId, model.OrderStatus(req.Status))
	if errors.Is(err, usecase.ErrOrderNotFound) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return nil, internal(ctx, "UpdateOrderStatus", err)
	}
	return modelToProto(order), nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.OrderIdRequest) (*pb.Order, error) {
	order, err := h.uc.CancelOrder(ctx, req.OrderId)
	if errors.Is(err, usecase.ErrOrderNotFound) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if errors.Is(err, usecase.ErrAlreadyCancelled) {
		return nil, status.Error(codes.FailedPrecondition, "order already cancelled")
	}
	if err != nil {
		return nil, internal(ctx, "CancelOrder", err)
	}
	return modelToProto(order), nil
}

func (h *OrderHandler) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	payment, err := h.uc.ProcessPayment(ctx, req.OrderId, req.Method, req.Amount)
	if errors.Is(err, usecase.ErrOrderNotFound) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return nil, internal(ctx, "ProcessPayment", err)
	}
	return &pb.PaymentResponse{
		Payment: &pb.Payment{
			Id:      payment.ID,
			OrderId: payment.OrderID,
			Status:  string(payment.Status),
			Amount:  payment.Amount,
			Method:  payment.Method,
		},
		Success: payment.Status == model.PaymentSuccess,
	}, nil
}

func (h *OrderHandler) GetOrderHistory(ctx context.Context, req *pb.UserIdRequest) (*pb.OrderList, error) {
	orders, err := h.uc.GetOrderHistory(ctx, req.UserId)
	if err != nil {
		return nil, internal(ctx, "GetOrderHistory", err)
	}
	return &pb.OrderList{Orders: modelsToProto(orders)}, nil
}

func (h *OrderHandler) CalculateTotal(ctx context.Context, req *pb.CartRequest) (*pb.TotalResponse, error) {
	items := protoItemsToModel(req.Items)
	subtotal, fee, total := h.uc.CalculateTotal(ctx, items)
	return &pb.TotalResponse{
		Subtotal:    subtotal,
		DeliveryFee: fee,
		Total:       total,
	}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func protoItemsToModel(pbItems []*pb.OrderItem) []model.OrderItem {
	items := make([]model.OrderItem, 0, len(pbItems))
	for _, i := range pbItems {
		items = append(items, model.OrderItem{
			MenuItemID: i.MenuItemId,
			Name:       i.Name,
			Quantity:   i.Quantity,
			Price:      i.Price,
		})
	}
	return items
}

func modelToProto(o *model.Order) *pb.Order {
	return &pb.Order{
		Id:           o.ID,
		UserId:       o.UserID,
		RestaurantId: o.RestaurantID,
		Status:       string(o.Status),
		Total:        o.Total,
		CreatedAt:    o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    o.UpdatedAt.Format(time.RFC3339),
	}
}

func modelsToProto(orders []model.Order) []*pb.Order {
	result := make([]*pb.Order, 0, len(orders))
	for i := range orders {
		result = append(result, modelToProto(&orders[i]))
	}
	return result
}
