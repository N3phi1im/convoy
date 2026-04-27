package models

import (
	"time"

	"github.com/google/uuid"
)

type Waypoint struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RouteID   uuid.UUID `json:"route_id" db:"route_id"`
	Latitude  float64   `json:"latitude" db:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64   `json:"longitude" db:"longitude" validate:"required,min=-180,max=180"`
	Order     int       `json:"order" db:"order" validate:"required,min=0"`
	Name      *string   `json:"name,omitempty" db:"name"`
	Address   *string   `json:"address,omitempty" db:"address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WaypointCreateRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Order     int     `json:"order" validate:"required,min=0"`
	Name      *string `json:"name,omitempty"`
	Address   *string `json:"address,omitempty"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (w *Waypoint) ToCoordinate() Coordinate {
	return Coordinate{
		Latitude:  w.Latitude,
		Longitude: w.Longitude,
	}
}
