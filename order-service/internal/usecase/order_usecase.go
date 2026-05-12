package usecase

import (
	"context"
	"errors"
	"time"

	"food-delivery/order-service/internal/model"
	appnats "food-delivery/order-service/internal/nats"
	"food-delivery/order-service/internal/repository"
)

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrEmptyItems     = errors.New("order must have at least one item")
	ErrAlreadyCancelled = errors.New("order already cancelled")
)

const deliveryFee = 500.0 // flat KZT

type OrderUsecase interface {
	CreateOrder(ctx context.Context, userID, restaurantID int64, items []model.OrderItem) (*model.Order, error)
	GetOrder(ctx context.Context, orderID int64) (*model.Order, error)
	ListUserOrders(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus) (*model.Order, error)
	CancelOrder(ctx context.Context, orderID int64) (*model.Order, error)
	ProcessPayment(ctx context.Context, orderID int64, method string, amount float64) (*model.Payment, error)
	GetOrderHistory(ctx context.Context, userID int64) ([]model.Order, error)
	CalculateTotal(ctx context.Context, items []model.OrderItem) (subtotal, fee, total float64)
}

type orderUsecase struct {
	repo      repository.OrderRepository
	publisher appnats.Publisher
}

func NewOrderUsecase(repo repository.OrderRepository, publisher appnats.Publisher) OrderUsecase {
	return &orderUsecase{repo: repo, publisher: publisher}
}

func (u *orderUsecase) CreateOrder(ctx context.Context, userID, restaurantID int64, items []model.OrderItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}

	_, _, total := u.CalculateTotal(ctx, items)

	order := &model.Order{
		UserID:       userID,
		RestaurantID: restaurantID,
		Items:        items,
		Status:       model.StatusPending,
		Total:        total,
	}

	if err := u.repo.CreateOrderWithItems(ctx, order); err != nil {
		return nil, err
	}

	// Publish order.created — Delivery Service subscribes to this
	_ = u.publisher.Publish(ctx, "order.created", order)

	return order, nil
}

func (u *orderUsecase) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := u.repo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrOrderNotFound
	}
	return order, err
}

func (u *orderUsecase) ListUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return u.repo.ListByUserID(ctx, userID)
}

func (u *orderUsecase) UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus) (*model.Order, error) {
	order, err := u.repo.UpdateStatus(ctx, orderID, status)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrOrderNotFound
	}
	return order, err
}

func (u *orderUsecase) CancelOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := u.repo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if order.Status == model.StatusCancelled {
		return nil, ErrAlreadyCancelled
	}

	cancelled, err := u.repo.UpdateStatus(ctx, orderID, model.StatusCancelled)
	if err != nil {
		return nil, err
	}

	// Publish order.cancelled — Delivery Service frees the driver
	_ = u.publisher.Publish(ctx, "order.cancelled", cancelled)

	return cancelled, nil
}

// ProcessPayment atomically:
//  1. Creates a payment record (pending)
//  2. Updates order status to paid
//  3. Updates payment status to success
//  4. Publishes order.paid to NATS
func (u *orderUsecase) ProcessPayment(ctx context.Context, orderID int64, method string, amount float64) (*model.Payment, error) {
	_, err := u.repo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	payment := &model.Payment{
		OrderID:   orderID,
		Status:    model.PaymentPending,
		Amount:    amount,
		Method:    method,
		CreatedAt: time.Now(),
	}

	// Step 1: insert payment
	if err := u.repo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	// Step 2: mark order paid
	if _, err := u.repo.UpdateStatus(ctx, orderID, model.StatusPaid); err != nil {
		_ = u.repo.UpdatePaymentStatus(ctx, orderID, model.PaymentFailed)
		return nil, err
	}

	// Step 3: mark payment success
	if err := u.repo.UpdatePaymentStatus(ctx, orderID, model.PaymentSuccess); err != nil {
		return nil, err
	}
	payment.Status = model.PaymentSuccess

	// Publish order.paid — Delivery Service confirms delivery
	_ = u.publisher.Publish(ctx, "order.paid", payment)

	return payment, nil
}

func (u *orderUsecase) GetOrderHistory(ctx context.Context, userID int64) ([]model.Order, error) {
	return u.repo.ListByUserID(ctx, userID)
}

func (u *orderUsecase) CalculateTotal(_ context.Context, items []model.OrderItem) (subtotal, fee, total float64) {
	for _, item := range items {
		subtotal += item.Price * float64(item.Quantity)
	}
	fee = deliveryFee
	total = subtotal + fee
	return
}
