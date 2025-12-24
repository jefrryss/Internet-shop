package messaging

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type OutboxMessage struct {
	ID      string
	Type    string
	Payload []byte
}

type OutboxStorage interface {
	GetUnsent(ctx context.Context) ([]OutboxMessage, error)
	MarkSent(ctx context.Context, id string) error
}

type Publisher struct {
	producer sarama.SyncProducer
	repo     OutboxStorage
}

func NewPublisher(p sarama.SyncProducer, r OutboxStorage) *Publisher {
	return &Publisher{
		producer: p,
		repo:     r,
	}
}

func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка Outbox Publisher...")
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *Publisher) processBatch(ctx context.Context) {
	msgs, err := p.repo.GetUnsent(ctx)
	if err != nil {
		log.Printf("Ошибка чтения Outbox: %v", err)
		return
	}

	for _, m := range msgs {

		_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
			Topic: m.Type,
			Value: sarama.ByteEncoder(m.Payload),
			Key:   sarama.StringEncoder(m.ID),
		})

		if err != nil {
			log.Printf("Ошибка отправки в Kafka (id=%s): %v", m.ID, err)
			continue
		}

		if err := p.repo.MarkSent(ctx, m.ID); err != nil {
			log.Printf("Ошибка обновления статуса Outbox (id=%s): %v", m.ID, err)
		}
	}
}
