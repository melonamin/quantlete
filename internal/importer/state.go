// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sasha/stata/internal/storage"
)

const importStateKey = "import_state"

// ImportPhase represents the current phase of import.
type ImportPhase string

const (
	PhaseIdle            ImportPhase = "idle"
	PhaseActivities      ImportPhase = "activities"
	PhaseGear            ImportPhase = "gear"
	PhaseStreams         ImportPhase = "streams"
	PhaseActivityDetails ImportPhase = "activity_details"
	PhaseSegmentDetails  ImportPhase = "segment_details"
	PhasePhotos          ImportPhase = "photos"
	PhaseCompleted       ImportPhase = "completed"
)

// ImportState tracks persistent import progress for resume capability.
type ImportState struct {
	// Current phase
	Phase ImportPhase `json:"phase"`

	// All activity IDs discovered during Phase 1
	ActivityIDs []int64 `json:"activity_ids,omitempty"`

	// Unique gear IDs to fetch
	GearIDs []string `json:"gear_ids,omitempty"`

	// Segment IDs needing detail fetch (collected during activity details phase)
	SegmentIDsToFetch []int64 `json:"segment_ids_to_fetch,omitempty"`

	// Progress tracking per phase
	ActivitiesLastPage   int `json:"activities_last_page"`
	GearLastIndex        int `json:"gear_last_index"`
	StreamsLastIndex     int `json:"streams_last_index"`
	DetailsLastIndex     int `json:"details_last_index"`
	SegmentsLastIndex    int `json:"segments_last_index"`
	PhotosLastIndex      int `json:"photos_last_index"`

	// Counters for display
	ActivitiesTotal int `json:"activities_total"`
	ActivitiesDone  int `json:"activities_done"`
	GearTotal       int `json:"gear_total"`
	GearDone        int `json:"gear_done"`
	StreamsTotal    int `json:"streams_total"`
	StreamsDone     int `json:"streams_done"`
	DetailsTotal    int `json:"details_total"`
	DetailsDone     int `json:"details_done"`
	SegmentsTotal   int `json:"segments_total"`
	SegmentsDone    int `json:"segments_done"`
	PhotosTotal     int `json:"photos_total"`
	PhotosDone      int `json:"photos_done"`

	// Import options (to know which phases to run)
	SkipStreams     bool `json:"skip_streams"`
	SkipSegments    bool `json:"skip_segments"`
	SkipBestEfforts bool `json:"skip_best_efforts"`
	SkipPhotos      bool `json:"skip_photos"`

	// Timestamps
	StartedAt time.Time `json:"started_at"`

	// Error tracking
	FailedCount int `json:"failed_count"`
}

// EstimatedAPICalls returns the estimated total API calls for the import.
func (s *ImportState) EstimatedAPICalls() int {
	calls := 0

	// Activities: ~N/100 pages (already have activities, so just count remaining)
	if s.Phase == PhaseActivities || s.ActivitiesTotal == 0 {
		// Unknown, estimate based on typical user
		calls += 10 // ~1000 activities / 100 per page
	}

	n := s.ActivitiesTotal
	if n == 0 {
		n = len(s.ActivityIDs)
	}

	// Gear: one per unique gear
	calls += s.GearTotal - s.GearDone

	// Streams: one per activity
	if !s.SkipStreams {
		calls += n - s.StreamsDone
	}

	// Activity details: one per activity (for segments + best efforts)
	if !s.SkipBestEfforts || !s.SkipSegments {
		calls += n - s.DetailsDone
	}

	// Segment details: one per unique segment
	if !s.SkipSegments {
		calls += s.SegmentsTotal - s.SegmentsDone
	}

	// Photos: one per activity
	if !s.SkipPhotos {
		calls += n - s.PhotosDone
	}

	return calls
}

// RemainingAPICalls returns API calls remaining from current position.
func (s *ImportState) RemainingAPICalls() int {
	calls := 0
	n := len(s.ActivityIDs)

	switch s.Phase {
	case PhaseActivities:
		// All remaining phases
		calls += 10 // Estimate for remaining activity pages
		calls += len(s.GearIDs)
		if !s.SkipStreams {
			calls += n
		}
		if !s.SkipBestEfforts || !s.SkipSegments {
			calls += n
		}
		if !s.SkipSegments {
			calls += len(s.SegmentIDsToFetch)
		}
		if !s.SkipPhotos {
			calls += n
		}
	case PhaseGear:
		calls += len(s.GearIDs) - s.GearLastIndex
		if !s.SkipStreams {
			calls += n
		}
		if !s.SkipBestEfforts || !s.SkipSegments {
			calls += n
		}
		if !s.SkipSegments {
			calls += len(s.SegmentIDsToFetch)
		}
		if !s.SkipPhotos {
			calls += n
		}
	case PhaseStreams:
		calls += n - s.StreamsLastIndex
		if !s.SkipBestEfforts || !s.SkipSegments {
			calls += n
		}
		if !s.SkipSegments {
			calls += len(s.SegmentIDsToFetch)
		}
		if !s.SkipPhotos {
			calls += n
		}
	case PhaseActivityDetails:
		calls += n - s.DetailsLastIndex
		if !s.SkipSegments {
			calls += len(s.SegmentIDsToFetch)
		}
		if !s.SkipPhotos {
			calls += n
		}
	case PhaseSegmentDetails:
		calls += len(s.SegmentIDsToFetch) - s.SegmentsLastIndex
		if !s.SkipPhotos {
			calls += n
		}
	case PhasePhotos:
		calls += n - s.PhotosLastIndex
	}

	return calls
}

// StateManager handles import state persistence.
type StateManager struct {
	appState *storage.AppStateRepository
}

// NewStateManager creates a new state manager.
func NewStateManager(appState *storage.AppStateRepository) *StateManager {
	return &StateManager{appState: appState}
}

// Load loads the import state from storage.
func (m *StateManager) Load(ctx context.Context) (*ImportState, error) {
	if m.appState == nil {
		return &ImportState{Phase: PhaseIdle}, nil
	}

	data, err := m.appState.Get(ctx, importStateKey)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return &ImportState{Phase: PhaseIdle}, nil
	}

	var state ImportState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save saves the import state to storage.
func (m *StateManager) Save(ctx context.Context, state *ImportState) error {
	if m.appState == nil {
		return nil
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return m.appState.Set(ctx, importStateKey, string(data))
}

// Clear clears the import state.
func (m *StateManager) Clear(ctx context.Context) error {
	if m.appState == nil {
		return nil
	}
	return m.appState.Delete(ctx, importStateKey)
}
