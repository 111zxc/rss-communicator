package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type RegistrationCodesRepository struct{ db *sql.DB }

func NewRegistrationCodesRepository(db *sql.DB) *RegistrationCodesRepository {
	return &RegistrationCodesRepository{db: db}
}

func (r *RegistrationCodesRepository) Create(ctx context.Context, code domain.RegistrationCode) (domain.RegistrationCode, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO registration_codes (code, name, description, enabled, max_uses, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, code, name, description, enabled, max_uses, use_count, expires_at, created_at, updated_at
	`, code.Code, code.Name, code.Description, code.Enabled, code.MaxUses, code.ExpiresAt)
	return scanRegistrationCode(row)
}

func (r *RegistrationCodesRepository) Update(ctx context.Context, codeID string, code string, name string, description *string, enabled bool, maxUses *int, expiresAt *time.Time) (domain.RegistrationCode, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE registration_codes
		SET code = $2, name = $3, description = $4, enabled = $5, max_uses = $6, expires_at = $7, updated_at = now()
		WHERE id = $1
		RETURNING id, code, name, description, enabled, max_uses, use_count, expires_at, created_at, updated_at
	`, codeID, code, name, description, enabled, maxUses, expiresAt)
	return scanRegistrationCode(row)
}

func (r *RegistrationCodesRepository) GetByID(ctx context.Context, codeID string) (domain.RegistrationCode, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, description, enabled, max_uses, use_count, expires_at, created_at, updated_at
		FROM registration_codes
		WHERE id = $1
	`, codeID)
	return scanRegistrationCode(row)
}

func (r *RegistrationCodesRepository) GetByCode(ctx context.Context, code string) (domain.RegistrationCode, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, description, enabled, max_uses, use_count, expires_at, created_at, updated_at
		FROM registration_codes
		WHERE code = $1
	`, code)
	return scanRegistrationCode(row)
}

func (r *RegistrationCodesRepository) List(ctx context.Context, limit, offset int) ([]domain.RegistrationCode, int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, description, enabled, max_uses, use_count, expires_at, created_at, updated_at
		FROM registration_codes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var codes []domain.RegistrationCode
	for rows.Next() {
		rc, err := scanRegistrationCode(rows)
		if err != nil {
			return nil, 0, err
		}
		codes = append(codes, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_codes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return codes, total, nil
}

func (r *RegistrationCodesRepository) Delete(ctx context.Context, codeID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM registration_codes WHERE id = $1`, codeID)
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

func (r *RegistrationCodesRepository) ListGroups(ctx context.Context, codeID string) ([]domain.Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.description, g.created_at, g.updated_at
		FROM registration_code_groups rcg
		JOIN groups g ON g.id = rcg.group_id
		WHERE rcg.registration_code_id = $1
		ORDER BY rcg.created_at DESC
	`, codeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *RegistrationCodesRepository) AddGroup(ctx context.Context, codeID, groupID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO registration_code_groups (registration_code_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, codeID, groupID)
	return err
}

func (r *RegistrationCodesRepository) RemoveGroup(ctx context.Context, codeID, groupID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM registration_code_groups
		WHERE registration_code_id = $1 AND group_id = $2
	`, codeID, groupID)
	return err
}

func (r *RegistrationCodesRepository) IncrementUse(ctx context.Context, codeID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE registration_codes
		SET use_count = use_count + 1, updated_at = now()
		WHERE id = $1
	`, codeID)
	return err
}

func scanRegistrationCode(scanner interface{ Scan(dest ...any) error }) (domain.RegistrationCode, error) {
	var rc domain.RegistrationCode
	var desc sql.NullString
	var maxUses sql.NullInt32
	var expiresAt sql.NullTime
	if err := scanner.Scan(&rc.ID, &rc.Code, &rc.Name, &desc, &rc.Enabled, &maxUses, &rc.UseCount, &expiresAt, &rc.CreatedAt, &rc.UpdatedAt); err != nil {
		return domain.RegistrationCode{}, err
	}
	if desc.Valid {
		rc.Description = &desc.String
	}
	if maxUses.Valid {
		v := int(maxUses.Int32)
		rc.MaxUses = &v
	}
	if expiresAt.Valid {
		t := expiresAt.Time
		rc.ExpiresAt = &t
	}
	return rc, nil
}
