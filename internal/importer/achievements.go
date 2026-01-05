// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"log/slog"
)

// AchievementType identifies the category of achievement.
type AchievementType string

const (
	AchievementPersonalRecord    AchievementType = "personal_record"
	AchievementSegmentPR         AchievementType = "segment_pr"
	AchievementEddingtonIncrease AchievementType = "eddington_increase"
	AchievementPowerRecord       AchievementType = "power_record"
	AchievementGearMilestone     AchievementType = "gear_milestone"
	AchievementTrainingLoadAlert AchievementType = "training_load_alert"
)

// Achievement represents a detected achievement during import.
type Achievement struct {
	Type        AchievementType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ActivityID  int64           `json:"activity_id,omitempty"`
	GearID      string          `json:"gear_id,omitempty"`
	Value       float64         `json:"value,omitempty"`
	PrevValue   float64         `json:"prev_value,omitempty"`
}

// AchievementStorage abstracts storage operations needed for achievement detection.
// This is a subset of ImportStorage plus additional query methods.
type AchievementStorage interface {
	// GetEddingtonNumber returns the current Eddington number for the athlete.
	GetEddingtonNumber(ctx context.Context, athleteID int64, sportTypes []string) (int, error)

	// GetGearDistances returns current distances for all gear owned by the athlete.
	// Returns map from gear ID to distance in meters.
	GetGearDistances(ctx context.Context, athleteID int64) (map[string]float64, error)

	// GetBestEffortPRsForActivities returns best efforts with pr_rank=1 for the given activity IDs.
	GetBestEffortPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]BestEffortPRInfo, error)

	// GetSegmentPRsForActivities returns segment efforts with pr_rank=1 for the given activity IDs.
	GetSegmentPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]SegmentPRInfo, error)

	// GetPowerRecords returns all-time best power values for standard durations.
	// Returns map from duration (seconds) to best watts.
	GetPowerRecords(ctx context.Context, athleteID int64) (map[int]float64, error)

	// GetPowerRecordsForActivities returns power records set in the given activity IDs.
	GetPowerRecordsForActivities(ctx context.Context, athleteID int64, activityIDs []int64, durations []int) ([]PowerRecordInfo, error)

	// GetLatestTrainingLoad returns the most recent training load metrics.
	GetLatestTrainingLoad(ctx context.Context, athleteID int64) (*TrainingLoadInfo, error)
}

// BestEffortPRInfo contains PR information for achievement detection.
type BestEffortPRInfo struct {
	ActivityID   int64
	ActivityName string
	DistanceType string
	Name         string
	ElapsedTime  int
}

// SegmentPRInfo contains segment PR information for achievement detection.
type SegmentPRInfo struct {
	ActivityID   int64
	ActivityName string
	SegmentID    int64
	SegmentName  string
	ElapsedTime  int
}

// PowerRecordInfo contains power record information for achievement detection.
type PowerRecordInfo struct {
	ActivityID int64
	DurationS  int
	Watts      float64
}

// TrainingLoadInfo contains training load metrics for achievement detection.
type TrainingLoadInfo struct {
	Day string
	CTL float64 // Chronic Training Load (fitness)
	ATL float64 // Acute Training Load (fatigue)
	TSB float64 // Training Stress Balance (form)
}

// AchievementDetector detects achievements during import.
type AchievementDetector struct {
	logger    *slog.Logger
	storage   AchievementStorage
	athleteID int64

	// Snapshots taken before import for comparison.
	eddingtonBefore     int
	gearDistancesBefore map[string]float64
	powerRecordsBefore  map[int]float64
}

// NewAchievementDetector creates a new achievement detector.
func NewAchievementDetector(logger *slog.Logger, storage AchievementStorage, athleteID int64) *AchievementDetector {
	if logger == nil {
		logger = slog.Default()
	}
	return &AchievementDetector{
		logger:    logger,
		storage:   storage,
		athleteID: athleteID,
	}
}

