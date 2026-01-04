package importer

import (
	"encoding/json"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// ConvertActivityToStorage converts importer.Activity to storage.Activity.
func ConvertActivityToStorage(athleteID int64, a *Activity) *storage.Activity {
	act := &storage.Activity{
		ID:                   a.ID,
		AthleteID:            athleteID,
		Name:                 a.Name,
		Description:          a.Description,
		SportType:            a.SportType,
		StartDate:            storage.SQLiteTime{Time: a.StartDate},
		StartDateLocal:       storage.SQLiteTime{Time: a.StartDateLocal},
		Timezone:             a.Timezone,
		LocationCity:         a.LocationCity,
		LocationState:        a.LocationState,
		LocationCountry:      a.LocationCountry,
		Distance:             a.Distance,
		MovingTime:           a.MovingTime,
		ElapsedTime:          a.ElapsedTime,
		TotalElevationGain:   a.TotalElevationGain,
		AverageSpeed:         a.AverageSpeed,
		MaxSpeed:             a.MaxSpeed,
		AverageHeartrate:     a.AverageHeartrate,
		MaxHeartrate:         a.MaxHeartrate,
		AverageWatts:         a.AverageWatts,
		MaxWatts:             a.MaxWatts,
		WeightedAverageWatts: a.WeightedAverageWatts,
		Kilojoules:           a.Kilojoules,
		AverageCadence:       a.AverageCadence,
		Calories:             a.Calories,
		KudosCount:           a.KudosCount,
		PhotoCount:           a.PhotoCount,
		Commute:              a.Commute,
		Private:              a.Private,
		Trainer:              a.Trainer,
		WorkoutType:          a.WorkoutType,
		DeviceName:           a.DeviceName,
		GearID:               a.GearID,
		StartLat:             a.StartLat,
		StartLng:             a.StartLng,
		SummaryPolyline:      a.SummaryPolyline,
	}

	return act
}

// ConvertSegmentToStorage converts importer.Segment to storage.Segment.
func ConvertSegmentToStorage(s *Segment) *storage.Segment {
	seg := &storage.Segment{
		ID:            s.ID,
		Name:          s.Name,
		ActivityType:  s.ActivityType,
		Distance:      s.Distance,
		AverageGrade:  s.AverageGrade,
		MaximumGrade:  s.MaximumGrade,
		ElevationHigh: s.ElevationHigh,
		ElevationLow:  s.ElevationLow,
		ClimbCategory: s.ClimbCategory,
		Starred:       s.Starred,
		Polyline:      s.Polyline,
	}

	// Handle latlng arrays
	if len(s.StartLatlng) >= 2 {
		seg.StartLat = &s.StartLatlng[0]
		seg.StartLng = &s.StartLatlng[1]
	}
	if len(s.EndLatlng) >= 2 {
		seg.EndLat = &s.EndLatlng[0]
		seg.EndLng = &s.EndLatlng[1]
	}

	// Handle athlete segment stats
	if s.AthleteSegmentStats.EffortCount > 0 {
		seg.AthleteEffortCount = &s.AthleteSegmentStats.EffortCount
	}
	if s.AthleteSegmentStats.PRElapsedTime > 0 {
		seg.AthletePRElapsedTime = &s.AthleteSegmentStats.PRElapsedTime
	}
	if s.AthleteSegmentStats.PRDate != nil {
		seg.AthletePRDate = &storage.SQLiteTime{Time: *s.AthleteSegmentStats.PRDate}
	}
	seg.AthleteKOMRank = s.AthleteSegmentStats.KOMRank

	return seg
}

// ConvertSegmentEffortToStorage converts importer.SegmentEffort to storage.SegmentEffort.
func ConvertSegmentEffortToStorage(athleteID, activityID int64, e *SegmentEffort, country string) *storage.SegmentEffort {
	effort := &storage.SegmentEffort{
		ID:          e.ID,
		SegmentID:   e.Segment.ID,
		ActivityID:  activityID,
		AthleteID:   athleteID,
		Name:        e.Name,
		ElapsedTime: e.ElapsedTime,
		MovingTime:  e.MovingTime,
		Distance:    e.Distance,
		PRRank:      e.PRRank,
		Country:     country,
	}

	if !e.StartDate.IsZero() {
		effort.StartDate = &storage.SQLiteTime{Time: e.StartDate}
	}
	if !e.StartDateLocal.IsZero() {
		effort.StartDateLocal = &storage.SQLiteTime{Time: e.StartDateLocal}
	}
	if e.AverageWatts > 0 {
		effort.AverageWatts = &e.AverageWatts
	}
	if e.AverageHeartrate > 0 {
		effort.AverageHeartrate = &e.AverageHeartrate
	}
	if e.MaxHeartrate > 0 {
		v := int(e.MaxHeartrate)
		effort.MaxHeartrate = &v
	}

	return effort
}

// BestPhotoURLsFromMap extracts best and thumbnail URLs from a URL map.
func BestPhotoURLsFromMap(urls map[string]string) (best, thumb string) {
	if len(urls) == 0 {
		return "", ""
	}

	// Prefer larger sizes for best
	for _, size := range []string{"1000", "600", "200", "100"} {
		if url, ok := urls[size]; ok && url != "" {
			if best == "" {
				best = url
			}
			thumb = url // Keep updating thumb to get smallest
		}
	}

	// Fallback to any URL
	if best == "" {
		for _, url := range urls {
			if url != "" {
				return url, url
			}
		}
	}

	return best, thumb
}

// ConvertBestEffortsToStorage converts importer.BestEffort slice to storage.BestEffort slice.
func ConvertBestEffortsToStorage(athleteID, activityID int64, sportType string, efforts []BestEffort) []storage.BestEffort {
	storageEfforts := make([]storage.BestEffort, len(efforts))
	for i, e := range efforts {
		dt, canonM := shared.CanonicalBestEffortDistanceType(e.Distance, e.Name)
		storageEfforts[i] = storage.BestEffort{
			AthleteID:    athleteID,
			ActivityID:   activityID,
			SportType:    sportType,
			DistanceType: dt,
			Name:         e.Name,
			DistanceM:    canonM,
			ElapsedTimeS: e.ElapsedTime,
			MovingTimeS:  &e.MovingTime,
			StartIndex:   e.StartIndex,
			EndIndex:     e.EndIndex,
			PRRank:       e.PRRank,
		}
		if !e.StartDate.IsZero() {
			t := storage.SQLiteTime{Time: e.StartDate}
			storageEfforts[i].StartDate = &t
		}
	}
	return storageEfforts
}

// ConvertStreamToStorage converts importer.Stream to storage.ActivityStream.
// Returns nil if stream or stream data is nil, or if marshaling fails.
func ConvertStreamToStorage(activityID int64, streamType string, s *Stream) (*storage.ActivityStream, error) {
	if s == nil || s.Data == nil {
		return nil, nil
	}

	data, err := json.Marshal(s.Data)
	if err != nil {
		return nil, err
	}

	return &storage.ActivityStream{
		ActivityID:   activityID,
		StreamType:   streamType,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
		Data:         data,
	}, nil
}

// ConvertGearToStorage converts importer.Gear to storage.Gear.
func ConvertGearToStorage(athleteID int64, g *Gear) *storage.Gear {
	return &storage.Gear{
		ID:          g.ID,
		AthleteID:   athleteID,
		Name:        g.Name,
		Primary:     g.Primary,
		Retired:     g.Retired,
		Distance:    g.Distance,
		BrandName:   g.BrandName,
		ModelName:   g.ModelName,
		Description: g.Description,
	}
}

// ConvertPhotoToStorage converts importer.Photo to storage.Photo.
// Returns nil if photo has no valid URL.
func ConvertPhotoToStorage(athleteID, activityID int64, p *Photo) *storage.Photo {
	url, thumb := BestPhotoURLsFromMap(p.URLs)
	if url == "" {
		return nil
	}

	var loc json.RawMessage
	if len(p.Location) > 0 {
		if b, err := json.Marshal(p.Location); err == nil {
			loc = b
		}
	}

	return &storage.Photo{
		ID:           p.UniqueID,
		AthleteID:    athleteID,
		ActivityID:   activityID,
		URL:          url,
		ThumbnailURL: thumb,
		Caption:      p.Caption,
		Location:     loc,
	}
}
