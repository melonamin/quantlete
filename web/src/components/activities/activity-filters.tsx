import { useState, useEffect } from 'react'
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
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2 border-t border-border">
          {/* Sport type dropdown */}
          <div>
            <label className="text-xs text-muted-foreground block mb-1">Sport Type</label>
            <Select
              value={filters.sport_type ?? ''}
              onValueChange={(value) =>
                onFiltersChange({ sport_type: value || undefined, page: 1 })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="All types" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All types</SelectItem>
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
