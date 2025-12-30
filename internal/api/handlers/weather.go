package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
	"github.com/melonamin/quantlete/internal/weather"
)

// WeatherHandler handles weather-related API requests.
type WeatherHandler struct {
	weatherRepo    *storage.WeatherRepository
	activityRepo   *storage.ActivityRepository
	streamsRepo    *storage.StreamRepository
	weatherService *weather.Service
	strava         *strava.Client
	logger         *slog.Logger
}

// NewWeatherHandler creates a new weather handler.
func NewWeatherHandler(
	weatherRepo *storage.WeatherRepository,
	activityRepo *storage.ActivityRepository,
	streamsRepo *storage.StreamRepository,
	stravaClient *strava.Client,
	logger *slog.Logger,
) *WeatherHandler {
	return &WeatherHandler{
		weatherRepo:    weatherRepo,
		activityRepo:   activityRepo,
		streamsRepo:    streamsRepo,
		weatherService: weather.NewService(),
		strava:         stravaClient,
		logger:         logger,
	}
}

// GetActivityWeather returns weather data for an activity.
// GET /api/v1/activities/{id}/weather
func (h *WeatherHandler) GetActivityWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	activityIDStr := chi.URLParam(r, "id")
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid activity ID"})
		return
	}

	// Check cache first
	cached, err := h.weatherRepo.GetByActivityID(ctx, activityID)
	if err != nil {
		h.logger.Error("failed to get cached weather", "error", err, "activity_id", activityID)
	}
	if cached != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cached)
		return
	}

	// Get activity to verify ownership and get location/time
	activity, err := h.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		h.logger.Error("failed to get activity", "error", err, "activity_id", activityID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch activity"})
		return
	}
	if activity == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "activity not found"})
		return
	}

	// Verify the activity belongs to the authenticated athlete
	if activity.AthleteID != athlete.ID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	// Try Strava temp stream first
	weatherData := h.tryStravaStream(ctx, activityID)

	// Fall back to Open-Meteo if no Strava data
	if weatherData == nil {
		weatherData = h.tryOpenMeteo(ctx, activity)
	}

	if weatherData == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "weather data not available"})
		return
	}

	// Cache the result
	weatherData.ActivityID = activityID
	if err := h.weatherRepo.Upsert(ctx, weatherData); err != nil {
		h.logger.Error("failed to cache weather", "error", err, "activity_id", activityID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(weatherData)
}

func (h *WeatherHandler) tryStravaStream(ctx context.Context, activityID int64) *weather.ActivityWeather {
	streams, err := h.streamsRepo.GetByActivityID(ctx, activityID)
	if err != nil {
		return nil
	}

	for _, stream := range streams {
		if stream.StreamType == "temp" {
			weatherData, err := h.weatherService.ComputeFromTempStream(string(stream.Data))
			if err != nil {
				h.logger.Debug("failed to compute weather from stream", "error", err)
				return nil
			}
			return weatherData
		}
	}

	return nil
}

func (h *WeatherHandler) tryOpenMeteo(ctx context.Context, activity *storage.Activity) *weather.ActivityWeather {
	// Need start coordinates
	if activity.StartLat == nil || activity.StartLng == nil {
		return nil
	}
	if *activity.StartLat == 0 && *activity.StartLng == 0 {
		return nil
	}

	// Check if activity has a valid start time
	if activity.StartDate.IsZero() {
		return nil
	}
	activityTime := activity.StartDate.Time

	// Don't fetch weather for very old activities (Open-Meteo archive has limits)
	if time.Since(activityTime) > 365*24*time.Hour*2 {
		return nil
	}

	weatherData, err := h.weatherService.FetchFromOpenMeteo(ctx, *activity.StartLat, *activity.StartLng, activityTime)
	if err != nil {
		h.logger.Debug("failed to fetch Open-Meteo data", "error", err, "activity_id", activity.ID)
		return nil
	}

	return weatherData
}
