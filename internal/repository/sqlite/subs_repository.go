package sqlite

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
		SELECT id, feed_id, contact_id, source, source_group_id, enabled, created_at
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
		var source string
		var groupID sql.NullString
		if err := rows.Scan(&s.ID, &s.FeedID, &s.ContactID, &source, &groupID, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Source = domain.SubscriptionSource(source)
		if groupID.Valid {
			s.GroupID = &groupID.String
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SubscriptionsRepository) ListByContact(ctx context.Context, contactID string) ([]domain.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, feed_id, contact_id, source, source_group_id, enabled, created_at
		FROM subscriptions
		WHERE contact_id=$1 AND enabled=true
		ORDER BY created_at DESC
	`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var source string
		var groupID sql.NullString
		if err := rows.Scan(&s.ID, &s.FeedID, &s.ContactID, &source, &groupID, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Source = domain.SubscriptionSource(source)
		if groupID.Valid {
			s.GroupID = &groupID.String
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SubscriptionsRepository) Add(ctx context.Context, feedID, contactID string) error {
	const q = `
INSERT INTO subscriptions (feed_id, contact_id, source, source_group_id, created_at)
VALUES ($1, $2, 'direct', NULL, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING
`
	_, err := r.db.ExecContext(ctx, q, feedID, contactID)
	return err
}

func (r *SubscriptionsRepository) Remove(ctx context.Context, feedID, contactID string) error {
	const q = `
DELETE FROM subscriptions
WHERE feed_id = $1 AND contact_id = $2 AND source = 'direct'
`
	_, err := r.db.ExecContext(ctx, q, feedID, contactID)
	return err
}

func (r *SubscriptionsRepository) AddGroup(ctx context.Context, feedID, contactID, groupID string) error {
	const q = `
INSERT INTO subscriptions (feed_id, contact_id, source, source_group_id, created_at)
VALUES ($1, $2, 'group', $3, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING
`
	_, err := r.db.ExecContext(ctx, q, feedID, contactID, groupID)
	return err
}

func (r *SubscriptionsRepository) RemoveGroupByFeed(ctx context.Context, groupID, feedID string) error {
	_, err := r.db.ExecContext(ctx, `
DELETE FROM subscriptions
WHERE source = 'group' AND source_group_id = $1 AND feed_id = $2
`, groupID, feedID)
	return err
}

func (r *SubscriptionsRepository) RemoveGroupByContact(ctx context.Context, groupID, contactID string) error {
	_, err := r.db.ExecContext(ctx, `
DELETE FROM subscriptions
WHERE source = 'group' AND source_group_id = $1 AND contact_id = $2
`, groupID, contactID)
	return err
}
