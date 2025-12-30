package demo

import (
	"fmt"
	"math/rand"

	"github.com/melonamin/quantlete/internal/storage"
)

// Segment name components for procedural generation.
var (
	// Cycling segment name parts
	cyclingPrefixes = []string{
		"Mt", "Old", "Upper", "Lower", "North", "South", "East", "West",
		"Hidden", "Scenic", "Rocky", "Pine", "Oak", "Redwood", "Canyon",
	}
	cyclingNames = []string{
		"Hill", "Ridge", "Summit", "Peak", "Pass", "Grade", "Climb",
		"Mountain", "Heights", "Bluff", "Vista", "Overlook", "Loop",
		"Road", "Way", "Lane", "Drive", "Boulevard", "Avenue", "Trail",
	}
	cyclingSuffixes = []string{
		"", "KOM", "Sprint", "Climb", "Challenge", "Classic", "Express",
		"Ascent", "Descent", "Switchbacks", "Hairpins", "North Face", "South Side",
	}

	// Running segment name parts
	runningPrefixes = []string{
		"Park", "Trail", "Lake", "River", "Beach", "Bay", "Harbor",
		"Forest", "Meadow", "Valley", "Creek", "Spring", "Downtown",
	}
	runningNames = []string{
		"Loop", "Mile", "Sprint", "Dash", "Run", "Path", "Track",
		"Circuit", "Course", "Route", "Trail", "Way", "Promenade",
	}
	runningSuffixes = []string{
		"", "5K", "10K", "1 Mile", "Half", "PR Course", "Flat", "Fast",
		"Scenic", "Challenge", "Classic", "Out & Back",
	}

	// Geographic regions with base coordinates
	regions = []struct {
		name    string
		baseLat float64
		baseLng float64
		country string
	}{
		{"San Francisco Bay Area", 37.7749, -122.4194, "United States"},
		{"Los Angeles", 34.0522, -118.2437, "United States"},
		{"New York Metro", 40.7128, -74.0060, "United States"},
		{"London", 51.5074, -0.1278, "United Kingdom"},
		{"Paris", 48.8566, 2.3522, "France"},
		{"Sydney", -33.8688, 151.2093, "Australia"},
		{"Tokyo", 35.6762, 139.6503, "Japan"},
		{"Berlin", 52.5200, 13.4050, "Germany"},
		{"Amsterdam", 52.3676, 4.9041, "Netherlands"},
		{"Denver", 39.7392, -104.9903, "United States"},
		{"Seattle", 47.6062, -122.3321, "United States"},
		{"Portland", 45.5152, -122.6784, "United States"},
		{"Chicago", 41.8781, -87.6298, "United States"},
		{"Boston", 42.3601, -71.0589, "United States"},
		{"Austin", 30.2672, -97.7431, "United States"},
		{"Vancouver", 49.2827, -123.1207, "Canada"},
		{"Toronto", 43.6532, -79.3832, "Canada"},
		{"Melbourne", -37.8136, 144.9631, "Australia"},
		{"Barcelona", 41.3851, 2.1734, "Spain"},
		{"Milan", 45.4642, 9.1900, "Italy"},
	}

	// Sample polylines (will be reused with slight variations)
	samplePolylines = []string{
		"ssfeF`ifjVuCjA}ClAeDbAgDr@mDd@uDTyD@{DEwDUsDe@oDw@mDaAiDgA",
		"cnfeF~yhjVuDdBaDvBsCfC_CrCcBlD{AfE}@bFi@`GQ~FBhGXrGr@nGfAfGvArFfBvE~BjE",
		"yngeF|dgjVmBuAaBgBiAoBw@wB_@aC",
		"cweeF|xdjV_@lBYpB]nBa@lBa@hBg@dBo@~A",
		"gyeeF~odjVkCiByBuB}AoCqAgDcA{Dq@eFa@gFEiFZeFp@cFdA}ErAwEdBmE",
		"afreF`rejVmA{FgBsFkCwEuDwDeFyCgFmB_GaBaGkA_Hq@yGYcH@kHPwGh@wGdAmG~AeGtBcGjCcF",
		"k~deF~nbjVaBf@qBb@cC`@oC^yCZwCVeDR_ERuED}E?cFGqFOyF]_Gk@",
		"_p~iF~ps|UdA~@dArAdA`BfA`BhAbB",
	}
)

