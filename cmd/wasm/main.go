//go:build js && wasm

// Package main provides a production WASM entry point that exposes the Go storage layer
// to JavaScript, using sql.js via go-sqlite3-js.
package main

import (
	"fmt"
	"syscall/js"

	_ "github.com/matrix-org/go-sqlite3-js"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// WasmBridge holds the database and all repositories for the WASM bridge
type WasmBridge struct {
	db                 *storage.DB
	athleteID          int64
	activities         *storage.ActivityRepository
	activityService    *services.ActivityService
	stats              *storage.StatsRepository
	statsService       *services.StatsService
	dashboardConfig    *storage.DashboardConfigRepository
	dashboardService   *services.DashboardService
	streams            *storage.StreamRepository
	athletes           *storage.AthleteRepository
	gear               *storage.GearRepository
	gearService        *services.GearService
	segments           *storage.SegmentRepository
	segmentsService    *services.SegmentsService
	bestEfforts        *storage.BestEffortsRepository
	photos             *storage.PhotoRepository
	photosService      *services.PhotosService
	appState           *storage.AppStateRepository
	syncHistory        *storage.SyncHistoryRepository
	power              *storage.PowerRepository
	athleteMetrics     *storage.AthleteMetricsRepository
	trainingLoad       *storage.TrainingLoadRepository
	goals              *storage.GoalsRepository
	challenges         *storage.ChallengeRepository
	challengesService  *services.ChallengesService
	maintenance        *storage.MaintenanceRepository
	maintenanceService *services.MaintenanceService
	settings           *storage.SettingsRepository
	zones              *storage.ZonesRepository
}

// Global bridge instance - single source of truth for all state
var bridge *WasmBridge

func main() {
	fmt.Println("Quantlete Go WASM Storage Layer (Production)")

	// Register all WASM functions (generated from //wasm:export comments)
	RegisterAll()

	fmt.Println("Go WASM storage functions registered")

	// Keep the Go program running
	select {}
}

// ============================================================================
// Initialization
// ============================================================================

//wasm:category Initialization

// initStorage initializes the database connection
// Called from JS: goStorage.init()
//wasm:export init
func initStorage(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("initStorage")

	// Open in-memory database (sql.js handles persistence externally)
	db, err := storage.OpenWasm()
	if err != nil {
		return errorJSON(err)
	}

	// Run migrations
	if err := storage.RunMigrations(db.Conn()); err != nil {
		return errorJSON(fmt.Errorf("running migrations: %w", err))
	}

	// Initialize repositories
	activities := storage.NewActivityRepository(db)
	stats := storage.NewStatsRepository(db)
	streams := storage.NewStreamRepository(db)
	athletes := storage.NewAthleteRepository(db)
	gear := storage.NewGearRepository(db)
	segments := storage.NewSegmentRepository(db)
	bestEfforts := storage.NewBestEffortsRepository(db)
	photos := storage.NewPhotoRepository(db)
	appState := storage.NewAppStateRepository(db)
	syncHistory := storage.NewSyncHistoryRepository(db, appState)
	power := storage.NewPowerRepository(db, streams)
	athleteMetrics := storage.NewAthleteMetricsRepository(db)
	zones := storage.NewZonesRepository(db)
	trainingLoad := storage.NewTrainingLoadRepository(db, streams, athleteMetrics, zones)
	goals := storage.NewGoalsRepository(db)
	challenges := storage.NewChallengeRepository(db)
	maintenance := storage.NewMaintenanceRepository(db)
	settings := storage.NewSettingsRepository(db)
	dashboardConfig := storage.NewDashboardConfigRepository(db)

	// Initialize services
	activityService := services.NewActivityService(activities, streams)
	gearService := services.NewGearService(gear)
	segmentsService := services.NewSegmentsService(segments)
	photosService := services.NewPhotosService(photos)
	statsService := services.NewStatsService(stats, power, bestEfforts, trainingLoad)
	dashboardService := services.NewDashboardService(db, stats, dashboardConfig)
	maintenanceService := services.NewMaintenanceService(maintenance)
	challengesService := services.NewChallengesService(challenges)

	// Initialize WasmBridge
	bridge = &WasmBridge{
		db:                 db,
		activities:         activities,
		activityService:    activityService,
		stats:              stats,
		statsService:       statsService,
		dashboardConfig:    dashboardConfig,
		dashboardService:   dashboardService,
		streams:            streams,
		athletes:           athletes,
		gear:               gear,
		gearService:        gearService,
		segments:           segments,
		segmentsService:    segmentsService,
		bestEfforts:        bestEfforts,
		photos:             photos,
		photosService:      photosService,
		appState:           appState,
		syncHistory:        syncHistory,
		power:              power,
		athleteMetrics:     athleteMetrics,
		trainingLoad:       trainingLoad,
		goals:              goals,
		challenges:         challenges,
		challengesService:  challengesService,
		maintenance:        maintenance,
		maintenanceService: maintenanceService,
		settings:           settings,
		zones:              zones,
	}

	return successJSON("Database initialized")
}

// setAthleteID sets the current athlete ID for queries
// Called from JS: goStorage.setAthleteId(id)
//wasm:export setAthleteId
func setAthleteID(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("setAthleteID")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing athlete ID"))
	}

	if bridge == nil {
		return errorJSON(fmt.Errorf("storage not initialized - call init() first"))
	}

	id := int64(args[0].Int())
	bridge.athleteID = id
	return successJSON(fmt.Sprintf("Athlete ID set to %d", id))
}

// exportDatabase exports the database for OPFS persistence
// Called from JS: goStorage.exportDb()
//wasm:export exportDb
func exportDatabase(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("exportDatabase")

	// Get the sql.js database and call export()
	dbMap := js.Global().Get("_go_sqlite_dbs")
	jsDb := dbMap.Call("get", ":memory:")
	if !jsDb.Truthy() {
		return errorJSON(fmt.Errorf("database not found"))
	}

	// Call db.export() to get Uint8Array
	data := jsDb.Call("export")
	return data // Return the Uint8Array directly
}
