package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type DeliveriesRepository struct {
	db   *sql.DB
	exec dbtx
}

func NewDeliveriesRepository(db *sql.DB) *DeliveriesRepository {
	return &DeliveriesRepository{db: db, exec: db}
}

func NewDeliveriesRepositoryTx(tx *sql.Tx) *DeliveriesRepository {
	return &DeliveriesRepository{exec: tx}
}

func (r *DeliveriesRepository) CreatePendingIfNotExists(ctx context.Context, contactID, itemID string, availableAt time.Time) (bool, string, error) {
	var id string
	err := r.exec.QueryRowContext(ctx, `
		INSERT INTO deliveries (contact_id, item_id, status, next_retry_at)
		VALUES ($1, $2, 'pending', $3)
		ON CONFLICT (contact_id, item_id) DO NOTHING
		RETURNING id
	`, contactID, itemID, availableAt).Scan(&id)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, id, nil
}

func (r *DeliveriesRepository) ClaimBatch(ctx context.Context, deliveryID string, now time.Time) (domain.Contact, domain.Feed, []domain.DeliveryWithItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	contact, feed, err := r.loadContactFeed(ctx, tx, deliveryID)
	if err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}

	claimIDs, err := r.selectClaimIDs(ctx, tx, contact.ID, feed.ID, now)
	if err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}
	if len(claimIDs) == 0 {
		return contact, feed, nil, tx.Commit()
	}

	query, args := buildIDsUpdate(`
		UPDATE deliveries
		SET status='in_progress', updated_at=CURRENT_TIMESTAMP
		WHERE id IN (%s)
	`, claimIDs)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}

	deliveries, err := r.loadDeliveries(ctx, tx, claimIDs)
	if err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}

	itemsByID, err := r.loadItems(ctx, tx, claimIDsToItemIDs(deliveries))
	if err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}

	batch := make([]domain.DeliveryWithItem, 0, len(deliveries))
	for _, d := range deliveries {
		it, ok := itemsByID[d.ItemID]
		if !ok {
			return domain.Contact{}, domain.Feed{}, nil, fmt.Errorf("item %s not found for delivery %s", d.ItemID, d.ID)
		}
		batch = append(batch, domain.DeliveryWithItem{Delivery: d, Item: it})
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, domain.Feed{}, nil, err
	}
	return contact, feed, batch, nil
}

func (r *DeliveriesRepository) RecoverInProgress(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.exec.ExecContext(ctx, `
		UPDATE deliveries
		SET status='failed', last_error=$1, next_retry_at=$2, updated_at=CURRENT_TIMESTAMP
		WHERE status='in_progress'
	`, "delivery recovered after process restart", now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *DeliveriesRepository) MarkSent(ctx context.Context, deliveryID string, sentAt time.Time) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE deliveries SET status='sent', sent_at=$2, next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
	`, deliveryID, sentAt)
	return err
}

func (r *DeliveriesRepository) MarkManySent(ctx context.Context, deliveryIDs []string, sentAt time.Time) error {
	if len(deliveryIDs) == 0 {
		return nil
	}
	query, args := buildIDsUpdate(`
		UPDATE deliveries
		SET status='sent', sent_at=$1, next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id IN (%s)
	`, sentAt, deliveryIDs)
	_, err := r.exec.ExecContext(ctx, query, args...)
	return err
}

func (r *DeliveriesRepository) MarkFailed(ctx context.Context, deliveryID string, errMsg string, nextRetryAt *time.Time) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE deliveries
		SET status='failed', attempt_count=attempt_count+1, last_error=$2, last_attempt_at=CURRENT_TIMESTAMP,
		    next_retry_at=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
	`, deliveryID, errMsg, nextRetryAt)
	return err
}

func (r *DeliveriesRepository) MarkManyFailed(ctx context.Context, deliveryIDs []string, errMsg string, nextRetryAt *time.Time) error {
	if len(deliveryIDs) == 0 {
		return nil
	}
	query, args := buildIDsUpdate(`
		UPDATE deliveries
		SET status='failed', attempt_count=attempt_count+1, last_error=$1, last_attempt_at=CURRENT_TIMESTAMP,
		    next_retry_at=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id IN (%s)
	`, errMsg, nextRetryAt, deliveryIDs)
	_, err := r.exec.ExecContext(ctx, query, args...)
	return err
}

func (r *DeliveriesRepository) GetByID(ctx context.Context, id string) (domain.Delivery, error) {
	row := r.exec.QueryRowContext(ctx, `
		SELECT id, item_id, contact_id, status, attempt_count, last_error, last_attempt_at, sent_at, next_retry_at, created_at, updated_at
		FROM deliveries WHERE id=$1
	`, id)
	return scanDelivery(row)
}

