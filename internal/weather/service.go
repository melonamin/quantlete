package weather

import (
	"context"
	"encoding/json"
	"time"
)

// StreamData represents Strava activity stream data for temperature.
type StreamData struct {
	Data []float64 `json:"data"`
}

// Service provides weather data for activities using hybrid sources.
type Service struct {
	openMeteo *OpenMeteoClient
}

// NewService creates a new weather service.
func NewService() *Service {
	return &Service{
		openMeteo: NewOpenMeteoClient(),
	}
}

// ComputeFromTempStream calculates weather stats from Strava temperature stream data.
func (s *Service) ComputeFromTempStream(tempStreamJSON string) (*ActivityWeather, error) {
	var stream StreamData
	if err := json.Unmarshal([]byte(tempStreamJSON), &stream); err != nil {
		return nil, err
	}

	if len(stream.Data) == 0 {
		return nil, nil
	}

	// Calculate min, max, avg
	minTemp := stream.Data[0]
	maxTemp := stream.Data[0]
	sum := 0.0

	for _, t := range stream.Data {
		if t < minTemp {
			minTemp = t
		}
		if t > maxTemp {
			maxTemp = t
		}
		sum += t
	}

	avgTemp := sum / float64(len(stream.Data))

	return &ActivityWeather{
		Source:     "strava",
		TempMinC:   &minTemp,
		TempMaxC:   &maxTemp,
		TempAvgC:   &avgTemp,
		TempStream: stream.Data,
		FetchedAt:  time.Now(),
	}, nil
}

// FetchFromOpenMeteo fetches weather data from Open-Meteo for an activity.
func (s *Service) FetchFromOpenMeteo(ctx context.Context, lat, lng float64, activityTime time.Time) (*ActivityWeather, error) {
	// Fetch data for the day of the activity
	req := OpenMeteoRequest{
		Latitude:  lat,
		Longitude: lng,
		StartDate: activityTime,
		EndDate:   activityTime,
	}

	resp, err := s.openMeteo.FetchHistorical(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.Hourly.Time) == 0 {
		return nil, nil
	}

	// Find the closest hour to the activity start time
	idx := FindClosestHour(resp.Hourly.Time, activityTime)
	if idx < 0 || idx >= len(resp.Hourly.Temperature2M) {
		return nil, nil
	}

	weather := &ActivityWeather{
		Source:    "open-meteo",
		FetchedAt: time.Now(),
	}

	// Set values from the closest hour
	if idx < len(resp.Hourly.Temperature2M) {
		temp := resp.Hourly.Temperature2M[idx]
		weather.TemperatureC = &temp
		weather.TempAvgC = &temp
	}

	if idx < len(resp.Hourly.FeelsLike) {
		feels := resp.Hourly.FeelsLike[idx]
		weather.FeelsLikeC = &feels
	}

	if idx < len(resp.Hourly.Humidity) {
		hum := resp.Hourly.Humidity[idx]
		weather.HumidityPct = &hum
	}

	if idx < len(resp.Hourly.WindSpeed) {
		wind := resp.Hourly.WindSpeed[idx]
		// Convert km/h to m/s
		windMps := wind / 3.6
		weather.WindSpeedMps = &windMps
	}

	if idx < len(resp.Hourly.WindDirection) {
		dir := resp.Hourly.WindDirection[idx]
		weather.WindDirDeg = &dir
	}

	if idx < len(resp.Hourly.Precipitation) {
		precip := resp.Hourly.Precipitation[idx]
		weather.PrecipMM = &precip
	}

	if idx < len(resp.Hourly.WeatherCode) {
		code := resp.Hourly.WeatherCode[idx]
		weather.WeatherCode = &code
	}

	// Calculate min/max from all hourly data for the day
	if len(resp.Hourly.Temperature2M) > 0 {
		minT := resp.Hourly.Temperature2M[0]
		maxT := resp.Hourly.Temperature2M[0]
		for _, t := range resp.Hourly.Temperature2M {
			if t < minT {
				minT = t
			}
			if t > maxT {
				maxT = t
			}
		}
		weather.TempMinC = &minT
		weather.TempMaxC = &maxT
	}

	return weather, nil
}
