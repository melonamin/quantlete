package analysis

import (
	"math"
	"testing"
)

func TestComputeSplits(t *testing.T) {
	t.Run("empty distance returns nil", func(t *testing.T) {
		result := ComputeSplits(nil, nil, nil, nil, nil, 1000)
		if result != nil {
			t.Error("expected nil for empty distance")
		}
	})

	t.Run("single point returns nil", func(t *testing.T) {
		result := ComputeSplits([]float64{0}, []float64{0}, nil, nil, nil, 1000)
		if result != nil {
			t.Error("expected nil for single point")
		}
	})

	t.Run("mismatched arrays returns nil", func(t *testing.T) {
		result := ComputeSplits([]float64{0, 1000}, []float64{0}, nil, nil, nil, 1000)
		if result != nil {
			t.Error("expected nil for mismatched arrays")
		}
	})

	t.Run("basic 2km run with 1km splits", func(t *testing.T) {
		// Simulate a 2km run at 5:00/km pace (300 sec/km)
		distance := []float64{0, 500, 1000, 1500, 2000}
		time := []float64{0, 150, 300, 450, 600}

		result := ComputeSplits(distance, time, nil, nil, nil, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalSplits != 2 {
			t.Errorf("expected 2 splits, got %d", result.TotalSplits)
		}

		if result.SplitLength != 1000 {
			t.Errorf("expected split length 1000, got %f", result.SplitLength)
		}

		// Both splits should be ~300 seconds
		for i, split := range result.Splits {
			if split.DurationS < 295 || split.DurationS > 305 {
				t.Errorf("split %d: expected ~300 duration, got %d", i+1, split.DurationS)
			}
		}
	})

	t.Run("with heart rate data", func(t *testing.T) {
		distance := []float64{0, 500, 1000}
		time := []float64{0, 150, 300}
		hr := []float64{140, 150, 160}

		result := ComputeSplits(distance, time, hr, nil, nil, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Should have average HR
		if result.Splits[0].AvgHR < 140 || result.Splits[0].AvgHR > 160 {
			t.Errorf("expected avg HR between 140-160, got %f", result.Splits[0].AvgHR)
		}
	})

	t.Run("with elevation data", func(t *testing.T) {
		distance := []float64{0, 500, 1000}
		time := []float64{0, 150, 300}
		altitude := []float64{100, 120, 110}

		result := ComputeSplits(distance, time, nil, nil, altitude, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Gain should be 20m (100->120), loss should be 10m (120->110)
		if result.Splits[0].ElevGain < 19 || result.Splits[0].ElevGain > 21 {
			t.Errorf("expected ~20m gain, got %f", result.Splits[0].ElevGain)
		}
		if result.Splits[0].ElevLoss < 9 || result.Splits[0].ElevLoss > 11 {
			t.Errorf("expected ~10m loss, got %f", result.Splits[0].ElevLoss)
		}
	})

	t.Run("fastest and slowest identification", func(t *testing.T) {
		// First km at 5:00, second km at 4:30, third km at 5:30
		distance := []float64{0, 1000, 2000, 3000}
		time := []float64{0, 300, 570, 900}

		result := ComputeSplits(distance, time, nil, nil, nil, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.FastestSplit != 2 {
			t.Errorf("expected fastest split 2, got %d", result.FastestSplit)
		}
		if result.SlowestSplit != 3 {
			t.Errorf("expected slowest split 3, got %d", result.SlowestSplit)
		}
	})

	t.Run("default split length when zero", func(t *testing.T) {
		distance := []float64{0, 500, 1000}
		time := []float64{0, 150, 300}

		result := ComputeSplits(distance, time, nil, nil, nil, 0)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.SplitLength != 1000 {
			t.Errorf("expected default split length 1000, got %f", result.SplitLength)
		}
	})

	t.Run("partial split under 10% excluded", func(t *testing.T) {
		// 1050m run - 50m partial should be excluded (under 10%)
		distance := []float64{0, 500, 1000, 1050}
		time := []float64{0, 150, 300, 315}

		result := ComputeSplits(distance, time, nil, nil, nil, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalSplits != 1 {
			t.Errorf("expected 1 split (partial excluded), got %d", result.TotalSplits)
		}
	})

	t.Run("partial split over 10% included", func(t *testing.T) {
		// 1500m run - 500m partial should be included (over 10%)
		distance := []float64{0, 500, 1000, 1500}
		time := []float64{0, 150, 300, 450}

		result := ComputeSplits(distance, time, nil, nil, nil, 1000)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalSplits != 2 {
			t.Errorf("expected 2 splits (partial included), got %d", result.TotalSplits)
		}
	})
}

func TestCalculateElevation(t *testing.T) {
	t.Run("empty returns zero", func(t *testing.T) {
		gain, loss := calculateElevation(nil)
		if gain != 0 || loss != 0 {
			t.Errorf("expected (0, 0), got (%f, %f)", gain, loss)
		}
	})

	t.Run("single point returns zero", func(t *testing.T) {
		gain, loss := calculateElevation([]float64{100})
		if gain != 0 || loss != 0 {
			t.Errorf("expected (0, 0), got (%f, %f)", gain, loss)
		}
	})

	t.Run("all gain", func(t *testing.T) {
		gain, loss := calculateElevation([]float64{100, 150, 200})
		if gain != 100 {
			t.Errorf("expected gain 100, got %f", gain)
		}
		if loss != 0 {
			t.Errorf("expected loss 0, got %f", loss)
		}
	})

	t.Run("all loss", func(t *testing.T) {
		gain, loss := calculateElevation([]float64{200, 150, 100})
		if gain != 0 {
			t.Errorf("expected gain 0, got %f", gain)
		}
		if loss != 100 {
			t.Errorf("expected loss 100, got %f", loss)
		}
	})

	t.Run("mixed gain and loss", func(t *testing.T) {
		gain, loss := calculateElevation([]float64{100, 150, 120, 180})
		// Gain: 100->150 (50) + 120->180 (60) = 110
		// Loss: 150->120 (30) = 30
		if gain != 110 {
			t.Errorf("expected gain 110, got %f", gain)
		}
		if loss != 30 {
			t.Errorf("expected loss 30, got %f", loss)
		}
	})
}

func TestComputeHRZoneDistribution(t *testing.T) {
	standardBounds := &ZoneBounds{
		Method: "absolute_bpm",
		Bounds: []float64{120, 140, 160, 175, 200},
	}

	t.Run("empty HR returns nil", func(t *testing.T) {
		result := ComputeHRZoneDistribution(nil, standardBounds)
		if result != nil {
			t.Error("expected nil for empty HR")
		}
	})

	t.Run("nil bounds returns nil", func(t *testing.T) {
		result := ComputeHRZoneDistribution([]float64{150}, nil)
		if result != nil {
			t.Error("expected nil for nil bounds")
		}
	})

	t.Run("insufficient bounds returns nil", func(t *testing.T) {
		result := ComputeHRZoneDistribution([]float64{150}, &ZoneBounds{
			Method: "absolute_bpm",
			Bounds: []float64{120, 140}, // Only 2 bounds, need 5
		})
		if result != nil {
			t.Error("expected nil for insufficient bounds")
		}
	})

	t.Run("all HR values zero returns nil", func(t *testing.T) {
		result := ComputeHRZoneDistribution([]float64{0, 0, 0}, standardBounds)
		if result != nil {
			t.Error("expected nil when all HR values are zero")
		}
	})

	t.Run("basic zone distribution", func(t *testing.T) {
		// 10 data points spread across zones
		hr := []float64{110, 115, 130, 135, 150, 155, 165, 170, 180, 190}

		result := ComputeHRZoneDistribution(hr, standardBounds)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalSeconds != 10 {
			t.Errorf("expected 10 total seconds, got %d", result.TotalSeconds)
		}

		if len(result.Zones) != 5 {
			t.Errorf("expected 5 zones, got %d", len(result.Zones))
		}

		// Verify average HR
		expectedAvg := (110.0 + 115 + 130 + 135 + 150 + 155 + 165 + 170 + 180 + 190) / 10
		if math.Abs(result.AvgHR-expectedAvg) > 0.01 {
			t.Errorf("expected avg HR %f, got %f", expectedAvg, result.AvgHR)
		}

		// Verify max HR
		if result.MaxHR != 190 {
			t.Errorf("expected max HR 190, got %f", result.MaxHR)
		}
	})

	t.Run("percent_hrmax method", func(t *testing.T) {
		percentBounds := &ZoneBounds{
			Method: "percent_hrmax",
			Bounds: []float64{0.6, 0.7, 0.8, 0.9, 1.0},
			HRMax:  200,
		}

		// HR values that should fall into different percentage zones
		hr := []float64{100, 130, 150, 175, 195}

		result := ComputeHRZoneDistribution(hr, percentBounds)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Zone bounds should be converted to absolute BPM
		// Zone 1 max: 0.6 * 200 = 120
		// Zone 2 max: 0.7 * 200 = 140
		// etc.
		if result.Zones[0].MaxBPM != 120 {
			t.Errorf("expected zone 1 max BPM 120, got %f", result.Zones[0].MaxBPM)
		}
	})

	t.Run("zone labels assigned", func(t *testing.T) {
		hr := []float64{150}
		result := ComputeHRZoneDistribution(hr, standardBounds)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		expectedLabels := []string{"Recovery", "Aerobic", "Tempo", "Threshold", "VO2 Max"}
		for i, zone := range result.Zones {
			if zone.Label != expectedLabels[i] {
				t.Errorf("zone %d: expected label %q, got %q", i+1, expectedLabels[i], zone.Label)
			}
		}
	})
}

func TestClassifyHRZone(t *testing.T) {
	bounds := &ZoneBounds{
		Method: "absolute_bpm",
		Bounds: []float64{120, 140, 160, 175, 200},
	}

	tests := []struct {
		hr       float64
		expected int
	}{
		{0, -1},   // Invalid HR
		{-10, -1}, // Negative HR
		{100, 0},  // Zone 1 (<=120)
		{120, 0},  // Zone 1 boundary
		{121, 1},  // Zone 2 (121-140)
		{140, 1},  // Zone 2 boundary
		{141, 2},  // Zone 3 (141-160)
		{160, 2},  // Zone 3 boundary
		{175, 3},  // Zone 4 boundary
		{190, 4},  // Zone 5 (176-200)
		{210, 4},  // Above all bounds -> highest zone
	}

	for _, tt := range tests {
		result := classifyHRZone(tt.hr, bounds)
		if result != tt.expected {
			t.Errorf("classifyHRZone(%f) = %d, want %d", tt.hr, result, tt.expected)
		}
	}

	t.Run("nil bounds returns -1", func(t *testing.T) {
		if classifyHRZone(150, nil) != -1 {
			t.Error("expected -1 for nil bounds")
		}
	})
}

func TestComputePaceDistribution(t *testing.T) {
	t.Run("empty velocity returns nil", func(t *testing.T) {
		result := ComputePaceDistribution(nil, 15)
		if result != nil {
			t.Error("expected nil for empty velocity")
		}
	})

	t.Run("all stopped returns nil", func(t *testing.T) {
		result := ComputePaceDistribution([]float64{0, 0.1, 0.2}, 15)
		if result != nil {
			t.Error("expected nil when all velocity below threshold")
		}
	})

	t.Run("all unrealistic pace returns nil", func(t *testing.T) {
		// Velocity that produces pace outside 2-15 min/km range
		result := ComputePaceDistribution([]float64{0.5}, 15) // ~33 min/km, too slow
		if result != nil {
			t.Error("expected nil for unrealistic pace")
		}
	})

	t.Run("basic pace distribution", func(t *testing.T) {
		// Velocities at 5:00/km pace (3.33 m/s) and 6:00/km pace (2.78 m/s)
		velocity := []float64{3.33, 3.33, 3.33, 2.78, 2.78}

		result := ComputePaceDistribution(velocity, 15)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalSeconds != 5 {
			t.Errorf("expected 5 total seconds, got %d", result.TotalSeconds)
		}

		// Fastest pace should be around 300 sec/km (5:00)
		if result.FastestPace < 295 || result.FastestPace > 305 {
			t.Errorf("expected fastest pace ~300, got %f", result.FastestPace)
		}

		// Slowest pace should be around 360 sec/km (6:00)
		if result.SlowestPace < 355 || result.SlowestPace > 365 {
			t.Errorf("expected slowest pace ~360, got %f", result.SlowestPace)
		}
	})

	t.Run("default bucket size when zero", func(t *testing.T) {
		velocity := []float64{3.33}
		result := ComputePaceDistribution(velocity, 0)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Should use default 15 sec/km buckets
		if len(result.Buckets) == 0 {
			t.Error("expected at least one bucket")
		}
	})

	t.Run("median calculation", func(t *testing.T) {
		// 5 velocities that give paces 300, 310, 320, 330, 340 sec/km
		velocity := []float64{
			1000.0 / 300, // 3.33 m/s
			1000.0 / 310,
			1000.0 / 320, // median should be ~320
			1000.0 / 330,
			1000.0 / 340,
		}

		result := ComputePaceDistribution(velocity, 15)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Median should be around 320 (middle value)
		if result.MedianPace < 315 || result.MedianPace > 325 {
			t.Errorf("expected median pace ~320, got %f", result.MedianPace)
		}
	})

	t.Run("bucket percentages sum to 100", func(t *testing.T) {
		velocity := []float64{3.33, 3.33, 3.33, 2.78, 2.78}

		result := ComputePaceDistribution(velocity, 15)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		var totalPercentage float64
		for _, bucket := range result.Buckets {
			totalPercentage += bucket.Percentage
		}

		if math.Abs(totalPercentage-100) > 0.01 {
			t.Errorf("expected percentages to sum to 100, got %f", totalPercentage)
		}
	})
}
