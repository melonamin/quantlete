package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
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
	cfg                *config.Config
	db                 *storage.DB
	stravaClient       *strava.Client
	authHandler        *handlers.AuthHandler
	activitiesHandler  *handlers.ActivitiesHandler
	importHandler      *handlers.ImportHandler
	dashboardHandler   *handlers.DashboardHandler
	goalsHandler       *handlers.GoalsHandler
	athleteHandler     *handlers.AthleteHandler
	statsHandler       *handlers.StatsHandler
	zonesHandler       *handlers.ZonesHandler
	settingsHandler    *handlers.SettingsHandler
	segmentsHandler    *handlers.SegmentsHandler
	gearHandler        *handlers.GearHandler
	maintenanceHandler *handlers.MaintenanceHandler
	photosHandler      *handlers.PhotosHandler
	challengesHandler  *handlers.ChallengesHandler
	exportHandler      *handlers.ExportHandler
}

// securityHeaders middleware adds security headers to all responses.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy for HTML responses
		// Allows: same-origin scripts/styles, inline styles for React, images from anywhere (for Strava badges),
		// and connections to Strava API
		if r.Header.Get("Accept") == "" || r.URL.Path == "/" || !isAPIPath(r.URL.Path) {
			csp := "default-src 'self'; " +
				"script-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' data:; " +
				"connect-src 'self' https://www.strava.com https://strava.com; " +
				"frame-ancestors 'none';"
			w.Header().Set("Content-Security-Policy", csp)
		}

		next.ServeHTTP(w, r)
	})
}

// isAPIPath checks if the request path is an API endpoint.
func isAPIPath(path string) bool {
	return len(path) >= 5 && path[:5] == "/api/"
}

