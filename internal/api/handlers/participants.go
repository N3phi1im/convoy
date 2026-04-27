package handlers

import (
	"net/http"

	"convoy/internal/api/middleware"
	"convoy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ParticipantHandler struct {
	participantService *service.ParticipantService
}

func NewParticipantHandler(participantService *service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{
		participantService: participantService,
	}
}

// JoinRoute handles POST /api/v1/routes/{id}/join
func (h *ParticipantHandler) JoinRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	routeIDStr := chi.URLParam(r, "id")
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid route ID")
		return
	}

	if err := h.participantService.JoinRoute(ctx, routeID, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "successfully joined route",
	})
}

// LeaveRoute handles POST /api/v1/routes/{id}/leave
func (h *ParticipantHandler) LeaveRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	routeIDStr := chi.URLParam(r, "id")
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid route ID")
		return
	}

	if err := h.participantService.LeaveRoute(ctx, routeID, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "successfully left route",
	})
}

// ListParticipants handles GET /api/v1/routes/{id}/participants
func (h *ParticipantHandler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routeIDStr := chi.URLParam(r, "id")
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid route ID")
		return
	}

	participants, err := h.participantService.ListParticipants(ctx, routeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list participants")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    participants,
	})
}

// RemoveParticipant handles DELETE /api/v1/routes/{id}/participants/{userId}
func (h *ParticipantHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	creatorID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	routeIDStr := chi.URLParam(r, "id")
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid route ID")
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.participantService.RemoveParticipant(ctx, routeID, userID, creatorID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "participant removed successfully",
	})
}
