// Package algorithms provides shared computation algorithms for power analysis,
// Eddington number calculation, and training load metrics.
//
// This package is designed to be compiled both as native Go code (for server use)
// and as WebAssembly via TinyGo (for browser use).
package algorithms

import "math"

// Power Analysis Functions

// RollingMaxAverage finds the maximum rolling average over a given window.
// Used for peak power calculations (e.g., 5s peak, 1min peak, 5min peak).
func RollingMaxAverage(values []float64, windowSeconds int) float64 {
	if windowSeconds <= 0 || len(values) < windowSeconds {
		return 0
	}

	// Initialize window sum
	windowSum := 0.0
	for i := 0; i < windowSeconds; i++ {
		v := values[i]
		if v < 0 {
			v = 0
		}
		windowSum += v
	}

	best := windowSum / float64(windowSeconds)

	// Slide window
	for i := windowSeconds; i < len(values); i++ {
		old := values[i-windowSeconds]
		if old < 0 {
			old = 0
		}
		new := values[i]
		if new < 0 {
			new = 0
		}
		windowSum = windowSum - old + new
		avg := windowSum / float64(windowSeconds)
		if avg > best {
			best = avg
		}
	}

	return best
}

// NormalizedPower calculates normalized power from power data.
// Uses 30-second rolling average, then takes the 4th root of the mean of 4th powers.
// Input should be 1 sample per second.
func NormalizedPower(watts []float64) float64 {
	if len(watts) == 0 {
		return 0
	}

	const window = 30
	if len(watts) < window {
		// Not enough data for 30s rolling average, fall back to simple average
		sum := 0.0
		for _, w := range watts {
			if w < 0 {
				w = 0
			}
			sum += w
		}
		return sum / float64(len(watts))
	}

	// Initialize window sum
	windowSum := 0.0
	for i := 0; i < window; i++ {
		w := watts[i]
		if w < 0 {
			w = 0
		}
		windowSum += w
	}

	// Calculate 4th power sum of rolling averages
	fourthPowerSum := 0.0
	avg := windowSum / float64(window)
	fourthPowerSum += avg * avg * avg * avg
	count := 1

	// Slide window
	for i := window; i < len(watts); i++ {
		old := watts[i-window]
		if old < 0 {
			old = 0
		}
		new := watts[i]
		if new < 0 {
			new = 0
		}
		windowSum = windowSum - old + new
		avg = windowSum / float64(window)
		fourthPowerSum += avg * avg * avg * avg
		count++
	}

	return math.Pow(fourthPowerSum/float64(count), 0.25)
}

// IntensityFactor calculates the intensity factor: IF = NP / FTP
func IntensityFactor(normalizedPower, ftp float64) float64 {
	if ftp <= 0 {
		return 0
	}
	return normalizedPower / ftp
}

// TrainingStressScore calculates TSS: (sec * NP * IF) / (FTP * 3600) * 100
func TrainingStressScore(durationSeconds int, normalizedPower, ftp float64) float64 {
	if durationSeconds <= 0 || ftp <= 0 {
		return 0
	}
	ifactor := IntensityFactor(normalizedPower, ftp)
	return (float64(durationSeconds) * normalizedPower * ifactor) / (ftp * 3600.0) * 100.0
}

// Eddington Number Functions

// EddingtonNumber calculates the Eddington number from daily distances.
// The Eddington number E is the largest number such that you have cycled
// at least E km on at least E days.
func EddingtonNumber(distances []float64) int {
	n := len(distances)
	if n == 0 {
		return 0
	}

	// Create sorted copy (descending)
	sorted := make([]float64, n)
	copy(sorted, distances)

	// Insertion sort (descending) - TinyGo compatible
	for i := 1; i < n; i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] < key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	// Find Eddington number
	eddington := 0
	for i := 0; i < n; i++ {
		dayNumber := i + 1
		if sorted[i] >= float64(dayNumber) {
			eddington = dayNumber
		} else {
			break
		}
	}

	return eddington
}

// NextStep represents the days needed to reach a target Eddington number.
type NextStep struct {
	Target     int
	DaysNeeded int
}

// EddingtonNextSteps calculates how many more days of riding are needed
// for each of the next Eddington numbers.
func EddingtonNextSteps(distances []float64, currentE, stepsToCalculate int) []NextStep {
	result := make([]NextStep, stepsToCalculate)

	for step := 0; step < stepsToCalculate; step++ {
		target := currentE + step + 1
		daysWithEnough := 0
		for _, d := range distances {
			if d >= float64(target) {
				daysWithEnough++
			}
		}
		daysNeeded := target - daysWithEnough
		if daysNeeded < 0 {
			daysNeeded = 0
		}
		result[step] = NextStep{Target: target, DaysNeeded: daysNeeded}
	}

	return result
}

// EddingtonHistory calculates progressive Eddington numbers over time.
// Given distances in chronological order, returns E for each day.
// Supports distances up to 1000km to accommodate ultra-distance events.
func EddingtonHistory(distances []float64) []int {
	n := len(distances)
	history := make([]int, n)

	// 1000km covers ultra-distance events (brevets, RAAM stages).
	// Using fixed array for TinyGo/WASM compatibility.
	const maxDistance = 1000
	counts := make([]int, maxDistance+1)
	currentE := 0

	for i := 0; i < n; i++ {
		distKm := int(distances[i])
		if distKm < 0 {
			distKm = 0
		}

		cap := distKm
		if cap > maxDistance {
			cap = maxDistance
		}
		for k := 1; k <= cap; k++ {
			counts[k]++
		}

		for currentE < maxDistance && counts[currentE+1] >= currentE+1 {
			currentE++
		}

		history[i] = currentE
	}

	return history
}

