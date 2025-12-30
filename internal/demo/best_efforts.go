package demo

import (
	"math/rand"
	"sort"

	"github.com/melonamin/quantlete/internal/storage"
)

// Standard running distance types and their distances in meters.
var standardDistances = []struct {
	distanceType string
	name         string
	distanceM    float64
}{
	{"400m", "400m", 400},
	{"half_mile", "1/2 mile", 804.672},
	{"1k", "1k", 1000},
	{"1_mile", "1 mile", 1609.34},
	{"2_mile", "2 mile", 3218.69},
	{"5k", "5k", 5000},
	{"10k", "10k", 10000},
	{"15k", "15k", 15000},
	{"10_mile", "10 mile", 16093.4},
	{"half_marathon", "Half-Marathon", 21097.5},
	{"marathon", "Marathon", 42195},
}

// generateBestEfforts creates best effort records for running activities.
// Only generates efforts for activities that are long enough for each distance.
func generateBestEfforts(rng *rand.Rand, athleteID int64, activities []storage.Activity) []storage.BestEffort {
	var allEfforts []storage.BestEffort

	// Filter to running activities only
	var runActivities []storage.Activity
	for _, a := range activities {
		switch a.SportType {
		case "Run", "VirtualRun", "TrailRun":
			runActivities = append(runActivities, a)
		}
	}

	if len(runActivities) == 0 {
		return nil
	}

	// Base pace for an amateur runner (around 5:30/km = 330 sec/km)
	// This will be varied per activity
	basePaceSecPerKm := 330.0

	// Track best times per distance type for PR ranking
	bestTimes := make(map[string][]effortWithTime)

	for _, activity := range runActivities {
		// Skip activities without meaningful distance
		if activity.Distance < 400 {
			continue
		}

		activityDistanceM := activity.Distance

		// Determine activity's pace variation (some runs faster, some slower)
		// Use activity's average speed if available, otherwise randomize
		paceMultiplier := 0.9 + rng.Float64()*0.3 // 0.9x to 1.2x of base pace

		for _, dist := range standardDistances {
			// Only generate effort if activity is long enough
			if activityDistanceM < dist.distanceM*0.95 {
				continue
			}

			// Calculate effort time based on distance and pace
			// Longer distances have slightly slower per-km pace (fatigue)
			fatigueFactor := 1.0 + (dist.distanceM/42195)*0.15 // up to 15% slower at marathon
			effortPace := basePaceSecPerKm * paceMultiplier * fatigueFactor

			// Add some randomness to each effort
			effortPace *= 0.95 + rng.Float64()*0.1

			elapsedTime := int(dist.distanceM / 1000 * effortPace)
			movingTime := elapsedTime - rng.Intn(elapsedTime/50+1)
			if movingTime < 1 {
				movingTime = elapsedTime
			}

			effort := storage.BestEffort{
				AthleteID:    athleteID,
				ActivityID:   activity.ID,
				SportType:    activity.SportType,
				DistanceType: dist.distanceType,
				Name:         dist.name,
				DistanceM:    dist.distanceM,
				ElapsedTimeS: elapsedTime,
				MovingTimeS:  ptr(movingTime),
				StartDate:    &activity.StartDate,
			}

			allEfforts = append(allEfforts, effort)
			bestTimes[dist.distanceType] = append(bestTimes[dist.distanceType], effortWithTime{
				idx:  len(allEfforts) - 1,
				time: elapsedTime,
			})
		}
	}

	// Assign PR ranks (1 = fastest, 2 = second fastest, etc.)
	for _, efforts := range bestTimes {
		// Sort by time
		sort.Slice(efforts, func(i, j int) bool {
			return efforts[i].time < efforts[j].time
		})

		// Assign ranks to top 3
		for rank, e := range efforts {
			if rank >= 3 {
				break
			}
			allEfforts[e.idx].PRRank = ptr(rank + 1)
		}
	}

	return allEfforts
}

type effortWithTime struct {
	idx  int
	time int
}
