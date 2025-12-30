package badges

import (
	"fmt"
)

const (
	metersPerKm   = 1000.0
	metersPerMile = 1609.34
	metersPerFoot = 0.3048
)

// FormatDistance formats a distance in meters to a human-readable string.
func FormatDistance(meters float64, unit UnitSystem) string {
	if unit == UnitImperial {
		miles := meters / metersPerMile
		return fmt.Sprintf("%.1f mi", miles)
	}
	km := meters / metersPerKm
	return fmt.Sprintf("%.1f km", km)
}

// FormatDuration formats a duration in seconds to a human-readable string.
func FormatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// FormatElevation formats elevation in meters to a human-readable string.
func FormatElevation(meters float64, unit UnitSystem) string {
	if unit == UnitImperial {
		feet := meters / metersPerFoot
		return FormatNumber(int(feet)) + " ft"
	}
	return FormatNumber(int(meters)) + " m"
}

// FormatNumber formats a number with thousands separators.
func FormatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Handle negative numbers
	negative := n < 0
	if negative {
		n = -n
	}

	// Build result from right to left
	result := ""
	for i := 0; n > 0; i++ {
		if i > 0 && i%3 == 0 {
			result = "," + result
		}
		result = fmt.Sprintf("%d", n%10) + result
		n /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}
