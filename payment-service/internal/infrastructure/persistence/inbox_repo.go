package persistence

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InboxRepository struct {
	db *sqlx.DB
}

func NewInboxRepository(db *sqlx.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Exists(ctx context.Context, tx *sql.Tx, id string) bool {
	var cnt int
	_ = tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM inbox_messages WHERE message_id = $1", id).Scan(&cnt)
	return cnt > 0
}

func (r *InboxRepository) Save(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO inbox_messages(message_id) VALUES ($1)", id)
	return err
}
