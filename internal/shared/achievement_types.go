// Package shared contains types and constants shared across packages.
package shared

// AchievementType identifies the category of achievement.
type AchievementType string

const (
	AchievementPersonalRecord    AchievementType = "personal_record"
	AchievementSegmentPR         AchievementType = "segment_pr"
	AchievementEddingtonIncrease AchievementType = "eddington_increase"
	AchievementPowerRecord       AchievementType = "power_record"
	AchievementGoalComplete      AchievementType = "goal_complete"
	AchievementGearMilestone     AchievementType = "gear_milestone"
	AchievementTrainingLoadAlert AchievementType = "training_load_alert"
)

// Training load alert subtypes.
const (
	AlertSubTypeFatigue      = "fatigue"
	AlertSubTypePeakForm     = "peak_form"
	AlertSubTypeOvertraining = "overtraining"
)
