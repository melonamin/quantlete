package demo

import (
	"math/rand"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// generateAthleteSettings creates demo athlete settings with useful defaults.
func generateAthleteSettings() storage.AthleteSettings {
	// Start with defaults and add some demo-specific customizations
	settings := storage.DefaultAthleteSettings()

	// Add a virtual world tile layer for Zwift-style maps
	settings.VirtualWorldTileLayers = map[string]storage.VirtualWorldTileLayer{
		"watopia": {
			Name:        "Watopia",
			URL:         "https://cdn.zwift.com/tiles/watopia/{z}/{x}/{y}.png",
			Attribution: "Zwift",
			MaxZoom:     18,
		},
	}

	// Eddington definitions are already set by DefaultAthleteSettings()
	// Add a few more for variety
	settings.EddingtonDefinitions = append(settings.EddingtonDefinitions,
		storage.EddingtonDefinition{
			ID:                    "outdoor_rides",
			Name:                  "Outdoor Rides",
			SportTypes:            []string{"Ride", "MountainBikeRide", "GravelRide"},
			ShowInNav:             false,
			ShowInDashboardWidget: false,
		},
		storage.EddingtonDefinition{
			ID:                    "trail_runs",
			Name:                  "Trail Runs",
			SportTypes:            []string{"TrailRun"},
			ShowInNav:             false,
			ShowInDashboardWidget: false,
		},
	)

	return settings
}

// generateTrainingGoals creates demo training goals with realistic targets.
func generateTrainingGoals() storage.TrainingGoalsConfig {
	// Helper to create pointers
	distPtr := func(d float64) *float64 { return &d }
	elevPtr := func(e float64) *float64 { return &e }
	timePtr := func(t int) *int { return &t }

	return storage.TrainingGoalsConfig{
		Version: 1,
		Sports: []storage.GoalsSportConfig{
			{
				Name:       "Rides",
				SportTypes: []string{"Ride", "VirtualRide", "GravelRide", "MountainBikeRide"},
				Targets: map[storage.GoalPeriod]storage.GoalsMetricTargets{
					storage.GoalPeriodWeek: {
						DistanceM:   distPtr(150000),   // 150km/week
						ElevationM:  elevPtr(1500),     // 1500m climbing/week
						MovingTimeS: timePtr(6 * 3600), // 6 hours/week
					},
					storage.GoalPeriodMonth: {
						DistanceM:   distPtr(600000),    // 600km/month
						ElevationM:  elevPtr(6000),      // 6000m climbing/month
						MovingTimeS: timePtr(24 * 3600), // 24 hours/month
					},
					storage.GoalPeriodYear: {
						DistanceM:   distPtr(8000000),    // 8000km/year
						ElevationM:  elevPtr(80000),      // 80,000m climbing/year
						MovingTimeS: timePtr(300 * 3600), // 300 hours/year
					},
				},
			},
			{
				Name:       "Runs",
				SportTypes: []string{"Run", "VirtualRun", "TrailRun"},
				Targets: map[storage.GoalPeriod]storage.GoalsMetricTargets{
					storage.GoalPeriodWeek: {
						DistanceM:   distPtr(40000),    // 40km/week
						ElevationM:  elevPtr(400),      // 400m climbing/week
						MovingTimeS: timePtr(4 * 3600), // 4 hours/week
					},
					storage.GoalPeriodMonth: {
						DistanceM:   distPtr(160000),    // 160km/month
						ElevationM:  elevPtr(1600),      // 1600m climbing/month
						MovingTimeS: timePtr(16 * 3600), // 16 hours/month
					},
					storage.GoalPeriodYear: {
						DistanceM:   distPtr(2000000),    // 2000km/year
						ElevationM:  elevPtr(20000),      // 20,000m climbing/year
						MovingTimeS: timePtr(200 * 3600), // 200 hours/year
					},
				},
			},
			{
				Name:       "Swims",
				SportTypes: []string{"Swim"},
				Targets: map[storage.GoalPeriod]storage.GoalsMetricTargets{
					storage.GoalPeriodWeek: {
						DistanceM:   distPtr(5000),     // 5km/week
						MovingTimeS: timePtr(2 * 3600), // 2 hours/week
					},
					storage.GoalPeriodMonth: {
						DistanceM:   distPtr(20000),    // 20km/month
						MovingTimeS: timePtr(8 * 3600), // 8 hours/month
					},
					storage.GoalPeriodYear: {
						DistanceM:   distPtr(250000),     // 250km/year
						MovingTimeS: timePtr(100 * 3600), // 100 hours/year
					},
				},
			},
		},
	}
}

// generateAthleteMetrics creates realistic weight and FTP history over time.
// Returns two slices: weight points and FTP points.
func generateAthleteMetrics(rng *rand.Rand, months int) (weightPoints, ftpPoints []storage.AthleteMetricPoint) {
	now := time.Now()
	startDate := now.AddDate(0, -months, 0)

	// Weight tracking: Start around 78kg, trend down to ~72kg with fluctuations
	// Typical athlete losing weight through training
	startWeight := 78.0
	targetWeight := 72.0
	weightRange := startWeight - targetWeight

	// FTP tracking: Start around 200W, build up to ~280W over time
	// Typical progression for a dedicated cyclist
	startFTP := 200.0
	peakFTP := 280.0
	ftpRange := peakFTP - startFTP

	// Generate monthly data points with some weekly variation
	for m := 0; m <= months; m++ {
		pointDate := startDate.AddDate(0, m, 0)

		// Progress ratio (0 to 1)
		progress := float64(m) / float64(months)

		// Weight: Gradual decrease with seasonal fluctuations
		// Add some noise and seasonal variation (heavier in winter)
		seasonalFactor := 0.5 * (1 + seasonalWave(pointDate))
		baseWeight := startWeight - (weightRange * progress * 0.8) // Don't quite reach target
		noise := (rng.Float64() - 0.5) * 2.0                       // +/- 1kg noise
		weight := baseWeight + seasonalFactor + noise
		if weight < targetWeight {
			weight = targetWeight + rng.Float64()*1.5
		}

		weightPoints = append(weightPoints, storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: pointDate},
			Value:      weight,
		})

		// FTP: Build up with plateaus and breakthroughs
		// FTP typically increases in steps, not linearly
		ftpProgress := ftpProgressCurve(progress)
		baseFTP := startFTP + (ftpRange * ftpProgress)
		// Add some variation (+/- 5W)
		ftpNoise := (rng.Float64() - 0.5) * 10.0
		ftp := baseFTP + ftpNoise
		if ftp < startFTP {
			ftp = startFTP
		}

		ftpPoints = append(ftpPoints, storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: pointDate},
			Value:      ftp,
		})

		// Add some extra weekly weight measurements for realism
		if m < months && rng.Float64() < 0.6 {
			midMonthDate := pointDate.AddDate(0, 0, 14+rng.Intn(7))
			if midMonthDate.Before(now) {
				midWeight := weight + (rng.Float64()-0.5)*1.5
				weightPoints = append(weightPoints, storage.AthleteMetricPoint{
					RecordedAt: storage.SQLiteTime{Time: midMonthDate},
					Value:      midWeight,
				})
			}
		}
	}

	return weightPoints, ftpPoints
}

