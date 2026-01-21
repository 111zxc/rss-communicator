package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type ContactsRepository struct {
	db *sql.DB
}

func NewContactsRepository(db *sql.DB) *ContactsRepository {
	return &ContactsRepository{db: db}
}

func (r *ContactsRepository) UpsertTelegramActive(
	ctx context.Context,
	chatID string,
	username *string,
	displayName *string,
	verifiedAt time.Time,
) (domain.Contact, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var c domain.Contact

	row := tx.QueryRowContext(ctx, `
		INSERT INTO contacts (type, status, value, display_name, verified_at)
		VALUES ($1, 'active', $2, $3, $4)
		ON CONFLICT (type, value) DO UPDATE SET
			status = EXCLUDED.status,
			display_name = COALESCE(EXCLUDED.display_name, contacts.display_name),
			verified_at = COALESCE(EXCLUDED.verified_at, contacts.verified_at),
			updated_at = now()
		RETURNING
			id,
			type,
			status,
			value,
			display_name,
			verified_at,
			created_at,
			updated_at
	`, string(domain.ContactTelegram), chatID, displayName, verifiedAt)

	var (
		typ   string
		stat  string
		disp  sql.NullString
		verif sql.NullTime
	)

	if err := row.Scan(
		&c.ID,
		&typ,
		&stat,
		&c.Value,
		&disp,
		&verif,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return domain.Contact{}, err
	}
	c.Type = domain.ContactType(typ)
	c.Status = domain.ContactStatus(stat)
	if disp.Valid {
		c.DisplayName = &disp.String
	}
	if verif.Valid {
		t := verif.Time
		c.VerifiedAt = &t
	}

	if username != nil {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO contact_telegram_config (contact_id, username)
			VALUES ($1, $2)
			ON CONFLICT (contact_id) DO UPDATE SET username = EXCLUDED.username
		`, c.ID, *username)
		if err != nil {
			return domain.Contact{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return c, nil
}

func (r *ContactsRepository) GetByID(ctx context.Context, id string) (domain.Contact, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, status, value, display_name, verified_at, created_at, updated_at
		FROM contacts
		WHERE id=$1
	`, id)

	var c domain.Contact
	var typ, stat string
	var disp sql.NullString
	var verif sql.NullTime

	if err := row.Scan(&c.ID, &typ, &stat, &c.Value, &disp, &verif, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return domain.Contact{}, err
	}
	c.Type = domain.ContactType(typ)
	c.Status = domain.ContactStatus(stat)
	if disp.Valid {
		c.DisplayName = &disp.String
	}
	if verif.Valid {
		t := verif.Time
		c.VerifiedAt = &t
	}
	return c, nil
}

func (r *ContactsRepository) List(ctx context.Context, limit, offset int) ([]domain.Contact, int, error) {
	const q = `
SELECT id, type, value, username, display_name, status,
       verified_at, created_at, updated_at
FROM contacts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	contacts := make([]domain.Contact, 0)
	for rows.Next() {
		var c domain.Contact
		if err := rows.Scan(
			&c.ID,
			&c.Type,
			&c.Value,
			&c.DisplayName,
			&c.Status,
			&c.VerifiedAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM contacts`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *ContactsRepository) GetByTypeValue(ctx context.Context, typ domain.ContactType, value string) (domain.Contact, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, status, value, display_name, verified_at, created_at, updated_at
		FROM contacts
		WHERE type = $1 AND value = $2
	`, string(typ), value)

	var c domain.Contact
	var t, stat string
	var disp sql.NullString
	var verif sql.NullTime

	if err := row.Scan(
		&c.ID,
		&t,
		&stat,
		&c.Value,
		&disp,
		&verif,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return domain.Contact{}, err
	}

	c.Type = domain.ContactType(t)
	c.Status = domain.ContactStatus(stat)

	if disp.Valid {
		c.DisplayName = &disp.String
	}
	if verif.Valid {
		tm := verif.Time
		c.VerifiedAt = &tm
	}

	return c, nil
}
