import type { Activity } from '@/lib/api'
import {
  formatDistance,
  formatDuration,
  formatElevation,
  formatSpeed,
  formatPace,
} from '@/lib/format'
import { getSportCategory } from '@/lib/sport-types'
import { useSettingsStore } from '@/stores'
import { Card, CardContent } from '@/components/ui/card'
import {
  Route,
  Clock,
  Mountain,
  Gauge,
  Heart,
  Zap,
  Flame,
  ThumbsUp,
  MessageCircle,
} from 'lucide-react'

interface ActivityStatsProps {
  activity: Activity
}

interface StatCardProps {
  icon: React.ElementType
  label: string
  value: string
  subValue?: string
}

function StatCard({ icon: Icon, label, value, subValue }: StatCardProps) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start gap-3">
          <div className="rounded-lg bg-muted p-2">
            <Icon className="h-5 w-5 text-muted-foreground" />
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{label}</p>
            <p className="text-2xl font-bold tabular-nums">{value}</p>
            {subValue && <p className="text-sm text-muted-foreground">{subValue}</p>}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function ActivityStats({ activity }: ActivityStatsProps) {
  const { unitSystem } = useSettingsStore()
  const category = getSportCategory(activity.sport_type)
  const isRunning = category === 'run' || category === 'walk'

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <StatCard
        icon={Route}
        label="Distance"
        value={formatDistance(activity.distance, unitSystem)}
      />
      <StatCard
        icon={Clock}
        label="Moving Time"
        value={formatDuration(activity.moving_time)}
        subValue={
          activity.elapsed_time !== activity.moving_time
            ? `${formatDuration(activity.elapsed_time)} elapsed`
            : undefined
        }
      />
      <StatCard
        icon={Mountain}
        label="Elevation"
        value={formatElevation(activity.total_elevation_gain, unitSystem)}
        subValue={
          activity.elev_high && activity.elev_low
            ? `${formatElevation(activity.elev_low, unitSystem)} - ${formatElevation(activity.elev_high, unitSystem)}`
            : undefined
        }
      />
      <StatCard
        icon={Gauge}
        label={isRunning ? 'Pace' : 'Speed'}
        value={
          isRunning
            ? formatPace(activity.average_speed, unitSystem)
            : formatSpeed(activity.average_speed, unitSystem)
        }
        subValue={
          isRunning
            ? `Max ${formatPace(activity.max_speed, unitSystem)}`
            : `Max ${formatSpeed(activity.max_speed, unitSystem)}`
        }
      />

      {activity.average_heartrate && (
        <StatCard
          icon={Heart}
          label="Heart Rate"
          value={`${Math.round(activity.average_heartrate)} bpm`}
          subValue={activity.max_heartrate ? `Max ${activity.max_heartrate} bpm` : undefined}
        />
      )}

      {activity.average_watts && (
        <StatCard
          icon={Zap}
          label="Power"
          value={`${Math.round(activity.average_watts)} W`}
          subValue={activity.max_watts ? `Max ${activity.max_watts} W` : undefined}
        />
      )}

      {activity.calories && (
        <StatCard
          icon={Flame}
          label="Calories"
          value={`${Math.round(activity.calories)}`}
          subValue="kcal"
        />
      )}

      {(activity.kudos_count ?? 0) > 0 && (
        <StatCard icon={ThumbsUp} label="Kudos" value={(activity.kudos_count ?? 0).toString()} />
      )}

      {(activity.comment_count ?? 0) > 0 && (
        <StatCard
          icon={MessageCircle}
          label="Comments"
          value={(activity.comment_count ?? 0).toString()}
        />
      )}
    </div>
  )
}
