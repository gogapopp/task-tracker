package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"scheduler/internal/broker"
	"scheduler/internal/config"
	"scheduler/internal/repository"
	"scheduler/internal/service"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const envPath = ".env"

func main() {
	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	cfg, err := config.New(envPath)
	if err != nil {
		sugar.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgx.Connect(ctx, cfg.Postgres.DSN())
	if err != nil {
		sugar.Fatal(err)
	}
	defer db.Close(ctx)

	brokers := strings.Split(cfg.Kafka.Brokers, ",")
	kafkaProducer, err := broker.NewProducer(
		sugar,
		brokers,
		cfg.Kafka.EmailTopic,
	)
	if err != nil {
		sugar.Fatal(err)
	}
	defer kafkaProducer.Close()

	userRepo := repository.NewUserRepository(sugar, db)
	taskRepo := repository.NewTaskRepository(sugar, db)

	taskScheduler := service.NewTaskScheduler(
		sugar,
		userRepo,
		taskRepo,
		kafkaProducer,
		"0 0 * * *", // run at midnight every day
	)

	go func() {
		if err := taskScheduler.Start(); err != nil {
			sugar.Fatalf("failed to start scheduler: %w", err)
		}
	}()

	// uncomment for test
	// sugar.Info("running task processor immediately...")
	// if err := taskScheduler.ProcessDailyReports(context.Background()); err != nil {
	// 	sugar.Errorf("Manual task processing failed: %v", err)
	// }

	<-ctx.Done()
	sugar.Info("shutdown signal received")

	taskScheduler.Stop()

	sugar.Info("shutdown completed")
}
