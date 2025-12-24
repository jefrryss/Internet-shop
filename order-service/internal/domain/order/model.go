package order

import "time"

// Статусы заказа
const (
	StatusPending = "PENDING"
	StatusPaid    = "PAID"
	StatusFailed  = "FAILED"
)

// Модель заказа
type Order struct {
	ID        string    `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Amount    int64     `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Событие для отправки в Payment Service
type PaymentRequestEvent struct {
	MessageID string `json:"message_id"`
	OrderID   string `json:"order_id"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
}

// Событие, которое мы ждем ОТ Payment Service
type PaymentResultEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
