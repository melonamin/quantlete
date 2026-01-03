// Package shared provides common utilities used across multiple packages.
package shared

import (
	"fmt"
	"strings"
)

// CanonicalDistance represents a standard best effort distance.
type CanonicalDistance struct {
	Type string  // Canonical type name (e.g., "5k", "marathon")
	M    float64 // Distance in meters
}

// CanonicalDistances defines the standard best effort distances.
// The tolerance for matching is max(25m, 1% of distance).
var CanonicalDistances = []CanonicalDistance{
	{Type: "400m", M: 400},
	{Type: "0.5mi", M: 804.672},
	{Type: "1k", M: 1000},
	{Type: "1mi", M: 1609.344},
	{Type: "2mi", M: 3218.688},
	{Type: "5k", M: 5000},
	{Type: "10k", M: 10000},
	{Type: "15k", M: 15000},
	{Type: "10mi", M: 16093.44},
	{Type: "20k", M: 20000},
	{Type: "half_marathon", M: 21097.5},
	{Type: "30k", M: 30000},
	{Type: "marathon", M: 42195},
	{Type: "50k", M: 50000},
	{Type: "100k", M: 100000},
}

// CanonicalBestEffortDistanceType maps a distance and name to a canonical
// distance type and meters.
//
// The function first attempts to match the distance against known canonical
// distances (within a tolerance of max(25m, 1%)). If no match is found, it
// falls back to sanitizing the name (converting spaces, slashes, and hyphens
// to underscores). If the name is empty, it generates a type like "m_5000".
//
// Returns:
//   - distanceType: The canonical type string (e.g., "5k", "marathon", "10mi")
//   - canonicalM: The canonical distance in meters (or original if no match)
func CanonicalBestEffortDistanceType(distanceM float64, name string) (distanceType string, canonicalM float64) {
	// Prefer matching by distance (tolerant to minor rounding)
	if distanceM > 0 {
		for _, c := range CanonicalDistances {
			// Accept within 1% or 25m, whichever is larger
			tol := 25.0
			if c.M*0.01 > tol {
				tol = c.M * 0.01
			}
			if distanceM >= c.M-tol && distanceM <= c.M+tol {
				return c.Type, c.M
			}
		}
	}

	// Fallback: sanitize the name
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.ReplaceAll(n, " ", "_")
	n = strings.ReplaceAll(n, "/", "_")
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.Trim(n, "_")
	if n != "" {
		return n, distanceM
	}

	// Last resort: use rounded meters
	return fmt.Sprintf("m_%d", int(distanceM+0.5)), distanceM
}
