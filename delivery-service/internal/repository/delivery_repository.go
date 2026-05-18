package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"food-delivery/delivery-service/internal/model"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrNoAvailableDriver  = errors.New("no available driver")
	ErrAlreadyRegistered  = errors.New("already registered as driver")
)

type DeliveryRepository interface {
	AssignDriver(ctx context.Context, orderID, userID int64, userEmail, address string) (*model.Delivery, bool, error)
	GetByID(ctx context.Context, deliveryID int64) (*model.Delivery, error)
	GetByOrderID(ctx context.Context, orderID int64) (*model.Delivery, error)
	UpdateDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error
	MarkInTransitByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error)
	CompleteDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error)
	CancelDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error)
	CancelByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error)
	ListDriverDeliveries(ctx context.Context, driverID int64) ([]model.Delivery, error)
	GetDeliveryHistory(ctx context.Context, userID int64) ([]model.Delivery, error)
	CreateDriver(ctx context.Context, d *model.Driver) (*model.Driver, error)
	GetDriverByUserID(ctx context.Context, userID int64) (*model.Driver, error)
	SetDriverAvailability(ctx context.Context, driverID int64, available bool) (*model.Driver, error)
}

type postgresDeliveryRepo struct {
	db *sql.DB
}

func NewPostgresDeliveryRepo(db *sql.DB) DeliveryRepository {
	return &postgresDeliveryRepo{db: db}
}

const deliverySelect = `
SELECT
	d.id,
	d.order_id,
	d.user_id,
	d.user_email,
	d.driver_id,
	COALESCE(dr.name, ''),
	COALESCE(dr.email, ''),
	COALESCE(dr.phone, ''),
	d.status,
	d.delivery_address,
	d.created_at,
	d.updated_at,
	d.assigned_at,
	d.completed_at,
	COALESCE(dr.current_latitude, 0),
	COALESCE(dr.current_longitude, 0)
FROM deliveries d
LEFT JOIN drivers dr ON dr.id = d.driver_id
`

func (r *postgresDeliveryRepo) AssignDriver(ctx context.Context, orderID, userID int64, userEmail, address string) (*model.Delivery, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	existing, err := r.getByOrderIDTx(ctx, tx, orderID)
	if err == nil {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, false, err
	}

	driver := &model.Driver{}
	err = tx.QueryRowContext(ctx, `
		SELECT id, name, email, phone, current_latitude, current_longitude
		FROM drivers
		WHERE is_available = TRUE
		ORDER BY id
		LIMIT 1
		FOR UPDATE SKIP LOCKED`,
	).Scan(
		&driver.ID,
		&driver.Name,
		&driver.Email,
		&driver.Phone,
		&driver.CurrentLatitude,
		&driver.CurrentLongitude,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, ErrNoAvailableDriver
	}
	if err != nil {
		return nil, false, err
	}

	var deliveryID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO deliveries (order_id, user_id, user_email, driver_id, status, delivery_address, assigned_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id`,
		orderID,
		userID,
		userEmail,
		driver.ID,
		model.StatusAssigned,
		address,
	).Scan(&deliveryID)
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}

	delivery, err := r.GetByID(ctx, deliveryID)
	return delivery, true, err
}

func (r *postgresDeliveryRepo) GetByID(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	return scanDelivery(r.db.QueryRowContext(ctx, deliverySelect+` WHERE d.id = $1`, deliveryID))
}

func (r *postgresDeliveryRepo) GetByOrderID(ctx context.Context, orderID int64) (*model.Delivery, error) {
	return scanDelivery(r.db.QueryRowContext(ctx, deliverySelect+` WHERE d.order_id = $1`, orderID))
}

func (r *postgresDeliveryRepo) UpdateDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE drivers
		SET current_latitude = $1, current_longitude = $2, updated_at = NOW()
		WHERE id = $3`,
		latitude,
		longitude,
		driverID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO locations (driver_id, latitude, longitude)
		VALUES ($1, $2, $3)`,
		driverID,
		latitude,
		longitude,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresDeliveryRepo) MarkInTransitByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error) {
	var deliveryID int64
	err := r.db.QueryRowContext(ctx, `
		UPDATE deliveries
		SET status = $1, updated_at = NOW()
		WHERE order_id = $2 AND status = $3
		RETURNING id`,
		model.StatusInTransit,
		orderID,
		model.StatusAssigned,
	).Scan(&deliveryID)
	if errors.Is(err, sql.ErrNoRows) {
		delivery, getErr := r.GetByOrderID(ctx, orderID)
		return delivery, false, getErr
	}
	if err != nil {
		return nil, false, err
	}
	delivery, err := r.GetByID(ctx, deliveryID)
	return delivery, true, err
}

func (r *postgresDeliveryRepo) CompleteDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error) {
	return r.finishDelivery(ctx, "id", deliveryID, model.StatusCompleted)
}

func (r *postgresDeliveryRepo) CancelDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, bool, error) {
	return r.finishDelivery(ctx, "id", deliveryID, model.StatusCancelled)
}

func (r *postgresDeliveryRepo) CancelByOrderID(ctx context.Context, orderID int64) (*model.Delivery, bool, error) {
	return r.finishDelivery(ctx, "order_id", orderID, model.StatusCancelled)
}

func (r *postgresDeliveryRepo) ListDriverDeliveries(ctx context.Context, driverID int64) ([]model.Delivery, error) {
	rows, err := r.db.QueryContext(ctx, deliverySelect+` WHERE d.driver_id = $1 ORDER BY d.created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeliveries(rows)
}

func (r *postgresDeliveryRepo) GetDeliveryHistory(ctx context.Context, userID int64) ([]model.Delivery, error) {
	rows, err := r.db.QueryContext(ctx, deliverySelect+` WHERE d.user_id = $1 ORDER BY d.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeliveries(rows)
}

func (r *postgresDeliveryRepo) finishDelivery(ctx context.Context, key string, id int64, status string) (*model.Delivery, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	query := `
		UPDATE deliveries
		SET status = $1,
		    updated_at = NOW(),
		    completed_at = CASE WHEN $1 = $2 THEN NOW() ELSE completed_at END
		WHERE ` + key + ` = $3 AND status NOT IN ($4, $5)
		RETURNING id, driver_id`

	var deliveryID, driverID int64
	err = tx.QueryRowContext(
		ctx,
		query,
		status,
		model.StatusCompleted,
		id,
		model.StatusCompleted,
		model.StatusCancelled,
	).Scan(&deliveryID, &driverID)
	if errors.Is(err, sql.ErrNoRows) {
		var existing *model.Delivery
		if key == "id" {
			existing, err = r.GetByID(ctx, id)
		} else {
			existing, err = r.GetByOrderID(ctx, id)
		}
		if err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}

	delivery, err := r.GetByID(ctx, deliveryID)
	return delivery, true, err
}

func (r *postgresDeliveryRepo) CreateDriver(ctx context.Context, d *model.Driver) (*model.Driver, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO drivers (user_id, name, email, phone)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, is_available, current_latitude, current_longitude, created_at, updated_at`,
		d.UserID, d.Name, d.Email, d.Phone,
	).Scan(&d.ID, &d.IsAvailable, &d.CurrentLatitude, &d.CurrentLongitude, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrAlreadyRegistered
		}
		return nil, err
	}
	return d, nil
}

