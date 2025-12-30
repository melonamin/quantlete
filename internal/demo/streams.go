package demo

import (
	"encoding/json"
	"math"
	"math/rand"

	"github.com/melonamin/quantlete/internal/storage"
)

// generateActivityStreams creates synthetic time-series data for activities.
// Returns streams for watts (cycling), heartrate, velocity_smooth, and cadence.
func generateActivityStreams(rng *rand.Rand, activities []storage.Activity) []storage.ActivityStream {
	var streams []storage.ActivityStream

	for _, a := range activities {
		// Calculate number of data points (1 point per second)
		duration := a.MovingTime
		if duration < 60 {
			continue // Skip very short activities
		}

		// Limit stream size for performance (max 2 hours of second-by-second data)
		if duration > 7200 {
			duration = 7200
		}

		isCycling := isCyclingActivity(a.SportType)
		isRunning := isRunningActivity(a.SportType)
		hasHR := a.AverageHeartrate != nil && *a.AverageHeartrate > 0
		hasPower := isCycling && a.AverageWatts != nil && *a.AverageWatts > 0

		// Generate heart rate stream (85% of activities with HR data)
		if hasHR && rng.Float64() < 0.85 {
			hrStream := generateHRStream(rng, duration, *a.AverageHeartrate, a.MaxHeartrate)
			streams = append(streams, storage.ActivityStream{
				ActivityID:   a.ID,
				StreamType:   "heartrate",
				OriginalSize: len(hrStream),
				Resolution:   "high",
				SeriesType:   "time",
				Data:         encodeStream(hrStream),
			})
		}

		// Generate power stream for cycling (70% of rides with power)
		if hasPower && rng.Float64() < 0.70 {
			wattsStream := generateWattsStream(rng, duration, *a.AverageWatts, a.MaxWatts)
			streams = append(streams, storage.ActivityStream{
				ActivityID:   a.ID,
				StreamType:   "watts",
				OriginalSize: len(wattsStream),
				Resolution:   "high",
				SeriesType:   "time",
				Data:         encodeStream(wattsStream),
			})
		}

		// Generate velocity stream for running/cycling
		if (isRunning || isCycling) && a.AverageSpeed > 0 {
			velocityStream := generateVelocityStream(rng, duration, a.AverageSpeed, a.MaxSpeed)
			streams = append(streams, storage.ActivityStream{
				ActivityID:   a.ID,
				StreamType:   "velocity_smooth",
				OriginalSize: len(velocityStream),
				Resolution:   "high",
				SeriesType:   "time",
				Data:         encodeStream(velocityStream),
			})
		}

		// Generate cadence stream
		if a.AverageCadence != nil && *a.AverageCadence > 0 {
			cadenceStream := generateCadenceStream(rng, duration, *a.AverageCadence, isRunning)
			streams = append(streams, storage.ActivityStream{
				ActivityID:   a.ID,
				StreamType:   "cadence",
				OriginalSize: len(cadenceStream),
				Resolution:   "high",
				SeriesType:   "time",
				Data:         encodeStream(cadenceStream),
			})
		}
	}

	return streams
}

// generateHRStream creates a realistic heart rate time series.
func generateHRStream(rng *rand.Rand, duration int, avgHR float64, maxHR *float64) []float64 {
	stream := make([]float64, duration)

	// Start with warmup HR (lower than average)
	warmupDuration := min(duration/10, 300) // up to 5 minutes warmup
	baseHR := avgHR * 0.85

	// Max HR defaults to 15% above average if not provided
	peakHR := avgHR * 1.15
	if maxHR != nil && *maxHR > avgHR {
		peakHR = *maxHR
	}

	// Generate with smooth transitions using random walk
	currentHR := baseHR
	targetHR := avgHR

	for i := 0; i < duration; i++ {
		// Warmup phase
		if i < warmupDuration {
			progress := float64(i) / float64(warmupDuration)
			targetHR = baseHR + (avgHR-baseHR)*progress
		} else if i > duration-warmupDuration/2 {
			// Cooldown phase
			progress := float64(i-(duration-warmupDuration/2)) / float64(warmupDuration/2)
			targetHR = avgHR - (avgHR-baseHR)*progress*0.5
		} else {
			// Main phase: occasional intervals/efforts
			if rng.Float64() < 0.002 { // ~0.2% chance per second to start interval
				targetHR = avgHR + (peakHR-avgHR)*rng.Float64()
			} else if rng.Float64() < 0.01 { // 1% chance to return to base
				targetHR = avgHR * (0.95 + rng.Float64()*0.1)
			}
		}

		// Smooth transition toward target
		diff := targetHR - currentHR
		currentHR += diff * 0.02 // gradual approach

		// Add small noise
		noise := (rng.Float64() - 0.5) * 4
		hr := currentHR + noise

		// Clamp to reasonable range
		hr = math.Max(60, math.Min(peakHR+5, hr))
		stream[i] = math.Round(hr)
	}

	return stream
}

