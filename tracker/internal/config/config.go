package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP       HTTPConfig     `yaml:"http"`
	Postgres   PostgresConfig `yaml:"postgres"`
	JWT        JWTConfig      `yaml:"jwt"`
	Kafka      KafkaConfig    `yaml:"kafka"`
	PassSecret string         `env:"PASS_SECRET" env-required:"true"`
}

type HTTPConfig struct {
	Port            string        `env:"BACKEND_HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"5s"`
}

type PostgresConfig struct {
	User     string `env:"POSTGRES_USER" env-default:"user"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"password"`
	Host     string `env:"POSTGRES_HOST" env-default:"postgres"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	Database string `env:"POSTGRES_DB" env-default:"tasktracker"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

type KafkaConfig struct {
	Brokers    string `env:"KAFKA_BROKERS" env-default:"kafka:9092"`
	EmailTopic string `env:"KAFKA_TOPIC_EMAIL_SENDING" env-default:"email.send"`
	GroupID    string `env:"KAFKA_GROUP_ID" env-default:"backend"`
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	skipEnvLoad := os.Getenv("SKIP_ENV_LOAD")
	if skipEnvLoad != "true" {
		err := godotenv.Load(path)
		if err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
	}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
