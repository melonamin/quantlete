package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// ActivityZoneDistribution stores the time spent in each HR zone for an activity.
type ActivityZoneDistribution struct {
	ActivityID   int64     `json:"activity_id"`
	ZoneDefID    string    `json:"zone_def_id"` // "sport_type:effective_from" for invalidation tracking
	SecondsZ1    int       `json:"seconds_z1"`
	SecondsZ2    int       `json:"seconds_z2"`
	SecondsZ3    int       `json:"seconds_z3"`
	SecondsZ4    int       `json:"seconds_z4"`
	SecondsZ5    int       `json:"seconds_z5"`
	TotalSeconds int       `json:"total_seconds"`
	ComputedAt   time.Time `json:"computed_at"`
}

// WeeklyZoneDistribution represents aggregated zone data for a week.
type WeeklyZoneDistribution struct {
	Week      string  `json:"week"` // "YYYY-WNN" format
	SecondsZ1 int     `json:"seconds_z1"`
	SecondsZ2 int     `json:"seconds_z2"`
	SecondsZ3 int     `json:"seconds_z3"`
	SecondsZ4 int     `json:"seconds_z4"`
	SecondsZ5 int     `json:"seconds_z5"`
	Total     int     `json:"total"`
	PercentZ1 float64 `json:"percent_z1"`
	PercentZ2 float64 `json:"percent_z2"`
	PercentZ3 float64 `json:"percent_z3"`
	PercentZ4 float64 `json:"percent_z4"`
	PercentZ5 float64 `json:"percent_z5"`
}

type ZoneDistributionRepository struct {
	db      *DB
	streams *StreamRepository
	zones   *ZonesRepository
}

func NewZoneDistributionRepository(db *DB, streams *StreamRepository, zones *ZonesRepository) *ZoneDistributionRepository {
	return &ZoneDistributionRepository{
		db:      db,
		streams: streams,
		zones:   zones,
	}
}

// Save stores or updates a zone distribution for an activity.
func (r *ZoneDistributionRepository) Save(ctx context.Context, dist ActivityZoneDistribution) error {
	q := NewQueries(r.db.Conn())
	return q.UpsertActivityZoneDistribution(
		ctx,
		dist.ActivityID,
		dist.ZoneDefID,
		strconv.Itoa(dist.SecondsZ1),
		strconv.Itoa(dist.SecondsZ2),
		strconv.Itoa(dist.SecondsZ3),
		strconv.Itoa(dist.SecondsZ4),
		strconv.Itoa(dist.SecondsZ5),
		strconv.Itoa(dist.TotalSeconds),
		SQLiteTime{Time: dist.ComputedAt},
	)
}

// GetWeeklyDistribution returns aggregated zone data for the last N weeks.
func (r *ZoneDistributionRepository) GetWeeklyDistribution(ctx context.Context, athleteID int64, weeks int) ([]WeeklyZoneDistribution, error) {
	q := NewQueries(r.db.Conn())
	rows, err := q.GetWeeklyZoneDistribution(ctx, athleteID, strconv.Itoa(weeks))
	if err != nil {
		return nil, fmt.Errorf("get weekly zone distribution: %w", err)
	}

	result := make([]WeeklyZoneDistribution, len(rows))
	for i, row := range rows {
		total := int(row.Total)
		result[i] = WeeklyZoneDistribution{
			Week:      row.Week,
			SecondsZ1: int(row.Z1),
			SecondsZ2: int(row.Z2),
			SecondsZ3: int(row.Z3),
			SecondsZ4: int(row.Z4),
			SecondsZ5: int(row.Z5),
			Total:     total,
		}
		// Calculate percentages
		if total > 0 {
			result[i].PercentZ1 = float64(result[i].SecondsZ1) / float64(total) * 100
			result[i].PercentZ2 = float64(result[i].SecondsZ2) / float64(total) * 100
			result[i].PercentZ3 = float64(result[i].SecondsZ3) / float64(total) * 100
			result[i].PercentZ4 = float64(result[i].SecondsZ4) / float64(total) * 100
			result[i].PercentZ5 = float64(result[i].SecondsZ5) / float64(total) * 100
		}
	}
	return result, nil
}

// GetActivitiesWithoutDistribution returns activity IDs that have HR streams but no zone distribution.
func (r *ZoneDistributionRepository) GetActivitiesWithoutDistribution(ctx context.Context, athleteID int64) ([]int64, error) {
	q := NewQueries(r.db.Conn())
	rows, err := q.GetActivityIDsWithoutZoneDistribution(ctx, athleteID)
	if err != nil {
		return nil, fmt.Errorf("get activities without distribution: %w", err)
	}

	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids, nil
}

// DeleteForSportType deletes zone distributions for activities matching the sport type and date range.
// Used when zone definitions change to trigger recomputation.
func (r *ZoneDistributionRepository) DeleteForSportType(ctx context.Context, athleteID int64, sportType string, effectiveFrom time.Time) error {
	q := NewQueries(r.db.Conn())
	return q.DeleteZoneDistributionsForSportType(ctx, athleteID, sportType, effectiveFrom.Format("2006-01-02"))
}

