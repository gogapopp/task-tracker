package service

import (
	"context"
	"emailsender/internal/entity"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type EmailService struct {
	logger *zap.SugaredLogger
	mailer *SMTPClient
}

func NewEmailService(logger *zap.SugaredLogger, mailer *SMTPClient) *EmailService {
	return &EmailService{
		logger: logger,
		mailer: mailer,
	}
}

func (s *EmailService) ProcessEmail(ctx context.Context, msg entity.EmailMessage) error {
	const op = "internal.service.email_service.ProcessEmail"

	s.logger.Infof("%s: process mail type '%s' for %s", op, msg.Type, msg.To)

	var htmlContent string

	recipientName := msg.To
	if name, exists := msg.Variables["name"]; exists && name != "" {
		recipientName = name
	} else if email, exists := msg.Variables["email"]; exists && email != "" {
		recipientName = email
	}

	switch msg.Type {
	case "daily_stats":
		htmlContent = s.generateDailyStatsEmail(msg)
	default:
		htmlContent = msg.Body
	}

	err := s.mailer.SendEmail(
		ctx,
		msg.To,
		recipientName,
		msg.Subject,
		msg.Body,
		htmlContent,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *EmailService) generateDailyStatsEmail(msg entity.EmailMessage) string {
	userName := "there"
	if name, exists := msg.Variables["name"]; exists && name != "" {
		userName = name
	}

	completedCount := msg.Variables["completedCount"]
	pendingCount := msg.Variables["pendingCount"]
	completedTasks := msg.Variables["completedTasks"]
	pendingTasks := msg.Variables["pendingTasks"]

	var completedTasksHTML string
	if completedTasks != "" {
		tasks := strings.Split(completedTasks, ",")
		completedTasksHTML = "<ul>"
		for _, task := range tasks {
			completedTasksHTML += fmt.Sprintf("<li>%s</li>", strings.TrimSpace(task))
		}
		completedTasksHTML += "</ul>"
	} else {
		completedTasksHTML = "<p>No tasks completed in the last 24 hours.</p>"
	}

	var pendingTasksHTML string
	if pendingTasks != "" {
		tasks := strings.Split(pendingTasks, ",")
		pendingTasksHTML = "<ul>"
		for _, task := range tasks {
			pendingTasksHTML += fmt.Sprintf("<li>%s</li>", strings.TrimSpace(task))
		}
		pendingTasksHTML += "</ul>"
	} else {
		pendingTasksHTML = "<p>No pending tasks.</p>"
	}

	return fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				h1 { color: #2c3e50; }
				h2 { color: #3498db; }
				.stats { background-color: #f9f9f9; padding: 15px; border-radius: 5px; margin: 20px 0; }
				.footer { margin-top: 30px; font-size: 12px; color: #7f8c8d; text-align: center; }
				ul { padding-left: 20px; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Hello %s!</h1>
				<p>Here's your daily task summary:</p>
				
				<div class="stats">
					<p><strong>Completed tasks:</strong> %s</p>
					<p><strong>Pending tasks:</strong> %s</p>
				</div>
				
				<h2>Completed Tasks</h2>
				%s
				
				<h2>Pending Tasks</h2>
				%s
				
				<div class="footer">
					<p>This is an automated message from TaskTracker. Please do not reply.</p>
				</div>
			</div>
		</body>
		</html>
	`, userName, completedCount, pendingCount, completedTasksHTML, pendingTasksHTML)
}
