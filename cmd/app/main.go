package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"

	"github.com/andrestrejos/fast-track-ingestion/internal/amqp"
	"github.com/andrestrejos/fast-track-ingestion/internal/db"
	"github.com/andrestrejos/fast-track-ingestion/internal/domain"
	"github.com/andrestrejos/fast-track-ingestion/internal/ingest"
	"github.com/andrestrejos/fast-track-ingestion/internal/util"
)

func mustEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def != "" {
		return def
	}
	log.Fatalf("missing env: %s", key)
	return ""
}

func main() {
	// Carga .env si existe (no falla si no está)
	_ = godotenv.Load()

	// --- ENV ---
	mysqlHost := mustEnv("MYSQL_HOST", "127.0.0.1")
	mysqlPort, _ := strconv.Atoi(mustEnv("MYSQL_PORT", "3306"))
	mysqlUser := mustEnv("MYSQL_USER", "root")
	mysqlPass := mustEnv("MYSQL_PASSWORD", "secret")
	mysqlDB := mustEnv("MYSQL_DB", "fastdb")

	rabbitURL := mustEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	exchange := mustEnv("RABBIT_EXCHANGE", "payments")
	queue := mustEnv("RABBIT_QUEUE", "payment_events_q")
	rk := mustEnv("RABBIT_ROUTING_KEY", "payment.event")

	// Contexto con cancel por señales (Ctrl+C)
	ctx, cancel := util.WithSignals(context.Background())
	defer cancel()

	// --- MySQL ---
	mysql, err := db.NewMySQL(mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDB)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer mysql.Close()

	repo := db.NewRepo(mysql.DB)
	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// --- RabbitMQ ---
	amqpCli, err := amqp.New(rabbitURL)
	if err != nil {
		log.Fatalf("amqp: %v", err)
	}
	defer amqpCli.Close()

	q, err := amqpCli.Declare(exchange, queue, rk)
	if err != nil {
		log.Fatalf("declare: %v", err)
	}

	// Canal para contar mensajes procesados (3 válidos + 1 duplicado)
	processed := make(chan struct{}, 4)

	// Consumer
	consumer := &ingest.Consumer{
		AMQP:      amqpCli,
		Repo:      repo,
		Queue:     q.Name,
		Processed: processed, // <- el consumer debe enviar una señal por cada mensaje ACKed
	}
	stopConsume := make(chan struct{})
	log.Println("consumer starting on queue:", q.Name)
	go func() {
		if err := consumer.Run(ctx, stopConsume); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()

	// Publisher (3 eventos) con sincronización
	publisher := &ingest.Publisher{AMQP: amqpCli, Exchange: exchange, RoutingKey: rk}
	var wg sync.WaitGroup
	done := make(chan struct{})
	publisher.Run(ctx, &wg, done)

	// Espera a que se publiquen los 3 eventos iniciales
	<-done

	// Publica el duplicado para forzar PK violation -> irá a skipped_messages
	dup := domain.PaymentEvent{UserID: 1, PaymentID: 1, DepositAmount: 10}
	b, _ := json.Marshal(dup)
	if err := amqpCli.Publish(ctx, exchange, rk, b); err != nil {
		log.Printf("publish dup err: %v", err)
	}

	// Espera a que termine la goroutine del publisher
	wg.Wait()

	// Espera hasta que el consumer procese exactamente 4 mensajes
	for i := 0; i < 4; i++ {
		<-processed
	}

	// Cierre ordenado del consumer
	close(stopConsume)
	log.Println("processed 4 messages; shutting down gracefully")
}
