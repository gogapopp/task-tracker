package api

import (
	"net/http"
	"tracker/internal/api/handlers"
	"tracker/internal/api/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Router(userHandler *handlers.UserHandler, taskHandler *handlers.TaskHandler, logger *zap.SugaredLogger, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	r.Post("/user", userHandler.Register())
	r.Post("/auth/login", userHandler.Login())

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(logger, jwtSecret))

		r.Get("/user", userHandler.GetCurrentUser())

		r.Post("/tasks", taskHandler.CreateTask())
		r.Get("/tasks", taskHandler.GetTasks())
		r.Get("/tasks/{id}", taskHandler.GetTask())
		r.Put("/tasks/{id}", taskHandler.UpdateTask())
		r.Delete("/tasks/{id}", taskHandler.DeleteTask())
	})

	return r
}