// seasonalWave returns a value between -1 and 1 based on month
// (higher in winter months, lower in summer for weight)
func seasonalWave(t time.Time) float64 {
	month := float64(t.Month())
	// Peak in January (month 1), trough in July (month 7)
	// Using cosine: cos(0) = 1, cos(pi) = -1
	return -1.0 * (2.0*3.14159*(month-1)/12.0 - 3.14159) / 3.14159
}

// ftpProgressCurve models realistic FTP progression
// Quick initial gains, then slower improvement with occasional plateaus
func ftpProgressCurve(progress float64) float64 {
	// Use a modified log curve for diminishing returns
	// Plus some step-like behavior
	if progress < 0.2 {
		// Quick initial gains (newbie gains)
		return progress * 2.0 // 0 -> 0.4
	} else if progress < 0.5 {
		// Steady improvement
		return 0.4 + (progress-0.2)*1.0 // 0.4 -> 0.7
	} else if progress < 0.7 {
		// Plateau
		return 0.7 + (progress-0.5)*0.3 // 0.7 -> 0.76
	} else {
		// Final push
		return 0.76 + (progress-0.7)*0.8 // 0.76 -> 1.0
	}
}

// generateFTPRunning generates running threshold pace (m/s) history.
// Typical amateur runner: 5:00/km threshold = 3.33 m/s, improving to 4:30/km = 3.7 m/s
func generateFTPRunning(rng *rand.Rand, months int) []storage.AthleteMetricPoint {
	var points []storage.AthleteMetricPoint

	now := time.Now()
	startDate := now.AddDate(0, -months, 0)

	// Start at 5:00/km pace (3.33 m/s), progress to 4:30/km (3.7 m/s)
	startPace := 3.33
	peakPace := 3.70
	paceRange := peakPace - startPace

	for m := 0; m <= months; m++ {
		pointDate := startDate.AddDate(0, m, 0)
		progress := float64(m) / float64(months)

		// Similar progression curve as FTP
		paceProgress := ftpProgressCurve(progress)
		basePace := startPace + (paceRange * paceProgress)

		// Add variation
		noise := (rng.Float64() - 0.5) * 0.1
		pace := basePace + noise
		if pace < startPace {
			pace = startPace
		}

		points = append(points, storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: pointDate},
			Value:      pace,
		})
	}

	return points
}

// generateHRZones creates HR zone definitions for different sport types.
// Returns definitions for "All" (general), "Run", and "Ride" sport types.
func generateHRZones(athleteID int64, startDate time.Time) []storage.HRZoneDefinition {
	// Typical amateur athlete: Max HR ~185, LTHR ~165
	// Using percent_hrmax method with standard 5-zone model

	// Standard zones as % of max HR:
	// Z1: 50-60% (Recovery)
	// Z2: 60-70% (Endurance)
	// Z3: 70-80% (Tempo)
	// Z4: 80-90% (Threshold)
	// Z5: 90-100% (VO2max)

	effectiveFrom := startDate.Format("2006-01-02")

	// Zone bounds as percentages of max HR
	// Format: [Z1 upper, Z2 upper, Z3 upper, Z4 upper] - Z5 is above Z4
	percentBounds := []float64{0.60, 0.70, 0.80, 0.90}

	zonesJSON := `{"bounds":[0.60,0.70,0.80,0.90],"hr_max":185}`

	// Silence unused variable warnings
	_ = percentBounds

	return []storage.HRZoneDefinition{
		{
			AthleteID:     athleteID,
			SportType:     "All",
			EffectiveFrom: effectiveFrom,
			Method:        "percent_hrmax",
			Zones:         []byte(zonesJSON),
		},
		{
			AthleteID:     athleteID,
			SportType:     "Run",
			EffectiveFrom: effectiveFrom,
			Method:        "percent_hrmax",
			Zones:         []byte(`{"bounds":[0.60,0.70,0.80,0.88],"hr_max":188}`), // Runners often have higher max HR
		},
		{
			AthleteID:     athleteID,
			SportType:     "Ride",
			EffectiveFrom: effectiveFrom,
			Method:        "percent_hrmax",
			Zones:         []byte(`{"bounds":[0.55,0.65,0.75,0.88],"hr_max":182}`), // Cycling zones slightly different
		},
	}
}
