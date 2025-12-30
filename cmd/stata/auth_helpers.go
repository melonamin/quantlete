package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"golang.org/x/oauth2"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// restoreAuth restores authentication from the database on startup.
func restoreAuth(
	ctx context.Context,
	stravaClient *strava.Client,
	tokenRepo *storage.TokenRepository,
	athleteRepo *storage.AthleteRepository,
) error {
	// Get active tokens (not expired)
	tokens, err := tokenRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("getting active tokens: %w", err)
	}

	slog.Info("checking for stored tokens", "active_count", len(tokens))

	if len(tokens) == 0 {
		slog.Info("no active tokens found in database")
		return nil
	}

	// Use the first active token (single-user app)
	storedToken := tokens[0]

	// Get the athlete
	storedAthlete, err := athleteRepo.GetByID(ctx, storedToken.AthleteID)
	if err != nil {
		return fmt.Errorf("getting athlete: %w", err)
	}
	if storedAthlete == nil {
		return fmt.Errorf("athlete not found for token")
	}

	// Convert to oauth2.Token
	token := &oauth2.Token{
		AccessToken:  storedToken.AccessToken,
		RefreshToken: storedToken.RefreshToken,
		TokenType:    storedToken.TokenType,
		Expiry:       storedToken.ExpiresAt.Time,
	}

	// Convert to strava.Athlete
	athlete := &strava.Athlete{
		ID:            storedAthlete.ID,
		Username:      storedAthlete.Username,
		FirstName:     storedAthlete.FirstName,
		LastName:      storedAthlete.LastName,
		City:          storedAthlete.City,
		State:         storedAthlete.State,
		Country:       storedAthlete.Country,
		Sex:           storedAthlete.Sex,
		Premium:       storedAthlete.Premium,
		Summit:        storedAthlete.Summit,
		ProfileMedium: storedAthlete.ProfileMedium,
		Profile:       storedAthlete.Profile,
		Weight:        storedAthlete.Weight,
	}

	stravaClient.SetToken(token, athlete)
	slog.Info("restored auth from database", "athlete_id", athlete.ID, "expires_at", token.Expiry)

	return nil
}

// loadDemoAthlete loads the demo athlete and sets it in the Strava client.
// This allows the dashboard to work without Strava OAuth.
func loadDemoAthlete(
	ctx context.Context,
	stravaClient *strava.Client,
	appStateRepo *storage.AppStateRepository,
	athleteRepo *storage.AthleteRepository,
) error {
	// Get the demo athlete ID
	athleteIDStr, err := appStateRepo.Get(ctx, storage.AppStateDemoAthleteID)
	if err != nil {
		return fmt.Errorf("getting demo athlete id: %w", err)
	}
	if athleteIDStr == "" {
		return fmt.Errorf("%s not found in app_state", storage.AppStateDemoAthleteID)
	}

	athleteID, parseErr := strconv.ParseInt(athleteIDStr, 10, 64)
	if parseErr != nil {
		return fmt.Errorf("parsing demo athlete id %q: %w", athleteIDStr, parseErr)
	}

	// Load the athlete
	storedAthlete, err := athleteRepo.GetByID(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting demo athlete: %w", err)
	}
	if storedAthlete == nil {
		return fmt.Errorf("demo athlete not found in database")
	}

	// Convert to strava.Athlete
	athlete := &strava.Athlete{
		ID:            storedAthlete.ID,
		Username:      storedAthlete.Username,
		FirstName:     storedAthlete.FirstName,
		LastName:      storedAthlete.LastName,
		City:          storedAthlete.City,
		State:         storedAthlete.State,
		Country:       storedAthlete.Country,
		Sex:           storedAthlete.Sex,
		Premium:       storedAthlete.Premium,
		Summit:        storedAthlete.Summit,
		ProfileMedium: storedAthlete.ProfileMedium,
		Profile:       storedAthlete.Profile,
		Weight:        storedAthlete.Weight,
	}

	// Set the athlete with a nil token (no API calls will be made in demo mode)
	stravaClient.SetToken(nil, athlete)
	slog.Info("loaded demo athlete", "athlete_id", athlete.ID, "name", athlete.FirstName+" "+athlete.LastName)

	return nil
}

// restoreRateLimitState restores rate limit state from the database on startup.
func restoreRateLimitState(
	ctx context.Context,
	stravaClient *strava.Client,
	appStateRepo *storage.AppStateRepository,
) error {
	stateJSON, err := appStateRepo.Get(ctx, storage.AppStateStravaRateLimit)
	if err != nil {
		return fmt.Errorf("getting rate limit state: %w", err)
	}

	if stateJSON == "" {
		slog.Debug("no stored rate limit state found")
		return nil
	}

	if err := stravaClient.RateLimiter().LoadFromJSON(stateJSON); err != nil {
		return fmt.Errorf("loading rate limit state: %w", err)
	}

	status := stravaClient.RateLimiter().Status()
	slog.Info("restored rate limit state",
		"usage_15min", status.Usage15Min,
		"limit_15min", status.Limit15Min,
		"usage_daily", status.UsageDaily,
		"limit_daily", status.LimitDaily,
	)

	return nil
}
