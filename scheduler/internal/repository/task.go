package repository

import (
	"context"
	"errors"
	"fmt"
	"scheduler/internal/entity"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository struct {
	logger *zap.SugaredLogger
	db     *pgx.Conn
}

func NewTaskRepository(logger *zap.SugaredLogger, db *pgx.Conn) *TaskRepository {
	return &TaskRepository{
		logger: logger,
		db:     db,
	}
}

func (r *TaskRepository) GetTasksByUserID(ctx context.Context, userID int64, completed *bool) ([]entity.Task, error) {
	const op = "internal.repository.task.GetTasksByUserID"

	query := `SELECT id, title, description, user_id, completed, completed_at, created_at, updated_at 
              FROM tasks 
              WHERE user_id = $1`
	args := []any{userID}

	if completed != nil {
		query += " AND completed = $2"
		args = append(args, *completed)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.UserID,
			&task.Completed,
			&task.CompletedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tasks, nil
}

func (r *TaskRepository) GetTasksCompletedLastDay(ctx context.Context, userID int64) ([]entity.Task, error) {
	const op = "internal.repository.task.GetTasksCompletedLastDay"

	query := `SELECT id, title, description, user_id, completed, completed_at, created_at, updated_at 
              FROM tasks 
              WHERE user_id = $1 
                AND completed = true 
                AND completed_at >= NOW() - INTERVAL '24 hours'
              ORDER BY completed_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.UserID,
			&task.Completed,
			&task.CompletedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tasks, nil
}
