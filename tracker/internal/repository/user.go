package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"
	"tracker/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	pool       *pgxpool.Pool
	passSecret string
}

func NewUserStorage(pool *pgxpool.Pool, passSecret string) *UserStorage {
	return &UserStorage{
		pool:       pool,
		passSecret: passSecret,
	}
}

func (s *UserStorage) RegisterUser(ctx context.Context, user entity.UserRegisterRequest) (int64, error) {
	const op = "internal.repository.user.RegisterUser"

	hashedPassword := s.generatePasswordHash(user.Password)
	now := time.Now()

	var userID int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users(email, password, created_at, updated_at) 
		 VALUES($1, $2, $3, $4) 
		 ON CONFLICT (email) DO NOTHING 
		 RETURNING id`,
		user.Email, hashedPassword, now, now,
	).Scan(&userID)
	if err == nil {
		return userID, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("%s: %w", op, ErrEmailTaken)
	}

	return 0, fmt.Errorf("%s: %w", op, err)
}

func (s *UserStorage) LoginUser(ctx context.Context, user entity.UserLoginRequest) (int64, error) {
	const op = "internal.repository.user.LoginUser"

	hashedPassword := s.generatePasswordHash(user.Password)

	var userID int64
	var passwordFromDB string
	err := s.pool.QueryRow(ctx,
		`SELECT id, password FROM users WHERE email = $1`,
		user.Email,
	).Scan(&userID, &passwordFromDB)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if passwordFromDB != hashedPassword {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	return userID, nil
}

func (s *UserStorage) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	const op = "internal.repository.user.GetUserByID"

	var user entity.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *UserStorage) generatePasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(s.passSecret)))
}