// generateWattsStream creates a realistic power time series for cycling.
func generateWattsStream(rng *rand.Rand, duration int, avgWatts float64, maxWatts *float64) []float64 {
	stream := make([]float64, duration)

	// Peak power defaults to 2x average if not provided
	peakWatts := avgWatts * 2
	if maxWatts != nil && *maxWatts > avgWatts {
		peakWatts = *maxWatts
	}

	// Power is much more variable than HR
	currentWatts := avgWatts * 0.7 // start lower
	targetWatts := avgWatts

	for i := 0; i < duration; i++ {
		// Occasional sprints/climbs
		if rng.Float64() < 0.005 {
			// Sprint: high power for short duration
			targetWatts = avgWatts * (1.5 + rng.Float64()*0.5)
		} else if rng.Float64() < 0.01 {
			// Sustained effort
			targetWatts = avgWatts * (1.1 + rng.Float64()*0.2)
		} else if rng.Float64() < 0.02 {
			// Recovery/descent
			targetWatts = avgWatts * (0.3 + rng.Float64()*0.4)
		} else if rng.Float64() < 0.03 {
			// Back to normal
			targetWatts = avgWatts * (0.9 + rng.Float64()*0.2)
		}

		// Smooth transition
		diff := targetWatts - currentWatts
		currentWatts += diff * 0.05

		// Power has more noise than HR
		noise := (rng.Float64() - 0.5) * avgWatts * 0.3
		watts := currentWatts + noise

		// Clamp (can have zeros during coasting)
		watts = math.Max(0, math.Min(peakWatts*1.1, watts))
		stream[i] = math.Round(watts)
	}

	return stream
}

// generateVelocityStream creates a speed time series (m/s).
func generateVelocityStream(rng *rand.Rand, duration int, avgSpeed, maxSpeed float64) []float64 {
	stream := make([]float64, duration)

	// If no max speed, estimate it
	if maxSpeed <= avgSpeed {
		maxSpeed = avgSpeed * 1.4
	}

	currentSpeed := avgSpeed * 0.5 // start slow
	targetSpeed := avgSpeed

	for i := 0; i < duration; i++ {
		// Warmup
		if i < duration/15 {
			progress := float64(i) / float64(duration/15)
			targetSpeed = avgSpeed * (0.5 + 0.5*progress)
		} else if i > duration-duration/20 {
			// Cooldown
			progress := float64(i-(duration-duration/20)) / float64(duration/20)
			targetSpeed = avgSpeed * (1.0 - 0.3*progress)
		} else {
			// Normal variation
			if rng.Float64() < 0.01 {
				targetSpeed = avgSpeed * (0.8 + rng.Float64()*0.5)
			}
		}

		// Smooth transition
		diff := targetSpeed - currentSpeed
		currentSpeed += diff * 0.03

		// Add noise
		noise := (rng.Float64() - 0.5) * avgSpeed * 0.15
		speed := currentSpeed + noise

		// Clamp
		speed = math.Max(0, math.Min(maxSpeed*1.05, speed))
		stream[i] = math.Round(speed*100) / 100 // 2 decimal places
	}

	return stream
}

// generateCadenceStream creates a cadence time series (rpm/spm).
func generateCadenceStream(rng *rand.Rand, duration int, avgCadence float64, isRunning bool) []float64 {
	stream := make([]float64, duration)

	// Running cadence is more stable than cycling
	variance := 0.15
	if isRunning {
		variance = 0.08
	}

	currentCadence := avgCadence
	targetCadence := avgCadence

	for i := 0; i < duration; i++ {
		// Occasional cadence changes
		if rng.Float64() < 0.01 {
			targetCadence = avgCadence * (1 - variance + rng.Float64()*variance*2)
		}

		// Smooth transition
		diff := targetCadence - currentCadence
		currentCadence += diff * 0.02

		// Add noise
		noise := (rng.Float64() - 0.5) * avgCadence * variance * 0.5
		cadence := currentCadence + noise

		// Clamp (can be 0 when coasting on bike)
		minCadence := 0.0
		if isRunning {
			minCadence = avgCadence * 0.7
		}
		cadence = math.Max(minCadence, math.Min(avgCadence*1.3, cadence))
		stream[i] = math.Round(cadence)
	}

	return stream
}

// encodeStream converts a float64 slice to JSON for storage.
func encodeStream(data []float64) json.RawMessage {
	encoded, _ := json.Marshal(data)
	return encoded
}

func isCyclingActivity(sportType string) bool {
	switch sportType {
	case "Ride", "VirtualRide", "GravelRide", "MountainBikeRide", "EBikeRide":
		return true
	}
	return false
}

func isRunningActivity(sportType string) bool {
	switch sportType {
	case "Run", "VirtualRun", "TrailRun":
		return true
	}
	return false
}
