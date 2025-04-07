package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Kafka KafkaConfig `yaml:"kafka"`
	SMTP  SMTPConfig  `yaml:"smtp"`
}

type KafkaConfig struct {
	Brokers    string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	EmailTopic string `env:"KAFKA_TOPIC_EMAIL_SENDING" env-default:"email.send"`
	GroupID    string `env:"KAFKA_GROUP_ID" env-default:"email-sender"`
}

type SMTPConfig struct {
	Host      string `env:"SMTP_HOST" env-default:"smtp.example.com"`
	Port      int    `env:"SMTP_PORT" env-default:"587"`
	User      string `env:"SMTP_USER" env-required:"true"`
	Password  string `env:"SMTP_PASSWORD" env-required:"true"`
	FromEmail string `env:"SMTP_FROM_EMAIL" env-required:"true"`
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
