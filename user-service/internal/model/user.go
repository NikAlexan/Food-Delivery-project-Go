package model

import "time"

type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Name         string    `db:"name"`
	Phone        string    `db:"phone"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Address struct {
	ID        int64  `db:"id"`
	UserID    int64  `db:"user_id"`
	Street    string `db:"street"`
	City      string `db:"city"`
	Zip       string `db:"zip"`
	IsDefault bool   `db:"is_default"`
}

type RefreshToken struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}
