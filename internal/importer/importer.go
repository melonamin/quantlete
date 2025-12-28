// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// Status represents the current import status.
type Status string

const (
	StatusIdle      Status = "idle"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Progress tracks import progress.
type Progress struct {
	Status          Status    `json:"status"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	TotalActivities int       `json:"total_activities"`
	ImportedCount   int       `json:"imported_count"`
	SkippedCount    int       `json:"skipped_count"`
	FailedCount     int       `json:"failed_count"`
	CurrentPage     int       `json:"current_page"`
	Error           string    `json:"error,omitempty"`
}

// Importer orchestrates data import from Strava.
type Importer struct {
	stravaClient *strava.Client
	activities   *storage.ActivityRepository
	athletes     *storage.AthleteRepository
	tokens       *storage.TokenRepository
	gear         *storage.GearRepository
	streams      *storage.StreamRepository
	segments     *storage.SegmentRepository
	bestEfforts  *storage.BestEffortsRepository
	maintenance  *storage.MaintenanceRepository
	photos       *storage.PhotoRepository

	segmentCacheMu sync.Mutex
	segmentCache   map[int64]*strava.Segment

	mu       sync.RWMutex
	progress Progress
	cancel   context.CancelFunc
}

// New creates a new importer.
func New(
	stravaClient *strava.Client,
	activities *storage.ActivityRepository,
	athletes *storage.AthleteRepository,
	tokens *storage.TokenRepository,
	gear *storage.GearRepository,
	streams *storage.StreamRepository,
	segments *storage.SegmentRepository,
	bestEfforts *storage.BestEffortsRepository,
	maintenance *storage.MaintenanceRepository,
	photos *storage.PhotoRepository,
) *Importer {
	return &Importer{
		stravaClient: stravaClient,
		activities:   activities,
		athletes:     athletes,
		tokens:       tokens,
		gear:         gear,
		streams:      streams,
		segments:     segments,
		bestEfforts:  bestEfforts,
		maintenance:  maintenance,
		photos:       photos,
		segmentCache: make(map[int64]*strava.Segment),
		progress:     Progress{Status: StatusIdle},
	}
}

// ImportOptions configures an import run.
type ImportOptions struct {
	FullSync           bool // If true, re-import all activities
	IncludeStreams     bool // If true, also import stream data
	IncludeSegments    bool // If true, import segments and segment efforts
	IncludeBestEfforts bool // If true, import Strava best efforts (PRs)
	IncludePhotos      bool // If true, import activity photos
}

// Start begins an import operation.
func (i *Importer) Start(ctx context.Context, opts ImportOptions) error {
	i.mu.Lock()
	if i.progress.Status == StatusRunning {
		i.mu.Unlock()
		return fmt.Errorf("import already in progress")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	i.cancel = cancel
	i.progress = Progress{
		Status:    StatusRunning,
		StartedAt: time.Now(),
	}
	i.mu.Unlock()

	// Run import in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("import panicked", "panic", r)
				i.mu.Lock()
				i.progress.Status = StatusFailed
				i.progress.Error = fmt.Sprintf("panic: %v", r)
				i.progress.CompletedAt = time.Now()
				i.mu.Unlock()
			}
		}()

		err := i.runImport(ctx, opts)
		i.mu.Lock()
		defer i.mu.Unlock()

		i.progress.CompletedAt = time.Now()
		if err != nil {
			slog.Error("import failed", "error", err)
			if ctx.Err() == context.Canceled {
				i.progress.Status = StatusCanceled
			} else {
				i.progress.Status = StatusFailed
				i.progress.Error = err.Error()
			}
		} else {
			i.progress.Status = StatusCompleted
		}
	}()

	return nil
}

// Cancel cancels a running import.
func (i *Importer) Cancel() {
	i.mu.RLock()
	cancel := i.cancel
	i.mu.RUnlock()

	if cancel != nil {
		cancel()
	}
}

// Progress returns the current import progress.
func (i *Importer) Progress() Progress {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.progress
}

// runImport performs the actual import.
func (i *Importer) runImport(ctx context.Context, opts ImportOptions) error {
	slog.Info("starting import",
		"full_sync", opts.FullSync,
		"include_streams", opts.IncludeStreams,
		"include_segments", opts.IncludeSegments,
		"include_best_efforts", opts.IncludeBestEfforts,
		"include_photos", opts.IncludePhotos,
	)

	stravaAthlete := i.stravaClient.GetAthlete()
	if stravaAthlete == nil {
		slog.Error("import failed: not authenticated")
		return fmt.Errorf("not authenticated")
	}
	slog.Info("import authenticated", "athlete_id", stravaAthlete.ID)

	// Save athlete profile
	slog.Info("saving athlete profile")
	athlete := convertAthlete(stravaAthlete)
	athlete.ID = stravaAthlete.ID
	if err := i.athletes.Upsert(ctx, athlete); err != nil {
		slog.Warn("failed to save athlete profile", "error", err)
	}
	slog.Info("athlete profile saved")

	// Import activities page by page
	slog.Info("starting activity fetch loop")
	page := 1
	perPage := 100
	seenIDs := make(map[int64]bool)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		i.mu.Lock()
		i.progress.CurrentPage = page
		i.mu.Unlock()

		slog.Info("fetching activities page", "page", page)

		activities, err := i.stravaClient.GetActivities(ctx, page, perPage)
		if err != nil {
			slog.Error("failed to fetch activities", "page", page, "error", err)
			return fmt.Errorf("fetching activities page %d: %w", page, err)
		}

		slog.Info("fetched activities", "page", page, "count", len(activities))

		if len(activities) == 0 {
			slog.Info("no more activities, import done")
			break
		}

		for _, a := range activities {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if seenIDs[a.ID] {
				continue
			}
			seenIDs[a.ID] = true

			i.mu.Lock()
			i.progress.TotalActivities++
			i.mu.Unlock()

			// Convert and store activity
			if err := i.importActivity(ctx, &a, stravaAthlete.ID, opts); err != nil {
				slog.Warn("failed to import activity", "id", a.ID, "error", err)
				i.mu.Lock()
				i.progress.FailedCount++
				i.mu.Unlock()
				continue
			}

			i.mu.Lock()
			i.progress.ImportedCount++
			i.mu.Unlock()

			// Import gear if referenced
			if a.GearID != "" {
				if err := i.importGear(ctx, a.GearID); err != nil {
					slog.Warn("failed to import gear", "id", a.GearID, "error", err)
				}
			}
		}

		if len(activities) < perPage {
			break
		}
		page++
	}

	slog.Info("import completed",
		"total", i.progress.TotalActivities,
		"imported", i.progress.ImportedCount,
		"failed", i.progress.FailedCount,
	)

	return nil
}

// importActivity imports a single activity.
func (i *Importer) importActivity(ctx context.Context, a *strava.Activity, athleteID int64, opts ImportOptions) error {
	actSource := a
	var detail *strava.Activity

	// Segment/best-efforts/photos are only available on the detailed activity response.
	if opts.IncludeSegments || opts.IncludeBestEfforts {
		var err error
		detail, err = i.stravaClient.GetActivity(ctx, a.ID)
		if err != nil {
			slog.Debug("failed to fetch activity detail", "activity_id", a.ID, "error", err)
		} else {
			actSource = detail
		}
	}

	// Convert Strava activity to storage activity
	act := convertActivity(actSource, athleteID)

	// Hashtag-based custom gear linking (only if Strava gear_id is empty).
	if act.GearID == "" {
		if gearID, err := i.gear.ResolveCustomGearIDFromActivityName(ctx, athleteID, act.Name); err == nil && gearID != "" {
			act.GearID = gearID
		}
	}

	if err := i.activities.Upsert(ctx, act); err != nil {
		return fmt.Errorf("storing activity: %w", err)
	}

	// Optionally import streams
	if opts.IncludeStreams {
		if err := i.importStreams(ctx, a.ID); err != nil {
			slog.Debug("failed to import streams", "activity_id", a.ID, "error", err)
		}
	}

	if opts.IncludeSegments {
		source := detail
		if source == nil {
			source = actSource
		}
		if err := i.importSegments(ctx, source, athleteID); err != nil {
			slog.Debug("failed to import segments", "activity_id", a.ID, "error", err)
		}
	}

	if opts.IncludeBestEfforts {
		source := detail
		if source == nil {
			source = actSource
		}
		if err := i.importBestEfforts(ctx, source, athleteID); err != nil {
			slog.Debug("failed to import best efforts", "activity_id", a.ID, "error", err)
		}
	}

	if opts.IncludePhotos && i.photos != nil {
		if err := i.importPhotos(ctx, act.ID, athleteID); err != nil {
			slog.Debug("failed to import photos", "activity_id", act.ID, "error", err)
		}
	}

	// Hashtag-based maintenance logging.
	if i.maintenance != nil {
		if _, err := i.maintenance.LogFromActivityHashtags(ctx, athleteID, act.ID, act.StartDateLocal, act.Name); err != nil {
			slog.Debug("failed to log maintenance from hashtags", "activity_id", act.ID, "error", err)
		}
	}

	return nil
}

func canonicalBestEffortDistanceType(distanceM float64, name string) (distanceType string, canonicalM float64) {
	// Prefer matching by distance (tolerant to minor rounding).
	type candidate struct {
		Type string
		M    float64
	}
	candidates := []candidate{
		{Type: "400m", M: 400},
		{Type: "0.5mi", M: 804.672},
		{Type: "1k", M: 1000},
		{Type: "1mi", M: 1609.344},
		{Type: "2mi", M: 3218.688},
		{Type: "5k", M: 5000},
		{Type: "10k", M: 10000},
		{Type: "15k", M: 15000},
		{Type: "10mi", M: 16093.44},
		{Type: "20k", M: 20000},
		{Type: "half_marathon", M: 21097.5},
		{Type: "30k", M: 30000},
		{Type: "marathon", M: 42195},
		{Type: "50k", M: 50000},
		{Type: "100k", M: 100000},
	}

	if distanceM > 0 {
		for _, c := range candidates {
			// Accept within 1% or 25m.
			tol := 25.0
			if c.M*0.01 > tol {
				tol = c.M * 0.01
			}
			if distanceM >= c.M-tol && distanceM <= c.M+tol {
				return c.Type, c.M
			}
		}
	}

	// Fallback: best-effort name or rounded meters.
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.ReplaceAll(n, " ", "_")
	n = strings.ReplaceAll(n, "/", "_")
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.Trim(n, "_")
	if n != "" {
		return n, distanceM
	}
	return fmt.Sprintf("m_%d", int(distanceM+0.5)), distanceM
}

func (i *Importer) importBestEfforts(ctx context.Context, a *strava.Activity, athleteID int64) error {
	if i.bestEfforts == nil {
		return nil
	}
	if a == nil || len(a.BestEfforts) == 0 {
		return nil
	}

	var out []storage.BestEffort
	for _, be := range a.BestEfforts {
		if be.ElapsedTime <= 0 || be.Distance <= 0 {
			continue
		}
		dt, canonM := canonicalBestEffortDistanceType(be.Distance, be.Name)
		start := be.StartDate
		moving := be.MovingTime
		out = append(out, storage.BestEffort{
			AthleteID:    athleteID,
			ActivityID:   a.ID,
			SportType:    a.SportType,
			DistanceType: dt,
			Name:         be.Name,
			DistanceM:    canonM,
			ElapsedTimeS: be.ElapsedTime,
			MovingTimeS:  &moving,
			StartIndex:   be.StartIndex,
			EndIndex:     be.EndIndex,
			PRRank:       be.PRRank,
			StartDate:    &start,
		})
	}

	return i.bestEfforts.ReplaceForActivity(ctx, athleteID, a.ID, a.SportType, out)
}

func (i *Importer) getSegment(ctx context.Context, segmentID int64) (*strava.Segment, error) {
	i.segmentCacheMu.Lock()
	if cached, ok := i.segmentCache[segmentID]; ok && cached != nil {
		i.segmentCacheMu.Unlock()
		return cached, nil
	}
	i.segmentCacheMu.Unlock()

	seg, err := i.stravaClient.GetSegment(ctx, segmentID)
	if err != nil {
		return nil, err
	}

	i.segmentCacheMu.Lock()
	i.segmentCache[segmentID] = seg
	i.segmentCacheMu.Unlock()

	return seg, nil
}

func shouldFetchSegmentDetail(seg strava.Segment) bool {
	if seg.Map.Polyline == "" {
		return true
	}
	if seg.AthleteSegmentStats.PRElapsedTime == 0 && seg.AthleteSegmentStats.PRDate == nil && seg.AthleteSegmentStats.EffortCount == 0 && seg.AthleteSegmentStats.KOMRank == nil {
		return true
	}
	return false
}

func ptrFloat64(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ptrInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func floatSliceLatLng(latlng []float64) (lat *float64, lng *float64) {
	if len(latlng) < 2 {
		return nil, nil
	}
	latV := latlng[0]
	lngV := latlng[1]
	return &latV, &lngV
}

func (i *Importer) importSegments(ctx context.Context, a *strava.Activity, athleteID int64) error {
	if i.segments == nil {
		return nil
	}
	if a == nil || len(a.SegmentEfforts) == 0 {
		return nil
	}

	for _, effort := range a.SegmentEfforts {
		seg := effort.Segment
		segDetail := &seg
		if shouldFetchSegmentDetail(seg) {
			if full, err := i.getSegment(ctx, seg.ID); err == nil && full != nil {
				segDetail = full
			}
		}

		startLat, startLng := floatSliceLatLng(segDetail.StartLatlng)
		endLat, endLng := floatSliceLatLng(segDetail.EndLatlng)

		var effortCount *int
		if segDetail.AthleteSegmentStats.EffortCount >= 0 {
			v := segDetail.AthleteSegmentStats.EffortCount
			effortCount = &v
		}
		var prElapsed *int
		if segDetail.AthleteSegmentStats.PRElapsedTime > 0 {
			v := segDetail.AthleteSegmentStats.PRElapsedTime
			prElapsed = &v
		}

		s := &storage.Segment{
			ID:                   segDetail.ID,
			Name:                 segDetail.Name,
			ActivityType:         segDetail.ActivityType,
			Distance:             segDetail.Distance,
			AverageGrade:         segDetail.AverageGrade,
			MaximumGrade:         segDetail.MaximumGrade,
			ElevationHigh:        segDetail.ElevationHigh,
			ElevationLow:         segDetail.ElevationLow,
			ClimbCategory:        segDetail.ClimbCategory,
			StartLat:             startLat,
			StartLng:             startLng,
			EndLat:               endLat,
			EndLng:               endLng,
			Starred:              segDetail.Starred,
			Polyline:             segDetail.Map.Polyline,
			AthleteKOMRank:       segDetail.AthleteSegmentStats.KOMRank,
			AthleteEffortCount:   effortCount,
			AthletePRElapsedTime: prElapsed,
			AthletePRDate:        segDetail.AthleteSegmentStats.PRDate,
		}
		if err := i.segments.UpsertSegment(ctx, s); err != nil {
			return fmt.Errorf("upserting segment %d: %w", segDetail.ID, err)
		}

		startDate := effort.StartDate
		startDateLocal := effort.StartDateLocal
		e := &storage.SegmentEffort{
			ID:               effort.ID,
			SegmentID:        segDetail.ID,
			ActivityID:       a.ID,
			AthleteID:        athleteID,
			Name:             effort.Name,
			ElapsedTime:      effort.ElapsedTime,
			MovingTime:       effort.MovingTime,
			StartDate:        &startDate,
			StartDateLocal:   &startDateLocal,
			Distance:         effort.Distance,
			AverageWatts:     ptrFloat64(effort.AverageWatts),
			AverageHeartrate: ptrFloat64(effort.AverageHeartrate),
			MaxHeartrate:     ptrInt(int(effort.MaxHeartrate)),
			PRRank:           effort.PRRank,
			Country:          a.LocationCountry,
		}
		if err := i.segments.UpsertEffort(ctx, e); err != nil {
			return fmt.Errorf("upserting segment effort %d: %w", effort.ID, err)
		}
	}

	return nil
}

func (i *Importer) importPhotos(ctx context.Context, activityID int64, athleteID int64) error {
	photos, err := i.stravaClient.GetActivityPhotos(ctx, activityID)
	if err != nil {
		return err
	}

	for _, p := range photos {
		url, thumb := bestPhotoURLs(p.URLs)
		if url == "" {
			continue
		}

		id := p.UniqueID
		if id == "" {
			id = strconv.FormatInt(p.ID, 10)
		}

		var loc json.RawMessage
		if len(p.Location) > 0 {
			if b, err := json.Marshal(p.Location); err == nil {
				loc = b
			}
		}

		err := i.photos.Upsert(ctx, &storage.Photo{
			ID:           id,
			AthleteID:    athleteID,
			ActivityID:   activityID,
			URL:          url,
			ThumbnailURL: thumb,
			Caption:      p.Caption,
			Location:     loc,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func bestPhotoURLs(urls map[string]string) (best string, thumb string) {
	if len(urls) == 0 {
		return "", ""
	}
	type kv struct {
		k int
		v string
	}
	var items []kv
	for k, v := range urls {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if n, err := strconv.Atoi(k); err == nil {
			items = append(items, kv{k: n, v: v})
		}
	}
	if len(items) == 0 {
		for _, v := range urls {
			if strings.TrimSpace(v) != "" {
				return v, ""
			}
		}
		return "", ""
	}

	sort.Slice(items, func(i, j int) bool { return items[i].k < items[j].k })
	thumb = items[0].v
	best = items[len(items)-1].v
	return best, thumb
}

// importStreams imports stream data for an activity.
func (i *Importer) importStreams(ctx context.Context, activityID int64) error {
	streams, err := i.stravaClient.GetActivityStreams(ctx, activityID, nil)
	if err != nil {
		return err
	}

	// Store each stream type
	if streams.Time != nil {
		if err := i.storeStream(ctx, activityID, streams.Time); err != nil {
			return err
		}
	}
	if streams.Distance != nil {
		if err := i.storeStream(ctx, activityID, streams.Distance); err != nil {
			return err
		}
	}
	if streams.Altitude != nil {
		if err := i.storeStream(ctx, activityID, streams.Altitude); err != nil {
			return err
		}
	}
	if streams.Heartrate != nil {
		if err := i.storeStream(ctx, activityID, streams.Heartrate); err != nil {
			return err
		}
	}
	if streams.Watts != nil {
		if err := i.storeStream(ctx, activityID, streams.Watts); err != nil {
			return err
		}
	}
	if streams.Cadence != nil {
		if err := i.storeStream(ctx, activityID, streams.Cadence); err != nil {
			return err
		}
	}
	if streams.VelocitySmooth != nil {
		if err := i.storeStream(ctx, activityID, streams.VelocitySmooth); err != nil {
			return err
		}
	}
	if streams.Latlng != nil {
		if err := i.storeStream(ctx, activityID, streams.Latlng); err != nil {
			return err
		}
	}

	return nil
}

// storeStream stores a single stream.
func (i *Importer) storeStream(ctx context.Context, activityID int64, s *strava.Stream) error {
	if s == nil || len(s.Data) == 0 {
		return nil
	}

	stream := &storage.ActivityStream{
		ActivityID:   activityID,
		StreamType:   s.Type,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}

	// Encode data as JSON
	data, err := encodeStreamData(s.Data)
	if err != nil {
		return err
	}
	stream.Data = data

	return i.streams.Upsert(ctx, stream)
}

// importGear imports gear details.
func (i *Importer) importGear(ctx context.Context, gearID string) error {
	gear, err := i.stravaClient.GetGear(ctx, gearID)
	if err != nil {
		return err
	}

	athlete := i.stravaClient.GetAthlete()
	if athlete == nil {
		return fmt.Errorf("not authenticated")
	}

	g := &storage.Gear{
		ID:          gear.ID,
		AthleteID:   athlete.ID,
		Name:        gear.Name,
		Primary:     gear.Primary,
		Retired:     gear.Retired,
		Distance:    gear.Distance,
		BrandName:   gear.BrandName,
		ModelName:   gear.ModelName,
		Description: gear.Description,
	}

	return i.gear.Upsert(ctx, g)
}
