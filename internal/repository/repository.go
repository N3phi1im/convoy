package repository

import (
	"convoy/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
}

type RouteRepository interface {
	Create(route *models.Route) error
	GetByID(id uuid.UUID) (*models.Route, error)
	List(query *models.RouteListQuery) ([]models.Route, int, error)
	Update(route *models.Route) error
	Delete(id uuid.UUID) error
	GetWithDetails(id uuid.UUID) (*models.Route, error)
}

type WaypointRepository interface {
	CreateBatch(waypoints []models.Waypoint) error
	GetByRouteID(routeID uuid.UUID) ([]models.Waypoint, error)
	DeleteByRouteID(routeID uuid.UUID) error
}

type ParticipantRepository interface {
	Create(participant *models.Participant) error
	GetByRouteAndUser(routeID, userID uuid.UUID) (*models.Participant, error)
	ListByRoute(routeID uuid.UUID) ([]models.Participant, error)
	Update(participant *models.Participant) error
	CountActiveByRoute(routeID uuid.UUID) (int, error)
}
