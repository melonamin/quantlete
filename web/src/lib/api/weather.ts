// Weather data for an activity
export interface ActivityWeather {
  activity_id: number
  source: 'strava' | 'open-meteo'
  temperature_c?: number
  feels_like_c?: number
  humidity_percent?: number
  wind_speed_mps?: number
  wind_direction_deg?: number
  precipitation_mm?: number
  weather_code?: number
  temp_min_c?: number
  temp_max_c?: number
  temp_avg_c?: number
  temp_stream?: number[]
  fetched_at: string
}

// WMO Weather Codes to description and icon mapping
// https://open-meteo.com/en/docs#weathervariables
export const WMO_CODES: Record<number, { description: string; icon: string }> = {
  0: { description: 'Clear sky', icon: 'sun' },
  1: { description: 'Mainly clear', icon: 'sun' },
  2: { description: 'Partly cloudy', icon: 'cloud-sun' },
  3: { description: 'Overcast', icon: 'cloud' },
  45: { description: 'Fog', icon: 'cloud-fog' },
  48: { description: 'Depositing rime fog', icon: 'cloud-fog' },
  51: { description: 'Light drizzle', icon: 'cloud-drizzle' },
  53: { description: 'Moderate drizzle', icon: 'cloud-drizzle' },
  55: { description: 'Dense drizzle', icon: 'cloud-drizzle' },
  56: { description: 'Light freezing drizzle', icon: 'cloud-drizzle' },
  57: { description: 'Dense freezing drizzle', icon: 'cloud-drizzle' },
  61: { description: 'Slight rain', icon: 'cloud-rain' },
  63: { description: 'Moderate rain', icon: 'cloud-rain' },
  65: { description: 'Heavy rain', icon: 'cloud-rain' },
  66: { description: 'Light freezing rain', icon: 'cloud-rain' },
  67: { description: 'Heavy freezing rain', icon: 'cloud-rain' },
  71: { description: 'Slight snow', icon: 'snowflake' },
  73: { description: 'Moderate snow', icon: 'snowflake' },
  75: { description: 'Heavy snow', icon: 'snowflake' },
  77: { description: 'Snow grains', icon: 'snowflake' },
  80: { description: 'Slight rain showers', icon: 'cloud-rain' },
  81: { description: 'Moderate rain showers', icon: 'cloud-rain' },
  82: { description: 'Violent rain showers', icon: 'cloud-rain' },
  85: { description: 'Slight snow showers', icon: 'snowflake' },
  86: { description: 'Heavy snow showers', icon: 'snowflake' },
  95: { description: 'Thunderstorm', icon: 'cloud-lightning' },
  96: { description: 'Thunderstorm with slight hail', icon: 'cloud-lightning' },
  99: { description: 'Thunderstorm with heavy hail', icon: 'cloud-lightning' },
}

// Get weather description from WMO code
export function getWeatherDescription(code: number): string {
  return WMO_CODES[code]?.description ?? 'Unknown'
}

// Get weather icon name from WMO code
export function getWeatherIcon(code: number): string {
  return WMO_CODES[code]?.icon ?? 'cloud'
}

// Format temperature with unit
export function formatTemperature(tempC: number, unit: 'celsius' | 'fahrenheit' = 'celsius'): string {
  if (unit === 'fahrenheit') {
    const tempF = (tempC * 9) / 5 + 32
    return `${Math.round(tempF)}°F`
  }
  return `${Math.round(tempC)}°C`
}

// Format wind speed with unit
export function formatWindSpeed(mps: number, unit: 'mps' | 'kmh' | 'mph' = 'mps'): string {
  switch (unit) {
    case 'kmh':
      return `${Math.round(mps * 3.6)} km/h`
    case 'mph':
      return `${Math.round(mps * 2.237)} mph`
    default:
      return `${Math.round(mps)} m/s`
  }
}

// Get wind direction as compass point
export function getWindDirection(deg: number): string {
  const directions = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW']
  const index = Math.round(deg / 45) % 8
  return directions[index]
}
