package persistence

import (
	"context"
	"encoding/json"
	"order-service/internal/domain/order"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateOrderWithOutbox: Создает заказ + запись в Outbox атомарно
func (r *Repository) CreateOrderWithOutbox(ctx context.Context, ord *order.Order, event order.PaymentRequestEvent) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//Сохраняем заказ
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (id, user_id, amount, status, created_at) VALUES ($1, $2, $3, $4, $5)",
		ord.ID, ord.UserID, ord.Amount, ord.Status, time.Now())
	if err != nil {
		return err
	}

	//Сохраняем сообщение в Outbox
	payload, _ := json.Marshal(event)
	_, err = tx.ExecContext(ctx,
		"INSERT INTO outbox (topic, key, payload) VALUES ($1, $2, $3)",
		"pay-order", ord.ID, payload)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateStatus: Обновляет статус заказа (вызывается Kafka Consumer'ом)
func (r *Repository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = $1 WHERE id = $2",
		status, orderID)
	return err
}

func (r *Repository) GetByUserID(ctx context.Context, userID int64) ([]order.Order, error) {
	var orders []order.Order
	err := r.db.SelectContext(ctx, &orders,
		"SELECT id, user_id, amount, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC",
		userID)
	return orders, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	var ord order.Order
	err := r.db.GetContext(ctx, &ord, "SELECT id, user_id, amount, status, created_at FROM orders WHERE id = $1", id)
	return &ord, err
}

// Outbox методы
type OutboxMsg struct {
	ID      int64  `db:"id"`
	Topic   string `db:"topic"`
	Key     string `db:"key"`
	Payload []byte `db:"payload"`
}

func (r *Repository) FetchUnprocessedOutbox(ctx context.Context, limit int) ([]OutboxMsg, error) {
	var msgs []OutboxMsg
	err := r.db.SelectContext(ctx, &msgs,
		"SELECT id, topic, key, payload FROM outbox WHERE processed = FALSE LIMIT $1 FOR UPDATE SKIP LOCKED",
		limit)
	return msgs, err
}

func (r *Repository) MarkOutboxProcessed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox SET processed = TRUE WHERE id = $1", id)
	return err
}
