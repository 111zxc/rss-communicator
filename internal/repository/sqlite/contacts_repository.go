package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
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
			updated_at = CURRENT_TIMESTAMP
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
	c.Telegram = &domain.TelegramContactConfig{Username: username}
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

func (r *ContactsRepository) CreateTelegram(
	ctx context.Context,
	chatID string,
	username *string,
	displayName *string,
	status domain.ContactStatus,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	contact, err := upsertTelegramContact(ctx, tx, "", chatID, username, displayName, status, verifiedAt)
	if err != nil {
		return domain.Contact{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return contact, nil
}

func (r *ContactsRepository) UpdateTelegram(
	ctx context.Context,
	contactID string,
	chatID string,
	username *string,
	displayName *string,
	status domain.ContactStatus,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	contact, err := upsertTelegramContact(ctx, tx, contactID, chatID, username, displayName, status, verifiedAt)
	if err != nil {
		return domain.Contact{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return contact, nil
}

func (r *ContactsRepository) CreateEmail(
	ctx context.Context,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.EmailContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	return r.createOrUpdateEmail(ctx, "", value, displayName, status, cfg, verifiedAt)
}

func (r *ContactsRepository) UpdateEmail(
	ctx context.Context,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.EmailContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	return r.createOrUpdateEmail(ctx, contactID, value, displayName, status, cfg, verifiedAt)
}

func (r *ContactsRepository) CreateHTTP(
	ctx context.Context,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.HTTPContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	return r.createOrUpdateHTTP(ctx, "", value, displayName, status, cfg, verifiedAt)
}

func (r *ContactsRepository) UpdateHTTP(
	ctx context.Context,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.HTTPContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	return r.createOrUpdateHTTP(ctx, contactID, value, displayName, status, cfg, verifiedAt)
}

func (r *ContactsRepository) createOrUpdateEmail(
	ctx context.Context,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.EmailContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	contact, err := upsertEmailContact(ctx, tx, contactID, value, displayName, status, verifiedAt)
	if err != nil {
		return domain.Contact{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_email_config (contact_id, format)
		VALUES ($1, $2)
		ON CONFLICT (contact_id) DO UPDATE SET format = EXCLUDED.format
	`, contact.ID, cfg.Format)
	if err != nil {
		return domain.Contact{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}

	contact.Email = &domain.EmailContactConfig{Format: cfg.Format}
	return contact, nil
}

func (r *ContactsRepository) createOrUpdateHTTP(
	ctx context.Context,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	cfg domain.HTTPContactConfig,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	contact, err := upsertHTTPContact(ctx, tx, contactID, value, displayName, status, verifiedAt)
	if err != nil {
		return domain.Contact{}, err
	}

	headersJSON, err := json.Marshal(cfg.Headers)
	if err != nil {
		return domain.Contact{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_http_config (contact_id, method, url, headers_json, body_template)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (contact_id) DO UPDATE SET
			method = EXCLUDED.method,
			url = EXCLUDED.url,
			headers_json = EXCLUDED.headers_json,
			body_template = EXCLUDED.body_template
	`, contact.ID, cfg.Method, cfg.URL, string(headersJSON), cfg.BodyTemplate)
	if err != nil {
		return domain.Contact{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}

	contact.HTTP = &domain.HTTPContactConfig{
		Method:       cfg.Method,
		URL:          cfg.URL,
		Headers:      cloneHeaders(cfg.Headers),
		BodyTemplate: cfg.BodyTemplate,
	}
	return contact, nil
}

func (r *ContactsRepository) GetEmailConfig(ctx context.Context, contactID string) (domain.EmailContactConfig, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT format
		FROM contact_email_config
		WHERE contact_id = $1
	`, contactID)

	var cfg domain.EmailContactConfig
	var format sql.NullString
	if err := row.Scan(&format); err != nil {
		return domain.EmailContactConfig{}, err
	}
	cfg.Format = "plain"
	if format.Valid && format.String != "" {
		cfg.Format = format.String
	}
	return cfg, nil
}

func (r *ContactsRepository) GetHTTPConfig(ctx context.Context, contactID string) (domain.HTTPContactConfig, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT method, url, headers_json, body_template
		FROM contact_http_config
		WHERE contact_id = $1
	`, contactID)

	var cfg domain.HTTPContactConfig
	var headersJSON []byte
	var body sql.NullString
	if err := row.Scan(&cfg.Method, &cfg.URL, &headersJSON, &body); err != nil {
		return domain.HTTPContactConfig{}, err
	}
	if len(headersJSON) == 0 {
		cfg.Headers = map[string]string{}
	} else if err := json.Unmarshal(headersJSON, &cfg.Headers); err != nil {
		return domain.HTTPContactConfig{}, err
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if body.Valid {
		cfg.BodyTemplate = &body.String
	}
	return cfg, nil
}

func (r *ContactsRepository) GetByID(ctx context.Context, id string) (domain.Contact, error) {
	cs, _, err := r.listByQuery(ctx, `
		SELECT c.id, c.type, c.status, c.value, c.display_name, c.verified_at, c.created_at, c.updated_at,
		       tt.username, ee.format, hh.method, hh.url, hh.headers_json, hh.body_template
		FROM contacts c
		LEFT JOIN contact_telegram_config tt ON tt.contact_id = c.id
		LEFT JOIN contact_email_config ee ON ee.contact_id = c.id
		LEFT JOIN contact_http_config hh ON hh.contact_id = c.id
		WHERE c.id = $1
	`, id)
	if err != nil {
		return domain.Contact{}, err
	}
	if len(cs) == 0 {
		return domain.Contact{}, sql.ErrNoRows
	}
	return cs[0], nil
}

func (r *ContactsRepository) List(ctx context.Context, limit, offset int) ([]domain.Contact, int, error) {
	const q = `
SELECT c.id, c.type, c.status, c.value, c.display_name, c.verified_at, c.created_at, c.updated_at,
       tt.username, ee.format, hh.method, hh.url, hh.headers_json, hh.body_template
FROM contacts c
LEFT JOIN contact_telegram_config tt ON tt.contact_id = c.id
LEFT JOIN contact_email_config ee ON ee.contact_id = c.id
LEFT JOIN contact_http_config hh ON hh.contact_id = c.id
ORDER BY c.created_at DESC
LIMIT $1 OFFSET $2
`
	contacts, _, err := r.listByQuery(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM contacts`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *ContactsRepository) GetByTypeValue(ctx context.Context, typ domain.ContactType, value string) (domain.Contact, error) {
	cs, _, err := r.listByQuery(ctx, `
		SELECT c.id, c.type, c.status, c.value, c.display_name, c.verified_at, c.created_at, c.updated_at,
		       tt.username, ee.format, hh.method, hh.url, hh.headers_json, hh.body_template
		FROM contacts c
		LEFT JOIN contact_telegram_config tt ON tt.contact_id = c.id
		LEFT JOIN contact_email_config ee ON ee.contact_id = c.id
		LEFT JOIN contact_http_config hh ON hh.contact_id = c.id
		WHERE c.type = $1 AND c.value = $2
	`, string(typ), value)
	if err != nil {
		return domain.Contact{}, err
	}
	if len(cs) == 0 {
		return domain.Contact{}, sql.ErrNoRows
	}
	return cs[0], nil
}

func (r *ContactsRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM contacts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ContactsRepository) listByQuery(ctx context.Context, q string, args ...any) ([]domain.Contact, int, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	contacts := make([]domain.Contact, 0)
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return contacts, len(contacts), nil
}

func scanContact(scanner interface {
	Scan(dest ...any) error
}) (domain.Contact, error) {
	var c domain.Contact
	var typ, stat string
	var disp, username, emailFormat, method, targetURL, body sql.NullString
	var verif sql.NullTime
	var headersJSON []byte

	if err := scanner.Scan(
		&c.ID,
		&typ,
		&stat,
		&c.Value,
		&disp,
		&verif,
		&c.CreatedAt,
		&c.UpdatedAt,
		&username,
		&emailFormat,
		&method,
		&targetURL,
		&headersJSON,
		&body,
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
	if username.Valid {
		c.Telegram = &domain.TelegramContactConfig{Username: &username.String}
	}
	if emailFormat.Valid || c.Type == domain.ContactEmail {
		format := "plain"
		if emailFormat.Valid && emailFormat.String != "" {
			format = emailFormat.String
		}
		c.Email = &domain.EmailContactConfig{Format: format}
	}
	if method.Valid && targetURL.Valid {
		c.HTTP = &domain.HTTPContactConfig{
			Method:  method.String,
			URL:     targetURL.String,
			Headers: map[string]string{},
		}
		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &c.HTTP.Headers); err != nil {
				return domain.Contact{}, err
			}
		}
		if c.HTTP.Headers == nil {
			c.HTTP.Headers = map[string]string{}
		}
		if body.Valid {
			c.HTTP.BodyTemplate = &body.String
		}
	}

	return c, nil
}

func cloneHeaders(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func upsertTelegramContact(
	ctx context.Context,
	tx *sql.Tx,
	contactID string,
	chatID string,
	username *string,
	displayName *string,
	status domain.ContactStatus,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	var row *sql.Row
	if contactID == "" {
		row = tx.QueryRowContext(ctx, `
			INSERT INTO contacts (type, status, value, display_name, verified_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, string(domain.ContactTelegram), string(status), chatID, displayName, verifiedAt)
	} else {
		row = tx.QueryRowContext(ctx, `
			UPDATE contacts
			SET value = $2, display_name = $3, status = $4, verified_at = $5, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND type = $6
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, contactID, chatID, displayName, string(status), verifiedAt, string(domain.ContactTelegram))
	}

	contact, err := scanBaseContact(row)
	if err != nil {
		return domain.Contact{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_telegram_config (contact_id, username)
		VALUES ($1, $2)
		ON CONFLICT (contact_id) DO UPDATE SET username = EXCLUDED.username
	`, contact.ID, username)
	if err != nil {
		return domain.Contact{}, err
	}
	contact.Telegram = &domain.TelegramContactConfig{Username: username}
	return contact, nil
}

func upsertHTTPContact(
	ctx context.Context,
	tx *sql.Tx,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	var row *sql.Row
	if contactID == "" {
		row = tx.QueryRowContext(ctx, `
			INSERT INTO contacts (type, status, value, display_name, verified_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, string(domain.ContactHTTP), string(status), value, displayName, verifiedAt)
	} else {
		row = tx.QueryRowContext(ctx, `
			UPDATE contacts
			SET value = $2, display_name = $3, status = $4, verified_at = $5, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND type = $6
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, contactID, value, displayName, string(status), verifiedAt, string(domain.ContactHTTP))
	}
	return scanBaseContact(row)
}

func upsertEmailContact(
	ctx context.Context,
	tx *sql.Tx,
	contactID string,
	value string,
	displayName *string,
	status domain.ContactStatus,
	verifiedAt *time.Time,
) (domain.Contact, error) {
	var row *sql.Row
	if contactID == "" {
		row = tx.QueryRowContext(ctx, `
			INSERT INTO contacts (type, status, value, display_name, verified_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, string(domain.ContactEmail), string(status), value, displayName, verifiedAt)
	} else {
		row = tx.QueryRowContext(ctx, `
			UPDATE contacts
			SET value = $2, display_name = $3, status = $4, verified_at = $5, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND type = $6
			RETURNING id, type, status, value, display_name, verified_at, created_at, updated_at
		`, contactID, value, displayName, string(status), verifiedAt, string(domain.ContactEmail))
	}
	return scanBaseContact(row)
}

func scanBaseContact(row *sql.Row) (domain.Contact, error) {
	var c domain.Contact
	var typ, stat string
	var disp sql.NullString
	var verif sql.NullTime

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
	return c, nil
}
