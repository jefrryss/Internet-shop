package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "order-service/internal/infrastructure/http"
	"order-service/internal/infrastructure/messaging"
	"order-service/internal/infrastructure/persistence"

	"github.com/IBM/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	//Конфигурация
	dbDSN := os.Getenv("DB_DSN")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaBrokers := []string{kafkaBroker}

	log.Println("Запуск Order Service...")
	log.Printf("DB: %s", dbDSN)
	log.Printf("Kafka: %s", kafkaBrokers)

	// 2. Подключение к БД
	var db *sqlx.DB
	var err error
	for i := 0; i < 15; i++ {
		db, err = sqlx.Connect("postgres", dbDSN)
		if err == nil {
			log.Println("Успешное подключение к БД")
			break
		}
		log.Printf("БД недоступна (%v). Ждем 2 сек... (%d/15)", err, i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	// Инициализация репозитория
	repo := persistence.NewRepository(db)

	// Kafka
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	//Producer
	var producer sarama.SyncProducer
	for i := 0; i < 15; i++ {
		producer, err = sarama.NewSyncProducer(kafkaBrokers, config)
		if err == nil {
			log.Println("Kafka Producer подключен")
			break
		}
		log.Printf("Kafka Producer недоступен (%v). Ждем 3 сек... (%d/15)", err, i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось создать Kafka Producer: %v", err)
	}

	// Запуск Outbox
	ctx, cancel := context.WithCancel(context.Background())
	relay := messaging.NewOutboxRelay(repo, producer)
	go relay.Run(ctx)

	//Consumer
	var consumerGroup sarama.ConsumerGroup
	for i := 0; i < 15; i++ {
		consumerGroup, err = sarama.NewConsumerGroup(kafkaBrokers, "orders-service-group", config)
		if err == nil {
			log.Println("Kafka Consumer Group создана")
			break
		}
		log.Printf("Kafka Consumer недоступен (%v). Ждем 3 сек... (%d/15)", err, i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось создать Kafka Consumer: %v", err)
	}

	//Запуск чтения топика в отдельной горутине
	go func() {
		h := messaging.NewResultConsumer(repo)
		for {

			if err := consumerGroup.Consume(ctx, []string{"payment-results"}, h); err != nil {
				log.Printf("Ошибка Consumer: %v. Рестарт через 5 сек...", err)
				time.Sleep(5 * time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// HTTP Сервер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := httpHandler.NewHandler(repo)

	// Хендлер регистрируется по пути /orders
	r.Route("/orders", h.Register)

	srv := &http.Server{Addr: ":8081", Handler: r}

	// Запуск сервера
	go func() {
		log.Println("Order Service запущен на порту :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Ошибка при остановке HTTP: %v", err)
	}
	producer.Close()
	consumerGroup.Close()
	log.Println("Сервис остановлен корректно")
}
