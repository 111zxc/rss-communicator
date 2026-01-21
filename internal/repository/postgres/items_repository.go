package postgres

import (
	"context"
	"database/sql"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type ItemsRepository struct{ db *sql.DB }

func NewItemsRepository(db *sql.DB) *ItemsRepository { return &ItemsRepository{db: db} }

func (r *ItemsRepository) InsertMany(ctx context.Context, items []domain.Item) ([]domain.Item, error) {
	if len(items) == 0 {
		return nil, nil
	}

	var inserted []domain.Item
	for _, it := range items {
		var id string
		err := r.db.QueryRowContext(ctx, `
			INSERT INTO items (feed_id, external_id, uniq_key, title, link, summary, author, published_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT DO NOTHING
			RETURNING id
		`, it.FeedID, it.ExternalID, it.UniqKey, it.Title, it.Link, it.Summary, it.Author, it.PublishedAt).Scan(&id)

		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		it.ID = id
		inserted = append(inserted, it)
	}
	return inserted, nil
}

func (r *ItemsRepository) GetByID(ctx context.Context, id string) (domain.Item, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, feed_id, external_id, uniq_key, title, link, summary, author, published_at, created_at
		FROM items WHERE id=$1
	`, id)

	var it domain.Item
	var ext, sum, auth sql.NullString
	var pub sql.NullTime
	if err := row.Scan(&it.ID, &it.FeedID, &ext, &it.UniqKey, &it.Title, &it.Link, &sum, &auth, &pub, &it.CreatedAt); err != nil {
		return domain.Item{}, err
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
	return it, nil
}
