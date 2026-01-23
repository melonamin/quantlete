import { useState, useEffect, useCallback } from 'react'
import { Search, X, Filter, HelpCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import type { ActivityFilters as Filters } from '@/lib/api'
import { useSettingsStore } from '@/stores/settings'
import {
  distanceToMeters,
  metersToDisplayUnit,
  parseDurationInput,
  formatDurationInput,
  getDistanceUnit,
} from '@/lib/format'

interface ActivityFiltersProps {
  filters: Filters
  onFiltersChange: (filters: Partial<Filters>) => void
  onReset: () => void
}

const SPORT_TYPES = [
  'Ride',
  'VirtualRide',
  'Run',
  'VirtualRun',
  'Walk',
  'Hike',
  'Swim',
  'WeightTraining',
  'Workout',
  'Yoga',
]

// Helper to get initial display value for distance
function getDisplayDistance(meters: number | undefined, unitSystem: 'metric' | 'imperial'): string {
  if (meters === undefined) return ''
  return metersToDisplayUnit(meters, unitSystem).toString()
}

// Helper to get initial display value for duration
function getDisplayDuration(seconds: number | undefined): string {
  if (seconds === undefined) return ''
  return formatDurationInput(seconds)
}

export function ActivityFiltersPanel({ filters, onFiltersChange, onReset }: ActivityFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [localSearch, setLocalSearch] = useState(filters.search ?? '')
  const unitSystem = useSettingsStore((s) => s.unitSystem)

  // Local state for debounced inputs - we use state (not useMemo) because:
  // 1. User types → updates local state (allows typing without immediate filter updates)
  // 2. Debounce timer fires → updates parent filter
  // 3. External changes (reset, unit change) → useEffect syncs back to local state
  // useMemo would make inputs read-only, breaking the debounce pattern
  const [localMinDistance, setLocalMinDistance] = useState(() =>
    getDisplayDistance(filters.min_distance_m, unitSystem)
  )
  const [localMaxDistance, setLocalMaxDistance] = useState(() =>
    getDisplayDistance(filters.max_distance_m, unitSystem)
  )
  const [localMinDuration, setLocalMinDuration] = useState(() =>
    getDisplayDuration(filters.min_duration_s)
  )
  const [localMaxDuration, setLocalMaxDuration] = useState(() =>
    getDisplayDuration(filters.max_duration_s)
  )

  // Sync distance inputs when parent filters change (external reset) or unit system changes.
  // This setState-in-effect is intentional: we need to sync external changes to local state
  // for debounced controlled inputs. Alternative patterns (useMemo, key reset) don't support
  // bidirectional data flow needed for user typing + external sync.
  useEffect(() => {
    const nextMinDistance = getDisplayDistance(filters.min_distance_m, unitSystem)
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLocalMinDistance((prev) => (prev === nextMinDistance ? prev : nextMinDistance))

    const nextMaxDistance = getDisplayDistance(filters.max_distance_m, unitSystem)
    setLocalMaxDistance((prev) => (prev === nextMaxDistance ? prev : nextMaxDistance))
  }, [filters.min_distance_m, filters.max_distance_m, unitSystem])

  // Sync duration inputs when parent filters are cleared externally
  useEffect(() => {
    if (filters.min_duration_s === undefined && localMinDuration !== '') {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setLocalMinDuration('')
    }
    if (filters.max_duration_s === undefined && localMaxDuration !== '') {
      setLocalMaxDuration('')
    }
  }, [
    filters.min_duration_s,
    filters.max_duration_s,
    localMinDuration,
    localMaxDuration,
  ])

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== filters.search) {
        onFiltersChange({ search: localSearch || undefined, page: 1 })
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [localSearch, filters.search, onFiltersChange])

  // Handle distance filter changes with validation
  const handleDistanceChange = useCallback(
    (field: 'min_distance_m' | 'max_distance_m', value: string) => {
      const parsed = parseFloat(value)
      if (value === '' || isNaN(parsed)) {
        onFiltersChange({ [field]: undefined, page: 1 })
        return
      }
      if (parsed < 0) return // Ignore negative values
      const meters = distanceToMeters(parsed, unitSystem)
      onFiltersChange({ [field]: meters, page: 1 })
    },
    [onFiltersChange, unitSystem]
  )

  // Handle duration filter changes with validation
  const handleDurationChange = useCallback(
    (field: 'min_duration_s' | 'max_duration_s', value: string) => {
      const parsed = parseDurationInput(value)
      if (value === '' || parsed === null) {
        onFiltersChange({ [field]: undefined, page: 1 })
        return
      }
      if (parsed < 0) return // Ignore negative values
      onFiltersChange({ [field]: parsed, page: 1 })
    },
    [onFiltersChange]
  )

  // Debounce distance/duration filters
  useEffect(() => {
    const timer = setTimeout(() => {
      handleDistanceChange('min_distance_m', localMinDistance)
    }, 500)
    return () => clearTimeout(timer)
  }, [localMinDistance, handleDistanceChange])

  useEffect(() => {
    const timer = setTimeout(() => {
      handleDistanceChange('max_distance_m', localMaxDistance)
    }, 500)
    return () => clearTimeout(timer)
  }, [localMaxDistance, handleDistanceChange])

  useEffect(() => {
    const timer = setTimeout(() => {
      handleDurationChange('min_duration_s', localMinDuration)
    }, 500)
    return () => clearTimeout(timer)
  }, [localMinDuration, handleDurationChange])

  useEffect(() => {
    const timer = setTimeout(() => {
      handleDurationChange('max_duration_s', localMaxDuration)
    }, 500)
    return () => clearTimeout(timer)
  }, [localMaxDuration, handleDurationChange])

  const distanceUnit = getDistanceUnit(unitSystem)

  const hasActiveFilters =
    filters.sport_type ||
    filters.after ||
    filters.before ||
    filters.search ||
    filters.commute !== undefined ||
    filters.trainer !== undefined ||
    filters.gear_id ||
    filters.min_distance_m !== undefined ||
    filters.max_distance_m !== undefined ||
    filters.min_duration_s !== undefined ||
    filters.max_duration_s !== undefined

  return (
    <div className="space-y-4 rounded-lg border border-border bg-card p-4 mb-6">
      {/* Search input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search activities by name, description, location, hashtags..."
          className="pl-10 pr-10"
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
        />
        {localSearch && (
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setLocalSearch('')}
            className="absolute right-1 top-1/2 -translate-y-1/2"
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>

      {/* Search syntax help */}
      <div className="flex items-center gap-1 text-xs text-muted-foreground">
        <HelpCircle className="h-3 w-3" />
        <span>Tip: Search supports #hashtags, gear names, and location text</span>
      </div>

      {/* Quick sport filters */}
      <div className="flex flex-wrap items-center gap-2">
        <ToggleGroup
          type="single"
          value={filters.sport_type ?? ''}
          onValueChange={(value) => onFiltersChange({ sport_type: value || undefined, page: 1 })}
          variant="outline"
          size="sm"
        >
          <ToggleGroupItem value="">All</ToggleGroupItem>
          {SPORT_TYPES.slice(0, 5).map((type) => (
            <ToggleGroupItem key={type} value={type}>
              {type}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
        <Button
          variant={showAdvanced ? 'default' : 'outline'}
          size="sm"
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="gap-1"
        >
          <Filter className="h-3 w-3" />
          More
        </Button>
      </div>

      {/* Advanced filters */}
      {showAdvanced && (
        <div className="space-y-4 pt-2 border-t border-border">
          {/* Row 1: Sport type, Date from, Date to, Options */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Sport type dropdown */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">Sport Type</label>
              <Select
                value={filters.sport_type ?? '__all__'}
                onValueChange={(value) =>
                  onFiltersChange({ sport_type: value === '__all__' ? undefined : value, page: 1 })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">All types</SelectItem>
                  {SPORT_TYPES.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Date from */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">From</label>
              <Input
                type="date"
                value={filters.after ?? ''}
                onChange={(e) => onFiltersChange({ after: e.target.value || undefined, page: 1 })}
              />
            </div>

            {/* Date to */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">To</label>
              <Input
                type="date"
                value={filters.before ?? ''}
                onChange={(e) => onFiltersChange({ before: e.target.value || undefined, page: 1 })}
              />
            </div>

            {/* Toggles */}
            <div className="space-y-2">
              <label className="text-xs text-muted-foreground block">Options</label>
              <div className="flex flex-wrap gap-3">
                <label className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={filters.commute === true}
                    onCheckedChange={(checked) =>
                      onFiltersChange({
                        commute: checked === true ? true : undefined,
                        page: 1,
                      })
                    }
                  />
                  Commute
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={filters.trainer === true}
                    onCheckedChange={(checked) =>
                      onFiltersChange({
                        trainer: checked === true ? true : undefined,
                        page: 1,
                      })
                    }
                  />
                  Trainer
                </label>
              </div>
            </div>
          </div>

          {/* Row 2: Distance and Duration filters */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Min Distance */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">
                Min Distance ({distanceUnit})
              </label>
              <Input
                type="number"
                min="0"
                step="0.1"
                placeholder={`e.g., 10`}
                value={localMinDistance}
                onChange={(e) => setLocalMinDistance(e.target.value)}
              />
            </div>

            {/* Max Distance */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">
                Max Distance ({distanceUnit})
              </label>
              <Input
                type="number"
                min="0"
                step="0.1"
                placeholder={`e.g., 50`}
                value={localMaxDistance}
                onChange={(e) => setLocalMaxDistance(e.target.value)}
              />
            </div>

            {/* Min Duration */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">
                Min Duration (h:mm)
              </label>
              <Input
                type="text"
                placeholder="e.g., 1:00 or 30"
                value={localMinDuration}
                onChange={(e) => setLocalMinDuration(e.target.value)}
              />
            </div>

            {/* Max Duration */}
            <div>
              <label className="text-xs text-muted-foreground block mb-1">
                Max Duration (h:mm)
              </label>
              <Input
                type="text"
                placeholder="e.g., 2:30 or 120"
                value={localMaxDuration}
                onChange={(e) => setLocalMaxDuration(e.target.value)}
              />
            </div>
          </div>
        </div>
      )}

      {/* Reset filters */}
      {hasActiveFilters && (
        <div className="flex justify-end">
          <Button variant="ghost" size="sm" onClick={onReset}>
            <X className="h-4 w-4 mr-1" />
            Clear filters
          </Button>
        </div>
      )}
    </div>
  )
}
