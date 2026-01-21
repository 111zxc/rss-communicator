package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type DeliveriesRepository struct{ db *sql.DB }

func NewDeliveriesRepository(db *sql.DB) *DeliveriesRepository { return &DeliveriesRepository{db: db} }

func (r *DeliveriesRepository) CreatePendingIfNotExists(ctx context.Context, contactID, itemID string) (bool, string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO deliveries (contact_id, item_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (contact_id, item_id) DO NOTHING
		RETURNING id
	`, contactID, itemID).Scan(&id)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, id, nil
}

func (r *DeliveriesRepository) MarkSent(ctx context.Context, deliveryID string, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE deliveries SET status='sent', sent_at=$2, updated_at=now()
		WHERE id=$1
	`, deliveryID, sentAt)
	return err
}

func (r *DeliveriesRepository) MarkFailed(ctx context.Context, deliveryID string, errMsg string, nextRetryAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE deliveries
		SET status='failed', attempt_count=attempt_count+1, last_error=$2, last_attempt_at=now(),
		    next_retry_at=$3, updated_at=now()
		WHERE id=$1
	`, deliveryID, errMsg, nextRetryAt)
	return err
}

func (r *DeliveriesRepository) GetByID(ctx context.Context, id string) (domain.Delivery, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, item_id, contact_id, status, attempt_count, last_error, last_attempt_at, sent_at, next_retry_at, created_at, updated_at
		FROM deliveries WHERE id=$1
	`, id)

	var d domain.Delivery
	var st string
	var lastErr sql.NullString
	var lastAttempt, sentAt, nextRetry sql.NullTime

	if err := row.Scan(&d.ID, &d.ItemID, &d.ContactID, &st, &d.AttemptCount, &lastErr, &lastAttempt, &sentAt, &nextRetry, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return domain.Delivery{}, err
	}
	d.Status = domain.DeliveryStatus(st)
	if lastErr.Valid {
		d.LastError = &lastErr.String
	}
	if lastAttempt.Valid {
		t := lastAttempt.Time
		d.LastAttemptAt = &t
	}
	if sentAt.Valid {
		t := sentAt.Time
		d.SentAt = &t
	}
	if nextRetry.Valid {
		t := nextRetry.Time
		d.NextRetryAt = &t
	}
	return d, nil
}

func (r *DeliveriesRepository) ListRetryDue(ctx context.Context, now time.Time, limit int, maxAttempts int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM deliveries
		WHERE status = 'failed'
		  AND next_retry_at IS NOT NULL
		  AND next_retry_at <= $1
		  AND attempt_count < $2
		ORDER BY next_retry_at ASC
		LIMIT $3
	`, now, maxAttempts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
