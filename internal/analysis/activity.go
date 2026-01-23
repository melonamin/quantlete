// Package analysis provides activity analysis functions.
package analysis

import (
	"slices"
)

// Split calculation constants.
const (
	// minPartialSplitRatio is the minimum ratio of a partial split to include (10% of full split).
	minPartialSplitRatio = 0.1
	// fullSplitThreshold is the threshold ratio to consider a split as "full" for comparison (90%).
	fullSplitThreshold = 0.9
	// defaultSplitLengthM is the default split length in meters (1km).
	defaultSplitLengthM = 1000.0
	// maxSplitDuration is the maximum reasonable duration for a single split (24 hours).
	// Durations exceeding this indicate corrupt stream data (e.g., GPS gaps, clock drift).
	// This prevents both integer overflow and nonsensical values in analysis output.
	maxSplitDuration = 24 * 60 * 60 // 86400 seconds = 24 hours
)

// Split represents a distance-based split (e.g., per-km or per-mile).
type Split struct {
	Index        int     `json:"index"`       // 1-indexed split number
	DistanceM    float64 `json:"distance_m"`  // Distance of this split in meters
	DurationS    int     `json:"duration_s"`  // Duration of this split in seconds
	PaceSecsPerM float64 `json:"pace_secs_m"` // Pace in seconds per meter
	AvgHR        float64 `json:"avg_hr"`      // Average heart rate (if available)
	AvgWatts     float64 `json:"avg_watts"`   // Average power (if available)
	ElevGain     float64 `json:"elev_gain"`   // Elevation gain in meters
	ElevLoss     float64 `json:"elev_loss"`   // Elevation loss in meters
}

// SplitsResult contains computed splits and summary.
type SplitsResult struct {
	Splits       []Split `json:"splits"`
	SplitLength  float64 `json:"split_length_m"` // Length of each split in meters (e.g., 1000 for 1km)
	TotalSplits  int     `json:"total_splits"`
	FastestSplit int     `json:"fastest_split"` // 1-indexed
	SlowestSplit int     `json:"slowest_split"` // 1-indexed
}

