package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"tracker/internal/entity"
	"tracker/internal/libs/jwt"

	"github.com/go-playground/validator/v10"
)

type UserService struct {
	repo      UserRepository
	validator *validator.Validate
	jwtSecret string
	kafka     KafkaProducer
}

type UserRepository interface {
	RegisterUser(ctx context.Context, user entity.UserRegisterRequest) (int64, error)
	LoginUser(ctx context.Context, user entity.UserLoginRequest) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
}

type KafkaProducer interface {
	SendWelcomeEmail(ctx context.Context, email string) error
}

func NewUserService(repo UserRepository, jwtSecret string, kafka KafkaProducer) *UserService {
	return &UserService{
		repo:      repo,
		validator: validator.New(),
		jwtSecret: jwtSecret,
		kafka:     kafka,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req entity.UserRegisterRequest) (string, error) {
	const op = "internal.service.user.RegisterUser"

	if err := s.validator.Struct(req); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	userID, err := s.repo.RegisterUser(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.GenerateJWTToken(s.jwtSecret, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.kafka.SendWelcomeEmail(ctx, req.Email); err != nil {
			log.Printf("failed to send welcome email: %v\n", err)
		}
	}()

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, req entity.UserLoginRequest) (string, error) {
	const op = "internal.service.user.LoginUser"

	if err := s.validator.Struct(req); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	userID, err := s.repo.LoginUser(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.GenerateJWTToken(s.jwtSecret, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID int64) (*entity.UserResponse, error) {
	const op = "internal.service.user.GetCurrentUser"

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}
