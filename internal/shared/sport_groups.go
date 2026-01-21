package shared

// SportGroup represents a named group of related sport types.
type SportGroup struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	SportTypes []string `json:"sport_types"`
}

// PredefinedSportGroups returns the list of predefined sport groups for Eddington calculations.
// These groups aggregate related Strava activity types into meaningful categories.
func PredefinedSportGroups() []SportGroup {
	return []SportGroup{
		{
			ID:   "cycling",
			Name: "Cycling",
			SportTypes: []string{
				"Ride",
				"MountainBikeRide",
				"GravelRide",
				"EBikeRide",
				"VirtualRide",
				"Velomobile",
				"Handcycle",
			},
		},
		{
			ID:   "running",
			Name: "Running",
			SportTypes: []string{
				"Run",
				"TrailRun",
				"VirtualRun",
			},
		},
		{
			ID:   "swimming",
			Name: "Swimming",
			SportTypes: []string{
				"Swim",
				"OpenWaterSwim",
			},
		},
		{
			ID:   "walking",
			Name: "Walking",
			SportTypes: []string{
				"Walk",
				"Hike",
			},
		},
		{
			ID:   "winter",
			Name: "Winter Sports",
			SportTypes: []string{
				"AlpineSki",
				"BackcountrySki",
				"NordicSki",
				"Snowboard",
				"Snowshoe",
				"IceSkate",
			},
		},
		{
			ID:   "water",
			Name: "Water Sports",
			SportTypes: []string{
				"Rowing",
				"Kayaking",
				"Canoeing",
				"StandUpPaddling",
				"Surfing",
				"Kitesurf",
				"Windsurf",
				"Sail",
			},
		},
	}
}

// SportGroupByID returns the sport group with the given ID, or nil if not found.
func SportGroupByID(id string) *SportGroup {
	for _, g := range PredefinedSportGroups() {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

// AllSportGroupIDs returns all predefined sport group IDs.
func AllSportGroupIDs() []string {
	groups := PredefinedSportGroups()
	ids := make([]string, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	return ids
}
