package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"convoy/internal/maps"
	"convoy/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockRouteRepository struct {
	mock.Mock
}

func (m *MockRouteRepository) Create(ctx context.Context, route *models.Route) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockRouteRepository) GetByID(ctx context.Context, id string) (*models.Route, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Route), args.Error(1)
}

func (m *MockRouteRepository) List(ctx context.Context, filters *models.RouteFilters) ([]*models.Route, int, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Route), args.Int(1), args.Error(2)
}

func (m *MockRouteRepository) Update(ctx context.Context, route *models.Route) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockRouteRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRouteRepository) GetWithDetails(ctx context.Context, id string) (*models.RouteWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RouteWithDetails), args.Error(1)
}

type MockWaypointRepository struct {
	mock.Mock
}

func (m *MockWaypointRepository) CreateBatch(ctx context.Context, waypoints []models.Waypoint) error {
	args := m.Called(ctx, waypoints)
	return args.Error(0)
}

func (m *MockWaypointRepository) GetByRouteID(ctx context.Context, routeID string) ([]models.Waypoint, error) {
	args := m.Called(ctx, routeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Waypoint), args.Error(1)
}

func (m *MockWaypointRepository) DeleteByRouteID(ctx context.Context, routeID string) error {
	args := m.Called(ctx, routeID)
	return args.Error(0)
}

func (m *MockWaypointRepository) Update(ctx context.Context, routeID string, waypoints []models.Waypoint) error {
	args := m.Called(ctx, routeID, waypoints)
	return args.Error(0)
}

type MockParticipantRepository struct {
	mock.Mock
}

func (m *MockParticipantRepository) Create(ctx context.Context, participant *models.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockParticipantRepository) GetByRouteAndUser(ctx context.Context, routeID, userID string) (*models.Participant, error) {
	args := m.Called(ctx, routeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Participant), args.Error(1)
}

func (m *MockParticipantRepository) ListByRoute(ctx context.Context, routeID string) ([]models.ParticipantWithUser, error) {
	args := m.Called(ctx, routeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantWithUser), args.Error(1)
}

func (m *MockParticipantRepository) CountByRoute(ctx context.Context, routeID string) (int, error) {
	args := m.Called(ctx, routeID)
	return args.Int(0), args.Error(1)
}

func (m *MockParticipantRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockParticipantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockParticipantRepository) DeleteByRouteAndUser(ctx context.Context, routeID, userID string) error {
	args := m.Called(ctx, routeID, userID)
	return args.Error(0)
}

type MockMapProvider struct {
	mock.Mock
}

func (m *MockMapProvider) GetDirections(ctx context.Context, req *maps.DirectionsRequest) (*maps.DirectionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*maps.DirectionsResponse), args.Error(1)
}

func (m *MockMapProvider) Geocode(ctx context.Context, req *maps.GeocodingRequest) (*maps.GeocodingResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*maps.GeocodingResponse), args.Error(1)
}

func (m *MockMapProvider) ReverseGeocode(ctx context.Context, req *maps.ReverseGeocodingRequest) (*maps.GeocodingResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*maps.GeocodingResponse), args.Error(1)
}

