package demo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// Challenge templates - real Strava challenges with realistic badge URLs.
var challengeTemplates = []struct {
	name     string
	slug     string
	badgeURL string
	monthly  bool // true if it's a monthly challenge
}{
	// Monthly challenges
	{"January Running Challenge", "january-running-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/january-running-challenge.png", true},
	{"February Cycling Challenge", "february-cycling-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/february-cycling-challenge.png", true},
	{"March Distance Challenge", "march-distance-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/march-distance-challenge.png", true},
	{"April Run Challenge", "april-run-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/april-run-challenge.png", true},
	{"May Movement Challenge", "may-movement-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/may-movement-challenge.png", true},
	{"June Cycling Challenge", "june-cycling-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/june-cycling-challenge.png", true},
	{"July Gran Fondo", "july-gran-fondo", "https://dgalywyr863hv.cloudfront.net/challenges/badges/july-gran-fondo.png", true},
	{"August Running Challenge", "august-running-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/august-running-challenge.png", true},
	{"September Cycling Challenge", "september-cycling-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/september-cycling-challenge.png", true},
	{"October Running Streak", "october-running-streak", "https://dgalywyr863hv.cloudfront.net/challenges/badges/october-running-streak.png", true},
	{"November Distance Challenge", "november-distance-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/november-distance-challenge.png", true},
	{"December Cycling Challenge", "december-cycling-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/december-cycling-challenge.png", true},

	// Yearly/brand challenges (non-monthly)
	{"100K in a Week", "100k-week-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/100k-week.png", false},
	{"Gran Fondo", "gran-fondo", "https://dgalywyr863hv.cloudfront.net/challenges/badges/gran-fondo.png", false},
	{"Marathon Month", "marathon-month", "https://dgalywyr863hv.cloudfront.net/challenges/badges/marathon-month.png", false},
	{"Climb Challenge", "climb-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/climb-challenge.png", false},
	{"Century Ride", "century-ride", "https://dgalywyr863hv.cloudfront.net/challenges/badges/century-ride.png", false},
	{"Half Marathon Badge", "half-marathon", "https://dgalywyr863hv.cloudfront.net/challenges/badges/half-marathon.png", false},
	{"50K Trail Challenge", "50k-trail", "https://dgalywyr863hv.cloudfront.net/challenges/badges/50k-trail.png", false},
	{"Gravel Challenge", "gravel-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/gravel-challenge.png", false},
	{"Indoor Challenge", "indoor-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/indoor-challenge.png", false},
	{"Zwift Challenge", "zwift-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/zwift-challenge.png", false},
	{"Run Streak", "run-streak", "https://dgalywyr863hv.cloudfront.net/challenges/badges/run-streak.png", false},
	{"Bike to Work", "bike-to-work", "https://dgalywyr863hv.cloudfront.net/challenges/badges/bike-to-work.png", false},
	{"Le Col Challenge", "le-col", "https://dgalywyr863hv.cloudfront.net/challenges/badges/le-col.png", false},
	{"Rapha Festive 500", "rapha-festive-500", "https://dgalywyr863hv.cloudfront.net/challenges/badges/rapha-festive-500.png", false},
	{"Red Bull Rise Up", "redbull-rise-up", "https://dgalywyr863hv.cloudfront.net/challenges/badges/redbull-rise-up.png", false},
	{"Endurance Challenge", "endurance-challenge", "https://dgalywyr863hv.cloudfront.net/challenges/badges/endurance-challenge.png", false},
	{"Power Hour", "power-hour", "https://dgalywyr863hv.cloudfront.net/challenges/badges/power-hour.png", false},
}

// generateChallenges creates sample challenge/badge data.
func generateChallenges(rng *rand.Rand, athleteID int64, months int) []storage.Challenge {
	var challenges []storage.Challenge

	now := time.Now()
	startDate := now.AddDate(0, -months, 0)

	// Generate monthly challenges (one per month, 60% completion rate)
	for m := 0; m < months; m++ {
		monthDate := startDate.AddDate(0, m, 0)
		monthNum := int(monthDate.Month()) - 1 // 0-indexed

		// Find the monthly challenge for this month
		for _, tmpl := range challengeTemplates {
			if !tmpl.monthly {
				continue
			}

			// Match challenge to month by name prefix
			monthNames := []string{"january", "february", "march", "april", "may", "june",
				"july", "august", "september", "october", "november", "december"}
			matchMonth := -1
			for i, name := range monthNames {
				if len(tmpl.slug) >= len(name) && tmpl.slug[:len(name)] == name {
					matchMonth = i
					break
				}
			}

			if matchMonth != monthNum {
				continue
			}

			// 60% chance to complete monthly challenges
			if rng.Float64() > 0.60 {
				continue
			}

			// Completion date is random day in the month
			completionDate := monthDate.AddDate(0, 0, rng.Intn(28))
			monthStr := monthDate.Format("2006-01")

			challenges = append(challenges, storage.Challenge{
				ID:             fmt.Sprintf("%s-%d", tmpl.slug, monthDate.Year()),
				AthleteID:      athleteID,
				Name:           fmt.Sprintf("%s %d", tmpl.name, monthDate.Year()),
				Slug:           tmpl.slug,
				BadgeURL:       tmpl.badgeURL,
				CompletionDate: &storage.SQLiteTime{Time: completionDate},
				Month:          monthStr,
			})
			break
		}
	}

	// Generate some non-monthly challenges (10-20 over the time period)
	numBrandChallenges := 10 + rng.Intn(11)
	brandChallenges := make([]struct {
		name     string
		slug     string
		badgeURL string
		monthly  bool
	}, 0)

	for _, t := range challengeTemplates {
		if !t.monthly {
			brandChallenges = append(brandChallenges, t)
		}
	}

	for i := 0; i < numBrandChallenges && len(brandChallenges) > 0; i++ {
		idx := rng.Intn(len(brandChallenges))
		tmpl := brandChallenges[idx]

		// Random completion date within the time period
		daysRange := int(now.Sub(startDate).Hours() / 24)
		if daysRange < 1 {
			daysRange = 1
		}
		completionDate := startDate.AddDate(0, 0, rng.Intn(daysRange))
		monthStr := completionDate.Format("2006-01")
		year := completionDate.Year()

		challenges = append(challenges, storage.Challenge{
			ID:             fmt.Sprintf("%s-%d-%d", tmpl.slug, year, i),
			AthleteID:      athleteID,
			Name:           tmpl.name,
			Slug:           tmpl.slug,
			BadgeURL:       tmpl.badgeURL,
			CompletionDate: &storage.SQLiteTime{Time: completionDate},
			Month:          monthStr,
		})

		// Remove used template to avoid duplicates
		brandChallenges = append(brandChallenges[:idx], brandChallenges[idx+1:]...)
	}

	return challenges
}
