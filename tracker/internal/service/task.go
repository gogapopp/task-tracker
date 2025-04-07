package service

import (
	"context"
	"errors"
	"fmt"
	"tracker/internal/entity"
	"tracker/internal/repository"
)

type TaskStorage interface {
	CreateTask(ctx context.Context, task entity.TaskCreateRequest, userID int64) (int64, error)
	GetTaskByID(ctx context.Context, taskID, userID int64) (*entity.Task, error)
	GetTasksByUserID(ctx context.Context, userID int64, completed *bool) ([]entity.Task, error)
	UpdateTask(ctx context.Context, taskID, userID int64, task entity.TaskUpdateRequest) error
	DeleteTask(ctx context.Context, taskID, userID int64) error
}

type TaskService struct {
	taskStorage TaskStorage
}

func NewTaskService(taskStorage TaskStorage) *TaskService {
	return &TaskService{
		taskStorage: taskStorage,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req entity.TaskCreateRequest, userID int64) (*entity.TaskResponse, error) {
	const op = "internal.service.task.CreateTask"

	taskID, err := s.taskStorage.CreateTask(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	task, err := s.taskStorage.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID, userID int64) (*entity.TaskResponse, error) {
	const op = "internal.service.task.GetTaskByID"

	task, err := s.taskStorage.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *TaskService) GetTasks(ctx context.Context, userID int64, completed *bool) (*entity.TaskListResponse, error) {
	const op = "internal.service.task.GetTasks"

	tasks, err := s.taskStorage.GetTasksByUserID(ctx, userID, completed)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var response entity.TaskListResponse
	response.Tasks = make([]entity.TaskResponse, len(tasks))
	for i, task := range tasks {
		response.Tasks[i] = entity.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Completed:   task.Completed,
			CompletedAt: task.CompletedAt,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		}
	}

	return &response, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID int64, req entity.TaskUpdateRequest) (*entity.TaskResponse, error) {
	const op = "internal.service.task.UpdateTask"

	err := s.taskStorage.UpdateTask(ctx, taskID, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, fmt.Errorf("%s: %w", op, repository.ErrTaskNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	task, err := s.taskStorage.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID int64) error {
	const op = "internal.service.task.DeleteTask"

	err := s.taskStorage.DeleteTask(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return fmt.Errorf("%s: %w", op, repository.ErrTaskNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
