package persistence

import (
	"context"
	"encoding/json"
	"time"

	"order-service/internal/domain/order"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateOrderWithOutbox - Атомарное создание заказа и записи в Outbox
func (r *Repository) CreateOrderWithOutbox(ctx context.Context, ord *order.Order, event order.PaymentRequestEvent) error {
	//Начинаем транзакцию
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Сохраняем заказ
	queryOrder := `INSERT INTO orders (id, user_id, amount, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, queryOrder,
		ord.ID, ord.UserID, ord.Amount, ord.Status, time.Now())
	if err != nil {
		return err
	}

	//Сохраняем событие в Outbox
	payload, _ := json.Marshal(event)
	queryOutbox := `INSERT INTO outbox (topic, key, payload) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, queryOutbox,
		"pay-order", ord.ID, payload)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetByUserID - Получение списка заказов пользователя
func (r *Repository) GetByUserID(ctx context.Context, userID int64) ([]order.Order, error) {
	var orders []order.Order
	err := r.db.SelectContext(ctx, &orders,
		"SELECT id, user_id, amount, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC",
		userID)
	return orders, err
}

// GetByID - Получение одного заказа по ID
func (r *Repository) GetByID(ctx context.Context, orderID string) (*order.Order, error) {
	var ord order.Order
	err := r.db.GetContext(ctx, &ord,
		"SELECT id, user_id, amount, status, created_at FROM orders WHERE id = $1",
		orderID)
	if err != nil {
		return nil, err
	}
	return &ord, nil
}

// UpdateStatus - вызывается консумером при ответе от Payment Service
func (r *Repository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = $1 WHERE id = $2",
		status, orderID)
	return err
}

// OutboxMsg - внутренняя структура для чтения таблицы outbox
type OutboxMsg struct {
	ID      int64  `db:"id"`
	Topic   string `db:"topic"`
	Key     string `db:"key"`
	Payload []byte `db:"payload"`
}

// FetchUnprocessedOutbox - блокирует и выбирает необработанные сообщения
func (r *Repository) FetchUnprocessedOutbox(ctx context.Context, limit int) ([]OutboxMsg, error) {
	var msgs []OutboxMsg
	query := `
		SELECT id, topic, key, payload 
		FROM outbox 
		WHERE processed = FALSE 
		ORDER BY id ASC
		LIMIT $1 
		FOR UPDATE SKIP LOCKED
	`
	err := r.db.SelectContext(ctx, &msgs, query, limit)
	return msgs, err
}

// MarkOutboxProcessed - помечает сообщение как отправленное
func (r *Repository) MarkOutboxProcessed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox SET processed = TRUE WHERE id = $1", id)
	return err
}
