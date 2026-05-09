package model

import "time"

// Category represents a restaurant category (e.g. Pizza, Sushi, Burgers).
type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// Restaurant is the core domain entity.
type Restaurant struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	CategoryID   int64     `db:"category_id"`
	CategoryName string    `db:"category_name"` // joined field
	Address      string    `db:"address"`
	Phone        string    `db:"phone"`
	ImageURL     string    `db:"image_url"`
	IsActive     bool      `db:"is_active"`
	Rating       float64   `db:"rating"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// MenuItem belongs to a restaurant's menu.
type MenuItem struct {
	ID           int64     `db:"id"`
	RestaurantID int64     `db:"restaurant_id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	Price        float64   `db:"price"`
	Category     string    `db:"category"`
	ImageURL     string    `db:"image_url"`
	IsAvailable  bool      `db:"is_available"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ListFilter carries pagination and optional category filter for restaurant listing.
type ListFilter struct {
	CategoryID int64
	Page       int32
	PageSize   int32
}

// SearchFilter carries a full-text query with pagination.
type SearchFilter struct {
	Query    string
	Page     int32
	PageSize int32
}
