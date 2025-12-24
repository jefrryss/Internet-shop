package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"payment-service/internal/application/command"
	"payment-service/internal/config"
	httpHandler "payment-service/internal/infrastructure/http"
	"payment-service/internal/infrastructure/messaging"
	"payment-service/internal/infrastructure/persistence"

	"github.com/IBM/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	//Подключение к БД с ретраями
	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.Open("postgres", cfg.DBDSN)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			log.Println("Успешное подключение к БД")
			break
		}
		log.Printf("БД недоступна (%v). Попытка %d/10 через 3 сек...", err, i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД после всех попыток: %v", err)
	}
	defer db.Close()

	//Инициализация слоев
	accountRepo := persistence.NewAccountRepository(db)
	inboxRepo := persistence.NewInboxRepository(db)
	outboxRepo := persistence.NewOutboxRepository(db)
	txRepo := persistence.NewTransactionRepository(db)

	handler := &command.Handler{
		DB:          db,
		AccountRepo: accountRepo,
		InboxRepo:   inboxRepo,
		TxRepo:      txRepo,
		OutboxRepo:  outboxRepo,
	}

	ctx, cancel := context.WithCancel(context.Background())

	//Kafka
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaCfg.Producer.Return.Successes = true

	//ПОДКЛЮЧЕНИЕ CONSUMER
	var group sarama.ConsumerGroup
	for {
		group, err = sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, kafkaCfg)
		if err == nil {
			log.Println("Kafka Consumer подключен!")
			break
		}
		log.Printf("Kafka Consumer недоступна (%v). Ждем 5 сек...", err)
		time.Sleep(5 * time.Second)
	}

	consumer := messaging.NewConsumer(handler)
	go func() {
		for {
			if err := group.Consume(ctx, []string{"pay-order"}, consumer); err != nil {
				if err == context.Canceled {
					break
				}
				log.Printf("Ошибка потребления Kafka: %v. Рестарт через 5 сек...", err)
				time.Sleep(5 * time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	//ПОДКЛЮЧЕНИЕ PRODUCER
	var producer sarama.SyncProducer
	for {
		producer, err = sarama.NewSyncProducer(cfg.KafkaBrokers, kafkaCfg)
		if err == nil {
			log.Println("Kafka Producer подключен!")
			break
		}
		log.Printf("Kafka Producer недоступен (%v). Ждем 5 сек...", err)
		time.Sleep(5 * time.Second)
	}

	publisher := messaging.NewPublisher(producer, outboxRepo)
	go func() {
		publisher.Run(ctx)
	}()

	//HTTP Setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	accHandler := httpHandler.NewAccountHandler(accountRepo)
	r.Route("/accounts", accHandler.Register)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}

	go func() {
		log.Println("Payments Service запущен на порту", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Остановка сервиса...")
	cancel()

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	server.Shutdown(ctxShutdown)
	producer.Close()
	group.Close()
	log.Println("Сервис остановлен")
}
