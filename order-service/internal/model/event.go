package model

type OrderCreatedEvent struct {
	OrderID         int64  `json:"order_id"`
	UserID          int64  `json:"user_id"`
	UserEmail       string `json:"user_email"`
	DeliveryAddress string `json:"delivery_address"`
}
