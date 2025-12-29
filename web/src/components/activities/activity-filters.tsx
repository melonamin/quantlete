import { useState, useEffect } from 'react'
import { Search, X, Filter, HelpCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { ActivityFilters as Filters } from '@/lib/api'

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

export function ActivityFiltersPanel({ filters, onFiltersChange, onReset }: ActivityFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [localSearch, setLocalSearch] = useState(filters.search ?? '')

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== filters.search) {
        onFiltersChange({ search: localSearch || undefined, page: 1 })
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [localSearch, filters.search, onFiltersChange])

  const hasActiveFilters =
    filters.sport_type ||
    filters.after ||
    filters.before ||
    filters.search ||
    filters.commute !== undefined ||
    filters.trainer !== undefined ||
    filters.gear_id

  return (
    <div className="space-y-4 rounded-lg border border-border bg-card p-4 mb-6">
      {/* Search input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search activities by name, description, location, hashtags..."
          className="h-10 w-full rounded-md border border-border bg-background pl-10 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
        />
        {localSearch && (
          <button
            onClick={() => setLocalSearch('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Search syntax help */}
      <div className="flex items-center gap-1 text-xs text-muted-foreground">
        <HelpCircle className="h-3 w-3" />
        <span>Tip: Search supports #hashtags, gear names, and location text</span>
      </div>

      {/* Quick sport filters */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => onFiltersChange({ sport_type: undefined, page: 1 })}
          className={cn(
            'px-3 py-1.5 rounded-md text-sm transition-colors',
            !filters.sport_type
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground hover:bg-muted/80'
          )}
        >
          All
        </button>
        {SPORT_TYPES.slice(0, 5).map((type) => (
          <button
            key={type}
            onClick={() =>
              onFiltersChange({
                sport_type: filters.sport_type === type ? undefined : type,
                page: 1,
              })
            }
            className={cn(
              'px-3 py-1.5 rounded-md text-sm transition-colors',
              filters.sport_type === type
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-muted/80'
            )}
          >
            {type}
          </button>
        ))}
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className={cn(
            'px-3 py-1.5 rounded-md text-sm transition-colors flex items-center gap-1',
            showAdvanced
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground hover:bg-muted/80'
          )}
        >
          <Filter className="h-3 w-3" />
          More
        </button>
      </div>

      {/* Advanced filters */}
      {showAdvanced && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2 border-t border-border">
          {/* Sport type dropdown */}
          <div>
            <label className="text-xs text-muted-foreground block mb-1">Sport Type</label>
            <select
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={filters.sport_type ?? ''}
              onChange={(e) =>
                onFiltersChange({ sport_type: e.target.value || undefined, page: 1 })
              }
            >
              <option value="">All types</option>
              {SPORT_TYPES.map((type) => (
                <option key={type} value={type}>
                  {type}
                </option>
              ))}
            </select>
          </div>

          {/* Date from */}
          <div>
            <label className="text-xs text-muted-foreground block mb-1">From</label>
            <input
              type="date"
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={filters.after ?? ''}
              onChange={(e) => onFiltersChange({ after: e.target.value || undefined, page: 1 })}
            />
          </div>

          {/* Date to */}
          <div>
            <label className="text-xs text-muted-foreground block mb-1">To</label>
            <input
              type="date"
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={filters.before ?? ''}
              onChange={(e) => onFiltersChange({ before: e.target.value || undefined, page: 1 })}
            />
          </div>

          {/* Toggles */}
          <div className="space-y-2">
            <label className="text-xs text-muted-foreground block">Options</label>
            <div className="flex flex-wrap gap-3">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={filters.commute === true}
                  onChange={(e) =>
                    onFiltersChange({
                      commute: e.target.checked ? true : undefined,
                      page: 1,
                    })
                  }
                />
                Commute
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={filters.trainer === true}
                  onChange={(e) =>
                    onFiltersChange({
                      trainer: e.target.checked ? true : undefined,
                      page: 1,
                    })
                  }
                />
                Trainer
              </label>
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
