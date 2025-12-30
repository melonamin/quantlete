package demo

import (
	"math/rand"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// sportTypeWeights defines the probability distribution of sport types.
var sportTypeWeights = []struct {
	sportType string
	weight    float64
}{
	{"Run", 0.40},
	{"Ride", 0.25},
	{"VirtualRide", 0.08},
	{"TrailRun", 0.05},
	{"Walk", 0.05},
	{"Hike", 0.05},
	{"Swim", 0.04},
	{"GravelRide", 0.03},
	{"MountainBikeRide", 0.02},
	{"VirtualRun", 0.02},
	{"Workout", 0.01},
}

// sportMetrics defines typical ranges for each sport type.
type sportMetrics struct {
	distanceMin  float64 // meters
	distanceMax  float64
	paceMinSpeed float64 // m/s (for pace calculation)
	paceMaxSpeed float64
	elevationMin float64 // meters
	elevationMax float64
	hasHeartRate bool
	hrMin        float64
	hrMax        float64
	hasPower     bool
	powerMin     float64
	powerMax     float64
	hasCadence   bool
	cadenceMin   float64
	cadenceMax   float64
}

var sportMetricsMap = map[string]sportMetrics{
	"Run": {
		distanceMin: 3000, distanceMax: 21000,
		paceMinSpeed: 2.5, paceMaxSpeed: 4.2, // ~6:00/km to ~4:00/km
		elevationMin: 0, elevationMax: 300,
		hasHeartRate: true, hrMin: 130, hrMax: 175,
		hasCadence: true, cadenceMin: 160, cadenceMax: 190,
	},
	"VirtualRun": {
		distanceMin: 3000, distanceMax: 15000,
		paceMinSpeed: 2.5, paceMaxSpeed: 4.0,
		elevationMin: 0, elevationMax: 50,
		hasHeartRate: true, hrMin: 130, hrMax: 175,
		hasCadence: true, cadenceMin: 165, cadenceMax: 185,
	},
	"TrailRun": {
		distanceMin: 5000, distanceMax: 25000,
		paceMinSpeed: 2.0, paceMaxSpeed: 3.5,
		elevationMin: 100, elevationMax: 800,
		hasHeartRate: true, hrMin: 135, hrMax: 180,
		hasCadence: true, cadenceMin: 155, cadenceMax: 180,
	},
	"Walk": {
		distanceMin: 2000, distanceMax: 10000,
		paceMinSpeed: 1.0, paceMaxSpeed: 1.8,
		elevationMin: 0, elevationMax: 100,
		hasHeartRate: true, hrMin: 90, hrMax: 130,
	},
	"Hike": {
		distanceMin: 5000, distanceMax: 20000,
		paceMinSpeed: 0.8, paceMaxSpeed: 1.5,
		elevationMin: 200, elevationMax: 1200,
		hasHeartRate: true, hrMin: 100, hrMax: 150,
	},
	"Ride": {
		distanceMin: 20000, distanceMax: 150000,
		paceMinSpeed: 6.0, paceMaxSpeed: 10.0, // ~22-36 km/h
		elevationMin: 100, elevationMax: 2000,
		hasHeartRate: true, hrMin: 120, hrMax: 170,
		hasPower: true, powerMin: 150, powerMax: 280,
		hasCadence: true, cadenceMin: 75, cadenceMax: 100,
	},
	"VirtualRide": {
		distanceMin: 15000, distanceMax: 60000,
		paceMinSpeed: 7.0, paceMaxSpeed: 10.5,
		elevationMin: 0, elevationMax: 500,
		hasHeartRate: true, hrMin: 130, hrMax: 175,
		hasPower: true, powerMin: 160, powerMax: 300,
		hasCadence: true, cadenceMin: 80, cadenceMax: 100,
	},
	"GravelRide": {
		distanceMin: 30000, distanceMax: 120000,
		paceMinSpeed: 5.0, paceMaxSpeed: 8.0,
		elevationMin: 300, elevationMax: 1500,
		hasHeartRate: true, hrMin: 125, hrMax: 165,
		hasPower: true, powerMin: 140, powerMax: 260,
		hasCadence: true, cadenceMin: 70, cadenceMax: 95,
	},
	"MountainBikeRide": {
		distanceMin: 15000, distanceMax: 50000,
		paceMinSpeed: 3.5, paceMaxSpeed: 6.5,
		elevationMin: 400, elevationMax: 1500,
		hasHeartRate: true, hrMin: 130, hrMax: 175,
		hasCadence: true, cadenceMin: 60, cadenceMax: 90,
	},
	"Swim": {
		distanceMin: 500, distanceMax: 4000,
		paceMinSpeed: 0.8, paceMaxSpeed: 1.5, // ~1:40/100m to ~1:07/100m
		elevationMin: 0, elevationMax: 0,
		hasHeartRate: true, hrMin: 110, hrMax: 160,
	},
	"Workout": {
		distanceMin: 0, distanceMax: 0,
		paceMinSpeed: 0, paceMaxSpeed: 0,
		elevationMin: 0, elevationMax: 0,
		hasHeartRate: true, hrMin: 100, hrMax: 170,
	},
}

// activityNames provides templates for generating activity names.
var activityNames = map[string][]string{
	"Run": {
		"Morning Run", "Easy Run", "Tempo Run", "Long Run",
		"Recovery Run", "Lunch Run", "Evening Run", "Interval Session",
		"Fartlek", "Threshold Run", "Base Run", "Progression Run",
	},
	"VirtualRun": {
		"Treadmill Run", "Indoor Run", "Virtual Miles",
	},
	"TrailRun": {
		"Trail Run", "Mountain Run", "Hill Repeats", "Forest Run",
		"Nature Run", "Technical Trail",
	},
	"Walk": {
		"Morning Walk", "Evening Stroll", "Dog Walk", "Lunch Walk",
		"Neighborhood Walk", "City Walk",
	},
	"Hike": {
		"Mountain Hike", "Trail Hike", "Day Hike", "Summit Attempt",
		"Nature Hike", "Scenic Hike",
	},
	"Ride": {
		"Morning Ride", "Club Ride", "Solo Ride", "Long Ride",
		"Tempo Ride", "Recovery Spin", "Coffee Ride", "Hills",
		"Weekend Ride", "Base Miles",
	},
	"VirtualRide": {
		"Zwift Session", "Indoor Trainer", "Virtual Ride",
		"ERG Workout", "Trainer Road",
	},
	"GravelRide": {
		"Gravel Adventure", "Mixed Surface", "Backroads",
		"Dirt Road Explore", "Gravel Grind",
	},
	"MountainBikeRide": {
		"MTB Session", "Trail Ride", "Singletrack",
		"Technical Trails", "Mountain Adventure",
	},
	"Swim": {
		"Pool Swim", "Open Water Swim", "Lap Swim",
		"Morning Swim", "Masters Swim",
	},
	"Workout": {
		"Strength Training", "Gym Session", "CrossFit",
		"Core Workout", "HIIT Session",
	},
}

// generateActivities creates demo activities spread across the specified time period.
func generateActivities(rng *rand.Rand, athleteID int64, gear []storage.Gear, count, months int) []storage.Activity {
	activities := make([]storage.Activity, 0, count)

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)
	totalDays := int(endDate.Sub(startDate).Hours() / 24)

	// Generate activities distributed across the time period
	for i := 0; i < count; i++ {
		// Pick a random day within the range
		dayOffset := rng.Intn(totalDays)
		activityDate := startDate.AddDate(0, 0, dayOffset)

		// Randomize time of day (morning bias)
		hour := 6 + rng.Intn(14) // 6 AM to 8 PM
		if rng.Float64() < 0.5 {
			hour = 6 + rng.Intn(4) // 50% chance of morning (6-10 AM)
		}
		minute := rng.Intn(60)
		activityDate = time.Date(activityDate.Year(), activityDate.Month(), activityDate.Day(),
			hour, minute, 0, 0, time.Local)

		// Pick sport type
		sportType := pickWeightedSportType(rng)

		// Generate activity
		activity := generateActivity(rng, athleteID, int64(1000000+i), sportType, activityDate, gear)
		activities = append(activities, activity)
	}

	return activities
}

