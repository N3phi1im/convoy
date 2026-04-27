package mapbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"convoy/internal/maps"
)

func TestGeocode(t *testing.T) {
	tests := []struct {
		name        string
		request     *maps.GeocodingRequest
		mockResp    *mapboxGeocodingResponse
		expectError bool
		useCache    bool
	}{
		{
			name: "successful geocoding",
			request: &maps.GeocodingRequest{
				Address: "1600 Pennsylvania Avenue NW, Washington, DC",
			},
			mockResp: &mapboxGeocodingResponse{
				Type: "FeatureCollection",
				Features: []mapboxFeature{
					{
						PlaceName: "1600 Pennsylvania Avenue NW, Washington, DC 20500, United States",
						Center:    []float64{-77.0365, 38.8977},
						PlaceType: []string{"address"},
						Relevance: 0.99,
						Text:      "Pennsylvania Avenue NW",
						Context: []mapboxContext{
							{Text: "Washington"},
							{Text: "District of Columbia"},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty address",
			request: &maps.GeocodingRequest{
				Address: "",
			},
			expectError: true,
		},
		{
			name: "no results",
			request: &maps.GeocodingRequest{
				Address: "nonexistent place xyz123",
			},
			mockResp: &mapboxGeocodingResponse{
				Type:     "FeatureCollection",
				Features: []mapboxFeature{},
			},
			expectError: true,
		},
		{
			name: "cached result",
			request: &maps.GeocodingRequest{
				Address: "cached address",
			},
			mockResp: &mapboxGeocodingResponse{
				Type: "FeatureCollection",
				Features: []mapboxFeature{
					{
						PlaceName: "Cached Place",
						Center:    []float64{-122.4194, 37.7749},
						PlaceType: []string{"place"},
						Relevance: 1.0,
					},
				},
			},
			useCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if tt.mockResp != nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResp)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			originalBaseURL := baseURL
			baseURL = server.URL
			defer func() { baseURL = originalBaseURL }()

			client := NewClient(&Config{
				APIKey:       "test-key",
				CacheEnabled: tt.useCache,
			})

			ctx := context.Background()
			
			// First call
			resp, err := client.Geocode(ctx, tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}

			// Validate response
			if tt.mockResp != nil && len(tt.mockResp.Features) > 0 {
				feature := tt.mockResp.Features[0]
				if resp.PlaceName != feature.PlaceName {
					t.Errorf("expected place name %s, got %s", feature.PlaceName, resp.PlaceName)
				}
				if len(feature.Center) >= 2 {
					if resp.Coordinate.Longitude != feature.Center[0] {
						t.Errorf("expected longitude %f, got %f", feature.Center[0], resp.Coordinate.Longitude)
					}
					if resp.Coordinate.Latitude != feature.Center[1] {
						t.Errorf("expected latitude %f, got %f", feature.Center[1], resp.Coordinate.Latitude)
					}
				}
			}

			// Test caching
			if tt.useCache {
				// Second call should use cache
				resp2, err := client.Geocode(ctx, tt.request)
				if err != nil {
					t.Fatalf("unexpected error on cached call: %v", err)
				}
				if callCount > 1 {
					t.Error("expected cache to be used, but API was called again")
				}
				if resp2.PlaceName != resp.PlaceName {
					t.Error("cached response doesn't match original")
				}
			}
		})
	}
}

func TestReverseGeocode(t *testing.T) {
	tests := []struct {
		name        string
		request     *maps.ReverseGeocodingRequest
		mockResp    *mapboxGeocodingResponse
		expectError bool
	}{
		{
			name: "successful reverse geocoding",
			request: &maps.ReverseGeocodingRequest{
				Latitude:  38.8977,
				Longitude: -77.0365,
			},
			mockResp: &mapboxGeocodingResponse{
				Type: "FeatureCollection",
				Features: []mapboxFeature{
					{
						PlaceName: "1600 Pennsylvania Avenue NW, Washington, DC 20500, United States",
						Center:    []float64{-77.0365, 38.8977},
						PlaceType: []string{"address"},
						Relevance: 0.99,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid latitude",
			request: &maps.ReverseGeocodingRequest{
				Latitude:  91.0,
				Longitude: -77.0365,
			},
			expectError: true,
		},
		{
			name: "invalid longitude",
			request: &maps.ReverseGeocodingRequest{
				Latitude:  38.8977,
				Longitude: -181.0,
			},
			expectError: true,
		},
		{
			name: "no results",
			request: &maps.ReverseGeocodingRequest{
				Latitude:  0.0,
				Longitude: 0.0,
			},
			mockResp: &mapboxGeocodingResponse{
				Type:     "FeatureCollection",
				Features: []mapboxFeature{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResp != nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResp)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			originalBaseURL := baseURL
			baseURL = server.URL
			defer func() { baseURL = originalBaseURL }()

			client := NewClient(&Config{
				APIKey:       "test-key",
				CacheEnabled: true,
			})

			ctx := context.Background()
			resp, err := client.ReverseGeocode(ctx, tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}

			// Validate response
			if tt.mockResp != nil && len(tt.mockResp.Features) > 0 {
				feature := tt.mockResp.Features[0]
				if resp.PlaceName != feature.PlaceName {
					t.Errorf("expected place name %s, got %s", feature.PlaceName, resp.PlaceName)
				}
			}
		})
	}
}

func TestConvertGeocodingResponse(t *testing.T) {
	client := NewClient(&Config{APIKey: "test"})

	feature := &mapboxFeature{
		PlaceName: "San Francisco, California, United States",
		Center:    []float64{-122.4194, 37.7749},
		PlaceType: []string{"place"},
		Relevance: 0.95,
		Text:      "San Francisco",
		Context: []mapboxContext{
			{Text: "California"},
			{Text: "United States"},
		},
	}

	resp := client.convertGeocodingResponse(feature)

	if resp.PlaceName != feature.PlaceName {
		t.Errorf("expected place name %s, got %s", feature.PlaceName, resp.PlaceName)
	}
	if resp.Coordinate.Longitude != feature.Center[0] {
		t.Errorf("expected longitude %f, got %f", feature.Center[0], resp.Coordinate.Longitude)
	}
	if resp.Coordinate.Latitude != feature.Center[1] {
		t.Errorf("expected latitude %f, got %f", feature.Center[1], resp.Coordinate.Latitude)
	}
	if resp.PlaceType != "place" {
		t.Errorf("expected place type 'place', got %s", resp.PlaceType)
	}
	if resp.Relevance != 0.95 {
		t.Errorf("expected relevance 0.95, got %f", resp.Relevance)
	}
	if len(resp.Context) != 2 {
		t.Errorf("expected 2 context items, got %d", len(resp.Context))
	}
}

func TestGeocodingWithCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := &mapboxGeocodingResponse{
			Type: "FeatureCollection",
			Features: []mapboxFeature{
				{
					PlaceName: "Test Place",
					Center:    []float64{-122.0, 37.0},
					PlaceType: []string{"place"},
					Relevance: 1.0,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	client := NewClient(&Config{
		APIKey:       "test-key",
		CacheEnabled: true,
	})

	ctx := context.Background()
	req := &maps.GeocodingRequest{Address: "Test Address"}

	// First call - should hit API
	_, err := client.Geocode(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}

	// Second call - should use cache
	_, err = client.Geocode(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected cache to be used (still 1 API call), got %d calls", callCount)
	}

	// Different address - should hit API again
	req2 := &maps.GeocodingRequest{Address: "Different Address"}
	_, err = client.Geocode(ctx, req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}
