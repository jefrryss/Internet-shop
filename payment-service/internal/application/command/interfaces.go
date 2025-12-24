package command

import (
	"context"
	"database/sql"
)

// AccountRepository для списания денег
type AccountRepository interface {
	WithdrawInTx(ctx context.Context, tx *sql.Tx, userID int64, amount int64) error
}

// InboxRepository для идемпотентности
type InboxRepository interface {
	Exists(ctx context.Context, tx *sql.Tx, messageID string) bool
	Save(ctx context.Context, tx *sql.Tx, messageID string) error
}

// TxRepository для exactly-once по заказу
type TxRepository interface {
	ExistsByOrderID(ctx context.Context, tx *sql.Tx, orderID string) bool
	Save(ctx context.Context, tx *sql.Tx, orderID string, userID int64, amount int64) error
}

// OutboxRepository для сохранения событий
type OutboxRepository interface {
	SaveSuccess(ctx context.Context, tx *sql.Tx, orderID string, userID int64) error
	SaveFailed(ctx context.Context, tx *sql.Tx, orderID string, userID int64, reason string) error
}
