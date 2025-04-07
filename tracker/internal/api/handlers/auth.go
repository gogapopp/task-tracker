package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"tracker/internal/api/middleware"
	"tracker/internal/entity"
	"tracker/internal/repository"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger    *zap.SugaredLogger
	userSvc   UserService
	validator *validator.Validate
}

type UserService interface {
	RegisterUser(ctx context.Context, req entity.UserRegisterRequest) (string, error)
	LoginUser(ctx context.Context, req entity.UserLoginRequest) (string, error)
	GetCurrentUser(ctx context.Context, userID int64) (*entity.UserResponse, error)
}

func NewUserHandler(logger *zap.SugaredLogger, userSvc UserService) *UserHandler {
	return &UserHandler{
		logger:    logger,
		userSvc:   userSvc,
		validator: validator.New(),
	}
}

func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.auth.Register"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		var req entity.UserRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid request")
			return
		}

		if err := h.validator.Struct(req); err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid request fields")
			return
		}

		token, err := h.userSvc.RegisterUser(ctx, req)
		if err != nil {
			logger.Errorf("%s: %w", op, err)

			switch {
			case errors.Is(err, repository.ErrEmailTaken):
				jsonErrorResponse(w, http.StatusConflict, "email is already taken")
			default:
				jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		jsonResponse(w, http.StatusOK, map[string]string{"token": token})
	}
}

func (h *UserHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.auth.Login"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		var req entity.UserLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid request")
			return
		}

		if err := h.validator.Struct(req); err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid request fields")
			return
		}

		token, err := h.userSvc.LoginUser(ctx, req)
		if err != nil {
			logger.Errorf("%s: %w", op, err)
			switch {
			case errors.Is(err, repository.ErrInvalidCredentials):
				jsonErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
			default:
				jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		jsonResponse(w, http.StatusOK, map[string]string{"token": token})
	}
}

func (h *UserHandler) GetCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.auth.GetCurrentUser"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exist := middleware.GetUserIDFromContext(ctx)
		if !exist {
			logger.Errorf("%s: %w", op, "user not found in context")
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := h.userSvc.GetCurrentUser(ctx, userID)
		if err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		jsonResponse(w, http.StatusOK, user)
	}
}
