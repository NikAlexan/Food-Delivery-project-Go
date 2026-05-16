package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"food-delivery/order-service/internal/model"
)

var ErrNotFound = errors.New("not found")

type OrderRepository interface {
	CreateOrderWithItems(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus) (*model.Order, error)

	CreatePayment(ctx context.Context, payment *model.Payment) error
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*model.Payment, error)
	UpdatePaymentStatus(ctx context.Context, orderID int64, status model.PaymentStatus) error
}

type postgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) OrderRepository {
	return &postgresOrderRepo{db: db}
}

// CreateOrderWithItems inserts order + items atomically in one transaction.
func (r *postgresOrderRepo) CreateOrderWithItems(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, restaurant_id, status, total, delivery_address, user_email)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		order.UserID, order.RestaurantID, order.Status, order.Total, order.DeliveryAddress, order.UserEmail,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for i := range order.Items {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO order_items (order_id, menu_item_id, name, quantity, price)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			order.ID, order.Items[i].MenuItemID, order.Items[i].Name,
			order.Items[i].Quantity, order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return fmt.Errorf("insert order_item: %w", err)
		}
		order.Items[i].OrderID = order.ID
	}

	return tx.Commit()
}

func (r *postgresOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, restaurant_id, status, total, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *postgresOrderRepo) getItemsByOrderID(ctx context.Context, orderID int64) ([]model.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, menu_item_id, name, quantity, price
		 FROM order_items WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID,
			&item.Name, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *postgresOrderRepo) ListByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, restaurant_id, status, total, created_at, updated_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.RestaurantID,
			&o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *postgresOrderRepo) UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, user_id, restaurant_id, status, total, created_at, updated_at`,
		status, orderID,
	).Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return o, err
}

// ── Payment ───────────────────────────────────────────────────────────────────

// CreatePayment inserts a payment row; atomically pairs with UpdateStatus
// via ProcessPayment usecase.
func (r *postgresOrderRepo) CreatePayment(ctx context.Context, payment *model.Payment) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO payments (order_id, status, amount, method)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		payment.OrderID, payment.Status, payment.Amount, payment.Method,
	).Scan(&payment.ID, &payment.CreatedAt)
}

func (r *postgresOrderRepo) GetPaymentByOrderID(ctx context.Context, orderID int64) (*model.Payment, error) {
	p := &model.Payment{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, status, amount, method, created_at
		 FROM payments WHERE order_id = $1`, orderID,
	).Scan(&p.ID, &p.OrderID, &p.Status, &p.Amount, &p.Method, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *postgresOrderRepo) UpdatePaymentStatus(ctx context.Context, orderID int64, status model.PaymentStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payments SET status = $1 WHERE order_id = $2`,
		status, orderID,
	)
	return err
}
