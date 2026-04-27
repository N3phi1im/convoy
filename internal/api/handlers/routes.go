package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"convoy/internal/api/middleware"
	"convoy/internal/models"
	"convoy/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RouteHandler struct {
	routeService *service.RouteService
}

func NewRouteHandler(routeService *service.RouteService) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
	}
}

// CreateRoute handles POST /api/v1/routes
func (h *RouteHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.RouteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := validateRouteCreateRequest(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation error: "+err.Error())
		return
	}

	route, err := h.routeService.Create(ctx, userID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create route: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    route,
	})
}

// GetRoute handles GET /api/v1/routes/{id}
func (h *RouteHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
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

	route, err := h.routeService.GetByID(ctx, routeID, userID)
	if err != nil {
		if err.Error() == "route not found" {
			respondError(w, http.StatusNotFound, "route not found")
			return
		}
		if err.Error() == "access denied: route is private" || err.Error() == "access denied: route is invite-only" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    route,
	})
}

// ListRoutes handles GET /api/v1/routes
func (h *RouteHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := parseRouteListQuery(r)

	routes, total, err := h.routeService.List(ctx, userID, query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    routes,
		"meta": map[string]interface{}{
			"total": total,
			"page":  query.Page,
			"limit": query.Limit,
		},
	})
}

// UpdateRoute handles PUT /api/v1/routes/{id}
func (h *RouteHandler) UpdateRoute(w http.ResponseWriter, r *http.Request) {
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

	var req models.RouteUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.routeService.Update(ctx, routeID, userID, &req); err != nil {
		if err.Error() == "route not found" {
			respondError(w, http.StatusNotFound, "route not found")
			return
		}
		if err.Error() == "access denied: only creator can update route" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "route updated successfully",
	})
}

// DeleteRoute handles DELETE /api/v1/routes/{id}
func (h *RouteHandler) DeleteRoute(w http.ResponseWriter, r *http.Request) {
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

	if err := h.routeService.Delete(ctx, routeID, userID); err != nil {
		if err.Error() == "route not found" {
			respondError(w, http.StatusNotFound, "route not found")
			return
		}
		if err.Error() == "access denied: only creator can delete route" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "route deleted successfully",
	})
}

// Helper functions

func parseRouteListQuery(r *http.Request) *models.RouteListQuery {
	query := &models.RouteListQuery{
		Page:  1,
		Limit: 20,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	if routeType := r.URL.Query().Get("route_type"); routeType != "" {
		rt := models.RouteType(routeType)
		query.RouteType = &rt
	}
	if visibility := r.URL.Query().Get("visibility"); visibility != "" {
		v := models.Visibility(visibility)
		query.Visibility = &v
	}
	if status := r.URL.Query().Get("status"); status != "" {
		s := models.RouteStatus(status)
		query.Status = &s
	}

	if search := r.URL.Query().Get("search"); search != "" {
		query.Search = &search
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	return query
}

func validateRouteCreateRequest(req *models.RouteCreateRequest) error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if len(req.Name) < 3 || len(req.Name) > 100 {
		return &ValidationError{Field: "name", Message: "name must be between 3 and 100 characters"}
	}
	if req.RouteType == "" {
		return &ValidationError{Field: "route_type", Message: "route_type is required"}
	}
	if req.Visibility == "" {
		return &ValidationError{Field: "visibility", Message: "visibility is required"}
	}
	if len(req.Waypoints) < 2 {
		return &ValidationError{Field: "waypoints", Message: "at least 2 waypoints are required"}
	}
	for i, wp := range req.Waypoints {
		if wp.Latitude < -90 || wp.Latitude > 90 {
			return &ValidationError{Field: "waypoints", Message: "invalid latitude at waypoint " + strconv.Itoa(i)}
		}
		if wp.Longitude < -180 || wp.Longitude > 180 {
			return &ValidationError{Field: "waypoints", Message: "invalid longitude at waypoint " + strconv.Itoa(i)}
		}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
