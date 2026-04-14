package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
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
		VALUES ($1, $2::jsonb, $3)
	`, topic, string(body), availableAt)
	return err
}

func (r *OutboxRepository) ClaimBatch(ctx context.Context, now time.Time, leaseUntil time.Time, limit int) ([]repository.OutboxMessage, error) {
	rows, err := r.exec.QueryContext(ctx, `
		WITH claimed AS (
			SELECT id
			FROM outbox
			WHERE status IN ('pending', 'failed')
			  AND available_at <= $1
			ORDER BY created_at ASC
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox o
		SET available_at=$3, updated_at=now()
		FROM claimed
		WHERE o.id = claimed.id
		RETURNING o.id, o.topic, o.payload::text, o.attempt_count
	`, now, limit, leaseUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []repository.OutboxMessage
	for rows.Next() {
		var msg repository.OutboxMessage
		var payload string
		if err := rows.Scan(&msg.ID, &msg.Topic, &payload, &msg.AttemptCount); err != nil {
			return nil, err
		}
		msg.Payload = []byte(payload)
		out = append(out, msg)
	}
	return out, rows.Err()
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE outbox
		SET status='published', last_error=NULL, updated_at=now()
		WHERE id=$1
	`, id)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id string, errMsg string, nextAvailableAt time.Time) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE outbox
		SET status='failed', attempt_count=attempt_count+1, last_error=$2, available_at=$3, updated_at=now()
		WHERE id=$1
	`, id, errMsg, nextAvailableAt)
	return err
}
