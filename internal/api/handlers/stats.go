package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

type StatsHandler struct {
	db           *storage.DB
	stats        *storage.StatsRepository
	power        *storage.PowerRepository
	streams      *storage.StreamRepository
	metrics      *storage.AthleteMetricsRepository
	zones        *storage.ZonesRepository
	bestEfforts  *storage.BestEffortsRepository
	trainingLoad *storage.TrainingLoadRepository
	strava       *strava.Client
}

func NewStatsHandler(
	db *storage.DB,
	stats *storage.StatsRepository,
	power *storage.PowerRepository,
	streams *storage.StreamRepository,
	metrics *storage.AthleteMetricsRepository,
	zones *storage.ZonesRepository,
	bestEfforts *storage.BestEffortsRepository,
	trainingLoad *storage.TrainingLoadRepository,
	stravaClient *strava.Client,
) *StatsHandler {
	return &StatsHandler{
		db:           db,
		stats:        stats,
		power:        power,
		streams:      streams,
		metrics:      metrics,
		zones:        zones,
		bestEfforts:  bestEfforts,
		trainingLoad: trainingLoad,
		strava:       stravaClient,
	}
}

// HeatmapResponse contains the heatmap data and summary statistics.
type HeatmapResponse struct {
	Activities []storage.HeatmapActivity    `json:"activities"`
	Total      int                          `json:"total"`
	Limit      int                          `json:"limit,omitempty"`
	Offset     int                          `json:"offset,omitempty"`
	Countries  []storage.HeatmapCountryStat `json:"countries,omitempty"`
}

// GetHeatmapData handles GET /api/v1/stats/heatmap
func (h *StatsHandler) GetHeatmapData(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	q := r.URL.Query()
	filters := storage.HeatmapFilters{}

	// Parse sport types
	if sportTypes := q.Get("sport_type"); sportTypes != "" {
		filters.SportTypes = strings.Split(sportTypes, ",")
	}

	// Parse date range
	if after := q.Get("after"); after != "" {
		if t, err := time.Parse("2006-01-02", after); err == nil {
			filters.StartAfter = &t
		}
	}

	if before := q.Get("before"); before != "" {
		if t, err := time.Parse("2006-01-02", before); err == nil {
			filters.StartBefore = &t
		}
	}

	// Parse commute filter
	if commute := q.Get("commute"); commute != "" {
		v := commute == "true" || commute == "1"
		filters.Commute = &v
	}

	// Parse workout type filter
	if wt := q.Get("workout_type"); wt != "" {
		if parsed, err := strconv.Atoi(wt); err == nil {
			v := parsed
			filters.WorkoutType = &v
		}
	}

	// Parse pagination
	if limit := q.Get("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			if parsed > MaxPageSize {
				parsed = MaxPageSize
			}
			filters.Limit = parsed
		}
	}
	if offset := q.Get("offset"); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil && parsed >= 0 {
			filters.Offset = parsed
		}
	}

	// Get total count for pagination metadata (before applying limit/offset)
	total, err := h.stats.CountHeatmapActivities(r.Context(), athlete.ID, filters)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to count heatmap data"})
		return
	}

	activities, err := h.stats.GetHeatmapData(r.Context(), athlete.ID, filters)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get heatmap data"})
		return
	}

	if activities == nil {
		activities = []storage.HeatmapActivity{}
	}

	countries, _ := h.stats.GetHeatmapCountries(r.Context(), athlete.ID, filters)

	writeJSON(w, http.StatusOK, HeatmapResponse{
		Activities: activities,
		Total:      total,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
		Countries:  countries,
	})
}

