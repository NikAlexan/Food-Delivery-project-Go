package model

import "time"

type OrderEvent struct {
	OrderID         int64  `json:"order_id"`
	UserID          int64  `json:"user_id"`
	UserEmail       string `json:"user_email"`
	DeliveryAddress string `json:"delivery_address"`
}

type DeliveryCompletedEvent struct {
	DeliveryID  int64     `json:"delivery_id"`
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	DriverID    int64     `json:"driver_id"`
	CompletedAt time.Time `json:"completed_at"`
}
