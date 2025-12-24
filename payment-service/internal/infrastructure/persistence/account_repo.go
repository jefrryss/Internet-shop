package persistence

import (
	"context"
	"database/sql"
	"payment-service/internal/domain/account"

	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create: Вставка нового счета
func (r *AccountRepository) Create(ctx context.Context, acc *account.Account) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO accounts(user_id, balance) VALUES ($1, $2)",
		acc.UserID, acc.Balance)
	return err
}

// GetByUserID: Получение счета
func (r *AccountRepository) GetByUserID(ctx context.Context, userID int64) (*account.Account, error) {
	var acc account.Account
	err := r.db.GetContext(ctx, &acc,
		"SELECT id, user_id, balance FROM accounts WHERE user_id=$1",
		userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

// Deposit: Атомарное пополнение
func (r *AccountRepository) Deposit(ctx context.Context, userID int64, amount int64) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE user_id = $2",
		amount, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return account.ErrAccountNotFound
	}
	return nil
}

// WithdrawInTx: Атомарное списание
func (r *AccountRepository) WithdrawInTx(ctx context.Context, tx *sql.Tx, userID int64, amount int64) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE accounts 
		SET balance = balance - $1 
		WHERE user_id = $2 AND balance >= $1
	`, amount, userID)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		var exists bool
		_ = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE user_id = $1)", userID).Scan(&exists)
		if !exists {
			return account.ErrAccountNotFound
		}
		return account.ErrInsufficientFunds
	}
	return nil
}
