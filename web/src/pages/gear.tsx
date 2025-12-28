import { useState } from 'react'
import { useGear, type Gear } from '@/lib/api'
import { formatDistance } from '@/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Bike, Footprints, Check } from 'lucide-react'

export function GearPage() {
  const [includeRetired, setIncludeRetired] = useState(false)
  const { data: gear, isLoading, error } = useGear(includeRetired)

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Gear</h1>
          <p className="text-muted-foreground">
            Manage your equipment and track usage
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setIncludeRetired(!includeRetired)}
        >
          {includeRetired ? 'Hide Retired' : 'Show Retired'}
        </Button>
      </div>

      {isLoading ? (
        <GearSkeleton />
      ) : error ? (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">Failed to load gear</p>
        </div>
      ) : gear && gear.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {gear.map((item) => (
            <GearCard key={item.id} gear={item} />
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">
            No gear found. Import activities with gear to see them here.
          </p>
        </div>
      )}
    </div>
  )
}

function GearCard({ gear }: { gear: Gear }) {
  const isBike = gear.id.startsWith('b')

  return (
    <Card className={gear.retired ? 'opacity-60' : ''}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            {isBike ? (
              <Bike className="h-5 w-5 text-sport-ride" />
            ) : (
              <Footprints className="h-5 w-5 text-sport-run" />
            )}
            <CardTitle className="text-lg">{gear.name}</CardTitle>
          </div>
          <div className="flex gap-1">
            {gear.primary && (
              <Badge variant="secondary" className="gap-1">
                <Check className="h-3 w-3" />
                Primary
              </Badge>
            )}
            {gear.retired && <Badge variant="outline">Retired</Badge>}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {(gear.brand_name || gear.model_name) && (
            <p className="text-sm text-muted-foreground">
              {[gear.brand_name, gear.model_name].filter(Boolean).join(' ')}
            </p>
          )}
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Distance</span>
            <span className="font-medium">{formatDistance(gear.distance)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Activities</span>
            <span className="font-medium">{gear.activity_count}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function GearSkeleton() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i}>
          <CardHeader className="pb-3">
            <div className="flex items-center gap-2">
              <Skeleton className="h-5 w-5" />
              <Skeleton className="h-5 w-32" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
