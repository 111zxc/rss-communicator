package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type FeedsRepository struct{ db dbtx }

func NewFeedsRepository(db *sql.DB) *FeedsRepository   { return &FeedsRepository{db: db} }
func NewFeedsRepositoryTx(tx *sql.Tx) *FeedsRepository { return &FeedsRepository{db: tx} }

func (r *FeedsRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Feed, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, url, name, enabled, interval_seconds, batch_enabled, batch_window_seconds, etag, last_modified, last_fetch_at, next_fetch_at, initialized_at
		FROM feeds
		WHERE enabled = true AND (next_fetch_at IS NULL OR next_fetch_at <= $1)
		ORDER BY COALESCE(next_fetch_at, to_timestamp(0)) ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Feed
	for rows.Next() {
		var f domain.Feed
		var etag, lm sql.NullString
		var lastFetch, nextFetch, initAt sql.NullTime

		if err := rows.Scan(&f.ID, &f.URL, &f.Name, &f.Enabled, &f.IntervalSeconds, &f.BatchEnabled, &f.BatchWindowSecs, &etag, &lm, &lastFetch, &nextFetch, &initAt); err != nil {
			return nil, err
		}
		if etag.Valid {
			f.ETag = &etag.String
		}
		if lm.Valid {
			f.LastModified = &lm.String
		}
		if lastFetch.Valid {
			t := lastFetch.Time
			f.LastFetchAt = &t
		}
		if nextFetch.Valid {
			t := nextFetch.Time
			f.NextFetchAt = &t
		}
		if initAt.Valid {
			t := initAt.Time
			f.InitializedAt = &t
		}

		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeedsRepository) ClaimDue(ctx context.Context, now time.Time, nextAt time.Time, limit int) ([]domain.Feed, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH due AS (
			SELECT id
			FROM feeds
			WHERE enabled = true AND (next_fetch_at IS NULL OR next_fetch_at <= $1)
			ORDER BY COALESCE(next_fetch_at, to_timestamp(0)) ASC
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE feeds f
		SET next_fetch_at = $3, updated_at = now()
		FROM due
		WHERE f.id = due.id
		RETURNING f.id, f.url, f.name, f.enabled, f.interval_seconds, f.batch_enabled, f.batch_window_seconds,
		          f.etag, f.last_modified, f.last_fetch_at, f.next_fetch_at, f.initialized_at
	`, now, limit, nextAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Feed
	for rows.Next() {
		var f domain.Feed
		var etag, lm sql.NullString
		var lastFetch, nextFetch, initAt sql.NullTime

		if err := rows.Scan(&f.ID, &f.URL, &f.Name, &f.Enabled, &f.IntervalSeconds, &f.BatchEnabled, &f.BatchWindowSecs, &etag, &lm, &lastFetch, &nextFetch, &initAt); err != nil {
			return nil, err
		}
		if etag.Valid {
			f.ETag = &etag.String
		}
		if lm.Valid {
			f.LastModified = &lm.String
		}
		if lastFetch.Valid {
			t := lastFetch.Time
			f.LastFetchAt = &t
		}
		if nextFetch.Valid {
			t := nextFetch.Time
			f.NextFetchAt = &t
		}
		if initAt.Valid {
			t := initAt.Time
			f.InitializedAt = &t
		}

		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeedsRepository) MarkFetched(ctx context.Context, feedID string, fetchedAt, nextAt time.Time, etag, lastModified *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE feeds
		SET last_fetch_at=$2, next_fetch_at=$3, etag=$4, last_modified=$5, last_error=NULL, error_count=0, updated_at=now()
		WHERE id=$1
	`, feedID, fetchedAt, nextAt, etag, lastModified)
	return err
}

func (r *FeedsRepository) MarkFetchError(ctx context.Context, feedID string, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE feeds
		SET last_error=$2, error_count=error_count+1, updated_at=now()
		WHERE id=$1
	`, feedID, errMsg)
	return err
}

func (r *FeedsRepository) MarkInitialized(ctx context.Context, feedID string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE feeds SET initialized_at=$2, updated_at=now()
		WHERE id=$1 AND initialized_at IS NULL
	`, feedID, at)
	return err
}