func (r *postgresDeliveryRepo) GetDriverByUserID(ctx context.Context, userID int64) (*model.Driver, error) {
	d := &model.Driver{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, email, phone, is_available,
		        current_latitude, current_longitude, created_at, updated_at
		 FROM drivers WHERE user_id = $1`,
		userID,
	).Scan(&d.ID, &d.UserID, &d.Name, &d.Email, &d.Phone, &d.IsAvailable,
		&d.CurrentLatitude, &d.CurrentLongitude, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return d, err
}

func (r *postgresDeliveryRepo) getByOrderIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*model.Delivery, error) {
	return scanDelivery(tx.QueryRowContext(ctx, deliverySelect+` WHERE d.order_id = $1 FOR UPDATE OF d`, orderID))
}

type scanRow interface {
	Scan(dest ...any) error
}

func scanDelivery(row scanRow) (*model.Delivery, error) {
	delivery := &model.Delivery{}
	var assignedAt sql.NullTime
	var completedAt sql.NullTime

	err := row.Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.UserID,
		&delivery.UserEmail,
		&delivery.DriverID,
		&delivery.DriverName,
		&delivery.DriverEmail,
		&delivery.DriverPhone,
		&delivery.Status,
		&delivery.DeliveryAddress,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
		&assignedAt,
		&completedAt,
		&delivery.CurrentLatitude,
		&delivery.CurrentLongitude,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if assignedAt.Valid {
		t := assignedAt.Time
		delivery.AssignedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		delivery.CompletedAt = &t
	}

	return delivery, nil
}

func scanDeliveries(rows *sql.Rows) ([]model.Delivery, error) {
	var deliveries []model.Delivery
	for rows.Next() {
		delivery, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, *delivery)
	}
	return deliveries, rows.Err()
}

func (r *postgresDeliveryRepo) SetDriverAvailability(ctx context.Context, driverID int64, available bool) (*model.Driver, error) {
	d := &model.Driver{}
	err := r.db.QueryRowContext(ctx,
		`UPDATE drivers SET is_available = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, user_id, name, email, phone, is_available, current_latitude, current_longitude, created_at, updated_at`,
		available, driverID,
	).Scan(&d.ID, &d.UserID, &d.Name, &d.Email, &d.Phone, &d.IsAvailable,
		&d.CurrentLatitude, &d.CurrentLongitude, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return d, err
}
