package persistence

import (
	"context"
	"database/sql"
	"encoding/json"

	"payment-service/internal/infrastructure/messaging"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) SaveSuccess(ctx context.Context, tx *sql.Tx, orderID string, userID int64) error {

	return r.save(ctx, tx, "payment-results", map[string]any{
		"order_id": orderID,
		"user_id":  userID,
		"status":   "success",
	})
}

func (r *OutboxRepository) SaveFailed(ctx context.Context, tx *sql.Tx, orderID string, userID int64, reason string) error {
	return r.save(ctx, tx, "payment-results", map[string]any{
		"order_id": orderID,
		"user_id":  userID,
		"status":   "failed",
		"reason":   reason,
	})
}

func (r *OutboxRepository) save(ctx context.Context, tx *sql.Tx, topic string, payload any) error {
	data, _ := json.Marshal(payload)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO outbox_messages(id, type, payload, sent, created_at)
		VALUES ($1, $2, $3, false, NOW())
	`, uuid.New().String(), topic, data)
	return err
}

func (r *OutboxRepository) GetUnsent(ctx context.Context) ([]messaging.OutboxMessage, error) {
	var rows []struct {
		ID      string `db:"id"`
		Type    string `db:"type"`
		Payload []byte `db:"payload"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, type, payload FROM outbox_messages 
		WHERE sent = false 
		ORDER BY created_at ASC 
		LIMIT 50 
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return nil, err
	}

	result := make([]messaging.OutboxMessage, 0, len(rows))
	for _, row := range rows {
		result = append(result, messaging.OutboxMessage{
			ID:      row.ID,
			Type:    row.Type,
			Payload: row.Payload,
		})
	}
	return result, nil
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_messages SET sent = true WHERE id = $1", id)
	return err
}
