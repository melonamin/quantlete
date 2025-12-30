package demo

import (
	"math/rand"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// Bike component templates with typical maintenance thresholds.
var bikeComponentTemplates = []struct {
	name               string
	maintenanceHashtag string
	rules              []struct {
		ruleType       string
		thresholdValue float64
	}
}{
	{
		name:               "Chain",
		maintenanceHashtag: "chain",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 5000000}, // 5000km
		},
	},
	{
		name:               "Cassette",
		maintenanceHashtag: "cassette",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 15000000}, // 15000km (every 3 chains)
		},
	},
	{
		name:               "Front Tire",
		maintenanceHashtag: "fronttire",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 5000000}, // 5000km
		},
	},
	{
		name:               "Rear Tire",
		maintenanceHashtag: "reartire",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 4000000}, // 4000km (wears faster)
		},
	},
	{
		name:               "Brake Pads",
		maintenanceHashtag: "brakepads",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 3000000}, // 3000km
		},
	},
	{
		name:               "Bar Tape",
		maintenanceHashtag: "bartape",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"days", 365}, // Yearly
		},
	},
	{
		name:               "Cables",
		maintenanceHashtag: "cables",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 10000000}, // 10000km
			{"days", 730},            // Or 2 years
		},
	},
}

// Shoe component templates.
var shoeComponentTemplates = []struct {
	name               string
	maintenanceHashtag string
	rules              []struct {
		ruleType       string
		thresholdValue float64
	}
}{
	{
		name:               "Midsole",
		maintenanceHashtag: "replace",
		rules: []struct {
			ruleType       string
			thresholdValue float64
		}{
			{"distance_m", 800000}, // 800km typical shoe lifespan
		},
	},
}

// GeneratedComponent holds component data for insertion.
type GeneratedComponent struct {
	GearID             string
	Name               string
	MaintenanceHashtag string
	CreatedAt          time.Time
	Rules              []GeneratedRule
}

// GeneratedRule holds maintenance rule data.
type GeneratedRule struct {
	Type           string
	ThresholdValue float64
}

// GeneratedMaintenanceLog holds maintenance log entry data.
type GeneratedMaintenanceLog struct {
	ComponentIdx int // Index in components slice (will be replaced with real ID after insert)
	CompletedAt  time.Time
}

// generateComponents creates components for all gear items.
func generateComponents(rng *rand.Rand, gear []storage.Gear, startDate time.Time) []GeneratedComponent {
	var components []GeneratedComponent

	for _, g := range gear {
		var templates []struct {
			name               string
			maintenanceHashtag string
			rules              []struct {
				ruleType       string
				thresholdValue float64
			}
		}

		// Select templates based on gear type
		switch {
		case g.Primary && isBikeGear(g.ID):
			templates = bikeComponentTemplates
		case g.Primary && isShoeGear(g.ID):
			templates = shoeComponentTemplates
		case isBikeGear(g.ID):
			// Non-primary bikes get fewer components
			templates = bikeComponentTemplates[:4] // Chain, cassette, tires only
		case isShoeGear(g.ID):
			templates = shoeComponentTemplates
		default:
			continue
		}

		// Create components with slightly randomized creation dates
		for _, tmpl := range templates {
			// Component created sometime between start date and 6 months after
			createdOffset := rng.Intn(180)
			createdAt := startDate.AddDate(0, 0, createdOffset)

			var rules []GeneratedRule
			for _, r := range tmpl.rules {
				// Add some variation to thresholds (+/- 10%)
				variation := 0.9 + rng.Float64()*0.2
				rules = append(rules, GeneratedRule{
					Type:           r.ruleType,
					ThresholdValue: r.thresholdValue * variation,
				})
			}

			components = append(components, GeneratedComponent{
				GearID:             g.ID,
				Name:               tmpl.name,
				MaintenanceHashtag: tmpl.maintenanceHashtag,
				CreatedAt:          createdAt,
				Rules:              rules,
			})
		}
	}

	return components
}

// generateMaintenanceLogs creates historical maintenance records.
// This simulates that the user has been maintaining their gear over time.
func generateMaintenanceLogs(rng *rand.Rand, components []GeneratedComponent, gear []storage.Gear, activities []storage.Activity) []GeneratedMaintenanceLog {
	var logs []GeneratedMaintenanceLog

	// Build gear distance map
	gearDistance := make(map[string]float64)
	gearFirstActivity := make(map[string]time.Time)
	for _, a := range activities {
		if a.GearID == "" {
			continue
		}
		gearDistance[a.GearID] += a.Distance
		if _, ok := gearFirstActivity[a.GearID]; !ok {
			gearFirstActivity[a.GearID] = a.StartDateLocal.Time
		}
	}

	for compIdx, comp := range components {
		totalDistance := gearDistance[comp.GearID]

		for _, rule := range comp.Rules {
			if rule.Type != "distance_m" {
				continue
			}

			// How many times should this component have been replaced?
			numReplacements := int(totalDistance / rule.ThresholdValue)

			// Cap at reasonable number
			if numReplacements > 10 {
				numReplacements = 10
			}

			// Generate replacement dates spread across the time period
			if numReplacements > 0 {
				firstActivity := gearFirstActivity[comp.GearID]
				if firstActivity.IsZero() {
					firstActivity = comp.CreatedAt
				}

				timeRange := time.Now().Sub(firstActivity)
				interval := timeRange / time.Duration(numReplacements+1)

				for i := 1; i <= numReplacements; i++ {
					// Add some randomness to the maintenance date
					jitter := time.Duration(rng.Intn(14)-7) * 24 * time.Hour
					completedAt := firstActivity.Add(interval * time.Duration(i)).Add(jitter)

					// Don't log future maintenance
					if completedAt.After(time.Now()) {
						continue
					}

					logs = append(logs, GeneratedMaintenanceLog{
						ComponentIdx: compIdx,
						CompletedAt:  completedAt,
					})
				}
			}
		}

		// For time-based rules (like bar tape yearly), add periodic maintenance
		for _, rule := range comp.Rules {
			if rule.Type != "days" {
				continue
			}

			daysSinceCreation := int(time.Now().Sub(comp.CreatedAt).Hours() / 24)
			numMaintenance := daysSinceCreation / int(rule.ThresholdValue)

			if numMaintenance > 5 {
				numMaintenance = 5
			}

			for i := 1; i <= numMaintenance; i++ {
				daysOffset := int(rule.ThresholdValue) * i
				jitter := rng.Intn(30) - 15
				completedAt := comp.CreatedAt.AddDate(0, 0, daysOffset+jitter)

				if completedAt.After(time.Now()) {
					continue
				}

				logs = append(logs, GeneratedMaintenanceLog{
					ComponentIdx: compIdx,
					CompletedAt:  completedAt,
				})
			}
		}
	}

	return logs
}

func isBikeGear(gearID string) bool {
	// Gear IDs starting with 'b' are bikes
	return len(gearID) > 0 && gearID[0] == 'b'
}

func isShoeGear(gearID string) bool {
	// Gear IDs starting with 'g' are shoes (from Strava convention)
	return len(gearID) > 0 && gearID[0] == 'g'
}
