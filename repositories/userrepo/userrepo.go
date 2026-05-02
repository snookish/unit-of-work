package userrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/snookish/unit-of-work/models"
)

type Repository interface {
	Create(ctx context.Context, u *models.User) error
	QueryByID(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, u *models.User) error
}

type repository struct {
	tx *sql.Tx
}

func NewRepository(tx *sql.Tx) *repository {
	return &repository{tx: tx}
}

func (r *repository) Create(ctx context.Context, u *models.User) error {
	const q = `
		INSERT INTO users (first_name, last_name, email, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	u.CreatedAt = time.Now().UTC()
	if err := r.tx.QueryRowContext(ctx, q, u.FirstName, u.LastName, u.Email, u.CreatedAt).Scan(&u.ID); err != nil {
		return fmt.Errorf("user repo create: %w", err)
	}

	return nil
}

func (r *repository) QueryByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	const q = `
		SELECT id, first_name, last_name, email, created_at
		FROM users
		WHERE id = $1
	`

	if err := r.tx.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("user repo query by id: %w", err)
	}
	return u, nil
}

func (r *repository) Update(ctx context.Context, u *models.User) error {
	const q = `
		UPDATE users
		SET first_name = $1, last_name = $2, email = $3
		WHERE id = $4
	`

	res, err := r.tx.ExecContext(ctx, q, u.FirstName, u.LastName, u.Email, u.ID)
	if err != nil {
		return fmt.Errorf("user repo update: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("user repo update rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user repo update: no rows affected for id %d", u.ID)
	}

	return nil
}