func (r *FeedsRepository) GetByID(ctx context.Context, id string) (domain.Feed, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, url, name, enabled, interval_seconds, batch_enabled, batch_window_seconds, etag, last_modified, last_fetch_at, next_fetch_at, initialized_at
		FROM feeds WHERE id=$1
	`, id)

	var f domain.Feed
	var etag, lm sql.NullString
	var lastFetch, nextFetch, initAt sql.NullTime

	if err := row.Scan(&f.ID, &f.URL, &f.Name, &f.Enabled, &f.IntervalSeconds, &f.BatchEnabled, &f.BatchWindowSecs, &etag, &lm, &lastFetch, &nextFetch, &initAt); err != nil {
		return domain.Feed{}, err
	}
	if etag.Valid {
		f.ETag = &etag.String
	}
	if lm.Valid {
		f.LastModified = &lm.String
	}
	if lastFetch.Valid {
		t := lastFetch.Time
		f.LastFetchAt = &t
	}
	if nextFetch.Valid {
		t := nextFetch.Time
		f.NextFetchAt = &t
	}
	if initAt.Valid {
		t := initAt.Time
		f.InitializedAt = &t
	}
	return f, nil
}

func (r *FeedsRepository) UpdateBatching(ctx context.Context, feedID string, batchEnabled bool, batchWindowSecs int) (domain.Feed, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE feeds
		SET batch_enabled=$2, batch_window_seconds=$3, updated_at=now()
		WHERE id=$1
		RETURNING id, url, name, enabled, interval_seconds, batch_enabled, batch_window_seconds, etag, last_modified, last_fetch_at, next_fetch_at, initialized_at, created_at, updated_at
	`, feedID, batchEnabled, batchWindowSecs)

	var out domain.Feed
	var etag, lm sql.NullString
	var lastFetch, nextFetch, initAt sql.NullTime
	if err := row.Scan(
		&out.ID,
		&out.URL,
		&out.Name,
		&out.Enabled,
		&out.IntervalSeconds,
		&out.BatchEnabled,
		&out.BatchWindowSecs,
		&etag,
		&lm,
		&lastFetch,
		&nextFetch,
		&initAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return domain.Feed{}, err
	}
	if etag.Valid {
		out.ETag = &etag.String
	}
	if lm.Valid {
		out.LastModified = &lm.String
	}
	if lastFetch.Valid {
		t := lastFetch.Time
		out.LastFetchAt = &t
	}
	if nextFetch.Valid {
		t := nextFetch.Time
		out.NextFetchAt = &t
	}
	if initAt.Valid {
		t := initAt.Time
		out.InitializedAt = &t
	}
	return out, nil
}

func (r *FeedsRepository) List(ctx context.Context, limit, offset int) ([]domain.Feed, int, error) {
	const q = `
SELECT id, name, url, enabled, interval_seconds,
       batch_enabled, batch_window_seconds,
       next_fetch_at, last_fetch_at, initialized_at,
       created_at, updated_at
FROM feeds
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	feeds := make([]domain.Feed, 0)
	for rows.Next() {
		var f domain.Feed
		if err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.URL,
			&f.Enabled,
			&f.IntervalSeconds,
			&f.BatchEnabled,
			&f.BatchWindowSecs,
			&f.NextFetchAt,
			&f.LastFetchAt,
			&f.InitializedAt,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		feeds = append(feeds, f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM feeds`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return feeds, total, nil
}

func (r *FeedsRepository) Delete(ctx context.Context, feedID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM feeds WHERE id = $1`, feedID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *FeedsRepository) Create(ctx context.Context, f domain.Feed) (domain.Feed, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO feeds (url, name, enabled, interval_seconds, batch_enabled, batch_window_seconds)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, url, name, enabled, interval_seconds, batch_enabled, batch_window_seconds,
		          etag, last_modified, last_fetch_at, next_fetch_at, initialized_at,
		          created_at, updated_at
	`,
		f.URL,
		f.Name,
		f.Enabled,
		f.IntervalSeconds,
		f.BatchEnabled,
		f.BatchWindowSecs,
	)

	var out domain.Feed
	var etag, lm sql.NullString
	var lastFetch, nextFetch, initAt sql.NullTime

	if err := row.Scan(
		&out.ID,
		&out.URL,
		&out.Name,
		&out.Enabled,
		&out.IntervalSeconds,
		&out.BatchEnabled,
		&out.BatchWindowSecs,
		&etag,
		&lm,
		&lastFetch,
		&nextFetch,
		&initAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return domain.Feed{}, err
	}

	if etag.Valid {
		out.ETag = &etag.String
	}
	if lm.Valid {
		out.LastModified = &lm.String
	}
	if lastFetch.Valid {
		t := lastFetch.Time
		out.LastFetchAt = &t
	}
	if nextFetch.Valid {
		t := nextFetch.Time
		out.NextFetchAt = &t
	}
	if initAt.Valid {
		t := initAt.Time
		out.InitializedAt = &t
	}

	return out, nil
}
