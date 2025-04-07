package service

import (
	"context"
	"fmt"
	"scheduler/internal/broker"
	"scheduler/internal/entity"
	"scheduler/internal/repository"
	"strings"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	maxTasksInReport = 5
)

type TaskScheduler struct {
	logger         *zap.SugaredLogger
	cron           *cron.Cron
	userRepo       *repository.UserRepository
	taskRepo       *repository.TaskRepository
	producer       *broker.Producer
	cronExpression string
}

func NewTaskScheduler(
	logger *zap.SugaredLogger,
	userRepo *repository.UserRepository,
	taskRepo *repository.TaskRepository,
	producer *broker.Producer,
	cronExpression string,
) *TaskScheduler {
	if cronExpression == "" {
		cronExpression = "0 0 * * *" // midnight every day
	}

	return &TaskScheduler{
		logger:         logger,
		cron:           cron.New(),
		userRepo:       userRepo,
		taskRepo:       taskRepo,
		producer:       producer,
		cronExpression: cronExpression,
	}
}

func (s *TaskScheduler) Start() error {
	const op = "internal.service.service.Start"

	_, err := s.cron.AddFunc(s.cronExpression, func() {
		ctx := context.Background()
		if err := s.ProcessDailyReports(ctx); err != nil {
			s.logger.Errorf("%s: failed to process daily reports: %w", op, err)
		}
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.cron.Start()
	s.logger.Infof("%s: scheduler started with cron expression: %s", op, s.cronExpression)
	return nil
}

func (s *TaskScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("scheduler stopped")
}

func (s *TaskScheduler) ProcessDailyReports(ctx context.Context) error {
	const op = "internal.service.service.ProcessDailyReports"

	s.logger.Info("starting daily reports processing")

	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to get users: %w", op, err)
	}

	s.logger.Infof("processing reports for %d users", len(users))

	for _, user := range users {
		if err := s.processUserReport(ctx, user); err != nil {
			s.logger.Errorf("%s: failed to process report for user %d: %w", op, user.ID, err)
			continue
		}
	}

	s.logger.Info("daily reports processing completed")
	return nil
}

func (s *TaskScheduler) processUserReport(ctx context.Context, user entity.User) error {
	const op = "internal.service.service.processUserReport"

	completedFalse := false
	pendingTasks, err := s.taskRepo.GetTasksByUserID(ctx, user.ID, &completedFalse)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	completedTasks, err := s.taskRepo.GetTasksCompletedLastDay(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(pendingTasks) == 0 && len(completedTasks) == 0 {
		s.logger.Infof("user %d has no tasks to report", user.ID)
		return nil
	}

	return s.sendTaskReportProducer(ctx, user, pendingTasks, completedTasks)
}

func (s *TaskScheduler) sendTaskReportProducer(ctx context.Context, user entity.User, pendingTasks []entity.Task, completedTasks []entity.Task) error {
	const op = "internal.service.service.sendTaskReportProducer"

	hasPending := len(pendingTasks) > 0
	hasCompleted := len(completedTasks) > 0

	if !hasPending && !hasCompleted {
		return nil
	}

	pendingTitles := []string{}
	if hasPending {
		limit := min(len(pendingTasks), maxTasksInReport)
		for i := 0; i < limit; i++ {
			pendingTitles = append(pendingTitles, pendingTasks[i].Title)
		}
	}

	completedTitles := []string{}
	if hasCompleted {
		limit := min(len(completedTasks), maxTasksInReport)
		for i := 0; i < limit; i++ {
			completedTitles = append(completedTitles, completedTasks[i].Title)
		}
	}

	var pendingStr, completedStr string
	if len(pendingTitles) > 0 {
		pendingStr = strings.Join(pendingTitles, ", ")
	}
	if len(completedTitles) > 0 {
		completedStr = strings.Join(completedTitles, ", ")
	}

	err := s.producer.SendDailyStats(
		ctx,
		user.Email,
		len(completedTasks),
		len(pendingTasks),
		completedStr,
		pendingStr,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to send daily stats email: %w", op, err)
	}

	s.logger.Infof("sent daily report to user %d (%s): %d completed, %d pending tasks",
		user.ID, user.Email, len(completedTasks), len(pendingTasks))

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
