package mapbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"convoy/internal/maps"
)

func TestGetDirections(t *testing.T) {
	tests := []struct {
		name        string
		request     *maps.DirectionsRequest
		mockResp    *mapboxDirectionsResponse
		expectError bool
	}{
		{
			name: "successful directions request",
			request: &maps.DirectionsRequest{
				Waypoints: []maps.Coordinate{
					{Latitude: 40.7128, Longitude: -74.0060},
					{Latitude: 34.0522, Longitude: -118.2437},
				},
				Profile: "driving",
			},
			mockResp: &mapboxDirectionsResponse{
				Code: "Ok",
				Routes: []mapboxRoute{
					{
						Distance: 4500000.0,
						Duration: 144000.0,
						Geometry: "encodedPolyline",
						Legs: []mapboxLeg{
							{
								Distance: 4500000.0,
								Duration: 144000.0,
								Steps: []mapboxStep{
									{
										Distance: 100.0,
										Duration: 10.0,
										Name:     "Main Street",
										Maneuver: mapboxManeuver{
											Location:    []float64{-74.0060, 40.7128},
											Instruction: "Turn left",
											Type:        "turn",
										},
									},
								},
							},
						},
					},
				},
				Waypoints: []mapboxWaypoint{
					{Name: "Start", Location: []float64{-74.0060, 40.7128}},
					{Name: "End", Location: []float64{-118.2437, 34.0522}},
				},
			},
			expectError: false,
		},
		{
			name: "insufficient waypoints",
			request: &maps.DirectionsRequest{
				Waypoints: []maps.Coordinate{
					{Latitude: 40.7128, Longitude: -74.0060},
				},
				Profile: "driving",
			},
			expectError: true,
		},
		{
			name: "mapbox error response",
			request: &maps.DirectionsRequest{
				Waypoints: []maps.Coordinate{
					{Latitude: 40.7128, Longitude: -74.0060},
					{Latitude: 34.0522, Longitude: -118.2437},
				},
				Profile: "driving",
			},
			mockResp: &mapboxDirectionsResponse{
				Code:    "InvalidInput",
				Message: "Invalid coordinates",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResp != nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResp)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			// Override baseURL for testing
			originalBaseURL := baseURL
			baseURL = server.URL
			defer func() { baseURL = originalBaseURL }()

			// Create client
			client := NewClient(&Config{
				APIKey:       "test-key",
				CacheEnabled: false,
			})

			// Execute request
			ctx := context.Background()
			resp, err := client.GetDirections(ctx, tt.request)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Validate response
			if resp == nil {
				t.Fatal("expected non-nil response")
			}

			if tt.mockResp != nil && len(tt.mockResp.Routes) > 0 {
				expectedRoute := tt.mockResp.Routes[0]
				if resp.Distance != expectedRoute.Distance {
					t.Errorf("expected distance %f, got %f", expectedRoute.Distance, resp.Distance)
				}
				if resp.Duration != expectedRoute.Duration {
					t.Errorf("expected duration %f, got %f", expectedRoute.Duration, resp.Duration)
				}
				if resp.Geometry != expectedRoute.Geometry {
					t.Errorf("expected geometry %s, got %s", expectedRoute.Geometry, resp.Geometry)
				}
			}
		})
	}
}

func TestConvertDirectionsResponse(t *testing.T) {
	client := NewClient(&Config{APIKey: "test"})

	mapboxResp := &mapboxDirectionsResponse{
		Code: "Ok",
		Routes: []mapboxRoute{
			{
				Distance: 1000.0,
				Duration: 300.0,
				Geometry: "test_polyline",
				Legs: []mapboxLeg{
					{
						Steps: []mapboxStep{
							{
								Distance: 500.0,
								Duration: 150.0,
								Maneuver: mapboxManeuver{
									Location:    []float64{-74.0060, 40.7128},
									Instruction: "Turn right",
								},
							},
						},
					},
				},
			},
		},
		Waypoints: []mapboxWaypoint{
			{Name: "Point A", Location: []float64{-74.0060, 40.7128}},
		},
	}

	resp, err := client.convertDirectionsResponse(mapboxResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Distance != 1000.0 {
		t.Errorf("expected distance 1000.0, got %f", resp.Distance)
	}
	if resp.Duration != 300.0 {
		t.Errorf("expected duration 300.0, got %f", resp.Duration)
	}
	if len(resp.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(resp.Steps))
	}
	if len(resp.Waypoints) != 1 {
		t.Errorf("expected 1 waypoint, got %d", len(resp.Waypoints))
	}
}

func TestDirectionsProfiles(t *testing.T) {
	profiles := []string{"driving", "cycling", "walking", "running"}
	
	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify profile is in URL
				if profile == "running" {
					// Running should use walking profile
					if !contains(r.URL.Path, "walking") {
						t.Errorf("expected walking profile for running, got %s", r.URL.Path)
					}
				} else {
					if !contains(r.URL.Path, profile) {
						t.Errorf("expected %s profile in URL, got %s", profile, r.URL.Path)
					}
				}
				
				resp := &mapboxDirectionsResponse{
					Code: "Ok",
					Routes: []mapboxRoute{{Distance: 1000, Duration: 300}},
				}
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			originalBaseURL := baseURL
			baseURL = server.URL
			defer func() { baseURL = originalBaseURL }()

			client := NewClient(&Config{APIKey: "test"})
			req := &maps.DirectionsRequest{
				Waypoints: []maps.Coordinate{
					{Latitude: 40.7128, Longitude: -74.0060},
					{Latitude: 34.0522, Longitude: -118.2437},
				},
				Profile: profile,
			}

			_, err := client.GetDirections(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
