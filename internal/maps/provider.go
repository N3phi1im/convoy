package maps

import "context"

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Waypoint struct {
	Coordinate
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type DirectionsRequest struct {
	Waypoints []Coordinate `json:"waypoints"`
	Profile   string       `json:"profile"` // driving, cycling, walking
}

type DirectionsResponse struct {
	Distance  float64     `json:"distance"` // in meters
	Duration  float64     `json:"duration"` // in seconds
	Geometry  string      `json:"geometry"` // encoded polyline
	Steps     []RouteStep `json:"steps,omitempty"`
	Waypoints []Waypoint  `json:"waypoints"`
}

type RouteStep struct {
	Distance    float64    `json:"distance"`
	Duration    float64    `json:"duration"`
	Instruction string     `json:"instruction"`
	Coordinate  Coordinate `json:"coordinate"`
}

type GeocodingRequest struct {
	Address string `json:"address"`
}

type ReverseGeocodingRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeocodingResponse struct {
	Coordinate Coordinate `json:"coordinate"`
	Address    string     `json:"address"`
	PlaceName  string     `json:"place_name"`
	PlaceType  string     `json:"place_type"`
	Relevance  float64    `json:"relevance"`
	Context    []string   `json:"context,omitempty"`
}

type Provider interface {
	GetDirections(ctx context.Context, req *DirectionsRequest) (*DirectionsResponse, error)
	Geocode(ctx context.Context, req *GeocodingRequest) (*GeocodingResponse, error)
	ReverseGeocode(ctx context.Context, req *ReverseGeocodingRequest) (*GeocodingResponse, error)
}
