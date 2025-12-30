package demo

import (
	"strings"

	"github.com/melonamin/quantlete/internal/storage"
)

// generateAthlete creates a demo athlete profile.
func (g *Generator) generateAthlete(firstName, lastName string) *storage.Athlete {
	// Generate a deterministic ID based on the name
	id := int64(100000 + g.rng.Intn(900000))

	username := strings.ToLower(firstName)
	if lastName != "" {
		username = strings.ToLower(firstName) + "_" + strings.ToLower(lastName)
	}

	// Random weight between 55-95 kg
	weight := g.randBetween(55, 95)

	// Random sex
	sex := "M"
	if g.rng.Float64() < 0.5 {
		sex = "F"
	}

	return &storage.Athlete{
		ID:        id,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		City:      "San Francisco",
		State:     "California",
		Country:   "United States",
		Sex:       sex,
		Premium:   true,
		Summit:    true,
		Weight:    weight,
	}
}
