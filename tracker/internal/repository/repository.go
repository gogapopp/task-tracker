package repository

import (
	"context"
	"errors"
	"tracker/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmailTaken         = errors.New("email is already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, err
	}

	return dbpool, nil
}
