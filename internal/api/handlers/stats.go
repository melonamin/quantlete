package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type StatsHandler struct {
	svc              *services.StatsService
	db               *storage.DB
	streams          *storage.StreamRepository
	metrics          *storage.AthleteMetricsRepository
	zones            *storage.ZonesRepository
	zoneDistribution *storage.ZoneDistributionRepository
	strava           *strava.Client
}

func NewStatsHandler(
	svc *services.StatsService,
	db *storage.DB,
	streams *storage.StreamRepository,
	metrics *storage.AthleteMetricsRepository,
	zones *storage.ZonesRepository,
	zoneDistribution *storage.ZoneDistributionRepository,
	stravaClient *strava.Client,
) *StatsHandler {
	return &StatsHandler{
		svc:              svc,
		db:               db,
		streams:          streams,
		metrics:          metrics,
		zones:            zones,
		zoneDistribution: zoneDistribution,
		strava:           stravaClient,
	}
}

// GetHeatmapData handles GET /api/v1/stats/heatmap
func (h *StatsHandler) GetHeatmapData(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()
	input := services.GetHeatmapDataInput{
		AthleteID: athlete.ID,
	}

	// Parse sport types
	if sportTypes := q.Get("sport_type"); sportTypes != "" {
		input.SportTypes = strings.Split(sportTypes, ",")
	}

	// Parse date range
	if t, ok := shared.ParseDateParam(q.Get("after")); ok {
		input.StartAfter = &t
	}
	if t, ok := shared.ParseDateParam(q.Get("before")); ok {
		input.StartBefore = &t
	}

	// Parse commute filter
	if commute := q.Get("commute"); commute != "" {
		v := commute == "true" || commute == "1"
		input.Commute = &v
	}

	// Parse workout type filter
	if wt := q.Get("workout_type"); wt != "" {
		if parsed, err := strconv.Atoi(wt); err == nil {
			input.WorkoutType = &parsed
		}
	}

	// Parse pagination
	if limit := q.Get("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			if parsed > MaxPageSize {
				parsed = MaxPageSize
			}
			input.Limit = parsed
		}
	}
	if offset := q.Get("offset"); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil && parsed >= 0 {
			input.Offset = parsed
		}
	}

	result, err := h.svc.GetHeatmapData(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// GetEddingtonData handles GET /api/v1/stats/eddington
func (h *StatsHandler) GetEddingtonData(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	result, err := h.svc.GetEddingtonData(r.Context(), services.GetEddingtonDataInput{
		AthleteID:  athlete.ID,
		SportTypes: sportTypes,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// GetEddingtonHistory handles GET /api/v1/stats/eddington/history
func (h *StatsHandler) GetEddingtonHistory(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	points, err := h.svc.GetEddingtonHistory(r.Context(), services.GetEddingtonHistoryInput{
		AthleteID:  athlete.ID,
		SportTypes: sportTypes,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, points)
}

// GetBestEfforts handles GET /api/v1/stats/best-efforts
// Returns one all-time PR (fastest elapsed time) per distance_type.
func (h *StatsHandler) GetBestEfforts(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	prs, err := h.svc.GetBestEffortPRs(r.Context(), services.GetBestEffortPRsInput{
		AthleteID:  athlete.ID,
		SportTypes: sportTypes,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, prs)
}

// GetBestEffortsByDistance handles GET /api/v1/stats/best-efforts/{distanceType}
func (h *StatsHandler) GetBestEffortsByDistance(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	distanceType := chi.URLParam(r, "distanceType")
	var sportTypes []string
	if st := r.URL.Query().Get("sport_type"); st != "" {
		sportTypes = strings.Split(st, ",")
	}

	items, err := h.svc.GetBestEffortsForType(r.Context(), services.GetBestEffortsForTypeInput{
		AthleteID:    athlete.ID,
		DistanceType: distanceType,
		SportTypes:   sportTypes,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, items)
}

// GetPowerStats handles GET /api/v1/stats/power
func (h *StatsHandler) GetPowerStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	input := services.GetPowerStatsInput{
		AthleteID: athlete.ID,
	}

	if t, ok := shared.ParseDateParam(r.URL.Query().Get("after")); ok {
		input.After = &t
	}
	if t, ok := shared.ParseDateParam(r.URL.Query().Get("before")); ok {
		input.Before = &t
	}

	if st := r.URL.Query().Get("sport_type"); st != "" {
		input.SportTypes = strings.Split(st, ",")
	}

	result, err := h.svc.GetPowerStats(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
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
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	// Get zone definitions for the response
	defs, err := h.zones.ListHR(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load zone definitions"))
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

	// Find the "All" definition for display, or use the first one
	var method string
	var cfg storage.HRZoneConfig
	for _, def := range defs {
		if def.SportType == "All" {
			method = def.Method
			if err := json.Unmarshal(def.Zones, &cfg); err == nil {
				break
			}
		}
	}
	if method == "" && len(defs) > 0 {
		method = defs[0].Method
		_ = json.Unmarshal(defs[0].Zones, &cfg)
	}

	// Ensure zone distributions are computed
	if h.zoneDistribution != nil {
		if err := h.zoneDistribution.EnsureComputed(r.Context(), athlete.ID); err != nil {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to compute zone distributions"))
			return
		}

		// Get aggregated totals from pre-computed data
		totals, err := h.zoneDistribution.GetTotalDistribution(r.Context(), athlete.ID)
		if err != nil {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to get zone distribution"))
			return
		}

		shared.WriteSuccess(w, HRZonesResponse{
			Method: method,
			Zones:  cfg,
			SecondsByZone: []int{
				totals.SecondsZ1,
				totals.SecondsZ2,
				totals.SecondsZ3,
				totals.SecondsZ4,
				totals.SecondsZ5,
			},
			TotalSeconds: totals.TotalSeconds,
		})
		return
	}

	// Fallback: return zeros if zone distribution not configured
	shared.WriteSuccess(w, HRZonesResponse{
		Method:        method,
		Zones:         cfg,
		SecondsByZone: []int{0, 0, 0, 0, 0},
		TotalSeconds:  0,
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
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var after *time.Time
	var before *time.Time
	if t, ok := shared.ParseDateParam(r.URL.Query().Get("after")); ok {
		after = &t
	}
	if t, ok := shared.ParseDateParam(r.URL.Query().Get("before")); ok {
		before = &t
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
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to query activities"))
		return
	}

	// Collect all activity info first to avoid nested queries with open rows cursor.
	// SQLite with single connection can deadlock if we query while rows are open.
	type activityInfo struct {
		id    int64
		start storage.SQLiteTime
	}
	var activities []activityInfo
	for rows.Next() {
		var a activityInfo
		if err := rows.Scan(&a.id, &a.start); err != nil {
			continue
		}
		activities = append(activities, a)
	}
	_ = rows.Close()

	var ftpUsed float64
	for _, a := range activities {
		ftpPoint, err := h.metrics.LatestBefore(r.Context(), athlete.ID, "ftp_cycling_watts", a.start.Time)
		if err != nil || ftpPoint == nil || ftpPoint.Value <= 0 {
			continue
		}
		ftp := ftpPoint.Value
		ftpUsed = ftp

		streams, err := h.streams.GetByActivityID(r.Context(), a.id)
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

	shared.WriteSuccess(w, PowerZonesResponse{
		FTPWatts:      ftpUsed,
		SecondsByZone: secondsByZone,
		TotalSeconds:  totalSeconds,
		Bounds:        bounds[:4],
	})
}

// GetTrainingLoad handles GET /api/v1/stats/training-load
func (h *StatsHandler) GetTrainingLoad(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	input := services.GetTrainingLoadInput{
		AthleteID: athlete.ID,
	}

	if t, ok := shared.ParseDateParam(r.URL.Query().Get("after")); ok {
		input.After = &t
	}
	if t, ok := shared.ParseDateParam(r.URL.Query().Get("before")); ok {
		input.Before = &t
	}

	result, err := h.svc.GetTrainingLoad(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

type DistributionSlice struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetDaytimeDistribution handles GET /api/v1/stats/daytime
func (h *StatsHandler) GetDaytimeDistribution(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
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
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to query distribution"))
		return
	}
	defer func() { _ = rows.Close() }()

	counts := map[string]int{}
	for rows.Next() {
		var bucket string
		var count int
		if err := rows.Scan(&bucket, &count); err != nil {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to scan distribution"))
			return
		}
		counts[bucket] = count
	}

	order := []string{"Morning", "Afternoon", "Evening", "Night"}
	out := make([]DistributionSlice, 0, len(order))
	for _, k := range order {
		out = append(out, DistributionSlice{Label: k, Count: counts[k]})
	}

	shared.WriteSuccess(w, out)
}

// GetWeekdayDistribution handles GET /api/v1/stats/weekday
func (h *StatsHandler) GetWeekdayDistribution(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
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
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to query distribution"))
		return
	}
	defer func() { _ = rows.Close() }()

	counts := map[int]int{}
	for rows.Next() {
		var weekday int
		var count int
		if err := rows.Scan(&weekday, &count); err != nil {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to scan distribution"))
			return
		}
		counts[weekday] = count
	}

	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := make([]DistributionSlice, 0, 7)
	for i, n := range names {
		out = append(out, DistributionSlice{Label: n, Count: counts[i]})
	}

	shared.WriteSuccess(w, out)
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
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	defs, err := h.zones.ListHR(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load zone definitions"))
		return
	}
	shared.WriteSuccess(w, defs)
}

func (h *ZonesHandler) UpsertHR(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var def storage.HRZoneDefinition
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	if err := h.zones.UpsertHR(r.Context(), athlete.ID, def); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorResponse(err))
		return
	}

	shared.WriteMessage(w, "ok")
}

func (h *ZonesHandler) DeleteHR(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	sportType := r.URL.Query().Get("sport_type")
	effectiveFrom := r.URL.Query().Get("effective_from")
	if err := h.zones.DeleteHR(r.Context(), athlete.ID, sportType, effectiveFrom); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorResponse(err))
		return
	}
	shared.WriteMessage(w, "ok")
}
