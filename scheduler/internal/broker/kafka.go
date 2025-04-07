package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"scheduler/internal/entity"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	logger            *zap.SugaredLogger
	writer            *kafka.Writer
	topicEmailSending string
}

func NewProducer(logger *zap.SugaredLogger, brokers []string, topicEmailSending string) (*Producer, error) {
	logger.Infof("connecting to Kafka brokers: %v, topic: %s", brokers, topicEmailSending)

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topicEmailSending,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		WriteTimeout: 10 * time.Second,
		Async:        false,
	}

	logger.Infof("producer initialized for topic %s", topicEmailSending)

	return &Producer{
		logger:            logger,
		writer:            writer,
		topicEmailSending: topicEmailSending,
	}, nil
}

func (p *Producer) SendDailyStats(
	ctx context.Context,
	email string,
	completedCount int,
	pendingCount int,
	completedTaskTitles string,
	pendingTaskTitles string,
) error {
	const op = "internal.broker.kafka.SendDailyStats"

	var subject string
	body := "Here are your stats for the last 24 hours."

	if completedCount > 0 && pendingCount > 0 {
		subject = fmt.Sprintf("TaskTracker: %d tasks completed, %d tasks pending", completedCount, pendingCount)
	} else if completedCount > 0 {
		subject = fmt.Sprintf("TaskTracker: You've completed %d tasks today!", completedCount)
	} else if pendingCount > 0 {
		subject = fmt.Sprintf("TaskTracker: You have %d pending tasks", pendingCount)
	} else {
		subject = "TaskTracker: Your Daily Task Report"
	}

	msg := entity.EmailMessage{
		Type:    "daily_stats",
		To:      email,
		Subject: subject,
		Body:    body,
		Variables: map[string]string{
			"email":          email,
			"completedCount": fmt.Sprintf("%d", completedCount),
			"pendingCount":   fmt.Sprintf("%d", pendingCount),
			"completedTasks": completedTaskTitles,
			"pendingTasks":   pendingTaskTitles,
		},
	}

	return p.sendMessage(ctx, msg)
}

func (p *Producer) sendMessage(ctx context.Context, msg entity.EmailMessage) error {
	const op = "internal.broker.kafka.sendMessage"

	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal message: %w", op, err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.To),
		Value: value,
		Time:  time.Now(),
	}

	p.logger.Infof("%s: attempting to send message to topic '%s'", op, p.topicEmailSending)

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		p.logger.Errorf("%s: failed to write message: %+v", op, err)
		return fmt.Errorf("%s: failed to write message: %w", op, err)
	}

	p.logger.Infof("%s: email queued for %s", op, msg.To)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
