package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"tracker/internal/api"
	"tracker/internal/api/handlers"
	"tracker/internal/broker"
	"tracker/internal/config"
	"tracker/internal/repository"
	"tracker/internal/service"

	"go.uber.org/zap"
)

const envPath = ".env"

func main() {
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbpool, err := repository.NewPool(ctx, cfg)
	if err != nil {
		sugar.Fatal(err)
	}
	defer dbpool.Close()

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

	userRepo := repository.NewUserStorage(dbpool, cfg.PassSecret)
	taskRepo := repository.NewTaskStorage(dbpool)

	userService := service.NewUserService(userRepo, cfg.JWT.Secret, kafkaProducer)
	taskService := service.NewTaskService(taskRepo)

	userHandlers := handlers.NewUserHandler(sugar, userService)
	taskHandlers := handlers.NewTaskHandler(sugar, taskService)

	routes := api.Router(userHandlers, taskHandlers, sugar, cfg.JWT.Secret)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      routes,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.ShutdownTimeout,
	}

	go func() {
		sugar.Infof("starting HTTP server on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("HTTP server error: %w", err)
		}
	}()

	<-ctx.Done()
	sugar.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		sugar.Fatalf("HTTP server shutdown error: %w", err)
	}

	sugar.Info("shutdown completed")
}
