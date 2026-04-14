package sqlite

import (
	"context"
	"database/sql"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type GroupsRepository struct{ db *sql.DB }

func NewGroupsRepository(db *sql.DB) *GroupsRepository { return &GroupsRepository{db: db} }

func (r *GroupsRepository) Create(ctx context.Context, g domain.Group) (domain.Group, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO groups (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, g.Name, g.Description)
	return scanGroup(row)
}

func (r *GroupsRepository) Update(ctx context.Context, groupID string, name string, description *string) (domain.Group, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE groups
		SET name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at
	`, groupID, name, description)
	return scanGroup(row)
}

func (r *GroupsRepository) GetByID(ctx context.Context, groupID string) (domain.Group, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM groups WHERE id = $1
	`, groupID)
	return scanGroup(row)
}

func (r *GroupsRepository) List(ctx context.Context, limit, offset int) ([]domain.Group, int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM groups
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *GroupsRepository) Delete(ctx context.Context, groupID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM groups WHERE id = $1`, groupID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *GroupsRepository) ListContacts(ctx context.Context, groupID string) ([]domain.Contact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.type, c.status, c.value, c.display_name, c.verified_at, c.created_at, c.updated_at,
		       tt.username, hh.method, hh.url, hh.headers_json, hh.body_template
		FROM group_contacts gc
		JOIN contacts c ON c.id = gc.contact_id
		LEFT JOIN contact_telegram_config tt ON tt.contact_id = c.id
		LEFT JOIN contact_http_config hh ON hh.contact_id = c.id
		WHERE gc.group_id = $1
		ORDER BY gc.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []domain.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

func (r *GroupsRepository) AddContact(ctx context.Context, groupID, contactID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_contacts (group_id, contact_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, groupID, contactID)
	return err
}

func (r *GroupsRepository) RemoveContact(ctx context.Context, groupID, contactID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM group_contacts
		WHERE group_id = $1 AND contact_id = $2
	`, groupID, contactID)
	return err
}

func (r *GroupsRepository) ListFeeds(ctx context.Context, groupID string) ([]domain.Feed, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.id, f.name, f.url, f.enabled, f.interval_seconds,
		       f.batch_enabled, f.batch_window_seconds,
		       f.next_fetch_at, f.last_fetch_at, f.initialized_at,
		       f.created_at, f.updated_at
		FROM group_feeds gf
		JOIN feeds f ON f.id = gf.feed_id
		WHERE gf.group_id = $1
		ORDER BY gf.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []domain.Feed
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
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (r *GroupsRepository) AddFeed(ctx context.Context, groupID, feedID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_feeds (group_id, feed_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, groupID, feedID)
	return err
}

func (r *GroupsRepository) RemoveFeed(ctx context.Context, groupID, feedID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM group_feeds
		WHERE group_id = $1 AND feed_id = $2
	`, groupID, feedID)
	return err
}

func (r *GroupsRepository) ListFeedIDs(ctx context.Context, groupID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT feed_id FROM group_feeds WHERE group_id = $1
	`, groupID)
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

func (r *GroupsRepository) ListContactIDs(ctx context.Context, groupID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT contact_id FROM group_contacts WHERE group_id = $1
	`, groupID)
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

func scanGroup(scanner interface{ Scan(dest ...any) error }) (domain.Group, error) {
	var g domain.Group
	var desc sql.NullString
	if err := scanner.Scan(&g.ID, &g.Name, &desc, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return domain.Group{}, err
	}
	if desc.Valid {
		g.Description = &desc.String
	}
	return g, nil
}
