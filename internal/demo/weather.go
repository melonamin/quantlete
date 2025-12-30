package demo

import (
	"math"
	"math/rand"
	"time"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/weather"
)

// WMO weather codes for common conditions.
var commonWeatherCodes = []int{
	0,  // Clear sky
	1,  // Mainly clear
	2,  // Partly cloudy
	3,  // Overcast
	51, // Light drizzle
	61, // Slight rain
	80, // Slight rain showers
}

// Seasonal temperature ranges by month (Celsius, Northern Hemisphere mid-latitudes).
var seasonalTemps = []struct {
	low  float64
	high float64
}{
	{-5, 8},   // January
	{-3, 10},  // February
	{2, 14},   // March
	{6, 18},   // April
	{10, 22},  // May
	{14, 26},  // June
	{16, 30},  // July
	{15, 29},  // August
	{11, 24},  // September
	{6, 18},   // October
	{1, 12},   // November
	{-3, 8},   // December
}

// generateWeather creates realistic weather data for activities.
func generateWeather(rng *rand.Rand, activities []storage.Activity) []weather.ActivityWeather {
	var result []weather.ActivityWeather

	for _, a := range activities {
		// Only outdoor activities get weather
		if isIndoorActivity(a.SportType) {
			continue
		}

		// 90% of outdoor activities have weather data
		if rng.Float64() > 0.90 {
			continue
		}

		w := generateActivityWeather(rng, a)
		result = append(result, w)
	}

	return result
}

func generateActivityWeather(rng *rand.Rand, a storage.Activity) weather.ActivityWeather {
	month := int(a.StartDateLocal.Time.Month()) - 1 // 0-indexed
	temps := seasonalTemps[month]

	// Base temperature for the day
	baseTemp := temps.low + rng.Float64()*(temps.high-temps.low)

	// Morning activities are cooler, afternoon warmer
	hour := a.StartDateLocal.Time.Hour()
	timeAdjust := 0.0
	if hour < 10 {
		timeAdjust = -3.0
	} else if hour > 14 && hour < 18 {
		timeAdjust = 3.0
	}
	temperature := baseTemp + timeAdjust + (rng.Float64()-0.5)*4

	// Feels like (wind chill / heat index effect)
	feelsLike := temperature + (rng.Float64()-0.5)*4

	// Humidity (higher in morning, lower in afternoon)
	humidity := 50.0 + rng.Float64()*30
	if hour < 10 {
		humidity += 15
	} else if hour > 14 {
		humidity -= 10
	}
	humidity = math.Max(30, math.Min(95, humidity))

	// Wind
	windSpeed := rng.Float64() * 8 // 0-8 m/s typically
	if rng.Float64() < 0.1 {       // 10% chance of windy day
		windSpeed = 8 + rng.Float64()*7
	}
	windDirection := rng.Float64() * 360

	// Precipitation (rare, mostly 0)
	precipitation := 0.0
	weatherCode := commonWeatherCodes[rng.Intn(4)] // Mostly clear/cloudy
	if rng.Float64() < 0.15 {                      // 15% chance of rain
		precipitation = rng.Float64() * 5
		weatherCode = commonWeatherCodes[4+rng.Intn(3)] // Drizzle or rain
	}

	// Temperature range during activity
	tempVariation := 2.0 + rng.Float64()*3 // 2-5 degree variation
	tempMin := temperature - tempVariation*0.4
	tempMax := temperature + tempVariation*0.6
	tempAvg := temperature

	return weather.ActivityWeather{
		ActivityID:   a.ID,
		Source:       "demo",
		TemperatureC: ptr(math.Round(temperature*10) / 10),
		FeelsLikeC:   ptr(math.Round(feelsLike*10) / 10),
		HumidityPct:  ptr(math.Round(humidity*10) / 10),
		WindSpeedMps: ptr(math.Round(windSpeed*10) / 10),
		WindDirDeg:   ptr(math.Round(windDirection*10) / 10),
		PrecipMM:     ptr(math.Round(precipitation*10) / 10),
		WeatherCode:  ptrInt(weatherCode),
		TempMinC:     ptr(math.Round(tempMin*10) / 10),
		TempMaxC:     ptr(math.Round(tempMax*10) / 10),
		TempAvgC:     ptr(math.Round(tempAvg*10) / 10),
		FetchedAt:    time.Now(),
	}
}

func isIndoorActivity(sportType string) bool {
	switch sportType {
	case "VirtualRide", "VirtualRun", "Swim", "WeightTraining", "Yoga", "Workout":
		return true
	}
	return false
}

func ptrInt(v int) *int {
	return &v
}
