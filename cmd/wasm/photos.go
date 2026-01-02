//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Photos Write
// ============================================================================

// savePhoto stores a photo in the database
// Called from JS: goStorage.savePhoto(photoJSON)
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
	if err := photos.Upsert(ctx, photo); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Photo %s saved", req.ID))
}

// ============================================================================
// Photos Read
// ============================================================================

// getPhotos retrieves paginated photos list
// Called from JS: goStorage.getPhotos(filtersJSON)
func getPhotos(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getPhotos")

	var req struct {
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing photo filters: %w", err))
		}
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage == 0 {
		perPage = 50
	}

	ctx := context.Background()
	result, err := photos.List(ctx, athleteID, storage.PhotoListFilters{}, page, perPage)
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(result.Items))
	for i, p := range result.Items {
		items[i] = photoListItemToMap(p)
	}

	totalPages := (result.Total + perPage - 1) / perPage

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        items,
		"total":       result.Total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	})
}

// getActivityPhotos retrieves photos for a specific activity
// Called from JS: goStorage.getActivityPhotos(activityId)
func getActivityPhotos(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getActivityPhotos")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing activity ID"))
	}

	activityID := int64(args[0].Int())
	ctx := context.Background()

	photoList, err := photos.ListByActivity(ctx, athleteID, activityID)
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(photoList))
	for i, p := range photoList {
		items[i] = photoToMap(p)
	}

	return dataJSON(items)
}

// ============================================================================
// Photo Map Helpers
// ============================================================================

func photoListItemToMap(p storage.PhotoListItem) map[string]interface{} {
	m := map[string]interface{}{
		"id":               p.ID,
		"athlete_id":       p.AthleteID,
		"activity_id":      p.ActivityID,
		"url":              p.URL,
		"thumbnail_url":    p.ThumbnailURL,
		"created_at":       p.CreatedAt.Format(time.RFC3339),
		"activity_name":    p.ActivityName,
		"sport_type":       p.SportType,
		"start_date_local": p.StartDateLocal.Format(time.RFC3339),
	}
	if p.Caption != "" {
		m["caption"] = p.Caption
	}
	if len(p.Location) > 0 {
		m["location"] = string(p.Location)
	}
	if p.LocationCountry != "" {
		m["location_country"] = p.LocationCountry
	}
	return m
}

func photoToMap(p storage.Photo) map[string]interface{} {
	m := map[string]interface{}{
		"id":            p.ID,
		"athlete_id":    p.AthleteID,
		"activity_id":   p.ActivityID,
		"url":           p.URL,
		"thumbnail_url": p.ThumbnailURL,
		"created_at":    p.CreatedAt.Format(time.RFC3339),
	}
	if p.Caption != "" {
		m["caption"] = p.Caption
	}
	if len(p.Location) > 0 {
		m["location"] = string(p.Location)
	}
	return m
}