// ComputeSplits calculates distance-based splits from stream data.
// Parameters:
//   - distance: cumulative distance array in meters
//   - time: cumulative time array in seconds
//   - hr: heart rate array (optional, can be nil)
//   - watts: power array (optional, can be nil)
//   - altitude: altitude array (optional, can be nil)
//   - splitLengthM: distance per split in meters (e.g., 1000 for 1km)
//
//nolint:gosec // G602: bounds are checked at function entry before loop access
func ComputeSplits(distance, time, hr, watts, altitude []float64, splitLengthM float64) *SplitsResult {
	if len(distance) < 2 || len(time) < 2 || len(distance) != len(time) {
		return nil
	}

	if splitLengthM <= 0 {
		splitLengthM = defaultSplitLengthM
	}

	hasHR := len(hr) == len(distance)
	hasWatts := len(watts) == len(distance)
	hasAltitude := len(altitude) == len(distance)

	var splits []Split
	splitStart := 0
	splitStartTime := time[0]
	nextSplitDist := splitLengthM
	splitIndex := 1

	// Track running sums for averages
	var hrSum, wattsSum float64
	var hrCount, wattsCount int

	for i := 1; i < len(distance); i++ {
		// Accumulate HR and watts for averaging
		if hasHR && hr[i] > 0 {
			hrSum += hr[i]
			hrCount++
		}
		if hasWatts && watts[i] > 0 {
			wattsSum += watts[i]
			wattsCount++
		}

		// Handle multiple split boundaries crossed in a single sample
		for distance[i] >= nextSplitDist {
			prevDist := distance[i-1]
			currDist := distance[i]
			prevTime := time[i-1]
			currTime := time[i]

			// Guard against division by zero when consecutive samples have equal distance
			if currDist == prevDist {
				break
			}

			// Linear interpolation factor
			fraction := (nextSplitDist - prevDist) / (currDist - prevDist)
			splitEndTime := prevTime + fraction*(currTime-prevTime)

			// Calculate duration - skip splits with unreasonable durations (indicates corrupt data)
			duration := splitEndTime - splitStartTime
			if duration <= 0 || duration > maxSplitDuration {
				// Skip this split but continue processing - likely GPS gap or clock drift
				splitIndex++
				splitStartTime = splitEndTime
				nextSplitDist += splitLengthM
				splitStart = i
				hrSum, wattsSum = 0, 0
				hrCount, wattsCount = 0, 0
				continue
			}

			split := Split{
				Index:        splitIndex,
				DistanceM:    splitLengthM,
				DurationS:    int(duration),
				PaceSecsPerM: duration / splitLengthM,
			}

			// Calculate averages
			if hrCount > 0 {
				split.AvgHR = hrSum / float64(hrCount)
			}
			if wattsCount > 0 {
				split.AvgWatts = wattsSum / float64(wattsCount)
			}

			// Calculate elevation change
			if hasAltitude && splitStart < len(altitude) && i < len(altitude) {
				elevGain, elevLoss := calculateElevation(altitude[splitStart : i+1])
				split.ElevGain = elevGain
				split.ElevLoss = elevLoss
			}

			splits = append(splits, split)

			// Reset for next split - use interpolated time as new start
			splitIndex++
			splitStartTime = splitEndTime
			nextSplitDist += splitLengthM
			splitStart = i
			hrSum, wattsSum = 0, 0
			hrCount, wattsCount = 0, 0
		}
	}

	// Handle final partial split
	lastDist := distance[len(distance)-1]
	if lastDist > (float64(splitIndex)-1)*splitLengthM {
		partialDist := lastDist - (float64(splitIndex)-1)*splitLengthM
		partialTime := time[len(time)-1] - splitStartTime

		// Include partial split only if it meets minimum ratio and has valid duration
		if partialDist > splitLengthM*minPartialSplitRatio && partialDist > 0 &&
			partialTime > 0 && partialTime <= maxSplitDuration {
			split := Split{
				Index:        splitIndex,
				DistanceM:    partialDist,
				DurationS:    int(partialTime),
				PaceSecsPerM: partialTime / partialDist,
			}

			if hrCount > 0 {
				split.AvgHR = hrSum / float64(hrCount)
			}
			if wattsCount > 0 {
				split.AvgWatts = wattsSum / float64(wattsCount)
			}

			if hasAltitude && splitStart < len(altitude) {
				elevGain, elevLoss := calculateElevation(altitude[splitStart:])
				split.ElevGain = elevGain
				split.ElevLoss = elevLoss
			}

			splits = append(splits, split)
		}
	}

	if len(splits) == 0 {
		return nil
	}

	// Find fastest and slowest splits (by pace), only considering full splits
	fastestIdx, slowestIdx := -1, -1
	for i, s := range splits {
		// Skip partial splits for comparison
		if s.DistanceM < splitLengthM*fullSplitThreshold {
			continue
		}
		if fastestIdx == -1 || s.PaceSecsPerM < splits[fastestIdx].PaceSecsPerM {
			fastestIdx = i
		}
		if slowestIdx == -1 || s.PaceSecsPerM > splits[slowestIdx].PaceSecsPerM {
			slowestIdx = i
		}
	}

	// If no full splits found, use the first split as both fastest and slowest
	if fastestIdx == -1 {
		fastestIdx = 0
		slowestIdx = 0
	}

	return &SplitsResult{
		Splits:       splits,
		SplitLength:  splitLengthM,
		TotalSplits:  len(splits),
		FastestSplit: fastestIdx + 1, // Convert to 1-indexed
		SlowestSplit: slowestIdx + 1,
	}
}

