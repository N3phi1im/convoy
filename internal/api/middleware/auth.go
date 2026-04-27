package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"convoy/internal/models"
	"convoy/internal/service"
)

type contextKey string

const UserIDKey contextKey = "userID"

func RequireAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
				return
			}

			token := parts[1]
			userID, err := authService.ValidateToken(token)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.NewErrorResponse(code, message, nil)

	if err := writeJSON(w, response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
