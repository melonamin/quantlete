package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/sasha/stata/internal/api/handlers"
	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/importer"
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
	importHandler     *handlers.ImportHandler
	dashboardHandler  *handlers.DashboardHandler
	gearHandler       *handlers.GearHandler
}

// NewRouter creates a new HTTP router with all routes configured.
func NewRouter(cfg *config.Config, stravaClient *strava.Client, db *storage.DB, imp *importer.Importer) *Router {
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
	athleteRepo := storage.NewAthleteRepository(db)
	tokenRepo := storage.NewTokenRepository(db)
	statsRepo := storage.NewStatsRepository(db)
	gearRepo := storage.NewGearRepository(db)

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg, stravaClient, tokenRepo, athleteRepo)
	activitiesHandler := handlers.NewActivitiesHandler(activityRepo, stravaClient)
	importHandler := handlers.NewImportHandler(imp)
	dashboardHandler := handlers.NewDashboardHandler(statsRepo, stravaClient)
	gearHandler := handlers.NewGearHandler(gearRepo, stravaClient)

	router := &Router{
		Mux:               r,
		cfg:               cfg,
		db:                db,
		stravaClient:      stravaClient,
		authHandler:       authHandler,
		activitiesHandler: activitiesHandler,
		importHandler:     importHandler,
		dashboardHandler:  dashboardHandler,
		gearHandler:       gearHandler,
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

		// Import routes
		router.Route("/import", func(router chi.Router) {
			router.Post("/start", r.importHandler.Start)
			router.Get("/progress", r.importHandler.Progress)
			router.Post("/cancel", r.importHandler.Cancel)
		})

		// Dashboard routes
		router.Route("/dashboard", func(router chi.Router) {
			router.Get("/", r.dashboardHandler.GetDashboard)
			router.Get("/stats", r.dashboardHandler.GetStats)
			router.Get("/weekly", r.dashboardHandler.GetWeeklyStats)
			router.Get("/recent", r.dashboardHandler.GetRecentActivities)
			router.Get("/sports", r.dashboardHandler.GetSportTypeStats)
			router.Get("/monthly", r.dashboardHandler.GetMonthlyStats)
			router.Get("/yearly", r.dashboardHandler.GetYearlyStats)
			router.Get("/calendar", r.dashboardHandler.GetCalendarData)
			router.Get("/calendar/activities", r.dashboardHandler.GetCalendarActivities)
		})

		// Stats routes (heatmap, eddington, etc.)
		router.Route("/stats", func(router chi.Router) {
			router.Get("/heatmap", r.dashboardHandler.GetHeatmapData)
			router.Get("/eddington", r.dashboardHandler.GetEddingtonData)
		})

		// Gear routes
		router.Route("/gear", func(router chi.Router) {
			router.Get("/", r.gearHandler.List)
			router.Get("/{id}", r.gearHandler.GetByID)
		})
	})

	// Static file serving (placeholder for embedded files)
	r.Get("/*", handlers.ServeFrontend)
}
