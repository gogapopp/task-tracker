package repository

import (
	"context"
	"fmt"
	"scheduler/internal/entity"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type UserRepository struct {
	logger *zap.SugaredLogger
	db     *pgx.Conn
}

func NewUserRepository(logger *zap.SugaredLogger, db *pgx.Conn) *UserRepository {
	return &UserRepository{
		logger: logger,
		db:     db,
	}
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	const op = "internal.repository.user.GetAllUsers"

	query := `SELECT id, email FROM users`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User

		err := rows.Scan(
			&user.ID,
			&user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}
