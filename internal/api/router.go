package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/melonamin/quantlete/internal/api/handlers"
	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// badgeRateLimitRequests is the maximum number of badge requests per IP per window.
// Badges are public endpoints that need protection against abuse while remaining
// usable for embedding in READMEs and profiles.
const badgeRateLimitRequests = 60

// badgeRateLimitWindow is the time window for badge rate limiting.
const badgeRateLimitWindow = time.Minute

// Router holds the HTTP router and its dependencies.
type Router struct {
	*chi.Mux
	cfg          *config.Config
	db           *storage.DB
	stravaClient *strava.Client
	registry     *services.ServiceRegistry

	// Handlers for routes not yet using generated adapters
	webhooksHandler    *handlers.StravaWebhookHandler
	authHandler        *handlers.AuthHandler
	importHandler      *handlers.ImportHandler
	dashboardHandler   *handlers.DashboardHandler
	goalsHandler       *handlers.GoalsHandler
	athleteHandler     *handlers.AthleteHandler
	statsHandler       *handlers.StatsHandler
	zonesHandler       *handlers.ZonesHandler
	settingsHandler    *handlers.SettingsHandler
	segmentsHandler    *handlers.SegmentsHandler
	gearHandler        *handlers.GearHandler // for custom gear operations
	maintenanceHandler *handlers.MaintenanceHandler
	photosHandler      *handlers.PhotosHandler
	challengesHandler  *handlers.ChallengesHandler
	exportHandler      *handlers.ExportHandler
	weatherHandler     *handlers.WeatherHandler
	setupHandler       *handlers.SetupHandler
	badgesHandler      *handlers.BadgesHandler
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

	// Create service registry (shared with WASM)
	registry := services.NewServiceRegistry(db)

	// Server-only repositories (OAuth, weather API)
	tokenRepo := storage.NewTokenRepository(db)
	weatherRepo := storage.NewWeatherRepository(db.Conn())

	// Create handlers using registry
	// Note: Activities and Gear List/GetByID/MonthlyUsage use generated handlers from adapters.gen.go
	webhooksHandler := handlers.NewStravaWebhookHandler(cfg, imp, registry.Activities(), registry.Settings(), stravaClient)
	authHandler := handlers.NewAuthHandler(cfg, stravaClient, tokenRepo, registry.Athletes())
	importHandler := handlers.NewImportHandler(imp, registry.SyncHistory(), stravaClient)
	importHandler.SetAllowedOrigins(allowedOrigins) // Configure CORS for SSE endpoint
	dashboardHandler := handlers.NewDashboardHandler(registry.DashboardService, stravaClient)
	goalsHandler := handlers.NewGoalsHandler(registry.Goals(), stravaClient)
	athleteHandler := handlers.NewAthleteHandler(registry.AthleteMetrics(), stravaClient)
	statsHandler := handlers.NewStatsHandler(registry.StatsService, db, registry.Streams(), registry.AthleteMetrics(), registry.Zones(), stravaClient)
	zonesHandler := handlers.NewZonesHandler(registry.Zones(), stravaClient)
	settingsHandler := handlers.NewSettingsHandler(registry.Settings(), stravaClient)
	segmentsHandler := handlers.NewSegmentsHandler(registry.SegmentsService, stravaClient)
	gearHandler := handlers.NewGearHandler(registry.GearService, stravaClient) // for custom gear operations
	maintenanceHandler := handlers.NewMaintenanceHandler(registry.MaintenanceService, stravaClient)
	photosHandler := handlers.NewPhotosHandler(registry.PhotosService, stravaClient)
	challengesHandler := handlers.NewChallengesHandler(registry.ChallengesService, stravaClient, cfg.Storage.DataDir)
	exportHandler := handlers.NewExportHandler(registry.Activities(), stravaClient)
	weatherHandler := handlers.NewWeatherHandler(weatherRepo, registry.Activities(), registry.Streams(), stravaClient, slog.Default())
	setupHandler := handlers.NewSetupHandler(cfg, registry.AppState(), stravaClient)
	badgesHandler := handlers.NewBadgesHandler(registry.Stats(), registry.Settings(), stravaClient)

	router := &Router{
		Mux:          r,
		cfg:          cfg,
		db:           db,
		stravaClient: stravaClient,
		registry:     registry,

		webhooksHandler:    webhooksHandler,
		authHandler:        authHandler,
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
		weatherHandler:     weatherHandler,
		setupHandler:       setupHandler,
		badgesHandler:      badgesHandler,
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

		// Webhooks
		router.Route("/webhooks", func(router chi.Router) {
			router.Get("/strava", r.webhooksHandler.Validate)
			router.Post("/strava", r.webhooksHandler.Receive)
		})

		// Auth routes
		router.Route("/auth", func(router chi.Router) {
			router.Get("/strava", r.authHandler.InitiateOAuth)
			router.Get("/strava/callback", r.authHandler.HandleCallback)
			router.Get("/status", r.authHandler.Status)
			router.Post("/refresh", r.authHandler.RefreshToken)
		})

		// Setup routes (for credential configuration)
		router.Route("/setup", func(router chi.Router) {
			router.Get("/credentials", r.setupHandler.GetCredentialsStatus)
			router.Put("/credentials", r.setupHandler.UpdateCredentials)
		})

		// Activities routes (using generated adapters)
		router.Route("/activities", func(router chi.Router) {
			router.Get("/", handlers.GenActivityServiceList(r.registry.ActivityService, r.stravaClient))
			router.Get("/{id}", handlers.GenActivityServiceGetByID(r.registry.ActivityService, r.stravaClient))
			router.Get("/{id}/streams", handlers.GenActivityServiceGetStreams(r.registry.ActivityService, r.stravaClient))
			router.Get("/{id}/photos", handlers.GenPhotosServiceListByActivity(r.registry.PhotosService, r.stravaClient))
			router.Get("/{id}/weather", r.weatherHandler.GetActivityWeather) // Manual: weather API integration
		})

		// Import routes
		router.Route("/import", func(router chi.Router) {
			router.Post("/start", r.importHandler.Start)
			router.Get("/progress", r.importHandler.Progress)
			router.Get("/events", r.importHandler.Events) // SSE endpoint for real-time updates
			router.Post("/cancel", r.importHandler.Cancel)
			router.Post("/pause", r.importHandler.Pause)
			router.Post("/resume", r.importHandler.Resume)
			router.Get("/history", r.importHandler.History)
			router.Get("/watermark", r.importHandler.Watermark)
		})

		// Dashboard routes (using generated adapters)
		router.Route("/dashboard", func(router chi.Router) {
			router.Get("/", handlers.GenDashboardServiceGetDashboard(r.registry.DashboardService, r.stravaClient))
			router.Get("/stats", handlers.GenDashboardServiceGetStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/weekly", handlers.GenDashboardServiceGetWeeklyStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/recent", handlers.GenDashboardServiceGetRecentActivities(r.registry.DashboardService, r.stravaClient))
			router.Get("/sports", handlers.GenDashboardServiceGetSportTypeStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/config", handlers.GenDashboardServiceGetConfig(r.registry.DashboardService, r.stravaClient))
			router.Put("/config", handlers.GenDashboardServiceUpdateConfig(r.registry.DashboardService, r.stravaClient))
			router.Get("/monthly", handlers.GenDashboardServiceGetMonthlyStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/yearly", handlers.GenDashboardServiceGetYearlyStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/calendar", handlers.GenDashboardServiceGetCalendarData(r.registry.DashboardService, r.stravaClient))
			router.Get("/calendar/summary", handlers.GenDashboardServiceGetCalendarSummary(r.registry.DashboardService, r.stravaClient))
			router.Get("/calendar/activities", handlers.GenDashboardServiceGetCalendarActivities(r.registry.DashboardService, r.stravaClient))
		})

		// Stats routes (using generated adapters where available)
		router.Route("/stats", func(router chi.Router) {
			router.Get("/heatmap", handlers.GenStatsServiceGetHeatmapData(r.registry.StatsService, r.stravaClient))
			router.Get("/eddington", handlers.GenStatsServiceGetEddingtonData(r.registry.StatsService, r.stravaClient))
			router.Get("/eddington/history", handlers.GenStatsServiceGetEddingtonHistory(r.registry.StatsService, r.stravaClient))
			router.Get("/best-efforts", handlers.GenStatsServiceGetBestEffortPRs(r.registry.StatsService, r.stravaClient))
			router.Get("/best-efforts/{distanceType}", handlers.GenStatsServiceGetBestEffortsForType(r.registry.StatsService, r.stravaClient))
			router.Get("/wrapped", handlers.GenStatsServiceGetWrapped(r.registry.StatsService, r.stravaClient))
			router.Get("/wrapped/years", handlers.GenStatsServiceGetWrappedYears(r.registry.StatsService, r.stravaClient))
			router.Get("/power", handlers.GenStatsServiceGetPowerStats(r.registry.StatsService, r.stravaClient))
			router.Get("/power-zones", r.statsHandler.GetPowerZones) // Manual: complex zone calculation
			router.Get("/hr-zones", r.statsHandler.GetHRZones)       // Manual: complex zone calculation
			router.Get("/training-load", handlers.GenStatsServiceGetTrainingLoad(r.registry.StatsService, r.stravaClient))
			router.Get("/daytime", handlers.GenDashboardServiceGetDaytimeDistribution(r.registry.DashboardService, r.stravaClient))
			router.Get("/weekday", handlers.GenDashboardServiceGetWeekdayDistribution(r.registry.DashboardService, r.stravaClient))
		})

		// Gear routes (using generated adapters where available)
		router.Route("/gear", func(router chi.Router) {
			router.Get("/", handlers.GenGearServiceList(r.registry.GearService, r.stravaClient))
			router.Get("/custom", r.gearHandler.ListCustom)           // Manual: custom gear CRUD
			router.Post("/custom", r.gearHandler.CreateCustom)        // Manual: custom gear CRUD
			router.Put("/custom/{id}", r.gearHandler.UpdateCustom)    // Manual: custom gear CRUD
			router.Delete("/custom/{id}", r.gearHandler.DeleteCustom) // Manual: custom gear CRUD
			router.Get("/stats/monthly", handlers.GenGearServiceMonthlyUsage(r.registry.GearService, r.stravaClient))
			router.Get("/{id}", handlers.GenGearServiceGetByID(r.registry.GearService, r.stravaClient))
			router.Get("/{id}/components", handlers.GenMaintenanceServiceListComponents(r.registry.MaintenanceService, r.stravaClient))
			router.Post("/{id}/components", handlers.GenMaintenanceServiceCreateComponent(r.registry.MaintenanceService, r.stravaClient))
		})

		// Maintenance routes (using generated adapters where available)
		router.Route("/components", func(router chi.Router) {
			router.Put("/{id}", handlers.GenMaintenanceServiceUpdateComponent(r.registry.MaintenanceService, r.stravaClient))
			router.Delete("/{id}", r.maintenanceHandler.DeleteComponent) // Manual: needs refactoring
			router.Post("/{id}/maintenance", r.maintenanceHandler.LogMaintenance)
		})
		router.Route("/maintenance", func(router chi.Router) {
			router.Get("/due", r.maintenanceHandler.Due) // Manual: needs refactoring
		})

		// Photos routes (using generated adapters)
		router.Route("/photos", func(router chi.Router) {
			router.Get("/", handlers.GenPhotosServiceList(r.registry.PhotosService, r.stravaClient))
		})

		// Challenges routes (using generated adapters where available)
		router.Route("/challenges", func(router chi.Router) {
			router.Get("/", handlers.GenChallengesServiceList(r.registry.ChallengesService, r.stravaClient))
			router.Post("/import", r.challengesHandler.Import)                    // Manual: file upload handling
			router.Post("/import-profile", r.challengesHandler.ImportFromProfile) // Manual: web scraping
		})

		// Segments routes (using generated adapters where available)
		router.Route("/segments", func(router chi.Router) {
			router.Get("/", handlers.GenSegmentsServiceList(r.registry.SegmentsService, r.stravaClient))
			router.Get("/countries", r.segmentsHandler.Countries) // Manual: needs refactoring
			router.Get("/{id}", handlers.GenSegmentsServiceGetByID(r.registry.SegmentsService, r.stravaClient))
			router.Get("/{id}/efforts", handlers.GenSegmentsServiceListEfforts(r.registry.SegmentsService, r.stravaClient))
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

		// Export routes (using generated adapters where available)
		router.Route("/export", func(router chi.Router) {
			router.Get("/stats", handlers.GenDashboardServiceGetExportStats(r.registry.DashboardService, r.stravaClient))
			router.Get("/activities/csv", r.exportHandler.ExportActivitiesCSV)   // Manual: binary CSV
			router.Get("/activities/json", r.exportHandler.ExportActivitiesJSON) // Manual: binary JSON
		})
	})

	// Public badge endpoints (no auth required, but checks setting)
	r.Route("/badges", func(router chi.Router) {
		router.Use(httprate.LimitByIP(badgeRateLimitRequests, badgeRateLimitWindow))
		router.Get("/distance.svg", r.badgesHandler.GetDistanceBadge)
		router.Get("/time.svg", r.badgesHandler.GetTimeBadge)
		router.Get("/elevation.svg", r.badgesHandler.GetElevationBadge)
		router.Get("/activities.svg", r.badgesHandler.GetActivitiesBadge)
		router.Get("/eddington.svg", r.badgesHandler.GetEddingtonBadge)
		router.Get("/year-{year}.svg", r.badgesHandler.GetYearBadge)
		router.Get("/month-{month}.svg", r.badgesHandler.GetMonthBadge)
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
