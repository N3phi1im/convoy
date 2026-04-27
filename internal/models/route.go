package models

import (
	"time"

	"github.com/google/uuid"
)

type RouteType string

const (
	RouteTypeDriving RouteType = "driving"
	RouteTypeCycling RouteType = "cycling"
	RouteTypeWalking RouteType = "walking"
	RouteTypeRunning RouteType = "running"
)

type Visibility string

const (
	VisibilityPublic     Visibility = "public"
	VisibilityPrivate    Visibility = "private"
	VisibilityInviteOnly Visibility = "invite_only"
)

type RouteStatus string

const (
	RouteStatusPlanned    RouteStatus = "planned"
	RouteStatusInProgress RouteStatus = "in_progress"
	RouteStatusCompleted  RouteStatus = "completed"
	RouteStatusCancelled  RouteStatus = "cancelled"
)

type Route struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	CreatorID       uuid.UUID   `json:"creator_id" db:"creator_id"`
	Name            string      `json:"name" db:"name" validate:"required,min=3,max=100"`
	Description     *string     `json:"description,omitempty" db:"description"`
	RouteType       RouteType   `json:"route_type" db:"route_type" validate:"required,oneof=driving cycling walking running"`
	Visibility      Visibility  `json:"visibility" db:"visibility" validate:"required,oneof=public private invite_only"`
	StartTime       *time.Time  `json:"start_time,omitempty" db:"start_time"`
	MaxParticipants *int        `json:"max_participants,omitempty" db:"max_participants" validate:"omitempty,min=2,max=1000"`
	Distance        float64     `json:"distance" db:"distance"` // in meters
	Duration        int         `json:"duration" db:"duration"` // in seconds
	Difficulty      *string     `json:"difficulty,omitempty" db:"difficulty"`
	Status          RouteStatus `json:"status" db:"status"`
	Geometry        *string     `json:"geometry,omitempty" db:"geometry"` // Mapbox polyline
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`

	Creator          *UserResponse  `json:"creator,omitempty" db:"-"`
	Waypoints        []Waypoint     `json:"waypoints,omitempty" db:"-"`
	Participants     []UserResponse `json:"participants,omitempty" db:"-"`
	ParticipantCount int            `json:"participant_count,omitempty" db:"-"`
}

type RouteCreateRequest struct {
	Name            string                  `json:"name" validate:"required,min=3,max=100"`
	Description     *string                 `json:"description,omitempty"`
	RouteType       RouteType               `json:"route_type" validate:"required,oneof=driving cycling walking running"`
	Visibility      Visibility              `json:"visibility" validate:"required,oneof=public private invite_only"`
	StartTime       *time.Time              `json:"start_time,omitempty"`
	MaxParticipants *int                    `json:"max_participants,omitempty" validate:"omitempty,min=2,max=1000"`
	Difficulty      *string                 `json:"difficulty,omitempty"`
	Waypoints       []WaypointCreateRequest `json:"waypoints" validate:"required,min=2,dive"`
}

type RouteUpdateRequest struct {
	Name            *string     `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description     *string     `json:"description,omitempty"`
	StartTime       *time.Time  `json:"start_time,omitempty"`
	MaxParticipants *int        `json:"max_participants,omitempty" validate:"omitempty,min=2,max=1000"`
	Difficulty      *string     `json:"difficulty,omitempty"`
	Visibility      *Visibility `json:"visibility,omitempty" validate:"omitempty,oneof=public private invite_only"`
}

type RouteListQuery struct {
	Page       int          `json:"page" validate:"min=1"`
	Limit      int          `json:"limit" validate:"min=1,max=100"`
	RouteType  *RouteType   `json:"route_type,omitempty"`
	Visibility *Visibility  `json:"visibility,omitempty"`
	Status     *RouteStatus `json:"status,omitempty"`
	Search     *string      `json:"search,omitempty"`
	SortBy     string       `json:"sort_by" validate:"omitempty,oneof=created_at start_time distance"`
	SortOrder  string       `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

type RouteFilters struct {
	CreatorID  string
	Status     string
	RouteType  string
	Visibility string
	Limit      int
	Offset     int
}

type RouteWithDetails struct {
	Route
	Waypoints    []Waypoint            `json:"waypoints"`
	Participants []ParticipantWithUser `json:"participants"`
}
