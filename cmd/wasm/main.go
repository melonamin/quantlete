//go:build js && wasm

// Package main provides a production WASM entry point that exposes the Go storage layer
// to JavaScript, using sql.js via go-sqlite3-js.
package main

import (
	"fmt"
	"syscall/js"

	_ "github.com/matrix-org/go-sqlite3-js"
	"github.com/melonamin/quantlete/internal/storage"
)

// WasmBridge holds the database and all repositories for the WASM bridge
type WasmBridge struct {
	db             *storage.DB
	athleteID      int64
	activities     *storage.ActivityRepository
	stats          *storage.StatsRepository
	streams        *storage.StreamRepository
	athletes       *storage.AthleteRepository
	gear           *storage.GearRepository
	segments       *storage.SegmentRepository
	bestEfforts    *storage.BestEffortsRepository
	photos         *storage.PhotoRepository
	appState       *storage.AppStateRepository
	syncHistory    *storage.SyncHistoryRepository
	power          *storage.PowerRepository
	athleteMetrics *storage.AthleteMetricsRepository
	trainingLoad   *storage.TrainingLoadRepository
	goals          *storage.GoalsRepository
	challenges     *storage.ChallengeRepository
	maintenance    *storage.MaintenanceRepository
	settings       *storage.SettingsRepository
	zones          *storage.ZonesRepository
}

// Global bridge instance
var bridge *WasmBridge

// Legacy aliases for backward compatibility during migration
// These will be removed once all functions are converted to methods
var (
	db             *storage.DB
	activities     *storage.ActivityRepository
	stats          *storage.StatsRepository
	streams        *storage.StreamRepository
	athletes       *storage.AthleteRepository
	gear           *storage.GearRepository
	segments       *storage.SegmentRepository
	bestEfforts    *storage.BestEffortsRepository
	photos         *storage.PhotoRepository
	appState       *storage.AppStateRepository
	syncHistory    *storage.SyncHistoryRepository
	power          *storage.PowerRepository
	athleteMetrics *storage.AthleteMetricsRepository
	trainingLoad   *storage.TrainingLoadRepository
	goals          *storage.GoalsRepository
	challenges     *storage.ChallengeRepository
	maintenance    *storage.MaintenanceRepository
	settings       *storage.SettingsRepository
	zones          *storage.ZonesRepository
	athleteID      int64
)