// pickWeightedSportType selects a sport type based on defined weights.
func pickWeightedSportType(rng *rand.Rand) string {
	r := rng.Float64()
	cumulative := 0.0
	for _, sw := range sportTypeWeights {
		cumulative += sw.weight
		if r < cumulative {
			return sw.sportType
		}
	}
	return "Run" // fallback
}

// generateActivity creates a single demo activity.
func generateActivity(rng *rand.Rand, athleteID, activityID int64, sportType string, startTime time.Time, gear []storage.Gear) storage.Activity {
	metrics, ok := sportMetricsMap[sportType]
	if !ok {
		metrics = sportMetricsMap["Run"]
	}

	// Get appropriate route
	route := getRouteForSportTypeRng(rng, sportType)

	// Generate distance
	var distance float64
	if metrics.distanceMax > 0 {
		distance = randBetweenRng(rng, metrics.distanceMin, metrics.distanceMax)
	}

	// Generate speed and calculate time
	var movingTime int
	var avgSpeed, maxSpeed float64
	if metrics.paceMaxSpeed > 0 {
		avgSpeed = randBetweenRng(rng, metrics.paceMinSpeed, metrics.paceMaxSpeed)
		maxSpeed = avgSpeed * (1.1 + rng.Float64()*0.3) // 10-40% faster than avg
		if distance > 0 {
			movingTime = int(distance / avgSpeed)
		}
	} else {
		// Workout type without distance
		movingTime = randIntBetweenRng(rng, 1800, 5400) // 30-90 minutes
	}

	// Elapsed time slightly longer than moving time
	elapsedTime := movingTime + randIntBetweenRng(rng, 0, movingTime/10)

	// Generate elevation
	var elevationGain float64
	var elevHigh, elevLow *float64
	if metrics.elevationMax > 0 {
		elevationGain = randBetweenRng(rng, metrics.elevationMin, metrics.elevationMax)
		low := randBetweenRng(rng, 0, 200)
		high := low + elevationGain*0.8 + randBetweenRng(rng, 0, 100)
		elevLow = ptr(low)
		elevHigh = ptr(high)
	}

	// Generate heart rate
	var avgHR, maxHR *float64
	if metrics.hasHeartRate && rng.Float64() < 0.85 { // 85% have HR data
		avg := randBetweenRng(rng, metrics.hrMin, metrics.hrMax)
		max := avg + randBetweenRng(rng, 5, 25)
		avgHR = ptr(avg)
		maxHR = ptr(max)
	}

	// Generate power (cycling)
	var avgWatts, maxWatts, weightedWatts, kj *float64
	if metrics.hasPower && rng.Float64() < 0.7 { // 70% of rides have power
		avg := randBetweenRng(rng, metrics.powerMin, metrics.powerMax)
		max := avg * (1.5 + rng.Float64()*0.5)        // 50-100% higher than avg
		weighted := avg * (1.02 + rng.Float64()*0.08) // NP slightly higher
		energy := avg * float64(movingTime) / 1000    // kJ
		avgWatts = ptr(avg)
		maxWatts = ptr(max)
		weightedWatts = ptr(weighted)
		kj = ptr(energy)
	}

	// Generate cadence
	var avgCadence *float64
	if metrics.hasCadence && rng.Float64() < 0.75 {
		cadence := randBetweenRng(rng, metrics.cadenceMin, metrics.cadenceMax)
		avgCadence = ptr(cadence)
	}

	// Generate calories
	var calories *float64
	if rng.Float64() < 0.8 {
		// Rough estimate: distance-based for cardio, time-based for others
		var cal float64
		if distance > 0 {
			cal = distance / 1000 * (60 + rng.Float64()*40) // ~60-100 cal/km
		} else {
			cal = float64(movingTime) / 60 * (8 + rng.Float64()*4) // ~8-12 cal/min
		}
		calories = ptr(cal)
	}

	// Generate name
	names := activityNames[sportType]
	if len(names) == 0 {
		names = []string{sportType}
	}
	name := names[rng.Intn(len(names))]

	// Get gear
	gearID := getGearForSportTypeRng(rng, sportType, gear)

	// Kudos and social
	kudosCount := 0
	if rng.Float64() < 0.6 { // 60% of activities have kudos
		kudosCount = randIntBetweenRng(rng, 1, 25)
	}

	// Comments (20% of activities have comments)
	commentCount := 0
	if rng.Float64() < 0.2 {
		commentCount = randIntBetweenRng(rng, 1, 5)
	}

	// Photos (15% of activities have photos)
	photoCount := 0
	if rng.Float64() < 0.15 {
		photoCount = randIntBetweenRng(rng, 1, 4)
	}

	// Device names
	devices := []string{"Garmin Edge 840", "Garmin Forerunner 265", "Apple Watch Ultra", "Wahoo ELEMNT", "Coros Pace 3"}
	device := devices[rng.Intn(len(devices))]

	// Commute flag for short runs/rides
	commute := false
	if (sportType == "Ride" || sportType == "Run") && distance < 10000 && rng.Float64() < 0.1 {
		commute = true
	}

	// Trainer flag for virtual activities
	trainer := sportType == "VirtualRide" || sportType == "VirtualRun"

	activity := storage.Activity{
		ID:                   activityID,
		AthleteID:            athleteID,
		Name:                 name,
		SportType:            sportType,
		StartDate:            storage.SQLiteTime{Time: startTime.UTC()},
		StartDateLocal:       storage.SQLiteTime{Time: startTime},
		Timezone:             "America/Los_Angeles",
		LocationCity:         "San Francisco",
		LocationCountry:      "United States",
		Distance:             distance,
		MovingTime:           movingTime,
		ElapsedTime:          elapsedTime,
		TotalElevationGain:   elevationGain,
		ElevHigh:             elevHigh,
		ElevLow:              elevLow,
		AverageSpeed:         avgSpeed,
		MaxSpeed:             maxSpeed,
		AverageHeartrate:     avgHR,
		MaxHeartrate:         maxHR,
		AverageWatts:         avgWatts,
		MaxWatts:             maxWatts,
		WeightedAverageWatts: weightedWatts,
		Kilojoules:           kj,
		AverageCadence:       avgCadence,
		Calories:             calories,
		KudosCount:           kudosCount,
		CommentCount:         commentCount,
		PhotoCount:           photoCount,
		Commute:              commute,
		Trainer:              trainer,
		DeviceName:           device,
		GearID:               gearID,
		SummaryPolyline:      route.SummaryPolyline,
		StartLat:             ptr(route.StartLat),
		StartLng:             ptr(route.StartLng),
		EndLat:               ptr(route.EndLat),
		EndLng:               ptr(route.EndLng),
	}

	return activity
}

