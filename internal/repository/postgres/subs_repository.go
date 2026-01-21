package postgres

import (
	"context"
	"database/sql"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type SubscriptionsRepository struct{ db *sql.DB }

func NewSubscriptionsRepository(db *sql.DB) *SubscriptionsRepository {
	return &SubscriptionsRepository{db: db}
}

func (r *SubscriptionsRepository) ListByFeed(ctx context.Context, feedID string) ([]domain.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, feed_id, contact_id, enabled, created_at
		FROM subscriptions
		WHERE feed_id=$1 AND enabled=true
	`, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(&s.ID, &s.FeedID, &s.ContactID, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SubscriptionsRepository) Add(ctx context.Context, feedID, contactID string) error {
	const q = `
INSERT INTO subscriptions (feed_id, contact_id, created_at)
VALUES ($1, $2, now())
ON CONFLICT (feed_id, contact_id) DO NOTHING
`
	_, err := r.db.ExecContext(ctx, q, feedID, contactID)
	return err
}

func (r *SubscriptionsRepository) Remove(ctx context.Context, feedID, contactID string) error {
	const q = `
DELETE FROM subscriptions
WHERE feed_id = $1 AND contact_id = $2
`
	_, err := r.db.ExecContext(ctx, q, feedID, contactID)
	return err
}
