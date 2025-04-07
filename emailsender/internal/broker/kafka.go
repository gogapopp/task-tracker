package broker

import (
	"context"
	"emailsender/internal/entity"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EmailProcessor interface {
	ProcessEmail(ctx context.Context, msg entity.EmailMessage) error
}

type Consumer struct {
	logger         *zap.SugaredLogger
	reader         *kafka.Reader
	emailTopic     string
	emailProcessor EmailProcessor
}

func NewConsumer(
	logger *zap.SugaredLogger,
	brokers string,
	emailTopic string,
	groupID string,
	emailProcessor EmailProcessor,
) (*Consumer, error) {
	logger.Infof("connecting to Kafka brokers: %v, topic: %s, group: %s", brokers, emailTopic, groupID)

	brokersList := strings.Split(brokers, ",")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokersList,
		Topic:           emailTopic,
		GroupID:         groupID,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
		CommitInterval:  time.Second,
	})

	logger.Infof("consumer initialized for topic %s", emailTopic)

	return &Consumer{
		logger:         logger,
		reader:         reader,
		emailTopic:     emailTopic,
		emailProcessor: emailProcessor,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	const op = "internal.broker.kafka.Start"

	c.logger.Infof("%s: starting to consume messages from topic %s", op, c.emailTopic)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping consumer")
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				c.logger.Errorf("%s: error fetching message: %w", op, err)
				continue
			}

			c.logger.Infof("%s: received message from partition %d at offset %d",
				op, msg.Partition, msg.Offset)

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Errorf("%s: error processing message: %w", op, err)
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Errorf("%s: error committing message: %w", op, err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	const op = "internal.broker.kafka.processMessage"

	var emailMsg entity.EmailMessage
	if err := json.Unmarshal(msg.Value, &emailMsg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Infof("%s: processing email message to %s with subject: %s",
		op, emailMsg.To, emailMsg.Subject)

	return c.emailProcessor.ProcessEmail(ctx, emailMsg)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