func (r *DeliveriesRepository) ListRetryDue(ctx context.Context, now time.Time, limit int, maxAttempts int) ([]string, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT id
		FROM deliveries
		WHERE status IN ('pending', 'failed')
		  AND next_retry_at IS NOT NULL
		  AND next_retry_at <= $1
		  AND (status = 'pending' OR attempt_count < $2)
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

func (r *DeliveriesRepository) loadContactFeed(ctx context.Context, tx *sql.Tx, deliveryID string) (domain.Contact, domain.Feed, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			c.id, c.type, c.status, c.value, c.display_name, c.verified_at, c.created_at, c.updated_at,
			f.id, f.url, f.name, f.enabled, f.interval_seconds, f.batch_enabled, f.batch_window_seconds,
			f.etag, f.last_modified, f.last_fetch_at, f.next_fetch_at, f.initialized_at, f.created_at, f.updated_at
		FROM deliveries d
		JOIN contacts c ON c.id = d.contact_id
		JOIN items i ON i.id = d.item_id
		JOIN feeds f ON f.id = i.feed_id
		WHERE d.id = $1
	`, deliveryID)

	var contact domain.Contact
	var feed domain.Feed
	var contactType string
	var contactStatus string
	var contactDisplay sql.NullString
	var contactVerified sql.NullTime
	var feedETag, feedLastModified sql.NullString
	var feedLastFetch, feedNextFetch, feedInit sql.NullTime
	if err := row.Scan(
		&contact.ID, &contactType, &contactStatus, &contact.Value, &contactDisplay, &contactVerified, &contact.CreatedAt, &contact.UpdatedAt,
		&feed.ID, &feed.URL, &feed.Name, &feed.Enabled, &feed.IntervalSeconds, &feed.BatchEnabled, &feed.BatchWindowSecs,
		&feedETag, &feedLastModified, &feedLastFetch, &feedNextFetch, &feedInit, &feed.CreatedAt, &feed.UpdatedAt,
	); err != nil {
		return domain.Contact{}, domain.Feed{}, err
	}
	contact.Type = domain.ContactType(contactType)
	contact.Status = domain.ContactStatus(contactStatus)
	if contactDisplay.Valid {
		contact.DisplayName = &contactDisplay.String
	}
	if contactVerified.Valid {
		t := contactVerified.Time
		contact.VerifiedAt = &t
	}
	if feedETag.Valid {
		feed.ETag = &feedETag.String
	}
	if feedLastModified.Valid {
		feed.LastModified = &feedLastModified.String
	}
	if feedLastFetch.Valid {
		t := feedLastFetch.Time
		feed.LastFetchAt = &t
	}
	if feedNextFetch.Valid {
		t := feedNextFetch.Time
		feed.NextFetchAt = &t
	}
	if feedInit.Valid {
		t := feedInit.Time
		feed.InitializedAt = &t
	}
	return contact, feed, nil
}

func (r *DeliveriesRepository) selectClaimIDs(ctx context.Context, tx *sql.Tx, contactID string, feedID string, now time.Time) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT d.id
		FROM deliveries d
		JOIN items i ON i.id = d.item_id
		WHERE d.contact_id = $1
		  AND i.feed_id = $2
		  AND d.status IN ('pending', 'failed')
		  AND d.next_retry_at IS NOT NULL
		  AND d.next_retry_at <= $3
		ORDER BY d.created_at ASC
	`, contactID, feedID, now)
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

func (r *DeliveriesRepository) loadDeliveries(ctx context.Context, tx *sql.Tx, deliveryIDs []string) ([]domain.Delivery, error) {
	query, args := buildIDsSelect(`
		SELECT id, item_id, contact_id, status, attempt_count, last_error, last_attempt_at, sent_at, next_retry_at, created_at, updated_at
		FROM deliveries
		WHERE id IN (%s)
	`, deliveryIDs)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[string]domain.Delivery, len(deliveryIDs))
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		byID[d.ID] = d
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]domain.Delivery, 0, len(deliveryIDs))
	for _, id := range deliveryIDs {
		d, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("delivery %s not found after claim", id)
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *DeliveriesRepository) loadItems(ctx context.Context, tx *sql.Tx, itemIDs []string) (map[string]domain.Item, error) {
	query, args := buildIDsSelect(`
		SELECT id, feed_id, external_id, uniq_key, title, link, summary, author, published_at, created_at
		FROM items
		WHERE id IN (%s)
	`, itemIDs)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemsByID := make(map[string]domain.Item, len(itemIDs))
	for rows.Next() {
		var it domain.Item
		var ext, sum, auth sql.NullString
		var pub sql.NullTime
		if err := rows.Scan(&it.ID, &it.FeedID, &ext, &it.UniqKey, &it.Title, &it.Link, &sum, &auth, &pub, &it.CreatedAt); err != nil {
			return nil, err
		}
		if ext.Valid {
			it.ExternalID = &ext.String
		}
		if sum.Valid {
			it.Summary = &sum.String
		}
		if auth.Valid {
			it.Author = &auth.String
		}
		if pub.Valid {
			t := pub.Time
			it.PublishedAt = &t
		}
		itemsByID[it.ID] = it
	}
	return itemsByID, rows.Err()
}

func scanDelivery(scanner interface{ Scan(dest ...any) error }) (domain.Delivery, error) {
	var d domain.Delivery
	var st string
	var lastErr sql.NullString
	var lastAttempt, sentAt, nextRetry sql.NullTime

	if err := scanner.Scan(&d.ID, &d.ItemID, &d.ContactID, &st, &d.AttemptCount, &lastErr, &lastAttempt, &sentAt, &nextRetry, &d.CreatedAt, &d.UpdatedAt); err != nil {
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

func claimIDsToItemIDs(deliveries []domain.Delivery) []string {
	itemIDs := make([]string, 0, len(deliveries))
	for _, d := range deliveries {
		itemIDs = append(itemIDs, d.ItemID)
	}
	return itemIDs
}

func buildIDsSelect(pattern string, ids []string) (string, []any) {
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}
	return fmt.Sprintf(pattern, strings.Join(placeholders, ",")), args
}

func buildIDsUpdate(pattern string, prefixArgs ...any) (string, []any) {
	ids := prefixArgs[len(prefixArgs)-1].([]string)
	baseArgs := prefixArgs[:len(prefixArgs)-1]
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(baseArgs)+len(ids))
	args = append(args, baseArgs...)
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(baseArgs)+i+1))
		args = append(args, id)
	}
	return fmt.Sprintf(pattern, strings.Join(placeholders, ",")), args
}
