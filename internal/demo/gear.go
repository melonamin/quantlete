package demo

import (
	"fmt"

	"github.com/melonamin/quantlete/internal/storage"
)

// generateGear creates demo gear (bikes and running shoes).
func (g *Generator) generateGear(athleteID int64) []storage.Gear {
	var gear []storage.Gear

	// Bikes
	bikes := []struct {
		name      string
		brand     string
		model     string
		distance  float64
		isPrimary bool
	}{
		{"Road Bike", "Canyon", "Aeroad CF SLX", g.randBetween(5000, 15000) * 1000, true},
		{"Gravel Bike", "Specialized", "Diverge Expert", g.randBetween(1000, 8000) * 1000, false},
		{"MTB", "Santa Cruz", "Tallboy", g.randBetween(500, 3000) * 1000, false},
	}

	for i, b := range bikes {
		gear = append(gear, storage.Gear{
			ID:        fmt.Sprintf("b%d", 1000000+i),
			AthleteID: athleteID,
			Name:      b.name,
			Primary:   b.isPrimary,
			Retired:   false,
			Distance:  b.distance,
			BrandName: b.brand,
			ModelName: b.model,
			Source:    "demo",
		})
	}

	// Running shoes
	shoes := []struct {
		name      string
		brand     string
		model     string
		distance  float64
		isPrimary bool
		retired   bool
	}{
		{"Daily Trainers", "Nike", "Pegasus 41", g.randBetween(200, 600) * 1000, true, false},
		{"Race Shoes", "Nike", "Vaporfly 3", g.randBetween(100, 300) * 1000, false, false},
		{"Trail Shoes", "Salomon", "Speedcross 6", g.randBetween(150, 500) * 1000, false, false},
		{"Old Trainers", "Asics", "Gel-Nimbus 25", g.randBetween(700, 900) * 1000, false, true},
	}

	for i, s := range shoes {
		gear = append(gear, storage.Gear{
			ID:        fmt.Sprintf("g%d", 2000000+i),
			AthleteID: athleteID,
			Name:      s.name,
			Primary:   s.isPrimary,
			Retired:   s.retired,
			Distance:  s.distance,
			BrandName: s.brand,
			ModelName: s.model,
			Source:    "demo",
		})
	}

	return gear
}
