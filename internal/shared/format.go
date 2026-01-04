package shared

import "time"

// Date/time formatting constants.
const (
	// MonthKeyFormat formats a time as "YYYY-MM" for use as map keys or display.
	MonthKeyFormat = "2006-01"
)

// FormatMonthKey formats a time as "YYYY-MM".
func FormatMonthKey(t time.Time) string {
	return t.Format(MonthKeyFormat)
}

// WeekdayLabels provides consistent weekday names indexed by time.Weekday (0=Sunday).
var WeekdayLabels = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

// DaytimeLabels provides time-of-day bucket labels in display order.
var DaytimeLabels = []string{
	"Morning",
	"Afternoon",
	"Evening",
	"Night",
}

// DaytimeBucket defines a time-of-day bucket with start/end hours.
type DaytimeBucket struct {
	Label string
	Start int // hour (0-23), inclusive
	End   int // hour (0-23), inclusive
}

// DaytimeBuckets defines the hour ranges for each time-of-day category.
// Used for SQL CASE statements and Go-side bucketing.
var DaytimeBuckets = []DaytimeBucket{
	{Label: "Morning", Start: 5, End: 11},
	{Label: "Afternoon", Start: 12, End: 16},
	{Label: "Evening", Start: 17, End: 21},
	{Label: "Night", Start: 22, End: 4}, // wraps around midnight
}

// GetDaytimeBucket returns the daytime label for a given hour (0-23).
func GetDaytimeBucket(hour int) string {
	switch {
	case hour >= 5 && hour <= 11:
		return "Morning"
	case hour >= 12 && hour <= 16:
		return "Afternoon"
	case hour >= 17 && hour <= 21:
		return "Evening"
	default:
		return "Night"
	}
}
