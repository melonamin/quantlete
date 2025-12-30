// Package demo provides fake data generation for demonstration purposes.
package demo

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/sasha/stata/internal/storage"
)

// Options configures demo data generation.
type Options struct {
	NumActivities int    // Number of activities to generate
	Months        int    // Number of months to spread activities across
	AthleteName   string // Name of the demo athlete (format: "First Last")
}

// Generator creates demo data using the existing repository pattern.
type Generator struct {
	db              *storage.DB
	activityRepo    *storage.ActivityRepository
	athleteRepo     *storage.AthleteRepository
	gearRepo        *storage.GearRepository
	streamRepo      *storage.StreamRepository
	segmentRepo     *storage.SegmentRepository
	settingsRepo    *storage.SettingsRepository
	goalsRepo       *storage.GoalsRepository
	metricsRepo     *storage.AthleteMetricsRepository
	bestEffortsRepo *storage.BestEffortsRepository
	challengeRepo   *storage.ChallengeRepository
	zonesRepo       *storage.ZonesRepository
	maintenanceRepo *storage.MaintenanceRepository
	appStateRepo    *storage.AppStateRepository
	rng             *rand.Rand
}

// NewGenerator creates a new demo generator.
func NewGenerator(db *storage.DB) *Generator {
	return &Generator{
		db:              db,
		activityRepo:    storage.NewActivityRepository(db),
		athleteRepo:     storage.NewAthleteRepository(db),
		gearRepo:        storage.NewGearRepository(db),
		streamRepo:      storage.NewStreamRepository(db),
		segmentRepo:     storage.NewSegmentRepository(db),
		settingsRepo:    storage.NewSettingsRepository(db),
		goalsRepo:       storage.NewGoalsRepository(db),
		metricsRepo:     storage.NewAthleteMetricsRepository(db),
		bestEffortsRepo: storage.NewBestEffortsRepository(db),
		challengeRepo:   storage.NewChallengeRepository(db),
		zonesRepo:       storage.NewZonesRepository(db),
		maintenanceRepo: storage.NewMaintenanceRepository(db),
		appStateRepo:    storage.NewAppStateRepository(db),
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // G404: demo data doesn't need crypto-random
	}
}