// Tests
func TestRouteService_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful creation", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		req := &models.RouteCreateRequest{
			Name:      "Test Route",
			RouteType: models.RouteTypeCycling,
			Visibility: models.VisibilityPublic,
			Waypoints: []models.WaypointCreateRequest{
				{Latitude: 40.7128, Longitude: -74.0060, Order: 0},
				{Latitude: 34.0522, Longitude: -118.2437, Order: 1},
			},
		}

		// Mock map provider response
		mockMapProvider.On("GetDirections", ctx, mock.AnythingOfType("*maps.DirectionsRequest")).
			Return(&maps.DirectionsResponse{
				Distance: 5000.0,
				Duration: 3600.0,
			}, nil)

		// Mock route creation
		mockRouteRepo.On("Create", ctx, mock.AnythingOfType("*models.Route")).
			Return(nil)

		// Mock waypoint creation
		mockWaypointRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]models.Waypoint")).
			Return(nil)

		// Mock participant creation
		mockParticipantRepo.On("Create", ctx, mock.AnythingOfType("*models.Participant")).
			Return(nil)

		// Mock GetWithDetails
		mockRouteRepo.On("GetWithDetails", ctx, mock.AnythingOfType("string")).
			Return(&models.RouteWithDetails{
				Route: models.Route{
					ID:        uuid.New(),
					CreatorID: userID,
					Name:      "Test Route",
					Distance:  5000.0,
					Duration:  3600,
				},
			}, nil)

		route, err := service.Create(ctx, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, route)
		assert.Equal(t, "Test Route", route.Name)
		assert.Equal(t, 5000.0, route.Distance)
		assert.Equal(t, 3600, route.Duration)

		mockMapProvider.AssertExpectations(t)
		mockRouteRepo.AssertExpectations(t)
		mockWaypointRepo.AssertExpectations(t)
		mockParticipantRepo.AssertExpectations(t)
	})

	t.Run("insufficient waypoints", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		req := &models.RouteCreateRequest{
			Name:      "Test Route",
			RouteType: models.RouteTypeCycling,
			Waypoints: []models.WaypointCreateRequest{
				{Latitude: 40.7128, Longitude: -74.0060, Order: 0},
			},
		}

		route, err := service.Create(ctx, userID, req)
		assert.Error(t, err)
		assert.Nil(t, route)
		assert.Contains(t, err.Error(), "at least 2 waypoints")
	})

	t.Run("map provider error", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		req := &models.RouteCreateRequest{
			Name:      "Test Route",
			RouteType: models.RouteTypeCycling,
			Waypoints: []models.WaypointCreateRequest{
				{Latitude: 40.7128, Longitude: -74.0060, Order: 0},
				{Latitude: 34.0522, Longitude: -118.2437, Order: 1},
			},
		}

		mockMapProvider.On("GetDirections", ctx, mock.AnythingOfType("*maps.DirectionsRequest")).
			Return(nil, errors.New("map provider error"))

		route, err := service.Create(ctx, userID, req)
		assert.Error(t, err)
		assert.Nil(t, route)
		assert.Contains(t, err.Error(), "failed to get directions")

		mockMapProvider.AssertExpectations(t)
	})
}

func TestRouteService_GetByID(t *testing.T) {
	ctx := context.Background()
	routeID := uuid.New()
	creatorID := uuid.New()
	otherUserID := uuid.New()

	t.Run("creator can view route", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		mockRouteRepo.On("GetWithDetails", ctx, routeID.String()).
			Return(&models.RouteWithDetails{
				Route: models.Route{
					ID:         routeID,
					CreatorID:  creatorID,
					Name:       "Test Route",
					Visibility: models.VisibilityPrivate,
				},
			}, nil)

		route, err := service.GetByID(ctx, routeID, creatorID)
		assert.NoError(t, err)
		assert.NotNil(t, route)
		assert.Equal(t, "Test Route", route.Name)

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("public route visible to all", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		mockRouteRepo.On("GetWithDetails", ctx, routeID.String()).
			Return(&models.RouteWithDetails{
				Route: models.Route{
					ID:         routeID,
					CreatorID:  creatorID,
					Name:       "Test Route",
					Visibility: models.VisibilityPublic,
				},
			}, nil)

		route, err := service.GetByID(ctx, routeID, otherUserID)
		assert.NoError(t, err)
		assert.NotNil(t, route)

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("private route denied to non-participant", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		mockRouteRepo.On("GetWithDetails", ctx, routeID.String()).
			Return(&models.RouteWithDetails{
				Route: models.Route{
					ID:         routeID,
					CreatorID:  creatorID,
					Name:       "Test Route",
					Visibility: models.VisibilityPrivate,
				},
			}, nil)

		mockParticipantRepo.On("GetByRouteAndUser", ctx, routeID.String(), otherUserID.String()).
			Return(nil, nil)

		route, err := service.GetByID(ctx, routeID, otherUserID)
		assert.Error(t, err)
		assert.Nil(t, route)
		assert.Contains(t, err.Error(), "access denied")

		mockRouteRepo.AssertExpectations(t)
		mockParticipantRepo.AssertExpectations(t)
	})

	t.Run("route not found", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		mockRouteRepo.On("GetWithDetails", ctx, routeID.String()).
			Return(nil, errors.New("not found"))

		route, err := service.GetByID(ctx, routeID, creatorID)
		assert.Error(t, err)
		assert.Nil(t, route)
		assert.Contains(t, err.Error(), "not found")

		mockRouteRepo.AssertExpectations(t)
	})
}

