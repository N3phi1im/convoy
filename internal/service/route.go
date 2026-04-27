package service

import (
	"context"
	"fmt"

	"convoy/internal/maps"
	"convoy/internal/models"

	"github.com/google/uuid"
)

type RouteRepository interface {
	Create(ctx context.Context, route *models.Route) error
	GetByID(ctx context.Context, id string) (*models.Route, error)
	List(ctx context.Context, filters *models.RouteFilters) ([]*models.Route, int, error)
	Update(ctx context.Context, route *models.Route) error
	Delete(ctx context.Context, id string) error
	GetWithDetails(ctx context.Context, id string) (*models.RouteWithDetails, error)
}

type WaypointRepository interface {
	CreateBatch(ctx context.Context, waypoints []models.Waypoint) error
	GetByRouteID(ctx context.Context, routeID string) ([]models.Waypoint, error)
	DeleteByRouteID(ctx context.Context, routeID string) error
	Update(ctx context.Context, routeID string, waypoints []models.Waypoint) error
}

type ParticipantRepository interface {
	Create(ctx context.Context, participant *models.Participant) error
	GetByRouteAndUser(ctx context.Context, routeID, userID string) (*models.Participant, error)
	ListByRoute(ctx context.Context, routeID string) ([]models.ParticipantWithUser, error)
	CountByRoute(ctx context.Context, routeID string) (int, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	DeleteByRouteAndUser(ctx context.Context, routeID, userID string) error
}

type RouteService struct {
	routeRepo       RouteRepository
	waypointRepo    WaypointRepository
	participantRepo ParticipantRepository
	mapProvider     maps.Provider
}

func NewRouteService(
	routeRepo RouteRepository,
	waypointRepo WaypointRepository,
	participantRepo ParticipantRepository,
	mapProvider maps.Provider,
) *RouteService {
	return &RouteService{
		routeRepo:       routeRepo,
		waypointRepo:    waypointRepo,
		participantRepo: participantRepo,
		mapProvider:     mapProvider,
	}
}

// Create creates a new route with waypoints and adds creator as first participant
func (s *RouteService) Create(ctx context.Context, userID uuid.UUID, req *models.RouteCreateRequest) (*models.Route, error) {
	if len(req.Waypoints) < 2 {
		return nil, fmt.Errorf("route must have at least 2 waypoints")
	}

	// Convert waypoint requests to coordinates for map provider
	coordinates := make([]maps.Coordinate, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		coordinates[i] = maps.Coordinate{
			Latitude:  wp.Latitude,
			Longitude: wp.Longitude,
		}
	}

	// Call map provider for directions
	directionsReq := &maps.DirectionsRequest{
		Waypoints: coordinates,
		Profile:   string(req.RouteType),
	}

	directions, err := s.mapProvider.GetDirections(ctx, directionsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get directions: %w", err)
	}

	// Create route with calculated distance and duration, and Mapbox polyline
	route := &models.Route{
		ID:              uuid.New(),
		CreatorID:       userID,
		Name:            req.Name,
		Description:     req.Description,
		RouteType:       req.RouteType,
		Visibility:      req.Visibility,
		StartTime:       req.StartTime,
		MaxParticipants: req.MaxParticipants,
		Difficulty:      req.Difficulty,
		Distance:        directions.Distance,
		Duration:        int(directions.Duration),
		Geometry:        &directions.Geometry,
		Status:          models.RouteStatusPlanned,
	}

	// Save route
	if err := s.routeRepo.Create(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	// Create waypoints
	waypoints := make([]models.Waypoint, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		waypoints[i] = models.Waypoint{
			ID:        uuid.New(),
			RouteID:   route.ID,
			Order:     wp.Order,
			Latitude:  wp.Latitude,
			Longitude: wp.Longitude,
			Name:      wp.Name,
			Address:   wp.Address,
		}
	}

	if err := s.waypointRepo.CreateBatch(ctx, waypoints); err != nil {
		_ = s.routeRepo.Delete(ctx, route.ID.String())
		return nil, fmt.Errorf("failed to create waypoints: %w", err)
	}

	// Add creator as first participant
	participant := &models.Participant{
		ID:      uuid.New(),
		RouteID: route.ID,
		UserID:  userID,
		Status:  models.ParticipantStatusActive,
	}

	if err := s.participantRepo.Create(ctx, participant); err != nil {
		_ = s.waypointRepo.DeleteByRouteID(ctx, route.ID.String())
		_ = s.routeRepo.Delete(ctx, route.ID.String())
		return nil, fmt.Errorf("failed to add creator as participant: %w", err)
	}

	details, err := s.routeRepo.GetWithDetails(ctx, route.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load route details: %w", err)
	}

	return &details.Route, nil
}

// GetByID retrieves a route by ID with visibility permission checks
func (s *RouteService) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.RouteWithDetails, error) {
	details, err := s.routeRepo.GetWithDetails(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("route not found")
	}

	if err := s.checkViewPermission(ctx, details, userID); err != nil {
		return nil, err
	}

	return details, nil
}

// checkViewPermission checks if user can view the route based on visibility
func (s *RouteService) checkViewPermission(ctx context.Context, route *models.RouteWithDetails, userID uuid.UUID) error {
	if route.CreatorID == userID {
		return nil
	}

	switch route.Visibility {
	case models.VisibilityPublic:
		return nil

	case models.VisibilityPrivate:
		participant, err := s.participantRepo.GetByRouteAndUser(ctx, route.ID.String(), userID.String())
		if err != nil {
			return fmt.Errorf("failed to check participant status: %w", err)
		}
		if participant == nil {
			return fmt.Errorf("access denied: route is private")
		}
		return nil

	case models.VisibilityInviteOnly:
		participant, err := s.participantRepo.GetByRouteAndUser(ctx, route.ID.String(), userID.String())
		if err != nil {
			return fmt.Errorf("failed to check participant status: %w", err)
		}
		if participant == nil {
			return fmt.Errorf("access denied: route is invite-only")
		}
		return nil

	default:
		return fmt.Errorf("unknown visibility type")
	}
}

// List retrieves routes based on filters with pagination
func (s *RouteService) List(ctx context.Context, userID uuid.UUID, query *models.RouteListQuery) ([]*models.Route, int, error) {
	filters := &models.RouteFilters{
		Limit:  query.Limit,
		Offset: (query.Page - 1) * query.Limit,
	}

	if query.RouteType != nil {
		filters.RouteType = string(*query.RouteType)
	}
	if query.Visibility != nil {
		filters.Visibility = string(*query.Visibility)
	}
	if query.Status != nil {
		filters.Status = string(*query.Status)
	}

	routes, _, err := s.routeRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list routes: %w", err)
	}

	visibleRoutes := []*models.Route{}
	for _, route := range routes {
		if route.Visibility == models.VisibilityPublic {
			visibleRoutes = append(visibleRoutes, route)
			continue
		}

		if route.CreatorID == userID {
			visibleRoutes = append(visibleRoutes, route)
			continue
		}

		participant, err := s.participantRepo.GetByRouteAndUser(ctx, route.ID.String(), userID.String())
		if err == nil && participant != nil {
			visibleRoutes = append(visibleRoutes, route)
		}
	}

	return visibleRoutes, len(visibleRoutes), nil
}

