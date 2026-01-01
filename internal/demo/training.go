package demo

import (
	"encoding/json"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// Standard power durations for best efforts (in seconds).
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// PowerBestEffort represents a computed peak power for a duration.
type PowerBestEffort struct {
	ActivityID int64
	AthleteID  int64
	DurationS  int
	Watts      float64
}

// generatePowerBestEfforts computes peak power outputs from activity streams.
// This simulates what would normally be computed from actual power data.
func generatePowerBestEfforts(rng *rand.Rand, athleteID int64, activities []storage.Activity, streams []storage.ActivityStream) []PowerBestEffort {
	var efforts []PowerBestEffort

	// Build a map of activity ID to watts stream
	wattsStreams := make(map[int64][]float64)
	for _, s := range streams {
		if s.StreamType == "watts" {
			var data []float64
			if err := json.Unmarshal(s.Data, &data); err == nil && len(data) > 0 {
				wattsStreams[s.ActivityID] = data
			}
		}
	}

	for _, a := range activities {
		watts, ok := wattsStreams[a.ID]
		if !ok || len(watts) < 5 {
			continue
		}

		for _, dur := range powerDurations {
			if len(watts) < dur {
				continue
			}

			// Calculate rolling max average for this duration
			best := rollingMaxAverage(watts, dur)
			if best <= 0 {
				continue
			}

			efforts = append(efforts, PowerBestEffort{
				ActivityID: a.ID,
				AthleteID:  athleteID,
				DurationS:  dur,
				Watts:      best,
			})
		}
	}

	return efforts
}

// rollingMaxAverage calculates the maximum average over a window of given size.
func rollingMaxAverage(data []float64, window int) float64 {
	if len(data) < window {
		return 0
	}

	var sum float64
	for i := 0; i < window; i++ {
		sum += data[i]
	}

	maxAvg := sum / float64(window)
	for i := window; i < len(data); i++ {
		sum += data[i] - data[i-window]
		avg := sum / float64(window)
		if avg > maxAvg {
			maxAvg = avg
		}
	}

	return math.Round(maxAvg*10) / 10
}

// ActivityTrainingLoad represents TSS and related metrics for an activity.
type ActivityTrainingLoad struct {
	ActivityID      int64
	AthleteID       int64
	SportType       string
	Method          string // cycling_power, running_pace, running_hr
	FTPUsed         float64
	NormalizedPower float64
	IntensityFactor float64
	TSS             float64
}

// generateActivityTrainingLoad computes TSS for each activity that has stream data.
func generateActivityTrainingLoad(rng *rand.Rand, athleteID int64, activities []storage.Activity, streams []storage.ActivityStream, ftpWatts, ftpRunningMps float64) []ActivityTrainingLoad {
	var loads []ActivityTrainingLoad

	// Build stream maps
	wattsStreams := make(map[int64][]float64)
	velocityStreams := make(map[int64][]float64)
	hrStreams := make(map[int64][]float64)

	for _, s := range streams {
		var data []float64
		if err := json.Unmarshal(s.Data, &data); err != nil || len(data) == 0 {
			continue
		}
		switch s.StreamType {
		case "watts":
			wattsStreams[s.ActivityID] = data
		case "velocity_smooth":
			velocityStreams[s.ActivityID] = data
		case "heartrate":
			hrStreams[s.ActivityID] = data
		}
	}

	for _, a := range activities {
		var load *ActivityTrainingLoad

		// Priority 1: Cycling power-based TSS
		if watts, ok := wattsStreams[a.ID]; ok && ftpWatts > 0 && isCyclingActivity(a.SportType) {
			np := normalizedPower(watts)
			if np > 0 {
				ifactor := np / ftpWatts
				tss := calculateTSS(a.MovingTime, np, ftpWatts)
				load = &ActivityTrainingLoad{
					ActivityID:      a.ID,
					AthleteID:       athleteID,
					SportType:       a.SportType,
					Method:          "cycling_power",
					FTPUsed:         ftpWatts,
					NormalizedPower: np,
					IntensityFactor: ifactor,
					TSS:             tss,
				}
			}
		}

		// Priority 2: Running pace-based TSS
		if load == nil && isRunningActivity(a.SportType) {
			if velocity, ok := velocityStreams[a.ID]; ok && ftpRunningMps > 0 {
				np := normalizedPower(velocity) // Same algorithm works for speed
				if np > 0 {
					ifactor := np / ftpRunningMps
					tss := calculateTSS(a.MovingTime, np, ftpRunningMps)
					load = &ActivityTrainingLoad{
						ActivityID:      a.ID,
						AthleteID:       athleteID,
						SportType:       a.SportType,
						Method:          "running_pace",
						FTPUsed:         ftpRunningMps,
						NormalizedPower: np,
						IntensityFactor: ifactor,
						TSS:             tss,
					}
				}
			}
		}

		// Priority 3: HR-based TSS for running (using LTHR estimate)
		if load == nil && isRunningActivity(a.SportType) {
			if hr, ok := hrStreams[a.ID]; ok {
				// Estimate LTHR as 89% of max HR from the stream
				maxHR := maxValue(hr)
				lthr := maxHR * 0.89
				if lthr > 100 {
					np := normalizedPower(hr)
					if np > 0 {
						ifactor := np / lthr
						tss := calculateTSS(a.MovingTime, np, lthr)
						load = &ActivityTrainingLoad{
							ActivityID:      a.ID,
							AthleteID:       athleteID,
							SportType:       a.SportType,
							Method:          "running_hr",
							FTPUsed:         lthr,
							NormalizedPower: np,
							IntensityFactor: ifactor,
							TSS:             tss,
						}
					}
				}
			}
		}

		if load != nil {
			loads = append(loads, *load)
		}
	}

	return loads
}

// normalizedPower calculates NP using 30-second rolling average and 4th power.
func normalizedPower(data []float64) float64 {
	if len(data) < 30 {
		return avgValue(data)
	}

	// 30-second rolling average
	window := 30
	rollingAvg := make([]float64, len(data)-window+1)

	var sum float64
	for i := 0; i < window; i++ {
		sum += data[i]
	}
	rollingAvg[0] = sum / float64(window)

	for i := window; i < len(data); i++ {
		sum += data[i] - data[i-window]
		rollingAvg[i-window+1] = sum / float64(window)
	}

	// Raise to 4th power, average, then take 4th root
	var sum4th float64
	for _, v := range rollingAvg {
		sum4th += math.Pow(v, 4)
	}
	avg4th := sum4th / float64(len(rollingAvg))

	return math.Round(math.Pow(avg4th, 0.25)*10) / 10
}

// calculateTSS computes Training Stress Score.
func calculateTSS(durationS int, np, ftp float64) float64 {
	if ftp <= 0 {
		return 0
	}
	hours := float64(durationS) / 3600.0
	ifactor := np / ftp
	tss := hours * ifactor * ifactor * 100
	return math.Round(tss*10) / 10
}

func avgValue(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func maxValue(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// DailyTrainingLoad represents aggregated daily TSS and CTL/ATL/TSB.
type DailyTrainingLoad struct {
	AthleteID int64
	Day       time.Time
	TSS       float64
	CTL       float64
	ATL       float64
	TSB       float64
}

// generateDailyTrainingLoad aggregates activity TSS by day and computes CTL/ATL/TSB.
func generateDailyTrainingLoad(athleteID int64, activityLoads []ActivityTrainingLoad, activities []storage.Activity) []DailyTrainingLoad {
	// Map activity ID to start date
	activityDates := make(map[int64]time.Time)
	for _, a := range activities {
		activityDates[a.ID] = a.StartDateLocal.Time
	}

	// Aggregate TSS by day
	dailyTSS := make(map[string]float64)
	for _, load := range activityLoads {
		date, ok := activityDates[load.ActivityID]
		if !ok {
			continue
		}
		day := date.Format("2006-01-02")
		dailyTSS[day] += load.TSS
	}

	// Sort days
	var days []string
	for day := range dailyTSS {
		days = append(days, day)
	}
	sort.Strings(days)

	if len(days) == 0 {
		return nil
	}

	// Fill in gaps with zero TSS days for proper EWMA calculation
	startDate, _ := time.Parse("2006-01-02", days[0])
	endDate, _ := time.Parse("2006-01-02", days[len(days)-1])

	var allDays []DailyTrainingLoad
	const ctlTau = 42.0
	const atlTau = 7.0
	var ctl, atl float64

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		tss := dailyTSS[dayStr] // 0 if no activity

		// EWMA calculation
		if len(allDays) == 0 {
			ctl = tss
			atl = tss
		} else {
			ctl += (tss - ctl) * (1.0 / ctlTau)
			atl += (tss - atl) * (1.0 / atlTau)
		}
		tsb := ctl - atl

		allDays = append(allDays, DailyTrainingLoad{
			AthleteID: athleteID,
			Day:       d,
			TSS:       math.Round(tss*10) / 10,
			CTL:       math.Round(ctl*10) / 10,
			ATL:       math.Round(atl*10) / 10,
			TSB:       math.Round(tsb*10) / 10,
		})
	}

	return allDays
}
