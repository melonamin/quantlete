package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SyncRun represents a single import/sync run.
type SyncRun struct {
	ID        int64      `json:"id"`
	AthleteID int64      `json:"athlete_id"`
	StartedAt time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationSeconds *int  `json:"duration_seconds,omitempty"`

	Status string `json:"status"` // running, completed, failed, canceled
	Error  string `json:"error,omitempty"`

	// Counts
	ActivitiesTotal    int `json:"activities_total"`
	ActivitiesImported int `json:"activities_imported"`
	ActivitiesSkipped  int `json:"activities_skipped"`
	GearImported       int `json:"gear_imported"`
	StreamsImported    int `json:"streams_imported"`
	SegmentsImported   int `json:"segments_imported"`
	PhotosImported     int `json:"photos_imported"`
	FailedCount        int `json:"failed_count"`

	// Options
	FullSync        bool `json:"full_sync"`
	SkipStreams     bool `json:"skip_streams"`
	SkipSegments    bool `json:"skip_segments"`
	SkipBestEfforts bool `json:"skip_best_efforts"`
	SkipPhotos      bool `json:"skip_photos"`

	// Watermark
	NewestActivityDate *time.Time `json:"newest_activity_date,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// SyncWatermark stores the sync watermark for incremental syncs.
type SyncWatermark struct {
	LastSyncedAt       time.Time  `json:"last_synced_at"`
	NewestActivityDate *time.Time `json:"newest_activity_date,omitempty"`
}

// SyncHistoryRepository provides access to sync history storage.
type SyncHistoryRepository struct {
	db       *DB
	appState *AppStateRepository
}

// NewSyncHistoryRepository creates a new sync history repository.
func NewSyncHistoryRepository(db *DB, appState *AppStateRepository) *SyncHistoryRepository {
	return &SyncHistoryRepository{db: db, appState: appState}
}

// StartRun creates a new sync run record with status "running".
func (r *SyncHistoryRepository) StartRun(ctx context.Context, athleteID int64, opts SyncRunOptions) (*SyncRun, error) {
	now := time.Now()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sync_history (
			athlete_id, started_at, status,
			full_sync, skip_streams, skip_segments, skip_best_efforts, skip_photos
		) VALUES (?, ?, 'running', ?, ?, ?, ?, ?)
	`, athleteID, SQLiteTime{Time: now},
		boolToInt(opts.FullSync), boolToInt(opts.SkipStreams), boolToInt(opts.SkipSegments),
		boolToInt(opts.SkipBestEfforts), boolToInt(opts.SkipPhotos))
	if err != nil {
		return nil, fmt.Errorf("inserting sync run: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &SyncRun{
		ID:              id,
		AthleteID:       athleteID,
		StartedAt:       now,
		Status:          "running",
		FullSync:        opts.FullSync,
		SkipStreams:     opts.SkipStreams,
		SkipSegments:    opts.SkipSegments,
		SkipBestEfforts: opts.SkipBestEfforts,
		SkipPhotos:      opts.SkipPhotos,
		CreatedAt:       now,
	}, nil
}

// SyncRunOptions contains options for a sync run.
type SyncRunOptions struct {
	FullSync        bool
	SkipStreams     bool
	SkipSegments    bool
	SkipBestEfforts bool
	SkipPhotos      bool
}

// CompleteRun marks a sync run as completed with final counts.
func (r *SyncHistoryRepository) CompleteRun(ctx context.Context, runID int64, counts SyncRunCounts) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE sync_history SET
			completed_at = ?,
			duration_seconds = (
				SELECT CAST(strftime('%s', ?) - strftime('%s', started_at) AS INTEGER)
				FROM sync_history WHERE id = ?
			),
			status = 'completed',
			activities_total = ?,
			activities_imported = ?,
			activities_skipped = ?,
			gear_imported = ?,
			streams_imported = ?,
			segments_imported = ?,
			photos_imported = ?,
			failed_count = ?,
			newest_activity_date = ?
		WHERE id = ?
	`, SQLiteTime{Time: now}, SQLiteTime{Time: now}, runID,
		counts.ActivitiesTotal, counts.ActivitiesImported, counts.ActivitiesSkipped,
		counts.GearImported, counts.StreamsImported, counts.SegmentsImported,
		counts.PhotosImported, counts.FailedCount,
		nullableSQLiteTime(counts.NewestActivityDate), runID)
	return err
}

// SyncRunCounts contains the final counts for a sync run.
type SyncRunCounts struct {
	ActivitiesTotal    int
	ActivitiesImported int
	ActivitiesSkipped  int
	GearImported       int
	StreamsImported    int
	SegmentsImported   int
	PhotosImported     int
	FailedCount        int
	NewestActivityDate *time.Time
}

// FailRun marks a sync run as failed with an error message.
func (r *SyncHistoryRepository) FailRun(ctx context.Context, runID int64, errMsg string, counts SyncRunCounts) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE sync_history SET
			completed_at = ?,
			duration_seconds = (
				SELECT CAST(strftime('%s', ?) - strftime('%s', started_at) AS INTEGER)
				FROM sync_history WHERE id = ?
			),
			status = 'failed',
			error = ?,
			activities_total = ?,
			activities_imported = ?,
			activities_skipped = ?,
			gear_imported = ?,
			streams_imported = ?,
			segments_imported = ?,
			photos_imported = ?,
			failed_count = ?
		WHERE id = ?
	`, SQLiteTime{Time: now}, SQLiteTime{Time: now}, runID, errMsg,
		counts.ActivitiesTotal, counts.ActivitiesImported, counts.ActivitiesSkipped,
		counts.GearImported, counts.StreamsImported, counts.SegmentsImported,
		counts.PhotosImported, counts.FailedCount, runID)
	return err
}

