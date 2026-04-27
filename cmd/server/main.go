package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"convoy/internal/api"
	"convoy/internal/auth"
	"convoy/internal/config"
	"convoy/internal/maps"
	"convoy/internal/maps/mapbox"
	"convoy/internal/repository/postgres"
	"convoy/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger := initLogger(cfg)
	logger.Info("Starting Convoy server...")
	logger.Info(fmt.Sprintf("Environment: %s", cfg.Server.Env))

	db, err := postgres.NewDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	logger.Info("Database connected successfully")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	routeRepo := postgres.NewRouteRepository(db.DB.DB)
	waypointRepo := postgres.NewWaypointRepository(db.DB.DB)
	participantRepo := postgres.NewParticipantRepository(db.DB.DB)

	// Initialize services
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiry)
	authService := service.NewAuthService(userRepo, tokenManager)

	// Initialize map provider (Mapbox)
	var routeService *service.RouteService
	if cfg.Map.MapboxAPIKey != "" {
		mapProvider := initMapProvider(cfg, logger)
		routeService = service.NewRouteService(routeRepo, waypointRepo, participantRepo, mapProvider)
		logger.Info("Route service initialized with Mapbox provider")
	} else {
		logger.Info("Mapbox API key not configured, route service disabled")
		routeService = nil
	}

	participantService := service.NewParticipantService(participantRepo, routeRepo)

	router := api.NewRouter(cfg, authService, routeService, participantService)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info(fmt.Sprintf("Server listening on port %s", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
}

type SimpleLogger struct {
	logger *log.Logger
}

func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[ERROR] %s %v", msg, keysAndValues)
}

func (l *SimpleLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalf("[FATAL] %s %v", msg, keysAndValues)
}

func initLogger(cfg *config.Config) Logger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func initMapProvider(cfg *config.Config, logger Logger) maps.Provider {
	switch cfg.Map.Provider {
	case "mapbox":
		return mapbox.NewClient(&mapbox.Config{
			APIKey:       cfg.Map.MapboxAPIKey,
			Timeout:      10 * time.Second,
			RateLimit:    100,
			CacheEnabled: true,
			CacheTTL:     1 * time.Hour,
			MaxRetries:   3,
		})
	default:
		logger.Fatal("Unsupported map provider", "provider", cfg.Map.Provider)
		return nil
	}
}