// calculateElevation calculates elevation gain and loss from altitude data.
//
//nolint:gosec // G602: bounds are checked at function entry before loop access
func calculateElevation(altitude []float64) (gain, loss float64) {
	if len(altitude) < 2 {
		return 0, 0
	}

	for i := 1; i < len(altitude); i++ {
		diff := altitude[i] - altitude[i-1]
		if diff > 0 {
			gain += diff
		} else {
			loss -= diff // Make loss positive
		}
	}

	return gain, loss
}

// ZoneDistribution represents time spent in each HR zone.
type ZoneDistribution struct {
	Zone       int     `json:"zone"`       // Zone number (1-5)
	SecondsIn  int     `json:"seconds"`    // Time spent in zone
	Percentage float64 `json:"percentage"` // Percentage of total time
	MinBPM     float64 `json:"min_bpm"`    // Lower bound of zone (BPM)
	MaxBPM     float64 `json:"max_bpm"`    // Upper bound of zone (BPM)
	Label      string  `json:"label"`      // Zone label (e.g., "Recovery", "Aerobic")
}

// ZoneDistributionResult contains HR zone distribution data.
type ZoneDistributionResult struct {
	Zones        []ZoneDistribution `json:"zones"`
	TotalSeconds int                `json:"total_seconds"`
	AvgHR        float64            `json:"avg_hr"`
	MaxHR        float64            `json:"max_hr"`
}

// ZoneBounds defines the upper bounds for each zone (5 zones).
type ZoneBounds struct {
	Method string    // "absolute_bpm" or "percent_hrmax"
	Bounds []float64 // Upper bounds for zones 1-5
	HRMax  float64   // Max HR (for percent method)
}

// ZoneLabels are the standard labels for 5-zone HR models.
var ZoneLabels = []string{"Recovery", "Aerobic", "Tempo", "Threshold", "VO2 Max"}

// ComputeHRZoneDistribution calculates time spent in each HR zone.
func ComputeHRZoneDistribution(hr []float64, bounds *ZoneBounds) *ZoneDistributionResult {
	if len(hr) == 0 || bounds == nil || len(bounds.Bounds) < 5 {
		return nil
	}

	secondsPerZone := make([]int, 5)
	totalSeconds := 0
	var hrSum, maxHR float64

	for _, h := range hr {
		if h <= 0 {
			continue
		}

		totalSeconds++
		hrSum += h
		if h > maxHR {
			maxHR = h
		}

		zoneIdx := classifyHRZone(h, bounds)
		if zoneIdx >= 0 && zoneIdx < 5 {
			secondsPerZone[zoneIdx]++
		}
	}

	if totalSeconds == 0 {
		return nil
	}

	// Build zone distribution with bounds
	zones := make([]ZoneDistribution, 5)
	for i := 0; i < 5; i++ {
		minBPM := 0.0
		if i > 0 {
			minBPM = bounds.Bounds[i-1]
		}

		// Convert to absolute BPM if using percent method
		if bounds.Method == "percent_hrmax" && bounds.HRMax > 0 {
			minBPM *= bounds.HRMax
		}

		maxBPM := bounds.Bounds[i]
		if bounds.Method == "percent_hrmax" && bounds.HRMax > 0 {
			maxBPM *= bounds.HRMax
		}

		zones[i] = ZoneDistribution{
			Zone:       i + 1,
			SecondsIn:  secondsPerZone[i],
			Percentage: float64(secondsPerZone[i]) / float64(totalSeconds) * 100,
			MinBPM:     minBPM,
			MaxBPM:     maxBPM,
			Label:      ZoneLabels[i],
		}
	}

	return &ZoneDistributionResult{
		Zones:        zones,
		TotalSeconds: totalSeconds,
		AvgHR:        hrSum / float64(totalSeconds),
		MaxHR:        maxHR,
	}
}

// classifyHRZone returns the zone index (0-4) for a given HR value.
func classifyHRZone(hr float64, bounds *ZoneBounds) int {
	if bounds == nil || hr <= 0 {
		return -1
	}

	value := hr
	if bounds.Method == "percent_hrmax" && bounds.HRMax > 0 {
		value = hr / bounds.HRMax
	}

	for i, bound := range bounds.Bounds {
		if value <= bound {
			return i
		}
	}

	return len(bounds.Bounds) - 1 // Return highest zone if above all bounds
}