func TestRouteService_Update(t *testing.T) {
	ctx := context.Background()
	routeID := uuid.New()
	creatorID := uuid.New()
	otherUserID := uuid.New()

	t.Run("successful update", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Name:      "Old Name",
			Status:    models.RouteStatusPlanned,
		}

		newName := "New Name"
		req := &models.RouteUpdateRequest{
			Name: &newName,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		mockRouteRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Route) bool {
			return r.Name == "New Name"
		})).Return(nil)

		err := service.Update(ctx, routeID, creatorID, req)
		assert.NoError(t, err)

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("non-creator cannot update", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Name:      "Old Name",
			Status:    models.RouteStatusPlanned,
		}

		newName := "New Name"
		req := &models.RouteUpdateRequest{
			Name: &newName,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		err := service.Update(ctx, routeID, otherUserID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("cannot update completed route", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Name:      "Old Name",
			Status:    models.RouteStatusCompleted,
		}

		newName := "New Name"
		req := &models.RouteUpdateRequest{
			Name: &newName,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		err := service.Update(ctx, routeID, creatorID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot update")

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("validate max participants", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Name:      "Old Name",
			Status:    models.RouteStatusPlanned,
		}

		newMax := 2
		req := &models.RouteUpdateRequest{
			MaxParticipants: &newMax,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		mockParticipantRepo.On("CountByRoute", ctx, routeID.String()).
			Return(5, nil)

		err := service.Update(ctx, routeID, creatorID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be less than current participant count")

		mockRouteRepo.AssertExpectations(t)
		mockParticipantRepo.AssertExpectations(t)
	})
}

func TestRouteService_Delete(t *testing.T) {
	ctx := context.Background()
	routeID := uuid.New()
	creatorID := uuid.New()
	otherUserID := uuid.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Status:    models.RouteStatusPlanned,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		mockRouteRepo.On("Delete", ctx, routeID.String()).
			Return(nil)

		err := service.Delete(ctx, routeID, creatorID)
		assert.NoError(t, err)

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("non-creator cannot delete", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Status:    models.RouteStatusPlanned,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		err := service.Delete(ctx, routeID, otherUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("cannot delete in-progress route", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Status:    models.RouteStatusInProgress,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		err := service.Delete(ctx, routeID, creatorID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete route that is in progress")

		mockRouteRepo.AssertExpectations(t)
	})
}

func TestRouteService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	routeID := uuid.New()
	creatorID := uuid.New()

	t.Run("valid status transition", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Status:    models.RouteStatusPlanned,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		mockRouteRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Route) bool {
			return r.Status == models.RouteStatusInProgress
		})).Return(nil)

		err := service.UpdateStatus(ctx, routeID, creatorID, models.RouteStatusInProgress)
		assert.NoError(t, err)

		mockRouteRepo.AssertExpectations(t)
	})

	t.Run("invalid status transition", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		existingRoute := &models.Route{
			ID:        routeID,
			CreatorID: creatorID,
			Status:    models.RouteStatusCompleted,
		}

		mockRouteRepo.On("GetByID", ctx, routeID.String()).
			Return(existingRoute, nil)

		err := service.UpdateStatus(ctx, routeID, creatorID, models.RouteStatusPlanned)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")

		mockRouteRepo.AssertExpectations(t)
	})
}

func TestRouteService_List(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("list public routes", func(t *testing.T) {
		mockRouteRepo := new(MockRouteRepository)
		mockWaypointRepo := new(MockWaypointRepository)
		mockParticipantRepo := new(MockParticipantRepository)
		mockMapProvider := new(MockMapProvider)

		service := NewRouteService(mockRouteRepo, mockWaypointRepo, mockParticipantRepo, mockMapProvider)

		query := &models.RouteListQuery{
			Page:  1,
			Limit: 20,
		}

		routes := []*models.Route{
			{
				ID:         uuid.New(),
				CreatorID:  uuid.New(),
				Name:       "Public Route",
				Visibility: models.VisibilityPublic,
			},
		}

		mockRouteRepo.On("List", ctx, mock.AnythingOfType("*models.RouteFilters")).
			Return(routes, 1, nil)

		result, total, err := service.List(ctx, userID, query)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 1, total)

		mockRouteRepo.AssertExpectations(t)
	})
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
