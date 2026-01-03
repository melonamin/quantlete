//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Photos Write
// ============================================================================

//wasm:category Photos - Write

// savePhoto stores a photo in the database
// Called from JS: goStorage.savePhoto(photoJSON)
//wasm:export
func savePhoto(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("savePhoto")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing photo JSON"))
	}

	photoJSON := args[0].String()
	var req struct {
		ID           string `json:"id"`
		AthleteID    int64  `json:"athlete_id"`
		ActivityID   int64  `json:"activity_id"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Caption      string `json:"caption"`
		Location     string `json:"location"` // JSON string of location array
		CreatedAt    string `json:"created_at"`
	}
	if err := json.Unmarshal([]byte(photoJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing photo: %w", err))
	}

	createdAt := time.Now()
	if req.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.CreatedAt); err == nil {
			createdAt = t
		}
	}

	photo := &storage.Photo{
		ID:           req.ID,
		AthleteID:    req.AthleteID,
		ActivityID:   req.ActivityID,
		URL:          req.URL,
		ThumbnailURL: req.ThumbnailURL,
		Caption:      req.Caption,
		CreatedAt:    storage.SQLiteTime{Time: createdAt},
	}

	// Handle location JSON
	if req.Location != "" {
		photo.Location = json.RawMessage(req.Location)
	}

	ctx := context.Background()
	if err := bridge.photos.Upsert(ctx, photo); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Photo %s saved", req.ID))
}

// ============================================================================
// Photos Read
// ============================================================================

//wasm:category Photos - Read

// getPhotos retrieves paginated photos list
// Called from JS: goStorage.getPhotos(filtersJSON)
//wasm:export
func getPhotos(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getPhotos")

	var req struct {
		Page       int      `json:"page"`
		PerPage    int      `json:"per_page"`
		SportTypes []string `json:"sport_types"`
		Country    string   `json:"country"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing photo filters: %w", err))
		}
	}

	ctx := context.Background()
	result, err := bridge.photosService.List(ctx, services.ListPhotosInput{
		AthleteID:  bridge.athleteID,
		SportTypes: req.SportTypes,
		Country:    req.Country,
		Page:       req.Page,
		PerPage:    req.PerPage,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
		"countries":   result.Countries,
		"sport_types": result.SportTypes,
	})
}

// getActivityPhotos retrieves photos for a specific activity
// Called from JS: goStorage.getActivityPhotos(activityId)
//wasm:export
func getActivityPhotos(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getActivityPhotos")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing activity ID"))
	}

	activityID := int64(args[0].Int())
	ctx := context.Background()

	items, err := bridge.photosService.ListByActivity(ctx, services.GetActivityPhotosInput{
		AthleteID:  bridge.athleteID,
		ActivityID: activityID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(items)
}