// Helper functions with rng parameter

func randBetweenRng(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}

func randIntBetweenRng(rng *rand.Rand, min, max int) int {
	if max <= min {
		return min
	}
	return min + rng.Intn(max-min+1)
}

func getRouteForSportTypeRng(rng *rand.Rand, sportType string) *RouteTemplate {
	var suitable []RouteTemplate

	for _, route := range routeTemplates {
		for _, st := range route.SportTypes {
			if st == sportType {
				suitable = append(suitable, route)
				break
			}
		}
	}

	if len(suitable) == 0 {
		return &routeTemplates[0]
	}

	return &suitable[rng.Intn(len(suitable))]
}

func getGearForSportTypeRng(rng *rand.Rand, sportType string, gear []storage.Gear) string {
	var bikes, shoes []string

	for _, item := range gear {
		if item.Retired {
			continue
		}
		if item.ID[0] == 'b' {
			bikes = append(bikes, item.ID)
		} else if item.ID[0] == 'g' {
			shoes = append(shoes, item.ID)
		}
	}

	switch sportType {
	case "Ride", "VirtualRide", "GravelRide", "MountainBikeRide":
		if len(bikes) > 0 {
			return bikes[rng.Intn(len(bikes))]
		}
	case "Run", "VirtualRun", "TrailRun":
		if len(shoes) > 0 {
			return shoes[rng.Intn(len(shoes))]
		}
	case "Walk", "Hike":
		if len(shoes) > 0 {
			return shoes[rng.Intn(len(shoes))]
		}
	}

	return ""
}