// PaceBucket represents a pace histogram bucket.
type PaceBucket struct {
	MinPace    float64 `json:"min_pace"`   // Min pace in seconds per km
	MaxPace    float64 `json:"max_pace"`   // Max pace in seconds per km
	Count      int     `json:"count"`      // Number of data points in bucket
	Seconds    int     `json:"seconds"`    // Time spent at this pace
	Percentage float64 `json:"percentage"` // Percentage of total time
}

// PaceDistributionResult contains pace histogram data.
type PaceDistributionResult struct {
	Buckets      []PaceBucket `json:"buckets"`
	TotalSeconds int          `json:"total_seconds"`
	AvgPace      float64      `json:"avg_pace"`     // Average pace in sec/km
	FastestPace  float64      `json:"fastest_pace"` // Fastest pace in sec/km
	SlowestPace  float64      `json:"slowest_pace"` // Slowest pace in sec/km
	MedianPace   float64      `json:"median_pace"`  // Median pace in sec/km
}

// ComputePaceDistribution calculates pace distribution from velocity data.
// Parameters:
//   - velocity: velocity array in m/s
//   - bucketSize: bucket width in seconds per km (e.g., 15 for 15s/km buckets)
func ComputePaceDistribution(velocity []float64, bucketSize float64) *PaceDistributionResult {
	if len(velocity) == 0 {
		return nil
	}

	if bucketSize <= 0 {
		bucketSize = 15 // Default to 15 sec/km buckets
	}

	// Convert velocities to pace (sec/km) and find range
	var paces []float64
	var paceSum, minPace, maxPace float64
	minPace = 1e10 // Large initial value

	for _, v := range velocity {
		if v <= 0.5 { // Ignore very slow/stopped movement (< 0.5 m/s)
			continue
		}

		// Convert m/s to sec/km
		pace := 1000 / v

		// Ignore unrealistic paces (faster than 2 min/km or slower than 15 min/km)
		if pace < 120 || pace > 900 {
			continue
		}

		paces = append(paces, pace)
		paceSum += pace

		if pace < minPace {
			minPace = pace
		}
		if pace > maxPace {
			maxPace = pace
		}
	}

	if len(paces) == 0 {
		return nil
	}

	// Create buckets from min to max pace
	bucketStart := float64(int(minPace/bucketSize)) * bucketSize
	bucketEnd := (float64(int(maxPace/bucketSize)) + 1) * bucketSize

	bucketCounts := make(map[int]int)
	for _, pace := range paces {
		bucketIdx := int((pace - bucketStart) / bucketSize)
		bucketCounts[bucketIdx]++
	}

	// Build bucket slice
	numBuckets := int((bucketEnd-bucketStart)/bucketSize) + 1
	buckets := make([]PaceBucket, 0, numBuckets)

	totalSeconds := len(paces)
	for i := 0; i < numBuckets; i++ {
		count := bucketCounts[i]
		if count == 0 {
			continue // Skip empty buckets
		}

		buckets = append(buckets, PaceBucket{
			MinPace:    bucketStart + float64(i)*bucketSize,
			MaxPace:    bucketStart + float64(i+1)*bucketSize,
			Count:      count,
			Seconds:    count, // Assuming 1Hz sampling
			Percentage: float64(count) / float64(totalSeconds) * 100,
		})
	}

	// Calculate median
	sortedPaces := make([]float64, len(paces))
	copy(sortedPaces, paces)
	slices.Sort(sortedPaces)
	medianPace := sortedPaces[len(sortedPaces)/2]

	return &PaceDistributionResult{
		Buckets:      buckets,
		TotalSeconds: totalSeconds,
		AvgPace:      paceSum / float64(len(paces)),
		FastestPace:  minPace,
		SlowestPace:  maxPace,
		MedianPace:   medianPace,
	}
}