// GetEddingtonData handles GET /api/v1/stats/eddington
func (h *StatsHandler) GetEddingtonData(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Parse sport types filter
	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	result, err := h.stats.GetEddingtonData(r.Context(), athlete.ID, sportTypes)
	if err != nil {
		slog.Error("failed to get eddington data", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get eddington data"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetEddingtonHistory handles GET /api/v1/stats/eddington/history
func (h *StatsHandler) GetEddingtonHistory(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	points, err := h.stats.GetEddingtonHistory(r.Context(), athlete.ID, sportTypes)
	if err != nil {
		slog.Error("failed to get eddington history", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get eddington history"})
		return
	}

	if points == nil {
		points = []storage.EddingtonHistoryPoint{}
	}

	writeJSON(w, http.StatusOK, points)
}

type BestEffortPRResponse struct {
	DistanceType   string  `json:"distance_type"`
	Name           string  `json:"name"`
	DistanceM      float64 `json:"distance_m"`
	ElapsedTimeS   int     `json:"elapsed_time_s"`
	MovingTimeS    *int    `json:"moving_time_s,omitempty"`
	PRRank         *int    `json:"pr_rank,omitempty"`
	ActivityID     int64   `json:"activity_id"`
	ActivityName   string  `json:"activity_name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
}

// GetBestEfforts handles GET /api/v1/stats/best-efforts
// Returns one all-time PR (fastest elapsed time) per distance_type.
func (h *StatsHandler) GetBestEfforts(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	if h.bestEfforts == nil {
		writeJSON(w, http.StatusOK, []BestEffortPRResponse{})
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	prs, err := h.bestEfforts.ListPRs(r.Context(), athlete.ID, sportTypes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load best efforts"})
		return
	}

	out := make([]BestEffortPRResponse, 0, len(prs))
	for _, p := range prs {
		out = append(out, BestEffortPRResponse{
			DistanceType:   p.DistanceType,
			Name:           p.Name,
			DistanceM:      p.DistanceM,
			ElapsedTimeS:   p.ElapsedTimeS,
			MovingTimeS:    p.MovingTimeS,
			PRRank:         p.PRRank,
			ActivityID:     p.ActivityID,
			ActivityName:   p.ActivityName,
			SportType:      p.SportType,
			StartDateLocal: p.StartDateLocal.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

type BestEffortListItemResponse struct {
	DistanceType   string  `json:"distance_type"`
	Name           string  `json:"name"`
	DistanceM      float64 `json:"distance_m"`
	ElapsedTimeS   int     `json:"elapsed_time_s"`
	MovingTimeS    *int    `json:"moving_time_s,omitempty"`
	PRRank         *int    `json:"pr_rank,omitempty"`
	StartIndex     *int    `json:"start_index,omitempty"`
	EndIndex       *int    `json:"end_index,omitempty"`
	ActivityID     int64   `json:"activity_id"`
	ActivityName   string  `json:"activity_name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
}

// GetBestEffortsByDistance handles GET /api/v1/stats/best-efforts/{distanceType}
func (h *StatsHandler) GetBestEffortsByDistance(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	if h.bestEfforts == nil {
		writeJSON(w, http.StatusOK, []BestEffortListItemResponse{})
		return
	}

	distanceType := chi.URLParam(r, "distanceType")
	if strings.TrimSpace(distanceType) == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "distanceType is required"})
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	items, err := h.bestEfforts.ListByDistanceType(r.Context(), athlete.ID, distanceType, sportTypes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load best efforts"})
		return
	}

	out := make([]BestEffortListItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, BestEffortListItemResponse{
			DistanceType:   it.DistanceType,
			Name:           it.Name,
			DistanceM:      it.DistanceM,
			ElapsedTimeS:   it.ElapsedTimeS,
			MovingTimeS:    it.MovingTimeS,
			PRRank:         it.PRRank,
			StartIndex:     it.StartIndex,
			EndIndex:       it.EndIndex,
			ActivityID:     it.ActivityID,
			ActivityName:   it.ActivityName,
			SportType:      it.SportType,
			StartDateLocal: it.StartDateLocal.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

// GetRewindYears handles GET /api/v1/stats/rewind/years
func (h *StatsHandler) GetRewindYears(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	years, err := h.stats.ListRewindYears(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load rewind years"})
		return
	}
	if years == nil {
		years = []int{}
	}
	writeJSON(w, http.StatusOK, years)
}

// GetRewind handles GET /api/v1/stats/rewind?year=YYYY
func (h *StatsHandler) GetRewind(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	year := time.Now().Year()
	if s := strings.TrimSpace(r.URL.Query().Get("year")); s != "" {
		if s == "all" || s == "0" {
			year = 0
		} else if parsed, err := strconv.Atoi(s); err == nil && parsed >= 1900 && parsed <= time.Now().Year()+1 {
			year = parsed
		}
	}

	report, err := h.stats.GetRewind(r.Context(), athlete.ID, year)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to build rewind"})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

type PowerStatsResponse struct {
	DurationsS []int                                   `json:"durations_s"`
	Best       []storage.PeakPowerBest                 `json:"best"`
	History    map[int][]storage.PeakPowerHistoryPoint `json:"history"`
}

// GetPowerStats handles GET /api/v1/stats/power
func (h *StatsHandler) GetPowerStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	durations := []int{5, 10, 30, 60, 300, 480, 1200, 3600}

	var after *time.Time
	var before *time.Time
	if s := r.URL.Query().Get("after"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			after = &t
		}
	}
	if s := r.URL.Query().Get("before"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			before = &t
		}
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	if err := h.power.EnsureComputedForRange(r.Context(), athlete.ID, after, before, sportTypes, durations); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to compute power stats"})
		return
	}

	best, err := h.power.GetBest(r.Context(), athlete.ID, durations, after, before, sportTypes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load power stats"})
		return
	}

	history := make(map[int][]storage.PeakPowerHistoryPoint)
	for _, d := range durations {
		points, err := h.power.GetHistory(r.Context(), athlete.ID, d, after, before, sportTypes)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load power history"})
			return
		}
		history[d] = points
	}

	writeJSON(w, http.StatusOK, PowerStatsResponse{
		DurationsS: durations,
		Best:       best,
		History:    history,
	})
}

type HRZonesResponse struct {
	Method        string               `json:"method"`
	Zones         storage.HRZoneConfig `json:"zones"`
	SecondsByZone []int                `json:"seconds_by_zone"`
	TotalSeconds  int                  `json:"total_seconds"`
}

// GetHRZones handles GET /api/v1/stats/hr-zones
func (h *StatsHandler) GetHRZones(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var after *time.Time
	var before *time.Time
	if s := r.URL.Query().Get("after"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			after = &t
		}
	}
	if s := r.URL.Query().Get("before"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			before = &t
		}
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	defs, err := h.zones.ListHR(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load zone definitions"})
		return
	}
	if len(defs) == 0 {
		// Default: percent_hrmax with a placeholder HR max.
		placeholder := storage.HRZoneConfig{HRMax: 190, Bounds: []float64{0.6, 0.7, 0.8, 0.9, 1.0}}
		b, _ := json.Marshal(placeholder)
		_ = h.zones.UpsertHR(r.Context(), athlete.ID, storage.HRZoneDefinition{
			SportType:     "All",
			EffectiveFrom: "1970-01-01",
			Method:        "percent_hrmax",
			Zones:         b,
		})
		defs, _ = h.zones.ListHR(r.Context(), athlete.ID)
	}

	// Fetch activities with HR streams in range.
	query := `
		SELECT DISTINCT a.id, a.sport_type, a.start_date
		FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id AND s.stream_type = 'heartrate'
		WHERE a.athlete_id = ?
	`
	args := []any{athlete.ID}
	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, *before)
	}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + storage.JoinStrings(placeholders, ",") + ")"
	}

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to query activities"})
		return
	}
	defer func() { _ = rows.Close() }()

	secondsByZone := make([]int, 5)
	totalSeconds := 0

	var method string
	var cfg storage.HRZoneConfig
	cfgSet := false

	for rows.Next() {
		var id int64
		var sportType string
		var start time.Time
		if err := rows.Scan(&id, &sportType, &start); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to scan activities"})
			return
		}

		group := sportGroup(sportType)
		def, zcfg, err := h.zones.GetApplicableHR(r.Context(), athlete.ID, group, start)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load zone definition"})
			return
		}
		if def == nil || zcfg == nil || len(zcfg.Bounds) < 5 {
			continue
		}
		if !cfgSet {
			method = def.Method
			cfg = *zcfg
			cfgSet = true
		}

		streams, err := h.streams.GetByActivityID(r.Context(), id)
		if err != nil {
			continue
		}
		var hrRaw []byte
		for _, s := range streams {
			if s.StreamType == "heartrate" {
				hrRaw = s.Data
				break
			}
		}
		hrs, err := storage.DecodeFloat64Array(hrRaw)
		if err != nil {
			continue
		}
		for _, hr := range hrs {
			idx := storage.HRZoneIndex(def.Method, zcfg, hr)
			if idx >= 0 && idx < 5 {
				secondsByZone[idx]++
				totalSeconds++
			}
		}
	}

	writeJSON(w, http.StatusOK, HRZonesResponse{
		Method:        method,
		Zones:         cfg,
		SecondsByZone: secondsByZone,
		TotalSeconds:  totalSeconds,
	})
}