// csrfProtection middleware validates Origin header for state-changing requests.
// This prevents CSRF attacks by ensuring requests come from the same origin.
func csrfProtection(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check state-changing methods
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check Origin header
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Fallback to Referer if Origin not present (some browsers don't send Origin)
				origin = r.Header.Get("Referer")
				if origin != "" {
					// Extract just the origin from referer URL
					if len(origin) > 0 {
						// Simple extraction: take everything up to the third slash
						slashCount := 0
						for i, c := range origin {
							if c == '/' {
								slashCount++
								if slashCount == 3 {
									origin = origin[:i]
									break
								}
							}
						}
					}
				}
			}

			// If no origin provided and it's an API request with auth token, allow it
			// Token-based auth is inherently CSRF-resistant since tokens must be explicitly sent
			if origin == "" {
				if r.Header.Get("Authorization") != "" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Validate origin against allowlist
			if origin != "" && !originSet[origin] {
				http.Error(w, "CSRF validation failed: invalid origin", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
	r.Use(securityHeaders)

	// Configure allowed origins for CORS and CSRF
	var allowedOrigins []string
	if cfg.Server.DevMode {
		allowedOrigins = []string{"http://localhost:5173", "http://localhost:8081"}
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	} else {
		// In production, allow same-origin requests
		// The frontend is embedded and served from the same origin
		host := cfg.Server.Host
		if host == "" {
			host = "localhost"
		}
		origin := "http://" + host
		if cfg.Server.Port != 80 && cfg.Server.Port != 0 {
			origin = fmt.Sprintf("http://%s:%d", host, cfg.Server.Port)
		}
		allowedOrigins = []string{origin}
	}

	// CSRF protection for state-changing requests
	r.Use(csrfProtection(allowedOrigins))

	// Create repositories
	activityRepo := storage.NewActivityRepository(db)
	athleteRepo := storage.NewAthleteRepository(db)
	tokenRepo := storage.NewTokenRepository(db)
	statsRepo := storage.NewStatsRepository(db)
	gearRepo := storage.NewGearRepository(db)
	streamRepo := storage.NewStreamRepository(db)
	dashboardConfigRepo := storage.NewDashboardConfigRepository(db)
	goalsRepo := storage.NewGoalsRepository(db)
	metricsRepo := storage.NewAthleteMetricsRepository(db)
	segmentRepo := storage.NewSegmentRepository(db)
	zonesRepo := storage.NewZonesRepository(db)
	powerRepo := storage.NewPowerRepository(db, streamRepo)
	bestEffortsRepo := storage.NewBestEffortsRepository(db)
	trainingLoadRepo := storage.NewTrainingLoadRepository(db, streamRepo, metricsRepo, zonesRepo)
	settingsRepo := storage.NewSettingsRepository(db)
	maintenanceRepo := storage.NewMaintenanceRepository(db)
	photoRepo := storage.NewPhotoRepository(db)
	challengeRepo := storage.NewChallengeRepository(db)
	appStateRepo := storage.NewAppStateRepository(db)
	syncHistoryRepo := storage.NewSyncHistoryRepository(db, appStateRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg, stravaClient, tokenRepo, athleteRepo)
	activitiesHandler := handlers.NewActivitiesHandler(activityRepo, streamRepo, stravaClient)
	importHandler := handlers.NewImportHandler(imp, syncHistoryRepo, stravaClient)
	dashboardHandler := handlers.NewDashboardHandler(statsRepo, dashboardConfigRepo, stravaClient)
	goalsHandler := handlers.NewGoalsHandler(goalsRepo, stravaClient)
	athleteHandler := handlers.NewAthleteHandler(metricsRepo, stravaClient)
	statsHandler := handlers.NewStatsHandler(db, statsRepo, powerRepo, streamRepo, metricsRepo, zonesRepo, bestEffortsRepo, trainingLoadRepo, stravaClient)
	zonesHandler := handlers.NewZonesHandler(zonesRepo, stravaClient)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, stravaClient)
	segmentsHandler := handlers.NewSegmentsHandler(segmentRepo, stravaClient)
	gearHandler := handlers.NewGearHandler(gearRepo, stravaClient)
	maintenanceHandler := handlers.NewMaintenanceHandler(maintenanceRepo, stravaClient)
	photosHandler := handlers.NewPhotosHandler(photoRepo, stravaClient)
	challengesHandler := handlers.NewChallengesHandler(challengeRepo, stravaClient, cfg.Storage.DataDir)
	exportHandler := handlers.NewExportHandler(activityRepo, stravaClient)

	router := &Router{
		Mux:                r,
		cfg:                cfg,
		db:                 db,
		stravaClient:       stravaClient,
		authHandler:        authHandler,
		activitiesHandler:  activitiesHandler,
		importHandler:      importHandler,
		dashboardHandler:   dashboardHandler,
		goalsHandler:       goalsHandler,
		athleteHandler:     athleteHandler,
		statsHandler:       statsHandler,
		zonesHandler:       zonesHandler,
		settingsHandler:    settingsHandler,
		segmentsHandler:    segmentsHandler,
		gearHandler:        gearHandler,
		maintenanceHandler: maintenanceHandler,
		photosHandler:      photosHandler,
		challengesHandler:  challengesHandler,
		exportHandler:      exportHandler,
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
			router.Get("/{id}/streams", r.activitiesHandler.GetStreams)
			router.Get("/{id}/photos", r.photosHandler.ActivityPhotos)
		})

		// Import routes
		router.Route("/import", func(router chi.Router) {
			router.Post("/start", r.importHandler.Start)
			router.Get("/progress", r.importHandler.Progress)
			router.Post("/cancel", r.importHandler.Cancel)
			router.Get("/history", r.importHandler.History)
			router.Get("/watermark", r.importHandler.Watermark)
		})

		// Dashboard routes
		router.Route("/dashboard", func(router chi.Router) {
			router.Get("/", r.dashboardHandler.GetDashboard)
			router.Get("/stats", r.dashboardHandler.GetStats)
			router.Get("/weekly", r.dashboardHandler.GetWeeklyStats)
			router.Get("/recent", r.dashboardHandler.GetRecentActivities)
			router.Get("/sports", r.dashboardHandler.GetSportTypeStats)
			router.Get("/config", r.dashboardHandler.GetDashboardConfig)
			router.Put("/config", r.dashboardHandler.UpdateDashboardConfig)
			router.Get("/monthly", r.dashboardHandler.GetMonthlyStats)
			router.Get("/yearly", r.dashboardHandler.GetYearlyStats)
			router.Get("/calendar", r.dashboardHandler.GetCalendarData)
			router.Get("/calendar/summary", r.dashboardHandler.GetCalendarSummary)
			router.Get("/calendar/activities", r.dashboardHandler.GetCalendarActivities)
		})

		// Stats routes (heatmap, eddington, etc.)
		router.Route("/stats", func(router chi.Router) {
			router.Get("/heatmap", r.statsHandler.GetHeatmapData)
			router.Get("/eddington", r.statsHandler.GetEddingtonData)
			router.Get("/eddington/history", r.statsHandler.GetEddingtonHistory)
			router.Get("/best-efforts", r.statsHandler.GetBestEfforts)
			router.Get("/best-efforts/{distanceType}", r.statsHandler.GetBestEffortsByDistance)
			router.Get("/rewind", r.statsHandler.GetRewind)
			router.Get("/rewind/years", r.statsHandler.GetRewindYears)
			router.Get("/power", r.statsHandler.GetPowerStats)
			router.Get("/power-zones", r.statsHandler.GetPowerZones)
			router.Get("/hr-zones", r.statsHandler.GetHRZones)
			router.Get("/training-load", r.statsHandler.GetTrainingLoad)
			router.Get("/daytime", r.statsHandler.GetDaytimeDistribution)
			router.Get("/weekday", r.statsHandler.GetWeekdayDistribution)
		})

		// Gear routes
		router.Route("/gear", func(router chi.Router) {
			router.Get("/", r.gearHandler.List)
			router.Get("/custom", r.gearHandler.ListCustom)
			router.Post("/custom", r.gearHandler.CreateCustom)
			router.Put("/custom/{id}", r.gearHandler.UpdateCustom)
			router.Delete("/custom/{id}", r.gearHandler.DeleteCustom)
			router.Get("/stats/monthly", r.gearHandler.MonthlyUsage)
			router.Get("/{id}", r.gearHandler.GetByID)
			router.Get("/{id}/components", r.maintenanceHandler.ListGearComponents)
			router.Post("/{id}/components", r.maintenanceHandler.CreateGearComponent)
		})

		// Maintenance routes
		router.Route("/components", func(router chi.Router) {
			router.Put("/{id}", r.maintenanceHandler.UpdateComponent)
			router.Delete("/{id}", r.maintenanceHandler.DeleteComponent)
			router.Post("/{id}/maintenance", r.maintenanceHandler.LogMaintenance)
		})
		router.Route("/maintenance", func(router chi.Router) {
			router.Get("/due", r.maintenanceHandler.Due)
		})

		// Photos routes
		router.Route("/photos", func(router chi.Router) {
			router.Get("/", r.photosHandler.List)
		})

		// Challenges routes
		router.Route("/challenges", func(router chi.Router) {
			router.Get("/", r.challengesHandler.List)
			router.Post("/import", r.challengesHandler.Import)
			router.Post("/import-profile", r.challengesHandler.ImportFromProfile)
		})

		// Segments routes
		router.Route("/segments", func(router chi.Router) {
			router.Get("/", r.segmentsHandler.List)
			router.Get("/countries", r.segmentsHandler.Countries)
			router.Get("/{id}", r.segmentsHandler.GetByID)
			router.Get("/{id}/efforts", r.segmentsHandler.ListEfforts)
		})

		// Goals routes
		router.Route("/goals", func(router chi.Router) {
			router.Get("/", r.goalsHandler.GetGoals)
			router.Put("/", r.goalsHandler.UpdateGoals)
		})

		// Athlete metrics routes
		router.Route("/athlete", func(router chi.Router) {
			router.Get("/ftp", r.athleteHandler.GetFTP)
			router.Put("/ftp", r.athleteHandler.UpdateFTP)
			router.Get("/weight", r.athleteHandler.GetWeight)
			router.Put("/weight", r.athleteHandler.UpdateWeight)
		})

		// Zones configuration routes
		router.Route("/zones", func(router chi.Router) {
			router.Route("/hr", func(router chi.Router) {
				router.Get("/", r.zonesHandler.ListHR)
				router.Put("/", r.zonesHandler.UpsertHR)
				router.Delete("/", r.zonesHandler.DeleteHR)
			})
		})

		// Settings routes
		router.Route("/settings", func(router chi.Router) {
			router.Get("/", r.settingsHandler.Get)
			router.Put("/", r.settingsHandler.Update)
		})

		// Export routes
		router.Route("/export", func(router chi.Router) {
			router.Get("/stats", r.exportHandler.ExportStats)
			router.Get("/activities/csv", r.exportHandler.ExportActivitiesCSV)
			router.Get("/activities/json", r.exportHandler.ExportActivitiesJSON)
		})
	})

	// Serve locally stored challenge badge images
	// URL: /files/challenges/{filename} -> {dataDir}/challenges/{filename}
	r.Get("/files/challenges/*", r.serveDataFile)

	// Static file serving (placeholder for embedded files)
	r.Get("/*", handlers.ServeFrontend)
}

// serveDataFile serves static files from the data directory.
func (r *Router) serveDataFile(w http.ResponseWriter, req *http.Request) {
	// Extract the file path from the URL
	urlPath := strings.TrimPrefix(req.URL.Path, "/files/")
	if urlPath == "" || strings.Contains(urlPath, "..") {
		http.NotFound(w, req)
		return
	}

	// Construct the full file path
	filePath := filepath.Join(r.cfg.Storage.DataDir, urlPath)

	// Serve the file
	http.ServeFile(w, req, filePath)
}
