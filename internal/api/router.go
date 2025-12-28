package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/sasha/stata/internal/api/handlers"
	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// Router holds the HTTP router and its dependencies.
type Router struct {
	*chi.Mux
	cfg               *config.Config
	db                *storage.DB
	stravaClient      *strava.Client
	authHandler       *handlers.AuthHandler
	activitiesHandler *handlers.ActivitiesHandler
}

// NewRouter creates a new HTTP router with all routes configured.
func NewRouter(cfg *config.Config, stravaClient *strava.Client, db *storage.DB) *Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS for development
	if cfg.Server.DevMode {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8081"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Create repositories
	activityRepo := storage.NewActivityRepository(db)

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg, stravaClient)
	activitiesHandler := handlers.NewActivitiesHandler(activityRepo, stravaClient)

	router := &Router{
		Mux:               r,
		cfg:               cfg,
		db:                db,
		stravaClient:      stravaClient,
		authHandler:       authHandler,
		activitiesHandler: activitiesHandler,
	}

	// Mount routes
	router.mountRoutes()

	return router
}

func (r *Router) mountRoutes() {
	// API routes
	r.Route("/api/v1", func(router chi.Router) {
		// Health check
		router.Get("/health", handlers.HealthCheck)

		// Auth routes
		router.Route("/auth", func(router chi.Router) {
			router.Get("/strava", r.authHandler.InitiateOAuth)
			router.Get("/strava/callback", r.authHandler.HandleCallback)
			router.Get("/status", r.authHandler.Status)
			router.Post("/refresh", r.authHandler.RefreshToken)
		})

		// Activities routes
		router.Route("/activities", func(router chi.Router) {
			router.Get("/", r.activitiesHandler.List)
			router.Get("/{id}", r.activitiesHandler.GetByID)
		})
	})

	// Static file serving (placeholder for embedded files)
	r.Get("/*", handlers.ServeFrontend)
}
