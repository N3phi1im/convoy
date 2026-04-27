package handlers

import (
	"net/http"

	"convoy/internal/api/middleware"
	"convoy/internal/service"

	"github.com/google/uuid"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    user,
	})
}