// Training Load Functions

// TrainingLoadPoint represents CTL/ATL/TSB values for a single day.
type TrainingLoadPoint struct {
	CTL float64
	ATL float64
	TSB float64
}

// CalculateTrainingLoad calculates training load metrics from daily TSS values.
// Uses EWMA: new_value = old_value + (tss - old_value) * (1 / tau)
// Default tau values: CTL = 42 days, ATL = 7 days
func CalculateTrainingLoad(dailyTss []float64, ctlTau, atlTau float64) []TrainingLoadPoint {
	n := len(dailyTss)
	result := make([]TrainingLoadPoint, n)

	if n == 0 {
		return result
	}

	ctl := 0.0
	atl := 0.0
	ctlDecay := 1.0 / ctlTau
	atlDecay := 1.0 / atlTau

	for i := 0; i < n; i++ {
		tss := dailyTss[i]
		ctl += (tss - ctl) * ctlDecay
		atl += (tss - atl) * atlDecay
		result[i] = TrainingLoadPoint{
			CTL: ctl,
			ATL: atl,
			TSB: ctl - atl,
		}
	}

	return result
}

// CalculateTrainingLoadWithInitial calculates training load starting from
// existing CTL/ATL values. Useful for continuing from a known state.
func CalculateTrainingLoadWithInitial(dailyTss []float64, initialCtl, initialAtl, ctlTau, atlTau float64) []TrainingLoadPoint {
	n := len(dailyTss)
	result := make([]TrainingLoadPoint, n)

	if n == 0 {
		return result
	}

	ctl := initialCtl
	atl := initialAtl
	ctlDecay := 1.0 / ctlTau
	atlDecay := 1.0 / atlTau

	for i := 0; i < n; i++ {
		tss := dailyTss[i]
		ctl += (tss - ctl) * ctlDecay
		atl += (tss - atl) * atlDecay
		result[i] = TrainingLoadPoint{
			CTL: ctl,
			ATL: atl,
			TSB: ctl - atl,
		}
	}

	return result
}

// PredictAfterWorkout calculates predicted TSB after a planned workout.
func PredictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau float64) TrainingLoadPoint {
	ctlDecay := 1.0 / ctlTau
	atlDecay := 1.0 / atlTau

	newCtl := currentCtl + (plannedTss-currentCtl)*ctlDecay
	newAtl := currentAtl + (plannedTss-currentAtl)*atlDecay

	return TrainingLoadPoint{
		CTL: newCtl,
		ATL: newAtl,
		TSB: newCtl - newAtl,
	}
}

// TssForTargetTsb calculates the TSS needed to reach a target TSB.
// This inverts the EWMA equations to find what TSS would produce
// the desired TSB after one day.
func TssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau float64) float64 {
	ctlDecay := 1.0 / ctlTau
	atlDecay := 1.0 / atlTau

	numerator := targetTsb - currentCtl*(1.0-ctlDecay) + currentAtl*(1.0-atlDecay)
	denominator := ctlDecay - atlDecay

	if math.Abs(denominator) < 1e-10 {
		return 0 // Edge case: ctlTau == atlTau
	}

	return numerator / denominator
}

// Monotony calculates training monotony from daily TSS values.
// Monotony = Mean TSS / StdDev TSS
// High monotony (>2.0) indicates repetitive training which can increase injury risk.
// Returns 0 if there's no variance (stddev is 0).
func Monotony(dailyTSS []float64) float64 {
	n := len(dailyTSS)
	if n == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, tss := range dailyTSS {
		sum += tss
	}
	mean := sum / float64(n)

	// Calculate standard deviation
	sumSqDiff := 0.0
	for _, tss := range dailyTSS {
		diff := tss - mean
		sumSqDiff += diff * diff
	}

	variance := sumSqDiff / float64(n)
	if variance <= 0 {
		return 0 // No variance means infinite monotony, return 0 to indicate undefined
	}

	stdDev := math.Sqrt(variance)
	if stdDev == 0 {
		return 0
	}

	return mean / stdDev
}

// DefaultTRIMPWeights are the standard zone weights for TRIMP calculation.
// Zone 1 = 1, Zone 2 = 2, Zone 3 = 3, Zone 4 = 4, Zone 5 = 5
var DefaultTRIMPWeights = []float64{1, 2, 3, 4, 5}

// ZoneBasedTRIMP calculates Training Impulse from time spent in each HR zone.
// zoneSeconds: array of seconds spent in [Z1, Z2, Z3, Z4, Z5]
// weights: zone multipliers (use DefaultTRIMPWeights if nil)
// TRIMP = Σ(time_in_zone × zone_weight)
func ZoneBasedTRIMP(zoneSeconds []int, weights []float64) float64 {
	if len(zoneSeconds) == 0 {
		return 0
	}

	if weights == nil || len(weights) == 0 {
		weights = DefaultTRIMPWeights
	}

	trimp := 0.0
	for i := 0; i < len(zoneSeconds) && i < len(weights); i++ {
		// Convert seconds to minutes for TRIMP calculation
		minutes := float64(zoneSeconds[i]) / 60.0
		trimp += minutes * weights[i]
	}

	return trimp
}