func main() {
	fmt.Println("Quantlete Go WASM Storage Layer (Production)")

	// Register JavaScript-callable functions under goStorage namespace
	js.Global().Set("goStorage", map[string]interface{}{
		// Initialization
		"init":         js.FuncOf(initStorage),
		"setAthleteId": js.FuncOf(setAthleteID),
		"exportDb":     js.FuncOf(exportDatabase),

		// Auth
		"getAuthStatus": js.FuncOf(getAuthStatus),

		// Activities - Read
		"getActivities":      js.FuncOf(getActivities),
		"getActivity":        js.FuncOf(getActivity),
		"getActivityStreams": js.FuncOf(getActivityStreams),

		// Activities - Write
		"saveActivity": js.FuncOf(saveActivity),
		"saveStream":   js.FuncOf(saveStream),

		// Athlete - Write
		"saveAthlete": js.FuncOf(saveAthlete),

		// Gear - Write
		"saveGear": js.FuncOf(saveGear),

		// Segments - Write
		"saveSegment":       js.FuncOf(saveSegment),
		"saveSegmentEffort": js.FuncOf(saveSegmentEffort),

		// Best Efforts - Write
		"saveBestEfforts": js.FuncOf(saveBestEfforts),

		// Photos - Write
		"savePhoto": js.FuncOf(savePhoto),

		// Power - Write (compute and store)
		"computePowerBestEfforts": js.FuncOf(computePowerBestEfforts),

		// Sync History - Write
		"createSyncRun":   js.FuncOf(createSyncRun),
		"updateSyncRun":   js.FuncOf(updateSyncRun),
		"completeSyncRun": js.FuncOf(completeSyncRun),

		// Dashboard
		"getDashboardStats":   js.FuncOf(getDashboardStats),
		"getWeeklyStats":         js.FuncOf(getWeeklyStats),
		"getRecentActivities":    js.FuncOf(getRecentActivities),
		"getSportTypeStats":      js.FuncOf(getSportTypeStats),
		"getMonthlyStats":        js.FuncOf(getMonthlyStats),
		"getYearlyStats":         js.FuncOf(getYearlyStats),
		"getDaytimeDistribution": js.FuncOf(getDaytimeDistribution),
		"getWeekdayDistribution": js.FuncOf(getWeekdayDistribution),
		"getExportStats":         js.FuncOf(getExportStats),

		// Heatmap
		"getHeatmapData": js.FuncOf(getHeatmapData),

		// Gear - Read
		"getGear":       js.FuncOf(getGear),
		"getGearDetail": js.FuncOf(getGearDetail),

		// Segments - Read
		"getSegments":      js.FuncOf(getSegments),
		"getSegmentDetail": js.FuncOf(getSegmentDetail),

		// Photos - Read
		"getPhotos":         js.FuncOf(getPhotos),
		"getActivityPhotos": js.FuncOf(getActivityPhotos),

		// Calendar - Read
		"getCalendarData": js.FuncOf(getCalendarData),

		// Best Efforts - Read
		"getBestEffortPRs":        js.FuncOf(getBestEffortPRs),
		"getBestEffortsForType":   js.FuncOf(getBestEffortsForType),

		// Sync History - Read
		"getSyncHistory": js.FuncOf(getSyncHistory),

		// Algorithms - Power
		"normalizedPower":        js.FuncOf(normalizedPowerFn),
		"rollingMaxAverage":      js.FuncOf(rollingMaxAverageFn),
		"intensityFactor":        js.FuncOf(intensityFactorFn),
		"trainingStressScore":    js.FuncOf(trainingStressScoreFn),

		// Algorithms - Eddington
		"eddingtonNumber":    js.FuncOf(eddingtonNumberFn),
		"eddingtonNextSteps": js.FuncOf(eddingtonNextStepsFn),
		"eddingtonHistory":   js.FuncOf(eddingtonHistoryFn),

		// Algorithms - Training Load
		"calculateTrainingLoad":            js.FuncOf(calculateTrainingLoadFn),
		"calculateTrainingLoadWithInitial": js.FuncOf(calculateTrainingLoadWithInitialFn),
		"predictAfterWorkout":              js.FuncOf(predictAfterWorkoutFn),
		"tssForTargetTsb":                  js.FuncOf(tssForTargetTsbFn),

		// Eddington
		"getEddingtonData": js.FuncOf(getEddingtonData),

		// Gear Stats
		"getGearMonthlyUsage": js.FuncOf(getGearMonthlyUsage),

		// Segment Efforts
		"getSegmentEfforts":   js.FuncOf(getSegmentEfforts),
		"getSegmentCountries": js.FuncOf(getSegmentCountries),

		// Rewind
		"getRewindYears": js.FuncOf(getRewindYears),
		"getRewind":      js.FuncOf(getRewind),

		// Athlete Metrics (FTP/Weight)
		"getFtpHistory":     js.FuncOf(getFtpHistory),
		"getWeightHistory":  js.FuncOf(getWeightHistory),
		"updateFtpHistory":  js.FuncOf(updateFtpHistory),
		"updateWeightHistory": js.FuncOf(updateWeightHistory),

		// Training Load
		"getTrainingLoad": js.FuncOf(getTrainingLoad),

		// Power Stats
		"getPowerStats": js.FuncOf(getPowerStats),

		// Challenges
		"getChallenges": js.FuncOf(getChallenges),

		// Training Goals
		"getTrainingGoals":    js.FuncOf(getTrainingGoals),
		"updateTrainingGoals": js.FuncOf(updateTrainingGoals),

		// Maintenance
		"getMaintenanceDue":  js.FuncOf(getMaintenanceDue),
		"getGearComponents":  js.FuncOf(getGearComponents),
		"createComponent":    js.FuncOf(createComponent),
		"updateComponent":    js.FuncOf(updateComponent),
		"deleteComponent":    js.FuncOf(deleteComponent),
		"logMaintenance":     js.FuncOf(logMaintenance),

		// Settings
		"getAppSettings":    js.FuncOf(getAppSettings),
		"updateAppSettings": js.FuncOf(updateAppSettings),

		// Custom Gear
		"getCustomGear":    js.FuncOf(getCustomGear),
		"createCustomGear": js.FuncOf(createCustomGear),
		"updateCustomGear": js.FuncOf(updateCustomGear),
		"deleteCustomGear": js.FuncOf(deleteCustomGear),

		// HR Zones
		"getHrZoneDefinitions":    js.FuncOf(getHrZoneDefinitions),
		"upsertHrZoneDefinition":  js.FuncOf(upsertHrZoneDefinition),
		"deleteHrZoneDefinition":  js.FuncOf(deleteHrZoneDefinition),
	})

	fmt.Println("Go WASM storage functions registered")

	// Keep the Go program running
	select {}
}

