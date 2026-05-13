package model

import "time"

const (
	StatusAssigned  = "assigned"
	StatusInTransit = "in_transit"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

type Delivery struct {
	ID               int64      `db:"id"`
	OrderID          int64      `db:"order_id"`
	UserID           int64      `db:"user_id"`
	UserEmail        string     `db:"user_email"`
	DriverID         int64      `db:"driver_id"`
	DriverName       string     `db:"driver_name"`
	DriverEmail      string     `db:"driver_email"`
	DriverPhone      string     `db:"driver_phone"`
	Status           string     `db:"status"`
	DeliveryAddress  string     `db:"delivery_address"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	AssignedAt       *time.Time `db:"assigned_at"`
	CompletedAt      *time.Time `db:"completed_at"`
	CurrentLatitude  float64    `db:"current_latitude"`
	CurrentLongitude float64    `db:"current_longitude"`
}

type Driver struct {
	ID               int64     `db:"id"`
	Name             string    `db:"name"`
	Email            string    `db:"email"`
	Phone            string    `db:"phone"`
	IsAvailable      bool      `db:"is_available"`
	CurrentLatitude  float64   `db:"current_latitude"`
	CurrentLongitude float64   `db:"current_longitude"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type Location struct {
	ID        int64     `db:"id"`
	DriverID  int64     `db:"driver_id"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	CreatedAt time.Time `db:"created_at"`
}