// generateSegments creates a large number of procedurally generated segments.
func generateSegments(rng *rand.Rand, athleteID int64, count int) []storage.Segment {
	if count <= 0 {
		count = 4000
	}

	segments := make([]storage.Segment, count)
	usedNames := make(map[string]bool)

	// 60% cycling, 40% running
	cyclingCount := count * 60 / 100

	for i := 0; i < count; i++ {
		isCycling := i < cyclingCount

		// Pick a random region
		region := regions[rng.Intn(len(regions))]

		// Generate unique name
		var name string
		for {
			if isCycling {
				name = generateCyclingName(rng)
			} else {
				name = generateRunningName(rng)
			}
			// Add region hint for uniqueness
			if rng.Float64() < 0.3 {
				name = fmt.Sprintf("%s %s", region.name[:3], name)
			}
			if !usedNames[name] {
				usedNames[name] = true
				break
			}
		}

		var seg storage.Segment
		if isCycling {
			seg = generateCyclingSegment(rng, i, name, region.baseLat, region.baseLng, region.country)
		} else {
			seg = generateRunningSegment(rng, i, name, region.baseLat, region.baseLng, region.country)
		}

		segments[i] = seg
	}

	return segments
}

func generateCyclingName(rng *rand.Rand) string {
	prefix := cyclingPrefixes[rng.Intn(len(cyclingPrefixes))]
	name := cyclingNames[rng.Intn(len(cyclingNames))]
	suffix := cyclingSuffixes[rng.Intn(len(cyclingSuffixes))]

	if suffix == "" {
		return fmt.Sprintf("%s %s", prefix, name)
	}
	return fmt.Sprintf("%s %s %s", prefix, name, suffix)
}

func generateRunningName(rng *rand.Rand) string {
	prefix := runningPrefixes[rng.Intn(len(runningPrefixes))]
	name := runningNames[rng.Intn(len(runningNames))]
	suffix := runningSuffixes[rng.Intn(len(runningSuffixes))]

	if suffix == "" {
		return fmt.Sprintf("%s %s", prefix, name)
	}
	return fmt.Sprintf("%s %s %s", prefix, name, suffix)
}

func generateCyclingSegment(rng *rand.Rand, idx int, name string, baseLat, baseLng float64, country string) storage.Segment {
	// Segment type distribution: 40% climbs, 30% sprints, 30% mixed
	segType := rng.Float64()

	var distance, avgGrade, maxGrade, elevLow, elevHigh float64
	var climbCat int

	if segType < 0.4 {
		// Climb segment
		distance = randBetweenRng(rng, 500, 15000)      // 0.5-15km
		avgGrade = randBetweenRng(rng, 3.0, 12.0)       // 3-12%
		maxGrade = avgGrade + randBetweenRng(rng, 2, 8) // Max is higher
		elevLow = randBetweenRng(rng, 0, 500)
		elevHigh = elevLow + distance*avgGrade/100

		// Climb category based on difficulty (distance * grade)
		difficulty := distance * avgGrade / 1000
		switch {
		case difficulty > 80:
			climbCat = 0 // HC
		case difficulty > 50:
			climbCat = 1 // Cat 1
		case difficulty > 25:
			climbCat = 2 // Cat 2
		case difficulty > 12:
			climbCat = 3 // Cat 3
		case difficulty > 5:
			climbCat = 4 // Cat 4
		default:
			climbCat = 5 // Not categorized
		}
	} else if segType < 0.7 {
		// Sprint segment (flat/fast)
		distance = randBetweenRng(rng, 200, 2000) // 200m-2km
		avgGrade = randBetweenRng(rng, -1.0, 1.0) // Flat
		maxGrade = randBetweenRng(rng, 0, 3.0)    // Slight undulation
		elevLow = randBetweenRng(rng, 0, 200)     // Low elevation
		elevHigh = elevLow + randBetweenRng(rng, 0, 20)
		climbCat = 5 // Not categorized
	} else {
		// Mixed/rolling segment
		distance = randBetweenRng(rng, 1000, 8000) // 1-8km
		avgGrade = randBetweenRng(rng, 0.5, 4.0)   // Gentle grade
		maxGrade = avgGrade + randBetweenRng(rng, 3, 10)
		elevLow = randBetweenRng(rng, 0, 300)
		elevHigh = elevLow + distance*avgGrade/100
		climbCat = 5 // Usually not categorized
		if avgGrade > 2.5 && distance > 3000 {
			climbCat = 4 // Maybe Cat 4
		}
	}

	// Generate coordinates with some variation from base
	latOffset := randBetweenRng(rng, -0.5, 0.5)
	lngOffset := randBetweenRng(rng, -0.5, 0.5)
	startLat := baseLat + latOffset
	startLng := baseLng + lngOffset
	// End point based on distance
	distDegrees := distance / 111000 // rough meters to degrees
	endLat := startLat + distDegrees*0.7*rng.Float64()
	endLng := startLng + distDegrees*0.7*rng.Float64()*(1+0.5*rng.Float64())

	// Calculate PR time
	speed := 30.0 - avgGrade*2
	if speed < 12 {
		speed = 12
	}
	prTime := int(distance / (speed * 1000 / 3600))
	prTime = int(float64(prTime) * (0.85 + rng.Float64()*0.3))

	return storage.Segment{
		ID:                   int64(3000000 + idx),
		Name:                 name,
		ActivityType:         "Ride",
		Distance:             distance,
		AverageGrade:         avgGrade,
		MaximumGrade:         maxGrade,
		ElevationHigh:        elevHigh,
		ElevationLow:         elevLow,
		ClimbCategory:        climbCat,
		StartLat:             ptr(startLat),
		StartLng:             ptr(startLng),
		EndLat:               ptr(endLat),
		EndLng:               ptr(endLng),
		Starred:              rng.Float64() < 0.05, // 5% starred
		Polyline:             samplePolylines[rng.Intn(len(samplePolylines))],
		AthleteEffortCount:   ptr(randIntBetweenRng(rng, 0, 50)),
		AthletePRElapsedTime: ptr(prTime),
	}
}