type PowerZonesResponse struct {
	FTPWatts      float64   `json:"ftp_watts,omitempty"`
	SecondsByZone []int     `json:"seconds_by_zone"`
	TotalSeconds  int       `json:"total_seconds"`
	Bounds        []float64 `json:"bounds"`
}

// GetPowerZones handles GET /api/v1/stats/power-zones
func (h *StatsHandler) GetPowerZones(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var after *time.Time
	var before *time.Time
	if s := r.URL.Query().Get("after"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			after = &t
		}
	}
	if s := r.URL.Query().Get("before"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			before = &t
		}
	}

	// Standard 5-zone bounds based on FTP percentage.
	bounds := []float64{0.55, 0.75, 0.9, 1.05, 999}
	secondsByZone := make([]int, 5)
	totalSeconds := 0

	query := `
		SELECT DISTINCT a.id, a.start_date
		FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id AND s.stream_type = 'watts'
		WHERE a.athlete_id = ?
	`
	args := []any{athlete.ID}
	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, *before)
	}

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to query activities"})
		return
	}
	defer func() { _ = rows.Close() }()

	var ftpUsed float64
	for rows.Next() {
		var id int64
		var start time.Time
		if err := rows.Scan(&id, &start); err != nil {
			continue
		}
		ftpPoint, err := h.metrics.LatestBefore(r.Context(), athlete.ID, "ftp_cycling_watts", start)
		if err != nil || ftpPoint == nil || ftpPoint.Value <= 0 {
			continue
		}
		ftp := ftpPoint.Value
		ftpUsed = ftp

		streams, err := h.streams.GetByActivityID(r.Context(), id)
		if err != nil {
			continue
		}
		var wattsRaw []byte
		for _, s := range streams {
			if s.StreamType == "watts" {
				wattsRaw = s.Data
				break
			}
		}
		watts, err := storage.DecodeFloat64Array(wattsRaw)
		if err != nil {
			continue
		}

		for _, wv := range watts {
			if wv <= 0 {
				continue
			}
			p := wv / ftp
			for i, b := range bounds {
				if p <= b {
					secondsByZone[i]++
					totalSeconds++
					break
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, PowerZonesResponse{
		FTPWatts:      ftpUsed,
		SecondsByZone: secondsByZone,
		TotalSeconds:  totalSeconds,
		Bounds:        bounds[:4],
	})
}

type TrainingLoadResponse struct {
	Series  []storage.DailyTrainingLoadPoint `json:"series"`
	Summary *storage.DailyTrainingLoadPoint  `json:"summary,omitempty"`
}

// GetTrainingLoad handles GET /api/v1/stats/training-load
func (h *StatsHandler) GetTrainingLoad(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var after *time.Time
	var before *time.Time
	if s := r.URL.Query().Get("after"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			after = &t
		}
	}
	if s := r.URL.Query().Get("before"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			before = &t
		}
	}
	if after == nil && before == nil {
		// Default to last 180 days.
		t := time.Now().AddDate(0, 0, -180)
		after = &t
	}

	if err := h.trainingLoad.EnsureComputedForRange(r.Context(), athlete.ID, after, before); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to compute training load"})
		return
	}

	series, err := h.trainingLoad.GetDailySeries(r.Context(), athlete.ID, after, before)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load training load"})
		return
	}
	summary, _ := h.trainingLoad.GetSummary(r.Context(), athlete.ID)

	writeJSON(w, http.StatusOK, TrainingLoadResponse{
		Series:  series,
		Summary: summary,
	})
}

