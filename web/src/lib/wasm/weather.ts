import type { ActivityWeather } from '@/lib/api/weather'

// Open-Meteo Archive API response
interface OpenMeteoResponse {
  latitude: number
  longitude: number
  hourly: {
    time: string[]
    temperature_2m: number[]
    apparent_temperature: number[]
    relative_humidity_2m: number[]
    wind_speed_10m: number[]
    wind_direction_10m: number[]
    precipitation: number[]
    weather_code: number[]
  }
}

// Fetch weather data from Open-Meteo Archive API (CORS-enabled, free)
export async function fetchOpenMeteoWeather(
  lat: number,
  lng: number,
  activityDate: Date
): Promise<ActivityWeather | null> {
  const dateStr = activityDate.toISOString().split('T')[0]

  const params = new URLSearchParams({
    latitude: lat.toFixed(6),
    longitude: lng.toFixed(6),
    start_date: dateStr,
    end_date: dateStr,
    hourly:
      'temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,weather_code',
    timezone: 'auto',
  })

  const url = `https://archive-api.open-meteo.com/v1/archive?${params.toString()}`

  try {
    const response = await fetch(url)
    if (!response.ok) {
      console.warn('Open-Meteo request failed:', response.status)
      return null
    }

    const data: OpenMeteoResponse = await response.json()

    if (!data.hourly || data.hourly.time.length === 0) {
      return null
    }

    // Find the closest hour to the activity start time
    const targetHour = activityDate.getHours()
    const targetDateStr = dateStr

    let bestIdx = 0
    let bestDiff = 24

    for (let i = 0; i < data.hourly.time.length; i++) {
      const timeStr = data.hourly.time[i]
      // Parse "2006-01-02T15:00" format
      const [datePart, timePart] = timeStr.split('T')
      if (datePart !== targetDateStr) continue

      const hour = parseInt(timePart.split(':')[0], 10)
      const diff = Math.abs(hour - targetHour)
      if (diff < bestDiff) {
        bestDiff = diff
        bestIdx = i
      }
    }

    // Calculate min/max from all hourly data
    const temps = data.hourly.temperature_2m
    const minTemp = Math.min(...temps)
    const maxTemp = Math.max(...temps)

    const weather: ActivityWeather = {
      activity_id: 0, // Will be set by caller
      source: 'open-meteo',
      temperature_c: temps[bestIdx],
      temp_avg_c: temps[bestIdx],
      temp_min_c: minTemp,
      temp_max_c: maxTemp,
      feels_like_c: data.hourly.apparent_temperature[bestIdx],
      humidity_percent: data.hourly.relative_humidity_2m[bestIdx],
      wind_speed_mps: data.hourly.wind_speed_10m[bestIdx] / 3.6, // km/h to m/s
      wind_direction_deg: data.hourly.wind_direction_10m[bestIdx],
      precipitation_mm: data.hourly.precipitation[bestIdx],
      weather_code: data.hourly.weather_code[bestIdx],
      fetched_at: new Date().toISOString(),
    }

    return weather
  } catch (err) {
    console.warn('Failed to fetch Open-Meteo weather:', err)
    return null
  }
}

// Compute weather stats from Strava temperature stream data
export function computeFromTempStream(tempData: number[]): ActivityWeather | null {
  if (!tempData || tempData.length === 0) {
    return null
  }

  let minTemp = tempData[0]
  let maxTemp = tempData[0]
  let sum = 0

  for (const t of tempData) {
    if (t < minTemp) minTemp = t
    if (t > maxTemp) maxTemp = t
    sum += t
  }

  const avgTemp = sum / tempData.length

  return {
    activity_id: 0, // Will be set by caller
    source: 'strava',
    temp_min_c: minTemp,
    temp_max_c: maxTemp,
    temp_avg_c: avgTemp,
    temp_stream: tempData,
    fetched_at: new Date().toISOString(),
  }
}