func generateRunningSegment(rng *rand.Rand, idx int, name string, baseLat, baseLng float64, country string) storage.Segment {
	// Segment type: 50% flat/fast, 30% trail/hilly, 20% track
	segType := rng.Float64()

	var distance, avgGrade, maxGrade, elevLow, elevHigh float64

	if segType < 0.5 {
		// Flat/fast segment
		distance = randBetweenRng(rng, 200, 5000) // 200m-5km
		avgGrade = randBetweenRng(rng, -0.5, 0.5) // Very flat
		maxGrade = randBetweenRng(rng, 0, 2.0)
		elevLow = randBetweenRng(rng, 0, 100)
		elevHigh = elevLow + randBetweenRng(rng, 0, 10)
	} else if segType < 0.8 {
		// Trail/hilly segment
		distance = randBetweenRng(rng, 500, 10000) // 500m-10km
		avgGrade = randBetweenRng(rng, 1.0, 6.0)   // Uphill
		maxGrade = avgGrade + randBetweenRng(rng, 2, 8)
		elevLow = randBetweenRng(rng, 50, 300)
		elevHigh = elevLow + distance*avgGrade/100
	} else {
		// Track segment (very flat, standard distances)
		trackDistances := []float64{200, 400, 800, 1000, 1609, 3000, 5000}
		distance = trackDistances[rng.Intn(len(trackDistances))]
		avgGrade = 0
		maxGrade = 0.5
		elevLow = randBetweenRng(rng, 0, 50)
		elevHigh = elevLow + 2
	}

	// Coordinates
	latOffset := randBetweenRng(rng, -0.3, 0.3)
	lngOffset := randBetweenRng(rng, -0.3, 0.3)
	startLat := baseLat + latOffset
	startLng := baseLng + lngOffset
	distDegrees := distance / 111000
	endLat := startLat + distDegrees*0.5*rng.Float64()
	endLng := startLng + distDegrees*0.5*rng.Float64()

	// PR time (running pace 3:30-6:00/km depending on grade)
	pace := 4.5 + avgGrade*0.4
	prTime := int(distance / 1000 * pace * 60)
	prTime = int(float64(prTime) * (0.85 + rng.Float64()*0.3))

	return storage.Segment{
		ID:                   int64(3000000 + idx),
		Name:                 name,
		ActivityType:         "Run",
		Distance:             distance,
		AverageGrade:         avgGrade,
		MaximumGrade:         maxGrade,
		ElevationHigh:        elevHigh,
		ElevationLow:         elevLow,
		ClimbCategory:        5, // Running segments not categorized
		StartLat:             ptr(startLat),
		StartLng:             ptr(startLng),
		EndLat:               ptr(endLat),
		EndLng:               ptr(endLng),
		Starred:              rng.Float64() < 0.05, // 5% starred
		Polyline:             samplePolylines[rng.Intn(len(samplePolylines))],
		AthleteEffortCount:   ptr(randIntBetweenRng(rng, 0, 30)),
		AthletePRElapsedTime: ptr(prTime),
	}
}

