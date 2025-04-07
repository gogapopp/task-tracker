package main

import (
	"context"
	"emailsender/internal/broker"
	"emailsender/internal/config"
	"emailsender/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	mailClient, err := service.NewSMTPClient(
		sugar,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
		cfg.SMTP.FromEmail,
		"task tracker",
	)
	if err != nil {
		sugar.Fatal(err)
	}

	emailService := service.NewEmailService(sugar, mailClient)

	consumer, err := broker.NewConsumer(
		sugar,
		cfg.Kafka.Brokers,
		cfg.Kafka.EmailTopic,
		cfg.Kafka.GroupID,
		emailService,
	)
	if err != nil {
		sugar.Fatal(err)
	}

	go func() {
		if err := consumer.Start(ctx); err != nil {
			sugar.Fatalf("failed to start consumer: %w", err)
		}
	}()

	sugar.Info("email sender service started successfully")

	<-ctx.Done()
	sugar.Info("shutdown signal received")

	if err := mailClient.Close(); err != nil {
		sugar.Errorf("error closing SMTP client: %w", err)
	}

	if err := consumer.Close(); err != nil {
		sugar.Errorf("error closing consumer: %w", err)
	}

	sugar.Info("shutdown completed")
}
