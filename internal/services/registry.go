package services

import (
	"log/slog"

	"github.com/melonamin/quantlete/internal/storage"
)

// ServiceRegistry consolidates all repository and service initialization.
// This provides a single source of truth for the dependency graph,
// shared between server and WASM targets.
//
// # Usage Guidelines
//
// The registry exposes both Services (public fields) and Repositories (accessor methods).
// Choose the appropriate level based on your use case:
//
// ## When to use Services (e.g., ActivityService, GearService)
//
// Use services when you need:
//   - Business logic beyond CRUD operations
//   - Coordinated operations across multiple repositories
//   - Computed/derived data (e.g., statistics, aggregations)
//   - Validation or transformation logic
//
// Example: registry.ActivityService.GetWithStreams(ctx, id)
//
// ## When to use Repositories (e.g., Activities(), Gear())
//
// Use repository accessors when you need:
//   - Direct database CRUD operations
//   - Simple queries without business logic
//   - Import/migration operations (where the importer handles the logic)
//   - Low-level operations that services don't expose
//
// Example: registry.Activities().Upsert(ctx, activity)
//
// ## General Rule
//
// Prefer services when available. Use repositories directly only when:
//   - No corresponding service exists
//   - You're implementing a service that needs repository access
//   - The operation is purely data-layer (import adapters, migrations)
type ServiceRegistry struct {
	db *storage.DB

	// Repositories (private - use accessor methods)
	activities      *storage.ActivityRepository
	athletes        *storage.AthleteRepository
	streams         *storage.StreamRepository
	gear            *storage.GearRepository
	segments        *storage.SegmentRepository
	bestEfforts     *storage.BestEffortsRepository
	photos          *storage.PhotoRepository
	appState        *storage.AppStateRepository
	syncHistory     *storage.SyncHistoryRepository
	power           *storage.PowerRepository
	athleteMetrics  *storage.AthleteMetricsRepository
	zones           *storage.ZonesRepository
	trainingLoad    *storage.TrainingLoadRepository
	goals           *storage.GoalsRepository
	challenges      *storage.ChallengeRepository
	maintenance     *storage.MaintenanceRepository
	settings        *storage.SettingsRepository
	dashboardConfig *storage.DashboardConfigRepository
	stats           *storage.StatsRepository

	// Services (public)
	ActivityService     *ActivityService
	GearService         *GearService
	SegmentsService     *SegmentsService
	PhotosService       *PhotosService
	StatsService        *StatsService
	DashboardService    *DashboardService
	MaintenanceService  *MaintenanceService
	ChallengesService   *ChallengesService
	NotificationService *NotificationService
	InsightsService     *InsightsService
}

// NewServiceRegistry creates a new registry with all repositories and services initialized.
// Dependency ordering is handled automatically.
func NewServiceRegistry(db *storage.DB, logger *slog.Logger) *ServiceRegistry {
	r := &ServiceRegistry{db: db}

	// Phase 1: Leaf repositories (no dependencies)
	r.activities = storage.NewActivityRepository(db)
	r.athletes = storage.NewAthleteRepository(db)
	r.streams = storage.NewStreamRepository(db)
	r.gear = storage.NewGearRepository(db)
	r.segments = storage.NewSegmentRepository(db)
	r.bestEfforts = storage.NewBestEffortsRepository(db)
	r.photos = storage.NewPhotoRepository(db)
	r.appState = storage.NewAppStateRepository(db)
	r.athleteMetrics = storage.NewAthleteMetricsRepository(db)
	r.zones = storage.NewZonesRepository(db)
	r.goals = storage.NewGoalsRepository(db)
	r.challenges = storage.NewChallengeRepository(db)
	r.maintenance = storage.NewMaintenanceRepository(db)
	r.settings = storage.NewSettingsRepository(db)
	r.dashboardConfig = storage.NewDashboardConfigRepository(db)
	r.stats = storage.NewStatsRepository(db)

	// Phase 2: Repositories with dependencies
	r.syncHistory = storage.NewSyncHistoryRepository(db, r.appState)
	r.power = storage.NewPowerRepository(db, r.streams)
	r.trainingLoad = storage.NewTrainingLoadRepository(db, r.streams, r.athleteMetrics, r.zones)

	// Phase 3: Services
	r.ActivityService = NewActivityService(r.activities, r.streams)
	r.GearService = NewGearService(r.gear)
	r.SegmentsService = NewSegmentsService(r.segments)
	r.PhotosService = NewPhotosService(r.photos)
	r.StatsService = NewStatsService(r.stats, r.power, r.bestEfforts, r.trainingLoad)
	r.DashboardService = NewDashboardService(db, r.stats, r.dashboardConfig)
	r.MaintenanceService = NewMaintenanceService(r.maintenance)
	r.ChallengesService = NewChallengesService(r.challenges)
	r.NotificationService = NewNotificationService(logger, r.settings)
	r.InsightsService = NewInsightsService(r.trainingLoad, r.power, r.activities)

	return r
}

// DB returns the database connection.
func (r *ServiceRegistry) DB() *storage.DB {
	return r.db
}

// Repository accessors - for handlers that need direct repository access

func (r *ServiceRegistry) Activities() *storage.ActivityRepository {
	return r.activities
}

func (r *ServiceRegistry) Athletes() *storage.AthleteRepository {
	return r.athletes
}

func (r *ServiceRegistry) Streams() *storage.StreamRepository {
	return r.streams
}

func (r *ServiceRegistry) Gear() *storage.GearRepository {
	return r.gear
}

func (r *ServiceRegistry) Segments() *storage.SegmentRepository {
	return r.segments
}

func (r *ServiceRegistry) BestEfforts() *storage.BestEffortsRepository {
	return r.bestEfforts
}

func (r *ServiceRegistry) Photos() *storage.PhotoRepository {
	return r.photos
}

func (r *ServiceRegistry) AppState() *storage.AppStateRepository {
	return r.appState
}

func (r *ServiceRegistry) SyncHistory() *storage.SyncHistoryRepository {
	return r.syncHistory
}

func (r *ServiceRegistry) Power() *storage.PowerRepository {
	return r.power
}

func (r *ServiceRegistry) AthleteMetrics() *storage.AthleteMetricsRepository {
	return r.athleteMetrics
}

func (r *ServiceRegistry) Zones() *storage.ZonesRepository {
	return r.zones
}

func (r *ServiceRegistry) TrainingLoad() *storage.TrainingLoadRepository {
	return r.trainingLoad
}

func (r *ServiceRegistry) Goals() *storage.GoalsRepository {
	return r.goals
}

func (r *ServiceRegistry) Challenges() *storage.ChallengeRepository {
	return r.challenges
}

func (r *ServiceRegistry) Maintenance() *storage.MaintenanceRepository {
	return r.maintenance
}

func (r *ServiceRegistry) Settings() *storage.SettingsRepository {
	return r.settings
}

func (r *ServiceRegistry) DashboardConfig() *storage.DashboardConfigRepository {
	return r.dashboardConfig
}

func (r *ServiceRegistry) Stats() *storage.StatsRepository {
	return r.stats
}