// generateSegmentEfforts creates segment efforts linked to activities.
// With many segments, each activity only hits a small subset of them.
func generateSegmentEfforts(rng *rand.Rand, athleteID int64, activities []storage.Activity, segments []storage.Segment) []storage.SegmentEffort {
	var efforts []storage.SegmentEffort
	effortID := int64(4000000)

	// Separate segments by type
	rideSegments := make([]storage.Segment, 0)
	runSegments := make([]storage.Segment, 0)
	for _, s := range segments {
		if s.ActivityType == "Ride" {
			rideSegments = append(rideSegments, s)
		} else {
			runSegments = append(runSegments, s)
		}
	}

	// Track effort counts per segment to limit to <100
	effortCounts := make(map[int64]int)
	const maxEffortsPerSegment = 99

	for _, activity := range activities {
		// Determine applicable segments
		var applicableSegments []storage.Segment
		switch activity.SportType {
		case "Ride", "VirtualRide", "GravelRide", "MountainBikeRide":
			applicableSegments = rideSegments
		case "Run", "VirtualRun", "TrailRun":
			applicableSegments = runSegments
		default:
			continue
		}

		if len(applicableSegments) == 0 {
			continue
		}

		// 25% chance an activity has segment efforts (reduced from 60%)
		if rng.Float64() > 0.25 {
			continue
		}

		// Pick 1-5 random segments for this activity
		numSegments := 1 + rng.Intn(5)
		if numSegments > 10 {
			numSegments = 10
		}

		// Pick random segments (not shuffling entire array for efficiency)
		picked := make(map[int]bool)
		for j := 0; j < numSegments && j < len(applicableSegments); j++ {
			// Try to find a segment that hasn't hit the limit
			for attempts := 0; attempts < 20; attempts++ {
				idx := rng.Intn(len(applicableSegments))
				if picked[idx] {
					continue
				}
				seg := applicableSegments[idx]
				if effortCounts[seg.ID] >= maxEffortsPerSegment {
					continue
				}
				picked[idx] = true

				// Create effort
				effort := createSegmentEffort(rng, effortID, seg, activity, athleteID)
				efforts = append(efforts, effort)
				effortCounts[seg.ID]++
				effortID++
				break
			}
		}
	}

	return efforts
}

func createSegmentEffort(rng *rand.Rand, effortID int64, seg storage.Segment, activity storage.Activity, athleteID int64) storage.SegmentEffort {
	prTime := 0
	if seg.AthletePRElapsedTime != nil {
		prTime = *seg.AthletePRElapsedTime
	} else {
		prTime = int(seg.Distance / 5)
	}

	// Effort is 0-40% slower than PR
	elapsedTime := int(float64(prTime) * (1.0 + rng.Float64()*0.4))
	movingTime := elapsedTime - randIntBetweenRng(rng, 0, elapsedTime/15)
	if movingTime < 1 {
		movingTime = 1
	}

	// PR rank
	var prRank *int
	if rng.Float64() < 0.05 {
		prRank = ptr(1)
	} else if rng.Float64() < 0.1 {
		prRank = ptr(2)
	} else if rng.Float64() < 0.15 {
		prRank = ptr(3)
	}

	// Heart rate and power
	var avgWatts, avgHR *float64
	var maxHR *int
	if activity.AverageWatts != nil {
		w := *activity.AverageWatts * (1.1 + rng.Float64()*0.2)
		avgWatts = ptr(w)
	}
	if activity.AverageHeartrate != nil {
		hr := *activity.AverageHeartrate + randBetweenRng(rng, 5, 15)
		avgHR = ptr(hr)
		maxHR = ptr(int(hr + randBetweenRng(rng, 5, 20)))
	}

	return storage.SegmentEffort{
		ID:               effortID,
		SegmentID:        seg.ID,
		ActivityID:       activity.ID,
		AthleteID:        athleteID,
		Name:             seg.Name,
		ElapsedTime:      elapsedTime,
		MovingTime:       movingTime,
		StartDate:        &activity.StartDate,
		StartDateLocal:   &activity.StartDateLocal,
		Distance:         seg.Distance,
		AverageWatts:     avgWatts,
		AverageHeartrate: avgHR,
		MaxHeartrate:     maxHR,
		PRRank:           prRank,
		Country:          "United States",
	}
}
