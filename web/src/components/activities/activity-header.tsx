import type { Activity } from '@/lib/api'
import { formatDateTime } from '@/lib/format'
import { getSportIcon, getSportTextColor, formatSportType } from '@/lib/sport-types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ExternalLink } from 'lucide-react'

interface ActivityHeaderProps {
  activity: Activity
}

export function ActivityHeader({ activity }: ActivityHeaderProps) {
  const SportIcon = getSportIcon(activity.sport_type)
  const sportColor = getSportTextColor(activity.sport_type)
  const stravaUrl = `https://www.strava.com/activities/${activity.id}`

  return (
    <div className="mb-6">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-4">
          <div className={`rounded-lg bg-muted p-3 ${sportColor}`}>
            <SportIcon className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">{activity.name}</h1>
            <div className="flex items-center gap-2 mt-1 text-muted-foreground">
              <span>{formatSportType(activity.sport_type)}</span>
              <span>•</span>
              <span>{formatDateTime(activity.start_date_local)}</span>
            </div>
            {activity.description && (
              <p className="mt-2 text-muted-foreground">{activity.description}</p>
            )}
            <div className="flex gap-2 mt-3">
              {activity.commute && (
                <Badge variant="secondary">Commute</Badge>
              )}
              {activity.trainer && (
                <Badge variant="secondary">Indoor</Badge>
              )}
              {activity.private && (
                <Badge variant="outline">Private</Badge>
              )}
            </div>
          </div>
        </div>
        <Button variant="outline" size="sm" asChild>
          <a href={stravaUrl} target="_blank" rel="noopener noreferrer">
            <ExternalLink className="h-4 w-4 mr-2" />
            View on Strava
          </a>
        </Button>
      </div>
    </div>
  )
}
