package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"convoy/internal/maps"
)

type DirectionsProfile string

const (
	ProfileDriving        DirectionsProfile = "driving"
	ProfileDrivingTraffic DirectionsProfile = "driving-traffic"
	ProfileWalking        DirectionsProfile = "walking"
	ProfileCycling        DirectionsProfile = "cycling"
)

type mapboxDirectionsResponse struct {
	Code      string           `json:"code"`
	Message   string           `json:"message,omitempty"`
	Routes    []mapboxRoute    `json:"routes"`
	Waypoints []mapboxWaypoint `json:"waypoints"`
}

type mapboxRoute struct {
	Distance float64     `json:"distance"` // meters
	Duration float64     `json:"duration"` // seconds
	Geometry string      `json:"geometry"` // polyline
	Legs     []mapboxLeg `json:"legs"`
}

type mapboxLeg struct {
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Steps    []mapboxStep `json:"steps"`
}

type mapboxStep struct {
	Distance float64        `json:"distance"`
	Duration float64        `json:"duration"`
	Geometry string         `json:"geometry"`
	Name     string         `json:"name"`
	Maneuver mapboxManeuver `json:"maneuver"`
}

type mapboxManeuver struct {
	Location    []float64 `json:"location"` // [lng, lat]
	Instruction string    `json:"instruction"`
	Type        string    `json:"type"`
}

type mapboxWaypoint struct {
	Name     string    `json:"name"`
	Location []float64 `json:"location"` // [lng, lat]
}

// GetDirections retrieves directions from Mapbox API
func (c *Client) GetDirections(ctx context.Context, req *maps.DirectionsRequest) (*maps.DirectionsResponse, error) {
	if len(req.Waypoints) < 2 {
		return nil, fmt.Errorf("at least 2 waypoints required")
	}

	profile := ProfileDriving
	switch req.Profile {
	case "driving":
		profile = ProfileDriving
	case "cycling":
		profile = ProfileCycling
	case "walking":
		profile = ProfileWalking
	case "running":
		profile = ProfileWalking
	}

	// Build coordinates string: "lng,lat;lng,lat;..."
	coords := make([]string, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		coords[i] = fmt.Sprintf("%f,%f", wp.Longitude, wp.Latitude)
	}
	coordsStr := strings.Join(coords, ";")

	url := fmt.Sprintf("%s/directions/v5/mapbox/%s/%s?geometries=polyline&steps=true&access_token=%s",
		baseURL, profile, coordsStr, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var mapboxResp mapboxDirectionsResponse
	if err := json.Unmarshal(body, &mapboxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if mapboxResp.Code != "Ok" {
		return nil, fmt.Errorf("mapbox error: %s - %s", mapboxResp.Code, mapboxResp.Message)
	}

	if len(mapboxResp.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	return c.convertDirectionsResponse(&mapboxResp)
}

// convertDirectionsResponse converts Mapbox response to our format
func (c *Client) convertDirectionsResponse(resp *mapboxDirectionsResponse) (*maps.DirectionsResponse, error) {
	if len(resp.Routes) == 0 {
		return nil, fmt.Errorf("no routes in response")
	}

	route := resp.Routes[0]

	waypoints := make([]maps.Waypoint, len(resp.Waypoints))
	for i, wp := range resp.Waypoints {
		if len(wp.Location) < 2 {
			continue
		}
		waypoints[i] = maps.Waypoint{
			Coordinate: maps.Coordinate{
				Longitude: wp.Location[0],
				Latitude:  wp.Location[1],
			},
			Name: wp.Name,
		}
	}

	var steps []maps.RouteStep
	for _, leg := range route.Legs {
		for _, step := range leg.Steps {
			if len(step.Maneuver.Location) < 2 {
				continue
			}
			steps = append(steps, maps.RouteStep{
				Distance:    step.Distance,
				Duration:    step.Duration,
				Instruction: step.Maneuver.Instruction,
				Coordinate: maps.Coordinate{
					Longitude: step.Maneuver.Location[0],
					Latitude:  step.Maneuver.Location[1],
				},
			})
		}
	}

	return &maps.DirectionsResponse{
		Distance:  route.Distance,
		Duration:  route.Duration,
		Geometry:  route.Geometry,
		Steps:     steps,
		Waypoints: waypoints,
	}, nil
}
