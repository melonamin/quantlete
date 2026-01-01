package weather

import "time"

// ActivityWeather represents cached weather data for an activity.
type ActivityWeather struct {
	ActivityID   int64     `json:"activity_id"`
	Source       string    `json:"source"` // "strava" or "open-meteo"
	TemperatureC *float64  `json:"temperature_c,omitempty"`
	FeelsLikeC   *float64  `json:"feels_like_c,omitempty"`
	HumidityPct  *float64  `json:"humidity_percent,omitempty"`
	WindSpeedMps *float64  `json:"wind_speed_mps,omitempty"`
	WindDirDeg   *float64  `json:"wind_direction_deg,omitempty"`
	PrecipMM     *float64  `json:"precipitation_mm,omitempty"`
	WeatherCode  *int      `json:"weather_code,omitempty"`
	TempMinC     *float64  `json:"temp_min_c,omitempty"`
	TempMaxC     *float64  `json:"temp_max_c,omitempty"`
	TempAvgC     *float64  `json:"temp_avg_c,omitempty"`
	TempStream   []float64 `json:"temp_stream,omitempty"`
	FetchedAt    time.Time `json:"fetched_at"`
}

// OpenMeteoRequest represents parameters for an Open-Meteo API request.
type OpenMeteoRequest struct {
	Latitude  float64
	Longitude float64
	StartDate time.Time
	EndDate   time.Time
}

// OpenMeteoResponse represents the response from Open-Meteo Archive API.
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hourly    struct {
		Time          []string  `json:"time"`
		Temperature2M []float64 `json:"temperature_2m"`
		FeelsLike     []float64 `json:"apparent_temperature"`
		Humidity      []float64 `json:"relative_humidity_2m"`
		WindSpeed     []float64 `json:"wind_speed_10m"`
		WindDirection []float64 `json:"wind_direction_10m"`
		Precipitation []float64 `json:"precipitation"`
		WeatherCode   []int     `json:"weather_code"`
	} `json:"hourly"`
}

// WMOCodeDescription maps WMO weather codes to descriptions.
var WMOCodeDescription = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	56: "Light freezing drizzle",
	57: "Dense freezing drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	66: "Light freezing rain",
	67: "Heavy freezing rain",
	71: "Slight snow",
	73: "Moderate snow",
	75: "Heavy snow",
	77: "Snow grains",
	80: "Slight rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	85: "Slight snow showers",
	86: "Heavy snow showers",
	95: "Thunderstorm",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

// GetWMODescription returns a human-readable description for a WMO code.
func GetWMODescription(code int) string {
	if desc, ok := WMOCodeDescription[code]; ok {
		return desc
	}
	return "Unknown"
}
