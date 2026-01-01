import { useActivityWeather } from '@/lib/data/hooks'
import {
  formatTemperature,
  formatWindSpeed,
  getWeatherDescription,
  getWindDirection,
} from '@/lib/api/weather'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Sun,
  Cloud,
  CloudSun,
  CloudFog,
  CloudDrizzle,
  CloudRain,
  Snowflake,
  CloudLightning,
  Thermometer,
  Droplets,
  Wind,
  type LucideProps,
} from 'lucide-react'

interface WeatherBadgeProps {
  activityId: number
}

const weatherIcons = {
  sun: Sun,
  cloudSun: CloudSun,
  cloud: Cloud,
  cloudFog: CloudFog,
  cloudDrizzle: CloudDrizzle,
  cloudRain: CloudRain,
  snowflake: Snowflake,
  cloudLightning: CloudLightning,
  thermometer: Thermometer,
} as const

type WeatherIconType = keyof typeof weatherIcons

function getWeatherIconType(code: number | undefined): WeatherIconType {
  if (code === undefined) return 'thermometer'

  // WMO code to icon mapping
  if (code <= 1) return 'sun'
  if (code <= 3) return code === 2 ? 'cloudSun' : 'cloud'
  if (code <= 48) return 'cloudFog'
  if (code <= 57) return 'cloudDrizzle'
  if (code <= 67) return 'cloudRain'
  if (code <= 77) return 'snowflake'
  if (code <= 82) return 'cloudRain'
  if (code <= 86) return 'snowflake'
  if (code >= 95) return 'cloudLightning'

  return 'cloud'
}

function WeatherIcon({ code, ...props }: { code: number | undefined } & LucideProps) {
  const iconType = getWeatherIconType(code)
  const Icon = weatherIcons[iconType]
  return <Icon {...props} />
}

export function WeatherBadge({ activityId }: WeatherBadgeProps) {
  const { data: weather, isLoading } = useActivityWeather(activityId)

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Weather</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <Skeleton className="h-12 w-12 rounded-full" />
            <div className="space-y-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-3 w-32" />
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!weather) {
    return null
  }

  const temp = weather.temperature_c ?? weather.temp_avg_c
  const description = weather.weather_code !== undefined
    ? getWeatherDescription(weather.weather_code)
    : weather.source === 'strava' ? 'From sensor' : 'Unknown'

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">Weather</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            <WeatherIcon code={weather.weather_code} className="h-6 w-6 text-muted-foreground" />
          </div>
          <div>
            <div className="flex items-baseline gap-2">
              {temp !== undefined && (
                <span className="text-2xl font-bold">{formatTemperature(temp)}</span>
              )}
              <span className="text-sm text-muted-foreground">{description}</span>
            </div>
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted-foreground">
              {weather.feels_like_c !== undefined && (
                <span className="flex items-center gap-1">
                  <Thermometer className="h-3 w-3" />
                  Feels {formatTemperature(weather.feels_like_c)}
                </span>
              )}
              {weather.humidity_percent !== undefined && (
                <span className="flex items-center gap-1">
                  <Droplets className="h-3 w-3" />
                  {Math.round(weather.humidity_percent)}%
                </span>
              )}
              {weather.wind_speed_mps !== undefined && (
                <span className="flex items-center gap-1">
                  <Wind className="h-3 w-3" />
                  {formatWindSpeed(weather.wind_speed_mps, 'kmh')}
                  {weather.wind_direction_deg !== undefined && (
                    <span className="text-xs">
                      {getWindDirection(weather.wind_direction_deg)}
                    </span>
                  )}
                </span>
              )}
            </div>
            {(weather.temp_min_c !== undefined || weather.temp_max_c !== undefined) &&
             weather.source === 'strava' && (
              <div className="mt-1 text-xs text-muted-foreground">
                {weather.temp_min_c !== undefined && (
                  <span>Min: {formatTemperature(weather.temp_min_c)}</span>
                )}
                {weather.temp_min_c !== undefined && weather.temp_max_c !== undefined && (
                  <span> / </span>
                )}
                {weather.temp_max_c !== undefined && (
                  <span>Max: {formatTemperature(weather.temp_max_c)}</span>
                )}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