// Update updates a route
func (s *RouteService) Update(ctx context.Context, id, userID uuid.UUID, req *models.RouteUpdateRequest) error {
	route, err := s.routeRepo.GetByID(ctx, id.String())
	if err != nil {
		return fmt.Errorf("route not found")
	}

	if route.CreatorID != userID {
		return fmt.Errorf("access denied: only creator can update route")
	}

	if route.Status == models.RouteStatusCompleted || route.Status == models.RouteStatusCancelled {
		return fmt.Errorf("cannot update %s route", route.Status)
	}

	if req.Name != nil {
		route.Name = *req.Name
	}
	if req.Description != nil {
		route.Description = req.Description
	}
	if req.StartTime != nil {
		route.StartTime = req.StartTime
	}
	if req.MaxParticipants != nil {
		count, err := s.participantRepo.CountByRoute(ctx, id.String())
		if err != nil {
			return fmt.Errorf("failed to count participants: %w", err)
		}
		if *req.MaxParticipants < count {
			return fmt.Errorf("max participants cannot be less than current participant count (%d)", count)
		}
		route.MaxParticipants = req.MaxParticipants
	}
	if req.Difficulty != nil {
		route.Difficulty = req.Difficulty
	}
	if req.Visibility != nil {
		route.Visibility = *req.Visibility
	}

	if err := s.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	return nil
}

// Delete deletes a route
func (s *RouteService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	route, err := s.routeRepo.GetByID(ctx, id.String())
	if err != nil {
		return fmt.Errorf("route not found")
	}
	if route.CreatorID != userID {
		return fmt.Errorf("access denied: only creator can delete route")
	}

	if route.Status == models.RouteStatusInProgress {
		return fmt.Errorf("cannot delete route that is in progress")
	}
	if err := s.routeRepo.Delete(ctx, id.String()); err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a route
func (s *RouteService) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status models.RouteStatus) error {
	route, err := s.routeRepo.GetByID(ctx, id.String())
	if err != nil {
		return fmt.Errorf("route not found")
	}
	if route.CreatorID != userID {
		return fmt.Errorf("access denied: only creator can update route status")
	}

	if err := s.validateStatusTransition(route.Status, status); err != nil {
		return err
	}
	route.Status = status
	if err := s.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route status: %w", err)
	}

	return nil
}

// validateStatusTransition validates if status transition is allowed
func (s *RouteService) validateStatusTransition(current, new models.RouteStatus) error {
	validTransitions := map[models.RouteStatus][]models.RouteStatus{
		models.RouteStatusPlanned: {
			models.RouteStatusInProgress,
			models.RouteStatusCancelled,
		},
		models.RouteStatusInProgress: {
			models.RouteStatusCompleted,
			models.RouteStatusCancelled,
		},
		models.RouteStatusCompleted: {},
		models.RouteStatusCancelled: {},
	}

	allowed := validTransitions[current]
	for _, status := range allowed {
		if status == new {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", current, new)
}