func sportGroup(sportType string) string {
	if strings.Contains(sportType, "Ride") || sportType == "Handcycle" {
		return "Ride"
	}
	if strings.Contains(sportType, "Run") {
		return "Run"
	}
	return "All"
}

type DistributionSlice struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetDaytimeDistribution handles GET /api/v1/stats/daytime
func (h *StatsHandler) GetDaytimeDistribution(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	query := `
		SELECT
			CASE
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 5 AND 11 THEN 'Morning'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 12 AND 16 THEN 'Afternoon'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 17 AND 21 THEN 'Evening'
				ELSE 'Night'
			END AS bucket,
			COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{athlete.ID}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + storage.JoinStrings(placeholders, ",") + ")"
	}
	query += " GROUP BY bucket"

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to query distribution"})
		return
	}
	defer func() { _ = rows.Close() }()

	counts := map[string]int{}
	for rows.Next() {
		var bucket string
		var count int
		if err := rows.Scan(&bucket, &count); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to scan distribution"})
			return
		}
		counts[bucket] = count
	}

	order := []string{"Morning", "Afternoon", "Evening", "Night"}
	out := make([]DistributionSlice, 0, len(order))
	for _, k := range order {
		out = append(out, DistributionSlice{Label: k, Count: counts[k]})
	}

	writeJSON(w, http.StatusOK, out)
}

// GetWeekdayDistribution handles GET /api/v1/stats/weekday
func (h *StatsHandler) GetWeekdayDistribution(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	query := `
		SELECT CAST(strftime('%w', start_date_local) AS INTEGER) AS weekday, COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{athlete.ID}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + storage.JoinStrings(placeholders, ",") + ")"
	}
	query += " GROUP BY weekday"

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to query distribution"})
		return
	}
	defer func() { _ = rows.Close() }()

	counts := map[int]int{}
	for rows.Next() {
		var weekday int
		var count int
		if err := rows.Scan(&weekday, &count); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to scan distribution"})
			return
		}
		counts[weekday] = count
	}

	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := make([]DistributionSlice, 0, 7)
	for i, n := range names {
		out = append(out, DistributionSlice{Label: n, Count: counts[i]})
	}

	writeJSON(w, http.StatusOK, out)
}

type ZonesHandler struct {
	zones  *storage.ZonesRepository
	strava *strava.Client
}

func NewZonesHandler(zones *storage.ZonesRepository, stravaClient *strava.Client) *ZonesHandler {
	return &ZonesHandler{zones: zones, strava: stravaClient}
}

func (h *ZonesHandler) ListHR(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	defs, err := h.zones.ListHR(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load zone definitions"})
		return
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *ZonesHandler) UpsertHR(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var def storage.HRZoneDefinition
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	if err := h.zones.UpsertHR(r.Context(), athlete.ID, def); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ZonesHandler) DeleteHR(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	sportType := r.URL.Query().Get("sport_type")
	effectiveFrom := r.URL.Query().Get("effective_from")
	if err := h.zones.DeleteHR(r.Context(), athlete.ID, sportType, effectiveFrom); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
