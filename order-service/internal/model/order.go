package model

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
	StatusDelivered OrderStatus = "delivered"
)

type OrderItem struct {
	ID         int64   `db:"id"`
	OrderID    int64   `db:"order_id"`
	MenuItemID int64   `db:"menu_item_id"`
	Name       string  `db:"name"`
	Quantity   int32   `db:"quantity"`
	Price      float64 `db:"price"`
}

type Order struct {
	ID           int64       `db:"id"`
	UserID       int64       `db:"user_id"`
	RestaurantID int64       `db:"restaurant_id"`
	Items        []OrderItem `db:"-"`
	Status       OrderStatus `db:"status"`
	Total        float64     `db:"total"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentSuccess PaymentStatus = "success"
	PaymentFailed  PaymentStatus = "failed"
)

type Payment struct {
	ID        int64         `db:"id"`
	OrderID   int64         `db:"order_id"`
	Status    PaymentStatus `db:"status"`
	Amount    float64       `db:"amount"`
	Method    string        `db:"method"`
	CreatedAt time.Time     `db:"created_at"`
}
