package entity

import "time"

type EmailMessage struct {
	Type      string            `json:"type"`
	To        string            `json:"to"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	Variables map[string]string `json:"variables,omitempty"`
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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