// TakeSnapshot captures current state before import for comparison after.
func (d *AchievementDetector) TakeSnapshot(ctx context.Context) error {
	if d.storage == nil {
		return nil
	}

	var err error

	// Snapshot Eddington number (for cycling by default).
	d.eddingtonBefore, err = d.storage.GetEddingtonNumber(ctx, d.athleteID, []string{"Ride", "VirtualRide", "GravelRide", "MountainBikeRide"})
	if err != nil {
		d.logger.Debug("failed to snapshot Eddington number", "error", err)
		d.eddingtonBefore = 0
	}

	// Snapshot gear distances.
	d.gearDistancesBefore, err = d.storage.GetGearDistances(ctx, d.athleteID)
	if err != nil {
		d.logger.Debug("failed to snapshot gear distances", "error", err)
		d.gearDistancesBefore = make(map[string]float64)
	}

	// Snapshot power records.
	d.powerRecordsBefore, err = d.storage.GetPowerRecords(ctx, d.athleteID)
	if err != nil {
		d.logger.Debug("failed to snapshot power records", "error", err)
		d.powerRecordsBefore = make(map[int]float64)
	}

	d.logger.Debug("achievement snapshot taken",
		"eddington_before", d.eddingtonBefore,
		"gear_count", len(d.gearDistancesBefore),
		"power_records", len(d.powerRecordsBefore))

	return nil
}

// DetectAchievements detects achievements based on imported activities.
// Returns a slice of achievements detected during this import.
func (d *AchievementDetector) DetectAchievements(ctx context.Context, importedActivityIDs []int64) ([]Achievement, error) {
	if d.storage == nil || len(importedActivityIDs) == 0 {
		return nil, nil
	}

	var achievements []Achievement

	// Detect personal records (best efforts).
	prs, err := d.detectPersonalRecords(ctx, importedActivityIDs)
	if err != nil {
		d.logger.Debug("failed to detect personal records", "error", err)
	} else {
		achievements = append(achievements, prs...)
	}

	// Detect segment PRs.
	segmentPRs, err := d.detectSegmentPRs(ctx, importedActivityIDs)
	if err != nil {
		d.logger.Debug("failed to detect segment PRs", "error", err)
	} else {
		achievements = append(achievements, segmentPRs...)
	}

	// Detect Eddington number increase.
	eddington, err := d.detectEddingtonChange(ctx)
	if err != nil {
		d.logger.Debug("failed to detect Eddington change", "error", err)
	} else if eddington != nil {
		achievements = append(achievements, *eddington)
	}

	// Detect power records.
	powerRecords, err := d.detectPowerRecords(ctx, importedActivityIDs)
	if err != nil {
		d.logger.Debug("failed to detect power records", "error", err)
	} else {
		achievements = append(achievements, powerRecords...)
	}

	// Detect gear milestones.
	gearMilestones, err := d.detectGearMilestones(ctx)
	if err != nil {
		d.logger.Debug("failed to detect gear milestones", "error", err)
	} else {
		achievements = append(achievements, gearMilestones...)
	}

	// Detect training load alerts.
	trainingAlerts, err := d.detectTrainingLoadAlerts(ctx)
	if err != nil {
		d.logger.Debug("failed to detect training load alerts", "error", err)
	} else {
		achievements = append(achievements, trainingAlerts...)
	}

	d.logger.Info("achievements detected", "count", len(achievements))
	return achievements, nil
}

// detectPersonalRecords checks best_efforts table for pr_rank=1 on imported activity IDs.
func (d *AchievementDetector) detectPersonalRecords(ctx context.Context, activityIDs []int64) ([]Achievement, error) {
	prs, err := d.storage.GetBestEffortPRsForActivities(ctx, d.athleteID, activityIDs)
	if err != nil {
		return nil, err
	}

	achievements := make([]Achievement, 0, len(prs))
	for _, pr := range prs {
		achievements = append(achievements, Achievement{
			Type:        AchievementPersonalRecord,
			Title:       "New Personal Record",
			Description: formatPRDescription(pr.Name, pr.ElapsedTime),
			ActivityID:  pr.ActivityID,
		})
	}
	return achievements, nil
}

