package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"

	"payment-service/internal/application/command"
)

type Consumer struct {
	handler *command.Handler
}

func NewConsumer(handler *command.Handler) *Consumer {
	return &Consumer{handler: handler}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {

	for msg := range claim.Messages() {
		var cmd command.PayOrderCommand

		if err := json.Unmarshal(msg.Value, &cmd); err != nil {
			log.Printf("Ошибка декодирования Kafka сообщения: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		cmd.MessageID = string(msg.Key)

		err := c.handler.Handle(context.Background(), cmd)
		if err != nil {
			log.Printf("Ошибка обработки заказа %s: %v", cmd.OrderID, err)
			continue
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
