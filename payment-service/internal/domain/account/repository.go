package account

import (
	"context"
	"database/sql"
)

type Repository interface {
	// Синхронные методы
	Create(ctx context.Context, acc *Account) error
	GetByUserID(ctx context.Context, userID int64) (*Account, error)
	Deposit(ctx context.Context, userID int64, amount int64) error

	// Асинхронный метод
	WithdrawInTx(ctx context.Context, tx *sql.Tx, userID int64, amount int64) error
}
