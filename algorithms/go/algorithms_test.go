package algorithms

import (
	"math"
	"testing"
)

const epsilon = 1e-6

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// Power Analysis Tests

func TestRollingMaxAverage(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		window   int
		expected float64
	}{
		{"empty", []float64{}, 5, 0},
		{"window too large", []float64{1, 2, 3}, 5, 0},
		{"window zero", []float64{1, 2, 3}, 0, 0},
		{"single window", []float64{100, 200, 300}, 3, 200},
		{"find max", []float64{100, 200, 300, 200, 100}, 3, 233.33333333333334},
		{"peak at start", []float64{300, 200, 100, 50, 50}, 2, 250},
		{"peak at end", []float64{50, 50, 100, 200, 300}, 2, 250},
		{"negative values clamped", []float64{-10, 100, 200}, 2, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RollingMaxAverage(tt.values, tt.window)
			if !almostEqual(got, tt.expected) {
				t.Errorf("RollingMaxAverage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNormalizedPower(t *testing.T) {
	tests := []struct {
		name     string
		watts    []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single value", []float64{200}, 200},
		{"short array (< 30)", make30Values(150), 150}, // Average of constant values
		{"constant power", make100Values(200), 200},    // NP of constant power = that power
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizedPower(tt.watts)
			if !almostEqual(got, tt.expected) {
				t.Errorf("NormalizedPower() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNormalizedPowerVariability(t *testing.T) {
	// NP should be higher than average for variable power with longer intervals
	// Using 60-second intervals (longer than 30s rolling window) to show the effect
	variable := make([]float64, 120)
	for i := range variable {
		if (i/60)%2 == 0 {
			variable[i] = 100
		} else {
			variable[i] = 300
		}
	}
	np := NormalizedPower(variable)
	avg := 200.0 // (100+300)/2

	// With 60-second intervals, the 30s rolling avg will see pure 100 or pure 300
	// so NP = (100^4 + 300^4)^0.25 / 2^0.25 ≈ 224 (higher than avg due to 4th power weighting)
	if np <= avg {
		t.Errorf("NP should be > average for variable power, got NP=%v, avg=%v", np, avg)
	}
}

func TestIntensityFactor(t *testing.T) {
	tests := []struct {
		name     string
		np, ftp  float64
		expected float64
	}{
		{"normal", 200, 250, 0.8},
		{"at threshold", 250, 250, 1.0},
		{"above threshold", 300, 250, 1.2},
		{"zero ftp", 200, 0, 0},
		{"negative ftp", 200, -100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntensityFactor(tt.np, tt.ftp)
			if !almostEqual(got, tt.expected) {
				t.Errorf("IntensityFactor() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrainingStressScore(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		np, ftp  float64
		expected float64
	}{
		{"one hour at threshold", 3600, 250, 250, 100}, // TSS = 100 for 1hr at FTP
		{"zero duration", 0, 250, 250, 0},
		{"zero ftp", 3600, 250, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrainingStressScore(tt.duration, tt.np, tt.ftp)
			if !almostEqual(got, tt.expected) {
				t.Errorf("TrainingStressScore() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Eddington Number Tests

func TestEddingtonNumber(t *testing.T) {
	tests := []struct {
		name      string
		distances []float64
		expected  int
	}{
		{"empty", []float64{}, 0},
		{"single ride", []float64{100}, 1},
		{"E=3", []float64{50, 50, 50}, 3},          // 3 days of 50+ km
		{"E=2", []float64{50, 50, 1}, 2},           // only 2 days of 2+ km
		{"mixed", []float64{10, 20, 30, 40, 50}, 5}, // 5 days of 5+ km each
		{"high E", generateEddingtonData(50), 50},  // exactly E=50
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EddingtonNumber(tt.distances)
			if got != tt.expected {
				t.Errorf("EddingtonNumber() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEddingtonNextSteps(t *testing.T) {
	distances := []float64{50, 50, 50} // E=3
	steps := EddingtonNextSteps(distances, 3, 3)

	if len(steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(steps))
	}

	// To reach E=4: need 4 days of 4+ km, have 3, need 1 more
	if steps[0].Target != 4 || steps[0].DaysNeeded != 1 {
		t.Errorf("Step 1: got target=%d, needed=%d, want target=4, needed=1",
			steps[0].Target, steps[0].DaysNeeded)
	}

	// To reach E=5: need 5 days of 5+ km, have 3, need 2 more
	if steps[1].Target != 5 || steps[1].DaysNeeded != 2 {
		t.Errorf("Step 2: got target=%d, needed=%d, want target=5, needed=2",
			steps[1].Target, steps[1].DaysNeeded)
	}
}

func TestEddingtonHistory(t *testing.T) {
	// Start with no rides, then add progressively
	distances := []float64{10, 20, 30} // After each: E=1, E=2, E=3
	history := EddingtonHistory(distances)

	if len(history) != 3 {
		t.Fatalf("Expected 3 history points, got %d", len(history))
	}

	expected := []int{1, 2, 3}
	for i, e := range expected {
		if history[i] != e {
			t.Errorf("history[%d] = %d, want %d", i, history[i], e)
		}
	}
}

// Training Load Tests

func TestCalculateTrainingLoad(t *testing.T) {
	dailyTss := []float64{100, 100, 100} // 3 days of TSS 100
	result := CalculateTrainingLoad(dailyTss, 42, 7)

	if len(result) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result))
	}

	// After first day with TSS=100, CTL and ATL > 0
	if result[0].CTL <= 0 {
		t.Errorf("CTL should be > 0 after training, got %v", result[0].CTL)
	}
	if result[0].ATL <= 0 {
		t.Errorf("ATL should be > 0 after training, got %v", result[0].ATL)
	}

	// ATL rises faster than CTL (shorter tau)
	if result[0].ATL <= result[0].CTL {
		t.Errorf("ATL should rise faster than CTL, got ATL=%v, CTL=%v",
			result[0].ATL, result[0].CTL)
	}

	// TSB = CTL - ATL, should be negative after fresh training
	if result[0].TSB >= 0 {
		t.Errorf("TSB should be negative after hard training, got %v", result[0].TSB)
	}
}

func TestCalculateTrainingLoadWithInitial(t *testing.T) {
	dailyTss := []float64{0} // Rest day
	result := CalculateTrainingLoadWithInitial(dailyTss, 50, 70, 42, 7)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// After rest day, both should decrease but CTL slower
	if result[0].CTL >= 50 {
		t.Errorf("CTL should decrease on rest day, got %v", result[0].CTL)
	}
	if result[0].ATL >= 70 {
		t.Errorf("ATL should decrease on rest day, got %v", result[0].ATL)
	}

	// TSB should improve (become less negative/more positive)
	initialTsb := 50.0 - 70.0 // -20
	if result[0].TSB <= initialTsb {
		t.Errorf("TSB should improve on rest day, got %v, was %v", result[0].TSB, initialTsb)
	}
}

func TestPredictAfterWorkout(t *testing.T) {
	result := PredictAfterWorkout(50, 30, 100, 42, 7)

	// CTL should increase (was 50, adding TSS 100)
	if result.CTL <= 50 {
		t.Errorf("CTL should increase after workout, got %v", result.CTL)
	}

	// ATL should increase more (faster tau)
	if result.ATL <= 30 {
		t.Errorf("ATL should increase after workout, got %v", result.ATL)
	}

	// TSB = CTL - ATL
	expectedTsb := result.CTL - result.ATL
	if !almostEqual(result.TSB, expectedTsb) {
		t.Errorf("TSB should be CTL-ATL, got %v, expected %v", result.TSB, expectedTsb)
	}
}

func TestTssForTargetTsb(t *testing.T) {
	currentCtl := 50.0
	currentAtl := 30.0
	targetTsb := 25.0 // want to increase TSB from 20 to 25

	tss := TssForTargetTsb(currentCtl, currentAtl, targetTsb, 42, 7)

	// Verify by applying the TSS
	result := PredictAfterWorkout(currentCtl, currentAtl, tss, 42, 7)
	if !almostEqual(result.TSB, targetTsb) {
		t.Errorf("TssForTargetTsb returned %v, but applying it gives TSB=%v, wanted %v",
			tss, result.TSB, targetTsb)
	}
}

func TestTssForTargetTsbEdgeCase(t *testing.T) {
	// Edge case: when ctlTau == atlTau, denominator is 0
	tss := TssForTargetTsb(50, 30, 25, 42, 42)
	if tss != 0 {
		t.Errorf("Expected 0 when taus are equal, got %v", tss)
	}
}

// Helper functions

func make30Values(val float64) []float64 {
	result := make([]float64, 29) // Less than 30 to test fallback
	for i := range result {
		result[i] = val
	}
	return result
}

func make100Values(val float64) []float64 {
	result := make([]float64, 100)
	for i := range result {
		result[i] = val
	}
	return result
}

func generateEddingtonData(targetE int) []float64 {
	// Generate exactly enough data to achieve E=targetE
	// Need targetE days of targetE+ km
	result := make([]float64, targetE)
	for i := range result {
		result[i] = float64(targetE)
	}
	return result
}
