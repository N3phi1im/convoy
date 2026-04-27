package mapbox

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"convoy/internal/maps"
)

type mapboxGeocodingResponse struct {
	Type     string          `json:"type"`
	Query    []interface{}   `json:"query"`
	Features []mapboxFeature `json:"features"`
}

type mapboxFeature struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	PlaceType  []string        `json:"place_type"`
	Relevance  float64         `json:"relevance"`
	Properties json.RawMessage `json:"properties"`
	Text       string          `json:"text"`
	PlaceName  string          `json:"place_name"`
	Center     []float64       `json:"center"` // [lng, lat]
	Geometry   mapboxGeometry  `json:"geometry"`
	Context    []mapboxContext `json:"context,omitempty"`
}

type mapboxGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [lng, lat]
}

type mapboxContext struct {
	ID        string `json:"id"`
	ShortCode string `json:"short_code,omitempty"`
	Text      string `json:"text"`
}

// Geocode converts an address to coordinates
func (c *Client) Geocode(ctx context.Context, req *maps.GeocodingRequest) (*maps.GeocodingResponse, error) {
	if req.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	cacheKey := c.geocodeCacheKey(req.Address)
	if cached, ok := c.cache.Get(cacheKey); ok {
		if result, ok := cached.(*maps.GeocodingResponse); ok {
			return result, nil
		}
	}

	encodedAddress := url.QueryEscape(req.Address)

	apiURL := fmt.Sprintf("%s/geocoding/v5/mapbox.places/%s.json?access_token=%s&limit=1",
		baseURL, encodedAddress, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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

	var mapboxResp mapboxGeocodingResponse
	if err := json.Unmarshal(body, &mapboxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(mapboxResp.Features) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", req.Address)
	}

	result := c.convertGeocodingResponse(&mapboxResp.Features[0])

	c.cache.Set(cacheKey, result)

	return result, nil
}

// ReverseGeocode converts coordinates to an address
func (c *Client) ReverseGeocode(ctx context.Context, req *maps.ReverseGeocodingRequest) (*maps.GeocodingResponse, error) {
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, fmt.Errorf("invalid latitude: %f", req.Latitude)
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, fmt.Errorf("invalid longitude: %f", req.Longitude)
	}

	cacheKey := c.reverseGeocodeCacheKey(req.Latitude, req.Longitude)
	if cached, ok := c.cache.Get(cacheKey); ok {
		if result, ok := cached.(*maps.GeocodingResponse); ok {
			return result, nil
		}
	}

	apiURL := fmt.Sprintf("%s/geocoding/v5/mapbox.places/%f,%f.json?access_token=%s&limit=1",
		baseURL, req.Longitude, req.Latitude, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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

	var mapboxResp mapboxGeocodingResponse
	if err := json.Unmarshal(body, &mapboxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(mapboxResp.Features) == 0 {
		return nil, fmt.Errorf("no results found for coordinates: %f,%f", req.Latitude, req.Longitude)
	}

	result := c.convertGeocodingResponse(&mapboxResp.Features[0])

	c.cache.Set(cacheKey, result)

	return result, nil
}

// convertGeocodingResponse converts Mapbox feature to our format
func (c *Client) convertGeocodingResponse(feature *mapboxFeature) *maps.GeocodingResponse {
	var coord maps.Coordinate
	if len(feature.Center) >= 2 {
		coord = maps.Coordinate{
			Longitude: feature.Center[0],
			Latitude:  feature.Center[1],
		}
	} else if len(feature.Geometry.Coordinates) >= 2 {
		coord = maps.Coordinate{
			Longitude: feature.Geometry.Coordinates[0],
			Latitude:  feature.Geometry.Coordinates[1],
		}
	}

	placeType := ""
	if len(feature.PlaceType) > 0 {
		placeType = feature.PlaceType[0]
	}

	context := make([]string, len(feature.Context))
	for i, ctx := range feature.Context {
		context[i] = ctx.Text
	}

	return &maps.GeocodingResponse{
		Coordinate: coord,
		Address:    feature.PlaceName,
		PlaceName:  feature.PlaceName,
		PlaceType:  placeType,
		Relevance:  feature.Relevance,
		Context:    context,
	}
}

// geocodeCacheKey generates a cache key for geocoding
func (c *Client) geocodeCacheKey(address string) string {
	hash := sha256.Sum256([]byte("geocode:" + address))
	return fmt.Sprintf("%x", hash)
}

// reverseGeocodeCacheKey generates a cache key for reverse geocoding
func (c *Client) reverseGeocodeCacheKey(lat, lng float64) string {
	key := fmt.Sprintf("reverse:%f,%f", lat, lng)
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}
