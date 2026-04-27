package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Map       MapConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Logging   LoggingConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type MapConfig struct {
	Provider         string
	MapboxAPIKey     string
	GoogleMapsAPIKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	RequestsPerMinute     int
	AuthRequestsPerMinute int
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
	setDefaults()

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			Expiry: viper.GetDuration("JWT_EXPIRY"),
		},
		Map: MapConfig{
			Provider:         viper.GetString("MAP_PROVIDER"),
			MapboxAPIKey:     viper.GetString("MAPBOX_API_KEY"),
			GoogleMapsAPIKey: viper.GetString("GOOGLE_MAPS_API_KEY"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseAllowedOrigins(viper.GetString("ALLOWED_ORIGINS")),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute:     viper.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),
			AuthRequestsPerMinute: viper.GetInt("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	// Split by comma and trim whitespace
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func setDefaults() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("MAP_PROVIDER", "mapbox")
	viper.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:3000"})
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	viper.SetDefault("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE", 5)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
}

func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if c.Database.Host == "" || c.Database.Name == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.Map.Provider == "mapbox" && c.Map.MapboxAPIKey == "" {
		return fmt.Errorf("MAPBOX_API_KEY is required when using mapbox provider")
	}

	return nil
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
