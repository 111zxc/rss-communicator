package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type OutboxRepository struct{ db *sql.DB }

func NewOutboxRepository(db *sql.DB) *OutboxRepository { return &OutboxRepository{db: db} }

func (r *OutboxRepository) Enqueue(ctx context.Context, topic string, payload any, availableAt time.Time) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO outbox (topic, payload, available_at)
		VALUES ($1, $2::jsonb, $3)
	`, topic, string(body), availableAt)
	return err
}