// CancelRun marks a sync run as canceled.
func (r *SyncHistoryRepository) CancelRun(ctx context.Context, runID int64, counts SyncRunCounts) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE sync_history SET
			completed_at = ?,
			duration_seconds = (
				SELECT CAST(strftime('%s', ?) - strftime('%s', started_at) AS INTEGER)
				FROM sync_history WHERE id = ?
			),
			status = 'canceled',
			activities_total = ?,
			activities_imported = ?,
			activities_skipped = ?,
			gear_imported = ?,
			streams_imported = ?,
			segments_imported = ?,
			photos_imported = ?,
			failed_count = ?
		WHERE id = ?
	`, SQLiteTime{Time: now}, SQLiteTime{Time: now}, runID,
		counts.ActivitiesTotal, counts.ActivitiesImported, counts.ActivitiesSkipped,
		counts.GearImported, counts.StreamsImported, counts.SegmentsImported,
		counts.PhotosImported, counts.FailedCount, runID)
	return err
}

// GetLatest returns the most recent sync runs for an athlete.
func (r *SyncHistoryRepository) GetLatest(ctx context.Context, athleteID int64, limit int) ([]SyncRun, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, athlete_id, started_at, completed_at, duration_seconds,
			status, error,
			activities_total, activities_imported, activities_skipped,
			gear_imported, streams_imported, segments_imported, photos_imported, failed_count,
			full_sync, skip_streams, skip_segments, skip_best_efforts, skip_photos,
			newest_activity_date, created_at
		FROM sync_history
		WHERE athlete_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, athleteID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var runs []SyncRun
	for rows.Next() {
		var run SyncRun
		var startedAt, completedAt, newestActivityDate, createdAt SQLiteTime
		var fullSync, skipStreams, skipSegments, skipBestEfforts, skipPhotos int

		err := rows.Scan(
			&run.ID, &run.AthleteID, &startedAt, &completedAt, &run.DurationSeconds,
			&run.Status, &run.Error,
			&run.ActivitiesTotal, &run.ActivitiesImported, &run.ActivitiesSkipped,
			&run.GearImported, &run.StreamsImported, &run.SegmentsImported,
			&run.PhotosImported, &run.FailedCount,
			&fullSync, &skipStreams, &skipSegments, &skipBestEfforts, &skipPhotos,
			&newestActivityDate, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		run.StartedAt = startedAt.Time
		if !completedAt.Time.IsZero() {
			run.CompletedAt = &completedAt.Time
		}
		if !newestActivityDate.Time.IsZero() {
			run.NewestActivityDate = &newestActivityDate.Time
		}
		run.CreatedAt = createdAt.Time

		run.FullSync = fullSync != 0
		run.SkipStreams = skipStreams != 0
		run.SkipSegments = skipSegments != 0
		run.SkipBestEfforts = skipBestEfforts != 0
		run.SkipPhotos = skipPhotos != 0

		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// GetWatermark returns the sync watermark for an athlete.
func (r *SyncHistoryRepository) GetWatermark(ctx context.Context, athleteID int64) (*SyncWatermark, error) {
	if r.appState == nil {
		return nil, nil
	}

	key := fmt.Sprintf("sync_watermark:%d", athleteID)
	data, err := r.appState.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}

	var wm SyncWatermark
	if err := json.Unmarshal([]byte(data), &wm); err != nil {
		return nil, err
	}

	return &wm, nil
}

// SetWatermark sets the sync watermark for an athlete.
func (r *SyncHistoryRepository) SetWatermark(ctx context.Context, athleteID int64, wm *SyncWatermark) error {
	if r.appState == nil {
		return nil
	}

	key := fmt.Sprintf("sync_watermark:%d", athleteID)
	data, err := json.Marshal(wm)
	if err != nil {
		return err
	}

	return r.appState.Set(ctx, key, string(data))
}

// Helper functions
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableSQLiteTime(t *time.Time) interface{} {
	if t == nil || t.IsZero() {
		return nil
	}
	return SQLiteTime{Time: *t}
}
