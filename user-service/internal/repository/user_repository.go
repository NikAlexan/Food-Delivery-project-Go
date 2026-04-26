package repository

import (
	"context"
	"database/sql"
	"errors"

	"food-delivery/user-service/internal/model"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	CreateUserWithAddress(ctx context.Context, user *model.User, addr *model.Address) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	UpdateProfile(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id int64) error

	AddAddress(ctx context.Context, addr *model.Address) error
	GetAddresses(ctx context.Context, userID int64) ([]model.Address, error)

	SaveRefreshToken(ctx context.Context, rt *model.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type postgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) UserRepository {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) CreateUserWithAddress(ctx context.Context, user *model.User, addr *model.Address) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, name, phone) VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at`,
		user.Email, user.PasswordHash, user.Name, user.Phone,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}

	if addr != nil {
		addr.UserID = user.ID
		_, err = tx.ExecContext(ctx,
			`INSERT INTO addresses (user_id, street, city, zip, is_default) VALUES ($1,$2,$3,$4,$5)`,
			addr.UserID, addr.Street, addr.City, addr.Zip, addr.IsDefault,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, phone, created_at, updated_at FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, phone, created_at, updated_at FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *postgresUserRepo) UpdateProfile(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name=$1, phone=$2, updated_at=NOW() WHERE id=$3`,
		user.Name, user.Phone, user.ID,
	)
	return err
}

func (r *postgresUserRepo) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *postgresUserRepo) AddAddress(ctx context.Context, addr *model.Address) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO addresses (user_id, street, city, zip, is_default) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		addr.UserID, addr.Street, addr.City, addr.Zip, addr.IsDefault,
	).Scan(&addr.ID)
}

func (r *postgresUserRepo) GetAddresses(ctx context.Context, userID int64) ([]model.Address, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, street, city, zip, is_default FROM addresses WHERE user_id=$1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Address
	for rows.Next() {
		var a model.Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.Street, &a.City, &a.Zip, &a.IsDefault); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *postgresUserRepo) SaveRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)
		 ON CONFLICT (token) DO UPDATE SET expires_at=$3`,
		rt.UserID, rt.Token, rt.ExpiresAt,
	)
	return err
}

func (r *postgresUserRepo) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	rt := &model.RefreshToken{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token=$1`,
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return rt, err
}

func (r *postgresUserRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token=$1`, token)
	return err
}
