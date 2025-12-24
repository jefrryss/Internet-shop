package persistence

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) ExistsByOrderID(ctx context.Context, tx *sql.Tx, orderID string) bool {
	var cnt int
	_ = tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM payment_transactions WHERE order_id = $1", orderID).Scan(&cnt)
	return cnt > 0
}

func (r *TransactionRepository) Save(ctx context.Context, tx *sql.Tx, orderID string, userID int64, amount int64) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO payment_transactions(order_id, user_id, amount) VALUES ($1, $2, $3)",
		orderID, userID, amount,
	)
	return err
}
