package messaging

import (
	"encoding/json"
	"log"
	"order-service/internal/domain/order"
	"order-service/internal/infrastructure/persistence"

	"github.com/IBM/sarama"
)

type ResultConsumer struct {
	repo *persistence.Repository
}

func NewResultConsumer(repo *persistence.Repository) *ResultConsumer {
	return &ResultConsumer{repo: repo}
}

func (c *ResultConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *ResultConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c *ResultConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event order.PaymentResultEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Ошибка декодирования ответа: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		// Определяем новый статус
		newStatus := order.StatusFailed
		if event.Status == "success" {
			newStatus = order.StatusPaid
		}

		log.Printf("Получен ответ по заказу %s: %s", event.OrderID, newStatus)

		// Обновляем статус в БД
		if err := c.repo.UpdateStatus(sess.Context(), event.OrderID, newStatus); err != nil {
			log.Printf("Ошибка обновления статуса заказа: %v", err)
			continue
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}
