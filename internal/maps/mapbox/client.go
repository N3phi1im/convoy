package mapbox

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	baseURL = "https://api.mapbox.com"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultRateLimit = 100
	maxRetries       = 3
	retryDelay       = 1 * time.Second
)

type Client struct {
	httpClient  *http.Client
	apiKey      string
	rateLimiter *rate.Limiter
	cache       *Cache
}

type Config struct {
	APIKey       string
	Timeout      time.Duration
	RateLimit    int
	CacheEnabled bool
	CacheTTL     time.Duration
	MaxRetries   int
}

type Cache struct {
	mu      sync.RWMutex
	data    map[string]*cacheEntry
	ttl     time.Duration
	enabled bool
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.RateLimit == 0 {
		config.RateLimit = defaultRateLimit
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = maxRetries
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}

	rps := float64(config.RateLimit) / 60.0
	limiter := rate.NewLimiter(rate.Limit(rps), config.RateLimit)

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:      config.APIKey,
		rateLimiter: limiter,
		cache: &Cache{
			data:    make(map[string]*cacheEntry),
			ttl:     config.CacheTTL,
			enabled: config.CacheEnabled,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(retryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			continue
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			defer resp.Body.Close()
			return nil, fmt.Errorf("client error: status %d", resp.StatusCode)
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, err)
	}

	return nil, fmt.Errorf("request failed after %d retries", maxRetries)
}

// Cache methods

func (c *Cache) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Clear() {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiration) {
			delete(c.data, key)
		}
	}
}

func (c *Cache) StartCleanup(interval time.Duration) {
	if !c.enabled {
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.Clear()
		}
	}()
}