// ============================================================================
// Initialization
// ============================================================================

// initStorage initializes the database connection
// Called from JS: goStorage.init()
func initStorage(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("initStorage")

	var err error
	// Open in-memory database (sql.js handles persistence externally)
	db, err = storage.OpenWasm()
	if err != nil {
		return errorJSON(err)
	}

	// Run migrations
	if err := storage.RunMigrations(db.Conn()); err != nil {
		return errorJSON(fmt.Errorf("running migrations: %w", err))
	}

	// Initialize repositories (legacy globals)
	activities = storage.NewActivityRepository(db)
	stats = storage.NewStatsRepository(db)
	streams = storage.NewStreamRepository(db)
	athletes = storage.NewAthleteRepository(db)
	gear = storage.NewGearRepository(db)
	segments = storage.NewSegmentRepository(db)
	bestEfforts = storage.NewBestEffortsRepository(db)
	photos = storage.NewPhotoRepository(db)
	appState = storage.NewAppStateRepository(db)
	syncHistory = storage.NewSyncHistoryRepository(db, appState)
	power = storage.NewPowerRepository(db, streams)
	athleteMetrics = storage.NewAthleteMetricsRepository(db)
	zones = storage.NewZonesRepository(db)
	trainingLoad = storage.NewTrainingLoadRepository(db, streams, athleteMetrics, zones)
	goals = storage.NewGoalsRepository(db)
	challenges = storage.NewChallengeRepository(db)
	maintenance = storage.NewMaintenanceRepository(db)
	settings = storage.NewSettingsRepository(db)

	// Initialize WasmBridge (new pattern)
	bridge = &WasmBridge{
		db:             db,
		activities:     activities,
		stats:          stats,
		streams:        streams,
		athletes:       athletes,
		gear:           gear,
		segments:       segments,
		bestEfforts:    bestEfforts,
		photos:         photos,
		appState:       appState,
		syncHistory:    syncHistory,
		power:          power,
		athleteMetrics: athleteMetrics,
		trainingLoad:   trainingLoad,
		goals:          goals,
		challenges:     challenges,
		maintenance:    maintenance,
		settings:       settings,
		zones:          zones,
	}

	return successJSON("Database initialized")
}

// setAthleteID sets the current athlete ID for queries
// Called from JS: goStorage.setAthleteId(id)
func setAthleteID(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("setAthleteID")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing athlete ID"))
	}

	id := int64(args[0].Int())
	athleteID = id // Legacy global
	if bridge != nil {
		bridge.athleteID = id
	}
	return successJSON(fmt.Sprintf("Athlete ID set to %d", id))
}

// exportDatabase exports the database for OPFS persistence
// Called from JS: goStorage.exportDb()
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
