package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/api/middleware"
	"tracker/internal/entity"
	"tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type TaskHandler struct {
	logger    *zap.SugaredLogger
	taskSvc   TaskService
	validator *validator.Validate
}

type TaskService interface {
	CreateTask(ctx context.Context, req entity.TaskCreateRequest, userID int64) (*entity.TaskResponse, error)
	GetTaskByID(ctx context.Context, taskID, userID int64) (*entity.TaskResponse, error)
	GetTasks(ctx context.Context, userID int64, completed *bool) (*entity.TaskListResponse, error)
	UpdateTask(ctx context.Context, taskID, userID int64, req entity.TaskUpdateRequest) (*entity.TaskResponse, error)
	DeleteTask(ctx context.Context, taskID, userID int64) error
}

func NewTaskHandler(logger *zap.SugaredLogger, taskSvc TaskService) *TaskHandler {
	return &TaskHandler{
		logger:    logger,
		taskSvc:   taskSvc,
		validator: validator.New(),
	}
}

func (h *TaskHandler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.task.CreateTask"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exists := middleware.GetUserIDFromContext(ctx)
		if !exists {
			logger.Errorf("%s: user not found in context", op)
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req entity.TaskCreateRequest
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

		task, err := h.taskSvc.CreateTask(ctx, req, userID)
		if err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		jsonResponse(w, http.StatusCreated, task)
	}
}

func (h *TaskHandler) GetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.task.GetTask"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exists := middleware.GetUserIDFromContext(ctx)
		if !exists {
			logger.Errorf("%s: user not found in context", op)
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		taskIDStr := chi.URLParam(r, "id")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			logger.Errorf("%s: invalid task ID: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid task ID")
			return
		}

		task, err := h.taskSvc.GetTaskByID(ctx, taskID, userID)
		if err != nil {
			logger.Errorf("%s: %w", op, err)

			if errors.Is(err, repository.ErrTaskNotFound) {
				jsonErrorResponse(w, http.StatusNotFound, "task not found")
				return
			}
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		jsonResponse(w, http.StatusOK, task)
	}
}

func (h *TaskHandler) GetTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.task.GetTasks"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exists := middleware.GetUserIDFromContext(ctx)
		if !exists {
			logger.Errorf("%s: user not found in context", op)
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var completed *bool
		completedStr := r.URL.Query().Get("completed")
		if completedStr != "" {
			completedVal, err := strconv.ParseBool(completedStr)
			if err != nil {
				logger.Errorf("%s: invalid completed parameter: %v", op, err)
				jsonErrorResponse(w, http.StatusBadRequest, "invalid completed parameter")
				return
			}
			completed = &completedVal
		}

		tasks, err := h.taskSvc.GetTasks(ctx, userID, completed)
		if err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		jsonResponse(w, http.StatusOK, tasks)
	}
}

func (h *TaskHandler) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.task.UpdateTask"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exists := middleware.GetUserIDFromContext(ctx)
		if !exists {
			logger.Errorf("%s: user not found in context", op)
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		taskIDStr := chi.URLParam(r, "id")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			logger.Errorf("%s: invalid task ID: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid task ID")
			return
		}

		var req entity.TaskUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("%s: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid request")
			return
		}

		task, err := h.taskSvc.UpdateTask(ctx, taskID, userID, req)
		if err != nil {
			logger.Errorf("%s: %w", op, err)

			if errors.Is(err, repository.ErrTaskNotFound) {
				jsonErrorResponse(w, http.StatusNotFound, "task not found")
				return
			}
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		jsonResponse(w, http.StatusOK, task)
	}
}

func (h *TaskHandler) DeleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.handlers.task.DeleteTask"
		ctx := r.Context()
		logger := h.logger.With("req_id", chimiddleware.GetReqID(ctx))

		userID, exists := middleware.GetUserIDFromContext(ctx)
		if !exists {
			logger.Errorf("%s: user not found in context", op)
			jsonErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		taskIDStr := chi.URLParam(r, "id")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			logger.Errorf("%s: invalid task ID: %w", op, err)
			jsonErrorResponse(w, http.StatusBadRequest, "invalid task ID")
			return
		}

		err = h.taskSvc.DeleteTask(ctx, taskID, userID)
		if err != nil {
			logger.Errorf("%s: %w", op, err)

			if errors.Is(err, repository.ErrTaskNotFound) {
				jsonErrorResponse(w, http.StatusNotFound, "task not found")
				return
			}
			jsonErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
