package payment

import "time"

type Transaction struct {
	ID        int64
	OrderID   string
	UserID    int64
	Amount    int64
	CreatedAt time.Time
}