// detectSegmentPRs checks segment_efforts table for pr_rank=1.
func (d *AchievementDetector) detectSegmentPRs(ctx context.Context, activityIDs []int64) ([]Achievement, error) {
	prs, err := d.storage.GetSegmentPRsForActivities(ctx, d.athleteID, activityIDs)
	if err != nil {
		return nil, err
	}

	achievements := make([]Achievement, 0, len(prs))
	for _, pr := range prs {
		achievements = append(achievements, Achievement{
			Type:        AchievementSegmentPR,
			Title:       "Segment PR",
			Description: formatSegmentPRDescription(pr.SegmentName, pr.ElapsedTime),
			ActivityID:  pr.ActivityID,
		})
	}
	return achievements, nil
}

// detectEddingtonChange compares current Eddington vs snapshot.
func (d *AchievementDetector) detectEddingtonChange(ctx context.Context) (*Achievement, error) {
	current, err := d.storage.GetEddingtonNumber(ctx, d.athleteID, []string{"Ride", "VirtualRide", "GravelRide", "MountainBikeRide"})
	if err != nil {
		return nil, err
	}

	if current > d.eddingtonBefore {
		return &Achievement{
			Type:        AchievementEddingtonIncrease,
			Title:       "Eddington Number Increased",
			Description: formatEddingtonDescription(d.eddingtonBefore, current),
			Value:       float64(current),
			PrevValue:   float64(d.eddingtonBefore),
		}, nil
	}
	return nil, nil
}

// Standard power durations to check for records.
var standardPowerDurations = []int{5, 60, 300, 1200, 3600} // 5s, 1m, 5m, 20m, 1hr

// detectPowerRecords checks power_best_efforts for new records.
func (d *AchievementDetector) detectPowerRecords(ctx context.Context, activityIDs []int64) ([]Achievement, error) {
	records, err := d.storage.GetPowerRecordsForActivities(ctx, d.athleteID, activityIDs, standardPowerDurations)
	if err != nil {
		return nil, err
	}

	var achievements []Achievement
	for _, rec := range records {
		prevBest := d.powerRecordsBefore[rec.DurationS]
		if rec.Watts > prevBest {
			achievements = append(achievements, Achievement{
				Type:        AchievementPowerRecord,
				Title:       "New Power Record",
				Description: formatPowerRecordDescription(rec.DurationS, rec.Watts, prevBest),
				ActivityID:  rec.ActivityID,
				Value:       rec.Watts,
				PrevValue:   prevBest,
			})
		}
	}
	return achievements, nil
}

// Gear milestone thresholds in meters (5000km increments).
var gearMilestoneThresholds = []float64{
	5000 * 1000,  // 5,000 km
	10000 * 1000, // 10,000 km
	15000 * 1000, // 15,000 km
	20000 * 1000, // 20,000 km
	25000 * 1000, // 25,000 km
	30000 * 1000, // 30,000 km
}

// detectGearMilestones checks if any gear crossed a 5000km threshold.
func (d *AchievementDetector) detectGearMilestones(ctx context.Context) ([]Achievement, error) {
	currentDistances, err := d.storage.GetGearDistances(ctx, d.athleteID)
	if err != nil {
		return nil, err
	}

	var achievements []Achievement
	for gearID, currentDist := range currentDistances {
		prevDist := d.gearDistancesBefore[gearID]

		for _, threshold := range gearMilestoneThresholds {
			if prevDist < threshold && currentDist >= threshold {
				achievements = append(achievements, Achievement{
					Type:        AchievementGearMilestone,
					Title:       "Gear Milestone",
					Description: formatGearMilestoneDescription(threshold),
					GearID:      gearID,
					Value:       currentDist,
					PrevValue:   prevDist,
				})
				break // Only report the highest threshold crossed
			}
		}
	}
	return achievements, nil
}

