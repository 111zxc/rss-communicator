package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/repository"
)

type OutboxRepository struct {
	db   *sql.DB
	exec dbtx
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository   { return &OutboxRepository{db: db, exec: db} }
func NewOutboxRepositoryTx(tx *sql.Tx) *OutboxRepository { return &OutboxRepository{exec: tx} }

func (r *OutboxRepository) Enqueue(ctx context.Context, topic string, payload any, availableAt time.Time) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = r.exec.ExecContext(ctx, `
		INSERT INTO outbox (topic, payload, available_at)
		VALUES ($1, $2, $3)
	`, topic, string(body), availableAt)
	return err
}

func (r *OutboxRepository) ClaimBatch(ctx context.Context, now time.Time, leaseUntil time.Time, limit int) ([]repository.OutboxMessage, error) {
	tx, ok := r.exec.(*sql.Tx)
	if !ok {
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = tx.Rollback() }()
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM outbox
		WHERE status IN ('pending', 'failed')
		  AND available_at <= $1
		ORDER BY created_at ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	if len(ids) == 0 {
		return nil, tx.Commit()
	}

	updateQuery, updateArgs := buildOutboxIDsQuery(`
		UPDATE outbox
		SET available_at=$1, updated_at=CURRENT_TIMESTAMP
		WHERE id IN (%s)
		  AND status IN ('pending', 'failed')
		  AND available_at <= $2
	`, leaseUntil, now, ids)
	if _, err := tx.ExecContext(ctx, updateQuery, updateArgs...); err != nil {
		return nil, err
	}

	selectQuery, selectArgs := buildOutboxIDsQuery(`
		SELECT id, topic, payload, attempt_count
		FROM outbox
		WHERE id IN (%s)
		  AND status IN ('pending', 'failed')
		  AND available_at = $1
	`, leaseUntil, ids)
	msgRows, err := tx.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, err
	}
	defer msgRows.Close()

	byID := make(map[string]repository.OutboxMessage, len(ids))
	for msgRows.Next() {
		var msg repository.OutboxMessage
		var payload string
		if err := msgRows.Scan(&msg.ID, &msg.Topic, &payload, &msg.AttemptCount); err != nil {
			return nil, err
		}
		msg.Payload = []byte(payload)
		byID[msg.ID] = msg
	}
	if err := msgRows.Err(); err != nil {
		return nil, err
	}

	out := make([]repository.OutboxMessage, 0, len(ids))
	for _, id := range ids {
		msg, ok := byID[id]
		if ok {
			out = append(out, msg)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE outbox
		SET status='published', last_error=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
	`, id)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id string, errMsg string, nextAvailableAt time.Time) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE outbox
		SET status='failed', attempt_count=attempt_count+1, last_error=$2, available_at=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
	`, id, errMsg, nextAvailableAt)
	return err
}

func buildOutboxIDsQuery(pattern string, prefixArgs ...any) (string, []any) {
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
