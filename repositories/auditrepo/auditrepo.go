package auditrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/snookish/unit-of-work/models"
)

type Repository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	ListByUserID(ctx context.Context, userID int64) ([]*models.AuditLog, error)
}

type repository struct {
	tx *sql.Tx
}

func NewRepository(tx *sql.Tx) *repository {
	return &repository{tx: tx}
}

func (r *repository) Create(ctx context.Context, log *models.AuditLog) error {
	const q = `
		INSERT INTO audit_logs (user_id, action, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	log.Timestamp = time.Now().UTC()
	if err := r.tx.QueryRowContext(ctx, q, log.UserID, log.Action, log.Timestamp).Scan(&log.ID); err != nil {
		return fmt.Errorf("audit log repo create: %w", err)
	}

	return nil
}

func (r *repository) ListByUserID(ctx context.Context, userID int64) ([]*models.AuditLog, error) {
	const q = `
		SELECT id, user_id, action, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.tx.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("audit log repo list: %w", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		l := &models.AuditLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("audit log repo scan: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}