// Generate creates demo data based on the provided options.
func (g *Generator) Generate(ctx context.Context, opts Options) error {
	// Wipe existing data
	slog.Info("Wiping existing data...")
	if err := g.wipeData(ctx); err != nil {
		return fmt.Errorf("wiping data: %w", err)
	}

	// Parse athlete name
	firstName, lastName := parseAthleteName(opts.AthleteName)

	// Generate athlete
	slog.Info("Creating demo athlete...")
	athlete := g.generateAthlete(firstName, lastName)
	if err := g.athleteRepo.Upsert(ctx, athlete); err != nil {
		return fmt.Errorf("creating athlete: %w", err)
	}

	// Generate gear
	slog.Info("Creating demo gear...")
	gear := g.generateGear(athlete.ID)
	for _, item := range gear {
		if err := g.gearRepo.Upsert(ctx, &item); err != nil {
			return fmt.Errorf("creating gear: %w", err)
		}
	}

	// Generate activities
	slog.Info("Generating activities...", "count", opts.NumActivities)
	activities := generateActivities(g.rng, athlete.ID, gear, opts.NumActivities, opts.Months)
	for i, a := range activities {
		if err := g.activityRepo.Upsert(ctx, &a); err != nil {
			return fmt.Errorf("creating activity %d: %w", i, err)
		}
		if (i+1)%25 == 0 || i == len(activities)-1 {
			slog.Debug("Activities progress", "completed", i+1, "total", len(activities))
		}
	}

	// Generate athlete metrics early (weight, FTP) - needed for training load calculations
	slog.Info("Generating athlete metrics (weight, FTP)...")
	weightPoints, ftpPoints := generateAthleteMetrics(g.rng, opts.Months)

	// Generate segments (200 for demo data)
	segmentCount := 200
	slog.Info("Generating segments...", "count", segmentCount)
	segments := generateSegments(g.rng, athlete.ID, segmentCount)
	for _, seg := range segments {
		if err := g.segmentRepo.UpsertSegment(ctx, &seg); err != nil {
			return fmt.Errorf("creating segment: %w", err)
		}
	}

	// Generate segment efforts linked to activities
	slog.Info("Generating segment efforts...")
	segEfforts := generateSegmentEfforts(g.rng, athlete.ID, activities, segments)
	for i, e := range segEfforts {
		if err := g.segmentRepo.UpsertEffort(ctx, &e); err != nil {
			return fmt.Errorf("creating segment effort %d: %w", i, err)
		}
	}

	// Generate best efforts (running PRs)
	slog.Info("Generating best efforts (running PRs)...")
	bestEfforts := generateBestEfforts(g.rng, athlete.ID, activities)
	// Group by activity for the repository API
	effortsByActivity := make(map[int64][]storage.BestEffort)
	for _, e := range bestEfforts {
		effortsByActivity[e.ActivityID] = append(effortsByActivity[e.ActivityID], e)
	}
	for activityID, actEfforts := range effortsByActivity {
		if len(actEfforts) == 0 {
			continue
		}
		if err := g.bestEffortsRepo.ReplaceForActivity(ctx, athlete.ID, activityID, actEfforts[0].SportType, actEfforts); err != nil {
			return fmt.Errorf("creating best efforts for activity %d: %w", activityID, err)
		}
	}
	slog.Info("Best efforts generated", "total", len(bestEfforts), "activities_with_efforts", len(effortsByActivity))

	// Generate activity streams (watts, HR, velocity, cadence)
	slog.Info("Generating activity streams...")
	streams := generateActivityStreams(g.rng, activities)
	for _, s := range streams {
		if err := g.streamRepo.Upsert(ctx, &s); err != nil {
			slog.Warn("Failed to create stream", "activity_id", s.ActivityID, "type", s.StreamType, "error", err)
		}
	}
	slog.Info("Activity streams generated", "total", len(streams))

	// Generate power best efforts from streams
	slog.Info("Generating power best efforts...")
	powerEfforts := generatePowerBestEfforts(g.rng, athlete.ID, activities, streams)
	for _, pe := range powerEfforts {
		_, err := g.db.ExecContext(ctx, `
			INSERT INTO power_best_efforts (activity_id, athlete_id, duration_s, best_avg_watts, computed_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (activity_id, duration_s) DO UPDATE SET
				best_avg_watts = EXCLUDED.best_avg_watts,
				computed_at = EXCLUDED.computed_at
		`, pe.ActivityID, pe.AthleteID, pe.DurationS, pe.Watts, time.Now())
		if err != nil {
			slog.Warn("Failed to create power best effort", "activity_id", pe.ActivityID, "duration", pe.DurationS, "error", err)
		}
	}
	slog.Info("Power best efforts generated", "total", len(powerEfforts))

	// Generate activity training load (TSS)
	// First, get the latest FTP values for calculations
	latestFTPWatts := 250.0 // default
	latestFTPRunning := 3.5 // default (m/s)
	if len(ftpPoints) > 0 {
		latestFTPWatts = ftpPoints[len(ftpPoints)-1].Value
	}

	slog.Info("Generating activity training load (TSS)...")
	activityLoads := generateActivityTrainingLoad(g.rng, athlete.ID, activities, streams, latestFTPWatts, latestFTPRunning)
	for _, load := range activityLoads {
		_, err := g.db.ExecContext(ctx, `
			INSERT INTO activity_training_load (activity_id, athlete_id, sport_type, method, ftp_used, normalized_power, intensity_factor, tss, computed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (activity_id) DO UPDATE SET
				method = EXCLUDED.method,
				ftp_used = EXCLUDED.ftp_used,
				normalized_power = EXCLUDED.normalized_power,
				intensity_factor = EXCLUDED.intensity_factor,
				tss = EXCLUDED.tss,
				computed_at = EXCLUDED.computed_at
		`, load.ActivityID, load.AthleteID, load.SportType, load.Method, load.FTPUsed, load.NormalizedPower, load.IntensityFactor, load.TSS, time.Now())
		if err != nil {
			slog.Warn("Failed to create activity training load", "activity_id", load.ActivityID, "error", err)
		}
	}
	slog.Info("Activity training load generated", "total", len(activityLoads))

	// Generate daily training load (CTL/ATL/TSB)
	slog.Info("Generating daily training load (CTL/ATL/TSB)...")
	dailyLoads := generateDailyTrainingLoad(athlete.ID, activityLoads, activities)
	for _, dl := range dailyLoads {
		_, err := g.db.ExecContext(ctx, `
			INSERT INTO daily_training_load (athlete_id, day, tss, ctl, atl, tsb)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (athlete_id, day) DO UPDATE SET
				tss = EXCLUDED.tss,
				ctl = EXCLUDED.ctl,
				atl = EXCLUDED.atl,
				tsb = EXCLUDED.tsb
		`, dl.AthleteID, dl.Day, dl.TSS, dl.CTL, dl.ATL, dl.TSB)
		if err != nil {
			slog.Warn("Failed to create daily training load", "day", dl.Day, "error", err)
		}
	}
	slog.Info("Daily training load generated", "total", len(dailyLoads))

	// Generate FTP running metric
	slog.Info("Creating FTP running history...")
	ftpRunningPoints := generateFTPRunning(g.rng, opts.Months)
	if err := g.metricsRepo.Replace(ctx, athlete.ID, "ftp_running_mps", ftpRunningPoints); err != nil {
		return fmt.Errorf("creating FTP running history: %w", err)
	}

	// Generate HR zone definitions
	slog.Info("Creating HR zone definitions...")
	startDate := time.Now().AddDate(0, -opts.Months, 0)
	hrZones := generateHRZones(athlete.ID, startDate)
	for _, zone := range hrZones {
		if err := g.zonesRepo.UpsertHR(ctx, athlete.ID, zone); err != nil {
			slog.Warn("Failed to create HR zone", "sport_type", zone.SportType, "error", err)
		}
	}

	// Generate challenges/badges
	slog.Info("Generating challenges...")
	challenges := generateChallenges(g.rng, athlete.ID, opts.Months)
	for _, c := range challenges {
		if err := g.challengeRepo.Upsert(ctx, &c); err != nil {
			slog.Warn("Failed to create challenge", "name", c.Name, "error", err)
		}
	}
	slog.Info("Challenges generated", "total", len(challenges))

	// Generate weather data
	slog.Info("Generating weather data...")
	weatherData := generateWeather(g.rng, activities)
	for _, w := range weatherData {
		_, err := g.db.Conn().ExecContext(ctx, `
			INSERT INTO activity_weather (
				activity_id, source, temperature_c, feels_like_c, humidity_percent,
				wind_speed_mps, wind_direction_deg, precipitation_mm, weather_code,
				temp_min_c, temp_max_c, temp_avg_c, fetched_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(activity_id) DO UPDATE SET
				source = excluded.source,
				temperature_c = excluded.temperature_c,
				feels_like_c = excluded.feels_like_c,
				humidity_percent = excluded.humidity_percent,
				wind_speed_mps = excluded.wind_speed_mps,
				wind_direction_deg = excluded.wind_direction_deg,
				precipitation_mm = excluded.precipitation_mm,
				weather_code = excluded.weather_code,
				temp_min_c = excluded.temp_min_c,
				temp_max_c = excluded.temp_max_c,
				temp_avg_c = excluded.temp_avg_c,
				fetched_at = excluded.fetched_at
		`, w.ActivityID, w.Source, w.TemperatureC, w.FeelsLikeC, w.HumidityPct,
			w.WindSpeedMps, w.WindDirDeg, w.PrecipMM, w.WeatherCode,
			w.TempMinC, w.TempMaxC, w.TempAvgC, w.FetchedAt)
		if err != nil {
			slog.Warn("Failed to create weather", "activity_id", w.ActivityID, "error", err)
		}
	}
	slog.Info("Weather data generated", "total", len(weatherData))

	// Generate gear components and maintenance
	slog.Info("Generating gear components...")
	startDate = time.Now().AddDate(0, -opts.Months, 0) // reuse startDate variable
	components := generateComponents(g.rng, gear, startDate)
	componentIDs := make([]int64, len(components))
	for i, comp := range components {
		createdAt := storage.SQLiteTime{Time: comp.CreatedAt}
		var componentID int64
		err := g.db.QueryRowContext(ctx, `
			INSERT INTO components (gear_id, name, description, installed_at, created_at)
			VALUES (?, ?, ?, ?, ?)
			RETURNING id
		`, comp.GearID, comp.Name, nil, createdAt, createdAt).Scan(&componentID)
		if err != nil {
			slog.Warn("Failed to create component", "name", comp.Name, "error", err)
			continue
		}
		componentIDs[i] = componentID

		// Create maintenance rules for this component
		for _, rule := range comp.Rules {
			// Map rule types to schema metric/unit
			var metric, unit string
			switch rule.Type {
			case "distance_m":
				metric = "distance"
				unit = "meters"
			case "days":
				metric = "interval"
				unit = "days"
			default:
				metric = "distance"
				unit = "meters"
			}
			_, err := g.db.ExecContext(ctx, `
				INSERT INTO maintenance_rules (component_id, metric, threshold_value, threshold_unit, created_at)
				VALUES (?, ?, ?, ?, ?)
			`, componentID, metric, rule.ThresholdValue, unit, createdAt)
			if err != nil {
				slog.Warn("Failed to create maintenance rule", "component_id", componentID, "error", err)
			}
		}
	}
	slog.Info("Components generated", "total", len(components))

	// Generate maintenance logs
	slog.Info("Generating maintenance history...")
	maintenanceLogs := generateMaintenanceLogs(g.rng, components, gear, activities)
	for _, log := range maintenanceLogs {
		componentID := componentIDs[log.ComponentIdx]
		if componentID == 0 {
			continue
		}
		completedAt := storage.SQLiteTime{Time: log.CompletedAt}
		_, err := g.db.ExecContext(ctx, `
			INSERT INTO maintenance_log (component_id, activity_id, completed_at, created_at)
			VALUES (?, NULL, ?, ?)
		`, componentID, completedAt, completedAt)
		if err != nil {
			slog.Warn("Failed to create maintenance log", "component_id", componentID, "error", err)
		}
	}
	slog.Info("Maintenance history generated", "total", len(maintenanceLogs))

	// Generate athlete settings
	slog.Info("Creating athlete settings...")
	settings := generateAthleteSettings()
	if err := g.settingsRepo.Upsert(ctx, athlete.ID, settings); err != nil {
		return fmt.Errorf("creating athlete settings: %w", err)
	}

	// Generate training goals
	slog.Info("Creating training goals...")
	goals := generateTrainingGoals()
	if err := g.goalsRepo.UpsertConfig(ctx, athlete.ID, goals); err != nil {
		return fmt.Errorf("creating training goals: %w", err)
	}

	// Save weight and FTP history (generated earlier for training load calculations)
	// Note: API expects specific metric names: weight_kg, ftp_cycling_watts, ftp_running_mps
	slog.Info("Saving athlete metrics (weight, FTP)...")
	if err := g.metricsRepo.Replace(ctx, athlete.ID, "weight_kg", weightPoints); err != nil {
		return fmt.Errorf("creating weight history: %w", err)
	}
	if err := g.metricsRepo.Replace(ctx, athlete.ID, "ftp_cycling_watts", ftpPoints); err != nil {
		return fmt.Errorf("creating FTP history: %w", err)
	}

	// Set demo mode flag so the server knows to bypass Strava auth
	if err := g.appStateRepo.Set(ctx, storage.AppStateDemoMode, "true"); err != nil {
		return fmt.Errorf("setting demo mode flag: %w", err)
	}

	// Store the demo athlete ID for the server to load
	if err := g.appStateRepo.Set(ctx, storage.AppStateDemoAthleteID, fmt.Sprintf("%d", athlete.ID)); err != nil {
		return fmt.Errorf("storing demo athlete id: %w", err)
	}

	slog.Info("Demo data generation complete",
		"athlete", opts.AthleteName,
		"activities", len(activities),
		"gear", len(gear),
		"components", len(components),
		"maintenance_logs", len(maintenanceLogs),
		"segments", len(segments),
		"segment_efforts", len(segEfforts),
		"best_efforts", len(bestEfforts),
		"streams", len(streams),
		"power_efforts", len(powerEfforts),
		"activity_loads", len(activityLoads),
		"daily_loads", len(dailyLoads),
		"challenges", len(challenges),
		"weather", len(weatherData),
	)

	return nil
}

// wipeData removes all existing data from the database.
func (g *Generator) wipeData(ctx context.Context) error {
	// Order matters due to foreign key constraints
	tables := []string{
		"activity_streams",
		"activity_weather",
		"activity_training_load",
		"segment_efforts",
		"best_efforts",
		"power_best_efforts",
		"photos",
		"maintenance_log",
		"activities",
		"segments",
		"maintenance_rules",
		"components",
		"gear",
		"daily_training_load",
		"challenges",
		"athlete_metrics",
		"hr_zone_definitions",
		"dashboard_config",
		"training_goals",
		"athlete_settings",
		"sync_history",
		"auth_tokens",
		"athletes",
		"app_state",
	}

	for _, table := range tables {
		if _, err := g.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			slog.Warn("Failed to clear table", "table", table, "error", err)
		}
	}

	return nil
}

// parseAthleteName splits a name into first and last parts.
func parseAthleteName(name string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// randBetween returns a random float64 between min and max.
func (g *Generator) randBetween(min, max float64) float64 {
	return min + g.rng.Float64()*(max-min)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}
