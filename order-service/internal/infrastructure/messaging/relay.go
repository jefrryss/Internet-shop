package messaging

import (
	"context"
	"log"
	"order-service/internal/infrastructure/persistence"
	"time"

	"github.com/IBM/sarama"
)

type OutboxRelay struct {
	repo     *persistence.Repository
	producer sarama.SyncProducer
}

func NewOutboxRelay(repo *persistence.Repository, producer sarama.SyncProducer) *OutboxRelay {
	return &OutboxRelay{repo: repo, producer: producer}
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *OutboxRelay) processBatch(ctx context.Context) {
	msgs, err := r.repo.FetchUnprocessedOutbox(ctx, 50)
	if err != nil {
		log.Printf("Outbox fetch error: %v", err)
		return
	}

	for _, msg := range msgs {
		// Отправляем в Kafka
		_, _, err := r.producer.SendMessage(&sarama.ProducerMessage{
			Topic: msg.Topic,
			Key:   sarama.StringEncoder(msg.Key),
			Value: sarama.ByteEncoder(msg.Payload),
		})

		if err != nil {
			log.Printf("Kafka produce error for msg %d: %v", msg.ID, err)
			continue
		}

		// Помечаем как обработанное в БД
		if err := r.repo.MarkOutboxProcessed(ctx, msg.ID); err != nil {
			log.Printf("DB update error for msg %d: %v", msg.ID, err)
		}
	}
}