// ComputeFromHRStream computes zone distribution from a heart rate stream.
// Returns the distribution with seconds spent in each zone.
func ComputeFromHRStream(hrStream json.RawMessage, method string, cfg *HRZoneConfig) (*ActivityZoneDistribution, error) {
	if len(hrStream) == 0 || cfg == nil {
		return nil, nil
	}

	hrs, err := decodeFloat64Array(hrStream)
	if err != nil {
		return nil, fmt.Errorf("decode HR stream: %w", err)
	}

	if len(hrs) == 0 {
		return nil, nil
	}

	// Count seconds in each zone (assuming 1Hz sampling)
	secondsPerZone := make([]int, 5)
	totalSeconds := 0

	for _, hr := range hrs {
		idx := HRZoneIndex(method, cfg, hr)
		if idx >= 0 && idx < 5 {
			secondsPerZone[idx]++
			totalSeconds++
		}
	}

	if totalSeconds == 0 {
		return nil, nil
	}

	return &ActivityZoneDistribution{
		SecondsZ1:    secondsPerZone[0],
		SecondsZ2:    secondsPerZone[1],
		SecondsZ3:    secondsPerZone[2],
		SecondsZ4:    secondsPerZone[3],
		SecondsZ5:    secondsPerZone[4],
		TotalSeconds: totalSeconds,
		ComputedAt:   time.Now(),
	}, nil
}

// ActivityForZone holds the minimal activity data needed for zone computation.
type ActivityForZone struct {
	ID        int64
	SportType string
	StartDate time.Time
}

// EnsureComputed ensures zone distributions are computed for all activities with HR streams.
// This is called lazily when zone trend data is requested.
func (r *ZoneDistributionRepository) EnsureComputed(ctx context.Context, athleteID int64) error {
	if r.streams == nil || r.zones == nil {
		return nil
	}

	// Get activities with HR streams but no zone distribution
	activityIDs, err := r.GetActivitiesWithoutDistribution(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("get activities without distribution: %w", err)
	}

	if len(activityIDs) == 0 {
		return nil
	}

	// Process each activity
	for _, activityID := range activityIDs {
		if err := r.computeAndSaveForActivity(ctx, athleteID, activityID); err != nil {
			slog.Warn("failed to compute zone distribution",
				"error", err,
				"athlete_id", athleteID,
				"activity_id", activityID)
			continue
		}
	}

	return nil
}

func (r *ZoneDistributionRepository) computeAndSaveForActivity(ctx context.Context, athleteID, activityID int64) error {
	// Get activity info
	activity, err := r.getActivityForZone(ctx, activityID)
	if err != nil || activity == nil {
		return err
	}

	// Get applicable zone definition
	def, cfg, err := r.zones.GetApplicableHR(ctx, athleteID, activity.SportType, activity.StartDate)
	if err != nil {
		return fmt.Errorf("get zone definition: %w", err)
	}
	if def == nil || cfg == nil {
		return nil // No zone definition configured
	}

	// Get HR stream
	streams, err := r.streams.GetByActivityID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("get streams: %w", err)
	}

	var hrStream json.RawMessage
	for _, s := range streams {
		if s.StreamType == "heartrate" {
			hrStream = s.Data
			break
		}
	}

	if len(hrStream) == 0 {
		return nil // No HR data
	}

	// Compute zone distribution
	dist, err := ComputeFromHRStream(hrStream, def.Method, cfg)
	if err != nil {
		return fmt.Errorf("compute zone distribution: %w", err)
	}
	if dist == nil {
		return nil
	}

	// Set activity ID and zone definition ID
	dist.ActivityID = activityID
	dist.ZoneDefID = fmt.Sprintf("%s:%s", def.SportType, def.EffectiveFrom)

	// Save
	return r.Save(ctx, *dist)
}

func (r *ZoneDistributionRepository) getActivityForZone(ctx context.Context, activityID int64) (*ActivityForZone, error) {
	q := NewQueries(r.db.Conn())
	row, err := q.GetActivityForZoneComputation(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity for zone: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	startDate, err := time.Parse("2006-01-02T15:04:05Z", row.StartDate)
	if err != nil {
		return nil, fmt.Errorf("parse start date: %w", err)
	}

	return &ActivityForZone{
		ID:        row.ID,
		SportType: row.SportType,
		StartDate: startDate,
	}, nil
}

// GetByActivityID returns the zone distribution for a specific activity, or nil if not found.
func (r *ZoneDistributionRepository) GetByActivityID(ctx context.Context, activityID int64) (*ActivityZoneDistribution, error) {
	q := NewQueries(r.db.Conn())
	row, err := q.GetActivityZoneDistribution(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity zone distribution: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	computedAt, err := time.Parse("2006-01-02T15:04:05Z", row.ComputedAt)
	if err != nil {
		return nil, fmt.Errorf("parse computed_at: %w", err)
	}

	return &ActivityZoneDistribution{
		ActivityID:   row.ActivityID,
		ZoneDefID:    row.ZoneDefID,
		SecondsZ1:    row.SecondsZ1,
		SecondsZ2:    row.SecondsZ2,
		SecondsZ3:    row.SecondsZ3,
		SecondsZ4:    row.SecondsZ4,
		SecondsZ5:    row.SecondsZ5,
		TotalSeconds: row.TotalSeconds,
		ComputedAt:   computedAt,
	}, nil
}
