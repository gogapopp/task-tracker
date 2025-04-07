package entity

import "time"

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	UserID      int64      `json:"user_id"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskCreateRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type TaskUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   *bool  `json:"completed,omitempty"`
}

type TaskResponse struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

type TaskStatsResponse struct {
	CompletedCount int `json:"completed_count"`
	PendingCount   int `json:"pending_count"`
}
