package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"tracker/internal/libs/jwt"

	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func JWTAuthMiddleware(logger *zap.SugaredLogger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "internal.api.middleware.middlewares.JWTAuthMiddleware"

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "authorization header is required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "authorization header must be in format: Bearer {token}")
				return
			}

			tokenString := parts[1]

			userID, err := jwt.ExtractUserID(tokenString, jwtSecret)
			if err != nil {
				logger.Errorf("%s: invalid token: %w", op, err)

				switch err {
				case jwt.ErrExpiredToken:
					respondWithError(w, http.StatusUnauthorized, "token has expired")
				default:
					respondWithError(w, http.StatusUnauthorized, "invalid token")
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserIDFromContext(r.Context())
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": message}); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
