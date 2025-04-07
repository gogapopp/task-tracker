package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tracker/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskStorage struct {
	pool *pgxpool.Pool
}

func NewTaskStorage(pool *pgxpool.Pool) *TaskStorage {
	return &TaskStorage{
		pool: pool,
	}
}

func (s *TaskStorage) CreateTask(ctx context.Context, task entity.TaskCreateRequest, userID int64) (int64, error) {
	const op = "internal.repository.task.CreateTask"

	now := time.Now()

	var taskID int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks(title, description, user_id, created_at, updated_at) 
		 VALUES($1, $2, $3, $4, $5) 
		 RETURNING id`,
		task.Title, task.Description, userID, now, now,
	).Scan(&taskID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return taskID, nil
}

func (s *TaskStorage) GetTaskByID(ctx context.Context, taskID, userID int64) (*entity.Task, error) {
	const op = "internal.repository.task.GetTaskByID"

	var task entity.Task
	err := s.pool.QueryRow(ctx,
		`SELECT id, title, description, user_id, completed, completed_at, created_at, updated_at 
		 FROM tasks 
		 WHERE id = $1 AND user_id = $2`,
		taskID, userID,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.UserID,
		&task.Completed, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrTaskNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &task, nil
}

func (s *TaskStorage) GetTasksByUserID(ctx context.Context, userID int64, completed *bool) ([]entity.Task, error) {
	const op = "internal.repository.task.GetTasksByUserID"

	var query string
	var args []any

	if completed != nil {
		query = `SELECT id, title, description, user_id, completed, completed_at, created_at, updated_at 
		         FROM tasks 
		         WHERE user_id = $1 AND completed = $2
		         ORDER BY created_at DESC`
		args = []any{userID, *completed}
	} else {
		query = `SELECT id, title, description, user_id, completed, completed_at, created_at, updated_at 
		         FROM tasks 
		         WHERE user_id = $1
		         ORDER BY created_at DESC`
		args = []any{userID}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	tasks := make([]entity.Task, 0)
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.UserID,
			&task.Completed, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
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

func (s *TaskStorage) UpdateTask(ctx context.Context, taskID, userID int64, task entity.TaskUpdateRequest) error {
	const op = "internal.repository.task.UpdateTask"

	now := time.Now()

	var completedAt *time.Time
	if task.Completed != nil && *task.Completed {
		completedAt = &now
	}

	query := `
		UPDATE tasks 
		SET 
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			completed = COALESCE($3, completed),
			completed_at = CASE
				WHEN $3 IS TRUE THEN $4
				WHEN $3 IS FALSE THEN NULL
				ELSE completed_at
			END,
			updated_at = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id
	`

	var id int64
	err := s.pool.QueryRow(ctx,
		query,
		task.Title, task.Description, task.Completed, completedAt, now, taskID, userID,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, ErrTaskNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *TaskStorage) DeleteTask(ctx context.Context, taskID, userID int64) error {
	const op = "internal.repository.task.DeleteTask"

	result, err := s.pool.Exec(ctx,
		`DELETE FROM tasks WHERE id = $1 AND user_id = $2`,
		taskID, userID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrTaskNotFound)
	}

	return nil
}
