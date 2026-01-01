import { Link } from '@tanstack/react-router'
import type { Activity } from '@/lib/api'
import { formatDistance, formatDuration, formatDate, formatElevation } from '@/lib/format'
import { SportIcon } from '@/lib/sport-icon'
import { getSportTextColor, formatSportType } from '@/lib/sport-types'
import { useSettingsStore } from '@/stores'
import { Badge } from '@/components/ui/badge'
import { TableCell, TableRow } from '@/components/ui/table'

interface ActivityRowProps {
  activity: Activity
}

export function ActivityRow({ activity }: ActivityRowProps) {
  const { unitSystem } = useSettingsStore()
  const sportColor = getSportTextColor(activity.sport_type)

  return (
    <TableRow className="hover:bg-muted/50">
      <TableCell className="font-medium">
        <Link
          to="/activities/$activityId"
          params={{ activityId: String(activity.id) }}
          className="hover:underline"
        >
          {activity.name}
        </Link>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <SportIcon sportType={activity.sport_type} className={`h-4 w-4 ${sportColor}`} />
          <span className="text-sm">{formatSportType(activity.sport_type)}</span>
        </div>
      </TableCell>
      <TableCell className="text-muted-foreground">
        {formatDate(activity.start_date_local)}
      </TableCell>
      <TableCell className="text-right tabular-nums">
        {formatDistance(activity.distance, unitSystem)}
      </TableCell>
      <TableCell className="text-right tabular-nums">
        {formatDuration(activity.moving_time)}
      </TableCell>
      <TableCell className="text-right tabular-nums">
        {formatElevation(activity.total_elevation_gain, unitSystem)}
      </TableCell>
      <TableCell>
        <div className="flex gap-1">
          {activity.commute && (
            <Badge variant="secondary" className="text-xs">
              Commute
            </Badge>
          )}
          {activity.trainer && (
            <Badge variant="secondary" className="text-xs">
              Indoor
            </Badge>
          )}
        </div>
      </TableCell>
    </TableRow>
  )
}
