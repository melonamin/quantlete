package importer

import (
	"encoding/json"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// convertActivity converts a Strava activity to a storage activity.
func convertActivity(a *strava.Activity) *storage.Activity {
	act := &storage.Activity{
		ID:                 a.ID,
		Name:               a.Name,
		Description:        a.Description,
		SportType:          a.SportType,
		StartDate:          a.StartDate,
		StartDateLocal:     a.StartDateLocal,
		Timezone:           a.Timezone,
		Distance:           a.Distance,
		MovingTime:         a.MovingTime,
		ElapsedTime:        a.ElapsedTime,
		TotalElevationGain: a.TotalElevationGain,
		AverageSpeed:       a.AverageSpeed,
		MaxSpeed:           a.MaxSpeed,
		KudosCount:         a.KudosCount,
		CommentCount:       a.CommentCount,
		PhotoCount:         a.PhotoCount,
		Commute:            a.Commute,
		Private:            a.Private,
		Trainer:            a.Trainer,
		DeviceName:         a.DeviceName,
		GearID:             a.GearID,
		SummaryPolyline:    a.Map.SummaryPolyline,
		Polyline:           a.Map.Polyline,
	}

	// Handle optional numeric fields
	if a.ElevHigh > 0 {
		act.ElevHigh = &a.ElevHigh
	}
	if a.ElevLow > 0 {
		act.ElevLow = &a.ElevLow
	}
	if a.AverageHeartrate > 0 {
		act.AverageHeartrate = &a.AverageHeartrate
	}
	if a.MaxHeartrate > 0 {
		act.MaxHeartrate = &a.MaxHeartrate
	}
	if a.AverageWatts > 0 {
		act.AverageWatts = &a.AverageWatts
	}
	if a.MaxWatts > 0 {
		act.MaxWatts = &a.MaxWatts
	}
	if a.WeightedAverageWatts > 0 {
		act.WeightedAverageWatts = &a.WeightedAverageWatts
	}
	if a.Kilojoules > 0 {
		act.Kilojoules = &a.Kilojoules
	}
	if a.AverageCadence > 0 {
		act.AverageCadence = &a.AverageCadence
	}
	if a.Calories > 0 {
		act.Calories = &a.Calories
	}
	if a.WorkoutType > 0 {
		act.WorkoutType = &a.WorkoutType
	}

	// Handle location
	if len(a.StartLatlng) >= 2 {
		act.StartLat = &a.StartLatlng[0]
		act.StartLng = &a.StartLatlng[1]
	}
	if len(a.EndLatlng) >= 2 {
		act.EndLat = &a.EndLatlng[0]
		act.EndLng = &a.EndLatlng[1]
	}

	return act
}

// convertAthlete converts a Strava athlete to a storage athlete.
func convertAthlete(a *strava.Athlete) *storage.Athlete {
	return &storage.Athlete{
		ID:            a.ID,
		Username:      a.Username,
		FirstName:     a.FirstName,
		LastName:      a.LastName,
		City:          a.City,
		State:         a.State,
		Country:       a.Country,
		Sex:           a.Sex,
		Premium:       a.Premium,
		Summit:        a.Summit,
		ProfileMedium: a.ProfileMedium,
		Profile:       a.Profile,
		Weight:        a.Weight,
	}
}

// encodeStreamData encodes stream data as JSON.
func encodeStreamData(data []any) (json.RawMessage, error) {
	return json.Marshal(data)
}
