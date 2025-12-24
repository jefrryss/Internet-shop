package account

type Account struct {
	ID      int64 `db:"id" json:"id"`
	UserID  int64 `db:"user_id" json:"user_id"`
	Balance int64 `db:"balance" json:"balance"`
}

func NewAccount(userID int64) *Account {
	return &Account{
		UserID:  userID,
		Balance: 0,
	}
}

// Deposit пополняет баланс
func (a *Account) Deposit(amount int64) {
	a.Balance += amount
}

// Withdraw списывает средства
func (a *Account) Withdraw(amount int64) error {
	if a.Balance < amount {
		return ErrInsufficientFunds
	}
	a.Balance -= amount
	return nil
}