// Training load thresholds.
const (
	tsbOvertrainingThreshold = -30.0 // TSB below this suggests overtraining
	tsbPeakFormThreshold     = 15.0  // TSB above this suggests peak form
	atlHighFatigueThreshold  = 80.0  // ATL above this suggests high fatigue
	ctlFitnessGainThreshold  = 60.0  // CTL above this suggests good fitness
)

// detectTrainingLoadAlerts checks TSB/ATL/CTL thresholds.
func (d *AchievementDetector) detectTrainingLoadAlerts(ctx context.Context) ([]Achievement, error) {
	load, err := d.storage.GetLatestTrainingLoad(ctx, d.athleteID)
	if err != nil {
		return nil, err
	}
	if load == nil {
		return nil, nil
	}

	var achievements []Achievement

	// Check for overtraining risk (very negative TSB).
	if load.TSB < tsbOvertrainingThreshold {
		achievements = append(achievements, Achievement{
			Type:        AchievementTrainingLoadAlert,
			Title:       "Overtraining Warning",
			Description: formatOvertrainingDescription(load.TSB),
			Value:       load.TSB,
		})
	}

	// Check for peak form (positive TSB after good training block).
	if load.TSB > tsbPeakFormThreshold && load.CTL > ctlFitnessGainThreshold {
		achievements = append(achievements, Achievement{
			Type:        AchievementTrainingLoadAlert,
			Title:       "Peak Form",
			Description: formatPeakFormDescription(load.TSB, load.CTL),
			Value:       load.TSB,
		})
	}

	// Check for high fatigue.
	if load.ATL > atlHighFatigueThreshold {
		achievements = append(achievements, Achievement{
			Type:        AchievementTrainingLoadAlert,
			Title:       "High Fatigue",
			Description: formatHighFatigueDescription(load.ATL),
			Value:       load.ATL,
		})
	}

	return achievements, nil
}

// Formatting helpers.

func formatPRDescription(name string, elapsedTime int) string {
	return name + " in " + formatDuration(elapsedTime)
}

func formatSegmentPRDescription(segmentName string, elapsedTime int) string {
	return segmentName + " in " + formatDuration(elapsedTime)
}

func formatEddingtonDescription(before, after int) string {
	return "E" + itoa(before) + " -> E" + itoa(after)
}

func formatPowerRecordDescription(durationS int, watts, prevBest float64) string {
	durationStr := formatPowerDuration(durationS)
	if prevBest > 0 {
		return durationStr + ": " + itoa(int(watts)) + "W (+" + itoa(int(watts-prevBest)) + "W)"
	}
	return durationStr + ": " + itoa(int(watts)) + "W"
}

func formatGearMilestoneDescription(thresholdMeters float64) string {
	km := int(thresholdMeters / 1000)
	return itoa(km) + " km milestone reached"
}

func formatOvertrainingDescription(tsb float64) string {
	return "Training Stress Balance at " + itoa(int(tsb)) + " - consider recovery"
}

func formatPeakFormDescription(tsb, ctl float64) string {
	return "TSB " + itoa(int(tsb)) + " with CTL " + itoa(int(ctl)) + " - ideal for racing"
}

func formatHighFatigueDescription(atl float64) string {
	return "Acute Training Load at " + itoa(int(atl)) + " - high fatigue"
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return itoa(h) + "h " + itoa(m) + "m " + itoa(s) + "s"
	}
	if m > 0 {
		return itoa(m) + "m " + itoa(s) + "s"
	}
	return itoa(s) + "s"
}

func formatPowerDuration(seconds int) string {
	switch seconds {
	case 5:
		return "5s"
	case 60:
		return "1min"
	case 300:
		return "5min"
	case 1200:
		return "20min"
	case 3600:
		return "1hr"
	default:
		return itoa(seconds) + "s"
	}
}

// itoa converts int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b) - 1
	for i > 0 {
		b[pos] = byte('0' + i%10)
		pos--
		i /= 10
	}
	if neg {
		b[pos] = '-'
		pos--
	}
	return string(b[pos+1:])
}
