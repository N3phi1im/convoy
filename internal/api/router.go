package api

import (
	"net/http"
	"time"

	"convoy/internal/api/handlers"
	"convoy/internal/api/middleware"
	"convoy/internal/config"
	"convoy/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Router struct {
	chi.Router
}

func NewRouter(cfg *config.Config, authService *service.AuthService, routeService *service.RouteService, participantService *service.ParticipantService) *Router {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)
	participantHandler := handlers.NewParticipantHandler(participantService)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"convoy"}`))
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authService))

			r.Get("/users/me", userHandler.GetCurrentUser)

			if routeService != nil {
				routeHandler := handlers.NewRouteHandler(routeService)
				r.Get("/routes", routeHandler.ListRoutes)
				r.Post("/routes", routeHandler.CreateRoute)
				r.Get("/routes/{id}", routeHandler.GetRoute)
				r.Put("/routes/{id}", routeHandler.UpdateRoute)
				r.Delete("/routes/{id}", routeHandler.DeleteRoute)
			} else {
				r.Get("/routes", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(`{"success":false,"error":"route service not configured - check MAPBOX_API_KEY"}`))
				})
			}

			r.Get("/routes/{id}/participants", participantHandler.ListParticipants)
			r.Post("/routes/{id}/join", participantHandler.JoinRoute)
			r.Post("/routes/{id}/leave", participantHandler.LeaveRoute)
			r.Delete("/routes/{id}/participants/{userId}", participantHandler.RemoveParticipant)
		})
	})

	return &Router{Router: r}
}
