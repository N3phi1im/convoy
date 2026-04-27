package mapbox

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "default config",
			config: &Config{
				APIKey: "test-key",
			},
		},
		{
			name: "custom config",
			config: &Config{
				APIKey:       "test-key",
				Timeout:      5 * time.Second,
				RateLimit:    50,
				CacheEnabled: true,
				CacheTTL:     30 * time.Minute,
				MaxRetries:   5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if client.apiKey != tt.config.APIKey {
				t.Errorf("expected API key %s, got %s", tt.config.APIKey, client.apiKey)
			}
			if client.httpClient == nil {
				t.Error("expected non-nil HTTP client")
			}
			if client.rateLimiter == nil {
				t.Error("expected non-nil rate limiter")
			}
			if client.cache == nil {
				t.Error("expected non-nil cache")
			}
		})
	}
}

func TestCache(t *testing.T) {
	cache := &Cache{
		data:    make(map[string]*cacheEntry),
		ttl:     100 * time.Millisecond,
		enabled: true,
	}

	t.Run("set and get", func(t *testing.T) {
		cache.Set("key1", "value1")
		val, ok := cache.Get("key1")
		if !ok {
			t.Error("expected to find key1")
		}
		if val != "value1" {
			t.Errorf("expected value1, got %v", val)
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, ok := cache.Get("nonexistent")
		if ok {
			t.Error("expected not to find nonexistent key")
		}
	})

	t.Run("expiration", func(t *testing.T) {
		cache.Set("expiring", "value")
		time.Sleep(150 * time.Millisecond)
		_, ok := cache.Get("expiring")
		if ok {
			t.Error("expected key to be expired")
		}
	})

	t.Run("disabled cache", func(t *testing.T) {
		disabledCache := &Cache{
			data:    make(map[string]*cacheEntry),
			ttl:     1 * time.Hour,
			enabled: false,
		}
		disabledCache.Set("key", "value")
		_, ok := disabledCache.Get("key")
		if ok {
			t.Error("expected disabled cache to not store values")
		}
	})

	t.Run("clear expired", func(t *testing.T) {
		cache.Set("key1", "value1")
		cache.Set("key2", "value2")
		time.Sleep(150 * time.Millisecond)
		cache.Clear()
		
		if len(cache.data) > 0 {
			t.Errorf("expected cache to be empty after clear, got %d entries", len(cache.data))
		}
	})
}

func TestRateLimiter(t *testing.T) {
	config := &Config{
		APIKey:    "test-key",
		RateLimit: 60,
	}
	client := NewClient(config)

	if client.rateLimiter == nil {
		t.Fatal("expected non-nil rate limiter")
	}

	ctx := context.Background()
	
	// Verify rate limiter allows requests
	if err := client.rateLimiter.Wait(ctx); err != nil {
		t.Fatalf("rate limiter error: %v", err)
	}
	
	// Verify it respects context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := client.rateLimiter.Wait(cancelCtx); err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestCacheKey(t *testing.T) {
	client := NewClient(&Config{APIKey: "test"})

	t.Run("geocode cache key", func(t *testing.T) {
		key1 := client.geocodeCacheKey("123 Main St")
		key2 := client.geocodeCacheKey("123 Main St")
		key3 := client.geocodeCacheKey("456 Oak Ave")

		if key1 != key2 {
			t.Error("expected same address to generate same cache key")
		}
		if key1 == key3 {
			t.Error("expected different addresses to generate different cache keys")
		}
	})

	t.Run("reverse geocode cache key", func(t *testing.T) {
		key1 := client.reverseGeocodeCacheKey(40.7128, -74.0060)
		key2 := client.reverseGeocodeCacheKey(40.7128, -74.0060)
		key3 := client.reverseGeocodeCacheKey(34.0522, -118.2437)

		if key1 != key2 {
			t.Error("expected same coordinates to generate same cache key")
		}
		if key1 == key3 {
			t.Error("expected different coordinates to generate different cache keys")
		}
	})
}
