package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"food-delivery/restaurant-service/internal/model"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// RestaurantRepository defines all persistence operations.
type RestaurantRepository interface {
	// Restaurant CRUD
	CreateRestaurant(ctx context.Context, r *model.Restaurant) error
	GetRestaurantByID(ctx context.Context, id int64) (*model.Restaurant, error)
	UpdateRestaurant(ctx context.Context, r *model.Restaurant) error
	DeleteRestaurant(ctx context.Context, id int64) error
	ListRestaurants(ctx context.Context, f model.ListFilter) ([]*model.Restaurant, int, error)
	SearchRestaurants(ctx context.Context, f model.SearchFilter) ([]*model.Restaurant, int, error)

	// MenuItem CRUD
	CreateMenuItem(ctx context.Context, item *model.MenuItem) error
	GetMenuItemByID(ctx context.Context, id int64) (*model.MenuItem, error)
	UpdateMenuItem(ctx context.Context, item *model.MenuItem) error
	DeleteMenuItem(ctx context.Context, id int64) error
	GetMenuByRestaurant(ctx context.Context, restaurantID int64) ([]*model.MenuItem, error)
}

type postgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo creates a PostgreSQL-backed RestaurantRepository.
func NewPostgresRepo(db *sql.DB) RestaurantRepository {
	return &postgresRepo{db: db}
}

// ── Restaurant ────────────────────────────────────────────────────────────────

func (r *postgresRepo) CreateRestaurant(ctx context.Context, rest *model.Restaurant) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO restaurants (name, description, category_id, address, phone, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, is_active, rating, created_at, updated_at`,
		rest.Name, rest.Description, rest.CategoryID,
		rest.Address, rest.Phone, rest.ImageURL,
	).Scan(&rest.ID, &rest.IsActive, &rest.Rating, &rest.CreatedAt, &rest.UpdatedAt)
}

func (r *postgresRepo) GetRestaurantByID(ctx context.Context, id int64) (*model.Restaurant, error) {
	rest := &model.Restaurant{}
	err := r.db.QueryRowContext(ctx,
		`SELECT r.id, r.name, r.description, r.category_id,
		        COALESCE(c.name, '') AS category_name,
		        r.address, r.phone, r.image_url,
		        r.is_active, r.rating, r.created_at, r.updated_at
		 FROM restaurants r
		 LEFT JOIN categories c ON c.id = r.category_id
		 WHERE r.id = $1`,
		id,
	).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.CategoryID, &rest.CategoryName,
		&rest.Address, &rest.Phone, &rest.ImageURL, &rest.IsActive, &rest.Rating,
		&rest.CreatedAt, &rest.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return rest, err
}

func (r *postgresRepo) UpdateRestaurant(ctx context.Context, rest *model.Restaurant) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE restaurants
		 SET name=$1, description=$2, address=$3, phone=$4, image_url=$5, is_active=$6, updated_at=NOW()
		 WHERE id=$7`,
		rest.Name, rest.Description, rest.Address, rest.Phone,
		rest.ImageURL, rest.IsActive, rest.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepo) DeleteRestaurant(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM restaurants WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepo) ListRestaurants(ctx context.Context, f model.ListFilter) ([]*model.Restaurant, int, error) {
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PageSize

	baseWhere := "WHERE r.is_active = true"
	args := []interface{}{}
	argN := 1

	if f.CategoryID > 0 {
		baseWhere += fmt.Sprintf(" AND r.category_id = $%d", argN)
		args = append(args, f.CategoryID)
		argN++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM restaurants r " + baseWhere
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.PageSize, offset)
	listQuery := fmt.Sprintf(
		`SELECT r.id, r.name, r.description, r.category_id,
		        COALESCE(c.name,'') AS category_name,
		        r.address, r.phone, r.image_url,
		        r.is_active, r.rating, r.created_at, r.updated_at
		 FROM restaurants r
		 LEFT JOIN categories c ON c.id = r.category_id
		 %s
		 ORDER BY r.rating DESC, r.created_at DESC
		 LIMIT $%d OFFSET $%d`, baseWhere, argN, argN+1,
	)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanRestaurants(rows, total)
}

func (r *postgresRepo) SearchRestaurants(ctx context.Context, f model.SearchFilter) ([]*model.Restaurant, int, error) {
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PageSize

	pattern := "%" + f.Query + "%"

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM restaurants r
		 WHERE r.is_active = true
		   AND (r.name ILIKE $1 OR r.description ILIKE $1 OR r.address ILIKE $1)`,
		pattern,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.name, r.description, r.category_id,
		        COALESCE(c.name,'') AS category_name,
		        r.address, r.phone, r.image_url,
		        r.is_active, r.rating, r.created_at, r.updated_at
		 FROM restaurants r
		 LEFT JOIN categories c ON c.id = r.category_id
		 WHERE r.is_active = true
		   AND (r.name ILIKE $1 OR r.description ILIKE $1 OR r.address ILIKE $1)
		 ORDER BY r.rating DESC
		 LIMIT $2 OFFSET $3`,
		pattern, f.PageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanRestaurants(rows, total)
}

func scanRestaurants(rows *sql.Rows, total int) ([]*model.Restaurant, int, error) {
	var list []*model.Restaurant
	for rows.Next() {
		rest := &model.Restaurant{}
		if err := rows.Scan(
			&rest.ID, &rest.Name, &rest.Description, &rest.CategoryID, &rest.CategoryName,
			&rest.Address, &rest.Phone, &rest.ImageURL, &rest.IsActive, &rest.Rating,
			&rest.CreatedAt, &rest.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, rest)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ── MenuItem ──────────────────────────────────────────────────────────────────

func (r *postgresRepo) CreateMenuItem(ctx context.Context, item *model.MenuItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify restaurant exists
	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM restaurants WHERE id=$1)`, item.RestaurantID,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO menu_items (restaurant_id, name, description, price, category, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, is_available, created_at, updated_at`,
		item.RestaurantID, item.Name, item.Description,
		item.Price, item.Category, item.ImageURL,
	).Scan(&item.ID, &item.IsAvailable, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresRepo) GetMenuItemByID(ctx context.Context, id int64) (*model.MenuItem, error) {
	item := &model.MenuItem{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, restaurant_id, name, description, price, category, image_url,
		        is_available, created_at, updated_at
		 FROM menu_items WHERE id=$1`,
		id,
	).Scan(
		&item.ID, &item.RestaurantID, &item.Name, &item.Description,
		&item.Price, &item.Category, &item.ImageURL,
		&item.IsAvailable, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *postgresRepo) UpdateMenuItem(ctx context.Context, item *model.MenuItem) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE menu_items
		 SET name=$1, description=$2, price=$3, category=$4, image_url=$5, is_available=$6, updated_at=NOW()
		 WHERE id=$7 AND restaurant_id=$8`,
		item.Name, item.Description, item.Price, item.Category,
		item.ImageURL, item.IsAvailable, item.ID, item.RestaurantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepo) DeleteMenuItem(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM menu_items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepo) GetMenuByRestaurant(ctx context.Context, restaurantID int64) ([]*model.MenuItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, name, description, price, category, image_url,
		        is_available, created_at, updated_at
		 FROM menu_items
		 WHERE restaurant_id=$1
		 ORDER BY category, name`,
		restaurantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.MenuItem
	for rows.Next() {
		item := &model.MenuItem{}
		if err := rows.Scan(
			&item.ID, &item.RestaurantID, &item.Name, &item.Description,
			&item.Price, &item.Category, &item.ImageURL,
			&item.IsAvailable, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}
